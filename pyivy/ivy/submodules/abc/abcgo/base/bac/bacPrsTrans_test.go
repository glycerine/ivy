package bac

import (
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

func TestPsrSourceRangeAndConstantSizes(t *testing.T) {
	names := util.Abc_NamStart(100, 16)
	p := &Psr_Ntk_t{pStrs: names}
	for _, tc := range []struct {
		text string
		want int32
	}{{"[7=3]", 5}, {"[2=5]", 4}, {"[4]", 1}, {"[-2=2]", 5}} {
		id := util.Abc_NamStrFindOrAdd(names, tc.text, nil)
		if got := Psr_ManRangeSizeRange(p, id); got != tc.want {
			t.Fatalf("%q => %d, want %d", tc.text, got, tc.want)
		}
	}
	id := util.Abc_NamStrFindOrAdd(names, "12'h000", nil)
	if got := Psr_ManRangeSizeConst(p, id); got != 12 {
		t.Fatalf("constant width %d", got)
	}
	if got := Psr_ManRangeSizeRange(p, 0); got != 1 {
		t.Fatalf("absent range width %d", got)
	}
}
