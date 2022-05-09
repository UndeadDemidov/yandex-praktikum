package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/server"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages"
)

type Config struct {
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080/"`
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	var s *http.Server
	var r handlers.Repository
	if r, err = storages.NewFileStorage(cfg.FileStoragePath); err == nil {
	} else {
		r = storages.NewLinkStorage()
	}
	s = server.NewServer(cfg.BaseURL, cfg.ServerAddress, r)

	// Пришлось сделать graceful shutdown, чтобы правильно закрывать файлы при остановке сервера
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Print("Server Started")

	<-done
	log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		_ = r.Close()
		cancel()
	}()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")
}
