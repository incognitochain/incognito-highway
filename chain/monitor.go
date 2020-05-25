package chain

import (
	"encoding/json"
	"highway/common"
	"highway/grafana"
	"sync"
	"time"

	peer "github.com/libp2p/go-libp2p-core/peer"
)

type Reporter struct {
	name    string
	manager *Manager

	requestCounts struct {
		m map[string]uint64
		sync.RWMutex
		lm map[string]uint64
	}

	requestsPerPeer struct {
		m PeerRequestMap
		sync.RWMutex
		lm PeerRequestMap
	}
	gralog *grafana.GrafanaLog
}

func (r *Reporter) Start(_ time.Duration) {
	go r.pushDataToGrafana()
}

func (r *Reporter) ReportJSON() (string, json.Marshaler, error) {
	validators := r.manager.GetAllPeers()
	totalConns := r.manager.GetTotalConnections()

	// // Make a copy of request stats
	// requests := map[string]uint64{}
	// r.requestCounts.RLock()
	// for key, val := range r.requestCounts.m {
	// 	requests[key] = val
	// }
	// r.requestCounts.RUnlock()

	// // Make a copy of request per peer stats
	// requestsPerPeer := PeerRequestMap{}
	// r.requestsPerPeer.RLock()
	// for key, val := range r.requestsPerPeer.m {
	// 	requestsPerPeer[key] = val
	// }
	// r.requestsPerPeer.RUnlock()

	// // Get memcache info
	// cacheInfo := map[string]interface{}{}
	// providers := r.manager.server.Providers
	// if len(providers) > 0 {
	// 	if cache, ok := providers[0].(*MemCache); ok {
	// 		cacheInfo = cache.Metrics()
	// 	}
	// }

	data := map[string]interface{}{
		"peers":               validators,
		"inbound_connections": totalConns,
		// "requests":            requests,
		// "request_per_peer":    requestsPerPeer,
		// "cache":               cacheInfo,
	}
	marshaler := common.NewDefaultMarshaler(data)
	return r.name, marshaler, nil
}

func NewReporter(manager *Manager) *Reporter {
	r := &Reporter{
		manager: manager,
		name:    "chain",
	}
	r.requestCounts.m = map[string]uint64{}
	r.requestCounts.lm = map[string]uint64{}
	r.requestCounts.RWMutex = sync.RWMutex{}
	r.requestsPerPeer.m = PeerRequestMap{}
	r.requestsPerPeer.lm = PeerRequestMap{}
	r.requestsPerPeer.RWMutex = sync.RWMutex{}
	r.gralog = manager.gralog
	return r
}

func (r *Reporter) watchRequestCounts(msg string) {
	r.requestCounts.Lock()
	defer r.requestCounts.Unlock()
	r.requestCounts.m[msg] += 1
}

func (r *Reporter) watchRequestsPerPeer(msg string, pid peer.ID, err error) {
	r.requestsPerPeer.Lock()
	defer r.requestsPerPeer.Unlock()
	key := PeerRequestKey{Msg: msg}
	if err == nil {
		key.PeerID = pid.String()
	} else {
		key.PeerID = "error"
	}
	r.requestsPerPeer.m[key] += 1
}

func (r *Reporter) pushDataToGrafana() {
	if r.gralog == nil {
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		r.requestCounts.Lock()
		for k, v := range r.requestCounts.m {
			c := v - r.requestCounts.lm[k]
			r.requestCounts.lm[k] = v
			r.gralog.Add(k, c)
		}

		// Get memcache info
		cacheInfo := map[string]interface{}{}
		providers := r.manager.server.Providers
		if len(providers) > 0 {
			if cache, ok := providers[0].(*MemCache); ok {
				cacheInfo = cache.Metrics()
			}
		}
		r.requestCounts.Unlock()

		for k, v := range cacheInfo {
			r.gralog.Add(k, v)
		}
	}

}

type PeerRequestKey struct {
	Msg    string
	PeerID string
}

type PeerRequestMap map[PeerRequestKey]int

// MarshalJSON helps flatten PeerRequestKey into a nested map
// for prettier results when json.Marshal
func (m PeerRequestMap) MarshalJSON() ([]byte, error) {
	splat := map[string]map[string]int{}
	for key, val := range m {
		if splat[key.Msg] == nil {
			splat[key.Msg] = map[string]int{}
		}
		splat[key.Msg][key.PeerID] = val
	}
	return json.Marshal(splat)
}
