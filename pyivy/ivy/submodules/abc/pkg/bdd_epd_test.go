package abc

// Go-only tests for src/bdd/epd/epd.h and epd.c.

import (
	"errors"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"testing"
)

func epdBig(number *EPDouble) *big.Float {
	if number == nil || number.IsNaNOrInf() {
		return nil
	}
	value := new(big.Float).SetPrec(256).SetFloat64(number.value)
	if number.mode == EPDBinary {
		return value.SetMantExp(value, number.exponent)
	}
	exponent := number.exponent
	if exponent < 0 {
		exponent = -exponent
	}
	power := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil)
	factor := new(big.Float).SetPrec(256).SetInt(power)
	if number.exponent < 0 {
		value.Quo(value, factor)
	} else {
		value.Mul(value, factor)
	}
	return value
}

func epdCloseToBig(t *testing.T, got *EPDouble, want *big.Float) {
	t.Helper()
	actual := epdBig(got)
	if actual == nil || want == nil {
		t.Fatalf("nonfinite comparison: %v %v", actual, want)
	}
	difference := new(big.Float).SetPrec(256).Sub(actual, want)
	difference.Abs(difference)
	scale := new(big.Float).SetPrec(256).Abs(want)
	if scale.Cmp(big.NewFloat(1)) < 0 {
		scale.SetInt64(1)
	}
	tolerance := new(big.Float).SetPrec(256).Mul(scale, big.NewFloat(4e-14))
	if difference.Cmp(tolerance) > 0 {
		t.Fatalf("got %s want %s difference %s", actual.Text('g', 18), want.Text('g', 18), difference.Text('g', 8))
	}
}

func TestEPDConstructionNormalizationAndString(t *testing.T) {
	for _, value := range []float64{0, math.Copysign(0, -1), 1, -1, 0.5, 3.25, math.SmallestNonzeroFloat64, math.MaxFloat64} {
		number, err := NewEPDouble(value, EPDLimits{})
		if err != nil {
			t.Fatal(value, err)
		}
		if value == 0 {
			if !number.IsZero() || math.Signbit(value) != math.Signbit(number.value) {
				t.Fatalf("zero sign %x %x", math.Float64bits(value), math.Float64bits(number.value))
			}
		} else if math.Abs(number.value) < 1 || math.Abs(number.value) >= 2 {
			t.Fatalf("not binary-normalized: %g -> %+v", value, number)
		}
		epdCloseToBig(t, number, new(big.Float).SetPrec(256).SetFloat64(value))
		if !strings.Contains(number.String(), "e") {
			t.Fatal(number.String())
		}
	}
	decimal, err := NewDecimalEPDouble(-12345.5, EPDLimits{})
	if err != nil || decimal.value != -1.23455 || decimal.exponent != 4 {
		t.Fatalf("decimal %+v %v", decimal, err)
	}
	if EPDGetBinaryExponent(1) != EPDMaxBinaryExponent || EPDGetDecimalExponent(0.00125) != -3 {
		t.Fatal("exponent helpers")
	}
	if EPDIsInfDouble(math.Inf(1)) != 1 || EPDIsInfDouble(math.Inf(-1)) != -1 || !EPDIsNaNDouble(math.NaN()) || !EPDIsNaNOrInfDouble(math.Inf(1)) {
		t.Fatal("classification")
	}
	decimalSubnormal, err := NewDecimalEPDouble(math.SmallestNonzeroFloat64, EPDLimits{})
	if err != nil || decimalSubnormal.IsNaNOrInf() || decimalSubnormal.exponent != -324 || decimalSubnormal.value < 4.9 || decimalSubnormal.value > 5 {
		t.Fatalf("decimal subnormal %+v %v", decimalSubnormal, err)
	}
	epdCloseToBig(t, decimalSubnormal, new(big.Float).SetPrec(256).SetFloat64(math.SmallestNonzeroFloat64))
}

func TestEPDArithmeticAndExceptionalCases(t *testing.T) {
	first, _ := NewEPDouble(3.5, EPDLimits{})
	second, _ := NewEPDouble(-2.25, EPDLimits{})
	product, err := EPDProduct(first, second)
	if err != nil {
		t.Fatal(err)
	}
	epdCloseToBig(t, product, big.NewFloat(-7.875))
	quotient, err := EPDQuotient(first, second)
	if err != nil {
		t.Fatal(err)
	}
	epdCloseToBig(t, quotient, new(big.Float).SetPrec(256).Quo(big.NewFloat(3.5), big.NewFloat(-2.25)))
	sum, _ := EPDSum(first, second)
	epdCloseToBig(t, sum, big.NewFloat(1.25))
	difference, _ := EPDDifference(first, second)
	epdCloseToBig(t, difference, big.NewFloat(5.75))
	if err = first.MultiplyFloat(4); err != nil {
		t.Fatal(err)
	}
	epdCloseToBig(t, first, big.NewFloat(14))
	if err = first.DivideFloat(7); err != nil {
		t.Fatal(err)
	}
	epdCloseToBig(t, first, big.NewFloat(2))
	zero, _ := NewEPDouble(0, EPDLimits{})
	if err = first.Divide(zero); err != nil || !first.IsNaN() {
		t.Fatalf("divide zero %+v %v", first, err)
	}
	inf, _ := NewEPDouble(math.Inf(1), EPDLimits{})
	negativeInf, _ := NewEPDouble(math.Inf(-1), EPDLimits{})
	if err = inf.Add(negativeInf); err != nil || !inf.IsNaN() {
		t.Fatalf("infinity cancellation %+v %v", inf, err)
	}
	finite, _ := NewEPDouble(-2, EPDLimits{})
	if err = finite.Multiply(negativeInf); err != nil || !finite.IsInf() || math.Signbit(finite.value) {
		t.Fatalf("infinity product %+v %v", finite, err)
	}
	negativeZero, _ := NewEPDouble(math.Copysign(0, -1), EPDLimits{})
	positiveInf, _ := NewEPDouble(math.Inf(1), EPDLimits{})
	if err = positiveInf.Divide(negativeZero); err != nil || !positiveInf.IsInf() || !math.Signbit(positiveInf.value) {
		t.Fatalf("infinity divided by signed zero %+v %v", positiveInf, err)
	}
	decimalLeft, _ := NewDecimalEPDouble(2.5e200, EPDLimits{})
	decimalRight, _ := NewDecimalEPDouble(-4e-150, EPDLimits{})
	decimalProduct, err := EPDProduct(decimalLeft, decimalRight)
	if err != nil || decimalProduct.mode != EPDDecimal || math.Abs(decimalProduct.value+1) > 1e-15 || decimalProduct.exponent != 51 {
		t.Fatalf("decimal product %+v %v", decimalProduct, err)
	}
	special, _ := NewEPDouble(1, EPDLimits{})
	if err = special.MakeZero(true); err != nil || !special.IsZero() || !math.Signbit(special.value) {
		t.Fatalf("negative zero %+v %v", special, err)
	}
	if err = special.MakeInf(true); err != nil || !special.IsInf() || !math.Signbit(special.value) {
		t.Fatalf("negative infinity %+v %v", special, err)
	}
	if err = special.MakeNaN(); err != nil || !special.IsNaN() {
		t.Fatalf("nan %+v %v", special, err)
	}
}

func TestEPDSpecialValueMatrixAndFormatting(t *testing.T) {
	limits := EPDLimits{}
	for _, test := range []struct {
		name      string
		operation func(*EPDouble, *EPDouble) error
		left      float64
		right     float64
		wantNaN   bool
		wantInf   int
	}{
		{"nan product", (*EPDouble).Multiply, math.NaN(), 1, true, 0},
		{"zero infinite product", (*EPDouble).Multiply, math.Copysign(0, -1), math.Inf(-1), false, 1},
		{"infinite quotient", (*EPDouble).Divide, math.Inf(1), math.Copysign(0, -1), false, -1},
		{"infinite cancellation", (*EPDouble).Add, math.Inf(1), math.Inf(-1), true, 0},
		{"same infinite subtraction", (*EPDouble).Subtract, math.Inf(-1), math.Inf(-1), true, 0},
		{"finite minus infinity", (*EPDouble).Subtract, 1, math.Inf(1), false, -1},
	} {
		t.Run(test.name, func(t *testing.T) {
			left, _ := NewEPDouble(test.left, limits)
			right, _ := NewEPDouble(test.right, limits)
			if err := test.operation(left, right); err != nil {
				t.Fatal(err)
			}
			if left.IsNaN() != test.wantNaN || EPDIsInfDouble(left.value) != test.wantInf {
				t.Fatalf("result %+v", left)
			}
		})
	}
	for value, want := range map[float64]string{1: "1.000000e+00", 0.00125: "1.250000e-03"} {
		number, _ := NewEPDouble(value, limits)
		if got := number.String(); got != want {
			t.Fatalf("String(%g) = %q, want %q", value, got, want)
		}
	}
	for value, want := range map[float64]string{math.Inf(1): "Inf", math.Inf(-1): "-Inf"} {
		number, _ := NewEPDouble(value, limits)
		if got := number.String(); got != want {
			t.Fatalf("String(%g) = %q, want %q", value, got, want)
		}
	}
	nan, _ := NewEPDouble(math.Float64frombits(0x7ff0000000000001), limits)
	if !nan.IsNaN() || nan.String() != "NaN" || !EPDIsNaNDouble(nan.value) {
		t.Fatalf("noncanonical NaN %+v", nan)
	}
}

func TestEPDPowersDecimalPartsCloneAndLifecycle(t *testing.T) {
	for _, exponent := range []int{-10000, -1024, -1, 0, 1, 1023, 1024, 10000} {
		number, err := EPDPow2(exponent, EPDLimits{MaxAbsExponent: 20000})
		if err != nil || number.value != 1 || number.exponent != exponent {
			t.Fatalf("pow %d %+v %v", exponent, number, err)
		}
		value, decimalExponent, err := number.DecimalParts()
		if err != nil || math.Abs(math.Log10(math.Abs(value))+float64(decimalExponent)-float64(exponent)*math.Log10(2)) > 1e-12 {
			t.Fatalf("decimal %d %g %d %v", exponent, value, decimalExponent, err)
		}
		decimal, err := EPDPow2Decimal(exponent, EPDLimits{MaxAbsExponent: 20000})
		if err != nil || decimal.mode != EPDDecimal || decimal.exponent != decimalExponent || math.Abs(decimal.value-value) > 1e-12 {
			t.Fatalf("decimal power %d %+v %v", exponent, decimal, err)
		}
	}
	number, _ := NewEPDouble(5, EPDLimits{})
	clone, _ := number.Clone()
	equal, _ := number.EqualStorage(clone)
	if !equal || clone == number {
		t.Fatal("clone")
	}
	destination, _ := NewEPDouble(1, EPDLimits{})
	if err := destination.CopyFrom(number); err != nil {
		t.Fatal(err)
	}
	equal, _ = number.EqualStorage(destination)
	if !equal {
		t.Fatal("copy")
	}
	if err := destination.Convert(0.125); err != nil || destination.value != 1 || destination.exponent != -3 || destination.mode != EPDBinary {
		t.Fatalf("convert %+v %v", destination, err)
	}
	destination.value, destination.exponent = 125, 0
	if err := destination.NormalizeDecimal(); err != nil || destination.value != 1.25 || destination.exponent != 2 || destination.mode != EPDDecimal {
		t.Fatalf("decimal normalize %+v %v", destination, err)
	}
	if err := destination.NormalizeBinary(); err != nil || destination.mode != EPDBinary {
		t.Fatalf("binary normalize %+v %v", destination, err)
	}
	clone.Close()
	clone.Close()
	if _, err := clone.Significand(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestEPDErrorsLimitsCancellationAndAtomicity(t *testing.T) {
	if _, err := NewEPDouble(1, EPDLimits{MaxAbsExponent: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := EPDPow2(11, EPDLimits{MaxAbsExponent: 10}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	cancelled, err := NewEPDouble(1, EPDLimits{Stop: func() bool { return true }})
	if cancelled != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(cancelled, err)
	}
	number, _ := NewEPDouble(1, EPDLimits{MaxAbsExponent: 10})
	large, _ := EPDPow2(10, EPDLimits{MaxAbsExponent: 10})
	before := *number
	if err = number.Multiply(large); err != nil {
		t.Fatal(err)
	}
	if err = number.Multiply(large); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if number.exponent != before.exponent+10 {
		t.Fatal("failed mutation was not atomic")
	}
	decimal, _ := NewDecimalEPDouble(1, EPDLimits{})
	if err = number.Add(decimal); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	extremeLimits := EPDLimits{MaxAbsExponent: math.MaxInt}
	dominant, err := EPDPow2(math.MaxInt, extremeLimits)
	if err != nil {
		t.Fatal(err)
	}
	negligible, err := EPDPow2(-math.MaxInt, extremeLimits)
	if err != nil {
		t.Fatal(err)
	}
	if err = dominant.Add(negligible); err != nil || dominant.value != 1 || dominant.exponent != math.MaxInt {
		t.Fatalf("extreme exponent alignment %+v %v", dominant, err)
	}
}

func TestEPDIndependentBigFloatModel(t *testing.T) {
	random := rand.New(rand.NewSource(17))
	for execution := 0; execution < 5000; execution++ {
		leftValue := math.Ldexp(random.Float64()*2-1, random.Intn(1201)-600)
		rightValue := math.Ldexp(random.Float64()*2-1, random.Intn(1201)-600)
		left, err := NewEPDouble(leftValue, EPDLimits{MaxAbsExponent: 10000})
		if err != nil {
			t.Fatal(execution, err)
		}
		right, err := NewEPDouble(rightValue, EPDLimits{MaxAbsExponent: 10000})
		if err != nil {
			t.Fatal(execution, err)
		}
		leftBig := new(big.Float).SetPrec(256).SetFloat64(leftValue)
		rightBig := new(big.Float).SetPrec(256).SetFloat64(rightValue)
		var result *EPDouble
		var want *big.Float
		switch execution & 3 {
		case 0:
			result, err = EPDProduct(left, right)
			want = new(big.Float).SetPrec(256).Mul(leftBig, rightBig)
		case 1:
			if rightValue == 0 {
				continue
			}
			result, err = EPDQuotient(left, right)
			want = new(big.Float).SetPrec(256).Quo(leftBig, rightBig)
		case 2:
			result, err = EPDSum(left, right)
			want = new(big.Float).SetPrec(256).Add(leftBig, rightBig)
		default:
			result, err = EPDDifference(left, right)
			want = new(big.Float).SetPrec(256).Sub(leftBig, rightBig)
		}
		if err != nil {
			t.Fatal(execution, err)
		}
		epdCloseToBig(t, result, want)
	}
}

func FuzzEPDBinaryArithmetic(f *testing.F) {
	f.Add(uint64(0x3ff0000000000000), uint64(0x4000000000000000), uint8(0))
	f.Add(uint64(0x0010000000000000), uint64(0x7fefffffffffffff), uint8(3))
	f.Fuzz(func(t *testing.T, leftBits, rightBits uint64, operation uint8) {
		leftValue, rightValue := math.Float64frombits(leftBits), math.Float64frombits(rightBits)
		if math.IsNaN(leftValue) || math.IsInf(leftValue, 0) || math.IsNaN(rightValue) || math.IsInf(rightValue, 0) {
			return
		}
		left, err := NewEPDouble(leftValue, EPDLimits{MaxAbsExponent: 100000})
		if err != nil {
			return
		}
		right, err := NewEPDouble(rightValue, EPDLimits{MaxAbsExponent: 100000})
		if err != nil {
			return
		}
		leftBig := new(big.Float).SetPrec(256).SetFloat64(leftValue)
		rightBig := new(big.Float).SetPrec(256).SetFloat64(rightValue)
		var result *EPDouble
		var want *big.Float
		switch operation & 3 {
		case 0:
			result, err = EPDProduct(left, right)
			want = new(big.Float).SetPrec(256).Mul(leftBig, rightBig)
		case 1:
			if rightValue == 0 {
				return
			}
			result, err = EPDQuotient(left, right)
			want = new(big.Float).SetPrec(256).Quo(leftBig, rightBig)
		case 2:
			result, err = EPDSum(left, right)
			want = new(big.Float).SetPrec(256).Add(leftBig, rightBig)
		default:
			result, err = EPDDifference(left, right)
			want = new(big.Float).SetPrec(256).Sub(leftBig, rightBig)
		}
		if err != nil {
			return
		}
		epdCloseToBig(t, result, want)
	})
}
