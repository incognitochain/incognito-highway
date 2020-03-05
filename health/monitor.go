package health

import (
	"encoding/json"
	"fmt"
	"highway/common"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

type Reporter struct {
	name string

	idle     uint64
	total    uint64
	cpuUsage float64

	memUsage uint64

	ifmons   []*Iface
	netStats map[string]NetStat // Sample of Tx/Rx
}

func (r *Reporter) Start(_ time.Duration) {
	healthTimestep := time.NewTicker(60 * time.Second)
	defer healthTimestep.Stop()
	for ; true; <-healthTimestep.C {
		r.updateSample()
		r.memProfile()
	}
}

func (r *Reporter) ReportJSON() (string, json.Marshaler, error) {
	data := map[string]interface{}{}
	data["cpu"] = r.cpuUsage
	data["mem"] = r.memUsage
	data["net"] = r.netStats
	marshaler := common.NewDefaultMarshaler(data)
	return r.name, marshaler, nil
}

func NewReporter() *Reporter {
	// Get all net interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		logger.Warn(err)
	}
	ifmons := []*Iface{}
	for _, iface := range ifaces {
		ifmon, err := NewIfmon(iface.Name, "")
		if err != nil {
			logger.Warn(err)
			continue
		}
		ifmons = append(ifmons, ifmon)
	}

	return &Reporter{
		name:     "health",
		ifmons:   ifmons,
		netStats: map[string]NetStat{},
	}
}

func (r *Reporter) updateSample() {
	// CPU sample
	idle0, total0 := r.idle, r.total
	idle, total := getCPUSample()
	idleTicks := float64(idle - idle0)
	totalTicks := float64(total - total0)
	r.cpuUsage = 100 * (totalTicks - idleTicks) / totalTicks
	r.idle, r.total = idle, total

	// Memory sample
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	r.memUsage = m.Sys >> 20

	// Net sample
	for _, ifmon := range r.ifmons {
		tx, rx, err := ifmon.GetStats()
		if err != nil {
			continue
		}
		r.netStats[ifmon.Name()] = NetStat{Tx: tx, Rx: rx}
	}
}

func (r *Reporter) memProfile() {
	f, err := os.Create("mem.prof")
	if err != nil {
		logger.Warn("Could not create memory profile: ", err)
		return
	}
	defer f.Close()
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		logger.Warn("Could not write memory profile: ", err)
	}
}

func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}

type NetStat struct {
	Rx uint64
	Tx uint64
}
