package auth

import (
	"errors"
	"net/rpc"
	"strings"

	"github.com/henrylee2cn/rpc2/client"
	"github.com/henrylee2cn/rpc2/plugin"
	"github.com/henrylee2cn/rpc2/server"
)

type (
	// AuthorizationPlugin authorization plugin
	AuthorizationPlugin struct {
		token             string // Authorization token
		tag               string // extra tag for Authorization
		authorizationFunc AuthorizationFunc
		uriFormator       server.URIFormator
	}

	// AuthorizationFunc defines a method type which handles Authorization info
	AuthorizationFunc func(serviceMethod, tag, token string) error
)

// NewServerAuthorizationPlugin means as its name
func NewServerAuthorizationPlugin(authorizationFunc AuthorizationFunc) *AuthorizationPlugin {
	return &AuthorizationPlugin{
		authorizationFunc: authorizationFunc,
	}
}

// NewClientAuthorizationPlugin means as its name
func NewClientAuthorizationPlugin(uriFormator server.URIFormator, tag string, token string) *AuthorizationPlugin {
	return &AuthorizationPlugin{
		token:       token,
		tag:         tag,
		uriFormator: uriFormator,
	}
}

var _ plugin.IPlugin = new(AuthorizationPlugin)

// Name returns plugin name.
func (auth *AuthorizationPlugin) Name() string {
	return "AuthorizationPlugin"
}

var _ client.IPreWriteRequestPlugin = new(AuthorizationPlugin)

func (auth *AuthorizationPlugin) PreWriteRequest(r *rpc.Request, body interface{}) error {
	p, v, err := auth.uriFormator.URIParse(r.ServiceMethod)
	if err != nil {
		return err
	}
	v.Add("auth", auth.tag+"\x1f"+auth.token)
	r.ServiceMethod = auth.uriFormator.URIEncode(v, p)
	return nil
}

var _ server.IPreReadRequestBodyPlugin = new(AuthorizationPlugin)

func (auth *AuthorizationPlugin) PreReadRequestBody(ctx *server.Context, _ interface{}) error {
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
