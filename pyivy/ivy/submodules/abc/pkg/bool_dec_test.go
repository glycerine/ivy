package abc

// Standalone tests for src/bool/dec.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func decTestEvaluateAll(t *testing.T, sop SOPCover, graph *DecoGraph) {
	t.Helper()
	for assignment := 0; assignment < 1<<uint(sop.Variables); assignment++ {
		inputs := make([]bool, sop.Variables)
		for variable := range inputs {
			inputs[variable] = assignment&(1<<uint(variable)) != 0
		}
		expected, err := sop.Evaluate(inputs)
		if err != nil {
			t.Fatal("SOP evaluation", err)
		}
		actual, err := graph.Evaluate(inputs)
		if err != nil {
			t.Fatal("graph evaluation", err)
		}
		if actual != expected {
			t.Fatalf("assignment %d: graph=%v SOP=%v cover=%+v", assignment, actual, expected, sop)
		}
	}
}

func TestDecFactorConstantsManagerAndErrors(t *testing.T) {
	for _, row := range []struct {
		text string
		want bool
	}{{" 0\n", false}, {" 1\n", true}} {
		graph, err := FactorSOPText(row.text, DecFactorLimits{})
		if err != nil {
			t.Fatal(err)
		}
		value, ok := graph.ConstantValue()
		if !ok || value != row.want {
			t.Fatalf("constant %q = %v,%v", row.text, value, ok)
		}
		graph.Close()
	}
	manager, err := StartDecoManager()
	if err != nil {
		t.Fatal(err)
	}
	if manager.MVC == nil || len(manager.Cubes) != 0 || cap(manager.Cubes) != 8 || len(manager.Canons) != 1<<16 || len(manager.Phases) != 1<<16 || len(manager.Permutations) != 1<<16 || len(manager.ClassMap) != 1<<16 {
		t.Fatal("manager layout")
	}
	manager.Canons[1] = 99
	second, err := StartDecoManager()
	if err != nil || second.Canons[1] == 99 {
		t.Fatal("manager ownership", err)
	}
	if err = manager.Stop(); err != nil {
		t.Fatal(err)
	}
	if err = manager.Stop(); !errors.Is(err, ErrInvariant) {
		t.Fatal("repeated stop", err)
	}
	_ = second.Stop()

	malformed := []string{"", "1 1", "1 q\n", "-- 1\n"}
	for _, text := range malformed {
		if graph, factorErr := FactorSOPText(text, DecFactorLimits{}); factorErr == nil || graph != nil {
			t.Fatalf("accepted malformed/invariant SOP %q", text)
		}
	}
	if graph, factorErr := FactorSOP(SOPCover{Variables: 2, Cubes: []string{"10"}, Output: '1'}, DecFactorLimits{MaxVariables: 1}); !errors.Is(factorErr, ErrLimit) || graph != nil {
		t.Fatal("variable limit", factorErr)
	}
	if graph, factorErr := FactorSOP(SOPCover{Variables: 2, Cubes: []string{"10"}, Output: '1'}, DecFactorLimits{Stop: func() bool { return true }}); !errors.Is(factorErr, ErrCancelled) || graph != nil {
		t.Fatal("cancellation", factorErr)
	}
	if graph, factorErr := FactorSOP(SOPCover{Variables: 3, Cubes: []string{"100", "010", "001"}, Output: '1'}, DecFactorLimits{MaxNodes: 3}); !errors.Is(factorErr, ErrLimit) || graph != nil {
		t.Fatal("graph node limit", factorErr)
	}
	wide, _ := NewDecoGraph(6, DecoGraphLimits{})
	wide.root = DecoEdge{Node: 0}
	if _, truthErr := DecoGraphTruth32(wide); !errors.Is(truthErr, ErrDimension) {
		t.Fatal("truth width", truthErr)
	}
	wide.Close()
}

func TestDecManagerStateNPNInventoryExhaustive(t *testing.T) {
	state, err := StartDecoManager()
	if err != nil {
		t.Fatal(err)
	}
	defer state.Stop()
	manager, err := NewDecManager()
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	classes := make(map[byte]bool)
	for truth := 0; truth < 1<<16; truth++ {
		canonical, phase, permutation, class, queryErr := manager.NPN4(uint16(truth))
		if queryErr != nil {
			t.Fatal(queryErr)
		}
		if state.Canons[truth] != canonical || byte(state.Phases[truth]) != phase || byte(state.Permutations[truth]) != permutation || state.ClassMap[truth] != class {
			t.Fatalf("truth=%04x state=(%04x,%d,%d,%d) manager=(%04x,%d,%d,%d)", truth, state.Canons[truth], state.Phases[truth], state.Permutations[truth], state.ClassMap[truth], canonical, phase, permutation, class)
		}
		if phase >= 32 || permutation >= 24 || canonical > uint16(truth) && canonical != 0 {
			t.Fatalf("truth=%04x canonical=%04x phase=%d permutation=%d", truth, canonical, phase, permutation)
		}
		classes[class] = true
	}
	if len(classes) != 222 {
		t.Fatalf("classes=%d", len(classes))
	}
}

func TestDecFactorSourceCasesTruthAndFormatting(t *testing.T) {
	cases := []SOPCover{
		{Variables: 1, Cubes: []string{"1"}, Output: '1'},
		{Variables: 2, Cubes: []string{"11"}, Output: '1'},
		{Variables: 2, Cubes: []string{"1-", "-1"}, Output: '1'},
		{Variables: 2, Cubes: []string{"10", "01"}, Output: '1'},
		{Variables: 3, Cubes: []string{"11-", "1-1", "-11"}, Output: '1'},
		{Variables: 3, Cubes: []string{"00-", "-01"}, Output: '0'},
	}
	for _, sop := range cases {
		graph, err := FactorSOP(sop, DecFactorLimits{})
		if err != nil {
			t.Fatalf("factor %+v: %v", sop, err)
		}
		decTestEvaluateAll(t, sop, graph)
		verified, verifyErr := VerifyDecoFactor(sop, graph, 64)
		if verifyErr != nil || !verified {
			t.Fatalf("factor verification %+v: %v %v", sop, verified, verifyErr)
		}
		if sop.Variables <= 5 {
			truth, truthErr := DecoGraphTruth32(graph)
			if truthErr != nil {
				t.Fatal(truthErr)
			}
			var expected uint32
			for assignment := 0; assignment < 32; assignment++ {
				inputs := make([]bool, sop.Variables)
				for variable := range inputs {
					inputs[variable] = assignment&(1<<uint(variable)) != 0
				}
				value, _ := sop.Evaluate(inputs)
				if value {
					expected |= uint32(1) << uint(assignment)
				}
			}
			if truth != expected {
				t.Fatalf("truth got %08x want %08x", truth, expected)
			}
		}
		graph.Close()
	}
	graph6, _ := FactorSOP(SOPCover{Variables: 6, Cubes: []string{"1-----", "-1----"}, Output: '1'}, DecFactorLimits{})
	table6, err := DecoGraphTruthTable(graph6, 2)
	if err != nil || table6.Variables() != 6 {
		t.Fatal("general truth derivation", err)
	}
	if _, err = DecoGraphTruthTable(graph6, -1); !errors.Is(err, ErrDimension) {
		t.Fatal("truth word limit validation", err)
	}
	graph6.Close()

	andGraph, _ := FactorSOP(SOPCover{Variables: 2, Cubes: []string{"11"}, Output: '1'}, DecFactorLimits{})
	formatted, err := FormatDecoGraph(andGraph, []string{"a", "b"}, "Y")
	if err != nil || formatted != "     Y = (ab)\n" {
		t.Fatalf("AND format %q %v", formatted, err)
	}
	original, err := FormatDecoGraphOriginal(andGraph, []string{"a", "b"}, "Y")
	if err != nil || original != "     Y = a b\n" {
		t.Fatalf("original AND format %q %v", original, err)
	}
	andGraph.Close()
	orGraph, _ := FactorSOP(SOPCover{Variables: 2, Cubes: []string{"1-", "-1"}, Output: '1'}, DecFactorLimits{})
	formatted, err = FormatDecoGraph(orGraph, nil, "")
	if err != nil || formatted != "     F = (b + a)\n" {
		t.Fatalf("OR format %q %v", formatted, err)
	}
	original, err = FormatDecoGraphOriginal(orGraph, nil, "")
	if err != nil || original != "     F = b + a\n" {
		t.Fatalf("original OR format %q %v", original, err)
	}
	orGraph.Close()
	constant, _ := NewConstantDecoGraph(false, DecoGraphLimits{})
	if text, formatErr := FormatDecoGraph(constant, nil, ""); formatErr != nil || text != "     F = Constant 0\n" {
		t.Fatalf("constant format %q err=%v", text, formatErr)
	}
	constant.Close()
	leaf, _ := NewLeafDecoGraph(0, 1, true, DecoGraphLimits{})
	if text, formatErr := FormatDecoGraph(leaf, []string{"input"}, "Y"); formatErr != nil || text != "     Y = !input\n" {
		t.Fatalf("leaf format %q err=%v", text, formatErr)
	}
	leaf.Close()
	xorGraph, _ := NewDecoGraph(2, DecoGraphLimits{})
	xorRoot, _ := xorGraph.AddXOR(DecoEdge{Node: 0}, DecoEdge{Node: 1}, false)
	_ = xorGraph.SetRoot(xorRoot)
	if text, formatErr := FormatDecoGraph(xorGraph, []string{"a", "b"}, "Y"); formatErr != nil || text != "     Y = (a # b)\n" {
		t.Fatalf("XOR format %q err=%v", text, formatErr)
	}
	xorGraph.Close()
}

func TestDecConcreteAIGConversionsCountUpdateOwnership(t *testing.T) {
	sop := SOPCover{Variables: 3, Cubes: []string{"11-", "0-1"}, Output: '1'}
	graph, err := FactorSOP(sop, DecFactorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	manager, err := NewDecAIG(3, DecFactorLimits{MaxNodes: 128})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	inputs := append([]DecAIGEdge(nil), manager.variables...)
	root, err := manager.BuildGraph(graph, inputs, true)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		expected, _ := sop.Evaluate(values)
		actual, aigErr := manager.Evaluate(root, values)
		if aigErr != nil || actual != expected {
			t.Fatalf("AIG assignment %d: %v %v want %v", assignment, actual, aigErr, expected)
		}
	}
	nodesBefore := len(manager.nodes)
	rootAgain, err := manager.BuildGraph(graph, inputs, true)
	if err != nil || rootAgain != root || len(manager.nodes) != nodesBefore {
		t.Fatal("strash reuse", rootAgain, root, err, nodesBefore, len(manager.nodes))
	}
	if count, countErr := manager.CountGraph(root, graph, inputs, 100, 100); count != 0 || !errors.Is(countErr, ErrInvariant) {
		t.Fatal("root reuse rejection", count, countErr)
	}
	count, err := manager.CountGraph(inputs[0], graph, inputs, 100, 100)
	if err != nil || count != 0 {
		t.Fatal("existing graph count", count, err)
	}
	for _, row := range []struct {
		second DecoEdge
		name   string
	}{{DecoEdge{Node: 0}, "idempotence"}, {DecoEdge{Node: 0, Complement: true}, "complement"}} {
		reduced, graphErr := NewDecoGraph(2, DecoGraphLimits{})
		if graphErr != nil {
			t.Fatal(graphErr)
		}
		reducedRoot, graphErr := reduced.AddAND(DecoEdge{Node: 0}, row.second)
		if graphErr == nil {
			graphErr = reduced.SetRoot(reducedRoot)
		}
		if graphErr != nil {
			t.Fatal(graphErr)
		}
		if got, countErr := manager.CountGraph(inputs[1], reduced, inputs[:2], 0, 0); countErr != nil || got != 0 {
			t.Fatalf("trivial lookup %s count=%d err=%v", row.name, got, countErr)
		}
		reduced.Close()
	}
	markedGraph, graphErr := NewDecoGraph(3, DecoGraphLimits{})
	if graphErr != nil {
		t.Fatal(graphErr)
	}
	markedFirst, _ := markedGraph.AddAND(DecoEdge{Node: 0}, DecoEdge{Node: 1})
	markedRoot, _ := markedGraph.AddAND(markedFirst, DecoEdge{Node: 2})
	_ = markedGraph.SetRoot(markedRoot)
	markedManager, _ := NewDecAIG(3, DecFactorLimits{MaxNodes: 32})
	existingPair, _ := markedManager.and(markedManager.variables[0], markedManager.variables[1], true)
	if count, countErr := markedManager.CountGraph(markedManager.variables[0], markedGraph, markedManager.variables, 2, 2); countErr != nil || count != 1 {
		t.Fatalf("unmarked count=%d err=%v", count, countErr)
	}
	if markErr := markedManager.MarkNode(existingPair.Node, true); markErr != nil {
		t.Fatal(markErr)
	}
	if count, countErr := markedManager.CountGraph(markedManager.variables[0], markedGraph, markedManager.variables, 2, 2); countErr != nil || count != 2 {
		t.Fatalf("marked count=%d err=%v", count, countErr)
	}
	markedManager.Close()
	markedGraph.Close()
	plain, _ := NewDecAIG(3, DecFactorLimits{MaxNodes: 128})
	if _, err = plain.BuildGraph(graph, plain.variables, false); err != nil {
		t.Fatal("unstrashed first build", err)
	}
	plainBefore := len(plain.nodes)
	if _, err = plain.BuildGraph(graph, plain.variables, false); err != nil || len(plain.nodes) != plainBefore+len(graph.nodes)-graph.leaves {
		t.Fatal("unstrashed node preservation", err, plainBefore, len(plain.nodes))
	}
	plain.Close()
	snapshot, err := manager.Nodes()
	if err != nil {
		t.Fatal(err)
	}
	snapshot[0].Alive = false
	if !manager.nodes[0].Alive {
		t.Fatal("node snapshot aliases manager")
	}

	limited, _ := NewDecAIG(3, DecFactorLimits{MaxNodes: 4})
	before := len(limited.nodes)
	if _, err = limited.BuildGraph(graph, limited.variables, true); !errors.Is(err, ErrLimit) || len(limited.nodes) != before {
		t.Fatal("atomic node limit", err, before, len(limited.nodes))
	}
	limited.Close()
	if _, err = NewDecAIG(2, DecFactorLimits{MaxNodes: decoEdgeNodeLimit + 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("edge representation limit", err)
	}

	orGraph, _ := FactorSOP(SOPCover{Variables: 2, Cubes: []string{"1-", "-1"}, Output: '1'}, DecFactorLimits{})
	updateManager, _ := NewDecAIG(2, DecFactorLimits{MaxNodes: 32})
	andGraph, _ := FactorSOP(SOPCover{Variables: 2, Cubes: []string{"11"}, Output: '1'}, DecFactorLimits{})
	old, _ := updateManager.BuildGraph(andGraph, updateManager.variables, true)
	consumer, _ := updateManager.and(old, updateManager.variables[0], true)
	replacement, err := updateManager.UpdateGraph(old, orGraph, updateManager.variables)
	if err != nil {
		t.Fatal("update", err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0}
		actual, evalErr := updateManager.Evaluate(consumer, values)
		want := (values[0] || values[1]) && values[0]
		if evalErr != nil || actual != want {
			t.Fatalf("updated consumer %d: %v %v want %v replacement %+v", assignment, actual, evalErr, want, replacement)
		}
	}
	andGraph.Close()
	orGraph.Close()
	updateManager.Close()

	stop := false
	atomicManager, _ := NewDecAIG(2, DecFactorLimits{MaxNodes: 32, Stop: func() bool { return stop }})
	atomicOld, _ := atomicManager.and(atomicManager.variables[0], atomicManager.variables[1], true)
	_, _ = atomicManager.and(atomicOld, atomicManager.variables[0], true)
	atomicBefore, _ := atomicManager.Nodes()
	replacementGraph, _ := FactorSOP(SOPCover{Variables: 2, Cubes: []string{"1-", "-1"}, Output: '1'}, DecFactorLimits{})
	stop = true
	if _, updateErr := atomicManager.UpdateGraph(atomicOld, replacementGraph, atomicManager.variables); !errors.Is(updateErr, ErrCancelled) {
		t.Fatalf("cancelled update: %v", updateErr)
	}
	atomicAfter, _ := atomicManager.Nodes()
	if !slices.Equal(atomicBefore, atomicAfter) {
		t.Fatal("cancelled update mutated manager")
	}
	stop = false
	if _, updateErr := atomicManager.UpdateGraph(atomicOld, replacementGraph, []DecAIGEdge{atomicOld, atomicManager.variables[1]}); !errors.Is(updateErr, ErrInvariant) {
		t.Fatalf("cyclic update: %v", updateErr)
	}
	replacementGraph.Close()
	atomicManager.Close()
}

func decRandomSOP(random *rand.Rand, data []byte) SOPCover {
	variables := 1 + random.Intn(4)
	cubeCount := 1 + random.Intn(5)
	cover := SOPCover{Variables: variables, Cubes: make([]string, cubeCount), Output: '1'}
	if random.Intn(2) != 0 {
		cover.Output = '0'
	}
	position := 0
	for cubeIndex := range cover.Cubes {
		cube := make([]byte, variables)
		care := false
		for variable := range cube {
			value := random.Intn(3)
			if len(data) > 0 {
				value = int(data[position%len(data)] % 3)
				position++
			}
			cube[variable] = "-01"[value]
			care = care || value != 0
		}
		if !care {
			cube[cubeIndex%variables] = byte('0' + cubeIndex%2)
		}
		cover.Cubes[cubeIndex] = string(cube)
	}
	return cover
}

func runDecFactorModel(data []byte, executions int) error {
	seed := int64(0xdec0)
	for _, value := range data {
		seed = seed*257 + int64(value)
	}
	random := rand.New(rand.NewSource(seed))
	for execution := 0; execution < executions; execution++ {
		sop := decRandomSOP(random, data)
		graph, err := FactorSOP(sop, DecFactorLimits{MaxNodes: 256, MaxDepth: 256})
		if err != nil {
			return err
		}
		manager, err := NewDecAIG(sop.Variables, DecFactorLimits{MaxNodes: 256, MaxDepth: 256})
		if err != nil {
			graph.Close()
			return err
		}
		root, err := manager.BuildGraph(graph, manager.variables, true)
		if err != nil {
			manager.Close()
			graph.Close()
			return err
		}
		for assignment := 0; assignment < 1<<uint(sop.Variables); assignment++ {
			inputs := make([]bool, sop.Variables)
			for variable := range inputs {
				inputs[variable] = assignment&(1<<uint(variable)) != 0
			}
			expected, evalErr := sop.Evaluate(inputs)
			if evalErr != nil {
				manager.Close()
				graph.Close()
				return evalErr
			}
			actual, evalErr := graph.Evaluate(inputs)
			if evalErr != nil || actual != expected {
				manager.Close()
				graph.Close()
				if evalErr != nil {
					return evalErr
				}
				return ErrInvariant
			}
			aigActual, aigErr := manager.Evaluate(root, inputs)
			if aigErr != nil || aigActual != expected {
				manager.Close()
				graph.Close()
				if aigErr != nil {
					return aigErr
				}
				return ErrInvariant
			}
		}
		manager.Close()
		graph.Close()
	}
	return nil
}

func TestDecIndependentFactorModel5000(t *testing.T) {
	if err := runDecFactorModel([]byte{1, 9, 2, 8, 3, 7, 4, 6, 5}, 5000); err != nil {
		t.Fatal(err)
	}
}

func FuzzDecFactorModel(f *testing.F) {
	f.Add([]byte{0})
	f.Add([]byte{1, 2, 0, 2, 1, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			data = data[:64]
		}
		if err := runDecFactorModel(data, 2); err != nil {
			t.Fatal(err)
		}
	})
}
