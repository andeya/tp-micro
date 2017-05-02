// +build linux darwin dragonfly freebsd netbsd openbsd rumprun
//
// Copyright 2016 HenryLee. All Rights Reserved.
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

package server

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/henrylee2cn/rpc2/log"
)

func graceSignal() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	defer func() {
		signal.Stop(ch)
		close(exit)
	}()
	sig := <-ch
	switch sig {
	case syscall.SIGINT, syscall.SIGTERM:
		Shutdown()
		return
	case syscall.SIGUSR2:
		Reboot()
	}
}

// Reboot all the rpc services gracefully.
// Notes: Windows system are not supported!
func Reboot(timeout ...time.Duration) {
	serversLock.Lock()
	defer serversLock.Unlock()
	log.Infof("rebooting servers...")

	ppid := os.Getppid()

	// Starts a new process passing it the active listeners. It
	// doesn't fork, but starts a new process using the same environment and
	// arguments as when it was originally started. This allows for a newly
	// deployed binary to be started.
	_, err := grace.StartProcess()
	if err != nil {
		log.Errorf("%s", err.Error())
		log.Infof("reboot servers failed, so close parent.")
		return
	}

	// Shut down gracefully, but wait no longer than global.shutdownTimeout before halting
	if len(timeout) > 0 {
		SetShutdown(timeout[0], finalizers...)
	}
	graceful := shutdown()

	// Close the parent if we inherited and it wasn't init that started us.
	if ppid != 1 {
		if err := syscall.Kill(ppid, syscall.SIGTERM); err != nil {
			log.Errorf("failed to close parent: %s", err.Error())
			log.Infof("servers reboot failed, so close parent.")
			return
		}
	}

	if graceful {
		log.Infof("servers are rebooted gracefully.")
	} else {
		log.Infof("servers are rebooted, but not gracefully.")
	}
}
