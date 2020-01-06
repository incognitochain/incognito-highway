package route

import (
	"highway/route/mocks"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/mock"
)

func TestEnlistLoop(t *testing.T) {
	h, _ := setupHost()
	publisher := &mocks.Publisher{}
	publisher.On("Publish", mock.Anything, mock.Anything).Return(nil)
	connector := &Connector{
		host:      h,
		publisher: publisher,
		stop:      make(chan int),
	}
	go connector.Start()
	time.Sleep(5 * time.Second)
	connector.stop <- 1
	publisher.AssertNumberOfCalls(t, "Publish", 2)
}

func TestCloseBothStreams(t *testing.T) {
	h, net := setupHost()
	net.On("ClosePeer", mock.Anything).Return(nil)
	connector := &Connector{host: h}
	h.closePeer(peer.ID(""))

	net.AssertNumberOfCalls(t, "ClosePeer", 1)
}
