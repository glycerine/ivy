package pdr

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestFlopPriorityMuxControlAndReferenceCleanup(t *testing.T) {
	g := gia.Gia_ManStart(32)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	control, highFanout, unused := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	mux := gia.Gia_ManAppendMux(g, control, a, b)
	gia.Gia_ManAppendCo(g, mux)
	for i := 0; i < 4; i++ {
		gia.Gia_ManAppendCo(g, highFanout)
	}
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManSetRegNum(g, 3)
	_ = unused
	for _, priority := range []func(*gia.Gia_Man_t, int32) *vec.Vec_Int_t{Pdr_ManDeriveFlopPriorities2, Pdr_ManDeriveFlopPriorities3} {
		if got := pdrFlopCosts(g, false); !reflect.DeepEqual(got, []int32{2, 4, 0}) {
			t.Fatalf("fanout costs = %v", got)
		}
		if got := pdrFlopCosts(g, true); !reflect.DeepEqual(got, []int32{7, 4, 0}) {
			t.Fatalf("boosted costs = %v", got)
		}
		plain := append([]int32(nil), vec.Vec_IntArray(priority(g, 0))...)
		boosted := vec.Vec_IntArray(priority(g, 1))
		if plain[1] <= plain[0] || boosted[0] <= boosted[1] || boosted[2] != 0 {
			t.Fatalf("unexpected ranks without control boost %v, with boost %v", plain, boosted)
		}
		if g.Refs != nil {
			t.Fatal("temporary reference counts were retained")
		}
	}
}

func TestFlopPriorityEmptyAndExistingReferences(t *testing.T) {
	g := gia.Gia_ManStart(8)
	for _, priority := range []func(*gia.Gia_Man_t, int32) *vec.Vec_Int_t{Pdr_ManDeriveFlopPriorities2, Pdr_ManDeriveFlopPriorities3} {
		if got := vec.Vec_IntSize(priority(g, 1)); got != 0 {
			t.Fatalf("empty graph produced %d priorities", got)
		}
	}
	gia.Gia_ManCreateRefs(g)
	old := g.Refs
	func() {
		defer func() {
			if recover() == nil {
				t.Error("existing reference data was overwritten")
			}
		}()
		Pdr_ManDeriveFlopPriorities2(g, 1)
	}()
	if len(g.Refs) != len(old) || &g.Refs[0] != &old[0] {
		t.Fatal("existing reference data changed")
	}
}

func TestFlopPriorityThreePreservesEqualCostOrder(t *testing.T) {
	g := gia.Gia_ManStart(16)
	for i := 0; i < 3; i++ {
		gia.Gia_ManAppendCi(g)
	}
	for i := 0; i < 3; i++ {
		gia.Gia_ManAppendCo(g, 0)
	}
	gia.Gia_ManSetRegNum(g, 3)
	if got := vec.Vec_IntArray(Pdr_ManDeriveFlopPriorities3(g, 0)); !reflect.DeepEqual(got, []int32{0, 1, 2}) {
		t.Fatalf("equal-cost priorities = %v", got)
	}
}

func TestSharedMuxControlReceivesOneBoost(t *testing.T) {
	g := gia.Gia_ManStart(32)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	control, highFanout := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	mux1 := gia.Gia_ManAppendMux(g, control, a, b)
	mux2 := gia.Gia_ManAppendMux(g, control, b, c)
	gia.Gia_ManAppendCo(g, mux1)
	gia.Gia_ManAppendCo(g, mux2)
	for i := 0; i < 5; i++ {
		gia.Gia_ManAppendCo(g, highFanout)
	}
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManSetRegNum(g, 2)
	if got := pdrFlopCosts(g, false); !reflect.DeepEqual(got, []int32{4, 5}) {
		t.Fatalf("plain costs %v", got)
	}
	if got := pdrFlopCosts(g, true); !reflect.DeepEqual(got, []int32{10, 5}) {
		t.Fatalf("control boosted more than once: %v", got)
	}
}
