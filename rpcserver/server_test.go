package rpcserver

import (
	"fmt"
	"net/rpc"
	"testing"
	"time"
)

func TestRpcServer_GetHWForAll(t *testing.T) {
	selfIPFSAddr := fmt.Sprintf("/ip4/%v/tcp/%v/p2p/%v", "127.0.0.1", "9330", "QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv")
	rpcConf := &RpcServerConfig{
		Port:     9330,
		IPFSAddr: selfIPFSAddr,
	}
	server, err := NewRPCServer(rpcConf)
	if err != nil {
		t.Error(err)
	}
	go server.Start()
	time.Sleep(3 * time.Second)

	client, err := rpc.Dial("tcp", "127.0.0.1:9330")
	if err != nil {
		t.Error(err)
	}

	peerExpected := "/ip4/127.0.0.1/tcp/9330/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv"

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
				if peerHW == peerExpected {
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
