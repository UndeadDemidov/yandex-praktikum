package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/server"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func main() {
	srv, repo := CreateServer()
	Run(srv, repo)
}

// CreateServer создает сервер и возвращает его и репозиторий.
// Можно заменить параметры на глобальные переменные, вроде как от этого ничего плохого не будет.
func CreateServer() (*http.Server, handlers.Repository) {
	var (
		srv  *http.Server
		db   *sql.DB
		repo handlers.Repository
		err  error
	)

	cs := viper.GetString("database-dsn")
	if len(cs) != 0 {
		db, err = sql.Open("postgres", cs)
		if err != nil {
			db = nil
		}
	}

	repo = chooseRepo(viper.GetString("file-storage-path"), db)
	srv = server.NewServer(viper.GetString("base-url"), viper.GetString("server-address"), repo, db)
	return srv, repo
}

// chooseRepo возвращает сервер в зависимости от того, какие параметры были переданы
func chooseRepo(filename string, db *sql.DB) (repo handlers.Repository) {
	var err error

	if db != nil {
		repo, err = storages.NewDBStorage(db)
		if err == nil {
			log.Print("In database storage will be used")
			return repo
		}
	}

	if len(filename) != 0 {
		repo, err = storages.NewFileStorage(filename)
		if err == nil {
			log.Print("In file storage will be used")
			return repo
		}
	}

	log.Print("In memory storage will be used")
	return storages.NewLinkStorage()
}

// Run запускает сервер с указанным репозиторием и реализуем graceful shutdown
func Run(srv *http.Server, repo handlers.Repository) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %+v\n", err)
		}
	}()
	log.Print("Server started")

	<-ctx.Done()

	log.Print("Server stopped")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		err := repo.Close()
		if err != nil {
			log.Printf("Caught an error due closing repository:%+v", err)
		}

		log.Println("Everything is closed properly")
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server Shutdown Failed:%+v", err)
	}
	stop()
	log.Print("Server exited properly")
}
