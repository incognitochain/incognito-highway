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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestKeepConnectionAtStart(t *testing.T) {
	discoverer, _ := setupDiscoverer(1)
	manager := setupKeepConnectionTest(discoverer)
	bootstrap := []string{"123123"}
	go manager.keepHighwayConnection(bootstrap)
	time.Sleep(1 * time.Second)
	assert.Len(t, manager.Hmap.Peers[0], 2)
	assert.Len(t, manager.Hmap.Peers[4], 1)
}

func TestQueryRandomPeer(t *testing.T) {
	discoverer, rpcUsed := setupDiscoverer(2)
	manager := setupKeepConnectionTest(discoverer)
	bootstrap := []string{"123123"}
	go manager.keepHighwayConnection(bootstrap)
	// n := 10
	time.Sleep(10 * common.RouteKeepConnectionTimestep)
	assert.Len(t, rpcUsed, 3)
}

func setupDiscoverer(cnt int) (*mocks.HighwayDiscoverer, map[string]int) {
	addrPort := 7337
	rpcPort := 9330
	hwAddrs := map[string][]rpcserver.HighwayAddr{
		"all": []rpcserver.HighwayAddr{},
	}
	pids := []string{
		"QmQMsPDbhHZyQMLYPY5WZCLKa2uR9f9QrZWNjE8Yz7gaDF",
		"QmSxPwHv8FPKAcimQ73gL2TTqRwohsTymuDarBckVuR3yK",
		"QmYdTMj3T3eswGoBtvUYD8ifDcaL9ZZssyLqqRqx9vEAhL",
		"QmdeNyfdr6mvcMfDRDbwwoQYpdyw8LyaYwgoc58bLTWVvx",
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

func setupKeepConnectionTest(discoverer HighwayDiscoverer) *Manager {
	hmap := hmap.NewMap(peer.AddrInfo{}, []byte{0, 1, 2, 3}, "")
	h, net := setupHost()
	setupConnectedness(net, []network.Connectedness{network.NotConnected, network.Connected})
	manager := &Manager{
		host:       h,
		discoverer: discoverer,
		Hmap:       hmap,
	}
	return manager
}

func setupHost() (*mocks.Host, *mocks.Network) {
	net := &mocks.Network{}
	h := &mocks.Host{}
	h.On("Network").Return(net)
	h.On("ID").Return(peer.ID(""))
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
