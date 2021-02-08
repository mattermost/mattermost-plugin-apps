package apps

import "encoding/json"

type AppID string
type AppType string
type AppVersion string
type AppVersionMap map[AppID]AppVersion

// default is HTTP
const (
	AppTypeHTTP      = "http"
	AppTypeAWSLambda = "aws_lambda"
	AppTypeBuiltin   = "builtin"
)

func (at AppType) IsValid() bool {
	return at == AppTypeHTTP ||
		at == AppTypeAWSLambda ||
		at == AppTypeBuiltin
}

// AssetType describes static assets of the Mattermost App.
// Assets can be saved in S3 with appropriate permissions,
// or they could be fetched as ordinary http resources.
type AssetType string

const (
	s3Asset   AssetType = "s3_asset"
	httpAsset AssetType = "http_asset"
)

func (at AssetType) IsValid() bool {
	return at == s3Asset ||
		at == httpAsset
}

// AppStatus describes status of the app
type AppStatus string

const (
	AppStatusRegistered AppStatus = "registered"
	AppStatusEnabled    AppStatus = "enabled"
	AppStatusDisabled   AppStatus = "disabled"
)

// Function describes app's function mapping
// For now Function can be either AWS Lambda or HTTP function
type Function struct {
	Name    string `json:"name"`
	Handler string `json:"handler"`
	Runtime string `json:"runtime"`
}

// Asset describes app's static asset.
// For now asset can be an S3 file or an http resource
type Asset struct {
	Name   string    `json:"name"`
	Type   AssetType `json:"type"`
	URL    string    `json:"url"`
	Bucket string    `json:"bucket"`
	Key    string    `json:"key"`
}

type Manifest struct {
	AppID       AppID      `json:"app_id"`
	Type        AppType    `json:"app_type"`
	Version     AppVersion `json:"version"`
	DisplayName string     `json:"display_name,omitempty"`
	Description string     `json:"description,omitempty"`

	HomepageURL string `json:"homepage_url,omitempty"`

	// HTTPRootURL applicable For AppTypeHTTP.
	//
	// TODO: check if it is used in the // user-agent, consider naming
	// consistently.
	HTTPRootURL string `json:"root_url,omitempty"`

	RequestedPermissions Permissions `json:"requested_permissions,omitempty"`

	// RequestedLocations is the list of top-level locations that the
	// application intends to bind to, e.g. `{"/post_menu", "/channel_header",
	// "/command/apptrigger"}``.
	RequestedLocations Locations `json:"requested_locations,omitempty"`

	// By default invoke "/install", expanding App, AdminAccessToken, and
	// Config.
	OnInstall   *Call `json:"on_install,omitempty"`
	OnUninstall *Call `json:"on_uninstall,omitempty"`
	OnStartup   *Call `json:"on_startup,omitempty"`

	// By default invoke "/bindings".
	Bindings *Call `json:"bindings,omitempty"`

	// Deployment manifest for hostable apps will include path->invoke mappings
	Functions []Function
	Assets    []Asset
}

// Conventions for Apps paths, and field names
const (
	DefaultInstallCallPath  = "/install"
	DefaultBindingsCallPath = "/bindings"
)

var DefaultInstallCall = &Call{
	URL: DefaultInstallCallPath,
	Expand: &Expand{
		App:              ExpandAll,
		AdminAccessToken: ExpandAll,
	},
}

var DefaultBindingsCall = &Call{
	URL: DefaultBindingsCallPath,
}

type App struct {
	AppID    AppID     `json:"app_id"`
	Manifest *Manifest `json:"manifest"`
	Status   AppStatus `json:"app_status"`

	// Secret is used to issue JWT
	Secret string `json:"secret,omitempty"`

	OAuth2ClientID     string `json:"oauth2_client_id,omitempty"`
	OAuth2ClientSecret string `json:"oauth2_client_secret,omitempty"`
	OAuth2TrustedApp   bool   `json:"oauth2_trusted_app,omitempty"`

	BotUserID      string `json:"bot_user_id,omitempty"`
	BotUsername    string `json:"bot_username,omitempty"`
	BotAccessToken string `json:"bot_access_token,omitempty"`

	// Grants should be scopable in the future, per team, channel, post with
	// regexp.
	GrantedPermissions Permissions `json:"granted_permissions,omitempty"`

	// GrantedLocations contains the list of top locations that the
	// application is allowed to bind to.
	GrantedLocations Locations `json:"granted_locations,omitempty"`
}

func (a *App) ConfigMap() map[string]interface{} {
	b, _ := json.Marshal(a)
	var out map[string]interface{}
	_ = json.Unmarshal(b, &out)
	return out
}

func AppFromConfigMap(in interface{}) *App {
	b, _ := json.Marshal(in)
	var out App
	_ = json.Unmarshal(b, &out)
	return &out
}
