package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestRetimingManagerUsesSourceObjectGroups(t *testing.T) {
	g := NewGraph()
	a, b, q, r := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	gate := g.AND(q, r)
	po, li0, li1 := g.AddOutput(gate), g.AddOutput(a), g.AddOutput(b)
	g.nRegs = 2
	p := Rtm_ManFromAig(g)
	want := []Edge{g.Const1(), a, b, po, q, r, li0, li1, gate}
	for id, e := range want {
		if got := p.sourceMap[e.NodeID()].Id; got != int32(id) {
			t.Errorf("object %d retiming ID = %d, want source group position %d", e, got, id)
		}
	}
}

func TestRetimingReconstructionCleansDanglingLogic(t *testing.T) {
	g := NewGraph()
	a, b, lo := g.NewInput(), g.NewInput(), g.NewInput()
	g.AND(a, b) // Deliberately dangling, but valid input logic.
	g.AddOutput(lo)
	g.AddOutput(a)
	g.nRegs = 1
	p := Rtm_ManFromAig(g)
	for _, o := range p.latchObjs {
		Rtm_ObjAddFirst(p, o.edges[0], RTM_VAL_ZERO)
	}
	out := Rtm_ManToAig(p)
	if Aig_ManNodeNum(out) != 0 {
		t.Errorf("reconstruction retained %d dangling gates", Aig_ManNodeNum(out))
	}
	if out.nRegs != 1 {
		t.Error("reconstruction lost original delay")
	}
	if err := out.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestEdgeRetimingPerformsSourceRegisterSharing(t *testing.T) {
	g := NewGraph()
	a, q, r := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(q)
	g.AddOutput(r)
	g.AddOutput(a)
	g.AddOutput(a)
	g.nRegs = 2
	for _, depth := range []int32{-1, 0, 1} {
		out := Rtm_ManRetime(g, 1, depth, 0)
		if out.nRegs != 1 {
			t.Errorf("depth %d left %d identical registers, source shares them", depth, out.nRegs)
		}
		checkRetFTraces(t, g, out)
	}
}

func TestRetimingAllocationRetainsSourceConstructionState(t *testing.T) {
	p := Rtm_ManAlloc(nil)
	driver := Rtm_ObjAlloc(p, 0, 1)
	po := Rtm_ObjAlloc(p, 1, 0)
	if !po.Type || po.Num != 1 || po.Temp != 0 || driver.Temp != 1 {
		t.Error("allocation omitted source PO flag and construction counts")
	}
}

func TestRetimingConnectionsEnforceSourceCapacities(t *testing.T) {
	for _, which := range []string{"fanin", "fanout", "third-pin"} {
		t.Run(which, func(t *testing.T) {
			p := Rtm_ManAlloc(nil)
			driver := Rtm_ObjAlloc(p, 0, 3)
			sink := Rtm_ObjAlloc(p, 1, 0)
			defer func() {
				if recover() == nil {
					t.Error("connection exceeded source capacity")
				}
			}()
			switch which {
			case "fanin":
				Rtm_ObjAddFanin(sink, driver, 0)
				Rtm_ObjAddFanin(sink, driver, 0)
			case "fanout":
				driver.Temp = 0
				Rtm_ObjAddFanin(sink, driver, 0)
			case "third-pin":
				sink.Num = 3
				Rtm_ObjAddFanin(sink, driver, 0)
				Rtm_ObjAddFanin(sink, driver, 0)
				Rtm_ObjAddFanin(sink, driver, 0)
			}
		})
	}
}

func TestRetimingManagerAndForwardMovesAgainstCompiledSource(t *testing.T) {
	out := runAigSourceFixture(t, "retiming_manager_compare.c", "aigMem.c")
	g := NewGraph()
	a, b, q, r := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	gate := g.AND(q.Not(), r.Not())
	g.AddOutput(gate)
	g.AddOutput(a)
	g.AddOutput(b)
	g.nRegs = 2
	p := Rtm_ManFromAig(g)
	defer Rtm_ManFree(p)
	var got strings.Builder
	got.WriteString("mapping:")
	for id := range g.nodes {
		fmt.Fprintf(&got, " %d", p.sourceMap[id].Id)
	}
	got.WriteByte('\n')
	for _, o := range p.vObjs {
		typeFlag := 0
		if o.Type {
			typeFlag = 1
		}
		fmt.Fprintf(&got, "%d type %d capacity %d/%d fanins:", o.Id, typeFlag, o.Num, o.Temp)
		for i, f := range o.fanins {
			c := 0
			if o.complements[i] {
				c = 1
			}
			fmt.Fprintf(&got, " %d/%d", f.Id, c)
		}
		got.WriteString(" fanouts:")
		for _, f := range o.fanouts {
			fmt.Fprintf(&got, " %d", f.Id)
		}
		got.WriteByte('\n')
		o.Num = 0
	}
	fmt.Fprintf(&got, "autonomous %d %d\n", Rtm_ManMarkAutoFwd(p), Rtm_ManMarkAutoBwd(p))
	for _, o := range p.latchObjs {
		Rtm_ObjAddFirst(p, o.edges[0], RTM_VAL_ZERO)
	}
	for _, e := range []Edge{q, r, gate} {
		Rtm_ObjRetimeFwd(p, p.sourceMap[e.NodeID()])
	}
	fmt.Fprintf(&got, "max %d\n", Rtm_ManLatchMax(p))
	for _, o := range p.vObjs {
		for i, e := range o.edges {
			fmt.Fprintf(&got, "edge %d/%d %d:", o.Id, i, e.nLats)
			for j := int32(0); j < e.nLats; j++ {
				fmt.Fprintf(&got, " %d", Rtm_ObjGetOne(p, e, j))
			}
			got.WriteByte('\n')
		}
	}
	if got.String() != out {
		t.Fatalf("Go retiming manager:\n%sC retiming manager:\n%s", got.String(), out)
	}
	result := Rtm_ManToAig(p)
	checkRetFTraces(t, g, result)
}

func TestEdgeRetimingDropsStaleRegisterOneHotAnnotations(t *testing.T) {
	g := NewGraph()
	a, b, q, r := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(q, r))
	g.AddOutput(a)
	g.AddOutput(b)
	g.nRegs = 2
	g.OneHotGroups = [][]int32{{0, 1}}
	g.OneHotInputGroups = [][]int32{{0, 1}, {2, 3}, {0, 2}}
	out := Rtm_ManRetime(g, 1, 1, 0)
	if out.nRegs != 1 || len(out.OneHotGroups) != 0 {
		t.Error("retimed graph retained one-hot claims about replaced registers")
	}
	if len(out.OneHotInputGroups) != 1 || len(out.OneHotInputGroups[0]) != 2 || out.OneHotInputGroups[0][0] != 0 || out.OneHotInputGroups[0][1] != 1 {
		t.Errorf("one-hot input annotations = %v, want only unchanged true-PI group", out.OneHotInputGroups)
	}
	if len(g.OneHotGroups) != 1 || len(g.OneHotInputGroups) != 3 {
		t.Error("retiming changed source annotations")
	}
}

func TestRetimingFreeReleasesRetainedOwnership(t *testing.T) {
	g := NewGraph()
	a, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(a)
	g.nRegs = 1
	p := Rtm_ManFromAig(g)
	for i := 0; i < 20; i++ {
		Rtm_ObjAddLast(p, p.latchObjs[0].edges[0], RTM_VAL_ZERO)
	}
	Rtm_ManFree(p)
	if p.source != nil || p.latchObjs != nil || p.nExtraCur != 0 {
		t.Error("freed retiming manager retains its graph, latch objects or allocation state")
	}
	if err := g.Check(); err != nil {
		t.Fatalf("freeing retiming state damaged borrowed graph: %v", err)
	}
}
