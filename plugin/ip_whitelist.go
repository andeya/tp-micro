package plugin

import (
	"errors"
	"net"
	"strings"
	"sync"

	"github.com/henrylee2cn/rpc2"
)

type IPWhitelistPlugin struct {
	match  map[string]bool
	prefix map[string]bool
	enable bool
	sync.RWMutex
}

func NewIPWhitelistPlugin() *IPWhitelistPlugin {
	return new(IPWhitelistPlugin).FreeAccess()
}

// Name returns plugin name.
func (ipWhitelist *IPWhitelistPlugin) Name() string {
	return "IPWhitelistPlugin"
}

var _ rpc2.IPostConnAcceptPlugin = new(IPWhitelistPlugin)

func (ipWhitelist *IPWhitelistPlugin) PostConnAccept(codecConn rpc2.ServerCodecConn) error {
	ip, _, _ := net.SplitHostPort(codecConn.RemoteAddr().String())
	if !ipWhitelist.IsAllowed(ip) {
		return errors.New("not allowed client ip: " + ip)
	}
	return nil
}

func (ipWhitelist *IPWhitelistPlugin) IsAllowed(addr string) bool {
	ipWhitelist.RLock()
	defer ipWhitelist.RUnlock()
	if !ipWhitelist.enable || ipWhitelist.match[addr] {
		return true
	}
	for ipPrefix := range ipWhitelist.prefix {
		if strings.HasPrefix(addr, ipPrefix) {
			return true
		}
	}
	return false
}

func (ipWhitelist *IPWhitelistPlugin) CancelAllow(pattern ...string) {
	if len(pattern) == 0 {
		return
	}

	ipWhitelist.Lock()
	defer ipWhitelist.Unlock()

	ipWhitelist.enable = true

	for _, ip := range pattern {
		ip = strings.TrimSpace(ip)
		length := len(ip)
		if length == 0 {
			continue
		}
		if !strings.HasSuffix(ip, "*") {
			delete(ipWhitelist.match, ip)
			continue
		}
		if length == 1 {
			go ipWhitelist.Clean()
			return
		}
		delete(ipWhitelist.prefix, ip[:length-1])
	}
}

func (ipWhitelist *IPWhitelistPlugin) Allow(pattern ...string) *IPWhitelistPlugin {
	if len(pattern) == 0 {
		return ipWhitelist
	}

	ipWhitelist.Lock()
	defer ipWhitelist.Unlock()

	ipWhitelist.enable = true

	for _, ip := range pattern {
		ip = strings.TrimSpace(ip)
		length := len(ip)
		if length == 0 {
			continue
		}
		if !strings.HasSuffix(ip, "*") {
			ipWhitelist.match[ip] = true
			continue
		}
		if length == 1 {
			go ipWhitelist.FreeAccess()
			return ipWhitelist
		}
		ipWhitelist.prefix[ip[:length-1]] = true
	}
	return ipWhitelist
}

func (ipWhitelist *IPWhitelistPlugin) OnlyLAN() *IPWhitelistPlugin {
	ipWhitelist.NoAccess()
	return ipWhitelist.Allow(
		"[*",
		"127.*",
		"192.168.*",
		"10.*",
	)
}

func (ipWhitelist *IPWhitelistPlugin) FreeAccess() *IPWhitelistPlugin {
	ipWhitelist.Lock()
	defer ipWhitelist.Unlock()
	ipWhitelist.prefix = map[string]bool{}
	ipWhitelist.match = map[string]bool{}
	ipWhitelist.enable = false
	return ipWhitelist
}

func (ipWhitelist *IPWhitelistPlugin) NoAccess() *IPWhitelistPlugin {
	return ipWhitelist.Clean()
}

func (ipWhitelist *IPWhitelistPlugin) Clean() *IPWhitelistPlugin {
	ipWhitelist.Lock()
	defer ipWhitelist.Unlock()
	ipWhitelist.prefix = map[string]bool{}
	ipWhitelist.match = map[string]bool{}
	ipWhitelist.enable = true
	return ipWhitelist
}
