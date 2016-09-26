package rpc2

import (
	"errors"
	"io"
	"net"
	"net/rpc"
	"time"
)

type serverCodecWrapper struct {
	rpc.ServerCodec
	conn         io.ReadWriteCloser
	groupPlugins []Plugin
	*Server
}

func (w *serverCodecWrapper) ReadRequestHeader(r *rpc.Request) error {
	var (
		conn net.Conn
		ok   bool
	)
	if w.Server.timeout > 0 {
		if conn, ok = w.conn.(net.Conn); ok {
			conn.SetDeadline(time.Now().Add(w.Server.timeout))
		}
	}
	if ok && w.Server.readTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(w.Server.readTimeout))
	}

	// decode
	err := w.ServerCodec.ReadRequestHeader(r)
	if err != nil {
		return err
	}

	for _, plugin := range w.Server.plugins {
		err = plugin.PostReadRequestHeader(r)
		if err != nil {
			return err
		}
	}

	serviceMethod := ParseServiceMethod(r.ServiceMethod)
	for _, groupPrefix := range serviceMethod.Groups() {
		group, ok := w.groupMap[groupPrefix]
		if !ok {
			return errors.New("rpc: can't find group " + groupPrefix)
		}
		for _, plugin := range group.plugins {
			err = plugin.PostReadRequestHeader(r)
			if err != nil {
				return err
			}
			w.groupPlugins = append(w.groupPlugins, plugin)
		}
	}

	r.ServiceMethod = serviceMethod.Path

	return err
}

func (w *serverCodecWrapper) ReadRequestBody(body interface{}) error {
	err := w.ServerCodec.ReadRequestBody(body)
	if err != nil {
		return err
	}
	for _, plugin := range w.Server.plugins {
		err = plugin.PostReadRequestBody(body)
		if err != nil {
			return err
		}
	}
	for _, plugin := range w.groupPlugins {
		err = plugin.PostReadRequestBody(body)
		if err != nil {
			return err
		}
	}
	w.groupPlugins = nil
	return err
}

// WriteResponse must be safe for concurrent use by multiple goroutines.
func (w *serverCodecWrapper) WriteResponse(resp *rpc.Response, body interface{}) error {
	var (
		conn net.Conn
		ok   bool
	)
	if w.Server.timeout > 0 {
		if conn, ok = w.conn.(net.Conn); ok {
			conn.SetDeadline(time.Now().Add(w.Server.timeout))
		}
	}
	if ok && w.Server.writeTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(w.Server.writeTimeout))
	}

	err := w.ServerCodec.WriteResponse(resp, body)

	return err
}
