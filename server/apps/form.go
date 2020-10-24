package apps

type Form struct {
	// RefreshOnChangeTo indicates that changes to the listed fields must reload
	// the form. Values of the fields with values that are not included in the
	// refreshed form are lost. Values that no longer apply are reset.
	RefreshOnChangeTo []string `json:"refresh_on_change_to,omitempty"`
	RefreshURL        string

	Elements []interface{} // of *XXXElement
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

type ElementProps struct {
	Type ElementType `json:"type"`

	// Name is the name of the JSON field to use when submitting
	Name string `json:"name"`

	// Description is long description, can be used in modals in addition to
	// Label, and in Autocomplete as Help
	Description string `json:"description,omitempty"`

	IsRequired        bool     `json:"is_required,omitempty"`
	RefreshOnChangeTo []string `json:"refresh_on_change_to,omitempty"`
}

type StaticSelectElementProps struct {
	Options []SelectOption `json:"options,omitempty"`
}

type DynamicSelectElementProps struct {
	// RefreshOnChangeTo indicates that changes to the listed fields must reload
	// the list (and reset the current value if the old one is not available).
	RefreshURL string `json:"refresh_url,omitempty"`
}

type TextElementProps struct {
	Subtype   string `json:"subtype,omitempty"`
	MinLength int    `json:"min_length,omitempty"`
	MaxLength int    `json:"max_length,omitempty"`
	// options - encoding, regexp, etc.
}

type SelectOption struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	IconData string `json:"icon_data"`
}

// type unmarshalledElement struct {
// 	elementProps

// 	autocompleteElementProps
// 	modalElementProps

// 	staticSelectElementProps
// 	dynamicSelectElementProps
// 	textElementProps
// }
