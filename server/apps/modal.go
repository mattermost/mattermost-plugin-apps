package apps

type Modal struct {
	Name    string
	Title   string
	Header  string
	Footer  string
	IconURL string

	Form
}

type modalElementProps struct {
	Label string // Autocomplete:
}

type modalProps struct {
	elementProps
	modalElementProps
}

type ModalTextElement struct {
	modalProps
	textElementProps
}

type ModalStaticSelectElement struct {
	modalProps
	staticSelectElementProps
}

type ModalDynamicSelectElement struct {
	modalProps
	dynamicSelectElementProps
}

type ModalBoolElement modalProps
type ModalUserElement modalProps
type ModalChannelElement modalProps

func (s *Service) CallModal(call *Call) (*CallResponse, error) {
	return nil, nil
}
