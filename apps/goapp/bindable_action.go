package goapp

import (
	"net/url"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type BindableAction struct {
	bindable
	submit        *apps.Call
	submitHandler HandlerFunc
}

var _ Bindable = BindableAction{}
var _ Initializer = BindableAction{}
var _ Requirer = BindableAction{}

func NewBindableAction(name string, submit apps.Call, submitHandler HandlerFunc) BindableAction {
	return BindableAction{
		bindable: bindable{
			name: name,
		},
		submit:        &submit,
		submitHandler: submitHandler,
	}
}

func (b BindableAction) Init(app *App) {
	app.HandleCall(b.path(), b.checkedHandler(b.submitHandler))
}

func (b BindableAction) path() string {
	return url.PathEscape("/" + b.name)
}

func (b BindableAction) getSubmit() *apps.Call {
	s := b.submit
	if s == nil {
		s = apps.NewCall(b.path())
	}
	s = s.PartialCopy()
	if s.Path == "" {
		s.Path = b.path()
	}
	return s
}

func (b BindableAction) Binding(creq CallRequest) *apps.Binding {
	binding := b.prepareBinding(creq)
	if binding == nil {
		return nil
	}

	binding.Submit = b.getSubmit()
	return binding
}
