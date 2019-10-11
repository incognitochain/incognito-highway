package process

import (
	context "context"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	peer "github.com/libp2p/go-libp2p-peer"
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

func NewHighwayClient(pr *p2pgrpc.GRPCProtocol) *HighwayClient {
	hc := &HighwayClient{
		NewPeers: make(chan peer.ID, 1000),
		cc:       NewClientConnector(pr),
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
	// TODO(@0xbunyip): check if connection is alive or not; maybe return a list of conn for HighwayClient to retry if fail to connect
	if _, ok := cc.conns[peerID]; !ok { // TODO(@0xbunyip): lock access to cc.conns
		conn, err := cc.pr.Dial(
			context.Background(),
			peerID,
			grpc.WithInsecure(),
			grpc.WithBlock(),
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		cc.conns[peerID] = conn
	}
	client := NewHighwayServiceClient(cc.conns[peerID])
	return client, nil
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
