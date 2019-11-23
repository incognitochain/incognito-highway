package chain

import (
	"encoding/json"
	"highway/common"
	"time"
)

type Reporter struct {
	name string

	manager *Manager
}

func (r *Reporter) Start(_ time.Duration) {}

func (r *Reporter) ReportJSON() (string, json.Marshaler, error) {
	validators := r.manager.GetAllPeers()
	totalConns := r.manager.GetTotalConnections()
	data := map[string]interface{}{
		"peers":               validators,
		"inbound_connections": totalConns,
	}
	marshaler := common.NewDefaultMarshaler(data)
	return r.name, marshaler, nil
}

func NewReporter(manager *Manager) *Reporter {
	return &Reporter{
		manager: manager,
		name:    "chain",
	}
}
