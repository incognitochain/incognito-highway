package hmap

import (
	"highway/common"
	"sync"

	"github.com/libp2p/go-libp2p-core/peer"
)

type Map struct {
	Peers    map[byte][]peer.AddrInfo // shard => peers
	Supports map[peer.ID][]byte       // peerID => shards supported
	RPCs     map[peer.ID]string       // peerID => RPC endpoint (to call GetPeers)

	peerConnected map[peer.ID]bool // keep track of connected peers
	*sync.RWMutex
}

func NewMap(p peer.AddrInfo, supportShards []byte, rpcUrl string) *Map {
	m := &Map{
		Peers:         map[byte][]peer.AddrInfo{},
		Supports:      map[peer.ID][]byte{},
		RPCs:          map[peer.ID]string{},
		peerConnected: map[peer.ID]bool{},
		RWMutex:       &sync.RWMutex{},
	}
	m.AddPeer(p, supportShards, rpcUrl)
	m.ConnectToShardOfPeer(p)
	return m
}

func (h *Map) isConnectedToShard(s byte) bool {
	for pid, conn := range h.peerConnected {
		if conn == false {
			continue
		}

		for _, sup := range h.Supports[pid] {
			if sup == s {
				return true
			}
		}
	}
	return false
}

func (h *Map) IsConnectedToShard(s byte) bool {
	h.RLock()
	defer h.RUnlock()
	return h.isConnectedToShard(s)
}

func (h *Map) IsConnectedToPeer(p peer.ID) bool {
	h.RLock()
	defer h.RUnlock()
	return h.peerConnected[p]
}

func (h *Map) ConnectToShardOfPeer(p peer.AddrInfo) {
	h.Lock()
	defer h.Unlock()
	h.peerConnected[p.ID] = true
}

func (h *Map) DisconnectToShardOfPeer(p peer.AddrInfo) {
	h.Lock()
	defer h.Unlock()
	h.peerConnected[p.ID] = false
}

// IsEnlisted checks if a peer has already registered as a valid highway
func (h *Map) IsEnlisted(p peer.AddrInfo) bool {
	h.RLock()
	defer h.RUnlock()
	_, ok := h.Supports[p.ID]
	return ok
}

func (h *Map) AddPeer(p peer.AddrInfo, supportShards []byte, rpcUrl string) {
	h.Lock()
	defer h.Unlock()
	mcopy := func(b []byte) []byte { // Create new slice and copy
		c := make([]byte, len(b))
		copy(c, b)
		return c
	}

	added := false
	h.Supports[p.ID] = mcopy(supportShards)
	for _, s := range supportShards {
		found := false
		for _, q := range h.Peers[s] {
			if q.ID == p.ID {
				found = true
				break
			}
		}
		if !found {
			added = true
			h.Peers[s] = append(h.Peers[s], p)
		}
	}
	if added {
		logger.Infof("Adding peer %+v, rpcUrl %s, support %v", p, rpcUrl, supportShards)
	}
	h.RPCs[p.ID] = rpcUrl
}

func (h *Map) RemovePeer(p peer.AddrInfo) {
	h.Lock()
	defer h.Unlock()
	delete(h.Supports, p.ID)
	delete(h.RPCs, p.ID)
	for i, addrs := range h.Peers {
		k := 0
		for _, addr := range addrs {
			if addr.ID == p.ID {
				continue
			}
			h.Peers[i][k] = addr
			k++
		}
		if k < len(h.Peers[i]) {
			logger.Infof("Removed peer from map of shard %d: %+v", i, p)
		}
		h.Peers[i] = h.Peers[i][:k]
	}
}

func (h *Map) CopyPeersMap() map[byte][]peer.AddrInfo {
	h.RLock()
	defer h.RUnlock()
	m := map[byte][]peer.AddrInfo{}
	for s, addrs := range h.Peers {
		for _, addr := range addrs { // NOTE: Addrs in AddrInfo are still referenced
			m[s] = append(m[s], addr)
		}
	}
	return m
}

func (h *Map) CopySupports() map[peer.ID][]byte {
	h.RLock()
	defer h.RUnlock()
	m := map[peer.ID][]byte{}
	for pid, cids := range h.Supports {
		c := make([]byte, len(cids))
		copy(c, cids)
		m[pid] = c
	}
	return m
}

func (h *Map) CopyRPCUrls() map[peer.ID]string {
	h.RLock()
	defer h.RUnlock()
	m := map[peer.ID]string{}
	for pid, rpc := range h.RPCs {
		m[pid] = rpc
	}
	return m
}

func (h *Map) CopyConnected() []byte {
	h.RLock()
	defer h.RUnlock()
	c := []byte{}
	for i := byte(0); i < common.NumberOfShard; i++ {
		if h.isConnectedToShard(i) {
			c = append(c, i)
		}
	}
	if h.isConnectedToShard(common.BEACONID) {
		c = append(c, common.BEACONID)
	}
	return c
}

func (h *Map) RemoveRPCUrl(url string) {
	h.Lock()
	defer h.Unlock()
	oldPID := peer.ID("")
	for pid, rpc := range h.RPCs {
		if rpc == url {
			oldPID = pid
		}
	}
	h.RemovePeer(peer.AddrInfo{ID: oldPID})
}
