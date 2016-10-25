package rpc2

import (
	"strings"
	"sync"
)

type IPWhitelist struct {
	match  map[string]bool
	prefix map[string]bool
	enable bool
	sync.RWMutex
}

func NewIPWhitelist() *IPWhitelist {
	return new(IPWhitelist).FreeAccess()
}

func (this *IPWhitelist) IsAllowed(addr string) bool {
	this.RLock()
	defer this.RUnlock()
	if !this.enable || this.match[addr] {
		return true
	}
	for ipPrefix := range this.prefix {
		if strings.HasPrefix(addr, ipPrefix) {
			return true
		}
	}
	return false
}

func (this *IPWhitelist) CancelAllow(pattern ...string) {
	if len(pattern) == 0 {
		return
	}

	this.Lock()
	defer this.Unlock()

	this.enable = true

	for _, ip := range pattern {
		ip = strings.TrimSpace(ip)
		length := len(ip)
		if length == 0 {
			continue
		}
		if !strings.HasSuffix(ip, "*") {
			delete(this.match, ip)
			continue
		}
		if length == 1 {
			go this.Clean()
			return
		}
		delete(this.prefix, ip[:length-1])
	}
}

func (this *IPWhitelist) Allow(pattern ...string) *IPWhitelist {
	if len(pattern) == 0 {
		return this
	}

	this.Lock()
	defer this.Unlock()

	this.enable = true

	for _, ip := range pattern {
		ip = strings.TrimSpace(ip)
		length := len(ip)
		if length == 0 {
			continue
		}
		if !strings.HasSuffix(ip, "*") {
			this.match[ip] = true
			continue
		}
		if length == 1 {
			go this.FreeAccess()
			return this
		}
		this.prefix[ip[:length-1]] = true
	}
	return this
}

func (this *IPWhitelist) OnlyLAN() *IPWhitelist {
	this.NoAccess()
	return this.Allow(
		"[*",
		"127.*",
		"192.168.*",
		"10.*",
	)
}

func (this *IPWhitelist) FreeAccess() *IPWhitelist {
	this.Lock()
	defer this.Unlock()
	this.prefix = map[string]bool{}
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
	this.prefix = map[string]bool{}
	this.match = map[string]bool{}
	this.enable = true
	return this
}
