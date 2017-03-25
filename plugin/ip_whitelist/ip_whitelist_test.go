package ip_whitelist

import (
	"testing"
)

func TestEmptyIPWhitelistPlugin(t *testing.T) {
	w := NewIPWhitelistPlugin()
	w.Allow([]string{}...)
	if !w.IsAllowed("202.0.123.12") {
		t.Log("not allowed")
	} else {
		t.Log("allowed")
	}
}

func TestIPWhitelistPlugin(t *testing.T) {
	w := NewIPWhitelistPlugin()
	w.Allow("9*")
	if !w.IsAllowed("202.0.123.12") {
		t.Log("not allowed")
	} else {
		t.Log("allowed")
	}
}
