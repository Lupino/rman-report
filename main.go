package main

import (
    "os"
    "flag"
    "os/signal"
    "github.com/golang/glog"
    "github.com/Lupino/rman-report/monitor"
)

func main() {
    defer glog.Flush()
    flag.Parse()

    m := monitor.NewRiemannMonitor()
    m.Run()
    s := make(chan os.Signal, 1)
    signal.Notify(s, os.Interrupt, os.Kill)
    <-s
}
