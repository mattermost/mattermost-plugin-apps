package goapp

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type BindableMulti struct {
	bindable
	children []Bindable
}

type asMulti interface {
	multiPtr() *BindableMulti
}

func (b *BindableMulti) multiPtr() *BindableMulti { return b }

var _ Bindable = BindableMulti{}
var _ Initializer = BindableMulti{}
var _ Requirer = BindableMulti{}

func MakeBindableMultiOrPanic(name string, opts ...BindableOption) *BindableMulti {
	b, err := MakeBindableMulti(name, opts...)
	if err != nil {
		panic(err)
	}
	return b
}

func MakeBindableMulti(name string, opts ...BindableOption) (*BindableMulti, error) {
	b := &BindableMulti{
		bindable: bindable{
			name: name,
		},
	}

	for _, opt := range opts {
		if err := opt(b); err != nil {
			return nil, err
		}
	}

	return b, nil
}

func WithChildren(children ...Bindable) BindableOption {
	return func(bb Bindable) error {
		i, ok := bb.(asMulti)
		if !ok {
			return errors.Errorf("bindable multi  %s: WithChildren method called on a wrong type: %T", bb, bb)
		}
		b := i.multiPtr()
		b.children = children
		return nil
	}
}

func (b BindableMulti) Init(app *App) error {
	return runInitializers(b.children, app)
}

func runInitializers(list []Bindable, app *App) error {
	for _, sub := range list {
		if i, ok := sub.(Initializer); ok {
			if err := i.Init(app); err != nil {
				return err
			}
		}
	}
	return nil
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
