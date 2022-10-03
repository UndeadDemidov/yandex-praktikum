package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
)

// Storage реализует хранение ссылок в памяти.
// Является потоко безопасной реализацией Repository
type Storage struct {
	storage map[string]map[string]string
	mx      sync.Mutex
}

var _ handlers.Repository = (*Storage)(nil)

// NewStorage cоздает и возвращает экземпляр Storage
func NewStorage() *Storage {
	s := Storage{}
	s.storage = make(map[string]map[string]string)
	return &s
}

// Store сохраняет ссылку в хранилище с указанным id
func (s *Storage) Store(ctx context.Context, user string, link string) (id string, err error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	id, err = utils.CreateShortID(ctx, s.isExist)
	if err != nil {
		return "", err
	}

	if _, ok := s.storage[user]; !ok {
		s.storage[user] = make(map[string]string)
	}

	s.storage[user][id] = link
	return id, nil
}

// isExist проверяет наличие id в сторадже
func (s *Storage) isExist(_ context.Context, id string) bool {
	for _, user := range s.storage {
		_, ok := user[id]
		if ok {
			return true
		}
	}
	return false
}

// Restore возвращает исходную ссылку по переданному короткому ID
func (s *Storage) Restore(_ context.Context, id string) (link string, err error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	for _, user := range s.storage {
		l, ok := user[id]
		if ok {
			return l, nil
		}
	}

	return "", fmt.Errorf(storages.ErrLinkNotFound, id)
}

// Unstore - помечает список ранее сохраненных ссылок удаленными
// только тех ссылок, которые принадлежат пользователю
// Только для совместимости контракта
func (s *Storage) Unstore(_ context.Context, _ string, _ []string) {
	// ToDo реализовать для практики
	panic("not implemented for memory storage")
}

// GetUserStorage возвращает map[id]link ранее сокращенных ссылок указанным пользователем
func (s *Storage) GetUserStorage(_ context.Context, user string) map[string]string {
	s.mx.Lock()
	defer s.mx.Unlock()

	ub, ok := s.storage[user]
	if !ok {
		return map[string]string{}
	}
	return ub
}

// StoreBatch сохраняет пакет ссылок из map[correlation_id]original_link и возвращает map[correlation_id]short_link
func (s *Storage) StoreBatch(ctx context.Context, user string, batchIn map[string]string) (batchOut map[string]string, err error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	if _, ok := s.storage[user]; !ok {
		s.storage[user] = make(map[string]string)
	}

	batchOut = make(map[string]string)
	var id string
	// требуется go 1.18, а в yandex_practicum видимо еще не обновили go
	// maps.Copy(s.storage[user], batch)
	for corrID, link := range batchIn {
		id, err = utils.CreateShortID(ctx, s.isExist)
		if err != nil {
			return nil, err
		}
		s.storage[user][id] = link
		batchOut[corrID] = id
	}

	return batchOut, nil
}

// Ping проверяет, что экземпляр Storage создан корректно, например с помощью NewStorage()
func (s *Storage) Ping(_ context.Context) error {
	if s.storage == nil {
		return storages.ErrStorageIsUnavailable
	}
	return nil
}

// Close ничего не делает, требуется только для совместимости с контрактом
func (s *Storage) Close() error {
	// Do nothing
	return nil
}
