package path

// Paths for the REST APIs exposed by the Apps Plugin itself

const (
	// API paths: {PluginURL}/api/v1/...
	API = "/api/v1"

	// User-agent ping
	Ping = "/ping"

	// Services for Apps.
	KV                = "/kv"
	OAuth2App         = "/oauth2/app"
	OAuth2CreateState = "/oauth2/create-state"
	OAuth2User        = "/oauth2/user"
	Subscribe         = "/subscribe"
	Unsubscribe       = "/unsubscribe"

	// Invoke.
	Call = "/call"

	// Administration.
	EnableApp        = "/enable-app"
	DisableApp       = "/disable-app"
	InstallApp       = "/install-app"
	UninstallApp     = "/uninstall-app"
	UpdateAppListing = "/update-app-listing"

	// Marketplace and local manifest store.
	Marketplace = "/marketplace"

	// APIs for user agents.
	BotIDs      = "/bot-ids"
	OAuthAppIDs = "/oauth-app-ids"
)

// App (proxy) paths: {PluginURL}/apps/{AppID}/...
const (
	Apps = "/apps"

	// OAuth2 App's HTTP endpoints in the {PluginURL}/apps/{AppID} space.
	MattermostOAuth2Connect  = "/oauth2/mattermost/connect"
	MattermostOAuth2Complete = "/oauth2/mattermost/complete"
	RemoteOAuth2Connect      = "/oauth2/remote/connect"
	RemoteOAuth2Complete     = "/oauth2/remote/complete"

	// Root Call path for incoming webhooks from remote (3rd party) systems.
	// Each webhook URL should be in the form:
	// "{PluginURL}/apps/{AppID}/webhook/{PATH}/.../?secret=XYZ", and it will
	// invoke a Call with "/webhook/{PATH}"."
	Webhook = "/webhook"

	Bindings = "/bindings"

	// Static assets are served from {PluginURL}/static/...
	StaticFolder = "static"
	Static       = "/" + StaticFolder
)
