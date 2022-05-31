package main

import (
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
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
}
