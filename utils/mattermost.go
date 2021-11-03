package utils

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-server/v6/model"
)

func CodeBlock(in string) string {
	return fmt.Sprintf("\n```\n%s\n```\n", in)
}

func JSONBlock(in interface{}) string {
	return CodeBlock(Pretty(in))
}

func EnsureSysAdmin(mm *pluginapi.Client, userID string) error {
	if !mm.User.HasPermissionTo(userID, model.PermissionManageSystem) {
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

func GetLocale(mm *pluginapi.Client, config *model.Config, userID string) string {
	u, _ := mm.User.Get(userID)
	return GetLocaleWithUser(config, u)
}

func GetLocaleWithUser(config *model.Config, user *model.User) string {
	if user != nil && user.Locale != "" {
		return user.Locale
	}

	if locale := config.LocalizationSettings.DefaultClientLocale; locale != nil && *locale != "" {
		return *locale
	}

	if locale := config.LocalizationSettings.DefaultServerLocale; locale != nil && *locale != "" {
		return *locale
	}

	return "en"
}

func LastN(s string, n int) string {
	out := []byte(s)
	if len(out) > n+3 {
		out = out[len(out)-n-3:]
	}
	for i := range out {
		if i < len(out)-n {
			out[i] = '*'
		}
	}
	return string(out)
}
