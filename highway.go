//+build !test
package main

import (
	"highway/common"
	logger "highway/customizelog"
	"highway/p2p"
	"highway/process"
	_ "net/http/pprof"
)

func main() {
	config := GetProxyConfig()
	config.printConfig()

	//process proxy stream
	proxyHost := p2p.NewHost(config.version, config.host, config.proxyPort, []byte(config.privateKey))
	process.ProcessConnection(proxyHost)
	err := common.InitGenesisCommitteeFromFile("keylist.json", common.NumberOfShard+1, common.CommitteeSize)
	if err != nil {
		return
	}

	// Pubsub
	// proxyHost := p2p.NewHost(
	// 	config.version,
	// 	config.host,
	// 	config.proxyPort,
	// 	[]byte(config.privateKey),
	// )
	if err := process.InitPubSub(proxyHost.Host); err != nil {
		logger.Error(err)
		return
	} else {
		go process.GlobalPubsub.WatchingChain()
		logger.Println("Init ok")
	}

	// go ProcessConnection(proxyHost)
	// time.Sleep(1 * time.Second)
	select {}
	//web server
	// StartMonitorServer(config.adminPort)
}
