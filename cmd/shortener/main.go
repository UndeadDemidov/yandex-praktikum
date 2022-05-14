package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/server"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages"
)

// ToDo Поленился вынести все в что-нибудь типа Execute()
// Сделаю в первый же выходной перед 3-м спринтом
func main() {
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

	// Пришлось сделать graceful shutdown, чтобы правильно закрывать файлы при остановке сервера
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %+v\n", err)
		}
	}()
	log.Print("Server started")

	<-done
	log.Print("Server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		err := repo.Close()
		if err != nil {
			log.Printf("Caught error due closing file:%+v", err)
		}
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server exited properly")
}
