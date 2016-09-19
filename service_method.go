package rpc2

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

type ServiceMethod struct {
	Service string // service name
	Method  string // method name
	Query   string // query values, without '?', only can be use in 'ReadRequestHeader(*rpc.Request)'.
}

func ParseServiceMethod(serviceMethodString string) (*ServiceMethod, error) {
	dot := strings.Index(serviceMethodString, ".")
	if dot < 0 || dot+1 == len(serviceMethodString) {
		return nil, errors.New("rpc: service/method request ill-formed: " + serviceMethodString)
	}

	var serviceMethod = &ServiceMethod{
		Service: serviceMethodString[:dot],
		Method:  serviceMethodString[dot+1:],
	}

	boundary := strings.Index(serviceMethod.Method, "?")
	if boundary == 0 {
		return nil, errors.New("rpc: service/method request ill-formed: " + serviceMethodString)
	}
	if boundary > 0 {
		serviceMethod.Query = serviceMethod.Method[boundary+1:]
		serviceMethod.Method = serviceMethod.Method[:boundary]
	}

	return serviceMethod, nil
}

// 'Path' return whole name for index.
func (s *ServiceMethod) Path() string {
	return s.Service + "." + s.Method
}

// ParseQuery parses the URL-encoded query string and returns
// a map listing the values specified for each key.
// ParseQuery always returns a non-nil map containing all the
// valid query parameters found; err describes the first decoding error
// encountered, if any.
func (s *ServiceMethod) ParseQuery() (url.Values, error) {
	return url.ParseQuery(s.Query)
}

// 'String' return original name.
func (s *ServiceMethod) String() string {
	return s.Service + "." + s.Method + "?" + s.Query
}

// 'Groups' return router groups prefixes.
func (s *ServiceMethod) Groups() []string {
	var prefixes []string
	for k, v := range s.Service {
		if v == '/' {
			prefixes = append(prefixes, s.Service[:k])
		}
	}
	return prefixes
}

var nameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func nameCharsFunc(r rune) bool {
	// A-Z
	if r >= 65 && r <= 90 {
		return false
	}
	// a-z
	if r >= 97 && r <= 122 {
		return false
	}
	// _
	if r == 95 {
		return false
	}
	// 0-9
	if r >= 48 && r <= 57 {
		return false
	}

	return true
}
