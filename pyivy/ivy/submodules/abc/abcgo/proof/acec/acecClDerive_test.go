package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMirrorDerivation(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	c := gia.Gia_ManAppendCi(p)
	ab := gia.Gia_ManAppendAnd(p, a, b)
	gia.Gia_ManAppendCo(p, ab)
	mirrors := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	vec.Vec_IntWriteEntry(mirrors, int(b>>1), c^1)
	dst := Acec_ManDerive(p, mirrors)
	for _, in := range [][]bool{{false, false, false}, {false, true, true}, {true, false, false}, {true, true, true}} {
		out, err := dst.Evaluate(in)
		if err != nil || out[0] != (in[0] && !in[2]) {
			t.Fatalf("mirror outputs %v error %v", out, err)
		}
	}
}
func TestDetectLiteralPolarity(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	ab := gia.Gia_ManAppendAnd(p, a, b^1)
	if got := Acec_DetectLitPolarity(p, ab>>1, b>>1); got != b^1 {
		t.Fatalf("b polarity %d", got)
	}
	if got := Acec_DetectLitPolarity(p, ab>>1, (ab>>1)+1); got != -1 {
		t.Fatalf("absent leaf %d", got)
	}
}

func TestLiteralPolarityStopsAtUnrelatedPrimaryInput(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	root := gia.Gia_ManAppendAnd(p, a^1, b)
	if got := Acec_DetectLitPolarity(p, root>>1, a>>1); got != a^1 {
		t.Fatalf("a polarity = %d, want %d", got, a^1)
	}
}
