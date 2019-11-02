package chain

import (
	"highway/process"
	"sync"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Manager struct {
	newPeers chan PeerInfo

	peers struct {
		ids map[int][]peer.ID
		sync.RWMutex
	}
}

func ManageChainConnections(h host.Host, prtc *p2pgrpc.GRPCProtocol, chainData *process.ChainData) {
	// Manage incoming connections
	m := &Manager{
		newPeers: make(chan PeerInfo, 1000),
	}
	m.peers.ids = map[int][]peer.ID{}
	m.peers.RWMutex = sync.RWMutex{}

	// Server and client instance to communicate to Incognito nodes
	RegisterServer(m, prtc.GetGRPCServer(), NewClient(m, prtc, chainData))

	h.Network().Notify(m)
	m.start()
}

func (m *Manager) GetPeers(cid int) []peer.ID {
	m.peers.RLock()
	defer m.peers.RUnlock()
	ids := []peer.ID{}
	for _, id := range m.peers.ids[cid] {
		s := string(id) // make a copy
		ids = append(ids, peer.ID(s))
	}
	return ids
}

func (m *Manager) start() {
	for {
		select {
		case p := <-m.newPeers:
			m.addNewPeer(p.ID, p.CID)
		}
	}
}

func (m *Manager) addNewPeer(pid peer.ID, cid int) {
	m.peers.Lock()
	defer m.peers.Unlock()

	// Remove from previous lists
	m.peers.ids = remove(m.peers.ids, pid)

	// Append to list
	m.peers.ids[cid] = append(m.peers.ids[cid], pid)
	logger.Infof("Appended new peer to shard %d, pid = %v", cid, pid)
}

func remove(ids map[int][]peer.ID, rid peer.ID) map[int][]peer.ID {
	for cid, peers := range ids {
		k := 0
		for _, pid := range peers {
			if pid == rid {
				continue
			}
			peers[k] = pid
			k++
		}

		if k < len(peers) {
			logger.Infof("Removed peer %s from shard %d, remaining %d peers", rid.String(), cid, k)
		}

		ids[cid] = peers[:k]

	}
	return ids
}

func (m *Manager) Listen(network.Network, multiaddr.Multiaddr)      {}
func (m *Manager) ListenClose(network.Network, multiaddr.Multiaddr) {}
func (m *Manager) Connected(n network.Network, c network.Conn) {
	// logger.Println("chain/manager: new conn")
}
func (m *Manager) OpenedStream(network.Network, network.Stream) {}
func (m *Manager) ClosedStream(network.Network, network.Stream) {}

func (m *Manager) Disconnected(_ network.Network, conn network.Conn) {
	m.peers.Lock()
	defer m.peers.Unlock()
	pid := conn.RemotePeer()
	logger.Info("Peer disconnected:", pid.String())

	// Remove from m.peers to prevent Client from requesting later
	m.peers.ids = remove(m.peers.ids, pid)
}

type PeerInfo struct {
	ID  peer.ID
	CID int // CommitteeID
}
