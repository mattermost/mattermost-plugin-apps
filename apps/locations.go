// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"fmt"
	"strings"
)

const (
	LocationPostMenu      Location = "/post_menu"
	LocationChannelHeader Location = "/channel_header"
	LocationCommand       Location = "/command"
	LocationInPost        Location = "/in_post"
)

type Location string

type Locations []Location

func (l Location) IsTop() bool {
	switch l {
	case LocationChannelHeader,
		LocationCommand,
		LocationPostMenu:
		return true
	}
	return false
}

func (l Location) In(other Location) bool {
	return strings.HasPrefix(string(l), string(other))
}

func (l Location) Sub(sub Location) Location {
	out := l
	if len(sub) == 0 {
		return out
	}
	if sub[0] != '/' {
		out += "/"
	}
	return out + sub
}

func (l Location) Markdown() string {
	if l[0] != '/' {
		return string(l)
	}

	tokens := strings.Split(string(l)[1:], "/")
	if len(tokens) == 0 {
		return string(l)
	}

	switch Location("/" + tokens[0]) {
	case LocationPostMenu:
		return "Post Menu items"
	case LocationChannelHeader:
		return "Channel Header buttons"
	case LocationCommand:
		if len(tokens) < 2 {
			return "Slash commands"
		}
		return fmt.Sprintf("`/%s` command", strings.Join(tokens[1:], " "))
	}
	return string(l)
}

func (list Locations) Contains(loc Location) bool {
	for _, current := range list {
		if current == loc {
			return true
		}
	}
	return false
}
