package route

import (
	"fmt"
	"highway/common"
	hmap "highway/route/hmap"
	"highway/route/mocks"
	"highway/rpcserver"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestKeepConnectionAtStart(t *testing.T) {
	discoverer, _ := setupDiscoverer(1)
	manager := setupKeepConnectionTest(discoverer, nil)
	bootstrap := []string{"123123"}
	go manager.keepHighwayConnection(bootstrap)
	time.Sleep(100 * time.Millisecond)
	assert.Len(t, manager.Hmap.Peers[0], 2)
	assert.Len(t, manager.Hmap.Peers[4], 1)
}

func TestQueryRandomPeer(t *testing.T) {
	defer configTime()()

	discoverer, rpcUsed := setupDiscoverer(2)
	manager := setupKeepConnectionTest(discoverer, nil)
	bootstrap := []string{"123123"}
	go manager.keepHighwayConnection(bootstrap)
	// n := 10
	time.Sleep(10 * common.RouteKeepConnectionTimestep)
	assert.Len(t, rpcUsed, 3)
}

func TestConnectInboundPeers(t *testing.T) {
	discoverer, _ := setupDiscoverer(1)
	manager := setupKeepConnectionTest(discoverer, nil)
	manager.Hmap.AddPeer(peer.AddrInfo{ID: peer.ID(pids[0])}, []byte{0, 1, 2}, "")
	manager.Hmap.AddPeer(peer.AddrInfo{ID: peer.ID(pids[1])}, []byte{0, 1, 2}, "")
	manager.Hmap.AddPeer(peer.AddrInfo{ID: peer.ID(pids[2])}, []byte{0, 1, 2}, "")
	manager.checkConnectionStatus()
	assert.True(t, manager.Hmap.IsConnectedToPeer(peer.ID(pids[0])))
	assert.True(t, manager.Hmap.IsConnectedToPeer(peer.ID(pids[1])))
	assert.True(t, manager.Hmap.IsConnectedToPeer(peer.ID(pids[2])))
}

func TestRemoveDeadPeers(t *testing.T) {
	defer configTime()()
	discoverer, _ := setupDiscoverer(1)
	connectedness := []network.Connectedness{network.Connected, network.Connected, network.Connected, network.NotConnected}
	manager := setupKeepConnectionTest(discoverer, connectedness)
	manager.Hmap.AddPeer(peer.AddrInfo{ID: peer.ID(pids[0])}, []byte{0}, "")
	manager.Hmap.AddPeer(peer.AddrInfo{ID: peer.ID(pids[1])}, []byte{0}, "")
	manager.Hmap.AddPeer(peer.AddrInfo{ID: peer.ID(pids[2])}, []byte{0}, "")
	manager.checkConnectionStatus()
	assert.True(t, manager.Hmap.IsConnectedToPeer(peer.ID(pids[0])))
	assert.True(t, manager.Hmap.IsConnectedToPeer(peer.ID(pids[1])))
	assert.True(t, manager.Hmap.IsConnectedToPeer(peer.ID(pids[2])))
	manager.checkConnectionStatus()
	time.Sleep(common.RouteHighwayRemoveDeadline)
	manager.checkConnectionStatus()
	assert.False(t, manager.Hmap.IsConnectedToPeer(peer.ID(pids[0])))
	assert.False(t, manager.Hmap.IsConnectedToPeer(peer.ID(pids[1])))
	assert.False(t, manager.Hmap.IsConnectedToPeer(peer.ID(pids[2])))
}

func TestConnectAllPeers(t *testing.T) {
	discoverer, _ := setupDiscoverer(1)
	manager := setupKeepConnectionTest(discoverer, nil)
	manager.Hmap.AddPeer(peer.AddrInfo{ID: peer.ID(pids[0])}, []byte{0}, "")
	manager.Hmap.AddPeer(peer.AddrInfo{ID: peer.ID(pids[1])}, []byte{1}, "")
	manager.Hmap.AddPeer(peer.AddrInfo{ID: peer.ID(pids[2])}, []byte{2}, "")
	manager.UpdateConnection()
	assert.Len(t, manager.hc.outPeers, 3)
}

func setupDiscoverer(cnt int) (*mocks.HighwayDiscoverer, map[string]int) {
	addrPort := 7337
	rpcPort := 9330
	hwAddrs := map[string][]rpcserver.HighwayAddr{
		"all": []rpcserver.HighwayAddr{},
	}
	for i := 0; i < cnt; i++ {
		hwAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/p2p/%s", addrPort+i, pids[i])
		rpc := fmt.Sprintf("0.0.0.0:%d", rpcPort+i)
		hwAddrs["all"] = append(hwAddrs["all"], rpcserver.HighwayAddr{Libp2pAddr: hwAddr, RPCUrl: rpc})
	}
	discoverer := &mocks.HighwayDiscoverer{}
	rpcUsed := map[string]int{}
	discoverer.On("DiscoverHighway", mock.Anything, mock.Anything).Return(hwAddrs, nil).Run(
		func(args mock.Arguments) {
			rpcUsed[args.Get(0).(string)] = 1
		},
	)
	return discoverer, rpcUsed
}

func setupKeepConnectionTest(discoverer HighwayDiscoverer, connectedness []network.Connectedness) *Manager {
	hmap := hmap.NewMap(peer.AddrInfo{}, []byte{0, 1, 2, 3}, "")
	h, net := setupHost()
	if connectedness == nil {
		connectedness = []network.Connectedness{network.NotConnected, network.Connected}
	}
	setupConnectedness(net, connectedness)
	manager := &Manager{
		host:          h,
		supportShards: []byte{0, 1, 2, 3, 4, 5, 6, 7, 255},
		discoverer:    discoverer,
		Hmap:          hmap,
		hc: &Connector{
			closePeers: make(chan peer.ID, 10),
			outPeers:   make(chan peer.AddrInfo, 10),
		},
	}
	return manager
}

func setupHost() (*mocks.Host, *mocks.Network) {
	net := &mocks.Network{}
	h := &mocks.Host{}
	h.On("Network").Return(net)
	h.On("ID").Return(peer.ID(""))
	h.On("Addrs").Return([]ma.Multiaddr{})
	return h, net
}

func setupConnectedness(net *mocks.Network, values []network.Connectedness) {
	idx := -1
	net.On("Connectedness", mock.Anything).Return(func(_ peer.ID) network.Connectedness {
		if idx+1 < len(values) {
			idx += 1
		}
		return values[idx]
	})
}

func init() {
	cf := zap.NewDevelopmentConfig()
	cf.Level.SetLevel(zapcore.FatalLevel)
	l, _ := cf.Build()
	logger = l.Sugar()

	// chain.InitLogger(logger)
	// chaindata.InitLogger(logger)
	InitLogger(logger)
	// process.InitLogger(logger)
	// topic.InitLogger(logger)
	// health.InitLogger(logger)
	// rpcserver.InitLogger(logger)
	hmap.InitLogger(logger)
	// datahandler.InitLogger(logger)
}

var pids = []string{
	"QmQMsPDbhHZyQMLYPY5WZCLKa2uR9f9QrZWNjE8Yz7gaDF",
	"QmSxPwHv8FPKAcimQ73gL2TTqRwohsTymuDarBckVuR3yK",
	"QmYdTMj3T3eswGoBtvUYD8ifDcaL9ZZssyLqqRqx9vEAhL",
	"QmdeNyfdr6mvcMfDRDbwwoQYpdyw8LyaYwgoc58bLTWVvx",
}
