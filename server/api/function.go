package api

type Function struct {
	Form   *Form   `json:"form"`
	Expand *Expand `json:"expand,omitempty"`
}
