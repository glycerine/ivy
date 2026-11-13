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
