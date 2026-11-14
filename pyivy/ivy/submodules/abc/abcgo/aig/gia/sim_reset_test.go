package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSimulationResetImplications(t *testing.T) {
	p := Gia_ManStart(24)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	for i := 0; i < 5; i++ {
		Gia_ManAppendCi(p)
	}
	x := Gia_ManAppendAnd(p, a, b^1)
	for i := 0; i < 5; i++ {
		Gia_ManAppendCo(p, x)
	}
	Gia_ManSetRegNum(p, 5)
	super := vec.Vec_IntAlloc(8)
	Gia_ManSimCollect(p, Gia_ManObj(p, x>>1), super)
	want := []int32{a, b ^ 1, x}
	got := vec.Vec_IntArray(super)
	if len(got) != len(want) {
		t.Fatalf("supergate %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("supergate %v", got)
		}
	}
	resets := Gia_ManSimDeriveResets(p)
	if vec.Vec_IntEntry(resets, int(a>>1)) != 1 || vec.Vec_IntEntry(resets, int(b>>1)) != 0 || vec.Vec_IntEntry(resets, int(x>>1)) != 1 {
		t.Fatalf("unexpected reset implications %v", vec.Vec_IntArray(resets))
	}
	var pars Gia_ParSim_t
	Gia_ManSimSetDefaultParams(&pars)
	if pars.NWords != 8 || pars.NIters != 32 || pars.TimeLimit != 60 || pars.iOutFail != -1 {
		t.Fatalf("wrong defaults: %+v", pars)
	}
}
