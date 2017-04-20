package client

import (
	"net/rpc"

	"github.com/henrylee2cn/rpc2/common"
	"github.com/henrylee2cn/rpc2/plugin"
)

// for client
type (
	//IPlugin represents a plugin.
	IPlugin plugin.IPlugin

	//IPostConnectedPlugin represents connected plugin.
	IPostConnectedPlugin interface {
		PostConnected(ClientCodecConn) error
	}

	//IPreWriteRequestPlugin means as its name.
	IPreWriteRequestPlugin interface {
		PreWriteRequest(*rpc.Request, interface{}) error
	}

	//IPostWriteRequestPlugin means as its name.
	IPostWriteRequestPlugin interface {
		PostWriteRequest(*rpc.Request, interface{}) error
	}

	//IPreReadResponseHeaderPlugin means as its name.
	IPreReadResponseHeaderPlugin interface {
		PreReadResponseHeader(*rpc.Response) error
	}

	//IPostReadResponseHeaderPlugin means as its name.
	IPostReadResponseHeaderPlugin interface {
		PostReadResponseHeader(*rpc.Response) error
	}

	//IPreReadResponseBodyPlugin means as its name.
	IPreReadResponseBodyPlugin interface {
		PreReadResponseBody(interface{}) error
	}

	//IPostReadResponseBodyPlugin means as its name.
	IPostReadResponseBodyPlugin interface {
		PostReadResponseBody(interface{}) error
	}

	//IClientPluginContainer represents a plugin container that defines all methods to manage plugins.
	//And it also defines all extension points.
	IClientPluginContainer interface {
		plugin.IPluginContainer

		doPostConnected(ClientCodecConn) error

		doPreWriteRequest(*rpc.Request, interface{}) error
		doPostWriteRequest(*rpc.Request, interface{}) error

		doPreReadResponseHeader(*rpc.Response) error
		doPostReadResponseHeader(*rpc.Response) error

		doPreReadResponseBody(interface{}) error
		doPostReadResponseBody(interface{}) error
	}
)

// ClientPluginContainer implements IPluginContainer interface.
type ClientPluginContainer struct {
	plugin.PluginContainer
}

// doPostConnected handles connected.
func (p *ClientPluginContainer) doPostConnected(codecConn ClientCodecConn) error {
	var err error
	for i := range p.Plugins {
		if plugin, ok := p.Plugins[i].(IPostConnectedPlugin); ok {
			err = plugin.PostConnected(codecConn)
			if err != nil { //interrupt
				codecConn.Close()
				return common.ErrPostConnected.Format(p.Plugins[i].Name(), err.Error())
			}
		}
	}
	return nil
}

// doPreWriteRequest invokes doPreWriteRequest plugin.
func (p *ClientPluginContainer) doPreWriteRequest(r *rpc.Request, body interface{}) error {
	for i := range p.Plugins {
		if plugin, ok := p.Plugins[i].(IPreWriteRequestPlugin); ok {
			err := plugin.PreWriteRequest(r, body)
			if err != nil {
				return common.ErrPreWriteRequest.Format(p.Plugins[i].Name(), err.Error())
			}
		}
	}

	return nil
}

// doPostWriteRequest invokes doPostWriteRequest plugin.
func (p *ClientPluginContainer) doPostWriteRequest(r *rpc.Request, body interface{}) error {
	for i := range p.Plugins {
		if plugin, ok := p.Plugins[i].(IPostWriteRequestPlugin); ok {
			err := plugin.PostWriteRequest(r, body)
			if err != nil {
				return common.ErrPostWriteRequest.Format(p.Plugins[i].Name(), err.Error())
			}
		}
	}

	return nil
}

// doPreReadResponseHeader invokes doPreReadResponseHeader plugin.
func (p *ClientPluginContainer) doPreReadResponseHeader(r *rpc.Response) error {
	for i := range p.Plugins {
		if plugin, ok := p.Plugins[i].(IPreReadResponseHeaderPlugin); ok {
			err := plugin.PreReadResponseHeader(r)
			if err != nil {
				return common.ErrPreReadResponseHeader.Format(p.Plugins[i].Name(), err.Error())
			}
		}
	}
	return nil
}

// doPostReadResponseHeader invokes doPostReadResponseHeader plugin.
func (p *ClientPluginContainer) doPostReadResponseHeader(r *rpc.Response) error {
	for i := range p.Plugins {
		if plugin, ok := p.Plugins[i].(IPostReadResponseHeaderPlugin); ok {
			err := plugin.PostReadResponseHeader(r)
			if err != nil {
				return common.ErrPostReadResponseHeader.Format(p.Plugins[i].Name(), err.Error())
			}
		}
	}
	return nil
}

// doPreReadResponseBody invokes doPreReadResponseBody plugin.
func (p *ClientPluginContainer) doPreReadResponseBody(body interface{}) error {
	for i := range p.Plugins {
		if plugin, ok := p.Plugins[i].(IPreReadResponseBodyPlugin); ok {
			err := plugin.PreReadResponseBody(body)
			if err != nil {
				return common.ErrPreReadResponseBody.Format(p.Plugins[i].Name(), err.Error())
			}
		}
	}
	return nil
}

// doPostReadResponseBody invokes doPostReadResponseBody plugin.
func (p *ClientPluginContainer) doPostReadResponseBody(body interface{}) error {
	for i := range p.Plugins {
		if plugin, ok := p.Plugins[i].(IPostReadResponseBodyPlugin); ok {
			err := plugin.PostReadResponseBody(body)
			if err != nil {
				return common.ErrPostReadResponseBody.Format(p.Plugins[i].Name(), err.Error())
			}
		}
	}

	return nil
}
