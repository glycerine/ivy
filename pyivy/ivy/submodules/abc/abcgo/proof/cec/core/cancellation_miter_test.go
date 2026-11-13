package core

import (
	"context"
	"testing"
)

func TestCancelledMiterWithConstantOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, lit := range []int32{0, 1} {
		c := newCircuit(0)
		c.outputs = []int32{lit}
		got, err := VerifyMiter(c, Options{Context: ctx})
		if err != nil || got.Status != Undecided || got.Counterexample != nil {
			t.Fatalf("constant output %d: %+v err=%v", lit, got, err)
		}
	}
}
