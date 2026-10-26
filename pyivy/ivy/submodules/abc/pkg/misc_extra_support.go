package abc

// Support minimization from src/misc/extra/extraUtilSupp.c.

import (
	"fmt"
	"math/bits"
	"os"
	"slices"
	"strconv"
	"strings"
)

type ExtraSupportLimits struct {
	MaxRows, MaxPairs, MaxOperations, MaxOutputBytes int
	Stop                                             func() bool
}

type ExtraSupportMinimizationStep struct {
	Index, Remaining int
	Mask             uint64
}

type ExtraSupportMinimization struct {
	Matrix []uint64
	Steps  []ExtraSupportMinimizationStep
}

type ExtraSupportVerification struct {
	Codes      []uint64
	Collisions int
}

type ExtraSupportReductionProfile struct {
	Singles, Doubles []int
}

type ExtraSupportReduction struct {
	Remaining int
	Removed   []int
	Active    uint64
	Profile   ExtraSupportReductionProfile
}

type ExtraSupportProblem struct {
	Variables int
	Cubes     []uint64
	Pairs     []uint64
	Greedy    ExtraSupportMinimization
	Reduction ExtraSupportReduction
}

func normalizeExtraSupportLimits(limits ExtraSupportLimits) (ExtraSupportLimits, error) {
	if limits.MaxRows < 0 || limits.MaxPairs < 0 || limits.MaxOperations < 0 || limits.MaxOutputBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxRows == 0 {
		limits.MaxRows = 1 << 20
	}
	if limits.MaxPairs == 0 {
		limits.MaxPairs = 1 << 24
	}
	if limits.MaxOperations == 0 {
		limits.MaxOperations = 1 << 28
	}
	if limits.MaxOutputBytes == 0 {
		limits.MaxOutputBytes = 16 << 20
	}
	return limits, nil
}

func extraSupportMask(bitsCount int) uint64 {
	if bitsCount == 64 {
		return ^uint64(0)
	}
	return uint64(1)<<uint(bitsCount) - 1
}

func checkExtraSupportRows(rows []uint64, bitsCount int, limits ExtraSupportLimits, rejectZero bool) error {
	if bitsCount < 0 || bitsCount > 64 || len(rows) > limits.MaxRows {
		return ErrDimension
	}
	mask := extraSupportMask(bitsCount)
	for _, row := range rows {
		if row&^mask != 0 {
			return ErrDimension
		}
		if rejectZero && row == 0 {
			return fmt.Errorf("%w: zero difference row", ErrFormat)
		}
	}
	return nil
}

func GenerateExtraSupportMasks(ones, bitsCount int, limits ExtraSupportLimits) ([]uint64, error) {
	limits, err := normalizeExtraSupportLimits(limits)
	if err != nil || ones < 0 || bitsCount < 0 || bitsCount > 62 || ones > bitsCount {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	space := uint64(1) << uint(bitsCount)
	if space > uint64(limits.MaxOperations) {
		return nil, ErrLimit
	}
	result := make([]uint64, 0)
	for mask := uint64(0); mask < space; mask++ {
		if mask&255 == 0 && limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		if bits.OnesCount64(mask) == ones {
			if len(result) == limits.MaxRows {
				return nil, ErrLimit
			}
			result = append(result, mask)
		}
	}
	return result, nil
}

func GenerateExtraSupportPairs(rows []uint64, bitsCount int, limits ExtraSupportLimits) ([]uint64, error) {
	limits, err := normalizeExtraSupportLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = checkExtraSupportRows(rows, bitsCount, limits, false); err != nil {
		return nil, err
	}
	pairs := int64(len(rows)) * int64(len(rows)-1) / 2
	if pairs > int64(limits.MaxOperations) {
		return nil, ErrLimit
	}
	result := make([]uint64, 0, min(int(pairs), limits.MaxPairs))
	seen := make(map[uint64]bool)
	operations := 0
	for first := 0; first < len(rows); first++ {
		for second := first + 1; second < len(rows); second++ {
			if operations&255 == 0 && limits.Stop != nil && limits.Stop() {
				return nil, ErrCancelled
			}
			operations++
			value := rows[first] ^ rows[second]
			if seen[value] {
				continue
			}
			if len(result) == limits.MaxPairs {
				return nil, ErrLimit
			}
			seen[value] = true
			result = append(result, value)
		}
	}
	return result, nil
}

func GenerateExtraSupportEvenMasks(maxHalfWeight, bitsCount int, limits ExtraSupportLimits) ([]uint64, error) {
	limits, err := normalizeExtraSupportLimits(limits)
	if err != nil || maxHalfWeight < 0 || bitsCount < 0 || bitsCount > 62 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	space := uint64(1) << uint(bitsCount)
	if space > uint64(limits.MaxOperations) {
		return nil, ErrLimit
	}
	result := []uint64{}
	for mask := uint64(0); mask < space; mask++ {
		if mask&255 == 0 && limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		weight := bits.OnesCount64(mask)
		if weight >= 2 && weight&1 == 0 && weight/2 <= maxHalfWeight {
			if len(result) == limits.MaxRows {
				return nil, ErrLimit
			}
			result = append(result, mask)
		}
	}
	return result, nil
}

type ExtraSupportGreedy struct {
	bits   int
	rows   []uint64
	parity []bool
	limits ExtraSupportLimits
	work   int
}

func NewExtraSupportGreedy(rows []uint64, bitsCount int, limits ExtraSupportLimits) (*ExtraSupportGreedy, error) {
	limits, err := normalizeExtraSupportLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = checkExtraSupportRows(rows, bitsCount, limits, false); err != nil || bitsCount == 0 && len(rows) != 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &ExtraSupportGreedy{bits: bitsCount, rows: slices.Clone(rows), parity: make([]bool, len(rows)), limits: limits}, nil
}

func (state *ExtraSupportGreedy) check() error {
	if state == nil {
		return ErrDimension
	}
	if state.bits < 0 || state.bits > 64 || len(state.rows) != len(state.parity) || state.limits.MaxOperations <= 0 {
		return ErrInvariant
	}
	return nil
}

func (state *ExtraSupportGreedy) charge(amount int) error {
	if amount < 0 || state.work > state.limits.MaxOperations-amount {
		return ErrLimit
	}
	if state.limits.Stop != nil && state.limits.Stop() {
		return ErrCancelled
	}
	state.work += amount
	return nil
}

func (state *ExtraSupportGreedy) Rows() ([]uint64, error) {
	if err := state.check(); err != nil {
		return nil, err
	}
	return slices.Clone(state.rows), nil
}

func (state *ExtraSupportGreedy) Profile() ([]int, error) {
	if err := state.check(); err != nil {
		return nil, err
	}
	if err := state.charge(len(state.rows) * state.bits); err != nil {
		return nil, err
	}
	counts := make([]int, state.bits)
	for index, row := range state.rows {
		for variable := 0; variable < state.bits; variable++ {
			if row&(uint64(1)<<uint(variable)) != 0 != state.parity[index] {
				counts[variable]++
			}
		}
	}
	return counts, nil
}

func (state *ExtraSupportGreedy) BestVariable() (int, int, error) {
	if err := state.check(); err != nil {
		return 0, 0, err
	}
	if state.bits == 0 {
		return 0, 0, ErrDimension
	}
	counts, err := state.Profile()
	if err != nil {
		return 0, 0, err
	}
	best := 0
	for variable := 1; variable < state.bits; variable++ {
		if counts[best] < counts[variable] {
			best = variable
		}
	}
	return best, counts[best], nil
}

func (state *ExtraSupportGreedy) SelectVariable(variable int) error {
	if err := state.check(); err != nil {
		return err
	}
	if variable < 0 || variable >= state.bits {
		return ErrDimension
	}
	if err := state.charge(len(state.rows)); err != nil {
		return err
	}
	for index, row := range state.rows {
		if row&(uint64(1)<<uint(variable)) != 0 {
			state.parity[index] = !state.parity[index]
		}
	}
	return nil
}

func (state *ExtraSupportGreedy) FilterCovered() (int, error) {
	if err := state.check(); err != nil {
		return 0, err
	}
	if err := state.charge(len(state.rows)); err != nil {
		return 0, err
	}
	kept := 0
	for index, row := range state.rows {
		if state.parity[index] {
			continue
		}
		state.rows[kept] = row
		state.parity[kept] = false
		kept++
	}
	removed := len(state.rows) - kept
	state.rows, state.parity = state.rows[:kept], state.parity[:kept]
	return removed, nil
}

func (state *ExtraSupportGreedy) FindOne() (uint64, error) {
	if err := state.check(); err != nil {
		return 0, err
	}
	if len(state.rows) == 0 {
		return 0, ErrDimension
	}
	for _, row := range state.rows {
		if row == 0 {
			return 0, fmt.Errorf("%w: zero difference row", ErrFormat)
		}
	}
	previous, mask := -1, uint64(0)
	for {
		variable, merit, err := state.BestVariable()
		if err != nil {
			return 0, err
		}
		if previous >= merit {
			return mask, nil
		}
		previous = merit
		if err = state.SelectVariable(variable); err != nil {
			return 0, err
		}
		mask |= uint64(1) << uint(variable)
	}
}

func MinimizeExtraSupport(rows []uint64, bitsCount int, limits ExtraSupportLimits) (ExtraSupportMinimization, error) {
	state, err := NewExtraSupportGreedy(rows, bitsCount, limits)
	if err != nil {
		return ExtraSupportMinimization{}, err
	}
	result := ExtraSupportMinimization{}
	for len(state.rows) > 0 {
		mask, findErr := state.FindOne()
		if findErr != nil {
			return ExtraSupportMinimization{}, findErr
		}
		removed, filterErr := state.FilterCovered()
		if filterErr != nil {
			return ExtraSupportMinimization{}, filterErr
		}
		if removed == 0 {
			return ExtraSupportMinimization{}, ErrInvariant
		}
		result.Matrix = append(result.Matrix, mask)
		result.Steps = append(result.Steps, ExtraSupportMinimizationStep{Index: len(result.Matrix) - 1, Remaining: len(state.rows), Mask: mask})
	}
	return result, nil
}

func VerifyExtraSupport(cubes, matrix []uint64, bitsCount int, limits ExtraSupportLimits) (ExtraSupportVerification, error) {
	limits, err := normalizeExtraSupportLimits(limits)
	if err != nil || len(matrix) > 64 {
		if err != nil {
			return ExtraSupportVerification{}, err
		}
		return ExtraSupportVerification{}, ErrDimension
	}
	if err = checkExtraSupportRows(cubes, bitsCount, limits, false); err != nil || checkExtraSupportRows(matrix, bitsCount, limits, false) != nil {
		return ExtraSupportVerification{}, ErrDimension
	}
	if len(cubes)*len(matrix) > limits.MaxOperations {
		return ExtraSupportVerification{}, ErrLimit
	}
	result := ExtraSupportVerification{Codes: make([]uint64, len(cubes))}
	seen := make(map[uint64]int)
	for index, cube := range cubes {
		if index&255 == 0 && limits.Stop != nil && limits.Stop() {
			return ExtraSupportVerification{}, ErrCancelled
		}
		for variable, row := range matrix {
			if bits.OnesCount64(cube&row)&1 != 0 {
				result.Codes[index] |= uint64(1) << uint(variable)
			}
		}
		result.Collisions += seen[result.Codes[index]]
		seen[result.Codes[index]]++
	}
	return result, nil
}

func DifferenceExtraSupport(cubes []uint64, bitsCount int, limits ExtraSupportLimits) ([]uint64, error) {
	limits, err := normalizeExtraSupportLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = checkExtraSupportRows(cubes, bitsCount, limits, false); err != nil {
		return nil, err
	}
	pairs := int64(len(cubes)) * int64(len(cubes)-1) / 2
	if pairs > int64(limits.MaxPairs) || pairs > int64(limits.MaxOperations) {
		return nil, ErrLimit
	}
	result := make([]uint64, 0, int(pairs))
	for first := 0; first < len(cubes); first++ {
		for second := first + 1; second < len(cubes); second++ {
			if len(result)&255 == 0 && limits.Stop != nil && limits.Stop() {
				return nil, ErrCancelled
			}
			result = append(result, cubes[first]^cubes[second])
		}
	}
	return result, nil
}

func extraSupportReductionProfile(rows []uint64, active uint64, bitsCount int) ExtraSupportReductionProfile {
	result := ExtraSupportReductionProfile{Singles: make([]int, bitsCount), Doubles: make([]int, bitsCount)}
	for _, row := range rows {
		remaining := row & active
		switch bits.OnesCount64(remaining) {
		case 1:
			result.Singles[bits.TrailingZeros64(remaining)]++
		case 2:
			for value := remaining; value != 0; value &= value - 1 {
				result.Doubles[bits.TrailingZeros64(value)]++
			}
		}
	}
	return result
}

func SolveExtraSupport(rows []uint64, bitsCount int, limits ExtraSupportLimits) (ExtraSupportReduction, error) {
	limits, err := normalizeExtraSupportLimits(limits)
	if err != nil {
		return ExtraSupportReduction{}, err
	}
	if err = checkExtraSupportRows(rows, bitsCount, limits, true); err != nil {
		return ExtraSupportReduction{}, err
	}
	active := extraSupportMask(bitsCount)
	result := ExtraSupportReduction{Active: active}
	work := 0
	for {
		if work > limits.MaxOperations-len(rows)*max(bitsCount, 1) {
			return ExtraSupportReduction{}, ErrLimit
		}
		if limits.Stop != nil && limits.Stop() {
			return ExtraSupportReduction{}, ErrCancelled
		}
		work += len(rows) * max(bitsCount, 1)
		profile := extraSupportReductionProfile(rows, active, bitsCount)
		best, merit := -1, -1
		for variable := 0; variable < bitsCount; variable++ {
			if active&(uint64(1)<<uint(variable)) == 0 || profile.Singles[variable] != 0 {
				continue
			}
			if best == -1 || merit > profile.Doubles[variable] {
				best, merit = variable, profile.Doubles[variable]
			}
		}
		if best == -1 {
			result.Active = active
			result.Remaining = bits.OnesCount64(active)
			result.Profile = profile
			return result, nil
		}
		active &^= uint64(1) << uint(best)
		result.Removed = append(result.Removed, best)
		for _, row := range rows {
			if row&active == 0 {
				return ExtraSupportReduction{}, ErrInvariant
			}
		}
	}
}

func FormatExtraSupportMask(mask uint64, bitsCount int) (string, error) {
	if bitsCount < 0 || bitsCount > 64 || mask&^extraSupportMask(bitsCount) != 0 {
		return "", ErrDimension
	}
	text := make([]byte, bitsCount+1)
	for variable := 0; variable < bitsCount; variable++ {
		text[variable] = '0' + byte(mask>>uint(variable)&1)
	}
	text[bitsCount] = '\n'
	return string(text), nil
}

func FormatExtraSupportProfile(rows []uint64, bitsCount int, limits ExtraSupportLimits) (string, error) {
	state, err := NewExtraSupportGreedy(rows, bitsCount, limits)
	if err != nil {
		return "", err
	}
	counts, err := state.Profile()
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	for variable, count := range counts {
		percent := 0.0
		if len(rows) != 0 {
			percent = 100 * float64(count) / float64(len(rows))
		}
		fmt.Fprintf(&builder, "%2d : %6d  %6.2f %%\n", variable, count, percent)
		if builder.Len() > state.limits.MaxOutputBytes {
			return "", ErrLimit
		}
	}
	return builder.String(), nil
}

func ParseExtraSupportMIN(data []byte, limits ExtraSupportLimits) (int, []uint64, error) {
	limits, err := normalizeExtraSupportLimits(limits)
	if err != nil {
		return 0, nil, err
	}
	if len(data) > limits.MaxOutputBytes {
		return 0, nil, ErrLimit
	}
	const marker = "INPUT F-COVER"
	start := strings.Index(string(data), marker)
	if start < 0 {
		return 0, nil, fmt.Errorf("%w: missing %s", ErrFormat, marker)
	}
	tokens := strings.FieldsFunc(string(data[start+len(marker):]), func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == ','
	})
	if len(tokens) == 0 {
		return 0, nil, ErrFormat
	}
	declared, err := strconv.Atoi(tokens[0])
	if err != nil || declared < 1 || declared > 1_000_000 || declared > limits.MaxRows {
		return 0, nil, ErrFormat
	}
	cubes := make([]uint64, 0, declared)
	variables, variable, cube := -1, 0, uint64(0)
	operations := 0
	for _, token := range tokens[1:] {
		if operations == limits.MaxOperations {
			return 0, nil, ErrLimit
		}
		if operations&255 == 0 && limits.Stop != nil && limits.Stop() {
			return 0, nil, ErrCancelled
		}
		operations++
		if len(token) > 2 {
			if strings.HasPrefix(token, "INPUT") {
				break
			}
			if variable > 64 || variables >= 0 && variables != variable {
				return 0, nil, ErrFormat
			}
			if variables == -1 {
				variables = variable
			}
			cubes = append(cubes, cube)
			cube, variable = 0, 0
			continue
		}
		if len(token) != 2 || token != "10" && token != "01" || variable >= 64 {
			return 0, nil, ErrFormat
		}
		if token == "10" {
			cube |= uint64(1) << uint(variable)
		}
		variable++
	}
	if len(cubes) != declared || variables < 0 {
		return 0, nil, ErrFormat
	}
	return variables, cubes, nil
}

func LoadExtraSupportMIN(path string, limits ExtraSupportLimits) (int, []uint64, error) {
	limits, err := normalizeExtraSupportLimits(limits)
	if err != nil {
		return 0, nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, nil, err
	}
	if info.Size() > int64(limits.MaxOutputBytes) {
		return 0, nil, ErrLimit
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, nil, err
	}
	if len(data) > limits.MaxOutputBytes {
		return 0, nil, ErrLimit
	}
	return ParseExtraSupportMIN(data, limits)
}

func RunExtraSupportProblem(cubes []uint64, bitsCount int, limits ExtraSupportLimits) (ExtraSupportProblem, error) {
	pairs, err := DifferenceExtraSupport(cubes, bitsCount, limits)
	if err != nil {
		return ExtraSupportProblem{}, err
	}
	greedy, err := MinimizeExtraSupport(pairs, bitsCount, limits)
	if err != nil {
		return ExtraSupportProblem{}, err
	}
	reduction, err := SolveExtraSupport(pairs, bitsCount, limits)
	if err != nil {
		return ExtraSupportProblem{}, err
	}
	return ExtraSupportProblem{Variables: bitsCount, Cubes: slices.Clone(cubes), Pairs: pairs, Greedy: greedy, Reduction: reduction}, nil
}
