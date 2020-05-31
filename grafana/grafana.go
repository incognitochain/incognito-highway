package grafana

import (
	"context"
	"fmt"
	"net/http"
	"syscall"

	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

func NewLog(
	uID string,
	networkinfo string,
	dbURL string,
) *GrafanaLog {
	taginfo := fmt.Sprintf("hwsystem,uID=%v,networkinfo=%v", uID, networkinfo)
	return &GrafanaLog{
		tags:        make(map[string]interface{}),
		param:       make(map[string]interface{}),
		uID:         uID,
		networkinfo: networkinfo,
		dbURL:       dbURL,
		fixedTag:    taginfo,
	}
}

func (gl *GrafanaLog) Start() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		idle0, total0 := getCPUSample()
		var m runtime.MemStats
		for _ = range ticker.C {
			l := gl.CopyLog()
			idle1, total1 := getCPUSample()
			idleTicks := float64(idle1 - idle0)
			totalTicks := float64(total1 - total0)
			cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks
			runtime.ReadMemStats(&m)
			//disk usage
			fs := syscall.Statfs_t{}
			err := syscall.Statfs("./", &fs)
			if err == nil {
				All := fs.Blocks * uint64(fs.Bsize)
				Free := fs.Bfree * uint64(fs.Bsize)
				Used := All - Free
				l.Add("disk", fmt.Sprintf("%.2f", float64(Used*100)/float64(All)))
			}
			l.Add("cpu", fmt.Sprintf("%.2f", cpuUsage), "mem", m.Sys>>20)
			ngr := runtime.NumGoroutine()
			l.Add("numgoroutine", fmt.Sprintf("%d", ngr))
			idle0, total0 = getCPUSample()
			l.Write()
		}
	}()
}

type GrafanaLog struct {
	tags  map[string]interface{}
	param map[string]interface{}
	sync.RWMutex
	dbURL       string
	networkinfo string
	uID         string
	fixedTag    string
}

func (gl *GrafanaLog) CopyLog(p ...interface{}) *GrafanaLog {
	nl := (&GrafanaLog{
		tags:        make(map[string]interface{}),
		param:       make(map[string]interface{}),
		dbURL:       gl.dbURL,
		networkinfo: gl.networkinfo,
		uID:         gl.uID,
		fixedTag:    gl.fixedTag,
	}).Add(p...)
	gl.RLock()
	for k, v := range gl.param {
		nl.param[k] = v
	}
	for k, v := range gl.tags {
		nl.tags[k] = v
	}
	gl.RUnlock()
	return nl
}

func (gl *GrafanaLog) Add(p ...interface{}) *GrafanaLog {
	if len(p) == 0 || len(p)%2 != 0 {
		// fmt.Println(len(p))
		return gl
	}
	gl.Lock()
	defer gl.Unlock()
	for i, v := range p {
		if i%2 == 0 {
			gl.param[v.(string)] = p[i+1]
		}
	}
	return gl
}

func (gl *GrafanaLog) AddTags(p ...interface{}) *GrafanaLog {
	if len(p) == 0 || len(p)%2 != 0 {
		return gl
	}
	gl.Lock()
	defer gl.Unlock()
	for i, v := range p {
		if i%2 == 0 {
			gl.tags[v.(string)] = p[i+1]
		}
	}
	return gl
}

func (gl *GrafanaLog) AllFields() string {
	mt := ""
	gl.RLock()
	defer gl.RUnlock()
	for k, v := range gl.param {
		mt += fmt.Sprintf("%s=%v,", k, v)
	}
	if len(mt) > 0 {
		mt = mt[:len(mt)-1]
	}
	return mt
}

func (gl *GrafanaLog) AllTags() string {
	mt := ""
	gl.RLock()
	defer gl.RUnlock()
	for k, v := range gl.tags {
		mt += fmt.Sprintf("%s=%v,", k, v)
	}
	if len(mt) > 0 {
		mt = mt[:len(mt)-1]
		mt = fmt.Sprintf(",%s", mt)
	}
	return mt
}

func (gl *GrafanaLog) Write() {
	bodyStr := fmt.Sprintf("%s%s %s %v", gl.fixedTag, gl.AllTags(), gl.AllFields(), time.Now().UnixNano())
	go makeRequest(bodyStr, gl.dbURL)
}

func (gl *GrafanaLog) WriteContent(content string) {
	go makeRequest(content, gl.dbURL)
}

func (gl *GrafanaLog) GetFixedTag() string {
	return gl.fixedTag
}

func makeRequest(bodyStr, dbURL string) {
	body := strings.NewReader(bodyStr)
	// body := bytes.NewReader([]byte(bodyStr))
	req, err := http.NewRequest(http.MethodPost, dbURL, body)
	req.ContentLength = int64(len(bodyStr))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		logger.Debug("Create Request failed with err: ", err)
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	client := &http.Client{}
	client.Do(req)
}
