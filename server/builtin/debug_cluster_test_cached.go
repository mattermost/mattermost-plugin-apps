// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"strconv"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-server/v6/model"
)

func (a *builtinApp) debugClusterTestCachedCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "test-cached",
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.cluster.test_cached.label",
			Other: "test-cached",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.cluster.test_cached.description",
			Other: "Runs a test of cluster aware cached store.",
		}),
		Hint: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.cluster.test_cached.hint",
			Other: "[ options... ]",
		}),
		Form: &apps.Form{
			Submit: newUserCall(pDebugClusterTestCached),
			Fields: []apps.Field{
				{
					Name: fNumberPuts,
					Type: apps.FieldTypeText,
					ModalLabel: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.cluster.num_puts.modal_label",
						Other: "Number of PUTs",
					}),
					Value: "10",
				},
				{
					Name: fWaitToSync,
					Type: apps.FieldTypeText,
					ModalLabel: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.cluster.wait_to_sync.modal_label",
						Other: "Wait before checking the index, in seconds",
					}),
					Value: "2",
				},
				{
					Name: fStoreKind,
					Type: apps.FieldTypeStaticSelect,
					SelectStaticOptions: []apps.SelectOption{
						{
							Label: string(store.SimpleCachedStoreKind),
							Value: string(store.SimpleCachedStoreKind),
						},
						{
							Label: string(store.SingleWriterCachedStoreKind),
							Value: string(store.SingleWriterCachedStoreKind),
						},
						{
							Label: string(store.MutexCachedStoreKind),
							Value: string(store.MutexCachedStoreKind),
						},
						{
							Label: string(store.TestCachedStoreKind),
							Value: string(store.TestCachedStoreKind),
						},
					},
					ModalLabel: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.cluster.store_type.modal_label",
						Other: "Cluster replication type",
					}),
					Value: apps.SelectOption{
						Label: string(store.MutexCachedStoreKind),
						Value: string(store.MutexCachedStoreKind),
					},
				},
			},
		},
	}
}

func (a *builtinApp) debugClusterTestCached(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	numPuts, _ := strconv.Atoi(creq.GetValue(fNumberPuts, ""))
	waitSeconds, _ := strconv.Atoi(creq.GetValue(fWaitToSync, ""))
	storeKind := store.CachedStoreClusterKind(creq.GetValue(fStoreKind, ""))
	name := "test/" + model.NewId()

	a.appservices.RunCachedStoreTest(r,
		store.CachedStoreTestParams{
			Kind:               storeKind,
			Name:               name,
			NumberOfPuts:       numPuts,
			WaitForIndexToSync: time.Duration(waitSeconds) * time.Second,
		},
	)

	loc := a.newLocalizer(creq)
	return apps.NewTextResponse(a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
		ID:    "message.cluster.started_test",
		Other: "Started the test. Will report as a direct conversation with @appsbot.",
	}))
}
