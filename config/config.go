package config

import (
	"encoding/base64"
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
	ListenAddr    string
	PublicIP      string
	Masternode    string
	Loglevel      string
	BootnodePort  int
	GrafanaDBURL  string
	HighwayIndex  int
	PrivateSeed   []byte
}

func GetProxyConfig() (*ProxyConfig, error) {
	// get config from process
	proxyPort := flag.Int("proxy_port", 7337, "port for communication with other node (optional, default 7337)")
	bootnodePort := flag.Int("bootnode_port", 9330, "rpc port, returns list of known highway (optional, default 9330)")
	adminPort := flag.Int("admin_port", 8080, "rest api /websocket port for administration, monitoring (optional, default 8080)")
	isProfiling := flag.Bool("profiling", true, "enable profiling through admin port")
	supportShards := flag.String("support_shards", "all", "shard list that this proxy will work for (optional, default \"all\")")
	privateSeed := flag.String("privateseed", "", "private seed of this proxy, use  for authentication with other node")
	index := flag.Int("index", 0, "index of this proxy in list of proxy")
	privateKey := flag.String("privatekey", "", "private key of this proxy, use  for authentication with other node")
	bootstrap := flag.String("bootstrap", "", "address of another highway to get list of highways from (ip:port)")
	version := flag.String("version", "0.1-local", "proxy version")
	listenAddr := flag.String("listenAddr", "0.0.0.0", "listenning address")
	publicIP := flag.String("host", "127.0.0.1", "public ip address")
	masternode := flag.String("masternode", "QmVsCnV9kRZ182MX11CpcHMyFAReyXV49a599AbqmwtNrV", "libp2p PeerID of master node")
	loglevel := flag.String("loglevel", "info", "loglevel for highway, info or debug")
	gDBURL := flag.String("gdburl", "", "Grafana DB URL")
	flag.Parse()

	ss, err := parseSupportShards(*supportShards)
	if err != nil {
		return nil, err
	}
	fmt.Println(*privateSeed)
	seed, err := base64.StdEncoding.DecodeString(*privateSeed)
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
		ListenAddr:    *listenAddr,
		PublicIP:      *publicIP,
		Masternode:    *masternode,
		Loglevel:      *loglevel,
		BootnodePort:  *bootnodePort,
		GrafanaDBURL:  *gDBURL,
		HighwayIndex:  *index,
		PrivateSeed:   seed,
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
	fmt.Println("ListenAddr: ", s.ListenAddr)
	fmt.Println("PublicIP: ", s.PublicIP)
	fmt.Println("Proxy: ", s.ProxyPort)
	fmt.Println("Admin: ", s.AdminPort)
	fmt.Println("IsProfiling: ", s.IsProfiling)
	fmt.Println("Support shards: ", s.SupportShards)
	fmt.Println("Private Key: ", s.PrivateKey)
	fmt.Println("Version: ", s.Version)
	fmt.Println("GrafanaDBURL: ", s.GrafanaDBURL)
	fmt.Println("============== End Config =============")
}
