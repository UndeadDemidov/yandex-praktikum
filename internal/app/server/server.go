package server

import (
	"net/http"
	_ "net/http/pprof"

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

	r.Use(midware.Decompress)
	r.Use(midware.UserCookie)

	r.Group(func(r chi.Router) {
		r.Post("/", handler.HandlePostShortenPlain)
		r.Post("/api/shorten", handler.HandlePostShortenJSON)
		r.Get("/{id}", handler.HandleGet)
		r.Get("/ping", handler.HeartBeat)
		r.Delete("/api/user/urls", handler.HandleDelete)
		r.NotFound(handler.HandleNotFound)
		r.MethodNotAllowed(handler.HandleMethodNotAllowed)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Compress(5))
		r.Post("/api/shorten/batch", handler.HandlePostShortenBatch)
		r.Get("/api/user/urls", handler.HandleGetUserURLsBucket)
	})

	r.Mount("/", http.DefaultServeMux)

	s := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	return s
}
