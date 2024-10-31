package middleware

import (
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/utils"
)

func NewValidateHTMXMw() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !utils.IsHTMXRequest(r) {
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
			if utils.IsHTMXRequest(r) {
				http.Error(w, "invalid request, does not accept HTMX", http.StatusUnprocessableEntity)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
