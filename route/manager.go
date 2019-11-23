package route

import (
	"context"
	"encoding/json"
	"highway/common"
	"highway/process"
	"highway/proto"
	"time"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/libp2p/go-libp2p-core/host"
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
	h host.Host,
	prtc *p2pgrpc.GRPCProtocol,
) *Manager {
	// TODO(@0xbunyip): use bootstrap to get initial highways
	p := peer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	hmap := NewMap(p, supportShards)

	hw := &Manager{
		ID:   h.ID(),
		hmap: hmap,
		hc: NewConnector(
			h,
			prtc,
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
	if len(bootstrap) == 0 || len(bootstrap[0]) == 0 {
		return
	}
	hInfos, err := h.getListHighwaysFromPeer(bootstrap[0])
	if err != nil {
		logger.Errorf("Failed getting list of highways from peer %+v, err = %+v", bootstrap[0], err)
		return
	}

	for _, b := range hInfos {
		// Get peer info
		addr, err := multiaddr.NewMultiaddr(b.PeerInfo)
		if err != nil {
			logger.Warnf("Invalid highway addr: %v", b)
			continue
		}
		addrInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			logger.Warnf("Invalid highway addr: %v, %v", b, addr)
			continue
		}

		// Remember this highway
		ss := []byte{}
		for _, s := range b.SupportShards {
			ss = append(ss, byte(s))
		}
		h.hmap.AddPeer(*addrInfo, ss)
	}

	// TODO(@0xbunyip): Get latest committee from bootstrap highways if available
	// err = h.hc.Dial(*addrInfo)
	// if err != nil {
	// 	logger.Warn("Failed dialing to bootstrap node", addrInfo, err)
	// 	continue
	// }

	// cc, err := h.GetChainCommittee(addrInfo.ID)
	// if err != nil {
	// 	logger.Warnf("Failed get chain committtee: %+v", err)
	// 	continue
	// }
	// logger.Info("Received chain committee:", cc)

	// // TOOD(@0xbunyip): update chain committee to ChainData here
}

func (h *Manager) getListHighwaysFromPeer(ma string) ([]*proto.HighwayInfo, error) {
	addrInfo, err := common.StringToAddrInfo(ma)
	if err != nil {
		return nil, err
	}

	err = h.hc.Dial(*addrInfo)
	if err != nil {
		return nil, err
	}

	hInfos, err := h.GetListHighways(addrInfo.ID)
	if err != nil {
		return nil, err
	}
	return hInfos, nil
}

func (h *Manager) GetChainCommittee(pid peer.ID) (*incognitokey.ChainCommittee, error) {
	c, err := h.hc.GetHWClient(pid)
	if err != nil {
		return nil, err
	}
	resp, err := c.GetChainCommittee(context.Background(), &proto.GetChainCommitteeRequest{})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	comm := &incognitokey.ChainCommittee{}
	if err := json.Unmarshal(resp.Data, comm); err != nil {
		return nil, errors.Wrapf(err, "comm: %s", comm)
	}
	return comm, nil
}

func (h *Manager) GetListHighways(pid peer.ID) ([]*proto.HighwayInfo, error) {
	c, err := h.hc.GetHWClient(pid)
	if err != nil {
		return nil, err
	}
	resp, err := c.GetHighwayInfos(context.Background(), &proto.GetHighwayInfosRequest{})
	logger.Infof("resp: %+v", resp)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp.Highways, nil
}

func (h *Manager) Start() {
	// Update connection when new highway comes online or old one goes offline
	for range time.Tick(10 * time.Second) { // TODO(@xbunyip): move params to config
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

	logger.Info("Connecting to chain ", sid)
	highways := h.hmap.Peers[sid]
	if len(highways) == 0 {
		return errors.Errorf("found no highway supporting chain %d", sid)
	}

	// TODO(@0xbunyip): repick if fail to connect
	p, err := choosePeer(highways, h.ID)
	if err != nil {
		return errors.WithMessagef(err, "shardID: %v", sid)
	}
	if err := h.connectTo(p); err != nil {
		return err
	}
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

// GetRouteClientWithBlock returns the grpc client with connection to a highway
// supporting a specific shard
func (h *Manager) GetClientSupportShard(cid int) (proto.HighwayServiceClient, error) {
	// TODO(@0xbunyip): make sure peer is still connected
	peers := h.hmap.Peers[byte(cid)]
	if len(peers) == 0 {
		return nil, errors.Errorf("no route client with block for cid = %v", cid)
	}

	conn, err := h.hc.hwc.GetConnection(peers[0].ID)
	if err != nil {
		return nil, err
	}

	return proto.NewHighwayServiceClient(conn), nil
}

func (h *Manager) GetShardsConnected() []byte {
	return h.hmap.CopyConnected()
}
