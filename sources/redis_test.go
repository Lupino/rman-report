package sources

import (
    "fmt"
    "testing"
)

func TestRedisInfo(t *testing.T) {
    pool := newRedisPool("127.0.0.1:6379")
    stats, err := getRedisInfo(pool)
    fmt.Printf("%v, %v\n", stats, err)
}
