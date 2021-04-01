package apps

import (
	"encoding/json"
	"net/url"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

// Where static assets are.
const StaticFolder = "static"

type Manifest struct {
	AppID   AppID      `json:"app_id"`
	AppType AppType    `json:"app_type"`
	Version AppVersion `json:"version"`

	// HomepageURL is required.
	HomepageURL string `json:"homepage_url"`

	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`

	// Callbacks

	// Bindings must be implemented by the Apps to add any UX elements to the
	// Mattermost UI. The default values for its fields are,
	//  "path":"/bindings",
	Bindings *Call `json:"bindings,omitempty"`

	// OnInstall gets invoked when a sysadmin installs the App with a `/apps
	// install` command. It may return another call to the app, or a form to
	// display. The default values for its fields are,
	//  "path":"/install",
	//  "expand":{
	//    "app":"all",
	//	  "admin_access_token":"all"
	//   }
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
	AWSLambda []AWSLambdaFunction `json:"aws_lambda,omitempty"`
}

var DefaultOnInstall = &Call{
	Path: "/install",
	Expand: &Expand{
		App:              ExpandAll,
		AdminAccessToken: ExpandAll,
	},
}

var DefaultBindings = &Call{
	Path: "/bindings",
}

var DefaultGetOAuth2ConnectURL = &Call{
	Path: "/oauth2/connect",
	Expand: &Expand{
		ActingUser:            ExpandSummary,
		ActingUserAccessToken: ExpandAll,
		OAuth2App:             ExpandAll,
	},
}

var DefaultOnOAuth2Complete = &Call{
	Path: "/oauth2/complete",
	Expand: &Expand{
		ActingUser:            ExpandSummary,
		ActingUserAccessToken: ExpandAll,
		OAuth2App:             ExpandAll,
	},
}

func (m Manifest) IsValid() error {
	for _, f := range []func() error{
		m.AppID.IsValid,
		m.Version.IsValid,
		m.AppType.IsValid,
		m.RequestedPermissions.IsValid,
	} {
		if err := f(); err != nil {
			return err
		}
	}

	switch m.AppType {
	case AppTypeHTTP:
		_, err := url.Parse(m.HTTPRootURL)
		if err != nil {
			return utils.NewInvalidError(errors.Wrapf(err, "invalid root_url: %q", m.HTTPRootURL))
		}

	case AppTypeAWSLambda:
		if len(m.AWSLambda) == 0 {
			return utils.NewInvalidError("must provide at least 1 function in aws_lambda")
		}
		for _, l := range m.AWSLambda {
			err := l.IsValid()
			if err != nil {
				return errors.Wrapf(err, "%q is not valid", l.Name)
			}
		}
	}

	return nil
}

func ManifestFromJSON(data []byte) (*Manifest, error) {
	var m Manifest
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	err = m.IsValid()
	if err != nil {
		return nil, err
	}
	return &m, nil
}
