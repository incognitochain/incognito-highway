package main

import (
	"context"
	"fmt"
	"highway/process"
	"time"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func (c *HWClient) GetClient(peerID peer.ID) (process.HighwayConnectorServiceClient, error) {
	fmt.Println("GetClient", peerID)
	// TODO(@0xbunyip): check if connection is alive or not; maybe return a list of conn for HighwayClient to retry if fail to connect
	if _, ok := c.conns[peerID]; !ok { // TODO(@0xbunyip): lock access to c.conns
		conn, err := c.pr.Dial(
			context.Background(),
			peerID,
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithTimeout(3*time.Second),
		)
		fmt.Println("GetClient", conn, err)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		c.conns[peerID] = conn
	}
	client := process.NewHighwayConnectorServiceClient(c.conns[peerID])
	return client, nil
}

type HWClient struct {
	pr    *p2pgrpc.GRPCProtocol
	conns map[peer.ID]*grpc.ClientConn
}

func NewHWClient(pr *p2pgrpc.GRPCProtocol) *HWClient {
	return &HWClient{
		pr:    pr,
		conns: map[peer.ID]*grpc.ClientConn{},
	}
}
