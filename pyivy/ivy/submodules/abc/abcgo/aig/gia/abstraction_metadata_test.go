package gia

import (
	"reflect"
	"testing"
)

func TestAbstractionPPIsAreImmediateGateAndRegisterBoundary(t *testing.T) {
	p := Gia_ManStart(16)
	pi, ro := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	lower := Gia_ManAppendAnd(p, pi, ro)
	upper := Gia_ManAppendAnd(p, lower, pi^1)
	Gia_ManAppendCo(p, upper)
	Gia_ManAppendCo(p, lower^1)
	Gia_ManSetRegNum(p, 1)
	p.SetGateClass(upper>>1, 1)
	if got := p.AbstractionPPIs(); !reflect.DeepEqual(got, []int32{lower >> 1}) {
		t.Fatal("gate boundary", got)
	}
	p.SetGateClass(ro>>1, 2)
	if got := p.AbstractionPPIs(); !reflect.DeepEqual(got, []int32{lower >> 1}) {
		t.Fatal("register input boundary", got)
	}
	p.SetGateClass(lower>>1, 1)
	if got := p.AbstractionPPIs(); len(got) != 0 {
		t.Fatal("ordinary PI is not PPI", got)
	}
}
