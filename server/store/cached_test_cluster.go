// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type testDataType struct {
	Key                string
	UpdatedByRequestID string
}

func (t testDataType) Clone() *testDataType {
	return &t
}

type CachedStoreTestParams struct {
	Kind               CachedStoreClusterKind
	Name               string
	NumberOfPuts       int
	WaitForIndexToSync time.Duration
	ActingUserID       string
	RootPostID         string
}

type CachedStoreTestReport struct {
	Errors       []string
	HostNickname string
	IndexSHA     string
	Params       CachedStoreTestParams
	Index        *StoredIndex[testDataType]
}

func (s *Service) RunCachedStoreTest(r *incoming.Request, params CachedStoreTestParams) {
	mutex, err := cluster.NewMutex(r.API.Plugin, "cluster_cached_store_test_mutex")
	if err != nil {
		r.Log.WithError(err).Errorf("Failed to create cluster mutex")
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	if s.getTestReportChannel() != nil {
		r.Log.Errorf("Another cluster test is already running on this host (unreachable)")
	}
	testReportChan := make(chan CachedStoreTestReport)
	s.setTestReportChannel(testReportChan)
	defer func() { s.setTestReportChannel(nil) }()

	params.ActingUserID = r.ActingUserID()
	params.RootPostID = dm(r, params, "Starting a cluster test run from host %s. Parameters:\n%s", utils.HostNickname(), utils.JSONBlock(params))

	// Initialize the stores on all hosts.
	store, report := s.initLocalCachedStoreTest(r, params)
	if store == nil || len(report.Errors) > 0 {
		dm(r, params, "(Local) host %s failed to initialize: %v", report.HostNickname, report.Errors)
		return
	}
	defer store.Stop()

	nOtherHosts := 0
	s.broadcastTestEvent(r, testInitEventID, params, params)

	acks, _ := s.collectReports(r, params, 1*time.Second, -1)
	responded := []string{}
	for _, report := range acks {
		if len(report.Errors) > 0 {
			dm(r, params, "Host %s failed to initialize: %v", report.HostNickname, report.Errors)
			return
		}
		responded = append(responded, report.HostNickname)
	}
	nOtherHosts = len(responded)
	dm(r, params, "Found %v other hosts in the cluster (%v)", nOtherHosts, responded)

	// Run the test locally.
	go func() {
		s.getTestReportChannel() <- s.runLocalCachedStoreTest(r, params, store)
	}()

	// Run the test on other hosts, and receive all reports including the local
	// one.
	timeout := 10 * time.Second
	s.broadcastTestEvent(r, testRunEventID, params, params)
	reports, timedOut := s.collectReports(r, params, timeout, 1+nOtherHosts)
	if !timedOut {
		dm(r, params, "Finished with %v reports", len(reports))
	} else {
		dm(r, params, "Timed out after %v, received %v reports", timeout, len(reports))
	}

	// Check the test reports
	sha := ""
	for _, report := range reports {
		switch {
		case sha == "":
			sha = report.IndexSHA
			dm(r, params, "Test report from `%s`: ok: set SHA to `%s`: ok", report.HostNickname, utils.FirstN(report.IndexSHA, 10))
		case sha == report.IndexSHA:
			dm(r, params, "Test report from `%s`: `%s`: ok: matched\n", report.HostNickname, utils.FirstN(report.IndexSHA, 10))
		default:
			dm(r, params, "Test report from `%s`: `%s`: FAIL: expected `%s`", report.HostNickname, utils.FirstN(report.IndexSHA, 10), utils.FirstN(sha, 10))
			return
		}
	}

	// Verify against the stored index, again.
	if params.Kind == TestCachedStoreKind {
		dm(r, params, "Skipping final verification sine KV was never updated.")
	} else {
		dm(r, params, "Starting final verification against the KV")
		store.Stop()
		time.Sleep(500 * time.Millisecond)
		store, err = makeCachedStore[testDataType](params.Kind, params.Name, s.cluster, r.Log)
		if err != nil {
			dm(r, params, "FAIL: final verification from %s", params.Name)
			return
		}
		if sha != store.Index().Stored().hash() {
			dm(r, params, "FAIL: final verification from %s: expected %s, got %s", params.Name, sha, store.Index().Stored())
			return
		}
		dm(r, params, "Final verification: OK")
	}
}

func (s *Service) processTestPluginClusterEvent(r *incoming.Request, ev model.PluginClusterEvent) bool {
	switch ev.Id {
	case testInitEventID, testRunEventID:
		params := CachedStoreTestParams{}
		if err := json.Unmarshal(ev.Data, &params); err != nil {
			r.Log.WithError(err).Errorw("failed to unmarshal test params")
			return true
		}

		go func() {
			var report CachedStoreTestReport
			if ev.Id == testInitEventID {
				// dm(r, params, "received cluster message INIT TEST %s", params.Name)
				store, initReport := s.initLocalCachedStoreTest(r, params)
				s.setTestStore(store)
				report = initReport
			} else { // run
				// dm(r, params, "received cluster message RUN TEST %s", params.Name)
				report = s.runLocalCachedStoreTest(r, params, s.getTestStore())
			}
			s.broadcastTestEvent(r, testReportEventID, report.Params, report)
		}()
		return true

	case testReportEventID:
		report := CachedStoreTestReport{}
		if err := json.Unmarshal(ev.Data, &report); err != nil {
			r.Log.WithError(err).Errorf("failed to unmarshal test report: %v", err)
			return true
		}
		if ch := s.getTestReportChannel(); ch != nil {
			// dm(r, report.Params, "received cluster message TEST REPORT %s:\n%s", report.Params.Name, utils.JSONBlock(report))
			ch <- report
		}
		return true

	default:
		return false
	}
}

func (s *Service) initLocalCachedStoreTest(r *incoming.Request, params CachedStoreTestParams) (store CachedStore[testDataType], report CachedStoreTestReport) {
	var allErrors []string
	hostNickname := utils.HostNickname()
	defer func() {
		report.HostNickname = hostNickname
		report.Params = params
		report.Errors = allErrors
	}()
	dm(r, params, "preparing a local test run on host %s.", hostNickname)

	store, err := makeCachedStore[testDataType](params.Kind, params.Name, s.cluster, r.Log)
	if err != nil {
		allErrors = append(allErrors, err.Error())
		return nil, report
	}
	return store, CachedStoreTestReport{
		IndexSHA: store.Index().Stored().hash(),
	}
}

func (s *Service) runLocalCachedStoreTest(r *incoming.Request, params CachedStoreTestParams, store CachedStore[testDataType]) (report CachedStoreTestReport) {
	var allErrors []string
	hostNickname := utils.HostNickname()
	defer func() {
		report.HostNickname = hostNickname
		report.Errors = allErrors
		report.Params = params
		dm(r, params, "Done running local test.")
	}()

	for i := 0; i < params.NumberOfPuts; i++ {
		cloneR := r.Clone()
		key := fmt.Sprintf("%s-test-%d", hostNickname, i)
		value := testDataType{
			Key:                key,
			UpdatedByRequestID: cloneR.RequestID,
		}
		if err := store.Put(r.Clone(), key, &value); err != nil {
			allErrors = append(allErrors, err.Error())
		}
	}

	stored := store.Index().Stored()
	dm(r, params, "Done with %v puts, index has %v items, waiting %v for index to sync...", params.NumberOfPuts, len(stored.Data), params.WaitForIndexToSync)
	time.Sleep(params.WaitForIndexToSync)
	stored = store.Index().Stored()
	dm(r, params, "Done waiting, index has %v items", len(stored.Data))

	store.Stop()
	return CachedStoreTestReport{
		IndexSHA: stored.hash(),
		Index:    stored,
	}
}

func dm(r *incoming.Request, params CachedStoreTestParams, message string, args ...interface{}) string {
	r.Log.Debugf(message, args...)
	post := &model.Post{
		Message: fmt.Sprintf(utils.HostNickname()+": "+message, args...),
		RootId:  params.RootPostID,
	}
	if err := r.API.Mattermost.Post.DM(
		r.Config.Get().BotUserID,
		params.ActingUserID,
		post,
	); err != nil {
		return ""
	}
	return post.Id
}

func (s *Service) collectReports(r *incoming.Request, params CachedStoreTestParams, timeout time.Duration, n int) (out []CachedStoreTestReport, timedOut bool) {
	timer := time.NewTimer(timeout)
	for {
		select {
		case report := <-s.getTestReportChannel():
			out = append(out, report)
			if n >= 0 && len(out) >= n {
				return out, false
			}

		case <-timer.C:
			return out, true
		}
	}
}

func (s *Service) setTestReportChannel(ch chan CachedStoreTestReport) {
	s.testDataMutex.Lock()
	defer s.testDataMutex.Unlock()
	s.testReportChan = ch
}

func (s *Service) getTestReportChannel() chan CachedStoreTestReport {
	s.testDataMutex.RLock()
	defer s.testDataMutex.RUnlock()
	return s.testReportChan
}

func (s *Service) setTestStore(store CachedStore[testDataType]) {
	s.testDataMutex.Lock()
	defer s.testDataMutex.Unlock()
	s.testStore = store
}

func (s *Service) getTestStore() CachedStore[testDataType] {
	s.testDataMutex.RLock()
	defer s.testDataMutex.RUnlock()
	return s.testStore
}

func (s *Service) broadcastTestEvent(r *incoming.Request, id string, params CachedStoreTestParams, data any) {
	runRemoteTests := params.Kind == SingleWriterCachedStoreKind || params.Kind == MutexCachedStoreKind
	if runRemoteTests {
		s.cluster.broadcastEvent(r, id, data)
	}
}
