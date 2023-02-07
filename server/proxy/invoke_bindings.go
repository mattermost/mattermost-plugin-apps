package proxy

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// InvokeGetBindings fetches bindings for a specific apps. We should avoid
// unnecessary logging here as this route is called very often.
func (p *Proxy) InvokeGetBindings(r *incoming.Request, cc apps.Context) ([]apps.Binding, error) {
	app, err := p.getEnabledDestination(r)
	if err != nil {
		return nil, err
	}

	if len(app.GrantedLocations) == 0 {
		return nil, utils.NewForbiddenError("no location granted to bind to")
	}

	var problems error
	conf := p.conf.Get()
	// TODO PERF: Add caching
	bindingsCall := app.Bindings.WithDefault(apps.DefaultBindings)

	// no need to clean the context, Call will do it.
	resp := p.call(r, app, bindingsCall, &cc)
	switch resp.Type {
	case apps.CallResponseTypeOK:
		var bindings = []apps.Binding{}
		b, _ := json.Marshal(resp.Data)
		err := json.Unmarshal(b, &bindings)
		if err != nil {
			problems = multierror.Append(problems, errors.Wrap(err, "failed to decode bindings"))
			return nil, problems
		}
		bindings, err = cleanAppBindings(app, bindings, "", cc.UserAgent, conf)
		if err != nil {
			problems = multierror.Append(problems, err)
		}
		return bindings, problems

	case apps.CallResponseTypeError:
		problems = multierror.Append(problems, errors.Wrap(resp, "received app error"))
		return nil, problems

	default:
		problems = multierror.Append(problems, errors.Errorf("unexpected response type %q", string(resp.Type)))
		return nil, problems
	}
}

// cleanAppBindings removes bindings to locations that have not been granted to
// the App, and sets the AppID on the relevant elements.
func cleanAppBindings(app *apps.App, bindings []apps.Binding, locPrefix apps.Location, userAgent string, conf config.Config) ([]apps.Binding, error) {
	out := []apps.Binding{}
	usedLocations := map[apps.Location]bool{}
	usedCommandLabels := map[string]bool{}

	var problems error
	for _, b := range bindings {
		clean, err := cleanAppBinding(app, b, locPrefix, userAgent, conf)
		if err != nil {
			problems = multierror.Append(problems, err)
		}
		if clean == nil {
			continue
		}

		fql := locPrefix.Sub(clean.Location)
		if usedLocations[clean.Location] {
			problems = multierror.Append(problems,
				errors.Errorf("ignored duplicate command binding for location %q", clean.Location))
			continue
		}
		if fql.In(apps.LocationCommand) && usedCommandLabels[clean.Label] {
			problems = multierror.Append(problems,
				errors.Errorf("ignored duplicate command binding for label %q (location %q)", clean.Label, clean.Location))
			continue
		}

		out = append(out, *clean)
	}

	return out, problems
}

func cleanAppBinding(
	app *apps.App,
	b apps.Binding,
	locPrefix apps.Location,
	userAgent string,
	conf config.Config,
) (*apps.Binding, error) {
	var problems error
	if b.Location == "" && b.Label == "" {
		return nil, multierror.Append(problems, errors.Errorf("%s: sub-binding with no location nor label", locPrefix))
	}

	// Cleanup Location.
	if b.Location == "" {
		b.Location = apps.Location(b.Label)
	}
	if trimmed := apps.Location(strings.TrimSpace(string(b.Location))); trimmed != b.Location {
		problems = multierror.Append(problems, errors.Errorf("%s: trimmed whitespace from location", locPrefix.Sub(trimmed)))
		b.Location = trimmed
	}

	fql := locPrefix.Sub(b.Location)
	allowed := false
	for _, grantedLoc := range app.GrantedLocations {
		if fql.In(grantedLoc) {
			allowed = true
			break
		}
	}
	if !allowed {
		problems = multierror.Append(problems, utils.NewForbiddenError("%s: location is not granted", fql))
		return nil, problems
	}

	// Ensure all bindings except the top-level have the AppID set, forcefully.
	if !fql.IsTop() {
		b.AppID = app.AppID
	}

	// Cleanup (command) label.
	if fql != apps.LocationCommand && fql.In(apps.LocationCommand) {
		// A command binding must have a valid label. Default to Location if needed.
		if b.Label == "" {
			b.Label = string(b.Location)
		}
		if trimmed := strings.TrimSpace(b.Label); trimmed != b.Label {
			problems = multierror.Append(problems, errors.Errorf("%s: trimmed whitespace from label %s", fql, trimmed))
			b.Label = trimmed
		}
		if strings.ContainsAny(b.Label, " \t") {
			problems = multierror.Append(problems, errors.Errorf("%s: command label %q has multiple words", fql, b.Label))
			// A command binding with a white space in it will not parse, so bail.
			return nil, problems
		}
	}

	// Cleanup Icon.
	if b.Icon != "" {
		icon, err := normalizeStaticPath(conf, app.AppID, b.Icon)
		if err == nil {
			b.Icon = icon
		} else {
			problems = multierror.Append(problems, errors.Errorf("%s: invalid icon path %q in binding", fql, b.Icon))
			b.Icon = ""
		}
	}

	if fql == apps.LocationChannelHeader.Sub(b.Location) {
		// A channel header binding must have an icon, for webapp anyway.
		if b.Icon == "" && userAgent == "webapp" {
			problems = multierror.Append(problems, errors.Errorf("%s: no icon in channel header binding", fql))
			return nil, problems
		}
	}

	// A binding can have sub-bindings, a direct submit, or a form.
	hasBindings := len(b.Bindings) > 0
	hasForm := b.Form != nil
	hasSubmit := b.Submit != nil
	switch {
	// valid cases
	case hasBindings && !hasForm && !hasSubmit:
		var newProblems error
		b.Bindings, newProblems = cleanAppBindings(app, b.Bindings, fql, userAgent, conf)
		if newProblems != nil {
			problems = multierror.Append(problems, newProblems)
		}
		if len(b.Bindings) == 0 {
			// We do not add bindings without any valid sub-bindings
			return nil, problems
		}

	case hasForm && !hasSubmit && !hasBindings:
		clean, err := cleanForm(*b.Form, conf, app.AppID)
		if err != nil {
			problems = multierror.Append(problems, err)
		}
		b.Form = &clean

	case hasSubmit && !hasBindings && !hasForm:
		// nothing to clean for submit

	default:
		problems = multierror.Append(problems, errors.Errorf(`%s: (only) one of  "submit", "form", or "bindings" must be set in a binding`, fql))
		return nil, problems
	}

	return &b, problems
}
