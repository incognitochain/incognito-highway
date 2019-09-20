package main

import (
	_ "net/http/pprof"
)

func main() {
	config := GetProxyConfig()
	config.printConfig()

	//process proxy stream
	proxyHost := NewProxyHost(config.version, config.host, config.proxyPort, nil)

	go proxyHost.ProcessConnection()

	//web server
	StartMonitorServer(config.adminPort)
}
