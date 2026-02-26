package handler

import (
	"log"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/mytkom/AliceTraINT/internal/utils"
)

const (
	errMsgUserUnauthorized string = "user unauthorized"
)

// writeError logs the internal error with rich context and sends a safe,
// user-facing error message. For HTMX requests it returns just the text body
// (so it can be swapped into a target), while for normal requests it uses
// http.Error to render a minimal error page.
func writeError(w http.ResponseWriter, r *http.Request, status int, publicMsg string, err error) {
	log.Printf("HTTP ERROR %s %s [%d]: %s; internal error: %v",
		r.Method, r.URL.Path, status, publicMsg, err)

	if utils.IsHTMXRequest(r) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(publicMsg))
		return
	}

	http.Error(w, publicMsg, status)
}

// handleServiceError maps well-known service-layer errors to HTTP status codes
// and safe public messages, then delegates to writeError for logging and
// response formatting.
func handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch err.(type) {
	case *service.ErrHandlerNotFound:
		writeError(w, r, http.StatusNotFound, err.Error(), err)
	case *service.ErrHandlerValidation:
		writeError(w, r, http.StatusUnprocessableEntity, err.Error(), err)
	case *service.ErrExternalServiceTimeout:
		writeError(w, r, http.StatusServiceUnavailable, err.Error(), err)
	default:
		// Do not leak internal error details to the client.
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
	}
}
