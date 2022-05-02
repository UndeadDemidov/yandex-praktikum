package main

import (
	"github.com/UndeadDemidov/yandex-praktikum/internal/app"
	"log"
)

func main() {
	s := app.NewServerWithBuiltinRepository("http://localhost:8080/", ":8080")
	log.Fatalln(s.ListenAndServe())
}
