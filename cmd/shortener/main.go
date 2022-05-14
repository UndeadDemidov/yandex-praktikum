package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/server"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages"
	"github.com/spf13/viper"
)

func main() {
	srv, repo := CreateServer()
	Run(srv, repo)
}

func CreateServer() (*http.Server, handlers.Repository) {
	var (
		srv  *http.Server
		repo handlers.Repository
		err  error
	)

	if repo, err = storages.NewFileStorage(viper.GetString("file-storage-path")); err != nil {
		repo = storages.NewLinkStorage()
		log.Print("In memory storage will be used")
	}
	srv = server.NewServer(viper.GetString("base-url"), viper.GetString("server-address"), repo)
	return srv, repo
}

func Run(srv *http.Server, repo handlers.Repository) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %+v\n", err)
		}
	}()
	log.Print("Server started")

	if <-ctx.Done(); true {
		log.Print("Server stopped")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			err := repo.Close()
			if err != nil {
				log.Printf("Caught an error due closing file:%+v", err)
			}
			cancel()
		}()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server Shutdown Failed:%+v", err)
		}
		stop()
		log.Print("Server exited properly")
	}
}
