package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestEcoMiter(t *testing.T) {
	makeGraph := func(extra bool) *gia.Gia_Man_t {
		g := gia.Gia_ManStart(16)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		root := gia.Gia_ManAppendAnd(g, a, b)
		gia.Gia_ManAppendCo(g, root)
		if extra {
			gia.Gia_ManAppendCo(g, root)
		}
		return g
	}
	gold, old := makeGraph(false), makeGraph(true)
	fans := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(fans, 1)
	vec.Vec_IntPush(fans, 2)
	q := Bmc_EcoMiter(gold, old, fans)
	if q == nil || gia.Gia_ManCiNum(q) != 3 || gia.Gia_ManCoNum(q) != 4 {
		t.Fatal("incorrect miter interface")
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		out, err := q.Evaluate(in)
		want := in[0] && in[1]
		if err != nil || out[0] != (want != in[2]) || out[1] != (want == in[2]) || out[2] != in[0] || out[3] != in[1] {
			t.Fatalf("%v -> %v (%v)", in, out, err)
		}
	}
}

func TestEcoPatchFromMiter(t *testing.T) {
	makeGraph := func(extra bool) *gia.Gia_Man_t {
		g := gia.Gia_ManStart(16)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		root := gia.Gia_ManAppendAnd(g, a, b)
		gia.Gia_ManAppendCo(g, root)
		if extra {
			gia.Gia_ManAppendCo(g, root)
		}
		return g
	}
	fans := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(fans, 1)
	vec.Vec_IntPush(fans, 2)
	q := Bmc_EcoMiter(makeGraph(false), makeGraph(true), fans)
	if got := Bmc_EcoPatch(q, 2, 1); got != 1 {
		t.Fatalf("patch result %d", got)
	}
}
