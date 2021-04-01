package proxy

import (
	"sync"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/mattermost/mattermost-server/v5/utils/markdown"
	"github.com/pkg/errors"
)

func (e *expander) getExplicitMentions(post *model.Post) []*model.User {
	mentions := map[string]*model.User{}

	buf := ""
	mentionsEnabledFields := getMentionsEnabledFields(post)
	for _, message := range mentionsEnabledFields {
		markdown.Inspect(message, func(node interface{}) bool {
			text, ok := node.(*markdown.Text)
			if !ok {
				users := e.processText(buf)
				mentions = join(mentions, users)
				buf = ""
				return true
			}
			buf += text.Text
			return false
		})
	}
	users := e.processText(buf)
	mentions = join(mentions, users)

	return mapToSlice(mentions)
}

// Given a post returns the values of the fields in which mentions are possible.
// post.message, preText and text in the attachment are enabled.
func getMentionsEnabledFields(post *model.Post) model.StringArray {
	ret := []string{}

	ret = append(ret, post.Message)
	for _, attachment := range post.Attachments() {

		if attachment.Pretext != "" {
			ret = append(ret, attachment.Pretext)
		}
		if attachment.Text != "" {
			ret = append(ret, attachment.Text)
		}
	}
	return ret
}

// Processes text to filter mentioned users and other potential mentions
func (e *expander) processText(text string) map[string]*model.User {
	type mentionMapItem struct {
		Name string
		User *model.User
	}

	possibleMentions := model.PossibleAtMentions(text)
	mentionChan := make(chan *mentionMapItem, len(possibleMentions))

	var wg sync.WaitGroup
	for _, mention := range possibleMentions {
		wg.Add(1)
		go func(mention string) {
			defer wg.Done()
			user, nErr := e.mm.User.GetByUsername(mention)

			var nfErr *store.ErrNotFound
			if nErr != nil && !errors.As(nErr, &nfErr) {
				e.mm.Log.Warn("Failed to retrieve user @"+mention, mlog.Err(nErr))
				return
			}

			// If it's a http.StatusNotFound error, check for usernames in substrings
			// without trailing punctuation
			if nErr != nil {
				trimmed, ok := model.TrimUsernameSpecialChar(mention)
				for ; ok; trimmed, ok = model.TrimUsernameSpecialChar(trimmed) {
					userFromTrimmed, nErr := e.mm.User.GetByUsername(trimmed)
					if nErr != nil && !errors.As(nErr, &nfErr) {
						return
					}

					if nErr != nil {
						continue
					}

					mentionChan <- &mentionMapItem{trimmed, userFromTrimmed}
					return
				}

				return
			}

			mentionChan <- &mentionMapItem{mention, user}
		}(mention)
	}

	wg.Wait()
	close(mentionChan)

	atMentionMap := make(map[string]*model.User)
	for mention := range mentionChan {
		atMentionMap[mention.Name] = mention.User
	}
	return atMentionMap
}

func join(a map[string]*model.User, b map[string]*model.User) map[string]*model.User {
	for k, v := range b {
		a[k] = v
	}
	return a
}

func mapToSlice(a map[string]*model.User) []*model.User {
	slice := make([]*model.User, 0, len(a))
	for _, v := range a {
		slice = append(slice, v)
	}
	return slice
}
