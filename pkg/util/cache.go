package util

import (
	"fmt"
	"sync"
)

type CacheStatus string

type GetCacheDataFunc = func() (interface{}, error)

const (
	ReadyCacheStatus 	CacheStatus = "Ready"
	FetchingCacheStatus CacheStatus = "Fetching"
	EmptyCacheStatus 	CacheStatus = "Empty"
)

// Cache design to be able to save async data in lazy mode (request the data only if someone ask it) and provide it to any goroutine thread 
// for example, we have the following function 
// 
//    func GetSomeData(id string) Data {
//       client := clientCache.Get()
//       return client.GetDataById(id)
//    }
// 
// and in another area, we call it in parallel
// 
//    go GetSomeData(id1)
//    go GetSomeData(id2)
//    go GetSomeData(id3)
// 
// the cache will keep that the function "clientCache.Get()" will be call to the data only once
type Cache interface  {
	Refresh() (interface{}, error)
	Get() (interface{}, error)
	Status() CacheStatus
}

type cacheImpl struct {
	get_data GetCacheDataFunc
	data interface{}
	err error
	status CacheStatus
	waitChan chan GetCacheDataFunc
	waitersCounter int
	mutex sync.Mutex
}

func NewCache(get GetCacheDataFunc) Cache {
	cache := cacheImpl {
		status: EmptyCacheStatus,
		get_data: get,
		waitChan: make(chan func ()(interface{}, error)),
	}
	return &cache
}

func (c *cacheImpl) wait() (interface{}, error) {
	return (<- c.waitChan)()
}

func (c *cacheImpl) Get() (interface{}, error) {
	// need to ensure that only one thread check the status to prevent more than one to call c.Refresh
	c.mutex.Lock()
	switch c.status {
	case EmptyCacheStatus:
		c.status = FetchingCacheStatus
		c.mutex.Unlock()
		return c.Refresh()
	case FetchingCacheStatus:
		c.waitersCounter++
		c.mutex.Unlock()
		return c.wait()
	case ReadyCacheStatus:
		c.mutex.Unlock()
		return c.data, c.err
	default: 
		c.mutex.Unlock()
		return nil, fmt.Errorf("[Cache] Unknown cache status: %s", c.status)
	}
}

func (c *cacheImpl) Status() CacheStatus {
	return c.status
}


func (c *cacheImpl) Refresh() (interface{}, error) {
	c.mutex.Lock()
	c.status = FetchingCacheStatus
	c.mutex.Unlock()
	data, err := c.get_data()
	c.mutex.Lock()
	c.data = data
	c.err = err
	c.status = ReadyCacheStatus
	for i := 0; i < c.waitersCounter; i++ {
		c.waitChan <- (func() (interface{}, error) { return data, err })
	}
	c.waitersCounter = 0
	c.mutex.Unlock()
	return data, err
}