// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upkubeless

import (
	"os"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// ProvisionApp creates Kubeless functions from an app bundle, as declared by
// the app's manifest.
func ProvisionApp(bundlePath string, log utils.Logger, shouldUpdate bool) (*apps.Manifest, error) {
	m, dir, err := upstream.GetAppBundle(bundlePath, log)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	if m.Kubeless == nil {
		return nil, errors.Wrap(err, "no 'kubeless' section in manifest.json")
	}

	kubelessPath, err := exec.LookPath("kubeless")
	if err != nil {
		return nil, errors.Wrap(err, "failed to find kubeless command. Please follow the steps from https://kubeless.io/docs/quick-start/")
	}

	// Provision functions.
	for _, kf := range m.Kubeless.Functions {
		name := FunctionName(m.AppID, m.Version, kf.Handler)

		verb := "deploy"
		if shouldUpdate {
			verb = "update"
		}
		args := []string{kubelessPath, "function", verb, name}

		args = append(args, "--handler", kf.Handler)
		args = append(args, "--namespace", Namespace)
		args = append(args, "--from-file", kf.File)
		args = append(args, "--runtime", kf.Runtime)
		if kf.DepsFile != "" {
			args = append(args, "--dependencies", kf.DepsFile)
		}
		if kf.Port != 0 {
			args = append(args, "--port", strconv.Itoa(int(kf.Port)))
		}
		if kf.Timeout != 0 {
			args = append(args, "--timeout", strconv.Itoa(kf.Timeout))
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
