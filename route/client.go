package route

import (
	"context"
	"highway/proto"
	"time"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func (c *Client) GetClient(peerID peer.ID) (proto.HighwayConnectorServiceClient, error) {
	// TODO(@0xbunyip): check if connection is alive or not; maybe return a list of conn for HighwayClient to retry if fail to connect
	conn, err := c.GetConnection(peerID)
	if err != nil {
		return nil, err
	}
	return proto.NewHighwayConnectorServiceClient(conn), nil
}

func (c *Client) GetConnection(peerID peer.ID) (*grpc.ClientConn, error) {
	if _, ok := c.conns[peerID]; !ok { // TODO(@0xbunyip): lock access to c.conns
		conn, err := c.pr.Dial(
			context.Background(),
			peerID,
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithTimeout(3*time.Second),
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		c.conns[peerID] = conn
	}
	return c.conns[peerID], nil
}

type Client struct {
	pr    *p2pgrpc.GRPCProtocol
	conns map[peer.ID]*grpc.ClientConn
}

func NewClient(pr *p2pgrpc.GRPCProtocol) *Client {
	return &Client{
		pr:    pr,
		conns: map[peer.ID]*grpc.ClientConn{},
	}
}
