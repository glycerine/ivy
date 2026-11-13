package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCec4PatternGeneratorHandlesConstantFanins(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a := gia.Gia_ManAppendCi(g)
	zero := gia.Gia_ManAppendAnd(g, a, 0)
	identity := gia.Gia_ManAppendAnd(g, a, 1)
	pat, visited := vec.Vec_IntAlloc(4), vec.Vec_IntAlloc(4)
	if got := Cec4_ManGeneratePatternOne(g, 0, 0, zero>>1, 1, pat, visited); got != 0 {
		t.Fatal("constant-zero gate cannot be one")
	}
	if got := Cec4_ManGeneratePatternOne(g, 0, 0, identity>>1, 1, pat, visited); got != 1 {
		t.Fatal("identity gate can be one")
	}
	if vec.Vec_IntSize(pat) != 1 || vec.Vec_IntEntry(pat, 0) != a|1 {
		t.Fatal("wrong satisfying pattern", vec.Vec_IntArray(pat))
	}
}
