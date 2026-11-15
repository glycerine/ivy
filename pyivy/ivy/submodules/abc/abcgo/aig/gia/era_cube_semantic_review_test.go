package gia

import (
	"fmt"
	"testing"
)

func cubeReviewSemanticZero(real bool, riMode int) *Gia_Man_t {
	g := Gia_ManStart(64)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	ro := Gia_ManAppendCi(g)
	// Distributivity: (a&b)|(a&c) == a&(b|c). Strashing alone does not
	// identify these expressions, so their XOR is nonconstant structurally.
	x := Gia_ManAppendOr(g, Gia_ManAppendAnd(g, a, b), Gia_ManAppendAnd(g, a, c))
	y := Gia_ManAppendAnd(g, a, Gia_ManAppendOr(g, b, c))
	z := Gia_ManAppendXor(g, x, y)
	if real {
		z = Gia_ManAppendXorReal(g, x, y)
	}
	if riMode < 0 {
		Gia_ManAppendCo(g, z)
		Gia_ManAppendCo(g, ro)
	} else {
		Gia_ManAppendCo(g, ro)
		Gia_ManAppendCo(g, z^int32(riMode))
	}
	Gia_ManSetRegNum(g, 1)
	return g
}

func TestCubeReachabilitySemanticZeroMiter(t *testing.T) {
	for _, real := range []bool{false, true} {
		for _, mode := range []int{-1, 0} {
			t.Run(fmt.Sprintf("real%t/mode%d", real, mode), func(t *testing.T) {
				defer func() {
					if failure := recover(); failure != nil {
						t.Errorf("safe semantic-zero machine panicked: %v", failure)
					}
				}()
				g := cubeReviewSemanticZero(real, mode)
				defer Gia_ManStop(g)
				for mask := 0; mask < 8; mask++ {
					out, err := g.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, false})
					if err != nil || out[0] || out[1] {
						t.Fatalf("scalar zero check%v %v", out, err)
					}
				}
				var got int32
				hashReviewStdout(t, func() { got = Gia_ManArePerform(g, 100, 1, 0) })
				if got != 1 || g.pCexSeq != nil {
					t.Fatalf("safe semantic-zero machine returned%d trace%+v", got, g.pCexSeq)
				}
			})
		}
	}
}

func TestCubeReachabilitySemanticConstantSuccessors(t *testing.T) {
	for _, real := range []bool{false, true} {
		for mode := 0; mode < 2; mode++ {
			t.Run(fmt.Sprintf("real%t/value%d", real, mode), func(t *testing.T) {
				g := cubeReviewSemanticZero(real, mode)
				p := Gia_ManAreCreate(g)
				defer Gia_ManAreFree(p)
				initial := Gia_ManAreCreateStaInit(p)
				Gia_ManAreCubeProcess(p, initial)
				p.iStaCur = 1
				Gia_ManAreDeriveNexts(p, Gia_Int2Ptr(1))
				if mode == 0 {
					if p.nStas != 2 || Gia_StaIsUnused(initial) != 0 {
						t.Fatalf("constant-zero RI fabricated a successor: states%d unused%d", p.nStas, Gia_StaIsUnused(initial))
					}
				} else {
					if p.nStas != 3 || Gia_StaHasValue1(p.states[2], 0) != 1 || Gia_StaHasValue0(p.states[2], 0) != 0 || Gia_StaIsUnused(initial) != 0 {
						t.Fatal("constant-one RI was treated as a don't-care")
					}
				}
			})
		}
	}
}
