package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	midware "github.com/UndeadDemidov/yandex-praktikum/internal/app/middleware"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
	"github.com/go-chi/chi/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var ErrUnableCreateShortID = errors.New("couldn't create unique ID in 10 tries")

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

	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	user := midware.GetUserID(ctx)
	shortenedURL, err := s.shorten(ctx, user, link)
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

	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	user := midware.GetUserID(ctx)
	shortenedURL, err := s.shorten(ctx, user, req.URL)
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
func (s URLShortener) shorten(ctx context.Context, user string, originalURL string) (shortenedURL string, err error) {
	id, err := s.createShortID(ctx)
	if err != nil {
		return "", err
	}
	err = s.linkRepo.Store(ctx, user, id, originalURL)
	if err != nil {
		return "", err
	}
	shortenedURL = fmt.Sprintf("%s%s", s.baseURL, id)
	return shortenedURL, nil
}

// createShortID создает короткий ID с проверкой на валидность
func (s URLShortener) createShortID(ctx context.Context) (id string, err error) {
	for i := 0; i < 10; i++ {
		id, err = gonanoid.New(8)
		if err != nil {
			return "", err
		}
		if !s.linkRepo.IsExist(ctx, id) {
			return id, nil
		}
	}
	return "", ErrUnableCreateShortID
}

// HandleGet - ручка для открытия по короткой ссылке
func (s URLShortener) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	u, err := s.linkRepo.Restore(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", u)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// HandleGetUserURLsBucket - ручка для получения всех ссылок пользователя
func (s URLShortener) HandleGetUserURLsBucket(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	user := midware.GetUserID(ctx)
	urlsMap := s.linkRepo.GetAllUserLinks(ctx, user)
	if len(urlsMap) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(MapToBucket(s.baseURL, urlsMap))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandlePostShortenBatch - ручка для создания коротких ссылок пакетом
// Оригинальные ссылки передаются через JSON Body
// ToDo есть желание переписать на stream
func (s URLShortener) HandlePostShortenBatch(w http.ResponseWriter, r *http.Request) {
	var req []URLShortenCorrelatedRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "proper JSON request is expected", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	user := midware.GetUserID(ctx)
	var resp []URLShortenCorrelatedResponse
	resp, err = s.shortenBatch(ctx, user, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s URLShortener) shortenBatch(ctx context.Context, user string, req []URLShortenCorrelatedRequest) (resp []URLShortenCorrelatedResponse, err error) {
	if req == nil || len(req) == 0 {
		return nil, errors.New("nothing to short")
	}

	batch := map[string]string{}
	resp = make([]URLShortenCorrelatedResponse, 0, len(req))
	for _, r := range req {
		id, err := s.createShortID(ctx)
		if err != nil {
			// вообще то не очень здорово так делать
			// лучше если просто continue - тогда из 100 ссылок сократиться 99,
			// а не отстрелится весь батч
			return nil, err
		}
		// ToDo Проверить что json корректный
		resp = append(resp, URLShortenCorrelatedResponse{
			CorrelationID: r.CorrelationID,
			ShortURL:      fmt.Sprintf("%s%s", s.baseURL, id),
		})
		batch[id] = r.OriginalURL
	}

	err = s.linkRepo.StoreBatch(ctx, user, batch)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// HeartBeat - ручка для проверки, что подключение к БД живое
func (s URLShortener) HeartBeat(w http.ResponseWriter, r *http.Request) {
	if s.database == nil {
		http.Error(w, "db is not initialized", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := s.database.PingContext(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("I'm alive (c)Helloween"))
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
	IsExist(ctx context.Context, id string) bool
	Store(ctx context.Context, user string, id string, link string) error
	Restore(ctx context.Context, id string) (link string, err error)
	Close() error
	GetAllUserLinks(ctx context.Context, user string) map[string]string
	StoreBatch(ctx context.Context, user string, batch map[string]string) error
}
