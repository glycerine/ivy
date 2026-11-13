package live

import (
	"bytes"
	"strings"
	"testing"
)

func TestSourceCommandErrorsPreserveNetwork(t *testing.T) {
	for _, args := range [][]string{{"l2s", "-z"}, {"l2s", "-s", "-l"}, {"l2s", "-h"}, {"l2s", "operand"}} {
		n := liveFixture(false)
		frame := &Frame{Network: n, Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		if got := Abc_CommandAbcLivenessToSafety(frame, int32(len(args)), args); got != 1 || frame.Network != n {
			t.Fatalf("args %v: status=%d network changed=%v", args, got, frame.Network != n)
		}
	}
	n := liveFixture(false)
	frame := &Frame{Network: n, In: strings.NewReader("2\n"), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if got := Abc_CommandNChooseK(frame, 1, []string{"nck"}); got != 1 || frame.Network != n {
		t.Fatalf("too-wide combination: status=%d network changed=%v", got, frame.Network != n)
	}
	frame = &Frame{Network: n, Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if got := Abc_CommandCS_kLiveness(frame, 1, []string{"klive"}); got != 1 || frame.Network != n {
		t.Fatalf("missing source target: status=%d network changed=%v", got, frame.Network != n)
	}
}
