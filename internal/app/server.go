package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

// NewServer создает и возвращает новый сервер
// Эндпоинт POST / принимает в теле запроса строку URL для сокращения и возвращает ответ с кодом 201 и сокращённым URL в виде текстовой строки в теле.
// Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL и возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
// Нужно учесть некорректные запросы и возвращать для них ответ с кодом 400.
func NewServer(baseURL string, addr string) *http.Server {
	linkStore := NewLinkStorage()
	handler := NewURLShortenerHandler(baseURL, linkStore)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{shortID}", handler.GetHandler())
	r.Post("/", handler.PostHandler())

	s := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	return s
}
