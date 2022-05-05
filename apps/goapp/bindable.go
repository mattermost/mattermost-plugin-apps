package goapp

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type Bindable interface {
	Binding(CallRequest) *apps.Binding
}

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

func (b bindable) WithDescription(description, hint string) bindable {
	b.hint = hint
	b.description = description
	return b
}

func (b bindable) WithIcon(icon string) bindable {
	b.icon = icon
	return b
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
	if b.requireConnectedUser && creq.IsConnectedUser() {
		return nil
	}

	binding := apps.Binding{
		Location:    apps.Location(pathFromName(b.name)),
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
