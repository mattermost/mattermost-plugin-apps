// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upstream

import (
	"context"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// GetAppBundle loads and unzips the bundle into a temp folder, parses the
// manifest.
func GetAppBundle(bundlePath string, log utils.Logger) (*apps.Manifest, string, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to create temp directory to unpack the bundle")
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to get current working directory")
	}

	getBundle := getter.Client{
		Mode: getter.ClientModeDir,
		Src:  bundlePath,
		Dst:  dir,
		Pwd:  pwd,
		Ctx:  context.Background(),
	}
	err = getBundle.Get()
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to get app bundle "+bundlePath)
	}

	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to load manifest.json")
	}
	m, err := apps.DecodeCompatibleManifest(data)
	if err != nil {
		return nil, "", errors.Wrap(err, "invalid manifest.json")
	}
	log.Debugw("loaded App bundle",
		"bundle", bundlePath,
		"app_id", m.AppID)

	return m, dir, nil
}
