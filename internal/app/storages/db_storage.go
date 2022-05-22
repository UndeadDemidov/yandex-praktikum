package storages

import (
	"context"
	"database/sql"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	"time"
)

const (
	checkQuery = `SELECT count(1)
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
						create index shortened_urls_user_id on shortened_urls (user_id);`
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
	row := db.QueryRowContext(ctx, checkQuery)
	err = row.Scan(&cnt)
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

func (d DBStorage) IsExist(id string) bool {
	//TODO implement me
	panic("implement me")
}

func (d DBStorage) Store(user string, id string, link string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (d DBStorage) Restore(id string) (link string, err error) {
	//TODO implement me
	panic("implement me")
}

func (d DBStorage) Close() error {
	return d.database.Close()
}

func (d DBStorage) GetUserBucket(baseURL, user string) (bucket []handlers.BucketItem) {
	//TODO implement me
	panic("implement me")
}
