package apps

type Autocomplete struct {
	Form
}

type AutocompleteElementProps struct {
	FlagName string `json:"flag_name"`

	// Hint describes what choices may come after selecting this choice
	Hint string `json:"hint"`

	RoleID     string `json:"role_id"`
	Positional bool   `json:"positional"`
}

type AutocompleteProps struct {
	*ElementProps
	*AutocompleteElementProps
}

type AutocompleteText struct {
	AutocompleteProps
	TextElementProps
}

type AutocompleteStaticSelect struct {
	*AutocompleteProps
	*StaticSelectElementProps
}

type AutocompleteDynamicSelect struct {
	AutocompleteProps
	DynamicSelectElementProps
}

type AutocompleteBool AutocompleteProps
type AutocompleteUser AutocompleteProps
type AutocompleteChannel AutocompleteProps
