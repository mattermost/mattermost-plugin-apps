package apps

type AutocompleteDef struct {
	Form
}

type autocompleteElementProps struct {
	FlagName string `json:"flag_name"`

	// Hint describes what choices may come after selecting this choice
	Hint string `json:"label"`

	RoleID string `json:"role_id"`
}

type autocompleteProps struct {
	elementProps
	autocompleteElementProps
}

type AutocompleteText struct {
	autocompleteProps
	textElementProps
}

type AutocompleteStaticSelect struct {
	autocompleteProps
	staticSelectElementProps
}

type AutocompleteDynamicSelect struct {
	autocompleteProps
	dynamicSelectElementProps
}

type AutocompleteBool autocompleteProps
type AutocompleteUser autocompleteProps
type AutocompleteChannel autocompleteProps
