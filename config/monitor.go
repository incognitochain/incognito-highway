package config

import (
	"encoding/json"
	"highway/common"
	"time"
)

type Reporter struct {
	conf *ProxyConfig

	name      string
	marshaler json.Marshaler
}

func (r *Reporter) Start(_ time.Duration) {
	r.name = "config"
	supports := []int{}
	for _, s := range r.conf.SupportShards {
		supports = append(supports, int(s))
	}
	data := map[string]interface{}{
		"port":           r.conf.ProxyPort,
		"masternode":     r.conf.Masternode,
		"bootstrap":      r.conf.Bootstrap,
		"shard_supports": supports,
	}
	r.marshaler = common.NewDefaultMarshaler(data)
}

func (r *Reporter) ReportJSON() (string, json.Marshaler, error) {
	return r.name, r.marshaler, nil
}

func NewReporter(conf *ProxyConfig) *Reporter {
	return &Reporter{conf: conf}
}
