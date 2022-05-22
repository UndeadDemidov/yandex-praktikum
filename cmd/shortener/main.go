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
	srv, repo, db := CreateServer()
	Run(srv, repo, db)
}

func CreateServer() (*http.Server, handlers.Repository, *sql.DB) {
	var (
		srv  *http.Server
		db   *sql.DB
		repo handlers.Repository
		err  error
	)

	cs := viper.GetString("database-dsn")
	if len(cs) != 0 {
		db, err = sql.Open("postgres", cs)
		if err == nil {
			log.Print("Database is initialized")
		}
	}

	if repo, err = storages.NewFileStorage(viper.GetString("file-storage-path")); err != nil {
		repo = storages.NewLinkStorage()
		log.Print("In memory storage will be used")
	}
	srv = server.NewServer(viper.GetString("base-url"), viper.GetString("server-address"), repo, db)
	return srv, repo, db
}

func Run(srv *http.Server, repo handlers.Repository, db *sql.DB) {
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
			log.Printf("Caught an error due closing file:%+v", err)
		}
		if db != nil {
			err = db.Close()
			if err != nil {
				log.Printf("Caught an error due closing db:%+v", err)
			}
		}

		log.Println("Everything closed properly")
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server Shutdown Failed:%+v", err)
	}
	stop()
	log.Print("Server exited properly")
}
