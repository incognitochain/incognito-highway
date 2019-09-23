package main

import (
	"highway/p2p"
	_ "net/http/pprof"
)

func main() {
	config := GetProxyConfig()
	config.printConfig()

	//process proxy stream
	proxyHost := p2p.NewHost(config.version, config.host, config.proxyPort, []byte(config.privateKey))

	go ProcessConnection(proxyHost)

	//web server
	StartMonitorServer(config.adminPort)
}
