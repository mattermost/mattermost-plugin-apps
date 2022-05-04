package goapp

import (
	"net/url"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type BindableForm struct {
	BindableAction
	form *apps.Form
	// TODO: add a formHandler when source= actually works
}

var _ Bindable = BindableForm{}

func NewBindableForm(name string, submitHandler HandlerFunc, form apps.Form) BindableForm {
	if form.Submit == nil {
		form.Submit = apps.NewCall("/" + url.PathEscape(name))
	}
	return BindableForm{
		BindableAction: NewBindableAction(name, submitHandler, *form.Submit),
		form:           &form,
	}
}

func (b BindableForm) prepareForm(creq CallRequest) *apps.Form {
	if b.form == nil {
		return nil
	}
	form := *b.form.PartialCopy()
	if form.Icon == "" {
		form.Icon = creq.App.Manifest.Icon
	}
	form.Submit = b.submit
	return &form
}

func (b BindableForm) Binding(creq CallRequest) *apps.Binding {
	binding := b.prepareBinding(creq)
	if binding == nil {
		return nil
	}

	binding.Form = b.prepareForm(creq)
	return binding
}
