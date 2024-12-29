package handler

import (
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/service"
)

const (
	errMsgUserUnauthorized string = "user unauthorized"
)

func handleServiceError(w http.ResponseWriter, err error) {
	switch err.(type) {
	case *service.ErrHandlerNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	case *service.ErrHandlerValidation:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	case *service.ErrExternalServiceTimeout:
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
