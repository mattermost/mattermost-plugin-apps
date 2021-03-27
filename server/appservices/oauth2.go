package appservices

import (
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/pkg/errors"
)

func (a *AppServices) CreateOAuth2State(actingUserID string) (string, error) {
	if err := a.ensureFromUser(actingUserID); err != nil {
		return "", err
	}
	return a.store.OAuth2.CreateState(actingUserID)
}

func (a *AppServices) ValidateOAuth2State(actingUserID string, urlState string) error {
	if err := a.ensureFromUser(actingUserID); err != nil {
		return err
	}

	urlUserId := strings.Split(urlState, "_")[1]
	if urlUserId != actingUserID {
		return utils.ErrForbidden
	}

	storedState, err := a.store.OAuth2.GetStateOnce(urlState)
	if err != nil {
		return err
	}
	if storedState != urlState {
		return errors.Wrap(utils.ErrForbidden, "state mismatch")
	}

	return nil
}

func (a *AppServices) StoreRemoteOAuth2App(botUserID string, oapp apps.OAuth2App) error {
	if err := a.ensureFromBot(botUserID); err != nil {
		return err
	}

	app, err := a.findBot(botUserID)
	if err != nil {
		return err
	}
	app.RemoteOAuth2 = oapp
	err = a.store.App.Save(app)
	if err != nil {
		return err
	}

	return nil
}

func (a *AppServices) StoreRemoteOAuth2User(appID apps.AppID, actingUserID string, ref interface{}) error {
	if err := a.ensureFromUser(actingUserID); err != nil {
		return err
	}
	return a.store.OAuth2.SaveRemoteUser(actingUserID, string(appID), ref)
}

func (a *AppServices) GetRemoteOAuth2User(appID apps.AppID, actingUserID string, ref interface{}) error {
	if err := a.ensureFromUser(actingUserID); err != nil {
		return err
	}
	return a.store.OAuth2.GetRemoteUser(actingUserID, string(appID), ref)
}

func (a *AppServices) findBot(botUserID string) (*apps.App, error) {
	for _, app := range a.store.App.AsMap() {
		if app.BotUserID == botUserID {
			return app, nil
		}
	}
	return nil, utils.ErrNotFound
}
