package apps

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type LocationType string

const (
	LocationPostMenuItem      LocationType = "post_menu_item"
	LocationChannelHeaderIcon LocationType = "channel_header_icon"
)

type LocationRegistry struct {
	FetchURL string
	AppID    AppID
}

type LocationInt interface {
	GetType() LocationType
}

type Location struct {
	LocationType LocationType
	LocationID   string
	Wish         Wish
}

func (l *Location) GetType() LocationType {
	return l.LocationType
}

type PostMenuItemLocation struct {
	Location
	Icon string
	Text string
}

type ChannelHeaderIconLocation struct {
	Location
	DropdownText string
	AriaText     string
	Icon         string
}

func LocationFromMap(m map[string]interface{}) (LocationInt, error) {
	buf, err := json.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling map")
	}

	var bareLocation Location

	err = json.Unmarshal(buf, &bareLocation)
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
