package rpcserver

import (
	"fmt"
	"net/rpc"
	"testing"
)

func TestHandler_GetPeers(t *testing.T) {
	// func (rpcClient *RPCClient) DiscoverHighway(
	// 	discoverPeerAddress string,
	// 	shardsStr []string,
	// ) (
	// 	map[string][]HighwayAddr,
	// 	error,
	// ) {
	discoverPeerAddress := "139.162.9.169:9330" // == "" {
	client := new(rpc.Client)
	var err error
	client, err = rpc.Dial("tcp", discoverPeerAddress)
	fmt.Println("Dialing...")
	if err != nil {
		t.Errorf("Connect to discover peer %v return error %v:", discoverPeerAddress, err)
	}
	defer client.Close()

	fmt.Printf("Connected to %v \n", discoverPeerAddress)
	req := Request{Shard: []string{"all"}}
	var res Response
	fmt.Printf("Start dialing RPC server with param %v\n", req)

	err = client.Call("Handler.GetPeers", req, &res)

	if err != nil {
		t.Errorf("Call Handler.GetPeers return error %v", err)
	}
	fmt.Printf("Bootnode return %v", res.PeerPerShard)
	// return res.PeerPerShard, nil
}
