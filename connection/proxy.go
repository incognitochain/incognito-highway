package proxy

import (
	"bufio"
	"context"
	"fmt"

	crypto2 "crypto"

	"github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/protocol"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

type PeerConn struct {
	RemotePeer *Peer
	RW         *bufio.ReadWriter
}

type Peer struct {
	IP            string
	Port          int
	TargetAddress []core.Multiaddr
	PeerID        peer.ID
	PublicKey     crypto2.PublicKey
}

type ProxyHostConfig struct {
	MaxConnection int
	PublicIP      string
	Port          int
	PrivateKey    crypto.PrivKey
}

type ProxyHost struct {
	Version  string
	Host     host.Host
	SelfPeer *Peer
}

func (s ProxyHost) GetProxyStreamProtocolID() protocol.ID {
	return protocol.ID("proxy/stream/" + s.Version)
}

func (s ProxyHost) GetMetadataProtocolID() protocol.ID {
	return protocol.ID("proxy/meta/" + s.Version)
}

func NewProxyHost(version string, pubIP string, port int, prv crypto.Key) *ProxyHost {
	listenAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", pubIP, port))
	catchError(err)

	ctx := context.Background()
	opts := []libp2p.Option{
		libp2p.ConnectionManager(nil),
		libp2p.ListenAddrs(listenAddr),
	}

	p2pHost, err := libp2p.New(ctx, opts...)

	selfPeer := &Peer{
		PeerID:        p2pHost.ID(),
		IP:            pubIP,
		Port:          port,
		TargetAddress: append([]multiaddr.Multiaddr{}, listenAddr),
	}

	node := &ProxyHost{
		Host:     p2pHost,
		SelfPeer: selfPeer,
		Version:  version,
	}

	return node
}

func catchError(err error) {
	if err != nil {
		panic(err)
	}
}
