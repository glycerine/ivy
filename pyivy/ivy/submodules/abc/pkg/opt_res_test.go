package abc

// Go-only tests for src/opt/res.

import (
	"errors"
	"slices"
	"testing"
)

type resTestGraph struct {
	graph       *AigManager
	base, root  AigLit
	inputPhases [2]bool
	inputIDs    [2]int
}

func resBuildTestGraph(t testing.TB, seed uint64) resTestGraph {
	t.Helper()
	graph := NewAigManager()
	inputs := make([]AigLit, 3)
	for i := range inputs {
		var err error
		inputs[i], err = graph.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	first := int(seed % 3)
	second := int((seed / 3) % 2)
	if second >= first {
		second++
	}
	phase0, phase1 := seed&(1<<4) != 0, seed&(1<<5) != 0
	a, b := inputs[first].NotIf(phase0), inputs[second].NotIf(phase1)
	base, err := graph.And(a, b)
	if err != nil {
		t.Fatal(err)
	}
	// This is logically base but deliberately not structurally absorbed.
	root, err := graph.And(base, a)
	if err != nil {
		t.Fatal(err)
	}
	if root == base {
		t.Fatal("fixture was structurally absorbed")
	}
	if _, err = graph.CreateCO(base); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(root); err != nil {
		t.Fatal(err)
	}
	if err = graph.Check(); err != nil {
		t.Fatal(err)
	}
	return resTestGraph{graph: graph, base: base, root: root, inputPhases: [2]bool{phase0, phase1}, inputIDs: [2]int{first, second}}
}

func resExpected(fixture resTestGraph, assignment int) bool {
	first := assignment&(1<<fixture.inputIDs[0]) != 0
	second := assignment&(1<<fixture.inputIDs[1]) != 0
	if fixture.inputPhases[0] {
		first = !first
	}
	if fixture.inputPhases[1] {
		second = !second
	}
	return first && second
}

func resCheckTruth(t testing.TB, fixture resTestGraph, graph *AigManager) {
	t.Helper()
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		outputs, err := graph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		want := resExpected(fixture, assignment)
		if len(outputs) != 2 || outputs[0] != want || outputs[1] != want {
			t.Fatalf("assignment=%03b outputs=%v want=%v", assignment, outputs, want)
		}
	}
}

func TestResynthesisWindowSimulationStrashAndFilter(t *testing.T) {
	fixture := resBuildTestGraph(t, 19)
	parameters := ResynthesisParameters{Window: 21, SimulationWords: 1, Candidates: 32, Area: true}
	window, err := ComputeResynthesisWindow(fixture.graph, fixture.root.ID(), parameters, ResynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if window.Pivot != fixture.root.ID() || len(window.Leaves) == 0 || len(window.Nodes) == 0 || len(window.Roots) == 0 {
		t.Fatalf("incomplete window: %+v", window)
	}
	inputWords := []uint64{truthVariablePattern(0), truthVariablePattern(1), truthVariablePattern(2)}
	simulation, err := SimulateResynthesis(fixture.graph, 1, inputWords, ResynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if simulation.Patterns != 64 || simulation.Values[fixture.base.ID()][0]&0xff == 0 {
		t.Fatalf("bad simulation: %+v", simulation)
	}
	candidates, err := FilterResynthesisCandidates(fixture.graph, window, simulation, 32, ResynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, candidate := range candidates {
		if candidate.Kind == ResynthesisLiteral && len(candidate.Inputs) == 1 && candidate.Inputs[0] == fixture.base {
			found = true
		}
	}
	if !found {
		t.Fatalf("base divisor was filtered: %+v", candidates)
	}
	strashed, outputs, err := StrashResynthesisWindow(fixture.graph, window, ResynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = strashed.Check(); err != nil {
		t.Fatal(err)
	}
	if len(outputs) < 4 || strashed.OutputCount() != len(outputs) {
		t.Fatalf("outputs=%d/%d", len(outputs), strashed.OutputCount())
	}
}

func TestResynthesisSATProofAndAtomicRewrite(t *testing.T) {
	fixture := resBuildTestGraph(t, 7)
	beforeObjects, beforeAnds := fixture.graph.ObjectCount(), fixture.graph.AndCount()
	bad := ResynthesisCandidate{Kind: ResynthesisConstant0, Simulated: true}
	if trial, proved, err := ProveResynthesisCandidate(fixture.graph, fixture.root.ID(), bad, ResynthesisLimits{}); err != nil || trial != nil || proved.Proved {
		t.Fatalf("bad candidate accepted: graph=%v candidate=%+v err=%v", trial, proved, err)
	}
	good := ResynthesisCandidate{Kind: ResynthesisLiteral, Inputs: []AigLit{fixture.base}, Simulated: true}
	trial, proved, err := ProveResynthesisCandidate(fixture.graph, fixture.root.ID(), good, ResynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if trial == nil || !proved.Proved || proved.Gain < 1 {
		t.Fatalf("candidate=%+v", proved)
	}
	resCheckTruth(t, fixture, trial)
	if fixture.graph.ObjectCount() != beforeObjects || fixture.graph.AndCount() != beforeAnds {
		t.Fatal("source graph mutated")
	}
	resCheckTruth(t, fixture, fixture.graph)
}

func TestResynthesizeAIGDeterminismOwnershipLimitsAndCancellation(t *testing.T) {
	fixture := resBuildTestGraph(t, 3)
	parameters := ResynthesisParameters{Window: 21, SimulationWords: 2, Candidates: 32, Area: true}
	first, err := ResynthesizeAIG(fixture.graph, parameters, ResynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := ResynthesizeAIG(fixture.graph, parameters, ResynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if first.Graph == fixture.graph || first.Graph.AndCount() >= fixture.graph.AndCount() || first.Stats.Replacements == 0 {
		t.Fatalf("no useful owned rewrite: %+v", first.Stats)
	}
	if first.Stats != second.Stats || len(first.Replacements) != len(second.Replacements) {
		t.Fatalf("nondeterministic stats: %+v %+v", first.Stats, second.Stats)
	}
	for i := range first.Replacements {
		if first.Replacements[i].Kind != second.Replacements[i].Kind || !slices.Equal(first.Replacements[i].Inputs, second.Replacements[i].Inputs) {
			t.Fatal("nondeterministic candidate order")
		}
	}
	resCheckTruth(t, fixture, first.Graph)
	if result, err := ResynthesizeAIG(fixture.graph, parameters, ResynthesisLimits{MaxObjects: 1}); result.Graph != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := ResynthesizeAIG(fixture.graph, parameters, ResynthesisLimits{Stop: func() bool { return true }}); result.Graph != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	if _, err := ResynthesizeAIG(fixture.graph, ResynthesisParameters{Window: 9}, ResynthesisLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := ResynthesizeAIG(fixture.graph, ResynthesisParameters{Window: 21, Candidates: 100}, ResynthesisLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("native candidate bound: %v", err)
	}
	if _, err := SimulateResynthesis(fixture.graph, 1, []uint64{0}, ResynthesisLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := SimulateResynthesis(fixture.graph, 2, nil, ResynthesisLimits{MaxSimulationWords: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	window, _ := ComputeResynthesisWindow(fixture.graph, fixture.root.ID(), parameters, ResynthesisLimits{})
	if _, err := FilterResynthesisCandidates(fixture.graph, window, ResynthesisSimulation{Words: 1}, 1, ResynthesisLimits{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestResynthesisExportedStageLimitValidation(t *testing.T) {
	fixture := resBuildTestGraph(t, 11)
	parameters := ResynthesisParameters{Window: 21, SimulationWords: 1, Candidates: 16, Area: true}
	window, err := ComputeResynthesisWindow(fixture.graph, fixture.root.ID(), parameters, ResynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	simulation, err := SimulateResynthesis(fixture.graph, 1, nil, ResynthesisLimits{})
	if err != nil {
		t.Fatal(err)
	}
	negative := ResynthesisLimits{MaxWork: -1}
	if candidates, filterErr := FilterResynthesisCandidates(fixture.graph, window, simulation, 16, negative); candidates != nil || !errors.Is(filterErr, ErrDimension) {
		t.Fatalf("filter negative limit: candidates=%v err=%v", candidates, filterErr)
	}
	if strashed, outputs, strashErr := StrashResynthesisWindow(fixture.graph, window, negative); strashed != nil || outputs != nil || !errors.Is(strashErr, ErrDimension) {
		t.Fatalf("strash negative limit: graph=%v outputs=%v err=%v", strashed, outputs, strashErr)
	}
	candidate := ResynthesisCandidate{Kind: ResynthesisLiteral, Inputs: []AigLit{fixture.base}, Simulated: true}
	if trial, _, proveErr := ProveResynthesisCandidate(fixture.graph, fixture.root.ID(), candidate, negative); trial != nil || !errors.Is(proveErr, ErrDimension) {
		t.Fatalf("proof negative limit: graph=%v err=%v", trial, proveErr)
	}
	if trial, _, proveErr := ProveResynthesisCandidate(fixture.graph, fixture.root.ID(), candidate, ResynthesisLimits{MaxObjects: 1}); trial != nil || !errors.Is(proveErr, ErrLimit) {
		t.Fatalf("proof object limit: graph=%v err=%v", trial, proveErr)
	}
	if candidates, filterErr := FilterResynthesisCandidates(fixture.graph, window, simulation, 100, ResynthesisLimits{}); candidates != nil || !errors.Is(filterErr, ErrDimension) {
		t.Fatalf("filter native candidate bound: candidates=%v err=%v", candidates, filterErr)
	}
	badWindow := window
	badWindow.Leaves = append(slices.Clone(window.Leaves), AigID(fixture.graph.ObjectCount()+1))
	if strashed, outputs, strashErr := StrashResynthesisWindow(fixture.graph, badWindow, ResynthesisLimits{}); strashed != nil || outputs != nil || !errors.Is(strashErr, ErrInvalidLiteral) {
		t.Fatalf("strash invalid boundary: graph=%v outputs=%v err=%v", strashed, outputs, strashErr)
	}
	if simulation, simulationErr := SimulateResynthesis(fixture.graph, 2, nil, ResynthesisLimits{MaxWork: fixture.graph.ObjectCount()}); simulation.Values != nil || !errors.Is(simulationErr, ErrLimit) {
		t.Fatalf("simulation allocation work bound: simulation=%+v err=%v", simulation, simulationErr)
	}
}

func TestResynthesisIndependentTruthCampaign5000(t *testing.T) {
	for execution := 0; execution < 5000; execution++ {
		fixture := resBuildTestGraph(t, uint64(execution)*0x9e3779b97f4a7c15)
		result, err := ResynthesizeAIG(fixture.graph, ResynthesisParameters{Window: 21, SimulationWords: 1, Candidates: 16, Area: true}, ResynthesisLimits{})
		if err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
		if result.Stats.Replacements == 0 || result.Graph.AndCount() >= fixture.graph.AndCount() {
			t.Fatalf("execution %d: %+v", execution, result.Stats)
		}
		resCheckTruth(t, fixture, result.Graph)
	}
}

func FuzzResynthesizeAIG(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(19))
	f.Add(^uint64(0))
	f.Fuzz(func(t *testing.T, seed uint64) {
		fixture := resBuildTestGraph(t, seed)
		result, err := ResynthesizeAIG(fixture.graph, ResynthesisParameters{Window: 21, SimulationWords: 1 + int(seed&1), Candidates: 16, Area: true}, ResynthesisLimits{})
		if err != nil {
			t.Fatal(err)
		}
		if result.Graph == fixture.graph || result.Stats.Replacements == 0 {
			t.Fatalf("%+v", result.Stats)
		}
		resCheckTruth(t, fixture, result.Graph)
		resCheckTruth(t, fixture, fixture.graph)
	})
}
