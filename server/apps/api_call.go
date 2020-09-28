package apps

import (
	"github.com/pkg/errors"
)

func (s *Service) Call(call Call) (*CallResponse, error) {
	switch {
	case call.Wish != nil && call.Modal == nil:
		return s.PostWish(call)
	case call.Modal != nil && call.Wish == nil:
		return s.CallModal(call)
	default:
		return nil, errors.New("invalid Call, only one of Wish, Modal can be specified")
	}
}
