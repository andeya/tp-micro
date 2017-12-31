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

// NewPeer creates a peer that has the StructArgsBinder plugin.
func NewPeer(cfg *tp.PeerConfig, plugin ...tp.Plugin) *tp.Peer {
	plugin = append([]tp.Plugin{NewStructArgsBinder(antBindErrCode, antBindErrMessage)}, plugin...)
	peer := tp.NewPeer(cfg, plugin...)
	return peer
}

var (
	antBindErrCode    int32  = 10000
	antBindErrMessage string = "Invalid parameter"
)

// SetBindErr custom settings error message after parameter binding or verification failed.
func SetBindErr(bindErrCode int32, bindErrMessage string) {
	antBindErrCode = bindErrCode
	antBindErrMessage = bindErrMessage
}
