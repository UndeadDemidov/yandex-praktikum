package app

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsURL(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "well formatted URL 1",
			args: args{"https://habr.com/ru/post/66931/"},
			want: true,
		},
		{
			name: "well formatted URL 2",
			args: args{"https://habr.com/ru/post/66931/?"},
			want: true,
		},
		{
			name: "well formatted URL 3",
			args: args{"http://habr.com/ru/post/66931/"},
			want: true,
		},
		{
			name: "well formatted URL 4",
			args: args{"https://ya.ru"},
			want: true,
		},
		{
			name: "bad URL 1",
			args: args{"ya"},
			want: false,
		},
		{
			name: "bad URL 2",
			args: args{"1234"},
			want: false,
		},
		{
			name: "bad URL 3",
			args: args{"https://ya.ru/!@#$%^&*()"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsURL(tt.args.str); got != tt.want {
				t.Errorf("IsURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewURLShortenerHandlerPost(t *testing.T) {
	type want struct {
		status int
		isURL  bool
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "valid link",
			body: "https://habr.com/ru/post/66931/",
			want: want{
				status: http.StatusCreated,
				isURL:  true,
			},
		},
		{
			name: "invalid link",
			body: "yaru",
			want: want{
				status: http.StatusBadRequest,
				isURL:  false,
			},
		},
		{
			name: "empty link",
			body: "",
			want: want{
				status: http.StatusBadRequest,
				isURL:  false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.body)
			request := httptest.NewRequest(http.MethodPost, "/", reader)
			w := httptest.NewRecorder()
			h := NewURLShortenerHandler(RepoMock{})
			h.ServeHTTP(w, request)
			result := w.Result()

			urlResult, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.status, result.StatusCode)
			assert.Equal(t, tt.want.isURL, IsURL(string(urlResult)))
		})
	}
}

func TestNewURLShortenerHandlerGet(t *testing.T) {
	type args struct {
		repo RepoMock
	}
	type want struct {
		status   int
		location string
	}
	tests := []struct {
		name string
		link string
		args args
		want want
	}{
		{
			name: "valid link",
			link: "http://localhost:8080/1111",
			args: args{RepoMock{singleItemStorage: "https://ya.ru"}},
			want: want{
				status:   http.StatusTemporaryRedirect,
				location: "https://ya.ru",
			},
		},
		{
			name: "invalid link",
			link: "http://localhost:8080/2222",
			args: args{RepoMock{singleItemStorage: "https://ya.ru"}},
			want: want{
				status:   http.StatusBadRequest,
				location: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.link, nil)
			w := httptest.NewRecorder()
			h := NewURLShortenerHandler(tt.args.repo)
			h.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			require.Equal(t, tt.want.status, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}

type RepoMock struct {
	singleItemStorage string
}

func (rm RepoMock) Store(link string) (id string, err error) {
	// не используется для тестов - go vet ругается
	// rm.singleItemStorage = link
	return "1111", nil
}

func (rm RepoMock) Restore(id string) (link string, err error) {
	if id != "1111" {
		return "", errors.New("mocked fail, use id = 1111 to get stored link")
	}
	return rm.singleItemStorage, nil
}
