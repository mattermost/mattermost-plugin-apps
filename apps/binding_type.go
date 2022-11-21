package apps

type BindingType string
type BindingSubtype string

const (
	BindingTypeView      BindingType = "view"
	BindingTypeSection   BindingType = "section"
	BindingTypeListBlock BindingType = "list_block"
	BindingTypeListItem  BindingType = "list_item"
	BindingTypeMenu      BindingType = "menu"
	BindingTypeMenuItem  BindingType = "menu_item"
	BindingTypeMarkdown  BindingType = "markdown"
	BindingTypeDivider   BindingType = "divider"
)

const (
	// select variants
	BindingSubtypeSelectCategories   BindingSubtype = "categories"
	BindingSubtypeSelectButtonSelect BindingSubtype = "button_select"
)

func (t BindingType) IsValid() bool {
	switch t {
	case BindingTypeView, BindingTypeSection, BindingTypeListBlock, BindingTypeListItem, BindingTypeMenu, BindingTypeMenuItem, BindingTypeMarkdown, BindingTypeDivider:
		return true
	default:
		return false
	}
}
