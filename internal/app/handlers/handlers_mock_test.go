package handlers

import (
	"errors"
	"fmt"
)

const mockedID = "1111"

var ErrNotExistedID = errors.New("mocked fail, use id = 1111 to get stored link")

// RepoMock - простейший мок для Repository интерфейса
type RepoMock struct {
	singleItemStorage string
}

var _ Repository = (*RepoMock)(nil)

func (rm RepoMock) IsExist(_ string) bool {
	return false
}

func (rm RepoMock) Store(_ string, _ string, _ string) (err error) {
	// rm.singleItemStorage = link
	return nil
}

func (rm RepoMock) Restore(_ string, id string) (link string, err error) {
	if id != mockedID {
		return "", ErrNotExistedID
	}
	return rm.singleItemStorage, nil
}

func (rm RepoMock) Close() error {
	return nil
}

func (rm RepoMock) GetUserBucket(_, _ string) (bucket []BucketItem) {
	return []BucketItem{
		{
			ShortURL:    fmt.Sprintf("http://localhost:8080/%s", mockedID),
			OriginalURL: rm.singleItemStorage,
		},
	}
}
