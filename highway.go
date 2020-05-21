//+build !test

package main

import (
	"fmt"
	"highway/chain"
	"highway/chaindata"
	"highway/common"
	"highway/config"
	"highway/grafana"
	"highway/health"
	"highway/key"
	"highway/monitor"
	"highway/p2p"
	"highway/process"
	"highway/process/topic"
	"highway/route"
	"highway/rpcserver"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

var _ monitor.Monitor = (*config.Reporter)(nil)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	conf, err := config.GetProxyConfig()
	if err != nil {
		fmt.Println(fmt.Errorf("%+v", err))
		return
	}

	// Setup logging
	initLogger(conf.Loglevel)

	conf.PrintConfig()

	//Init grafana log
	gl := grafana.NewLog(
		fmt.Sprintf("%v:%v", conf.PublicIP, conf.ProxyPort),
		conf.Version,
		conf.GrafanaDBURL,
	)
	gl.Start()

	masterPeerID, err := peer.IDB58Decode(conf.Masternode)
	if err != nil {
		logger.Error(err)
		return
	}

	chainData := new(chaindata.ChainData)
	chainData.Init(common.NumberOfShard)

	whitelisthw, hostPriKey, err := key.GenWhiteList(conf.PrivateSeed, conf.HighwayIndex, common.NumberOfHighway)

	// New libp2p host
	proxyHost := p2p.NewHost(conf.Version, conf.ListenAddr, conf.ProxyPort, hostPriKey)

	// Setup topic
	topic.Handler = topic.TopicManager{}
	topic.Handler.Init(proxyHost.Host.ID().String())
	topic.Handler.UpdateSupportShards(conf.SupportShards)

	// Pubsub
	floodPubSub, err := process.NewPubSub(
		proxyHost.Host,
		conf.SupportShards,
		chainData)
	if err != nil {
		logger.Fatal(err)
		return
	}
	logger.Info("Init pubsub ok")
	go floodPubSub.WatchingChain()

	// Highway manager: connect cross highways
	multiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", conf.PublicIP, conf.ProxyPort))
	if err != nil {
		logger.Fatal(err)
		return
	}

	rman := route.NewManager(
		conf.SupportShards,
		conf.Bootstrap,
		masterPeerID,
		proxyHost.Host,
		proxyHost.GRPC,
		multiAddr,
		fmt.Sprintf("%s:%d", conf.PublicIP, conf.BootnodePort),
		floodPubSub,
		gl,
		whitelisthw,
	)
	go rman.Start()

	// RPCServer
	rpcServer, err := rpcserver.NewRPCServer(
		&rpcserver.RpcServerConfig{
			Port: conf.BootnodePort,
		},
		rman.Hmap,
	)
	if err != nil {
		logger.Fatal(err)
		return
	}
	go rpcServer.Start()

	// Chain-facing connections
	chainReporter, err := chain.ManageChainConnections(
		proxyHost.Host,
		rman,
		proxyHost.GRPC,
		chainData,
		conf.SupportShards,
		gl, //GrafanaLog
		conf.HighwayIndex,
	)
	if err != nil {
		logger.Fatal(err)
		return
	}

	// // Subscribe to receive new committee
	// process.GlobalPubsub.SubHandlers <- process.SubHandler{
	// 	Topic:   "chain_committee",
	// 	Handler: chainData.ProcessChainCommitteeMsg,
	// }

	// Setup monitoring
	confReporter := config.NewReporter(conf)
	routeReporter := route.NewReporter(rman, gl)
	healthReporter := health.NewReporter()
	processReporter := process.NewReporter(chainData)
	reporters := []monitor.Monitor{confReporter, chainReporter, routeReporter, healthReporter, processReporter}
	timestep := 10 * time.Second // TODO(@0xbunyip): move to config
	monitor.StartMonitorServer(conf.AdminPort, timestep, reporters)

	logger.Info("Serving...")
	proxyHost.GRPC.Serve() // NOTE: must serve after registering all services

}
