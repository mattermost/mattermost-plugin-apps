package goapp

import (
	"io/fs"
	"net/http"
	"net/url"
	"os"

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

func NewApp(m apps.Manifest, log utils.Logger) *App {
	if len(m.RequestedPermissions) == 0 {
		m.RequestedPermissions = []apps.Permission{
			apps.PermissionActAsBot,
		}
	}
	
	app := &App{
		Manifest: m,
		Log:      log,
		Router:   mux.NewRouter(),
	}

	// Ping.
	app.Router.Path("/ping").HandlerFunc(httputils.DoHandleJSONData([]byte("{}")))

	// GET manifest.json.
	app.Router.Path("/manifest.json").HandlerFunc(httputils.DoHandleJSON(&app.Manifest)).Methods("GET")

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

	if !app.Manifest.RequestedLocations.Contains(apps.LocationCommand) {
		app.Manifest.RequestedLocations = append(app.Manifest.RequestedLocations, apps.LocationCommand)
	}
	return app
}

func (app *App) WithPostMenu(items ...Bindable) *App {
	app.postMenu = items
	runInitializers(app.postMenu, app)

	if !app.Manifest.RequestedLocations.Contains(apps.LocationPostMenu) {
		app.Manifest.RequestedLocations = append(app.Manifest.RequestedLocations, apps.LocationPostMenu)
	}
	return app
}

func (app *App) WithChannelHeader(items ...Bindable) *App {
	app.channelHeader = items
	runInitializers(app.channelHeader, app)

	if !app.Manifest.RequestedLocations.Contains(apps.LocationChannelHeader) {
		app.Manifest.RequestedLocations = append(app.Manifest.RequestedLocations, apps.LocationChannelHeader)
	}
	return app
}

func (app *App) RunHTTP() error {
	app.Mode = apps.DeployHTTP
	if app.Manifest.Deploy.HTTP == nil {
		app.Log.Debugf("Using default HTTP deploy settings")
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
	app.Log.Infof("%s app started, listening on port %s, manifest at `%s/manifest.json`", app.Manifest.AppID, portStr, app.Manifest.Deploy.HTTP.RootURL)
	panic(http.ListenAndServe(listen, nil))
}
