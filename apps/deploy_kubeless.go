// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

// Kubeless is not longer supported: https://mattermost.atlassian.net/browse/MM-40011

package apps

import (
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// Kubeless contains metadata for an app that can be deployed to Kubeless
// running on a Kubernetes cluster, and is accessed using the Kubernetes APIs
// and HTTP. The JSON name `kubeless` must match the type.
type Kubeless struct {
	Functions []KubelessFunction `json:"functions,omitempty"`
}

func (k *Kubeless) Validate() error {
	if k == nil {
		return nil
	}
	if len(k.Functions) == 0 {
		return utils.NewInvalidError("must provide at least 1 function in kubeless_functions")
	}
	for _, kf := range k.Functions {
		err := kf.Validate()
		if err != nil {
			return errors.Wrapf(err, "invalid function %q", kf.Handler)
		}
	}
	return nil
}

// KubelessFunction describes a distinct Kubeless function defined by the app,
// and what path should be mapped to it.
//
// cmd/appsctl will create or update the functions in a kubeless service.
//
// upkubeless will find the closest match for the call's path, and then to
// invoke the kubeless function.
type KubelessFunction struct {
	// Path is used to match/map incoming Call requests.
	Path string `json:"path"`

	// Handler refers to the actual language function being invoked.
	// TODO examples py, go
	Handler string `json:"handler"`

	// File is the file ath (relative, in the bundle) to the function (source?)
	// file. Checksum is the expected checksum of the file.
	File string `json:"file"`

	// DepsFile is the path to the file with runtime-specific dependency list,
	// e.g. go.mod.
	DepsFile string `json:"deps_file"`

	// Kubeless runtime to use.
	Runtime string `json:"runtime"`

	// Timeout for the function to complete its execution, in seconds.
	Timeout int `json:"timeout"`

	// Port is the local ipv4 port that the function listens to, default 8080.
	Port int32 `json:"port"`
}

func (kf KubelessFunction) Validate() error {
	var result error
	if kf.Path == "" {
		result = multierror.Append(result,
			utils.NewInvalidError("invalid Kubeless function: path must not be empty"))
	}
	if kf.Handler == "" {
		result = multierror.Append(result,
			utils.NewInvalidError("invalid Kubeless function: handler must not be empty"))
	}
	if kf.Runtime == "" {
		result = multierror.Append(result,
			utils.NewInvalidError("invalid Kubeless function: runtime must not be empty"))
	}
	_, err := utils.CleanPath(kf.File)
	if err != nil {
		result = multierror.Append(result,
			errors.Wrap(err, "invalid Kubeless function: invalid file"))
	}
	if kf.DepsFile != "" {
		_, err := utils.CleanPath(kf.DepsFile)
		if err != nil {
			result = multierror.Append(result,
				errors.Wrap(err, "invalid Kubeless function: invalid deps_file"))
		}
	}
	if kf.Port < 0 || kf.Port > 65535 {
		result = multierror.Append(result,
			utils.NewInvalidError("invalid Kubeless function: port must be between 0 and 65535"))
	}
	return result
}
