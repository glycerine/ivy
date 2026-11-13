package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCareNodeCounting(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	inner := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, inner, c))
	for _, test := range []struct {
		pattern []int32
		want    int
	}{{[]int32{1, 1, 1}, 2}, {[]int32{1, 1, 0}, 1}, {[]int32{0, 1, 1}, 2}} {
		if got := giaCountCareNodes(g, test.pattern); got != test.want {
			t.Fatalf("%v got %d want %d", test.pattern, got, test.want)
		}
	}
}
