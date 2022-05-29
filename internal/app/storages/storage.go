package storages

import (
	"context"
	"errors"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
)

var (
	ErrUnableCreateShortID = errors.New("couldn't create unique ID in 10 tries")
)

// createShortID создает короткий ID с проверкой на валидность
func createShortID(ctx context.Context, isExist func(context.Context, string) bool) (id string, err error) {
	for i := 0; i < 10; i++ {
		id = utils.NewUniqueID()
		if !isExist(ctx, id) {
			return id, nil
		}
	}
	return "", ErrUnableCreateShortID
}
