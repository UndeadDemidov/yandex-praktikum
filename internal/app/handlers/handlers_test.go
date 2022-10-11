package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mock_handlers "github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers/mocks"
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8080/"

var dumbError = errors.New("dumb error")

//nolint:funlen
func TestURLShortenerHandler_HandlePostShortenPlain(t *testing.T) {
	type fields struct {
		repo *mock_handlers.MockRepository
	}
	type want struct {
		status        int
		isURLInResult bool
	}
	tests := []struct {
		name    string
		body    string
		want    want
		prepare func(f *fields)
	}{
		{
			name: "valid link",
			body: "https://habr.com/ru/post/66931/",
			want: want{
				status:        http.StatusCreated,
				isURLInResult: true,
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any()).Return("1111", nil),
				)
			},
		},
		{
			name: "link already shortened",
			body: "https://habr.com/ru/post/66931/",
			want: want{
				status:        http.StatusConflict,
				isURLInResult: true,
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any()).Return("", ErrLinkIsAlreadyShortened),
				)
			},
		},
		{
			name: "repository error",
			body: "https://habr.com/ru/post/66931/",
			want: want{
				status:        http.StatusInternalServerError,
				isURLInResult: false,
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any()).Return("", dumbError),
				)
			},
		},
		{
			name: "invalid link",
			body: "yaru",
			want: want{
				status:        http.StatusBadRequest,
				isURLInResult: false,
			},
		},
		{
			name: "empty link",
			body: "",
			want: want{
				status:        http.StatusBadRequest,
				isURLInResult: false,
			},
		},
	}
	zerolog.SetGlobalLevel(zerolog.Disabled)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockRepo := mock_handlers.NewMockRepository(mockCtrl)

			f := fields{
				repo: mockRepo,
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			reader := strings.NewReader(tt.body)
			request := httptest.NewRequest(http.MethodPost, "/", reader)
			w := httptest.NewRecorder()
			h := NewURLShortener(baseURL, mockRepo)
			h.HandlePostShortenPlain(w, request)
			result := w.Result()

			urlResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.status, result.StatusCode)
			assert.Equal(t, tt.want.isURLInResult, utils.IsURL(string(urlResult)))
		})
	}
}

// Пример использования HandlePostShortenPlain
//
//nolint:funlen
func ExampleURLShortener_HandlePostShortenPlain() {
	reader := strings.NewReader(`https://habr.com/ru/post/66931/`)
	request := httptest.NewRequest(http.MethodPost, "/", reader)
	w := httptest.NewRecorder()
	h := NewURLShortener("http://localhost:8080/", RepoMock{})
	h.HandlePostShortenPlain(w, request)
	result := w.Result()

	urlResult, err := io.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	err = result.Body.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(urlResult))
	// Output: http://localhost:8080/1111
}

//nolint:funlen
func TestURLShortenerHandler_HandlePostShorten(t *testing.T) {
	type fields struct {
		repo *mock_handlers.MockRepository
	}
	type want struct {
		wantErr assert.ErrorAssertionFunc
		status  int
	}
	tests := []struct {
		want    want
		name    string
		reqBody string
		prepare func(f *fields)
	}{
		{
			name:    "valid request",
			reqBody: `{"url":"https://habr.com/ru/post/66931/"}`,
			want: want{
				status:  http.StatusCreated,
				wantErr: assert.NoError,
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any()).Return("1111", nil),
				)
			},
		},
		{
			name:    "link already shortened",
			reqBody: `{"url":"https://habr.com/ru/post/66931/"}`,
			want: want{
				status:  http.StatusConflict,
				wantErr: assert.NoError,
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any()).Return("", ErrLinkIsAlreadyShortened),
				)
			},
		},
		{
			name:    "repository error",
			reqBody: `{"url":"https://habr.com/ru/post/66931/"}`,
			want: want{
				status:  http.StatusInternalServerError,
				wantErr: assert.Error,
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any()).Return("", dumbError),
				)
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
	zerolog.SetGlobalLevel(zerolog.Disabled)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockRepo := mock_handlers.NewMockRepository(mockCtrl)

			f := fields{
				repo: mockRepo,
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			reader := strings.NewReader(tt.reqBody)
			request := httptest.NewRequest(http.MethodPost, "/", reader)
			w := httptest.NewRecorder()
			h := NewURLShortener(baseURL, mockRepo)
			h.HandlePostShortenJSON(w, request)
			result := w.Result()

			require.Equal(t, tt.want.status, result.StatusCode)
			var resp URLShortenResponse
			err := json.NewDecoder(result.Body).Decode(&resp)
			if !tt.want.wantErr(t, err, fmt.Sprintf("request: (%v)", tt.reqBody)) {
				return
			}
			err = result.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestURLShortenerHandler_HandleGet(t *testing.T) {
	type fields struct {
		repo *mock_handlers.MockRepository
	}
	type want struct {
		location string
		status   int
	}
	tests := []struct {
		name    string
		link    string
		want    want
		prepare func(f *fields)
	}{
		{
			name: "valid link",
			link: "http://localhost:8080",
			want: want{
				status:   http.StatusTemporaryRedirect,
				location: "https://ya.ru",
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().Restore(gomock.Any(), gomock.Any()).Return("https://ya.ru", nil),
				)
			},
		},
		{
			name: "deleted link",
			link: "http://localhost:8080",
			want: want{
				status:   http.StatusGone,
				location: "",
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().Restore(gomock.Any(), gomock.Any()).Return("", ErrLinkIsDeleted),
				)
			},
		},
		{
			name: "invalid link",
			link: "http://localhost:8080",
			want: want{
				status:   http.StatusBadRequest,
				location: "",
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().Restore(gomock.Any(), gomock.Any()).Return("", dumbError),
				)
			},
		},
	}
	zerolog.SetGlobalLevel(zerolog.Disabled)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockRepo := mock_handlers.NewMockRepository(mockCtrl)

			f := fields{
				repo: mockRepo,
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			r := httptest.NewRequest(http.MethodGet, tt.link, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1111")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h := NewURLShortener(baseURL, mockRepo)
			w := httptest.NewRecorder()
			h.HandleGet(w, r)
			result := w.Result()
			err := result.Body.Close()
			require.NoError(t, err)

			require.Equal(t, tt.want.status, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}

func TestURLShortener_HandleDelete(t *testing.T) {
	type fields struct {
		repo *mock_handlers.MockRepository
	}
	type want struct {
		status int
	}
	tests := []struct {
		name    string
		reqBody string
		want    want
		prepare func(f *fields)
	}{
		{
			name:    "valid request",
			reqBody: `["111","222"]`,
			want: want{
				status: http.StatusAccepted,
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().Unstore(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes(),
				)
			},
		},
		{
			name:    "invalid request",
			reqBody: `["111","222]`,
			want: want{
				status: http.StatusBadRequest,
			},
		},
	}
	zerolog.SetGlobalLevel(zerolog.Disabled)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockRepo := mock_handlers.NewMockRepository(mockCtrl)

			f := fields{
				repo: mockRepo,
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			reader := strings.NewReader(tt.reqBody)
			request := httptest.NewRequest(http.MethodDelete, "/", reader)
			w := httptest.NewRecorder()
			h := NewURLShortener(baseURL, mockRepo)
			h.HandleDelete(w, request)
			result := w.Result()
			err := result.Body.Close()
			require.NoError(t, err)

			require.Equal(t, tt.want.status, result.StatusCode)
		})
	}
}

// Пример использования HandleDelete
func ExampleURLShortener_HandleDelete() {
	reader := strings.NewReader(`["111","222"]`)
	request := httptest.NewRequest(http.MethodDelete, "/", reader)
	w := httptest.NewRecorder()
	h := NewURLShortener("http://localhost:8080/", RepoMock{})
	h.HandleDelete(w, request)
	result := w.Result()
	err := result.Body.Close()
	if err != nil {
		panic(err)
	}
}

//nolint:funlen
func TestURLShortener_HandleGetUserURLsBucket(t *testing.T) {
	type fields struct {
		repo *mock_handlers.MockRepository
	}
	type want struct {
		status int
		result string
	}
	tests := []struct {
		name    string
		want    want
		prepare func(f *fields)
	}{
		{
			name: "single item bucket",
			want: want{
				status: http.StatusOK,
				result: `[
					  {
					    "short_url": "http://localhost:8080/1111",
					    "original_url": "https://ya.ru"
					  }
					]`,
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().GetUserStorage(gomock.Any(), gomock.Any()).
						Return(map[string]string{"1111": "https://ya.ru"}),
				)
			},
		},
		{
			name: "couple item bucket",
			want: want{
				status: http.StatusOK,
				result: `[
					  {
					    "short_url": "http://localhost:8080/1111",
					    "original_url": "https://ya.ru"
					  },
					  {
					    "short_url": "http://localhost:8080/2222",
					    "original_url": "https://yandex.ru"
					  }
					]`,
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().GetUserStorage(gomock.Any(), gomock.Any()).
						Return(map[string]string{"1111": "https://ya.ru", "2222": "https://yandex.ru"}),
				)
			},
		},
		{
			name: "empty bucket",
			want: want{
				status: http.StatusNoContent,
				result: "",
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().GetUserStorage(gomock.Any(), gomock.Any()).
						Return(map[string]string{}),
				)
			},
		},
	}
	zerolog.SetGlobalLevel(zerolog.Disabled)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockRepo := mock_handlers.NewMockRepository(mockCtrl)

			f := fields{
				repo: mockRepo,
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			h := NewURLShortener(baseURL, mockRepo)
			h.HandleGetUserURLsBucket(w, request)
			result := w.Result()
			assert.Equal(t, tt.want.status, result.StatusCode)

			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(result.Body)
			require.NoError(t, err)
			if tt.want.result == "" || buf.String() == "" {
				assert.EqualValues(t, tt.want.result, buf.String())
			} else {
				assert.JSONEq(t, tt.want.result, buf.String())
			}
			err = result.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestURLShortener_HandlePostShortenBatch(t *testing.T) {
	type fields struct {
		repo *mock_handlers.MockRepository
	}
	type want struct {
		status int
		result string
	}
	tests := []struct {
		name    string
		reqBody string
		want    want
		prepare func(f *fields)
	}{
		{
			name:    "empty body",
			reqBody: "",
			want:    want{status: http.StatusBadRequest, result: "proper JSON request is expected\n"},
			prepare: nil,
		},
		{
			name: "single item",
			reqBody: `
[
  {
    "correlation_id": "xxxx",
    "original_url": "https://ya.ru"
  }
]
`,
			want: want{
				status: http.StatusCreated,
				result: `
[
   {
     "correlation_id": "xxxx",
     "short_url": "http://localhost:8080/1111"
   }
 ]
`,
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().StoreBatch(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(map[string]string{"xxxx": "1111"}, nil),
				)
			},
		},
		{
			name: "couple items",
			reqBody: `
[
  {
    "correlation_id": "xxxx",
    "original_url": "https://ya.ru"
  },
  {
    "correlation_id": "yyyy",
    "original_url": "https://yandex.ru"
  }
]
`,
			want: want{
				status: http.StatusCreated,
				result: `
[
   {
     "correlation_id": "xxxx",
     "short_url": "http://localhost:8080/1111"
   },
   {
     "correlation_id": "yyyy",
     "short_url": "http://localhost:8080/2222"
   }
 ]
`,
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().StoreBatch(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(map[string]string{"xxxx": "1111", "yyyy": "2222"}, nil),
				)
			},
		},
		{
			name: "already shortened",
			reqBody: `
[
  {
    "correlation_id": "xxxx",
    "original_url": "https://ya.ru"
  }
]
`,
			want: want{
				status: http.StatusConflict,
				result: `
[
   {
     "correlation_id": "xxxx",
     "short_url": "http://localhost:8080/1111"
   }
 ]
`,
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().StoreBatch(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(map[string]string{"xxxx": "1111"}, ErrLinkIsAlreadyShortened),
				)
			},
		},
		{
			name: "internal server error",
			reqBody: `
[
  {
    "correlation_id": "xxxx",
    "original_url": "https://ya.ru"
  }
]
`,
			want: want{
				status: http.StatusInternalServerError,
				result: fmt.Sprintf("%s\n", dumbError.Error()),
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().StoreBatch(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(map[string]string{}, dumbError),
				)
			},
		},
	}
	zerolog.SetGlobalLevel(zerolog.Disabled)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockRepo := mock_handlers.NewMockRepository(mockCtrl)

			f := fields{
				repo: mockRepo,
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			reader := strings.NewReader(tt.reqBody)
			request := httptest.NewRequest(http.MethodPost, "/", reader)
			w := httptest.NewRecorder()
			h := NewURLShortener(baseURL, mockRepo)
			h.HandlePostShortenBatch(w, request)
			result := w.Result()
			assert.Equal(t, tt.want.status, result.StatusCode)

			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(result.Body)
			require.NoError(t, err)
			if tt.want.result == "" || buf.String() == "" ||
				tt.want.status == http.StatusBadRequest ||
				tt.want.status == http.StatusInternalServerError {
				assert.EqualValues(t, tt.want.result, buf.String())
			} else {
				assert.JSONEq(t, tt.want.result, buf.String())
			}
			err = result.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestURLShortener_HeartBeat(t *testing.T) {
	type fields struct {
		repo *mock_handlers.MockRepository
	}
	type want struct {
		status int
		result string
	}
	tests := []struct {
		name    string
		want    want
		prepare func(f *fields)
	}{
		{
			name: "ok",
			want: want{status: http.StatusOK, result: "I'm alive (c)Helloween"},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().Ping(gomock.Any()).Return(nil),
				)
			},
		},
		{
			name: "not ok",
			want: want{status: http.StatusInternalServerError, result: "dumb error\n"},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.repo.EXPECT().Ping(gomock.Any()).Return(dumbError),
				)
			},
		},
	}
	zerolog.SetGlobalLevel(zerolog.Disabled)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockRepo := mock_handlers.NewMockRepository(mockCtrl)

			f := fields{
				repo: mockRepo,
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			h := NewURLShortener(baseURL, mockRepo)
			h.HeartBeat(w, request)
			result := w.Result()
			assert.Equal(t, tt.want.status, result.StatusCode)

			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(result.Body)
			require.NoError(t, err)
			assert.EqualValues(t, tt.want.result, buf.String())
			err = result.Body.Close()
			require.NoError(t, err)
		})
	}
}
