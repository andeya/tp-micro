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
	tp "github.com/henrylee2cn/teleport"
)

func init() {
	go tp.GraceSignal()
}

var (
	antBindErrCode    int32  = 10000
	antBindErrMessage string = "invalid parameter"
)

// SetBindErr custom settings error message after parameter binding or verification failed.
func SetBindErr(bindErrCode int32, bindErrMessage string) {
	antBindErrCode = bindErrCode
	antBindErrMessage = bindErrMessage
}

// teleport types' alias
type (
	BaseCtx                   = tp.BaseCtx
	Plugin                    = tp.Plugin
	PluginContainer           = tp.PluginContainer
	PreNewPeerPlugin          = tp.PreNewPeerPlugin
	PostNewPeerPlugin         = tp.PostNewPeerPlugin
	PostAcceptPlugin          = tp.PostAcceptPlugin
	PostDialPlugin            = tp.PostDialPlugin
	PostDisconnectPlugin      = tp.PostDisconnectPlugin
	PostReadPullBodyPlugin    = tp.PostReadPullBodyPlugin
	PostReadPullHeaderPlugin  = tp.PostReadPullHeaderPlugin
	PostReadPushBodyPlugin    = tp.PostReadPushBodyPlugin
	PostReadPushHeaderPlugin  = tp.PostReadPushHeaderPlugin
	PostReadReplyBodyPlugin   = tp.PostReadReplyBodyPlugin
	PostReadReplyHeaderPlugin = tp.PostReadReplyHeaderPlugin
	PostRegPlugin             = tp.PostRegPlugin
	PostSession               = tp.PostSession
	PostWritePullPlugin       = tp.PostWritePullPlugin
	PostWritePushPlugin       = tp.PostWritePushPlugin
	PostWriteReplyPlugin      = tp.PostWriteReplyPlugin
	PreReadHeaderPlugin       = tp.PreReadHeaderPlugin
	PreReadPullBodyPlugin     = tp.PreReadPullBodyPlugin
	PreReadPushBodyPlugin     = tp.PreReadPushBodyPlugin
	PreReadReplyBodyPlugin    = tp.PreReadReplyBodyPlugin
	PreSession                = tp.PreSession
	PreWritePullPlugin        = tp.PreWritePullPlugin
	PreWritePushPlugin        = tp.PreWritePushPlugin
	PreWriteReplyPlugin       = tp.PreWriteReplyPlugin
	PullCmd                   = tp.PullCmd
	PullCtx                   = tp.PullCtx
	PushCtx                   = tp.PushCtx
	ReadCtx                   = tp.ReadCtx
	Rerror                    = tp.Rerror
	Session                   = tp.Session
	UnknownPullCtx            = tp.UnknownPullCtx
	UnknownPushCtx            = tp.UnknownPushCtx
	WriteCtx                  = tp.WriteCtx
)

// teleport functions
var (
	// GraceSignal func GraceSignal()
	GraceSignal = tp.GraceSignal

	// SetShutdown func SetShutdown(timeout time.Duration, firstSweep, beforeExiting func() error)
	SetShutdown = tp.SetShutdown

	// Shutdown func Shutdown(timeout ...time.Duration)
	Shutdown = tp.Shutdown

	// Reboot func Reboot(timeout ...time.Duration)
	Reboot = tp.Reboot

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
