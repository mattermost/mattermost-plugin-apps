// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-api/cluster"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Service interface {
	// Subscriptions

	Subscribe(*incoming.Request, apps.Subscription) error
	GetSubscriptions(*incoming.Request) ([]apps.Subscription, error)
	Unsubscribe(*incoming.Request, apps.Event) error
	UnsubscribeApp(*incoming.Request, apps.AppID) error

	// Timer
	CreateTimer(*incoming.Request, apps.Timer) error

	// KV

	KVSet(_ *incoming.Request, prefix, id string, data []byte) (bool, error)
	KVGet(_ *incoming.Request, prefix, id string) ([]byte, error)
	KVDelete(_ *incoming.Request, prefix, id string) error
	KVList(_ *incoming.Request, namespace string, processf func(key string) error) error
	KVDebugInfo(*incoming.Request) (*store.KVDebugInfo, error)
	KVDebugAppInfo(*incoming.Request, apps.AppID) (*store.KVDebugAppInfo, error)

	// Remote (3rd party) OAuth2

	StoreOAuth2App(_ *incoming.Request, data []byte) error
	StoreOAuth2User(_ *incoming.Request, data []byte) error
	GetOAuth2User(_ *incoming.Request) ([]byte, error)
}

type Caller interface {
	InvokeCall(*incoming.Request, apps.CallRequest) (*apps.App, apps.CallResponse)
	NewIncomingRequest() *incoming.Request
}

type AppServices struct {
	store     *store.Service
	scheduler *cluster.JobOnceScheduler
	caller    Caller

	conf config.Service
	log  utils.Logger
}

var _ Service = (*AppServices)(nil)

// SetCaller must be called before calling any other methods of AppsServies.
// TODO: Remove this uggly hack.
func (a *AppServices) SetCaller(caller Caller) {
	a.caller = caller
}

func NewService(log utils.Logger, confService config.Service, store *store.Service, scheduler *cluster.JobOnceScheduler) (*AppServices, error) {
	service := &AppServices{
		store:     store,
		scheduler: scheduler,
		conf:      confService,
		log:       log,
	}

	err := scheduler.SetCallback(service.ExecuteTimer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set timer callback")
	}

	err = scheduler.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start timer scheduler")
	}

	return service, nil
}
