package abc

// Stand-alone combinatorial and truth-table utilities from src/misc/extra/extraUtilMisc.c.

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"math/bits"
	"sort"
	"strconv"
	"strings"
)

// ExtraBase2LogDouble returns ceil(log2(value)).
func ExtraBase2LogDouble(value float64) (int, error) {
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, ErrDimension
	}
	result := math.Ceil(math.Log2(value))
	if result < math.MinInt || result > math.MaxInt {
		return 0, ErrLimit
	}
	return int(result), nil
}

func ExtraPower2(degree int) (float64, error) {
	if degree < 0 || degree > 1023 {
		return 0, ErrDimension
	}
	return math.Ldexp(1, degree), nil
}

func ExtraPower3(degree int) (int, error) {
	if degree < 0 {
		return 0, ErrDimension
	}
	result := 1
	for range degree {
		if result > math.MaxInt/3 {
			return 0, ErrLimit
		}
		result *= 3
	}
	return result, nil
}

func ExtraFactorial(value int) (int, error) {
	if value < 0 {
		return 0, ErrDimension
	}
	result := 1
	for factor := 2; factor <= value; factor++ {
		if result > math.MaxInt/factor {
			return 0, ErrLimit
		}
		result *= factor
	}
	return result, nil
}

func ExtraCombinationCount(k, n int) (int, error) {
	if k < 0 || n < 0 || k > n {
		return 0, ErrDimension
	}
	if k > n-k {
		k = n - k
	}
	result := uint64(1)
	for i := 1; i <= k; i++ {
		factor := uint64(n - i + 1)
		divisor := uint64(i)
		common := uint64(gcdInt(int(factor), int(divisor)))
		factor, divisor = factor/common, divisor/common
		common = uint64(gcdInt(int(result), int(divisor)))
		result, divisor = result/common, divisor/common
		if divisor != 1 || factor != 0 && result > uint64(math.MaxInt)/factor {
			return 0, ErrLimit
		}
		result *= factor
	}
	return int(result), nil
}

func gcdInt(first, second int) int {
	for second != 0 {
		first, second = second, first%second
	}
	return first
}

func ExtraRadixCode(number, radix, digits int) ([]int, error) {
	if number < 0 || radix < 2 || digits < 0 || digits >= 100 {
		return nil, ErrDimension
	}
	result := make([]int, digits)
	for index := range result {
		result[index] = number % radix
		number /= radix
	}
	if number != 0 {
		return nil, ErrLimit
	}
	return result, nil
}

func ExtraCountOnes(data []byte) int {
	result := 0
	for _, value := range data {
		result += bits.OnesCount8(value)
	}
	return result
}

// ExtraPermutations preserves the recursive section ordering of the source.
func ExtraPermutations(size int) ([][]int, error) {
	if size < 1 || size > 10 {
		return nil, ErrDimension
	}
	count, err := ExtraFactorial(size)
	if err != nil {
		return nil, err
	}
	result := make([][]int, count)
	for index := range result {
		result[index] = make([]int, size)
	}
	current := make([]int, size)
	for index := range current {
		current[index] = index
	}
	var fill func(offset, factorial, remaining int)
	fill = func(offset, factorial, remaining int) {
		if remaining == 1 {
			result[offset][0] = current[0]
			return
		}
		next := factorial / remaining
		last := remaining - 1
		for selected := 0; selected < remaining; selected++ {
			current[selected], current[last] = current[last], current[selected]
			section := offset + (remaining-1-selected)*next
			for row := 0; row < next; row++ {
				result[section+row][last] = current[last]
			}
			fill(section, next, remaining-1)
			current[selected], current[last] = current[last], current[selected]
		}
	}
	fill(0, count, size)
	return result, nil
}

func checkExtraSmallPermutation(permutation []int, variables int) error {
	if variables < 0 || variables > 5 || len(permutation) != variables {
		return ErrDimension
	}
	seen := uint32(0)
	for _, value := range permutation {
		if value < 0 || value >= variables || seen&(1<<uint(value)) != 0 {
			return ErrDimension
		}
		seen |= 1 << uint(value)
	}
	return nil
}

func ExtraPermuteMinterms(minterms, permutation []int, variables int) ([]int, error) {
	if err := checkExtraSmallPermutation(permutation, variables); err != nil {
		return nil, err
	}
	limit := 1 << uint(variables)
	result := make([]int, len(minterms))
	for index, minterm := range minterms {
		if minterm < 0 || minterm >= limit {
			return nil, ErrDimension
		}
		for variable, target := range permutation {
			if minterm&(1<<uint(variable)) != 0 {
				result[index] |= 1 << uint(target)
			}
		}
	}
	return result, nil
}

func ExtraPermuteTruth32(truth uint32, permutation []int, variables int, reverse bool) (uint32, error) {
	if err := checkExtraSmallPermutation(permutation, variables); err != nil {
		return 0, err
	}
	assignments := 1 << uint(variables)
	var result uint32
	for source := 0; source < assignments; source++ {
		target := 0
		for variable, mapped := range permutation {
			if source&(1<<uint(variable)) != 0 {
				target |= 1 << uint(mapped)
			}
		}
		if reverse {
			if truth&(1<<uint(target)) != 0 {
				result |= 1 << uint(source)
			}
		} else if truth&(1<<uint(source)) != 0 {
			result |= 1 << uint(target)
		}
	}
	return result, nil
}

func extraSmallTruthMask(variables int) (uint32, error) {
	if variables < 0 || variables > 5 {
		return 0, ErrDimension
	}
	assignments := 1 << uint(variables)
	if assignments == 32 {
		return ^uint32(0), nil
	}
	return 1<<uint(assignments) - 1, nil
}

func ExtraPolarizeTruth32(truth uint32, phase, variables int) (uint32, error) {
	mask, err := extraSmallTruthMask(variables)
	if err != nil || phase < 0 || phase >= 1<<uint(variables) {
		if err != nil {
			return 0, err
		}
		return 0, ErrDimension
	}
	truth &= mask
	assignments := 1 << uint(variables)
	var result uint32
	for assignment := 0; assignment < assignments; assignment++ {
		if truth&(1<<uint(assignment^phase)) != 0 {
			result |= 1 << uint(assignment)
		}
	}
	return result, nil
}

type ExtraTruthCanonicalMode uint8

const (
	ExtraTruthCanonN ExtraTruthCanonicalMode = iota
	ExtraTruthCanonNN
	ExtraTruthCanonP
	ExtraTruthCanonNP
	ExtraTruthCanonNPN
)

// ExtraCanonicalTruth32 returns the numerical minimum in the selected orbit.
func ExtraCanonicalTruth32(truth uint32, variables int, mode ExtraTruthCanonicalMode) (uint32, error) {
	mask, err := extraSmallTruthMask(variables)
	if err != nil || mode > ExtraTruthCanonNPN {
		if err != nil {
			return 0, err
		}
		return 0, ErrDimension
	}
	truth &= mask
	permutations := [][]int{{}}
	if mode == ExtraTruthCanonP || mode == ExtraTruthCanonNP || mode == ExtraTruthCanonNPN {
		if variables == 0 {
			permutations = [][]int{{}}
		} else if permutations, err = ExtraPermutations(variables); err != nil {
			return 0, err
		}
	}
	phases := 1
	if mode == ExtraTruthCanonN || mode == ExtraTruthCanonNN || mode == ExtraTruthCanonNP || mode == ExtraTruthCanonNPN {
		phases = 1 << uint(variables)
	}
	outputs := 1
	if mode == ExtraTruthCanonNN || mode == ExtraTruthCanonNPN {
		outputs = 2
	}
	best := mask
	for output := 0; output < outputs; output++ {
		base := truth
		if output != 0 {
			base = ^base & mask
		}
		for phase := 0; phase < phases; phase++ {
			candidate, _ := ExtraPolarizeTruth32(base, phase, variables)
			for _, permutation := range permutations {
				permuted := candidate
				if variables > 0 && len(permutation) > 0 {
					permuted, _ = ExtraPermuteTruth32(candidate, permutation, variables, false)
				}
				if permuted < best {
					best = permuted
				}
			}
		}
	}
	return best, nil
}

type ExtraTruth4NPNTable struct {
	Canonical   []uint16
	Phase       []uint8
	Permutation []uint8
	Class       []uint8
	Classes     int
}

// ExtraTruth4NPN builds the complete four-input NPN orbit table.
func ExtraTruth4NPN() (ExtraTruth4NPNTable, error) {
	permutations, _ := ExtraPermutations(4)
	result := ExtraTruth4NPNTable{
		Canonical: make([]uint16, 1<<16), Phase: make([]uint8, 1<<16),
		Permutation: make([]uint8, 1<<16), Class: make([]uint8, 1<<16),
	}
	assigned := make([]bool, 1<<16)
	class := 0
	for representative := 0; representative < 1<<16; representative++ {
		if assigned[representative] {
			continue
		}
		class++
		for phase := 0; phase < 16; phase++ {
			polar, _ := ExtraPolarizeTruth32(uint32(representative), phase, 4)
			for permutationIndex, permutation := range permutations {
				value, _ := ExtraPermuteTruth32(polar, permutation, 4, false)
				for output := 0; output < 2; output++ {
					index := int(value)
					if output != 0 {
						index ^= 0xffff
					}
					if !assigned[index] {
						assigned[index] = true
						result.Canonical[index] = uint16(representative)
						result.Phase[index] = uint8(phase | output<<4)
						result.Permutation[index] = uint8(permutationIndex)
					}
					result.Class[index] = uint8(class - 1)
				}
			}
		}
	}
	result.Classes = class
	if class != 222 {
		return ExtraTruth4NPNTable{}, ErrInvariant
	}
	return result, nil
}

type ExtraTruthNPhaseTable struct {
	Variables int
	Canonical []uint16
	Phases    [][]uint8
	Classes   int
}

func ExtraTruthNPhases(variables, maxPhases int) (ExtraTruthNPhaseTable, error) {
	if (variables != 3 && variables != 4) || maxPhases < 1 || maxPhases > 1<<uint(variables) {
		return ExtraTruthNPhaseTable{}, ErrDimension
	}
	functions := 1 << uint(1<<uint(variables))
	result := ExtraTruthNPhaseTable{Variables: variables, Canonical: make([]uint16, functions), Phases: make([][]uint8, functions)}
	assigned := make([]bool, functions)
	for representative := 0; representative < functions; representative++ {
		if assigned[representative] {
			continue
		}
		result.Classes++
		for phase := 0; phase < 1<<uint(variables); phase++ {
			value, _ := ExtraPolarizeTruth32(uint32(representative), phase, variables)
			index := int(value)
			if !assigned[index] {
				assigned[index] = true
				result.Canonical[index] = uint16(representative)
			}
			if len(result.Phases[index]) < maxPhases {
				result.Phases[index] = append(result.Phases[index], uint8(phase))
			}
		}
	}
	return result, nil
}

func extraSupportPermutation(variables int, phase uint32) ([]int, int, bool, error) {
	if variables < 1 || variables > 8 || phase >= 1<<uint(variables) {
		return nil, 0, false, ErrDimension
	}
	ones := bits.OnesCount32(phase)
	if phase == 0 || phase == 1<<uint(ones)-1 {
		return nil, 0, true, nil
	}
	if ones == 1 {
		return nil, bits.TrailingZeros32(phase), false, nil
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
	return permutation, -1, false, nil
}

// ExtraExpandTruth compacts the variables selected by phase into the low positions.
func ExtraExpandTruth(words []uint32, variables int, phase uint32) ([]uint32, error) {
	wordCount := 1
	if variables > 5 {
		wordCount = 1 << uint(variables-5)
	}
	if variables < 1 || variables > 8 || len(words) != wordCount {
		return nil, ErrDimension
	}
	permutation, singleton, unchanged, err := extraSupportPermutation(variables, phase)
	if err != nil {
		return nil, err
	}
	result := make([]uint32, wordCount)
	if unchanged {
		copy(result, words)
		return result, nil
	}
	assignments := 1 << uint(variables)
	if singleton >= 0 {
		for assignment := 0; assignment < assignments; assignment++ {
			if assignment&(1<<uint(singleton)) != 0 {
				result[assignment>>5] |= 1 << uint(assignment&31)
			}
		}
		return result, nil
	}
	for source := 0; source < assignments; source++ {
		if words[source>>5]&(1<<uint(source&31)) == 0 {
			continue
		}
		target := 0
		for output, input := range permutation {
			if source&(1<<uint(input)) != 0 {
				target |= 1 << uint(output)
			}
		}
		result[target>>5] |= 1 << uint(target&31)
	}
	return result, nil
}

func ExtraTruthVariables8() [8][8]uint32 {
	var result [8][8]uint32
	for variable := range result {
		for assignment := 0; assignment < 256; assignment++ {
			if assignment&(1<<uint(variable)) != 0 {
				result[variable][assignment>>5] |= 1 << uint(assignment&31)
			}
		}
	}
	return result
}

// ExtraBubbleSortOrder stably orders indices by costs.
func ExtraBubbleSortOrder(order, costs []int, increasing bool) ([]int, error) {
	if len(order) == 0 {
		return []int{}, nil
	}
	if len(costs) == 0 {
		return nil, ErrDimension
	}
	result := append([]int(nil), order...)
	for _, index := range result {
		if index < 0 || index >= len(costs) {
			return nil, ErrDimension
		}
	}
	sort.SliceStable(result, func(first, second int) bool {
		if increasing {
			return costs[result[first]] < costs[result[second]]
		}
		return costs[result[first]] > costs[result[second]]
	})
	return result, nil
}

func ExtraTruth6MinimumExact(truth uint64) (uint64, error) {
	gray, err := ExtraGrayCodeSchedule(6)
	if err != nil {
		return 0, err
	}
	permutations, err := ExtraPermutationSchedule(6)
	if err != nil {
		return 0, err
	}
	minimum := ^uint64(0)
	for output := 0; output < 2; output++ {
		current := truth
		if output != 0 {
			current = ^current
		}
		for _, swap := range permutations {
			for _, variable := range gray {
				if current < minimum {
					minimum = current
				}
				current = extraTruth6Flip(current, variable)
			}
			current = extraTruth6SwapAdjacent(current, swap)
		}
	}
	return minimum, nil
}

func extraTruth6MinimumRoundOne(truth uint64, variable int) uint64 {
	minimum := truth
	first := extraTruth6Flip(truth, variable)
	if first < minimum {
		minimum = first
	}
	second := extraTruth6Flip(first, variable+1)
	if second < minimum {
		minimum = second
	}
	both := extraTruth6Flip(second, variable)
	if both < minimum {
		minimum = both
	}
	swapped := extraTruth6SwapAdjacent(truth, variable)
	if swapped < minimum {
		minimum = swapped
	}
	first = extraTruth6Flip(swapped, variable)
	if first < minimum {
		minimum = first
	}
	second = extraTruth6Flip(first, variable+1)
	if second < minimum {
		minimum = second
	}
	both = extraTruth6Flip(second, variable)
	if both < minimum {
		minimum = both
	}
	return minimum
}

func extraTruth6MinimumRounds(truth uint64) uint64 {
	minimum := truth
	for range 10 {
		before := minimum
		for variable := 4; variable >= 0; variable-- {
			candidate := extraTruth6MinimumRoundOne(minimum, variable)
			if candidate < minimum {
				minimum = candidate
			}
		}
		if minimum == before {
			break
		}
	}
	return minimum
}

func ExtraTruth6MinimumHeuristic(truth uint64) uint64 {
	ones := bits.OnesCount64(truth)
	if ones < 32 {
		return extraTruth6MinimumRounds(truth)
	}
	if ones > 32 {
		return extraTruth6MinimumRounds(^truth)
	}
	first, second := extraTruth6MinimumRounds(truth), extraTruth6MinimumRounds(^truth)
	if second < first {
		return second
	}
	return first
}

func ParseExtraNPNFunctions(data []byte, expected int) ([]uint64, error) {
	if expected < 0 {
		return nil, ErrDimension
	}
	result := make([]uint64, 0, expected)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			return nil, ErrFormat
		}
		line = strings.TrimPrefix(strings.TrimPrefix(line, "0x"), "0X")
		if len(line) != 16 {
			return nil, ErrFormat
		}
		value, err := strconv.ParseUint(line, 16, 64)
		if err != nil {
			return nil, ErrFormat
		}
		result = append(result, value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(result) != expected {
		return nil, ErrFormat
	}
	return result, nil
}

func FormatExtraBits(value uint64, width int) (string, error) {
	if width < 0 || width > 64 {
		return "", ErrDimension
	}
	var builder strings.Builder
	builder.Grow(width)
	for bit := width - 1; bit >= 0; bit-- {
		builder.WriteByte(byte('0' + value>>uint(bit)&1))
	}
	return builder.String(), nil
}

type ExtraPowerRow struct {
	Base, Exponent int
	Value          uint64
}

func ExtraPowerRows(bitsWide int, maxRows int) ([]ExtraPowerRow, error) {
	if bitsWide < 1 || bitsWide > 16 || maxRows < 0 {
		return nil, ErrDimension
	}
	count := 1 << uint(bitsWide)
	if maxRows == 0 || maxRows > count*count {
		maxRows = count * count
	}
	rows := make([]ExtraPowerRow, 0, maxRows)
	for base := 0; base < count && len(rows) < maxRows; base++ {
		for exponent := 0; exponent < count && len(rows) < maxRows; exponent++ {
			value := uint64(base)
			for factor := 1; factor < exponent; factor++ {
				value *= uint64(base)
			}
			rows = append(rows, ExtraPowerRow{base, exponent, value})
		}
	}
	return rows, nil
}

func ExtraTranspose32(matrix *[32]uint32) error {
	if matrix == nil {
		return ErrDimension
	}
	mask := uint32(0x0000ffff)
	for width := uint(16); width != 0; {
		for k := uint(0); k < 32; k = ((k | width) + 1) &^ width {
			temporary := (matrix[k] ^ matrix[k|width]>>width) & mask
			matrix[k] ^= temporary
			matrix[k|width] ^= temporary << width
		}
		width >>= 1
		mask ^= mask << width
	}
	return nil
}

// ExtraTransposeBitMatrix transposes a packed row-major matrix. Input has
// 64*outWords rows of inWords words; output has 64*inWords rows of outWords.
func ExtraTransposeBitMatrix(input []uint64, inWords, outWords int) ([]uint64, error) {
	if inWords < 1 || outWords < 1 || len(input) != 64*inWords*outWords {
		return nil, ErrDimension
	}
	output := make([]uint64, len(input))
	for row := 0; row < 64*outWords; row++ {
		for column := 0; column < 64*inWords; column++ {
			if input[row*inWords+column/64]&(uint64(1)<<uint(column&63)) != 0 {
				output[column*outWords+row/64] |= uint64(1) << uint(row&63)
			}
		}
	}
	return output, nil
}

func ExtraTransposeBitMatrixRows(input [][]uint64, inWords, outWords int) ([]uint64, error) {
	if inWords < 1 || outWords < 1 || len(input) != 64*outWords {
		return nil, ErrDimension
	}
	flat := make([]uint64, 0, 64*inWords*outWords)
	for _, row := range input {
		if len(row) != inWords {
			return nil, ErrDimension
		}
		flat = append(flat, row...)
	}
	return ExtraTransposeBitMatrix(flat, inWords, outWords)
}

func FormatExtraBitMatrix(words []uint64, rowWords int) (string, error) {
	if rowWords < 1 || len(words)%rowWords != 0 {
		return "", ErrDimension
	}
	var builder strings.Builder
	for row := 0; row < len(words)/rowWords; row++ {
		if row%64 == 0 {
			builder.WriteByte('\n')
		}
		for word := 0; word < rowWords; word++ {
			if word != 0 {
				builder.WriteByte(' ')
			}
			fmt.Fprintf(&builder, "%064b", words[row*rowWords+word])
		}
		builder.WriteByte('\n')
	}
	builder.WriteByte('\n')
	return builder.String(), nil
}
