package sources

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "flag"
)

var HostsFile = flag.String("hosts_file", "hosts", "specify the cadvisors location")

type EntryPoints struct {
    Redis     RedisHosts     `json:"redis"`
    Memcached MemcachedHosts `json:"memcached"`
}

type ExternalSource struct {
    redis redisSource
    memcached memcachedSource
    entrypoints EntryPoints
}

func (self *ExternalSource) loadHosts() (error) {
    fi, err := os.Stat(*HostsFile)
    if err != nil {
        return err
    }
    if fi.Size() == 0 {
        return nil
    }
    contents, err := ioutil.ReadFile(*HostsFile)
    if err != nil {
        return err
    }

    err = json.Unmarshal(contents, &self.entrypoints)
    if err != nil {
        return fmt.Errorf("failed to unmarshal contents of file %s. Error: %s", HostsFile, err)
    }
    return nil
}

func (self *ExternalSource) GetInfo() ([]SourceData, error) {
    var sourceData = make([]SourceData, 0)
    if self.entrypoints.Redis != nil {
        stdatas, _ := self.redis.getAllInfo()
        for _, stdata := range stdatas {
            sourceData = append(sourceData, stdata)
        }
    }

    if self.entrypoints.Memcached != nil {
        stdatas, _ := self.memcached.getAllInfo()
        for _, stdata := range stdatas {
            sourceData = append(sourceData, stdata)
        }
    }

    return sourceData, nil
}


func newExternalSource() (Source, error) {
    if _, err := os.Stat(*HostsFile); err != nil {
        return nil, fmt.Errorf("Cannot stat hosts_file %s. Error: %s", *HostsFile, err)
    }
    source := new(ExternalSource)
    err := source.loadHosts()
    if err != nil {
        return nil, err
    }
    source.redis = newRedisSource(source.entrypoints.Redis)
    source.memcached = newMemcachedSource(source.entrypoints.Memcached)
    return source, nil
}
