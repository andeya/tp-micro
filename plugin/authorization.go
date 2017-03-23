package plugin

import (
	"errors"
	"net/rpc"
	"strings"

	"github.com/henrylee2cn/rpc2"
)

type (
	// AuthorizationPlugin authorization plugin
	AuthorizationPlugin struct {
		token             string // Authorization token
		tag               string // extra tag for Authorization
		authorizationFunc AuthorizationFunc
		serviceMethodFunc rpc2.ServiceMethodFunc
	}

	// AuthorizationFunc defines a method type which handles Authorization info
	AuthorizationFunc func(serviceMethod, tag, token string) error
)

// NewServerAuthorizationPlugin is by name
func NewServerAuthorizationPlugin(authorizationFunc AuthorizationFunc) *AuthorizationPlugin {
	return &AuthorizationPlugin{
		authorizationFunc: authorizationFunc,
	}
}

// NewClientAuthorizationPlugin is by name
func NewClientAuthorizationPlugin(serviceMethodFunc rpc2.ServiceMethodFunc, tag string, token string) *AuthorizationPlugin {
	return &AuthorizationPlugin{
		token:             token,
		tag:               tag,
		serviceMethodFunc: serviceMethodFunc,
	}
}

// Name returns plugin name.
func (auth *AuthorizationPlugin) Name() string {
	return "AuthorizationPlugin"
}

var _ rpc2.IPreWriteRequestPlugin = new(AuthorizationPlugin)

func (auth *AuthorizationPlugin) PreWriteRequest(r *rpc.Request, body interface{}) error {
	sm := auth.serviceMethodFunc()
	if err := sm.ParseInto(r.ServiceMethod); err != nil {
		return err
	}
	sm.Query().Add("auth", auth.tag+"\x1f"+auth.token)
	r.ServiceMethod = sm.Encode()
	return nil
}

var _ rpc2.IPostReadRequestHeaderPlugin = new(AuthorizationPlugin)

func (auth *AuthorizationPlugin) PostReadRequestHeader(ctx *rpc2.Context) error {
	if auth.authorizationFunc == nil {
		return nil
	}
	s := ctx.Query().Get("auth")
	a := strings.Split(s, "\x1f")
	if len(a) != 2 {
		return errors.New("The authorization is not formatted correctly: " + s)
	}
	return auth.authorizationFunc(ctx.Path(), a[0], a[1])
}
