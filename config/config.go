package config

import (
	"flag"
	"fmt"
	"highway/common"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type ProxyConfig struct {
	ProxyPort     int
	AdminPort     int
	IsProfiling   bool
	SupportShards []byte
	PrivateKey    string
	Bootstrap     []string
	Version       string
	Host          string
	Masternode    string
	Loglevel      string
	BootnodePort  int
}

func GetProxyConfig() (*ProxyConfig, error) {
	// get config from process
	proxyPort := flag.Int("proxy_port", 9330, "port for communication with other node (optional, default 3333)")
	bootnodePort := flag.Int("bootnode_port", 9334, "port for communication with other node (optional, default 9334)")
	adminPort := flag.Int("admin_port", 8080, "rest api /websocket port for administration, monitoring (optional, default 8080)")
	isProfiling := flag.Bool("profiling", false, "enable profiling through admin port")
	supportShards := flag.String("support_shards", "all", "shard list that this proxy will work for (optional, default \"all\")")
	privateKey := flag.String("privatekey", "", "private key of this proxy, use  for authentication with other node")
	bootstrap := flag.String("bootstrap", "", "specify other proxy node to join proxy network as bootstrap node")
	version := flag.String("version", "0.1", "proxy version")
	host := flag.String("host", "0.0.0.0", "listenning address")
	masternode := flag.String("masternode", "QmVsCnV9kRZ182MX11CpcHMyFAReyXV49a599AbqmwtNrV", "libp2p PeerID of master node")
	loglevel := flag.String("loglevel", "info", "loglevel for highway, info or debug")
	flag.Parse()

	ss, err := parseSupportShards(*supportShards)
	if err != nil {
		return nil, err
	}

	config := &ProxyConfig{
		ProxyPort:     *proxyPort,
		AdminPort:     *adminPort,
		IsProfiling:   *isProfiling,
		SupportShards: ss,
		PrivateKey:    *privateKey,
		Bootstrap:     strings.Split(*bootstrap, ","),
		Version:       *version,
		Host:          *host,
		Masternode:    *masternode,
		Loglevel:      *loglevel,
		BootnodePort:  *bootnodePort,
	}
	// if config.privateKey == "" {
	// 	config.printConfig()
	// 	log.Fatal("Need private key")
	// }
	return config, nil
}

func parseSupportShards(s string) ([]byte, error) {
	sup := []byte{}
	switch s {
	case "all":
		sup = append(sup, 255) // beacon
		for i := byte(0); i < common.NumberOfShard; i++ {
			sup = append(sup, i)
		}
	case "beacon":
		sup = append(sup, 255) // beacon
	case "shard":
		for i := byte(0); i < common.NumberOfShard; i++ {
			sup = append(sup, i)
		}
	default:
		for _, v := range strings.Split(s, ",") {
			j, err := strconv.Atoi(v)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid support shard: %v", v)
			}
			if j > common.NumberOfShard || j < 0 {
				return nil, errors.Errorf("invalid support shard: %v", v)
			}
			sup = append(sup, byte(j))
		}
	}
	return sup, nil
}

func (s ProxyConfig) PrintConfig() {
	fmt.Println("============== Config =============")
	fmt.Println("Bootstrap node: ", s.Bootstrap)
	fmt.Println("Host: ", s.Host)
	fmt.Println("Proxy: ", s.ProxyPort)
	fmt.Println("Admin: ", s.AdminPort)
	fmt.Println("IsProfiling: ", s.IsProfiling)
	fmt.Println("Support shards: ", s.SupportShards)
	fmt.Println("Private Key: ", s.PrivateKey)
	fmt.Println("============== End Config =============")
}
