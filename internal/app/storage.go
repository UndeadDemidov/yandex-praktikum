package app

import (
	"fmt"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"sync"
)

const LinkNotFoundError = "link not found with passed id: %s"

type Repository interface {
	Store(link string) (id string, err error)
	Restore(id string) (link string, err error)
}

// LinkStorage является потоко НЕ безопасная реализация Repository
type LinkStorage struct {
	mx      sync.Mutex
	storage map[string]string
}

var _ Repository = (*LinkStorage)(nil)

// NewLinkStorage cоздает и возвращает экземпляр LinkStorage
func NewLinkStorage() *LinkStorage {
	s := LinkStorage{}
	s.storage = make(map[string]string)
	return &s
}

// Store сохраняет ссылку в хранилище и возвращает короткий ID
func (ls *LinkStorage) Store(link string) (id string, err error) {
	// ToDo Пока без реализации избегания дубликатов. Если скажете - доделаю.
	// Но лучше оставить до перехода на БД
	ls.mx.Lock()
	defer ls.mx.Unlock()

	for {
		id, err = gonanoid.New(8)
		if err != nil {
			return "", err
		}
		if _, ok := ls.storage[id]; !ok {
			ls.storage[id] = link
			break
		}
	}
	return id, nil
}

// Restore возвращает исходную ссылку по переданному короткому ID
func (ls *LinkStorage) Restore(id string) (link string, err error) {
	ls.mx.Lock()
	defer ls.mx.Unlock()

	l, ok := ls.storage[id]
	if !ok {
		return "", fmt.Errorf(LinkNotFoundError, id)
	}
	return l, nil
}
