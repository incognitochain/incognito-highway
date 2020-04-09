package route

import (
	"highway/common"
	hmap "highway/route/hmap"
	"highway/route/mocks"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

func configTime() func() {
	broadcastMsgEnlistTimestep := common.BroadcastMsgEnlistTimestep
	routeKeepConnectionTimestep := common.RouteKeepConnectionTimestep
	routeHighwayRemoveDeadline := common.RouteHighwayRemoveDeadline
	common.BroadcastMsgEnlistTimestep = 100 * time.Millisecond
	common.RouteKeepConnectionTimestep = 20 * time.Millisecond
	common.RouteHighwayRemoveDeadline = 100 * time.Millisecond

	return func() {
		// Revert time configuration after a test is done
		common.BroadcastMsgEnlistTimestep = broadcastMsgEnlistTimestep
		common.RouteKeepConnectionTimestep = routeKeepConnectionTimestep
		common.RouteHighwayRemoveDeadline = routeHighwayRemoveDeadline
	}
}

func TestEnlistLoop(t *testing.T) {
	defer configTime()()

	h, _ := setupHost()
	publisher := &mocks.Publisher{}
	publisher.On("Publish", mock.Anything, mock.Anything).Return(nil)
	connector := &Connector{
		host:      h,
		publisher: publisher,
		stop:      make(chan int),
	}
	go connector.Start()
	time.Sleep(2*common.BroadcastMsgEnlistTimestep + 200*time.Microsecond)
	connector.stop <- 1
	publisher.AssertNumberOfCalls(t, "Publish", 2)
}

func TestCloseBothStreams(t *testing.T) {
	h, net := setupHost()
	net.On("ClosePeer", mock.Anything).Return(nil)
	keeper := &mocks.ConnKeeper{}
	keeper.On("CloseConnection", mock.Anything).Return(nil)
	connector := &Connector{
		host: h,
		hwc:  keeper,
	}
	connector.closePeer(peer.ID(""))

	net.AssertNumberOfCalls(t, "ClosePeer", 1)
	keeper.AssertNumberOfCalls(t, "CloseConnection", 1)
}

func TestBroadcastEnlistMsg(t *testing.T) {
	publisher := &mocks.Publisher{}
	publisher.On("Publish", mock.Anything, mock.Anything).Return(nil)
	connector := &Connector{
		publisher: publisher,
	}
	assert.Nil(t, connector.enlist())
	publisher.AssertNumberOfCalls(t, "Publish", 1)
}

func TestDialAndEnlistNewPeer(t *testing.T) {
	h, net := setupHost()
	h.On("Connect", mock.Anything, mock.Anything).Return(nil)
	setupConnectedness(net, []network.Connectedness{network.NotConnected})
	publisher := &mocks.Publisher{}
	publisher.On("Publish", mock.Anything, mock.Anything).Return(nil)
	hmap := hmap.NewMap(peer.AddrInfo{}, []byte{0, 1, 2, 3}, "")
	connector := &Connector{
		host:      h,
		hmap:      hmap,
		publisher: publisher,
	}
	assert.Nil(t, connector.dialAndEnlist(peer.AddrInfo{}))
	publisher.AssertNumberOfCalls(t, "Publish", 1)
	assert.True(t, hmap.IsConnectedToPeer(peer.ID("")))
}

func TestDialNewClientConnection(t *testing.T) {
	dialer := &mocks.Dialer{}
	dialer.On("Dial", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&grpc.ClientConn{}, nil)
	client := &Client{
		dialer: dialer,
	}
	client.conns.connMap = map[peer.ID]*grpc.ClientConn{}
	client.conns.RWMutex = &sync.RWMutex{}

	_, err := client.GetConnection(peer.ID(""))
	assert.Nil(t, err)
}
