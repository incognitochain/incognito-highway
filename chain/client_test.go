package chain

import (
	context "context"
	"highway/chain/mocks"
	"highway/chaindata"
	"highway/route"
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

	client := &Client{
		m:            m,
		routeManager: &route.Manager{ID: peer.ID("123")},
		peerStore:    peerStore,
	}

	pid, hw, err := client.choosePeerIDWithBlock(ctx, cid, blk)
	if assert.Nil(t, err) {
		assert.Equal(t, peer.ID(""), pid)
		assert.Equal(t, peer.ID(""), hw)
	}
}

func TestGroupPeers(t *testing.T) {
	// Connected, with blk
	selfPeerID := peer.ID("abc")
	blk := uint64(100)
	a := []chaindata.PeerWithBlk{}
	for i := 0; i < 3; i++ {
		a = append(a, chaindata.PeerWithBlk{
			HW:     selfPeerID,
			Height: uint64(123),
		})
	}

	// Not connected, with blk
	b := []chaindata.PeerWithBlk{}
	for i := 0; i < 5; i++ {
		b = append(b, chaindata.PeerWithBlk{
			Height: uint64(123),
		})
	}

	// Connected, without blk
	c := []chaindata.PeerWithBlk{}
	for i := 0; i < 7; i++ {
		c = append(c, chaindata.PeerWithBlk{
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

	expected := [][]chaindata.PeerWithBlk{a, b, c, d}

	peers := []chaindata.PeerWithBlk{}
	peers = append(peers, a...)
	peers = append(peers, b...)
	peers = append(peers, c...)
	peers = append(peers, d...)
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})

	connectedPeers := []PeerInfo{PeerInfo{ID: peer.ID("")}, PeerInfo{ID: selfPeerID}}
	groups := groupPeersByDistance(peers, blk, selfPeerID, connectedPeers)
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
			expectedTo: 115,
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

func init() {
	cf := zap.NewDevelopmentConfig()
	cf.Level.SetLevel(zapcore.FatalLevel)
	l, _ := cf.Build()
	logger = l.Sugar()

	// chain.InitLogger(logger)
	// chaindata.InitLogger(logger)
	InitLogger(logger)
	// process.InitLogger(logger)
	// topic.InitLogger(logger)
	// health.InitLogger(logger)
	// rpcserver.InitLogger(logger)
	// hmap.InitLogger(logger)
	// datahandler.InitLogger(logger)
}
