package rpc2

import (
	"testing"
)

func TestParseServiceMethod(t *testing.T) {
	s := ParseServiceMethod("a/b/c.d?auth_tag=ljqr3456l%26%26asdlj%E5%B0%B1%E5%95%A6%E7%9C%8B%E7%94%B5%E8%A7%86%E6%84%9F%E8%A7%89%25%3D&auth_token=adflj%E8%AE%B0%E5%BD%95%EF%BC%9B%EF%BC%9B%E6%8E%A5%E5%95%8A%E5%9C%B0%E6%96%B9%26ljf%E4%B8%9C%E6%96%B9%E5%B7%A8%E9%BE%99%3D%3D+%E5%95%8A%E4%B8%A4%E5%9C%B0%E5%88%86%E5%B1%85")
	t.Logf("%#v\n", s)

	t.Logf("String(): %v\n", s)

	t.Logf("Groups(): %#v\n", s.Groups())

	v, err := s.ParseQuery()
	t.Logf("ParseQuery(): %#v   %#v\n", v, err)

	service, method := s.Split()
	t.Logf("Split(): %#v   %#v\n", service, method)
}
