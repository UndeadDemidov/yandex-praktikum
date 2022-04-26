package app

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

// URLShortenerHandler - реализация интерфейса http.Handler
// Согласно заданию 1-го инкремента
type URLShortenerHandler struct {
	//non-persistent storage
	//just for starting
	linkRepo Repository
}

// NewURLShortenerHandler создает URLShortenerHandler и инициализирует его
func NewURLShortenerHandler(repo Repository) *URLShortenerHandler {
	h := URLShortenerHandler{}
	h.linkRepo = repo
	return &h
}

// ServeHTTP - реализация метода интефрейса http.Handler
// Эндпоинт POST / принимает в теле запроса строку URL для сокращения и возвращает ответ с кодом 201 и сокращённым URL в виде текстовой строки в теле.
// Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL и возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
// Нужно учесть некорректные запросы и возвращать для них ответ с кодом 400.
func (s URLShortenerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
	default:
		http.Error(w, "Only GET and POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
}

// handlePost - ручка для создания короткой ссылки
func (s URLShortenerHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Something went wrong due reading body request", http.StatusInternalServerError)
		return
	}
	if len(b) == 0 {
		http.Error(w, "The link is not provided", http.StatusBadRequest)
		return
	}

	link := string(b)
	if !IsURL(link) {
		http.Error(w, "Hey, Dude! Provide a link! Not a crap!", http.StatusBadRequest)
		return
	}
	shortenedURL, err := s.createShortLink(link)
	if err != nil {
		http.Error(w, "Couldn't store link", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortenedURL))
	if err != nil {
		http.Error(w, "Something went wrong due writing response", http.StatusInternalServerError)
		return
	}
}

// createShortLink - создает короткую ссылку в ответ на исходную
// Ссылка не очень короткая :) Простейшая реализация через UUID.
func (s URLShortenerHandler) createShortLink(originalLink string) (string, error) {
	id, err := s.linkRepo.Store(originalLink)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("http://localhost:8080/")
	sb.WriteString(id)
	shortenedURL := sb.String()
	return shortenedURL, nil
}

// handlePost - ручка для для открытия по короткой ссылке
func (s URLShortenerHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	u, err := s.linkRepo.Restore(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", u)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// IsURL проверяет ссылку на валидность.
// Хотел сначала на регулярках сделать, потом со стековерфлоу согрешил
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
