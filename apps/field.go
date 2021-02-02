package apps

type FieldType string

const (
	FieldTypeText          = FieldType("text")
	FieldTypeStaticSelect  = FieldType("static_select")
	FieldTypeDynamicSelect = FieldType("dynamic_select")
	FieldTypeBool          = FieldType("bool")
	FieldTypeUser          = FieldType("user")
	FieldTypeChannel       = FieldType("channel")
)

type SelectOption struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	IconData string `json:"icon_data"`
}

type Field struct {
	// Name is the name of the JSON field to use.
	Name       string    `json:"name"`
	Type       FieldType `json:"type"`
	IsRequired bool      `json:"is_required,omitempty"`

	// Present (default) value of the field
	Value interface{} `json:"value,omitempty"`

	Description string `json:"description,omitempty"`

	Label                string `json:"label,omitempty"`
	AutocompleteHint     string `json:"hint,omitempty"`
	AutocompletePosition int    `json:"position,omitempty"`

	ModalLabel string `json:"modal_label"`

	// Select props
	SelectRefresh       bool           `json:"refresh,omitempty"`
	SelectStaticOptions []SelectOption `json:"options,omitempty"`

	// Text props
	TextSubtype   string `json:"subtype,omitempty"`
	TextMinLength int    `json:"min_length,omitempty"`
	TextMaxLength int    `json:"max_length,omitempty"`
}
