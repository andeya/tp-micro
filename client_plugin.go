package rpc2

import (
	"net"
	"net/rpc"
)

// for client
type (
	//IPostConnectedPlugin represents connected plugin.
	IPostConnectedPlugin interface {
		PostConnected(net.Conn) (net.Conn, error)
	}

	//IPreReadResponseHeaderPlugin is by name.
	IPreReadResponseHeaderPlugin interface {
		PreReadResponseHeader(*rpc.Response) error
	}

	//IPostReadResponseHeaderPlugin is by name.
	IPostReadResponseHeaderPlugin interface {
		PostReadResponseHeader(*rpc.Response) error
	}

	//IPreReadResponseBodyPlugin is by name.
	IPreReadResponseBodyPlugin interface {
		PreReadResponseBody(interface{}) error
	}

	//IPostReadResponseBodyPlugin is by name.
	IPostReadResponseBodyPlugin interface {
		PostReadResponseBody(interface{}) error
	}

	//IPreWriteRequestPlugin is by name.
	IPreWriteRequestPlugin interface {
		PreWriteRequest(*rpc.Request, interface{}) error
	}

	//IPostWriteRequestPlugin is by name.
	IPostWriteRequestPlugin interface {
		PostWriteRequest(*rpc.Request, interface{}) error
	}

	//IClientPluginContainer represents a plugin container that defines all methods to manage plugins.
	//And it also defines all extension points.
	IClientPluginContainer interface {
		IPluginContainer

		doPostConnected(net.Conn) (net.Conn, error)

		doPreReadResponseHeader(*rpc.Response) error
		doPostReadResponseHeader(*rpc.Response) error

		doPreReadResponseBody(interface{}) error
		doPostReadResponseBody(interface{}) error

		doPreWriteRequest(*rpc.Request, interface{}) error
		doPostWriteRequest(*rpc.Request, interface{}) error
	}
)

// ClientPluginContainer implements IPluginContainer interface.
type ClientPluginContainer struct {
	PluginContainer
}

// doPreReadResponseHeader invokes doPreReadResponseHeader plugin.
func (p *ClientPluginContainer) doPreReadResponseHeader(r *rpc.Response) error {
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPreReadResponseHeaderPlugin); ok {
			err := plugin.PreReadResponseHeader(r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// doPostReadResponseHeader invokes doPostReadResponseHeader plugin.
func (p *ClientPluginContainer) doPostReadResponseHeader(r *rpc.Response) error {
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPostReadResponseHeaderPlugin); ok {
			err := plugin.PostReadResponseHeader(r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// doPreReadResponseBody invokes doPreReadResponseBody plugin.
func (p *ClientPluginContainer) doPreReadResponseBody(body interface{}) error {
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPreReadResponseBodyPlugin); ok {
			err := plugin.PreReadResponseBody(body)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// doPostReadResponseBody invokes doPostReadResponseBody plugin.
func (p *ClientPluginContainer) doPostReadResponseBody(body interface{}) error {
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPostReadResponseBodyPlugin); ok {
			err := plugin.PostReadResponseBody(body)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// doPreWriteRequest invokes doPreWriteRequest plugin.
func (p *ClientPluginContainer) doPreWriteRequest(r *rpc.Request, body interface{}) error {
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPreWriteRequestPlugin); ok {
			err := plugin.PreWriteRequest(r, body)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// doPostWriteRequest invokes doPostWriteRequest plugin.
func (p *ClientPluginContainer) doPostWriteRequest(r *rpc.Request, body interface{}) error {
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPostWriteRequestPlugin); ok {
			err := plugin.PostWriteRequest(r, body)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// doPostConnected handles connected.
func (p *ClientPluginContainer) doPostConnected(conn net.Conn) (net.Conn, error) {
	var err error
	for i := range p.plugins {
		if plugin, ok := p.plugins[i].(IPostConnectedPlugin); ok {
			conn, err = plugin.PostConnected(conn)
			if err != nil { //interrupt
				conn.Close()
				return conn, err
			}
		}
	}
	return conn, nil
}
