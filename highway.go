//+build !test
package main

import (
	"highway/chain"
	"highway/common"
	"highway/config"
	"highway/health"
	"highway/monitor"
	"highway/p2p"
	"highway/process"
	"highway/process/topic"
	"highway/route"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

var _ monitor.Monitor = (*config.Reporter)(nil)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	conf, err := config.GetProxyConfig()
	if err != nil {
		logger.Errorf("%+v", err)
		return
	}

	// Setup logging
	initLogger(conf.Loglevel)

	conf.PrintConfig()
	topic.Handler.UpdateSupportShards(conf.SupportShards)
	masterPeerID, err := peer.IDB58Decode(conf.Masternode)
	if err != nil {
		logger.Error(err)
		return
	}

	chainData := new(process.ChainData)
	chainData.Init(common.NumberOfShard)

	// New libp2p host
	proxyHost := p2p.NewHost(conf.Version, conf.Host, conf.ProxyPort, conf.PrivateKey)

	// Pubsub
	if err := process.InitPubSub(proxyHost.Host, conf.SupportShards, chainData); err != nil {
		logger.Error(err)
		return
	}
	logger.Info("Init pubsub ok")
	go process.GlobalPubsub.WatchingChain()

	// Highway manager: connect cross highways
	rman := route.NewManager(
		conf.SupportShards,
		conf.Bootstrap,
		masterPeerID,
		proxyHost.Host,
		proxyHost.GRPC,
	)
	go rman.Start()

	// Chain-facing connections
	chainReporter := chain.ManageChainConnections(proxyHost.Host, rman, proxyHost.GRPC, chainData, conf.SupportShards)

	// // Subscribe to receive new committee
	// process.GlobalPubsub.SubHandlers <- process.SubHandler{
	// 	Topic:   "chain_committee",
	// 	Handler: chainData.ProcessChainCommitteeMsg,
	// }

	// Setup monitoring
	confReporter := config.NewReporter(conf)
	routeReporter := route.NewReporter(rman)
	healthReporter := health.NewReporter()
	processReporter := process.NewReporter(chainData)
	reporters := []monitor.Monitor{confReporter, chainReporter, routeReporter, healthReporter, processReporter}
	timestep := 10 * time.Second // TODO(@0xbunyip): move to config
	monitor.StartMonitorServer(conf.AdminPort, timestep, reporters)

	logger.Info("Serving...")
	proxyHost.GRPC.Serve() // NOTE: must serve after registering all services
}
