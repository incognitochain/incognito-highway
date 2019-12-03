package process

import (
	"encoding/json"
	"highway/common"
	"sync"
	"time"
)

type Reporter struct {
	name string

	chainData    *ChainData
	networkState struct {
		sync.RWMutex
		state NetworkState
	}
}

func (r *Reporter) Start(_ time.Duration) {
	stateTimestep := 5 * time.Second
	for ; true; <-time.Tick(stateTimestep) {
		r.updateNetworkState()
	}
}

func (r *Reporter) ReportJSON() (string, json.Marshaler, error) {
	r.networkState.RLock()
	defer r.networkState.RUnlock()
	data := map[string]interface{}{
		"network_state": r.networkState.state,
	}
	marshaler := common.NewDefaultMarshaler(data)
	return r.name, marshaler, nil
}

func (r *Reporter) updateNetworkState() {
	r.networkState.Lock()
	r.networkState.state = r.chainData.CopyNetworkState()
	r.networkState.Unlock()
}

func NewReporter(chainData *ChainData) *Reporter {
	r := &Reporter{
		chainData: chainData,
		name:      "process",
	}
	r.networkState.state = NetworkState{}
	r.networkState.RWMutex = sync.RWMutex{}
	return r
}
