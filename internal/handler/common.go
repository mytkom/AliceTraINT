package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

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
	log.Printf("HTTP ERROR %s %s [%d]: %s; internal error: %s",
		r.Method, r.URL.Path, status, publicMsg, formatErrorChain(err))

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
	var notFound *service.ErrHandlerNotFound
	var validation *service.ErrHandlerValidation
	var external *service.ErrExternalServiceTimeout
	var internal *service.ErrInternal

	switch {
	case errors.As(err, &notFound):
		writeError(w, r, http.StatusNotFound, err.Error(), err)
	case errors.As(err, &validation):
		writeError(w, r, http.StatusUnprocessableEntity, err.Error(), err)
	case errors.As(err, &external):
		writeError(w, r, http.StatusServiceUnavailable, err.Error(), err)
	case errors.As(err, &internal):
		writeError(w, r, http.StatusInternalServerError, internal.PublicMessage(), err)
	default:
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
	}
}

func formatErrorChain(err error) string {
	if err == nil {
		return "<nil>"
	}

	parts := make([]string, 0, 4)
	for err != nil {
		parts = append(parts, err.Error())
		err = errors.Unwrap(err)
	}

	return strings.Join(parts, " -> ")
}
