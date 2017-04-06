package common

// RPCError call error
type RPCError struct {
	Type  ErrorType
	Error string
}

// NewRPCError creates rpc error.
func NewRPCError(errorType ErrorType, errMsg string) *RPCError {
	return &RPCError{
		Type:  errorType,
		Error: errMsg,
	}
}

// ErrorType error type
type ErrorType int8

const (
	// ErrorTypeUnknown unknown error type
	ErrorTypeUnknown ErrorType = 0
)

// RPC Client error type codes.
const (
	ErrorTypeClientShutdown ErrorType = -1 * (iota + 1)
	ErrorTypeClientConnect
	ErrorTypeClientPreWriteRequest
	ErrorTypeClientWriteRequest
	ErrorTypeClientPostWriteRequest
	ErrorTypeClientPreReadResponseHeader
	ErrorTypeClientReadResponseHeader
	ErrorTypeClientPostReadResponseHeader
	ErrorTypeClientPreReadResponseBody
	ErrorTypeClientReadResponseBody
	ErrorTypeClientPostReadResponseBody
)

// RPC Server error type codes.
const (
	ErrorTypeServerPreReadRequestHeader ErrorType = iota + 1
	ErrorTypeServerReadRequestHeader
	ErrorTypeServerInvalidServiceMethod
	ErrorTypeServerNotFoundService
	ErrorTypeServerPostReadRequestHeader
	ErrorTypeServerPreReadRequestBody
	ErrorTypeServerReadRequestBody
	ErrorTypeServerPostReadRequestBody
	ErrorTypeServerServicePanic
	ErrorTypeServerService
	ErrorTypeServerPreWriteResponse
)

// ErrShutdown returns an error with message: 'connection is shut down'
var RPCErrShutdown = &RPCError{
	Type:  ErrorTypeClientShutdown,
	Error: "connection is shut down",
}

var RPCErrBroadCast = &RPCError{
	Type:  ErrorTypeUnknown,
	Error: "some invokers return Error",
}

var RPCErrForking = &RPCError{
	Type:  ErrorTypeUnknown,
	Error: "all invokers return Error",
}
