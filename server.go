// Copyright 2017 HenryLee. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ants

import (
	"net"
	"time"

	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	binder "github.com/henrylee2cn/tp-ext/plugin-binder"
	heartbeat "github.com/henrylee2cn/tp-ext/plugin-heartbeat"
)

// SrvConfig server config
// Note:
//  yaml tag is used for github.com/henrylee2cn/cfgo
//  ini tag is used for github.com/henrylee2cn/ini
type SrvConfig struct {
	TlsCertFile         string        `yaml:"tls_cert_file"          ini:"tls_cert_file"          comment:"TLS certificate file path"`
	TlsKeyFile          string        `yaml:"tls_key_file"           ini:"tls_key_file"           comment:"TLS key file path"`
	DefaultReadTimeout  time.Duration `yaml:"default_read_timeout"   ini:"default_read_timeout"   comment:"Default maximum duration for reading; ns,µs,ms,s,m,h"`
	DefaultWriteTimeout time.Duration `yaml:"default_write_timeout"  ini:"default_write_timeout"  comment:"Default maximum duration for writing; ns,µs,ms,s,m,h"`
	SlowCometDuration   time.Duration `yaml:"slow_comet_duration"    ini:"slow_comet_duration"    comment:"Slow operation alarm threshold; ns,µs,ms,s ..."`
	DefaultBodyCodec    string        `yaml:"default_body_codec"     ini:"default_body_codec"     comment:"Default body codec type id"`
	PrintBody           bool          `yaml:"print_body"             ini:"print_body"             comment:"Is print body or not"`
	CountTime           bool          `yaml:"count_time"             ini:"count_time"             comment:"Is count cost time or not"`
	Network             string        `yaml:"network"                ini:"network"                comment:"Network; tcp, tcp4, tcp6, unix or unixpacket"`
	ListenAddress       string        `yaml:"listen_address"         ini:"listen_address"         comment:"Listen address; for server role"`
	Heartbeat           time.Duration `yaml:"heartbeat"              ini:"heartbeat"              comment:"When the heartbeat interval is greater than 0, heartbeat is enabled; ns,µs,ms,s,m,h"`
}

// Reload Bi-directionally synchronizes config between YAML file and memory.
func (s *SrvConfig) Reload(bind cfgo.BindFunc) error {
	return bind()
}

func (s *SrvConfig) peerConfig() tp.PeerConfig {
	return tp.PeerConfig{
		TlsCertFile:         s.TlsCertFile,
		TlsKeyFile:          s.TlsKeyFile,
		DefaultReadTimeout:  s.DefaultReadTimeout,
		DefaultWriteTimeout: s.DefaultWriteTimeout,
		SlowCometDuration:   s.SlowCometDuration,
		DefaultBodyCodec:    s.DefaultBodyCodec,
		PrintBody:           s.PrintBody,
		CountTime:           s.CountTime,
		Network:             s.Network,
		ListenAddress:       s.ListenAddress,
	}
}

// Server server peer
type Server struct {
	peer *tp.Peer
}

// NewServer creates a server peer.
func NewServer(cfg SrvConfig, plugin ...tp.Plugin) *Server {
	plugin = append(
		[]tp.Plugin{binder.NewStructArgsBinder(antBindErrCode, antBindErrMessage)},
		plugin...,
	)
	if cfg.Heartbeat > 0 {
		plugin = append(plugin, heartbeat.NewPong(cfg.Heartbeat))
	}
	peer := tp.NewPeer(cfg.peerConfig(), plugin...)
	return &Server{
		peer: peer,
	}
}

// Close closes server.
func (s *Server) Close() error {
	return s.peer.Close()
}

// CountSession returns the number of sessions.
func (s *Server) CountSession() int {
	return s.peer.CountSession()
}

// GetSession gets the session by id.
func (s *Server) GetSession(sessionId string) (tp.Session, bool) {
	return s.peer.GetSession(sessionId)
}

// Listen turns on the listening service.
func (s *Server) Listen(protoFunc ...socket.ProtoFunc) error {
	return s.peer.Listen(protoFunc...)
}

// RangeSession ranges all sessions. If fn returns false, stop traversing.
func (s *Server) RangeSession(fn func(sess tp.Session) bool) {
	s.peer.RangeSession(fn)
}

// ServeConn serves the connection and returns a session.
func (s *Server) ServeConn(conn net.Conn, protoFunc ...socket.ProtoFunc) (tp.Session, error) {
	return s.peer.ServeConn(conn, protoFunc...)
}
