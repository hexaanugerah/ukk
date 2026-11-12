// Package apperror menyediakan taksonomi Kind yang dipakai errors.go tiap domain untuk membangun error sentinel kustom. Domain tidak pernah mengembalikan error DB mentah ke handler; dibungkus dengan Kind agar handler memetakan ke status HTTP.
package apperror

import "errors"

type Kind string

const (
	KindValidation   Kind = "validation"
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindInternal     Kind = "internal"
	KindRateLimited  Kind = "rate_limited"
)

type AppError struct {
	Kind    Kind
	Message string
	Code    string // e.g. "AUTH_001" — Part 2.6's error code catalog; empty for errors not in that catalog
	Err     error
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Kind)
}

func (e *AppError) Unwrap() error { return e.Err }

func New(kind Kind, message string, wrapped error) *AppError {
	return &AppError{Kind: kind, Message: message, Err: wrapped}
}

// WithCode melampirkan kode katalog ke AppError yang ada, mis. WithCode(AUTH_001). Dipakai sebagai panggilan berantai agar domain memilih kode hanya di tempat yang katalognya memang mendefinisikan.
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

func NotFound(message string) *AppError     { return New(KindNotFound, message, nil) }
func Validation(message string) *AppError   { return New(KindValidation, message, nil) }
func Conflict(message string) *AppError     { return New(KindConflict, message, nil) }
func Unauthorized(message string) *AppError { return New(KindUnauthorized, message, nil) }
func Forbidden(message string) *AppError    { return New(KindForbidden, message, nil) }
func Internal(err error) *AppError {
	return New(KindInternal, "Terjadi kesalahan. Silakan coba kembali.", err)
}

func As(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
