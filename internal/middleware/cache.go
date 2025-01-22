package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/utils"
)

type myResponseWriter struct {
	http.ResponseWriter
	buf    *bytes.Buffer
	status int
}

func (mrw *myResponseWriter) WriteHeader(code int) {
	mrw.status = code
	mrw.ResponseWriter.WriteHeader(code)
}

func (mrw *myResponseWriter) Write(p []byte) (int, error) {
	return mrw.buf.Write(p)
}

func NewCacheMw(c *utils.Cache) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Cache only GET requests
			if r.Method != "GET" {
				next.ServeHTTP(w, r)
				return
			}

			cacheKey := r.URL.String()
			if cached, found := c.Get(cacheKey); found {
				cachedBuffer := bytes.NewBufferString(cached)
				if _, err := io.Copy(w, cachedBuffer); err != nil {
					log.Printf("Failed to obtain cached response from CacheMiddleware")
				}
				return
			}

			// if not cached, pass it to the next handlerFunc
			mrw := &myResponseWriter{
				ResponseWriter: w,
				buf:            &bytes.Buffer{},
				status:         http.StatusOK, // Default to 200
			}

			next.ServeHTTP(mrw, r)

			// Cache only 200 OK responses
			if mrw.status == http.StatusOK {
				cacheBuffer := bytes.Buffer{}
				multiWriter := io.MultiWriter(w, &cacheBuffer)

				if _, err := io.Copy(multiWriter, mrw.buf); err != nil {
					log.Printf("Failed to send out response from CacheMiddleware")
				}

				// Save to cache
				c.Set(cacheKey, cacheBuffer.String())
			} else {
				// Write response directly if not 200 OK
				_, _ = io.Copy(w, mrw.buf)
			}
		})
	}
}
