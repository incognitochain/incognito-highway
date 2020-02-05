package chain

import (
	context "context"
	"highway/chaindata"
	"highway/common"
	"highway/proto"
	"math/rand"
	"sync"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func (hc *Client) GetBlockByHeight(
	ctx context.Context,
	req requestByHeight,
	heights []uint64,
) (resp [][]byte, errOut error) {
	logger := Logger(ctx)

	serviceClient, pid, err := hc.getClientWithBlock(ctx, int(req.fromShard), heights[len(heights)-1])
	logger.Debugf("Requesting block by height: shard %v -> %v, heights = %v", req.fromShard, req.toShard, heights)

	// Monitor, defer here to make sure even failed requests are logged
	defer func() {
		hc.reporter.watchRequestsPerPeer("get_by_height", pid, errOut)
	}()

	if err != nil {
		logger.Debugf("No serviceClient with block, shardID = %v, heights = %v, err = %+v", req.fromShard, heights, err)
		return nil, err
	}

	data, err := getBlockByHeight(serviceClient, req, heights)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Data len: %v", len(data))
	return data, nil
}

func getBlockByHeight(serviceClient proto.HighwayServiceClient, req requestByHeight, heights []uint64) ([][]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()

	var data [][]byte
	var err error
	crossShard := req.fromShard != req.toShard
	beacon := byte(req.fromShard) == common.BEACONID || byte(req.toShard) == common.BEACONID
	if !crossShard {
		if !beacon {
			data, err = getBlockShardByHeight(ctx, serviceClient, req, heights)
		} else {
			data, err = getBlockBeaconByHeight(ctx, serviceClient, req, heights)
		}
	} else {
		if !beacon {
			data, err = getBlockCrossShardByHeight(ctx, serviceClient, req, heights)
		} else {
			data, err = getBlockShardToBeaconByHeight(ctx, serviceClient, req, heights)
		}
	}

	if err != nil {
		return nil, err
	}
	return data, nil
}

func getBlockShardByHeight(
	ctx context.Context,
	serviceClient proto.HighwayServiceClient,
	req requestByHeight,
	heights []uint64,
) ([][]byte, error) {
	reply, err := serviceClient.GetBlockShardByHeight(
		ctx,
		&proto.GetBlockShardByHeightRequest{
			Shard:     req.fromShard,
			Specific:  true,
			Heights:   heights,
			FromPool:  false,
			CallDepth: req.callDepth + 1,
		},
		grpc.MaxCallRecvMsgSize(common.ChainMaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

func getBlockBeaconByHeight(
	ctx context.Context,
	serviceClient proto.HighwayServiceClient,
	req requestByHeight,
	heights []uint64,
) ([][]byte, error) {
	reply, err := serviceClient.GetBlockBeaconByHeight(
		ctx,
		&proto.GetBlockBeaconByHeightRequest{
			Specific:  true,
			Heights:   heights,
			FromPool:  false,
			CallDepth: req.callDepth + 1,
		},
		grpc.MaxCallRecvMsgSize(common.ChainMaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

func getBlockCrossShardByHeight(
	ctx context.Context,
	serviceClient proto.HighwayServiceClient,
	req requestByHeight,
	heights []uint64,
) ([][]byte, error) {
	reply, err := serviceClient.GetBlockCrossShardByHeight(
		ctx,
		&proto.GetBlockCrossShardByHeightRequest{
			FromShard: req.fromShard,
			ToShard:   req.toShard,
			Heights:   heights,
			FromPool:  req.fromPool,
			CallDepth: req.callDepth + 1,
		},
		grpc.MaxCallRecvMsgSize(common.ChainMaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

func getBlockShardToBeaconByHeight(
	ctx context.Context,
	serviceClient proto.HighwayServiceClient,
	req requestByHeight,
	heights []uint64,
) ([][]byte, error) {
	reply, err := serviceClient.GetBlockShardToBeaconByHeight(
		ctx,
		&proto.GetBlockShardToBeaconByHeightRequest{
			FromShard: req.fromShard,
			Heights:   heights,
			FromPool:  false,
			CallDepth: req.callDepth + 1,
		},
		grpc.MaxCallRecvMsgSize(common.ChainMaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

func (hc *Client) GetBlockByHash(
	ctx context.Context,
	req requestByHash,
	hashes [][]byte,
) (resp [][]byte, errOut error) {
	logger := Logger(ctx)

	serviceClient, pid, err := hc.getClientWithHashes(int(req.shard), hashes)
	logger.Debugf("Requesting block by hash: shard = %v, hashes %v ", req.shard, hashes)

	// Monitor, defer here to make sure even failed requests are logged
	defer func() {
		hc.reporter.watchRequestsPerPeer("get_by_hash", pid, errOut)
	}()

	if err != nil {
		logger.Debugf("No client with block hashes, shardID = %v, hashes %v, err = %+v", req.shard, hashes, err)
		return nil, err
	}

	data, err := getBlockByHash(serviceClient, req, hashes)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Data len: %v", len(data))
	return data, nil
}

func getBlockByHash(serviceClient proto.HighwayServiceClient, req requestByHash, hashes [][]byte) ([][]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()

	var data [][]byte
	var err error
	beacon := byte(req.shard) == common.BEACONID
	if !beacon {
		data, err = getBlockShardByHash(ctx, serviceClient, req, hashes)
	} else {
		data, err = getBlockBeaconByHash(ctx, serviceClient, req, hashes)
	}

	if err != nil {
		return nil, err
	}
	return data, nil
}

func getBlockShardByHash(
	ctx context.Context,
	serviceClient proto.HighwayServiceClient,
	req requestByHash,
	hashes [][]byte,
) ([][]byte, error) {
	reply, err := serviceClient.GetBlockShardByHash(
		ctx,
		&proto.GetBlockShardByHashRequest{
			Shard:     req.shard,
			Hashes:    hashes,
			CallDepth: req.callDepth + 1,
		},
		grpc.MaxCallRecvMsgSize(common.ChainMaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

func getBlockBeaconByHash(
	ctx context.Context,
	serviceClient proto.HighwayServiceClient,
	req requestByHash,
	hashes [][]byte,
) ([][]byte, error) {
	reply, err := serviceClient.GetBlockBeaconByHash(
		ctx,
		&proto.GetBlockBeaconByHashRequest{
			Hashes:    hashes,
			CallDepth: req.callDepth + 1,
		},
		grpc.MaxCallRecvMsgSize(common.ChainMaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
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
	return hc.router.GetClientSupportShard(cid)
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

	if hw != hc.router.GetID() { // Peer not connected, let ask the other highway
		logger.Debugf("Chosen peer not connected, connect to hw %s", hw.String())
		return hc.router.GetHighwayServiceClient(hw)
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
	logger := Logger(ctx)

	peersHasBlk, err := hc.peerStore.GetPeerHasBlk(blk, byte(cid)) // Get all peers from peerstate
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
	groups := groupPeersByDistance(peersHasBlk, blk, hc.router.GetID(), connectedPeers)
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

type Client struct {
	DisconnectedIDs chan peer.ID

	m             *Manager
	reporter      *Reporter
	router        Router
	cc            *ClientConnector
	peerStore     PeerStore
	supportShards []byte // to know if we should query node or other highways
}

func NewClient(
	m *Manager,
	reporter *Reporter,
	router Router,
	pr *p2pgrpc.GRPCProtocol,
	peerStore PeerStore,
	supportShards []byte,
) *Client {
	hc := &Client{
		m:               m,
		reporter:        reporter,
		router:          router,
		cc:              NewClientConnector(pr),
		peerStore:       peerStore,
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
		conn, err := cc.dialer.Dial(
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
	dialer Dialer
	conns  struct {
		connMap map[peer.ID]*grpc.ClientConn
		sync.RWMutex
	}
}

func NewClientConnector(dialer Dialer) *ClientConnector {
	connector := &ClientConnector{dialer: dialer}
	connector.conns.connMap = map[peer.ID]*grpc.ClientConn{}
	connector.conns.RWMutex = sync.RWMutex{}
	return connector
}

type PeerStore interface {
	GetPeerHasBlk(blkHeight uint64, committeeID byte) ([]chaindata.PeerWithBlk, error)
}

type Dialer interface {
	Dial(ctx context.Context, peerID peer.ID, dialOpts ...grpc.DialOption) (*grpc.ClientConn, error)
}

type Router interface {
	GetClientSupportShard(cid int) (proto.HighwayServiceClient, peer.ID, error)
	GetHighwayServiceClient(pid peer.ID) (proto.HighwayServiceClient, peer.ID, error)
	GetID() peer.ID
}

type requestByHeight struct {
	fromShard int32
	toShard   int32
	callDepth int32
	fromPool  bool
}

type getBlockByHeightRequest interface {
	GetCallDepth() int32
	GetFromPool() bool
}

func ParseGetBlockByHeight(inp getBlockByHeightRequest) requestByHeight {
	req := requestByHeight{}
	req.callDepth = inp.GetCallDepth()
	req.fromPool = inp.GetFromPool()
	return req
}

func ParseGetBlockShardByHeight(inp *proto.GetBlockShardByHeightRequest) requestByHeight {
	req := ParseGetBlockByHeight(inp)
	req.fromShard = inp.Shard
	req.toShard = inp.Shard
	return req
}

func ParseGetBlockBeaconByHeight(inp *proto.GetBlockBeaconByHeightRequest) requestByHeight {
	req := ParseGetBlockByHeight(inp)
	req.fromShard = int32(common.BEACONID)
	req.toShard = int32(common.BEACONID)
	return req
}

func ParseGetBlockCrossShardByHeight(inp *proto.GetBlockCrossShardByHeightRequest) requestByHeight {
	// NOTE: requesting crossshard block transfering PRV from `fromShard` to `toShard`
	// => request from peer of shard `fromShard`
	req := ParseGetBlockByHeight(inp)
	req.fromShard = inp.FromShard
	req.toShard = inp.ToShard
	return req
}

func ParseGetBlockShardToBeaconByHeight(inp *proto.GetBlockShardToBeaconByHeightRequest) requestByHeight {
	req := ParseGetBlockByHeight(inp)
	req.fromShard = inp.FromShard
	req.toShard = int32(common.BEACONID)
	return req
}

type requestByHash struct {
	shard     int32
	callDepth int32
}

func ParseGetBlockShardByHash(inp *proto.GetBlockShardByHashRequest) requestByHash {
	req := requestByHash{}
	req.shard = inp.Shard
	req.callDepth = inp.CallDepth
	return req
}

func ParseGetBlockBeaconByHash(inp *proto.GetBlockBeaconByHashRequest) requestByHash {
	req := requestByHash{}
	req.shard = int32(common.BEACONID)
	req.callDepth = inp.CallDepth
	return req
}
