package rpc2

import (
	"net/url"
	"regexp"
	"strings"
)

type ServiceMethod struct {
	Path  string // whole name(service + method) for index.
	Query string // query values, without '?', only can be use in 'ReadRequestQuery(*rpc.Request)'.
}

func ParseServiceMethod(serviceMethodString string) *ServiceMethod {
	boundary := strings.Index(serviceMethodString, "?")

	var serviceMethod = new(ServiceMethod)

	if boundary < 0 {
		serviceMethod.Path = serviceMethodString

	} else {
		serviceMethod.Path = serviceMethodString[:boundary]
		serviceMethod.Query = serviceMethodString[boundary+1:]
	}

	return serviceMethod
}

// 'Split' return service name and method name.
func (s *ServiceMethod) Split() (service, method string) {
	dot := strings.LastIndex(s.Path, ".")

	if dot <= 0 || dot+1 == len(s.Path) {
		service = s.Path
		return
	}

	service = s.Path[:dot]
	method = s.Path[dot+1:]

	return
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
	return s.Path + "?" + s.Query
}

// 'Groups' return router groups prefixes.
func (s *ServiceMethod) Groups() []string {
	var prefixes []string
	for k, v := range s.Path {
		if v == '/' {
			prefixes = append(prefixes, s.Path[:k])
		}
	}
	service, _ := s.Split()
	return append(prefixes, service+".")
}

var nameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_\.]+$`)
