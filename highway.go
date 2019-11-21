//+build !test
package main

import (
	"highway/chain"
	"highway/common"
	"highway/p2p"
	"highway/process"
	"highway/process/topic"
	"highway/route"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	config, err := GetProxyConfig()
	if err != nil {
		logger.Errorf("%+v", err)
		return
	}

	// Setup logging
	initLogger(config.loglevel)

	config.printConfig()
	topic.Handler.UpdateSupportShards(config.supportShards)
	masterPeerID, err := peer.IDB58Decode(config.masternode)
	if err != nil {
		logger.Error(err)
		return
	}

	chainData := new(process.ChainData)
	chainData.Init("keylist.json", common.NumberOfShard, common.CommitteeSize, masterPeerID)

	// New libp2p host
	proxyHost := p2p.NewHost(config.version, config.host, config.proxyPort, config.privateKey)

	// Pubsub
	if err := process.InitPubSub(proxyHost.Host, config.supportShards, chainData); err != nil {
		logger.Error(err)
		return
	}
	logger.Info("Init pubsub ok")
	go process.GlobalPubsub.WatchingChain()

	// Highway manager: connect cross highways
	rman := route.NewManager(
		config.supportShards,
		config.bootstrap,
		masterPeerID,
		proxyHost.Host,
		proxyHost.GRPC,
	)
	go rman.Start()

	// Chain-facing connections
	chain.ManageChainConnections(proxyHost.Host, rman, proxyHost.GRPC, chainData, config.supportShards)

	// Subscribe to receive new committee
	process.GlobalPubsub.SubHandlers <- process.SubHandler{
		Topic:   "chain_committee",
		Handler: chainData.ProcessChainCommitteeMsg,
	}

	logger.Info("Serving...")
	proxyHost.GRPC.Serve() // NOTE: must serve after registering all services
}
