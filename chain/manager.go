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

	h.Network().Notify((*notifiee)(m))
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

type notifiee Manager

func (no *notifiee) Listen(network.Network, multiaddr.Multiaddr)      {}
func (no *notifiee) ListenClose(network.Network, multiaddr.Multiaddr) {}
func (no *notifiee) Connected(n network.Network, c network.Conn)      {}
func (no *notifiee) OpenedStream(network.Network, network.Stream)     {}
func (no *notifiee) ClosedStream(network.Network, network.Stream)     {}

func (no *notifiee) Disconnected(network.Network, network.Conn) {
}

type PeerInfo struct {
	ID  peer.ID
	CID int // CommitteeID
}
