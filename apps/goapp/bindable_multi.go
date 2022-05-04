package goapp

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type BindableMulti struct {
	bindable
	children []Bindable
}

var _ Bindable = BindableMulti{}
var _ Initializer = BindableMulti{}
var _ Requirer = BindableMulti{}

func NewBindableMulti(name string, children ...Bindable) BindableMulti {
	return BindableMulti{
		bindable: bindable{
			name: name,
		},
		children: children,
	}
}

func (b BindableMulti) Init(app *App) {
	runInitializers(b.children, app)
}

func runInitializers(list []Bindable, app *App) {
	for _, sub := range list {
		if i, ok := sub.(Initializer); ok {
			i.Init(app)
		}
	}
}

func (b BindableMulti) Binding(creq CallRequest) *apps.Binding {
	binding := b.prepareBinding(creq)
	if binding == nil {
		return nil
	}

	for _, sub := range b.children {
		subBinding := sub.Binding(creq)
		if subBinding != nil {
			binding.Bindings = append(binding.Bindings, *subBinding)
		}
	}
	return binding
}
