package abc

// Fixed-radix arbitrary-precision arithmetic from src/bdd/cudd/cuddApa.c.

import (
	"math"
	"math/bits"
	"strconv"
	"strings"
)

// CuddAPANumber stores base-2^32 digits most-significant first, matching the
// 64-bit-host representation selected by the source.
type CuddAPANumber []uint32

type CuddAPAOptions struct {
	MaxDigits     int
	MaxOperations int
	Stop          func() bool
}

const cuddAPAHardMaxDigits = 1 << 20

func normalizeCuddAPAOptions(options CuddAPAOptions) (CuddAPAOptions, error) {
	if options.MaxDigits < 0 || options.MaxOperations < 0 {
		return options, ErrDimension
	}
	if options.MaxDigits == 0 {
		options.MaxDigits = cuddAPAHardMaxDigits
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 26
	}
	if options.MaxDigits > cuddAPAHardMaxDigits {
		options.MaxDigits = cuddAPAHardMaxDigits
	}
	return options, nil
}

func CuddAPANumberOfDigits(binaryDigits int) (int, error) {
	if binaryDigits <= 0 {
		return 0, ErrDimension
	}
	digits := (binaryDigits + 31) / 32
	if digits > cuddAPAHardMaxDigits {
		return 0, ErrLimit
	}
	return digits, nil
}

func NewCuddAPANumber(digits int) (CuddAPANumber, error) {
	if digits <= 0 {
		return nil, ErrDimension
	}
	if digits > cuddAPAHardMaxDigits {
		return nil, ErrLimit
	}
	return make(CuddAPANumber, digits), nil
}

func checkCuddAPANumber(number CuddAPANumber) error {
	if len(number) == 0 {
		return ErrDimension
	}
	if len(number) > cuddAPAHardMaxDigits {
		return ErrLimit
	}
	return nil
}

func checkCuddAPATernary(first, second, result CuddAPANumber) error {
	if err := checkCuddAPANumber(first); err != nil {
		return err
	}
	if len(second) != len(first) || len(result) != len(first) {
		return ErrDimension
	}
	return nil
}

func CuddAPACopy(source, destination CuddAPANumber) error {
	if err := checkCuddAPANumber(source); err != nil {
		return err
	}
	if len(destination) != len(source) {
		return ErrDimension
	}
	copy(destination, source)
	return nil
}

// CuddAPAAdd returns the carry beyond the most-significant digit. Inputs may
// alias result.
func CuddAPAAdd(first, second, result CuddAPANumber) (uint32, error) {
	if err := checkCuddAPATernary(first, second, result); err != nil {
		return 0, err
	}
	var carry uint32
	for index := len(first) - 1; index >= 0; index-- {
		result[index], carry = bits.Add32(first[index], second[index], carry)
	}
	return carry, nil
}

// CuddAPASubtract returns zero when no borrow leaves the most-significant
// digit and all one bits when a borrow does, matching the unsigned C result.
func CuddAPASubtract(first, second, result CuddAPANumber) (uint32, error) {
	if err := checkCuddAPATernary(first, second, result); err != nil {
		return 0, err
	}
	var borrow uint32
	for index := len(first) - 1; index >= 0; index-- {
		result[index], borrow = bits.Sub32(first[index], second[index], borrow)
	}
	return 0 - borrow, nil
}

func cuddAPADivision(dividend CuddAPANumber, divisor uint32, quotient CuddAPANumber) (uint32, error) {
	if err := checkCuddAPANumber(dividend); err != nil {
		return 0, err
	}
	if divisor == 0 || len(quotient) != len(dividend) {
		return 0, ErrDimension
	}
	var remainder uint32
	for index, digit := range dividend {
		partial := uint64(remainder)<<32 | uint64(digit)
		quotient[index] = uint32(partial / uint64(divisor))
		remainder = uint32(partial % uint64(divisor))
	}
	return remainder, nil
}

func CuddAPAShortDivision(dividend CuddAPANumber, divisor uint32, quotient CuddAPANumber) (uint32, error) {
	return cuddAPADivision(dividend, divisor, quotient)
}

func CuddAPAIntDivision(dividend CuddAPANumber, divisor uint32, quotient CuddAPANumber) (uint32, error) {
	return cuddAPADivision(dividend, divisor, quotient)
}

func CuddAPAShiftRight(in uint32, source, destination CuddAPANumber) error {
	if err := checkCuddAPANumber(source); err != nil {
		return err
	}
	if len(destination) != len(source) {
		return ErrDimension
	}
	carry := (in & 1) << 31
	for index := 0; index < len(source); index++ {
		next := source[index] << 31
		destination[index] = source[index]>>1 | carry
		carry = next
	}
	return nil
}

func CuddAPASetToLiteral(number CuddAPANumber, literal uint32) error {
	if err := checkCuddAPANumber(number); err != nil {
		return err
	}
	clear(number)
	number[len(number)-1] = literal
	return nil
}

func CuddAPAPowerOfTwo(number CuddAPANumber, power int) error {
	if err := checkCuddAPANumber(number); err != nil {
		return err
	}
	if power < 0 {
		return ErrDimension
	}
	clear(number)
	index := len(number) - 1 - power/32
	if index >= 0 {
		number[index] = uint32(1) << uint(power&31)
	}
	return nil
}

func CuddAPACompare(first, second CuddAPANumber) (int, error) {
	if err := checkCuddAPANumber(first); err != nil {
		return 0, err
	}
	if err := checkCuddAPANumber(second); err != nil {
		return 0, err
	}
	firstStart, secondStart := 0, 0
	for firstStart < len(first) && first[firstStart] == 0 {
		firstStart++
	}
	for secondStart < len(second) && second[secondStart] == 0 {
		secondStart++
	}
	firstLength, secondLength := len(first)-firstStart, len(second)-secondStart
	if firstLength > secondLength {
		return 1, nil
	}
	if firstLength < secondLength {
		return -1, nil
	}
	for index := 0; index < firstLength; index++ {
		if first[firstStart+index] > second[secondStart+index] {
			return 1, nil
		}
		if first[firstStart+index] < second[secondStart+index] {
			return -1, nil
		}
	}
	return 0, nil
}

func CuddAPACompareRatios(first CuddAPANumber, firstDenominator uint32, second CuddAPANumber, secondDenominator uint32) (int, error) {
	if firstDenominator == 0 || secondDenominator == 0 {
		return 0, ErrDimension
	}
	firstQuotient, err := NewCuddAPANumber(len(first))
	if err != nil {
		return 0, err
	}
	secondQuotient, err := NewCuddAPANumber(len(second))
	if err != nil {
		return 0, err
	}
	firstRemainder, err := CuddAPAIntDivision(first, firstDenominator, firstQuotient)
	if err != nil {
		return 0, err
	}
	secondRemainder, err := CuddAPAIntDivision(second, secondDenominator, secondQuotient)
	if err != nil {
		return 0, err
	}
	comparison, err := CuddAPACompare(firstQuotient, secondQuotient)
	if err != nil || comparison != 0 {
		return comparison, err
	}
	left := uint64(firstRemainder) * uint64(secondDenominator)
	right := uint64(secondRemainder) * uint64(firstDenominator)
	if left > right {
		return 1, nil
	}
	if left < right {
		return -1, nil
	}
	return 0, nil
}

func CuddAPAHex(number CuddAPANumber) (string, error) {
	if err := checkCuddAPANumber(number); err != nil {
		return "", err
	}
	var builder strings.Builder
	if len(number) > math.MaxInt/8 {
		return "", ErrLimit
	}
	builder.Grow(len(number) * 8)
	for _, digit := range number {
		text := strconv.FormatUint(uint64(digit), 16)
		builder.WriteString("00000000"[:8-len(text)])
		builder.WriteString(text)
	}
	return builder.String(), nil
}

func cuddAPADecimalDigits(number CuddAPANumber, options CuddAPAOptions) ([]byte, error) {
	if err := checkCuddAPANumber(number); err != nil {
		return nil, err
	}
	options, err := normalizeCuddAPAOptions(options)
	if err != nil {
		return nil, err
	}
	if len(number) > options.MaxDigits {
		return nil, ErrLimit
	}
	work := append(CuddAPANumber(nil), number...)
	capacity := int(float64(len(number))*math.Log10(4294967296.0)) + 1
	digits := make([]byte, capacity)
	operations := 0
	for index := capacity - 1; index >= 0; index-- {
		if len(number) > options.MaxOperations-operations {
			return nil, ErrLimit
		}
		operations += len(number)
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		remainder, err := CuddAPAShortDivision(work, 10, work)
		if err != nil {
			return nil, err
		}
		digits[index] = byte(remainder)
	}
	return digits, nil
}

func CuddAPADecimal(number CuddAPANumber) (string, error) {
	return CuddAPADecimalWithOptions(number, CuddAPAOptions{})
}

func CuddAPADecimalWithOptions(number CuddAPANumber, options CuddAPAOptions) (string, error) {
	digits, err := cuddAPADecimalDigits(number, options)
	if err != nil {
		return "", err
	}
	first := 0
	for first < len(digits)-1 && digits[first] == 0 {
		first++
	}
	text := make([]byte, len(digits)-first)
	for index := range text {
		text[index] = '0' + digits[first+index]
	}
	return string(text), nil
}

func CuddAPAExponential(number CuddAPANumber, precision int) (string, error) {
	return CuddAPAExponentialWithOptions(number, precision, CuddAPAOptions{})
}

func CuddAPAExponentialWithOptions(number CuddAPANumber, precision int, options CuddAPAOptions) (string, error) {
	if precision <= 0 {
		return "", ErrDimension
	}
	digits, err := cuddAPADecimalDigits(number, options)
	if err != nil {
		return "", err
	}
	first := 0
	for first < len(digits)-1 && digits[first] == 0 {
		first++
	}
	last := first + precision
	if last > len(digits) {
		last = len(digits)
	}
	var builder strings.Builder
	for index := first; index < last; index++ {
		if index == first+1 {
			builder.WriteByte('.')
		}
		builder.WriteByte('0' + digits[index])
	}
	builder.WriteString("e+")
	builder.WriteString(strconv.Itoa(len(digits) - first - 1))
	return builder.String(), nil
}

type cuddAPACountWork struct {
	manager    *ExtraBDDManager
	maximum    CuddAPANumber
	options    CuddAPAOptions
	operations int
	memo       map[uint32]CuddAPANumber
}

func (work *cuddAPACountWork) step() error {
	work.operations++
	if work.operations > work.options.MaxOperations {
		return ErrLimit
	}
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (work *cuddAPACountWork) count(literal ExtraBDDLit) (CuddAPANumber, error) {
	if err := work.step(); err != nil {
		return nil, err
	}
	if literal == ExtraBDDFalse {
		return make(CuddAPANumber, len(work.maximum)), nil
	}
	if literal == ExtraBDDTrue {
		return append(CuddAPANumber(nil), work.maximum...), nil
	}
	if literal.Complemented() {
		regular, err := work.count(literal.Not())
		if err != nil {
			return nil, err
		}
		result := make(CuddAPANumber, len(work.maximum))
		if _, err = CuddAPASubtract(work.maximum, regular, result); err != nil {
			return nil, err
		}
		return result, nil
	}
	if cached, present := work.memo[literal.ID()]; present {
		return append(CuddAPANumber(nil), cached...), nil
	}
	thenCount, err := work.count(work.manager.thenOf(literal))
	if err != nil {
		return nil, err
	}
	elseCount, err := work.count(work.manager.elseOf(literal))
	if err != nil {
		return nil, err
	}
	result := make(CuddAPANumber, len(work.maximum))
	carry, err := CuddAPAAdd(thenCount, elseCount, result)
	if err != nil {
		return nil, err
	}
	if err = CuddAPAShiftRight(carry, result, result); err != nil {
		return nil, err
	}
	work.memo[literal.ID()] = append(CuddAPANumber(nil), result...)
	return result, nil
}

func (manager *ExtraBDDManager) cuddAPASupportSize(root ExtraBDDLit, work *cuddAPACountWork) (int, error) {
	seenNodes := make([]bool, len(manager.nodes))
	seenVariables := make([]bool, manager.variables)
	stack := []ExtraBDDLit{root}
	for len(stack) > 0 {
		if err := work.step(); err != nil {
			return 0, err
		}
		literal := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if literal < 2 || seenNodes[literal.ID()] {
			continue
		}
		seenNodes[literal.ID()] = true
		variable := manager.variableOf(literal)
		if variable < 0 || variable >= len(seenVariables) {
			return 0, ErrInvariant
		}
		seenVariables[variable] = true
		stack = append(stack, manager.thenOf(literal), manager.elseOf(literal))
	}
	count := 0
	for _, present := range seenVariables {
		if present {
			count++
		}
	}
	return count, nil
}

func (manager *ExtraBDDManager) CuddAPACountMinterms(root ExtraBDDLit, variables int, options CuddAPAOptions) (CuddAPANumber, error) {
	if err := manager.checkLiteral(root); err != nil {
		return nil, err
	}
	if variables < 0 || variables == math.MaxInt {
		return nil, ErrDimension
	}
	options, err := normalizeCuddAPAOptions(options)
	if err != nil {
		return nil, err
	}
	digits, err := CuddAPANumberOfDigits(variables + 1)
	if err != nil {
		return nil, err
	}
	if digits > options.MaxDigits {
		return nil, ErrLimit
	}
	maximum := make(CuddAPANumber, digits)
	if err = CuddAPAPowerOfTwo(maximum, variables); err != nil {
		return nil, err
	}
	work := cuddAPACountWork{manager: manager, maximum: maximum, options: options, memo: make(map[uint32]CuddAPANumber)}
	support, err := manager.cuddAPASupportSize(root, &work)
	if err != nil {
		return nil, err
	}
	if support > variables {
		return nil, ErrDimension
	}
	return work.count(root)
}

func (manager *ExtraBDDManager) CuddAPAMintermDecimal(root ExtraBDDLit, variables int, options CuddAPAOptions) (string, error) {
	count, err := manager.CuddAPACountMinterms(root, variables, options)
	if err != nil {
		return "", err
	}
	return CuddAPADecimalWithOptions(count, options)
}

func (manager *ExtraBDDManager) CuddAPAMintermExponential(root ExtraBDDLit, variables, precision int, options CuddAPAOptions) (string, error) {
	count, err := manager.CuddAPACountMinterms(root, variables, options)
	if err != nil {
		return "", err
	}
	return CuddAPAExponentialWithOptions(count, precision, options)
}

func (manager *ExtraBDDManager) CuddAPADensity(root ExtraBDDLit, variables int, options CuddAPAOptions) (string, error) {
	options, err := normalizeCuddAPAOptions(options)
	if err != nil {
		return "", err
	}
	count, err := manager.CuddAPACountMinterms(root, variables, options)
	if err != nil {
		return "", err
	}
	seen := make([]bool, len(manager.nodes))
	seenTerminal := false
	stack := []ExtraBDDLit{root}
	size := uint32(0)
	operations := 0
	for len(stack) > 0 {
		operations++
		if operations > options.MaxOperations {
			return "", ErrLimit
		}
		if options.Stop != nil && options.Stop() {
			return "", ErrCancelled
		}
		literal := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if literal < 2 {
			if !seenTerminal {
				seenTerminal = true
				size++
			}
			continue
		}
		if seen[literal.ID()] {
			continue
		}
		seen[literal.ID()] = true
		size++
		stack = append(stack, manager.thenOf(literal), manager.elseOf(literal))
	}
	if size == 0 {
		return "", ErrInvariant
	}
	quotient := make(CuddAPANumber, len(count))
	remainder, err := CuddAPAIntDivision(count, size, quotient)
	if err != nil {
		return "", err
	}
	whole, err := CuddAPADecimalWithOptions(quotient, options)
	if err != nil {
		return "", err
	}
	fractional := uint32(uint64(remainder) * 1000000 / uint64(size))
	return whole + "." + strconv.FormatUint(uint64(fractional), 10), nil
}
