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
	data := map[string]interface{}{
		"validators_connecting": validators,
		// "pending_connecting":    pendings,
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
