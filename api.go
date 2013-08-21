package golink

import (
	"fmt"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type API struct {
	BaseURL, KeyID, VCode            string
	Cache                            APICache
	lastCachedUntil, lastCurrentTime time.Time
	Client URLFetcher
}

type URLFetcher func(path string, params url.Values) (result *http.Response, err error)

type APICache interface {
	Get(k string) []byte
	Put(k string, v []byte, duration time.Duration)
}

type InMemoryCacheValue struct {
	V          []byte
	Expiration time.Time
}

type InMemoryAPICache map[string]*InMemoryCacheValue

func (c InMemoryAPICache) Get(k string) []byte {
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

func (c InMemoryAPICache) Put(k string, v []byte, duration time.Duration) {
	c[k] = &InMemoryCacheValue{V: v, Expiration: time.Now().Add(duration)}
}

type EveError struct {
	Msg  string `xml:"chardata"`
	Code string `xml:"code,attr"`
}

type GenericAPIResponse struct {
	Result      []byte   `xml:"result,innerxml"`
	CurrentTime string   `xml:"currentTime"`
	CachedUntil string   `xml:"cachedUntil"`
	Error       EveError `xml:"error"`
}

func NewAPI() *API {
	return &API{}
}

//Request a specific path from the EVE API.
func (a *API) Get(path string, params url.Values) ([]byte, error) {
	if a.KeyID != "" {
		params["keyID"] = []string{a.KeyID}
		params["vCode"] = []string{a.VCode}
	}
	cacheKey := genCacheKey(path, params)
	response := a.Cache.Get(cacheKey)
	cached := response != nil
	if !cached {
		r, err := a.Client(fmt.Sprintf("https://%v/%v.xml.aspx", a.BaseURL, path), params)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()
		response, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
	}

	xmlResponse := GenericAPIResponse{}
	err := xml.Unmarshal(response, &xmlResponse)
	if err != nil {
		return nil, err
	}
	currentTime, err := parseEveTs(xmlResponse.CurrentTime)
	if err != nil {
		return nil, err
	}
	a.lastCurrentTime = currentTime
	expiresTime, err := parseEveTs(xmlResponse.CachedUntil)
	if err != nil {
		return nil, err
	}
	a.lastCachedUntil = expiresTime

	if !cached {
		a.Cache.Put(cacheKey, response, expiresTime.Sub(currentTime))
	}

	if xmlResponse.Error.Code != "" {
		return nil, fmt.Errorf("API reported error with code %v and message \"%v\"", xmlResponse.Error.Code, xmlResponse.Error.Msg)
	}

	return xmlResponse.Result, nil
}

func genCacheKey(path string, params url.Values) string {
	ks := make([]string, 0)
	for k, v := range params {
		ks = append(ks, fmt.Sprint(k, v))
	}
	sort.StringSlice(ks).Sort()
	return fmt.Sprintf("%v#%v", path, ks)
}

func parseEveTs(in string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", in)
}
