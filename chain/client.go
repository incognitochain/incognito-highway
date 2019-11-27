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
) ([][]byte, error) {
	to, heights = capBlocksPerRequest(specific, from, to, heights)
	client, err := hc.getClientWithBlock(int(shardID), to)
	logger.Debugf("Requesting Shard block: shard = %v, height %v -> %v, heights = %v", shardID, from, to, heights)
	if err != nil {
		logger.Debugf("No client with Shard block, shardID = %v, height %v -> %v, specificHeights = %v", shardID, from, to, heights)
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
	to, heights = capBlocksPerRequest(specific, from, to, heights)
	client, err := hc.getClientWithBlock(int(shardID), to)
	logger.Debugf("Requesting S2B block: shard = %v, height %v -> %v, heights = %v", shardID, from, to, heights)
	if err != nil {
		logger.Debugf("No client with S2B block, shardID = %v, from %v to %v, specificHeights = %v", shardID, from, to, heights)
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
) ([][]byte, error) {
	// NOTE: requesting crossshard block transfering PRV from `fromShard` to `toShard`
	// => request from peer of shard `fromShard`
	toHeight, heights = capBlocksPerRequest(specific, fromHeight, toHeight, heights)
	client, err := hc.getClientWithBlock(int(fromShard), toHeight)
	logger.Debugf("Requesting CrossShard block: shard %v -> %v, height %v -> %v, heights = %v, pool = %v", fromShard, toShard, fromHeight, toHeight, heights, fromPool)
	if err != nil {
		logger.Debugf("No client with CrossShard block, shard %v -> %v, height %v -> %v, specificHeights = %v", fromShard, toShard, fromHeight, toHeight, heights)
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
) ([][]byte, error) {
	to, heights = capBlocksPerRequest(specific, from, to, heights)
	client, err := hc.getClientWithBlock(int(common.BEACONID), to)
	logger.Debugf("Requesting Beacon block: height %v -> %v, heights = %v", from, to, heights)
	if err != nil {
		logger.Debugf("No client with Beacon block, height %v -> %v, specificHeights = %v", from, to, heights)
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
	height uint64,
) (proto.HighwayServiceClient, error) {
	if hc.supported(cid) {
		return hc.getChainClientWithBlock(cid, height)
	}
	return hc.routeManager.GetClientSupportShard(cid)
}

func (hc *Client) getChainClientWithBlock(cid int, height uint64) (proto.HighwayServiceClient, error) {
	peerID, err := hc.choosePeerIDWithBlock(cid, height)
	// logger.Debugf("Chosen peer: %v", peerID)
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
	peersHasBlk, err := hc.chainData.GetPeerHasBlk(blk, byte(cid))
	// logger.Debugf("PeersHasBlk for cid %v: %+v", cid, peersHasBlk)
	if err != nil {
		return peer.ID(""), err
	}

	// Filter out disconnected peers
	connectedPeers := hc.m.GetPeers(cid)
	// logger.Debugf("ConnectedPeers for cid %v: %+v", cid, connectedPeers)
	var peers []process.PeerWithBlk
	for _, p := range peersHasBlk {
		for _, cp := range connectedPeers {
			if p.ID == cp {
				peers = append(peers, p)
			}
		}
	}
	// logger.Debugf("PeersLeft: %+v", peers)

	// Pick randomly
	p, err := pickWeightedRandomPeer(peers, blk)
	if err != nil {
		return peer.ID(""), err
	}
	// logger.Debugf("Peer picked: %+v", p)
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

// capBlocksPerRequest returns the maximum height allowed for a single request
// If the request is for a range, this function returns the maximum block height allowed
// If the request is for some blocks, this caps the number blocks requested
func capBlocksPerRequest(specific bool, from, to uint64, heights []uint64) (uint64, []uint64) {
	if specific {
		heights = heights[:common.MaxBlocksPerRequest]
		return heights[len(heights)-1], heights
	}

	maxHeight := from + common.MaxBlocksPerRequest
	if to > maxHeight {
		return maxHeight, heights
	}
	return to, heights
}

type Client struct {
	m             *Manager
	routeManager  *route.Manager
	cc            *ClientConnector
	chainData     *process.ChainData
	supportShards []byte // to know if we should query node or other highways
}

func NewClient(
	m *Manager,
	rman *route.Manager,
	pr *p2pgrpc.GRPCProtocol,
	incChainData *process.ChainData,
	supportShards []byte,
) *Client {
	hc := &Client{
		m:             m,
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
