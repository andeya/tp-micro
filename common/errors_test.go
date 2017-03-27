package common

import (
	"fmt"
	"testing"
)

func TestMultiError(t *testing.T) {
	mult := NewMultiError([]error{
		ErrShutdown,
		ErrAccessDenied,
	})
	fmt.Println(mult)
}
