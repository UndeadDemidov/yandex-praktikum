package main

import (
	"database/sql"
	"strings"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages/database"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages/file"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages/memory"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.With().Caller().Logger()

	pflag.StringP("base-url", "b", "http://localhost:8080/", "sets base URL for shortened link")
	pflag.StringP("server-address", "a", ":8080", "sets address of service server")
	pflag.StringP("file-storage-path", "f", "", "sets path for file storage")
	pflag.StringP("database-dsn", "d", "", "sets connection string for postgres DB")
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatal().Err(err).Msgf("can't bind argument flags %v", pflag.CommandLine)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	initRepository()
}

func initRepository() {
	var (
		err error
		db  *sql.DB
	)

	cs := viper.GetString("database-dsn")
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

	filename := viper.GetString("file-storage-path")
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
