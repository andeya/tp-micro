// Copyright 2018 HenryLee. All Rights Reserved.
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

package ant

import (
	"strings"
	"sync"
	"time"

	"github.com/henrylee2cn/cfgo"
	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	cliSession "github.com/henrylee2cn/tp-ext/mod-cliSession"
	heartbeat "github.com/henrylee2cn/tp-ext/plugin-heartbeat"
)

// CliConfig client config
// Note:
//  yaml tag is used for github.com/henrylee2cn/cfgo
//  ini tag is used for github.com/henrylee2cn/ini
type CliConfig struct {
	TlsCertFile         string        `yaml:"tls_cert_file"          ini:"tls_cert_file"          comment:"TLS certificate file path"`
	TlsKeyFile          string        `yaml:"tls_key_file"           ini:"tls_key_file"           comment:"TLS key file path"`
	DefaultSessionAge   time.Duration `yaml:"default_session_age"    ini:"default_session_age"    comment:"Default session max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
	DefaultContextAge   time.Duration `yaml:"default_context_age"    ini:"default_context_age"    comment:"Default PULL or PUSH context max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
	DefaultDialTimeout  time.Duration `yaml:"default_dial_timeout"   ini:"default_dial_timeout"   comment:"Default maximum duration for dialing; for client role; ns,µs,ms,s,m,h"`
	RedialTimes         int           `yaml:"redial_times"           ini:"redial_times"           comment:"The maximum times of attempts to redial, after the connection has been unexpectedly broken; for client role"`
	Failover            int           `yaml:"failover"               ini:"failover"               comment:"The maximum times of failover"`
	SlowCometDuration   time.Duration `yaml:"slow_comet_duration"    ini:"slow_comet_duration"    comment:"Slow operation alarm threshold; ns,µs,ms,s ..."`
	DefaultBodyCodec    string        `yaml:"default_body_codec"     ini:"default_body_codec"     comment:"Default body codec type id"`
	PrintBody           bool          `yaml:"print_body"             ini:"print_body"             comment:"Is print body or not"`
	CountTime           bool          `yaml:"count_time"             ini:"count_time"             comment:"Is count cost time or not"`
	Network             string        `yaml:"network"                ini:"network"                comment:"Network; tcp, tcp4, tcp6, unix or unixpacket"`
	HeartbeatSecond     int           `yaml:"heartbeat_second"       ini:"heartbeat_second"       comment:"When the heartbeat interval(second) is greater than 0, heartbeat is enabled; if it's smaller than 3, change to 3 default"`
	SessMaxQuota        int           `yaml:"sess_max_quota"         ini:"sess_max_quota"         comment:"The maximum number of sessions in the connection pool"`
	SessMaxIdleDuration time.Duration `yaml:"sess_max_idle_duration" ini:"sess_max_idle_duration" comment:"The maximum time period for the idle session in the connection pool; ns,µs,ms,s,m,h"`
}

// Reload Bi-directionally synchronizes config between YAML file and memory.
func (c *CliConfig) Reload(bind cfgo.BindFunc) error {
	err := bind()
	if err != nil {
		return err
	}
	return c.Check()
}

func (c *CliConfig) Check() error {
	if c.SessMaxQuota <= 0 {
		c.SessMaxQuota = 100
	}
	if c.SessMaxIdleDuration <= 0 {
		c.SessMaxIdleDuration = time.Minute * 3
	}
	if c.Failover < 0 {
		c.Failover = 0
	}
	if c.HeartbeatSecond <= 0 {
		c.HeartbeatSecond = 0
	} else if c.HeartbeatSecond < 3 {
		c.HeartbeatSecond = 3
	}
	return nil
}

func (c *CliConfig) peerConfig() tp.PeerConfig {
	return tp.PeerConfig{
		DefaultSessionAge:  c.DefaultSessionAge,
		DefaultContextAge:  c.DefaultContextAge,
		DefaultDialTimeout: c.DefaultDialTimeout,
		RedialTimes:        int32(c.RedialTimes),
		SlowCometDuration:  c.SlowCometDuration,
		DefaultBodyCodec:   c.DefaultBodyCodec,
		PrintBody:          c.PrintBody,
		CountTime:          c.CountTime,
		Network:            c.Network,
	}
}

// Client client peer
type Client struct {
	peer                tp.Peer
	linker              Linker
	protoFunc           socket.ProtoFunc
	cliSessPool         goutil.Map
	sessMaxQuota        int
	sessMaxIdleDuration time.Duration
	closeCh             chan struct{}
	closeMu             sync.Mutex
	maxTry              int
}

// NewClient creates a client peer.
func NewClient(cfg CliConfig, linker Linker, plugin ...tp.Plugin) *Client {
	doInit()
	if err := cfg.Check(); err != nil {
		tp.Fatalf("%v", err)
	}
	if cfg.HeartbeatSecond > 0 {
		plugin = append(plugin, heartbeat.NewPing(cfg.HeartbeatSecond))
	}
	peer := tp.NewPeer(cfg.peerConfig(), plugin...)
	if len(cfg.TlsCertFile) > 0 && len(cfg.TlsKeyFile) > 0 {
		err := peer.SetTlsConfigFromFile(cfg.TlsCertFile, cfg.TlsKeyFile)
		if err != nil {
			tp.Fatalf("%v", err)
		}
	}
	cli := &Client{
		peer:                peer,
		protoFunc:           socket.DefaultProtoFunc(),
		linker:              linker,
		cliSessPool:         goutil.AtomicMap(),
		sessMaxQuota:        cfg.SessMaxQuota,
		sessMaxIdleDuration: cfg.SessMaxIdleDuration,
		closeCh:             make(chan struct{}),
		maxTry:              cfg.Failover + 1,
	}
	go cli.watchEventDel()
	return cli
}

// SetProtoFunc sets socket.ProtoFunc.
func (c *Client) SetProtoFunc(protoFunc socket.ProtoFunc) {
	if protoFunc == nil {
		protoFunc = socket.DefaultProtoFunc()
	}
	c.protoFunc = protoFunc
}

// Peer returns the peer
func (c *Client) Peer() tp.Peer {
	return c.peer
}

// SubRoute adds handler group.
func (c *Client) SubRoute(pathPrefix string, plugin ...tp.Plugin) *tp.SubRouter {
	return c.peer.SubRoute(pathPrefix, plugin...)
}

// RoutePush registers PUSH handlers, and returns the paths.
func (c *Client) RoutePush(ctrlStruct interface{}, plugin ...tp.Plugin) []string {
	return c.peer.RoutePush(ctrlStruct, plugin...)
}

// RoutePushFunc registers PUSH handler, and returns the path.
func (c *Client) RoutePushFunc(pushHandleFunc interface{}, plugin ...tp.Plugin) string {
	return c.peer.RoutePushFunc(pushHandleFunc, plugin...)
}

// AsyncPull sends a packet and receives reply asynchronously.
// If the args is []byte or *[]byte type, it can automatically fill in the body codec name.
func (c *Client) AsyncPull(uri string, args interface{}, reply interface{}, done chan tp.PullCmd, setting ...socket.PacketSetting) {
	cliSess, rerr := c.getCliSession(uri)
	if rerr != nil {
		done <- tp.NewFakePullCmd(uri, args, reply, rerr)
		return
	}
	cliSess.AsyncPull(uri, args, reply, done, setting...)
}

// Pull sends a packet and receives reply.
// Note:
// If the args is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func (c *Client) Pull(uri string, args interface{}, reply interface{}, setting ...socket.PacketSetting) tp.PullCmd {
	var (
		cliSess *cliSession.CliSession
		rerr    *tp.Rerror
		r       tp.PullCmd
		uriPath = getUriPath(uri)
	)
	for i := 0; i < c.maxTry; i++ {
		cliSess, rerr = c.getCliSession(uriPath)
		if rerr != nil {
			return tp.NewFakePullCmd(uri, args, reply, rerr)
		}
		r = cliSess.Pull(uri, args, reply, setting...)
		if !tp.IsConnRerror(r.Rerror()) {
			return r
		}
		c.linker.Sick(cliSess.Addr())
		if i > 0 {
			tp.Debugf("the %dth failover is triggered because: %s", i, r.Rerror().String())
		}
	}
	return r
}

// Push sends a packet, but do not receives reply.
// Note:
// If the args is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func (c *Client) Push(uri string, args interface{}, setting ...socket.PacketSetting) *tp.Rerror {
	var (
		cliSess *cliSession.CliSession
		rerr    *tp.Rerror
		uriPath = getUriPath(uri)
	)
	for i := 0; i < c.maxTry; i++ {
		cliSess, rerr = c.getCliSession(uriPath)
		if rerr != nil {
			return rerr
		}
		rerr = cliSess.Push(uri, args, setting...)
		if !tp.IsConnRerror(rerr) {
			return rerr
		}
		c.linker.Sick(cliSess.Addr())
		if i > 0 {
			tp.Debugf("the %dth failover is triggered because: %s", i, rerr.String())
		}
	}
	return rerr
}

// Close closes client.
func (c *Client) Close() {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()
	select {
	case <-c.closeCh:
		return
	default:
		close(c.closeCh)
		c.peer.Close()
		c.linker.Close()
	}
}

func getUriPath(uri string) string {
	if idx := strings.Index(uri, "?"); idx != -1 {
		return uri[:idx]
	}
	return uri
}

func (c *Client) getCliSession(uriPath string) (*cliSession.CliSession, *tp.Rerror) {
	select {
	case <-c.closeCh:
		return nil, RerrClosed
	default:
	}
	addr, rerr := c.linker.Select(uriPath)
	if rerr != nil {
		return nil, rerr
	}
	_cliSess, ok := c.cliSessPool.Load(addr)
	if ok {
		return _cliSess.(*cliSession.CliSession), nil
	}
	cliSess := cliSession.New(
		c.peer,
		addr,
		c.sessMaxQuota,
		c.sessMaxIdleDuration,
		c.protoFunc,
	)
	c.cliSessPool.Store(addr, cliSess)
	return cliSess, nil
}

func (c *Client) watchEventDel() {
	ch := c.linker.EventDel()
	for addr := range <-ch {
		_cliSess, ok := c.cliSessPool.Load(addr)
		if !ok {
			continue
		}
		c.cliSessPool.Delete(addr)
		tp.Go(_cliSess.(*cliSession.CliSession).Close)
	}
}
