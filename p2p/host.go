package p2p

import (
	"context"
	crypto2 "crypto"
	"fmt"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Peer struct {
	IP            string
	Port          int
	TargetAddress []core.Multiaddr
	PeerID        peer.ID
	PublicKey     crypto2.PublicKey
}

type HostConfig struct {
	MaxConnection int
	PublicIP      string
	Port          int
	PrivateKey    crypto.PrivKey
}

type Host struct {
	Version  string
	Host     host.Host
	SelfPeer *Peer
	GRPC     *p2pgrpc.GRPCProtocol
}

func NewHost(version string, pubIP string, port int, privKeyStr string) *Host {
	var privKey crypto.PrivKey
	if len(privKeyStr) == 0 {
		privKey, _, _ = crypto.GenerateKeyPair(crypto.ECDSA, 2048)
		m, _ := crypto.MarshalPrivateKey(privKey)
		encoded := crypto.ConfigEncodeKey(m)
		fmt.Println("encoded libp2p key:", encoded)
	} else {
		b, err := crypto.ConfigDecodeKey(privKeyStr)
		catchError(err)
		privKey, err = crypto.UnmarshalPrivateKey(b)
		catchError(err)
	}

	listenAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", pubIP, port))
	catchError(err)

	ctx := context.Background()
	opts := []libp2p.Option{
		libp2p.ConnectionManager(nil),
		libp2p.ListenAddrs(listenAddr),
		libp2p.Identity(privKey),
	}

	p2pHost, err := libp2p.New(ctx, opts...)
	catchError(err)
	fmt.Println(p2pHost.Addrs())

	selfPeer := &Peer{
		PeerID:        p2pHost.ID(),
		IP:            pubIP,
		Port:          port,
		TargetAddress: append([]multiaddr.Multiaddr{}, listenAddr),
	}

	node := &Host{
		Host:     p2pHost,
		SelfPeer: selfPeer,
		Version:  version,
		GRPC:     p2pgrpc.NewGRPCProtocol(context.Background(), p2pHost),
	}

	return node
}

func catchError(err error) {
	if err != nil {
		panic(err)
	}
}
