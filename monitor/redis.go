package monitor

import (
    "bytes"
    "bufio"
    "strings"
    "strconv"
    "github.com/golang/glog"
    "github.com/garyburd/redigo/redis"
)

var REDIS_STRING_VALUES = []string{
    "used_memory_human",
    "used_memory_peak_human",
}

type RedisHost string
type RedisHosts map[string]RedisHost

type redisSource struct {
    pools map[string]*redis.Pool
    m Monitor
}

func newRedisSource(hosts RedisHosts, m Monitor) redisSource {
    pools := make(map[string]*redis.Pool)
    for k, v := range hosts {
        pools[k] = newRedisPool(string(v))
    }
    return redisSource{pools:pools,m:m}
}

func (rs redisSource) monitor() {
    for k, pool := range rs.pools {
        go rs.monitorOne(k, pool)
    }
}

func (rs redisSource) monitorOne(hostname string, pool *redis.Pool) {
    ticker := rs.m.NewTicker().C
    for {
        select {
        case <-ticker:
            stats, err := getRedisInfo(pool)
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
                go rs.m.HandleStat("redis", stat)
            }
        }
    }
}

func newRedisPool(host_port string) *redis.Pool {
    pool := redis.NewPool(func() (conn redis.Conn, err error) {
        conn, err = redis.Dial("tcp", host_port)
        return
    }, 1)
    return pool
}


func getRedisInfo(pool *redis.Pool) ([]Stat, error) {
    conn := pool.Get()
    defer conn.Close()

    reply, err := redis.Bytes(conn.Do("info"))
    if err != nil {
        glog.Errorf("unable to request redis info: %s", err)
        return nil, err
    }
    reader := bufio.NewReader(bytes.NewReader(reply))
    retval := make(map[string]string)
    for {
        line, _, err := reader.ReadLine()
        if err != nil {
            break
        }
        line = bytes.Trim(line, " ")
        if bytes.HasPrefix(line, []byte("#")) || len(line) == 0 {
            continue
        }
        parts := strings.SplitN(string(line), ":", 2)
        if len(parts) == 2 {
            retval[parts[0]] = parts[1]
        }
    }

    stats := make([]Stat, len(REDIS_STRING_VALUES))
    for i, name := range REDIS_STRING_VALUES {
        desc := retval[name]
        var stat = Stat{
            Name: name,
            Desc: desc,
            State: "ok",
        }

        if strings.Contains("used_memory_human used_memory_peak_human", name) {
            metric, _ := strconv.ParseFloat(desc[:len(desc) - 1], 64)

            if strings.HasSuffix(desc, "M") {
                metric /= 1024
            }

            name += " (G)"

            stat.Metric = metric
        }
        stats[i] =  stat
    }
    return stats, nil
}
