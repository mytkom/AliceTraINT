package middleware

import (
	"net/http"
)

func NewValidateHTMXMw() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("HX-Request") != "true" {
				http.Error(w, "invalid request, accepts only HTMX", http.StatusUnprocessableEntity)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

func NewBlockHTMXMw() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("HX-Request") == "true" {
				http.Error(w, "invalid request, does not accept HTMX", http.StatusUnprocessableEntity)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
