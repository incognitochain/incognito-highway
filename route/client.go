package route

import (
	"context"
	"highway/common"
	"highway/proto"
	"sync"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
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
	// We might not write but still do a Lock() since we don't want to Dial to a same peerID twice
	c.conns.Lock()
	defer c.conns.Unlock()
	if _, ok := c.conns.connMap[peerID]; !ok {
		ctx, cancel := context.WithTimeout(context.Background(), common.RouteClientDialTimeout)
		defer cancel()
		conn, err := c.pr.Dial(
			ctx,
			peerID,
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:    common.RouteClientKeepaliveTime,
				Timeout: common.RouteClientKeepaliveTimeout,
			}),
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		c.conns.connMap[peerID] = conn
	}
	return c.conns.connMap[peerID], nil
}

func (c *Client) CloseConnection(peerID peer.ID) error {
	c.conns.Lock()
	defer c.conns.Unlock()
	if conn, ok := c.conns.connMap[peerID]; ok {
		err := conn.Close()
		if err != nil {
			return errors.WithStack(err)
		}
		c.conns.connMap[peerID] = nil
	}
	return nil
}

type Client struct {
	pr    *p2pgrpc.GRPCProtocol
	conns struct {
		connMap map[peer.ID]*grpc.ClientConn
		*sync.RWMutex
	}
}

func NewClient(pr *p2pgrpc.GRPCProtocol) *Client {
	client := &Client{pr: pr}
	client.conns.connMap = map[peer.ID]*grpc.ClientConn{}
	client.conns.RWMutex = &sync.RWMutex{}
	return client
}
