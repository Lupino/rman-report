package monitor

import (
    "fmt"
    "net"
    "bytes"
    "bufio"
    "strings"
    "github.com/golang/glog"
)

var MEMCACHED_STRING_VALUES = []string{
    // "uptime",
    // "version",
    // "libevent",
    // "pointer_size",
    "rusage_user",
    "rusage_system",
    "curr_connections",
    "total_connections",
    // "connection_structures",
    // "reserved_fds",
    // "cmd_get",
    // "cmd_set",
    // "cmd_flush",
    // "cmd_touch",
    // "get_hits",
    // "get_misses",
    // "delete_misses",
    // "delete_hits",
    // "incr_misses",
    // "incr_hits",
    // "decr_misses",
    // "decr_hits",
    // "cas_misses",
    // "cas_hits",
    // "cas_badval",
    // "touch_hits",
    // "touch_misses",
    // "auth_cmds",
    // "auth_errors",
    "bytes_read",
    "bytes_written",
    "limit_maxbytes",
    // "accepting_conns",
    // "listen_disabled_num",
    // "threads",
    // "conn_yields",
    // "hash_power_level",
    // "hash_bytes",
    // "hash_is_expanding",
    // "malloc_fails",
    // "bytes",
    // "curr_items",
    // "total_items",
    // "expired_unfetched",
    // "evicted_unfetched",
    // "evictions",
    // "reclaimed",
    // "crawler_reclaimed",
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
            if err != nil {
                stats = make([]Stat, 1)
                stats[0] = Stat{
                    Name: "stat",
                    Value: "error",
                }
            } else {
                stat := Stat{
                    Name: "stat",
                    Value: "ok",
                }
                stats = append(stats, stat)
            }

            for _, stat := range stats {
                go ms.m.HandleStat("memcached", hostname, stat)
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
    for i, str := range MEMCACHED_STRING_VALUES {
        stats[i] = Stat{
            Name: str,
            Value: retval[str],
        }
    }
    return stats, nil
}
