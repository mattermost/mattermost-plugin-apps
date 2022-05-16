package goapp

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type Bindable interface {
	Binding(CallRequest) *apps.Binding
}

type BindableOption func(Bindable) error

type asBindable interface {
	bindablePtr() *bindable
}

func (b *bindable) bindablePtr() *bindable { return b }

type bindable struct {
	// name is the short "location" of the function, will also be used as the
	// default for command names and such.
	name string

	// Display parameters, the use may be location-specific.
	hint        string
	description string
	icon        string

	// Filters on who should see this.
	requireSystemAdmin   bool
	requireConnectedUser bool
}

var _ Requirer = bindable{}

func WithDescription(description string) BindableOption {
	return func(bb Bindable) error {
		i, ok := bb.(asBindable)
		if !ok {
			return errors.Errorf("bindable  %s: WithDescription method called on a wrong type: %T", bb, bb)
		}
		b := i.bindablePtr()
		b.description = description
		return nil
	}
}

func WithHint(hint string) BindableOption {
	return func(bb Bindable) error {
		i, ok := bb.(asBindable)
		if !ok {
			return errors.Errorf("bindable  %s: WithHint method called on a wrong type: %T", bb, bb)
		}
		b := i.bindablePtr()
		b.hint = hint
		return nil
	}
}

func WithIcon(icon string) BindableOption {
	return func(bb Bindable) error {
		i, ok := bb.(asBindable)
		if !ok {
			return errors.Errorf("bindable  %s: WithIcon method called on a wrong type: %T", bb, bb)
		}
		b := i.bindablePtr()
		b.icon = icon
		return nil
	}
}

func (b bindable) String() string {
	return b.name
}

func (b bindable) RequireSystemAdmin() bool   { return b.requireSystemAdmin }
func (b bindable) RequireConnectedUser() bool { return b.requireConnectedUser }

func (b bindable) prepareBinding(creq CallRequest) *apps.Binding {
	if b.requireSystemAdmin && !creq.IsSystemAdmin() {
		return nil
	}
	if b.requireConnectedUser && !creq.IsConnectedUser() {
		return nil
	}

	binding := apps.Binding{
		Location:    locationFromName(b.name),
		Icon:        b.icon,
		Hint:        b.hint,
		Description: b.description,
	}
	if binding.Icon == "" {
		binding.Icon = creq.App.Manifest.Icon
	}
	return &binding
}

func (b bindable) checkedHandler(h HandlerFunc) HandlerFunc {
	if b.requireSystemAdmin {
		h = RequireAdmin(h)
	}
	if b.requireConnectedUser {
		h = RequireConnectedUser(h)
	}
	return h
}
