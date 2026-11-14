package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSimulationPatternHashAndMiter(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendAnd(p, a, b^1)
	p.VSimsPi = vec.Vec_WrdAlloc(2)
	vec.Vec_WrdPush(p.VSimsPi, 0xaaaaaaaaaaaaaaaa)
	vec.Vec_WrdPush(p.VSimsPi, 0xaaaaaaaaaaaaaaaa)
	sims := Gia_ManSimPatSim(p)
	var zeros, ones int32
	if unique := Gia_ManSimPatHashPatterns(p, 1, sims, &zeros, &ones); unique != 2 || zeros != 1 || ones != 0 {
		t.Fatalf("unique=%d zero=%d one=%d", unique, zeros, ones)
	}
	miter := Gia_ManSimPatGenMiter(p, sims)
	if Gia_ManCoNum(miter) != 1 {
		t.Fatal("expected one candidate output")
	}
	out, err := miter.Evaluate([]bool{true, false})
	if err != nil || !out[0] {
		t.Fatalf("candidate miter output %v, error %v", out, err)
	}
	out, err = miter.Evaluate([]bool{true, true})
	if err != nil || out[0] {
		t.Fatalf("candidate miter output %v, error %v", out, err)
	}
}
