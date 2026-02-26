package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/mytkom/AliceTraINT/internal/utils"
)

// NewRecoverMw returns middleware that recovers from panics, logs them with a
// stack trace, and returns a generic 500 response without leaking internals.
func NewRecoverMw() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("PANIC recovered in HTTP handler %s %s: %v\n%s",
						r.Method, r.URL.Path, rec, debug.Stack())

					const publicMsg = "unexpected internal server error"

					if utils.IsHTMXRequest(r) {
						w.Header().Set("Content-Type", "text/plain; charset=utf-8")
						w.WriteHeader(http.StatusInternalServerError)
						_, _ = w.Write([]byte(publicMsg))
						return
					}

					http.Error(w, publicMsg, http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

