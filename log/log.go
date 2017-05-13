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

package log

import (
	"log"

	"github.com/henrylee2cn/myrpc/log/logging"
	"github.com/henrylee2cn/myrpc/log/logging/color"
)

// type Level int

// // Log levels.
// const (
// 	CRITICAL Level = iota
// 	ERROR
// 	WARNING
// 	NOTICE
// 	INFO
// 	DEBUG
// )

// Logger interface
type Logger interface {
	// AddCalldepth adjusts the calling path depth to print the correct call location.
	AddCalldepth(int)
	// // SetLevel sets log level.
	// SetLevel(Level)
	// Fatal is equivalent to l.Critica followed by a call to os.Exit(1).
	Fatal(args ...interface{})
	// Fatalf is equivalent to l.Criticalf followed by a call to os.Exit(1).
	Fatalf(format string, args ...interface{})

	// Panic is equivalent to l.Critical followed by a call to panic().
	Panic(args ...interface{})
	// Panicf is equivalent to l.Criticalf followed by a call to panic().
	Panicf(format string, args ...interface{})

	// Critical logs a message using CRITICAL as log level.
	Critical(args ...interface{})
	// Criticalf logs a message using CRITICAL as log level.
	Criticalf(format string, args ...interface{})

	// Error logs a message using ERROR as log level.
	Error(args ...interface{})
	// Errorf logs a message using ERROR as log level.
	Errorf(format string, args ...interface{})

	// Warn logs a message using WARNING as log level.
	Warn(args ...interface{})
	// Warnf logs a message using WARNING as log level.
	Warnf(format string, args ...interface{})

	// Notice logs a message using NOTICE as log level.
	Notice(args ...interface{})
	// Noticef logs a message using NOTICE as log level.
	Noticef(format string, args ...interface{})

	// Info logs a message using INFO as log level.
	Info(args ...interface{})
	// Infof logs a message using INFO as log level.
	Infof(format string, args ...interface{})

	// Debug logs a message using DEBUG as log level.
	Debug(args ...interface{})
	// Debugf logs a message using DEBUG as log level.
	Debugf(format string, args ...interface{})
}

// global logger
var global = newDefaultLogger()

// SetLogger sets global logger.
// Note: Concurrent is not safe!
func SetLogger(logger Logger) {
	if logger == nil {
		return
	}
	logger.AddCalldepth(1)
	global = logger
}

const __loglevel__ = "DEBUG"

func newDefaultLogger() Logger {
	var consoleLogBackend = &logging.LogBackend{
		Logger: log.New(color.NewColorableStdout(), "", 0),
		Color:  true,
	}
	consoleFormat := logging.MustStringFormatter("[%{time:01/02 15:04:05}] %{color}[%{level:.1s}]%{color:reset} %{message} <%{longfile}>")
	consoleBackendLevel := logging.AddModuleLevel(logging.NewBackendFormatter(consoleLogBackend, consoleFormat))
	level, err := logging.LogLevel(__loglevel__)
	if err != nil {
		panic(err)
	}
	consoleBackendLevel.SetLevel(level, "")
	logger := logging.NewLogger("myrpc")
	logger.SetBackend(consoleBackendLevel)
	logger.ExtraCalldepth++
	return logger
}

// Fatal is equivalent to l.Critica followed by a call to os.Exit(1).
func Fatal(args ...interface{}) {
	global.Fatal(args...)
}

// Fatalf is equivalent to l.Criticalf followed by a call to os.Exit(1).
func Fatalf(format string, args ...interface{}) {
	global.Fatalf(format, args...)
}

// Panic is equivalent to l.Critical followed by a call to panic().
func Panic(args ...interface{}) {
	global.Panic(args...)
}

// Panicf is equivalent to l.Criticalf followed by a call to panic().
func Panicf(format string, args ...interface{}) {
	global.Panicf(format, args...)
}

// Critical logs a message using CRITICAL as log level.
func Critical(args ...interface{}) {
	global.Critical(args...)
}

// Criticalf logs a message using CRITICAL as log level.
func Criticalf(format string, args ...interface{}) {
	global.Criticalf(format, args...)
}

// Error logs a message using ERROR as log level.
func Error(args ...interface{}) {
	global.Error(args...)
}

// Errorf logs a message using ERROR as log level.
func Errorf(format string, args ...interface{}) {
	global.Errorf(format, args...)
}

// Warn logs a message using WARNING as log level.
func Warn(args ...interface{}) {
	global.Warn(args...)
}

// Warnf logs a message using WARNING as log level.
func Warnf(format string, args ...interface{}) {
	global.Warnf(format, args...)
}

// Notice logs a message using NOTICE as log level.
func Notice(args ...interface{}) {
	global.Notice(args...)
}

// Noticef logs a message using NOTICE as log level.
func Noticef(format string, args ...interface{}) {
	global.Noticef(format, args...)
}

// Info logs a message using INFO as log level.
func Info(args ...interface{}) {
	global.Info(args...)
}

// Infof logs a message using INFO as log level.
func Infof(format string, args ...interface{}) {
	global.Infof(format, args...)
}

// Debug logs a message using DEBUG as log level.
func Debug(args ...interface{}) {
	global.Debug(args...)
}

// Debugf logs a message using DEBUG as log level.
func Debugf(format string, args ...interface{}) {
	global.Debugf(format, args...)
}
