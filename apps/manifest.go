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
	AppID AppID `json:"app_id"`

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

	// Deploy types, only those supported by the App should be populated.

	// HTTP contains metadata for an app that is already, deployed externally
	// and us accessed over HTTP. The JSON name `http` must match the type.
	HTTP *HTTP `json:"http,omitempty"`

	// AWSLambda contains metadata for an app that can be deployed to AWS Lambda
	// and S3 services, and is accessed using the AWS APIs. The JSON name
	// `aws_lambda` must match the type.
	AWSLambda *AWSLambda `json:"aws_lambda,omitempty"`

	// Kubeless contains metadata for an app that can be deployed to Kubeless
	// running on a Kubernetes cluster, and is accessed using the Kubernetes
	// APIs and HTTP. The JSON name `kubeless` must match the type.
	Kubeless *Kubeless `json:"kubeless,omitempty"`

	// Plugin contains metadata for an app that is implemented and is deployed
	// and accessed as a local Plugin. The JSON name `plugin` must match the
	// type.
	Plugin *Plugin `json:"plugin,omitempty"`
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

	// At least one deploy type must be supported.
	if m.HTTP == nil &&
		m.Plugin == nil &&
		m.AWSLambda == nil &&
		m.Kubeless == nil {
		return utils.NewInvalidError("manifest does not define an app type (http, aws_lambda, etc.)")
	}

	for _, v := range []validator{
		m.AppID,
		m.Version,
		m.RequestedPermissions,
		m.HTTP,
		m.AWSLambda,
		m.Kubeless,
		m.Plugin,
	} {
		if v != nil {
			if err := v.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m Manifest) MustDeployAs() DeployType {
	tt := m.DeployTypes()
	if len(tt) == 1 {
		return tt[0]
	}
	return ""
}

func (m Manifest) DeployTypes() (out []DeployType) {
	if m.AWSLambda != nil {
		out = append(out, DeployAWSLambda)
	}
	if m.HTTP != nil {
		out = append(out, DeployHTTP)
	}
	if m.Kubeless != nil {
		out = append(out, DeployKubeless)
	}
	if m.Plugin != nil {
		out = append(out, DeployPlugin)
	}
	return out
}

func (m Manifest) SupportsDeploy(dtype DeployType) bool {
	switch dtype {
	case DeployAWSLambda:
		return m.AWSLambda != nil && m.AWSLambda.Validate() == nil
	case DeployHTTP:
		return m.HTTP != nil && m.HTTP.Validate() == nil
	case DeployKubeless:
		return m.Kubeless != nil && m.Kubeless.Validate() == nil
	case DeployPlugin:
		return m.Plugin != nil && m.Plugin.Validate() == nil
	}
	return false
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
