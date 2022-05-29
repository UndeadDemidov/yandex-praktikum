package storages

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
)

const (
	checkQuery = `SELECT COUNT(1)
					FROM information_schema.tables
				   WHERE table_schema = 'public'
					 AND table_type = 'BASE TABLE'
					 AND table_name = 'shortened_urls'`
	createDBStatement = `create table shortened_urls
						(
						    id           VARCHAR not null constraint shortened_urls_pk primary key,
						    user_id      uuid    not null,
						    original_url VARCHAR not null
						);
						create unique index shortened_urls_id_uindex on shortened_urls (id);
						create unique index shortened_urls_original_url_uindex on shortened_urls (original_url);
						create index shortened_urls_user_id on shortened_urls (user_id);`
	isExistQuery = `SELECT COUNT(1) FROM shortened_urls WHERE id=$1`
	// Как говорит великий Том Кайт - если можно сделать одним SQL statement - сделай это!
	// Если original_url уже есть, то возвращается его ID (независимо от user_id),
	// Если
	storeQuery = `WITH inserted_rows AS (
						INSERT INTO shortened_urls (id, user_id, original_url)
        				VALUES ($1, $2, $3)
        				ON CONFLICT (original_url) DO NOTHING
						RETURNING id
					  )
						SELECT id
						FROM shortened_urls
   						 WHERE NOT exists (SELECT 1 FROM inserted_rows)
   						   AND original_url=$3;`
	restoreQuery    = `SELECT original_url FROM shortened_urls WHERE id=$1`
	userBucketQuery = `SELECT id, original_url FROM shortened_urls WHERE user_id=$1`
)

// DBStorage реализует хранение ссылок в файле.
// Выполнена простейшая реализация для сдачи работы.
type DBStorage struct {
	database *sql.DB
}

var _ handlers.Repository = (*DBStorage)(nil)

// NewDBStorage cоздает и возвращает экземпляр DBStorage
func NewDBStorage(db *sql.DB) (st *DBStorage, err error) {
	st = &DBStorage{database: db}
	err = createDB(db)
	if err != nil {
		return &DBStorage{}, err
	}
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

// IsExist проверяет наличие id в базе
func (d DBStorage) isExist(ctx context.Context, id string) bool {
	var cnt int64
	err := d.database.QueryRowContext(ctx, isExistQuery, id).Scan(&cnt)
	if err != nil {
		// после 10 попытке свалиться в генерации уникального ID
		return true
	}

	if cnt != 0 {
		return true
	}
	return false
}

// Store сохраняет ссылку в хранилище с указанным id. В случае конфликта c уже ранее сохраненным link
// возвращает ошибку handlers.ErrLinkIsAlreadyShortened и id с раннего сохранения.
// ToDo Вообще это не очень чистая реализация в Go, потому что в случае ошибки прозрачней указывать id пустой.
// Или я не прав - и для Go так нормально???
func (d DBStorage) Store(ctx context.Context, user string, link string) (id string, err error) {
	var actualID string
	// две попытки для генерации уникального id
	for i := 0; i < 2; i++ {
		id = utils.NewUniqueID()
		err = d.database.QueryRowContext(ctx, storeQuery, id, user, link).Scan(&actualID)
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
func (d DBStorage) Restore(ctx context.Context, id string) (link string, err error) {
	err = d.database.QueryRowContext(ctx, restoreQuery, id).Scan(&link)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf(ErrLinkNotFound, id)
	}
	if err != nil {
		return "", err
	}
	return
}

// GetAllUserLinks возвращает map[id]link ранее сокращенных ссылок указанным пользователем
func (d DBStorage) GetAllUserLinks(ctx context.Context, user string) map[string]string {
	rows, err := d.database.QueryContext(ctx, userBucketQuery, user)
	if err != nil {
		log.Println(err)
		return map[string]string{}
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Println(err)
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
			log.Println(err)
			return map[string]string{}
		}
		m[id] = originalURL
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return map[string]string{}
	}

	return m
}

// StoreBatch сохраняет пакет ссылок из map[correlation_id]original_link и возвращает map[correlation_id]short_link.
// В случае конфликта c уже ранее сохраненным link возвращает ошибку handlers.ErrLinkIsAlreadyShortened и id с раннего сохранения.
func (d DBStorage) StoreBatch(ctx context.Context, user string, batchIn map[string]string) (batchOut map[string]string, err error) {
	// шаг 1 — объявляем транзакцию
	tx, err := d.database.Begin()
	if err != nil {
		return nil, err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer tx.Rollback()

	// шаг 2 — готовим инструкцию
	query, err := tx.PrepareContext(ctx, storeQuery)
	if err != nil {
		return nil, err
	}
	// шаг 2.1 — не забываем закрыть инструкцию, когда она больше не нужна
	defer query.Close()

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

// Close закрывает базу данных
func (d DBStorage) Close() error {
	return d.database.Close()
}
