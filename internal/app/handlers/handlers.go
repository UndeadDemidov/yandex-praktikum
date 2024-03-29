package handlers

import (
	"context"
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
	_ "github.com/golang/mock/mockgen/model"
	"github.com/rs/zerolog/log"
)

//go:generate mockgen -destination=./mocks/mock_repository.go . Repository

var (
	ErrLinkIsAlreadyShortened = errors.New("link is already shortened")
	ErrEmptyBatchToShort      = errors.New("nothing to short")
	ErrLinkIsDeleted          = errors.New("link is deleted")
	ErrMethodNotAllowed       = errors.New("method is not allowed, read task description carefully")
	ErrProperJSONIsExpected   = errors.New("proper JSON is expected, read task description carefully")
)

// URLShortener - реализует набор методов для сокращения ссылок, хранение их оригинального состояние
// и открытие по сокращенному варианту. Обеспечивается контроль авторства сокращенных ссылок.
type URLShortener struct {
	linkRepo Repository
	baseURL  string
}

// NewURLShortener создает URLShortener и инициализирует его адресом, по которому будут доступны методы,
// и репозиторием хранения ссылок.
func NewURLShortener(base string, repo Repository) *URLShortener {
	h := URLShortener{}
	h.linkRepo = repo
	if utils.IsURL(base) {
		h.baseURL = fmt.Sprintf("%s/", strings.TrimRight(base, "/"))
	} else {
		h.baseURL = "http://localhost:8080/"
	}

	return &h
}

// HandlePostShortenPlain - метод для создания короткой ссылки, где оригинальная ссылка передается через Text Body.
func (s URLShortener) HandlePostShortenPlain(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	// validate
	if len(b) == 0 {
		http.Error(w, "The link is not provided", http.StatusBadRequest)
		log.Debug().Msg("User provided not a single link")
		return
	}
	link := string(b)
	if !utils.IsURL(link) {
		http.Error(w, fmt.Sprintf("Hey, Dude! Provide a link! Not the crap: %v", link), http.StatusBadRequest)
		log.Debug().Msg(fmt.Sprintf("User provided data: %v", link))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	user := midware.GetUserID(ctx)
	shortenedURL, err := s.shorten(ctx, user, link)
	switch {
	case errors.Is(err, ErrLinkIsAlreadyShortened):
		w.WriteHeader(http.StatusConflict)
	case err != nil:
		utils.InternalServerError(w, err)
		return
	default:
		w.WriteHeader(http.StatusCreated)
	}

	_, err = w.Write([]byte(shortenedURL))
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
}

// HandlePostShortenJSON - метод для создания короткой ссылки, где оригинальная ссылка передается с помощью JSON.
func (s URLShortener) HandlePostShortenJSON(w http.ResponseWriter, r *http.Request) {
	req := URLShortenRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, ErrProperJSONIsExpected.Error(), http.StatusBadRequest)
		return
	}
	if !utils.IsURL(req.URL) {
		http.Error(w, fmt.Sprintf("Hey, Dude! Provide a link! Not the crap: %v", req.URL), http.StatusBadRequest)
		log.Debug().Msg(fmt.Sprintf("User provided data: %v", req.URL))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	user := midware.GetUserID(ctx)
	shortenedURL, err := s.shorten(ctx, user, req.URL)
	switch {
	case errors.Is(err, ErrLinkIsAlreadyShortened):
		w.WriteHeader(http.StatusConflict)
	case err != nil:
		utils.InternalServerError(w, err)
		return
	default:
		w.WriteHeader(http.StatusCreated)
	}

	resp := URLShortenResponse{Result: shortenedURL}
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		utils.InternalServerError(w, err)
	}
}

// shorten возвращает короткую ссылку в ответ на оригинальную
func (s URLShortener) shorten(ctx context.Context, user string, originalURL string) (shortenedURL string, err error) {
	var id string
	id, err = s.linkRepo.Store(ctx, user, originalURL)
	if err == nil || errors.Is(err, ErrLinkIsAlreadyShortened) {
		return fmt.Sprintf("%s%s", s.baseURL, id), err
	}
	return "", err
}

// HandleGet - метод для открытия оригинальной ссылки по короткому варианту.
func (s URLShortener) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	url, err := s.linkRepo.Restore(ctx, id)
	switch {
	case errors.Is(err, ErrLinkIsDeleted):
		http.Error(w, err.Error(), http.StatusGone)
		log.Debug().Err(err)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Debug().Err(err)
		return
	}
	w.Header().Add("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// HandleDelete - метод для удаления раннее созданных коротких ссылок.
// На вход принимается json массив токенов коротких ссылок для удаления.
func (s URLShortener) HandleDelete(w http.ResponseWriter, r *http.Request) {
	req := make([]URLID, 0)
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, ErrProperJSONIsExpected.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	user := midware.GetUserID(ctx)
	s.unstore(ctx, user, req)
	w.WriteHeader(http.StatusAccepted)
}

func (s URLShortener) unstore(ctx context.Context, user string, req []URLID) {
	list := make([]string, 0, len(req))
	for _, urlID := range req {
		list = append(list, string(urlID))
	}
	// Стартуем в отдельном thread, чтобы не блокировать handler, т.е. удаляем список id асинхронно
	// Сколько вызовов - столько новых thread
	go s.linkRepo.Unstore(ctx, user, list)
}

// HandleGetUserURLsBucket - метод для получения всех сокращенных пользователем ссылок.
func (s URLShortener) HandleGetUserURLsBucket(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	user := midware.GetUserID(ctx)
	urlsMap := s.linkRepo.GetUserStorage(ctx, user)
	if len(urlsMap) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(MapToBucket(s.baseURL, urlsMap))
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
}

// HandlePostShortenBatch - метод для создания коротких ссылок одним пакетом,
// где оригинальные ссылки передаются через JSON.
func (s URLShortener) HandlePostShortenBatch(w http.ResponseWriter, r *http.Request) {
	var req []URLShortenCorrelatedRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "proper JSON request is expected", http.StatusBadRequest)
		log.Debug().Msg("corrupted JSON provided")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	user := midware.GetUserID(ctx)
	var resp []URLShortenCorrelatedResponse
	resp, err = s.shortenBatch(ctx, user, req)
	switch {
	case errors.Is(err, ErrLinkIsAlreadyShortened):
		w.WriteHeader(http.StatusConflict)
	case err != nil:
		utils.InternalServerError(w, err)
		return
	default:
		w.WriteHeader(http.StatusCreated)
	}

	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
}

// shortenBatch производит сокращение ссылок пакетом и возвращает слайс сокращенных ссылок
func (s URLShortener) shortenBatch(ctx context.Context, user string, req []URLShortenCorrelatedRequest) (resp []URLShortenCorrelatedResponse, err error) {
	if len(req) == 0 {
		return nil, ErrEmptyBatchToShort
	}

	batchIn := map[string]string{} // map[correlation_id]original_link
	for _, request := range req {
		batchIn[request.CorrelationID] = request.OriginalURL
	}

	batchOut, err := s.linkRepo.StoreBatch(ctx, user, batchIn) // batchOut = map[correlation_id]short_id
	if err != nil && !errors.Is(err, ErrLinkIsAlreadyShortened) {
		return []URLShortenCorrelatedResponse{}, err
	}

	resp = make([]URLShortenCorrelatedResponse, 0, len(req))
	for corrID, id := range batchOut {
		resp = append(resp, URLShortenCorrelatedResponse{
			CorrelationID: corrID,
			ShortURL:      fmt.Sprintf("%s%s", s.baseURL, id),
		})
	}
	// err либо nil, либо ErrLinkIsAlreadyShortened
	return resp, err
}

// HeartBeat - метод для проверки, что подключение к репозиторию живое.
func (s URLShortener) HeartBeat(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := s.linkRepo.Ping(ctx); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("I'm alive (c)Helloween"))
	if err != nil {
		utils.InternalServerError(w, err)
	}
}

// HandleMethodNotAllowed обрабатывает не валидный HTTP метод
func (s URLShortener) HandleMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
}

// HandleNotFound обрабатывает не найденный путь
func (s URLShortener) HandleNotFound(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, ErrMethodNotAllowed.Error(), http.StatusNotFound)
}

// Repository описывает контракт работы с хранилищем.
// Используется для удобства тестирования и для дальнейшей легкой миграции на другой "движок".
type Repository interface {
	// Store сохраняет оригинальную ссылку и возвращает id (токен) сокращенного варианта.
	Store(ctx context.Context, user string, link string) (id string, err error)
	// Restore возвращает оригинальную ссылку по его id.
	// если error == ErrLinkIsDeleted значит короткая ссылка (id) была удалена.
	Restore(ctx context.Context, id string) (link string, err error)
	// Unstore - помечает ссылки удаленными.
	// Согласно заданию - результат работы пользователю не возвращается.
	Unstore(ctx context.Context, user string, ids []string)
	// GetUserStorage возвращает массив всех ранее сокращенных пользователей ссылок.
	GetUserStorage(ctx context.Context, user string) map[string]string
	// StoreBatch сохраняет пакет ссылок в хранилище и возвращает список пакет id.
	// batchIn = map[correlation_id]original_link
	// batchOut= map[correlation_id]short_link
	// если error == ErrLinkIsAlreadyShortened значит среди пакета были ранее сокращенные ссылки.
	StoreBatch(ctx context.Context, user string, batchIn map[string]string) (batchOut map[string]string, err error)
	// Ping проверяет готовность к работе репозитория.
	Ping(context.Context) error
	// Close завершает работу репозитория в стиле graceful shutdown.
	Close() error
}
