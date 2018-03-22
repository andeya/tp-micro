package sdk

import (
	"testing"

	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/samples/template/types"
)

// TestSdk test SDK.
func TestSdk(t *testing.T) {
	Init(
		micro.CliConfig{
			Failover:        3,
			HeartbeatSecond: 4,
		},
		micro.NewStaticLinker(":9090"),
	)
	reply, rerr := PullMathDivide(&types.MathDivideArgs{A: 10, B: 5})
	if rerr != nil {
		t.Logf("rerr: %v", rerr)
	} else {
		t.Logf("10 / 5 = %d", *reply)
	}
}
