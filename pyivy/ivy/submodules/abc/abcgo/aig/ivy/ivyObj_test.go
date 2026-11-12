package ivy

import "testing"

func TestReplaceIdentityAndBufferedComplement(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	old := Ivy_And(p, a, b)
	Ivy_ObjCreatePo(p, old)
	Ivy_ManStartFanout(p)
	id, negative := old.Id, Ivy_Not(old)
	replacement := Ivy_And(p, a, c)
	Ivy_ObjReplace(p, old, replacement, 1, 0, 1)
	if old.Id != id || Ivy_Not(old) != negative || !replacement.recycled || Ivy_ManCheck(p) == 0 {
		t.Fatal("identity transfer failed")
	}
	for bits := 0; bits < 8; bits++ {
		if ivyBalanceEvaluate(p, bits)[0] != (bits&1 != 0 && bits&4 != 0) {
			t.Fatal("replacement truth changed")
		}
	}
	Ivy_ObjReplace(p, old, Ivy_Not(c), 1, 0, 1)
	if Ivy_ObjIsBuf(old) == 0 {
		t.Fatal("complemented replacement must retain buffer identity")
	}
	Ivy_ManPropagateBuffers(p, 1)
	if Ivy_ManBufNum(p) != 0 || Ivy_ManCheck(p) == 0 {
		t.Fatal("buffer propagation failed")
	}
	for bits := 0; bits < 8; bits++ {
		if ivyBalanceEvaluate(p, bits)[0] != (bits&4 == 0) {
			t.Fatal("buffer polarity lost")
		}
	}
}
func TestReplaceRejectsDependencyCycleBeforeMutation(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	old := Ivy_And(p, a, b)
	dependent := Ivy_And(p, old, c)
	Ivy_ObjCreatePo(p, dependent)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("cycle replacement accepted")
			}
		}()
		Ivy_ObjReplace(p, old, dependent, 1, 0, 0)
	}()
	if Ivy_ManCheck(p) == 0 {
		t.Fatal("rejected replacement mutated graph")
	}
}
func TestMakeSeqBoundaryAndAccounting(t *testing.T) {
	p := Ivy_ManStart()
	input, state := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	next := Ivy_Exor(p, input, state)
	Ivy_ObjCreatePo(p, state)
	Ivy_ObjCreatePo(p, next)
	Ivy_ManMakeSeq(p, 1, []int32{int32(IVY_INIT_1)})
	if Ivy_ManPiNum(p) != 1 || Ivy_ManPoNum(p) != 1 || Ivy_ManLatchNum(p) != 1 || Ivy_ManCheck(p) == 0 {
		t.Fatal("invalid sequential boundary/accounting")
	}
	var latch *Ivy_Obj_t
	for _, n := range p.vObjs {
		if n != nil && Ivy_ObjIsLatch(n) != 0 {
			latch = n
		}
	}
	if latch == nil {
		t.Fatal("initial state lost")
	}
	for assignment := 0; assignment < 4; assignment++ {
		values := map[*Ivy_Obj_t]bool{p.pConst1: true, p.vPis[0]: assignment&1 != 0, latch: assignment&2 != 0}
		var eval func(*Ivy_Obj_t) bool
		eval = func(edge *Ivy_Obj_t) bool {
			n := Ivy_Regular(edge)
			value, known := values[n]
			if !known {
				if n.Type == IVY_AND {
					value = eval(n.pFanin0) && eval(n.pFanin1)
				} else if n.Type == IVY_EXOR {
					value = eval(n.pFanin0) != eval(n.pFanin1)
				} else {
					value = eval(n.pFanin0)
				}
				values[n] = value
			}
			if Ivy_IsComplement(edge) != 0 {
				return !value
			}
			return value
		}
		visible := eval(p.vPos[0])
		if assignment == 0 && latch.Init == IVY_INIT_0 && !visible {
			t.Fatal("initial visible state lost")
		}
		if assignment == 2 && latch.Init == IVY_INIT_1 && !visible {
			t.Fatal("initial visible state lost")
		}
		nextValue := eval(latch.pFanin0)
		values = map[*Ivy_Obj_t]bool{p.pConst1: true, p.vPis[0]: assignment&1 != 0, latch: nextValue}
		if eval(p.vPos[0]) != ((assignment&1 != 0) != visible) {
			t.Fatalf("sequential transition differs for %d", assignment)
		}
	}
}

func TestBoundedDependencyRevisitsSharedConeWithGreaterDepth(t *testing.T) {
	p := Ivy_ManStart()
	inputs := make([]*Ivy_Obj_t, 8)
	for i := range inputs {
		inputs[i] = Ivy_ObjCreatePi(p)
	}
	target := Ivy_And(p, inputs[0], inputs[1])
	middle := Ivy_And(p, target, inputs[2])
	shared := Ivy_And(p, middle, inputs[3])
	long := Ivy_And(p, Ivy_And(p, shared, inputs[4]), inputs[5])
	short := Ivy_And(p, shared, inputs[6])
	root := Ivy_And(p, long, short)
	if Ivy_ObjIsInTfi_rec(root, target, 4) != 1 {
		t.Fatal("bounded search discarded shorter route to shared cone")
	}
}
