package abc

// Extended-range floating arithmetic from src/bdd/epd/epd.h and epd.c.

import (
	"math"
	"strconv"
	"strings"
)

const (
	EPDMaxBinaryExponent  = 1023
	EPDMaxDecimalExponent = 308
	EPDInfinityExponent   = 0x7ff
)

type EPDMode uint8

const (
	EPDBinary EPDMode = iota
	EPDDecimal
)

type EPDLimits struct {
	MaxAbsExponent int
	Stop           func() bool
}

func normalizeEPDLimits(limits EPDLimits) (EPDLimits, error) {
	if limits.MaxAbsExponent < 0 {
		return limits, ErrDimension
	}
	if limits.MaxAbsExponent == 0 {
		limits.MaxAbsExponent = math.MaxInt32
	}
	return limits, nil
}

// EPDouble stores a normalized significand and a separate binary or decimal
// exponent. It replaces the source union and endian-dependent bit fields with
// an endian-independent owned value.
type EPDouble struct {
	value    float64
	exponent int
	mode     EPDMode
	limits   EPDLimits
	closed   bool
}

func NewEPDouble(value float64, limits EPDLimits) (*EPDouble, error) {
	limits, err := normalizeEPDLimits(limits)
	if err != nil {
		return nil, err
	}
	number := &EPDouble{value: value, mode: EPDBinary, limits: limits}
	if err = number.normalize(); err != nil {
		return nil, err
	}
	return number, nil
}

func NewDecimalEPDouble(value float64, limits EPDLimits) (*EPDouble, error) {
	limits, err := normalizeEPDLimits(limits)
	if err != nil {
		return nil, err
	}
	number := &EPDouble{value: value, mode: EPDDecimal, limits: limits}
	if err = number.normalize(); err != nil {
		return nil, err
	}
	return number, nil
}

func (number *EPDouble) check() error {
	if number == nil || number.closed {
		return ErrInvariant
	}
	if number.mode != EPDBinary && number.mode != EPDDecimal {
		return ErrInvariant
	}
	if number.limits.MaxAbsExponent <= 0 {
		return ErrInvariant
	}
	return nil
}

func (number *EPDouble) stopped() bool {
	return number.limits.Stop != nil && number.limits.Stop()
}

func epdCheckedAdd(first, second int) (int, error) {
	if second > 0 && first > math.MaxInt-second || second < 0 && first < math.MinInt-second {
		return 0, ErrLimit
	}
	return first + second, nil
}

func epdAbsWithin(value, maximum int) bool {
	if value >= 0 {
		return value <= maximum
	}
	return value >= -maximum
}

func (number *EPDouble) normalize() error {
	if number.stopped() {
		return ErrCancelled
	}
	if math.IsNaN(number.value) {
		number.value, number.exponent = math.NaN(), 0
		return nil
	}
	if math.IsInf(number.value, 0) {
		number.exponent = 0
		return nil
	}
	if number.value == 0 {
		number.exponent = 0
		return nil
	}
	var shift int
	if number.mode == EPDBinary {
		fraction, binaryExponent := math.Frexp(number.value)
		number.value = fraction * 2
		shift = binaryExponent - 1
	} else {
		shift = EPDGetDecimalExponent(number.value)
		if shift < -EPDMaxDecimalExponent {
			// math.Pow10(-324) is zero. Scale subnormals in two finite
			// steps so their significand is preserved.
			number.value *= math.Pow10(EPDMaxDecimalExponent)
			number.value *= math.Pow10(-shift - EPDMaxDecimalExponent)
		} else {
			number.value /= math.Pow10(shift)
		}
	}
	exponent, err := epdCheckedAdd(number.exponent, shift)
	if err != nil || !epdAbsWithin(exponent, number.limits.MaxAbsExponent) {
		return ErrLimit
	}
	number.exponent = exponent
	return nil
}

func (number *EPDouble) cloneValue() EPDouble { return *number }

// Convert replaces EpdConvert and atomically resets this number to binary
// normalization of value.
func (number *EPDouble) Convert(value float64) error {
	if err := number.check(); err != nil {
		return err
	}
	candidate := number.cloneValue()
	candidate.value, candidate.exponent, candidate.mode = value, 0, EPDBinary
	if err := candidate.normalize(); err != nil {
		return err
	}
	*number = candidate
	return nil
}

func (number *EPDouble) NormalizeBinary() error {
	if err := number.check(); err != nil {
		return err
	}
	candidate := number.cloneValue()
	candidate.mode = EPDBinary
	if err := candidate.normalize(); err != nil {
		return err
	}
	*number = candidate
	return nil
}

func (number *EPDouble) NormalizeDecimal() error {
	if err := number.check(); err != nil {
		return err
	}
	candidate := number.cloneValue()
	candidate.mode = EPDDecimal
	if err := candidate.normalize(); err != nil {
		return err
	}
	*number = candidate
	return nil
}

func (number *EPDouble) Significand() (float64, error) {
	if err := number.check(); err != nil {
		return 0, err
	}
	return number.value, nil
}

func (number *EPDouble) Exponent() (int, error) {
	if err := number.check(); err != nil {
		return 0, err
	}
	return number.exponent, nil
}

func (number *EPDouble) Mode() (EPDMode, error) {
	if err := number.check(); err != nil {
		return 0, err
	}
	return number.mode, nil
}

func (number *EPDouble) Clone() (*EPDouble, error) {
	if err := number.check(); err != nil {
		return nil, err
	}
	clone := number.cloneValue()
	return &clone, nil
}

func (number *EPDouble) CopyFrom(other *EPDouble) error {
	if err := number.check(); err != nil {
		return err
	}
	if err := other.check(); err != nil {
		return err
	}
	if number.stopped() {
		return ErrCancelled
	}
	number.value, number.exponent, number.mode = other.value, other.exponent, other.mode
	return nil
}

func (number *EPDouble) EqualStorage(other *EPDouble) (bool, error) {
	if err := number.check(); err != nil {
		return false, err
	}
	if err := other.check(); err != nil {
		return false, err
	}
	// EpdCmp uses floating equality, so two NaNs compare unequal.
	return number.value == other.value && number.exponent == other.exponent && number.mode == other.mode, nil
}

func (number *EPDouble) IsInf() bool {
	return number != nil && !number.closed && math.IsInf(number.value, 0)
}
func (number *EPDouble) IsZero() bool { return number != nil && !number.closed && number.value == 0 }
func (number *EPDouble) IsNaN() bool {
	return number != nil && !number.closed && math.IsNaN(number.value)
}
func (number *EPDouble) IsNaNOrInf() bool { return number.IsNaN() || number.IsInf() }

func EPDIsInfDouble(value float64) int {
	if math.IsInf(value, 1) {
		return 1
	}
	if math.IsInf(value, -1) {
		return -1
	}
	return 0
}

func EPDIsNaNDouble(value float64) bool      { return math.IsNaN(value) }
func EPDIsNaNOrInfDouble(value float64) bool { return math.IsNaN(value) || math.IsInf(value, 0) }

func (number *EPDouble) MakeInf(negative bool) error {
	if err := number.check(); err != nil {
		return err
	}
	if number.stopped() {
		return ErrCancelled
	}
	number.value, number.exponent = math.Inf(1), 0
	if negative {
		number.value = math.Inf(-1)
	}
	return nil
}

func (number *EPDouble) MakeZero(negative bool) error {
	if err := number.check(); err != nil {
		return err
	}
	if number.stopped() {
		return ErrCancelled
	}
	number.value, number.exponent = 0, 0
	if negative {
		number.value = math.Copysign(0, -1)
	}
	return nil
}

func (number *EPDouble) MakeNaN() error {
	if err := number.check(); err != nil {
		return err
	}
	if number.stopped() {
		return ErrCancelled
	}
	number.value, number.exponent = math.NaN(), 0
	return nil
}

func (number *EPDouble) requireSameMode(other *EPDouble) error {
	if err := number.check(); err != nil {
		return err
	}
	if err := other.check(); err != nil {
		return err
	}
	if number.mode != other.mode {
		return ErrDimension
	}
	return nil
}

func (number *EPDouble) commitBinary(value float64, exponent int) error {
	candidate := number.cloneValue()
	candidate.value, candidate.exponent = value, exponent
	if err := candidate.normalize(); err != nil {
		return err
	}
	*number = candidate
	return nil
}

func epdSpecialProduct(first, second float64) (float64, bool) {
	if math.IsNaN(first) || math.IsNaN(second) {
		return math.NaN(), true
	}
	if math.IsInf(first, 0) || math.IsInf(second, 0) {
		negative := math.Signbit(first) != math.Signbit(second)
		if negative {
			return math.Inf(-1), true
		}
		return math.Inf(1), true
	}
	return 0, false
}

func (number *EPDouble) Multiply(other *EPDouble) error {
	if err := number.requireSameMode(other); err != nil {
		return err
	}
	if number.stopped() {
		return ErrCancelled
	}
	if special, ok := epdSpecialProduct(number.value, other.value); ok {
		return number.commitBinary(special, 0)
	}
	exponent, err := epdCheckedAdd(number.exponent, other.exponent)
	if err != nil {
		return err
	}
	return number.commitBinary(number.value*other.value, exponent)
}

func (number *EPDouble) MultiplyFloat(value float64) error {
	other, err := NewEPDouble(value, number.limits)
	if err != nil {
		return err
	}
	if number.mode == EPDDecimal {
		other, err = NewDecimalEPDouble(value, number.limits)
		if err != nil {
			return err
		}
	}
	return number.Multiply(other)
}

func EPDProduct(first, second *EPDouble) (*EPDouble, error) {
	clone, err := first.Clone()
	if err != nil {
		return nil, err
	}
	if err = clone.Multiply(second); err != nil {
		return nil, err
	}
	return clone, nil
}

func (number *EPDouble) Divide(other *EPDouble) error {
	if err := number.requireSameMode(other); err != nil {
		return err
	}
	if number.stopped() {
		return ErrCancelled
	}
	if math.IsNaN(number.value) || math.IsNaN(other.value) || math.IsInf(number.value, 0) && math.IsInf(other.value, 0) {
		return number.commitBinary(math.NaN(), 0)
	}
	if math.IsInf(number.value, 0) {
		negative := math.Signbit(number.value) != math.Signbit(other.value)
		return number.MakeInf(negative)
	}
	if math.IsInf(other.value, 0) {
		negative := math.Signbit(number.value) != math.Signbit(other.value)
		return number.MakeZero(negative)
	}
	if other.value == 0 {
		return number.commitBinary(math.NaN(), 0)
	}
	exponent, err := epdCheckedAdd(number.exponent, -other.exponent)
	if err != nil {
		return err
	}
	return number.commitBinary(number.value/other.value, exponent)
}

func (number *EPDouble) DivideFloat(value float64) error {
	other, err := NewEPDouble(value, number.limits)
	if err != nil {
		return err
	}
	if number.mode == EPDDecimal {
		other, err = NewDecimalEPDouble(value, number.limits)
		if err != nil {
			return err
		}
	}
	return number.Divide(other)
}

func EPDQuotient(first, second *EPDouble) (*EPDouble, error) {
	clone, err := first.Clone()
	if err != nil {
		return nil, err
	}
	if err = clone.Divide(second); err != nil {
		return nil, err
	}
	return clone, nil
}

func epdAligned(first, second *EPDouble, subtract bool) (float64, int) {
	if first.exponent > second.exponent {
		value := first.value
		// Subtraction in uint gives the exact non-negative distance across
		// the whole int range, including opposite-signed extreme exponents.
		difference := uint(first.exponent) - uint(second.exponent)
		if difference <= EPDMaxBinaryExponent {
			term := math.Ldexp(second.value, -int(difference))
			if subtract {
				value -= term
			} else {
				value += term
			}
		}
		return value, first.exponent
	}
	if first.exponent < second.exponent {
		value := second.value
		if subtract {
			value = -value
		}
		difference := uint(second.exponent) - uint(first.exponent)
		if difference <= EPDMaxBinaryExponent {
			term := math.Ldexp(first.value, -int(difference))
			if subtract {
				value = term - second.value
			} else {
				value = term + second.value
			}
		}
		return value, second.exponent
	}
	if subtract {
		return first.value - second.value, first.exponent
	}
	return first.value + second.value, first.exponent
}

func (number *EPDouble) Add(other *EPDouble) error {
	return number.addOrSubtract(other, false)
}

func (number *EPDouble) Subtract(other *EPDouble) error {
	return number.addOrSubtract(other, true)
}

func (number *EPDouble) addOrSubtract(other *EPDouble, subtract bool) error {
	if err := number.requireSameMode(other); err != nil {
		return err
	}
	if number.mode != EPDBinary {
		return ErrDimension
	}
	if number.stopped() {
		return ErrCancelled
	}
	left, right := number.value, other.value
	if subtract {
		right = -right
	}
	if math.IsNaN(left) || math.IsNaN(right) || math.IsInf(left, 0) && math.IsInf(right, 0) && math.Signbit(left) != math.Signbit(right) {
		return number.commitBinary(math.NaN(), 0)
	}
	if math.IsInf(left, 0) {
		return number.commitBinary(left, 0)
	}
	if math.IsInf(right, 0) {
		return number.commitBinary(right, 0)
	}
	value, exponent := epdAligned(number, other, subtract)
	return number.commitBinary(value, exponent)
}

func (number *EPDouble) AddFloat(value float64) error {
	other, err := NewEPDouble(value, number.limits)
	if err != nil {
		return err
	}
	return number.Add(other)
}

func (number *EPDouble) SubtractFloat(value float64) error {
	other, err := NewEPDouble(value, number.limits)
	if err != nil {
		return err
	}
	return number.Subtract(other)
}

func EPDSum(first, second *EPDouble) (*EPDouble, error) {
	clone, err := first.Clone()
	if err != nil {
		return nil, err
	}
	if err = clone.Add(second); err != nil {
		return nil, err
	}
	return clone, nil
}

func EPDDifference(first, second *EPDouble) (*EPDouble, error) {
	clone, err := first.Clone()
	if err != nil {
		return nil, err
	}
	if err = clone.Subtract(second); err != nil {
		return nil, err
	}
	return clone, nil
}

func EPDPow2(exponent int, limits EPDLimits) (*EPDouble, error) {
	limits, err := normalizeEPDLimits(limits)
	if err != nil {
		return nil, err
	}
	if !epdAbsWithin(exponent, limits.MaxAbsExponent) {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &EPDouble{value: 1, exponent: exponent, mode: EPDBinary, limits: limits}, nil
}

func EPDPow2Decimal(exponent int, limits EPDLimits) (*EPDouble, error) {
	limits, err := normalizeEPDLimits(limits)
	if err != nil {
		return nil, err
	}
	if !epdAbsWithin(exponent, limits.MaxAbsExponent) {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	// Preserve the source's direct-power base case and recursive positive
	// splitting. Extend its negative range by reciprocal construction instead
	// of silently underflowing below the hardware double range.
	if exponent >= -EPDMaxBinaryExponent && exponent <= EPDMaxBinaryExponent {
		return NewDecimalEPDouble(math.Pow(2, float64(exponent)), limits)
	}
	if exponent < -EPDMaxBinaryExponent {
		positive, powerErr := EPDPow2Decimal(-exponent, limits)
		if powerErr != nil {
			return nil, powerErr
		}
		result := &EPDouble{value: 1 / positive.value, exponent: -positive.exponent, mode: EPDDecimal, limits: limits}
		if powerErr = result.normalize(); powerErr != nil {
			return nil, powerErr
		}
		return result, nil
	}
	firstExponent := exponent / 2
	secondExponent := exponent - firstExponent
	first, err := EPDPow2Decimal(firstExponent, limits)
	if err != nil {
		return nil, err
	}
	second, err := EPDPow2Decimal(secondExponent, limits)
	if err != nil {
		return nil, err
	}
	return EPDProduct(first, second)
}

func (number *EPDouble) DecimalParts() (float64, int, error) {
	if err := number.check(); err != nil {
		return 0, 0, err
	}
	if number.IsNaNOrInf() {
		return number.value, 0, nil
	}
	if number.value == 0 {
		return number.value, 0, nil
	}
	if number.mode == EPDDecimal {
		return number.value, number.exponent, nil
	}
	power, err := EPDPow2Decimal(number.exponent, number.limits)
	if err != nil {
		return 0, 0, err
	}
	base := &EPDouble{value: number.value, mode: EPDDecimal, limits: number.limits}
	product, err := EPDProduct(base, power)
	if err != nil {
		return 0, 0, err
	}
	return product.value, product.exponent, nil
}

func EPDGetBinaryExponent(value float64) int {
	return int((math.Float64bits(value) >> 52) & EPDInfinityExponent)
}

func EPDGetDecimalExponent(value float64) int {
	if value == 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	text := strconv.FormatFloat(value, 'E', -1, 64)
	position := strings.LastIndexByte(text, 'E')
	exponent, err := strconv.Atoi(text[position+1:])
	if err != nil {
		return 0
	}
	return exponent
}

func (number *EPDouble) String() string {
	if number == nil || number.closed {
		return "<closed>"
	}
	if math.IsNaN(number.value) {
		return "NaN"
	}
	if math.IsInf(number.value, -1) {
		return "-Inf"
	}
	if math.IsInf(number.value, 1) {
		return "Inf"
	}
	value, exponent, err := number.DecimalParts()
	if err != nil {
		return "NaN"
	}
	text := strconv.FormatFloat(value, 'e', 6, 64)
	position := strings.LastIndexByte(text, 'e')
	sign := '+'
	if exponent < 0 {
		sign = '-'
		exponent = -exponent
	}
	digits := strconv.Itoa(exponent)
	if len(digits) < 2 {
		digits = "0" + digits
	}
	return text[:position+1] + string(sign) + digits
}

func (number *EPDouble) Close() {
	if number == nil || number.closed {
		return
	}
	number.value, number.exponent, number.closed = 0, 0, true
}
