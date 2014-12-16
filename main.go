package main

import (
    "time"
    "flag"
    "strconv"
    "strings"
    "github.com/golang/glog"
    "github.com/bigdatadev/goryman"
    "github.com/Lupino/rman-report/sources"
)

var riemannAddress = flag.String("riemann_address", "localhost:5555", "specify the riemann server location")
var sampleInterval = flag.Duration("interval", 5*time.Second, "Interval between sampling (default: 5s)")

func main() {
    defer glog.Flush()
    flag.Parse()

    source, err := sources.NewSource()
    if err != nil {
        glog.Fatalf("unable to setup source: %s", err)
    }
    // Setting up the Riemann client
    r := goryman.NewGorymanClient(*riemannAddress)
    err = r.Connect()
    if err != nil {
        glog.Fatalf("unable to connect to riemann: %s", err)
    }
    //defer r.Close()
    // Setting up the ticker
    ticker := time.NewTicker(*sampleInterval).C
    for {
        select {
        case <-ticker:
            // Make the call to get all the possible data points
            data, err := source.GetInfo()
            if err != nil {
                glog.Fatalf("unable to retrieve machine data: %s", err)
            }
            // Start dumping data into riemann
            // Loop into each ContainerInfo
            // Get stats
            // Push into riemann
            for _, rs := range data {
                for _, stat := range rs.Stats {
                    evt := &goryman.Event{
                        Service: rs.Name + " " + stat.Name,
                        Host:    rs.Hostname,
                        // State:   stat.Value,
                        Tags:    []string{rs.Name},
                        // Description: stat.Value,
                    }

                    if rs.Name == "redis" {
                        if strings.Contains("rdb_last_bgsave_status aof_last_bgrewrite_status stat", stat.Name) {
                            evt.State = stat.Value
                        } else {
                            evt.Description = stat.Value
                        }

                        if strings.Contains("used_memory_human used_memory_peak_human", stat.Name) {
                            evt.Metric, _ = strconv.ParseFloat(stat.Value[:len(stat.Value)-1], 64)
                        }
                    }

                    if rs.Name == "memcached" {
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

                    err = r.SendEvent(evt)
                    if err != nil {
                        glog.Fatalf("unable to write to riemann: %s", err)
                    }
                }
            }
        }
    }
}
