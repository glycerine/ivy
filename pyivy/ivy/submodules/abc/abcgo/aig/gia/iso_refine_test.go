package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestIsoPrepareAndAssignUnique(t *testing.T) {
	graph := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	root := Gia_ManAppendAnd(graph, a, b)
	co := Gia_ManAppendCo(graph, root)
	p := Gia_IsoManStart(graph)
	defer Gia_IsoManStop(p)
	Gia_IsoPrepare(p)
	if want := []int32{0, 1, 1, 2, 3, 1, 4, 1}; !reflect.DeepEqual(vec.Vec_IntArray(p.vClasses), want) {
		t.Fatalf("classes %v want %v", vec.Vec_IntArray(p.vClasses), want)
	}
	for id := int32(1); id < p.nObjs; id++ {
		if got := Gia_IsoGetItem(p, id); got != uint32(id) {
			t.Fatalf("item %d is %d", id, got)
		}
	}
	Gia_IsoSetValue(p, 1, 0x89abcdef)
	Gia_IsoSetItem(p, 1, 0x12345678)
	if Gia_IsoGetValue(p, 1) != 0x89abcdef || Gia_IsoGetItem(p, 1) != 0x12345678 {
		t.Fatal("packed fields changed each other")
	}
	Gia_IsoSetItem(p, 1, 1)
	Gia_IsoAssignUnique(p)
	if p.nSingles != 3 || p.nEntries != 2 || p.pUniques[0] == 0 || p.pUniques[root>>1] == 0 || p.pUniques[co>>1] == 0 {
		t.Fatalf("uniques %v singles %d entries %d", p.pUniques, p.nSingles, p.nEntries)
	}
	if got := vec.Vec_IntArray(p.vClasses); !reflect.DeepEqual(got, []int32{1, 2}) {
		t.Fatalf("remaining classes %v", got)
	}
	Gia_IsoManTransferUnique(p)
	if Gia_ManObj(graph, root>>1).Value != uint32(p.pUniques[root>>1]) {
		t.Fatal("unique value not transferred")
	}
}

func TestIsoSortRefinesSignatures(t *testing.T) {
	g := Gia_ManStart(12)
	for i := 0; i < 4; i++ {
		Gia_ManAppendCi(g)
	}
	Gia_ManAppendCo(g, 2)
	p := Gia_IsoManStart(g)
	defer Gia_IsoManStop(p)
	Gia_IsoPrepare(p)
	Gia_IsoAssignUnique(p)
	for i, value := range []uint32{5, 9, 5, 7} {
		Gia_ManCi(g, int32(i)).Value = value
	}
	if Gia_IsoSort(p) != 1 {
		t.Fatal("expected refinement")
	}
	classes := vec.Vec_IntArray(p.vClasses)
	if len(classes) != 2 || classes[1] != 2 {
		t.Fatalf("remaining classes %v", classes)
	}
	begin := classes[0]
	if Gia_IsoGetValue(p, begin) != 5 || Gia_IsoGetValue(p, begin+1) != 5 {
		t.Fatal("equal signatures not grouped")
	}
	if p.nSingles != 2 || p.nEntries != 2 {
		t.Fatalf("singles %d entries %d", p.nSingles, p.nEntries)
	}
	if Gia_IsoSort(p) != 0 {
		t.Fatal("equal signatures unexpectedly refined")
	}
}

func TestIsoForwardAndBackwardSignatures(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	node := Gia_ManAppendAnd(g, a, b^1)
	output := Gia_ManAppendCo(g, node)
	p := Gia_IsoManStart(g)
	defer Gia_IsoManStop(p)
	Gia_IsoPrepare(p)
	Gia_IsoSimulate(p, 0)
	av := s_256Primes[ISO_MASK-1]
	nodeWant := av + Gia_IsoUpdateValue(0, 0) + av + Gia_IsoUpdateValue(0, 1)
	coWant := nodeWant + Gia_IsoUpdateValue(1, 0)
	if Gia_ManObj(g, node>>1).Value != nodeWant || Gia_ManObj(g, output>>1).Value != coWant {
		t.Fatal("incorrect forward signature")
	}
	Gia_IsoSimulateBack(p, 0)
	if Gia_ManObj(g, node>>1).Value != nodeWant+coWant+Gia_IsoUpdateValue(2, 0) {
		t.Fatal("incorrect backward signature")
	}
}

func TestIsoDeriveOutputClasses(t *testing.T) {
	for _, forward := range []int32{0, 1} {
		g := Gia_ManStart(16)
		a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x, y, z := Gia_ManAppendAnd(g, a, b), Gia_ManAppendAnd(g, c, d), Gia_ManAppendAnd(g, a, b^1)
		Gia_ManAppendCo(g, x)
		Gia_ManAppendCo(g, y)
		Gia_ManAppendCo(g, z)
		result := Gia_IsoDeriveEquivPos(g, forward, 0)
		if result == nil {
			t.Fatal("missing output classes")
		}
		seen := make([]bool, 3)
		for _, item := range vec.Vec_PtrArray(result) {
			group := vec.Vec_IntArray(item.(*Vec_Int_t))
			for _, id := range group {
				if id < 0 || id >= 3 || seen[id] {
					t.Fatalf("invalid output class %v", group)
				}
				seen[id] = true
			}
			if forward != 0 && len(group) > 1 {
				if !reflect.DeepEqual(group, []int32{0, 1}) {
					t.Fatalf("unexpected forward class %v", group)
				}
			}
		}
		for i, ok := range seen {
			if !ok {
				t.Fatalf("output %d missing", i)
			}
		}
	}
}
