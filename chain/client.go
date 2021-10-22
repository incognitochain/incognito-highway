package chain

import (
	context "context"
	"highway/chaindata"
	"highway/common"
	"highway/proto"
	"math/rand"
	"sync"
	"time"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func (hc *Client) GetBlockByHash(
	ctx context.Context,
	req GetBlockByHashRequest,
	hashes [][]byte,
) (resp [][]byte, errOut error) {
	logger := Logger(ctx)

	serviceClient, pid, err := hc.getClientWithHashes(int(req.GetCID()), hashes)
	logger.Debugf("Requesting block by hash: shard = %v, hashes %v ", req.GetCID(), hashes)

	// Monitor, defer here to make sure even failed requests are logged
	defer func() {
		hc.reporter.watchRequestsPerPeer("get_by_hash", pid, errOut)
	}()

	if err != nil {
		logger.Debugf("No client with block hashes, shardID = %v, hashes %v, err = %+v", req.GetCID(), hashes, err)
		return nil, err
	}

	data, err := getBlockByHash(serviceClient, req, hashes)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Data len: %v", len(data))
	return data, nil
}

func getBlockByHash(serviceClient proto.HighwayServiceClient, req GetBlockByHashRequest, hashes [][]byte) ([][]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()

	var data [][]byte
	var err error
	beacon := byte(req.GetCID()) == common.BEACONID
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
	req GetBlockByHashRequest,
	hashes [][]byte,
) ([][]byte, error) {
	reply, err := serviceClient.GetBlockShardByHash(
		ctx,
		&proto.GetBlockShardByHashRequest{
			Shard:     req.GetCID(),
			Hashes:    hashes,
			CallDepth: req.GetCallDepth() + 1,
			UUID:      req.GetUUID(),
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
	req GetBlockByHashRequest,
	hashes [][]byte,
) ([][]byte, error) {
	reply, err := serviceClient.GetBlockBeaconByHash(
		ctx,
		&proto.GetBlockBeaconByHashRequest{
			Hashes:    hashes,
			CallDepth: req.GetCallDepth() + 1,
			UUID:      req.GetUUID(),
		},
		grpc.MaxCallRecvMsgSize(common.ChainMaxCallRecvMsgSize),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply.Data, nil
}

func (hc *Client) SetBlockByHeight(_ context.Context, _ GetBlockByHeightRequest, _ []uint64, _ [][]byte) error {
	// Client no needs to cache block
	return nil
}

func (hc *Client) SetSingleBlockByHeight(
	_ context.Context,
	_ RequestBlockByHeight,
	_ common.ExpectedBlkByHeight,
) error {
	// Client no needs to cache block
	return nil
}

func (hc *Client) SetSingleBlockByHeightv2(
	_ context.Context,
	_ RequestBlockByHeight,
	_ common.ExpectedBlk,
) error {
	// Client no needs to cache block
	return nil
}

func (hc *Client) SetSingleBlockByHash(
	_ context.Context,
	_ RequestBlockByHash,
	_ common.ExpectedBlk,
) error {
	// Client no needs to cache block
	return nil
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

func (hc *Client) getClientWithHashes(
	cid int,
	hashes [][]byte,
) (proto.HighwayServiceClient, peer.ID, error) {
	connectedPeers := hc.m.GetPeers(cid)
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	if len(connectedPeers) > 0 {
		p := PeerInfo{}
		// Find block proposer (position = 0) and ask it
		mark := map[peer.ID]struct{}{}
		for range connectedPeers {
			for {
				p = connectedPeers[r1.Intn(len(connectedPeers))]
				if _, ok := mark[p.ID]; !ok {
					mark[p.ID] = struct{}{}
					break
				}
				if len(mark) == len(connectedPeers) {
					break
				}
			}
			if pos, ok := hc.m.watcher.pos[p.ID]; (ok && ((pos.id > 0) && (pos.id <= 21))) || (len(mark) == len(connectedPeers)) {
				client, err := hc.FindServiceClient(p.ID)
				if err == nil {
					return client, p.ID, nil
				}
			}
		}
	}
	return hc.router.GetClientSupportShard(cid)
}

func (hc *Client) getClientWithHashesV2(
	cid int,
	hashes [][]byte,
) (proto.HighwayServiceClient, peer.ID, error) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	peers := hc.m.watcher.getFixedPeersOfCID(cid)
	if len(peers) > 0 {
		p := peers[r1.Intn(len(peers))]
		client, err := hc.FindServiceClient(p)
		if err == nil {
			return client, p, nil
		}
	}
	return hc.router.GetClientSupportShard(cid)
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

func (hc *Client) FindServiceClient(pID peer.ID) (proto.HighwayServiceClient, error) {
	hwID, err := hc.peerStore.GetHWIDOfPeer(pID)
	if err != nil {
		return nil, err
	}
	if hwID != hc.router.GetID() { // Peer not connected, let ask the other highway
		logger.Debugf("Peer %v is not connected, connect to hw %s", pID.String(), hwID.String())
		hwClient, _, err := hc.router.GetHighwayServiceClient(hwID)
		return hwClient, err
	}
	// Connected peer, get connection
	return hc.cc.GetServiceClient(pID)
}

// choosePeerIDWithBlock returns peerID of a node that holds some blocks
// and its corresponding highway's peerID
func (hc *Client) choosePeerIDWithBlock(ctx context.Context, cid int, blk uint64) (pid peer.ID, hw peer.ID, err error) {
	logger := Logger(ctx)
	_ = logger

	peersHasBlk, err := hc.peerStore.GetPeerHasBlkV2(blk, byte(cid)) // Get all peers from peerstate
	// logger.Debugf("PeersHasBlk for cid %v blk %v: %+v", cid, blk, peersHasBlk)
	// logger.Debugf("PeersHasBlk for cid %v: %+v", cid, peersHasBlk)
	if err != nil {
		return peer.ID(""), peer.ID(""), err
	}
	if len(peersHasBlk) == 0 {
		return peer.ID(""), peer.ID(""), errors.Errorf("no peer with blk %d %d", cid, blk)
	}

	// Prioritize peers and sort into different groups
	connectedPeers := hc.m.GetPeers(cid) // Filter out disconnected peers
	groups := groupPeersByDistance(peersHasBlk, blk, hc.router.GetID(), connectedPeers, hc.m.watcher)
	// logger.Debugf("Peers by groups: %+v", groups)

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
	w *watcher,
) [][]chaindata.PeerWithBlk {
	// Group peers into 4 groups:
	a := []chaindata.PeerWithBlk{}  // 1.  Fixed Nodes connected to this highway and have all needed blocks
	a2 := []chaindata.PeerWithBlk{} // 1.5 Nodes connected to this highway and have all needed blocks
	b := []chaindata.PeerWithBlk{}  // 2.  Nodes from other highways and have all needed blocks
	h := uint64(0)                  // Find maximum height
	for _, p := range peers {
		if p.Height >= blk {
			if p.HW == selfPeerID {
				_, ok := w.getPeerPosition(p.ID)
				if ok {
					a = append(a, p)
				} else {
					a2 = append(a2, p)
				}
			} else {
				b = append(b, p)
			}
		}
		if p.Height > h {
			h = p.Height
		}
	}
	a = filterPeers(a, connectedPeers)   // Retain only connected peers
	a2 = filterPeers(a2, connectedPeers) // Retain only connected peers

	c := []chaindata.PeerWithBlk{}  // 3.  Fixed Nodes connected to this highway and have the largest amount of blocks
	c2 := []chaindata.PeerWithBlk{} // 3.5 Nodes connected to this highway and have the largest amount of blocks
	d := []chaindata.PeerWithBlk{}  // 4.  Nodes from other highways and have the largest amount of blocks
	for _, p := range peers {
		if p.Height < blk && p.Height+common.ChoosePeerBlockDelta >= h {
			if p.HW == selfPeerID {
				_, ok := w.getPeerPosition(p.ID)
				if ok {
					c = append(c, p)
				} else {
					c2 = append(c2, p)
				}
			} else {
				d = append(d, p)
			}
		}
	}
	c = filterPeers(c, connectedPeers)   // Retain only connected peers
	c2 = filterPeers(c2, connectedPeers) // Retain only connected peers
	return [][]chaindata.PeerWithBlk{a, a2, b, c, c2, d}
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
	cc.conns.pLock.Lock()
	_, exist := cc.conns.peerLocker[peerID]
	if !exist {
		cc.conns.peerLocker[peerID] = sync.Mutex{}
	}
	locker := cc.conns.peerLocker[peerID]
	cc.conns.pLock.Unlock()
	locker.Lock()
	defer locker.Unlock()
	cc.conns.cLock.Lock()
	c, ok := cc.conns.connMap[peerID]
	cc.conns.cLock.Unlock()
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
		if err == nil {
			cc.conns.cLock.Lock()
			cc.conns.connMap[peerID] = conn
			c = cc.conns.connMap[peerID]
			cc.conns.cLock.Unlock()
		} else {
			return nil, errors.WithStack(err)
		}
	}
	return proto.NewHighwayServiceClient(c), nil
}

func (cc *ClientConnector) CloseDisconnected(peerID peer.ID) {
	cc.conns.pLock.Lock()
	cc.conns.cLock.Lock()
	defer cc.conns.cLock.Unlock()
	defer cc.conns.pLock.Unlock()

	if c, ok := cc.conns.connMap[peerID]; ok {
		logger.Infof("Closing connection to pID %s", peerID.String())
		if err := c.Close(); err != nil {
			logger.Warnf("Failed closing connection to pID %s: %s", peerID.String(), errors.WithStack(err))
		} else {
			delete(cc.conns.connMap, peerID)
			delete(cc.conns.peerLocker, peerID)
			logger.Infof("Closed connection to pID %s successfully", peerID.String())
		}
	}
}

type ClientConnector struct {
	dialer Dialer
	conns  struct {
		connMap    map[peer.ID]*grpc.ClientConn
		peerLocker map[peer.ID]sync.Mutex
		cLock      sync.RWMutex
		pLock      sync.RWMutex
	}
}

func NewClientConnector(dialer Dialer) *ClientConnector {
	connector := &ClientConnector{dialer: dialer}
	connector.conns.connMap = map[peer.ID]*grpc.ClientConn{}
	connector.conns.peerLocker = map[peer.ID]sync.Mutex{}
	connector.conns.cLock = sync.RWMutex{}
	connector.conns.pLock = sync.RWMutex{}
	return connector
}

type PeerStore interface {
	GetPeerHasBlk(blkHeight uint64, committeeID byte) ([]chaindata.PeerWithBlk, error)
	GetPeerHasBlkV2(blkHeight uint64, committeeID byte) ([]chaindata.PeerWithBlk, error)
	GetHWIDOfPeer(pID peer.ID) (peer.ID, error)
	GetPeerIDOfPubkey(pk string) (peer.ID, error)
}

type Dialer interface {
	Dial(ctx context.Context, peerID peer.ID, dialOpts ...grpc.DialOption) (*grpc.ClientConn, error)
}

type Router interface {
	GetClientSupportShard(cid int) (proto.HighwayServiceClient, peer.ID, error)
	GetHighwayServiceClient(pid peer.ID) (proto.HighwayServiceClient, peer.ID, error)
	GetID() peer.ID
	CheckHWPeerID(pID string) bool
}

type GetBlockByHeightRequest interface {
	GetCallDepth() int32
	GetFromPool() bool
	GetFrom() int32
	GetTo() int32
	GetSpecific() bool
	GetFromHeight() uint64
	GetToHeight() uint64
	GetHeights() []uint64
	GetUUID() string
}

type GetBlockByHashRequest interface {
	GetCallDepth() int32
	GetCID() int32
	GetHashes() [][]byte
	GetUUID() string
}

type RequestBlockByHeight interface {
	GetType() proto.BlkType
	GetCallDepth() int32
	GetFrom() int32
	GetTo() int32
	GetSpecific() bool
	GetHeights() []uint64
	GetSyncFromPeer() string
	GetUUID() string
}

type RequestBlockByHash interface {
	GetType() proto.BlkType
	GetCallDepth() int32
	GetFrom() int32
	GetTo() int32
	GetHashes() [][]byte
	GetSyncFromPeer() string
	GetUUID() string
}
