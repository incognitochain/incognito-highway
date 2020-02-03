package chain

import (
	"highway/chaindata"
	"highway/route"
	"sync"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Manager struct {
	client   *Client
	newPeers chan PeerInfo

	peers struct {
		ids map[int][]PeerInfo
		sync.RWMutex
	}

	conns struct {
		num int
		sync.RWMutex
	}
}

func ManageChainConnections(
	h host.Host,
	rman *route.Manager,
	prtc *p2pgrpc.GRPCProtocol,
	chainData *chaindata.ChainData,
	supportShards []byte,
) *Reporter {
	// Manage incoming connections
	m := &Manager{
		newPeers: make(chan PeerInfo, 1000),
	}
	m.peers.ids = map[int][]PeerInfo{}
	m.peers.RWMutex = sync.RWMutex{}
	m.conns.RWMutex = sync.RWMutex{}

	// Monitor
	reporter := NewReporter(m)

	// Server and client instance to communicate to Incognito nodes
	client := NewClient(m, reporter, rman, prtc, chainData, supportShards)
	RegisterServer(m, prtc.GetGRPCServer(), client, chainData, reporter)
	m.client = client

	h.Network().Notify(m)
	go m.start()
	return reporter
}

func (m *Manager) GetPeers(cid int) []PeerInfo {
	m.peers.RLock()
	defer m.peers.RUnlock()
	ids := m.getPeers(cid)
	return ids
}

func (m *Manager) GetAllPeers() map[int][]PeerInfo {
	m.peers.RLock()
	defer m.peers.RUnlock()
	ids := map[int][]PeerInfo{}
	for cid, _ := range m.peers.ids {
		ids[cid] = m.getPeers(cid)
	}
	return ids
}

func (m *Manager) getPeers(cid int) []PeerInfo {
	peers := []PeerInfo{}
	for _, pinfo := range m.peers.ids[cid] {
		pcopy := PeerInfo{
			ID:     peer.ID(string(pinfo.ID)), // make a copy
			CID:    pinfo.CID,
			Role:   pinfo.Role,
			Pubkey: pinfo.Pubkey,
		}
		peers = append(peers, pcopy)
	}
	return peers
}

func (m *Manager) start() {
	for {
		select {
		case p := <-m.newPeers:
			m.addNewPeer(p)
		}
	}
}

func (m *Manager) addNewPeer(pinfo PeerInfo) {
	m.peers.Lock()
	defer m.peers.Unlock()

	pid := pinfo.ID
	cid := pinfo.CID

	// Remove from previous lists
	m.peers.ids = remove(m.peers.ids, pid)

	// Append to list
	m.peers.ids[cid] = append(m.peers.ids[cid], pinfo)
	logger.Infof("Appended new peer to shard %d, pid = %v, cnt = %d peers", cid, pid, len(m.peers.ids[cid]))
}

func remove(ids map[int][]PeerInfo, rid peer.ID) map[int][]PeerInfo {
	for cid, peers := range ids {
		k := 0
		for _, p := range peers {
			if p.ID == rid {
				continue
			}
			peers[k] = p
			k++
		}

		if k < len(peers) {
			logger.Infof("Removed peer %s from shard %d, remaining %d peers", rid.String(), cid, k)
		}

		ids[cid] = peers[:k]
	}
	return ids
}

func (m *Manager) GetTotalConnections() int {
	m.conns.RLock()
	total := m.conns.num
	m.conns.RUnlock()
	return total
}

func (m *Manager) Listen(network.Network, multiaddr.Multiaddr)      {}
func (m *Manager) ListenClose(network.Network, multiaddr.Multiaddr) {}
func (m *Manager) Connected(n network.Network, c network.Conn) {
	// logger.Println("chain/manager: new conn")
	m.conns.Lock()
	m.conns.num++
	m.conns.Unlock()
}
func (m *Manager) OpenedStream(network.Network, network.Stream) {}
func (m *Manager) ClosedStream(network.Network, network.Stream) {}

func (m *Manager) Disconnected(_ network.Network, conn network.Conn) {
	m.conns.Lock()
	m.conns.num--
	m.conns.Unlock()

	m.peers.Lock()
	defer m.peers.Unlock()
	pid := conn.RemotePeer()
	logger.Infof("Peer disconnected: %s", pid.String())

	// Remove from m.peers to prevent Client from requesting later
	m.peers.ids = remove(m.peers.ids, pid)
	m.client.DisconnectedIDs <- pid
}

type PeerInfo struct {
	ID     peer.ID
	CID    int // CommitteeID
	Role   string
	Pubkey string
}
