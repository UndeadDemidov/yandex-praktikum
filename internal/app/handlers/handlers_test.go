package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
)

func TestURLShortenerHandler_HandlePost(t *testing.T) {
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
			h.HandlePost(w, request)
			result := w.Result()

			urlResult, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.status, result.StatusCode)
			assert.Equal(t, tt.want.isURL, utils.IsURL(string(urlResult)))
		})
	}
}

func TestURLShortenerHandler_HandlePostShorten(t *testing.T) {
	type want struct {
		status  int
		wantErr assert.ErrorAssertionFunc
	}
	tests := []struct {
		name     string
		reqBody  string
		respBody string
		want     want
	}{
		{
			name:    "valid request",
			reqBody: `{"url":"https://habr.com/ru/post/66931/"}`,
			want: want{
				status:  http.StatusCreated,
				wantErr: assert.NoError,
			},
		},
		{
			name:    "invalid link",
			reqBody: `{"url":"yaru"}`,
			want: want{
				status:  http.StatusBadRequest,
				wantErr: assert.Error,
			},
		},
		{
			name:    "invalid json",
			reqBody: `{"url":"yaru"`,
			want: want{
				status:  http.StatusBadRequest,
				wantErr: assert.Error,
			},
		},
		{
			name:    "empty body",
			reqBody: "",
			want: want{
				status:  http.StatusBadRequest,
				wantErr: assert.Error,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.reqBody)
			request := httptest.NewRequest(http.MethodPost, "/", reader)
			w := httptest.NewRecorder()
			h := NewURLShortenerHandler("http://localhost:8080/", RepoMock{})
			h.HandlePostShorten(w, request)
			result := w.Result()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					log.Fatalln(err)
				}
			}(result.Body)

			require.Equal(t, tt.want.status, result.StatusCode)
			var resp URLShortenResponse
			err := json.NewDecoder(result.Body).Decode(&resp)
			if !tt.want.wantErr(t, err, fmt.Sprintf("request: (%v)", tt.reqBody)) {
				return
			}
		})
	}
}

func TestURLShortenerHandler_HandleGet(t *testing.T) {
	type args struct {
		repo RepoMock
	}
	type want struct {
		status   int
		location string
	}
	tests := []struct {
		name  string
		link  string
		param string
		args  args
		want  want
	}{
		{
			name:  "valid link",
			link:  "http://localhost:8080",
			param: "1111",
			args:  args{RepoMock{singleItemStorage: "https://ya.ru"}},
			want: want{
				status:   http.StatusTemporaryRedirect,
				location: "https://ya.ru",
			},
		},
		{
			name:  "invalid link",
			link:  "http://localhost:8080",
			param: "2222",
			args:  args{RepoMock{singleItemStorage: "https://ya.ru"}},
			want: want{
				status:   http.StatusBadRequest,
				location: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.link, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.param)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h := NewURLShortenerHandler("http://localhost:8080/", tt.args.repo)
			w := httptest.NewRecorder()
			h.HandleGet(w, r)
			result := w.Result()
			_ = result.Body.Close()

			require.Equal(t, tt.want.status, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
