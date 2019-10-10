//+build !test
package main

import (
	"highway/common"
	logger "highway/customizelog"
	"highway/p2p"
	"highway/process"
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

	if err := process.InitPubSub(proxyHost.Host); err != nil {
		logger.Error(err)
		return
	}
	logger.Println("Init ok")

	go process.GlobalPubsub.WatchingChain()

	// Highway manager: connect cross shards
	h := NewHighway(config.supportShards, config.bootstrap, proxyHost.Host)
	go h.Start()

	select {}
}
