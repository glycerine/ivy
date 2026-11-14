package cec

import (
	"slices"
	"testing"
)

func TestDividedCorrespondencePreservesTraces(t *testing.T) {
	p := corrFixture(0)
	var params Cec_ParCor_t
	Cec_ManCorSetDefaultParams(&params)
	params.FLatchCorr = 1
	params.nWords = 1
	before := make([][]bool, 16)
	for pattern := range before {
		before[pattern] = corrTrace(p, pattern, 4)
	}
	out := Gia_ManScorrDivideTest(p, &params)
	for pattern, want := range before {
		if got := corrTrace(out, pattern, 4); !slices.Equal(got, want) {
			t.Fatalf("pattern %d: %v != %v", pattern, got, want)
		}
	}
}
