//+build !test
package main

import (
	"highway/common"
	"highway/p2p"
	"highway/process"
	_ "net/http/pprof"
)

func main() {
	config := GetProxyConfig()
	config.printConfig()

	//process proxy stream
	proxyHost := p2p.NewHost(config.version, config.host, config.proxyPort, []byte(config.privateKey))

	// go ProcessConnection(proxyHost)
	process.ProcessConnection(proxyHost)
	err := common.InitGenesisCommitteeFromFile("keylist.json", common.NumberOfShard+1, common.CommitteeSize)
	if err != nil {
		return
	}
	err = process.InitPubSub(proxyHost.Host)
	if err != nil {
		return
	} else {
		go process.GlobalPubsub.WatchingChain()
	}

	//web server
	StartMonitorServer(config.adminPort)
}
