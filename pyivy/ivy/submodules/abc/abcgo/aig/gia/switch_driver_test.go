package gia

import (
	"math"
	"testing"
)

func TestSwitchProbabilityDriver(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendCo(p, x^1)
	v := Gia_ManComputeSwitchProbs(p, 48, 16, 1)
	get := func(lit int32) float64 { return float64(math.Float32frombits(uint32(Vec_IntEntry(v, int(lit>>1))))) }
	if math.Abs(get(a)-.5) > .05 || math.Abs(get(x)-.25) > .05 || math.Abs(get(y)-.75) > .05 {
		t.Fatalf("probabilities a=%v and=%v out=%v", get(a), get(x), get(y))
	}
	v = Gia_ManComputeSwitchProbs(p, 48, 16, 0)
	if math.Abs(float64(math.Float32frombits(uint32(Vec_IntEntry(v, int(a>>1)))))-.5) > .05 {
		t.Fatal("unexpected input transition probability")
	}
}
