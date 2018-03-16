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
	"net"
	"sync"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
)

// InnerIpPort returns the service's intranet address, such as '192.168.1.120:8080'.
func InnerIpPort(port string) (string, error) {
	host, err := goutil.IntranetIP()
	if err != nil {
		return "", err
	}
	hostPort := net.JoinHostPort(host, port)
	return hostPort, nil
}

// OuterIpPort returns the service's extranet address, such as '113.116.141.121:8080'.
func OuterIpPort(port string) (string, error) {
	host, err := goutil.ExtranetIP()
	if err != nil {
		return "", err
	}
	hostPort := net.JoinHostPort(host, port)
	return hostPort, nil
}

var initOnce sync.Once

func doInit() {
	initOnce.Do(func() {
		go tp.GraceSignal()
	})
}

// error codes
const (
	RerrCodeClientClosed int32 = 10000
	RerrCodeBind         int32 = 10001
)

var (
	// RerrClosed reply error: client is closed
	RerrClosed = tp.NewRerror(RerrCodeClientClosed, "client is closed", "")
)

// functions that migrated from teleport

// GraceSignal func GraceSignal()
var GraceSignal = tp.GraceSignal

// SetShutdown func SetShutdown(timeout time.Duration, firstSweep, beforeExiting func() error)
var SetShutdown = tp.SetShutdown

// Shutdown func Shutdown(timeout ...time.Duration)
var Shutdown = tp.Shutdown

// Reboot func Reboot(timeout ...time.Duration)
var Reboot = tp.Reboot

// IsConnRerror determines whether the error is a connection error
//  func IsConnRerror(rerr *Rerror) bool
var IsConnRerror = tp.IsConnRerror

// Go func Go(fn func()) bool
var Go = tp.Go

// AnywayGo func AnywayGo(fn func())
var AnywayGo = tp.AnywayGo

// SetGopool func SetGopool(maxGoroutinesAmount int, maxGoroutineIdleDuration time.Duration)
var SetGopool = tp.SetGopool

// SetLogger func SetLogger(logger Logger)
var SetLogger = tp.SetLogger

// SetLoggerLevel func SetLoggerLevel(level string)
var SetLoggerLevel = tp.SetLoggerLevel

// Criticalf func Criticalf(format string, args ...interface{})
var Criticalf = tp.Criticalf

// Debugf func Debugf(format string, args ...interface{})
var Debugf = tp.Debugf

// Errorf func Errorf(format string, args ...interface{})
var Errorf = tp.Errorf

// Fatalf func Fatalf(format string, args ...interface{})
var Fatalf = tp.Fatalf

// Infof func Infof(format string, args ...interface{})
var Infof = tp.Infof

// Noticef func Noticef(format string, args ...interface{})
var Noticef = tp.Noticef

// Panicf func Panicf(format string, args ...interface{})
var Panicf = tp.Panicf

// Printf func Printf(format string, args ...interface{})
var Printf = tp.Printf

// Tracef func Tracef(format string, args ...interface{})
var Tracef = tp.Tracef

// Warnf func Warnf(format string, args ...interface{})
var Warnf = tp.Warnf

// WithRealId sets the real ID to metadata.
// func WithRealId(id string) socket.PacketSetting
var WithRealId = tp.WithRealId

// WithRealIp sets the real IP to metadata.
// func WithRealIp(ip string) socket.PacketSetting
var WithRealIp = tp.WithRealIp

// WithAcceptBodyCodec sets the body codec that the sender wishes to accept.
// Note: If the specified codec is invalid, the receiver will ignore the mate data.
//  func WithAcceptBodyCodec(bodyCodec byte) socket.PacketSetting
var WithAcceptBodyCodec = tp.WithAcceptBodyCodec

// WithContext sets the packet handling context.
//  func WithContext(ctx context.Context) socket.PacketSetting
var WithContext = tp.WithContext

// WithSeq sets the packet sequence.
//  func WithSeq(seq uint64) socket.PacketSetting
var WithSeq = tp.WithSeq

// WithPtype sets the packet type.
//  func WithPtype(ptype byte) socket.PacketSetting
var WithPtype = tp.WithPtype

// WithUri sets the packet URL string.
//  func WithUri(uri string) socket.PacketSetting
var WithUri = tp.WithUri

// WithQuery sets the packet URL query parameter.
//  func WithQuery(key, value string) socket.PacketSetting
var WithQuery = tp.WithQuery

// WithAddMeta adds 'key=value' metadata argument.
// Multiple values for the same key may be added.
//  func WithAddMeta(key, value string) socket.PacketSetting
var WithAddMeta = tp.WithAddMeta

// WithSetMeta sets 'key=value' metadata argument.
//  func WithSetMeta(key, value string) socket.PacketSetting
var WithSetMeta = tp.WithSetMeta

// WithBodyCodec sets the body codec.
//  func WithBodyCodec(bodyCodec byte) socket.PacketSetting
var WithBodyCodec = tp.WithBodyCodec

// WithBody sets the body object.
//  func WithBody(body interface{}) socket.PacketSetting
var WithBody = tp.WithBody

// WithNewBody resets the function of geting body.
//  func WithNewBody(newBodyFunc socket.NewBodyFunc) socket.PacketSetting
var WithNewBody = tp.WithNewBody

// WithXferPipe sets transfer filter pipe.
//  func WithXferPipe(filterId ...byte) socket.PacketSetting
var WithXferPipe = tp.WithXferPipe

// GetPacket gets a *Packet form packet stack.
// Note:
//  newBodyFunc is only for reading form connection;
//  settings are only for writing to connection.
//  func GetPacket(settings ...socket.PacketSetting) *socket.Packet
var GetPacket = tp.GetPacket

// PutPacket puts a *socket.Packet to packet stack.
//  func PutPacket(p *socket.Packet)
var PutPacket = tp.PutPacket

// DefaultProtoFunc gets the default builder of socket communication protocol
//  func DefaultProtoFunc() socket.ProtoFunc
var DefaultProtoFunc = tp.DefaultProtoFunc

// SetDefaultProtoFunc sets the default builder of socket communication protocol
//  func SetDefaultProtoFunc(protoFunc socket.ProtoFunc)
var SetDefaultProtoFunc = tp.SetDefaultProtoFunc

// GetReadLimit gets the packet size upper limit of reading.
//  GetReadLimit() uint32
var GetReadLimit = tp.GetReadLimit

// SetReadLimit sets max packet size.
// If maxSize<=0, set it to max uint32.
//  func SetReadLimit(maxPacketSize uint32)
var SetReadLimit = tp.SetReadLimit

// SetSocketKeepAlive sets whether the operating system should send
// keepalive messages on the connection.
// Note: If have not called the function, the system defaults are used.
//  func SetSocketKeepAlive(keepalive bool)
var SetSocketKeepAlive = tp.SetSocketKeepAlive

// SetSocketKeepAlivePeriod sets period between keep alives.
// Note: if d<0, don't change the value.
//  func SetSocketKeepAlivePeriod(d time.Duration)
var SetSocketKeepAlivePeriod = tp.SetSocketKeepAlivePeriod

// SocketReadBuffer returns the size of the operating system's
// receive buffer associated with the connection.
// Note: if using the system default value, bytes=-1 and isDefault=true.
//  func SocketReadBuffer() (bytes int, isDefault bool)
var SocketReadBuffer = tp.SocketReadBuffer

// SetSocketReadBuffer sets the size of the operating system's
// receive buffer associated with the connection.
// Note: if bytes<0, don't change the value.
//  func SetSocketReadBuffer(bytes int)
var SetSocketReadBuffer = tp.SetSocketReadBuffer

// SocketWriteBuffer returns the size of the operating system's
// transmit buffer associated with the connection.
// Note: if using the system default value, bytes=-1 and isDefault=true.
//  func SocketWriteBuffer() (bytes int, isDefault bool)
var SocketWriteBuffer = tp.SocketWriteBuffer

// SetSocketNoDelay controls whether the operating system should delay
// packet transmission in hopes of sending fewer packets (Nagle's
// algorithm).  The default is true (no delay), meaning that data is
// sent as soon as possible after a Write.
//  func SetSocketNoDelay(noDelay bool)
var SetSocketNoDelay = tp.SetSocketNoDelay
