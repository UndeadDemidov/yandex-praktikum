package storages

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	"log"
	"time"
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
	isExistQuery   = `SELECT COUNT(1) FROM shortened_urls WHERE id=$1`
	storeStatement = `INSERT INTO shortened_urls (id, user_id, original_url) VALUES ($1, $2, $3);`
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

func NewDBStorage(db *sql.DB) (st DBStorage, err error) {
	st = DBStorage{database: db}
	err = createDB(db)
	if err != nil {
		return DBStorage{}, err
	}
	return st, nil
}

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

func (d DBStorage) IsExist(ctx context.Context, id string) bool {
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

func (d DBStorage) Store(ctx context.Context, user string, id string, link string) (err error) {
	var actualID string
	err = d.database.QueryRowContext(ctx, storeQuery, id, user, link).Scan(&actualID)
	if errors.Is(err, sql.ErrNoRows) {
		// Если пустой сет записей, то успешно вставили запись
		return nil
	}
	if err != nil {
		return err
	}

	return handlers.NewUniqueIDViolatedError(errors.New("link is in database already"), map[string]string{id: actualID})
}

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

func (d DBStorage) StoreBatch(ctx context.Context, user string, batch map[string]string) error {
	// шаг 1 — объявляем транзакцию
	tx, err := d.database.Begin()
	if err != nil {
		return err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer tx.Rollback()

	// шаг 2 — готовим инструкцию
	stmt, err := tx.PrepareContext(ctx, storeStatement)
	if err != nil {
		return err
	}
	// шаг 2.1 — не забываем закрыть инструкцию, когда она больше не нужна
	defer stmt.Close()

	for k, v := range batch {
		// шаг 3 — указываем, что каждый элемент будет добавлен в транзакцию
		if _, err = stmt.ExecContext(ctx, k, user, v); err != nil {
			return err
		}
	}
	// шаг 4 — сохраняем изменения
	return tx.Commit()
}

func (d DBStorage) Close() error {
	return d.database.Close()
}
