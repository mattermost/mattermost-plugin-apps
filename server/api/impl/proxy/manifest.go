// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/mattermost/mattermost-plugin-apps/modelapps"
	"github.com/pkg/errors"
)

func LoadManifest(manifestURL string) (*modelapps.Manifest, error) {
	var manifest modelapps.Manifest
	resp, err := http.Get(manifestURL) // nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&manifest)
	if err != nil {
		return nil, err
	}
	err = validateManifest(&manifest)
	if err != nil {
		return nil, err
	}
	return &manifest, nil
}

func validateManifest(manifest *modelapps.Manifest) error {
	if manifest.AppID == "" {
		return errors.New("empty AppID")
	}
	if !manifest.Type.IsValid() {
		return errors.Errorf("invalid type: %s", manifest.Type)
	}

	if manifest.Type == modelapps.AppTypeHTTP {
		_, err := url.Parse(manifest.HTTPRootURL)
		if err != nil {
			return errors.Wrapf(err, "invalid manifest URL %q", manifest.HTTPRootURL)
		}
	}
	return nil
}
