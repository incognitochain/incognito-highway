package chain

import (
	"encoding/json"
	"fmt"
	"highway/grafana"
	"io/ioutil"
	"strings"
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

	watchingPubkeys map[string]position
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
	fmt.Println("debugging sending pos:", pos.cid, pos.id)
	fmt.Println("debugging sending id:", pos.cid, fmt.Sprintf("\"%s\"", pinfo.ID.String()))

	w.data[pos] = watchInfo{
		pid:       pinfo.ID,
		connected: 1,
		ip:        pinfo.ip,
	}
	w.pos[pinfo.ID] = pos
}

func (w *watcher) processOutPeer(pid peer.ID) {
	if pos, ok := w.pos[pid]; ok {
		fmt.Println("debugging processOutPeer:", pos)
		if winfo, ok := w.data[pos]; ok {
			fmt.Println("debugging processOutPeer found")
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
	fmt.Println("debugging len:", len(w.data))
	fmt.Println("debugging data:", w.data)
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
		} else {
			fmt.Println("debugging remove disconnected node:", pos, winfo)
		}
	}
	w.data = data

	fmt.Println("debugging points len:", len(points))
	for i, p := range points {
		fmt.Println("debugging points:", i, len(p), p)
	}
	content := strings.Join(points, "\n")
	fmt.Printf("debugging content: %d %s\n", len(content), content)
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
		w.watchingPubkeys[pubkey] = position{
			cid: cid,
			id:  id,
		}
	}
}
