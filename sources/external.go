package sources

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "flag"
)

var HostsFile = flag.String("hosts_file", "hosts", "specify the cadvisors location")

type ExternalSource struct {
    redis redisSource
}

func (self *ExternalSource) getRedisHosts() (RedisHosts, error) {
    fi, err := os.Stat(*HostsFile)
    if err != nil {
        return nil, err
    }
    if fi.Size() == 0 {
        return RedisHosts{}, nil
    }
    contents, err := ioutil.ReadFile(*HostsFile)
    if err != nil {
        return nil, err
    }

    type config map[string]map[string]string
    var c config
    err = json.Unmarshal(contents, &c)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal contents of file %s. Error: %s", HostsFile, err)
    }

    redisConfig, ok := c["redis"]
    if !ok {
        return nil, fmt.Errorf("failed to find redis config of file %s. Error: %s", HostsFile, err)
    }
    var redisHosts = make(RedisHosts)
    for k, v := range redisConfig {
        redisHosts[k] = RedisHost(v)
    }
    return redisHosts, nil
}

func (self *ExternalSource) GetInfo() ([]SourceData, error) {
    return self.redis.getAllInfo()
}


func newExternalSource() (Source, error) {
    if _, err := os.Stat(*HostsFile); err != nil {
        return nil, fmt.Errorf("Cannot stat hosts_file %s. Error: %s", *HostsFile, err)
    }
    source := new(ExternalSource)
    hosts, err := source.getRedisHosts()
    if err != nil {
        return nil, err
    }
    source.redis = newRedisSource(hosts)
    return source, nil
}
