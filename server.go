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
	"time"

	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	binder "github.com/henrylee2cn/tp-ext/plugin-binder"
)

// Server server peer
type Server struct {
	peer *tp.Peer
}

// NewServer creates a server peer.
func NewServer(cfg SrvConfig, plugin ...tp.Plugin) *Server {
	plugin = append([]tp.Plugin{binder.NewStructArgsBinder(antBindErrCode, antBindErrMessage)}, plugin...)
	peer := tp.NewPeer(cfg.peerConfig(), plugin...)
	return &Server{
		peer: peer,
	}
}

// CliConfig client config
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
