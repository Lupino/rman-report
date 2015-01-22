package monitor

import (
    "time"
)

type Stat struct {
    Name string
    Host string
    Metric interface{}
    Desc   string
    State string
}

type SourceHost struct {
    Port int    `json:"port"`
    Host string `json:"host"`
}

type Source interface {
    Monitor()
}

func NewSource(m Monitor) (Source, error) {
    return newExternalSource(m)
}

type Monitor interface {
    HandleStat(string, Stat)
    Run()
    NewTicker() *time.Ticker
}
