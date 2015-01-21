package monitor

import (
    "bytes"
    "bufio"
    "strings"
    "github.com/golang/glog"
    "github.com/garyburd/redigo/redis"
)

var REDIS_STRING_VALUES = []string{
    // "redis_version",
    // "redis_git_sha1",
    // "redis_mode",
    // "os",
    // "multiplexing_api",
    // "gcc_version",
    // "run_id",
    "used_memory_human",
    "used_memory_peak_human",
    // "mem_allocator",
    // "rdb_last_bgsave_status",
    // "aof_last_bgrewrite_status",
    // "role",
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

func (rs redisSource) monitorOne(k string, pool *redis.Pool) {
    ticker := rs.m.NewTicker().C
    for {
        select {
        case <-ticker:
            stats, err := getRedisInfo(pool)
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
                go rs.m.HandleStat("redis", k, stat)
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
    for i, str := range REDIS_STRING_VALUES {
        stats[i] = Stat{
            Name: str,
            Value: retval[str],
        }
    }
    return stats, nil
}
