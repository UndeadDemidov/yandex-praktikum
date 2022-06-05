package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/server"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var repo handlers.Repository

func main() {
	srv := CreateServer()
	Run(srv)
}

// CreateServer создает сервер и возвращает его и репозиторий.
// Можно заменить параметры на глобальные переменные, вроде как от этого ничего плохого не будет.
func CreateServer() *http.Server {
	return server.NewServer(viper.GetString("base-url"), viper.GetString("server-address"), repo)
}

// Run запускает сервер с указанным репозиторием и реализуем graceful shutdown
func Run(srv *http.Server) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("listen: %+v\n", err)
		}
	}()
	log.Info().Msg("Server started")

	<-ctx.Done()

	log.Info().Msg("Server stopped")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		err := repo.Close()
		if err != nil {
			log.Error().Msgf("Caught an error due closing repository:%+v", err)
		}

		log.Info().Msg("Everything is closed properly")
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Msgf("Server Shutdown Failed:%+v", err)
	}
	stop()
	log.Info().Msg("Server exited properly")
}
