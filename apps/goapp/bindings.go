package goapp

import "github.com/mattermost/mattermost-plugin-apps/apps"

func (app *App) getBindings(creq CallRequest) apps.CallResponse {
	return apps.NewDataResponse(app.Bindings(creq))
}

func (app *App) Bindings(creq CallRequest) []apps.Binding {
	var out []apps.Binding

	if app.command != nil {
		if binding := app.command.Binding(creq); binding != nil {
			out = append(out, apps.Binding{
				Location: apps.LocationCommand,
				Bindings: []apps.Binding{*binding},
			})
		}
	}
	if bindings := MakeBindings(creq, app.channelHeader); len(bindings) > 0 {
		out = append(out, apps.Binding{
			Location: apps.LocationChannelHeader,
			Bindings: bindings,
		})
	}
	if bindings := MakeBindings(creq, app.postMenu); len(bindings) > 0 {
		out = append(out, apps.Binding{
			Location: apps.LocationPostMenu,
			Bindings: bindings,
		})
	}

	return out
}

func MakeBindings(creq CallRequest, bindables []Bindable) []apps.Binding {
	var out []apps.Binding
	for _, b := range bindables {
		if r, ok := b.(Requirer); ok {
			if r.RequireSystemAdmin() && !creq.IsSystemAdmin() {
				continue
			}
			if r.RequireConnectedUser() && !creq.IsConnectedUser() {
				continue
			}
		}

		binding := b.Binding(creq)
		if binding == nil {
			continue
		}

		out = append(out, *binding)
	}
	return out
}

func AppendBindings(orig []apps.Binding, extra ...*apps.Binding) []apps.Binding {
	bb := orig
	for _, b := range extra {
		if b == nil {
			continue
		}
		bb = append(bb, *b)
	}
	return bb
}
