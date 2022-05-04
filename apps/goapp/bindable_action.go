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

func NewBindableAction(name string, submitHandler HandlerFunc, submit apps.Call) BindableAction {
	if submit.Path == "" {
		submit.Path = "/" + url.PathEscape(name)
	}

	return BindableAction{
		bindable: bindable{
			name: name,
		},
		submitHandler: submitHandler,
		submit:        &submit,
	}
}

func (b BindableAction) WithExpand(e apps.Expand) BindableAction {
	b.submit = b.submit.WithExpand(e)
	return b
}

func (b BindableAction) WithState(state interface{}) BindableAction {
	b.submit = b.submit.WithState(state)
	return b
}

func (b BindableAction) Init(app *App) {
	app.HandleCall(b.submit.Path, b.checkedHandler(b.submitHandler))
}

func completeSubmit(s apps.Call, name string) *apps.Call {
	return &s
}

func (b BindableAction) Binding(creq CallRequest) *apps.Binding {
	binding := b.prepareBinding(creq)
	if binding == nil {
		return nil
	}

	binding.Submit = b.submit
	return binding
}
