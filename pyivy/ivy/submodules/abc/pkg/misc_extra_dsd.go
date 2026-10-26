package abc

// Six-variable DSD class management from src/misc/extra/extraUtilDsd.c.

import (
	_ "embed"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
)

//go:embed test_vectors/extra_dsd/classes6.tsv
var extraDSDClassData string

const ExtraDSDClassCount = 595

type ExtraDSDClass struct {
	Variables, Ands, Clauses int
	Truth                    uint64
	Formula                  string
}

type ExtraDSDOptions struct {
	MaxChecks, MaxDivisorCalls int
	Stop                       func() bool
}

type ExtraDSDStatistics struct {
	ClassCounts []int
	NonDSD, All int
}

type ExtraDSDManager struct {
	classes    []ExtraDSDClass
	truthClass map[uint64]int
	perms      [][6]int
	permIndex  map[int]int
	permuted   []uint64
	options    ExtraDSDOptions
	statistics ExtraDSDStatistics
}

func parseExtraDSDClasses(text string) ([]ExtraDSDClass, error) {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) != ExtraDSDClassCount+1 || lines[0] != "variables\tands\tclauses\ttruth\tformula" {
		return nil, ErrFormat
	}
	result := make([]ExtraDSDClass, ExtraDSDClassCount)
	for index, line := range lines[1:] {
		fields := strings.Split(line, "\t")
		if len(fields) != 5 {
			return nil, fmt.Errorf("%w: DSD row %d", ErrFormat, index)
		}
		values := [3]int{}
		for field := 0; field < 3; field++ {
			value, err := strconv.Atoi(fields[field])
			if err != nil {
				return nil, ErrFormat
			}
			values[field] = value
		}
		truth, err := strconv.ParseUint(fields[3], 16, 64)
		if err != nil {
			return nil, ErrFormat
		}
		if values[0] < 0 || values[0] > 6 || values[1] < 0 || values[2] < 0 || fields[4] == "" {
			return nil, ErrFormat
		}
		result[index] = ExtraDSDClass{values[0], values[1], values[2], truth, fields[4]}
	}
	return result, nil
}

func ExtraGrayCodeSchedule(variables int) ([]int, error) {
	if variables < 1 || variables > 20 {
		return nil, ErrDimension
	}
	result := make([]int, 1<<uint(variables))
	offset := 0
	for variable := 0; variable < variables; variable++ {
		result[offset] = variable
		offset++
		for index := 1; index < 1<<uint(variable); index++ {
			result[offset] = result[index-1]
			offset++
		}
	}
	result[offset] = variables - 1
	return result, nil
}

func ExtraPermutationSchedule(size int) ([]int, error) {
	if size < 1 || size > 10 {
		return nil, ErrDimension
	}
	if size == 1 {
		return []int{0}, nil
	}
	if size == 2 {
		return []int{0, 0}, nil
	}
	previous, err := ExtraPermutationSchedule(size - 1)
	if err != nil {
		return nil, err
	}
	factorial := 1
	for value := 2; value <= size; value++ {
		factorial *= value
	}
	groups := factorial / size / 2
	result := make([]int, 0, factorial)
	for group := 0; group < groups; group++ {
		for index := size - 1; index > 0; index-- {
			result = append(result, index-1)
		}
		result = append(result, previous[2*group]+1)
		for index := 0; index < size-1; index++ {
			result = append(result, index)
		}
		result = append(result, previous[2*group+1])
	}
	return result, nil
}

func extraTruth6SwapAdjacent(truth uint64, variable int) uint64 {
	if variable < 0 || variable >= 5 {
		return truth
	}
	low, high := truthVariablePattern(variable), truthVariablePattern(variable+1)
	up, down := low&^high, high&^low
	shift := (uint(1) << uint(variable+1)) - (uint(1) << uint(variable))
	return truth&^(up|down) | truth&up<<shift | truth&down>>shift
}

func extraTruth6Flip(truth uint64, variable int) uint64 {
	if variable < 0 || variable >= 6 {
		return truth
	}
	shift := uint(1) << uint(variable)
	mask := truthVariablePattern(variable)
	return truth<<shift&mask | truth&mask>>shift
}

func extraTruth6SwapVariables(truth uint64, first, second int) uint64 {
	if first == second {
		return truth
	}
	if first > second {
		first, second = second, first
	}
	low, high := truthVariablePattern(first), truthVariablePattern(second)
	up, down := low&^high, high&^low
	shift := (uint(1) << uint(second)) - (uint(1) << uint(first))
	return truth&^(up|down) | truth&up<<shift | truth&down>>shift
}

func extraDSDPermuteTruth(truth uint64, permutation [6]int) uint64 {
	labels := permutation
	for target := 0; target < 6; target++ {
		position := target
		for labels[position] != target {
			position++
		}
		if position != target {
			truth = extraTruth6SwapVariables(truth, target, position)
			labels[target], labels[position] = labels[position], labels[target]
		}
	}
	return truth
}

func extraDSDPackPermutation(permutation [6]int) int {
	value := 0
	for index, item := range permutation {
		value |= item << uint(3*index)
	}
	return value
}

func NewExtraDSDManager(options ExtraDSDOptions) (*ExtraDSDManager, error) {
	if options.MaxChecks < 0 || options.MaxDivisorCalls < 0 {
		return nil, ErrDimension
	}
	if options.MaxChecks == 0 {
		options.MaxChecks = 1 << 20
	}
	if options.MaxDivisorCalls == 0 {
		options.MaxDivisorCalls = 1 << 20
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	classes, err := parseExtraDSDClasses(extraDSDClassData)
	if err != nil {
		return nil, err
	}
	schedule, _ := ExtraPermutationSchedule(6)
	perms := make([][6]int, len(schedule))
	current := [6]int{0, 1, 2, 3, 4, 5}
	indexes := make(map[int]int, len(schedule))
	for index, swap := range schedule {
		current[swap], current[swap+1] = current[swap+1], current[swap]
		perms[index] = current
		indexes[extraDSDPackPermutation(current)] = index
	}
	if current != ([6]int{0, 1, 2, 3, 4, 5}) || len(indexes) != 720 {
		return nil, ErrInvariant
	}
	m := &ExtraDSDManager{classes: classes, truthClass: make(map[uint64]int, len(classes)), perms: perms, permIndex: indexes, permuted: make([]uint64, len(classes)*len(perms)), options: options, statistics: ExtraDSDStatistics{ClassCounts: make([]int, len(classes))}}
	for class, row := range classes {
		normalized, representativeComplement := row.Truth, 0
		if normalized&1 != 0 {
			normalized, representativeComplement = ^normalized, 1
		}
		if _, duplicate := m.truthClass[normalized]; duplicate {
			return nil, ErrInvariant
		}
		m.truthClass[normalized] = class<<1 | representativeComplement
		truth := row.Truth
		for index, swap := range schedule {
			truth = extraTruth6SwapAdjacent(truth, swap)
			m.permuted[class*720+index] = truth
		}
		if truth != row.Truth {
			return nil, ErrInvariant
		}
	}
	return m, nil
}

func (m *ExtraDSDManager) usable() bool {
	return m != nil && len(m.classes) == ExtraDSDClassCount && len(m.perms) == 720
}
func (m *ExtraDSDManager) Classes() []ExtraDSDClass {
	if !m.usable() {
		return nil
	}
	return slices.Clone(m.classes)
}
func (m *ExtraDSDManager) Class(index int) (ExtraDSDClass, error) {
	if !m.usable() || index < 0 || index >= len(m.classes) {
		return ExtraDSDClass{}, ErrDimension
	}
	return m.classes[index], nil
}
func (m *ExtraDSDManager) CNFCosts() []int {
	if !m.usable() {
		return nil
	}
	result := make([]int, len(m.classes))
	for index := 2; index < len(result); index++ {
		result[index] = m.classes[index].Clauses
	}
	return result
}
func (m *ExtraDSDManager) ResetStatistics() {
	if !m.usable() {
		return
	}
	m.statistics = ExtraDSDStatistics{ClassCounts: make([]int, len(m.classes))}
}
func (m *ExtraDSDManager) Statistics() ExtraDSDStatistics {
	if !m.usable() {
		return ExtraDSDStatistics{}
	}
	result := m.statistics
	result.ClassCounts = slices.Clone(result.ClassCounts)
	return result
}
func (m *ExtraDSDManager) FormatPermutation(mask int) (string, error) {
	permutation, index, err := m.resolvePermutation(mask)
	if err != nil {
		return "", err
	}
	_ = index
	var b strings.Builder
	for _, value := range permutation {
		fmt.Fprintf(&b, "%d ", value)
	}
	b.WriteString(" ")
	return b.String(), nil
}

func (m *ExtraDSDManager) resolvePermutation(mask int) ([6]int, int, error) {
	if !m.usable() || mask < 0 || mask >= 1<<18 {
		return [6]int{}, 0, ErrDimension
	}
	used := [6]bool{}
	result := [6]int{}
	missing := make([]int, 0, 6)
	for variable := 0; variable < 6; variable++ {
		value := (mask >> uint(3*variable)) & 7
		if value == 6 || value < 0 {
			return result, 0, ErrDimension
		}
		if value < 6 {
			if used[value] {
				return result, 0, ErrDimension
			}
			used[value] = true
			result[variable] = value
		} else {
			result[variable] = -1
		}
	}
	for value := 0; value < 6; value++ {
		if !used[value] {
			missing = append(missing, value)
		}
	}
	next := 0
	for index := range result {
		if result[index] < 0 {
			result[index] = missing[next]
			next++
		}
	}
	permutationIndex, found := m.permIndex[extraDSDPackPermutation(result)]
	if !found {
		return result, 0, ErrInvariant
	}
	return result, permutationIndex, nil
}

func extraDSDTransformToClass(truth uint64, permutation [6]int, phase int) uint64 {
	var result uint64
	for classAssignment := 0; classAssignment < 64; classAssignment++ {
		source := 0
		for oldVariable := 0; oldVariable < 6; oldVariable++ {
			bit := (classAssignment >> uint(permutation[oldVariable])) & 1
			bit ^= (phase >> uint(oldVariable)) & 1
			source |= bit << uint(oldVariable)
		}
		if truth>>uint(source)&1 != 0 {
			result |= 1 << uint(classAssignment)
		}
	}
	return result
}

// Check returns the packed source configuration or -1 when truth is not DSD.
func (m *ExtraDSDManager) Check(truth uint64) (int, error) {
	if !m.usable() {
		return -1, ErrDimension
	}
	identity := [6]int{0, 1, 2, 3, 4, 5}
	identityIndex := m.permIndex[extraDSDPackPermutation(identity)]
	normalized, complement := truth, 0
	if normalized&1 != 0 {
		normalized, complement = ^normalized, 1
	}
	if entry, found := m.truthClass[normalized]; found {
		class, outputComplement := entry>>1, complement^(entry&1)
		return class<<17 | outputComplement<<16 | identityIndex<<6, nil
	}
	checks := 0
	gray, _ := ExtraGrayCodeSchedule(6)
	for permutationIndex, permutation := range m.perms {
		candidate := extraDSDPermuteTruth(truth, permutation)
		phase := 0
		for step := -1; step < 63; step++ {
			if m.options.Stop != nil && m.options.Stop() {
				return -1, ErrCancelled
			}
			if checks == m.options.MaxChecks {
				return -1, ErrLimit
			}
			checks++
			if step >= 0 {
				variable := gray[step]
				phase ^= 1 << uint(variable)
				candidate = extraTruth6Flip(candidate, permutation[variable])
			}
			normalizedCandidate := candidate
			outputComplement := 0
			if normalizedCandidate&1 != 0 {
				normalizedCandidate = ^normalizedCandidate
				outputComplement = 1
			}
			if entry, found := m.truthClass[normalizedCandidate]; found {
				class := entry >> 1
				outputComplement ^= entry & 1
				return class<<17 | outputComplement<<16 | permutationIndex<<6 | phase, nil
			}
		}
	}
	return -1, nil
}

type ExtraDSDComputeResult struct {
	Literal int
	Cut     []int
	Truth   uint64
	Class   int
}

func (m *ExtraDSDManager) Compute(firstLiteral, secondLiteral int, cut []int, mask int, xor bool) (ExtraDSDComputeResult, error) {
	fail := func(err error) (ExtraDSDComputeResult, error) { return ExtraDSDComputeResult{}, err }
	if !m.usable() || firstLiteral < 0 || secondLiteral < 0 || firstLiteral >= 2*len(m.classes) || secondLiteral >= 2*len(m.classes) || mask <= 1 {
		return fail(ErrDimension)
	}
	_, permutationIndex, err := m.resolvePermutation(mask & 0x3ffff)
	if err != nil {
		return fail(err)
	}
	first := m.classes[firstLiteral>>1].Truth
	second := m.permuted[(secondLiteral>>1)*720+permutationIndex]
	for variable := 0; variable < 6; variable++ {
		if mask>>uint(18+variable)&1 != 0 {
			second = extraTruth6Flip(second, variable)
		}
	}
	if firstLiteral&1 != 0 {
		first = ^first
	}
	if secondLiteral&1 != 0 {
		second = ^second
	}
	truth := first & second
	if xor {
		truth = first ^ second
	}
	m.statistics.All++
	config, err := m.Check(truth)
	if err != nil {
		return fail(err)
	}
	if config < 0 {
		m.statistics.NonDSD++
		return ExtraDSDComputeResult{Literal: -1, Truth: truth, Class: -1}, nil
	}
	class := config >> 17
	outputComplement := (config >> 16) & 1
	phase := config & 63
	permutationIndex = (config >> 6) & 0x3ff
	result := ExtraDSDComputeResult{Literal: class<<1 | outputComplement, Truth: truth, Class: class}
	if cut != nil {
		if len(cut) < 1 || cut[0] < 0 || cut[0] > 6 || len(cut) != cut[0]+1 {
			return fail(ErrDimension)
		}
		leaves := make([]int, 6)
		for index := range leaves {
			leaves[index] = -1
		}
		for index := 0; index < cut[0]; index++ {
			target := m.perms[permutationIndex][index]
			leaves[target] = cut[index+1] ^ (phase >> uint(index) & 1)
		}
		result.Cut = make([]int, m.classes[class].Variables+1)
		result.Cut[0] = m.classes[class].Variables
		for index := 0; index < result.Cut[0]; index++ {
			if leaves[index] < 0 {
				return fail(ErrInvariant)
			}
			result.Cut[index+1] = leaves[index]
		}
	}
	m.statistics.ClassCounts[class]++
	return result, nil
}

type ExtraDSDStatisticsReport struct {
	Rows                         []string
	Unused                       int
	UnusedPercent, NonDSDPercent float64
}

func (m *ExtraDSDManager) StatisticsReport(verbose bool) (ExtraDSDStatisticsReport, error) {
	if !m.usable() {
		return ExtraDSDStatisticsReport{}, ErrDimension
	}
	result := ExtraDSDStatisticsReport{}
	for index, count := range m.statistics.ClassCounts {
		if count == 0 {
			result.Unused++
			continue
		}
		if verbose {
			result.Rows = append(result.Rows, fmt.Sprintf("%5d  :  %-20s   %8d", index, m.classes[index].Formula, count))
		}
	}
	result.UnusedPercent = 100 * float64(result.Unused) / float64(len(m.classes))
	denominator := m.statistics.All
	if denominator < 1 {
		denominator = 1
	}
	result.NonDSDPercent = 100 * float64(m.statistics.NonDSD) / float64(denominator)
	return result, nil
}

func extraDSDHasVariable(truth uint64, variable int) bool {
	return dauDSD6Cofactor(truth, variable, false) != dauDSD6Cofactor(truth, variable, true)
}

func (m *ExtraDSDManager) CollectDivisors(truth uint64) (map[int][]uint64, error) {
	if !m.usable() {
		return nil, ErrDimension
	}
	result := map[int][]uint64{2: {}, 3: {}, 4: {}}
	seen := make(map[uint64]bool)
	calls := 0
	var visit func(uint64) error
	visit = func(current uint64) error {
		if m.options.Stop != nil && m.options.Stop() {
			return ErrCancelled
		}
		if calls == m.options.MaxDivisorCalls {
			return ErrLimit
		}
		calls++
		if seen[current] {
			return nil
		}
		seen[current] = true
		config, err := m.Check(current)
		if err != nil {
			return err
		}
		if config >= 0 && config>>17 < 2 {
			return nil
		}
		counter := 0
		for variable := 0; variable < 6; variable++ {
			if !extraDSDHasVariable(current, variable) {
				continue
			}
			if err = visit(dauDSD6Cofactor(current, variable, false)); err != nil {
				return err
			}
			if err = visit(dauDSD6Cofactor(current, variable, true)); err != nil {
				return err
			}
			counter++
		}
		if config >= 0 && counter >= 2 && counter <= 4 {
			normalized := current
			if normalized&1 != 0 {
				normalized = ^normalized
			}
			result[counter] = append(result[counter], normalized)
		}
		return nil
	}
	if err := visit(truth); err != nil {
		return nil, err
	}
	for size := 2; size <= 4; size++ {
		sort.Slice(result[size], func(i, j int) bool { return result[size][i] < result[size][j] })
		result[size] = slices.Compact(result[size])
	}
	return result, nil
}
