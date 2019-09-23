package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"highway/p2p"
	"testing"
	"time"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func TestProcessConnection(t *testing.T) {
	host1 := p2p.NewHost("0.1", "127.0.0.1", 10000, nil)
	ProcessConnection(host1)

	host2 := p2p.NewHost("0.1", "127.0.0.1", 10001, nil)
	ProcessConnection(host2)

	fmt.Println(peer.AddrInfo{host1.Host.ID(), host1.Host.Addrs()})
	fmt.Println(peer.AddrInfo{host2.Host.ID(), host2.Host.Addrs()})

	err := host1.Host.Connect(context.Background(), peer.AddrInfo{host2.Host.ID(), host2.Host.Addrs()})
	must(err)

	stream, err := host1.Host.NewStream(context.Background(), host2.Host.ID(), host2.GetProxyStreamProtocolID())
	must(err)
	//
	stream.Write([]byte("asdsd"))
	time.Sleep(1 * time.Second)

}
