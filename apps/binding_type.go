package apps

type BindingType string

const (
	BindingTypeView     BindingType = "view"
	BindingTypeSection  BindingType = "section"
	BindingTypeList     BindingType = "list"
	BindingTypeListItem BindingType = "list_block"
	BindingTypeMenu     BindingType = "menu"
	BindingTypeMenuItem BindingType = "menu_item"
	BindingTypeMarkdown BindingType = "markdown"
	BindingTypeDivider  BindingType = "divider"
)

func (t BindingType) IsValid() bool {
	switch t {
	case BindingTypeView, BindingTypeList, BindingTypeListItem, BindingTypeMarkdown, BindingTypeDivider:
		return true
	default:
		return false
	}
}
