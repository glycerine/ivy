package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestGhostPreservesSourceNodeKinds(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	cases := []struct {
		name          string
		kind          Aig_Type_t
		want          NodeKind
		first, second Edge
	}{
		{"CI", AIG_OBJ_CI, Input, InvalidEdge, InvalidEdge},
		{"constant CO", AIG_OBJ_CO, Output, g.Const1(), InvalidEdge},
		{"inverted buffer", AIG_OBJ_BUF, Buffer, a.Not(), InvalidEdge},
		{"AND", AIG_OBJ_AND, And, b, a.Not()},
		{"XOR", AIG_OBJ_EXOR, ExclusiveOr, b.Not(), a},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ghost := Aig_ObjCreateGhost(g, tc.first, tc.second, tc.kind)
			if ghost.Kind != tc.want {
				t.Errorf("ghost kind %v, want %v for C type %d", ghost.Kind, tc.want, tc.kind)
			}
			left, right := tc.first, tc.second
			if right != InvalidEdge && left > right {
				left, right = right, left
			}
			if ghost.Fanins != [2]Edge{left, right} {
				t.Errorf("ghost fanins %v, want %v", ghost.Fanins, [2]Edge{left, right})
			}
			if g.NodeCount() != 3 || g.Node(a).Refs != 0 || g.Node(b).Refs != 0 {
				t.Fatal("ghost construction changed the graph")
			}
		})
	}
}

func TestRawObjectReservationStartsEmpty(t *testing.T) {
	g := NewGraph()
	e := Aig_ManFetchMemory(g)
	if !Aig_ObjIsNone(g, e) {
		t.Error("zeroed raw reservation was exposed as a live object")
	}
	if got := Aig_ObjType(g, e); got != AIG_OBJ_NONE {
		t.Errorf("reserved object type %d, want NONE", got)
	}
	if g.NodeCount() != 2 || len(g.nodes) != 2 {
		t.Errorf("raw reservation changed source object count: live %d slots %d", g.NodeCount(), len(g.nodes))
	}
}

func TestRecycledObjectReportsSourceNoneType(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	e := g.AND(a, b)
	Aig_ObjDisconnect(g, e)
	Aig_ManRecycleMemory(g, e)
	if got := Aig_ObjType(g, e); got != AIG_OBJ_NONE {
		t.Fatalf("recycled object type %d, want NONE", got)
	}
	if !Aig_ObjIsNone(g, e) {
		t.Fatal("recycled object is still live")
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalChildrenUseMissingEdgeSentinel(t *testing.T) {
	g := NewGraph()
	input := g.NewInput()
	for _, e := range []Edge{g.Const1(), input} {
		if Aig_ObjChild0(g, e) != InvalidEdge || Aig_ObjChild1(g, e) != InvalidEdge {
			t.Errorf("terminal %d has nonmissing children", e.NodeID())
		}
	}
}

func TestClearedObjectReportsSourceNoneType(t *testing.T) {
	g := NewGraph()
	e := Aig_ManFetchMemory(g)
	// Populate the reserved record as a low-level constructor does.
	g.nodes[e.NodeID()] = Node{Kind: Buffer, Fanins: [2]Edge{g.Const1(), InvalidEdge}}
	Aig_ObjClean(g, e)
	if Aig_ObjType(g, e) != AIG_OBJ_NONE || !Aig_ObjIsNone(g, e) {
		t.Fatal("zeroed record became a constant instead of NONE")
	}
	if Aig_ObjChild0(g, e) != InvalidEdge || Aig_ObjChild1(g, e) != InvalidEdge {
		t.Fatal("zeroed record has nonmissing children")
	}
}

func TestHeaderGhostAndRawRecordsMatchOriginalC(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	kinds := []Aig_Type_t{AIG_OBJ_CI, AIG_OBJ_CO, AIG_OBJ_BUF, AIG_OBJ_AND, AIG_OBJ_EXOR}
	left := []Edge{InvalidEdge, g.Const1(), a.Not(), b, b.Not()}
	right := []Edge{InvalidEdge, InvalidEdge, InvalidEdge, a.Not(), a}
	types := []Aig_Type_t{AIG_OBJ_CONST1, AIG_OBJ_CI, AIG_OBJ_AND, AIG_OBJ_BUF, AIG_OBJ_EXOR, AIG_OBJ_CO, AIG_OBJ_NONE}
	literal := func(e Edge) int64 {
		if e == InvalidEdge {
			return -1
		}
		return int64(e)
	}
	bit := func(v bool) int {
		if v {
			return 1
		}
		return 0
	}
	var want strings.Builder
	fmt.Fprintf(&want, "initial ghost %d children %d %d\n", types[g.ghost.Kind], literal(g.ghost.Fanins[0]), literal(g.ghost.Fanins[1]))
	for i, kind := range kinds {
		ghost := Aig_ObjCreateGhost(g, left[i], right[i], kind)
		fmt.Fprintf(&want, "ghost %d kind %d children %d %d\n", i, types[ghost.Kind], literal(ghost.Fanins[0]), literal(ghost.Fanins[1]))
	}
	fmt.Fprintf(&want, "terminals: %d %d %d %d\n", bit(Aig_ObjChild0(g, g.Const1()) == InvalidEdge), bit(Aig_ObjChild1(g, g.Const1()) == InvalidEdge), bit(Aig_ObjChild0(g, a) == InvalidEdge), bit(Aig_ObjChild1(g, a) == InvalidEdge))
	raw := Aig_ManFetchMemory(g)
	fmt.Fprintf(&want, "reserved type %d none %d count %d slots %d children %d %d\n", Aig_ObjType(g, raw), bit(Aig_ObjIsNone(g, raw)), g.NodeCount(), len(g.nodes), literal(Aig_ObjChild0(g, raw)), literal(Aig_ObjChild1(g, raw)))
	g.nodes[raw.NodeID()] = Node{Kind: Buffer, Fanins: [2]Edge{g.Const1(), InvalidEdge}}
	Aig_ObjClean(g, raw)
	fmt.Fprintf(&want, "cleared type %d none %d count %d children %d %d\n", Aig_ObjType(g, raw), bit(Aig_ObjIsNone(g, raw)), g.NodeCount(), literal(Aig_ObjChild0(g, raw)), literal(Aig_ObjChild1(g, raw)))
	Aig_ManRecycleMemory(g, raw)
	fmt.Fprintf(&want, "recycled type %d none %d count %d slots %d\n", Aig_ObjType(g, raw), bit(Aig_ObjIsNone(g, raw)), g.NodeCount(), len(g.nodes))
	got := runAigSourceFixture(t, "header_records_compare.c", "aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	if got != want.String() {
		t.Fatalf("C header states:\n%sGo header states:\n%s", got, want.String())
	}
}

func TestMissingChildrenAreNotComplemented(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	output := g.AddOutput(a.Not())
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{a, InvalidEdge}})
	if Aig_IsComplement(InvalidEdge) != 0 {
		t.Error("missing edge has an inversion tag")
	}
	for _, e := range []Edge{g.Const1(), a} {
		if Aig_ObjFaninC0(g, e) != 0 || Aig_ObjFaninC1(g, e) != 0 {
			t.Errorf("terminal %d has complemented missing fanins", e.NodeID())
		}
	}
	if Aig_ObjFaninC0(g, output) != 1 || Aig_ObjFaninC1(g, output) != 0 {
		t.Error("output fanin phases differ from the source")
	}
	if Aig_ObjFaninC0(g, buffer) != 0 || Aig_ObjFaninC1(g, buffer) != 0 {
		t.Error("buffer has a complemented missing fanin")
	}
}
