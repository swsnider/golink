package golink

import "time"

type API struct {
	BaseURL, KeyID, VCode string
	Cache                 APICache
}

type APICache interface {
	Get(k string) interface{}
	Put(k string, v interface{}, duration time.Duration)
}

type InMemoryCacheValue struct {
	V          interface{}
	Expiration time.Time
}

type InMemoryAPICache map[string]*InMemoryCacheValue

func (c InMemoryAPICache) Get(k string) interface{} {
	r := c[k]
	if r == nil {
		return nil
	}
	if time.Now().After(r.Expiration) {
		c[k] = nil
		return nil
	}
	return r.V
}

func (c InMemoryAPICache) Set(k string, v interface{}, duration time.Duration) {
	c[k] = &InMemoryCacheValue{V: v, Expiration: time.Now().Add(duration)}
}

//Request a specific path from the EVE API.
func (a *API) Get(path string, params map[string]string) error
