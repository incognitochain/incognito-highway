//+build !test
package main

import (
	"highway/chain"
	"highway/common"
	logger "highway/customizelog"
	"highway/p2p"
	"highway/process"
	"highway/route"

	"github.com/libp2p/go-libp2p-core/peer"
)

func main() {
	logger.SetConfFile("customizelog/config.conf")
	config, err := GetProxyConfig()
	if err != nil {
		logger.Errorf("%+v", err)
		return
	}
	config.printConfig()

	chainData := new(process.ChainData)
	chainData.Init("keylist.json", common.NumberOfShard+1, common.CommitteeSize)

	// New libp2p host
	proxyHost := p2p.NewHost(config.version, config.host, config.proxyPort, config.privateKey)

	// Chain-facing connections
	go chain.ManageChainConnections(proxyHost.Host, proxyHost.GRPC, chainData)

	if err := process.InitPubSub(proxyHost.Host, config.supportShards, chainData); err != nil {
		logger.Error(err)
		return
	}
	logger.Println("Init pubsub ok")

	go process.GlobalPubsub.WatchingChain()

	// Highway manager: connect cross highways
	masterPeerID, err := peer.IDB58Decode(config.masternode)
	if err != nil {
		logger.Error(err)
		return
	}
	h := route.NewManager(
		config.supportShards,
		config.bootstrap,
		masterPeerID,
		proxyHost.Host,
		proxyHost.GRPC,
	)
	go h.Start()

	go proxyHost.GRPC.Serve() // NOTE: must serve after registering all services
	select {}
}
