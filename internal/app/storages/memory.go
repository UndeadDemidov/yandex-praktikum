package storages

import (
	"context"
	"fmt"
	"sync"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
)

const (
	ErrLinkNotFound = "link not found with passed id %s"
)

// MemoryStorage реализует хранение ссылок в памяти.
// Является потоко безопасной реализацией Repository
type MemoryStorage struct {
	mx      sync.Mutex
	storage map[string]map[string]string
}

var _ handlers.Repository = (*MemoryStorage)(nil)

// NewLinkStorage cоздает и возвращает экземпляр MemoryStorage
func NewLinkStorage() *MemoryStorage {
	s := MemoryStorage{}
	s.storage = make(map[string]map[string]string)
	return &s
}

// Store сохраняет ссылку в хранилище с указанным id
func (ms *MemoryStorage) Store(ctx context.Context, user string, link string) (id string, err error) {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	id, err = createShortID(ctx, ms.isExist)
	if err != nil {
		return "", err
	}

	_, ok := ms.storage[user]
	if !ok {
		ms.storage[user] = make(map[string]string)
	}

	ms.storage[user][id] = link
	return id, nil
}

// isExist проверяет наличие id в сторадже
func (ms *MemoryStorage) isExist(_ context.Context, id string) bool {
	for _, m := range ms.storage {
		_, ok := m[id]
		if ok {
			return true
		}
	}
	return false
}

// Restore возвращает исходную ссылку по переданному короткому ID
func (ms *MemoryStorage) Restore(_ context.Context, id string) (link string, err error) {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	for _, user := range ms.storage {
		l, ok := user[id]
		if ok {
			return l, nil
		}
	}

	return "", fmt.Errorf(ErrLinkNotFound, id)
}

// GetAllUserLinks возвращает map[id]link ранее сокращенных ссылок указанным пользователем
func (ms *MemoryStorage) GetAllUserLinks(_ context.Context, user string) map[string]string {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	ub, ok := ms.storage[user]
	if !ok {
		return map[string]string{}
	}
	return ub
}

// StoreBatch сохраняет пакет ссылок из map[id]link
func (ms *MemoryStorage) StoreBatch(_ context.Context, user string, batch map[string]string) error {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	_, ok := ms.storage[user]
	if !ok {
		ms.storage[user] = make(map[string]string)
	}
	// требуется go 1.18, а в yandex_practicum видимо еще не обновили go
	// maps.Copy(ms.storage[user], batch)
	for k, v := range batch {
		ms.storage[user][k] = v
	}

	return nil
}

// Close ничего не делает, требуется только для совместимости с контрактом
func (ms *MemoryStorage) Close() error {
	// Do nothing
	return nil
}
