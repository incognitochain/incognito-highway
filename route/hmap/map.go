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
	logger.Debugf("[map] IsConnectedToShard RLock START")
	defer func() {
		logger.Debugf("[map] IsConnectedToShard RUnlock")
	}()
	h.RLock()
	logger.Debugf("[map] IsConnectedToShard RLock END")
	defer h.RUnlock()
	return h.isConnectedToShard(s)
}

func (h *Map) IsConnectedToPeer(p peer.ID) bool {
	logger.Debugf("[map] IsConnectedToPeer RLock START")
	defer func() {
		logger.Debugf("[map] IsConnectedToPeer RUnlock")
	}()
	h.RLock()
	logger.Debugf("[map] IsConnectedToPeer RLock END")
	defer h.RUnlock()
	return h.peerConnected[p]
}

func (h *Map) ConnectToShardOfPeer(p peer.AddrInfo) {
	logger.Debugf("[map] ConnectToShardOfPeer Lock START")
	defer func() {
		logger.Debugf("[map] ConnectToShardOfPeer Unlock")
	}()
	h.Lock()
	logger.Debugf("[map] ConnectToShardOfPeer Lock END")
	defer h.Unlock()
	h.peerConnected[p.ID] = true
}

func (h *Map) DisconnectToShardOfPeer(p peer.AddrInfo) {
	logger.Debugf("[map] DisconnectToShardOfPeer Lock START")
	defer func() {
		logger.Debugf("[map] DisconnectToShardOfPeer Unlock")
	}()
	h.Lock()
	logger.Debugf("[map] DisconnectToShardOfPeer Lock END")
	defer h.Unlock()
	h.peerConnected[p.ID] = false
}

// IsEnlisted checks if a peer has already registered as a valid highway
func (h *Map) IsEnlisted(p peer.AddrInfo) bool {
	logger.Debugf("[map] IsEnlisted RLock START")
	defer func() {
		logger.Debugf("[map] IsEnlisted RUnlock")
	}()
	h.RLock()
	logger.Debugf("[map] IsEnlisted RLock END")
	defer h.RUnlock()
	_, ok := h.Supports[p.ID]
	return ok
}

func (h *Map) AddPeer(p peer.AddrInfo, supportShards []byte, rpcUrl string) {
	logger.Debugf("[map] AddPeer Lock START")
	defer func() {
		logger.Debugf("[map] AddPeer Unlock")
	}()
	h.Lock()
	logger.Debugf("[map] AddPeer Lock END")
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
	logger.Debugf("[map] RemovePeer Lock START")
	defer func() {
		logger.Debugf("[map] RemovePeer Unlock")
	}()
	h.Lock()
	logger.Debugf("[map] RemovePeer Lock END")
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
	logger.Debugf("[map] CopyPeersMap RLock START")
	defer func() {
		logger.Debugf("[map] CopyPeersMap RUnlock")
	}()
	h.RLock()
	logger.Debugf("[map] CopyPeersMap RLock END")
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
	logger.Debugf("[map] CopySupports RLock START")
	defer func() {
		logger.Debugf("[map] CopySupports RUnlock")
	}()
	h.RLock()
	logger.Debugf("[map] CopySupports RLock END")
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
	logger.Debugf("[map] CopyRPCUrls RLock START")
	defer func() {
		logger.Debugf("[map] CopyRPCUrls RUnlock")
	}()
	h.RLock()
	logger.Debugf("[map] CopyRPCUrls RLock END")
	defer h.RUnlock()
	m := map[peer.ID]string{}
	for pid, rpc := range h.RPCs {
		m[pid] = rpc
	}
	return m
}

func (h *Map) CopyConnected() []byte {
	logger.Debugf("[map] CopyConnected RLock START")
	defer func() {
		logger.Debugf("[map] CopyConnected RUnlock")
	}()
	h.RLock()
	logger.Debugf("[map] CopyConnected RLock END")
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
