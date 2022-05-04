package goapp

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type BindableForm struct {
	BindableAction
	form *apps.Form
	// TODO: add a formHandler when source= actually works
}

var _ Bindable = BindableForm{}

func (b BindableAction) WithForm(form apps.Form) BindableForm {
	if form.Submit == nil {
		form.Submit = b.getSubmit()
	}
	if form.Submit.Path == "" {
		form.Submit.Path = b.path()
	}
	return BindableForm{
		BindableAction: b,
		form:           &form,
	}
}

func (b BindableForm) getForm(creq CallRequest) *apps.Form {
	if b.form == nil {
		return nil
	}
	form := *b.form.PartialCopy()
	if form.Icon == "" {
		form.Icon = creq.App.Manifest.Icon
	}
	return &form
}

func (b BindableForm) Binding(creq CallRequest) *apps.Binding {
	binding := b.prepareBinding(creq)
	if binding == nil {
		return nil
	}

	binding.Form = b.getForm(creq)
	return binding
}
