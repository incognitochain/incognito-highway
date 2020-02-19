package monitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "net/http/pprof"

	"github.com/pkg/errors"
)

func StartMonitorServer(port int, timestep time.Duration, monitors []Monitor) {
	// Start all monitors
	for _, m := range monitors {
		go m.Start(timestep)
	}

	// Run http server
	m := &poller{Monitors: monitors, reports: map[string]interface{}{}}
	go m.start(timestep)
	http.Handle("/monitor", m)
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			logger.Errorf("Error in ListenAndServe: %s", err)
		}
	}()
}

func (p *poller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(p.reports)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (p *poller) start(timestep time.Duration) {
	t := time.NewTicker(timestep)
	defer t.Stop()
	for ; true; <-t.C {
		reports := map[string]interface{}{}
		for _, m := range p.Monitors {
			name, value, err := m.ReportJSON()
			if err != nil {
				fmt.Println(errors.WithStack(err))
				continue
			}
			reports[name] = value
		}

		p.reports = reports // No need to lock, only save a reference
	}
}

type poller struct {
	Monitors []Monitor

	reports map[string]interface{}
}
