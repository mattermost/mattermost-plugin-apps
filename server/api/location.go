package api

import (
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Location string

const (
	LocationPostMenu      Location = "/post_menu"
	LocationChannelHeader Location = "/channel_header"
	LocationCommand       Location = "/command"
	LocationInPost        Location = "/in_post"
)

func (l Location) In(other Location) bool {
	return strings.HasPrefix(string(l), string(other))
}

func (l Location) Markdown() md.MD {
	if l[0] != '/' {
		return md.MD(l)
	}

	tokens := strings.Split(string(l)[1:], "/")
	if len(tokens) == 0 {
		return md.MD(l)
	}

	switch Location(tokens[0]) {
	case LocationPostMenu:
		return "Post Menu items"
	case LocationChannelHeader:
		return "Channel Header buttons"
	case LocationCommand:
		if len(tokens) < 2 {
			return "Arbitrary /-commands"
		}
		return md.Markdownf("`/%s` command", strings.Join(tokens[1:], " "))
	}
	return md.MD(l)
}
