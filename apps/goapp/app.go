package goapp

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime/debug"
	"testing"
	"unicode"

	"github.com/gorilla/mux"
	"go.uber.org/zap/zapcore"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type App struct {
	Manifest apps.Manifest

	log    utils.Logger
	Mode   apps.DeployType
	Router *mux.Router

	command       *BindableMulti
	postMenu      []Bindable
	channelHeader []Bindable
	appBar        []Bindable
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
	if app.Manifest.Bindings == nil {
		app.Manifest.Bindings = apps.NewCall("/bindings")
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
		app.log.Debugf("App request: not found: %q", req.URL.String())
		http.NotFound(w, req)
	})

	return app, nil
}

func WithLog(log utils.Logger) AppOption {
	return func(app *App) error {
		app.log = log
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
		command, err := MakeBindableMulti(string(app.Manifest.AppID), WithChildren(subcommands...))
		if err != nil {
			return err
		}
		app.command = command

		err = app.command.Init(app)
		if err != nil {
			return err
		}

		if !app.Manifest.RequestedLocations.Contains(apps.LocationCommand) {
			app.Manifest.RequestedLocations = append(app.Manifest.RequestedLocations, apps.LocationCommand)
		}
		return nil
	}
}

func WithBindingsPath(path string) AppOption {
	return func(app *App) error {
		app.Manifest.Bindings.Path = path
		return nil
	}
}

func WithBindingsExpand(e *apps.Expand) AppOption {
	return func(app *App) error {
		app.Manifest.Bindings.Expand = e
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

func WithAppBar(items ...Bindable) AppOption {
	return func(app *App) error {
		app.appBar = items
		err := runInitializers(app.appBar, app)
		if err != nil {
			return err
		}

		if !app.Manifest.RequestedLocations.Contains(apps.LocationAppBar) {
			app.Manifest.RequestedLocations = append(app.Manifest.RequestedLocations, apps.LocationAppBar)
		}
		return nil
	}
}

func (app *App) RunHTTP() {
	if app.log == nil {
		app.log = utils.MustMakeCommandLogger(zapcore.DebugLevel)
	}

	app.Mode = apps.DeployHTTP
	if app.Manifest.Deploy.HTTP == nil {
		app.log.Debugf("Using default HTTP deploy settings")
		app.Manifest.Deploy.HTTP = &apps.HTTP{}
	}

	rootURL := os.Getenv("ROOT_URL")
	if rootURL != "" {
		app.Manifest.Deploy.HTTP.RootURL = rootURL
	}

	portStr := os.Getenv("PORT")
	if portStr == "" {
		u, err := url.Parse(app.Manifest.Deploy.HTTP.RootURL)
		if err != nil {
			panic(err)
		}
		portStr = u.Port()
		if portStr == "" {
			portStr = "8080"
		}
	}

	if app.Manifest.Deploy.HTTP.RootURL == "" {
		app.Manifest.Deploy.HTTP.RootURL = "http://localhost:" + portStr
	}

	http.Handle("/", app.Router)

	listen := ":" + portStr
	app.log.Infof("%s started, listening on port %s, manifest at `%s/manifest.json`; use environment variables PORT and ROOT_URL to customize.", app.Manifest.AppID, portStr, app.Manifest.Deploy.HTTP.RootURL)
	panic(http.ListenAndServe(listen, nil))
}

func pathFromName(name string) string {
	return "/" + url.PathEscape(string(locationFromName(name)))
}

func locationFromName(name string) apps.Location {
	b := bytes.Buffer{}
	for _, c := range name {
		if unicode.IsSpace(c) || c == '_' {
			c = '-'
		}
		_, _ = b.WriteRune(c)
	}
	return apps.Location(b.String())
}

func (app *App) NewTestServer(t testing.TB) *httptest.Server {
	if app.log == nil {
		app.log = utils.NewTestLogger()
	}
	app.Mode = apps.DeployHTTP
	if app.Manifest.Deploy.HTTP == nil {
		app.log.Debugf("Using default HTTP deploy settings")
		app.Manifest.Deploy.HTTP = &apps.HTTP{}
	}

	withPanicLog := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Output panics in dev. mode.
		defer func() {
			if x := recover(); x != nil {
				txt := fmt.Sprintf("Recovered from a panic in an HTTP handler, url: %q, error: %q, stack:\n%s",
					req.URL.String(), x, string(debug.Stack()))
				t.Log(txt)
				http.Error(w, txt, http.StatusInternalServerError)
			}
		}()

		app.Router.ServeHTTP(w, req)
	})

	appServer := httptest.NewServer(withPanicLog)
	rootURL := appServer.URL
	app.Manifest.Deploy.HTTP.RootURL = rootURL

	u, _ := url.Parse(rootURL)
	portStr := u.Port()
	app.log.Infof("%s started, listening on port %s, manifest at `%s/manifest.json`; use environment variables PORT and ROOT_URL to customize.", app.Manifest.AppID, portStr, app.Manifest.Deploy.HTTP.RootURL)
	return appServer
}

func TestWithBindingsHandler(h HandlerFunc) AppOption {
	return func(app *App) error {
		app.HandleCall("/bindings", h)
		return nil
	}
}
