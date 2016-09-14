package rpc2

import (
	"strings"
	"sync"
)

type IPWhitelist struct {
	prefix []string
	match  map[string]bool
	count  int
	sync.RWMutex
}

var ipWhitelist = func() *IPWhitelist {
	that := new(IPWhitelist).Clean()
	that.AddAllowedPrefix(
		"[",
		"127.",
		"192.168.",
		"10.",
	)
	return that
}()

func (this *IPWhitelist) allowAccess(addr string) bool {
	this.RLock()
	defer this.RUnlock()
	for i := 0; i < this.count; i++ {
		if strings.HasPrefix(addr, this.prefix[i]) {
			return true
		}
	}
	return false
}

func (this *IPWhitelist) AddAllowedPrefix(ipPrefix ...string) *IPWhitelist {
	this.Lock()
	defer this.Unlock()
	if this.match[""] {
		this.Clean()
	}
	for _, ip := range ipPrefix {
		if !this.match[ip] {
			this.match[ip] = true
			this.prefix = append(this.prefix, ip)
			this.count++
		}
	}
	return this
}

func (this *IPWhitelist) Clean() *IPWhitelist {
	this.Lock()
	defer this.Unlock()
	this.prefix = []string{}
	this.match = map[string]bool{}
	return this
}

func (this *IPWhitelist) AllowLAN() *IPWhitelist {
	return this.AddAllowedPrefix(
		"[",
		"127.",
		"192.168.",
		"10.",
	)
}

func (this *IPWhitelist) AllowAny() *IPWhitelist {
	return this.AddAllowedPrefix("")
}

// Add the client ip that is allowed to connect,
// LAN match are always allowed.
func AddAllowedIPPrefix(ipPrefix ...string) {
	ipWhitelist.AddAllowedPrefix(ipPrefix...)
}

func CleanIPWhitelist() {
	ipWhitelist.Clean()
}

func AllowLANAccess() {
	ipWhitelist.AllowLAN()
}

func AllowAnyAccess() {
	ipWhitelist.AllowAny()
}
