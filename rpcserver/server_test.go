package rpcserver

import (
	"fmt"
	"highway/route/hmap"
	"net/rpc"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestRpcServer_GetHWForAll(t *testing.T) {
	// selfIPFSAddr := fmt.Sprintf("/ip4/%v/tcp/%v/p2p/%v", "127.0.0.1", "9330", "QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv")
	rpcConf := &RpcServerConfig{
		Port: 9330,
		// IPFSAddr: selfIPFSAddr,
	}
	pid, _ := peer.IDB58Decode("QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv")
	addr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", "127.0.0.1", 9330))
	p := peer.AddrInfo{
		ID:    pid,
		Addrs: []multiaddr.Multiaddr{addr},
	}

	server, err := NewRPCServer(rpcConf, hmap.NewMap(p, []byte{0, 1, 2}, "127.0.0.1:9330"))
	if err != nil {
		t.Error(err)
	}
	go server.Start()
	time.Sleep(100 * time.Millisecond)

	client, err := rpc.Dial("tcp", "127.0.0.1:9330")
	if err != nil {
		t.Error(err)
	}

	peerExpected := "/ip4/127.0.0.1/tcp/9330/ipfs/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"

	if client != nil {
		defer client.Close()
		// var rawResponse map[string][]string
		var raw Response
		err := client.Call("Handler.GetPeers", Request{
			Shard: []string{"all"},
		}, &raw)
		if err != nil {
			t.Errorf("Calling RPC Server err %v", err)
		} else {
			peersHWForAll, ok := raw.PeerPerShard["all"]
			if !ok {
				t.Error("RPC server is not contain HW peer for all shard")
			}
			ok = false
			for _, peerHW := range peersHWForAll {
				if peerHW.Libp2pAddr == peerExpected {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf("got %v, want %v", peersHWForAll, peerExpected)
			}
		}
	}
}

func init() {
	cf := zap.NewDevelopmentConfig()
	cf.Level.SetLevel(zapcore.FatalLevel)
	l, _ := cf.Build()
	logger = l.Sugar()

	// chain.InitLogger(logger)
	// chaindata.InitLogger(logger)
	InitLogger(logger)
	// process.InitLogger(logger)
	// topic.InitLogger(logger)
	// health.InitLogger(logger)
	// rpcserver.InitLogger(logger)
	hmap.InitLogger(logger)
	// datahandler.InitLogger(logger)
}
