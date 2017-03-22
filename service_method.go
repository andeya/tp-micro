package rpc2

import (
	"bytes"
	"net/url"
	"path"
)

type (
	// IServiceMethod ServiceMethod
	IServiceMethod interface {
		// Path returns ServiceMethod without query string.
		Path() string
		// Service returns ServiceMethod's service.
		Service() string
		// Method returns ServiceMethod's method.
		Method() string
		// Query returns ServiceMethod's query params.
		Query() url.Values
		// Encode returns ServiceMethod's string.
		Encode() string
		// ParseInto parses the ServiceMethod string into IServiceMethod.
		ParseInto(string) error
		// Reset reset IServiceMethod.
		Reset(service string, method string, query url.Values)
	}

	// ServiceMethodFunc creates IServiceMethod
	ServiceMethodFunc func() IServiceMethod

	// URLServiceMethod implements a IServiceMethod in URL format.
	URLServiceMethod struct {
		service string
		method  string
		query   url.Values
	}
)

// NewURLServiceMethod creates URL IServiceMethod.
func NewURLServiceMethod() IServiceMethod {
	return new(URLServiceMethod)
}

// Path returns ServiceMethod without query string.
func (u *URLServiceMethod) Path() string {
	return path.Join(u.service, u.method)
}

// Service returns ServiceMethod's service.
func (u *URLServiceMethod) Service() string {
	return u.service
}

// Method returns ServiceMethod's method.
func (u *URLServiceMethod) Method() string {
	return u.method
}

// Query returns ServiceMethod's query params.
func (u *URLServiceMethod) Query() url.Values {
	return u.query
}

// Encode returns ServiceMethod's string.
func (u *URLServiceMethod) Encode() (serviceMethod string) {
	sm := url.URL{
		Path:     path.Join("/", SnakeString(u.service), SnakeString(u.method)),
		RawQuery: u.query.Encode(),
	}
	return sm.String()
}

// ParseInto parses the ServiceMethod string into IServiceMethod.
func (u *URLServiceMethod) ParseInto(serviceMethod string) error {
	sm, err := url.Parse(serviceMethod)
	if err != nil {
		return ErrServiceMethod.Format(serviceMethod)
	}
	p := []byte(sm.Path)
	i := bytes.LastIndexByte(p, '/')

	if i <= 0 ||
		// method can not be ''
		i == len(p)-1 {
		return ErrServiceMethod.Format(serviceMethod)
	}

	u.method = string(p[i+1:])

	p = p[:i]
	// service can not be ''
	if len(p) == 0 {
		return ErrServiceMethod.Format(serviceMethod)
	}
	u.service = string(p)
	u.query = sm.Query()
	return nil
}

// Reset reset IServiceMethod.
func (u *URLServiceMethod) Reset(service string, method string, query url.Values) {
	u.service = path.Join("/", SnakeString(service))
	u.method = SnakeString(method)
	u.query = query
}
