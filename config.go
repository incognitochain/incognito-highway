package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const NumShards = 8

type ProxyConfig struct {
	proxyPort     int
	adminPort     int
	isProfiling   bool
	supportShards []byte
	privateKey    string
	bootstrap     []string
	version       string
	host          string
}

var proxyConfig *ProxyConfig

func GetProxyConfig() (*ProxyConfig, error) {
	if proxyConfig != nil {
		return proxyConfig, nil
	}
	// get config from process
	proxyPort := flag.Int("proxy_port", 9330, "port for communication with other node (optional, default 3333)")
	adminPort := flag.Int("admin_port", 8080, "rest api /websocket port for administration, monitoring (optional, default 8080)")
	isProfiling := flag.Bool("profiling", false, "enable profiling through admin port")
	supportShards := flag.String("support_shards", "all", "shard list that this proxy will work for (optional, default \"all\")")
	privateKey := flag.String("privatekey", "", "private key of this proxy, use  for authentication with other node")
	bootstrap := flag.String("bootstrap", "", "specify other proxy node to join proxy network as bootstrap node")
	version := flag.String("version", "0.1", "proxy version")
	host := flag.String("host", "0.0.0.0", "listenning address")
	flag.Parse()

	ss, err := parseSupportShards(*supportShards)
	if err != nil {
		return nil, err
	}

	config := &ProxyConfig{
		proxyPort:     *proxyPort,
		adminPort:     *adminPort,
		isProfiling:   *isProfiling,
		supportShards: ss,
		privateKey:    *privateKey,
		bootstrap:     strings.Split(*bootstrap, ","),
		version:       *version,
		host:          *host,
	}
	// if config.privateKey == "" {
	// 	config.printConfig()
	// 	log.Fatal("Need private key")
	// }
	proxyConfig = config
	return proxyConfig, nil
}

func parseSupportShards(s string) ([]byte, error) {
	sup := []byte{}
	switch s {
	case "all":
		sup = append(sup, 255) // beacon
		for i := byte(0); i < NumShards; i++ {
			sup = append(sup, i)
		}
	case "beacon":
		sup = append(sup, 255) // beacon
	case "shard":
		for i := byte(0); i < NumShards; i++ {
			sup = append(sup, i)
		}
	default:
		for _, v := range strings.Split(s, ",") {
			j, err := strconv.Atoi(v)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid support shard: %v", v)
			}
			if j > NumShards || j < 0 {
				return nil, errors.Errorf("invalid support shard: %v", v)
			}
			sup = append(sup, byte(j))
		}
	}
	return sup, nil
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
