// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

// Provides persistent storage capabilities for the apps framework.
// Specifically, stores apps, manifests, subscriptions, and apps data: KV, user,
// and OAUth2 state data.
//
// [CachedStore] provides a cluster-aware write-through in-memory cache for
// Apps, Manifests, and Subscriptions. There are 2 important implementations. To
// ensure the store consistency [MutexCachedStore] which uses a cluster.Mutex,
// and [Sing;eWriterCachedStore] which performs all updates on the cluster
// leader node.
package store
