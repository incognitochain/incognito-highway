package process

import (
	context "context"

	peer "github.com/libp2p/go-libp2p-peer"
	p2pgrpc "github.com/paralin/go-libp2p-grpc"
	"github.com/pkg/errors"
	grpc "google.golang.org/grpc"
)

func (hc *HighwayClient) GetBlockShardByHeight(
	peerID peer.ID,
	shardID int32,
	from uint64,
	to uint64,
) ([][]byte, error) {
	client, err := hc.cc.GetServiceClient(peerID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	reply, err := client.GetBlockShardByHeight(
		context.Background(),
		&GetBlockShardByHeightRequest{
			Shard:      shardID,
			Specific:   false,
			FromHeight: from,
			ToHeight:   to,
			Heights:    nil,
			FromPool:   false,
		},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

type HighwayClient struct {
	cc *ClientConnector
}

func NewHighwayClient() *HighwayClient {
	return &HighwayClient{
		cc: nil,
	}
}

func (cc *ClientConnector) GetServiceClient(peerID peer.ID) (HighwayServiceClient, error) {
	return nil, nil
}

type ClientConnector struct {
	pr    *p2pgrpc.GRPCProtocol
	conns map[peer.ID]*grpc.ClientConn
}

func NewClientConnector(pr *p2pgrpc.GRPCProtocol) *ClientConnector {
	return &ClientConnector{
		pr:    pr,
		conns: map[peer.ID]*grpc.ClientConn{},
	}
}
