package chain

import (
	"encoding/json"
	"fmt"
	"highway/grafana"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

type watcher struct {
	hwid     int
	gralog   *grafana.GrafanaLog
	inPeers  chan PeerInfoWithIP
	outPeers chan peer.ID

	data map[position]watchInfo
	pos  map[peer.ID]position

	posLocker sync.RWMutex

	watchingPubkeys map[string]position
	anchorK         map[int]string
}

type PeerInfoWithIP struct {
	PeerInfo
	ip string
}

func newWatcher(gralog *grafana.GrafanaLog, hwid int) *watcher {
	w := &watcher{
		hwid:            hwid,
		inPeers:         make(chan PeerInfoWithIP, 100),
		outPeers:        make(chan peer.ID, 100),
		data:            make(map[position]watchInfo),
		pos:             make(map[peer.ID]position),
		watchingPubkeys: make(map[string]position),
		gralog:          gralog,
		anchorK:         make(map[int]string),
	}
	w.readKeys()
	return w
}

type watchInfo struct {
	connected int
	pid       peer.ID
	ip        string
}

type position struct {
	cid int
	id  int
}

func (w *watcher) processInPeer(pinfo PeerInfoWithIP) {
	pos := getWatchingPosition(pinfo.Pubkey, w.watchingPubkeys)
	logger.Infof("Start watching peer: cid = %v, id = %v, ip = %v, pid = %s", pos.cid, pos.id, pinfo.ip, pinfo.ID.String())

	w.data[pos] = watchInfo{
		pid:       pinfo.ID,
		connected: 1,
		ip:        pinfo.ip,
	}
	w.setPeerPosition(pinfo.ID, pos)
}

func (w *watcher) processOutPeer(pid peer.ID) {
	pos, ok := w.getPeerPosition(pid)
	if ok {
		if winfo, ok := w.data[pos]; ok {
			logger.Infof("Stop watching peer: cid = %v, id = %v, ip = %v, pid = %s", pos.cid, pos.id, winfo.ip, pid.String())
			w.data[pos] = watchInfo{
				pid:       pid,
				connected: 0,
				ip:        winfo.ip,
			}
		}
	}
}

func (w *watcher) pushData() {
	if len(w.data) == 0 {
		return
	}

	points := []string{}
	for pos, winfo := range w.data {
		tags := map[string]interface{}{
			"watch_id": pos.id,
		}
		fields := map[string]interface{}{
			"watch_libp2p_id": fmt.Sprintf("\"%s\"", winfo.pid),
			"watch_cid":       pos.cid,
			"watch_connected": winfo.connected,
			"watch_ip":        fmt.Sprintf("\"%s\"", winfo.ip),
			"watch_hwid":      w.hwid,
		}

		points = append(points, buildPoint(w.gralog.GetFixedTag(), tags, fields))
	}

	// Remove disconnected nodes so other highway can report its status
	data := map[position]watchInfo{}
	for pos, winfo := range w.data {
		if winfo.connected == 1 {
			data[pos] = winfo
		}
	}
	w.data = data

	content := strings.Join(points, "\n")
	w.gralog.WriteContent(content)
}

func buildPoint(fixedTag string, tags map[string]interface{}, fields map[string]interface{}) string {
	point := fixedTag
	for key, val := range tags {
		point = point + "," + fmt.Sprintf("%s=%v", key, val)
	}
	if len(fields) > 0 {
		point = point + " "
	}
	firstField := true
	for key, val := range fields {
		if !firstField {
			point = point + ","
		}

		point = point + fmt.Sprintf("%s=%v", key, val)
		firstField = false
	}
	return fmt.Sprintf("%s %v", point, time.Now().UnixNano())
}

func (w *watcher) process() {
	if w.gralog == nil {
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case pinfo := <-w.inPeers:
			w.processInPeer(pinfo)

		case pid := <-w.outPeers:
			w.processOutPeer(pid)

		case <-ticker.C:
			w.pushData()
		}
	}
}

func isWatching(pubkey string, watchingPubkeys map[string]position) bool {
	pos := getWatchingPosition(pubkey, watchingPubkeys)
	return pos.id != -1
}

func getWatchingPosition(pubkey string, watchingPubkeys map[string]position) position {
	if pos, ok := watchingPubkeys[pubkey]; ok {
		return pos
	}
	return position{-1, -1}
}

func (w *watcher) markPeer(pinfo PeerInfo, ip string) {
	if !isWatching(pinfo.Pubkey, w.watchingPubkeys) {
		return
	}
	w.inPeers <- PeerInfoWithIP{
		pinfo,
		ip,
	}
}

func (w *watcher) unmarkPeer(pid peer.ID) {
	w.outPeers <- pid
}

func (w *watcher) readKeys() {
	keyData, err := ioutil.ReadFile("keylist.json")
	if err != nil {
		logger.Error(err)
		return
	}

	type AccountKey struct {
		PaymentAddress     string
		CommitteePublicKey string
	}

	type KeyList struct {
		Shard  map[int][]AccountKey
		Beacon []AccountKey
	}

	keylist := KeyList{}

	err = json.Unmarshal(keyData, &keylist)
	if err != nil {
		logger.Error(err)
		return
	}

	for cid, keys := range keylist.Shard {
		for id, committeeKey := range keys {
			k := &incognitokey.CommitteePublicKey{}
			if err := k.FromString(committeeKey.CommitteePublicKey); err != nil {
				logger.Error(err)
				continue
			}
			pubkey := k.GetMiningKeyBase58(common.BlsConsensus)
			if id == 21 {
				w.anchorK[cid] = pubkey
			}
			w.watchingPubkeys[pubkey] = position{
				cid: cid,
				id:  id,
			}
		}
	}

	cid := 255 // for beacon
	for id, committeeKey := range keylist.Beacon {
		k := &incognitokey.CommitteePublicKey{}
		if err := k.FromString(committeeKey.CommitteePublicKey); err != nil {
			logger.Error(err)
			continue
		}
		pubkey := k.GetMiningKeyBase58(common.BlsConsensus)
		logger.Infof("Beacon pubkey: %d %v", id, pubkey)
		if id == 6 {
			w.anchorK[cid] = pubkey
		}
		w.watchingPubkeys[pubkey] = position{
			cid: cid,
			id:  id,
		}
	}
}

func (w *watcher) getPeerPosition(pID peer.ID) (position, bool) {
	w.posLocker.RLock()
	pos, ok := w.pos[pID]
	w.posLocker.RUnlock()
	return pos, ok
}

func (w *watcher) setPeerPosition(pID peer.ID, pos position) {
	w.posLocker.Lock()
	w.pos[pID] = pos
	w.posLocker.Unlock()
}

func (w *watcher) getFixedPeersOfCID(cid int) []peer.ID {
	infos := []peer.ID{}
	w.posLocker.Lock()
	for pos, info := range w.data {
		if (pos.cid != cid) || (info.connected == 0) {
			continue
		} else {
			infos = append(infos, info.pid)
		}
	}
	w.posLocker.Unlock()
	return infos
}
