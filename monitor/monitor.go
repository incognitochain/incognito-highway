package monitor

import (
	"encoding/json"
	"time"
)

type Monitor interface {
	Start(timestep time.Duration)
	ReportJSON() (string, json.Marshaler, error)
}
