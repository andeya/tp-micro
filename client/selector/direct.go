package selector

import (
	"time"

	"github.com/henrylee2cn/rpc2/client"
)

// DirectSelector is used to a direct rpc server.
// It don't select a node from service cluster but a specific rpc server.
type DirectSelector struct {
	Network        string
	Address        string
	DialTimeout    time.Duration
	newInvokerFunc client.NewInvokerFunc
	invoker        client.Invoker
}

var _ client.Selector = new(DirectSelector)

//SetNewInvokerFunc sets the NewInvokerFunc.
func (s *DirectSelector) SetNewInvokerFunc(newInvokerFunc client.NewInvokerFunc) {
	s.newInvokerFunc = newInvokerFunc
}

//SetSelectMode is meaningless for DirectSelector because there is only one invoker.
func (s *DirectSelector) SetSelectMode(_ client.SelectMode) {}

//Select returns a rpc invoker.
func (s *DirectSelector) Select(options ...interface{}) (client.Invoker, error) {
	if s.invoker != nil {
		return s.invoker, nil
	}
	c, err := s.newInvokerFunc(s.Network, s.Address, s.DialTimeout)
	s.invoker = c
	return c, err
}

//List returns Invokers to all servers
func (s *DirectSelector) List() []client.Invoker {
	if s.invoker == nil {
		return []client.Invoker{}
	}
	return []client.Invoker{s.invoker}
}

//HandleFailed handle failed Invoker
func (s *DirectSelector) HandleFailed(invoker client.Invoker) {
	invoker.Close()
	s.invoker = nil // reset
}
