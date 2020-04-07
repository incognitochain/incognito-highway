package hmap

import (
	"highway/common"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

type Map struct {
	Peers    map[byte][]peer.AddrInfo // shard => peers
	Supports map[peer.ID][]byte       // peerID => shards supported
	RPCs     map[peer.ID]string       // peerID => RPC endpoint (to call GetPeers)

	status        map[peer.ID]Status
	peerConnected map[peer.ID]bool // keep track of connected peers
	*sync.RWMutex
}

type Status struct {
	Connecting bool      `json:"connecting"`
	Start      time.Time `json:"since"`
}

func NewMap(p peer.AddrInfo, supportShards []byte, rpcUrl string) *Map {
	m := &Map{
		Peers:         map[byte][]peer.AddrInfo{},
		Supports:      map[peer.ID][]byte{},
		RPCs:          map[peer.ID]string{},
		status:        map[peer.ID]Status{},
		peerConnected: map[peer.ID]bool{},
		RWMutex:       &sync.RWMutex{},
	}
	m.AddPeer(p, supportShards, rpcUrl)
	m.ConnectToShardOfPeer(p)
	m.status[p.ID] = Status{
		Connecting: true,
		Start:      time.Now().Add(-time.Duration(common.MinStableDuration)), // Consider ourself as a stable peer
	}
	return m
}

func (m *Map) isConnectedToShard(s byte) bool {
	for pid, conn := range m.peerConnected {
		if conn == false {
			continue
		}

		for _, sup := range m.Supports[pid] {
			if sup == s {
				return true
			}
		}
	}
	return false
}

func (m *Map) IsConnectedToShard(s byte) bool {
	m.RLock()
	defer m.RUnlock()
	return m.isConnectedToShard(s)
}

func (m *Map) IsConnectedToPeer(p peer.ID) bool {
	m.RLock()
	defer m.RUnlock()
	return m.peerConnected[p]
}

func (m *Map) ConnectToShardOfPeer(p peer.AddrInfo) {
	m.Lock()
	defer m.Unlock()
	m.peerConnected[p.ID] = true
	m.updateStatus(p.ID, true)
}

func (m *Map) DisconnectToShardOfPeer(p peer.AddrInfo) {
	m.Lock()
	defer m.Unlock()
	m.peerConnected[p.ID] = false
}

// IsEnlisted checks if a peer has already registered as a valid highway
func (m *Map) IsEnlisted(p peer.AddrInfo) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.Supports[p.ID]
	return ok
}

func (m *Map) updateStatus(pid peer.ID, connecting bool) {
	if s, ok := m.status[pid]; !ok || s.Connecting != connecting {
		m.status[pid] = Status{
			Connecting: connecting,
			Start:      time.Now(),
		}
	}
}

func (m *Map) UpdateStatus(pid peer.ID, connecting bool) {
	m.Lock()
	defer m.Unlock()
	m.updateStatus(pid, connecting)
}

// Status returns connection status of a single peer
func (m *Map) Status(pid peer.ID) (Status, bool) {
	m.RLock()
	defer m.RUnlock()
	if _, ok := m.status[pid]; !ok {
		return Status{}, false
	}
	return m.status[pid], true
}

func (m *Map) AddPeer(p peer.AddrInfo, supportShards []byte, rpcUrl string) {
	m.Lock()
	defer m.Unlock()
	mcopy := func(b []byte) []byte { // Create new slice and copy
		c := make([]byte, len(b))
		copy(c, b)
		return c
	}

	added := false
	m.Supports[p.ID] = mcopy(supportShards)
	for _, s := range supportShards {
		found := false
		for _, q := range m.Peers[s] {
			if q.ID == p.ID {
				found = true
				break
			}
		}
		if !found {
			added = true
			m.Peers[s] = append(m.Peers[s], p)
		}
	}
	if added {
		logger.Infof("Adding peer %+v, rpcUrl %s, support %v", p, rpcUrl, supportShards)
	}
	m.RPCs[p.ID] = rpcUrl
}

func (m *Map) RemovePeer(p peer.AddrInfo) {
	m.Lock()
	defer m.Unlock()
	delete(m.Supports, p.ID)
	delete(m.RPCs, p.ID)
	for i, addrs := range m.Peers {
		k := 0
		for _, addr := range addrs {
			if addr.ID == p.ID {
				continue
			}
			m.Peers[i][k] = addr
			k++
		}
		if k < len(m.Peers[i]) {
			logger.Infof("Removed peer from map of shard %d: %+v", i, p)
		}
		m.Peers[i] = m.Peers[i][:k]
	}
}

func (m *Map) CopyPeersMap() map[byte][]peer.AddrInfo {
	m.RLock()
	defer m.RUnlock()
	peers := map[byte][]peer.AddrInfo{}
	for s, addrs := range m.Peers {
		for _, addr := range addrs { // NOTE: Addrs in AddrInfo are still referenced
			peers[s] = append(peers[s], addr)
		}
	}
	return peers
}

func (m *Map) CopySupports() map[peer.ID][]byte {
	m.RLock()
	defer m.RUnlock()
	sup := map[peer.ID][]byte{}
	for pid, cids := range m.Supports {
		c := make([]byte, len(cids))
		copy(c, cids)
		sup[pid] = c
	}
	return sup
}

func (m *Map) CopyRPCUrls() map[peer.ID]string {
	m.RLock()
	defer m.RUnlock()
	rpcs := map[peer.ID]string{}
	for pid, rpc := range m.RPCs {
		rpcs[pid] = rpc
	}
	return rpcs
}

func (m *Map) CopyConnected() []byte {
	m.RLock()
	defer m.RUnlock()
	c := []byte{}
	for i := byte(0); i < common.NumberOfShard; i++ {
		if m.isConnectedToShard(i) {
			c = append(c, i)
		}
	}
	if m.isConnectedToShard(common.BEACONID) {
		c = append(c, common.BEACONID)
	}
	return c
}

func (m *Map) CopyStatus() map[peer.ID]Status {
	m.RLock()
	defer m.RUnlock()
	status := map[peer.ID]Status{}
	for pid, s := range m.status {
		status[pid] = s // Make a copy
	}
	return status
}
func (m *Map) RemoveRPCUrl(url string) {
	m.Lock()
	oldPID := peer.ID("")
	for pid, rpc := range m.RPCs {
		if rpc == url {
			oldPID = pid
		}
	}
	m.Unlock()
	m.RemovePeer(peer.AddrInfo{ID: oldPID})
}
