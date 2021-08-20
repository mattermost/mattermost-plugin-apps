package utils

import (
	"encoding/json"
	"fmt"
	"os"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/utils/fileutils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
)

func ToJSON(in interface{}) string {
	bb, err := json.Marshal(in)
	if err != nil {
		return ""
	}
	return string(bb)
}

func Pretty(in interface{}) string {
	bb, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return ""
	}
	return string(bb)
}

func CodeBlock(in string) string {
	return fmt.Sprintf("```\n%s\n```\n", in)
}

func JSONBlock(in interface{}) string {
	return CodeBlock(Pretty(in))
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

func ClientFromSession(mm *pluginapi.Client, mattermostSiteURL, sessionID, actingUserID string) (*model.Client4, LocError, error) {
	session, err := LoadSession(mm, sessionID, actingUserID)
	if err != nil {
		return nil, NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.utils.error.clienFromSession.failedSession",
				Other: "failed to load session",
			},
		}), errors.Wrap(err, "failed to load session")
	}

	client := model.NewAPIv4Client(mattermostSiteURL)
	client.SetToken(session.Token)

	return client, nil, nil
}

// DumpObject pretty prints any object to the standard output. Only used for debug.
func DumpObject(c interface{}) {
	b, _ := json.MarshalIndent(c, "", "    ")
	fmt.Printf("%s\n", string(b))
}

func LastN(s string, n int) string {
	out := []byte(s)
	for i := range out {
		if i < len(out)-n {
			out[i] = '*'
		}
	}
	return string(out)
}
