package apps

type BindingType string

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

func (t BindingType) IsValid() bool {
	switch t {
	case BindingTypeView, BindingTypeSection, BindingTypeListBlock, BindingTypeListItem, BindingTypeMenu, BindingTypeMenuItem, BindingTypeMarkdown, BindingTypeDivider:
		return true
	default:
		return false
	}
}
