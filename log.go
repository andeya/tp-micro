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

package rpc2

import (
	"log"

	"github.com/henrylee2cn/rpc2/logging"
	"github.com/henrylee2cn/rpc2/logging/color"
)

// Logger interface
type Logger interface {
	// Fatalf is equivalent to l.Critical followed by a call to os.Exit(1).
	Fatalf(format string, args ...interface{})
	// Panicf is equivalent to l.Critical followed by a call to panic().
	Panicf(format string, args ...interface{})
	// Criticalf logs a message using CRITICAL as log level.
	Criticalf(format string, args ...interface{})
	// Errorf logs a message using ERROR as log level.
	Errorf(format string, args ...interface{})
	// Warningf logs a message using WARNING as log level.
	Warningf(format string, args ...interface{})
	// Warnf is an alias for Warningf.
	Warnf(format string, args ...interface{})
	// Noticef logs a message using NOTICE as log level.
	Noticef(format string, args ...interface{})
	// Infof logs a message using INFO as log level.
	Infof(format string, args ...interface{})
	// Debugf logs a message using DEBUG as log level.
	Debugf(format string, args ...interface{})
}

func newDefaultLogger() Logger {
	var consoleLogBackend = &logging.LogBackend{
		Logger: log.New(color.NewColorableStdout(), "", 0),
		Color:  true,
	}
	consoleFormat := logging.MustStringFormatter("[%{time:01/02 15:04:05}] %{color}[%{level:.1s}]%{color:reset} %{message}")
	consoleBackendLevel := logging.AddModuleLevel(logging.NewBackendFormatter(consoleLogBackend, consoleFormat))
	level, err := logging.LogLevel("DEBUG")
	if err != nil {
		panic(err)
	}
	consoleBackendLevel.SetLevel(level, "")
	logger := logging.NewLogger("rpc2")
	logger.SetBackend(consoleBackendLevel)
	return logger
}
