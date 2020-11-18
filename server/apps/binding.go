package apps

type Binding struct {
	// For internal use by Mattermost, Apps do not need to set.
	AppID AppID `json:"app_id,omitempty"`

	// Location allows the App to identify where in the UX the Call request
	// comes from. It is optional. For /command bindings, Location is
	// defaulted to Label.
	Location Location `json:"location,omitempty"`

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
	// i.e. menu sub-items or subcommands. An app-defined Modal can be displayed
	// by setting AsModal.
	Call     *Call      `json:"call,omitempty"`
	Bindings []*Binding `json:"bindings,omitempty"`

	// Form allows to embed a form into a binding, and avoid the need to
	// Call(type=Form). At the moment, the sole use case is in-post forms, but
	// this may prove useful in other contexts.
	// TODO: Can embedded forms be mutable, and what does it mean?
	Form *Form `json:"form,omitempty"`
}
