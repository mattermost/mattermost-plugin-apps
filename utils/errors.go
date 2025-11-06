package utils

import (
	"github.com/mattermost/mattermost/server/public/pluginapi/i18n"
	"github.com/pkg/errors"
)

var ErrAlreadyExists = errors.New("already exists")
var ErrForbidden = errors.New("forbidden")
var ErrInvalid = errors.New("invalid input")
var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")

func NewError(source error, args ...any) error {
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

func NewAlreadyExistsError(args ...any) error { return NewError(ErrAlreadyExists, args...) }
func NewForbiddenError(args ...any) error     { return NewError(ErrForbidden, args...) }
func NewInvalidError(args ...any) error       { return NewError(ErrInvalid, args...) }
func NewNotFoundError(args ...any) error      { return NewError(ErrNotFound, args...) }
func NewUnauthorizedError(args ...any) error  { return NewError(ErrUnauthorized, args...) }

type LocError []*i18n.LocalizeConfig

func NewLocError(err *i18n.LocalizeConfig) LocError {
	return LocError{err}
}
func (err LocError) Error(bundle *i18n.Bundle, loc *i18n.Localizer) string {
	errStr := ""
	for _, e := range err {
		if e.TemplateData == nil {
			e.TemplateData = map[string]any{}
		}
		e.TemplateData.(map[string]any)["Error"] = errStr
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
