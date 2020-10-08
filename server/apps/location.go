package apps

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
