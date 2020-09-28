package apps

type LocationType string

const (
	LocationPostMenuItem = "post_menu_item"
)

type Location interface {
	Type() LocationType
}

type location struct {
	LocationType LocationType
	LocationID   string
}

func (l location) Type() LocationType {
	return l.LocationType
}

type PostMenuItemLocation struct {
	location
}
