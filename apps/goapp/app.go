package goapp

import (
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
	Manifest apps.Manifest

	Log    utils.Logger
	Mode   apps.DeployType
	Router *mux.Router

	command       *BindableMulti
	postMenu      []Bindable
	channelHeader []Bindable
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

	// Bindings.
	app.HandleCall("/bindings", app.getBindings)

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

func (app *App) WithCommand(subcommands ...Bindable) *App {
	appCommand := NewBindableMulti(string(app.Manifest.AppID), subcommands...)
	app.command = &appCommand
	app.command.Init(app)
	return app
}

func (app *App) RunHTTP() error {
	if app.Manifest.Deploy.HTTP == nil {
		app.Manifest.Deploy.HTTP = &apps.HTTP{}
	}
	app.Mode = apps.DeployHTTP

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

	app.Log = utils.MustMakeCommandLogger(zapcore.DebugLevel)
	http.Handle("/", app.Router)

	listen := ":" + portStr
	app.Log.Infof("%s app started, listening on port %s, manifest at `%s/manifest.json`", app.Manifest.AppID, portStr, app.Manifest.Deploy.HTTP.RootURL)
	panic(http.ListenAndServe(listen, nil))
}
