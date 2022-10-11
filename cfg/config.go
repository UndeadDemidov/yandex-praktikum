package cfg

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultBaseURL       = "http://localhost:8080/"
	defaultServerAddress = ":8080"
)

type Config struct {
	ServerAddress   string `json:"server_address"`
	BaseUrl         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDsn     string `json:"database_dsn"`
	EnableHttps     bool   `json:"enable_https"`
}

func GetConfig() *Config {
	pflag.StringP("config", "c", "", "sets path to config file")
	pflag.StringP("base-url", "b", defaultBaseURL, "sets base URL for shortened link")
	pflag.StringP("server-address", "a", defaultServerAddress, "sets address of service server")
	pflag.StringP("file-storage-path", "f", "", "sets path for file storage")
	pflag.StringP("database-dsn", "d", "", "sets connection string for postgres DB")
	pflag.BoolP("enable-https", "s", false, "enable https protocol")
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatal().Err(err).Msgf("can't bind argument flags %v", pflag.CommandLine)
		return nil
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	cfg := new(Config)
	cfgPath := viper.GetString("config")
	if len(cfgPath) != 0 {
		cfg.loadConfigFromFile(cfgPath)
	}
	cfg.expandConfigFromFlags()

	return cfg
}

func (c *Config) loadConfigFromFile(filepath string) {
	log.Info().Msgf("trying to load config from file %s", filepath)
	file, err := os.Open(filepath)
	if err != nil {
		log.Err(err).Msgf("can't open given config file: %s", filepath)
		path, _ := os.Getwd()
		log.Err(err).Msgf("current work path is %s", path)
		return
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	}(file)

	err = json.NewDecoder(file).Decode(c)
	if err != nil {
		log.Err(err).Msgf("can't read config from given file: %s", filepath)
	}
}

func (c *Config) expandConfigFromFlags() {
	if viper.GetString("base-url") != defaultBaseURL {
		c.BaseUrl = viper.GetString("base-url")
	}
	if viper.GetString("server-address") != defaultServerAddress {
		c.ServerAddress = viper.GetString("server-address")
	}
	if viper.GetString("file-storage-path") != "" {
		c.FileStoragePath = viper.GetString("file-storage-path")
	}
	if viper.GetString("database-dsn") != "" {
		c.DatabaseDsn = viper.GetString("database-dsn")
	}
	if viper.GetBool("enable-https") {
		c.EnableHttps = viper.GetBool("enable-https")
	}
}
