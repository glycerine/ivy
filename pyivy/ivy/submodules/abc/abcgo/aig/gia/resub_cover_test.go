package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestResubUnateCoverAndSorting(t *testing.T) {
	a, b := word(0xaaaaaaaaaaaaaaaa), word(0xcccccccccccccccc)
	on := a | b
	off := ^on
	divs := vec.Vec_PtrAlloc(4)
	for _, v := range []word{off, on, a, b} {
		vec.Vec_PtrPush(divs, []word{v})
	}
	lits, weights, sorter := vec.Vec_IntAlloc(2), vec.Vec_IntAlloc(2), vec.Vec_WecAlloc(0)
	vec.Vec_IntPush(lits, 6)
	vec.Vec_IntPush(lits, 4)
	Gia_ManSortUnatesInt([]word{off}, []word{on}, divs, 1, lits, weights, sorter)
	if vec.Vec_IntEntry(weights, 0) != 32 || vec.Vec_IntEntry(weights, 1) != 32 {
		t.Fatal("incorrect coverage")
	}
	var tested int32
	pair := Gia_ManFindTwoUnateInt([]word{off}, []word{on}, divs, 1, lits, weights, &tested)
	if pair < 0 || pair&1 == 0 || tested != 1 {
		t.Fatalf("pair %d tested %d", pair, tested)
	}
	result := []word{0}
	Gia_ManDeriveDivPair(pair^1, divs, 1, result)
	if ^result[0] != on {
		t.Fatal("incorrect OR cover")
	}
}
