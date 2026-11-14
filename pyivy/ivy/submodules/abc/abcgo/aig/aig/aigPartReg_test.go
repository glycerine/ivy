package aig

import (
	"reflect"
	"testing"
)

func registerPartitionFixture() *Graph {
	p := NewGraph()
	pi, a, b, c := p.NewInput(), p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(p.OR(pi, a))
	p.AddOutput(p.AND(pi, b))
	p.AddOutput(p.OR(a, c))
	p.AddOutput(c)
	p.nRegs = 3
	return p
}
func TestRegisterPartitionSimpleReference(t *testing.T) {
	p := NewGraph()
	for i := 0; i < 10; i++ {
		e := p.NewInput()
		p.AddOutput(e)
	}
	p.nRegs = 10
	want := [][]int32{{0, 1, 2, 3}, {3, 4, 5, 6}, {6, 7, 8, 9}}
	if got := Aig_ManRegPartitionSimple(p, 4, 1); !reflect.DeepEqual(got, want) {
		t.Fatal(got)
	}
	if got := Aig_ManRegPartitionSimple(p, 4, 4); !reflect.DeepEqual(got, [][]int32{{0, 1, 2, 3}, {2, 3, 4, 5}, {4, 5, 6, 7}, {6, 7, 8, 9}}) {
		t.Fatal(got)
	}
}
func TestRegisterPartitionSupportHeuristic(t *testing.T) {
	p := registerPartitionFixture()
	manager := Aig_ManRegManStart(p, 2)
	if !reflect.DeepEqual(manager.supports, [][]int32{{1}, {0, 2}, {2}}) {
		t.Fatal(manager.supports)
	}
	if Aig_ManRegFindSeed(manager) != 1 {
		t.Fatal("largest free support not selected")
	}
	parts := Aig_ManRegPartitionSmart(p, 2)
	seen := map[int32]bool{}
	for _, part := range parts {
		if len(part) > 2 {
			t.Fatal(part)
		}
		for _, reg := range part {
			seen[reg] = true
		}
	}
	if len(seen) != 3 {
		t.Fatal(parts)
	}
}
func TestRegisterPartitionCreatePreservesNextStateCone(t *testing.T) {
	p := registerPartitionFixture()
	var pis, external int32
	var back []int32
	part := Aig_ManRegCreatePart(p, []int32{0}, &pis, &external, &back)
	if pis != 1 || external != 1 || part.nRegs != 1 || len(part.inputs) != 3 || len(part.outputs) != 2 {
		t.Fatal(pis, external, part.inputs, part.outputs)
	}
	if back[0] != 0 {
		t.Fatal(back)
	}
	for mask := 0; mask < 8; mask++ {
		pi, b, a := mask&1 != 0, mask&2 != 0, mask&4 != 0
		got, err := part.Evaluate([]bool{pi, b, a})
		if err != nil || got[0] != a || got[1] != (pi && b) {
			t.Fatal(got, err)
		}
	}
	if err := part.Check(); err != nil {
		t.Fatal(err)
	}
	groups := Aig_ManRegProjectOnehots(p, part, [][]int32{{0, 1}, {1, 2}, {2}}, 0)
	if !reflect.DeepEqual(groups, [][]int32{{2, 1}}) {
		t.Fatal(groups)
	}
	for _, id := range back {
		if id >= int32(len(p.nodes)) {
			t.Fatal("invalid back map")
		}
	}
}
func TestRegisterPartitionInvalidSize(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("zero size accepted")
		}
	}()
	Aig_ManRegPartitionSimple(registerPartitionFixture(), 0, 0)
}

func TestRegisterPartitionSmallDomainIsNotDroppedByOverlap(t *testing.T) {
	p := NewGraph()
	for i := 0; i < 2; i++ {
		e := p.NewInput()
		p.AddOutput(e)
	}
	p.nRegs = 2
	if got := Aig_ManRegPartitionSimple(p, 4, 3); !reflect.DeepEqual(got, [][]int32{{0, 1}}) {
		t.Fatal("overlap dropped entire domain", got)
	}
}
func TestRegisterPartitionInvalidDuplicateMember(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("duplicate register accepted")
		}
	}()
	Aig_ManRegCreatePart(registerPartitionFixture(), []int32{1, 1}, nil, nil, nil)
}

func TestRegisterPartitionTraversalWrapClearsOldMarks(t *testing.T) {
	p := registerPartitionFixture()
	p.travID = (1 << 30) - 1
	for i := range p.nodes {
		p.nodes[i].TravID = 1
	}
	part := Aig_ManRegCreatePart(p, []int32{0}, nil, nil, nil)
	groups := Aig_ManRegProjectOnehots(p, part, [][]int32{{0, 1, 2}}, 0)
	if !reflect.DeepEqual(groups, [][]int32{{2, 1}}) {
		t.Fatal("stale traversal marks projected absent registers", groups)
	}
}

func TestRegisterPartitionTransfersRealEquivalentLatchClasses(t *testing.T) {
	p := NewGraph()
	pi, a, b := p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(p.OR(a, b))
	p.AddOutput(pi)
	p.AddOutput(pi)
	p.nRegs = 2
	var back []int32
	part := Aig_ManRegCreatePart(p, []int32{0, 1}, nil, nil, &back)
	Aig_ManReprStart(part, int32(len(part.nodes)))
	Aig_ObjCreateRepr(part, part.inputs[1], part.inputs[2])
	Aig_ManReprStart(p, int32(len(p.nodes)))
	if count := Aig_TransferMappedClasses(p, part, back); count != 1 {
		t.Fatal("class not transferred", count)
	}
	if Aig_ObjFindRepr(p, b) != a {
		t.Fatal("partition IDs were transferred as original IDs instead of using mapBack")
	}
}
