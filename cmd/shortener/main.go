package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UndeadDemidov/yandex-praktikum/cfg"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/server"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
	repo         handlers.Repository
	config       *cfg.Config
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	srv := CreateServer()
	g := CreateGRPCServer()
	Run(srv, g)
}

// CreateServer создает сервер и возвращает его и репозиторий.
// Можно заменить параметры на глобальные переменные, вроде как от этого ничего плохого не будет.
func CreateServer() *http.Server {
	return server.NewServer(config.BaseUrl, config.ServerAddress, config.TrustedSubnet, repo)
}

// CreateGRPCServer создает grpc сервер и возвращает его и репозиторий.
func CreateGRPCServer() *grpc.Server {
	return server.NewGRPCServer(config.BaseUrl, repo)
}

// Run запускает сервер с указанным репозиторием и реализуем graceful shutdown
// Более читаемый вариант: https://github.com/rudderlabs/graceful-shutdown-examples/blob/main/httpserver/main.go
func Run(srv *http.Server, grpc *grpc.Server) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	log.Info().Msg("Start HTTP server...")
	go func() {
		const (
			cert = "cert.pem"
			key  = "key.pem"
		)
		var err error
		if config.EnableHttps {
			log.Info().Msg("HTTPS enabled")
			err = utils.CreateTLSCert(cert, key)
			if err != nil {
				log.Fatal().Msgf("cert creation: %+v\n", err)
			}
			err = srv.ListenAndServeTLS(cert, key)
		} else {
			log.Info().Msg("HTTPS is not enabled")
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("listen: %+v\n", err)
		}
	}()

	log.Info().Msg("Start GRPC server...")
	go func() {
		listen, err := net.Listen("tcp", ":3200")
		if err != nil {
			log.Fatal().Msgf("listen: %+v\n", err)
		}
		if err = grpc.Serve(listen); err != nil {
			log.Fatal().Msgf("serve: %+v\n", err)
		}
	}()

	<-ctx.Done()

	log.Info().Msg("Servers stopping...")
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
	grpc.GracefulStop()
	stop()
	log.Info().Msg("Servers exited properly")
}
