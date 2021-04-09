package utils

import (
	"encoding/json"
	"fmt"
	"os"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/utils/fileutils"
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

func EnsureSysAdmin(mm *pluginapi.Client, userID string) error {
	if !mm.User.HasPermissionTo(userID, model.PERMISSION_MANAGE_SYSTEM) {
		return NewUnauthorizedError("user must be a sysadmin")
	}
	return nil
}

func LoadSession(mm *pluginapi.Client, sessionID, actingUserID string) (*model.Session, error) {
	if actingUserID == "" {
		return nil, ErrUnauthorized
	}
	if sessionID == "" {
		return nil, NewUnauthorizedError("no user session")
	}
	session, err := mm.Session.Get(sessionID)
	if err != nil {
		return nil, NewUnauthorizedError(err)
	}
	if session.UserId != actingUserID {
		return nil, NewUnauthorizedError("user ID mismatch")
	}
	return session, nil
}

func DumpObject(c interface{}) {
	b, _ := json.MarshalIndent(c, "", "    ")
	fmt.Printf("%s\n", string(b))
}
