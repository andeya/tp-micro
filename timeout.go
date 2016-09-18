package rpc2

import (
	"fmt"
	"time"
)

var (
	afterTime = time.Minute
)

func SetTimeout(timeout time.Duration) {
	afterTime = timeout
	if afterTime <= 0 {
		afterTime = 1<<63 - 1
	}
}

func timeoutCoder(f func(interface{}) error, e interface{}, msg string) error {
	echan := make(chan error, 1)
	go func() { echan <- f(e) }()
	select {
	case e := <-echan:
		return e
	case <-time.After(afterTime):
		return fmt.Errorf("Timeout %s", msg)
	}
}
