package golink

import (
	"bytes"
	"code.google.com/p/go-etree"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type APIFetcher interface {
  Get(path string, params url.Values, c *APICredentials) (etree.Element, error)
}

type API struct {
	BaseURL                          string
	Cache                            APICache
	lastCachedUntil, lastCurrentTime time.Time
	Client                           URLFetcher
}

type CredentialedAPI struct {
  api APIFetcher
  credentials APICredentials
}

// Returns a new API object, filling in defaults as needed.
func NewAPI(base string, c APICache, uf URLFetcher) *API {
	if base == "" {
		base = "api.eveonline.com"
	}
	if c == nil {
		c = make(InMemoryAPICache)
	}
	if uf == nil {
		uf = http.PostForm
	}
	return &API{BaseURL: base, Cache: c, Client: uf}
}

func NewCredentialedAPI(a APIFetcher, c APICredentials) *CredentialedAPI {
  return &CredentialedAPI{api: a, credentials: c}
}

func (a *CredentialedAPI) Get(path string, params url.Values) (etree.Element, error) {
  return a.api.Get(path, params, &a.credentials)
}

type APICredentials struct {
	KeyID, VCode string
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

//Request a specific path from the EVE API.
func (a *API) Get(path string, params url.Values, c *APICredentials) (etree.Element, error) {
	if c != nil {
		params["keyID"] = []string{c.KeyID}
		params["vCode"] = []string{c.VCode}
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

	tree, err := etree.Parse(bytes.NewBuffer(response))
	if err != nil {
		return nil, err
	}
	tree = tree.Find("eveapi")
	elem := tree.Find("currentTime")
	if elem == nil {
		return nil, fmt.Errorf("Unable to parse currentTime.")
	}
	currentTime, err := parseEveTs(elem.Text())
	if err != nil {
		return nil, err
	}
	a.lastCurrentTime = currentTime
	elem = tree.Find("cachedUntil")
	if elem == nil {
		return nil, fmt.Errorf("Unable to parse cachedUntil.")
	}
	expiresTime, err := parseEveTs(elem.Text())
	if err != nil {
		return nil, err
	}
	a.lastCachedUntil = expiresTime

	if !cached {
		a.Cache.Put(cacheKey, response, expiresTime.Sub(currentTime))
	}

	xmlErr := tree.Find("error")
	if xmlErr != nil {
		code, _ := xmlErr.Get("code")
		return nil, fmt.Errorf("API reported error with code %v and message \"%v\"", code, xmlErr.Text())
	}

	return tree.Find("result"), nil
}

func genCacheKey(path string, params url.Values) string {
	ks := make([]string, 0)
	for k, v := range params {
		ks = append(ks, fmt.Sprint(k, v))
	}
	sort.StringSlice(ks).Sort()
	return fmt.Sprintf("%v#%v", path, ks)
}
