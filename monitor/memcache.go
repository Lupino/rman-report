package monitor

import (
    "fmt"
    "net"
    "bytes"
    "bufio"
    "strings"
    "strconv"
    "github.com/golang/glog"
)

var MEMCACHED_STRING_VALUES = []string{
    "curr_connections",
    "limit_maxbytes",
    "bytes",
    "curr_items",
}

type MemcachedHost string
type MemcachedHosts map[string]MemcachedHost

type memcachedSource struct {
    hosts MemcachedHosts
    m Monitor
}

func newMemcachedSource(hosts MemcachedHosts, monitor Monitor) memcachedSource {
    return memcachedSource{hosts: hosts, m: monitor}
}

func (ms memcachedSource) monitor() {
    for hostname, host := range ms.hosts {
        go ms.monitorOne(hostname, host)

    }
}

func (ms memcachedSource) monitorOne(hostname string, host MemcachedHost) {
    ticker := ms.m.NewTicker().C
    for {
        select {
        case <-ticker:
            stats, err := getMamcachedStat(host)
            var stat = Stat{
                Name: "state",
                State: "ok",
            }
            if err != nil {
                stat.State = "error"
            }
            stats = append(stats, stat)
            for _, stat := range stats {
                stat.Host = hostname
                go ms.m.HandleStat("memcached", stat)
            }
        }
    }
}

func getMamcachedStat(host MemcachedHost) ([]Stat, error) {
    conn, err := net.Dial("tcp", string(host))
    if err != nil {
        glog.Errorf("connect memcached %s fail: %s", host, err)
        return nil, err
    }
    fmt.Fprintf(conn, "stats\n")
    retval := make(map[string]string)
    reader := bufio.NewReader(conn)
    for {
        line, _, err := reader.ReadLine()
        if err != nil {
            break
        }
        line = bytes.Trim(line, " ")
        if bytes.HasPrefix(line, []byte("END")) {
            break
        }
        parts := strings.SplitN(string(line), " ", 3)
        if len(parts) == 3 {
            retval[parts[1]] = parts[2]
        }
    }

    stats := make([]Stat, len(MEMCACHED_STRING_VALUES))

    limitMaxBytes, _ := strconv.Atoi(retval["limit_maxbytes"])
    usageBytes, _ := strconv.Atoi(retval["bytes"])

    var usage float64

    if limitMaxBytes > 0 && limitMaxBytes > usageBytes {
        usage = float64(usageBytes) / float64(limitMaxBytes)
    }

    var state = "ok"
    if usage > 0.95 {
        state = "critical"
    } else if usage > 0.9 {
        state = "warning"
    }

    usageStat := Stat{
        Name: "usage",
        Metric: usage,
        State: state,
    }

    for i, name := range MEMCACHED_STRING_VALUES {
        desc := retval[name]
        var stat = Stat{
            Name: name,
            // Desc: desc,
            State: "ok",
        }

        stat.Metric, _ = strconv.Atoi(desc)

        stats[i] =  stat
    }
    stats = append(stats, usageStat)
    return stats, nil
}
