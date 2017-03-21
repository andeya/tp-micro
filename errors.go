package rpc2

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	// ErrShutdown returns an error with message: 'rpc: connection is shut down'
	ErrShutdown = NewRPCError("connection is shut down")
	// ErrPluginIsNil returns an error with message: 'rpc: Taccess denied'
	ErrAccessDenied = NewRPCError("Access denied!")
	// ErrServiceMethod returns an error with message: 'rpc: service/method request ill-formed: '+serviceMethod'
	ErrServiceMethod = NewRPCError("service/method request ill-formed: '%s'")
	// ErrPluginIsNil returns an error with message: 'rpc: The plugin cannot be nil'
	ErrPluginIsNil = NewRPCError("The plugin cannot be nil!")
	// ErrPluginAlreadyExists returns an error with message: 'rpc: Cannot activate the same plugin again,'+plugin name' is already exists'
	ErrPluginAlreadyExists = NewRPCError("Cannot use the same plugin again, '%s' is already exists")
	// ErrPluginActivate returns an error with message: 'rpc: While trying to activate plugin '+plugin name'. Trace: +specific error'
	ErrPluginActivate = NewRPCError("While trying to activate plugin '%s'. Trace: %s")
	// ErrPluginRemoveNoPlugins returns an error with message: 'rpc: No plugins are registed yet, you cannot remove a plugin from an empty list!'
	ErrPluginRemoveNoPlugins = NewRPCError("No plugins are registed yet, you cannot remove a plugin from an empty list!")
	// ErrPluginRemoveEmptyName returns an error with message: 'rpc: Plugin with an empty name cannot be removed'
	ErrPluginRemoveEmptyName = NewRPCError("Plugin with an empty name cannot be removed")
	// ErrPluginRemoveNotFound returns an error with message: 'rpc: Cannot remove a plugin which doesn't exists'
	ErrPluginRemoveNotFound = NewRPCError("Cannot remove a plugin which doesn't exists")
	// ErrInvalidPath  returns an error with message: 'rpc: The service path contains invalid string '+string', need to meet '/^[a-zA-Z0-9_\\.]+$/'
	ErrInvalidPath = NewRPCError("The service path contains invalid string '%s', need to meet '/^[a-zA-Z0-9_\\./]+$/'")
	// ErrServiceAlreadyExists returns an error with message: 'rpc: Cannot activate the same service again, '+service name' is already exists'
	ErrServiceAlreadyExists = NewRPCError("Cannot use the same service again, '%s' is already exists")
)

// RPCError holds the error
type RPCError struct {
	message string
}

const (
	errPrefix = "[RPC] "
)

// NewRPCError creates and returns an Error with a 'rpc: ' prefix.
func NewRPCError(errMsg ...string) *RPCError {
	e := &RPCError{message: errPrefix}
	for _, msg := range errMsg {
		e.Append(msg)
	}
	return e
}

// Error returns the message of the actual error
func (e *RPCError) Error() string {
	return e.message
}

// Format returns a formatted new error based on the arguments
func (e *RPCError) Format(args ...interface{}) error {
	return fmt.Errorf(e.message, args...)
}

// Append appends a error message
func (e *RPCError) Append(errMsg string) *RPCError {
	e.message += strings.TrimPrefix(errMsg, errPrefix)
	return e
}

// With does the same thing as Format but it receives an error type which if it's nil it returns a nil error
func (e *RPCError) With(err error) error {
	if err == nil {
		return nil
	}
	return e.Format(err.Error())
}

// Return returns the actual error as it is
func (e *RPCError) Return() error {
	return fmt.Errorf(e.message)
}

// Panic output the message and after panics
func (e *RPCError) Panic() {
	if e == nil {
		return
	}
	_, fn, line, _ := runtime.Caller(1)
	errMsg := e.message + "\nCaller was: " + fmt.Sprintf("%s:%d", fn, line)
	panic(errMsg)
}

// Panicf output the formatted message and after panics
func (e *RPCError) Panicf(args ...interface{}) {
	if e == nil {
		return
	}
	_, fn, line, _ := runtime.Caller(1)
	errMsg := e.Format(args...).Error()
	errMsg = errMsg + "\nCaller was: " + fmt.Sprintf("%s:%d", fn, line)
	panic(errMsg)
}

// MultiError holds multiple errors
type MultiError struct {
	Errors []error
}

// Error returns the message of the actual error
func (e *MultiError) Error() string {
	return fmt.Sprintf("%v", e.Errors)
}

// NewMultiError creates and returns an Error with error splice
func NewMultiError(errors []error) *MultiError {
	return &MultiError{Errors: errors}
}
