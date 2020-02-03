package chain

import (
	context "context"
	"highway/chaindata"
	"highway/common"
	"highway/proto"
	"highway/route"
	"math/rand"
	"sync"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const MaxCallRecvMsgSize = 50 << 20 // 50 MBs per gRPC response

func (hc *Client) GetBlockShardByHeight(
	ctx context.Context,
	shardID int32,
	specific bool,
	from uint64,
	to uint64,
	heights []uint64,
	callDepth int32,
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
		logger.Debugf("No client with Shard block, shardID = %v, height %v -> %v, specificHeights = %v, err = %+v", shardID, from, to, heights, err)
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
			CallDepth:  callDepth + 1,
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
	callDepth int32,
) (resp [][]byte, errOut error) {
	logger := Logger(ctx)

	client, pid, err := hc.getClientWithHashes(int(shardID), hashes)
	logger.Debugf("Requesting Shard block: shard = %v, hashes %v ", shardID, hashes)

	// Monitor, defer here to make sure even failed requests are logged
	defer func() {
		hc.reporter.watchRequestsPerPeer("get_block_shard", pid, errOut)
	}()

	if err != nil {
		logger.Debugf("No client with Shard block hashes, shardID = %v, hashes %v, err = %+v", shardID, hashes, err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockShardByHash(
		ctx,
		&proto.GetBlockShardByHashRequest{
			Shard:     shardID,
			Hashes:    hashes,
			CallDepth: callDepth + 1,
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
	callDepth int32,
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
		logger.Debugf("No client with S2B block, shardID = %v, from %v to %v, specificHeights = %v, err = %+v", shardID, from, to, heights, err)
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
			CallDepth:  callDepth + 1,
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
	callDepth int32,
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
		logger.Debugf("No client with CrossShard block, shard %v -> %v, height %v -> %v, specificHeights = %v, err = %+v", fromShard, toShard, fromHeight, toHeight, heights, err)
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
			CallDepth:  callDepth + 1,
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
	callDepth int32,
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
		logger.Debugf("No client with Beacon block, height %v -> %v, specificHeights = %v, err = %+v", from, to, heights, err)
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
			CallDepth:  callDepth + 1,
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
	callDepth int32,
) (resp [][]byte, errOut error) {
	logger := Logger(ctx)

	client, pid, err := hc.getClientWithHashes(int(common.BEACONID), hashes)
	logger.Debugf("Requesting Beacon block: shard = %v, hashes %v ", int(common.BEACONID), hashes)

	// Monitor, defer here to make sure even failed requests are logged
	defer func() {
		hc.reporter.watchRequestsPerPeer("get_block_beacon", pid, errOut)
	}()

	if err != nil {
		logger.Debugf("No client with Beacon block hashes, shardID = %v, hashes %v, err = %+v", int(common.BEACONID), hashes, err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	reply, err := client.GetBlockBeaconByHash(
		ctx,
		&proto.GetBlockBeaconByHashRequest{
			Hashes:    hashes,
			CallDepth: callDepth + 1,
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
		return hc.getClientOfSupportedShard(ctx, cid, height)
	}
	return hc.routeManager.GetClientSupportShard(cid)
}

// TODO(@0xakk0r0kamui) replace this function, it just for fix special case in "1 HW for all"-mode.
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

// getClientOfSupportedShard returns a client (node or another highway)
// that has the needed block height
// This func prioritizes getting from a node to reduce load to highways
func (hc *Client) getClientOfSupportedShard(ctx context.Context, cid int, height uint64) (proto.HighwayServiceClient, peer.ID, error) {
	logger := Logger(ctx)

	peerID, hw, err := hc.choosePeerIDWithBlock(ctx, cid, height)
	logger.Debugf("Chosen peer: %s %s", peerID.String(), hw.String())
	if err != nil {
		return nil, peerID, err
	}

	if hw != hc.routeManager.ID { // Peer not connected, let ask the other highway
		logger.Debugf("Chosen peer not connected, connect to hw %s", hw.String())
		return hc.routeManager.GetHighwayServiceClient(hw)
	}

	// Connected peer, get connection
	client, err := hc.cc.GetServiceClient(peerID)
	if err != nil {
		return nil, peerID, err
	}
	return client, peerID, nil
}

// choosePeerIDWithBlock returns peerID of a node that holds some blocks
// and its corresponding highway's peerID
func (hc *Client) choosePeerIDWithBlock(ctx context.Context, cid int, blk uint64) (pid peer.ID, hw peer.ID, err error) {
	// logger := Logger(ctx)

	peersHasBlk, err := hc.chainData.GetPeerHasBlk(blk, byte(cid)) // Get all peers from peerstate
	logger.Debugf("PeersHasBlk for cid %v blk %v: %+v", cid, blk, peersHasBlk)
	// logger.Debugf("PeersHasBlk for cid %v: %+v", cid, peersHasBlk)
	if err != nil {
		return peer.ID(""), peer.ID(""), err
	}
	if len(peersHasBlk) == 0 {
		return peer.ID(""), peer.ID(""), errors.Errorf("no peer with blk %d %d", cid, blk)
	}

	// Prioritize peers and sort into different groups
	connectedPeers := hc.m.GetPeers(cid) // Filter out disconnected peers
	groups := groupPeersByDistance(peersHasBlk, blk, hc.routeManager.ID, connectedPeers)
	logger.Debugf("Peers by groups: %+v", groups)

	// Choose a single peer from the sorted groups
	p, err := choosePeerFromGroup(groups)
	if err != nil {
		return peer.ID(""), peer.ID(""), errors.WithMessagef(err, "groups: %+v", groups)
	}

	// logger.Debugf("Peer picked: %+v", p)
	return p.ID, p.HW, nil
}

// groupPeersByDistance prioritizes peers by grouping them into
// different groups based on their distance to this highway
func groupPeersByDistance(
	peers []chaindata.PeerWithBlk,
	blk uint64,
	selfPeerID peer.ID,
	connectedPeers []PeerInfo,
) [][]chaindata.PeerWithBlk {
	// Group peers into 4 groups:
	a := []chaindata.PeerWithBlk{} // 1. Nodes connected to this highway and have all needed blocks
	b := []chaindata.PeerWithBlk{} // 2. Nodes from other highways and have all needed blocks
	h := uint64(0)                 // Find maximum height
	for _, p := range peers {
		if p.Height >= blk {
			if p.HW == selfPeerID {
				a = append(a, p)
			} else {
				b = append(b, p)
			}
		}
		if p.Height > h {
			h = p.Height
		}
	}
	a = filterPeers(a, connectedPeers) // Retain only connected peers

	c := []chaindata.PeerWithBlk{} // 3. Nodes connected to this highway and have the largest amount of blocks
	d := []chaindata.PeerWithBlk{} // 4. Nodes from other highways and have the largest amount of blocks
	for _, p := range peers {
		if p.Height < blk && p.Height+common.ChoosePeerBlockDelta >= h {
			if p.HW == selfPeerID {
				c = append(c, p)
			} else {
				d = append(d, p)
			}
		}
	}
	c = filterPeers(c, connectedPeers) // Retain only connected peers
	return [][]chaindata.PeerWithBlk{a, b, c, d}
}

func choosePeerFromGroup(groups [][]chaindata.PeerWithBlk) (chaindata.PeerWithBlk, error) {
	// Pick randomly
	for _, group := range groups {
		if len(group) > 0 {
			return group[rand.Intn(len(group))], nil
		}
	}
	return chaindata.PeerWithBlk{}, errors.New("no group of peers to choose")
}

func filterPeers(allPeers []chaindata.PeerWithBlk, allows []PeerInfo) []chaindata.PeerWithBlk {
	// logger.Debugf("ConnectedPeers for cid %v: %+v", cid, connectedPeers)
	var peers []chaindata.PeerWithBlk
	for _, p := range allPeers {
		for _, a := range allows {
			if p.ID == a.ID {
				peers = append(peers, p)
				break
			}
		}
	}
	// logger.Debugf("PeersLeft: %+v", peers)
	return peers
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
	chainData     *chaindata.ChainData
	supportShards []byte // to know if we should query node or other highways
}

func NewClient(
	m *Manager,
	reporter *Reporter,
	rman *route.Manager,
	pr *p2pgrpc.GRPCProtocol,
	incChainData *chaindata.ChainData,
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
		ctx, cancel := context.WithTimeout(context.Background(), common.ChainClientDialTimeout)
		defer cancel()
		conn, err := cc.pr.Dial(
			ctx,
			peerID,
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:    common.ChainClientKeepaliveTime,
				Timeout: common.ChainClientKeepaliveTimeout,
			}),
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

	if conn, ok := cc.conns.connMap[peerID]; ok {
		logger.Infof("Closing connection to pID %s", peerID.String())
		if err := conn.Close(); err != nil {
			logger.Warnf("Failed closing connection to pID %s: %s", peerID.String(), errors.WithStack(err))
		} else {
			delete(cc.conns.connMap, peerID)
			logger.Infof("Closed connection to pID %s successfully", peerID.String())
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
