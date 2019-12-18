package hmap

import (
	"sync"

	"github.com/libp2p/go-libp2p-core/peer"
)

type Map struct {
	Peers    map[byte][]peer.AddrInfo // shard => peers
	Supports map[peer.ID][]byte       // peerID => shards supported

	connected []byte // keep track of connected shards
	*sync.RWMutex
}

func NewMap(p peer.AddrInfo, supportShards []byte) *Map {
	m := &Map{
		Peers:     map[byte][]peer.AddrInfo{},
		Supports:  map[peer.ID][]byte{},
		connected: supportShards,
		RWMutex:   &sync.RWMutex{},
	}
	m.AddPeer(p, supportShards)
	return m
}

func (h *Map) IsConnectedToShard(s byte) bool {
	h.RLock()
	defer h.RUnlock()
	for _, c := range h.connected {
		if c == s {
			return true
		}
	}
	return false
}

func (h *Map) connectToShard(s byte) {
	h.connected = append(h.connected, s)
}

func (h *Map) ConnectToShardOfPeer(p peer.AddrInfo) {
	h.Lock()
	defer h.Unlock()
	for _, s := range h.Supports[p.ID] {
		h.connectToShard(s)
	}
}

// IsEnlisted checks if a peer has already registered as a valid highway
func (h *Map) IsEnlisted(p peer.AddrInfo) bool {
	h.RLock()
	defer h.RUnlock()
	_, ok := h.Supports[p.ID]
	return ok
}

func (h *Map) AddPeer(p peer.AddrInfo, supportShards []byte) {
	h.Lock()
	defer h.Unlock()
	mcopy := func(b []byte) []byte { // Create new slice and copy
		c := make([]byte, len(b))
		copy(c, b)
		return c
	}

	h.Supports[p.ID] = mcopy(supportShards)
	for _, s := range supportShards {
		h.Peers[s] = append(h.Peers[s], p) // TODO(@0xbunyip): clear h.Peers before appending
	}
}

func (h *Map) RemovePeer(p peer.AddrInfo) {
	h.Lock()
	defer h.Unlock()
	delete(h.Supports, p.ID)
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
			logger.Infof("Remove peer from map of shard %d: %+v", i, p)
		}
		h.Peers[i] = h.Peers[i][:k]
	}
	// TODO(@0xbunyip): update connected
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

func (h *Map) CopyConnected() []byte {
	h.RLock()
	defer h.RUnlock()
	c := make([]byte, len(h.connected))
	copy(c, h.connected)
	return c
}
