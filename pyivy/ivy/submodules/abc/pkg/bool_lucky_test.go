package abc

// Go-only tests for src/bool/lucky.

import (
	"errors"
	"math/bits"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func luckyTestTable(t testing.TB, variables int, words []uint64) *TruthTable {
	t.Helper()
	if variables < 6 {
		var err error
		words = slices.Clone(words)
		words[0], err = NormalizeSmallTruthWord(words[0], variables)
		if err != nil {
			t.Fatal(err)
		}
	}
	table, err := TruthTableFromWords(variables, words, len(words))
	if err != nil {
		t.Fatal(err)
	}
	return table
}

func luckyModelResult(source *TruthTable, result LuckyCanonicalization) []uint64 {
	words := make([]uint64, len(source.words))
	assignments := 1 << uint(source.variables)
	for target := 0; target < assignments; target++ {
		sourceAssignment := 0
		for position, sourceVariable := range result.Permutation {
			value := target>>uint(position)&1 != 0
			if result.InputPhases[position] {
				value = !value
			}
			if value {
				sourceAssignment |= 1 << uint(sourceVariable)
			}
		}
		value := source.words[sourceAssignment>>6]>>uint(sourceAssignment&63)&1 != 0
		if result.OutputPhase {
			value = !value
		}
		if value {
			words[target>>6] |= uint64(1) << uint(target&63)
		}
	}
	if source.variables < 6 {
		words[0], _ = NormalizeSmallTruthWord(words[0], source.variables)
	}
	return words
}

func TestLuckyExhaustiveSmallTruths(t *testing.T) {
	for variables := 0; variables <= 3; variables++ {
		functions := 1 << uint(1<<uint(variables))
		for logical := 0; logical < functions; logical++ {
			source := luckyTestTable(t, variables, []uint64{uint64(logical)})
			result, err := source.LuckyCanonicalize(LuckyOptions{})
			if err != nil {
				t.Fatalf("variables=%d truth=%x: %v", variables, logical, err)
			}
			if valid, verifyErr := VerifyLuckyCanonicalization(source, result); verifyErr != nil || !valid {
				t.Fatalf("variables=%d truth=%x valid=%v err=%v result=%+v", variables, logical, valid, verifyErr, result)
			}
			if want := luckyModelResult(source, result); !slices.Equal(result.Table.words, want) {
				t.Fatalf("variables=%d truth=%x got=%x want=%x", variables, logical, result.Table.words, want)
			}
			again, err := source.LuckyCanonicalize(LuckyOptions{})
			if err != nil || !slices.Equal(result.Table.words, again.Table.words) || !slices.Equal(result.Permutation, again.Permutation) || !slices.Equal(result.InputPhases, again.InputPhases) || result.OutputPhase != again.OutputPhase {
				t.Fatalf("nondeterministic variables=%d truth=%x first=%+v second=%+v err=%v", variables, logical, result, again, err)
			}
			copyResult := result.Clone()
			copyResult.Table.words[0] ^= 1
			if variables > 0 {
				copyResult.Permutation[0] ^= 1
				copyResult.InputPhases[0] = !copyResult.InputPhases[0]
			}
			if slices.Equal(copyResult.Table.words, result.Table.words) || variables > 0 && (slices.Equal(copyResult.Permutation, result.Permutation) || slices.Equal(copyResult.InputPhases, result.InputPhases)) {
				t.Fatal("result clone aliases owned storage")
			}
		}
	}
}

func TestLuckyIndependentPermutationModelCampaign(t *testing.T) {
	state := uint64(0xd1b54a32d192ed03)
	for execution := 0; execution < 6000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		variables := 2 + int(state%9)
		words := make([]uint64, 1<<uint(max(0, variables-6)))
		for i := range words {
			state ^= state << 13
			state ^= state >> 7
			state ^= state << 17
			words[i] = state
		}
		source := luckyTestTable(t, variables, words)
		result, err := source.LuckyCanonicalize(LuckyOptions{Thorough: execution&1 != 0})
		if err != nil {
			t.Fatalf("execution=%d: %v", execution, err)
		}
		if want := luckyModelResult(source, result); !slices.Equal(result.Table.words, want) {
			t.Fatalf("execution=%d variables=%d metadata mismatch", execution, variables)
		}
		if valid, verifyErr := VerifyLuckyCanonicalization(source, result); verifyErr != nil || !valid {
			t.Fatalf("execution=%d valid=%v err=%v", execution, valid, verifyErr)
		}
		if len(result.Permutation) != variables || len(result.InputPhases) != variables || result.Operations < 0 {
			t.Fatalf("execution=%d malformed result %+v", execution, result)
		}
	}
}

func TestLuckySixAndSixteenVariablePaths(t *testing.T) {
	for _, variables := range []int{6, 7, 12, 16} {
		words := make([]uint64, 1<<uint(max(0, variables-6)))
		for i := range words {
			words[i] = uint64(i)*0x9e3779b97f4a7c15 ^ 0x6996966996696996
		}
		source := luckyTestTable(t, variables, words)
		result, err := source.LuckyCanonicalize(LuckyOptions{})
		if err != nil {
			t.Fatal(variables, err)
		}
		if valid, err := VerifyLuckyCanonicalization(source, result); err != nil || !valid {
			t.Fatal(variables, valid, err)
		}
	}
}

func TestLuckyStoreRoundTripAndOwnership(t *testing.T) {
	input := "0x6 trailing\r\n9\r\n\r\nF\r\n"
	store, err := ParseLuckyTruthStore(input, 2)
	if err != nil || store.Variables() != 2 || len(store.Tables()) != 2 {
		t.Fatal(store, err)
	}
	formatted, err := store.Format()
	if err != nil || formatted != "6\n9\n" {
		t.Fatalf("format=%q err=%v", formatted, err)
	}
	tables := store.Tables()
	tables[0].words[0] = 0
	if store.Tables()[0].words[0] == 0 {
		t.Fatal("store accessor aliases storage")
	}
	path := filepath.Join(t.TempDir(), "truths.hex")
	if err = store.Write(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := ReadLuckyTruthStore(path, 2)
	if err != nil {
		t.Fatal(err)
	}
	loadedText, _ := loaded.Format()
	if loadedText != formatted {
		t.Fatalf("loaded=%q want=%q", loadedText, formatted)
	}
	if _, err = os.Stat(path); err != nil {
		t.Fatal(err)
	}
	for _, malformed := range []string{"", "XYZ\n", "000\n", "0\n00\n"} {
		if parsed, parseErr := ParseLuckyTruthStore(malformed, 0); parsed != nil || parseErr == nil {
			t.Fatalf("malformed=%q parsed=%v err=%v", malformed, parsed, parseErr)
		}
	}
	if _, err = ParseLuckyTruthStore("0\n1\n", 1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	for _, nonNative := range []string{"\n6\n", "0X6\n", "6\ttrailing\n", "6\n   \n"} {
		if parsed, parseErr := ParseLuckyTruthStore(nonNative, 0); parsed != nil || parseErr == nil {
			t.Fatalf("non-native syntax %q parsed=%v err=%v", nonNative, parsed, parseErr)
		}
	}
}

func luckyTestFast1Passes(t testing.TB, source *TruthTable) (luckyState, luckyState) {
	t.Helper()
	budget, err := newLuckyBudget(LuckyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	state := newLuckyState(source)
	counts, inputTied, outputTied, err := luckyInitial(&state)
	if err != nil {
		t.Fatal(err)
	}
	initial := state.clone()
	state, err = luckyRefine(state, counts, inputTied, budget)
	if err != nil {
		t.Fatal(err)
	}
	if outputTied {
		complement := initial.clone()
		complement.complement()
		complement, err = luckyRefine(complement, counts, inputTied, budget)
		if err != nil {
			t.Fatal(err)
		}
		if luckyLess(complement, state) {
			state = complement
		}
	}
	first := state.clone()
	luckyInvertAllInputs(&state)
	inverted := state.clone()
	state, err = luckyRefine(state, counts, inputTied, budget)
	if err != nil {
		t.Fatal(err)
	}
	if outputTied {
		complement := inverted.clone()
		complement.complement()
		complement, err = luckyRefine(complement, counts, inputTied, budget)
		if err != nil {
			t.Fatal(err)
		}
		if luckyLess(complement, state) {
			state = complement
		}
	}
	return first, state
}

func TestLuckyFast1ReturnsSecondBasin(t *testing.T) {
	state := uint64(0x243f6a8885a308d3)
	found := false
	for trial := 0; trial < 20000 && !found; trial++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		source := luckyTestTable(t, 6, []uint64{state})
		first, second := luckyTestFast1Passes(t, source)
		if !luckyLess(first, second) {
			continue
		}
		found = true
		result, err := source.LuckyCanonicalize(LuckyOptions{Thorough: true})
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(result.Table.words, second.table.words) || !slices.Equal(result.Permutation, second.permutation) || !slices.Equal(result.InputPhases, second.phases) || result.OutputPhase != second.output {
			t.Fatalf("fast1 retained first basin: first=%x second=%x result=%x", first.table.words, second.table.words, result.Table.words)
		}
	}
	if !found {
		t.Fatal("campaign did not find distinct fast1 basins")
	}
}

func TestLuckyIndependentPairTransformModel5000(t *testing.T) {
	state := uint64(0x13198a2e03707344)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 13
		state ^= state >> 7
		state ^= state << 17
		variables := 2 + execution%9
		words := make([]uint64, 1<<uint(max(0, variables-6)))
		for index := range words {
			state ^= state << 7
			state ^= state >> 9
			words[index] = state
		}
		source := luckyTestTable(t, variables, words)
		position := execution % (variables - 1)
		allowPhases := execution&1 != 0
		transforms := [][3]bool{{}}
		if allowPhases {
			transforms = append(transforms, [3]bool{false, true, false}, [3]bool{false, false, true}, [3]bool{false, true, true})
		}
		transforms = append(transforms, [3]bool{true, false, false})
		if allowPhases {
			transforms = append(transforms, [3]bool{true, true, false}, [3]bool{true, false, true}, [3]bool{true, true, true})
		}
		bestWords := slices.Clone(source.words)
		bestPermutation := make([]int, variables)
		for index := range bestPermutation {
			bestPermutation[index] = index
		}
		bestPhases := make([]bool, variables)
		for transformIndex, transform := range transforms[1:] {
			permutation := make([]int, variables)
			for index := range permutation {
				permutation[index] = index
			}
			phases := make([]bool, variables)
			if transform[0] {
				permutation[position], permutation[position+1] = permutation[position+1], permutation[position]
			}
			phases[position], phases[position+1] = transform[1], transform[2]
			candidate := luckyModelResult(source, LuckyCanonicalization{Permutation: permutation, InputPhases: phases})
			if truthWordsLessReverse(candidate, bestWords) {
				bestWords, bestPermutation, bestPhases = candidate, permutation, phases
			}
			_ = transformIndex
		}
		budget, err := newLuckyBudget(LuckyOptions{})
		if err != nil {
			t.Fatal(err)
		}
		actual, err := luckyRefinePair(newLuckyState(source), position, allowPhases, budget)
		if err != nil || !slices.Equal(actual.table.words, bestWords) || !slices.Equal(actual.permutation, bestPermutation) || !slices.Equal(actual.phases, bestPhases) {
			t.Fatalf("execution=%d pair=%d phases=%v actual=%x/%v/%v expected=%x/%v/%v err=%v", execution, position, allowPhases, actual.table.words, actual.permutation, actual.phases, bestWords, bestPermutation, bestPhases, err)
		}
	}
}

func TestLuckyLimitsCancellationAndErrors(t *testing.T) {
	source := luckyTestTable(t, 6, []uint64{0x6996966996696996})
	if result, err := source.LuckyCanonicalize(LuckyOptions{MaxOperations: -1}); err == nil || result.Table != nil {
		t.Fatal(result, err)
	}
	if result, err := source.LuckyCanonicalize(LuckyOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) || result.Table != nil {
		t.Fatal(result, err)
	}
	if result, err := source.LuckyCanonicalize(LuckyOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) || result.Table != nil || result.Permutation != nil || result.InputPhases != nil {
		t.Fatal(result, err)
	}
	wide := &TruthTable{variables: LuckyMaximumVariables + 1, limit: 1, words: make([]uint64, 1<<(LuckyMaximumVariables+1-6))}
	if result, err := wide.LuckyCanonicalize(LuckyOptions{}); !errors.Is(err, ErrDimension) || result.Table != nil {
		t.Fatal(result, err)
	}
	if valid, err := VerifyLuckyCanonicalization(source, LuckyCanonicalization{}); valid || !errors.Is(err, ErrDimension) {
		t.Fatal(valid, err)
	}
}

func FuzzLuckyCanonicalize(f *testing.F) {
	f.Add(uint64(0), uint64(0), uint8(3), false)
	f.Add(uint64(0x6996966996696996), uint64(0x0123456789abcdef), uint8(6), true)
	f.Add(^uint64(0), uint64(1), uint8(10), false)
	f.Fuzz(func(t *testing.T, first, second uint64, rawVariables uint8, thorough bool) {
		variables := 2 + int(rawVariables%9)
		words := make([]uint64, 1<<uint(max(0, variables-6)))
		state := first ^ bits.RotateLeft64(second, 17)
		for i := range words {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			words[i] = state
		}
		source := luckyTestTable(t, variables, words)
		result, err := source.LuckyCanonicalize(LuckyOptions{Thorough: thorough})
		if err != nil {
			t.Fatal(err)
		}
		want := luckyModelResult(source, result)
		if !slices.Equal(result.Table.words, want) {
			t.Fatalf("metadata mismatch vars=%d result=%+v", variables, result)
		}
		valid, err := VerifyLuckyCanonicalization(source, result)
		if err != nil || !valid {
			t.Fatal(valid, err)
		}
		formattedStore := &LuckyTruthStore{variables: variables, tables: []*TruthTable{source}}
		formatted, err := formattedStore.Format()
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseLuckyTruthStore(strings.TrimSpace(formatted), 1)
		if err != nil || !slices.Equal(parsed.tables[0].words, source.words) {
			t.Fatal(err)
		}
	})
}
