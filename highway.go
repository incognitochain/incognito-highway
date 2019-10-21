//+build !test
package main

import (
	"highway/common"
	logger "highway/customizelog"
	"highway/p2p"
	"highway/process"

	"github.com/libp2p/go-libp2p-core/peer"
)

func main() {
	config, err := GetProxyConfig()
	if err != nil {
		logger.Errorf("%+v", err)
		return
	}
	config.printConfig()

	// Process proxy stream
	proxyHost := p2p.NewHost(config.version, config.host, config.proxyPort, config.privateKey)
	process.RunHighwayServer(proxyHost, process.NewHighwayClient(proxyHost.GRPC))

	if err := common.InitGenesisCommitteeFromFile("keylist.json", common.NumberOfShard+1, common.CommitteeSize); err != nil {
		logger.Error(err)
		return
	}

	if err := process.InitPubSub(proxyHost.Host, config.supportShards); err != nil {
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
	h := NewHighway(
		config.supportShards,
		config.bootstrap,
		masterPeerID,
		proxyHost,
	)
	go h.Start()

	select {}
}
