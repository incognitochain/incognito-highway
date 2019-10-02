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
	proxyHost := p2p.NewHost(config.version, config.host, config.proxyPort, []byte(config.privateKey))
	process.ProcessConnection(proxyHost)

	if err := common.InitGenesisCommitteeFromFile("keylist.json", common.NumberOfShard+1, common.CommitteeSize); err != nil {
		logger.Error(err)
		return
	}

	if err := process.InitPubSub(proxyHost.Host); err != nil {
		logger.Error(err)
		return
	}

	process.GlobalPubsub.WatchingChain()

	// Highway manager: connect cross shards
	h := NewHighway(config.supportShards, config.bootstrap, proxyHost.Host)
	go h.Start()

	logger.Println("Init ok")
}
