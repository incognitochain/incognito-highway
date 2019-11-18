package route

import (
	"github.com/libp2p/go-libp2p-core/peer"
)

type Map struct {
	Peers    map[byte][]peer.AddrInfo // shard => peers
	Supports map[peer.ID][]byte       // peerID => shards supported

	connected []byte // keep track of connected shards
}

func NewMap(p peer.AddrInfo, supportShards []byte) *Map {
	m := &Map{
		Peers:     map[byte][]peer.AddrInfo{},
		Supports:  map[peer.ID][]byte{},
		connected: supportShards,
	}
	m.AddPeer(p, supportShards)
	return m
}

func (h *Map) IsConnectedToShard(s byte) bool {
	for _, c := range h.connected {
		if c == s {
			return true
		}
	}
	return false
}

func (h *Map) ConnectToShard(s byte) {
	logger.Debugf("Connected to shard %d", s)
	h.connected = append(h.connected, s)
}

func (h *Map) ConnectToShardOfPeer(p peer.AddrInfo) {
	for _, s := range h.Supports[p.ID] {
		h.ConnectToShard(s)
	}
}

// IsEnlisted checks if a peer has already registered as a valid highway
func (h *Map) IsEnlisted(p peer.AddrInfo) bool {
	_, ok := h.Supports[p.ID]
	return ok
}

func (h *Map) AddPeer(p peer.AddrInfo, supportShards []byte) {
	// TODO(@0xbunyip): serialize all access to prevent race condition

	mcopy := func(b []byte) []byte { // Create new slice and copy
		c := make([]byte, len(b))
		copy(c, b)
		return c
	}

	h.Supports[p.ID] = mcopy(supportShards)
	for _, s := range supportShards {
		h.Peers[s] = append(h.Peers[s], p)
	}
	logger.Info("added peer", p, supportShards)
}
