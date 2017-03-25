package server

import (
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

type (
	// IServiceBuilder is tool about IService.
	IServiceBuilder interface {
		// NewServices creates and returns IService array.
		NewServices(rcvr interface{}, pathSegment ...string) ([]IService, error)
		// URIFormator URI format tool
		URIFormator
	}

	// IService handles request.
	IService interface {
		// SetPluginContainer means as its name
		SetPluginContainer(IServerPluginContainer)
		// GetPluginContainer means as its name
		GetPluginContainer() IServerPluginContainer
		// GetPath returns the name of service
		GetPath() string
		// GetArgType returns the receiver type of request body.
		GetArgType() reflect.Type
		// GetReplyType returns the receiver type of response body.
		GetReplyType() reflect.Type
		// Call calls service method.
		Call(argv, replyv reflect.Value, ctx *Context) error
	}
)

type (
	NormServiceBuilder struct {
		URIFormator
	}
	NormService struct {
		path            string        // name of service
		rcvr            reflect.Value // receiver of methods for the service
		typ             reflect.Type  // type of the receiver
		method          reflect.Method
		ArgType         reflect.Type
		ReplyType       reflect.Type
		numCalls        uint
		sync.Mutex      // protects counters
		pluginContainer IServerPluginContainer
	}
)

func NewNormServiceBuilder(uriFormat URIFormator) *NormServiceBuilder {
	return &NormServiceBuilder{
		URIFormator: uriFormat,
	}
}

// NewServices creates and returns IService array.
func (b *NormServiceBuilder) NewServices(rcvr interface{}, pathSegment ...string) ([]IService, error) {
	rcvrt := reflect.TypeOf(rcvr)
	rcvrv := reflect.ValueOf(rcvr)
	var services []IService
	for k, v := range b.suitableMethods(rcvrt, true) {
		v.typ = rcvrt
		v.rcvr = rcvrv
		v.path = b.URIEncode(nil, append(pathSegment, k)...)
		services = append(services, v)
	}
	return services, nil
}

// SetPluginContainer means as its name
func (n *NormService) SetPluginContainer(p IServerPluginContainer) {
	n.pluginContainer = p
}

// GetPluginContainer means as its name
func (n *NormService) GetPluginContainer() IServerPluginContainer {
	return n.pluginContainer
}

// GetArgType returns the receiver type of request body.
func (n *NormService) GetArgType() reflect.Type {
	return n.ArgType
}

// GetReplyType returns the receiver type of request body.
func (n *NormService) GetReplyType() reflect.Type {
	return n.ReplyType
}

// Call calls service method, and returns response result.
func (n *NormService) Call(argv, replyv reflect.Value, _ *Context) error {
	n.Lock()
	n.numCalls++
	n.Unlock()
	function := n.method.Func
	// Invoke the method, providing a new value for the reply.
	returnValues := function.Call([]reflect.Value{n.rcvr, argv, replyv})
	// The return value for the method is an error.
	errInter := returnValues[0].Interface()
	if errInter != nil {
		return errInter.(error)
	}
	return nil
}

// GetPath returns the name of service
func (n *NormService) GetPath() string {
	return n.path
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

// suitableMethods returns suitable Rpc methods of typ, it will report
// error using log if reportErr is true.
func (*NormServiceBuilder) suitableMethods(typ reflect.Type, reportErr bool) map[string]*NormService {
	methods := make(map[string]*NormService)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		// Method needs three ins: receiver, *args, *reply.
		if mtype.NumIn() != 3 {
			if reportErr {
				// log.Notice("[RPC] method", mname, "has wrong number of ins:", mtype.NumIn())
			}
			continue
		}
		// First arg need not be a pointer.
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			if reportErr {
				// log.Notice("[RPC]", mname, "argument type not exported:", argType)
			}
			continue
		}
		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			if reportErr {
				// log.Notice("[RPC] method", mname, "reply type not a pointer:", replyType)
			}
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			if reportErr {
				// log.Notice("[RPC] method", mname, "reply type not exported:", replyType)
			}
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			if reportErr {
				// log.Notice("[RPC] method", mname, "has wrong number of outs:", mtype.NumOut())
			}
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			if reportErr {
				// log.Notice("[RPC] method", mname, "returns", returnType.String(), "not error")
			}
			continue
		}
		methods[mname] = &NormService{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}
