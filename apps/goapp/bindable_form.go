package goapp

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// BindableForm is a bindable action, with a form attached to it. It allows
// binding submittable forms to commands for the autocomplete of parameters and
// flags, and to the channel header and post actions menu where they open as
// modal dialogs.
type BindableForm struct {
	*BindableAction

	// form is the template of the form to be used. If it contains a Submit, it
	// will be used as a template. Internally, the implementation copies it over
	// to the BindableAction to leverage its methods.
	form *apps.Form
	// TODO: add a formHandler when source= actually works
}

type asForm interface {
	formActionPtr() *BindableForm
}

func (b *BindableForm) formActionPtr() *BindableForm { return b }

var _ Bindable = (*BindableForm)(nil)
var _ Initializer = (*BindableForm)(nil)
var _ Requirer = (*BindableForm)(nil)
var _ asForm = (*BindableForm)(nil)
var _ asAction = (*BindableForm)(nil)

func MakeBindableFormOrPanic(name string, form apps.Form, submitHandler HandlerFunc, opts ...BindableOption) *BindableForm {
	b, err := MakeBindableForm(name, form, submitHandler, opts...)
	if err != nil {
		panic(err)
	}
	return b
}

func MakeBindableForm(name string, form apps.Form, submitHandler HandlerFunc, opts ...BindableOption) (*BindableForm, error) {
	action, err := MakeBindableAction(name, submitHandler, WithSubmit(form.Submit))
	if err != nil {
		return nil, err
	}

	b := &BindableForm{
		BindableAction: action,
		form:           &form,
	}

	for _, opt := range opts {
		err := opt(b)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
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
