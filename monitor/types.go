package monitor

import (
    "time"
)

type Stat struct {
    Name string
    Value string
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
    HandleStat(string, string, Stat)
    Run()
    NewTicker() *time.Ticker
}
