package abc

// Go-only models, exhaustive checks, and fuzz tests for extraUtilMisc.c behavior.

import (
	"errors"
	"math/bits"
	"slices"
	"strings"
	"testing"
)

func TestExtraMiscArithmeticRadixAndPermutations(t *testing.T) {
	logs := []struct {
		value float64
		want  int
	}{{1, 0}, {2, 1}, {3, 2}, {4, 2}, {4.01, 3}}
	for _, row := range logs {
		got, err := ExtraBase2LogDouble(row.value)
		if err != nil || got != row.want {
			t.Fatalf("log2(%g) = %d %v want %d", row.value, got, err, row.want)
		}
	}
	if value, err := ExtraPower2(40); err != nil || value != 1<<40 {
		t.Fatal(value, err)
	}
	if value, err := ExtraPower3(10); err != nil || value != 59049 {
		t.Fatal(value, err)
	}
	if value, err := ExtraFactorial(10); err != nil || value != 3628800 {
		t.Fatal(value, err)
	}
	if value, err := ExtraCombinationCount(4, 10); err != nil || value != 210 {
		t.Fatal(value, err)
	}
	code, err := ExtraRadixCode(123, 5, 4)
	if err != nil || !slices.Equal(code, []int{3, 4, 4, 0}) {
		t.Fatal(code, err)
	}
	code[0] = 99
	again, _ := ExtraRadixCode(123, 5, 4)
	if again[0] != 3 || ExtraCountOnes([]byte{0, 1, 3, 0xff}) != 11 {
		t.Fatal("ownership or popcount", again)
	}
	permutations, err := ExtraPermutations(3)
	want := [][]int{{0, 1, 2}, {1, 0, 2}, {0, 2, 1}, {2, 0, 1}, {2, 1, 0}, {1, 2, 0}}
	if err != nil || !slices.EqualFunc(permutations, want, slices.Equal[[]int]) {
		t.Fatal(permutations, err)
	}
	seen := map[[3]int]bool{}
	for _, permutation := range permutations {
		seen[[3]int(permutation)] = true
	}
	if len(seen) != 6 {
		t.Fatal("duplicate permutation", permutations)
	}
	for _, call := range []func() error{
		func() error { _, e := ExtraBase2LogDouble(0); return e },
		func() error { _, e := ExtraPower2(-1); return e },
		func() error { _, e := ExtraPower3(-1); return e },
		func() error { _, e := ExtraFactorial(-1); return e },
		func() error { _, e := ExtraCombinationCount(3, 2); return e },
		func() error { _, e := ExtraRadixCode(10, 1, 2); return e },
		func() error { _, e := ExtraPermutations(0); return e },
	} {
		if err = call(); !errors.Is(err, ErrDimension) {
			t.Fatal(err)
		}
	}
}

func extraMiscPermuteModel(truth uint32, permutation []int, variables int) uint32 {
	var result uint32
	for source := 0; source < 1<<uint(variables); source++ {
		if truth&(1<<uint(source)) == 0 {
			continue
		}
		target := 0
		for sourceVariable, targetVariable := range permutation {
			target |= (source >> uint(sourceVariable) & 1) << uint(targetVariable)
		}
		result |= 1 << uint(target)
	}
	return result
}

func TestExtraMiscTruthPermutationPolarizationAndCanonicalization(t *testing.T) {
	permutation := []int{2, 0, 3, 1}
	minterms := []int{0, 1, 2, 5, 15}
	gotMinterms, err := ExtraPermuteMinterms(minterms, permutation, 4)
	if err != nil || !slices.Equal(gotMinterms, []int{0, 4, 1, 12, 15}) {
		t.Fatal(gotMinterms, err)
	}
	for truth := 0; truth < 1<<16; truth += 137 {
		got, permErr := ExtraPermuteTruth32(uint32(truth), permutation, 4, false)
		want := extraMiscPermuteModel(uint32(truth), permutation, 4)
		if permErr != nil || got != want {
			t.Fatal(truth, got, want, permErr)
		}
		reverse, reverseErr := ExtraPermuteTruth32(got, permutation, 4, true)
		if reverseErr != nil || reverse != uint32(truth) {
			t.Fatal("reverse", truth, reverse, reverseErr)
		}
		for phase := 0; phase < 16; phase++ {
			polar, polarErr := ExtraPolarizeTruth32(uint32(truth), phase, 4)
			var model uint32
			for assignment := 0; assignment < 16; assignment++ {
				model |= (uint32(truth) >> uint(assignment^phase) & 1) << uint(assignment)
			}
			if polarErr != nil || polar != model {
				t.Fatal("polar", truth, phase, polar, model, polarErr)
			}
		}
	}
	truth := uint32(0x963c)
	for mode := ExtraTruthCanonN; mode <= ExtraTruthCanonNPN; mode++ {
		minimum, canonErr := ExtraCanonicalTruth32(truth, 4, mode)
		if canonErr != nil {
			t.Fatal(mode, canonErr)
		}
		permutations := [][]int{{0, 1, 2, 3}}
		if mode == ExtraTruthCanonP || mode == ExtraTruthCanonNP || mode == ExtraTruthCanonNPN {
			permutations, _ = ExtraPermutations(4)
		}
		outputs := 1
		if mode == ExtraTruthCanonNN || mode == ExtraTruthCanonNPN {
			outputs = 2
		}
		phases := 1
		if mode == ExtraTruthCanonN || mode == ExtraTruthCanonNN || mode == ExtraTruthCanonNP || mode == ExtraTruthCanonNPN {
			phases = 16
		}
		modelMinimum := uint32(0xffff)
		for output := 0; output < outputs; output++ {
			base := truth
			if output != 0 {
				base = ^base & 0xffff
			}
			for phase := 0; phase < phases; phase++ {
				polar, _ := ExtraPolarizeTruth32(base, phase, 4)
				for _, candidatePermutation := range permutations {
					candidate := extraMiscPermuteModel(polar, candidatePermutation, 4)
					if candidate < modelMinimum {
						modelMinimum = candidate
					}
				}
			}
		}
		if minimum != modelMinimum {
			t.Fatal("canonical minimum", mode, minimum, modelMinimum)
		}
	}
	if _, err = ExtraPermuteTruth32(0, []int{0, 0}, 2, false); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = ExtraPolarizeTruth32(0, 8, 3); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestExtraMiscCanonicalTablesExhaustive(t *testing.T) {
	phase3, err := ExtraTruthNPhases(3, 8)
	if err != nil || phase3.Classes != 46 || len(phase3.Canonical) != 256 {
		t.Fatal(phase3.Classes, err)
	}
	for truth := 0; truth < 256; truth++ {
		minimum, _ := ExtraCanonicalTruth32(uint32(truth), 3, ExtraTruthCanonN)
		if phase3.Canonical[truth] != uint16(minimum) {
			t.Fatal("3N", truth, phase3.Canonical[truth], minimum)
		}
		for _, phase := range phase3.Phases[truth] {
			got, _ := ExtraPolarizeTruth32(uint32(phase3.Canonical[truth]), int(phase), 3)
			if got != uint32(truth) {
				t.Fatal("3N replay", truth, phase, got)
			}
		}
	}
	phase4, err := ExtraTruthNPhases(4, 3)
	if err != nil || len(phase4.Canonical) != 1<<16 {
		t.Fatal(phase4.Classes, err)
	}
	classes := map[uint16]bool{}
	for truth := 0; truth < 1<<16; truth++ {
		minimum, _ := ExtraCanonicalTruth32(uint32(truth), 4, ExtraTruthCanonN)
		if phase4.Canonical[truth] != uint16(minimum) || len(phase4.Phases[truth]) < 1 || len(phase4.Phases[truth]) > 3 {
			t.Fatal("4N", truth, phase4.Canonical[truth], minimum, phase4.Phases[truth])
		}
		classes[phase4.Canonical[truth]] = true
		for _, phase := range phase4.Phases[truth] {
			got, _ := ExtraPolarizeTruth32(uint32(phase4.Canonical[truth]), int(phase), 4)
			if got != uint32(truth) {
				t.Fatal("4N replay", truth, phase, got)
			}
		}
	}
	if len(classes) != phase4.Classes {
		t.Fatal("4N classes", len(classes), phase4.Classes)
	}
	npn, err := ExtraTruth4NPN()
	if err != nil || npn.Classes != 222 || len(npn.Canonical) != 1<<16 {
		t.Fatal(npn.Classes, err)
	}
	permutations, _ := ExtraPermutations(4)
	classSeen := make([]bool, npn.Classes)
	for truth := 0; truth < 1<<16; truth++ {
		representative := uint32(npn.Canonical[truth])
		phase := int(npn.Phase[truth])
		if phase&16 != 0 {
			representative = ^representative & 0xffff
		}
		polar, _ := ExtraPolarizeTruth32(representative, phase&15, 4)
		replayed := extraMiscPermuteModel(polar, permutations[npn.Permutation[truth]], 4)
		if replayed != uint32(truth) {
			t.Fatal("4NPN replay", truth, replayed, npn.Canonical[truth], npn.Phase[truth], npn.Permutation[truth])
		}
		classSeen[npn.Class[truth]] = true
	}
	for class, seen := range classSeen {
		if !seen {
			t.Fatal("missing class", class)
		}
	}
	npn.Canonical[0] = 99
	again, _ := ExtraTruth4NPN()
	if again.Canonical[0] != 0 {
		t.Fatal("table ownership")
	}
}

func extraMiscExpandModel(words []uint32, variables int, phase uint32) []uint32 {
	result := make([]uint32, len(words))
	ones := bits.OnesCount32(phase)
	if phase == 0 || phase == 1<<uint(ones)-1 {
		copy(result, words)
		return result
	}
	if ones == 1 {
		variable := bits.TrailingZeros32(phase)
		for assignment := 0; assignment < 1<<uint(variables); assignment++ {
			if assignment&(1<<uint(variable)) != 0 {
				result[assignment>>5] |= 1 << uint(assignment&31)
			}
		}
		return result
	}
	permutation := make([]int, variables)
	one, zero := 0, ones
	for variable := 0; variable < variables; variable++ {
		if phase&(1<<uint(variable)) != 0 {
			permutation[variable] = one
			one++
		} else {
			permutation[variable] = zero
			zero++
		}
	}
	for source := 0; source < 1<<uint(variables); source++ {
		if words[source>>5]&(1<<uint(source&31)) == 0 {
			continue
		}
		target := 0
		for targetVariable, sourceVariable := range permutation {
			target |= (source >> uint(sourceVariable) & 1) << uint(targetVariable)
		}
		result[target>>5] |= 1 << uint(target&31)
	}
	return result
}

func TestExtraMiscSupportExpansionVariablesAndSort(t *testing.T) {
	variables := ExtraTruthVariables8()
	for variable := 0; variable < 8; variable++ {
		for assignment := 0; assignment < 256; assignment++ {
			got := variables[variable][assignment>>5] >> uint(assignment&31) & 1
			want := uint32(assignment >> uint(variable) & 1)
			if got != want {
				t.Fatal(variable, assignment, got, want)
			}
		}
	}
	words := []uint32{0x963ca55a, 0x69960ff0, 0x12345678, 0xabcdef01}
	for phase := uint32(0); phase < 128; phase++ {
		got, err := ExtraExpandTruth(words, 7, phase)
		want := extraMiscExpandModel(words, 7, phase)
		if err != nil || !slices.Equal(got, want) {
			t.Fatal(phase, got, want, err)
		}
	}
	order := []int{0, 1, 2, 3, 4}
	costs := []int{3, 1, 3, 2, 1}
	ascending, err := ExtraBubbleSortOrder(order, costs, true)
	if err != nil || !slices.Equal(ascending, []int{1, 4, 3, 0, 2}) {
		t.Fatal(ascending, err)
	}
	descending, _ := ExtraBubbleSortOrder(order, costs, false)
	if !slices.Equal(descending, []int{0, 2, 3, 1, 4}) {
		t.Fatal(descending)
	}
	if _, err = ExtraExpandTruth(nil, 7, 1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func extraMiscTransform6(truth uint64, permutation []int, phase int, output bool) uint64 {
	var result uint64
	for assignment := 0; assignment < 64; assignment++ {
		source := 0
		for targetVariable, sourceVariable := range permutation {
			bit := assignment >> uint(targetVariable) & 1
			bit ^= phase >> uint(sourceVariable) & 1
			source |= bit << uint(sourceVariable)
		}
		value := truth >> uint(source) & 1
		if output {
			value ^= 1
		}
		result |= value << uint(assignment)
	}
	return result
}

func TestExtraMiscTruth6MinimaParserBitsAndPower(t *testing.T) {
	truth := uint64(0xa2222aaa08888000)
	exact, err := ExtraTruth6MinimumExact(truth)
	if err != nil || exact > truth {
		t.Fatal(exact, err)
	}
	heuristic := ExtraTruth6MinimumHeuristic(truth)
	if heuristic < exact || heuristic > truth {
		t.Fatal("heuristic bounds", exact, heuristic, truth)
	}
	permutations, _ := ExtraPermutations(6)
	for index := 0; index < len(permutations); index += 37 {
		transformed := extraMiscTransform6(truth, permutations[index], index&63, index&1 != 0)
		minimum, minimumErr := ExtraTruth6MinimumExact(transformed)
		if minimumErr != nil || minimum != exact {
			t.Fatal("orbit minimum", index, minimum, exact, minimumErr)
		}
	}
	values, err := ParseExtraNPNFunctions([]byte("0x0123456789abcdef\nFEDCBA9876543210\n"), 2)
	if err != nil || !slices.Equal(values, []uint64{0x0123456789abcdef, 0xfedcba9876543210}) {
		t.Fatal(values, err)
	}
	if _, err = ParseExtraNPNFunctions([]byte("1\n"), 1); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
	formatted, err := FormatExtraBits(0x29, 8)
	if err != nil || formatted != "00101001" {
		t.Fatal(formatted, err)
	}
	rows, err := ExtraPowerRows(2, 0)
	if err != nil || len(rows) != 16 || rows[0].Value != 0 || rows[4].Value != 1 || rows[7].Value != 1 {
		t.Fatal(rows, err)
	}
}

func extraMiscTranspose32Model(input [32]uint32) [32]uint32 {
	var result [32]uint32
	for row := 0; row < 32; row++ {
		for column := 0; column < 32; column++ {
			result[31-column] |= (input[row] >> uint(column) & 1) << uint(31-row)
		}
	}
	return result
}

func TestExtraMiscPackedTransposeAndFormatting(t *testing.T) {
	var matrix [32]uint32
	for row := range matrix {
		matrix[row] = uint32(row*0x1020304) ^ 0x963ca55a
	}
	want32 := extraMiscTranspose32Model(matrix)
	if err := ExtraTranspose32(&matrix); err != nil || matrix != want32 {
		t.Fatal(err, matrix[0], want32[0])
	}
	inWords, outWords := 2, 3
	input := make([]uint64, 64*inWords*outWords)
	for index := range input {
		input[index] = uint64(index)*0x9e3779b97f4a7c15 ^ 0x6996966996696996
	}
	output, err := ExtraTransposeBitMatrix(input, inWords, outWords)
	if err != nil || len(output) != len(input) {
		t.Fatal(len(output), err)
	}
	for row := 0; row < 64*outWords; row++ {
		for column := 0; column < 64*inWords; column++ {
			before := input[row*inWords+column/64] >> uint(column&63) & 1
			after := output[column*outWords+row/64] >> uint(row&63) & 1
			if before != after {
				t.Fatal(row, column, before, after)
			}
		}
	}
	rows := make([][]uint64, 64*outWords)
	for row := range rows {
		rows[row] = slices.Clone(input[row*inWords : (row+1)*inWords])
	}
	fromRows, err := ExtraTransposeBitMatrixRows(rows, inWords, outWords)
	if err != nil || !slices.Equal(fromRows, output) {
		t.Fatal(err)
	}
	text, err := FormatExtraBitMatrix([]uint64{1, 2}, 1)
	if err != nil || !strings.Contains(text, "0000000000000000000000000000000000000000000000000000000000000001") {
		t.Fatal(text, err)
	}
	if _, err = ExtraTransposeBitMatrix(input[:1], inWords, outWords); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzExtraMiscTruthAndTransposeModels(f *testing.F) {
	f.Add(uint32(0x963ca55a), uint8(5), uint8(17))
	f.Add(uint32(0), uint8(1), uint8(0))
	f.Fuzz(func(t *testing.T, truth uint32, width, phaseByte uint8) {
		variables := int(width%5) + 1
		mask, _ := extraSmallTruthMask(variables)
		truth &= mask
		phase := int(phaseByte) & ((1 << uint(variables)) - 1)
		polar, err := ExtraPolarizeTruth32(truth, phase, variables)
		if err != nil {
			t.Fatal(err)
		}
		back, err := ExtraPolarizeTruth32(polar, phase, variables)
		if err != nil || back != truth {
			t.Fatal(variables, phase, truth, back, err)
		}
		permutations, err := ExtraPermutations(variables)
		if err != nil {
			t.Fatal(err)
		}
		permutation := permutations[int(phaseByte)%len(permutations)]
		got, err := ExtraPermuteTruth32(truth, permutation, variables, false)
		if err != nil || got != extraMiscPermuteModel(truth, permutation, variables) {
			t.Fatal(variables, truth, permutation, got, err)
		}
		var matrix [32]uint32
		for row := range matrix {
			matrix[row] = bits.RotateLeft32(truth, row) ^ uint32(row)*0x9e3779b9
		}
		want := extraMiscTranspose32Model(matrix)
		if err = ExtraTranspose32(&matrix); err != nil || matrix != want {
			t.Fatal(err)
		}
	})
}
