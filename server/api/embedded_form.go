package api

type EmbeddedForm struct {
	AppID    string     `json:"app_id"`
	Title    string     `json:"title"`
	Text     string     `json:"text"`
	Bindings []*Binding `json:"bindings"`
}
