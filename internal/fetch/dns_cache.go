package fetch

import (
	"sync"
)

var dnsCache = &DnsCache{
	m: &sync.Map{},
}

type DnsCache struct {
	m *sync.Map
}

func (c *DnsCache) Load(key string) (string, bool) {
	v, ok := c.m.Load(key)
	if !ok {
		return "", false
	}
	return v.(string), ok
}

func (c *DnsCache) Store(key string, val string) {
	c.m.Store(key, val)
}

func (c *DnsCache) Delete(key string) {
	c.m.Delete(key)
}
