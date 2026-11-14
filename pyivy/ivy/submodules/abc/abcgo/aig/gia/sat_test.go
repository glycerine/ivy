package gia

import (
	"context"
	"github.com/glycerine/abcgo/misc/util"
	ceccore "github.com/glycerine/abcgo/proof/cec/core"
	"testing"
)

func TestGiaSatUsesAllOutputConjunction(t *testing.T) {
	p := Gia_ManStart(4)
	a := Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, a^1)
	if status := Gia_ManSolveSat(p); status != 1 || p.PCexComb != nil {
		t.Fatal("contradictory outputs were not UNSAT", status)
	}
	sat := Gia_ManStart(6)
	x, y := Gia_ManAppendCi(sat), Gia_ManAppendCi(sat)
	Gia_ManAppendCo(sat, x)
	Gia_ManAppendCo(sat, y)
	before := sat.nObjs
	result, err := Gia_ManSolveSatResult(sat)
	if err != nil || result.Status != ceccore.Different || result.Counterexample == nil {
		t.Fatal(err, result)
	}
	if sat.nObjs != before || sat.PCexComb == nil || sat.PCexComb.PData[0]&3 != 3 {
		t.Fatal("invalid witness or mutated graph")
	}
	if sat.pObjs[x>>1].fMark0 != 1 || sat.pObjs[y>>1].fMark0 != 1 {
		t.Fatal("node marks not recorded")
	}
	for _, truth := range []int32{0, 1} {
		constant := Gia_ManStart(2)
		Gia_ManAppendCo(constant, truth)
		if result := Gia_ManSolveSat(constant); result != 1-truth {
			t.Fatal("constant status", truth, result)
		}
	}
}

func TestGiaCecRealSolverAndCounterexampleOwnership(t *testing.T) {
	left, right := Gia_ManStart(8), Gia_ManStart(8)
	a, b := Gia_ManAppendCi(left), Gia_ManAppendCi(left)
	x, y := Gia_ManAppendCi(right), Gia_ManAppendCi(right)
	Gia_ManAppendCo(left, Gia_ManAppendXor(left, a, b))
	Gia_ManAppendCo(right, Gia_ManAppendXor(right, y, x))
	if status := Cec_ManVerifyTwo(left, right, 0); status != 1 {
		t.Fatal("equivalent XORs", status)
	}
	different := Gia_ManStart(8)
	u, v := Gia_ManAppendCi(different), Gia_ManAppendCi(different)
	Gia_ManAppendCo(different, Gia_ManAppendAnd(different, u, v))
	existing := util.Abc_CexAlloc(0, 2, 1)
	existing.PData[0] = 2
	different.PCexComb = existing
	if status := Cec_ManVerifyTwo(left, different, 0); status != 0 || left.PCexComb == nil || different.PCexComb != existing {
		t.Fatal("missing difference witness", status)
	}
	bits := left.PCexComb.PData[0]
	inputs := []bool{bits&1 != 0, bits&2 != 0}
	l, _ := left.Evaluate(inputs)
	r, _ := different.Evaluate(inputs)
	if l[0] == r[0] {
		t.Fatal("witness does not distinguish networks")
	}
	other := different.PCexComb.PData[0]
	left.PCexComb.PData[0] ^= 1
	if different.PCexComb.PData[0] != other {
		t.Fatal("independent managers share mutable witness storage")
	}
}

func TestGiaSatCancellationAndExplicitFreeState(t *testing.T) {
	p := Gia_ManStart(6)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, b)
	p.nRegs = 1
	if _, err := Gia_ManSolveSatResult(p); err == nil {
		t.Fatal("sequential network accepted as combinational proof")
	}
	options := ceccore.DefaultSATOptions()
	options.TreatRegistersAsInputs = true
	result, err := Gia_ManSolveSatResult(p, options)
	if err != nil || result.Status != ceccore.Different {
		t.Fatal(err, result)
	}
	if p.PCexComb.NRegs != 1 || p.PCexComb.NPis != 1 || p.PCexComb.PData[0]&3 != 3 {
		t.Fatal("wrong initial-state/input witness layout")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	options.Context = ctx
	result, err = Gia_ManSolveSatResult(p, options)
	if err != nil || result.Status != ceccore.Undecided || p.PCexComb != nil {
		t.Fatal(err, result)
	}
}
