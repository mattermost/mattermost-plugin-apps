package goapp

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type BindableAction struct {
	bindable
	submit        *apps.Call
	submitHandler HandlerFunc
}

type asAction interface {
	actionPtr() *BindableAction
}

func (b *BindableAction) actionPtr() *BindableAction { return b }

var _ Bindable = (*BindableAction)(nil)
var _ Initializer = (*BindableAction)(nil)
var _ Requirer = (*BindableAction)(nil)

func MakeBindableActionOrPanic(name string, submitHandler HandlerFunc, opts ...BindableOption) *BindableAction {
	b, err := MakeBindableAction(name, submitHandler, opts...)
	if err != nil {
		panic(err)
	}
	return b
}

func MakeBindableAction(name string, submitHandler HandlerFunc, opts ...BindableOption) (*BindableAction, error) {
	b := &BindableAction{
		bindable: bindable{
			name: name,
		},
		submitHandler: submitHandler,
	}

	// Initialize the default submit.
	_ = WithSubmit(&apps.Call{})(b)

	for _, opt := range opts {
		err := opt(b)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func WithSubmit(submit *apps.Call) BindableOption {
	return optionWithActionPtr(func(b *BindableAction) {
		if submit == nil {
			submit = &apps.Call{}
		} else {
			submit = submit.PartialCopy()
		}

		if submit.Path == "" {
			submit.Path = pathFromName(b.name)
		}

		// Auto-fill Expand to satisfy RequireSystemAdmin, etc.
		if b.requireSystemAdmin {
			if b.submit.Expand == nil {
				b.submit.Expand = &apps.Expand{}
			}
			if b.submit.Expand.ActingUser == "" {
				b.submit.Expand.ActingUser = apps.ExpandSummary
			}
		}
		if b.requireConnectedUser {
			if b.submit.Expand == nil {
				b.submit.Expand = &apps.Expand{}
			}
			if b.submit.Expand.ActingUserAccessToken == "" {
				b.submit.Expand.ActingUserAccessToken = apps.ExpandAll
			}
			if b.submit.Expand.ActingUser == "" {
				b.submit.Expand.ActingUser = apps.ExpandSummary
			}
		}

		b.submit = submit
	})
}

func WithExpand(expand apps.Expand) BindableOption {
	return optionWithActionPtr(func(b *BindableAction) {
		b.submit.Expand = &expand
	})
}

func WithState(state interface{}) BindableOption {
	return optionWithActionPtr(func(b *BindableAction) {
		b.submit.State = state
	})
}

func (b BindableAction) Init(app *App) error {
	app.HandleCall(b.submit.Path, b.checkedHandler(b.submitHandler))
	return nil
}

func (b BindableAction) Binding(creq CallRequest) *apps.Binding {
	binding := b.prepareBinding(creq)
	if binding == nil {
		return nil
	}

	binding.Submit = b.submit
	return binding
}

func optionWithActionPtr(f func(*BindableAction)) BindableOption {
	return func(b Bindable) error {
		ba, ok := b.(asAction)
		if !ok {
			return errors.Errorf("bindable action %s: With... method called on a wrong type: %T", b, b)
		}
		f(ba.actionPtr())
		return nil
	}
}
