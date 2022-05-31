package utils

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/storages"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

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
	if err = ioutil.WriteFile(filename, d, 0644); err == nil { //nolint:gosec
		err = os.Remove(filename) // And delete it
		if err != nil {
			return err
		}
		return nil
	}

	return err
}

func NewUniqueID() (id string) {
	var err error
	id, err = gonanoid.New(8)
	if err != nil {
		panic(err)
	}
	return id
}

// CreateShortID создает короткий ID с проверкой на валидность
func CreateShortID(ctx context.Context, isExist func(context.Context, string) bool) (id string, err error) {
	for i := 0; i < 10; i++ {
		id = NewUniqueID()
		if !isExist(ctx, id) {
			return id, nil
		}
	}
	return "", storages.ErrUnableCreateShortID
}
