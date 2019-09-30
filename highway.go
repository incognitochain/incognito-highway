//+build !test
package main

import (
	"fmt"
	"highway/common"
	"highway/p2p"
	"highway/process"
	_ "net/http/pprof"
	"time"
)

func main() {
	config := GetProxyConfig()
	config.printConfig()

	//process proxy stream
	proxyHost := p2p.NewHost(config.version, config.host, config.proxyPort, []byte(config.privateKey))
	err := common.InitGenesisCommitteeFromFile("keylist.json", common.NumberOfShard+1, common.CommitteeSize)
	if err != nil {
		return
	}
	err = process.InitPubSub(proxyHost.Host)
	if err != nil {
		return
	} else {
		go process.GlobalPubsub.WatchingChain()
		fmt.Println("Init ok")
	}

	// go ProcessConnection(proxyHost)
	time.Sleep(1 * time.Second)
	process.ProcessConnection(proxyHost)

	//web server
	// StartMonitorServer(config.adminPort)
}
