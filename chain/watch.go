package chain

import (
	"fmt"
	"highway/grafana"
	"strings"
	"time"

	peer "github.com/libp2p/go-libp2p-core/peer"
)

type watcher struct {
	gralog   *grafana.GrafanaLog
	inPeers  chan PeerInfoWithIP
	outPeers chan peer.ID

	data map[position]watchInfo
	pos  map[peer.ID]position
}

type PeerInfoWithIP struct {
	PeerInfo
	ip string
}

func newWatcher(gralog *grafana.GrafanaLog) *watcher {
	return &watcher{
		inPeers:  make(chan PeerInfoWithIP, 100),
		outPeers: make(chan peer.ID, 100),
		data:     make(map[position]watchInfo),
		pos:      make(map[peer.ID]position),
		gralog:   gralog,
	}
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

var watchingPubkeys = map[string]position{
	"1DiqMXWSgNrBcEGPWt2coFVdmjwNnrG8de43SMziHFvff6A6NgZQe1A1a7Qc9DiRUHUV9vVgYCtpvfFszfTCC2J31SwoKzDiBpGXitLk66umDr9ECN2rTGGqmChF4tcG7R634JF4JDUhL63su6Fpq7ooKWPHMetAjbsnF3VNLX68VxD8nQecs": position{
		cid: 0,
		id:  0,
	},
	"1SVKWw7tUcCUUguytHav5nU3WaZdZYSADXBvGhXcd3eSPXvUEbS5vKkVKEj8bbh4ZkVMkaNgSZSz1KpBh97TjWJL3aiF6gCKKVvnDfjEG8KHZ9r8ZByYtHjA8UkHwPRUVZMvzmsDQLmVSjgSKm1dnUUAN3kkTqZYyY1K7vw1hqGxbyiMVBg3T": position{
		cid: 0,
		id:  1,
	},
}

func (w *watcher) processInPeer(pinfo PeerInfoWithIP) {
	pos := getWatchingPosition(pinfo.Pubkey)
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
		}

		points = append(points, buildPoint(w.gralog.GetFixedTag(), tags, fields))
	}

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

func isWatching(pubkey string) bool {
	pos := getWatchingPosition(pubkey)
	return pos.id != -1
}

func getWatchingPosition(pubkey string) position {
	if pos, ok := watchingPubkeys[pubkey]; ok {
		return pos
	}
	return position{-1, -1}
}

func (w *watcher) markPeer(pinfo PeerInfo, ip string) {
	if !isWatching(pinfo.Pubkey) {
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
