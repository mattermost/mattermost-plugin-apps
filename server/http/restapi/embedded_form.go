package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const (
	embeddedSubmissionPostIDKey = "mm_postID"
	embeddedSubmissionAppIDKey  = "mm_appID"
)

func (a *api) handleEmbeddedForm(w http.ResponseWriter, req *http.Request, userID string) {
	var dialogRequest model.SubmitDialogRequest
	err := json.NewDecoder(req.Body).Decode(&dialogRequest)
	if err != nil {
		writeDialogError(w, "Could not decode request.")
		return
	}

	if dialogRequest.Submission[embeddedSubmissionPostIDKey] == nil {
		writeDialogError(w, "No Post ID provided.")
		return
	}

	postID, ok := dialogRequest.Submission[embeddedSubmissionPostIDKey].(string)
	if !ok {
		writeDialogError(w, "Post ID not recognized.")
		return
	}

	delete(dialogRequest.Submission, embeddedSubmissionPostIDKey)

	if dialogRequest.Submission[embeddedSubmissionAppIDKey] == nil {
		writeDialogError(w, "No App ID provided")
		return
	}

	appID, ok := dialogRequest.Submission[embeddedSubmissionAppIDKey].(string)
	if !ok {
		writeDialogError(w, "App ID not recognized")
		return
	}

	delete(dialogRequest.Submission, embeddedSubmissionAppIDKey)

	c := &apps.Call{
		Wish: store.NewWish(appID, dialogRequest.URL),
		Request: &apps.CallRequest{
			Context: &apps.Context{
				AppID:        store.AppID(appID),
				ActingUserID: dialogRequest.UserId,
				ChannelID:    dialogRequest.ChannelId,
				TeamID:       dialogRequest.TeamId,
				UserID:       dialogRequest.UserId,
				PostID:       postID,
			},
			Values: apps.FormValues{
				Data: dialogRequest.Submission,
			},
		},
	}

	resp, err := a.apps.Client.PostWish(c)
	if err != nil {
		writeDialogError(w, "Error contacting the app: "+err.Error())
		return
	}

	var dialogResponse model.SubmitDialogResponse

	if resp.Type == apps.ResponseTypeError {
		if resp.Data["errors"] != nil {
			dialogResponse.Errors = make(map[string]string)
			if errors, ok := resp.Data["errors"].(map[string]interface{}); ok {

				for key, value := range errors {
					if svalue, ok := value.(string); ok {
						dialogResponse.Errors[key] = svalue
					}
				}
			}
		}
		dialogResponse.Error = resp.Error
	} else {
		if resp.Data["post"] != nil {
			if updatePost, parseErr := postFromInterface(resp.Data["post"]); parseErr == nil {
				updateErr := a.UpdatePost(postID, updatePost)
				if updateErr != nil {
					a.mm.Log.Debug("could not update post", "error", updateErr)
				}
			} else {
				a.mm.Log.Debug("could not transform post", "error", parseErr)
			}
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

func (a *api) UpdatePost(postID string, post *model.Post) error {
	// If the updated post does contain a replacement Props set, we still
	// need to preserve some original values, as listed in
	// model.PostActionRetainPropKeys. remove and retain track these.
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

	a.mm.Post.UpdatePost(post)

	return nil
}
