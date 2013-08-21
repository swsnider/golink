package golink

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type nopCloser struct {
	io.Reader
}

func TestTimeParse(t *testing.T) {
	ts, err := parseEveTs("2012-06-12 12:04:33")
	if err != nil {
		t.Error(err)
	}
	if ts.Unix() != 1339502673 {
		t.Errorf("Parsed time incorrect. Got: %+v", ts)
	}
}

func TestCache(t *testing.T) {
	cache := make(InMemoryAPICache)
	cache.Put("foo", []byte("bar"), 3600)
	v := cache.Get("foo")
	if string(v) != "bar" {
		t.Errorf("Incorrect cache value for foo: %v", v)
	}
}

func TestExpire(t *testing.T) {
	cache := make(InMemoryAPICache)
	cache.Put("foo", []byte("bar"), -1)
	v := cache.Get("foo")
	if string(v) != "" {
		t.Errorf("Incorrect cache value for foo: %v", v)
	}
}

func TestCacheKey(t *testing.T) {
	if genCacheKey("foo/bar", url.Values{"a": []string{"1"}, "b": []string{"2"}}) != genCacheKey("foo/bar", url.Values{"b": []string{"2"}, "a": []string{"1"}}) {
		t.Error("genCacheKey does not sort map keys.")
	}
}

func TestGet(t *testing.T) {
	a := NewAPI("", nil, URLTestFetcher)
	tree, err := a.Get("blagh", url.Values{}, nil)
	if err != nil {
		t.Error(err)
	}
	rowset := tree.Find("rowset")
	rows := rowset.FindAll("row")
	if len(rows) != 2 {
		t.Errorf("Got wrong number of rows: %v", len(rows))
	}
	if first(rows[0].Get("foo")) != "bar" {
		t.Error("Incorrect attribute parsing.")
	}
	if a.lastCachedUntil.Unix() != 1258563931 {
		t.Errorf("Incorrect cachedUntil. Got %v", a.lastCachedUntil)
	}
	if a.lastCurrentTime.Unix() != 1255885531 {
		t.Errorf("Incorrect currentTime. Got %v", a.lastCurrentTime)
	}
}

func TestGetCached(t *testing.T) {
	a := NewAPI("", nil, URLErrFetcher)
	a.Cache.Put(genCacheKey("blagh", url.Values{}), []byte(testXML), time.Hour)
	tree, err := a.Get("blagh", url.Values{}, nil)
	if err != nil {
		t.Error(err)
	}
	rowset := tree.Find("rowset")
	rows := rowset.FindAll("row")
	if len(rows) != 2 {
		t.Errorf("Got wrong number of rows: %v", len(rows))
	}
	if first(rows[0].Get("foo")) != "bar" {
		t.Error("Incorrect attribute parsing.")
	}
	if a.lastCachedUntil.Unix() != 1258563931 {
		t.Errorf("Incorrect cachedUntil. Got %v", a.lastCachedUntil)
	}
	if a.lastCurrentTime.Unix() != 1255885531 {
		t.Errorf("Incorrect currentTime. Got %v", a.lastCurrentTime)
	}
}

func TestGetErr(t *testing.T) {
	a := NewAPI("", nil, URLErrFetcher)
	_, err := a.Get("blagh", url.Values{}, nil)
	if err == nil {
		t.Error("Error XML was not parsed as an error!")
	}
	if a.lastCachedUntil.Unix() != 1258571131 {
		t.Errorf("Incorrect cachedUntil. Got %v (%v)", a.lastCachedUntil, a.lastCachedUntil.Unix())
	}
	if a.lastCurrentTime.Unix() != 1255885531 {
		t.Errorf("Incorrect currentTime. Got %v (%v)", a.lastCurrentTime, a.lastCurrentTime.Unix())
	}
}

func TestGetErrCached(t *testing.T) {
	a := NewAPI("", nil, URLTestFetcher)
	a.Cache.Put(genCacheKey("blagh", url.Values{}), []byte(errXML), time.Hour)
	_, err := a.Get("blagh", url.Values{}, nil)
	if err == nil {
		t.Error("Error XML was not parsed as an error!")
	}
	if a.lastCachedUntil.Unix() != 1258571131 {
		t.Errorf("Incorrect cachedUntil. Got %v (%v)", a.lastCachedUntil, a.lastCachedUntil.Unix())
	}
	if a.lastCurrentTime.Unix() != 1255885531 {
		t.Errorf("Incorrect currentTime. Got %v (%v)", a.lastCurrentTime, a.lastCurrentTime.Unix())
	}
}

func (b nopCloser) Close() error {
	return nil
}

func URLTestFetcher(path string, params url.Values) (result *http.Response, err error) {
	return &http.Response{Body: &nopCloser{bytes.NewBufferString(testXML)}}, nil
}

func URLErrFetcher(path string, params url.Values) (result *http.Response, err error) {
	return &http.Response{Body: &nopCloser{bytes.NewBufferString(errXML)}}, nil
}

const testXML = `
<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
    <currentTime>2009-10-18 17:05:31</currentTime>
    <result>
        <rowset>
            <row foo="bar" />
            <row foo="baz" />
        </rowset>
    </result>
    <cachedUntil>2009-11-18 17:05:31</cachedUntil>
</eveapi>
`

const errXML = `
<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
    <currentTime>2009-10-18 17:05:31</currentTime>
    <error code="123">
        Test error message.
    </error>
    <cachedUntil>2009-11-18 19:05:31</cachedUntil>
</eveapi>
`
