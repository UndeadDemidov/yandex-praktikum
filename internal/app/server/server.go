package server

import (
	"net/http"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	midware "github.com/UndeadDemidov/yandex-praktikum/internal/app/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewServer создает и возвращает новый сервер с указанным репозиторием коротких ссылок
// Эндпоинт POST / принимает в теле запроса строку URL для сокращения и возвращает ответ с кодом 201 и сокращённым URL в виде текстовой строки в теле.
// Эндпоинт POST /api/shorten принимает в теле запроса JSON {"url":"<some_url>"} для сокращения и возвращает ответ с кодом 201 и сокращённым URL в виде {"result":"<shorten_url>"}
// Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL и возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
// Нужно учесть некорректные запросы и возвращать для них ответ с кодом 400.
func NewServer(baseURL string, addr string, repo handlers.Repository) *http.Server {
	linkStore := repo
	handler := handlers.NewURLShortenerHandler(baseURL, linkStore)

	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(midware.Decompress)

	r.Post("/", handler.HandlePostShortenPlain)
	r.Post("/api/shorten", handler.HandlePostShortenJSON)
	r.Get("/{id}", handler.HandleGet)
	r.Get("/api/user/urls", handler.HandleGetUserURLsBucket)
	r.NotFound(handler.HandleNotFound)
	r.MethodNotAllowed(handler.HandleMethodNotAllowed)

	s := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	return s
}
