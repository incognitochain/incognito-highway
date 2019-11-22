package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func StartMonitorServer(port int) {
	// Run http server
	m := &monitor{}
	go m.start()
	http.Handle("/monitor", m)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		// TODO(@0xbunyip): change to logger and prevent os.Exit()
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func (m *monitor) ServeHTTP(w ResponseWriter, r *Request) {
	profile := &monitor{}
	js, err := json.Marshal(profile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (m *monitor) start() {
	for ; true; <-time.Tick(5 * time.Second) {
	}
}

type monitor struct {
}
