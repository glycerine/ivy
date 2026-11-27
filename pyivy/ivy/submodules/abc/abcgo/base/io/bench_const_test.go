package io

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

// TestBenchReadsGndVddGatesAsConstants: Io_ReadBenchNetwork reads "x = gnd"
// and "x = vdd" as constant nodes without fanins (the file ABC writes for a
// LUT with a constant input), not as buffers of extra "gnd"/"vdd" nets.
func TestBenchReadsGndVddGatesAsConstants(t *testing.T) {
	text := "INPUT(a)\nOUTPUT(z)\nz           = LUT 0x6 ( a, new_gnd_ )\nnew_gnd_    = gnd\n"
	n, err := ParseBench("c.bench", text, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(n.ObjectsOfKind(abc.ObjectNode)); got != 2 {
		t.Fatalf("%d nodes, want 2 (the LUT and a constant)", got)
	}
	checkBooleanNetwork(t, n, func(a []bool) bool { return a[0] })
}
