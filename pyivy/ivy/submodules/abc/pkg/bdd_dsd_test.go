package abc

// Go-only validation for the src/bdd/dsd port.

import (
	"slices"
	"testing"
)

func dsdTestTruth(t testing.TB, variables int, value uint64) *TruthTable {
	t.Helper()
	truth, err := NewConstantTruthTable(variables, false, 1<<14)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
		if value&(uint64(1)<<assignment) != 0 {
			if err = truth.Set(assignment, true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return truth
}

func dsdTruthValue(t testing.TB, truth *TruthTable) uint64 {
	t.Helper()
	var value uint64
	for assignment := uint64(0); assignment < uint64(1)<<uint(truth.Variables()); assignment++ {
		bit, err := truth.Get(assignment)
		if err != nil {
			t.Fatal(err)
		}
		if bit {
			value |= uint64(1) << assignment
		}
	}
	return value
}

// dsdIndependentEvaluate deliberately does not use Reconstruct or its helper.
func dsdIndependentEvaluate(t testing.TB, manager *DSDManager, edge DSDEdge, assignment uint64) bool {
	t.Helper()
	if edge.Node < 0 || edge.Node >= len(manager.nodes) {
		t.Fatalf("invalid edge %+v", edge)
	}
	node := manager.nodes[edge.Node]
	var err error
	var result bool
	switch node.Type {
	case DSDNodeConst1:
		result = true
	case DSDNodeBuffer:
		result = assignment&(uint64(1)<<uint(node.Variable)) != 0
	case DSDNodeOR:
		for _, child := range node.Children {
			result = result || dsdIndependentEvaluate(t, manager, child, assignment)
		}
	case DSDNodeXOR:
		for _, child := range node.Children {
			result = result != dsdIndependentEvaluate(t, manager, child, assignment)
		}
	case DSDNodePrime:
		var local uint64
		for index, child := range node.Children {
			if dsdIndependentEvaluate(t, manager, child, assignment) {
				local |= uint64(1) << uint(index)
			}
		}
		result, err = node.PrimeFunction.Get(local)
		if err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("bad node type %d", node.Type)
	}
	if edge.Complement {
		result = !result
	}
	return result
}

func checkDSDTruth(t testing.TB, manager *DSDManager, edge DSDEdge, expected *TruthTable) {
	t.Helper()
	rebuilt, err := manager.Reconstruct(edge)
	if err != nil {
		t.Fatal(err)
	}
	if !dsdTruthEqual(rebuilt, expected) {
		t.Fatalf("reconstruction mismatch got=%x want=%x", rebuilt.Words(), expected.Words())
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(expected.Variables()); assignment++ {
		want, _ := expected.Get(assignment)
		if got := dsdIndependentEvaluate(t, manager, edge, assignment); got != want {
			t.Fatalf("independent evaluation at %d got %v want %v", assignment, got, want)
		}
	}
}

func TestDSDNodeKindsMetricsOwnershipAndFormat(t *testing.T) {
	manager, err := NewDSDManager(3, DSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// x0 | x1 | x2, x0 ^ x1 ^ x2, and majority.
	functions := []*TruthTable{
		dsdTestTruth(t, 3, 0xfe),
		dsdTestTruth(t, 3, 0x96),
		dsdTestTruth(t, 3, 0xe8),
	}
	roots, err := manager.Decompose(functions)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds := []DSDNodeType{DSDNodeOR, DSDNodeXOR, DSDNodePrime}
	for index, root := range roots {
		stored, rootErr := manager.Root(index)
		if rootErr != nil || stored != root {
			t.Fatalf("stored root %d=%+v err=%v want=%+v", index, stored, rootErr, root)
		}
		node, nodeErr := manager.Node(root.Node)
		if nodeErr != nil {
			t.Fatal(nodeErr)
		}
		if node.Type != wantKinds[index] {
			t.Fatalf("root %d type=%d want=%d", index, node.Type, wantKinds[index])
		}
		checkDSDTruth(t, manager, root, functions[index])
		text, formatErr := manager.Format(root, []string{"a", "b", "c"})
		if formatErr != nil || text == "" {
			t.Fatalf("format=%q err=%v", text, formatErr)
		}
	}
	input, err := manager.Input(1)
	if err != nil || input.Node != 2 {
		t.Fatalf("input=%+v err=%v", input, err)
	}
	one, err := manager.ConstantOne()
	if err != nil || one != (DSDEdge{Node: 0}) {
		t.Fatalf("constant=%+v err=%v", one, err)
	}
	primeFunction, err := manager.PrimeFunction(roots[2].Node)
	if err != nil || dsdTruthValue(t, primeFunction) != 0xe8 {
		t.Fatalf("prime=%x err=%v", dsdTruthValue(t, primeFunction), err)
	}
	dfs, err := manager.DFS(roots[2])
	if err != nil || !slices.Equal(dfs, []int{roots[2].Node}) {
		t.Fatalf("dfs=%v err=%v", dfs, err)
	}
	inputDFS, err := manager.DFS(input)
	if err != nil || len(inputDFS) != 0 {
		t.Fatalf("input dfs=%v err=%v", inputDFS, err)
	}
	info, err := manager.TreeInfo(roots)
	if err != nil || info.Depth == 0 || info.MaximumGateSize < 2 {
		t.Fatalf("info=%+v err=%v", info, err)
	}
	cost, err := manager.AIGCost(roots[1])
	if err != nil || cost != 6 {
		t.Fatalf("xor cost=%d err=%v", cost, err)
	}
	nonterminal, prime, err := manager.NodeCounts(roots[2])
	if err != nil || nonterminal != 1 || prime != 1 {
		t.Fatalf("counts=%d,%d err=%v", nonterminal, prime, err)
	}
	variables, err := manager.DecomposableVariables(roots[0])
	if err != nil || !slices.Equal(variables, []int{0, 2, 1}) {
		t.Fatalf("decomposable=%v err=%v", variables, err)
	}
	node, _ := manager.Node(roots[0].Node)
	node.Children[0] = DSDEdge{Node: 999}
	checkDSDTruth(t, manager, roots[0], functions[0])
	if err = manager.SetMark(roots[0].Node, 17); err != nil {
		t.Fatal(err)
	}
	marked, _ := manager.Node(roots[0].Node)
	if marked.Mark != 17 {
		t.Fatalf("mark=%d", marked.Mark)
	}
	if err = manager.Unmark(); err != nil {
		t.Fatal(err)
	}
	unmarked, _ := manager.Node(roots[0].Node)
	if unmarked.Mark != 0 {
		t.Fatalf("unmark=%d", unmarked.Mark)
	}
}

func TestDSDExhaustiveFourVariableReconstruction(t *testing.T) {
	var manager *DSDManager
	for value := uint64(0); value < 1<<16; value++ {
		// Keep the exhaustive campaign while bounding manager validation cost.
		// Reusing one manager for all functions makes each checked operation scan
		// an ever-growing table and becomes quadratic under race instrumentation.
		if value%64 == 0 {
			var err error
			manager, err = NewDSDManager(4, DSDOptions{MaxNodes: 1 << 18, MaxOperations: 1 << 30})
			if err != nil {
				t.Fatal(err)
			}
		}
		truth := dsdTestTruth(t, 4, value)
		edge, buildErr := manager.DecomposeOne(truth)
		if buildErr != nil {
			t.Fatalf("value=%04x: %v", value, buildErr)
		}
		for assignment := uint64(0); assignment < 16; assignment++ {
			want := value&(uint64(1)<<assignment) != 0
			if got := dsdIndependentEvaluate(t, manager, edge, assignment); got != want {
				t.Fatalf("value=%04x assignment=%d got=%v want=%v", value, assignment, got, want)
			}
		}
		node := manager.nodes[edge.Node]
		if node.Support&^0xf != 0 {
			t.Fatalf("value=%04x support=%x", value, node.Support)
		}
		if node.Type == DSDNodeOR || node.Type == DSDNodeXOR || node.Type == DSDNodePrime {
			seen := uint64(0)
			for _, child := range node.Children {
				childNode := manager.nodes[child.Node]
				if seen&childNode.Support != 0 {
					t.Fatalf("value=%04x overlapping supports", value)
				}
				seen |= childNode.Support
			}
			if seen != node.Support {
				t.Fatalf("value=%04x child support=%x node=%x", value, seen, node.Support)
			}
		}
	}
}

func TestDSDBDDConversionCareLimitsCancellationAndLifecycle(t *testing.T) {
	bdd, err := NewExtraBDDManager(3, 128)
	if err != nil {
		t.Fatal(err)
	}
	x0, _ := bdd.Variable(0)
	x1, _ := bdd.Variable(1)
	root, err := bdd.And(x0, x1.Not(), ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	manager, _ := NewDSDManager(3, DSDOptions{})
	edge, err := manager.DecomposeBDD(bdd, root)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < 8; assignment++ {
		want := assignment&1 != 0 && assignment&2 == 0
		if got := dsdIndependentEvaluate(t, manager, edge, assignment); got != want {
			t.Fatalf("assignment=%d got=%v want=%v", assignment, got, want)
		}
	}
	all, _ := NewConstantTruthTable(3, true, 1)
	zero, _ := NewConstantTruthTable(3, false, 1)
	first := dsdTestTruth(t, 3, 0xaa)
	second := dsdTestTruth(t, 3, 0xaa)
	equal, err := DSDEquivalentOnCare(first, second, all, all)
	if err != nil || !equal {
		t.Fatalf("care equal=%v err=%v", equal, err)
	}
	equal, err = DSDEquivalentOnCare(first, zero, zero, all)
	if err != nil || !equal {
		t.Fatalf("empty care equal=%v err=%v", equal, err)
	}
	stopCalls := 0
	cancelled, _ := NewDSDManager(4, DSDOptions{Stop: func() bool {
		stopCalls++
		return stopCalls > 3
	}})
	before, _ := cancelled.Statistics()
	if _, err = cancelled.Decompose([]*TruthTable{dsdTestTruth(t, 4, 0x6996)}); err != ErrCancelled {
		t.Fatalf("cancel err=%v", err)
	}
	after, _ := cancelled.Statistics()
	if before != after {
		t.Fatalf("cancellation not atomic before=%+v after=%+v", before, after)
	}
	limited, _ := NewDSDManager(4, DSDOptions{MaxNodes: 5})
	if _, err = limited.DecomposeOne(dsdTestTruth(t, 4, 0xe8e8)); err != ErrLimit {
		t.Fatalf("node limit err=%v", err)
	}
	bddStopCalls := 0
	bddCancelled, _ := NewDSDManager(3, DSDOptions{Stop: func() bool {
		bddStopCalls++
		return bddStopCalls > 2
	}})
	bddBefore, _ := bddCancelled.Statistics()
	if _, err = bddCancelled.DecomposeBDD(bdd, root); err != ErrCancelled {
		t.Fatalf("bdd cancellation err=%v", err)
	}
	bddAfter, _ := bddCancelled.Statistics()
	if bddBefore != bddAfter {
		t.Fatalf("bdd cancellation not atomic before=%+v after=%+v", bddBefore, bddAfter)
	}
	statistics, err := manager.Close()
	if err != nil || statistics.Nodes < 4 {
		t.Fatalf("close stats=%+v err=%v", statistics, err)
	}
	if _, err = manager.Statistics(); err != ErrInvariant {
		t.Fatalf("closed err=%v", err)
	}
}

func TestDSDDecomposableVariablesLargePrimeBlocksAncestors(t *testing.T) {
	manager, err := NewDSDManager(6, DSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	function, err := NewConstantTruthTable(6, false, 1<<14)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < 64; assignment++ {
		firstFive := assignment & 31
		// This five-input function has no disjoint OR/XOR partition.
		prime := uint64(0x9e3779b8)&(uint64(1)<<firstFive) != 0
		value := prime || assignment&(1<<5) != 0
		if err = function.Set(assignment, value); err != nil {
			t.Fatal(err)
		}
	}
	edge, err := manager.DecomposeOne(function)
	if err != nil {
		t.Fatal(err)
	}
	node := manager.nodes[edge.Node]
	if node.Type != DSDNodeOR {
		t.Fatalf("root type=%d want OR", node.Type)
	}
	variables, err := manager.DecomposableVariables(edge)
	if err != nil {
		t.Fatal(err)
	}
	if len(variables) != 0 {
		t.Fatalf("large prime should block ancestor collection, got %v", variables)
	}
	checkDSDTruth(t, manager, edge, function)
}

func TestDSDIndependentRandomModel5000(t *testing.T) {
	manager, err := NewDSDManager(6, DSDOptions{MaxNodes: 1 << 18, MaxOperations: 1 << 30})
	if err != nil {
		t.Fatal(err)
	}
	state := uint64(0x9e3779b97f4a7c15)
	for iteration := 0; iteration < 5000; iteration++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		truth := dsdTestTruth(t, 6, state)
		edge, buildErr := manager.DecomposeOne(truth)
		if buildErr != nil {
			t.Fatalf("iteration=%d: %v", iteration, buildErr)
		}
		for assignment := uint64(0); assignment < 64; assignment++ {
			want := state&(uint64(1)<<assignment) != 0
			if got := dsdIndependentEvaluate(t, manager, edge, assignment); got != want {
				t.Fatalf("iteration=%d assignment=%d", iteration, assignment)
			}
		}
	}
}

func FuzzDSDDecomposition(f *testing.F) {
	f.Add(byte(3), []byte{0xe8})
	f.Add(byte(4), []byte{0x96, 0x69})
	f.Add(byte(6), []byte{0x81, 0x42, 0x24, 0x18, 0x18, 0x24, 0x42, 0x81})
	f.Fuzz(func(t *testing.T, rawVariables byte, data []byte) {
		variables := int(rawVariables % 7) // bounded independent truth oracle
		assignments := 1 << uint(variables)
		var value uint64
		for index := 0; index < assignments && index < 64; index++ {
			if len(data) != 0 && data[index%len(data)]&(1<<uint(index&7)) != 0 {
				value |= uint64(1) << uint(index)
			}
		}
		manager, err := NewDSDManager(variables, DSDOptions{MaxNodes: 4096, MaxOperations: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		truth := dsdTestTruth(t, variables, value)
		edge, err := manager.DecomposeOne(truth)
		if err != nil {
			t.Fatal(err)
		}
		checkDSDTruth(t, manager, edge, truth)
		text1, err := manager.Format(edge, nil)
		if err != nil {
			t.Fatal(err)
		}
		text2, _ := manager.Format(edge, nil)
		if text1 != text2 {
			t.Fatal("nondeterministic format")
		}
	})
}
