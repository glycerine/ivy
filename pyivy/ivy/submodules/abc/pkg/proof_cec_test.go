package abc

// Standalone Go tests for src/proof/cec semantics.

import (
	"errors"
	"math/rand"
	"testing"
)

func cecGraph(t testing.TB, inputs int, function func(*AigManager, []AigLit) AigLit) *AigManager {
	t.Helper()
	graph := NewAigManager()
	vars := make([]AigLit, inputs)
	for i := range vars {
		var err error
		vars[i], err = graph.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := graph.CreateCO(function(graph, vars)); err != nil {
		t.Fatal(err)
	}
	return graph
}

func cecAnd(t testing.TB, graph *AigManager, left, right AigLit) AigLit {
	t.Helper()
	result, err := graph.And(left, right)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func cecTruth2(t testing.TB, truth uint8) *AigManager {
	return cecGraph(t, 2, func(graph *AigManager, variables []AigLit) AigLit {
		result := AigConst0
		for pattern := 0; pattern < 4; pattern++ {
			if truth&(1<<uint(pattern)) == 0 {
				continue
			}
			term := AigConst1
			for input := 0; input < 2; input++ {
				term = cecAnd(t, graph, term, variables[input].NotIf(pattern&(1<<uint(input)) == 0))
			}
			var err error
			result, err = graph.Or(result, term)
			if err != nil {
				t.Fatal(err)
			}
		}
		return result
	})
}

func TestCECExhaustiveTwoInputFunctions(t *testing.T) {
	graphs := make([]*AigManager, 16)
	for truth := range graphs {
		graphs[truth] = cecTruth2(t, uint8(truth))
	}
	for left := range graphs {
		for right := range graphs {
			result, err := CompareAIGCombinational(graphs[left], graphs[right], CECOptions{})
			if err != nil {
				t.Fatalf("%x/%x: %v", left, right, err)
			}
			if (result.Status == CECEquivalent) != (left == right) {
				t.Fatalf("%x/%x: %+v", left, right, result)
			}
			if result.Status == CECDifferent {
				if err := ReplayCECCombinational(graphs[left], graphs[right], result.Counter, CECOptions{}); err != nil {
					t.Fatalf("%x/%x replay: %v", left, right, err)
				}
			}
		}
	}
}

func TestCECCombinationalExactAndReplay(t *testing.T) {
	left := cecGraph(t, 2, func(g *AigManager, v []AigLit) AigLit { return cecAnd(t, g, v[0], v[1]) })
	right := cecGraph(t, 2, func(g *AigManager, v []AigLit) AigLit { return cecAnd(t, g, v[1], v[0]) })
	result, err := CompareAIGCombinational(left, right, CECOptions{})
	if err != nil || result.Status != CECEquivalent || result.Patterns != 4 {
		t.Fatalf("equivalent: %+v %v", result, err)
	}
	different := cecGraph(t, 2, func(g *AigManager, v []AigLit) AigLit { return v[0] })
	result, err = CompareAIGCombinational(left, different, CECOptions{})
	if err != nil || result.Status != CECDifferent || result.Counter == nil {
		t.Fatalf("different: %+v %v", result, err)
	}
	if err := ReplayCECCombinational(left, different, result.Counter, CECOptions{}); err != nil {
		t.Fatal(err)
	}
	bad := *result.Counter
	bad.LeftValue = !bad.LeftValue
	if err := ReplayCECCombinational(left, different, &bad, CECOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("bad replay: %v", err)
	}
}

func cecSequentialGraph(t testing.TB, invertOutput bool) *AigManager {
	t.Helper()
	g := NewAigManager()
	pi, err := g.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	ro, err := g.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	driver := ro.NotIf(invertOutput)
	if _, err := g.CreateCO(driver); err != nil {
		t.Fatal(err)
	}
	next, err := g.Xor(pi, ro)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.CreateCO(next); err != nil {
		t.Fatal(err)
	}
	if err := g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestCECSequentialShortestAndReplay(t *testing.T) {
	left := cecSequentialGraph(t, false)
	right := cecSequentialGraph(t, false)
	result, err := CompareAIGSequential(left, right, []bool{false}, []bool{false}, CECOptions{MaxFrames: 4})
	if err != nil || result.Status != CECEquivalent {
		t.Fatalf("sequential equivalent: %+v %v", result, err)
	}
	different := cecSequentialGraph(t, true)
	result, err = CompareAIGSequential(left, different, []bool{false}, []bool{false}, CECOptions{MaxFrames: 4})
	if err != nil || result.Status != CECDifferent || result.Counter == nil || result.Counter.Frame != 0 {
		t.Fatalf("sequential different: %+v %v", result, err)
	}
	if err := ReplayCECSequential(left, different, result.Counter, CECOptions{}); err != nil {
		t.Fatal(err)
	}
	extra := *result.Counter
	extra.Inputs = cecCloneFrames(result.Counter.Inputs)
	extra.Inputs = append(extra.Inputs, []bool{false})
	if err := ReplayCECSequential(left, different, &extra, CECOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("trailing witness frame accepted: %v", err)
	}
}

func TestCECOutputClasses(t *testing.T) {
	g := NewAigManager()
	a, _ := g.CreateCI()
	b, _ := g.CreateCI()
	and := cecAnd(t, g, a, b)
	for _, lit := range []AigLit{and, and.Not(), a, a} {
		if _, err := g.CreateCO(lit); err != nil {
			t.Fatal(err)
		}
	}
	classes, err := CECOutputClasses(g, CECOptions{})
	if err != nil || len(classes) != 2 {
		t.Fatalf("classes: %+v %v", classes, err)
	}
	if len(classes[0].Members) != 2 || !classes[0].Complemented[1] || len(classes[1].Members) != 2 || classes[1].Complemented[1] {
		t.Fatalf("class contents: %+v", classes)
	}
	for _, class := range classes {
		if len(class.Members) != len(class.Complemented) {
			t.Fatalf("misaligned class metadata: %+v", class)
		}
	}
	classSetupWork := g.ObjectCount() + g.InputCount() + g.OutputCount() + 4*(g.ObjectCount()+g.OutputCount())
	if _, err := CECOutputClasses(g, CECOptions{MaxOperations: classSetupWork + 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("class comparison work limit: %v", err)
	}
	if _, err := CECOutputClasses(g, CECOptions{MaxBytes: 40}); !errors.Is(err, ErrLimit) {
		t.Fatalf("class result allocation limit: %v", err)
	}
	stopCalls := 0
	if _, err := CECOutputClasses(g, CECOptions{Stop: func() bool {
		stopCalls++
		return stopCalls == 6
	}}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("class comparison cancellation: %v", err)
	}
}

func TestCECLimitsCancellationDimensionsAndInventory(t *testing.T) {
	g := cecGraph(t, 3, func(g *AigManager, v []AigLit) AigLit { return v[0] })
	if _, err := CompareAIGCombinational(g, g, CECOptions{MaxPatterns: 7}); !errors.Is(err, ErrLimit) {
		t.Fatalf("pattern limit: %v", err)
	}
	if _, err := CompareAIGCombinational(g, g, CECOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %v", err)
	}
	if _, err := CompareAIGCombinational(nil, g, CECOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("nil: %v", err)
	}
	other := cecGraph(t, 2, func(g *AigManager, v []AigLit) AigLit { return v[0] })
	if _, err := CompareAIGCombinational(g, other, CECOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("dimension: %v", err)
	}
	for name, options := range map[string]CECOptions{
		"work":    {MaxOperations: 1},
		"bytes":   {MaxBytes: 1},
		"nodes":   {MaxNodes: g.ObjectCount()},
		"clauses": {MaxClauses: 1},
	} {
		if _, err := CompareAIGCombinational(g, g, options); !errors.Is(err, ErrLimit) {
			t.Errorf("%s limit: %v", name, err)
		}
	}
	deep := cecGraph(t, 3, func(graph *AigManager, values []AigLit) AigLit {
		return cecAnd(t, graph, cecAnd(t, graph, values[0], values[1]), values[2])
	})
	if _, err := CompareAIGCombinational(deep, deep, CECOptions{MaxDepth: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("depth limit: %v", err)
	}
	multi := NewAigManager()
	input, _ := multi.CreateCI()
	_, _ = multi.CreateCO(input)
	_, _ = multi.CreateCO(input.Not())
	if _, err := CompareAIGCombinational(multi, multi, CECOptions{MaxPartitions: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("partition limit: %v", err)
	}
	if len(CECBuildInventory) != 18 || CECBuildInventory[0] != "cecCec.c" || CECBuildInventory[17] != "cecSweep.c" {
		t.Fatalf("inventory: %v", CECBuildInventory)
	}
	if len(CECSourceInventory) != 23 || CECSourceInventory[0] != "cec.c" || CECSourceInventory[22] != "module.make" {
		t.Fatalf("source inventory: %v", CECSourceInventory)
	}
}

func TestCECOutputAlignmentMetricsBoundedUnknownAndRollback(t *testing.T) {
	left := NewAigManager()
	a, _ := left.CreateCI()
	for _, output := range []AigLit{a, a.Not()} {
		if _, err := left.CreateCO(output); err != nil {
			t.Fatal(err)
		}
	}
	right := NewAigManager()
	b, _ := right.CreateCI()
	for _, output := range []AigLit{b.Not(), b} {
		if _, err := right.CreateCO(output); err != nil {
			t.Fatal(err)
		}
	}
	beforeLeft, beforeRight := left.ObjectCount(), right.ObjectCount()
	result, err := CompareAIGCombinational(left, right, CECOptions{})
	if err != nil || result.Status != CECDifferent || result.Counter.Output != 0 {
		t.Fatalf("aligned outputs: %+v %v", result, err)
	}
	if result.Nodes == 0 || result.Clauses == 0 || result.Partitions != 2 || result.Work == 0 || result.Bytes == 0 {
		t.Fatalf("metrics: %+v", result)
	}
	miter, err := EvaluateCECCombinationalMiter(left, right, []bool{false}, CECOptions{})
	if err != nil || len(miter) != 2 || !miter[0] || !miter[1] {
		t.Fatalf("miter: %v %v", miter, err)
	}
	miter[0] = false
	again, err := EvaluateCECCombinationalMiter(left, right, []bool{false}, CECOptions{})
	if err != nil || !again[0] {
		t.Fatalf("miter ownership: %v %v", again, err)
	}
	if _, err := EvaluateCECCombinationalMiter(left, right, []bool{false}, CECOptions{MaxBytes: 3}); !errors.Is(err, ErrLimit) {
		t.Fatalf("miter result allocation limit: %v", err)
	}
	if left.ObjectCount() != beforeLeft || right.ObjectCount() != beforeRight {
		t.Fatal("comparison mutated caller graphs")
	}

	seq := cecSequentialGraph(t, false)
	bounded, err := CompareAIGSequential(seq, seq, []bool{false}, []bool{false}, CECOptions{MaxFrames: 1})
	if err != nil || bounded.Status != CECUnknown {
		t.Fatalf("bounded status: %+v %v", bounded, err)
	}
	if _, err := CompareAIGSequential(seq, seq, []bool{false}, []bool{false}, CECOptions{MaxStates: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("state limit: %v", err)
	}
	corrupt := cecGraph(t, 1, func(graph *AigManager, values []AigLit) AigLit { return values[0] })
	corrupt.objects[corrupt.cos[0]].Fanin0 = InvalidAigLit
	if _, err := CompareAIGCombinational(corrupt, corrupt, CECOptions{}); err == nil {
		t.Fatal("malformed topology accepted")
	}
	badName := cecGraph(t, 1, func(graph *AigManager, values []AigLit) AigLit { return values[0] })
	badName.Name = "bad\x00name"
	if _, err := CompareAIGCombinational(badName, badName, CECOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("malformed name: %v", err)
	}
}

func TestCECIndependentRandomModel5000(t *testing.T) {
	rng := rand.New(rand.NewSource(0xCEC))
	for trial := 0; trial < 5000; trial++ {
		inputs := 1 + rng.Intn(4)
		leftIndex := rng.Intn(inputs)
		rightIndex := rng.Intn(inputs)
		leftNeg := rng.Intn(2) != 0
		rightNeg := rng.Intn(2) != 0
		left := cecGraph(t, inputs, func(g *AigManager, v []AigLit) AigLit { return v[leftIndex].NotIf(leftNeg) })
		right := cecGraph(t, inputs, func(g *AigManager, v []AigLit) AigLit { return v[rightIndex].NotIf(rightNeg) })
		result, err := CompareAIGCombinational(left, right, CECOptions{})
		if err != nil {
			t.Fatal(err)
		}
		equivalent := leftIndex == rightIndex && leftNeg == rightNeg
		if (result.Status == CECEquivalent) != equivalent {
			t.Fatalf("trial %d: %+v", trial, result)
		}
		if !equivalent {
			if err := ReplayCECCombinational(left, right, result.Counter, CECOptions{}); err != nil {
				t.Fatalf("trial %d replay: %v", trial, err)
			}
		}
	}
}

func FuzzCECEquivalence(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x55, 0xaa, 0xffff} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		inputs := 1 + int(seed%5)
		leftIndex := int((seed >> 8) % uint64(inputs))
		rightIndex := int((seed >> 16) % uint64(inputs))
		leftNeg := seed&(1<<24) != 0
		rightNeg := seed&(1<<25) != 0
		left := cecGraph(t, inputs, func(g *AigManager, v []AigLit) AigLit { return v[leftIndex].NotIf(leftNeg) })
		right := cecGraph(t, inputs, func(g *AigManager, v []AigLit) AigLit { return v[rightIndex].NotIf(rightNeg) })
		result, err := CompareAIGCombinational(left, right, CECOptions{MaxPatterns: 64})
		if err != nil {
			t.Fatal(err)
		}
		if result.Status == CECDifferent {
			if err := ReplayCECCombinational(left, right, result.Counter, CECOptions{}); err != nil {
				t.Fatal(err)
			}
		}
	})
}
