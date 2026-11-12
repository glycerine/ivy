package gia

import (
	"reflect"
	"testing"
)

func TestWindowExtractInsertPreservesExternalSharedFanout(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	shared := Gia_ManAppendAnd(p, a, b)
	root := Gia_ManAppendAnd(p, shared, c)
	outside := Gia_ManAppendOr(p, shared, d)
	Gia_ManAppendCo(p, root)
	Gia_ManAppendCo(p, outside)
	selected := Vec_IntAlloc(1)
	Vec_IntPush(selected, 0)
	window := Gia_ManExtractWin(p, selected, 0)
	if Gia_ManCiNum(window) != 3 || Gia_ManCoNum(window) != 2 {
		t.Fatal("window interface", Gia_ManCiNum(window), Gia_ManCoNum(window))
	}
	balanced := Gia_ManBalanceInt(window, 1)
	rebuilt := Gia_ManInsertWin(p, selected, balanced)
	for mask := 0; mask < 16; mask++ {
		inputs := make([]bool, 4)
		for i := range inputs {
			inputs[i] = mask&(1<<uint(i)) != 0
		}
		want, e := p.Evaluate(inputs)
		if e != nil {
			t.Fatal(e)
		}
		got, e := rebuilt.Evaluate(inputs)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatal("window changed outside fanout", mask, got, want)
		}
	}
	for _, obj := range p.pObjs {
		if obj.fMark0 != 0 {
			t.Fatal("window leaked marks")
		}
	}
}
func TestWindowExtractIncludesZeroDepthOutputs(t *testing.T) {
	p := Gia_ManStart(8)
	a := Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, 0)
	selected := Vec_IntAlloc(2)
	Vec_IntPush(selected, 0)
	Vec_IntPush(selected, 1)
	window := Gia_ManExtractWin(p, selected, 1)
	if Gia_ManCiNum(window) != 1 || Gia_ManCoNum(window) != 2 {
		t.Fatal("lost selected CI/constant", Gia_ManCiNum(window), Gia_ManCoNum(window))
	}
}
func TestWindowInsertionPreservesRegistersAndConstraints(t *testing.T) {
	p := Gia_ManStart(16)
	pi, ro := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	node := Gia_ManAppendAnd(p, pi, ro)
	Gia_ManAppendCo(p, node)
	Gia_ManAppendCo(p, pi)
	Gia_ManAppendCo(p, ro^1)
	Gia_ManSetRegNum(p, 1)
	p.nConstrs = 1
	selected := Vec_IntAlloc(1)
	Vec_IntPush(selected, 0)
	window := Gia_ManExtractWin(p, selected, 0)
	rebuilt := Gia_ManInsertWin(p, selected, window)
	if rebuilt.nRegs != 1 || rebuilt.nConstrs != 1 {
		t.Fatal("lost sequential property metadata")
	}
}
func TestFindLatestDepthAndPercentage(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	and := Gia_ManAppendAnd(p, a, b)
	root := Gia_ManAppendAnd(p, and, c)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, and)
	Gia_ManAppendCo(p, root)
	if got := Vec_IntArray(Gia_ManFindLatest(p, 2, 0)); !reflect.DeepEqual(got, []int32{2}) {
		t.Fatal(got)
	}
	if got := Vec_IntArray(Gia_ManFindLatest(p, 0, 50)); !reflect.DeepEqual(got, []int32{1, 2}) {
		t.Fatal(got)
	}
}
