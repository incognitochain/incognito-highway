package database

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	cachedata *cache.Cache
)

func init() {
	cachedata = cache.New(5*time.Second, 10*time.Second)
}

func MarkData(data []byte) {
	cachedata.Add(string(data), nil, 5*time.Second)
}

func IsMarkedData(data []byte) bool {
	_, isExist := cachedata.Get(string(data))
	return isExist
}
