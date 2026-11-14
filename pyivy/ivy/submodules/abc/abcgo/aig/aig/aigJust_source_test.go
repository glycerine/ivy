package aig

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

func TestJustificationOutputIsTransparentScratch(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	out := g.AddOutput(g.AND(a, b).Not())
	Aig_ManSetCioIds(g)
	g.node(out).MarkB = true
	before := *g.node(out)
	var support []int32
	if !Aig_ObjFindSatAssign(g, out, 0, &support) {
		t.Fatal("satisfiable output was not justified")
	}
	if *g.node(out) != before {
		t.Error("source wrapper bypasses COs, but justification changed output scratch")
	}
}

func TestJustificationFailureRetainsLastPartialSupport(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(g.AND(a, b), g.AND(a.Not(), c))
	Aig_ManSetCioIds(g)
	var support []int32
	if Aig_ObjFindSatAssign(g, root, 1, &support) {
		t.Fatal("contradictory cone was justified")
	}
	if !slices.Equal(support, []int32{0, 2}) {
		t.Fatalf("last partial support = %v, want source [0 2]", support)
	}
}

func TestJustificationUsesSharedTraversalRollover(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	g.travID = (1 << 30) - 1
	var support []int32
	if !Aig_ObjFindSatAssign(g, a, 1, &support) || g.travID != 2 {
		t.Fatalf("justification bypassed shared rollover: generation %d, support %v", g.travID, support)
	}
}

func TestJustificationRejectsInvalidWitnessInputIndex(t *testing.T) {
	for _, index := range []int32{-1, 0, 7} {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			g := NewGraph()
			g.NewInput()
			b := g.NewInput()
			g.node(b).CioID = index
			defer func() {
				if recover() == nil {
					t.Error("invalid CI index was accepted as a support witness")
				}
			}()
			var support []int32
			Aig_ObjFindSatAssign(g, b, 1, &support)
		})
	}
}

func TestJustificationAgainstCompiledSource(t *testing.T) {
	rng := rand.New(rand.NewPCG(211, 239))
	var input, want strings.Builder
	var labels []string
	for trial := 0; trial < 10; trial++ {
		g := NewGraph()
		pool := []Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		if trial == 0 {
			pool = append(pool, g.AND(g.AND(pool[0], pool[1]), g.AND(pool[0].Not(), pool[2])))
		}
		for i := 0; i < 12; i++ {
			a, b := pool[rng.IntN(len(pool))], pool[rng.IntN(len(pool))]
			if rng.IntN(2) != 0 {
				a = a.Not()
			}
			if rng.IntN(2) != 0 {
				b = b.Not()
			}
			pool = append(pool, g.AND(a, b))
		}
		beforeOutputs := len(g.nodes)
		outputIndex := make(map[int]int)
		for id := 1; id < beforeOutputs; id++ {
			out := g.AddOutput(Edge(id) << 1)
			outputIndex[out.NodeID()] = len(g.outputs) - 1
		}
		g.AddOutput((Edge(beforeOutputs-1) << 1).Not())
		outputIndex[g.outputs[len(g.outputs)-1].NodeID()] = len(g.outputs) - 1
		g.AddOutput(g.Const0())
		outputIndex[g.outputs[len(g.outputs)-1].NodeID()] = len(g.outputs) - 1
		Aig_ManSetCioIds(g)
		var queries strings.Builder
		queryCount := 0
		var support []int32
		apply := func(mode int, e Edge, value, heur int32, literals []int32) {
			fmt.Fprintf(&queries, "%d %d %d %d %d", mode, e.NodeID(), value, heur, len(literals))
			for _, lit := range literals {
				fmt.Fprintf(&queries, " %d", lit)
			}
			queries.WriteByte('\n')
			queryCount++
			labels = append(labels, fmt.Sprintf("graph %d mode %d object %d value %d heuristic %d support %v", trial, mode, e.NodeID(), value, heur, literals))
			var result int32
			switch mode {
			case 0:
				if Aig_ObjFindSatAssign(g, e, value, &support) {
					result = 1
				}
			case 1:
				support = support[:0]
				Aig_ManIncrementTravId(g)
				if Aig_NtkFindSatAssign_rec(g, e, value, &support, heur) {
					result = 1
				}
			case 2:
				support = append(support[:0], literals...)
				result = Aig_ObjTerSimulate(g, e, support)
			}
			fmt.Fprintf(&want, "%d %d %d:", mode, result, g.travID)
			for _, lit := range support {
				fmt.Fprintf(&want, " %d", lit)
			}
			want.WriteString(" | ")
			for id, n := range g.nodes {
				fmt.Fprintf(&want, "%d/%d ", Aig_ObjGetTerValue(g, Edge(id)<<1), n.TravID)
			}
			want.WriteByte('\n')
			if out, ok := outputIndex[e.NodeID()]; ok && (mode == 0 && result == 1 || mode == 2 && result != AIG_VALX) {
				target := value != 0
				if mode == 2 {
					target = result == AIG_VAL1
				}
				consistent := 0
				for mask := 0; mask < 16; mask++ {
					assignment := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
					matches := true
					for _, lit := range support {
						matches = matches && assignment[lit>>1] == (lit&1 == 0)
					}
					if matches {
						consistent++
						values, err := g.Evaluate(assignment)
						if err != nil || values[out] != target {
							t.Fatalf("invalid witness/known ternary value: %s assignment %d: %v (%v)", labels[len(labels)-1], mask, values, err)
						}
					}
				}
				if consistent == 0 {
					t.Fatalf("contradictory support reported a known value: %s", labels[len(labels)-1])
				}
			}
		}
		apply(0, g.Const1(), 1, 0, nil)
		for id := 1; id < len(g.nodes); id++ {
			for target := int32(0); target < 2; target++ {
				if g.nodes[id].Kind == Output {
					driver := g.nodes[id].Fanins[0]
					if driver.NodeID() == 0 && (target != 0) == driver.Complemented() {
						continue // C incorrectly accepts an impossible constant target.
					}
				}
				apply(0, Edge(id)<<1, target, 0, nil)
				if g.nodes[id].Kind == And {
					for heur := int32(0); heur < 8; heur++ {
						apply(1, Edge(id)<<1, target, heur, nil)
					}
				}
			}
		}
		for code := 0; code < 81; code++ {
			var literals []int32
			for index, ternary := 0, code; index < 4; index, ternary = index+1, ternary/3 {
				if digit := ternary % 3; digit != 0 {
					lit := int32(2 * index)
					if digit == 1 {
						lit |= 1
					}
					literals = append(literals, lit)
				}
			}
			apply(2, g.outputs[beforeOutputs-2], 0, 0, literals)
			apply(2, g.outputs[len(g.outputs)-2], 0, 0, literals)
			apply(0, g.outputs[beforeOutputs-2], int32(code&1), 0, nil)
		}
		fmt.Fprintf(&input, "%d %d %d %d\n", len(g.inputs), beforeOutputs-1-len(g.inputs), len(g.outputs), queryCount)
		for id := 1 + len(g.inputs); id < beforeOutputs; id++ {
			fmt.Fprintf(&input, "%d %d\n", g.nodes[id].Fanins[0], g.nodes[id].Fanins[1])
		}
		for _, e := range g.outputs {
			fmt.Fprintf(&input, "%d ", g.outputDriver(e))
		}
		input.WriteByte('\n')
		input.WriteString(queries.String())
	}
	got := runAigSourceFixtureInput(t, "justification_compare.c", input.String(), "aigUtil.c")
	gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want.String(), "\n")
	for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
		if gotLines[i] != wantLines[i] {
			t.Fatalf("%s: C %q, Go %q", labels[i], gotLines[i], wantLines[i])
		}
	}
	if len(gotLines) != len(wantLines) {
		t.Fatalf("C trace length %d, Go %d", len(gotLines), len(wantLines))
	}
	t.Logf("compared %d justification/ternary queries", len(labels))
}
