package chain

import (
	"encoding/json"
	"highway/common"
	"sync"
	"time"

	peer "github.com/libp2p/go-libp2p-core/peer"
)

type Reporter struct {
	name    string
	manager *Manager

	requestCounts struct {
		m map[string]int
		sync.RWMutex
	}

	requestsPerPeer struct {
		m map[string]map[string]int
		sync.RWMutex
	}
}

func (r *Reporter) Start(_ time.Duration) {
	clearRequestTimestep := 5 * time.Minute
	for ; true; <-time.Tick(clearRequestTimestep) {
		r.clearRequestCounts()
		r.clearRequestsPerPeer()
	}
}

func (r *Reporter) ReportJSON() (string, json.Marshaler, error) {
	validators := r.manager.GetAllPeers()
	totalConns := r.manager.GetTotalConnections()

	// Make a copy of request stats
	requests := map[string]int{}
	r.requestCounts.RLock()
	for key, val := range r.requestCounts.m {
		requests[key] = val
	}
	r.requestCounts.RUnlock()

	// Make a copy of request per peer stats
	requestsPerPeer := map[string]map[string]int{}
	r.requestsPerPeer.RLock()
	for msg, peers := range r.requestsPerPeer.m {
		for pid, cnt := range peers {
			requestsPerPeer[msg][pid] = cnt
		}
	}
	r.requestsPerPeer.RUnlock()

	data := map[string]interface{}{
		"peers":               validators,
		"inbound_connections": totalConns,
		"requests":            r.requestCounts.m,
		"request_per_peer":    requestsPerPeer,
	}
	marshaler := common.NewDefaultMarshaler(data)
	return r.name, marshaler, nil
}

func NewReporter(manager *Manager) *Reporter {
	r := &Reporter{
		manager: manager,
		name:    "chain",
	}
	r.requestCounts.m = map[string]int{}
	r.requestCounts.RWMutex = sync.RWMutex{}
	return r
}

func (r *Reporter) watchRequestCounts(msg string) {
	r.requestCounts.Lock()
	defer r.requestCounts.Unlock()
	r.requestCounts.m[msg] += 1
}

func (r *Reporter) clearRequestCounts() {
	r.requestCounts.Lock()
	defer r.requestCounts.Unlock()
	for key := range r.requestCounts.m {
		r.requestCounts.m[key] = 0
	}
}

func (r *Reporter) watchRequestsPerPeer(msg string, pid peer.ID, err error) {
	r.requestsPerPeer.Lock()
	defer r.requestsPerPeer.Unlock()
	if err != nil { // NOTE: we can monitor failed requests too
		r.requestsPerPeer.m[msg][pid.String()] += 1
	} else {
		logger.Infof("watch is err")
	}
}

func (r *Reporter) clearRequestsPerPeer() {
	r.requestsPerPeer.Lock()
	defer r.requestsPerPeer.Unlock()
	for key := range r.requestsPerPeer.m {
		r.requestsPerPeer.m[key] = map[string]int{}
	}
}
