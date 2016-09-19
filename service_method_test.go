package rpc2

import (
	"testing"
)

func TestParseServiceMethod(t *testing.T) {
	s, err := ParseServiceMethod("a/b/c.d?e=1&f=2")
	t.Logf("%#v %#v\n", s, err)
	t.Logf("%v\n", s)
}

func TestParseQuery(t *testing.T) {
	s := &ServiceMethod{Query: "e=1&f=2"}
	v, err := s.ParseQuery()
	t.Logf("%#v %#v\n", v, err)
}

func TestGroups(t *testing.T) {
	s := &ServiceMethod{Service: "a/b/c"}
	t.Logf("%v\n", s.Groups())
	s.Service = "c"
	t.Logf("%v\n", s.Groups())
}
