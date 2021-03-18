package utils

import (
	"encoding/json"
	"os"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/utils/fileutils"
	"github.com/pkg/errors"
)

func ToJSON(in interface{}) string {
	bb, err := json.Marshal(in)
	if err != nil {
		return ""
	}
	return string(bb)
}

// FindDir looks for the given directory in nearby ancestors relative to the current working
// directory as well as the directory of the executable, falling back to `./` if not found.
func FindDir(dir string) (string, bool) {
	commonBaseSearchPaths := []string{
		".",
		"..",
		"../..",
		"../../..",
		"../../../..",
	}
	found := fileutils.FindPath(dir, commonBaseSearchPaths, func(fileInfo os.FileInfo) bool {
		return fileInfo.IsDir()
	})
	if found == "" {
		return "./", false
	}

	return found, true
}

func EnsureSysadmin(mm *pluginapi.Client, userID string) error {
	if !mm.User.HasPermissionTo(userID, model.PERMISSION_MANAGE_SYSTEM) {
		return errors.Wrapf(ErrUnauthorized, "user must be a sysadmin")
	}
	return nil
}
