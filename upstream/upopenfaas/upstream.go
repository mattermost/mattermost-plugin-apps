// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upopenfaas

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/upstream/uphttp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	// Environment variable used by appsctl to find OpenFaaS
	EnvGatewayURL = "OPENFAAS_URL"

	// Environment variables set for function execution
	FuncEnvMode = "MODE"
	FuncEnvSelf = "SELF"

	// YAML file for provisioning
	ManifestYaml = "manifest.yml"
)

type Upstream struct {
	uphttp.Upstream

	Gateway string
}

var _ upstream.Upstream = (*Upstream)(nil)

func MakeUpstream(httpOut httpout.Service, devMode bool) (*Upstream, error) {
	gateway := os.Getenv(EnvGatewayURL)
	if gateway == "" {
		return nil, utils.NewNotFoundError(EnvGatewayURL + " environment variable must be defined")
	}
	up := &Upstream{
		Gateway: gateway,
	}
	up.Upstream = *uphttp.NewUpstream(httpOut, devMode, up.appRootURL)
	return up, nil
}

func (u *Upstream) appRootURL(app apps.App, path string) (string, error) {
	return RootURL(app.Manifest, u.Gateway, path)
}

func RootURL(m apps.Manifest, gateway, path string) (string, error) {
	if m.OpenFAAS == nil {
		return "", errors.Errorf("failed to get root URL: app %s has no open_faas section in manifest.json", m.AppID)
	}

	matchedPath := ""
	var matched apps.OpenFAASFunction
	for _, f := range m.OpenFAAS.Functions {
		if strings.HasPrefix(path, f.Path) {
			if len(f.Path) > len(matchedPath) {
				matched = f
				matchedPath = f.Path
			}
		}
	}

	if matchedPath == "" {
		return "", utils.NewNotFoundError("no function matched %q", path)
	}

	return fmt.Sprintf("%s/function/%s", gateway, FunctionName(m.AppID, m.Version, matched.Name)), nil
}
