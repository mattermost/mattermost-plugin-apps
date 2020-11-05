package api

type LocationID string

const (
	LocationPostMenu      LocationID = "/post_menu"
	LocationChannelHeader LocationID = "/channel_header"
	LocationCommand       LocationID = "/command"
)

type Binding struct {
	// For use by Mattermost only, not for apps
	AppID AppID `json:"app_id,omitempty"`

	LocationID LocationID `json:"location_id,omitempty"`

	// For PostMenu, ChannelHeader locations specifies the icon.
	Icon string `json:"icon,omitempty"`

	// Label is the (usually short) primary text to display at the location.
	// - For LocationPostMenu is the menu item text.
	// - For LocationChannelHeader is the dropdown text.
	// - For LocationCommand is the name of the command
	Label string `json:"label,omitempty"`

	// Hint is the secondary text to display
	// - LocationPostMenu: not used
	// - LocationChannelHeader: tooltip
	// - LocationCommand: the "Hint" line
	Hint string `json:"hint,omitempty"`

	// Description is the (optional) extended help text, used in modals and autocomplete
	Description string `json:"description,omitempty"`

	RoleID           string `json:"role_id,omitempty"`
	DependsOnTeam    bool   `json:"depends_on_team,omitempty"`
	DependsOnChannel bool   `json:"depends_on_channel,omitempty"`
	DependsOnUser    bool   `json:"depends_on_user,omitempty"`
	DependsOnPost    bool   `json:"depends_on_post,omitempty"`

	// A Binding is either to a Call, or is a "container" for other locations -
	// i.e. menu sub-items or subcommands.
	Call     *Call      `json:"call,omitempty"`
	Bindings []*Binding `json:"bindings,omitempty"`
}
