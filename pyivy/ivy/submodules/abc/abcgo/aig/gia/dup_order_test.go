package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDupOrderDfsAndSelected(t *testing.T) {
	p := Gia_ManStart(24)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	and := Gia_ManAppendAnd(p, b, c)
	Gia_ManAppendCo(p, and)
	Gia_ManAppendCo(p, a^1)
	dup := Gia_ManDupOrderDfs(p)
	selected := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(selected, 1)
	sub := Gia_ManDupSelectedOutputs(p, selected)
	group := Gia_ManDupOutputGroup(p, 0, 1)
	for bits := 0; bits < 8; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, e := p.Evaluate(inputs)
		if e != nil {
			t.Fatal(e)
		}
		got, e := dup.Evaluate(inputs)
		if e != nil || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("DFS %v %v", got, e)
		}
		got, e = sub.Evaluate(inputs)
		if e != nil || len(got) != 1 || got[0] != want[1] {
			t.Fatalf("selected %v %v", got, e)
		}
		got, e = group.Evaluate([]bool{inputs[1], inputs[2]})
		if e != nil || got[0] != want[0] {
			t.Fatalf("group %v %v", got, e)
		}
	}
}
func TestDupAbsStateMapping(t *testing.T) {
	p := Gia_ManStart(24)
	pi, r0, r1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, r0, r1))
	Gia_ManAppendCo(p, pi)
	Gia_ManAppendCo(p, r0)
	Gia_ManSetRegNum(p, 2)
	ppi := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(ppi, 0)
	inverse := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(inverse, 0)
	vec.Vec_IntPush(inverse, -1)
	out := Gia_ManDupAbs(p, ppi, inverse)
	if Gia_ManPiNum(out) != 2 || Gia_ManRegNum(out) != 1 || Gia_ManPoNum(out) != 1 {
		t.Fatal("incorrect abstracted interface")
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, e := p.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != want[0] || got[1] != want[2] {
			t.Fatalf("abstracted %v vs %v (%v)", got, want, e)
		}
	}
}
func TestDupRemapLiterals(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	p.pObjs[a>>1].Value = 10
	p.pObjs[b>>1].Value = ^uint32(0)
	lits := vec.Vec_IntAlloc(3)
	vec.Vec_IntPush(lits, a^1)
	vec.Vec_IntPush(lits, b)
	vec.Vec_IntPush(lits, -1)
	Gia_ManDupRemapLiterals(lits, p)
	got := vec.Vec_IntArray(lits)
	if got[0] != 11 || got[1] != -1 || got[2] != -1 {
		t.Fatalf("remapped %v", got)
	}
}

func TestDupOrderRealGatesAndUnusedInputs(t *testing.T) {
	p := Gia_ManStart(20)
	p.EnableMuxes()
	a, b, c, unused := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendXorReal(p, b, c)
	mux := Gia_ManAppendMuxReal(p, a, x, b)
	Gia_ManAppendCo(p, mux)
	_ = unused
	out := Gia_ManDupOrderDfs(p)
	if Gia_ManCiNum(out) != 4 {
		t.Fatal("unused input was dropped")
	}
	for bits := 0; bits < 16; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
		want, e := p.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		got, e := out.Evaluate(input)
		if e != nil || got[0] != want[0] {
			t.Fatalf("input %v got %v want %v err %v", input, got, want, e)
		}
	}
}

func TestDupReverseAndAigerOrder(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, c^1)
	Gia_ManAppendAnd(p, a, c)
	for _, out := range []*Gia_Man_t{Gia_ManDupOrderDfsReverse(p), Gia_ManDupOrderAiger(p)} {
		for bits := 0; bits < 8; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			want, e := p.Evaluate(in)
			if e != nil {
				t.Fatal(e)
			}
			got, e := out.Evaluate(in)
			if e != nil || got[0] != want[0] || got[1] != want[1] {
				t.Fatalf("%v %v %v", got, want, e)
			}
		}
	}
}
