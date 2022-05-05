package goapp

import (
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

func NewBindableAction(name string, submitHandler HandlerFunc) BindableAction {
	return BindableAction{
		bindable: bindable{
			name: name,
		},
		submitHandler: submitHandler,
	}
}

func (b BindableAction) WithSubmit(submit *apps.Call) BindableAction {
	if submit == nil {
		submit = &apps.Call{}
	} else {
		submit = submit.PartialCopy()
	}
	if submit.Path == "" {
		submit.Path = pathFromName(b.name)
	}

	b.submit = submit
	return b
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

func (b BindableAction) Binding(creq CallRequest) *apps.Binding {
	binding := b.prepareBinding(creq)
	if binding == nil {
		return nil
	}

	binding.Submit = b.submit
	return binding
}
