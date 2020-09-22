package apps

type Wish struct {
	URL string
}

func (s *service) CallWish(appID AppID, w *WishManifest, ctx callContext) (*WishResponse) error {
	app, err := s.Registry.GetApp(appID)
	if err != nil {

	}
	s.Proxy.Post()
	return nil
}
