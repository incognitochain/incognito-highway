package main

import (
	logger "highway/customizelog"

	"github.com/libp2p/go-libp2p-core/peer"
)

type HighwayMap struct {
	Peers    map[byte][]peer.AddrInfo // shard => peers
	Supports map[peer.ID][]byte       // peerID => shards supported

	connected []byte // keep track of connected shards
}

func NewHighwayMap(p peer.AddrInfo, supportShards []byte) *HighwayMap {
	m := &HighwayMap{
		Peers:     map[byte][]peer.AddrInfo{},
		Supports:  map[peer.ID][]byte{},
		connected: supportShards,
	}
	m.AddPeer(p, supportShards)
	return m
}

func (h *HighwayMap) IsConnectedToShard(s byte) bool {
	for _, c := range h.connected {
		if c == s {
			return true
		}
	}
	return false
}

func (h *HighwayMap) ConnectToShard(s byte) {
	h.connected = append(h.connected, s)
}

func (h *HighwayMap) ConnectToShardOfPeer(p peer.AddrInfo) {
	for _, s := range h.Supports[p.ID] {
		h.ConnectToShard(s)
	}
}

// IsEnlisted checks if a peer has already registered as a valid highway
func (h *HighwayMap) IsEnlisted(p peer.AddrInfo) bool {
	_, ok := h.Supports[p.ID]
	return ok
}

func (h *HighwayMap) AddPeer(p peer.AddrInfo, supportShards []byte) {
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
