package utils

import "github.com/pkg/errors"

var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("forbidden")
var ErrUnauthorized = errors.New("unauthorized")
var ErrInvalid = errors.New("invalid input")

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

func NewNotFoundError(args ...interface{}) error     { return NewError(ErrNotFound, args...) }
func NewForbiddenError(args ...interface{}) error    { return NewError(ErrForbidden, args...) }
func NewUnauthorizedError(args ...interface{}) error { return NewError(ErrUnauthorized, args...) }
func NewInvalidError(args ...interface{}) error      { return NewError(ErrInvalid, args...) }
