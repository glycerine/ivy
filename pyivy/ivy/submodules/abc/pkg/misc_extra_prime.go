package abc

// Prime enumeration and full-minterm cube pairing from
// src/misc/extra/extraUtilPrime.c.

import (
	"bytes"
	"math"
	"slices"
	"strconv"
	"strings"
)

type ExtraPrimeLimits struct {
	MaxValue, MaxPrimes, MaxOperations, MaxBytes int
	Stop                                         func() bool
}

func normalizeExtraPrimeLimits(limits ExtraPrimeLimits) (ExtraPrimeLimits, error) {
	if limits.MaxValue < 0 || limits.MaxPrimes < 0 || limits.MaxOperations < 0 || limits.MaxBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxValue == 0 {
		limits.MaxValue = 1 << 24
	}
	if limits.MaxPrimes == 0 {
		limits.MaxPrimes = 2 << 20
	}
	if limits.MaxOperations == 0 {
		limits.MaxOperations = 1 << 28
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 256 << 20
	}
	return limits, nil
}

type ExtraPrimePair struct {
	First, Second int
}

type ExtraPrimeGeneration struct {
	Variables        int
	ExclusiveLimit   int
	Primes           []int
	DistanceOnePairs []ExtraPrimePair
}

type ExtraCubeDifference struct {
	Common               int
	Tagged               []int
	Duplicate, Contained bool
}

func checkExtraSortedCube(cube []int) error {
	for index, literal := range cube {
		if literal < 0 || literal > math.MaxInt/2 {
			return ErrDimension
		}
		if index > 0 && cube[index-1] >= literal {
			return ErrInvariant
		}
	}
	return nil
}

// ExtraCubeFreeDifference is the owned form of Tab_ManCubeFree. Each unmatched
// source literal is shifted left and tagged in its low bit to distinguish the
// two cubes, exactly as the source tags its temporary divisor vector.
func ExtraCubeFreeDifference(first, second []int) (ExtraCubeDifference, error) {
	if err := checkExtraSortedCube(first); err != nil {
		return ExtraCubeDifference{}, err
	}
	if err := checkExtraSortedCube(second); err != nil {
		return ExtraCubeDifference{}, err
	}
	if len(first) > math.MaxInt-len(second) {
		return ExtraCubeDifference{}, ErrLimit
	}
	result := ExtraCubeDifference{Tagged: make([]int, 0, len(first)+len(second))}
	firstIndex, secondIndex := 0, 0
	firstTag, secondTag := 0, 1
	for firstIndex < len(first) && secondIndex < len(second) {
		switch {
		case first[firstIndex] == second[secondIndex]:
			result.Common++
			firstIndex++
			secondIndex++
		case first[firstIndex] < second[secondIndex]:
			result.Tagged = append(result.Tagged, first[firstIndex]<<1|firstTag)
			firstIndex++
		default:
			if len(result.Tagged) == 0 {
				firstTag, secondTag = 1, 0
			}
			result.Tagged = append(result.Tagged, second[secondIndex]<<1|secondTag)
			secondIndex++
		}
	}
	for ; firstIndex < len(first); firstIndex++ {
		result.Tagged = append(result.Tagged, first[firstIndex]<<1|firstTag)
	}
	for ; secondIndex < len(second); secondIndex++ {
		result.Tagged = append(result.Tagged, second[secondIndex]<<1|secondTag)
	}
	result.Duplicate = len(result.Tagged) == 0
	result.Contained = len(result.Tagged) == 1
	return result, nil
}

// ExtraCubesEqualWithout compares sorted cubes after omitting one zero-based
// position from each. An omission of -1 retains the whole cube.
func ExtraCubesEqualWithout(first, second []int, omitFirst, omitSecond int) (bool, error) {
	if err := checkExtraSortedCube(first); err != nil {
		return false, err
	}
	if err := checkExtraSortedCube(second); err != nil {
		return false, err
	}
	if omitFirst < -1 || omitFirst >= len(first) || omitSecond < -1 || omitSecond >= len(second) {
		return false, ErrDimension
	}
	firstLength, secondLength := len(first), len(second)
	if omitFirst >= 0 {
		firstLength--
	}
	if omitSecond >= 0 {
		secondLength--
	}
	if firstLength != secondLength {
		return false, nil
	}
	firstIndex, secondIndex := 0, 0
	for firstIndex < len(first) && secondIndex < len(second) {
		if firstIndex == omitFirst {
			firstIndex++
			continue
		}
		if secondIndex == omitSecond {
			secondIndex++
			continue
		}
		if firstIndex == len(first) || secondIndex == len(second) {
			break
		}
		if first[firstIndex] != second[secondIndex] {
			return false, nil
		}
		firstIndex++
		secondIndex++
	}
	return true, nil
}

func extraPrimeDomainLimit(variables int) (int, error) {
	if variables < 0 || variables >= strconv.IntSize-1 {
		return 0, ErrDimension
	}
	limit64 := uint64(1) << uint(variables)
	if limit64 > uint64(math.MaxInt) {
		return 0, ErrLimit
	}
	return int(limit64), nil
}

func extraPrimeSieveLimit(variables int, limits ExtraPrimeLimits) (int, error) {
	limit, err := extraPrimeDomainLimit(variables)
	if err != nil {
		return 0, err
	}
	if limit > limits.MaxValue || limit > limits.MaxBytes {
		return 0, ErrLimit
	}
	return limit, nil
}

type extraPrimeBudget struct {
	limits     ExtraPrimeLimits
	operations int
}

func (budget *extraPrimeBudget) use(count int) error {
	if count < 0 || budget.operations > budget.limits.MaxOperations-count {
		return ErrLimit
	}
	budget.operations += count
	if budget.limits.Stop != nil && budget.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

// GenerateExtraPrimes returns every prime strictly below 2^variables and all
// unordered prime pairs whose binary representations differ in one bit.
func GenerateExtraPrimes(variables int, limits ExtraPrimeLimits) (ExtraPrimeGeneration, error) {
	limits, err := normalizeExtraPrimeLimits(limits)
	if err != nil {
		return ExtraPrimeGeneration{}, err
	}
	limit, err := extraPrimeSieveLimit(variables, limits)
	if err != nil {
		return ExtraPrimeGeneration{}, err
	}
	budget := extraPrimeBudget{limits: limits}
	if err = budget.use(1); err != nil {
		return ExtraPrimeGeneration{}, err
	}
	composite := make([]bool, limit)
	if limit > 0 {
		composite[0] = true
	}
	if limit > 1 {
		composite[1] = true
	}
	for candidate := 2; candidate < limit; candidate++ {
		if err = budget.use(1); err != nil {
			return ExtraPrimeGeneration{}, err
		}
		if composite[candidate] || candidate > (limit-1)/candidate {
			continue
		}
		for multiple := candidate * candidate; multiple < limit; multiple += candidate {
			if err = budget.use(1); err != nil {
				return ExtraPrimeGeneration{}, err
			}
			composite[multiple] = true
		}
	}
	primes := make([]int, 0)
	for value := 2; value < limit; value++ {
		if err = budget.use(1); err != nil {
			return ExtraPrimeGeneration{}, err
		}
		if composite[value] {
			continue
		}
		if len(primes) == math.MaxInt || len(primes) == limits.MaxPrimes {
			return ExtraPrimeGeneration{}, ErrLimit
		}
		primeBytes := extraPrimeIntBytes(len(primes) + 1)
		if primeBytes > limits.MaxBytes-limit {
			return ExtraPrimeGeneration{}, ErrLimit
		}
		primes = append(primes, value)
	}
	pairs, err := extraPrimeDistanceOnePairs(primes, variables, &budget)
	if err != nil {
		return ExtraPrimeGeneration{}, err
	}
	return ExtraPrimeGeneration{
		Variables:        variables,
		ExclusiveLimit:   limit,
		Primes:           primes,
		DistanceOnePairs: pairs,
	}, nil
}

func extraPrimeIntBytes(count int) int {
	width := strconv.IntSize / 8
	if count < 0 || count > math.MaxInt/width {
		return math.MaxInt
	}
	return count * width
}

func extraPrimePairBytes(count int) int {
	width := 2 * (strconv.IntSize / 8)
	if count < 0 || count > math.MaxInt/width {
		return math.MaxInt
	}
	return count * width
}

func validateExtraPrimeValues(values []int, variables int, limits ExtraPrimeLimits) ([]int, int, error) {
	limit, err := extraPrimeDomainLimit(variables)
	if err != nil {
		return nil, 0, err
	}
	if len(values) > limits.MaxPrimes || extraPrimeIntBytes(len(values)) > limits.MaxBytes {
		return nil, 0, ErrLimit
	}
	owned := slices.Clone(values)
	ordered := slices.Clone(values)
	slices.Sort(ordered)
	for index, value := range ordered {
		if value < 0 || value >= limit {
			return nil, 0, ErrDimension
		}
		if index > 0 && ordered[index-1] == value {
			return nil, 0, ErrInvariant
		}
	}
	return owned, limit, nil
}

func extraPrimeDistanceOnePairs(values []int, variables int, budget *extraPrimeBudget) ([]ExtraPrimePair, error) {
	ordered, _, err := validateExtraPrimeValues(values, variables, budget.limits)
	if err != nil {
		return nil, err
	}
	slices.Sort(ordered)
	if len(ordered) > 0 && variables > (budget.limits.MaxOperations-budget.operations)/len(ordered) {
		return nil, ErrLimit
	}
	pairs := make([]ExtraPrimePair, 0)
	for _, first := range ordered {
		for variable := 0; variable < variables; variable++ {
			if err = budget.use(1); err != nil {
				return nil, err
			}
			second := first ^ 1<<uint(variable)
			_, found := slices.BinarySearch(ordered, second)
			if first < second && found {
				if len(pairs) == math.MaxInt {
					return nil, ErrLimit
				}
				if extraPrimePairBytes(len(pairs)+1) > budget.limits.MaxBytes-extraPrimeIntBytes(len(ordered)) {
					return nil, ErrLimit
				}
				pairs = append(pairs, ExtraPrimePair{First: first, Second: second})
			}
		}
	}
	return pairs, nil
}

// ExtraPrimeDistanceOnePairs is the owned, non-printing replacement for
// Abc_GenCountHits1. Input order is irrelevant and output is numeric.
func ExtraPrimeDistanceOnePairs(values []int, variables int, limits ExtraPrimeLimits) ([]ExtraPrimePair, error) {
	limits, err := normalizeExtraPrimeLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	budget := extraPrimeBudget{limits: limits}
	pairs, err := extraPrimeDistanceOnePairs(values, variables, &budget)
	if err != nil {
		return nil, err
	}
	return pairs, nil
}

// MarshalExtraPrimePLA returns Abc_GenCountDump's PLA bytes. Timestamp is
// caller-owned so output is reproducible and independent of a process clock.
func MarshalExtraPrimePLA(values []int, variables int, timestamp string, limits ExtraPrimeLimits) ([]byte, error) {
	limits, err := normalizeExtraPrimeLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	ordered, _, err := validateExtraPrimeValues(values, variables, limits)
	if err != nil {
		return nil, err
	}
	if strings.ContainsAny(timestamp, "\r\n\x00") {
		return nil, ErrDimension
	}
	// Bound the variable-size header before concatenating it.
	fixedHeader := len("#  prime numbers up to 2^ generated by ABC on \n") + len(strconv.Itoa(len(ordered))) + len(strconv.Itoa(variables))
	if len(timestamp) > limits.MaxBytes || fixedHeader > limits.MaxBytes-len(timestamp) {
		return nil, ErrLimit
	}
	budget := extraPrimeBudget{limits: limits}
	var output bytes.Buffer
	appendString := func(text string) error {
		if output.Len() > limits.MaxBytes-len(text) {
			return ErrLimit
		}
		output.WriteString(text)
		return nil
	}
	if err = appendString("# " + strconv.Itoa(len(ordered)) + " prime numbers up to 2^" + strconv.Itoa(variables) + " generated by ABC on " + timestamp + "\n"); err != nil {
		return nil, err
	}
	if err = appendString(".i " + strconv.Itoa(variables) + "\n.o 1\n.p " + strconv.Itoa(len(ordered)) + "\n"); err != nil {
		return nil, err
	}
	for _, value := range ordered {
		if err = budget.use(max(1, variables)); err != nil {
			return nil, err
		}
		if output.Len() > limits.MaxBytes-variables-3 {
			return nil, ErrLimit
		}
		for bit := variables - 1; bit >= 0; bit-- {
			output.WriteByte(byte('0' + (value>>uint(bit))&1))
		}
		output.WriteString(" 1\n")
	}
	if err = appendString(".e\n\n"); err != nil {
		return nil, err
	}
	return slices.Clone(output.Bytes()), nil
}

// ExtraPrimeCubeTable is Tab_Man_t's observable full-minterm state. The C
// hash table only filters candidates; Go indexes values directly and cannot
// lose pairs to hash collisions.
type ExtraPrimeCubeTable struct {
	variables int
	values    []int
	closed    bool
}

func NewExtraPrimeCubeTable(variables int, values []int, limits ExtraPrimeLimits) (*ExtraPrimeCubeTable, error) {
	limits, err := normalizeExtraPrimeLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	ordered, _, err := validateExtraPrimeValues(values, variables, limits)
	if err != nil {
		return nil, err
	}
	return &ExtraPrimeCubeTable{variables: variables, values: ordered}, nil
}

func (table *ExtraPrimeCubeTable) check() error {
	if table == nil {
		return ErrDimension
	}
	if table.closed || table.variables < 0 {
		return ErrInvariant
	}
	return nil
}

func (table *ExtraPrimeCubeTable) Variables() (int, error) {
	if err := table.check(); err != nil {
		return 0, err
	}
	return table.variables, nil
}

func (table *ExtraPrimeCubeTable) Values() ([]int, error) {
	if err := table.check(); err != nil {
		return nil, err
	}
	return slices.Clone(table.values), nil
}

func (table *ExtraPrimeCubeTable) LiteralCount() (int, error) {
	if err := table.check(); err != nil {
		return 0, err
	}
	if table.variables != 0 && len(table.values) > math.MaxInt/table.variables {
		return 0, ErrLimit
	}
	return table.variables * len(table.values), nil
}

// ItemCount preserves Tab_ManCountItems for the used distance-one case and
// uses n*(n-1)/2 for the dormant distance-two case. The C expression n*n/2
// disagrees with its own nested a<b loop for most n.
func (table *ExtraPrimeCubeTable) ItemCount(distanceTwo bool) (int, error) {
	if err := table.check(); err != nil {
		return 0, err
	}
	perCube := 1 + table.variables
	if distanceTwo {
		if table.variables > 1 && table.variables > math.MaxInt/(table.variables-1) {
			return 0, ErrLimit
		}
		perCube += table.variables * (table.variables - 1) / 2
	}
	if perCube != 0 && len(table.values) > math.MaxInt/perCube {
		return 0, ErrLimit
	}
	return perCube * len(table.values), nil
}

// Cube returns source literals in ascending variable order. A set value bit
// is a positive literal and a clear bit is complemented.
func (table *ExtraPrimeCubeTable) Cube(index int) ([]int, error) {
	if err := table.check(); err != nil {
		return nil, err
	}
	if index < 0 || index >= len(table.values) {
		return nil, ErrDimension
	}
	value := table.values[index]
	cube := make([]int, table.variables)
	for variable := range cube {
		complement := 1
		if value&(1<<uint(variable)) != 0 {
			complement = 0
		}
		cube[variable] = variable<<1 | complement
	}
	return cube, nil
}

// FormatCube preserves the active Tab_ManPrintCube behavior. Its Var argument
// is accepted and range-checked, but the source's dash-printing branch is
// commented out, so every position is printed as zero or one.
func (table *ExtraPrimeCubeTable) FormatCube(index, variable int) (string, error) {
	if err := table.check(); err != nil {
		return "", err
	}
	if index < 0 || index >= len(table.values) || variable < -1 || variable >= table.variables {
		return "", ErrDimension
	}
	var text strings.Builder
	text.Grow(table.variables)
	for cubeVariable := 0; cubeVariable < table.variables; cubeVariable++ {
		if table.values[index]&(1<<uint(cubeVariable)) != 0 {
			text.WriteByte('1')
		} else {
			text.WriteByte('0')
		}
	}
	return text.String(), nil
}

func (table *ExtraPrimeCubeTable) CollectDistanceOne(limits ExtraPrimeLimits) ([]ExtraPrimePair, error) {
	if err := table.check(); err != nil {
		return nil, err
	}
	return ExtraPrimeDistanceOnePairs(table.values, table.variables, limits)
}

func (table *ExtraPrimeCubeTable) Close() error {
	if err := table.check(); err != nil {
		return err
	}
	table.values = nil
	table.closed = true
	return nil
}
