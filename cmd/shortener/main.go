package main

import (
	"log"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080/"`
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:":8080"`
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	s := app.NewServerWithBuiltinRepository(cfg.BaseURL, cfg.ServerAddress)
	log.Fatalln(s.ListenAndServe())
}
