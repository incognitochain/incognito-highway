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

const MaxCallRecvMsgSize = 50 << 20 // 50 MBs per gRPC response

func (hc *Client) GetBlockShardByHeight(
	ctx context.Context,
	shardID int32,
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
) (resp [][]byte, errOut error) {
	logger := Logger(ctx)

	to, heights = capBlocksPerRequest(specific, from, to, heights)
	client, pid, err := hc.getClientWithBlock(ctx, int(shardID), to)
	logger.Debugf("Requesting Shard block: shard = %v, height %v -> %v, heights = %v", shardID, from, to, heights)

	// Monitor, defer here to make sure even failed requests are logged
	defer func() {
		hc.reporter.watchRequestsPerPeer("get_block_shard", pid, errOut)
	}()

	if err != nil {
		logger.Debugf("No client with Shard block, shardID = %v, height %v -> %v, specificHeights = %v", shardID, from, to, heights)
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockShardByHeight(
		ctx,
		&proto.GetBlockShardByHeightRequest{
			Shard:      shardID,
			Specific:   specific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
		grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if reply == nil {
		return nil, errors.New("empty reply")
	}
	logger.Debugf("Reply len: %v", len(reply.Data))
	return reply.Data, nil
}

func (hc *Client) GetBlockShardByHash(
	ctx context.Context,
	shardID int32,
	hashes [][]byte,
) (resp [][]byte, errOut error) {
	logger := Logger(ctx)

	client, pid, err := hc.getClientWithHashes(int(shardID), hashes)
	logger.Debugf("Requesting Shard block: shard = %v, hashes %v ", shardID, hashes)

	// Monitor, defer here to make sure even failed requests are logged
	defer func() {
		hc.reporter.watchRequestsPerPeer("get_block_shard", pid, errOut)
	}()

	if err != nil {
		logger.Debugf("No client with Shard block hashes, shardID = %v, hashes %v", shardID, hashes)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockShardByHash(
		ctx,
		&proto.GetBlockShardByHashRequest{
			Shard:  shardID,
			Hashes: hashes,
		},
		grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if reply == nil {
		return nil, errors.New("empty reply")
	}
	logger.Debugf("Reply len: %v", len(reply.Data))
	return reply.Data, nil
}

func (hc *Client) GetBlockShardToBeaconByHeight(
	ctx context.Context,
	shardID int32,
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
) (resp [][]byte, errOut error) {
	logger := Logger(ctx)

	to, heights = capBlocksPerRequest(specific, from, to, heights)
	client, pid, err := hc.getClientWithBlock(ctx, int(shardID), to)
	logger.Debugf("Requesting S2B block: shard = %v, height %v -> %v, heights = %v", shardID, from, to, heights)

	// Monitor, defer here to make sure even failed requests are logged
	defer func() {
		hc.reporter.watchRequestsPerPeer("get_block_shard_to_beacon", pid, errOut)
	}()

	if err != nil {
		logger.Debugf("No client with S2B block, shardID = %v, from %v to %v, specificHeights = %v", shardID, from, to, heights)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockShardToBeaconByHeight(
		ctx,
		&proto.GetBlockShardToBeaconByHeightRequest{
			FromShard:  shardID,
			Specific:   specific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
		grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
	)
	if err != nil {
		logger.Warnf("err: %+v", err)
		return nil, errors.WithStack(err)
	}
	if reply == nil {
		return nil, errors.New("empty reply")
	}
	logger.Debugf("Reply len: %v", len(reply.Data))
	return reply.Data, nil
}

func (hc *Client) GetBlockCrossShardByHeight(
	ctx context.Context,
	fromShard int32,
	toShard int32,
	specific bool,
	fromHeight uint64,
	toHeight uint64,
	heights []uint64,
	fromPool bool,
) (resp [][]byte, errOut error) {
	logger := Logger(ctx)

	// NOTE: requesting crossshard block transfering PRV from `fromShard` to `toShard`
	// => request from peer of shard `fromShard`
	toHeight, heights = capBlocksPerRequest(specific, fromHeight, toHeight, heights)
	client, pid, err := hc.getClientWithBlock(ctx, int(fromShard), toHeight)
	logger.Debugf("Requesting CrossShard block: shard %v -> %v, height %v -> %v, heights = %v, pool = %v", fromShard, toShard, fromHeight, toHeight, heights, fromPool)

	// Monitor, defer here to make sure even failed requests are logged
	defer func() {
		hc.reporter.watchRequestsPerPeer("get_block_cross_shard", pid, errOut)
	}()

	if err != nil {
		logger.Debugf("No client with CrossShard block, shard %v -> %v, height %v -> %v, specificHeights = %v", fromShard, toShard, fromHeight, toHeight, heights)
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockCrossShardByHeight(
		ctx,
		&proto.GetBlockCrossShardByHeightRequest{
			FromShard:  fromShard,
			ToShard:    toShard,
			Specific:   specific,
			FromHeight: fromHeight,
			ToHeight:   toHeight,
			Heights:    heights,
			FromPool:   fromPool,
		},
		grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if reply == nil {
		return nil, errors.New("empty reply")
	}
	logger.Debugf("Reply len: %v", len(reply.Data))
	return reply.Data, nil
}

func (hc *Client) GetBlockBeaconByHeight(
	ctx context.Context,
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
) (resp [][]byte, errOut error) {
	logger := Logger(ctx)

	to, heights = capBlocksPerRequest(specific, from, to, heights)
	client, pid, err := hc.getClientWithBlock(ctx, int(common.BEACONID), to)
	logger.Debugf("Requesting Beacon block: height %v -> %v, heights = %v", from, to, heights)

	// Monitor, defer here to make sure even failed requests are logged
	defer func() {
		hc.reporter.watchRequestsPerPeer("get_block_beacon", pid, errOut)
	}()

	if err != nil {
		logger.Debugf("No client with Beacon block, height %v -> %v, specificHeights = %v", from, to, heights)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockBeaconByHeight(
		ctx,
		&proto.GetBlockBeaconByHeightRequest{
			Specific:   specific,
			FromHeight: from,
			ToHeight:   to,
			Heights:    heights,
			FromPool:   false,
		},
		grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if reply == nil {
		return nil, errors.New("empty reply")
	}
	logger.Debugf("Reply len: %v", len(reply.Data))
	return reply.Data, nil
}

func (hc *Client) GetBlockBeaconByHash(
	ctx context.Context,
	hashes [][]byte,
) (resp [][]byte, errOut error) {
	logger := Logger(ctx)

	client, pid, err := hc.getClientWithHashes(int(common.BEACONID), hashes)
	logger.Debugf("Requesting Beacon block: shard = %v, hashes %v ", int(common.BEACONID), hashes)

	// Monitor, defer here to make sure even failed requests are logged
	defer func() {
		hc.reporter.watchRequestsPerPeer("get_block_beacon", pid, errOut)
	}()

	if err != nil {
		logger.Debugf("No client with Beacon block hashes, shardID = %v, hashes %v", int(common.BEACONID), hashes)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockBeaconByHash(
		ctx,
		&proto.GetBlockBeaconByHashRequest{
			Hashes: hashes,
		},
		grpc.MaxCallRecvMsgSize(MaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if reply == nil {
		return nil, errors.New("empty reply")
	}
	logger.Debugf("Reply len: %v", len(reply.Data))
	return reply.Data, nil
}

func (hc *Client) getClientWithBlock(
	ctx context.Context,
	cid int,
	height uint64,
) (proto.HighwayServiceClient, peer.ID, error) {
	if hc.supported(cid) {
		return hc.getChainClientWithBlock(ctx, cid, height)
	}
	return hc.routeManager.GetClientSupportShard(cid)
}

// TODO replace this function, it just for fix special case in "1 HW for all"-mode.
func (hc *Client) getClientWithHashes(
	cid int,
	hashes [][]byte,
) (proto.HighwayServiceClient, peer.ID, error) {
	connectedPeers := hc.m.GetPeers(cid)
	if len(connectedPeers) == 0 {
		return nil, peer.ID(""), errors.Errorf("no route client with block for cid = %v", cid)
	}
	peerPicked := connectedPeers[rand.Intn(len(connectedPeers))]
	client, err := hc.cc.GetServiceClient(peerPicked.ID)
	return client, peerPicked.ID, err
}

func (hc *Client) getChainClientWithBlock(ctx context.Context, cid int, height uint64) (proto.HighwayServiceClient, peer.ID, error) {
	logger := Logger(ctx)

	peerID, err := hc.choosePeerIDWithBlock(ctx, cid, height)
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

func (hc *Client) choosePeerIDWithBlock(ctx context.Context, cid int, blk uint64) (peer.ID, error) {
	// logger := Logger(ctx)

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
			if p.ID == cp.ID {
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

func (hc *Client) Start() {
	for {
		select {
		case pid := <-hc.DisconnectedIDs:
			hc.cc.CloseDisconnected(pid)
		}
	}
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
		if len(heights) > common.MaxBlocksPerRequest {
			heights = heights[:common.MaxBlocksPerRequest]
		}
		return heights[len(heights)-1], heights
	}

	maxHeight := from + common.MaxBlocksPerRequest
	if to > maxHeight {
		return maxHeight, heights
	}
	return to, heights
}

type Client struct {
	DisconnectedIDs chan peer.ID

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
		m:               m,
		reporter:        reporter,
		routeManager:    rman,
		cc:              NewClientConnector(pr),
		chainData:       incChainData,
		supportShards:   supportShards,
		DisconnectedIDs: make(chan peer.ID, 1000),
	}
	go hc.Start()
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

func (cc *ClientConnector) CloseDisconnected(peerID peer.ID) {
	cc.conns.Lock()
	defer cc.conns.Unlock()

	logger.Infof("Closing connection to pID %s", peerID.String())
	if conn, ok := cc.conns.connMap[peerID]; ok {
		if err := conn.Close(); err != nil {
			logger.Warnf("Failed closing connection to pID %s: %s", peerID.String(), errors.WithStack(err))
		} else {
			delete(cc.conns.connMap, peerID)
		}
	}
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
