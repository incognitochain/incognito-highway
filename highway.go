//+build !test
package main

import (
	"highway/common"
	logger "highway/customizelog"
	"highway/p2p"
	"highway/process"
)

func main() {
	config := GetProxyConfig()
	config.printConfig()

	// Load committee
	if err := common.InitGenesisCommitteeFromFile("keylist.json", common.NumberOfShard+1, common.CommitteeSize); err != nil {
		logger.Error(err)
		return
	}

	// Pubsub
	proxyHost := p2p.NewHost(
		config.version,
		config.host,
		config.proxyPort,
		[]byte(config.privateKey),
	)
	if err := process.InitPubSub(proxyHost.Host); err != nil {
		logger.Error(err)
		return
	} else {
		go process.GlobalPubsub.WatchingChain()
		logger.Println("Init ok")
	}

	// gRPC
	// go process.ProcessConnection(proxyHost)
	// process.ProcessConnection(proxyHost)

	select {}

	//web server
	// StartMonitorServer(config.adminPort)
}
