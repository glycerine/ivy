package gia

import (
	"fmt"
	"testing"
)

func cubeReviewCounter(bits int) *Gia_Man_t {
	g := Gia_ManStart(int32(16 * bits))
	r, next := make([]int32, bits), make([]int32, bits)
	for i := range r {
		r[i] = Gia_ManAppendCi(g)
	}
	Gia_ManHashAlloc(g)
	carry := int32(1)
	for i := range r {
		next[i] = Gia_ManHashXor(g, r[i], carry)
		carry = Gia_ManHashAnd(g, carry, r[i])
	}
	Gia_ManAppendCo(g, r[bits-1])
	for _, lit := range next {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManHashStop(g)
	Gia_ManSetRegNum(g, int32(bits))
	return g
}

func TestCubeReachabilityCounterLimits(t *testing.T) {
	for _, limit := range []int32{0, 1, 2, 3, 4, 10, 32, 33, 34, 64, 65, 66, 100} {
		for miter := int32(0); miter < 2; miter++ {
			t.Run(fmt.Sprintf("limit%d/miter%d", limit, miter), func(t *testing.T) {
				g := cubeReviewCounter(6)
				defer Gia_ManStop(g)
				// The zero-input counter has 64 exact reachable states. Counts include
				// the reserved record and are checked after generating a successor.
				want := int32(-1)
				if miter == 0 && limit >= 65 {
					want = 1
				}
				if miter != 0 && limit >= 34 {
					want = 0
				}
				var got int32
				hashReviewStdout(t, func() { got = Gia_ManArePerform(g, limit, miter, 0) })
				if got != want {
					t.Fatalf("result%d want%d", got, want)
				}
				if want == 0 {
					if g.pCexSeq == nil || g.pCexSeq.IFrame != 32 || Gia_ManVerifyCex(g, g.pCexSeq, 0) == 0 {
						t.Fatalf("counter failure trace%+v", g.pCexSeq)
					}
				} else if g.pCexSeq != nil {
					t.Fatal("resource/safe result has a failure trace")
				}
			})
		}
	}
}

func TestCubeReachabilityZeroBudgetClearsOldTrace(t *testing.T) {
	g := Gia_ManStart(8)
	defer Gia_ManStop(g)
	a, r := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	Gia_ManAppendCo(g, r)
	Gia_ManSetRegNum(g, 1)
	hashReviewStdout(t, func() {
		if Gia_ManArePerform(g, 100, 1, 0) != 0 {
			t.Fatal("missing setup violation")
		}
	})
	if g.pCexSeq == nil {
		t.Fatal("missing setup trace")
	}
	if got := Gia_ManArePerform(g, 0, 1, 0); got != -1 {
		t.Fatalf("zero budget result%d", got)
	}
	if g.pCexSeq != nil {
		t.Fatal("zero-budget stop retained the previous run's failure trace")
	}
}

func TestCubeReachabilityRecursiveCallLimit(t *testing.T) {
	g := Gia_ManStart(128)
	parity := int32(0)
	var inputs []int32
	for i := 0; i < 20; i++ {
		inputs = append(inputs, Gia_ManAppendCi(g))
	}
	Gia_ManAppendCi(g)
	Gia_ManAppendCi(g)
	Gia_ManHashAlloc(g)
	for _, lit := range inputs {
		parity = Gia_ManHashXor(g, parity, lit)
	}
	Gia_ManAppendCo(g, 0)
	Gia_ManAppendCo(g, parity)
	Gia_ManAppendCo(g, parity)
	Gia_ManHashStop(g)
	Gia_ManSetRegNum(g, 2)
	p := Gia_ManAreCreate(g)
	defer Gia_ManAreFree(p)
	initial := Gia_ManAreCreateStaInit(p)
	Gia_ManAreCubeProcess(p, initial)
	// The two correlated parity RIs force PI splitting until all 20 values
	// are bound, exceeding one million recursive calls despite only two states.
	hashReviewStdout(t, func() { Gia_ManAreDeriveNexts(p, Gia_Int2Ptr(1)) })
	if p.nRecCalls < MAX_CALL_NUM || p.fStopped != 1 || p.pTarget != nil || p.pSat != nil {
		t.Fatalf("call stop: calls%d stopped%d target%v", p.nRecCalls, p.fStopped, p.pTarget)
	}
	// The recursive stop must restore cofactor literals while unwinding. Read
	// them through an independent scalar evaluator, with no further recursion.
	for _, mask := range []uint32{0, 1, 3, 7, 0xfffff, 0x55555, 0xaaaaa} {
		in := make([]bool, 22)
		want := false
		for i := 0; i < 20; i++ {
			in[i] = mask&(1<<i) != 0
			want = want != in[i]
		}
		values, _, err := p.pNew.evaluateObjects(in)
		if err != nil {
			t.Fatal(err)
		}
		for i := int32(0); i < 2; i++ {
			lit := Gia_ObjValue(Gia_ManRi(g, i))
			if got := values[lit>>1] != (lit&1 != 0); got != want {
				t.Fatalf("input%x RI%d failed to restore parity", mask, i)
			}
		}
	}
}
