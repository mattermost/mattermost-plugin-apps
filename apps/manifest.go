package apps

import (
	"encoding/json"
	"net/url"

	"github.com/pkg/errors"
)

type Manifest struct {
	AppID   AppID      `json:"app_id"`
	AppType AppType    `json:"app_type"`
	Version AppVersion `json:"version"`

	// HomepageURL is required.
	HomepageURL string `json:"homepage_url"`

	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
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

	// Bindings must be implemented by the Apps to add any UX elements to the
	// Mattermost UI. The default values for its fields are,
	//  "path":"/bindings",
	Bindings *Call `json:"bindings,omitempty"`

	// OnEnable, OnDisable are not yet supported
	OnDisable *Call `json:"on_disable,omitempty"`
	OnEnable  *Call `json:"on_enable,omitempty"`

	// OnOAuth2Redirect must return
	OnOAuth2Redirect *Call `json:"on_oauth2_redirect,omitempty"`

	OnOAuth2Complete *Call `json:"on_oauth2_complete,omitempty"`

	RequestedPermissions Permissions `json:"requested_permissions,omitempty"`

	// RequestedLocations is the list of top-level locations that the
	// application intends to bind to, e.g. `{"/post_menu", "/channel_header",
	// "/command/apptrigger"}``.
	RequestedLocations Locations `json:"requested_locations,omitempty"`

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

var DefaultOnInstallCall = &Call{
	Path: "/install",
	Expand: &Expand{
		App:              ExpandAll,
		AdminAccessToken: ExpandAll,
	},
}

var DefaultBindingsCall = &Call{
	Path: "/bindings",
}

var DefaultOnOAuth2RedirectCall = &Call{
	Path: "/oauth2/remote/redirect",
}

var DefaultOnOAuth2CompleteCall = &Call{
	Path: "/oauth2/remote/complete",
}

func (m Manifest) IsValid() error {
	for _, f := range []func() error{
		m.AppID.IsValid,
		m.Version.IsValid,
		m.AppType.IsValid,
	} {
		if err := f(); err != nil {
			return err
		}
	}

	switch m.AppType {
	case AppTypeHTTP:
		_, err := url.Parse(m.HTTPRootURL)
		if err != nil {
			return errors.Wrapf(err, "invalid root_url: %q", m.HTTPRootURL)
		}

	case AppTypeAWSLambda:
		if len(m.AWSLambda) == 0 {
			return errors.New("must provide at least 1 function in aws_lambda")
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
