package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const (
	sampleIcon = "http://www.mattermost.org/wp-content/uploads/2016/04/icon.png"
)

func (h *helloapp) handleLocations(w http.ResponseWriter, req *http.Request, userID, channelID string) {
	locations := []apps.LocationInt{
		&apps.ChannelHeaderIconLocation{
			Location: apps.Location{
				LocationID:   "pingSomeone",
				LocationType: apps.LocationChannelHeaderIcon,
				FormURL:      h.appURL(pathCreateEmbeddedPing),
			},
			DropdownText: "Ping someone",
			AriaText:     "Ping someone",
			Icon:         sampleIcon,
		},
		&apps.PostMenuItemLocation{
			Location: apps.Location{
				LocationID:   "pingMePost",
				LocationType: apps.LocationPostMenuItem,
				FormURL:      h.appURL(pathPing),
			},
			Text: "Ping me this message",
			Icon: sampleIcon,
		},
		&apps.PostMenuItemLocation{
			Location: apps.Location{
				LocationID:   "pingSomeonePost",
				LocationType: apps.LocationPostMenuItem,
				FormURL:      h.appURL(pathOpenPingDialog),
			},
			Text: "Ping someone else this message",
			Icon: sampleIcon,
		},
	}

	httputils.WriteJSON(w, locations)
}
