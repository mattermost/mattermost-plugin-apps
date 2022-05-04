package goapp

import (
	"context"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type CallRequest struct {
	apps.CallRequest

	App          *App
	GoContext    context.Context
	asBot        *appclient.Client
	asActingUser *appclient.Client
}

func (creq CallRequest) AsBot() *appclient.Client {
	if creq.asBot != nil {
		return creq.asBot
	}
	creq.asBot = appclient.AsBot(creq.Context)
	return creq.asBot
}

func (creq CallRequest) AsActingUser() *appclient.Client {
	if creq.asActingUser != nil {
		return creq.asActingUser
	}
	creq.asActingUser = appclient.AsActingUser(creq.Context)
	return creq.asActingUser
}

func (creq CallRequest) OAuth2User() *User {
	if creq.Context.OAuth2.User == nil {
		return nil
	}
	user := User{}
	utils.Remarshal(&user, creq.Context.OAuth2.User)
	return &user
}

// GetValue scans Values, then State if it is a map for a name, returns the
// value, or the default if not found.
func (creq CallRequest) GetValue(name, defaultValue string) string {
	if v, _ := creq.StringValue(name); v != "" {
		return v
	}
	return defaultValue
}

func (creq CallRequest) StringValue(name string) (string, bool) {
	value := func(m map[string]interface{}, n string) (string, bool) {
		s, ok := m[n].(string)
		if ok {
			return s, true
		}
		opt, ok := creq.Values[n].(map[string]interface{})
		if ok {
			if v, ok2 := opt["value"].(string); ok2 {
				return v, true
			}
		}
		return "", false
	}

	if s, found := value(creq.Values, name); found {
		return s, true
	}

	if stateInterfaces, ok := creq.CallRequest.State.(map[string]interface{}); ok {
		if s, found := value(stateInterfaces, name); found {
			return s, true
		}
	}

	if stateStrings, ok := creq.CallRequest.State.(map[string]string); ok {
		if s, found := stateStrings[name]; found {
			return s, true
		}
	}

	return "", false
}

func (creq CallRequest) BoolValue(name string) (value, found bool) {
	if len(creq.Values) == 0 {
		return false, false
	}

	isBool := func(v interface{}) (bool, bool) {
		if b, ok := v.(bool); ok {
			return b, true
		}
		if b, ok := creq.Values[name].(string); ok {
			switch b {
			case "true":
				return true, true
			case "false":
				return false, true
			}
		}
		return false, false
	}

	if b, ok := isBool(creq.Values[name]); ok {
		return b, true
	}

	if opt, ok := creq.Values[name].(map[string]interface{}); ok {
		if v, ok2 := isBool(opt["value"]); ok2 {
			return v, true
		}
	}

	if state, ok := creq.CallRequest.State.(map[string]interface{}); ok && len(state) > 0 {
		if b, ok2 := isBool(state[name]); ok2 {
			return b, true
		}
	}

	return false, false
}

func (creq CallRequest) IsSystemAdmin() bool {
	return creq.Context.ActingUser != nil && creq.Context.ActingUser.IsSystemAdmin()
}

func (creq CallRequest) IsConnectedUser() bool {
	return creq.Context.OAuth2.User != nil
}
