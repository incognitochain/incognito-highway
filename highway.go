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
	config, err := GetProxyConfig()
	if err != nil {
		logger.Errorf("%+v", err)
		return
	}
	config.printConfig()

	chainData := new(process.ChainData)
	chainData.Init("keylist.json", common.NumberOfShard+1, common.CommitteeSize)
	// Process proxy stream
	proxyHost := p2p.NewHost(config.version, config.host, config.proxyPort, config.privateKey)
	// process.RunHighwayServer(proxyHost, process.NewHighwayClient(proxyHost.GRPC, chainData))
	chain.RegisterServer(proxyHost, chain.NewClient(proxyHost.GRPC, chainData))
	if err := process.InitPubSub(proxyHost.Host, config.supportShards, chainData); err != nil {
		logger.Error(err)
		return
	}
	logger.Println("Init pubsub ok")

	go process.GlobalPubsub.WatchingChain()

	// Highway manager: connect cross shards
	masterPeerID, err := peer.IDB58Decode(config.masternode)
	if err != nil {
		logger.Error(err)
		return
	}
	h := route.NewManager(
		config.supportShards,
		config.bootstrap,
		masterPeerID,
		proxyHost,
	)
	go h.Start()

	go proxyHost.GRPC.Serve() // NOTE: must serve after registering all services
	select {}
}
