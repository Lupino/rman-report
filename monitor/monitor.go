package monitor

import (
    "time"
    "flag"
    "strconv"
    "strings"
    "github.com/golang/glog"
    "github.com/bigdatadev/goryman"
)

var riemannAddress = flag.String("riemann_address", "localhost:5555", "specify the riemann server location")
var sampleInterval = flag.Duration("interval", 5*time.Second, "Interval between sampling (default: 5s)")

type riemannMonitor struct {
    client *goryman.GorymanClient
}

func NewRiemannMonitor() *riemannMonitor{
    client := goryman.NewGorymanClient(*riemannAddress)
    if err := client.Connect(); err != nil {
        glog.Fatalf("unable to connect to riemann: %s", err)
    }
    return &riemannMonitor{client:client}
}

func (rm riemannMonitor) HandleStat(serverType, hostname string, stat Stat) {
    evt := &goryman.Event{
        Service: serverType + " " + stat.Name,
        Host:    hostname,
        // State:   stat.Value,
        Tags:    []string{serverType},
        // Description: stat.Value,
    }

    if serverType == "redis" {
        if strings.Contains("rdb_last_bgsave_status aof_last_bgrewrite_status stat", stat.Name) {
            evt.State = stat.Value
        } else {
            evt.Description = stat.Value
        }

        if strings.Contains("used_memory_human used_memory_peak_human", stat.Name) {
            evt.Metric, _ = strconv.ParseFloat(stat.Value[:len(stat.Value)-1], 64)
        }
    }

    if serverType == "memcached" {
        if strings.Contains("version libevent", stat.Name) {
            evt.Description = stat.Value
        } else if strings.Contains("rusage_user rusage_system", stat.Name) {
            evt.Metric, _ = strconv.ParseFloat(stat.Value, 64)

        } else if strings.Contains("stat", stat.Name) {
            evt.State = stat.Value
        } else {
            evt.Metric, _ = strconv.Atoi(stat.Value)
        }
    }

    err := rm.client.SendEvent(evt)
    if err != nil {
        glog.Fatalf("unable to write to riemann: %s", err)
    }
}


func (rm riemannMonitor) Run() {
    source, _ := NewSource(rm)
    source.Monitor()
}

func (rm riemannMonitor) NewTicker() *time.Ticker {
    return time.NewTicker(*sampleInterval)
}
