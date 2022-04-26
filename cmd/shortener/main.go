package main

import "net/http"

func main() {
	hand := NewURLShortenerHandler()
	// маршрутизация запросов обработчику
	http.Handle("/", hand)
	// запуск сервера с адресом localhost, порт 8080
	_ = http.ListenAndServe(":8080", nil)
}
