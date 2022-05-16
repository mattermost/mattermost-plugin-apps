// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
	"unicode"

	"github.com/hashicorp/go-multierror"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

const MaxManifestSize = 1024 * 1024 // MaxManifestSize is the maximum size of a Manifest in bytes

var DefaultPing = Call{
	Path: "/ping",
}

var DefaultBindings = Call{
	Path: path.Bindings,
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

var DefaultOnRemoteWebhook = Call{
	Path: path.Webhook,
}

type Manifest struct {
	// Set to the version of the Apps plugin that stores it, e.g. "v0.8.0"
	SchemaVersion string `json:",omitempty"`

	// The AppID is a globally unique identifier that represents your app. IDs
	// must be at least 3 characters, at most 32 characters and must contain
	// only alphanumeric characters, dashes, underscores and periods.
	AppID AppID `json:"app_id,omitempty"`

	// Version of the app, formatted as v00.00.000
	Version AppVersion `json:"version,omitempty"`

	// HomepageURL is required.
	HomepageURL string `json:"homepage_url,omitempty"`

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
	// appclient.StoreOAuth2User.
	OnOAuth2Complete *Call `json:"on_oauth2_complete,omitempty"`

	// OnRemoteWebhook gets invoked when an HTTP webhook is received from a
	// remote system, and is optionally authenticated by Mattermost. The request
	// is passed to the call serialized as HTTPCallRequest (JSON).
	OnRemoteWebhook *Call `json:"on_remote_webhook,omitempty"`

	// Requested Access
	RequestedPermissions Permissions `json:"requested_permissions,omitempty"`

	// RemoteWebhookAuthType specifies how incoming webhook messages from remote
	// systems should be authenticated by Mattermost.
	RemoteWebhookAuthType RemoteWebhookAuthType `json:"remote_webhook_auth_type,omitempty"`

	// RequestedLocations is the list of top-level locations that the
	// application intends to bind to, e.g. `{"/post_menu", "/channel_header",
	// "/command/apptrigger"}``.
	RequestedLocations Locations `json:"requested_locations,omitempty"`

	// Deployment information
	Deploy

	// unexported data

	// v7AppType is the AppType field value if the Manifest was decoded from a
	// v0.7.x version. It is used in App.DecodeCompatibleManifest to set
	// DeployType.
	v7AppType string
}

// DecodeCompatibleManifest decodes any known version of manifest.json into the
// current format. Since App embeds Manifest anonymously, it appears impossible
// to implement json.Unmarshaler without introducing all kinds of complexities.
// Thus, custom functions to encode/decode JSON, with backwards compatibility
// support for App and Manifest.
func DecodeCompatibleManifest(data []byte) (m *Manifest, err error) {
	defer func() {
		if m != nil {
			err = m.Validate()
			if err != nil {
				m = nil
			}
		}
	}()

	err = json.Unmarshal(data, &m)
	// If failed to decode as current version, opportunistically try as a
	// v0.7.x. There was no schema version before, this condition may need to be
	// updated in the future.
	if err != nil || m.SchemaVersion == "" {
		m7 := ManifestV0_7{}
		_ = json.Unmarshal(data, &m7)
		if from7 := m7.Manifest(); from7 != nil {
			return from7, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

type validator interface {
	Validate() error
}

func (m Manifest) Validate() error {
	var result error
	if m.HomepageURL == "" {
		result = multierror.Append(result,
			utils.NewInvalidError("homepage_url is empty"))
	}
	if err := httputils.IsValidURL(m.HomepageURL); err != nil {
		result = multierror.Append(result,
			utils.NewInvalidError("homepage_url %q invalid: %v", m.HomepageURL, err))
	}

	if m.Icon != "" {
		_, err := utils.CleanStaticPath(m.Icon)
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	for _, v := range []validator{
		m.AppID,
		m.Version,
		m.RequestedPermissions,
		m.Deploy,
	} {
		if v != nil {
			if err := v.Validate(); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	return result
}

// AppID is a globally unique identifier that represents a Mattermost App.
// An AppID is restricted to no more than 32 ASCII letters, numbers, '-', or '_'.
type AppID string

const (
	MinAppIDLength = 3
	MaxAppIDLength = 32
)

func (id AppID) Validate() error {
	var result error
	if len(id) < MinAppIDLength {
		result = multierror.Append(result,
			utils.NewInvalidError("appID %s too short, should be %d bytes", id, MinAppIDLength))
	}

	if len(id) > MaxAppIDLength {
		result = multierror.Append(result,
			utils.NewInvalidError("appID %s too long, should be %d bytes", id, MaxAppIDLength))
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

		result = multierror.Append(result,
			utils.NewInvalidError("invalid character '%c' in appID %q", c, id))
	}

	return result
}

// AppVersion is the version of a Mattermost App. AppVersion is expected to look
// like "v00_00_000".
type AppVersion string

const VersionFormat = "v00_00_000"

func (v AppVersion) Validate() error {
	var result error
	if len(v) > len(VersionFormat) {
		result = multierror.Append(result,
			utils.NewInvalidError("version %s too long, should be in %s format", v, VersionFormat))
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

		result = multierror.Append(result,
			utils.NewInvalidError("invalid character '%c' in appVersion", c))
	}

	return result
}

type RemoteWebhookAuthType string

const (
	// No authentication means that the message will be accepted, and passed to
	// tha app. The app can perform its own authentication then. This is also
	// the default type.
	NoAuth = RemoteWebhookAuthType("none")

	// Secret authentication expects the App secret to be passed in the incoming
	// request's query as ?secret=appsecret.
	SecretAuth = RemoteWebhookAuthType("secret")

	// JWT authentication: not implemented yet
	JWTAuth = RemoteWebhookAuthType("jwt")
)
