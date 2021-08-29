package apps

import (
	"encoding/json"
	"unicode"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// Where static assets are.
const StaticFolder = "static"

// Root Call path for incoming webhooks from remote (3rd party) systems. Each
// webhook URL should be in the form:
// "{PluginURL}/apps/{AppID}/webhook/{PATH}/.../?secret=XYZ", and it will invoke a
// Call with "/webhook/{PATH}"."
const PathWebhook = "/webhook"

var DefaultBindings = Call{
	Path: "/bindings",
}

var DefaultGetOAuth2ConnectURL = Call{
	Path: "/oauth2/connect",
	Expand: &Expand{
		ActingUser:            ExpandSummary,
		ActingUserAccessToken: ExpandAll,
		OAuth2App:             ExpandAll,
	},
}

var DefaultOnOAuth2Complete = Call{
	Path: "/oauth2/complete",
	Expand: &Expand{
		ActingUser:            ExpandSummary,
		ActingUserAccessToken: ExpandAll,
		OAuth2App:             ExpandAll,
		OAuth2User:            ExpandAll,
	},
}

type Manifest struct {
	// The AppID is a globally unique identifier that represents your app. IDs
	// must be at least 3 characters, at most 32 characters and must contain
	// only alphanumeric characters, dashes, underscores and periods.
	AppID   AppID   `json:"app_id"`
	AppType AppType `json:"app_type"`

	// Version of the app, formatted as v00.00.000
	Version AppVersion `json:"version"`

	// HomepageURL is required.
	HomepageURL string `json:"homepage_url"`

	// DisplayName and Description provide optional information about the App.
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`

	// Icon is a relative path in the static assets folder of an png image,
	// which is used to represent the App.
	Icon string `json:"icon,omitempty"`

	// Callbacks

	// Bindings must be implemented by the Apps to add any UX elements to the
	// Mattermost UI. The default values for its fields are,
	//  "path":"/bindings",
	Bindings *Call `json:"bindings,omitempty"`

	// OnInstall gets invoked when a sysadmin installs the App with a `/apps
	// install` command. It may return another call to the app, or a form to
	// display. It is not called unless explicitly provided in the manifest.
	OnInstall *Call `json:"on_install,omitempty"`

	// OnVersionChanged gets invoked when the Mattermost-recommended version of
	// the app no longer matches the previously installed one, and the app needs
	// to be upgraded/downgraded. It is not called unless explicitly provided in
	// the manifest.
	OnVersionChanged *Call `json:"on_version_changed,omitempty"`

	// OnUninstall gets invoked when a sysadmin uses the `/apps uninstall`
	// command, before the app is actually removed. It is not called unless
	// explicitly provided in the manifest.
	OnUninstall *Call `json:"on_uninstall,omitempty"`

	// OnEnable, OnDisable are not yet supported
	OnDisable *Call `json:"on_disable,omitempty"`
	OnEnable  *Call `json:"on_enable,omitempty"`

	// GetOAuth2ConnectURL is called when the App's "connect to 3rd party" link
	// is clicked, to be redirected to the OAuth flow. It must return Data set
	// to the remote OAuth2 redirect URL. A "state" string is created by the
	// proxy, and is passed to the app as a value. The state is  a 1-time secret
	// that is included in the connect URL, and will be used to validate OAuth2
	// complete callback.
	GetOAuth2ConnectURL *Call `json:"get_oauth2_connect_url,omitempty"`

	// OnOAuth2Complete gets called upon successful completion of the remote
	// (3rd party) OAuth2 flow, and after the "state" has already been
	// validated. It gets passed the URL query as Values. The App should obtain
	// the OAuth2 user token, and store it persistently for future use using
	// mmclient.StoreOAuth2User.
	OnOAuth2Complete *Call `json:"on_oauth2_complete,omitempty"`

	// Requested Access

	RequestedPermissions Permissions `json:"requested_permissions,omitempty"`

	// RequestedLocations is the list of top-level locations that the
	// application intends to bind to, e.g. `{"/post_menu", "/channel_header",
	// "/command/apptrigger"}``.
	RequestedLocations Locations `json:"requested_locations,omitempty"`

	// App type-specific fields

	// For HTTP Apps all paths are relative to the RootURL.
	HTTPRootURL string `json:"root_url,omitempty"`

	// AWSLambda must be included by the developer in the published manifest for
	// AWS apps. These declarations are used to:
	// - create AWS Lambda functions that will service requests in Mattermost
	// Cloud;
	// - define path->function mappings, aka "routes". The function with the
	// path matching as the longest prefix is used to handle a Call request.
	AWSLambda []AWSLambda `json:"aws_lambda,omitempty"`

	KubelessFunctions []KubelessFunction `json:"kubeless_functions,omitempty"`

	// PluginID is the ID of the plugin, which manages the app, if there is one.
	PluginID string `json:"plugin_id,omitempty"`
}

// KubelessFunction describes a distinct Kubeless function defined by the app, and
// what path should be mapped to it.
//
// cmd/appsctl will create or update the functions in a kubeless service.
//
// upkubeless will find the closest match for the call's path, and then to
// invoke the kubeless function.
type KubelessFunction struct {
	// CallPath is used to match/map incoming Call requests.
	CallPath string `json:"path"`

	// Handler refers to the actual language function being invoked.
	// TODO examples py, go
	Handler string `json:"handler"`

	// File is the file path (relative, in the bundle) to the function (source?)
	// file.
	File string `json:"file"`

	// DepsFile is the path to the file with runtime-specific dependency list,
	// e.g. go.mod.
	DepsFile string `json:"deps_file"`

	// Kubeless runtime to use. See https://kubeless.io/docs/runtimes/ for more.
	Runtime string `json:"runtime"`

	// Timeout for the function to complete its execution, in seconds.
	Timeout int `json:"timeout"`

	// Port is the local ipv4 port that the function listens to, default 8080.
	Port int32 `json:"port"`
}

func (kf KubelessFunction) Validate() error {
	if kf.CallPath == "" {
		return utils.NewInvalidError("invalid Kubeless function: path must not be empty")
	}
	if kf.Handler == "" {
		return utils.NewInvalidError("invalid Kubeless function: handler must not be empty")
	}
	if kf.Runtime == "" {
		return utils.NewInvalidError("invalid Kubeless function: runtime must not be empty")
	}
	_, err := utils.CleanPath(kf.File)
	if err != nil {
		return errors.Wrap(err, "invalid Kubeless function: invalid file")
	}
	if kf.DepsFile != "" {
		_, err := utils.CleanPath(kf.DepsFile)
		if err != nil {
			return errors.Wrap(err, "invalid Kubeless function: invalid deps_file")
		}
	}
	if kf.Port < 0 || kf.Port > 65535 {
		return utils.NewInvalidError("invalid Kubeless function: port must be between 0 and 65535")
	}
	return nil
}

// AWSLambda describes a distinct AWS Lambda function defined by the app, and
// what path should be mapped to it. See
// https://developers.mattermost.com/integrate/apps/deployment/#making-your-app-runnable-as-an-aws-lambda-function
// for more information.
//
// cmd/appsctl will create or update the manifest's aws_lambda functions in the
// AWS Lambda service.
//
// upawslambda will use the manifest's aws_lambda functions to find the closest
// match for the call's path, and then to invoke the AWS Lambda function.
type AWSLambda struct {
	// The lambda function with its Path the longest-matching prefix of the
	// call's Path will be invoked for a call.
	Path string `json:"path"`

	// TODO @iomodo
	Name    string `json:"name"`
	Handler string `json:"handler"`
	Runtime string `json:"runtime"`
}

func (f AWSLambda) Validate() error {
	if f.Path == "" {
		return utils.NewInvalidError("aws_lambda path must not be empty")
	}
	if f.Name == "" {
		return utils.NewInvalidError("aws_lambda name must not be empty")
	}
	if f.Handler == "" {
		return utils.NewInvalidError("aws_lambda handler must not be empty")
	}
	if f.Runtime == "" {
		return utils.NewInvalidError("aws_lambda runtime must not be empty")
	}
	return nil
}

func ManifestFromJSON(data []byte) (*Manifest, error) {
	var m Manifest
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	err = m.Validate()
	if err != nil {
		return nil, err
	}

	return &m, nil
}

type validator interface {
	Validate() error
}

func (m Manifest) Validate() error {
	if m.HomepageURL == "" {
		return utils.NewInvalidError(errors.New("homepage_url is empty"))
	}
	if err := utils.IsValidHTTPURL(m.HomepageURL); err != nil {
		return utils.NewInvalidError(errors.Wrapf(err, "homepage_url invalid: %q", m.HomepageURL))
	}

	if m.Icon != "" {
		_, err := utils.CleanStaticPath(m.Icon)
		if err != nil {
			return err
		}
	}

	for _, v := range []validator{
		m.AppID,
		m.Version,
		m.AppType,
		m.RequestedPermissions,
	} {
		if v != nil {
			if err := v.Validate(); err != nil {
				return err
			}
		}
	}

	switch m.AppType {
	case AppTypeHTTP:
		if m.HTTPRootURL == "" {
			return utils.NewInvalidError(errors.New("root_url must be set for HTTP apps"))
		}

		err := utils.IsValidHTTPURL(m.HTTPRootURL)
		if err != nil {
			return utils.NewInvalidError(errors.Wrapf(err, "invalid root_url: %q", m.HTTPRootURL))
		}

	case AppTypeAWSLambda:
		if len(m.AWSLambda) == 0 {
			return utils.NewInvalidError("must provide at least 1 function in aws_lambda")
		}
		for _, l := range m.AWSLambda {
			err := l.Validate()
			if err != nil {
				return errors.Wrapf(err, "%q is not valid", l.Name)
			}
		}

	case AppTypeKubeless:
		if len(m.KubelessFunctions) == 0 {
			return utils.NewInvalidError("must provide at least 1 function in kubeless_functions")
		}
		for _, kf := range m.KubelessFunctions {
			err := kf.Validate()
			if err != nil {
				return errors.Wrapf(err, "invalid function %q", kf.Handler)
			}
		}
	}
	return nil
}

// AppID is a globally unique identifier that represents a Mattermost App.
// An AppID is restricted to no more than 32 ASCII letters, numbers, '-', or '_'.
type AppID string

const (
	MinAppIDLength = 3
	MaxAppIDLength = 32
)

func (id AppID) Validate() error {
	if len(id) < MinAppIDLength {
		return utils.NewInvalidError("appID %s too short, should be %d bytes", id, MinAppIDLength)
	}

	if len(id) > MaxAppIDLength {
		return utils.NewInvalidError("appID %s too long, should be %d bytes", id, MaxAppIDLength)
	}

	for _, c := range id {
		if unicode.IsLetter(c) {
			continue
		}

		if unicode.IsNumber(c) {
			continue
		}

		if c == '-' || c == '_' || c == '.' {
			continue
		}

		return utils.NewInvalidError("invalid character '%c' in appID %q", c, id)
	}

	return nil
}

// AppType is the type of an app: http, aws_lambda, or builtin.
type AppType string

const (
	// HTTP app (default). All communications are done via HTTP requests. Paths
	// for both functions and static assets are appended to RootURL "as is".
	// Mattermost authenticates to the App with an optional shared secret based
	// JWT.
	AppTypeHTTP AppType = "http"

	// AWS Lambda app. All functions are called via AWS Lambda "Invoke" API,
	// using path mapping provided in the app's manifest. Static assets are
	// served out of AWS S3, using the "Download" method. Mattermost
	// authenticates to AWS, no authentication to the App is necessary.
	AppTypeAWSLambda AppType = "aws_lambda"

	AppTypeKubeless AppType = "kubeless"

	// Builtin app. All functions and resources are served by directly invoking
	// go functions. No manifest, no Mattermost to App authentication are
	// needed.
	AppTypeBuiltin AppType = "builtin"

	// An App running as a plugin. All communications are done via inter-plugin HTTP requests.
	// Authentication is done via the plugin.Context.SourcePluginId field.
	AppTypePlugin AppType = "plugin"
)

func (at AppType) Validate() error {
	switch at {
	case AppTypeHTTP, AppTypeAWSLambda, AppTypeBuiltin, AppTypeKubeless, AppTypePlugin:
		return nil
	default:
		return utils.NewInvalidError("%s is not a valid app type", at)
	}
}

// AppVersion is the version of a Mattermost App. AppVersion is expected to look
// like "v00_00_000".
type AppVersion string

const VersionFormat = "v00_00_000"

func (v AppVersion) Validate() error {
	if len(v) > len(VersionFormat) {
		return utils.NewInvalidError("version %s too long, should be in %s format", v, VersionFormat)
	}

	for _, c := range v {
		if unicode.IsLetter(c) {
			continue
		}

		if unicode.IsNumber(c) {
			continue
		}

		if c == '-' || c == '_' || c == '.' {
			continue
		}

		return utils.NewInvalidError("invalid character '%c' in appVersion", c)
	}

	return nil
}
