package middleware

import (
	"fmt"
	"net"
	"net/http"
)

func IPFilter(n *net.IPNet) func(next http.Handler) http.Handler {
	f := NewIPFilterer(n)
	return f.Handler
}

type IPFilterer struct {
	subnet *net.IPNet
}

func NewIPFilterer(n *net.IPNet) *IPFilterer {
	return &IPFilterer{subnet: n}
}

func (f *IPFilterer) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := net.ParseIP(r.Header.Get("X-Real-IP"))
		if !f.subnet.Contains(ip) {
			http.Error(w, fmt.Sprintf("ip %s is not from trusted subnet", ip.String()), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
