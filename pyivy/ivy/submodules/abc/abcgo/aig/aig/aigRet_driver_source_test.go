package aig

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
)

func TestRetimingCompleteDriverAgainstCompiledSource(t *testing.T) {
	graphs := []*Graph{
		retFExample(1, false, false, false),
		retFExample(1, true, true, false),
		retFExample(3, true, false, false),
		retimingBackwardInitialExample(),
	}
	// Observable autonomous feedback and an unobservable PI-driven component
	// distinguish forward PI reachability from backward PO reachability.
	g := NewGraph()
	a, q, r, s := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(q, r.Not()))
	g.AddOutput(a)
	g.AddOutput(s)
	g.AddOutput(r.Not())
	g.nRegs = 3
	graphs = append(graphs, g)
	g = NewGraph()
	a, q, r = g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(q)
	g.AddOutput(a)
	g.AddOutput(g.AND(a, r.Not()))
	g.nRegs = 2
	graphs = append(graphs, g)
	rng := rand.New(rand.NewPCG(89, 151))
	for trial := 0; trial < 20; trial++ {
		g := NewGraph()
		pool := []Edge{g.Const1(), g.NewInput(), g.NewInput()}
		for i := 0; i < 3; i++ {
			pool = append(pool, g.NewInput())
		}
		literal := func() Edge {
			e := pool[rng.IntN(len(pool))]
			if rng.IntN(2) != 0 {
				e = e.Not()
			}
			return e
		}
		for i := 0; i < 8; i++ {
			pool = append(pool, g.AND(literal(), literal()))
		}
		for i := 0; i < 5; i++ {
			g.AddOutput(literal())
		}
		g.nRegs = 3
		graphs = append(graphs, g)
	}
	var input, want strings.Builder
	var caseNames []string
	for index, g := range graphs {
		for _, forward := range []int32{0, 1} {
			for _, steps := range []int32{-1, 0, 1, 2, 5} {
				caseNames = append(caseNames, fmt.Sprintf("graph=%d forward=%d steps=%d", index, forward, steps))
				writeRetimingSourceCase(t, &input, g, forward, steps)
				out := Rtm_ManRetime(g, forward, steps, 0)
				if err := out.Check(); err != nil {
					t.Fatalf("%s: %v", caseNames[len(caseNames)-1], err)
				}
				fmt.Fprintf(&want, "registers %d\n", out.nRegs)
				npis := len(g.inputs) - int(g.nRegs)
				sequence := make([][]bool, 64)
				for frame := range sequence {
					sequence[frame] = make([]bool, npis)
					for i := range sequence[frame] {
						sequence[frame][i] = ((frame*13+(frame>>2)*7)>>i)&1 != 0
					}
				}
				for _, frame := range retFTrace(t, out, sequence) {
					for _, bit := range frame {
						if bit {
							want.WriteByte('1')
						} else {
							want.WriteByte('0')
						}
					}
					want.WriteByte('\n')
				}
			}
		}
	}
	got := runAigSourceFixtureInput(t, "retiming_driver_compare.c", input.String(),
		"aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c",
		"aigFanout.c", "aigUtil.c", "aigCheck.c", "aigScl.c", "aigDup.c", "aigDfs.c",
		"../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want.String(), "\n")
	if len(gotLines) != len(wantLines) {
		t.Fatalf("source trace length = %d, want %d: %s", len(gotLines), len(wantLines), got)
	}
	for i := range wantLines {
		if gotLines[i] != wantLines[i] {
			name := caseNames[i/65]
			t.Fatalf("%s line %d: Go %q, C %q", name, i%65, wantLines[i], gotLines[i])
		}
	}
}

func retimingBackwardInitialExample() *Graph {
	g := NewGraph()
	a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(q)
	g.AddOutput(g.AND(a.Not(), b.Not()))
	g.nRegs = 1
	return g
}

func TestRetimingBackwardDoesNotPreserveZeroReset(t *testing.T) {
	g := retimingBackwardInitialExample()
	sequence := [][]bool{{false, false}}
	if retFTrace(t, g, sequence)[0][0] {
		t.Fatal("original register must initially be zero")
	}
	// The complete compiled-source comparison includes this graph. C inserts
	// VOID latches and reconstructs them as zero, without solving their reset
	// values. Complemented inputs then make the new AND output initially one.
	out := Rtm_ManRetime(g, 0, 1, 0)
	if !retFTrace(t, out, sequence)[0][0] {
		t.Fatal("backward reconstruction no longer matches C's reset limitation")
	}
}

func TestRetimingEdgeRequiresOneCompleteClockDomain(t *testing.T) {
	for _, domain := range [][]int32{{}, {0}, {0, 0}, {-1, 1}, {0, 2}} {
		t.Run(fmt.Sprint(domain), func(t *testing.T) {
			g := retFExample(1, false, false, false)
			g.ClockDomains = [][]int32{domain}
			defer func() {
				if recover() == nil {
					t.Error("retiming accepted incomplete or invalid clock-domain membership")
				}
			}()
			Rtm_ManRetime(g, 1, 1, 0)
		})
	}
	g := retFExample(1, true, true, false)
	g.ClockDomains = [][]int32{{1, 0}}
	out := Rtm_ManRetime(g, 1, 1, 0)
	if len(out.ClockDomains) != 1 || len(out.ClockDomains[0]) != int(out.nRegs) {
		t.Fatalf("retimed domain does not match new register count: %v", out.ClockDomains)
	}
	for i, reg := range out.ClockDomains[0] {
		if reg != int32(i) {
			t.Fatalf("retimed domain membership = %v", out.ClockDomains)
		}
	}
	if g.ClockDomains[0][0] != 1 || g.ClockDomains[0][1] != 0 {
		t.Fatal("retiming mutated source clock-domain membership")
	}
	checkRetFTraces(t, g, out)
}

func writeRetimingSourceCase(t *testing.T, input *strings.Builder, g *Graph, forward, steps int32) {
	t.Helper()
	var nodes []Edge
	literals := map[int]int{0: 0}
	for i, e := range g.inputs {
		literals[e.NodeID()] = 2 * (i + 1)
	}
	for id, n := range g.nodes {
		if !g.deleted[id] && n.Kind == And {
			literals[id] = 2 * (1 + len(g.inputs) + len(nodes))
			nodes = append(nodes, Edge(id)<<1)
		}
	}
	writeLiteral := func(e Edge) {
		lit, ok := literals[e.NodeID()]
		if !ok {
			t.Fatalf("unsupported fixture edge %d", e)
		}
		if e.Complemented() {
			lit ^= 1
		}
		fmt.Fprintf(input, "%d ", lit)
	}
	fmt.Fprintf(input, "%d %d %d %d %d %d\n", len(g.inputs)-int(g.nRegs), g.nRegs,
		len(nodes), len(g.outputs)-int(g.nRegs), forward, steps)
	for _, e := range nodes {
		for _, f := range g.node(e).Fanins {
			writeLiteral(f)
		}
		input.WriteByte('\n')
	}
	for _, e := range g.outputs {
		writeLiteral(g.outputDriver(e))
	}
	input.WriteByte('\n')
}
