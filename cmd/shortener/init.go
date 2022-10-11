package main

import (
	"database/sql"
	"os"

	"github.com/UndeadDemidov/yandex-praktikum/cfg"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages/database"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages/file"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages/memory"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	config = cfg.GetConfig()
	initRepository()
}

func initRepository() {
	var (
		err error
		db  *sql.DB
	)

	cs := config.DatabaseDsn
	if len(cs) != 0 {
		db, err = sql.Open("postgres", cs)
		if err != nil {
			db = nil
		}
	}

	if db != nil {
		repo, err = database.NewStorage(db)
		if err == nil {
			log.Info().Msg("In database storage will be used")
			return
		}
	}

	filename := config.FileStoragePath
	if len(filename) != 0 {
		repo, err = file.NewStorage(filename)
		if err == nil {
			log.Info().Msg("In file storage will be used")
			return
		}
	}

	repo = memory.NewStorage()
	log.Info().Msg("In memory storage will be used")
}
