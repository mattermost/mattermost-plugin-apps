package api

type Function struct {
	Form     *Form      `json:"form"`
	Bindings []*Binding `json:"bindings,omitempty"`
	Expand   *Expand    `json:"expand,omitempty"`

	DependsOnTeam    bool `json:"depends_on_team,omitempty"`
	DependsOnChannel bool `json:"depends_on_channel,omitempty"`
	DependsOnUser    bool `json:"depends_on_user,omitempty"`
	DependsOnPost    bool `json:"depends_on_post,omitempty"`
}
