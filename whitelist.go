package rpc2

import (
	"strings"
	"sync"
)

type IPWhitelist struct {
	prefix []string
	match  map[string]bool
	count  int
	enable bool
	sync.RWMutex
}

func NewIPWhitelist() *IPWhitelist {
	return new(IPWhitelist).FreeAccess()
}

func (this *IPWhitelist) IsAllowed(addr string) bool {
	this.RLock()
	defer this.RUnlock()
	if !this.enable {
		return true
	}

	for i := 0; i < this.count; i++ {
		if strings.HasPrefix(addr, this.prefix[i]) {
			return true
		}
	}
	return false
}

func (this *IPWhitelist) AllowIPPrefix(ipPrefix ...string) *IPWhitelist {
	this.Lock()
	defer this.Unlock()
	this.enable = true
	for _, ip := range ipPrefix {
		if len(ip) == 0 {
			continue
		}
		if !this.match[ip] {
			this.match[ip] = true
			this.prefix = append(this.prefix, ip)
			this.count++
		}
	}
	return this
}

func (this *IPWhitelist) OnlyLAN() *IPWhitelist {
	this.NoAccess()
	return this.AllowIPPrefix(
		"[",
		"127.",
		"192.168.",
		"10.",
	)
}

func (this *IPWhitelist) FreeAccess() *IPWhitelist {
	this.Lock()
	defer this.Unlock()
	this.prefix = []string{}
	this.match = map[string]bool{}
	this.enable = false
	return this
}

func (this *IPWhitelist) NoAccess() *IPWhitelist {
	return this.Clean()
}

func (this *IPWhitelist) Clean() *IPWhitelist {
	this.Lock()
	defer this.Unlock()
	this.prefix = []string{}
	this.match = map[string]bool{}
	this.enable = true
	return this
}
