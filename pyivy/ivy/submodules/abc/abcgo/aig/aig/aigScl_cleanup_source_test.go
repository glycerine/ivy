package aig

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

func TestSeqCleanupBoundariesAgainstCompiledSource(t *testing.T) {
	got := runAigSourceFixture(t, "seq_cleanup_boundaries.c", "aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "aigScl.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	want := "full_buffer 0 1 1\nbasic_buffer 0 1 1\nbasic_fanouts 12 16 14\nbasic_patch 0 0 0\nfull_ids 1 1 1\nremoved_ci_buffers 2 1\n"
	if got != want {
		t.Fatalf("source sequential cleanup boundaries:\n%s, want\n%s", got, want)
	}
}

func TestSeqCleanupBasicPreservesCircularFanoutOrder(t *testing.T) {
	g := NewGraph()
	Aig_ManFanoutStart(g)
	pi, dead, live := g.NewInput(), g.NewInput(), g.NewInput()
	gate := g.AND(pi, dead)
	g.AddOutput(live)
	constantPO := g.AddOutput(g.Const1())
	patchedLI := g.AddOutput(gate)
	liveLI := g.AddOutput(g.Const1())
	Aig_ManSetRegNum(g, 2)
	Aig_ManSeqCleanupBasic(g)
	want := []int32{
		Aig_FanoutCreate(int32(constantPO.NodeID()), 0),
		Aig_FanoutCreate(int32(liveLI.NodeID()), 0),
		Aig_FanoutCreate(int32(patchedLI.NodeID()), 0),
	}
	if got := g.fanouts.Handles(g.Const1()); !slices.Equal(got, want) {
		t.Fatalf("fanout order %v, want source order %v", got, want)
	}
}

func TestSeqCleanupRetainsStandaloneBufferLikeSource(t *testing.T) {
	for _, cleanup := range []func(*Graph) int32{Aig_ManSeqCleanup, Aig_ManSeqCleanupBasic} {
		g := NewGraph()
		a, b := g.NewInput(), g.NewInput()
		gate := g.AND(a, b)
		buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{gate, InvalidEdge}})
		if removed := cleanup(g); removed != 0 || !g.valid(buffer) || !g.valid(gate) {
			t.Errorf("standalone buffer: removed %d, buffer live %t, gate live %t", removed, g.valid(buffer), g.valid(gate))
		}
	}
}

func TestSeqCleanupBasicCountsOnlyFinalCleanup(t *testing.T) {
	g := NewGraph()
	a, lo := g.NewInput(), g.NewInput()
	g.AddOutput(a)
	dead := g.AND(a, lo)
	g.AddOutput(dead)
	Aig_ManSetRegNum(g, 1)
	if removed := Aig_ManSeqCleanupBasic(g); removed != 0 {
		t.Errorf("returned %d gates, want zero: source patch already deletes the old cone", removed)
	}
	if g.valid(dead) {
		t.Error("patch did not delete the dangling gate")
	}
}

func TestSeqCleanupRefreshesTerminalIdsLikeSource(t *testing.T) {
	g := NewGraph()
	pi, dead, live := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(live)
	g.AddOutput(dead)
	g.AddOutput(pi)
	Aig_ManSetRegNum(g, 2)
	Aig_ManSeqCleanup(g)
	if g.node(live).CioID != 1 || g.node(g.outputs[1]).CioID != 1 {
		t.Fatalf("retained register/LI IDs %d/%d, want 1/1", g.node(live).CioID, g.node(g.outputs[1]).CioID)
	}
}

func TestSeqCleanupBasicClearsLatchOutputUnion(t *testing.T) {
	g := NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	Aig_ManSetRegNum(g, 1)
	g.next = map[int]Edge{lo.NodeID(): pi}
	Aig_ManSeqCleanupBasic(g)
	if g.node(lo).CioID != 0 || g.next[lo.NodeID()] != 0 {
		t.Fatalf("source clears the latch-output pNext/CioId union, got %d/%d", g.node(lo).CioID, g.next[lo.NodeID()])
	}
}

func TestSeqCleanupOnlyReclaimsBuffersWithRemovedInputs(t *testing.T) {
	g := NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	unsafe := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{lo, InvalidEdge}})
	safe := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{pi, InvalidEdge}})
	g.AddOutput(pi)
	g.AddOutput(pi)
	Aig_ManSetRegNum(g, 1)
	if removed := Aig_ManSeqCleanup(g); removed != 0 || !g.valid(safe) || g.valid(unsafe) || g.valid(lo) {
		t.Fatalf("removed %d gates; safe buffer %t, removed-input buffer %t, removed LO %t", removed, g.valid(safe), g.valid(unsafe), g.valid(lo))
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestSeqCleanupSharedRemovedInputBufferCones(t *testing.T) {
	for _, fanouts := range []bool{false, true} {
		g := NewGraph()
		if fanouts {
			Aig_ManFanoutStart(g)
		}
		pi, dead, live := g.NewInput(), g.NewInput(), g.NewInput()
		shared := g.AND(pi, dead)
		xor := g.NewXor(shared, dead.Not())
		buffer := func(driver Edge) Edge {
			return Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{driver, InvalidEdge}})
		}
		bad := []Edge{buffer(shared.Not()), buffer(buffer(xor))}
		safeGate := g.AND(pi, live)
		safe := buffer(safeGate)
		g.AddOutput(live)
		g.AddOutput(pi)
		g.AddOutput(pi.Not())
		Aig_ManSetRegNum(g, 2)
		original := Aig_ManDupSimple(g)
		if removed := Aig_ManSeqCleanup(g); removed != 2 {
			t.Fatalf("fanouts %t: removed %d gates, want two", fanouts, removed)
		}
		for _, edge := range append(bad, shared, xor, dead) {
			if g.valid(edge) {
				t.Fatalf("fanouts %t: removed-register cone object %d survived", fanouts, edge)
			}
		}
		if !g.valid(safe) || !g.valid(safeGate) || g.nRegs != 1 {
			t.Fatal("independent buffer cone or live register was lost")
		}
		if err := g.Check(); err != nil {
			t.Fatal(err)
		}
		sclEqualTrace(t, original, g)
	}
}

func TestSeqCleanupStatesAgainstCompiledSource(t *testing.T) {
	rng := rand.New(rand.NewPCG(23, 47))
	var input, want strings.Builder
	for trial := 0; trial < 32; trial++ {
		for regs := 0; regs <= 3; regs++ {
			for basic := 0; basic < 2; basic++ {
				for fanouts := 0; fanouts < 2; fanouts++ {
					g := NewGraph()
					available := []Edge{g.Const1(), g.NewInput(), g.NewInput(), g.NewInput()}
					for step := 0; step < 16; step++ {
						a, b := available[rng.IntN(len(available))], available[rng.IntN(len(available))]
						if rng.IntN(2) == 0 {
							a = a.Not()
						}
						var root Edge
						switch rng.IntN(3) {
						case 0:
							root = g.AND(a, b)
						case 1:
							root = g.NewXor(a, b)
						default:
							// Keep standalone buffer roots independent of removable
							// LOs; the upstream invalid-input corner has its own test.
							driver := available[rng.IntN(4-regs)]
							root = Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{driver, InvalidEdge}})
						}
						available = append(available, root)
					}
					for i := 0; i < 4; i++ {
						g.AddOutput(available[rng.IntN(len(available))])
					}
					Aig_ManSetRegNum(g, int32(regs))
					g.nTruePis, g.nTruePos = 77, 88
					g.FlopNumbers = make([]int32, regs)
					for i := range g.FlopNumbers {
						g.FlopNumbers[i] = int32(100 + i)
					}
					for i, ci := range g.inputs {
						g.node(ci).CioID = int32(10 + i)
					}
					for i, co := range g.outputs {
						g.node(co).CioID, g.node(co).Level = int32(20+i), 99
					}
					fmt.Fprintf(&input, "%d %d %d %d\n", len(g.nodes), fanouts, regs, basic)
					for _, node := range g.nodes[1:] {
						a, b := int64(-1), int64(-1)
						pins := nodeFanins(node)
						if len(pins) > 0 {
							a = int64(pins[0])
						}
						if len(pins) > 1 {
							b = int64(pins[1])
						}
						fmt.Fprintf(&input, "%d %d %d\n", node.Kind, a, b)
					}
					if fanouts != 0 {
						Aig_ManFanoutStart(g)
					}
					for pass := 0; pass < 2; pass++ {
						cleanup := Aig_ManSeqCleanup
						if basic != 0 {
							cleanup = Aig_ManSeqCleanupBasic
						}
						removed := cleanup(g)
						fmt.Fprintf(&want, "removed %d regs %d true %d/%d gen %d hash %d | ci", removed, g.nRegs, g.nTruePis, g.nTruePos, g.travID, Aig_TableCountEntries(g))
						for _, ci := range g.inputs {
							fmt.Fprintf(&want, " %d", ci.NodeID())
						}
						want.WriteString(" | co")
						for _, co := range g.outputs {
							fmt.Fprintf(&want, " %d", co.NodeID())
						}
						want.WriteString(" | flops")
						for _, number := range g.FlopNumbers {
							fmt.Fprintf(&want, " %d", number)
						}
						want.WriteString(" | objects")
						for id, node := range g.nodes {
							if g.deleted[id] {
								want.WriteString(" dead")
								continue
							}
							a, b, cio := int64(-1), int64(-1), int32(-1)
							pins := nodeFanins(node)
							if len(pins) > 0 {
								a = int64(pins[0])
							}
							if len(pins) > 1 {
								b = int64(pins[1])
							}
							if node.Kind == Input || node.Kind == Output {
								cio = node.CioID
							}
							fmt.Fprintf(&want, " %d/%d/%d/%d/%d/%d", node.Refs, a, b, node.Level, node.TravID, cio)
							if fanouts != 0 {
								want.WriteByte('[')
								for _, pin := range g.fanouts.Handles(Edge(id) << 1) {
									fmt.Fprintf(&want, "%d,", pin)
								}
								want.WriteByte(']')
							}
						}
						want.WriteByte('\n')
						if err := g.Check(); err != nil {
							t.Fatalf("trial %d regs %d basic %d pass %d: %v", trial, regs, basic, pass, err)
						}
					}
				}
			}
		}
	}
	got := runAigSourceFixtureInput(t, "seq_cleanup_compare.c", input.String(), "aigMan.c", "aigObj.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "aigScl.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	if got != want.String() {
		c, goLines := strings.Split(got, "\n"), strings.Split(want.String(), "\n")
		for i := 0; i < len(c) && i < len(goLines); i++ {
			if c[i] != goLines[i] {
				t.Fatalf("state %d: C %q, Go %q", i, c[i], goLines[i])
			}
		}
		t.Fatalf("state count C %d, Go %d", len(c), len(goLines))
	}
}
