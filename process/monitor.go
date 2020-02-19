package process

import (
	"encoding/json"
	"highway/chaindata"
	"highway/common"
	"sync"
	"time"

	peer "github.com/libp2p/go-libp2p-peer"
)

type Reporter struct {
	name string

	chainData    *chaindata.ChainData
	networkState struct {
		sync.RWMutex
		state     chaindata.NetworkState
		hwofpeers map[string]string
	}
}

func (r *Reporter) Start(_ time.Duration) {
	stateTimestep := time.NewTicker(5 * time.Second)
	defer stateTimestep.Stop()
	for ; true; <-stateTimestep.C {
		r.updateNetworkState()
	}
}

func (r *Reporter) ReportJSON() (string, json.Marshaler, error) {
	r.networkState.RLock()
	defer r.networkState.RUnlock()
	data := map[string]interface{}{
		"network_state": r.networkState.state,
		"hwid_of_peers": r.networkState.hwofpeers,
	}
	marshaler := common.NewDefaultMarshaler(data)
	return r.name, marshaler, nil
}

func (r *Reporter) updateNetworkState() {
	r.networkState.Lock()
	r.networkState.state = r.chainData.CopyNetworkState()
	r.networkState.hwofpeers = r.chainData.CurrentNetworkState.GetAllHWIDInfo()
	r.networkState.Unlock()
}

func NewReporter(chainData *chaindata.ChainData) *Reporter {
	r := &Reporter{
		chainData: chainData,
		name:      "process",
	}
	r.networkState.state = chaindata.NetworkState{}
	r.networkState.hwofpeers = map[string]string{}
	r.networkState.RWMutex = sync.RWMutex{}
	return r
}

type PubsubInfo struct {
	Info   map[string][]peer.ID
	Locker sync.RWMutex
}
