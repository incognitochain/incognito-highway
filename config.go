package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

type ProxyConfig struct {
	proxyPort     int
	adminPort     int
	isProfiling   bool
	supportShards []string
	privateKey    string
	bootstrap     []string
	version       string
	host          string
}

var proxyConfig *ProxyConfig

func GetProxyConfig() *ProxyConfig {
	if proxyConfig != nil {
		return proxyConfig
	}
	// get config from process
	proxyPort := flag.Int("proxy_port", 9330, "port for communication with other node (optional, default 3333)")
	adminPort := flag.Int("admin_port", 8080, "rest api /websocket port for administration, monitoring (optional, default 8080)")
	isProfiling := flag.Bool("profiling", false, "enable profiling through admin port")
	supportShards := flag.String("shard", "", "shard list that this proxy will work for (optional, default \"all\")")
	privateKey := flag.String("privatekey", "", "private key of this proxy, use  for authentication with other node")
	bootstrap := flag.String("bootstrap", "", "specify other proxy node to join proxy network as bootstrap node")
	version := flag.String("version", "0.1", "proxy version")
	host := flag.String("host", "127.0.0.1", "listenning address")
	flag.Parse()

	config := &ProxyConfig{
		proxyPort:     *proxyPort,
		adminPort:     *adminPort,
		isProfiling:   *isProfiling,
		supportShards: strings.Split(*supportShards, ","),
		privateKey:    *privateKey,
		bootstrap:     strings.Split(*bootstrap, ","),
		version:       *version,
		host:          *host,
	}
	if config.privateKey == "" {
		config.printConfig()
		log.Fatal("Need private key")
	}
	proxyConfig = config
	return proxyConfig
}

func (s ProxyConfig) printConfig() {
	fmt.Println("============== Config =============")
	fmt.Println("Bootstrap node: ", s.bootstrap)
	fmt.Println("Host: ", s.host)
	fmt.Println("Proxy: ", s.proxyPort)
	fmt.Println("Admin: ", s.adminPort)
	fmt.Println("IsProfiling: ", s.isProfiling)
	fmt.Println("Support shards: ", s.supportShards)
	fmt.Println("Private Key: ", s.privateKey)
	fmt.Println("============== End Config =============")
}
