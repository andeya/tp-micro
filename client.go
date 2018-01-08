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
	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	heartbeat "github.com/henrylee2cn/tp-ext/plugin-heartbeat"
)

// CliConfig client config
// Note:
//  yaml tag is used for github.com/henrylee2cn/cfgo
//  ini tag is used for github.com/henrylee2cn/ini
type CliConfig struct {
	TlsCertFile         string        `yaml:"tls_cert_file"          ini:"tls_cert_file"          comment:"TLS certificate file path"`
	TlsKeyFile          string        `yaml:"tls_key_file"           ini:"tls_key_file"           comment:"TLS key file path"`
	DefaultReadTimeout  time.Duration `yaml:"default_read_timeout"   ini:"default_read_timeout"   comment:"Default maximum duration for reading; ns,µs,ms,s,m,h"`
	DefaultWriteTimeout time.Duration `yaml:"default_write_timeout"  ini:"default_write_timeout"  comment:"Default maximum duration for writing; ns,µs,ms,s,m,h"`
	DefaultDialTimeout  time.Duration `yaml:"default_dial_timeout"   ini:"default_dial_timeout"   comment:"Default maximum duration for dialing; for client role; ns,µs,ms,s,m,h"`
	RedialTimes         int32         `yaml:"redial_times"           ini:"redial_times"           comment:"The maximum times of attempts to redial, after the connection has been unexpectedly broken; for client role"`
	SlowCometDuration   time.Duration `yaml:"slow_comet_duration"    ini:"slow_comet_duration"    comment:"Slow operation alarm threshold; ns,µs,ms,s ..."`
	DefaultBodyCodec    string        `yaml:"default_body_codec"     ini:"default_body_codec"     comment:"Default body codec type id"`
	PrintBody           bool          `yaml:"print_body"             ini:"print_body"             comment:"Is print body or not"`
	CountTime           bool          `yaml:"count_time"             ini:"count_time"             comment:"Is count cost time or not"`
	Network             string        `yaml:"network"                ini:"network"                comment:"Network; tcp, tcp4, tcp6, unix or unixpacket"`
	Heartbeat           time.Duration `yaml:"heartbeat"              ini:"heartbeat"              comment:"When the heartbeat interval is greater than 0, heartbeat is enabled; ns,µs,ms,s,m,h"`
}

// Reload Bi-directionally synchronizes config between YAML file and memory.
func (c *CliConfig) Reload(bind cfgo.BindFunc) error {
	return bind()
}

func (c *CliConfig) peerConfig() tp.PeerConfig {
	return tp.PeerConfig{
		TlsCertFile:         c.TlsCertFile,
		TlsKeyFile:          c.TlsKeyFile,
		DefaultReadTimeout:  c.DefaultReadTimeout,
		DefaultWriteTimeout: c.DefaultWriteTimeout,
		DefaultDialTimeout:  c.DefaultDialTimeout,
		RedialTimes:         c.RedialTimes,
		SlowCometDuration:   c.SlowCometDuration,
		DefaultBodyCodec:    c.DefaultBodyCodec,
		PrintBody:           c.PrintBody,
		CountTime:           c.CountTime,
		Network:             c.Network,
	}
}

// Client client peer
type Client struct {
	peer         *tp.Peer
	linker       Linker
	cacheSession goutil.Map
	protoFunc    socket.ProtoFunc
}

// NewClient creates a client peer.
func NewClient(cfg CliConfig, plugin ...tp.Plugin) *Client {
	if cfg.Heartbeat > 0 {
		plugin = append(plugin, heartbeat.NewPing(cfg.Heartbeat))
	}
	peer := tp.NewPeer(cfg.peerConfig(), plugin...)
	return &Client{
		peer:         peer,
		protoFunc:    socket.DefaultProtoFunc(),
		cacheSession: goutil.AtomicMap(),
	}
}

func (c *Client) SetProtoFunc(protoFunc socket.ProtoFunc) {
	c.protoFunc = protoFunc
}

// AsyncPull sends a packet and receives reply asynchronously.
// If the args is []byte or *[]byte type, it can automatically fill in the body codec name.
func (c *Client) AsyncPull(uri string, args interface{}, reply interface{}, done chan tp.PullCmd, setting ...socket.PacketSetting) {
	sess, rerr := c.getSession(uri)
	if rerr != nil {
		done <- c.fakePullCmd(uri, args, reply, rerr, setting...)
	} else {
		sess.AsyncPull(uri, args, reply, done, setting...)
	}
}

// Pull sends a packet and receives reply.
// Note:
// If the args is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func (c *Client) Pull(uri string, args interface{}, reply interface{}, setting ...socket.PacketSetting) tp.PullCmd {
	sess, rerr := c.getSession(uri)
	if rerr != nil {
		return c.fakePullCmd(uri, args, reply, rerr, setting...)
	}
	return sess.Pull(uri, args, reply, setting...)
}

func (c *Client) getSession(uri string) (tp.Session, *tp.Rerror) {
	addr, rerr := c.linker.SelectAddr(uri)
	if rerr != nil {
		return nil, rerr
	}
	_sess, ok := c.cacheSession.Load(addr)
	var sess tp.Session
	if ok {
		sess = _sess.(tp.Session)
		if sess.Health() {
			return sess, nil
		}
	}
	sess, rerr = c.peer.Dial(addr, c.protoFunc)
	if rerr != nil {
		return nil, rerr
	}
	c.cacheSession.Store(addr, sess)
	return sess, nil
}

func (c *Client) fakePullCmd(uri string, args, reply interface{}, rerr *tp.Rerror, setting ...socket.PacketSetting) tp.PullCmd {
	output := socket.NewPacket(
		socket.WithPtype(tp.TypePull),
		socket.WithUri(uri),
		socket.WithBody(args),
	)
	for _, fn := range setting {
		fn(output)
	}
	return &fakePullCmd{
		peer:   c.peer,
		reply:  reply,
		rerr:   rerr,
		output: output,
	}
}

type fakePullCmd struct {
	peer   *tp.Peer
	reply  interface{}
	rerr   *tp.Rerror
	output *socket.Packet
}

// Peer returns the peer.
func (c *fakePullCmd) Peer() *tp.Peer {
	return c.peer
}

// Session returns the session.
func (c *fakePullCmd) Session() tp.Session {
	return nil
}

// Ip returns the remote addr.
func (c *fakePullCmd) Ip() string {
	return ""
}

// Public returns temporary public data of context.
func (c *fakePullCmd) Public() goutil.Map {
	return nil
}

// PublicLen returns the length of public data of context.
func (c *fakePullCmd) PublicLen() int {
	return 0
}

// Output returns writed packet.
func (c *fakePullCmd) Output() *socket.Packet {
	return c.output
}

// Result returns the pull result.
func (c *fakePullCmd) Result() (interface{}, *tp.Rerror) {
	return c.reply, c.rerr
}

// *Rerror returns the pull error.
func (c *fakePullCmd) Rerror() *tp.Rerror {
	return c.rerr
}

// CostTime returns the pulled cost time.
// If PeerConfig.CountTime=false, always returns 0.
func (c *fakePullCmd) CostTime() time.Duration {
	return 0
}

// Push sends a packet, but do not receives reply.
// Note:
// If the args is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func (c *Client) Push(uri string, args interface{}, setting ...socket.PacketSetting) *tp.Rerror {
	sess, rerr := c.getSession(uri)
	if rerr != nil {
		return rerr
	}
	return sess.Push(uri, args, setting...)
}
