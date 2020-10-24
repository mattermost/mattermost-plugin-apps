package apps

type Modal struct {
	Name    string
	Title   string
	Header  string
	Footer  string
	IconURL string

	Form
}

type ModalElementProps struct {
	Label string // Autocomplete:
}

type ModalProps struct {
	ElementProps
	ModalElementProps
}

type ModalTextElement struct {
	ModalProps
	TextElementProps
}

type ModalStaticSelectElement struct {
	ModalProps
	StaticSelectElementProps
}

type ModalDynamicSelectElement struct {
	ModalProps
	DynamicSelectElementProps
}

type ModalBoolElement ModalProps
type ModalUserElement ModalProps
type ModalChannelElement ModalProps

func (s *service) CallModal(call *Call) (*CallResponse, error) {
	return nil, nil
}
