package app

import "net/http"

func NewServer(addr string) *http.Server {
	linkStore := NewLinkStorage()
	handler := NewURLShortenerHandler(linkStore)
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	s := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return s
}
