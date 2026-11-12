package epd

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func EpdAlloc() *EpDouble { return new(EpDouble) }
func EpdFree(*EpDouble)   {}
func EpdCmp(a, b *EpDouble) int {
	if a.Value != b.Value || a.Exponent != b.Exponent {
		return 1
	}
	return 0
}
func EpdCopy(a, b *EpDouble)            { *b = *a }
func EpdConvert(v float64, p *EpDouble) { p.Value = v; p.Exponent = 0; EpdNormalize(p) }
func EpdNormalize(p *EpDouble) {
	if p.Value == 0 || math.IsNaN(p.Value) || math.IsInf(p.Value, 0) {
		p.Exponent = 0
		return
	}
	f, e := math.Frexp(p.Value)
	p.Value = f * 2
	p.Exponent += e - 1
}
func EpdNormalizeDecimal(p *EpDouble) {
	if p.Value == 0 || math.IsNaN(p.Value) || math.IsInf(p.Value, 0) {
		p.Exponent = 0
		return
	}
	e := EpdGetExponentDecimal(p.Value)
	// 10^-324 is smaller than the least subnormal double. Split scaling
	// to avoid dividing a finite subnormal by an underflowed zero.
	if e < -308 {
		p.Value = (p.Value * 1e308) / math.Pow10(e+308)
	} else {
		p.Value /= math.Pow10(e)
	}
	p.Exponent += e
}
func EpdMultiply(p *EpDouble, v float64) { var q EpDouble; EpdConvert(v, &q); EpdMultiply2(p, &q) }
func EpdMultiply2(a, b *EpDouble)        { EpdMultiply3(a, b, a) }
func EpdMultiply3(a, b, c *EpDouble) {
	*c = EpDouble{a.Value * b.Value, a.Exponent + b.Exponent}
	EpdNormalize(c)
}
func EpdMultiply2Decimal(a, b *EpDouble) { EpdMultiply3Decimal(a, b, a) }
func EpdMultiply3Decimal(a, b, c *EpDouble) {
	*c = EpDouble{a.Value * b.Value, a.Exponent + b.Exponent}
	EpdNormalizeDecimal(c)
}
func EpdDivide(p *EpDouble, v float64) { var q EpDouble; EpdConvert(v, &q); EpdDivide2(p, &q) }
func EpdDivide2(a, b *EpDouble)        { EpdDivide3(a, b, a) }
func EpdDivide3(a, b, c *EpDouble) {
	*c = EpDouble{a.Value / b.Value, a.Exponent - b.Exponent}
	EpdNormalize(c)
}
func EpdAdd(p *EpDouble, v float64) { var q EpDouble; EpdConvert(v, &q); EpdAdd2(p, &q) }
func EpdAdd2(a, b *EpDouble)        { EpdAdd3(a, b, a) }
func EpdAdd3(a, b, c *EpDouble) {
	x, y := *a, *b
	if math.IsInf(x.Value, 0) || math.IsInf(y.Value, 0) || math.IsNaN(x.Value) || math.IsNaN(y.Value) {
		*c = EpDouble{Value: x.Value + y.Value}
		return
	}
	if x.Value == 0 {
		*c = y
		return
	}
	if y.Value == 0 {
		*c = x
		return
	}
	if x.Exponent < y.Exponent {
		x, y = y, x
	}
	d := x.Exponent - y.Exponent
	v := x.Value
	if d < 1075 {
		v += math.Ldexp(y.Value, -d)
	}
	*c = EpDouble{v, x.Exponent}
	EpdNormalize(c)
}
func EpdSubtract(p *EpDouble, v float64) { EpdAdd(p, -v) }
func EpdSubtract2(a, b *EpDouble)        { EpdSubtract3(a, b, a) }
func EpdSubtract3(a, b, c *EpDouble)     { q := *b; q.Value = -q.Value; EpdAdd3(a, &q, c) }
func EpdPow2(n int, p *EpDouble)         { *p = EpDouble{1, n} }

// Compute decimal powers by splitting the binary exponent, matching EPD's
// multiplication order without taking logarithms of huge exponents.
func EpdPow2Decimal(n int, p *EpDouble) {
	if n < -1022 {
		var positive EpDouble
		// -(n+1) avoids integer overflow for the most negative machine integer.
		EpdPow2Decimal(-(n + 1), &positive)
		*p = EpDouble{0.5 / positive.Value, -positive.Exponent}
		EpdNormalizeDecimal(p)
		return
	}
	if n <= EPD_MAX_BIN {
		*p = EpDouble{Value: math.Ldexp(1, n)}
		EpdNormalizeDecimal(p)
		return
	}
	var a, b EpDouble
	EpdPow2Decimal(n/2, &a)
	EpdPow2Decimal(n-n/2, &b)
	EpdMultiply3Decimal(&a, &b, p)
}
func EpdGetValueAndDecimalExponent(p *EpDouble, v *float64, e *int) {
	if p.Value == 0 || math.IsNaN(p.Value) || math.IsInf(p.Value, 0) {
		*v = p.Value
		*e = 0
		return
	}
	a := EpDouble{Value: p.Value}
	var scale EpDouble
	EpdPow2Decimal(p.Exponent, &scale)
	EpdMultiply2Decimal(&a, &scale)
	*v = a.Value
	*e = a.Exponent
}
func EpdGetString(p *EpDouble) string {
	if math.IsNaN(p.Value) {
		return "NaN"
	}
	if math.IsInf(p.Value, 1) {
		return "Inf"
	}
	if math.IsInf(p.Value, -1) {
		return "-Inf"
	}
	var v float64
	var e int
	EpdGetValueAndDecimalExponent(p, &v, &e)
	mantissa := strconv.FormatFloat(v, 'e', 6, 64)
	mantissa = mantissa[:strings.IndexByte(mantissa, 'e')]
	return fmt.Sprintf("%se%+03d", mantissa, e)
}
func EpdGetExponent(v float64) int { return int((math.Float64bits(v) >> 52) & 0x7ff) }
func EpdGetExponentDecimal(v float64) int {
	if v == 0 || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	text := strconv.FormatFloat(v, 'E', 6, 64)
	exponent, _ := strconv.Atoi(text[strings.IndexByte(text, 'E')+1:])
	return exponent
}
func EpdMakeInf(p *EpDouble, sign int) {
	v := 1
	if sign != 0 {
		v = -1
	}
	*p = EpDouble{Value: math.Inf(v)}
}
func EpdMakeZero(p *EpDouble, sign int) {
	*p = EpDouble{}
	if sign != 0 {
		p.Value = math.Copysign(0, -1)
	}
}
func EpdMakeNan(p *EpDouble) { *p = EpDouble{Value: math.NaN()} }
func flag(v bool) int {
	if v {
		return 1
	}
	return 0
}
func EpdIsInf(p *EpDouble) int       { return IsInfDouble(p.Value) }
func EpdIsZero(p *EpDouble) int      { return flag(p.Value == 0) }
func EpdIsNan(p *EpDouble) int       { return IsNanDouble(p.Value) }
func EpdIsNanOrInf(p *EpDouble) int  { return IsNanOrInfDouble(p.Value) }
func IsInfDouble(v float64) int      { return flag(math.IsInf(v, 0)) }
func IsNanDouble(v float64) int      { return flag(math.IsNaN(v)) }
func IsNanOrInfDouble(v float64) int { return flag(math.IsNaN(v) || math.IsInf(v, 0)) }
