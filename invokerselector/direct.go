package invokerselector

import (
	"time"

	"github.com/henrylee2cn/rpc2"
)

// DirectInvokerSelector is used to a direct rpc server.
// It don't select a node from service cluster but a specific rpc server.
type DirectInvokerSelector struct {
	Network        string
	Address        string
	DialTimeout    time.Duration
	newInvokerFunc rpc2.NewInvokerFunc
	invoker        rpc2.Invoker
}

var _ rpc2.InvokerSelector = new(DirectInvokerSelector)

//Select returns a rpc invoker.
func (s *DirectInvokerSelector) Select(options ...interface{}) (rpc2.Invoker, error) {
	if s.invoker != nil {
		return s.invoker, nil
	}
	c, err := s.newInvokerFunc(s.Network, s.Address, s.DialTimeout)
	s.invoker = c
	return c, err
}

//SetNewInvokerFunc sets the NewInvokerFunc.
func (s *DirectInvokerSelector) SetNewInvokerFunc(newInvokerFunc rpc2.NewInvokerFunc) {
	s.newInvokerFunc = newInvokerFunc
}

//SetSelectMode is meaningless for DirectInvokerSelector because there is only one invoker.
func (s *DirectInvokerSelector) SetSelectMode(_ rpc2.SelectMode) {}

//List returns Invokers to all servers
func (s *DirectInvokerSelector) List() []rpc2.Invoker {
	if s.invoker == nil {
		return []rpc2.Invoker{}
	}
	return []rpc2.Invoker{s.invoker}
}

//HandleFailed handle failed Invoker
func (s *DirectInvokerSelector) HandleFailed(invoker rpc2.Invoker) {
	invoker.Close()
	s.invoker = nil // reset
}
