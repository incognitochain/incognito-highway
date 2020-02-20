package rpcserver

import (
	"net"
	"net/rpc"
	"time"

	"github.com/pkg/errors"
)

type RPCClient struct{}

func (rpcClient *RPCClient) DiscoverHighway(
	discoverPeerAddress string,
	shardsStr []string,
) (
	map[string][]HighwayAddr,
	error,
) {
	if discoverPeerAddress == "" {
		return nil, errors.Errorf("empty address")
	}
	logger.Info("Dialing...")
	conn, err := net.DialTimeout("tcp", discoverPeerAddress, 10*time.Second)
	if err != nil {
		return nil, errors.WithMessagef(err, "fail to connect to discover peer %v", discoverPeerAddress)
	}

	logger.Infof("Connected to %v", discoverPeerAddress)
	client := rpc.NewClient(conn)
	defer client.Close()

	req := Request{Shard: shardsStr}
	var res Response
	logger.Infof("Start dialing RPC server with param %v", req)

	err = client.Call("Handler.GetPeers", req, &res)

	if err != nil {
		return nil, errors.Errorf("Call Handler.GetPeers return error %v", err)
	}
	logger.Infof("Bootnode return %v", res.PeerPerShard)
	return res.PeerPerShard, nil
}
