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
)

// NotFoundService error: not found service in linker.
var NotFoundService = tp.NewRerror(tp.CodeDialFailed, "Dial Failed", "not found service")

// Linker linker for client.
type Linker interface {
	// Select selects a service address by URI path.
	Select(uriPath string) (addr string, rerr *tp.Rerror)
	// EventDel pushs service node offline notification.
	EventDel() <-chan string
	// Sick marks the address status is unhealthy.
	Sick(addr string)
	// Close closes the linker.
	Close()
}

// static linker

// NewStaticLinker creates a static linker.
func NewStaticLinker(srvAddr string) Linker {
	return &staticLinker{
		srvAddr: srvAddr,
		ch:      make(chan string),
	}
}

type staticLinker struct {
	srvAddr string
	ch      chan string
}

// Select selects a service address by URI path.
func (d *staticLinker) Select(string) (string, *tp.Rerror) {
	return d.srvAddr, nil
}

// EventDel pushs service node offline notification.
func (d *staticLinker) EventDel() <-chan string {
	return d.ch
}

// Sick marks the address status is unhealthy.
func (d *staticLinker) Sick(string) {}

// Close closes the linker.
func (d *staticLinker) Close() {
	close(d.ch)
}
