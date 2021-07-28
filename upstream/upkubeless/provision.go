// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upkubeless

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// ProvisionApp creates all necessary functions in Kubeless, and outputs
// the manifest to use.
//
// Its input is a zip file containing:
//   |-- manifest.json
//   |-- function files referenced in manifest.json...
func ProvisionApp(bundlePath string, log utils.Logger, shouldUpdate bool) (*apps.Manifest, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp directory to unpack the bundle")
	}
	defer os.RemoveAll(dir)

	pwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain current working directory")
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
		return nil, errors.Wrap(err, "failed to get bundle "+bundlePath)
	}

	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to load manifest.json")
	}
	m, err := apps.ManifestFromJSON(data)
	if err != nil {
		return nil, errors.Wrap(err, "invalid manifest.json")
	}
	if log != nil {
		log.Debugw("Loaded App bundle",
			"bundle", bundlePath,
			"app_id", m.AppID)
	}

	kubelessPath, err := exec.LookPath("kubeless")
	if err != nil {
		return nil, errors.Wrap(err, "failed to find kubeless command")
	}

	// Provision functions.
	for _, kf := range m.KubelessFunctions {
		name := FunctionName(m.AppID, m.Version, kf.Handler)
		ns := namespace(m.AppID)

		verb := "deploy"
		if shouldUpdate {
			verb = "update"
		}
		args := []string{"", "function", verb, name}

		args = append(args, "--handler", kf.Handler)
		args = append(args, "--namespace", ns)
		args = append(args, "--from-file", kf.File)
		args = append(args, "--runtime", kf.Runtime)
		if kf.DepsFile != "" {
			args = append(args, "--dependencies", kf.DepsFile)
		}
		if kf.Port != 0 {
			args = append(args, "--port", strconv.Itoa(int(kf.Port)))
		}

		cmd := exec.Cmd{
			Path: kubelessPath,
			Args: args,
			Dir:  dir,
		}
		log.Debugf("Run %s\n", cmd.String())
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("Command failed, output:\n%s", string(out))
			return nil, errors.Wrap(err, "failed to run kubeless command")
		}
	}

	return m, nil
}
