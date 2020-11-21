// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

type Store interface {
	DeleteSub(*Subscription) error
	GetSubs(subject Subject, teamID, channelID string) ([]*Subscription, error)
	StoreSub(sub *Subscription) error
}
