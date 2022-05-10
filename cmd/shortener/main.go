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

func main() {
	var s *http.Server
	var r handlers.Repository
	var err error

	if r, err = storages.NewFileStorage(viper.GetString("file-storage-path")); err == nil {
	} else {
		r = storages.NewLinkStorage()
		log.Print("In memory storage will be used")
	}
	s = server.NewServer(viper.GetString("base-url"), viper.GetString("server-address"), r)

	// Пришлось сделать graceful shutdown, чтобы правильно закрывать файлы при остановке сервера
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Print("Server started")

	<-done
	log.Print("Server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		_ = r.Close()
		cancel()
	}()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server exited properly")
}
