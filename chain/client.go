package chain

import (
	context "context"
	"highway/common"
	"highway/process"
	"highway/proto"
	"sync"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	grpc "google.golang.org/grpc"
)

func (hc *Client) GetBlockByHeight(
	shardID int32,
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
	fromCandidate string,
) ([][]byte, error) {
	client, err := hc.getClientWithPublicKey(fromCandidate)
	if err != nil {
		return nil, err
	}
	if shardID != -1 {
		reply, err := client.GetBlockShardByHeight(
			context.Background(),
			&proto.GetBlockShardByHeightRequest{
				Shard:      shardID,
				Specific:   specific,
				FromHeight: from,
				ToHeight:   to,
				Heights:    heights,
				FromPool:   false,
			},
		)
		if err != nil {
			return nil, errors.WithStack(err)
		} else {
			logger.Infof("Reply: %v", reply)
			return reply.GetData(), nil
		}
	}

	reply, err := client.GetBlockBeaconByHeight(
		context.Background(),
		&proto.GetBlockBeaconByHeightRequest{
			Specific:   specific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	} else {
		logger.Infof("Reply: %v", reply)
		return reply.GetData(), nil
	}
}

func (hc *Client) getClientWithPublicKey(
	committeePublicKey string,
) (proto.HighwayServiceClient, error) {
	peerID, exist := hc.chainData.PeerIDByCommitteePubkey[committeePublicKey]
	if !exist {
		logger.Infof("Committee Publickey %v", committeePublicKey)
		PK := common.CommitteePublicKey{}
		PK.FromString(committeePublicKey)
		pkstring, _ := PK.MiningPublicKey()
		logger.Infof("Committee Publickey by mining key %v", hc.chainData.CommitteeKeyByMiningKey[pkstring])
		return nil, errors.Errorf("Can not find PeerID of this committee PublicKey %v", committeePublicKey)
	}
	client, err := hc.cc.GetServiceClient(peerID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (hc *Client) GetBlockShardByHeight(
	shardID int32,
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
) ([][]byte, error) {
	client, err := hc.getClientWithBlock(int(shardID), specific, to, heights)
	if err != nil {
		log.Warn("No client with blockshard, shardID = %v, from %v to %v", shardID, from, to)
		return nil, err
	}
	reply, err := client.GetBlockShardByHeight(
		context.Background(),
		&proto.GetBlockShardByHeightRequest{
			Shard:      shardID,
			Specific:   specific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
	)
	// logger.Infof("Reply: %v", reply)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

func (hc *Client) GetBlockShardToBeaconByHeight(
	shardID int32,
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
) ([][]byte, error) {
	client, err := hc.getClientWithBlock(int(shardID), specific, to, heights)
	if err != nil {
		log.Warn("No client with blocks2b, shardID = %v, from %v to %v", shardID, from, to)
		return nil, err
	}
	reply, err := client.GetBlockShardToBeaconByHeight(
		context.Background(),
		&proto.GetBlockShardToBeaconByHeightRequest{
			FromShard:  shardID,
			Specific:   specific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
	)
	// logger.Infof("Reply: %v", reply)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

func (hc *Client) GetBlockBeaconByHeight(
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
) ([][]byte, error) {
	client, err := hc.getClientWithBlock(int(common.BEACONID), specific, to, heights)
	if err != nil {
		log.Warn("No client with blockbeacon, from %v to %v", from, to)
		return nil, err
	}
	reply, err := client.GetBlockBeaconByHeight(
		context.Background(),
		&proto.GetBlockBeaconByHeightRequest{
			Specific:   specific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
	)
	// logger.Infof("Reply: %v", reply)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

func (hc *Client) getClientWithBlock(
	cid int,
	specific bool,
	to uint64,
	heights []uint64,
) (proto.HighwayServiceClient, error) {
	peerID := peer.ID("")
	maxHeight := to
	if specific {
		maxHeight = heights[len(heights)-1]
	}
	peerID, err := hc.choosePeerIDWithBlock(cid, maxHeight)
	// logger.Infof("Chosen peer: %v", peerID)
	if err != nil {
		return nil, err
	}

	client, err := hc.cc.GetServiceClient(peerID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (hc *Client) choosePeerIDWithBlock(cid int, blk uint64) (peer.ID, error) {
	peers := hc.m.GetPeers(cid)
	if len(peers) < 1 {
		return peer.ID(""), errors.Errorf("empty peer list for cid %v, block %v", cid, blk)
	}

	// TODO(0xbunyip): choose connected peer that has blk
	// peerID, err := hc.chainData.GetPeerHasBlk(blk, byte(cid))
	return peers[0], nil
}

type Client struct {
	m         *Manager
	cc        *ClientConnector
	chainData *process.ChainData
}

func NewClient(m *Manager, pr *p2pgrpc.GRPCProtocol, incChainData *process.ChainData) *Client {
	hc := &Client{
		m:         m,
		cc:        NewClientConnector(pr),
		chainData: incChainData,
	}
	return hc
}

func (cc *ClientConnector) GetServiceClient(peerID peer.ID) (proto.HighwayServiceClient, error) {
	// TODO(@0xbunyip): check if connection is alive or not; maybe return a list of conn for Client to retry if fail to connect
	// We might not write but still do a Lock() since we don't want to Dial to a same peerID twice
	cc.conns.Lock()
	defer cc.conns.Unlock()
	_, ok := cc.conns.connMap[peerID]

	if !ok {
		conn, err := cc.pr.Dial(
			context.Background(),
			peerID,
			grpc.WithInsecure(),
			grpc.WithBlock(),
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		cc.conns.connMap[peerID] = conn
	}
	client := proto.NewHighwayServiceClient(cc.conns.connMap[peerID])
	return client, nil
}

type ClientConnector struct {
	pr    *p2pgrpc.GRPCProtocol
	conns struct {
		connMap map[peer.ID]*grpc.ClientConn
		sync.RWMutex
	}
}

func NewClientConnector(pr *p2pgrpc.GRPCProtocol) *ClientConnector {
	connector := &ClientConnector{pr: pr}
	connector.conns.connMap = map[peer.ID]*grpc.ClientConn{}
	connector.conns.RWMutex = sync.RWMutex{}
	return connector
}
