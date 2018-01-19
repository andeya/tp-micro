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
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
)

func init() {
	go tp.GraceSignal()
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

var (
	// GraceSignal func GraceSignal()
	GraceSignal = tp.GraceSignal
	// SetShutdown func SetShutdown(timeout time.Duration, firstSweep, beforeExiting func() error)
	SetShutdown = tp.SetShutdown
	// Shutdown func Shutdown(timeout ...time.Duration)
	Shutdown = tp.Shutdown
	// Reboot func Reboot(timeout ...time.Duration)
	Reboot = tp.Reboot

	// IsConnRerror determines whether the error is a connection error
	//  func IsConnRerror(rerr *Rerror) bool
	IsConnRerror = tp.IsConnRerror

	// Go func Go(fn func()) bool
	Go = tp.Go
	// AnywayGo func AnywayGo(fn func())
	AnywayGo = tp.AnywayGo
	// SetGopool func SetGopool(maxGoroutinesAmount int, maxGoroutineIdleDuration time.Duration)
	SetGopool = tp.SetGopool

	// SetLogger func SetLogger(logger Logger)
	SetLogger = tp.SetLogger
	// SetLoggerLevel func SetLoggerLevel(level string)
	SetLoggerLevel = tp.SetLoggerLevel
	// Criticalf func Criticalf(format string, args ...interface{})
	Criticalf = tp.Criticalf
	// Debugf func Debugf(format string, args ...interface{})
	Debugf = tp.Debugf
	// Errorf func Errorf(format string, args ...interface{})
	Errorf = tp.Errorf
	// Fatalf func Fatalf(format string, args ...interface{})
	Fatalf = tp.Fatalf
	// Infof func Infof(format string, args ...interface{})
	Infof = tp.Infof
	// Noticef func Noticef(format string, args ...interface{})
	Noticef = tp.Noticef
	// Panicf func Panicf(format string, args ...interface{})
	Panicf = tp.Panicf
	// Printf func Printf(format string, args ...interface{})
	Printf = tp.Printf
	// Tracef func Tracef(format string, args ...interface{})
	Tracef = tp.Tracef
	// Warnf func Warnf(format string, args ...interface{})
	Warnf = tp.Warnf
)

// functions that migrated from socket

// GetPacket gets a *Packet form packet stack.
// Note:
//  newBodyFunc is only for reading form connection;
//  settings are only for writing to connection.
//  func GetPacket(settings ...socket.PacketSetting) *socket.Packet
var GetPacket = socket.GetPacket

// PutPacket puts a *socket.Packet to packet stack.
//  func PutPacket(p *socket.Packet)
var PutPacket = socket.PutPacket

// DefaultProtoFunc gets the default builder of socket communication protocol
//  func DefaultProtoFunc() socket.ProtoFunc
var DefaultProtoFunc = socket.DefaultProtoFunc

// SetDefaultProtoFunc sets the default builder of socket communication protocol
//  func SetDefaultProtoFunc(protoFunc socket.ProtoFunc)
var SetDefaultProtoFunc = socket.SetDefaultProtoFunc

// GetReadLimit gets the packet size upper limit of reading.
//  PacketSizeLimit() uint32
var GetReadLimit = socket.PacketSizeLimit

// SetReadLimit sets max packet size.
// If maxSize<=0, set it to max uint32.
//  func SetPacketSizeLimit(maxPacketSize uint32)
var SetReadLimit = socket.SetPacketSizeLimit

// SetSocketKeepAlive sets whether the operating system should send
// keepalive messages on the connection.
// Note: If have not called the function, the system defaults are used.
//  func SetSocketKeepAlive(keepalive bool)
var SetSocketKeepAlive = socket.SetKeepAlive

// SetSocketKeepAlivePeriod sets period between keep alives.
// Note: if d<0, don't change the value.
//  func SetSocketKeepAlivePeriod(d time.Duration)
var SetSocketKeepAlivePeriod = socket.SetKeepAlivePeriod

// SocketReadBuffer returns the size of the operating system's
// receive buffer associated with the connection.
// Note: if using the system default value, bytes=-1 and isDefault=true.
//  func ReadBuffer() (bytes int, isDefault bool)
var SocketReadBuffer = socket.ReadBuffer

// SetSocketReadBuffer sets the size of the operating system's
// receive buffer associated with the connection.
// Note: if bytes<0, don't change the value.
//  func SetReadBuffer(bytes int)
var SetSocketReadBuffer = socket.SetReadBuffer

// SocketWriteBuffer returns the size of the operating system's
// transmit buffer associated with the connection.
// Note: if using the system default value, bytes=-1 and isDefault=true.
//  func WriteBuffer() (bytes int, isDefault bool)
var SocketWriteBuffer = socket.WriteBuffer

// SetSocketNoDelay controls whether the operating system should delay
// packet transmission in hopes of sending fewer packets (Nagle's
// algorithm).  The default is true (no delay), meaning that data is
// sent as soon as possible after a Write.
//  func SetNoDelay(noDelay bool)
var SetSocketNoDelay = socket.SetNoDelay
