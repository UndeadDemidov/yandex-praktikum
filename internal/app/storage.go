package app

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
)

const LinkNotFoundError = "link not found with passed id: %s"

type Repository interface {
	Store(link string) (id string, err error)
	Restore(id string) (link string, err error)
}

type LinkStorage struct {
	storage map[string]string
}

func NewLinkStorage() *LinkStorage {
	s := LinkStorage{}
	s.storage = make(map[string]string)
	return &s
}

func (ls LinkStorage) Store(link string) (id string, err error) {
	//ToDo - пока без реализации избегания дубликатов. Если скажете - доделаю.
	//ToDo - заменить на красивую реализацию создания токена (id)
	//ToDo - можно так https://stackoverflow.com/questions/742013/how-do-i-create-a-url-shortener
	id = strings.Replace(uuid.New().String(), "-", "", -1)
	ls.storage[id] = link
	return id, nil
}

func (ls LinkStorage) Restore(id string) (link string, err error) {
	l, ok := ls.storage[id]
	if !ok {
		return "", fmt.Errorf(LinkNotFoundError, id)
	}
	return l, nil
}
