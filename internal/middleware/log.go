package middleware

import (
	"log"
	"net/http"
	"time"
)

type logResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *logResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func NewLogMw() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lrw := &logResponseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}
			start := time.Now()
			log.Printf("Started %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(lrw, r)
			log.Printf("Completed %s in %v with code %d", r.URL.Path, time.Since(start), lrw.status)
		})
	}
}
