package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"testing"
	"time"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func TestProcessConnection(t *testing.T) {
	host1 := NewProxyHost("0.1", "127.0.0.1", 10000, nil)
	host1.ProcessConnection()

	host2 := NewProxyHost("0.1", "127.0.0.1", 10001, nil)
	host2.ProcessConnection()

	fmt.Println(peer.AddrInfo{host1.Host.ID(), host1.Host.Addrs()})
	fmt.Println(peer.AddrInfo{host2.Host.ID(), host2.Host.Addrs()})

	err := host1.Host.Connect(context.Background(), peer.AddrInfo{host2.Host.ID(), host2.Host.Addrs()})
	must(err)

	stream, err := host1.Host.NewStream(context.Background(), host2.Host.ID(), host1.GetProxyStreamProtocolID())
	must(err)

	stream.Write([]byte("asdsd"))
	time.Sleep(time.Second * 1)

	stream, err = host1.Host.NewStream(context.Background(), host2.Host.ID(), host1.GetProxyStreamProtocolID())
	must(err)

	stream.Write([]byte("sdfdfd"))
	time.Sleep(time.Second * 1)

	stream, err = host1.Host.NewStream(context.Background(), host2.Host.ID(), host1.GetProxyStreamProtocolID())
	must(err)

	stream.Write([]byte("sdfdfd"))
	time.Sleep(time.Second * 1)

}
