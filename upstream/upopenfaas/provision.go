// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upopenfaas

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/openfaas/faas-cli/stack"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// ProvisionApp creates OpenFaaS functions from an app bundle, as declared by
// the app's manifest.
func ProvisionApp(bundlePath string, log utils.Logger, shouldUpdate bool, gateway, prefix string) (*apps.Manifest, error) {
	m, dir, err := upstream.GetAppBundle(bundlePath, log)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	yamlFile := filepath.Join(dir, ManifestYaml)
	parsedServices, err := stack.ParseYAMLFile(yamlFile, "", "", false)
	if err != nil {
		return nil, err
	}
	parsedServices.Provider.GatewayURL = gateway

	newFunctions := map[string]stack.Function{}
	for name, origF := range parsedServices.Functions {
		f := origF
		if f.Environment == nil {
			f.Environment = map[string]string{}
		}
		f.Name = FunctionName(m.AppID, m.Version, name)
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		f.Image = prefix + f.Image
		newFunctions[f.Name] = f
	}
	parsedServices.Functions = newFunctions

	yamlData, err := yaml.Marshal(parsedServices)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(yamlFile, yamlData, 0)
	if err != nil {
		return nil, err
	}

	// Provision functions.
	faascliPath, err := exec.LookPath("faas-cli")
	if err != nil {
		return nil, errors.Wrap(err, "failed to find faas-cli command. Please follow the steps from https://docs.openfaas.com/cli/install/")
	}
	cmd := exec.Cmd{
		Path:   faascliPath,
		Args:   []string{faascliPath, "up", "-f", ManifestYaml},
		Dir:    dir,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	log.Debugf("Run %s\n", cmd.String())
	err = cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, "failed to run faas-cli command")
	}

	return m, nil
}

func FunctionName(appID apps.AppID, version apps.AppVersion, name string) string {
	sanitizedAppID := strings.ReplaceAll(string(appID), ".", "-")
	sanitizedVersion := strings.ReplaceAll(string(version), ".", "-")
	sanitizedFunction := strings.ReplaceAll(name, " ", "-")
	sanitizedFunction = strings.ReplaceAll(sanitizedFunction, "_", "-")
	sanitizedFunction = strings.ReplaceAll(sanitizedFunction, ".", "-")
	sanitizedFunction = strings.ToLower(sanitizedFunction)
	return fmt.Sprintf("%s-%s-%s", sanitizedAppID, sanitizedVersion, sanitizedFunction)
}
