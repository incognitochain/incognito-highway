package chain

import (
	"encoding/json"
	"highway/common"
	"sync"
	"time"
)

type Reporter struct {
	name    string
	manager *Manager

	requestCounts struct {
		m map[string]int
		sync.RWMutex
	}
}

func (r *Reporter) Start(_ time.Duration) {
	clearRequestTimestep := 5 * time.Minute
	for ; true; <-time.Tick(clearRequestTimestep) {
		r.clearRequestCounts()
	}
}

func (r *Reporter) ReportJSON() (string, json.Marshaler, error) {
	validators := r.manager.GetAllPeers()
	totalConns := r.manager.GetTotalConnections()
	requests := map[string]int{}

	// Make a copy of request stats
	r.requestCounts.RLock()
	for key, val := range r.requestCounts.m {
		requests[key] = val
	}
	r.requestCounts.RUnlock()

	data := map[string]interface{}{
		"peers":               validators,
		"inbound_connections": totalConns,
		"requests":            r.requestCounts.m,
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

func (r *Reporter) watchRequestCounts(key string) {
	r.requestCounts.Lock()
	defer r.requestCounts.Unlock()
	r.requestCounts.m[key] += 1
}

func (r *Reporter) clearRequestCounts() {
	r.requestCounts.Lock()
	defer r.requestCounts.Unlock()
	for key := range r.requestCounts.m {
		r.requestCounts.m[key] = 0
	}
}
