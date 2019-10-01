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

	// Process proxy stream
	proxyHost := p2p.NewHost(config.version, config.host, config.proxyPort, []byte(config.privateKey))
	process.ProcessConnection(proxyHost)
	err := common.InitGenesisCommitteeFromFile("keylist.json", common.NumberOfShard+1, common.CommitteeSize)
	if err != nil {
		return
	}

	if err := process.InitPubSub(proxyHost.Host); err != nil {
		logger.Error(err)
		return
	}

	process.GlobalPubsub.WatchingChain()
	logger.Println("Init ok")
}
