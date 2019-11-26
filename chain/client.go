package chain

import (
	context "context"
	"highway/common"
	"highway/process"
	"highway/proto"
	"highway/route"
	"math/rand"
	"sync"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
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
) (resp [][]byte, errOut error) {
	client, pid, err := hc.getClientWithBlock(int(shardID), specific, to, heights)

	// Monitor, defer here to make sure even failed requests are logged
	logger.Infof("Defering get_block_shard")
	defer hc.reporter.watchRequestsPerPeer("get_block_shard", pid, errOut)

	if err != nil {
		logger.Debugf("No client with blockshard, shardID = %v, from %v to %v", shardID, from, to)
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
) (resp [][]byte, errOut error) {
	client, pid, err := hc.getClientWithBlock(int(shardID), specific, to, heights)

	// Monitor, defer here to make sure even failed requests are logged
	defer hc.reporter.watchRequestsPerPeer("get_block_shard_to_beacon", pid, errOut)

	if err != nil {
		logger.Debugf("No client with blocks2b, shardID = %v, from %v to %v", shardID, from, to)
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

func (hc *Client) GetBlockCrossShardByHeight(
	fromShard int32,
	toShard int32,
	specific bool,
	fromHeight uint64,
	toHeight uint64,
	heights []uint64,
	fromPool bool,
) (resp [][]byte, errOut error) {
	logger.Debugf("Requesting CrossShard block shard %v -> %v, height %v -> %v, %v, pool: %v", fromShard, toShard, fromHeight, toHeight, heights, fromPool)

	// NOTE: requesting crossshard block transfering PRV from `fromShard` to `toShard`
	// => request from peer of shard `fromShard`
	client, pid, err := hc.getClientWithBlock(int(fromShard), specific, toHeight, heights)

	// Monitor, defer here to make sure even failed requests are logged
	defer hc.reporter.watchRequestsPerPeer("get_block_cross_shard", pid, errOut)

	if err != nil {
		logger.Debugf("No client with blockCS, shard from = %v, to = %v, height from = %v, to = %v, heights = %v", fromShard, toShard, fromHeight, toHeight, heights)
		return nil, err
	}
	reply, err := client.GetBlockCrossShardByHeight(
		context.Background(),
		&proto.GetBlockCrossShardByHeightRequest{
			FromShard:  fromShard,
			ToShard:    toShard,
			Specific:   specific,
			FromHeight: fromHeight,
			ToHeight:   toHeight,
			Heights:    heights,
			FromPool:   fromPool,
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
) (resp [][]byte, errOut error) {
	client, pid, err := hc.getClientWithBlock(int(common.BEACONID), specific, to, heights)

	// Monitor, defer here to make sure even failed requests are logged
	logger.Infof("Defering get_block_beacon")
	defer hc.reporter.watchRequestsPerPeer("get_block_beacon", pid, errOut)

	if err != nil {
		logger.Debugf("No client with blockbeacon, from %v to %v", from, to)
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
) (proto.HighwayServiceClient, peer.ID, error) {
	if hc.supported(cid) {
		return hc.getChainClientWithBlock(cid, specific, to, heights)
	}
	return hc.routeManager.GetClientSupportShard(cid)
}

func (hc *Client) getChainClientWithBlock(
	cid int,
	specific bool,
	to uint64,
	heights []uint64,
) (proto.HighwayServiceClient, peer.ID, error) {
	maxHeight := to
	if specific {
		maxHeight = heights[len(heights)-1]
	}
	peerID, err := hc.choosePeerIDWithBlock(cid, maxHeight)
	logger.Debugf("Chosen peer: %v", peerID)
	if err != nil {
		return nil, peerID, err
	}

	client, err := hc.cc.GetServiceClient(peerID)
	if err != nil {
		return nil, peerID, err
	}
	return client, peerID, nil
}

func (hc *Client) choosePeerIDWithBlock(cid int, blk uint64) (peer.ID, error) {
	peersHasBlk, err := hc.chainData.GetPeerHasBlk(blk, byte(cid))
	logger.Debugf("PeersHasBlk for cid %v: %+v", cid, peersHasBlk)
	if err != nil {
		return peer.ID(""), err
	}

	// Filter out disconnected peers
	connectedPeers := hc.m.GetPeers(cid)
	logger.Debugf("ConnectedPeers for cid %v: %+v", cid, connectedPeers)
	var peers []process.PeerWithBlk
	for _, p := range peersHasBlk {
		for _, cp := range connectedPeers {
			if p.ID == cp.ID {
				peers = append(peers, p)
			}
		}
	}
	logger.Debugf("PeersLeft: %+v", peers)

	// Pick randomly
	p, err := pickWeightedRandomPeer(peers, blk)
	if err != nil {
		return peer.ID(""), err
	}
	logger.Debugf("Peer picked: %+v", p)
	return p.ID, nil
}

func (hc *Client) supported(cid int) bool {
	for _, s := range hc.supportShards {
		if byte(cid) == s {
			return true
		}
	}
	return false
}

func pickWeightedRandomPeer(peers []process.PeerWithBlk, blk uint64) (process.PeerWithBlk, error) {
	if len(peers) == 0 {
		return process.PeerWithBlk{}, errors.Errorf("empty peer list")
	}

	// Find peers have all the blocks
	last := -1
	for i, p := range peers {
		if p.Height < blk {
			break
		}
		last = i
	}

	end := last + 1 // Pick only from peers with all the blocks
	if last <= 0 {
		end = len(peers) // Otherwise, pick randomly from all peers
	}
	return peers[rand.Intn(end)], nil
}

type Client struct {
	m             *Manager
	reporter      *Reporter
	routeManager  *route.Manager
	cc            *ClientConnector
	chainData     *process.ChainData
	supportShards []byte // to know if we should query node or other highways
}

func NewClient(
	m *Manager,
	reporter *Reporter,
	rman *route.Manager,
	pr *p2pgrpc.GRPCProtocol,
	incChainData *process.ChainData,
	supportShards []byte,
) *Client {
	hc := &Client{
		m:             m,
		reporter:      reporter,
		routeManager:  rman,
		cc:            NewClientConnector(pr),
		chainData:     incChainData,
		supportShards: supportShards,
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
