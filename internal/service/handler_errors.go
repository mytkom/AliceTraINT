package service

import (
	"errors"
	"fmt"
	"os"
)

type ErrInternal struct {
	Op  string
	Err error
}

func (e *ErrInternal) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *ErrInternal) Unwrap() error {
	return e.Err
}

func (e *ErrInternal) PublicMessage() string {
	return "unexpected internal server error"
}

func InternalError(op string, err error) error {
	if err == nil {
		err = errors.New("unknown error")
	}
	return &ErrInternal{
		Op:  op,
		Err: err,
	}
}

func mapCCDBError(op string, err error) error {
	if err == nil {
		return nil
	}
	if os.IsTimeout(err) {
		return errCCDBUnreachable
	}
	return InternalError(op, err)
}

func mapJAlienError(_ string, err error) error {
	if err == nil {
		return nil
	}
	return errJAlienUnreachable
}

func mapMonalisaError(op string, err error) error {
	if err == nil {
		return nil
	}
	if os.IsTimeout(err) {
		return errMonalisaUnreachable
	}
	return InternalError(op, err)
}

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
