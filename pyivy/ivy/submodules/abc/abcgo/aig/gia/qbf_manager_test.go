package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestQbfManagerVerificationAndCofactor(t *testing.T) {
	g := Gia_ManStart(8)
	p, x := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManHashAlloc(g)
	Gia_ManAppendCo(g, Gia_ManHashOr(g, p, x))
	man := Gia_QbfAlloc(g, 1, 0, 0)
	defer Gia_QbfFree(man)
	values := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(values, 0)
	if Gia_QbfVerify(man, values) != 1 || vec.Vec_IntEntry(values, 0) != 0 {
		t.Fatal("missing counterexample")
	}
	mapping := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(mapping, -1)
	cofactor := Gia_QbfCofactor(g, 1, values, mapping)
	if Gia_QbfAddCofactor(man, cofactor) == 0 || man.synthesis.Solve() != satoko.SATOKO_SAT {
		t.Fatal("cofactor rejected")
	}
	Gia_QbfOnePattern(man, values)
	if vec.Vec_IntEntry(values, 0) != 1 || Gia_QbfVerify(man, values) != 0 {
		t.Fatal("incorrect satisfying parameter")
	}
}
