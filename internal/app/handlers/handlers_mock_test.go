package handlers

import (
	"context"
	"errors"
)

const mockedID = "1111"

var ErrNotExistedID = errors.New("mocked fail, use id = 1111 to get stored link")

// RepoMock - простейший мок для Repository интерфейса
type RepoMock struct {
	singleItemStorage string
}

var _ Repository = (*RepoMock)(nil)

func (rm RepoMock) IsExist(_ context.Context, _ string) bool {
	return false
}

func (rm RepoMock) Store(_ context.Context, _ string, _ string, _ string) (err error) {
	// rm.singleItemStorage = link
	return nil
}

func (rm RepoMock) Restore(_ context.Context, id string) (link string, err error) {
	if id != mockedID {
		return "", ErrNotExistedID
	}
	return rm.singleItemStorage, nil
}

func (rm RepoMock) Close() error {
	return nil
}

func (rm RepoMock) GetAllUserLinks(_ context.Context, _ string) map[string]string {
	return map[string]string{mockedID: rm.singleItemStorage}
}

func (rm RepoMock) StoreBatch(_ context.Context, _ string, _ map[string]string) error {
	return nil
}
