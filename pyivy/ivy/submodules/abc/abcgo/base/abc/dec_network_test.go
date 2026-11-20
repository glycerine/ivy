package abc

import (
	"github.com/glycerine/abcgo/bool/dec"
	"testing"
)

func TestDecNetworkCountsAndSharesMarkedAnd(t *testing.T) {
	n := NewNetwork("reuse")
	a, b := n.NewObject(ObjectPI).Signal(), n.NewObject(ObjectPI).Signal()
	gate := n.AndEdges(a, b)
	d := DecNetwork{n}
	Abc_NtkIncrementTravId(n)
	Abc_NodeSetTravIdCurrent(gate.Object)
	if !d.InDeletionCone(gate.Not()) {
		t.Error("marked AND omitted from candidate cost")
	}
	if got := d.And(a, b); got != gate {
		t.Error("decomposition duplicated a structurally identical AND")
	}
	Abc_NtkIncrementTravId(n)
	gate.Object.MarkA = true
	if d.InDeletionCone(gate) {
		t.Error("old traversal mark counted in a new candidate")
	}
}

func TestDecNetworkReplacementUsesAigUpdatePath(t *testing.T) {
	n := NewNetwork("dec replacement")
	aig := Abc_AigAlloc(n)
	n.FunctionManager = aig
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	a.SetName("a")
	b.SetName("b")
	c.SetName("c")
	ab := Abc_AigAnd(aig, a.Signal(), b.Signal())
	root := Abc_AigAnd(aig, ab, c.Signal())
	po := n.NewObject(ObjectPO)
	po.SetName("f")
	if err := n.AddFanin(po, root); err != nil {
		t.Fatal(err)
	}

	DecNetwork{Network: n}.Replace(root, ab, false)
	if n.Object(root.Object.ID) != nil {
		t.Fatal("replacement left the old AIG node alive")
	}
	if po.Fanins()[0].Object != ab.Object {
		t.Fatal("replacement did not redirect the output")
	}
	if Abc_AigCheck(aig) == 0 || Abc_NtkCheck(n) == 0 {
		t.Fatal("replacement did not preserve AIG/network integrity")
	}
}

func TestDecCountUsesManagerHashAndNeedsNoBoxesOnMiss(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 0)
	defer Abc_NtkDelete(n)
	d := DecNetwork{Network: n}
	m := n.FunctionManager.(*Abc_Aig_t)
	a, b := Abc_NtkCreatePi(n).Signal(), Abc_NtkCreatePi(n).Signal()
	root := any(Abc_NtkCreatePi(n).Signal())
	g := dec.Dec_GraphCreate(2)
	g.SetLeafSignal(0, a)
	g.SetLeafSignal(1, b)
	dec.Dec_GraphSetRoot(g, dec.Dec_GraphAddNodeAnd(g, dec.Dec_EdgeCreate(0, 1), dec.Dec_EdgeCreate(1, 1)))
	Abc_NtkIncrementTravId(n)
	count := func() int32 { return dec.Dec_GraphToNetworkCount(d, root, g, 1, 1) }
	if got := count(); got != 1 {
		t.Fatalf("absent AND cost %d, want 1", got)
	}
	if got := testing.AllocsPerRun(100, func() { count() }); got != 0 {
		t.Errorf("lookup miss allocated %.0f objects, want 0", got)
	}
	existing := Abc_AigAnd(m, a.Not(), b.Not())
	if got := count(); got != 0 {
		t.Fatalf("existing AND cost %d, want 0", got)
	}
	Abc_AigAndDelete(m, existing.Object)
	if got := d.LookupAnd(a.Not(), b.Not()); got != nil {
		t.Errorf("decomposition lookup returned deleted hash entry %v", got)
	}
	if got := count(); got != 1 {
		t.Errorf("excluded AND cost %d, want 1", got)
	}
	if got := d.And(a.Not(), b.Not()); got == existing {
		t.Error("graph construction reused excluded AND")
	}
}

type genericCountingNetwork struct{ dec.CountingNetwork }

func TestDecPhasedCandidateLevelsAndLimits(t *testing.T) {
	n := NewNetwork("phased candidate")
	a, b := n.NewObject(ObjectPI).Signal(), n.NewObject(ObjectPI).Signal()
	a.Object.Level, b.Object.Level = 2, 4
	root := n.NewObject(ObjectPI).Signal()
	existing := n.AndEdges(a.Not(), b)
	Abc_NtkIncrementTravId(n)
	d := DecNetwork{Network: n}
	for _, tc := range []struct {
		name                  string
		a, b, root            Edge
		invertA, invertB      int32
		nodeLimit, levelLimit int32
		cost                  int32
		signal                any
		level                 uint32
	}{
		{"existing", a, b, root, 1, 0, 0, 5, 0, existing, 5},
		{"both input phases", a.Not(), b.Not(), root, 0, 1, 0, 5, 0, existing, 5},
		{"missing", a, b, root, 1, 1, 1, 5, 1, nil, 5},
		{"node limit", a, b, root, 1, 1, 0, 5, -1, nil, 0},
		{"level limit", a, b, root, 1, 0, 0, 4, -1, nil, 0},
		{"self reference", a, b, existing, 1, 0, 1, 5, -1, nil, 0},
		{"constant zero", a, a, root, 1, 0, 0, 0, 0, n.One().Not(), 0},
		{"left operand", a, n.One(), root, 1, 0, 0, 2, 0, a.Not(), 2},
		{"right operand", n.One(), b, root, 0, 1, 0, 4, 0, b.Not(), 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, network := range []dec.CountingNetwork{d, genericCountingNetwork{d}} {
				g := dec.Dec_GraphCreate(2)
				g.SetLeafSignal(0, tc.a)
				g.SetLeafSignal(1, tc.b)
				dec.Dec_GraphSetRoot(g, dec.Dec_GraphAddNodeAnd(g, dec.Dec_EdgeCreate(0, tc.invertA), dec.Dec_EdgeCreate(1, tc.invertB)))
				got := dec.Dec_GraphToNetworkCount(network, tc.root, g, tc.nodeLimit, tc.levelLimit)
				level, signal := dec.Dec_GraphNodeLast(g).Level, dec.Dec_GraphNodeFunc(g, 2)
				if got != tc.cost || level != tc.level || signal != tc.signal {
					t.Fatalf("%T got cost %d, level %d and signal %v, want %d, %d and %v", network, got, level, signal, tc.cost, tc.level, tc.signal)
				}
			}
		})
	}
	Abc_NodeSetTravIdCurrent(existing.Object)
	g := dec.Dec_GraphCreate(2)
	g.SetLeafSignal(0, a)
	g.SetLeafSignal(1, b)
	dec.Dec_GraphSetRoot(g, dec.Dec_GraphAddNodeAnd(g, dec.Dec_EdgeCreate(0, 1), dec.Dec_EdgeCreate(1, 0)))
	if got := dec.Dec_GraphToNetworkCount(d, root, g, 1, 5); got != 1 {
		t.Errorf("existing AND in deletion cone cost %d, want 1", got)
	}
}

func TestAndLookupChecksEffectiveFanoutsAfterTrivialCases(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 0)
	defer Abc_NtkDelete(n)
	m := n.FunctionManager.(*Abc_Aig_t)
	a, b := Abc_NtkCreatePi(n).Signal(), Abc_NtkCreatePi(n).Signal()
	gate := Abc_AigAnd(m, a, b)
	for _, leaf := range []Edge{a, b} {
		Abc_ObjAdjustFanoutCount(leaf.Object, -1)
		if got := Abc_AigAndLookup(m, a, b); got.Object != nil {
			t.Errorf("leaf %d has zero effective fanouts but lookup found %v", leaf.Object.ID, got)
		}
		if got := Abc_AigAndLookup(m, leaf, n.One()); got != leaf {
			t.Errorf("zero-fanout operand lost trivial result %v", got)
		}
		Abc_ObjAdjustFanoutCount(leaf.Object, 1)
		if got := Abc_AigAndLookup(m, a, b); got != gate {
			t.Errorf("restored fanout lost gate %v", got)
		}
	}
}
