package middleware

import (
	"compress/gzip"
	"net/http"

	"github.com/UndeadDemidov/yandex-praktikum/internal/app/utils"
)

// Decompress реализует распаковку запроса переданного в сжатом gzip
// Сделано максимально топорно, согласно текущему уровню курса. Например, каждый раз создается новый reader.
// Может продержусь на курсе до Пула ресурсов ¯\_(ツ)_/¯
func Decompress(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				utils.InternalServerError(w, err)
			}
			defer func() {
				err := gz.Close()
				if err != nil {
					utils.InternalServerError(w, err)
				}
			}()
			r.Body = gz
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
