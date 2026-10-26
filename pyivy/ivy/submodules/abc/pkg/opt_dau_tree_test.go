package abc

// Go-only structural and assignment models for the DAU tree manager core.

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestDauTreeOperationsAndTruth(t *testing.T) {
	manager, err := NewDauTreeManager(DauTreeOptions{Variables: 4, NonDecomposableLimit: 3})
	if err != nil {
		t.Fatal(err)
	}
	v := manager.Variable()
	and, err := manager.Operation(DauDSDAnd, []DauTreeLiteral{v, v.Not()}, nil)
	if err != nil {
		t.Fatal(err)
	}
	reversed, err := manager.Operation(DauDSDAnd, []DauTreeLiteral{v.Not(), v}, nil)
	if err != nil || slices.Equal(manager.nodes[and.ID()].Fanins, manager.nodes[reversed.ID()].Fanins) {
		t.Fatal("complemented variable leaves did not preserve stable source order", err)
	}
	nested, err := manager.Operation(DauDSDAnd, []DauTreeLiteral{v, and}, nil)
	if err != nil || len(manager.nodes[nested.ID()].Fanins) != 3 {
		t.Fatal(nested, err)
	}
	xor, err := manager.Operation(DauDSDXor, []DauTreeLiteral{v.Not(), v}, nil)
	if err != nil || !xor.Complemented() {
		t.Fatal(xor, err)
	}
	muxA, err := manager.Operation(DauDSDMux, []DauTreeLiteral{v.Not(), v, v.Not()}, nil)
	if err != nil {
		t.Fatal(err)
	}
	muxB, err := manager.Operation(DauDSDMux, []DauTreeLiteral{v, v.Not(), v}, nil)
	if err != nil || muxA != muxB {
		t.Fatal(muxA, muxB, err)
	}
	primeTruth, _ := TruthTableFromWords(2, []uint64{0x6666666666666666}, 0)
	prime, err := manager.Operation(DauDSDPrime, []DauTreeLiteral{v, v}, primeTruth)
	if err != nil {
		t.Fatal(err)
	}
	permutation := []DauTreeLiteral{0, 2, 4}
	truth, err := manager.Truth(prime, permutation[:2], 0, nil)
	if err != nil || truth.words[0] != 0x6666666666666666 {
		t.Fatal(truth, err)
	}
	andTruth, err := manager.Truth(and, permutation[:2], 0, nil)
	if err != nil || andTruth.words[0] != 0x2222222222222222 {
		t.Fatal(andTruth, err)
	}
	nodes := manager.Nodes()
	nodes[2].Fanins[0] ^= 1
	if slices.Equal(nodes[2].Fanins, manager.Nodes()[2].Fanins) {
		t.Fatal("node snapshot aliases manager")
	}
}

func TestDauTreeLimitsAndErrors(t *testing.T) {
	if manager, err := NewDauTreeManager(DauTreeOptions{Variables: 13}); manager != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(manager, err)
	}
	manager, _ := NewDauTreeManager(DauTreeOptions{Variables: 3, MaxNodes: 3})
	v := manager.Variable()
	if _, err := manager.Operation(DauDSDMux, []DauTreeLiteral{v, v}, nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := manager.Operation(DauDSDAnd, []DauTreeLiteral{v, 999}, nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := manager.Operation(DauDSDAnd, []DauTreeLiteral{v, v}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Operation(DauDSDXor, []DauTreeLiteral{v, v}, nil); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	cancelled, _ := NewDauTreeManager(DauTreeOptions{Variables: 2, Stop: func() bool { return true }})
	if _, err := cancelled.Operation(DauDSDAnd, []DauTreeLiteral{cancelled.Variable(), cancelled.Variable()}, nil); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if truth, err := manager.Truth(2, []DauTreeLiteral{0}, 0, func() bool { return true }); truth != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(truth, err)
	}
}

func TestDauTreeAndWithPermutation(t *testing.T) {
	manager, err := NewDauTreeManager(DauTreeOptions{Variables: 3})
	if err != nil {
		t.Fatal(err)
	}
	v := manager.Variable()
	pair, err := manager.Operation(DauDSDAnd, []DauTreeLiteral{v, v}, nil)
	if err != nil {
		t.Fatal(err)
	}
	literal, permutation, err := manager.AndWithPermutation([]DauTreeLiteral{v.Not(), pair})
	if err != nil || !slices.Equal(permutation, []DauTreeLiteral{1, 2, 4}) {
		t.Fatal(literal, permutation, err)
	}
	node := manager.nodes[literal.ID()]
	if !slices.Equal(node.Fanins, []DauTreeLiteral{v, v, v}) {
		t.Fatal(node)
	}
	truth, err := manager.Truth(literal, permutation, 0, nil)
	if err != nil || truth.words[0]&0xff != 0x40 {
		t.Fatal(truth, err)
	}
	again, againPermutation, err := manager.AndWithPermutation([]DauTreeLiteral{v.Not(), pair})
	if err != nil || again != literal || !slices.Equal(againPermutation, permutation) {
		t.Fatal(again, againPermutation, err)
	}
}

func TestDauTreeFormatAndPrimeDetection(t *testing.T) {
	manager, _ := NewDauTreeManager(DauTreeOptions{Variables: 3})
	v := manager.Variable()
	primeTruth, _ := TruthTableFromWords(2, []uint64{0x6666666666666666}, 0)
	prime, err := manager.Operation(DauDSDPrime, []DauTreeLiteral{v, v}, primeTruth)
	if err != nil {
		t.Fatal(err)
	}
	root, permutation, err := manager.AndWithPermutation([]DauTreeLiteral{v.Not(), prime})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(permutation, []DauTreeLiteral{1, 2, 4}) {
		t.Fatal(permutation)
	}
	text, err := manager.Format(root, permutation, 0)
	if err != nil || text != "(!a6{bc})" {
		t.Fatal(text, err)
	}
	formula, err := ParseDauDSD(text, DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := formula.TruthTable(3, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	want, err := manager.Truth(root, permutation, 0, nil)
	if err != nil || !slices.Equal(got.words, want.words) {
		t.Fatal(got, want, err)
	}
	if found, err := manager.ContainsPrime(root); err != nil || !found {
		t.Fatal(found, err)
	}
	if found, err := manager.ContainsPrime(v); err != nil || found {
		t.Fatal(found, err)
	}
	if text, err := manager.Format(root, permutation, 3); text != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(text, err)
	}
}

func TestDauTreePrimeCanonicalization(t *testing.T) {
	manager, err := NewDauTreeManager(DauTreeOptions{Variables: 12})
	if err != nil {
		t.Fatal(err)
	}
	variable := manager.Variable()
	andNode, err := manager.Operation(DauDSDAnd, []DauTreeLiteral{variable, variable}, nil)
	if err != nil {
		t.Fatal(err)
	}
	xorNode, err := manager.Operation(DauDSDXor, []DauTreeLiteral{variable, andNode}, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs := []DauTreeLiteral{variable, andNode.Not(), xorNode}
	prime := dauDSDTestTable(t, 3, func(assignment uint64) bool {
		return uint8(0xd8)&(1<<uint(assignment)) != 0
	})
	canonical, err := prime.CanonicalizeNPN(TruthCanonicalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	literal, err := manager.Operation(DauDSDPrime, inputs, prime)
	if err != nil {
		t.Fatal(err)
	}
	if literal.Complemented() != canonical.OutputPhase {
		t.Fatal(literal, canonical.OutputPhase)
	}
	node := manager.nodes[literal.ID()]
	if node.Kind != DauDSDPrime || !slices.Equal(node.Truth, canonical.Table.Words()) {
		t.Fatal(node)
	}
	want := make([]DauTreeLiteral, len(inputs))
	for source, target := range canonical.Permutation {
		want[target] = inputs[source]
		if canonical.InputPhases[source] {
			want[target] = want[target].Not()
		}
	}
	if !slices.Equal(node.Fanins, want) {
		t.Fatal(node.Fanins, want)
	}
}

func TestDauTreeImportFormula(t *testing.T) {
	for _, text := range []string{
		"0", "1", "!a", "(a!b)", "![a(bc)]", "<a[bc](de)>", "0{}", "A{a}", "6{ab}", "!8{a!b}",
	} {
		formula, err := ParseDauDSD(text, DauDSDOptions{})
		if err != nil {
			t.Fatal(text, err)
		}
		variables := max(1, formula.VariableCount())
		manager, _ := NewDauTreeManager(DauTreeOptions{Variables: variables, MaxNodes: 128})
		literal, mapping, err := manager.ImportFormula(text, DauDSDOptions{})
		if err != nil {
			t.Fatal(text, err)
		}
		got, err := manager.Truth(literal, mapping, 0, nil)
		if err != nil {
			t.Fatal(text, err)
		}
		want, err := formula.TruthTable(variables, 0, nil)
		if err != nil || !slices.Equal(got.words, want.words) {
			t.Fatal(text, got, want, mapping, err)
		}
		roundTrip, err := manager.Format(literal, mapping, 0)
		if err != nil {
			t.Fatal(text, err)
		}
		reparsed, err := ParseDauDSD(roundTrip, DauDSDOptions{})
		if err != nil {
			t.Fatal(text, roundTrip, err)
		}
		replayed, err := reparsed.TruthTable(variables, 0, nil)
		if err != nil || !slices.Equal(replayed.words, want.words) {
			t.Fatal(text, roundTrip, replayed, want, err)
		}
	}
}

func TestDauTreeImportAtomicErrors(t *testing.T) {
	manager, _ := NewDauTreeManager(DauTreeOptions{Variables: 4, MaxNodes: 3})
	if literal, mapping, err := manager.ImportFormula("(ab)", DauDSDOptions{}); literal.ID() != 2 || len(mapping) != 2 || err != nil {
		t.Fatal(literal, mapping, err)
	}
	before := manager.Nodes()
	if literal, mapping, err := manager.ImportFormula("[ab]", DauDSDOptions{}); literal != 0 || mapping != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(literal, mapping, err)
	}
	if !slices.EqualFunc(before, manager.Nodes(), func(a, b DauTreeNode) bool {
		return a.Kind == b.Kind && a.Support == b.Support && slices.Equal(a.Fanins, b.Fanins) && slices.Equal(a.Truth, b.Truth)
	}) {
		t.Fatal("failed import mutated manager")
	}
	bindingManager, _ := NewDauTreeManager(DauTreeOptions{Variables: 5})
	if _, _, err := bindingManager.ImportFormula("<abb>{de}", DauDSDOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if literal, mapping, err := bindingManager.ImportFormula("<aaa>", DauDSDOptions{}); literal != 0 || mapping != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(literal, mapping, err)
	}
}

func TestDauTreeSharedMapAndTransparency(t *testing.T) {
	manager, _ := NewDauTreeManager(DauTreeOptions{Variables: 5})
	v := manager.Variable()
	first, _ := manager.Operation(DauDSDAnd, []DauTreeLiteral{v, v}, nil)
	second, _ := manager.Operation(DauDSDXor, []DauTreeLiteral{v, v}, nil)
	shared, err := manager.SharedMap(first, second,
		[]DauTreeLiteral{2, 5},
		[]DauTreeLiteral{4, 3},
		(1<<1)|(1<<2))
	if err != nil || !slices.Equal(shared.Variables, []DauTreeSharedVariable{{SecondIndex: 0, First: 3}, {SecondIndex: 1, First: 1}}) {
		t.Fatal(shared, err)
	}
	shared.Variables[0].First = 99
	again, err := manager.SharedMap(first, second, []DauTreeLiteral{2, 5}, []DauTreeLiteral{4, 3}, (1<<1)|(1<<2))
	if err != nil || again.Variables[0].First == 99 {
		t.Fatal(again, err)
	}
	if _, err := manager.SharedMap(first, second, []DauTreeLiteral{2, 2}, []DauTreeLiteral{4, 3}, 0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err := manager.SharedMap(first, second, []DauTreeLiteral{2, 5}, []DauTreeLiteral{8, 3}, 1<<4); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	andTransparent, err := manager.IsTransparent(first)
	xorTransparent, err2 := manager.IsTransparent(second)
	mux, _ := manager.Operation(DauDSDMux, []DauTreeLiteral{v, second, v}, nil)
	muxTransparent, err3 := manager.IsTransparent(mux)
	primeTruth, _ := TruthTableFromWords(2, []uint64{0x6666666666666666}, 0)
	prime, _ := manager.Operation(DauDSDPrime, []DauTreeLiteral{v, v}, primeTruth)
	primeTransparent, err4 := manager.IsTransparent(prime)
	if err != nil || err2 != nil || err3 != nil || err4 != nil || andTransparent || !xorTransparent || !muxTransparent || primeTransparent {
		t.Fatal(andTransparent, xorTransparent, muxTransparent, primeTransparent, err, err2, err3, err4)
	}
}

func TestDauTreeMergeDisjointAndConstants(t *testing.T) {
	manager, _ := NewDauTreeManager(DauTreeOptions{Variables: 4})
	first, firstMapping, err := manager.ImportFormula("(a!b)", DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	second, secondMapping, err := manager.ImportFormula("[cd]", DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if first > second {
		first, second = second, first
		firstMapping, secondMapping = secondMapping, firstMapping
	}
	result, err := manager.MergeAnd(first, second, firstMapping, secondMapping, 0, 4)
	if err != nil || len(result.Mapping) != 4 {
		t.Fatal(result, err)
	}
	got, err := manager.Truth(result.Literal, result.Mapping, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	firstTruth, _ := manager.Truth(first, firstMapping, 0, nil)
	secondTruth, _ := manager.Truth(second, secondMapping, 0, nil)
	for index := range got.words {
		if got.words[index] != firstTruth.words[index]&secondTruth.words[index] {
			t.Fatal(got, firstTruth, secondTruth)
		}
	}
	zero, err := manager.MergeAnd(0, second, nil, secondMapping, 0, 4)
	one, err2 := manager.MergeAnd(1, second, nil, secondMapping, 0, 4)
	if err != nil || err2 != nil || zero.Literal != 0 || zero.Mapping != nil || one.Literal != second || !slices.Equal(one.Mapping, secondMapping) {
		t.Fatal(zero, one, err, err2)
	}
	one.Mapping[0] ^= 1
	if slices.Equal(one.Mapping, secondMapping) {
		t.Fatal("constant merge mapping aliases input")
	}
	if _, err := manager.MergeAnd(second, first, secondMapping, firstMapping, 0, 4); second > first && !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err := manager.MergeAnd(first, second, firstMapping, secondMapping, 0, 3); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	invalidMap := slices.Clone(secondMapping)
	invalidMap[0] = firstMapping[0]
	if _, err := manager.MergeAnd(first, second, firstMapping, invalidMap, 0, 4); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	sharedMask := uint32(1) << uint(firstMapping[0].ID())
	sharedSecond := slices.Clone(secondMapping)
	sharedSecond[0] = firstMapping[0]
	sharedResult, err := manager.MergeAnd(first, second, firstMapping, sharedSecond, sharedMask, 3)
	if err != nil {
		t.Fatal(err)
	}
	sharedTruth, err := manager.Truth(sharedResult.Literal, sharedResult.Mapping, 0, nil)
	firstTruth, err2 = manager.Truth(first, firstMapping, 0, nil)
	secondTruth, err3 := manager.Truth(second, sharedSecond, 0, nil)
	if err != nil || err2 != nil || err3 != nil {
		t.Fatal(err, err2, err3)
	}
	for index := range sharedTruth.words {
		if sharedTruth.words[index] != firstTruth.words[index]&secondTruth.words[index] {
			t.Fatal(sharedResult, sharedTruth, firstTruth, secondTruth)
		}
	}
}

func TestDauTreeMergeSharedCacheAndPrimeLimit(t *testing.T) {
	manager, _ := NewDauTreeManager(DauTreeOptions{Variables: 4})
	if err := manager.EnableCache(4); err != nil {
		t.Fatal(err)
	}
	first, firstMap, _ := manager.ImportFormula("(ab)", DauDSDOptions{})
	second, secondMap, _ := manager.ImportFormula("[ac]", DauDSDOptions{})
	if first > second {
		first, second = second, first
		firstMap, secondMap = secondMap, firstMap
	}
	sharedMask := uint32(1)
	firstResult, err := manager.MergeAnd(first, second, firstMap, secondMap, sharedMask, 3)
	if err != nil {
		t.Fatal(err)
	}
	stats := manager.CacheStats()
	if stats.SharedMisses != 1 || stats.SharedEntries != 1 || stats.SharedHits != 0 {
		t.Fatal(stats)
	}
	remappedFirst := slices.Clone(firstMap)
	remappedSecond := slices.Clone(secondMap)
	for index := range remappedFirst {
		phase := remappedFirst[index].Complemented()
		remappedFirst[index] = DauTreeLiteral((3 - remappedFirst[index].ID()) << 1)
		if phase {
			remappedFirst[index] = remappedFirst[index].Not()
		}
	}
	for index := range remappedSecond {
		phase := remappedSecond[index].Complemented()
		remappedSecond[index] = DauTreeLiteral((3 - remappedSecond[index].ID()) << 1)
		if phase {
			remappedSecond[index] = remappedSecond[index].Not()
		}
	}
	remappedMask := uint32(1) << uint(remappedFirst[0].ID())
	secondResult, err := manager.MergeAnd(first, second, remappedFirst, remappedSecond, remappedMask, 3)
	if err != nil || secondResult.Literal != firstResult.Literal {
		t.Fatal(firstResult, secondResult, err)
	}
	stats = manager.CacheStats()
	if stats.SharedHits != 1 || stats.SharedMisses != 1 || stats.SharedEntries != 1 {
		t.Fatal(stats)
	}
	secondResult.Mapping[0] ^= 1
	again, err := manager.MergeAnd(first, second, remappedFirst, remappedSecond, remappedMask, 3)
	if err != nil || slices.Equal(secondResult.Mapping, again.Mapping) {
		t.Fatal("shared cached mapping aliases caller", err)
	}

	foundLimited := false
	candidates := []byte{0xe8, 0xd8, 0xca, 0xac, 0x96, 0xe2, 0xb4, 0x8e, 0xd2, 0xa6}
	for _, firstBits := range candidates {
		for _, secondBits := range candidates {
			limited, _ := NewDauTreeManager(DauTreeOptions{Variables: 4, NonDecomposableLimit: 2})
			firstText := dauTreePrimeDigits([]uint64{uint64(firstBits) * 0x0101010101010101}, 3) + "{abc}"
			secondText := dauTreePrimeDigits([]uint64{uint64(secondBits) * 0x0101010101010101}, 3) + "{abc}"
			firstPrime, firstPrimeMap, firstErr := limited.ImportFormula(firstText, DauDSDOptions{})
			secondPrime, secondPrimeMap, secondErr := limited.ImportFormula(secondText, DauDSDOptions{})
			if firstErr != nil || secondErr != nil {
				t.Fatal(firstErr, secondErr)
			}
			if firstPrime > secondPrime {
				firstPrime, secondPrime = secondPrime, firstPrime
				firstPrimeMap, secondPrimeMap = secondPrimeMap, firstPrimeMap
			}
			before := limited.NodeCount()
			_, mergeErr := limited.MergeAnd(firstPrime, secondPrime, firstPrimeMap, secondPrimeMap, 7, 3)
			if errors.Is(mergeErr, ErrLimit) {
				if limited.NodeCount() != before {
					t.Fatal("limited shared merge mutated manager", before, limited.NodeCount())
				}
				foundLimited = true
				break
			}
			if mergeErr != nil {
				t.Fatal(mergeErr)
			}
		}
		if foundLimited {
			break
		}
	}
	if !foundLimited {
		t.Fatal("candidate set had no irreducible shared conjunction")
	}
}

func TestDauTreeMergeCache(t *testing.T) {
	manager, _ := NewDauTreeManager(DauTreeOptions{Variables: 4})
	if err := manager.EnableCache(1); err != nil {
		t.Fatal(err)
	}
	first, firstMap, _ := manager.ImportFormula("(ab)", DauDSDOptions{})
	second, secondMap, _ := manager.ImportFormula("[cd]", DauDSDOptions{})
	if first > second {
		first, second = second, first
		firstMap, secondMap = secondMap, firstMap
	}
	initialNodes := manager.NodeCount()
	firstResult, err := manager.MergeAnd(first, second, firstMap, secondMap, 0, 4)
	if err != nil {
		t.Fatal(err)
	}
	stats := manager.CacheStats()
	if stats.UnsharedMisses != 1 || stats.UnsharedEntries != 1 || stats.UnsharedHits != 0 || manager.NodeCount() != initialNodes+1 {
		t.Fatal(stats, initialNodes, manager.NodeCount())
	}
	remappedFirst := []DauTreeLiteral{6, 1}
	remappedSecond := []DauTreeLiteral{4, 3}
	secondResult, err := manager.MergeAnd(first, second, remappedFirst, remappedSecond, 0, 4)
	if err != nil || secondResult.Literal != firstResult.Literal || manager.NodeCount() != initialNodes+1 {
		t.Fatal(firstResult, secondResult, err, manager.NodeCount())
	}
	stats = manager.CacheStats()
	if stats.UnsharedHits != 1 || stats.UnsharedMisses != 1 || stats.UnsharedEntries != 1 {
		t.Fatal(stats)
	}
	third, thirdMap, err := manager.ImportFormula("6{cd}", DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	left, right := first, third
	leftMap, rightMap := firstMap, thirdMap
	if left > right {
		left, right = right, left
		leftMap, rightMap = rightMap, leftMap
	}
	beforeLimit := manager.NodeCount()
	if _, err := manager.MergeAnd(left, right, leftMap, rightMap, 0, 4); !errors.Is(err, ErrLimit) || manager.NodeCount() != beforeLimit {
		t.Fatal(err, beforeLimit, manager.NodeCount())
	}
	secondResult.Mapping[0] ^= 1
	again, err := manager.MergeAnd(first, second, remappedFirst, remappedSecond, 0, 4)
	if err != nil || slices.Equal(secondResult.Mapping, again.Mapping) {
		t.Fatal("cached mapping aliases caller", err)
	}
	manager.DisableCache()
	if statsAfter := manager.CacheStats(); statsAfter != (DauTreeCacheStats{UnsharedHits: 2, UnsharedMisses: 2, UnsharedEntries: 1}) {
		t.Fatal(statsAfter)
	}
	if err := manager.EnableCache(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestDauTreeSummaryAndPrimeTruths(t *testing.T) {
	manager, _ := NewDauTreeManager(DauTreeOptions{Variables: 4})
	if _, _, err := manager.ImportFormula("(a6{bc})", DauDSDOptions{}); err != nil {
		t.Fatal(err)
	}
	summary, err := manager.Summary(0, 0)
	if err != nil || summary.Objects != manager.NodeCount() || summary.PrimeObjects != 1 || summary.PrimeStructures != 2 || len(summary.Rows) != manager.NodeCount() {
		t.Fatal(summary, err)
	}
	text, err := FormatDauTreeSummary(summary, 0)
	if err != nil || !strings.Contains(text, "Non-DSD objects = 1\n") || !strings.Contains(text, "6{bc}") {
		t.Fatal(text, err)
	}
	truths, err := manager.PrimeTruthTables()
	if err != nil || len(truths) != 1 || truths[0].Truth.variables != 4 || truths[0].Truth.words[0] != 0x6666666666666666 {
		t.Fatal(truths, err)
	}
	summary.Rows[0].Formula = "mutated"
	again, err := manager.Summary(1, 0)
	if err != nil || len(again.Rows) != 1 || again.Rows[0].Formula == "mutated" {
		t.Fatal(again, err)
	}
	if _, err := manager.Summary(0, 1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if text, err := FormatDauTreeSummary(summary, 1); text != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(text, err)
	}
	manager.stop = func() bool { return true }
	if _, err := manager.Summary(0, 0); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if truths, err := manager.PrimeTruthTables(); truths != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(truths, err)
	}
}

func FuzzDauTreeMergeDisjoint(f *testing.F) {
	f.Add(uint8(0), uint8(1), false, true)
	f.Fuzz(func(t *testing.T, firstVariable, secondVariable uint8, firstPhase, secondPhase bool) {
		firstIndex := int(firstVariable % 3)
		secondIndex := int(secondVariable%3) + 3
		manager, _ := NewDauTreeManager(DauTreeOptions{Variables: 6})
		variable := manager.Variable()
		firstMap := []DauTreeLiteral{DauTreeLiteral(firstIndex << 1)}
		secondMap := []DauTreeLiteral{DauTreeLiteral(secondIndex << 1)}
		if firstPhase {
			firstMap[0] = firstMap[0].Not()
		}
		if secondPhase {
			secondMap[0] = secondMap[0].Not()
		}
		result, err := manager.MergeAnd(variable, variable, firstMap, secondMap, 0, 2)
		if err != nil {
			t.Fatal(err)
		}
		truth, err := manager.Truth(result.Literal, result.Mapping, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 64; assignment++ {
			firstValue := assignment&(1<<uint(firstIndex)) != 0
			secondValue := assignment&(1<<uint(secondIndex)) != 0
			if firstPhase {
				firstValue = !firstValue
			}
			if secondPhase {
				secondValue = !secondValue
			}
			got := truth.words[0]&(1<<uint(assignment)) != 0
			if got != (firstValue && secondValue) {
				t.Fatal(assignment, got, firstValue, secondValue, result)
			}
		}
	})
}

func FuzzDauTreeMergeShared(f *testing.F) {
	f.Add(uint8(0), false, true)
	f.Fuzz(func(t *testing.T, variable uint8, firstPhase, secondPhase bool) {
		index := int(variable % 6)
		manager, _ := NewDauTreeManager(DauTreeOptions{Variables: 6})
		leaf := manager.Variable()
		firstMap := []DauTreeLiteral{DauTreeLiteral(index << 1)}
		secondMap := []DauTreeLiteral{DauTreeLiteral(index << 1)}
		if firstPhase {
			firstMap[0] = firstMap[0].Not()
		}
		if secondPhase {
			secondMap[0] = secondMap[0].Not()
		}
		result, err := manager.MergeAnd(leaf, leaf, firstMap, secondMap, 1<<uint(index), 1)
		if err != nil {
			t.Fatal(err)
		}
		truth, err := manager.Truth(result.Literal, result.Mapping, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 64; assignment++ {
			value := assignment&(1<<uint(index)) != 0
			firstValue, secondValue := value != firstPhase, value != secondPhase
			got := truth.words[0]&(1<<uint(assignment)) != 0
			if got != (firstValue && secondValue) {
				t.Fatal(assignment, got, firstValue, secondValue, result)
			}
		}
	})
}

func FuzzDauTreeMergeSharedStructured(f *testing.F) {
	f.Add(uint8(0), uint8(0))
	f.Fuzz(func(t *testing.T, offset, phases uint8) {
		manager, _ := NewDauTreeManager(DauTreeOptions{Variables: 6})
		first, firstMap, err := manager.ImportFormula("(ab)", DauDSDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		second, secondMap, err := manager.ImportFormula("[ac]", DauDSDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		base := int(offset % 6)
		firstMap = []DauTreeLiteral{DauTreeLiteral(base << 1), DauTreeLiteral(((base + 1) % 6) << 1)}
		secondMap = []DauTreeLiteral{DauTreeLiteral(base << 1), DauTreeLiteral(((base + 2) % 6) << 1)}
		for index := range firstMap {
			if phases&(1<<uint(index)) != 0 {
				firstMap[index] = firstMap[index].Not()
			}
		}
		for index := range secondMap {
			if phases&(1<<uint(index+2)) != 0 {
				secondMap[index] = secondMap[index].Not()
			}
		}
		firstTruth, err := manager.Truth(first, firstMap, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		secondTruth, err := manager.Truth(second, secondMap, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		if first > second {
			first, second = second, first
			firstMap, secondMap = secondMap, firstMap
		}
		result, err := manager.MergeAnd(first, second, firstMap, secondMap, 1<<uint(base), 3)
		if err != nil {
			t.Fatal(err)
		}
		got, err := manager.Truth(result.Literal, result.Mapping, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		for index := range got.words {
			if got.words[index] != firstTruth.words[index]&secondTruth.words[index] {
				t.Fatal(index, result, got, firstTruth, secondTruth)
			}
		}
	})
}

func FuzzDauTreeImportFormula(f *testing.F) {
	for _, seed := range []string{"(ab)", "![a(bc)]", "<abc>", "8{ab}", "[(ac)b]"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		formula, err := ParseDauDSD(text, DauDSDOptions{MaxBytes: 128, MaxNodes: 64, MaxDepth: 16, MaxVariables: 6})
		if err != nil {
			return
		}
		variables := max(1, formula.VariableCount())
		manager, err := NewDauTreeManager(DauTreeOptions{Variables: variables, MaxNodes: 128})
		if err != nil {
			t.Fatal(err)
		}
		literal, mapping, err := manager.ImportFormula(text, DauDSDOptions{MaxBytes: 128, MaxNodes: 64, MaxDepth: 16, MaxVariables: 6})
		if errors.Is(err, ErrUnsupported) || errors.Is(err, ErrInvariant) {
			return
		}
		if err != nil {
			t.Fatal(text, err)
		}
		got, err := manager.Truth(literal, mapping, 0, nil)
		if err != nil {
			t.Fatal(text, err)
		}
		want, err := formula.TruthTable(variables, 0, nil)
		if err != nil || !slices.Equal(got.words, want.words) {
			t.Fatal(text, got, want, mapping, err)
		}
	})
}

func FuzzDauTreeOperations(f *testing.F) {
	f.Add(uint8(0), uint8(1), uint8(2), false)
	f.Add(uint8(2), uint8(0), uint8(1), true)
	f.Fuzz(func(t *testing.T, first, second, third uint8, complement bool) {
		manager, err := NewDauTreeManager(DauTreeOptions{Variables: 3, MaxNodes: 64})
		if err != nil {
			t.Fatal(err)
		}
		v := manager.Variable()
		inputs := []DauTreeLiteral{v, v, v}
		phases := []uint8{first, second, third}
		for index := range inputs {
			if phases[index]&1 != 0 {
				inputs[index] = inputs[index].Not()
			}
		}
		kind := DauDSDNodeKind(uint8(DauDSDAnd) + first%3)
		literal, err := manager.Operation(kind, inputs, nil)
		if err != nil {
			t.Fatal(err)
		}
		if complement {
			literal = literal.Not()
		}
		mapped := DauTreeLiteral((second % 3) << 1)
		permutation := []DauTreeLiteral{mapped, mapped, mapped}
		truth, err := manager.Truth(literal, permutation, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			values := [3]bool{}
			for index := range values {
				values[index] = assignment&(1<<uint(mapped.ID())) != 0
				if inputs[index].Complemented() {
					values[index] = !values[index]
				}
			}
			want := values[0] && values[1] && values[2]
			if kind == DauDSDXor {
				want = values[0] != values[1] != values[2]
			} else if kind == DauDSDMux {
				if values[0] {
					want = values[1]
				} else {
					want = values[2]
				}
			}
			if complement {
				want = !want
			}
			got := truth.words[0]&(1<<uint(assignment)) != 0
			if got != want {
				t.Fatal(kind, inputs, permutation, assignment, got, want)
			}
		}
	})
}
