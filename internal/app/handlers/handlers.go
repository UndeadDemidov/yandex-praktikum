package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	midware "github.com/UndeadDemidov/yandex-praktikum/internal/app/middleware"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
	"github.com/go-chi/chi/v5"
)

// URLShortener - реализация интерфейса http.Handler
// Согласно заданию 1-го инкремента
type URLShortener struct {
	// non-persistent storage
	// just for starting
	linkRepo Repository
	baseURL  string
	database *sql.DB
}

// NewURLShortener создает URLShortener и инициализирует его
func NewURLShortener(base string, repo Repository, db *sql.DB) *URLShortener {
	h := URLShortener{}
	h.linkRepo = repo
	if utils.IsURL(base) {
		h.baseURL = fmt.Sprintf("%s/", strings.TrimRight(base, "/"))
	} else {
		h.baseURL = "http://localhost:8080/"
	}
	h.database = db

	return &h
}

// HandlePostShortenPlain - ручка для создания короткой ссылки
// Оригинальная ссылка передается через Text Body
func (s URLShortener) HandlePostShortenPlain(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// validate
	if len(b) == 0 {
		http.Error(w, "The link is not provided", http.StatusBadRequest)
		return
	}
	link := string(b)
	if !utils.IsURL(link) {
		http.Error(w, "Hey, Dude! Provide a link! Not the crap!", http.StatusBadRequest)
		return
	}

	user := midware.GetUserID(r.Context())
	shortenedURL, err := s.shorten(user, link)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortenedURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandlePostShortenJSON - ручка для создания короткой ссылки
// Оригинальная ссылка передается через JSON Body
func (s URLShortener) HandlePostShortenJSON(w http.ResponseWriter, r *http.Request) {
	req := URLShortenRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "JSON {\"url\":\"<some_url>\"} is expected", http.StatusBadRequest)
		return
	}
	if !utils.IsURL(req.URL) {
		http.Error(w, "Hey, Dude! Provide a link! Not the crap!", http.StatusBadRequest)
		return
	}

	user := midware.GetUserID(r.Context())
	shortenedURL, err := s.shorten(user, req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := URLShortenResponse{Result: shortenedURL}
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// shorten возвращает короткую ссылку в ответ на оригинальную
func (s URLShortener) shorten(user string, originalURL string) (shortenedURL string, err error) {
	id, err := utils.CreateShortID(s.linkRepo.IsExist)
	if err != nil {
		return "", err
	}
	err = s.linkRepo.Store(user, id, originalURL)
	if err != nil {
		return "", err
	}
	shortenedURL = fmt.Sprintf("%s%s", s.baseURL, id)
	return shortenedURL, nil
}

// HandleGet - ручка для открытия по короткой ссылке
func (s URLShortener) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u, err := s.linkRepo.Restore(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", u)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// HandleGetUserURLsBucket - ручка для получения всех ссылок пользователя
func (s URLShortener) HandleGetUserURLsBucket(w http.ResponseWriter, r *http.Request) {
	user := midware.GetUserID(r.Context())
	bucket := s.linkRepo.GetUserBucket(s.baseURL, user)
	if len(bucket) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(&bucket)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HeartBeat - ручка для открытия по короткой ссылке
func (s URLShortener) HeartBeat(w http.ResponseWriter, r *http.Request) {
	if s.database == nil {
		http.Error(w, "db is not initialized", http.StatusInternalServerError)
		return
	}

	var (
		cancel context.CancelFunc
		err    error
	)
	ctx := r.Context()
	ctx, cancel = context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err = s.database.PingContext(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("I'm alive (c)Helloween"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleMethodNotAllowed обрабатывает не валидный HTTP метод
func (s URLShortener) HandleMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Only GET and POST requests are allowed!", http.StatusMethodNotAllowed)
}

// HandleNotFound обрабатывает не найденный путь
func (s URLShortener) HandleNotFound(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `Only POST "/" with link in body and GET "/{short_link_id} are allowed" `, http.StatusNotFound)
}

// Repository описывает контракт работы с хранилищем.
// Используется для удобства тестирования и для дальнейшей легкой миграции на другой "движок".
type Repository interface {
	IsExist(id string) bool
	Store(user string, id string, link string) (err error)
	Restore(id string) (link string, err error)
	Close() error
	GetUserBucket(baseURL, user string) (bucket []BucketItem)
}

// BucketItem представляет собой структуру, в которой требуется сериализовать список ссылок
type BucketItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
