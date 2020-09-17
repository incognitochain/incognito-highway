package p2p

import (
	"context"
	crypto2 "crypto"
	"fmt"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Peer struct {
	TargetAddress []core.Multiaddr
	PeerID        peer.ID
	PublicKey     crypto2.PublicKey
}

type Host struct {
	Version  string
	Host     host.Host
	SelfPeer *Peer
	GRPC     *p2pgrpc.GRPCProtocol
}

func NewHost(version string, listenAddr string, listenPort int, privKey crypto.PrivKey) *Host {
	tracer := opentracing.GlobalTracer()
	// var privKey crypto.PrivKey
	if privKey == nil {
		privKey, _, _ = crypto.GenerateKeyPair(crypto.ECDSA, 2048)
		m, _ := crypto.MarshalPrivateKey(privKey)
		encoded := crypto.ConfigEncodeKey(m)
		fmt.Println("encoded libp2p key:", encoded)
	}
	// else {
	// 	b, err := crypto.ConfigDecodeKey(privKeyStr)
	// 	catchError(err)
	// 	privKey, err = crypto.UnmarshalPrivateKey(b)
	// 	catchError(err)
	// }

	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", listenAddr, listenPort))
	catchError(err)

	ctx := context.Background()
	opts := []libp2p.Option{
		libp2p.ConnectionManager(nil),
		libp2p.ListenAddrs(addr),
		libp2p.Identity(privKey),
	}

	p2pHost, err := libp2p.New(ctx, opts...)
	catchError(err)

	selfPeer := &Peer{
		PeerID:        p2pHost.ID(),
		TargetAddress: append([]multiaddr.Multiaddr{}, addr),
	}

	kasp := keepalive.ServerParameters{
		MaxConnectionIdle: GRPCMaxConnectionIdle,
		Time:              GRPCTime,
		Timeout:           GRPCTimeout,
	}
	node := &Host{
		Host:     p2pHost,
		SelfPeer: selfPeer,
		Version:  version,
		GRPC: p2pgrpc.NewGRPCProtocol(
			context.Background(),
			p2pHost,
			grpc.KeepaliveParams(kasp),
			grpc.UnaryInterceptor(
				otgrpc.OpenTracingServerInterceptor(tracer)),
			grpc.StreamInterceptor(
				otgrpc.OpenTracingStreamServerInterceptor(tracer))),
	}

	return node
}

func catchError(err error) {
	if err != nil {
		panic(err)
	}
}
