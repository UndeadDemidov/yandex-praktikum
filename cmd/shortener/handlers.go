package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// URLShortenerHandler - реализация интерфейса http.Handler
// Согласно заданию 1-го инкремента
type URLShortenerHandler struct {
	//non-persistent storage
	//just for starting
	storage map[string]string
}

// NewURLShortenerHandler создает URLShortenerHandler и инициализирует его
func NewURLShortenerHandler() *URLShortenerHandler {
	handler := URLShortenerHandler{}
	handler.storage = make(map[string]string)
	return &handler
}

// ServeHTTP - реализация метода интефрейса http.Handler
// Эндпоинт POST / принимает в теле запроса строку URL для сокращения и возвращает ответ с кодом 201 и сокращённым URL в виде текстовой строки в теле.
// Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL и возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
// Нужно учесть некорректные запросы и возвращать для них ответ с кодом 400.
func (short URLShortenerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		short.handleGet(w, r)
	case http.MethodPost:
		short.handlePost(w, r)
	default:
		http.Error(w, "Only GET and POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
}

// handlePost - ручка для создания короткой ссылки
func (short URLShortenerHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if len(b) == 0 {
		http.Error(w, "The link is not provided", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Fatalln("Something went wrong due reading body request")
		return
	}

	link := string(b)
	if !IsURL(link) {
		http.Error(w, "Hey, Dude! Provide a link! Not a crap!", http.StatusBadRequest)
		return
	}
	shortenedURL := short.createShortLink(link)

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortenedURL))
	if err != nil {
		log.Fatalln("Something went wrong due writing response")
	}
}

// createShortLink - создает короткую ссылку в ответ на исходную
// Ссылка не очень короткая :) Простейшая реализация через UUID.
func (short URLShortenerHandler) createShortLink(originalLink string) string {
	//ToDo - заменить на красивую реализацию создания токена (id)
	//ToDo - можно так https://stackoverflow.com/questions/742013/how-do-i-create-a-URL-shortener
	id := strings.Replace(uuid.New().String(), "-", "", -1)
	short.storeURL(originalLink, id)

	var sb strings.Builder
	sb.WriteString("http://localhost:8080/")
	sb.WriteString(id)
	shortenedURL := sb.String()
	return shortenedURL
}

// storeURL - запоминает пару исходная+короткая ссылки
// Выбрана самая простая реализация. Посмотрим, что будет дальше.
func (short URLShortenerHandler) storeURL(originalLink string, id string) {
	//ToDo - пока без реализации избегания дубликатов. Если скажете - доделаю.
	short.storage[id] = originalLink
}

// handlePost - ручка для для открытия по короткой ссылке
func (short URLShortenerHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	token := strings.Replace(r.URL.Path, "/", "", -1)
	u := short.storage[token]
	if u == "" {
		http.Error(w, "Invalid token passed. Create short link first.", http.StatusBadRequest)
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
