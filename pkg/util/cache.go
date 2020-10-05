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
	mux sync.Mutex
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
	c.mux.Lock()
	switch c.status {
	case EmptyCacheStatus:
		c.status = FetchingCacheStatus
		c.mux.Unlock()
		return c.Refresh()
	case FetchingCacheStatus:
		c.waitersCounter++
		c.mux.Unlock()
		return c.wait()
	case ReadyCacheStatus:
		c.mux.Unlock()
		return c.data, c.err
	default: 
		c.mux.Unlock()
		return nil, fmt.Errorf("[Cache] Unknown cache status: %s", c.status)
	}
}

func (c *cacheImpl) Status() CacheStatus {
	return c.status
}


func (c *cacheImpl) Refresh() (interface{}, error) {
	c.mux.Lock()
	c.status = FetchingCacheStatus
	c.mux.Unlock()
	data, err := c.get_data()
	c.mux.Lock()
	c.data = data
	c.err = err
	c.status = ReadyCacheStatus
	for i := 0; i < c.waitersCounter; i++ {
		c.waitChan <- (func() (interface{}, error) { return data, err })
	}
	c.waitersCounter = 0
	c.mux.Unlock()
	return data, err
}