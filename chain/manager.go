package chain

import (
	"fmt"
	"highway/chaindata"
	"highway/grafana"
	"highway/route"
	"sync"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Manager struct {
	server   *Server
	client   *Client
	newPeers chan PeerInfo

	peers struct {
		ids map[int][]PeerInfo
		sync.RWMutex
	}

	conns struct {
		num   int
		addrs map[peer.ID]multiaddr.Multiaddr
		sync.RWMutex
	}

	gralog  *grafana.GrafanaLog
	watcher *watcher
	host    host.Host
}

func ManageChainConnections(
	h host.Host,
	rman *route.Manager,
	prtc *p2pgrpc.GRPCProtocol,
	chainData *chaindata.ChainData,
	supportShards []byte,
	gl *grafana.GrafanaLog,
	hwid int,
) (*Reporter, error) {
	// Manage incoming connections
	m := &Manager{
		newPeers: make(chan PeerInfo, 1000),
		watcher:  newWatcher(gl, hwid),
		host:     h,
	}
	go m.watcher.process()
	m.peers.ids = map[int][]PeerInfo{}
	m.peers.RWMutex = sync.RWMutex{}
	m.conns.RWMutex = sync.RWMutex{}
	m.conns.addrs = map[peer.ID]multiaddr.Multiaddr{}
	m.gralog = gl

	// Monitor
	reporter := NewReporter(m)

	// Server and client instance to communicate to Incognito nodes
	client := NewClient(m, reporter, rman, prtc, chainData, supportShards)
	server, err := RegisterServer(m, prtc.GetGRPCServer(), client, chainData, reporter)
	if err != nil {
		return nil, err
	}

	m.server = server
	m.client = client

	h.Network().Notify(m)
	go m.start()
	return reporter, nil
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
	pid := pinfo.ID
	cid := pinfo.CID

	m.peers.Lock()
	// m.peers.ids = remove(m.peers.ids, pid, m.gralog)   // Remove from previous lists
	m.peers.ids[cid] = append(m.peers.ids[cid], pinfo) // Append to list
	logger.Infof("Appended new peer to shard %d, pid = %v, cnt = %d peers", cid, pid, len(m.peers.ids[cid]))
	m.peers.Unlock()

	if m.gralog != nil {
		m.gralog.Add(fmt.Sprintf("total_cid_%v", cid), len(m.peers.ids[cid]))

		m.watchPeer(pinfo)
	}
}

func (m *Manager) watchPeer(pinfo PeerInfo) {
	var maddr multiaddr.Multiaddr
	ip4 := ""
	port := ""

	m.conns.RLock()
	if savedAddr, ok := m.conns.addrs[pinfo.ID]; ok {
		// Use global ip address if we have it
		maddr = savedAddr
	} else if peerstoreAddrs := m.host.Peerstore().PeerInfo(pinfo.ID).Addrs; len(peerstoreAddrs) > 0 {
		// Otherwise, get an ip address (even if it's local)
		maddr = peerstoreAddrs[0]
	}
	m.conns.RUnlock()

	if maddr != nil {
		ip4, _ = maddr.ValueForProtocol(multiaddr.P_IP4)
		port, _ = maddr.ValueForProtocol(multiaddr.P_TCP)
	}
	m.watcher.markPeer(pinfo, fmt.Sprintf("%s:%s", ip4, port))
}

func remove(ids map[int][]PeerInfo, rid peer.ID, gl *grafana.GrafanaLog) map[int][]PeerInfo {
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
			if gl != nil {
				gl.Add(fmt.Sprintf("total_cid_%v", cid), k)
			}
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
func (m *Manager) Connected(n network.Network, conn network.Conn) {
	// logger.Println("chain/manager: new conn")
	m.conns.Lock()
	m.conns.num++
	m.conns.addrs[conn.RemotePeer()] = conn.RemoteMultiaddr()
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
	logger.Infof("Peer disconnected: %s %v", pid.String(), conn.RemoteMultiaddr())

	// Remove from m.peers to prevent Client from requesting later
	m.peers.ids = remove(m.peers.ids, pid, m.gralog)
	m.client.DisconnectedIDs <- pid
	m.watcher.unmarkPeer(pid)
}

type PeerInfo struct {
	ID     peer.ID
	CID    int // CommitteeID
	Role   string
	Pubkey string
}
