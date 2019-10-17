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

	chainData := new(process.ChainData)
	chainData.Init("keylist.json", common.NumberOfShard+1, common.CommitteeSize)
	// Process proxy stream
	proxyHost := p2p.NewHost(config.version, config.host, config.proxyPort, config.privateKey)
	process.RunHighwayServer(proxyHost, chainData, process.NewHighwayClient(proxyHost.GRPC))

	if err := process.InitPubSub(proxyHost.Host, config.supportShards, chainData); err != nil {
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
