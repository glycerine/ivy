package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCec4PatternGeneratorHandlesBuffer(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendBuf(g, a^1)
	pat, visited := vec.Vec_IntAlloc(4), vec.Vec_IntAlloc(4)
	if got := Cec4_ManGeneratePatternOne(g, 0, 0, b>>1, 1, pat, visited); got != 1 {
		t.Fatal("inverted buffer can be true")
	}
	if vec.Vec_IntSize(pat) != 1 || vec.Vec_IntEntry(pat, 0) != a {
		t.Fatalf("expected input 0: %v", vec.Vec_IntArray(pat))
	}
}
