package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/sat/bsat"
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

func TestCubeReachabilityInterruptedFeasibilityStops(t *testing.T) {
	g := Gia_ManStart(32)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCi(g)
	Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, b, c))
	Gia_ManAppendCo(g, a)
	Gia_ManAppendCo(g, a)
	Gia_ManSetRegNum(g, 2)
	p := Gia_ManAreCreate(g)
	defer Gia_ManAreFree(p)
	p.fMiter = 1
	Gia_ManAreCubeProcess(p, Gia_ManAreCreateStaInit(p))
	Gia_ManAreDeriveCexSatStart(p)
	calls := 0
	p.pSat.SetRuntimeDeadline(func() int64 { calls++; return 2 }, 1)
	stopped := Gia_ManAreDeriveNexts(p, Gia_Int2Ptr(1))
	if stopped != 1 || p.fStopped != 1 || p.pTarget != nil || p.nStas != 2 {
		t.Fatalf("interrupted proof: return%d stopped%d target%v states%d", stopped, p.fStopped, p.pTarget, p.nStas)
	}
	if calls < 1 || p.nRecCalls != 2 {
		t.Fatalf("proof stop did not unwind: proof clock reads%d recursive calls%d", calls, p.nRecCalls)
	}
	if len(p.cofTruth) != 0 {
		t.Fatal("interrupted query cached an unproved Boolean fact")
	}
}

func TestCubeReachabilityProofCachePolarities(t *testing.T) {
	for _, real := range []bool{false, true} {
		for mode := 0; mode < 2; mode++ {
			t.Run(fmt.Sprintf("real%t/value%d", real, mode), func(t *testing.T) {
				p := Gia_ManAreCreate(cubeReviewSemanticZero(real, mode))
				defer Gia_ManAreFree(p)
				Gia_ManAreCubeProcess(p, Gia_ManAreCreateStaInit(p))
				Gia_ManAreDeriveNexts(p, Gia_Int2Ptr(1))
				lit := Gia_ObjValue(Gia_ManRi(p.pAig, 0))
				if lit < 2 || p.pSat == nil {
					t.Fatal("semantic proof path was not exercised")
				}
				for value := uint8(0); value < 2; value++ {
					want := bsat.Unsatisfiable
					if int(value) == mode {
						want = bsat.Satisfiable
					}
					if got := giaAreCofactorPossible(p, 1, lit, value); got != want {
						t.Fatalf("polarity%d status%v want%v", value, got, want)
					}
				}
				reads := 0
				p.pSat.SetRuntimeDeadline(func() int64 { reads++; return 2 }, 1)
				for value := uint8(0); value < 2; value++ {
					want := bsat.Unsatisfiable
					if int(value) == mode {
						want = bsat.Satisfiable
					}
					if got := giaAreCofactorPossible(p, 1, lit, value); got != want {
						t.Fatalf("cached polarity%d status%v want%v", value, got, want)
					}
				}
				if reads != 0 {
					t.Fatal("cached proof unexpectedly invoked the expired solver")
				}
			})
		}
	}
}

func TestCubeReachabilityProofCacheGraphLifetime(t *testing.T) {
	p := Gia_ManAreCreate(cubeReviewSemanticZero(true, 0))
	defer Gia_ManAreFree(p)
	Gia_ManAreCubeProcess(p, Gia_ManAreCreateStaInit(p))
	Gia_ManAreDeriveNexts(p, Gia_Int2Ptr(1))
	oldGraph, oldFacts := p.pNew, p.cofTruth
	if len(oldFacts) == 0 {
		t.Fatal("missing initial proof facts")
	}
	// Exercise the real graph-replacement threshold with harmless dangling
	// buffers; they do not alter source logic or require a fabricated counter.
	for p.pNew.nObjs <= 1000000 {
		Gia_ManAppendBuf(p.pNew, 0)
	}
	Gia_ManAreDeriveNexts(p, Gia_Int2Ptr(1))
	if p.pNew == oldGraph || p.pNew.nObjs > 1000 {
		t.Fatal("temporary cofactor graph was not replaced")
	}
	lit := Gia_ObjValue(Gia_ManRi(p.pAig, 0))
	// Old graph owners must not be able to poison proofs in the new literal
	// namespace, even when the same object number is reused.
	oldFacts[lit] = giaAreCofTruth{known: 3, possible: 2}
	if got := giaAreCofactorPossible(p, 1, lit, 1); got != bsat.Unsatisfiable {
		t.Fatal("old graph facts leaked into the replacement graph")
	}
}

func TestCubeReachabilityOnlineSATPreservesSourceScratch(t *testing.T) {
	p := Gia_ManAreCreate(cubeReviewSemanticZero(true, 0))
	defer Gia_ManAreFree(p)
	saved := make([]uint32, len(p.pAig.pObjs))
	for i, obj := range p.pAig.pObjs {
		obj.Value = uint32(i*17 + 99)
		saved[i] = obj.Value
	}
	Gia_ManAreDeriveCexSatStart(p)
	for i, obj := range p.pAig.pObjs {
		if obj.Value != saved[i] {
			t.Fatalf("online typed SAT initialization changed source Value at%d", i)
		}
	}
}
