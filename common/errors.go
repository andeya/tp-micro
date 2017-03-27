package common

import (
	"fmt"
	"runtime"
)

var (
	// ErrShutdown returns an error with message: 'connection is shut down'
	ErrShutdown = NewRPCError("connection is shut down")
	// ErrPluginIsNil returns an error with message: 'Taccess denied'
	ErrAccessDenied = NewRPCError("Access denied!")
	// ErrServiceMethod returns an error with message: 'service/method request ill-formed: '+serviceMethod'
	ErrServiceMethod = NewRPCError("service/method request ill-formed: '%s'")
	// ErrPluginIsNil returns an error with message: 'The plugin cannot be nil'
	ErrPluginIsNil = NewRPCError("The plugin cannot be nil!")
	// ErrPluginAlreadyExists returns an error with message: 'Cannot activate the same plugin again,'+plugin name' is already exists'
	ErrPluginAlreadyExists = NewRPCError("Cannot use the same plugin again, '%s' is already exists")
	// ErrPluginActivate returns an error with message: 'While trying to activate plugin '+plugin name'. Trace: +specific error'
	ErrPluginActivate = NewRPCError("While trying to activate plugin '%s'. Trace: %s")
	// ErrPluginRemoveNoPlugins returns an error with message: 'No plugins are registed yet, you cannot remove a plugin from an empty list!'
	ErrPluginRemoveNoPlugins = NewRPCError("No plugins are registed yet, you cannot remove a plugin from an empty list!")
	// ErrPluginRemoveEmptyName returns an error with message: 'Plugin with an empty name cannot be removed'
	ErrPluginRemoveEmptyName = NewRPCError("Plugin with an empty name cannot be removed")
	// ErrPluginRemoveNotFound returns an error with message: 'Cannot remove a plugin which doesn't exists'
	ErrPluginRemoveNotFound = NewRPCError("Cannot remove a plugin which doesn't exists")
	// ErrInvalidPath  returns an error with message: 'The service name '+name' invalid, need to meet '/^[a-zA-Z0-9_\.\-/]*$/'
	ErrInvalidPath = NewRPCError("The service name '%s' invalid, need to meet '/^[a-zA-Z0-9_\\.\\-/]*$/'")
	// ErrServiceAlreadyExists returns an error with message: 'Cannot activate the same service again, '+service name' is already exists'
	ErrServiceAlreadyExists = NewRPCError("Cannot use the same service again, '%s' is already exists")

	// RegisterPlugin returns an error with message: 'RegisterPlugin(+plugin name): +errMsg'
	ErrRegisterPlugin = NewRPCError("RegisterPlugin(%s): %s")
	// PostConnAccept returns an error with message: 'PostConnAccept(+plugin name): +errMsg'
	ErrPostConnAccept = NewRPCError("PostConnAccept(%s): %s")
	// ErrPreReadRequestHeader returns an error with message: 'PreReadRequestHeader(+plugin name): +errMsg'
	ErrPreReadRequestHeader = NewRPCError("PreReadRequestHeader(%s): %s")
	// ErrPostReadRequestHeader returns an error with message: 'PostReadRequestHeader(+plugin name): +errMsg'
	ErrPostReadRequestHeader = NewRPCError("PostReadRequestHeader(%s): %s")
	// ErrPreReadRequestBody returns an error with message: 'PreReadRequestBody(+plugin name): +errMsg'
	ErrPreReadRequestBody = NewRPCError("PreReadRequestBody(%s): %s")
	// ErrPostReadRequestBody returns an error with message: 'PostReadRequestBody(+plugin name): +errMsg'
	ErrPostReadRequestBody = NewRPCError("PostReadRequestBody(%s): %s")
	// ErrPreWriteResponse returns an error with message: 'PreWriteResponse(+plugin name): +errMsg'
	ErrPreWriteResponse = NewRPCError("PreWriteResponse(%s): %s")
	// ErrPostWriteResponse returns an error with message: 'PostWriteResponse(+plugin name): +errMsg'
	ErrPostWriteResponse = NewRPCError("PostWriteResponse(%s): %s")

	// ErrPostConnected returns an error with message: 'PostConnected(+plugin name): +errMsg'
	ErrPostConnected = NewRPCError("PostConnected(%s): %s")
	// ErrPreReadResponseHeader returns an error with message: 'PreReadResponseHeader(+plugin name): +errMsg'
	ErrPreReadResponseHeader = NewRPCError("PreReadResponseHeader(%s): %s")
	// ErrPostReadResponseHeader returns an error with message: 'PostReadResponseHeader(+plugin name): +errMsg'
	ErrPostReadResponseHeader = NewRPCError("PostReadResponseHeader(%s): %s")
	// ErrPreReadResponseBody returns an error with message: 'PreReadResponseBody(+plugin name): +errMsg'
	ErrPreReadResponseBody = NewRPCError("PreReadResponseBody(%s): %s")
	// ErrPostReadResponseBody returns an error with message: 'PostReadResponseBody(+plugin name): +errMsg'
	ErrPostReadResponseBody = NewRPCError("PostReadResponseBody(%s): %s")
	// ErrPreWriteRequest returns an error with message: 'PreWriteRequest(+plugin name): +errMsg'
	ErrPreWriteRequest = NewRPCError("PreWriteRequest(%s): %s")
	// ErrPostWriteRequest returns an error with message: 'PostWriteRequest(+plugin name): +errMsg'
	ErrPostWriteRequest = NewRPCError("PostWriteRequest(%s): %s")
)

// RPCError holds the error
type RPCError struct {
	message string
}

// NewRPCError creates and returns an Error with a '' prefix.
func NewRPCError(errMsg string) *RPCError {
	return &RPCError{message: errMsg}
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
	e.message += errMsg
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
	errors []error
}

// Error returns the message of the actual error
func (e *MultiError) Error() string {
	s := "[MultiError]:"
	for _, err := range e.errors {
		if err != nil {
			s += "\n  - " + err.Error()
		} else {
			s += "\n  - " + fmt.Sprintf("%v", err)
		}
	}
	return s
	// return fmt.Sprintf("%v", e.errors)
}

// NewMultiError creates and returns an Error with error splice
func NewMultiError(errors []error) *MultiError {
	return &MultiError{errors: errors}
}
