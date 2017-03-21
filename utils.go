package rpc2

import (
	"net"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

func RealRemoteAddr(req *http.Request) string {
	var ip string
	if ip = req.Header.Get("X-Real-IP"); len(ip) == 0 {
		if ip = req.Header.Get("X-Forwarded-For"); len(ip) == 0 {
			ip, _, _ = net.SplitHostPort(req.RemoteAddr)
		}
	}
	return ip
}

// SnakeString converts the accepted string to a snake string (XxYy to xx_yy)
func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}

// CamelString converts the accepted string to a camel string (xx_yy to XxYy)
func CamelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

// ObjectName gets the type name of the object
func ObjectName(i interface{}) string {
	v := reflect.ValueOf(i)
	t := v.Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(v.Pointer()).Name()
	}
	return t.String()
}

var nameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_\./]+$`)

func CheckSname(sname string) error {
	if sname == "" || sname == "/" || !nameRegexp.MatchString(sname) {
		return ErrInvalidPath.Format(sname)
	}
	return nil
}
