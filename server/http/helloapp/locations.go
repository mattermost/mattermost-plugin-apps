package helloapp

import (
	"io"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
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
				Wish:         *apps.NewWish(AppID, h.AppURL(PathWishSample)),
			},
			DropdownText: user.Username,
			AriaText:     user.Username,
			Icon:         "https://www.wizcase.com/wp-content/uploads/2020/06/Zoom-Logo.jpg",
		},
		&apps.PostMenuItemLocation{
			Location: apps.Location{
				LocationType: apps.LocationPostMenuItem,
				Wish:         *apps.NewWish(AppID, h.AppURL(PathWishSample)),
			},
			Text: user.Username,
			Icon: "https://www.wizcase.com/wp-content/uploads/2020/06/Zoom-Logo.jpg",
		},
		&apps.PostMenuItemLocation{
			Location: apps.Location{
				LocationType: apps.LocationPostMenuItem,
				Wish:         *apps.NewWish(AppID, h.AppURL(PathWishSample)),
			},
			Text: "Remove " + user.Username,
			Icon: "https://www.wizcase.com/wp-content/uploads/2020/06/Zoom-Logo.jpg",
		},
	}

	httputils.WriteJSON(w, locations)
}
