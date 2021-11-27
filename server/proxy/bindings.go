package proxy

import (
	"encoding/json"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func mergeBindings(bb1, bb2 []apps.Binding) []apps.Binding {
	out := append([]apps.Binding(nil), bb1...)

	for _, b2 := range bb2 {
		found := false
		for i, o := range out {
			if b2.AppID == o.AppID && b2.Location == o.Location {
				found = true

				// b2 overrides b1, if b1 and b2 have Bindings, they are merged
				merged := b2
				if len(o.Bindings) != 0 {
					merged.Bindings = mergeBindings(o.Bindings, b2.Bindings)
				}
				out[i] = merged
			}
		}
		if !found {
			out = append(out, b2)
		}
	}
	return out
}

// GetBindings fetches bindings for all apps.
// We should avoid unnecessary logging here as this route is called very often.
func (p *Proxy) GetBindings(in Incoming, cc apps.Context) ([]apps.Binding, error) {
	all := make(chan []apps.Binding)
	defer close(all)

	allApps := store.SortApps(p.store.App.AsMap())
	for i := range allApps {
		app := allApps[i]

		go func(app apps.App) {
			bb, problems := p.GetAppBindings(in, cc, app)
			for _, problem := range problems {
				p.conf.Logger().WithError(problem).Debugw("Binding error")
			}
			all <- bb
		}(app)
	}

	ret := []apps.Binding{}
	for i := 0; i < len(allApps); i++ {
		bb := <-all
		ret = mergeBindings(ret, bb)
	}
	return ret, nil
}

// GetAppBindings fetches bindings for a specific apps. We should avoid
// unnecessary logging here as this route is called very often.
func (p *Proxy) GetAppBindings(in Incoming, cc apps.Context, app apps.App) ([]apps.Binding, []error) {
	var problems []error
	if !p.appIsEnabled(app) {
		return nil, problems
	}

	if len(app.GrantedLocations) == 0 {
		return nil, problems
	}

	conf, _, log := p.conf.Basic()
	log = log.With("app_id", app.AppID)
	appID := app.AppID
	cc.AppID = appID

	// TODO PERF: Add caching
	bindingsCall := app.Bindings.WithDefault(apps.DefaultBindings)

	// no need to clean the context, Call will do.
	resp := p.call(in, app, bindingsCall, &cc)
	switch resp.Type {
	case apps.CallResponseTypeOK:
		var bindings = []apps.Binding{}
		b, _ := json.Marshal(resp.Data)
		err := json.Unmarshal(b, &bindings)
		if err != nil {
			problems = append(problems, errors.Wrap(err, "failed to decode bindings"))
			return nil, problems
		}
		bindings, cleanupProblems := cleanAppBindings(app, bindings, "", cc.UserAgent, conf, log)
		return bindings, append(problems, cleanupProblems...)

	case apps.CallResponseTypeError:
		problems = append(problems, errors.Wrap(resp, "received app error"))
		return nil, problems

	default:
		problems = append(problems, errors.Errorf("unexpected response type %q", string(resp.Type)))
		return nil, problems
	}
}

// cleanAppBindings removes bindings to locations that have not been granted to
// the App, and sets the AppID on the relevant elements.
func cleanAppBindings(app apps.App, bindings []apps.Binding, locPrefix apps.Location, userAgent string, conf config.Config, baseLog utils.Logger) ([]apps.Binding, []error) {
	out := []apps.Binding{}
	usedLocations := map[apps.Location]bool{}
	usedCommandLabels := map[string]bool{}

	var problems []error
	for _, b := range bindings {
		fql := locPrefix.Sub(b.Location)
		log := baseLog.With("location", fql)

		clean, appProblems := cleanAppBinding(app, b, locPrefix, userAgent, conf, log)
		problems = append(problems, appProblems...)
		if clean == nil {
			continue
		}

		fql = locPrefix.Sub(clean.Location)
		if usedLocations[clean.Location] {
			problems = append(problems,
				errors.Errorf("ignored diplicate command binding for location %q", clean.Location))
			continue
		}
		if fql.In(apps.LocationCommand) && usedCommandLabels[clean.Label] {
			problems = append(problems,
				errors.Errorf("ignored diplicate command binding for label %q (location %q)", clean.Label, clean.Location))
			continue
		}

		out = append(out, *clean)
	}

	return out, problems
}

func cleanAppBinding(
	app apps.App,
	b apps.Binding,
	locPrefix apps.Location,
	userAgent string,
	conf config.Config,
	baseLog utils.Logger,
) (*apps.Binding, []error) {
	var problems []error

	// Cleanup Location.
	if b.Location == "" {
		b.Location = apps.Location(app.Manifest.AppID)
	}
	if trimmed := apps.Location(strings.TrimSpace(string(b.Location))); trimmed != b.Location {
		problems = append(problems, errors.Errorf("trimmed whitespace from location %s", trimmed))
		b.Location = trimmed
	}

	fql := locPrefix.Sub(b.Location)
	allowed := false
	for _, grantedLoc := range app.GrantedLocations {
		// TODO Why `grantedLoc.In(fql)`?
		if fql.In(grantedLoc) || grantedLoc.In(fql) {
			allowed = true
			break
		}
	}
	if !allowed {
		problems = append(problems, utils.NewForbiddenError("location %q is not granted", fql))
		return nil, problems
	}

	// Cleanup AppID.
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
			problems = append(problems, errors.Errorf("trimmed whitespace from label %s", trimmed))
			b.Label = trimmed
		}
		if strings.ContainsAny(b.Label, " \t") {
			problems = append(problems, errors.Errorf("command label %q has multiple words", b.Label))
			return nil, problems
		}
	}

	// Cleanup Icon.
	if b.Icon != "" {
		icon, err := normalizeStaticPath(conf, app.AppID, b.Icon)
		if err == nil {
			b.Icon = icon
		} else {
			problems = append(problems, errors.Errorf("invalid icon path %q in binding", b.Icon))
			b.Icon = ""
		}
	}

	if fql == apps.LocationChannelHeader.Sub(b.Location) {
		// A channel header binding must have an icon, for webapp anyway.
		if b.Icon == "" && userAgent == "webapp" {
			problems = append(problems, errors.Errorf("no icon in channel header binding %s", fql))
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
		var newProblems []error
		b.Bindings, newProblems = cleanAppBindings(app, b.Bindings, fql, userAgent, conf, baseLog)
		problems = append(problems, newProblems...)
		if len(b.Bindings) == 0 {
			// We do not add bindings without any valid sub-bindings
			return nil, problems
		}

	case hasForm && !hasSubmit && !hasBindings:
		clean, formProblems := cleanForm(*b.Form)
		problems = append(problems, formProblems...)
		b.Form = &clean

	case hasSubmit && !hasBindings && !hasForm:
		// nothing to clean for submit

	default:
		problems = append(problems, errors.New(`(only) one of  "submit", "form", or "bindings" must be set in a binding`))
		return nil, problems
	}

	return &b, problems
}

func (p *Proxy) dispatchRefreshBindingsEvent(userID string) {
	if userID != "" {
		p.conf.MattermostAPI().Frontend.PublishWebSocketEvent(
			config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: userID})
	}
}
