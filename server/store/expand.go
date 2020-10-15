// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

type ExpandLevel string

const (
	ExpandAll     = ExpandLevel("All")
	ExpandSummary = ExpandLevel("Summary")
)

type Expand struct {
	App        ExpandLevel `json:"app"`
	ActingUser ExpandLevel `json:"acting_user"`
	Channel    ExpandLevel `json:"channel,omitempty"`
	Config     ExpandLevel `json:"config,omitempty"`
	Mentioned  ExpandLevel `json:"mentioned,omitempty"`
	ParentPost ExpandLevel `json:"parent_post,omitempty"`
	Post       ExpandLevel `json:"post,omitempty"`
	RootPost   ExpandLevel `json:"root_post,omitempty"`
	Team       ExpandLevel `json:"team,omitempty"`
	User       ExpandLevel `json:"user,omitempty"`
}
