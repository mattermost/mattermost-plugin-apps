package helloapp

import (
	"io"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const (
	sampleIcon = "http://www.mattermost.org/wp-content/uploads/2016/04/icon.png"
)

func (h *helloapp) HandleLocations(w http.ResponseWriter, req *http.Request, userID, channelID string) {
	user, err := h.apps.Mattermost.User.Get(userID)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	reader, err := h.apps.Mattermost.User.GetProfileImage(userID)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}
	icon := new(strings.Builder)
	_, err = io.Copy(icon, reader)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	locations := []apps.LocationInt{
		&apps.ChannelHeaderIconLocation{
			Location: apps.Location{
				LocationType: apps.LocationChannelHeaderIcon,
				Wish:         store.NewWish(AppID, h.AppURL(PathWishPing)),
			},
			DropdownText: user.Username,
			AriaText:     user.Username,
			Icon:         sampleIcon,
		},
		&apps.PostMenuItemLocation{
			Location: apps.Location{
				LocationType: apps.LocationPostMenuItem,
				Wish:         store.NewWish(AppID, h.AppURL(PathWishPing)),
			},
			Text: user.Username,
			Icon: sampleIcon,
		},
		&apps.PostMenuItemLocation{
			Location: apps.Location{
				LocationType: apps.LocationPostMenuItem,
				Wish:         store.NewWish(AppID, h.AppURL(PathWishPing)),
			},
			Text: "Remove " + user.Username,
			Icon: sampleIcon,
		},
	}

	httputils.WriteJSON(w, locations)
}
