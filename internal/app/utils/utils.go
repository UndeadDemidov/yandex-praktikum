package utils

import (
	"errors"
	"io/ioutil"
	"net/url"
	"os"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var ErrUnableCreateShortID = errors.New("couldn't create unique ID in 10 tries")

// IsURL проверяет ссылку на валидность.
// Хотел сначала на регулярках сделать, потом со стековерфлоу согрешил
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func CheckFilename(filename string) (err error) {
	// Check if file already exists
	if _, err = os.Stat(filename); err == nil {
		return nil
	}

	// Attempt to create it
	var d []byte
	if err = ioutil.WriteFile(filename, d, 0644); err == nil {
		_ = os.Remove(filename) // And delete it
		return nil
	}

	return err
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
