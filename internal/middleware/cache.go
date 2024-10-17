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
	buf *bytes.Buffer
}

func (mrw *myResponseWriter) Write(p []byte) (int, error) {
	return mrw.buf.Write(p)
}

func NewCacheMw(c *utils.Cache) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			// if not cached pass it to the next handlerFunc
			mrw := &myResponseWriter{
				ResponseWriter: w,
				buf:            &bytes.Buffer{},
			}

			next.ServeHTTP(mrw, r)

			cacheBuffer := bytes.Buffer{}
			multiWriter := io.MultiWriter(w, &cacheBuffer)

			if _, err := io.Copy(multiWriter, mrw.buf); err != nil {
				log.Printf("Failed to send out response from CacheMiddleware")
			}

			// cache result
			c.Set(cacheKey, cacheBuffer.String())
		})
	}
}
