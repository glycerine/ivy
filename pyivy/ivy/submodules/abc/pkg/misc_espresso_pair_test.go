package abc

// Go-only pairing transforms and matching models for src/misc/espresso/pair.c.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func espressoPairedTruth(t testing.TB, original *EspressoPLA, paired EspressoPairResult, output int) uint64 {
	t.Helper()
	variables := original.Layout.binaryVariables
	pairedVariables := make([]bool, variables)
	for _, pair := range paired.Pairing.Pairs {
		pairedVariables[pair.First], pairedVariables[pair.Second] = true, true
	}
	truth := uint64(0)
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		for _, cube := range paired.PLA.On.sets {
			matches, newVariable := true, 0
			for variable := 0; variable < variables; variable++ {
				if pairedVariables[variable] {
					continue
				}
				value := (assignment >> uint(variable)) & 1
				present, _ := cube.Has(paired.PLA.Layout.firstParts[newVariable] + value)
				if !present {
					matches = false
					break
				}
				newVariable++
			}
			if !matches {
				continue
			}
			for pairIndex, pair := range paired.Pairing.Pairs {
				first := (assignment >> uint(pair.First)) & 1
				second := (assignment >> uint(pair.Second)) & 1
				part := 3 - (2*first + second)
				present, _ := cube.Has(paired.PLA.Layout.firstParts[paired.PLA.Layout.binaryVariables+pairIndex] + part)
				if !present {
					matches = false
					break
				}
			}
			if !matches {
				continue
			}
			outputVariable := paired.PLA.Layout.OutputVariable()
			present, _ := cube.Has(paired.PLA.Layout.firstParts[outputVariable] + output)
			if present {
				truth |= uint64(1) << uint(assignment)
				break
			}
		}
	}
	return truth
}

func TestEspressoPairVariablesLayoutLabelsTruthAndOwnership(t *testing.T) {
	truths := []uint64{0b10010110, 0b01101001}
	base := espressoPLAFromOutputTruths(t, 3, truths)
	labels := []string{"a.bar", "a", "b.bar", "b", "c.bar", "c", "y0", "y1"}
	pla, _ := NewEspressoPLA(base.Layout, base.On, base.DontCare, base.Off, labels)
	before := espressoFamilyMasks(t, pla.On)
	result, err := pla.PairVariables(EspressoPairing{Pairs: []EspressoVariablePair{{First: 0, Second: 2}}}, true, EspressoPairOptions{})
	if err != nil || result.PLA.Layout.binaryVariables != 1 || !slices.Equal(result.PLA.Layout.partSizes, []int{2, 4, 2}) || !slices.Equal(result.SparseVariables, []bool{false, false, true}) {
		t.Fatal(result.PLA, result.SparseVariables, err)
	}
	wantLabels := []string{"b.bar", "b", "a.bar+c.bar", "a.bar+c", "a+c.bar", "a+c", "y0", "y1"}
	if !slices.Equal(result.PLA.Labels, wantLabels) {
		t.Fatal(result.PLA.Labels, wantLabels)
	}
	for output, truth := range truths {
		if got := espressoPairedTruth(t, pla, result, output); got != truth {
			t.Fatal(output, got, truth)
		}
	}
	if !slices.Equal(espressoFamilyMasks(t, pla.On), before) {
		t.Fatal("pairing mutated input")
	}
}

func TestEspressoPairEnumerationBestAndFormat(t *testing.T) {
	pairings, err := EnumeratePairings(5, EspressoPairOptions{})
	if err != nil || len(pairings) != 15 {
		t.Fatal(len(pairings), err)
	}
	costs := [][]int{{0, 9, 1, 1}, {0, 0, 1, 8}, {0, 0, 0, 7}, {0, 0, 0, 0}}
	exact, exactCost, err := BestPairing(costs, true, EspressoPairOptions{})
	greedy, greedyCost, greedyErr := BestPairing(costs, false, EspressoPairOptions{})
	want := EspressoPairing{Pairs: []EspressoVariablePair{{First: 0, Second: 1}, {First: 2, Second: 3}}}
	if err != nil || greedyErr != nil || exactCost != 16 || greedyCost != 16 || !slices.Equal(exact.Pairs, want.Pairs) || !slices.Equal(greedy.Pairs, want.Pairs) || exact.Format() != "pair is (1 2) (3 4)" {
		t.Fatal(exact, exactCost, greedy, greedyCost, err, greedyErr)
	}
	clone := exact.Clone()
	clone.Pairs[0].First = 3
	if exact.Pairs[0].First != 0 {
		t.Fatal("pair clone aliases source")
	}
	negative := [][]int{{0, -5, -2}, {0, 0, -3}, {0, 0, 0}}
	negativeBest, negativeCost, err := BestPairing(negative, true, EspressoPairOptions{})
	if err != nil || len(negativeBest.Pairs) != 1 || negativeCost != -2 {
		t.Fatal(negativeBest, negativeCost, err)
	}
}

func TestEspressoPairCostsAndOptimizers(t *testing.T) {
	pla := espressoPLAFromOutputTruths(t, 2, []uint64{0b1111})
	options := EspressoPairOptions{
		Heuristic: EspressoMinimizeOptions{SingleExpand: true, MaxIterations: 16},
		Exact:     EspressoExactOptions{MaxSearchNodes: 1 << 16},
		Phase:     EspressoPhaseOptions{Heuristic: EspressoMinimizeOptions{SingleExpand: true, MaxIterations: 16}},
	}
	for _, strategy := range []EspressoPairCostStrategy{EspressoPairWeak, EspressoPairStrong, EspressoPairFull, EspressoPairExact} {
		costs, err := pla.PairingCosts(strategy, options)
		want := 3
		if strategy == EspressoPairStrong {
			want = 0
		}
		if err != nil || costs[0][1] != want {
			t.Fatal(strategy, costs, err)
		}
	}
	optimal, err := pla.FindOptimalPairing(EspressoPairFull, options)
	if err != nil || len(optimal.Pairing.Pairs) != 1 || len(optimal.PLA.On.sets) != 1 || espressoPairedTruth(t, pla, optimal, 0) != 0b1111 {
		t.Fatal(optimal.Pairing, espressoFamilyMasks(t, optimal.PLA.On), err)
	}
	for _, strategy := range []EspressoPairOptimizationStrategy{EspressoPairOptimizeHeuristic, EspressoPairOptimizeExact, EspressoPairOptimizePhase} {
		result, err := pla.OptimizePairing(strategy, options)
		if err != nil || result.Evaluated != 1 {
			t.Fatal(strategy, result, err)
		}
		want := uint64(0b1111)
		if len(result.PositivePhase) != 0 && !result.PositivePhase[0] {
			want = 0
		}
		if got := espressoPairedTruth(t, pla, result, 0); got != want {
			t.Fatal(strategy, got, want, result.PositivePhase)
		}
	}
}

func TestEspressoPairErrorsLimitsAndCancellation(t *testing.T) {
	pla := espressoPLAFromOutputTruths(t, 3, []uint64{0xaa})
	bad := []EspressoPairing{
		{Pairs: []EspressoVariablePair{{First: 0, Second: 0}}},
		{Pairs: []EspressoVariablePair{{First: 0, Second: 3}}},
		{Pairs: []EspressoVariablePair{{First: 0, Second: 1}, {First: 1, Second: 2}}},
	}
	for _, pairing := range bad {
		if result, err := pla.PairVariables(pairing, false, EspressoPairOptions{}); result.PLA != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(pairing, result, err)
		}
	}
	if result, err := pla.PairVariables(EspressoPairing{Pairs: []EspressoVariablePair{{First: 0, Second: 1}}}, false, EspressoPairOptions{MaxOperations: 1}); result.PLA != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if pairings, err := EnumeratePairings(4, EspressoPairOptions{MaxPairings: 1}); pairings != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(pairings, err)
	}
	if pairings, err := EnumeratePairings(4, EspressoPairOptions{Stop: func() bool { return true }}); pairings != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(pairings, err)
	}
	if _, _, err := BestPairing([][]int{{0}, {0}}, true, EspressoPairOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	overflow := [][]int{{0, math.MaxInt, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, math.MaxInt}, {0, 0, 0, 0}}
	if _, _, err := BestPairing(overflow, true, EspressoPairOptions{}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if costs, err := pla.PairingCosts(9, EspressoPairOptions{}); costs != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(costs, err)
	}
}

func FuzzEspressoPairVariables(f *testing.F) {
	f.Add([]byte{3, 2, 0x96, 0x69, 0, 2})
	f.Add([]byte{2, 1, 0xf, 0, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 5 {
			return
		}
		variables := int(data[0]%3) + 2
		outputs := int(data[1]%2) + 1
		assignments := 1 << uint(variables)
		universe := uint64(1)<<uint(assignments) - 1
		truths := make([]uint64, outputs)
		for output := range truths {
			truths[output] = uint64(data[(2+output)%len(data)]) & universe
		}
		first := int(data[len(data)-2]) % variables
		second := int(data[len(data)-1]) % (variables - 1)
		if second >= first {
			second++
		}
		if first > second {
			first, second = second, first
		}
		pla := espressoPLAFromOutputTruths(t, variables, truths)
		result, err := pla.PairVariables(EspressoPairing{Pairs: []EspressoVariablePair{{First: first, Second: second}}}, false, EspressoPairOptions{MaxCubes: 1 << 16, MaxOperations: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		for output, truth := range truths {
			if got := espressoPairedTruth(t, pla, result, output); got != truth {
				t.Fatal(variables, outputs, first, second, output, got, truth)
			}
		}
	})
}

func FuzzEspressoBestPairing(f *testing.F) {
	f.Add([]byte{4, 9, 1, 1, 1, 8, 7})
	f.Add([]byte{5, 1, 2, 3, 4, 5})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		variables := int(data[0]%6) + 1
		costs := make([][]int, variables)
		position := 1
		for first := 0; first < variables; first++ {
			costs[first] = make([]int, variables)
			for second := first + 1; second < variables; second++ {
				costs[first][second] = int(data[position%len(data)]%31) - 15
				position++
			}
		}
		got, gotCost, err := BestPairing(costs, true, EspressoPairOptions{MaxPairings: 1 << 14})
		if err != nil {
			t.Fatal(err)
		}
		pairings, err := EnumeratePairings(variables, EspressoPairOptions{MaxPairings: 1 << 14})
		if err != nil {
			t.Fatal(err)
		}
		wantCost := math.MinInt
		for _, pairing := range pairings {
			cost, _ := pairingCost(costs, pairing)
			wantCost = max(wantCost, cost)
		}
		if gotCost != wantCost || len(got.Pairs) != variables/2 {
			t.Fatal(variables, costs, got, gotCost, wantCost)
		}
	})
}
