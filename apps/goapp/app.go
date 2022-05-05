package goapp

import (
	"bytes"
	"io/fs"
	"net/http"
	"net/url"
	"unicode"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type App struct {
	Manifest apps.Manifest

	Log    utils.Logger
	Mode   apps.DeployType
	Router *mux.Router

	command       *BindableMulti
	postMenu      []Bindable
	channelHeader []Bindable
}

type AppOption func(app *App) error

func MakeAppOrPanic(m apps.Manifest, opts ...AppOption) *App {
	app, err := MakeApp(m, opts...)
	if err != nil {
		panic(err)
	}
	return app
}

func MakeApp(m apps.Manifest, opts ...AppOption) (*App, error) {
	// Default the app's permissions
	if len(m.RequestedPermissions) == 0 {
		m.RequestedPermissions = []apps.Permission{
			apps.PermissionActAsBot,
		}
	}

	app := &App{
		Manifest: m,
		Router:   mux.NewRouter(),
	}

	// Run the options.
	for _, opt := range opts {
		err := opt(app)
		if err != nil {
			return nil, err
		}
	}

	// Set up the auto-served HTTP routes.
	app.Router.Path("/ping").HandlerFunc(httputils.DoHandleJSONData([]byte("{}")))
	app.Router.Path("/manifest.json").HandlerFunc(httputils.DoHandleJSON(&app.Manifest)).Methods("GET")
	app.HandleCall("/bindings", app.getBindings)

	app.Router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		app.Log.Debugf("App request: not found: %q", req.URL.String())
		http.NotFound(w, req)
	})

	return app, nil
}

func WithLog(log utils.Logger) AppOption {
	return func(app *App) error {
		app.Log = log
		return nil
	}
}

func WithStatic(staticFS fs.FS) AppOption {
	return func(app *App) error {
		app.Router.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticFS)))
		return nil
	}
}

func WithCommand(subcommands ...Bindable) AppOption {
	return func(app *App) error {
		appCommand := NewBindableMulti(string(app.Manifest.AppID), subcommands...)
		app.command = &appCommand
		err := app.command.Init(app)
		if err != nil {
			return err
		}

		if !app.Manifest.RequestedLocations.Contains(apps.LocationCommand) {
			app.Manifest.RequestedLocations = append(app.Manifest.RequestedLocations, apps.LocationCommand)
		}
		return nil
	}
}

func WithPostMenu(items ...Bindable) AppOption {
	return func(app *App) error {
		app.postMenu = items
		err := runInitializers(app.postMenu, app)
		if err != nil {
			return err
		}

		if !app.Manifest.RequestedLocations.Contains(apps.LocationPostMenu) {
			app.Manifest.RequestedLocations = append(app.Manifest.RequestedLocations, apps.LocationPostMenu)
		}
		return nil
	}
}

func WithChannelHeader(items ...Bindable) AppOption {
	return func(app *App) error {
		app.channelHeader = items
		err := runInitializers(app.channelHeader, app)
		if err != nil {
			return err
		}

		if !app.Manifest.RequestedLocations.Contains(apps.LocationChannelHeader) {
			app.Manifest.RequestedLocations = append(app.Manifest.RequestedLocations, apps.LocationChannelHeader)
		}
		return nil
	}
}

func pathFromName(name string) string {
	b := bytes.Buffer{}
	for _, c := range name {
		if unicode.IsSpace(c) {
			c = '-'
		}
		_, _ = b.WriteRune(c)
	}
	return "/" + url.PathEscape(b.String())
}
