package health

import (
	"encoding/json"
	"fmt"
	"highway/common"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

type Reporter struct {
	name string

	idle     uint64
	total    uint64
	cpuUsage float64
}

func (r *Reporter) Start(_ time.Duration) {
	healthTimestep := 5 * time.Second
	for ; true; <-time.Tick(healthTimestep) {
		r.updateSample()
	}
}

func (r *Reporter) ReportJSON() (string, json.Marshaler, error) {
	data := map[string]interface{}{}
	data["cpu"] = r.cpuUsage
	marshaler := common.NewDefaultMarshaler(data)
	return r.name, marshaler, nil
}

func NewReporter() *Reporter {
	return &Reporter{
		name: "health",
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
