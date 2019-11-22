package main

import (
	"fmt"
	"log"
	"net/http"
)

func StartMonitorServer(port int) {
	//run http server
	http.HandleFunc("/admin", handler)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {

}
