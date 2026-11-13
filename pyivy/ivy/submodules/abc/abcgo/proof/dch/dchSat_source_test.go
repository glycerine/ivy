package dch

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestDchProvedImplicationsAndConstantShortcut(t *testing.T) {
	for _, flip := range []bool{false, true} {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		left := g.AND(g.AND(a, b), c)
		right := g.AND(a, g.AND(b, c))
		zero := g.AND(g.AND(a, b), g.AND(a, b.Not()))
		g.AddOutput(left)
		g.AddOutput(right)
		g.AddOutput(zero)
		pars := DefaultParams()
		pars.PolarFlip = flip
		m, err := newManager(g, &pars)
		if err != nil {
			t.Fatal(err)
		}
		m.recycleSolver()
		if _, err := m.satLiteral(left); err != nil {
			t.Fatal(err)
		}
		if _, err := m.satLiteral(right); err != nil {
			t.Fatal(err)
		}
		before := m.solver.Statistics().Clauses
		result, _, err := m.provePair(left, right)
		if err != nil || result != proofEquivalent {
			t.Fatalf("flip %v: result %v error %v", flip, result, err)
		}
		if m.solver.Statistics().Clauses < before+2 {
			t.Fatalf("flip %v: proved implications absent: before=%d after=%d", flip, before, m.solver.Statistics().Clauses)
		}
		beforeCalls := m.stats.SATCalls
		result, _, err = m.provePair(g.Const1(), zero)
		if err != nil || result != proofEquivalent || m.stats.SATCalls != beforeCalls+1 {
			t.Fatalf("flip %v: constant result %v calls %d error %v", flip, result, m.stats.SATCalls-beforeCalls, err)
		}
	}
}

func TestDchSweepSubstitutesProvedNodesBeforeLaterCones(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	x := g.AND(g.AND(a, b), c)
	y := g.AND(a, g.AND(b, c))
	z1, z2 := g.AND(x, d), g.AND(y, d)
	g.AddOutput(x)
	g.AddOutput(y)
	g.AddOutput(z1)
	g.AddOutput(z2)
	m, err := newManager(g, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.computeEquivalences(); err != nil {
		t.Fatal(err)
	}
	if m.mapped[x.NodeID()] != m.mapped[y.NodeID()] || m.mapped[z1.NodeID()] != m.mapped[z2.NodeID()] {
		t.Fatalf("proved nodes were not substituted in later cones: %v", m.mapped)
	}
	if m.stats.Proofs < 2 {
		t.Fatalf("expected both equivalences: %+v", m.stats)
	}
	for _, n := range []aig.Edge{x, y, z1, z2} {
		m.satGraph.AddOutput(m.mapped[n.NodeID()])
	}
	for bits := 0; bits < 16; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
		values := simulateAssignment(g, m.order, inputs)
		work, err := m.satGraph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		for i, n := range []aig.Edge{x, y, z1, z2} {
			if work[i] != values[n.NodeID()] {
				t.Fatalf("bits %d node %d changed", bits, n.NodeID())
			}
		}
	}
}

// Exercise source-style supergate and MUX clauses against direct truth tables.
func TestDchSourceCNFTruthTables(t *testing.T) {
	for _, flip := range []bool{false, true} {
		for _, mux := range []bool{false, true} {
			g := aig.NewGraph()
			a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
			var output aig.Edge
			if mux {
				output = g.Mux(a, b.Not(), c)
			} else {
				output = g.AND(g.AND(g.AND(a, b), c), d)
			}
			g.AddOutput(output)
			pars := DefaultParams()
			pars.PolarFlip = flip
			m, err := newManager(g, &pars)
			if err != nil {
				t.Fatal(err)
			}
			m.recycleSolver()
			out, err := m.satLiteral(output)
			if err != nil {
				t.Fatal(err)
			}
			if !mux && len(m.variables) != 6 {
				t.Fatalf("supergate introduced interior SAT variables: %v", m.variables)
			}
			for bits := 0; bits < 16; bits++ {
				in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
				expected, err := g.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				assumptions := make([]int32, 0, 5)
				for i, edge := range g.Inputs() {
					if _, used := m.variables[edge.NodeID()]; !used {
						continue
					}
					lit, err := m.satLiteral(edge)
					if err != nil {
						t.Fatal(err)
					}
					if !in[i] {
						lit ^= 1
					}
					assumptions = append(assumptions, lit)
				}
				target := out
				if !expected[0] {
					target ^= 1
				}
				if got := m.solver.Solve(append(assumptions, target)...); got != bsat.Satisfiable {
					t.Fatalf("flip=%v mux=%v bits=%d expected SAT got %v", flip, mux, bits, got)
				}
				if got := m.solver.Solve(append(assumptions, target^1)...); got != bsat.Unsatisfiable {
					t.Fatalf("flip=%v mux=%v bits=%d expected UNSAT got %v", flip, mux, bits, got)
				}
			}
		}
	}
}

func TestDchRejectsNativeNodesOutsideSourceContract(t *testing.T) {
	for _, kind := range []aig.NodeKind{aig.ExclusiveOr, aig.Buffer} {
		g := aig.NewGraph()
		a, b := g.NewInput(), g.NewInput()
		var node aig.Edge
		if kind == aig.ExclusiveOr {
			node = g.NewXor(a, b)
		} else {
			node = aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{a, aig.InvalidEdge}})
		}
		g.AddOutput(node)
		if result, _, err := ComputeChoicesWithStats(g, nil); err == nil || result != nil {
			t.Fatalf("node kind %v accepted: result=%v error=%v", kind, result, err)
		}
		if result, err := DeriveTotalAig([]*aig.Graph{g}); err == nil || result != nil {
			t.Fatalf("node kind %v accepted in snapshot: result=%v error=%v", kind, result, err)
		}
	}
}

// The C initial random passes refine nonconstant classes only; constant
// candidates discovered by the first pass wait for SAT counterexamples.
func TestDchInitialSimulationKeepsConstantCandidates(t *testing.T) {
	const inputs = 10
	random := aig.NewAigRandomSource()
	random.Uint32() // source reset draw
	samples := [2]map[int]bool{{}, {}}
	for pass := 0; pass < 2; pass++ {
		var rows [inputs]uint32
		for i := range rows {
			rows[i] = random.Uint32() << 1
		}
		for bit := 0; bit < 32; bit++ {
			mask := 0
			for i, row := range rows {
				if row&(1<<uint(bit)) != 0 {
					mask |= 1 << uint(i)
				}
			}
			samples[pass][mask] = true
		}
	}
	target := -1
	for mask := 1; mask < 1<<inputs; mask++ {
		if !samples[0][mask] && samples[1][mask] {
			target = mask
			break
		}
	}
	if target < 0 {
		t.Fatal("fixture has no second-round-only input")
	}
	g := aig.NewGraph()
	node := g.Const1()
	for i := 0; i < inputs; i++ {
		x := g.NewInput()
		if target&(1<<uint(i)) == 0 {
			x = x.Not()
		}
		node = g.AND(node, x)
	}
	g.AddOutput(node)
	order, err := graphOrder(g)
	if err != nil {
		t.Fatal(err)
	}
	classes := createClasses(g, order, 1)
	if got := classes.representative(node); got != g.Const1() {
		t.Fatalf("constant candidate lost during later random passes: representative %v", got)
	}
}
