package apps

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type LocationType string

const (
	LocationPostMenuItem      LocationType = "post_menu_item"
	LocationChannelHeaderIcon LocationType = "channel_header_icon"
	LocationSlashCommand      LocationType = "slash_command"
)

type LocationInt interface {
	GetType() LocationType
}

type Location struct {
	AppID        string       `json:"app_id"`
	LocationType LocationType `json:"location_type"`
	LocationID   string       `json:"location_id"`
	FormURL      string       `json:"form_url"`
}

func (l *Location) GetType() LocationType {
	return l.LocationType
}

type PostMenuItemLocation struct {
	Location
	Icon string `json:"icon"`
	Text string `json:"text"`
}

type ChannelHeaderIconLocation struct {
	Location
	DropdownText string `json:"dropdown_text"`
	AriaText     string `json:"aria_text"`
	Icon         string `json:"icon"`
}

type SlashCommandLocation struct {
	Location
	Trigger string `json:"trigger"`
	Icon    string `json:"icon"`
	Text    string `json:"text"`
}

func LocationFromMap(m map[string]interface{}) (LocationInt, error) {
	buf, err := json.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling map")
	}

	var bareLocation Location

	err = json.Unmarshal(buf, &bareLocation)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshaling bare location")
	}
	switch bareLocation.GetType() {
	case LocationChannelHeaderIcon:
		var specificLocation ChannelHeaderIconLocation
		err = json.Unmarshal(buf, &specificLocation)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding channel header icon location")
		}
		return &specificLocation, nil
	case LocationPostMenuItem:
		var specificLocation PostMenuItemLocation
		err = json.Unmarshal(buf, &specificLocation)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding post menu item location")
		}
		return &specificLocation, nil
	case LocationSlashCommand:
		var specificLocation SlashCommandLocation
		err = json.Unmarshal(buf, &specificLocation)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding post menu item location")
		}
		return &specificLocation, nil
	}

	return nil, errors.New("location not recognized")
}

// Alternative
// type Location struct {
// 	LocationType LocationType
// 	LocationID   string
// 	Wish         Wish
// 	Extra        interface{}
// }

// type PostMenuItemExtra struct {
// 	Icon string
// 	Text string
// }
