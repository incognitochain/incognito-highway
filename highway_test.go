package main

import (
	"context"
	"fmt"
	"highway/p2p"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"
)

func TestProcessConnection(t *testing.T) {
	host1 := p2p.NewHost("0.1", "127.0.0.1", 10000, nil)
	ProcessConnection(host1)

	host2 := p2p.NewHost("0.1", "127.0.0.1", 10001, nil)
	ProcessConnection(host2)

	fmt.Println(peer.AddrInfo{host1.Host.ID(), host1.Host.Addrs()})
	fmt.Println(peer.AddrInfo{host2.Host.ID(), host2.Host.Addrs()})

	err := host1.Host.Connect(context.Background(), peer.AddrInfo{host2.Host.ID(), host2.Host.Addrs()})
	must(err)

	// client := GRPCService_Client{host1.GRPC}
	// peerid, err := peer.IDB58Decode(host2.Host.ID().String())
	// res, err := client.ProxyRegister(context.Background(), peerid, "mypub")
	// fmt.Println(res, err)
}
