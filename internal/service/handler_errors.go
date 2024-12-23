package service

import (
	"errors"
	"fmt"
)

type ErrHandlerNotFound struct {
	Resource string
}

func NewErrHandlerNotFound(resource string) *ErrHandlerNotFound {
	return &ErrHandlerNotFound{
		Resource: resource,
	}
}

func (e *ErrHandlerNotFound) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

var (
	errMsgNotUnique = "must be unique"
	errMsgMissing   = "missing"
)

type ErrHandlerValidation struct {
	Field string
	Msg   string
}

func (e *ErrHandlerValidation) Error() string {
	return fmt.Sprintf("%s %s", e.Field, e.Msg)
}

var (
	errInternalServerError = errors.New("unexpected internal server error")
)

type ErrExternalServiceTimeout struct {
	Service string
}

func NewErrExternalServiceTimeout(service string) *ErrExternalServiceTimeout {
	return &ErrExternalServiceTimeout{
		Service: service,
	}
}

func (e *ErrExternalServiceTimeout) Error() string {
	return fmt.Sprintf(`"%s" external service is unreachable`, e.Service)
}
