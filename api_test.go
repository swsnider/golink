package golink

import (
	"net/url"
	"testing"
)

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
