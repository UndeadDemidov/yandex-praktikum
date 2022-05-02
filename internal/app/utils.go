package app

import (
	"errors"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"net/url"
)

var ErrUnableCreateShortID = errors.New("couldn't create unique ID in 10 tries")

// IsURL проверяет ссылку на валидность.
// Хотел сначала на регулярках сделать, потом со стековерфлоу согрешил
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// CreateShortID создает короткий ID с проверкой на валидность
func CreateShortID(isExist func(string) bool) (id string, err error) {
	for i := 0; i < 10; i++ {
		id, err = gonanoid.New(8)
		if err != nil {
			return "", err
		}
		if !isExist(id) {
			return id, nil
		}
	}
	return "", ErrUnableCreateShortID
}
