package chain

import (
	context "context"
	"highway/chain/mocks"
	"highway/chaindata"
	"highway/common"
	"highway/process/topic"
	"highway/proto"
	"math/rand"
	"sync"
	"testing"

	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	grpc "google.golang.org/grpc"
)

func TestRegisterTopics(t *testing.T) {
	s := Server{
		reporter: NewReporter(&Manager{}),
		m: &Manager{
			newPeers: make(chan PeerInfo, 10),
		},
	}

	req := &proto.RegisterRequest{
		PeerID:             "QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv",
		CommitteePublicKey: "1J9fpx9aZme3gF9YRYQAZJrhnHZUgihbwWcxxNFNzRuET7uFwGp2PTpiVA4rPPFKEBjSPs9YFC6vEmuhJ47BoG2oNyKzW8Xmab6A9duRZLnrAb7sBmPnHFvAX7k9yyNZvun1iPEUSJ4m2DB4yq2N6xrUQ2hz1eLNF1wsoChehcfZjPU218LvH",
		CommitteeID:        []byte{0},
		Role:               "",
		WantedMessages:     []string{"blockshard"},
	}
	topic.Handler = topic.TopicManager{}
	topic.Handler.Init("")
	topic.Handler.UpdateSupportShards([]byte{1})

	resp, err := s.Register(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, s.m.newPeers, 1)
	assert.Len(t, resp.Pair, 1)
	assert.Equal(t, proto.UserRole{
		Layer: common.ShardRole,
		Role:  "",
		Shard: int32(0),
	}, *resp.Role)
}

func TestValidatorRegister(t *testing.T) {
	chainData := &chaindata.ChainData{}
	chainData.Init(8)
	s := Server{
		reporter:  NewReporter(&Manager{}),
		chainData: chainData,
		m: &Manager{
			newPeers: make(chan PeerInfo, 10),
		},
	}

	req := &proto.RegisterRequest{
		PeerID:             "QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv",
		CommitteePublicKey: "1J9fpx9aZme3gF9YRYQAZJrhnHZUgihbwWcxxNFNzRuET7uFwGp2PTpiVA4rPPFKEBjSPs9YFC6vEmuhJ47BoG2oNyKzW8Xmab6A9duRZLnrAb7sBmPnHFvAX7k9yyNZvun1iPEUSJ4m2DB4yq2N6xrUQ2hz1eLNF1wsoChehcfZjPU218LvH",
		Role:               common.CommitteeRole,
	}
	resp, err := s.Register(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, chainData.MiningPubkeyByPeerID, 1)
	assert.Len(t, chainData.PeerIDByMiningPubkey, 1)
	assert.Len(t, chainData.ShardByMiningPubkey, 1)
	assert.Len(t, s.m.newPeers, 1)
}

func TestGetClientNode(t *testing.T) {
	hwPID := peer.ID("123")
	peerStore := &mocks.PeerStore{}
	peerStore.On("GetPeerHasBlk", mock.Anything, mock.Anything).Return([]chaindata.PeerWithBlk{chaindata.PeerWithBlk{HW: hwPID, Height: 1235}}, nil)

	cid := 0
	m := &Manager{
		watcher: newWatcher(nil, 0),
	}
	m.peers.ids = map[int][]PeerInfo{cid: []PeerInfo{PeerInfo{ID: peer.ID("")}}}
	m.peers.RWMutex = sync.RWMutex{}

	connector := &ClientConnector{}
	connector.conns.connMap = map[peer.ID]*grpc.ClientConn{
		peer.ID(""): &grpc.ClientConn{},
	}
	connector.conns.RWMutex = sync.RWMutex{}

	router := &mocks.Router{}
	router.On("GetID").Return(hwPID)
	client := &Client{
		m:         m,
		cc:        connector,
		router:    router,
		peerStore: peerStore,
	}

	ctx := context.Background()
	height := uint64(1234)
	_, pid, err := client.getClientOfSupportedShard(ctx, cid, height)
	assert.Nil(t, err)
	assert.Equal(t, peer.ID(""), pid)
}

func TestGetClientFromHighway(t *testing.T) {
	peerStore := &mocks.PeerStore{}
	peerStore.On("GetPeerHasBlk", mock.Anything, mock.Anything).Return([]chaindata.PeerWithBlk{chaindata.PeerWithBlk{}}, nil)

	cid := 0
	m := &Manager{}
	m.peers.ids = map[int][]PeerInfo{cid: []PeerInfo{}}
	m.peers.RWMutex = sync.RWMutex{}

	router := &mocks.Router{}
	router.On("GetID").Return(peer.ID("123"))
	router.On("GetHighwayServiceClient", mock.Anything).Return(proto.NewHighwayServiceClient(nil), peer.ID("123"), nil)
	client := &Client{
		m:         m,
		router:    router,
		peerStore: peerStore,
	}

	ctx := context.Background()
	height := uint64(1234)
	_, pid, err := client.getClientOfSupportedShard(ctx, cid, height)
	assert.Nil(t, err)
	assert.Equal(t, peer.ID("123"), pid)
}

func TestGetServiceClient(t *testing.T) {
	peerID := peer.ID("")

	calledWithTimeout := false
	dialer := &mocks.Dialer{}
	dialer.On("Dial", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&grpc.ClientConn{}, nil).Run(
		func(args mock.Arguments) {
			_, ok := args.Get(0).(context.Context).Deadline()
			calledWithTimeout = ok
		},
	)
	connector := &ClientConnector{dialer: dialer}
	connector.conns.connMap = map[peer.ID]*grpc.ClientConn{}
	connector.conns.RWMutex = sync.RWMutex{}
	_, err := connector.GetServiceClient(peerID)
	if assert.Nil(t, err) {
		assert.Len(t, connector.conns.connMap, 1)
		assert.True(t, calledWithTimeout)
	}
}

func TestChoosePeerID(t *testing.T) {
	ctx := context.Background()
	cid := 0
	blk := uint64(123)

	peerStore := &mocks.PeerStore{}
	peerStore.On("GetPeerHasBlk", mock.Anything, mock.Anything).Return([]chaindata.PeerWithBlk{chaindata.PeerWithBlk{}}, nil)

	m := &Manager{}
	m.peers.ids = map[int][]PeerInfo{cid: []PeerInfo{}}
	m.peers.RWMutex = sync.RWMutex{}

	router := &mocks.Router{}
	router.On("GetID").Return(peer.ID("123"))
	client := &Client{
		m:         m,
		router:    router,
		peerStore: peerStore,
	}

	pid, hw, err := client.choosePeerIDWithBlock(ctx, cid, blk)
	if assert.Nil(t, err) {
		assert.Equal(t, peer.ID(""), pid)
		assert.Equal(t, peer.ID(""), hw)
	}
}

func TestGroupPeers(t *testing.T) {
	w := newWatcher(nil, 0)
	w.pos = map[peer.ID]position{}
	w.setPeerPosition(peer.ID("123123123"), position{
		cid: 2,
		id:  1,
	})
	w.setPeerPosition(peer.ID("123123124"), position{
		cid: 2,
		id:  1,
	})
	// Connected, with blk
	selfPeerID := peer.ID("abc")
	blk := uint64(100)
	a2 := []chaindata.PeerWithBlk{}
	for i := 0; i < 3; i++ {
		a2 = append(a2, chaindata.PeerWithBlk{
			HW:     selfPeerID,
			Height: uint64(123),
		})
	}

	// Fixed node Connected, with blk
	a := []chaindata.PeerWithBlk{}
	a = append(a, chaindata.PeerWithBlk{
		HW:     selfPeerID,
		Height: uint64(123),
		ID:     peer.ID("123123123"),
	})

	// Not connected, with blk
	b := []chaindata.PeerWithBlk{}
	for i := 0; i < 5; i++ {
		b = append(b, chaindata.PeerWithBlk{
			Height: uint64(123),
		})
	}

	// Connected, without blk
	c := []chaindata.PeerWithBlk{}
	c = append(c, chaindata.PeerWithBlk{
		HW:     selfPeerID,
		Height: uint64(45),
		ID:     peer.ID("123123124"),
	})

	c2 := []chaindata.PeerWithBlk{}
	for i := 0; i < 7; i++ {
		c2 = append(c2, chaindata.PeerWithBlk{
			HW:     selfPeerID,
			Height: uint64(45),
		})
	}

	// Not connected, without blk
	d := []chaindata.PeerWithBlk{}
	for i := 0; i < 9; i++ {
		d = append(d, chaindata.PeerWithBlk{
			Height: uint64(45),
		})
	}

	expected := [][]chaindata.PeerWithBlk{a, a2, b, c, c2, d}

	peers := []chaindata.PeerWithBlk{}
	peers = append(peers, a...)
	peers = append(peers, a2...)
	peers = append(peers, b...)
	peers = append(peers, c...)
	peers = append(peers, c2...)
	peers = append(peers, d...)
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})

	connectedPeers := []PeerInfo{PeerInfo{ID: peer.ID("")}, PeerInfo{ID: selfPeerID}, PeerInfo{ID: peer.ID("123123124")}, PeerInfo{ID: peer.ID("123123123")}}
	groups := groupPeersByDistance(peers, blk, selfPeerID, connectedPeers, w)
	assert.Equal(t, expected, groups)
}

func TestChoosePeerFromGroup(t *testing.T) {
	buildGroup := func(height uint64) []chaindata.PeerWithBlk {
		g := []chaindata.PeerWithBlk{}
		for i := uint64(0); i < 10+height*100; i++ {
			g = append(g, chaindata.PeerWithBlk{
				Height: height,
			})
		}
		return g
	}

	testCases := []struct {
		desc          string
		groups        [][]chaindata.PeerWithBlk
		expectedGroup uint64
	}{
		{
			desc:          "Choose from 1st group",
			groups:        [][]chaindata.PeerWithBlk{buildGroup(0), buildGroup(1), buildGroup(2), buildGroup(3)},
			expectedGroup: 0,
		},
		{
			desc:          "Choose from 2nd group",
			groups:        [][]chaindata.PeerWithBlk{[]chaindata.PeerWithBlk{}, buildGroup(1), buildGroup(2), buildGroup(3)},
			expectedGroup: 1,
		},
		{
			desc:          "Choose from 3rd group",
			groups:        [][]chaindata.PeerWithBlk{[]chaindata.PeerWithBlk{}, []chaindata.PeerWithBlk{}, buildGroup(2), buildGroup(3)},
			expectedGroup: 2,
		},
		{
			desc:          "Choose from 4th group",
			groups:        [][]chaindata.PeerWithBlk{[]chaindata.PeerWithBlk{}, []chaindata.PeerWithBlk{}, []chaindata.PeerWithBlk{}, buildGroup(3)},
			expectedGroup: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			p, err := choosePeerFromGroup(tc.groups)
			if assert.Nil(t, err) {
				assert.Equal(t, tc.expectedGroup, p.Height)
			}
		})
	}
}

func TestCapBlocks(t *testing.T) {
	defer configMaxBlock()()

	testCases := []struct {
		desc            string
		specific        bool
		from            uint64
		to              uint64
		heights         []uint64
		expectedTo      uint64
		expectedHeights []uint64
	}{
		{
			desc:       "Range, not exceeded cap",
			specific:   false,
			from:       5,
			to:         85,
			expectedTo: 85,
		},
		{
			desc:       "Range, exceeded cap",
			specific:   false,
			from:       15,
			to:         155,
			expectedTo: 114,
		},
		{
			desc:            "Specific, not exceeded cap",
			specific:        true,
			heights:         []uint64{5, 7, 15, 22, 33, 99, 150, 1555},
			expectedHeights: []uint64{5, 7, 15, 22, 33, 99, 150, 1555},
		},
		{
			desc:     "Specific, exceeded cap",
			specific: true,
			heights: func() []uint64 {
				h := []uint64{}
				for i := 0; i < 333; i++ {
					h = append(h, uint64(i))
				}
				return h
			}(),
			expectedHeights: func() []uint64 {
				h := []uint64{}
				for i := 0; i < 100; i++ {
					h = append(h, uint64(i))
				}
				return h
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			to, heights := capBlocksPerRequest(tc.specific, tc.from, tc.to, tc.heights)
			if tc.specific {
				assert.Equal(t, tc.expectedHeights, heights)
			} else {
				assert.Equal(t, tc.expectedTo, to)
			}
		})
	}
}

func configMaxBlock() func() {
	maxBlockPerRequest := common.MaxBlocksPerRequest
	common.MaxBlocksPerRequest = 100

	return func() {
		// Revert configuration after a test is done
		common.MaxBlocksPerRequest = maxBlockPerRequest
	}
}

func init() {
	cf := zap.NewDevelopmentConfig()
	cf.Level.SetLevel(zapcore.FatalLevel)
	l, _ := cf.Build()
	logger = l.Sugar()

	// chain.InitLogger(logger)
	chaindata.InitLogger(logger)
	InitLogger(logger)
	// process.InitLogger(logger)
	// topic.InitLogger(logger)
	// health.InitLogger(logger)
	// rpcserver.InitLogger(logger)
	// hmap.InitLogger(logger)
	// datahandler.InitLogger(logger)
}
