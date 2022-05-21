package storages

import (
	"fmt"
	"sync"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
)

const (
	ErrLinkNotFound    = "link not found with passed id %s"
	ErrIDAlreadyExists = "passed id %s already exists in the storage"
)

// LinkStorage реализует хранение ссылок в памяти.
// Является потоко безопасной реализацией Repository
type LinkStorage struct {
	mx      sync.Mutex
	storage map[string]map[string]string
}

var _ handlers.Repository = (*LinkStorage)(nil)

// NewLinkStorage cоздает и возвращает экземпляр LinkStorage
func NewLinkStorage() *LinkStorage {
	s := LinkStorage{}
	s.storage = make(map[string]map[string]string)
	return &s
}

// IsExist проверяет наличие id в сторадже
func (ls *LinkStorage) IsExist(id string) bool {
	ls.mx.Lock()
	defer ls.mx.Unlock()

	return ls.isExist(id)
}

// Store сохраняет ссылку в хранилище и возвращает короткий ID
func (ls *LinkStorage) Store(user string, id string, link string) (err error) {
	ls.mx.Lock()
	defer ls.mx.Unlock()

	if ls.isExist(id) {
		return fmt.Errorf(ErrIDAlreadyExists, id)
	}
	_, ok := ls.storage[user]
	if !ok {
		ls.storage[user] = make(map[string]string)
	}

	ls.storage[user][id] = link
	return nil
}

// Restore возвращает исходную ссылку по переданному короткому ID
func (ls *LinkStorage) Restore(user string, id string) (link string, err error) {
	ls.mx.Lock()
	defer ls.mx.Unlock()

	l, ok := ls.storage[user][id]
	if !ok {
		return "", fmt.Errorf(ErrLinkNotFound, id)
	}
	return l, nil
}

// Close ничего не делает, требуется только для совместимости с контрактом
func (ls *LinkStorage) Close() error {
	// Do nothing
	return nil
}

// isExist проверяет наличие id в сторадже
// внутреняя реализация
func (ls *LinkStorage) isExist(id string) bool {
	for _, m := range ls.storage {
		_, ok := m[id]
		if ok {
			return true
		}
	}
	return false
}
