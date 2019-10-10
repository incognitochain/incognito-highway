package process

import (
	context "context"

	peer "github.com/libp2p/go-libp2p-peer"
	p2pgrpc "github.com/paralin/go-libp2p-grpc"
	"github.com/pkg/errors"
	grpc "google.golang.org/grpc"
)

func (hc *HighwayClient) choosePeerIDForShardBlock(shardID byte, from, to uint64) (peer.ID, error) {
	// TODO(0xakk0r0kamui): choose client from peer state
	if len(hc.peers) < 1 {
		return peer.ID(""), errors.Errorf("empty peer list for shardID %v, block %v to %v", shardID, from, to)
	}
	return hc.peers[0], nil
}

func (hc *HighwayClient) GetBlockShardByHeight(
	shardID int32,
	from uint64,
	to uint64,
) ([][]byte, error) {
	peerID, err := hc.choosePeerIDForShardBlock(byte(shardID), from, to)
	if err != nil {
		return nil, errors.WithStack(err)
	}

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
	NewPeers chan peer.ID

	cc    *ClientConnector
	peers []peer.ID
}

func NewHighwayClient() *HighwayClient {
	hc := &HighwayClient{
		NewPeers: make(chan peer.ID, 1000),
		cc:       NewClientConnector(nil),
	}
	go hc.start()
	return hc
}

func (hc *HighwayClient) start() {
	for {
		select {
		case pid := <-hc.NewPeers:
			hc.peers = append(hc.peers, pid)
		}
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
