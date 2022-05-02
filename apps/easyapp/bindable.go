package goapp

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type Bindable struct {
	// Name is the short "location" of the function, will also be used as the
	// default for command names and such.
	Name string

	// Display parameters, the use may be location-specific.
	Hint        string
	Description string
	Icon        string

	// Filters on who should see this.
	RequireAdmin         bool
	RequireConnectedUser bool

	BaseForm   *apps.Form
	BaseSubmit *apps.Call
	Handler func(CallRequest) apps.CallResponse
}

func (b Bindable) Path() string {
	return "/" + b.Name
}

func (b Bindable) Submit(creq CallRequest) *apps.Call {
	if b.BaseSubmit == nil {
		return nil
	}
	s := *b.BaseSubmit.PartialCopy()
	if s.Path == "" {
		s.Path = b.Path()
	}
	return &s
}

func (b Bindable) Form(creq CallRequest) *apps.Form {
	if b.BaseForm == nil {
		return nil
	}
	f := *b.BaseForm.PartialCopy()
	if f.Icon == "" {
		f.Icon = creq.App.Icon
	}
	if f.Submit == nil {
		f.Submit = b.Submit(creq)
	} else if f.Submit.Path == "" {
		f.Submit.Path = b.Path()
	}
	return &f
}

func (b Bindable) Binding(creq CallRequest) apps.Binding {
	binding := apps.Binding{
		Location:    apps.Location(b.Name),
		Icon:        b.Icon,
		Hint:        b.Hint,
		Description: b.Description,
		Submit:      b.Submit(creq),
		Form:        b.Form(creq),
	}
	if binding.Icon == "" {
		binding.Icon = creq.App.Icon
	}
	return binding
}
