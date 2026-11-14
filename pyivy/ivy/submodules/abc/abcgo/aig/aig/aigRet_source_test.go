package aig

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"slices"
	"strings"
	"testing"
	"weak"
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

func TestRetimingDirectReconstructionMetadata(t *testing.T) {
	for _, move := range []bool{false, true} {
		t.Run(fmt.Sprintf("move=%t", move), func(t *testing.T) {
			g := retFExample(1, false, false, false)
			g.FlopNumbers = []int32{17, 29}
			g.FlopReprs = []int32{17, 29}
			g.ClockDomains = [][]int32{{0, 1}}
			g.OneHotGroups = [][]int32{{0, 1}}
			g.OneHotInputGroups = [][]int32{{0, 1}, {2, 3}}
			p := Rtm_ManFromAig(g)
			for _, lo := range p.latchObjs {
				Rtm_ObjAddFirst(p, lo.edges[0], RTM_VAL_ZERO)
			}
			if move {
				for _, lo := range p.latchObjs {
					Rtm_ObjRetimeFwd(p, lo)
				}
				for _, o := range p.vObjs {
					if len(o.edges) == 2 {
						Rtm_ObjRetimeFwd(p, o)
					}
				}
			}
			out := Rtm_ManToAig(p)
			if move {
				if out.FlopNumbers != nil || out.FlopReprs != nil || !out.RetimingChanged || len(out.OneHotGroups) != 0 || len(out.OneHotInputGroups) != 1 {
					t.Errorf("direct reconstruction retained claims about replaced registers")
				}
				if len(out.ClockDomains) != 1 || !slices.Equal(out.ClockDomains[0], []int32{0}) {
					t.Errorf("direct reconstruction domain = %v, want [[0]]", out.ClockDomains)
				}
			} else if !slices.Equal(out.FlopNumbers, g.FlopNumbers) || !slices.Equal(out.FlopReprs, g.FlopReprs) || out.RetimingChanged || len(out.OneHotGroups) != 1 || len(out.OneHotInputGroups) != 2 {
				t.Error("unchanged reconstruction lost original register metadata")
			}
			if len(g.ClockDomains[0]) != 2 || g.FlopNumbers[0] != 17 || len(g.OneHotGroups) != 1 {
				t.Error("reconstruction changed source metadata")
			}
			if !t.Failed() {
				checkRetFTraces(t, g, out)
			}
		})
	}
}

func TestRetimingReconstructionInitialEncodingMetadata(t *testing.T) {
	for _, value := range []Rtm_Init_t{RTM_VAL_ZERO, RTM_VAL_ONE, RTM_VAL_VOID} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			g := NewGraph()
			a, q := g.NewInput(), g.NewInput()
			g.AddOutput(q)
			g.AddOutput(a)
			g.nRegs = 1
			g.FlopNumbers = []int32{19}
			p := Rtm_ManFromAig(g)
			Rtm_ObjAddFirst(p, p.latchObjs[0].edges[0], value)
			out := Rtm_ManToAig(p)
			if value == RTM_VAL_ONE {
				if out.FlopNumbers != nil || !out.RetimingChanged {
					t.Error("complemented register encoding retained original flop identity")
				}
			} else if !slices.Equal(out.FlopNumbers, g.FlopNumbers) || out.RetimingChanged {
				t.Error("unchanged zero encoding lost original flop identity")
			}
			if got := retFTrace(t, out, [][]bool{{false}})[0][0]; got != (value == RTM_VAL_ONE) {
				t.Errorf("initial output = %t for initialization %d", got, value)
			}
		})
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
	if p.source != nil || p.latchObjs != nil || p.sourceMap != nil || p.vObjs != nil {
		t.Error("freed retiming manager retains its graph, latch objects or allocation state")
	}
	if err := g.Check(); err != nil {
		t.Fatalf("freeing retiming state damaged borrowed graph: %v", err)
	}
}

func TestRetimingRetiredOverflowStorageIsCollectible(t *testing.T) {
	for _, grow := range []bool{false, true} {
		t.Run(fmt.Sprintf("grow=%t", grow), func(t *testing.T) {
			p := Rtm_ManAlloc(nil)
			edge := &Rtm_Edg_t{}
			retired := retireRetimingOverflow(p, edge, grow)
			runtime.GC()
			if retired.Value() != nil {
				t.Error("retiming manager retains overflow storage no longer used by its edge")
			}
			// Both owners remain live: only the obsolete backing array may go.
			runtime.KeepAlive(edge)
			runtime.KeepAlive(p)
		})
	}
}

func retireRetimingOverflow(p *Rtm_Man_t, edge *Rtm_Edg_t, grow bool) weak.Pointer[uint32] {
	for i := 0; i < 100; i++ {
		Rtm_ObjAddLast(p, edge, RTM_VAL_ZERO)
	}
	// Seven words exceed the runtime's tiny-allocation batching threshold.
	retired := weak.Make(&edge.big[0])
	if grow {
		for i := 0; i < 100; i++ {
			Rtm_ObjAddLast(p, edge, RTM_VAL_ONE)
		}
	} else {
		for edge.nLats > 0 {
			Rtm_ObjRemFirst(p, edge)
		}
	}
	return retired
}

func TestRetimingPackedStorageAgainstCompiledSource(t *testing.T) {
	p := Rtm_ManAlloc(nil)
	var edges [3]Rtm_Edg_t
	var input, want strings.Builder
	apply := func(id, action, value int) {
		fmt.Fprintf(&input, "%d %d %d\n", id, action, value)
		e := &edges[id]
		var result Rtm_Init_t
		switch action {
		case 0:
			Rtm_ObjAddFirst(p, e, Rtm_Init_t(value))
		case 1:
			Rtm_ObjAddLast(p, e, Rtm_Init_t(value))
		case 2:
			result = Rtm_ObjRemFirst(p, e)
		case 3:
			result = Rtm_ObjRemLast(p, e)
		}
		fmt.Fprintf(&want, "%d", result)
		for i := range edges {
			e := &edges[i]
			fmt.Fprintf(&want, " | %d:", e.nLats)
			for j := int32(0); j < e.nLats; j++ {
				fmt.Fprintf(&want, "%d", Rtm_ObjGetOne(p, e, j))
			}
		}
		want.WriteByte('\n')
	}
	// Interleaved edges exercise shared C-pool reallocations, compact/overflow
	// transitions, both ends and every two-bit value across six word boundaries.
	for cycle := 0; cycle < 2; cycle++ {
		for n := 0; n < 100; n++ {
			for id := range edges {
				apply(id, (n+id+cycle)%2, 1+(n+id)%3)
			}
		}
		for n := 0; n < 100; n++ {
			for id := range edges {
				apply(id, 2+(n+id+cycle)%2, 0)
			}
		}
	}
	rng := rand.New(rand.NewPCG(59, 127))
	for n := 0; n < 2000; n++ {
		id, action := rng.IntN(3), rng.IntN(4)
		if edges[id].nLats == 0 {
			action %= 2
		} else if edges[id].nLats >= 100 {
			action = 2 + action%2
		}
		apply(id, action, 1+rng.IntN(3))
	}
	got := runAigSourceFixtureInput(t, "retiming_storage_compare.c", input.String())
	gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want.String(), "\n")
	if len(gotLines) != len(wantLines) {
		t.Fatalf("source trace length = %d, want %d", len(gotLines), len(wantLines))
	}
	for i := range wantLines {
		if gotLines[i] != wantLines[i] {
			t.Fatalf("operation %d: Go %q, C %q", i, wantLines[i], gotLines[i])
		}
	}
}
