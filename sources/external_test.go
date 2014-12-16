package sources

import (
    "fmt"
    "testing"
)

func TestExternalSourceInfo(t *testing.T) {
    source := new(ExternalSource)
    hosts, err := source.getRedisHosts()
    if err != nil {
        t.Fatalf("%s", err)
    }
    source.redis = newRedisSource(hosts)
    sourceData, err := source.GetInfo()
    fmt.Printf("%s, %v\n", sourceData, err)
}
