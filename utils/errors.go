package utils

import (
	"github.com/mattermost/mattermost-plugin-api/i18n"
	"github.com/pkg/errors"
)

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

type LocError []*i18n.LocalizeConfig

func NewLocError(err *i18n.LocalizeConfig) LocError {
	return LocError{err}
}
func (err LocError) Error(bundle *i18n.Bundle, loc *i18n.Localizer) string {
	errStr := ""
	for _, e := range err {
		if e.TemplateData == nil {
			e.TemplateData = map[string]interface{}{}
		}
		e.TemplateData.(map[string]interface{})["Error"] = errStr
		errStr = bundle.LocalizeWithConfig(loc, e)
	}

	return errStr
}
func (err LocError) Wrap(e *i18n.LocalizeConfig) LocError {
	if err == nil {
		return LocError{e}
	}
	return append(err, e)
}
