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
	// ExpandNone means to include no data for the field.
	ExpandNone ExpandLevel = ""

	// ExpandID means include only the relevant ID(s), leaving other metadata
	// out.
	ExpandID ExpandLevel = "id"

	// ExpandSummary means to provide key metadata for the field.
	ExpandSummary ExpandLevel = "summary"

	// ExpandAll means to provide all the data available.
	ExpandAll ExpandLevel = "all"
)

func (l ExpandLevel) isRequired() bool {
	s := string(l)
	return len(s) > 0 && s[0] == '+'
}

func (l ExpandLevel) Required() ExpandLevel {
	if l.isRequired() {
		return l
	}
	return "+" + l
}

func ParseExpandLevel(raw ExpandLevel) (isRequired bool, clean ExpandLevel, err error) {
	level := raw
	if level.isRequired() {
		level = level[1:]
	}
	switch level {
	case ExpandNone, ExpandID, ExpandSummary, ExpandAll:
		return raw.isRequired(), level, nil
	default:
		return false, "", errors.Errorf("%q is not a known expand level", level)
	}
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
	// App (default: none, required). Details about the installed record of the
	// App. Of relevance to the app may be the version, and the Bot account
	// details.
	App ExpandLevel `json:"app,omitempty"`

	// ActingUser (default: none, required). Set to "all" for the entire
	// (sanitized) model.User; to "summary" for BotDescription, DeleteAt, Email,
	// FirstName, Id, IsBot, LastName, Locale, Nickname, Roles, Timezone,
	// Username; to "id" for Id only.
	ActingUser ExpandLevel `json:"acting_user,omitempty"`

	// ActingUserAccessToken (default: none, required): Set to "all" to include
	// user-level access token in the request. Requires act_as_user permission
	// to have been granted to the app. "summary" and "id" fail to expand.
	ActingUserAccessToken ExpandLevel `json:"acting_user_access_token,omitempty"`

	// Locale (default: none, optional) expands the locale to use for this call. There is
	// no difference between the modes.
	Locale ExpandLevel `json:"locale,omitempty"`

	// Channel (default: none, optional): Set to "all" for model.Channel; to
	// "summary" for Id, DeleteAt, TeamId, Type, DisplayName, Name; to "id" for
	// Id only.
	Channel ExpandLevel `json:"channel,omitempty"`

	// ChannelMember (default: none, optional): expand model.ChannelMember if
	// ChannelID and ActingUserID (or UserID) are set. if both ActingUserID and
	// UserID are set, it expands UserID, as may be relevant in
	// UserJoinedChannel notifications. "all" and "summary" include the same
	// full model.ChannelMember struct.
	ChannelMember ExpandLevel `json:"channel_member,omitempty"`

	// Team (default: none, optional): "all" for model.Team; "summary"
	// for Id, DisplayName, Name, Description, Email, Type; "id" for Id only.
	Team ExpandLevel `json:"team,omitempty"`

	// TeamMember (default: none, optional): expand model.TeamMember if TeamID
	// and ActingUserID (or UserID) are set. if both ActingUserID and UserID are
	// set, it expands UserID, as may be relevant in UserJoinedTeam
	// notifications. "all" and "summary" include the same full model.TeamMember
	// struct.
	TeamMember ExpandLevel `json:"team_member,omitempty"`

	// Post, RootPost (default: none, optional): all for model.Post, summary for
	// Id, Type, UserId, ChannelId, RootId, Message.
	Post     ExpandLevel `json:"post,omitempty"`
	RootPost ExpandLevel `json:"root_post,omitempty"`

	// User (default: none, optional): all for model.User, summary for
	// BotDescription, DeleteAt, Email, FirstName, Id, IsBot, LastName, Locale,
	// Nickname, Roles, Timezone, Username.
	User ExpandLevel `json:"user,omitempty"`

	// OAuth2App (default: none, required) expands the remote (3rd party) OAuth2
	// app configuration data.
	OAuth2App ExpandLevel `json:"oauth2_app,omitempty"`

	// OAuth2User (default: none, required) expands the remote (3rd party)
	// OAuth2 user (custom object, previously stored with
	// appclient.StoreOAuthUser).
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
