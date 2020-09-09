package cloudapps

type Form struct {
	RefreshURL string
	Elements   []interface{} // of *XXXElement
}

type ElementType string

const (
	ElementTypeCommand       = ElementType("command")
	ElementTypeText          = ElementType("text")
	ElementTypeStaticSelect  = ElementType("static_select")
	ElementTypeDynamicSelect = ElementType("dynamic_select")
	// ElementTypeTime          = ElementType("time")
	ElementTypeBool    = ElementType("bool")
	ElementTypeUser    = ElementType("user")
	ElementTypeChannel = ElementType("channel")
)

type elementProps struct {
	Type ElementType `json:"type"`

	// Name is the name of the JSON field to use when submitting
	Name string `json:"name"`

	// Description is long description, can be used in modals in addition to
	// Label, and in Autocomplete as Help
	Description string `json:"description,omitempty"`

	IsRequired      bool `json:"is_required,omitempty"`
	RefreshOnChange bool `json:"refresh_on_change,omitempty"`
}

type staticSelectElementProps struct {
	Options []SelectOption `json:"options,omitempty"`
}

type dynamicSelectElementProps struct {
	URL string `json:"url,omitempty"`

	// ReloadOnChangeTo indicates that changes to the 
	ReloadOnChangeTo []string `json:"load_on_change,omitempty"`
}

type textElementProps struct {
	Subtype   string `json:"subtype,omitempty"`
	MinLength int    `json:"min_length,omitempty"`
	MaxLength int    `json:"max_length,omitempty"`
	// options - encoding, regexp, etc.
}

type SelectOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type unmarshalledElement struct {
	elementProps

	autocompleteElementProps
	modalElementProps

	staticSelectElementProps
	dynamicSelectElementProps
	textElementProps
}
