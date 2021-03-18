package utils

import (
	"encoding/json"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

func ToJSON(in interface{}) string {
	bb, err := json.Marshal(in)
	if err != nil {
		return ""
	}
	return string(bb)
}

func EnsureSysadmin(mm *pluginapi.Client, userID string) error {
	if !mm.User.HasPermissionTo(userID, model.PERMISSION_MANAGE_SYSTEM) {
		return errors.Wrapf(ErrUnauthorized, "user must be a sysadmin")
	}
	return nil
}
