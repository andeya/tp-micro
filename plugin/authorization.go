package rpc2

import (
	"net/rpc"
	"net/url"

	"github.com/henrylee2cn/rpc2"
)

type (
	AuthorizationPlugin struct {
		Token             string `json:"auth_token"` // Authorization token
		Tag               string `json:"auth_tag"`   // extra tag for Authorization
		AuthorizationFunc `json:"-"`
	}

	// AuthorizationFunc defines a method type which handles Authorization info
	AuthorizationFunc func(token string, tag string, serviceMethod string) error
)

var _ rpc2.Plugin = new(AuthorizationPlugin)

func (auth *AuthorizationPlugin) String() string {
	return url.Values{
		"auth_token": []string{auth.Token},
		"auth_tag":   []string{auth.Tag},
	}.Encode()
}

func (auth *AuthorizationPlugin) PostReadRequestHeader(req *rpc.Request) error {
	if auth.AuthorizationFunc == nil {
		return nil
	}

	s := rpc2.ParseServiceMethod(req.ServiceMethod)

	v, err := s.ParseQuery()
	if err != nil {
		return err
	}

	return auth.AuthorizationFunc(v.Get("auth_token"), v.Get("auth_tag"), s.Path)
}

func (*AuthorizationPlugin) PostReadRequestBody(_ interface{}) error {
	return nil
}
