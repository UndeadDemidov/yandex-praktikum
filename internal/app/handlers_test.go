package app

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
			h := NewURLShortenerHandler("http://localhost:8080/", RepoMock{})
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
			h := NewURLShortenerHandler("http://localhost:8080/", tt.args.repo)
			h.ServeHTTP(w, request)
			result := w.Result()
			_ = result.Body.Close()

			require.Equal(t, tt.want.status, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
