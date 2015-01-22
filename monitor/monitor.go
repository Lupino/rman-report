package monitor

import (
    "time"
    "flag"
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

func (rm riemannMonitor) HandleStat(serverType string, stat Stat) {
    evt := &goryman.Event{
        Service: serverType + " " + stat.Name,
        Host:    stat.Host,
        State:   stat.State,
        Tags:    []string{serverType},
        Description: stat.Desc,
        Metric: stat.Metric,
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
