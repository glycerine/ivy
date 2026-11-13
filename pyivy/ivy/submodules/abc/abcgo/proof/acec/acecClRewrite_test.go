package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRewriteReplacePreservesOutputPhase(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	n := gia.Gia_ManAppendAnd(p, a, b)
	gia.Gia_ManAppendCo(p, n^1)
	roots := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(roots, n)
	dst := Acec_RewriteReplace(p, roots)
	for i := 0; i < 4; i++ {
		in := []bool{i&1 != 0, i&2 != 0}
		want, e1 := p.Evaluate(in)
		got, e2 := dst.Evaluate(in)
		if e1 != nil || e2 != nil || got[0] != want[0] {
			t.Fatalf("input %v original %v rewritten %v", in, want, got)
		}
	}
}
func TestRewriteTopSingleRoot(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a)
	levels := vec.Vec_WecStart(1)
	vec.Vec_WecPush(levels, 0, a)
	result := Acec_RewriteTop(p, &Acec_Box_t{vRootLits: levels})
	if vec.Vec_IntSize(result) != 1 || vec.Vec_IntEntry(result, 0) != a {
		t.Fatalf("result %v", vec.Vec_IntArray(result))
	}
}

// acecCl.c builds a fresh manager and never assigns its register count.
func TestRewriteReplaceSourceLeavesRegisterCountZero(t *testing.T) {
	p := gia.Gia_ManStart(8)
	q := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, q)
	gia.Gia_ManSetRegNum(p, 1)
	roots := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(roots, q)
	dst := Acec_RewriteReplace(p, roots)
	if got := gia.Gia_ManRegNum(dst); got != 0 {
		t.Fatalf("registers=%d, want source result 0", got)
	}
}

// acecCl.c copies every internal node with Gia_ManAppendAnd, including native
// XOR nodes; its subsequent phase correction only flips output polarity.
func TestRewriteReplaceSourceCopiesNativeXorAsAnd(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManAppendXorReal(p, a, b)
	gia.Gia_ManAppendCo(p, x)
	roots := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(roots, x)
	dst := Acec_RewriteReplace(p, roots)
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		out, err := dst.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		if out[0] != (in[0] && in[1]) {
			t.Fatalf("input %v got %v, want source AND", in, out)
		}
	}
}
