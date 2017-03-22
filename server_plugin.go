package rpc2

// for server
type (
	//IRegisterPlugin is register plugin.
	IRegisterPlugin interface {
		Register(nodePath string, rcvr interface{}, metadata ...string) error
	}

	//IPostConnAcceptPlugin is connection accept plugin.
	// if returns error, it means subsequent IPostConnAcceptPlugins should not contiune to handle this conn
	// and this conn has been closed.
	IPostConnAcceptPlugin interface {
		PostConnAccept(ServerCodecConn) error
	}

	//IPreReadRequestHeaderPlugin is by name.
	IPreReadRequestHeaderPlugin interface {
		PreReadRequestHeader(*Context) error
	}

	//IPostReadRequestHeaderPlugin is by name.
	IPostReadRequestHeaderPlugin interface {
		PostReadRequestHeader(*Context) error
	}

	//IPreReadRequestBodyPlugin is by name.
	IPreReadRequestBodyPlugin interface {
		PreReadRequestBody(ctx *Context, body interface{}) error
	}

	//IPostReadRequestBodyPlugin is by name.
	IPostReadRequestBodyPlugin interface {
		PostReadRequestBody(ctx *Context, body interface{}) error
	}

	//IPreWriteResponsePlugin is by name.
	IPreWriteResponsePlugin interface {
		PreWriteResponse(ctx *Context, body interface{}) error
	}

	//IPostWriteResponsePlugin is by name.
	IPostWriteResponsePlugin interface {
		PostWriteResponse(ctx *Context, body interface{}) error
	}

	//IServerPluginContainer is a plugin container that defines all methods to manage plugins.
	//And it also defines all extension points.
	IServerPluginContainer interface {
		IPluginContainer

		doRegister(nodePath string, rcvr interface{}, metadata ...string) error

		doPostConnAccept(ServerCodecConn) error

		doPreReadRequestHeader(*Context) error
		doPostReadRequestHeader(*Context) error

		doPreReadRequestBody(ctx *Context, body interface{}) error
		doPostReadRequestBody(ctx *Context, body interface{}) error

		doPreWriteResponse(ctx *Context, body interface{}) error
		doPostWriteResponse(ctx *Context, body interface{}) error
	}
)

// ServerPluginContainer implements IServerPluginContainer interface.
type ServerPluginContainer struct {
	PluginContainer
}

var _ IServerPluginContainer = new(ServerPluginContainer)

// doRegister invokes doRegister plugin.
func (p *ServerPluginContainer) doRegister(nodePath string, rcvr interface{}, metadata ...string) error {
	var errors []error
	for i := range p.plugins {

		if plugin, ok := p.plugins[i].(IRegisterPlugin); ok {
			err := plugin.Register(nodePath, rcvr, metadata...)
			if err != nil {
				errors = append(errors, ErrRegisterPlugin.Format(p.plugins[i].Name(), err.Error()))
			}
		}
	}

	if len(errors) > 0 {
		return NewMultiError(errors)
	}
	return nil
}

//doPostConnAccept handles accepted conn
func (p *ServerPluginContainer) doPostConnAccept(conn ServerCodecConn) error {
	var err error
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPostConnAcceptPlugin); ok {
			err = plugin.PostConnAccept(conn)
			if err != nil { //interrupt
				conn.Close()
				return ErrPostConnAccept.Format(p.plugins[i].Name(), err.Error())
			}
		}
	}
	return nil
}

// doPreReadRequestHeader invokes doPreReadRequestHeader plugin.
func (p *ServerPluginContainer) doPreReadRequestHeader(ctx *Context) error {
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPreReadRequestHeaderPlugin); ok {
			err := plugin.PreReadRequestHeader(ctx)
			if err != nil {
				return ErrPreReadRequestHeader.Format(p.plugins[i].Name(), err.Error())
			}
		}
	}

	return nil
}

// doPostReadRequestHeader invokes doPostReadRequestHeader plugin.
func (p *ServerPluginContainer) doPostReadRequestHeader(ctx *Context) error {
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPostReadRequestHeaderPlugin); ok {
			err := plugin.PostReadRequestHeader(ctx)
			if err != nil {
				return ErrPostReadRequestHeader.Format(p.plugins[i].Name(), err.Error())
			}
		}
	}

	return nil
}

// doPreReadRequestBody invokes doPreReadRequestBody plugin.
func (p *ServerPluginContainer) doPreReadRequestBody(ctx *Context, body interface{}) error {
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPreReadRequestBodyPlugin); ok {
			err := plugin.PreReadRequestBody(ctx, body)
			if err != nil {
				return ErrPreReadRequestBody.Format(p.plugins[i].Name(), err.Error())
			}
		}
	}

	return nil
}

// doPostReadRequestBody invokes doPostReadRequestBody plugin.
func (p *ServerPluginContainer) doPostReadRequestBody(ctx *Context, body interface{}) error {
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPostReadRequestBodyPlugin); ok {
			err := plugin.PostReadRequestBody(ctx, body)
			if err != nil {
				return ErrPostReadRequestBody.Format(p.plugins[i].Name(), err.Error())
			}
		}
	}

	return nil
}

// doPreWriteResponse invokes doPreWriteResponse plugin.
func (p *ServerPluginContainer) doPreWriteResponse(ctx *Context, body interface{}) error {
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPreWriteResponsePlugin); ok {
			err := plugin.PreWriteResponse(ctx, body)
			if err != nil {
				return ErrPreWriteResponse.Format(p.plugins[i].Name(), err.Error())
			}
		}
	}

	return nil
}

// doPostWriteResponse invokes doPostWriteResponse plugin.
func (p *ServerPluginContainer) doPostWriteResponse(ctx *Context, body interface{}) error {
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPostWriteResponsePlugin); ok {
			err := plugin.PostWriteResponse(ctx, body)
			if err != nil {
				return ErrPostWriteResponse.Format(p.plugins[i].Name(), err.Error())
			}
		}
	}

	return nil
}
