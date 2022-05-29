package storages

import (
	"context"
	"errors"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	ErrUnableCreateShortID = errors.New("couldn't create unique ID in 10 tries")
)

// createShortID создает короткий ID с проверкой на валидность
func createShortID(ctx context.Context, isExist func(context.Context, string) bool) (id string, err error) {
	for i := 0; i < 10; i++ {
		id, err = gonanoid.New(8)
		if err != nil {
			return "", err
		}
		if !isExist(ctx, id) {
			return id, nil
		}
	}
	return "", ErrUnableCreateShortID
}
