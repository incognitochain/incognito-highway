package route

import (
	"context"
	"encoding/json"
	"highway/common"
	logger "highway/customizelog"
	"highway/p2p"
	"highway/process"
	"time"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/stathat/consistent"
)

type Manager struct {
	ID peer.ID

	hmap *Map
	hc   *Connector
}

func NewManager(
	supportShards []byte,
	bootstrap []string,
	masternode peer.ID,
	h *p2p.Host,
) *Manager {
	// TODO(@0xbunyip): use bootstrap to get initial highways
	p := peer.AddrInfo{
		ID:    h.Host.ID(),
		Addrs: h.Host.Addrs(),
	}
	hmap := NewMap(p, supportShards)

	hw := &Manager{
		ID:   h.Host.ID(),
		hmap: hmap,
		hc: NewConnector(
			h,
			hmap,
			&process.GlobalPubsub,
			masternode,
		),
	}

	hw.setup(bootstrap)

	// Start highway connector event loop
	go hw.hc.Start()
	return hw
}

func (h *Manager) setup(bootstrap []string) {
	for _, b := range bootstrap {
		if len(b) == 0 {
			continue
		}

		// TODO(@0xbunyip): parse bootstrap nodes
		ss := []byte{0}
		id, _ := peer.IDB58Decode("QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv")
		addr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9330")
		addrInfo := peer.AddrInfo{
			ID:    id,
			Addrs: []multiaddr.Multiaddr{addr},
		}
		h.hmap.AddPeer(addrInfo, ss)

		// Get latest committee from bootstrap highways if available
		err := h.hc.Dial(addrInfo)
		if err != nil {
			logger.Warn("Failed dialing to bootstrap node", addrInfo, err)
			continue
		}

		cc, err := h.GetChainCommittee(id)
		if err != nil {
			logger.Warnf("Failed get chain committtee: %+v", err)
			continue
		}
		logger.Info("Received chain committee:", cc)

		// TOOD(@0xbunyip): update chain committee to ChainData here
	}
}

func (h *Manager) GetChainCommittee(pid peer.ID) (*incognitokey.ChainCommittee, error) {
	c, err := h.hc.GetHWClient(pid)
	if err != nil {
		return nil, err
	}
	resp, err := c.GetChainCommittee(context.Background(), &process.GetChainCommitteeRequest{})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	comm := &incognitokey.ChainCommittee{}
	if err := json.Unmarshal(resp.Data, comm); err != nil {
		return nil, errors.Wrapf(err, "comm: %s", comm)
	}
	return comm, nil
}

func (h *Manager) Start() {
	// Update connection when new highway comes online or old one goes offline
	for range time.Tick(5 * time.Second) { // TODO(@xbunyip): move params to config
		// TODO(@0xbunyip) Check for liveness of connected highways

		// New highways online: update map and reconnect to load-balance
		newManager := true
		_ = newManager

		// Connect to other highways if needed
		h.UpdateConnection()
	}
}

func (h *Manager) UpdateConnection() {
	for i := byte(0); i < common.NumberOfShard; i++ {
		if err := h.connectChain(i); err != nil {
			logger.Error(err)
		}
	}

	if err := h.connectChain(common.BEACONID); err != nil {
		logger.Error(err)
	}
}

// connectChain connects this highway to a peer in a chain (shard or beacon) if it hasn't connected to one yet
func (h *Manager) connectChain(sid byte) error {
	if h.hmap.IsConnectedToShard(sid) {
		return nil
	}

	logger.Info("connect chain", sid)
	highways := h.hmap.Peers[sid]
	if len(highways) == 0 {
		return errors.Errorf("found no highway supporting shard %d", sid)
	}

	// TODO(@0xbunyip): repick if fail to connect
	p, err := choosePeer(highways, h.ID)
	if err != nil {
		return errors.WithMessagef(err, "shardID: %v", sid)
	}
	if err := h.connectTo(p); err != nil {
		return err
	}

	// Update list of connected shards
	h.hmap.ConnectToShardOfPeer(p)
	return nil
}

func (h *Manager) connectTo(p peer.AddrInfo) error {
	return h.hc.ConnectTo(p)
}

// choosePeer picks a peer from a list using consistent hashing
func choosePeer(peers []peer.AddrInfo, id peer.ID) (peer.AddrInfo, error) {
	cst := consistent.New()
	for _, p := range peers {
		cst.Add(string(p.ID))
	}

	closest, err := cst.Get(string(id))
	if err != nil {
		return peer.AddrInfo{}, errors.New("could not get consistent-hashing peer")
	}

	for _, p := range peers {
		if string(p.ID) == closest {
			return p, nil
		}
	}
	return peer.AddrInfo{}, errors.New("failed choosing peer to connect")
}
