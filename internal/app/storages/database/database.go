package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
	"github.com/rs/zerolog/log"
)

const (
	checkQuery = `SELECT COUNT(1)
					FROM information_schema.tables
				   WHERE table_schema = 'public'
					 AND table_type = 'BASE TABLE'
					 AND table_name = 'shortened_urls'`
	createDBStatement = `CREATE TABLE shortened_urls
						(
						    id           VARCHAR NOT NULL CONSTRAINT shortened_urls_pk PRIMARY KEY,
						    user_id      uuid    NOT NULL,
						    original_url VARCHAR NOT NULL,
						    is_deleted   BOOLEAN NOT NULL DEFAULT FALSE
						);
						CREATE UNIQUE INDEX shortened_urls_id_uindex ON shortened_urls (id);
						CREATE UNIQUE INDEX shortened_urls_original_url_uindex ON shortened_urls (original_url);
						CREATE INDEX shortened_urls_user_id ON shortened_urls (user_id);`
	// Как говорит великий Том Кайт - если можно сделать одним SQL statement - сделай это!
	// Если original_url уже есть, то возвращается его ID (независимо от user_id),
	// Если original_url еще нет, то возвращается пустой row set
	storeQuery = `WITH inserted_rows AS (
						INSERT INTO shortened_urls (id, user_id, original_url)
        				VALUES ($1, $2, $3)
        				ON CONFLICT (original_url) DO NOTHING
						RETURNING id
					  )
						SELECT id
						FROM shortened_urls
   						 WHERE NOT EXISTS (SELECT 1 FROM inserted_rows)
   						   AND original_url=$3;`
	restoreQuery    = `SELECT original_url, is_deleted FROM shortened_urls WHERE id=$1`
	deleteStatement = `UPDATE shortened_urls SET is_deleted=true WHERE user_id=$1 AND id=$2`
	userBucketQuery = `SELECT id, original_url FROM shortened_urls WHERE user_id=$1`

	batchSize = 10
)

// Storage реализует хранение ссылок в файле.
// Выполнена простейшая реализация для сдачи работы.
type Storage struct {
	database *sql.DB
	delBatch chan userID
	done     chan bool
}

var _ handlers.Repository = (*Storage)(nil)

// NewStorage cоздает и возвращает экземпляр Storage
func NewStorage(db *sql.DB) (st *Storage, err error) {
	st = &Storage{database: db}
	err = createDB(db)
	if err != nil {
		return &Storage{}, err
	}

	st.delBatch = make(chan userID)
	st.done = make(chan bool)
	// Запускаем единственный consumer fanIn, в теории можно сделать пул consumers
	// ToDo Нужно вырезать слой Service и там делать метод Run, где и будет запущен этот consumer
	go st.unstoreConsume()

	return st, nil
}

// createDB проверяет, есть ли уже необходимая структура БД и создает ее, если нет
func createDB(db *sql.DB) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return err
	}

	var cnt int64
	err = db.QueryRowContext(ctx, checkQuery).Scan(&cnt)
	if err != nil {
		return err
	}

	if cnt == 0 {
		_, err = db.ExecContext(ctx, createDBStatement)
		if err != nil {
			return err
		}
	}
	return nil
}

// Store сохраняет ссылку в хранилище с указанным id. В случае конфликта c уже ранее сохраненным link
// возвращает ошибку handlers.ErrLinkIsAlreadyShortened и id с раннего сохранения.
func (s *Storage) Store(ctx context.Context, user string, link string) (id string, err error) {
	var actualID string
	// две попытки для генерации уникального id
	for i := 0; i < 2; i++ {
		id = utils.NewUniqueID()
		err = s.database.QueryRowContext(ctx, storeQuery, id, user, link).Scan(&actualID)
		if err == nil || errors.Is(err, sql.ErrNoRows) {
			break
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		// Если пустой сет записей, то успешно вставили запись
		return id, nil
	}
	if err != nil {
		return "", err
	}

	return actualID, handlers.ErrLinkIsAlreadyShortened
}

// Restore возвращает исходную ссылку по переданному короткому ID
func (s *Storage) Restore(ctx context.Context, id string) (link string, err error) {
	var deleted bool
	err = s.database.QueryRowContext(ctx, restoreQuery, id).Scan(&link, &deleted)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf(storages.ErrLinkNotFound, id)
	case err != nil:
		return "", err
	case deleted:
		return "", handlers.ErrLinkIsDeleted
	}

	return
}

// Unstore - помечает список ранее сохраненных ссылок удаленными
// только тех ссылок, которые принадлежат пользователю
func (s *Storage) Unstore(ctx context.Context, user string, ids []string) {
	ch := make(chan userID)
	go s.unstoreProduce(ctx, ch, user, ids)
	// на каждого продюсера один воркер
	// ToDo можно сделать пул воркеров
	go s.unstoreWork(ch)
}

func (s *Storage) unstoreProduce(_ context.Context, ch chan userID, user string, ids []string) {
	// Делаем for и шлем каждый элемент в channel.
	// Что успеет заслаться - то и обработается.
	for i, id := range ids {
		ch <- userID{User: user, ID: id}
		log.Debug().Msgf("%v", i)
	}
	close(ch)
}

func (s *Storage) unstoreWork(ch chan userID) {
	for uID := range ch {
		s.delBatch <- uID
	}
}

// unstoreConsume собирает пакет определенного размера и выталкивает в БД.
// Чтобы хвосты неполных пакетов не застревали, регулярно делаем flush
func (s *Storage) unstoreConsume() {
	flush := func() {
		for {
			time.Sleep(time.Second)
			s.done <- true
		}
	}

	go flush()

	var buf = make([]userID, batchSize)
	i := 0
	for {
		select {
		case <-s.done:
			if i != 0 {
				log.Debug().Msg(fmt.Sprint(buf[:i]))
				err := s.unstoreBatch(buf[:i])
				if err != nil {
					log.Err(err)
				}
				i = 0
			}
		case id, ok := <-s.delBatch:
			if !ok {
				return
			}
			if i == len(buf) {
				log.Debug().Msg(fmt.Sprint(buf))
				err := s.unstoreBatch(buf)
				if err != nil {
					log.Err(err)
				}
				i = 0
			}
			buf[i] = id
			i++
		}
	}
}

func (s *Storage) unstoreBatch(ids []userID) error {
	// шаг 1 — объявляем транзакцию
	tx, err := s.database.Begin()
	if err != nil {
		return err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer func() {
		if err = tx.Rollback(); err != nil {
			log.Err(err)
		}
	}()

	// Это чтобы мы тут тоже не зависли надолго
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	// шаг 2 — готовим инструкцию
	stmt, err := tx.PrepareContext(ctx, deleteStatement)
	if err != nil {
		return err
	}
	// шаг 2.1 — не забываем закрыть инструкцию, когда она больше не нужна
	defer func() {
		if err = stmt.Close(); err != nil {
			log.Err(err)
		}
	}()

	// шаг 3 - выполняем задачу
	for _, id := range ids {
		_, err = stmt.ExecContext(ctx, id.User, id.ID)
		if err != nil {
			return err
		}
	}

	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// GetUserStorage возвращает map[id]link ранее сокращенных ссылок указанным пользователем
func (s *Storage) GetUserStorage(ctx context.Context, user string) map[string]string {
	rows, err := s.database.QueryContext(ctx, userBucketQuery, user)
	if err != nil {
		log.Err(err)
		return map[string]string{}
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			log.Err(err)
		}
	}()

	m := map[string]string{}
	for rows.Next() {
		var (
			id          string
			originalURL string
		)
		err = rows.Scan(&id, &originalURL)
		if err != nil {
			log.Err(err)
			return map[string]string{}
		}
		m[id] = originalURL
	}

	err = rows.Err()
	if err != nil {
		log.Err(err)
		return map[string]string{}
	}

	return m
}

// StoreBatch сохраняет пакет ссылок из map[correlation_id]original_link и возвращает map[correlation_id]short_link.
// В случае конфликта c уже ранее сохраненным link возвращает ошибку handlers.ErrLinkIsAlreadyShortened и id с раннего сохранения.
func (s *Storage) StoreBatch(ctx context.Context, user string, batchIn map[string]string) (batchOut map[string]string, err error) {
	// шаг 1 — объявляем транзакцию
	tx, err := s.database.Begin()
	if err != nil {
		return nil, err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer func() {
		if err = tx.Rollback(); err != nil {
			log.Err(err)
		}
	}()

	// шаг 2 — готовим инструкцию
	query, err := tx.PrepareContext(ctx, storeQuery)
	if err != nil {
		return nil, err
	}
	// шаг 2.1 — не забываем закрыть инструкцию, когда она больше не нужна
	defer func() {
		if err = query.Close(); err != nil {
			log.Err(err)
		}
	}()

	batchOut = make(map[string]string)
	conflict := false
	for corrID, link := range batchIn {
		// шаг 3 — указываем, что каждый элемент будет добавлен в транзакцию
		var (
			id       string
			actualID string
		)
		// две попытки для генерации уникального id
		for i := 0; i < 2; i++ {
			id = utils.NewUniqueID()
			err = query.QueryRowContext(ctx, id, user, link).Scan(&actualID)
			if err == nil || errors.Is(err, sql.ErrNoRows) {
				break
			}
		}

		if errors.Is(err, sql.ErrNoRows) {
			// Если пустой сет записей, то успешно вставили запись
			batchOut[corrID] = id
			continue
		}
		if err != nil {
			return nil, err
		}
		batchOut[corrID] = actualID
		conflict = true
	}

	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if conflict {
		err = handlers.ErrLinkIsAlreadyShortened
	}
	return batchOut, err // err либо nil, либо ErrLinkIsAlreadyShortened
}

// Ping проверяет доступность БД
func (s *Storage) Ping(ctx context.Context) error {
	return s.database.PingContext(ctx)
}

// Close закрывает базу данных
func (s *Storage) Close() error {
	s.done <- true
	// важен порядок закрытия!
	close(s.delBatch)
	close(s.done)
	return s.database.Close()
}

type userID struct {
	User string
	ID   string
}
