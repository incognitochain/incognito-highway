package route

import (
	"encoding/json"
	"highway/common"
	"highway/grafana"
	hmap "highway/route/hmap"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

type Reporter struct {
	name string

	gralog  *grafana.GrafanaLog
	manager *Manager
}

func (r *Reporter) Start(_ time.Duration) {
	go r.pushDataToGrafana()
}

func (r *Reporter) ReportJSON() (string, json.Marshaler, error) {
	// PID of this highway
	peerID := r.manager.ID.String()

	// List of shards connected by this highway (directly and indirectly)
	connected := r.manager.GetShardsConnected()
	shardsConnected := common.BytesToInts(connected)

	// List of all connected highway, their full addrinfo and supported shards
	peers := r.manager.Hmap.CopyPeersMap()
	urls := r.manager.Hmap.CopyRPCUrls()
	supports := r.manager.Hmap.CopySupports()
	status := r.manager.Hmap.CopyStatus()
	highwayConnected := map[string]highwayInfo{}
	for pid, cids := range supports {
		// Find addrInfo from pid
		var addrInfo peer.AddrInfo
		for _, addrs := range peers {
			for _, addr := range addrs {
				if addr.ID == pid {
					addrInfo = addr
				}
			}
		}

		s := status[pid]
		highwayConnected[pid.String()] = highwayInfo{
			AddrInfo: addrInfo,
			Supports: common.BytesToInts(cids),
			RPCUrl:   urls[pid],
			Status:   s,
		}
	}

	data := map[string]interface{}{
		"peer_id":           peerID,
		"shards_connected":  shardsConnected,
		"highway_connected": highwayConnected,
	}
	marshaler := common.NewDefaultMarshaler(data)
	return r.name, marshaler, nil
}

func (r *Reporter) pushDataToGrafana() {
	if r.gralog == nil {
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		numConn := 0
		supports := r.manager.Hmap.CopySupports()
		status := r.manager.Hmap.CopyStatus()
		for pid, _ := range supports {
			s := status[pid]
			if s.Connecting {
				numConn++
			}
		}

		key := "highway_connected"
		value := numConn
		r.gralog.Add(key, value)
	}

}

func NewReporter(manager *Manager, gralog *grafana.GrafanaLog) *Reporter {
	return &Reporter{
		manager: manager,
		gralog:  gralog,
		name:    "route",
	}
}

type highwayInfo struct {
	AddrInfo peer.AddrInfo `json:"addr_info"`
	Supports []int         `json:"shards_support"`
	RPCUrl   string        `json:"rpc_url"`
	Status   hmap.Status   `json:"status"`
}
