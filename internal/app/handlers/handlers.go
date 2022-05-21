package handlers

import (
	"encoding/json"
	"fmt"
	midware "github.com/UndeadDemidov/yandex-praktikum/internal/app/middleware"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"strings"
)

// URLShortenerHandler - реализация интерфейса http.Handler
// Согласно заданию 1-го инкремента
type URLShortenerHandler struct {
	// non-persistent storage
	// just for starting
	linkRepo Repository
	baseURL  string
}

// NewURLShortenerHandler создает URLShortenerHandler и инициализирует его
func NewURLShortenerHandler(base string, repo Repository) *URLShortenerHandler {
	h := URLShortenerHandler{}
	h.linkRepo = repo
	if utils.IsURL(base) {
		h.baseURL = fmt.Sprintf("%s/", strings.TrimRight(base, "/"))
	} else {
		h.baseURL = "http://localhost:8080/"
	}
	return &h
}

// HandlePostShortenPlain - ручка для создания короткой ссылки
// Оригинальная ссылка передается через Text Body
func (s URLShortenerHandler) HandlePostShortenPlain(w http.ResponseWriter, r *http.Request) {
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
func (s URLShortenerHandler) HandlePostShortenJSON(w http.ResponseWriter, r *http.Request) {
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
func (s URLShortenerHandler) shorten(user string, originalURL string) (shortenedURL string, err error) {
	log.Println("store with user: ", user)

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
func (s URLShortenerHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	user := midware.GetUserID(r.Context())
	log.Println("expand with user: ", user)

	id := chi.URLParam(r, "id")
	u, err := s.linkRepo.Restore(user, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", u)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// HandleGetUserURLsBucket - ручка для получения всех ссылок пользователя
func (s URLShortenerHandler) HandleGetUserURLsBucket(w http.ResponseWriter, r *http.Request) {
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

// HandleMethodNotAllowed обрабатывает не валидный HTTP метод
func (s URLShortenerHandler) HandleMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Only GET and POST requests are allowed!", http.StatusMethodNotAllowed)
}

// HandleNotFound обрабатывает не найденный путь
func (s URLShortenerHandler) HandleNotFound(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `Only POST "/" with link in body and GET "/{short_link_id} are allowed" `, http.StatusNotFound)
}

// Repository описывает контракт работы с хранилищем.
// Используется для удобства тестирования и для дальнейшей легкой миграции на другой "движок".
type Repository interface {
	IsExist(id string) bool
	Store(user string, id string, link string) (err error)
	Restore(user string, id string) (link string, err error)
	Close() error
	GetUserBucket(baseURL, user string) (bucket []BucketItem)
}

// BucketItem представляет собой структуру, в которой требуется сериализовать список ссылок
type BucketItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
