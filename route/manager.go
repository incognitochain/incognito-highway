package route

import (
	"context"
	"encoding/json"
	"highway/common"
	"highway/process"
	"highway/proto"
	hmap "highway/route/hmap"
	"highway/rpcserver"
	"math/rand"
	"time"

	p2pgrpc "github.com/incognitochain/go-libp2p-grpc"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/stathat/consistent"
)

type Manager struct {
	ID peer.ID

	Hmap *hmap.Map
	hc   *Connector
	host host.Host
}

func NewManager(
	supportShards []byte,
	bootstrap []string,
	masternode peer.ID,
	h host.Host,
	prtc *p2pgrpc.GRPCProtocol,
	rpcUrl string,
) *Manager {
	p := peer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	hmap := hmap.NewMap(p, supportShards, rpcUrl)

	hw := &Manager{
		ID:   h.ID(),
		Hmap: hmap,
		hc: NewConnector(
			h,
			prtc,
			hmap,
			&process.GlobalPubsub,
			masternode,
		),
		host: h,
	}

	go hw.keepHighwayConnection(bootstrap)

	// Start highway connector event loop
	go hw.hc.Start()
	return hw
}

// keepHighwayConnection maintains the map of all highways including
// both the ones we want and don't want to connect to.
// This func keep the map up to date but doesn't make the decision of
// which highway to connect. That job is performed by route.Connector
// Every few seconds, we call a random known highway, get the list of
// all highways and merge with ours. If any connection is dead for too
// long, we remove it from the highway map.
// Note that other highways won't return a disconnected highway when queried.
// Therefore after awhile, the offline highway will be removed from all maps.
func (h *Manager) keepHighwayConnection(bootstrap []string) {
	if len(bootstrap) > 0 && len(bootstrap[0]) > 0 {
		hInfos, err := h.getListHighwaysFromPeer(bootstrap[0])
		if err != nil {
			logger.Warnf("Failed getting list of highways from peer %+v, err = %+v", bootstrap[0], err)
		} else {
			h.updateHighwayMap(hInfos)
		}
	}

	// We refer to each highway using their peerID. This is not totally correct
	// when the same highway is rerun with different supported shards.
	// Therefore, if we rerun a highway (in a short period of time) with different
	// supported shards, we need to use a new peerID.
	lastSeen := map[peer.ID]time.Time{}
	watchTimestep := 30 * time.Second
	removeDeadline := time.Duration(30 * time.Minute)
	for ; true; <-time.Tick(watchTimestep) {
		// Map from peerID to RPCUrl
		urls := h.Hmap.CopyRPCUrls()

		// Get a random highway from map to get list highway
		if randomPeer, ok := getRandomPeer(h.Hmap.CopyPeersMap(), []peer.ID{h.ID}); ok {
			url := urls[randomPeer.ID]
			hInfos, err := h.getListHighwaysFromPeer(url)
			if err != nil {
				logger.Warnf("Failed getting list highway from peer %+v, url = %+v err = %+v", randomPeer, url, err)
			} else {
				h.updateHighwayMap(hInfos)
			}
		}

		peerMap := h.Hmap.CopyPeersMap()
		for _, peers := range peerMap {
			for _, p := range peers {
				// No need to connect to ourself
				if p.ID == h.ID {
					continue
				}

				if h.host.Network().Connectedness(p.ID) == network.Connected {
					lastSeen[p.ID] = time.Now()
					continue
				}

				if _, ok := lastSeen[p.ID]; !ok {
					lastSeen[p.ID] = time.Now()
				}

				// Not connected, remove from map it's been too long
				if time.Since(lastSeen[p.ID]) > removeDeadline {
					h.Hmap.RemovePeer(p)
					delete(lastSeen, p.ID)
				}
			}
		}
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

func (h *Manager) updateHighwayMap(hInfos []HighwayInfo) {
	for _, b := range hInfos {
		// Get addr info from string
		addr, err := multiaddr.NewMultiaddr(b.AddrInfo)
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
		h.Hmap.AddPeer(*addrInfo, ss, b.RPCUrl)
	}
}

// getRandomPeer returns the a random peer, exluding some peers to
// make sure we don't connect to ourself
func getRandomPeer(peers map[byte][]peer.AddrInfo, excludes []peer.ID) (peer.AddrInfo, bool) {
	includes := []peer.AddrInfo{}
	for _, addrs := range peers {
		for _, p := range addrs {
			found := false
			for _, e := range excludes {
				if p.ID == e {
					found = true
					break
				}
			}
			if !found {
				includes = append(includes, p)
			}
		}
	}
	if len(includes) == 0 {
		return peer.AddrInfo{}, false
	}
	return includes[rand.Intn(len(includes))], true
}

func (h *Manager) getListHighwaysFromPeer(addr string) ([]HighwayInfo, error) {
	resps, err := rpcserver.DiscoverHighWay(addr, []string{"all"})
	if err != nil {
		return nil, err
	}

	// NOTE: assume each highway supports all shards
	// TODO(@0xbunyip): for v2, return correct list of supported shards for each addrinfo
	if _, ok := resps["all"]; !ok {
		return nil, errors.Errorf("no highway return from bootnode %s", addr)
	}
	ss := make([]int, common.NumberOfShard+1)
	for i := 0; i < common.NumberOfShard; i++ {
		ss[i] = i
	}
	ss[common.NumberOfShard] = int(common.BEACONID)

	hInfos := []HighwayInfo{}
	for _, resp := range resps["all"] {
		hInfos = append(hInfos, HighwayInfo{
			AddrInfo:      resp.Libp2pAddr,
			RPCUrl:        resp.RPCUrl,
			SupportShards: ss,
		})
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
	// TODO(@0xbunyip): connect to all highway same shards and
	// one highway of other shard
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
	if h.Hmap.IsConnectedToShard(sid) {
		return nil
	}

	logger.Info("Connecting to chain ", sid)
	highways := h.Hmap.Peers[sid]
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
func (h *Manager) GetClientSupportShard(cid int) (proto.HighwayServiceClient, peer.ID, error) {
	// TODO(@0xbunyip): make sure peer is still connected
	peers := h.Hmap.Peers[byte(cid)]
	if len(peers) == 0 {
		return nil, peer.ID(""), errors.Errorf("no route client with block for cid = %v", cid)
	}

	// TODO(@0xbunyip): get peer randomly here?
	pid := peers[0].ID
	conn, err := h.hc.hwc.GetConnection(pid)
	if err != nil {
		return nil, pid, err
	}

	return proto.NewHighwayServiceClient(conn), pid, nil
}

func (h *Manager) GetShardsConnected() []byte {
	return h.Hmap.CopyConnected()
}

type HighwayInfo struct {
	AddrInfo      string
	RPCUrl        string
	SupportShards []int
}
