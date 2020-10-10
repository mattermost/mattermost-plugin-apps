// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

func (s *store) StoreApp(app *App) error {
	_, err := s.Mattermost.KV.Set(prefixApp+string(app.Manifest.AppID), app)
	return err
}

func (s *store) GetApp(appID AppID) (*App, error) {
	app := App{}
	err := s.Mattermost.KV.Get(prefixApp+string(appID), &app)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (s *store) DeleteApp(appID AppID) error {
	return s.Mattermost.KV.Delete(prefixApp + string(appID))
}
