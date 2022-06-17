package storages

import (
	"errors"
)

var (
	ErrUnableCreateShortID  = errors.New("couldn't create unique ID in 10 tries")
	ErrStorageIsUnavailable = errors.New("storage is unavailable")
)

const (
	ErrLinkNotFound = "link not found with passed id %s"
)
