package client

import (
	"time"
)

// Selector manage Invokers.
type Selector interface {
	// SetSelectMode is just as the name implies
	SetSelectMode(SelectMode)
	// SetNewInvokerFunc is just as the name implies
	SetNewInvokerFunc(NewInvokerFunc)
	//Select returns a new Invoker and it also update current Invoker
	Select(options ...interface{}) (Invoker, error)
	//List returns all Invoker
	List() []Invoker
	//HandleFailed handle failed Invoker
	HandleFailed(Invoker)
}

// NewInvokerFunc the function to create a new Invoker.
type NewInvokerFunc func(network, address string, dialTimeout time.Duration) (Invoker, error)

// SelectMode defines the algorithm of selecting a services from cluster
type SelectMode int

const (
	//RandomSelect is selecting randomly
	RandomSelect SelectMode = iota
	//RoundRobin is selecting by round robin
	RoundRobin
	//WeightedRoundRobin is selecting by weighted round robin
	WeightedRoundRobin
	//WeightedICMP is selecting by weighted Ping time
	WeightedICMP
	//ConsistentHash is selecting by hashing
	ConsistentHash
	//Closest is selecting the closest server
	Closest
)

var selectModeStrs = [...]string{
	"RandomSelect",
	"RoundRobin",
	"WeightedRoundRobin",
	"WeightedICMP",
	"ConsistentHash",
	"Closest",
}

func (s SelectMode) String() string {
	return selectModeStrs[s]
}
