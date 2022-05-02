package app

import "errors"

type RepoMock struct {
	singleItemStorage string
}

func (rm RepoMock) IsExist(_ string) bool {
	return false
}

func (rm RepoMock) Store(_ string, _ string) (err error) {
	// rm.singleItemStorage = link
	return nil
}

func (rm RepoMock) Restore(id string) (link string, err error) {
	if id != "1111" {
		return "", errors.New("mocked fail, use id = 1111 to get stored link")
	}
	return rm.singleItemStorage, nil
}
