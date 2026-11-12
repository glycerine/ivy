package cgt

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"reflect"
	"testing"
)

func gatingFixture() *aig.Graph {
	g := aig.NewGraph()
	enable, data, q := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(q)
	g.AddOutput(g.Mux(enable, data, q))
	aig.Aig_ManSetRegNum(g, 1)
	return g
}
func checkTransitionEquivalent(t *testing.T, left, right *aig.Graph, care *aig.Graph) {
	t.Helper()
	for mask := 0; mask < 1<<uint(left.InputCount()); mask++ {
		input := make([]bool, left.InputCount())
		for i := range input {
			input[i] = mask&(1<<uint(i)) != 0
		}
		if care != nil {
			v, err := care.Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
			allowed := true
			for _, x := range v {
				allowed = allowed && x
			}
			if !allowed {
				continue
			}
		}
		a, err := left.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		b, err := right.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, b) {
			t.Fatal("transition changed", input, a, b)
		}
	}
}
func TestClockGatingFindsEnableAndPreservesTransitions(t *testing.T) {
	g := gatingFixture()
	g.Name = "enable"
	g.FlopNumbers = []int32{42}
	pars := DefaultParams()
	pars.ConflictMax = 0
	pars.FlopsMin = 1
	pars.VarsMin = 1
	all, err := Cgt_ClockGatingCandidates(g, nil, &pars, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || len(all[0]) == 0 {
		t.Fatal("no gating candidates")
	}
	found := false
	for _, gate := range all[0] {
		found = found || gate == g.Inputs()[0].Not()
	}
	if !found {
		t.Fatal("missing !enable hold condition", all)
	}
	for _, area := range []bool{false, true} {
		pars.AreaOnly = area
		out, err := Cgt_ClockGating(g, nil, &pars)
		if err != nil {
			t.Fatal(err)
		}
		if err = out.Check(); err != nil {
			t.Fatal(err)
		}
		checkTransitionEquivalent(t, g, out, nil)
		if out.Name != g.Name || !reflect.DeepEqual(out.FlopNumbers, g.FlopNumbers) {
			t.Fatal("lost metadata")
		}
		out.FlopNumbers[0] = 0
		if g.FlopNumbers[0] != 42 {
			t.Fatal("aliased metadata")
		}
	}
}
func TestClockGatingCandidatesAgainstExhaustiveOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(119))
	for trial := 0; trial < 30; trial++ {
		g := aig.NewGraph()
		pool := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for n := 0; n < 7; n++ {
			a, b := pool[rng.Intn(len(pool))], pool[rng.Intn(len(pool))]
			if rng.Intn(2) != 0 {
				a = a.Not()
			}
			if rng.Intn(2) != 0 {
				b = b.Not()
			}
			if n%3 == 0 {
				pool = append(pool, g.NewXor(a, b))
			} else {
				pool = append(pool, g.AND(a, b))
			}
		}
		g.AddOutput(pool[len(pool)-1])
		g.AddOutput(pool[len(pool)-2])
		g.AddOutput(pool[len(pool)-3])
		aig.Aig_ManSetRegNum(g, 2)
		pars := DefaultParams()
		pars.ConflictMax = 0
		pars.FlopsMin = 1
		pars.VarsMin = 1
		all, err := Cgt_ClockGatingCandidates(g, nil, &pars, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, pis, pos := layout(g)
		for r, set := range all {
			for _, gate := range set {
				for mask := 0; mask < 16; mask++ {
					in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
					values := evaluateNodes(g, in)
					enabled := values[gate.NodeID()] != (gate&1 != 0)
					out, _ := g.Evaluate(in)
					if enabled && out[pos+r] != in[pis+r] {
						t.Fatalf("unsound gate trial%d reg%d lit%d input%v", trial, r, gate, in)
					}
				}
			}
		}
		for _, area := range []bool{false, true} {
			pars.AreaOnly = area
			out, err := Cgt_ClockGating(g, nil, &pars)
			if err != nil {
				t.Fatal(err)
			}
			checkTransitionEquivalent(t, g, out, nil)
		}
	}
}
func TestClockGatingCareAndValidation(t *testing.T) {
	g := aig.NewGraph()
	a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
	next := g.AND(a, b)
	g.AddOutput(q)
	g.AddOutput(next)
	aig.Aig_ManSetRegNum(g, 1)
	care := aig.NewGraph()
	ca, cb, cq := care.NewInput(), care.NewInput(), care.NewInput()
	care.AddOutput(care.XOR(care.AND(ca, cb), cq).Not())
	pars := DefaultParams()
	pars.ConflictMax = 0
	out, err := Cgt_ClockGating(g, care, &pars)
	if err != nil {
		t.Fatal(err)
	}
	checkTransitionEquivalent(t, g, out, care)
	bad := aig.NewGraph()
	bad.NewInput()
	bad.AddOutput(bad.Const1())
	if _, err = Cgt_ClockGating(g, bad, &pars); err == nil {
		t.Fatal("mismatched care accepted")
	}
	pars.FlopsMin = 0
	if _, err = Cgt_ClockGating(g, nil, &pars); err == nil {
		t.Fatal("zero partition stride accepted")
	}
}
func TestClockGatingODCUnobservableState(t *testing.T) {
	g := aig.NewGraph()
	a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(a)
	g.AddOutput(g.AND(a, b))
	aig.Aig_ManSetRegNum(g, 1)
	_ = q
	pars := DefaultParams()
	pars.ConflictMax = 0
	pars.ODCMax = 2
	out, err := Cgt_ClockGating(g, nil, &pars)
	if err != nil {
		t.Fatal(err)
	}
	rng := rand.New(rand.NewSource(5))
	left, right := false, false
	for frame := 0; frame < 100; frame++ {
		a, b := rng.Intn(2) != 0, rng.Intn(2) != 0
		x, _ := g.Evaluate([]bool{a, b, left})
		y, _ := out.Evaluate([]bool{a, b, right})
		if x[0] != y[0] {
			t.Fatal("ODC changed observable trace")
		}
		left, right = x[1], y[1]
	}
}
func TestClockGatingPartitionsAndSimulationPatterns(t *testing.T) {
	g := gatingFixture()
	pars := DefaultParams()
	pars.ConflictMax = 0
	p, err := Cgt_ManCreate(g, nil, &pars)
	if err != nil {
		t.Fatal(err)
	}
	Cgt_ManDeriveAigForGating(p)
	next := Cgt_ClockGatingRange(p, 0)
	if next != 1 || p.Stats.Calls == 0 || p.Stats.UNSAT == 0 {
		t.Fatal("range not proved", p.Stats)
	}
	if p.Stats.Patterns == 0 || len(p.patterns) == 0 {
		t.Fatal("SAT witnesses not recorded")
	}
	Cgt_ManClean(p)
	if p.Part != nil || p.solver != nil || p.patterns != nil {
		t.Fatal("partition not cleaned")
	}
	Cgt_ManStop(p)
}
func TestClockGatingRandomODCTrace(t *testing.T) {
	rng := rand.New(rand.NewSource(1011))
	for trial := 0; trial < 25; trial++ {
		g := aig.NewGraph()
		pool := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for n := 0; n < 8; n++ {
			a, b := pool[rng.Intn(len(pool))], pool[rng.Intn(len(pool))]
			if rng.Intn(2) != 0 {
				a = a.Not()
			}
			pool = append(pool, g.AND(a, b))
		}
		g.AddOutput(pool[len(pool)-1])
		g.AddOutput(pool[len(pool)-2])
		g.AddOutput(pool[len(pool)-3])
		aig.Aig_ManSetRegNum(g, 2)
		pars := DefaultParams()
		pars.ConflictMax = 0
		pars.ODCMax = 1 + trial%3
		out, err := Cgt_ClockGating(g, nil, &pars)
		if err != nil {
			t.Fatal(err)
		}
		if err = out.Check(); err != nil {
			t.Fatal(err)
		}
		left, right := []bool{false, false}, []bool{false, false}
		for frame := 0; frame < 60; frame++ {
			a, b := rng.Intn(2) != 0, rng.Intn(2) != 0
			x, err := g.Evaluate([]bool{a, b, left[0], left[1]})
			if err != nil {
				t.Fatal(err)
			}
			y, err := out.Evaluate([]bool{a, b, right[0], right[1]})
			if err != nil {
				t.Fatal(err)
			}
			if x[0] != y[0] {
				t.Fatal("ODC random observable mismatch", trial, frame)
			}
			left, right = x[1:], y[1:]
		}
	}
}
func TestClockGatingAreaSelectionKeepsGatePhase(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.NewInput()
	g.NewInput()
	gate := g.AND(a, b)
	g.AddOutput(g.Const0())
	g.AddOutput(gate)
	g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 2)
	all := Gates{{gate, gate.Not()}, {gate}}
	selected := Cgt_ManDecideArea(g, all, 0, false)
	if len(selected[0]) != 1 || selected[0][0] != gate || len(selected[1]) != 1 || selected[1][0] != gate {
		t.Fatal("area selection conflated gate phases", selected)
	}
}
func TestClockGatingNilDefaultsAndPartitionBounds(t *testing.T) {
	g := gatingFixture()
	out, err := Cgt_ClockGating(g, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	checkTransitionEquivalent(t, g, out, nil)
	p, err := Cgt_ManCreate(g, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	frame := Cgt_ManDeriveAigForGating(p)
	for _, start := range []int{-1, frame.OutputCount()} {
		if _, _, err := Cgt_ManDupPartition(frame, 1, 1, start, nil); err == nil {
			t.Fatal("invalid partition accepted")
		}
	}
}
func TestClockGateBuilderRejectsMissingCopy(t *testing.T) {
	g := aig.NewGraph()
	edge := g.NewInput()
	defer func() {
		if recover() == nil {
			t.Error("missing gate copy silently became constant one")
		}
	}()
	Cgt_ManBuildClockGate(aig.NewGraph(), g, []aig.Edge{edge}, map[int]aig.Edge{}, nil)
}
