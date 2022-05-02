package goapp

import (
	"errors"
	"io/fs"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	"go.uber.org/zap/zapcore"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type App struct {
	apps.Manifest

	Icon   string
	Log    utils.Logger
	Mode   apps.DeployType
	Router *mux.Router

	Bindables map[apps.Location]Bindable
}

func NewApp(m apps.Manifest) *App {
	app := &App{
		Manifest: m,
		Router:   mux.NewRouter(),
	}

	// Ping.
	app.Router.Path("/ping").HandlerFunc(httputils.DoHandleJSONData([]byte("{}")))

	// GET manifest.json.
	app.Router.Path("/manifest.json").HandlerFunc(httputils.DoHandleJSON(app.Manifest)).Methods("GET")

	app.Router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		app.Log.Debugf("App request: not found: %q", req.URL.String())
		http.NotFound(w, req)
	})

	return app
}

func (app *App) WithStatic(staticFS fs.FS) *App {
	app.Router.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticFS)))
	return app
}

func (app *App) WithIcon(iconPath string) *App {
	app.Icon = iconPath
	return app
}

func (app *App) RunHTTP() error {
	if app.Deploy.HTTP == nil {
		return errors.New("no HTTP in the app's manifest")
	}
	app.Mode = apps.DeployHTTP

	rootURL := os.Getenv("ROOT_URL")
	if rootURL != "" {
		app.Deploy.HTTP.RootURL = rootURL
	}

	portStr := os.Getenv("PORT")
	if portStr == "" {
		u, err := url.Parse(app.Deploy.HTTP.RootURL)
		if err != nil {
			panic(err)
		}
		portStr = u.Port()
		if portStr == "" {
			portStr = "8080"
		}
	}

	app.Log = utils.MustMakeCommandLogger(zapcore.DebugLevel)
	http.Handle("/", app.Router)

	listen := ":" + portStr
	app.Log.Infof("%s app started, listening on port %s, manifest at `%s/manifest.json`", app.AppID, portStr, app.Deploy.HTTP.RootURL)
	panic(http.ListenAndServe(listen, nil))
}

func (app *App) Bindings(bindings ...apps.Binding) *apps.Binding {



func (app *App) CommandBindings(bindings ...apps.Binding) *apps.Binding {
	return &apps.Binding{
		Location: apps.LocationCommand,
		Bindings: []apps.Binding{
			{
				Label:       string(app.AppID),
				Description: app.Description,
				Icon:        app.Icon,
				Bindings:    bindings,
			},
		},
	}
}

func (app *App) PostMenuBindings(bindings ...apps.Binding) *apps.Binding {
	return &apps.Binding{
		Location: apps.LocationPostMenu,
		Bindings: bindings,
	}
}

func (app *App) ChannelHeaderBindings(bindings ...apps.Binding) *apps.Binding {
	return &apps.Binding{
		Location: apps.LocationChannelHeader,
		Bindings: bindings,
	}
}

func AppendBinding(bb []apps.Binding, b *apps.Binding) []apps.Binding {
	var out []apps.Binding
	out = append(out, bb...)
	if b != nil {
		out = append(out, *b)
	}
	return out
}
