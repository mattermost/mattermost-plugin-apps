package api

type Form struct {
	Title  string `json:"title,omitempty"`
	Header string `json:"header,omitempty"`
	Footer string `json:"footer,omitempty"`
	Icon   string `json:"icon,omitempty"`

	// DependsOn is the list of field names that when changed force reloading
	// the form. Values of the fields with values that are not included in the
	// refreshed form are lost.
	DependsOn        []string `json:"depends_on,omitempty"`
	DependsOnTeam    bool     `json:"depends_on_team,omitempty"`
	DependsOnChannel bool     `json:"depends_on_channel,omitempty"`
	DependsOnUser    bool     `json:"depends_on_user,omitempty"`
	DependsOnPost    bool     `json:"depends_on_post,omitempty"`

	Fields []*Field
}
