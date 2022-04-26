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
	baseURL  string
}

// NewURLShortenerHandler создает URLShortenerHandler и инициализирует его
func NewURLShortenerHandler(base string, repo Repository) *URLShortenerHandler {
	h := URLShortenerHandler{}
	h.linkRepo = repo
	if IsURL(base) {
		h.baseURL = base
	} else {
		h.baseURL = "http://localhost:8080/"
	}
	return &h
}

func (s URLShortenerHandler) PostHandler() func(w http.ResponseWriter, r *http.Request) {
	return s.handlePost
}

func (s URLShortenerHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return s.handleGet
}

// ServeHTTP - реализация метода интефрейса http.Handler
func (s URLShortenerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
	default:
		s.handleMethodNotAllowed(w, r)
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
		http.Error(w, "Hey, Dude! Provide a link! Not the crap!", http.StatusBadRequest)
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
func (s URLShortenerHandler) createShortLink(originalLink string) (string, error) {
	id, err := s.linkRepo.Store(originalLink)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(s.baseURL)
	sb.WriteString(id)
	shortenedURL := sb.String()
	return shortenedURL, nil
}

// handlePost - ручка для для открытия по короткой ссылке
func (s URLShortenerHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	// Не стал переписывать на получение ID через chi.URLParam
	// потому что в этом случае тесты надо тоже завязать на chi,
	// так как паттерн пути задается при создании сервера,
	// а обрабатывается в handlers.
	// То есть в тесте должен быть задан тот же паттерн через chi,
	// чтобы он смог его подхватить тут.
	// Можно тестировать сервер целиком, но не тестирвовать handlers,
	// тогда не будет этой коллизии.
	// ToDo - Если можно обойти - то как? Буду раз узнать!
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

// handleMethodNotAllowed обрабатывает не валидный HTTP метод
func (s URLShortenerHandler) handleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Only GET and POST requests are allowed!", http.StatusMethodNotAllowed)
}

// handleNotFound обрабатывает не найденный путь
func (s URLShortenerHandler) handleNotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `Only POST "/" with link in body and GET "/{short_link_id} are allowed" `, http.StatusNotFound)
}
