package epd

import (
	"math"
	"testing"
)

func TestArithmetic(t *testing.T) {
	var a, b, c EpDouble
	EpdConvert(3, &a)
	EpdConvert(5, &b)
	EpdMultiply3(&a, &b, &c)
	if math.Ldexp(c.Value, c.Exponent) != 15 {
		t.Fatal(c)
	}
	EpdAdd3(&a, &b, &c)
	if math.Ldexp(c.Value, c.Exponent) != 8 {
		t.Fatal(c)
	}
	EpdSubtract3(&a, &a, &c)
	if c.Value != 0 {
		t.Fatal(c)
	}
	EpdDivide3(&b, &a, &c)
	if math.Abs(math.Ldexp(c.Value, c.Exponent)-5.0/3) > 1e-15 {
		t.Fatal(c)
	}
}
func TestExtendedAndSpecial(t *testing.T) {
	var a, b EpDouble
	EpdPow2(10000, &a)
	EpdMultiply(&a, 2)
	if a.Exponent != 10001 || a.Value != 1 {
		t.Fatal(a)
	}
	EpdPow2(-10000, &b)
	EpdMultiply2(&a, &b)
	if math.Ldexp(a.Value, a.Exponent) != 2 {
		t.Fatal(a)
	}
	EpdConvert(math.SmallestNonzeroFloat64, &a)
	if math.Ldexp(a.Value, a.Exponent) != math.SmallestNonzeroFloat64 {
		t.Fatal(a)
	}
	EpdMakeInf(&a, 0)
	EpdMakeZero(&b, 0)
	EpdMultiply2(&a, &b)
	if EpdIsNan(&a) != 1 {
		t.Fatal(a)
	}
}
func TestDecimalAndFormatting(t *testing.T) {
	var a, b EpDouble
	EpdPow2Decimal(10000, &a)
	if a.Exponent != 3010 || a.Value < 1 || a.Value >= 10 {
		t.Fatal(a)
	}
	b = EpDouble{2, 5}
	a = EpDouble{3, 8}
	EpdMultiply2Decimal(&a, &b)
	if a.Value != 6 || a.Exponent != 13 {
		t.Fatal(a)
	}
	EpdConvert(123, &a)
	if EpdGetString(&a) != "1.230000e+02" {
		t.Fatal(EpdGetString(&a))
	}
	EpdMakeZero(&a, 1)
	if !math.Signbit(a.Value) {
		t.Fatal(a)
	}
	EpdMakeInf(&a, 1)
	if EpdGetString(&a) != "-Inf" {
		t.Fatal(a)
	}
}
func TestAliasingAndCopy(t *testing.T) {
	var a, b EpDouble
	EpdConvert(7, &a)
	EpdConvert(2, &b)
	EpdSubtract3(&a, &b, &b)
	if math.Ldexp(b.Value, b.Exponent) != 5 {
		t.Fatal(b)
	}
	EpdCopy(&a, &b)
	if EpdCmp(&a, &b) != 0 {
		t.Fatal(a, b)
	}
	EpdAdd(&b, 1)
	if EpdCmp(&a, &b) == 0 {
		t.Fatal(a, b)
	}
}

func TestDecimalExtremeReciprocal(t *testing.T) {
	var a, b EpDouble
	EpdPow2Decimal(10000, &a)
	EpdPow2Decimal(-10000, &b)
	EpdMultiply2Decimal(&a, &b)
	if math.Abs(a.Value*math.Pow10(a.Exponent)-1) > 2e-14 {
		t.Fatal(a)
	}
}
