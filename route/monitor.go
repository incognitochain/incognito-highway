package route

import (
	"encoding/json"
	"highway/common"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

type Reporter struct {
	name string

	manager *Manager
}

func (r *Reporter) Start(_ time.Duration) {}

func (r *Reporter) ReportJSON() (string, json.Marshaler, error) {
	// PID of this highway
	peerID := r.manager.ID.String()

	// List of shards connected by this highway (directly and indirectly)
	connected := r.manager.GetShardsConnected()
	shardsConnected := common.BytesToInts(connected)

	// List of all connected highway, their full addrinfo and supported shards
	peers := r.manager.hmap.CopyPeersMap()
	supports := r.manager.hmap.CopySupports()
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

		highwayConnected[pid.String()] = highwayInfo{
			AddrInfo: addrInfo,
			Supports: common.BytesToInts(cids),
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

func NewReporter(manager *Manager) *Reporter {
	return &Reporter{
		manager: manager,
		name:    "route",
	}
}

type highwayInfo struct {
	AddrInfo peer.AddrInfo `json:"addr_info"`
	Supports []int         `json:"shards_support"`
}
