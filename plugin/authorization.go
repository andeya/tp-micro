package rpc2

import (
	"net/rpc"
	"net/url"

	"github.com/henrylee2cn/rpc2"
)

type (
	AuthorizationPlugin struct {
		token             string // Authorization token
		tag               string // extra tag for Authorization
		authorizationFunc AuthorizationFunc
	}

	// AuthorizationFunc defines a method type which handles Authorization info
	AuthorizationFunc func(token string, tag string, serviceMethod string) error
)

var _ rpc2.Plugin = new(AuthorizationPlugin)

func NewServerAuthorization(fn AuthorizationFunc) *AuthorizationPlugin {
	return &AuthorizationPlugin{
		authorizationFunc: fn,
	}
}

func NewClientAuthorization(token string, tag string) *AuthorizationPlugin {
	return &AuthorizationPlugin{
		token: token,
		tag:   tag,
	}
}

func (auth *AuthorizationPlugin) String() string {
	return url.Values{
		"auth_token": []string{auth.token},
		"auth_tag":   []string{auth.tag},
	}.Encode()
}

func (auth *AuthorizationPlugin) PostReadRequestHeader(req *rpc.Request) error {
	if auth.authorizationFunc == nil {
		return nil
	}

	s := rpc2.ParseServiceMethod(req.ServiceMethod)

	v, err := s.ParseQuery()
	if err != nil {
		return err
	}

	return auth.authorizationFunc(v.Get("auth_token"), v.Get("auth_tag"), s.Path)
}

func (*AuthorizationPlugin) PostReadRequestBody(_ interface{}) error {
	return nil
}
