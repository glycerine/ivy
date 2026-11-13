package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestICheckFlopOrder(t *testing.T) {
	g := gia.Gia_ManStart(20)
	pi := gia.Gia_ManAppendCi(g)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, b)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManAppendCo(g, c)
	gia.Gia_ManSetRegNum(g, 3)
	order := vec.Vec_IntAlloc(3)
	Bmc_PerformFindFlopOrder(g, order)
	if got := vec.Vec_IntArray(order); !reflect.DeepEqual(got, []int32{0, 1, 2}) {
		t.Fatal(got)
	}
}

func TestICheckSearchStableRegister(t *testing.T) {
	g := gia.Gia_ManStart(8)
	ro := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, ro)
	gia.Gia_ManAppendCo(g, ro)
	gia.Gia_ManSetRegNum(g, 1)
	result := Bmc_PerformISearch(g, 1, 0, 0, 0, 0, 0)
	if result == nil {
		t.Fatal("stable register must be inductive")
	}
}
