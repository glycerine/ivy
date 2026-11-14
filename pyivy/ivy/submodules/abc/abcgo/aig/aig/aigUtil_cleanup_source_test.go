package aig

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestCleanNextClearsSourceCioUnion(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	output := g.AddOutput(gate)
	g.next = map[int]Edge{a.NodeID(): b, gate.NodeID(): a}
	for i := range g.nodes {
		g.nodes[i].CioID = int32(100 + i)
		g.nodes[i].Xsim = 2
	}
	g.objectData = map[int]any{gate.NodeID(): "keep"}
	g.copies = map[int]Edge{gate.NodeID(): output}
	before := g.Nodes()
	Aig_ManCleanNext(g)
	if len(g.next) != 0 {
		t.Error("next bindings survived cleanup")
	}
	for i, n := range g.nodes {
		if n.CioID != 0 {
			t.Errorf("object %d kept source CioId/pNext union %d", i, n.CioID)
		}
		before[i].CioID = 0
	}
	if !reflect.DeepEqual(before, g.Nodes()) || g.objectData[gate.NodeID()] != "keep" || g.copies[gate.NodeID()] != output {
		t.Error("next cleanup changed unrelated state")
	}
}

func TestRecursiveDataCleanRejectsSourceInvalidRoots(t *testing.T) {
	for _, co := range []bool{false, true} {
		g := NewGraph()
		a, b := g.NewInput(), g.NewInput()
		root := g.AND(a, b)
		if co {
			root = g.AddOutput(root)
		} else {
			root = root.Not()
		}
		g.objectData = map[int]any{root.NodeID(): "keep"}
		g.copies = map[int]Edge{root.NodeID(): a}
		panicked := false
		func() { defer func() { panicked = recover() != nil }(); Aig_ObjCleanData_rec(g, root) }()
		if !panicked {
			t.Errorf("accepted source-invalid root %v (CO=%t)", root, co)
		}
		if g.objectData[root.NodeID()] != "keep" || g.copies[root.NodeID()] != a {
			t.Error("invalid cleanup root changed payloads")
		}
	}
}

func TestUtilityCleanupMatchesSourceStorageUnions(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	g.AddOutput(root.Not())
	g.AND(a, c)
	for id := range g.nodes {
		g.nodes[id].CioID = int32(100 + id)
		g.nodes[id].Xsim = 2
		g.SetObjectData(Edge(id)<<1, int32(42+id))
	}
	var want strings.Builder
	dump := func(stage string) {
		for _, e := range g.LiveEdges() {
			id := e.NodeID()
			var data int32
			if payload := g.ObjectData(e); payload != nil {
				data = payload.(int32)
			}
			fmt.Fprintf(&want, "%s %d cio %d data %d xsim %d\n", stage, id, g.Node(e).CioID, data, g.Node(e).Xsim)
		}
	}
	Aig_ManCleanNext(g)
	dump("next")
	Aig_ManCleanData(g)
	dump("data")
	for id := range g.nodes {
		g.SetObjectData(Edge(id)<<1, int32(42+id))
	}
	Aig_ObjCleanData_rec(g, root)
	dump("cone")
	source := runAigSourceFixture(t, "utility_cleanup_compare.c", "aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	if source != want.String() {
		t.Fatalf("Go cleanup states:\n%sC cleanup states:\n%s", want.String(), source)
	}
}
