package app

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strings"
)

// URLShortenerHandler - реализация интерфейса http.Handler
// Согласно заданию 1-го инкремента
type URLShortenerHandler struct {
	//non-persistent storage
	//just for starting
	linkRepo Repository
	baseURL  string
}

// Repository описывает контракт работы с хранилищем.
// Используется для удобства тестирования и для дальнейшей легкой мигарции на другой "движок".
type Repository interface {
	IsExist(id string) bool
	Store(id string, link string) (err error)
	Restore(id string) (link string, err error)
}

// NewURLShortenerHandler создает URLShortenerHandler и инициализирует его
func NewURLShortenerHandler(base string, repo Repository) *URLShortenerHandler {
	h := URLShortenerHandler{}
	h.linkRepo = repo
	if IsURL(base) {
		h.baseURL = fmt.Sprintf("%s/", strings.TrimRight(base, "/"))
	} else {
		h.baseURL = "http://localhost:8080/"
	}
	return &h
}

// ServeHTTP - реализация метода интефрейса http.Handler
func (s URLShortenerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.HandleGet(w, r)
	case http.MethodPost:
		s.HandlePost(w, r)
	default:
		s.HandleMethodNotAllowed(w, r)
	}
}

// HandlePost - ручка для создания короткой ссылки
// Оригинальная ссылка передается через Text Body
func (s URLShortenerHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
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
	if !IsURL(link) {
		http.Error(w, "Hey, Dude! Provide a link! Not the crap!", http.StatusBadRequest)
		return
	}

	shortenedURL, err := s.shorten(link)
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

// HandlePostShorten - ручка для создания короткой ссылки
// Оригинальная ссылка передается через JSON Body
func (s URLShortenerHandler) HandlePostShorten(w http.ResponseWriter, r *http.Request) {
	req := URLShortenRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "JSON {\"url\":\"<some_url>\"} is expected", http.StatusBadRequest)
		return
	}
	if !IsURL(req.URL) {
		http.Error(w, "Hey, Dude! Provide a link! Not the crap!", http.StatusBadRequest)
		return
	}

	shortenedURL, err := s.shorten(req.URL)
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

// shorten - возвращает короткую ссылку в ответ на оригинальную
func (s URLShortenerHandler) shorten(originalURL string) (shortenedURL string, err error) {
	id, err := CreateShortID(s.linkRepo.IsExist)
	if err != nil {
		return "", err
	}
	err = s.linkRepo.Store(id, originalURL)
	if err != nil {
		return "", err
	}
	shortenedURL = fmt.Sprintf("%s%s", s.baseURL, id)
	return shortenedURL, nil
}

// HandleGet - ручка для для открытия по короткой ссылке
func (s URLShortenerHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u, err := s.linkRepo.Restore(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", u)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// HandleMethodNotAllowed обрабатывает не валидный HTTP метод
func (s URLShortenerHandler) HandleMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Only GET and POST requests are allowed!", http.StatusMethodNotAllowed)
}

// HandleNotFound обрабатывает не найденный путь
func (s URLShortenerHandler) HandleNotFound(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, `Only POST "/" with link in body and GET "/{short_link_id} are allowed" `, http.StatusNotFound)
}
