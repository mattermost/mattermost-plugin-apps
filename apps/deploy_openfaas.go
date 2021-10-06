package apps

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type OpenFAAS struct {
	Functions []OpenFAASFunction `json:"functions,omitempty"`
}

func (o *OpenFAAS) Validate() error {
	if o == nil {
		return nil
	}
	if len(o.Functions) == 0 {
		return utils.NewInvalidError("must provide at least 1 function")
	}
	for _, of := range o.Functions {
		err := of.Validate()
		if err != nil {
			return errors.Wrapf(err, "invalid function %q", of.Name)
		}
	}
	return nil
}

// OpenFAASFunction defines a mapping of call paths to a function name.
// Functions themselves are defined in manifest.yml, in the app bundle, see
// upopenfaas.DeployApp for details.
type OpenFAASFunction struct {
	// Path is used to match/map incoming Call requests.
	Path string `json:"path"`

	// Name is the "short" name of the fuinction, it is combined with the app's
	// ID+Version when deployed, see upopenfaas.FunctionName.
	Name string `json:"name"`
}

func (of OpenFAASFunction) Validate() error {
	if of.Path == "" {
		return utils.NewInvalidError("invalid OpenFaaS function: path must not be empty")
	}
	if of.Name == "" {
		return utils.NewInvalidError("invalid OpenFaaS function: name must not be empty")
	}
	return nil
}
