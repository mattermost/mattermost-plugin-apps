package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const (
	embeddedSubmissionPostIDKey = "mm_post_id"
	embeddedSubmissionAppIDKey  = "mm_app_id"
)

func (a *restapi) handleEmbeddedForm(w http.ResponseWriter, req *http.Request, userID string) {
	var dialogRequest model.SubmitDialogRequest
	err := json.NewDecoder(req.Body).Decode(&dialogRequest)
	if err != nil {
		writeDialogError(w, "Could not decode request.")
		return
	}

	postID, ok := dialogRequest.Submission[embeddedSubmissionPostIDKey].(string)
	if !ok {
		writeDialogError(w, "Post ID not recognized.")
		return
	}

	delete(dialogRequest.Submission, embeddedSubmissionPostIDKey)

	appID, ok := dialogRequest.Submission[embeddedSubmissionAppIDKey].(string)
	if !ok {
		writeDialogError(w, "App ID not recognized")
		return
	}

	delete(dialogRequest.Submission, embeddedSubmissionAppIDKey)

	values := mapInterfaceToMapString(dialogRequest.Submission)

	c := &api.Call{
		URL: dialogRequest.URL,
		Context: &api.Context{
			AppID:        api.AppID(appID),
			ActingUserID: dialogRequest.UserId,
			ChannelID:    dialogRequest.ChannelId,
			TeamID:       dialogRequest.TeamId,
			UserID:       dialogRequest.UserId,
			PostID:       postID,
		},
		Values: values,
	}

	resp, err := a.apps.Client.PostFunction(c)
	if err != nil {
		writeDialogError(w, "Error contacting the app: "+err.Error())
		return
	}

	var dialogResponse model.SubmitDialogResponse

	if resp.Type == api.CallResponseTypeError {
		if resp.Data[api.EmbeddedResponseDataErrors] != nil {
			dialogResponse.Errors = make(map[string]string)
			if errors, ok := resp.Data[api.EmbeddedResponseDataErrors].(map[string]interface{}); ok {
				for key, value := range errors {
					if svalue, ok := value.(string); ok {
						dialogResponse.Errors[key] = svalue
					}
				}
			}
		}
		dialogResponse.Error = resp.Error
	} else if resp.Data[api.EmbeddedResponseDataPost] != nil {
		if updatedPost, parseErr := postFromInterface(resp.Data[api.EmbeddedResponseDataPost]); parseErr == nil {
			updateErr := a.UpdatePost(postID, updatedPost)
			if updateErr != nil {
				a.mm.Log.Debug("could not update post", "error", updateErr)
			}
		} else {
			a.mm.Log.Debug("could not transform post", "error", parseErr)
		}
	}
	httputils.WriteJSON(w, dialogResponse)
}

func writeDialogError(w http.ResponseWriter, msg string) {
	httputils.WriteJSON(w, model.SubmitDialogResponse{
		Error: msg,
	})
}

func postFromInterface(v interface{}) (*model.Post, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var post model.Post
	err = json.Unmarshal(b, &post)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (a *restapi) UpdatePost(postID string, post *model.Post) error {
	// If the updated post does contain a replacement Props set, we still
	// need to preserve some original values, as listed in
	// model.PostActionRetainPropKeys. remove and retain track these.
	// Copied from: https://github.com/mattermost/mattermost-server/blob/20491c2585475c2218f964e0a882c65deac570a5/app/integration_action.go#L57
	remove := []string{}
	retain := map[string]interface{}{}

	for _, key := range model.PostActionRetainPropKeys {
		value, ok := post.GetProps()[key]
		if ok {
			retain[key] = value
		} else {
			remove = append(remove, key)
		}
	}

	originalPost, err := a.mm.Post.GetPost(postID)
	if err != nil {
		return errors.Wrap(err, "error getting the post")
	}

	post.Id = originalPost.Id
	post.RootId = originalPost.RootId
	if post.GetProps() == nil {
		post.SetProps(originalPost.GetProps())
	} else {
		for key, value := range retain {
			post.AddProp(key, value)
		}
		for _, key := range remove {
			post.DelProp(key)
		}
	}
	post.IsPinned = originalPost.IsPinned
	post.HasReactions = originalPost.HasReactions

	err = a.mm.Post.UpdatePost(post)

	if err != nil {
		return errors.Wrap(err, "error updating the post")
	}

	return nil
}

func mapInterfaceToMapString(in map[string]interface{}) map[string]string {
	out := make(map[string]string)
	for k, v := range in {
		if sv, ok := v.(string); ok {
			out[k] = sv
			continue
		}
		rv, err := json.Marshal(v)
		if err != nil {
			out[k] = ""
			continue
		}
		out[k] = string(rv)
	}

	return out
}
