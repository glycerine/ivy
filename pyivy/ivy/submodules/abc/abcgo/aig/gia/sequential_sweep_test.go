package gia

import "testing"

func TestSequentialSweepConstantsAndEquivalentRegisters(t *testing.T) {
	p := Gia_ManStart(30)
	a, r0, r1, r2 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManHashAlloc(p)
	out := Gia_ManHashOr(p, r0, Gia_ManHashXor(p, r1, r2))
	Gia_ManAppendCo(p, out)
	Gia_ManAppendCo(p, 0)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, a)
	Gia_ManHashStop(p)
	Gia_ManSetRegNum(p, 3)
	simplified := Gia_ManSeqStructSweep(p, 1, 1, 0)
	if Gia_ManRegNum(simplified) != 0 {
		t.Fatalf("registers not reduced: %d", Gia_ManRegNum(simplified))
	}
	for _, bit := range []bool{false, true} {
		got, e := simplified.Evaluate([]bool{bit})
		if e != nil || got[0] {
			t.Fatalf("wrong output %v %v", got, e)
		}
	}
}
func TestTernarySimulationToggleAndUnknown(t *testing.T) {
	p := Gia_ManStart(20)
	a, r0, r1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, r0)
	Gia_ManAppendCo(p, r1)
	Gia_ManAppendCo(p, r0^1)
	Gia_ManAppendCo(p, a)
	Gia_ManSetRegNum(p, 2)
	states := giaTernaryStates(p)
	if len(states) != 4 || states[0][0] != 0 || states[1][0] != 1 || states[2][0] != 0 || states[1][1] != 2 || p.nTerLoop != 1 {
		t.Fatalf("states %v loop %d", states, p.nTerLoop)
	}
	reduced := Gia_ManReduceConst(p, 0)
	if Gia_ManRegNum(reduced) != 2 {
		t.Fatal("register interface should remain until cleanup")
	}
}

func TestSequentialSweepPreservesResetTraces(t *testing.T) {
	for variant := 0; variant < 8; variant++ {
		p := Gia_ManStart(32)
		a, r0, r1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		Gia_ManHashAlloc(p)
		out := Gia_ManHashXor(p, r0, r1)
		Gia_ManAppendCo(p, out)
		next0 := a
		if variant&1 != 0 {
			next0 = r0 ^ 1
		}
		next1 := a
		if variant&2 != 0 {
			next1 = 0
		}
		if variant&4 != 0 {
			next1 = Gia_ManHashAnd(p, a, r0)
		}
		Gia_ManAppendCo(p, next0)
		Gia_ManAppendCo(p, next1)
		Gia_ManHashStop(p)
		Gia_ManSetRegNum(p, 2)
		reduced := Gia_ManSeqStructSweep(p, 1, 1, 0)
		for sequence := 0; sequence < 32; sequence++ {
			state0 := make([]bool, Gia_ManRegNum(p))
			state1 := make([]bool, Gia_ManRegNum(reduced))
			for frame := 0; frame < 5; frame++ {
				pi := sequence&(1<<frame) != 0
				original, e := p.Evaluate(append([]bool{pi}, state0...))
				if e != nil {
					t.Fatal(e)
				}
				transformed, e := reduced.Evaluate(append([]bool{pi}, state1...))
				if e != nil || original[0] != transformed[0] {
					t.Fatalf("variant %d seq %d frame %d: %v %v %v", variant, sequence, frame, original, transformed, e)
				}
				copy(state0, original[1:])
				copy(state1, transformed[1:])
			}
		}
	}
}
