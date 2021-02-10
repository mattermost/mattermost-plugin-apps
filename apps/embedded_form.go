package apps

type AppPostEmbed struct {
	Title    string     `json:"title"`
	Text     string     `json:"text"`
	Bindings []*Binding `json:"bindings"`
}
