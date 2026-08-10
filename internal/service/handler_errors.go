package service

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
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
	if isCCDBUnreachable(err) {
		return errCCDBUnreachable
	}
	return InternalError(op, err)
}

func isCCDBUnreachable(err error) bool {
	if err == nil {
		return false
	}
	if os.IsTimeout(err) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	var unknownAuth *x509.UnknownAuthorityError
	if errors.As(err, &unknownAuth) {
		return true
	}

	var certInvalid *x509.CertificateInvalidError
	if errors.As(err, &certInvalid) {
		return true
	}

	var hostnameErr x509.HostnameError
	if errors.As(err, &hostnameErr) {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "certificate") ||
		strings.Contains(msg, "tls:") ||
		strings.Contains(msg, "x509:") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "network is unreachable")
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
