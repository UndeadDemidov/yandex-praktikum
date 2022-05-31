package server

import (
	"net/http"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	midware "github.com/UndeadDemidov/yandex-praktikum/internal/app/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewServer создает и возвращает новый сервер с указанным репозиторием коротких ссылок
func NewServer(baseURL string, addr string, repo handlers.Repository) *http.Server {
	linkStore := repo
	handler := handlers.NewURLShortener(baseURL, linkStore)

	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(midware.Decompress)
	r.Use(midware.UserCookie)

	r.Post("/", handler.HandlePostShortenPlain)
	r.Post("/api/shorten", handler.HandlePostShortenJSON)
	r.Post("/api/shorten/batch", handler.HandlePostShortenBatch)
	r.Get("/{id}", handler.HandleGet)
	r.Get("/api/user/urls", handler.HandleGetUserURLsBucket)
	r.Get("/ping", handler.HeartBeat)
	r.NotFound(handler.HandleNotFound)
	r.MethodNotAllowed(handler.HandleMethodNotAllowed)

	s := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	return s
}
