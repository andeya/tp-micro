package plugin

import (
	"testing"
)

func TestEmptyIPWhitelist(t *testing.T) {
	w := NewIPWhitelist()
	w.Allow([]string{}...)
	if !w.IsAllowed("202.0.123.12") {
		t.Log("not allowed")
	} else {
		t.Log("allowed")
	}
}

func TestIPWhitelist(t *testing.T) {
	w := NewIPWhitelist()
	w.Allow("9*")
	if !w.IsAllowed("202.0.123.12") {
		t.Log("not allowed")
	} else {
		t.Log("allowed")
	}
}
