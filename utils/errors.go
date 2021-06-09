package utils

import "github.com/pkg/errors"

var ErrAlreadyExists = errors.New("already exists")
var ErrForbidden = errors.New("forbidden")
var ErrInvalid = errors.New("invalid input")
var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")

func NewError(source error, args ...interface{}) error {
	s, _ := args[0].(string)
	err, _ := args[0].(error)

	switch {
	case len(args) == 0:
		return source

	case s != "":
		return errors.Wrapf(source, s, args[1:]...)

	case err != nil:
		return errors.Wrap(source, err.Error())

	default:
		return source
	}
}

func NewAlreadyExistsError(args ...interface{}) error { return NewError(ErrAlreadyExists, args...) }
func NewForbiddenError(args ...interface{}) error     { return NewError(ErrForbidden, args...) }
func NewInvalidError(args ...interface{}) error       { return NewError(ErrInvalid, args...) }
func NewNotFoundError(args ...interface{}) error      { return NewError(ErrNotFound, args...) }
func NewUnauthorizedError(args ...interface{}) error  { return NewError(ErrUnauthorized, args...) }
