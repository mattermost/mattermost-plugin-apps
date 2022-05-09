// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// ExpandLevel uses the format of "[+|-]level", where level indicates how much
// data to include in the expanded context.
//
// "+" and "-" indicate what to do if the data needed to expand is not
// available. Setting as "+" means that the proxy will fail the call before
// calling the app. Setting as "-" means that the app will be called, with the
// indicated field missing or incomplete in the context. The +|- part itself is
// optional, by default some fields are required, and some are optional.
type ExpandLevel string

const (
	// ExpandDefault is the default Expand value.
	ExpandDefault ExpandLevel = ""

	// ExpandNone means to include no data for the field.
	ExpandNone ExpandLevel = "none"

	// ExpandID means include only the relevant ID(s), leaving other metadata
	// out.
	ExpandID ExpandLevel = "id"

	// ExpandSummary means to provide key metadata for the field.
	ExpandSummary ExpandLevel = "summary"

	// ExpandAll means to provide all the data available.
	ExpandAll ExpandLevel = "all"

	// ExpandRequired and ExpandOptional indicate how to treat the case when the
	// data is not available; required means fail the call, optional means pass
	// it through without the data.
	ExpandRequired ExpandLevel = "+"
	ExpandOptional ExpandLevel = "-"
)

func ParseExpandLevel(l ExpandLevel, optionalByDefault bool) (ExpandLevel, error) {
	prefix, l, err := l.parse()
	if err != nil {
		return "", err
	}
	if prefix == "" {
		if optionalByDefault {
			prefix = ExpandOptional
		} else {
			prefix = ExpandRequired
		}
	}
	return prefix + l, nil
}

func (l ExpandLevel) IsRequired() bool {
	prefix, _, _ := l.parse()
	return prefix == ExpandRequired
}

func (l ExpandLevel) IsOptional() bool {
	prefix, _, _ := l.parse()
	return prefix == ExpandOptional
}

func (l ExpandLevel) Level() ExpandLevel {
	_, level, _ := l.parse()
	return level
}

func (l ExpandLevel) parse() (prefix ExpandLevel, level ExpandLevel, err error) {
	switch {
	case strings.HasPrefix(string(l), string(ExpandRequired)):
		prefix = ExpandRequired
		level = l[len(ExpandRequired):]
	case strings.HasPrefix(string(l), string(ExpandOptional)):
		prefix = ExpandOptional
		level = l[len(ExpandOptional):]
	default:
		level = l
	}

	if level == ExpandNone {
		return "", level, nil
	}

	for _, known := range []ExpandLevel{ExpandDefault, ExpandID, ExpandSummary, ExpandAll} {
		if level == known {
			return prefix, level, nil
		}
	}
	return "", "", errors.Errorf("%q is not a known expand level", level)
}

// Expand is a clause in the Call struct that controls what additional
// information is to be provided in each request made.
//
// By default only the IDs of certain entities are provided in the request's
// Context. Expand allows to selectively add data to ExpandedContext, including
// privileged information such as access tokens, and detailed data on Mattermost
// entities, such as users and channels.
//
// Based on the app's GrantedPermissions, Bot, User, or Admin-level tokens may
// be provided in the request. If the app connects to a 3rd party, it may store
// authentication data in the Mattermost token store and get the token data
// expanded in the request.
//
// When expanding Mattermost data entities, the apps proxy must not exceed the
// highest available access level in the request's Context.
type Expand struct {
	// App: default:all,required. Details about the installed record of the App. Of relevance to
	// the app may be the version, and the Bot account details.
	App ExpandLevel `json:"app,omitempty"`

	// ActingUser: all for the entire model.User, summary for BotDescription,
	// DeleteAt, Email, FirstName, Id, IsBot, LastName, Locale, Nickname, Roles,
	// Timezone, Username.
	ActingUser ExpandLevel `json:"acting_user,omitempty"`

	// ActingUserAccessToken: all. Include user-level access token in the
	// request. Requires act_as_user permission to have been granted to the app.
	ActingUserAccessToken ExpandLevel `json:"acting_user_access_token,omitempty"`

	// Locale expands the locale to use for this call. There is no difference
	// between all or summary.
	Locale ExpandLevel `json:"locale,omitempty"`

	// Channel: all for model.Channel, summary for Id, DeleteAt, TeamId, Type,
	// DisplayName, Name
	Channel ExpandLevel `json:"channel,omitempty"`

	// ChannelMember: expand model.ChannelMember if ChannelID and
	// ActingUserID are set.
	ChannelMember ExpandLevel `json:"channel_member,omitempty"`

	// Team: all for model.Team, summary for Id, DisplayName, Name, Description,
	// Email, Type.
	Team ExpandLevel `json:"team,omitempty"`

	// TeamMember: expand model.TeamMember if TeamID and ActingUserID are set.
	TeamMember ExpandLevel `json:"team_member,omitempty"`

	// Post, RootPost: all for model.Post, summary for Id, Type, UserId,
	// ChannelId, RootId, Message.
	Post     ExpandLevel `json:"post,omitempty"`
	RootPost ExpandLevel `json:"root_post,omitempty"`

	// User: all for model.User, summary for BotDescription, DeleteAt, Email,
	// FirstName, Id, IsBot, LastName, Locale, Nickname, Roles, Timezone,
	// Username. Default is summary.
	User ExpandLevel `json:"user,omitempty"`

	// Not currently implemented
	Mentioned ExpandLevel `json:"mentioned,omitempty"`

	// OAuth2App expands the remote (3rd party) OAuth2 app configuration data.
	OAuth2App ExpandLevel `json:"oauth2_app,omitempty"`

	// OAuth2User expands the remote (3rd party) OAuth2 user (custom object,
	// previously stored with appclient.StoreOAuthUser).
	OAuth2User ExpandLevel `json:"oauth2_user,omitempty"`
}

func (e Expand) String() string {
	m := map[string]string{}
	utils.Remarshal(&m, e)
	ss := []string{}
	for k, v := range m {
		ss = append(ss, k+":"+v)
	}
	sort.Strings(ss)
	return strings.Join(ss, ",")
}
