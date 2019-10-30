package chain

import (
	logger "highway/customizelog"
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

func ManageChainConnections(h host.Host, prtc *p2pgrpc.GRPCProtocol) {
	// Manage incoming connections
	m := &Manager{
		newPeers: make(chan PeerInfo, 1000),
	}
	m.peers.ids = map[int][]peer.ID{}
	m.peers.RWMutex = sync.RWMutex{}

	// Server and client instance to communicate to Incognito nodes
	RegisterServer(m, prtc.GetGRPCServer(), NewClient(m, prtc))

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
			logger.Infof("Append new peer: cid = %v, pid = %v", p.CID, p.ID)
			m.peers.Lock()
			m.peers.ids[p.CID] = append(m.peers.ids[p.CID], p.ID)
			m.peers.Unlock()
		}
	}
}

func (m *Manager) Listen(network.Network, multiaddr.Multiaddr)      {}
func (m *Manager) ListenClose(network.Network, multiaddr.Multiaddr) {}
func (m *Manager) Connected(n network.Network, c network.Conn)      {}
func (m *Manager) OpenedStream(network.Network, network.Stream)     {}
func (m *Manager) ClosedStream(network.Network, network.Stream)     {}

func (m *Manager) Disconnected(_ network.Network, conn network.Conn) {
	m.peers.Lock()
	defer m.peers.Unlock()

	// Remove from m.peers to prevent Client from requesting later
	for cid, peers := range m.peers.ids {
		for i, pid := range peers {
			if pid == conn.RemotePeer() {
				l := len(peers)
				peers[i], peers[l-1] = peers[l-1], peers[i]
				m.peers.ids[cid] = peers
				return
			}
		}
	}
}

type PeerInfo struct {
	ID  peer.ID
	CID int // CommitteeID
}
