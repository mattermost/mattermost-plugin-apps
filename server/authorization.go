package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-api/experimental/common"
	"github.com/mattermost/mattermost-server/v5/model"
)

const ContextState = "state"
const StateKey = "s_"
const AppKey = "app_"

type Manifest struct {
	Name        string
	Description string
	URL         string
	Permissions string
}

func (m *Manifest) ParsePermissions() string {
	return m.Permissions
}

func (p *Plugin) InstallApp(url, adminID string) error {
	manifestURL := url + "/manifest"
	r, err := http.Get(manifestURL)
	if err != nil {
		return err
	}

	var manifest Manifest
	if r.StatusCode != 200 {
		return errors.New("cannot fetch manifest")
	}

	err = json.NewDecoder(r.Body).Decode(&manifest)
	if err != nil {
		return err
	}

	err = p.AskAuthorization(manifest, adminID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) AskAuthorization(manifest Manifest, adminID string) error {
	id := model.NewId()
	state := fmt.Sprintf("%s_%s", manifest.Name, id)
	key := fmt.Sprintf("%s_%s", StateKey, state)
	marshaled, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	appErr := p.API.KVSet(key, marshaled)
	if appErr != nil {
		return appErr
	}

	err = p.DM(p.getPost(manifest, state), adminID)
	if err != nil {
		return err
	}
	return nil
}

func (p *Plugin) DM(post *model.Post, userID string) error {
	//TODO
	return nil
}

func (p *Plugin) getPost(manifest Manifest, state string) *model.Post {
	actionTrue := model.PostAction{
		Name: "Authorize",
		Integration: &model.PostActionIntegration{
			URL: "authorize",
			Context: map[string]interface{}{
				ContextState: state,
			},
		},
	}

	actionFalse := model.PostAction{
		Name: "Not Authorize",
		Integration: &model.PostActionIntegration{
			URL: "notAuthorize",
			Context: map[string]interface{}{
				ContextState: state,
			},
		},
	}

	title := fmt.Sprintf("Authorize %s", manifest.Name)
	text := fmt.Sprintf("Description:\n%s\n\nPermissions:\n%s", manifest.Description, manifest.ParsePermissions())

	sa := model.SlackAttachment{
		Title:    title,
		Text:     text,
		Fallback: fmt.Sprintf("%s: %s", title, text),
		Actions:  []*model.PostAction{&actionTrue, &actionFalse},
	}

	post := &model.Post{}

	model.ParseSlackAttachment(post, []*model.SlackAttachment{&sa})

	return post
}

func (p *Plugin) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		common.SlackAttachmentError(w, "Error: Not authorized")
		return
	}

	request := model.PostActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		common.SlackAttachmentError(w, "Error: invalid request")
		return
	}

	state, ok := request.Context[ContextState].(string)
	if !ok {
		common.SlackAttachmentError(w, "Error: missing state")
		return
	}

	rawManifest, appErr := p.API.KVGet(fmt.Sprintf("%s_%s", StateKey, state))
	if appErr != nil {
		common.SlackAttachmentError(w, "Error: error getting manifest")
		return
	}

	var manifest Manifest
	err := json.Unmarshal(rawManifest, &manifest)
	if err != nil {
		common.SlackAttachmentError(w, "Error: error parsing manifest")
		return
	}

	err = p.AuthorizeApp(manifest)
	if err != nil {
		common.SlackAttachmentError(w, "Error: cannot authorize the app")
		return
	}

	response := model.PostActionIntegrationResponse{}
	post := model.Post{}
	model.ParseSlackAttachment(&post, []*model.SlackAttachment{p.getResponseSlackAttachments(true)})
	response.Update = &post

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(response.ToJson())
}

func (p *Plugin) handleNotauthorize(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		common.SlackAttachmentError(w, "Error: Not authorized")
		return
	}

	request := model.PostActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		common.SlackAttachmentError(w, "Error: invalid request")
		return
	}

	state, ok := request.Context[ContextState].(string)
	if !ok {
		common.SlackAttachmentError(w, "Error: missing state")
		return
	}

	p.API.KVDelete(fmt.Sprintf("%s_%s", StateKey, state))

	response := model.PostActionIntegrationResponse{}
	post := model.Post{}
	model.ParseSlackAttachment(&post, []*model.SlackAttachment{p.getResponseSlackAttachments(false)})
	response.Update = &post

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(response.ToJson())
}

func (p *Plugin) getResponseSlackAttachments(authorized bool) *model.SlackAttachment {
	//TODO
	return &model.SlackAttachment{}
}

type SecretRequest struct {
	secret string
}

type SecretResponse struct {
	secret string
}

type Secret struct {
	local    string
	app      string
	manifest Manifest
}

func (p *Plugin) AuthorizeApp(manifest Manifest) error {
	secret, err := p.CreateServiceAccount(manifest)
	if err != nil {
		return err
	}
	body, err := json.Marshal(SecretRequest{secret: secret})
	if err != nil {
		return err
	}
	r, err := http.Post(manifest.URL+"/authorization", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	if r.StatusCode != 200 {
		return errors.New("Could not authorize")
	}

	var appSecret SecretResponse
	json.NewDecoder(r.Body).Decode(&appSecret)

	secretObject := Secret{
		local:    secret,
		app:      appSecret.secret,
		manifest: manifest,
	}

	rawSecret, err := json.Marshal(secretObject)
	if err != nil {
		return err
	}

	p.API.KVSet(AppKey+manifest.Name, rawSecret)
	return nil
}

func (p *Plugin) CreateServiceAccount(manifest Manifest) (string, error) {
	//TODO
	secret := model.NewId()
	return secret, nil
}

type GetTokenResponse struct {
	Token string
}

func (p *Plugin) handleGetToken(w http.ResponseWriter, r *http.Request) {
	appName, userID := p.verifyJWT(r)
	if appName == "" || userID == "" {
		http.Error(w, "Could not verify JWT", http.StatusForbidden)
		return
	}

	rawSecret, appErr := p.API.KVGet(AppKey + appName)
	if appErr != nil {
		http.Error(w, "Error fetching app", http.StatusInternalServerError)
		return
	}

	var secret Secret
	err := json.Unmarshal(rawSecret, &secret)
	if err != nil {
		http.Error(w, "Error decoding app", http.StatusInternalServerError)
		return
	}

	token := p.createToken(userID, secret.manifest.Permissions)
	rawToken, err := json.Marshal(GetTokenResponse{Token: token})
	if err != nil {
		http.Error(w, "Error marshalling token", http.StatusInternalServerError)
		return
	}

	w.Write(rawToken)
}

func (p *Plugin) verifyJWT(r *http.Request) (string, string) {
	// TODO
	return "appName", "userID"
}

func (p *Plugin) createToken(userID, permissions string) string {
	//TODO
	return "token"
}
