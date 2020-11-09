package api

import (
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

const (
	LocationPostMenu      Location = "/post_menu"
	LocationChannelHeader Location = "/channel_header"
	LocationCommand       Location = "/command"
	LocationInPost        Location = "/in_post"
)

type Location string

type Locations []Location

func (ll Locations) toStringArray() []string {
	out := []string{}
	for _, current := range ll {
		out = append(out, string(current))
	}
	return out
}

func locationsFromConfigArray(in interface{}) Locations {
	out := Locations{}
	instr, _ := in.([]string)
	if len(instr) == 0 {
		return out
	}
	for _, current := range instr {
		out = append(out, Location(current))
	}
	return out
}

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

func (l Location) Make(sub Location) Location {
	out := l
	if sub[0] != '/' {
		out += "/"
	}
	return out + sub
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
