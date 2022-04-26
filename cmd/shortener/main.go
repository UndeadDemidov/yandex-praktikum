package main

import (
	"github.com/UndeadDemidov/yandex-praktikum/internal/app"
	"log"
)

func main() {
	s := app.NewServer(":8080")
	log.Fatalln(s.ListenAndServe())
}
