package process

import (
	"highway/process/topic"

	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

// Maybe we should merge them into 4 function
func (pubsub *PubSubManager) HandleBlkBeacon(
	fromInside bool,
	msgType string,
	cID int,
	data []byte,
) {
	var topicPubs []string
	topicPubsInside := topic.Handler.GetHWPubTopicsFromMsg(msgType, cID)
	topicPubs = append(topicPubs, topicPubsInside...)
	if fromInside {
		topicPub := topic.Handler.GetHWPubSubOutSideFromMsg(msgType, cID)
		topicPubs = append(topicPubs, topicPub)
	}
	for _, topicPub := range topicPubs {
		pubsub.FloodMachine.Publish(topicPub, data)
	}
}

//============
// Cross commitee block: blkshardtobeacon, cross shard block
func (pubsub *PubSubManager) HandleCrossCommitteeBlock(
	fromInside bool,
	msgType string,
	cID int,
	data []byte,
) {
	var topicPubs []string
	if fromInside {
		topicPub := topic.Handler.GetHWPubSubOutSideFromMsg(msgType, cID)
		topicPubs = append(topicPubs, topicPub)
	} else {
		topicPubs = topic.Handler.GetHWPubTopicsFromMsg(msgType, cID)
	}
	for _, topicPub := range topicPubs {
		pubsub.FloodMachine.Publish(topicPub, data)
	}
}

//============
// HandleBlkShard msg block shard just available inside a shard committee
func (pubsub *PubSubManager) HandleBlkShard(
	msgType string,
	cID int,
	data []byte,
) {
	topicPubs := topic.Handler.GetHWPubTopicsFromMsg(msgType, cID)
	for _, topicPub := range topicPubs {
		pubsub.FloodMachine.Publish(topicPub, data)
	}
}

func (pubsub *PubSubManager) HandleTX(
	msgType string,
	cID int,
	data []byte,
) {
	topicPubs := topic.Handler.GetHWPubTopicsFromMsg(msgType, cID)
	for _, topicPub := range topicPubs {
		pubsub.FloodMachine.Publish(topicPub, data)
	}
}

//============
func (pubsub *PubSubManager) HandlePeerState(
	pID peer.ID,
	data []byte,
) {
	err := pubsub.BlockChainData.UpdatePeerState(pubsub.BlockChainData.GetMiningPubkeyFromPeerID(pID), data)
	if err != nil {
		err = errors.WithMessagef(err, "from: %v", pID)
		logger.Warnf("Cannot update peerstate: %+v", err)
	}
}

//============
