package gia

import (
	"math/rand"
	"testing"
)

func TestChronologicalCircuitSATExhaustiveAndModels(t *testing.T) {
	rng := rand.New(rand.NewSource(177))
	for run := 0; run < 20; run++ {
		g := Gia_ManStart(100)
		Gia_ManHashStart(g)
		lits := []int32{0, 1}
		for i := 0; i < 5; i++ {
			lits = append(lits, Gia_ManAppendCi(g))
		}
		for i := 0; i < 30; i++ {
			a, b := lits[rng.Intn(len(lits))]^int32(rng.Intn(2)), lits[rng.Intn(len(lits))]^int32(rng.Intn(2))
			if i%3 == 0 {
				lits = append(lits, Gia_ManHashXor(g, a, b))
			} else {
				lits = append(lits, Gia_ManHashAnd(g, a, b))
			}
		}
		Gia_ManHashStop(g)
		p := Cbs0_ManAlloc(g)
		for heuristic := 0; heuristic < 3; heuristic++ {
			p.Pars.fUseHighest = 0
			p.Pars.fUseLowest = 0
			p.Pars.fUseMaxFF = 0
			switch heuristic {
			case 0:
				p.Pars.fUseHighest = 1
			case 1:
				p.Pars.fUseLowest = 1
			case 2:
				p.Pars.fUseMaxFF = 1
			}
			for _, literal := range lits {
				for complement := int32(0); complement < 2; complement++ {
					lit := literal ^ complement
					satisfiable := false
					for mask := 0; mask < 32; mask++ {
						inputs := make([]bool, 5)
						for i := range inputs {
							inputs[i] = mask>>i&1 != 0
						}
						values, e := giaEvaluateNodes(g, inputs)
						if e != nil {
							t.Fatal(e)
						}
						if values[lit>>1] != (lit&1 != 0) {
							satisfiable = true
							break
						}
					}
					status := Cbs0_ManSolve(p, Gia_ObjFromLit(g, lit))
					if status < 0 || (status == 0) != satisfiable {
						t.Fatalf("run%d heuristic%d literal%d: %d vs SAT%v", run, heuristic, lit, status, satisfiable)
					}
					if status == 0 {
						inputs := make([]bool, 5)
						for _, value := range Vec_IntArray(Cbs0_ReadModel(p)) {
							inputs[value>>1] = value&1 == 0
						}
						values, e := giaEvaluateNodes(g, inputs)
						if e != nil || values[lit>>1] == (lit&1 != 0) {
							t.Fatal("invalid chronological witness", lit, Vec_IntArray(p.vModel), e)
						}
					}
					for _, o := range g.pObjs {
						if o.fMark0 != 0 || o.fMark1 != 0 {
							t.Fatal("assignment leaked")
						}
					}
				}
			}
		}
		Cbs0_ManStop(p)
	}
}
func TestChronologicalLimitsAndQueueRestoration(t *testing.T) {
	g := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, b)
	p := Cbs0_ManAlloc(g)
	p.Pars.nJustLimit = 0
	if result := Cbs0_ManSolve(p, Gia_ObjFromLit(g, root^1)); result != -1 {
		t.Fatalf("expected frontier timeout, got %d", result)
	}
	if Vec_IntSize(p.vModel) != 0 {
		t.Fatal("timeout exposes stale model")
	}
	p.Pars.nJustLimit = 100
	if Cbs0_ManSolve(p, Gia_ObjFromLit(g, root^1)) != 0 {
		t.Fatal("solver not reusable")
	}
	q := Cbs0_Que_t{}
	Cbs0_QuePush(&q, Gia_ObjFromLit(g, a))
	Cbs0_QuePush(&q, Gia_ObjFromLit(g, b))
	var head, tail int32
	Cbs0_QueStore(&q, &head, &tail)
	q.pData[q.iHead] = nil
	Cbs0_QueRestore(&q, head, tail)
	if q.pData[0] != Gia_ObjFromLit(g, a) || q.pData[1] != Gia_ObjFromLit(g, b) {
		t.Fatal("suspended frontier overwritten")
	}
}
func TestChronologicalDegenerateFaninsAndMiter(t *testing.T) {
	g := Gia_ManStart(10)
	a := Gia_ManAppendCi(g)
	Gia_ManHashStart(g)
	opposite := Gia_ManHashAnd(g, a, a^1)
	same := Gia_ManHashAnd(g, a, a)
	Gia_ManHashStop(g)
	p := Cbs0_ManAlloc(g)
	for _, tc := range []struct{ lit, status int32 }{{0, 1}, {1, 0}, {opposite, 1}, {opposite ^ 1, 0}, {same, 0}, {same ^ 1, 0}} {
		if got := Cbs0_ManSolve(p, Gia_ObjFromLit(g, tc.lit)); got != tc.status {
			t.Fatal(tc, got)
		}
	}
	Cbs0_ManStop(p)
	for _, lit := range []int32{0, 1, opposite, a} {
		Gia_ManAppendCo(g, lit)
	}
	var statuses *Vec_Str_t
	store := Cbs_ManSolveMiter(g, 1000, &statuses, 0)
	if Vec_StrSize(statuses) != 4 || Vec_StrEntry(statuses, 0) != 1 || Vec_StrEntry(statuses, 1) != 0 || Vec_StrEntry(statuses, 2) != 1 || Vec_StrEntry(statuses, 3) != 0 {
		t.Fatal("miter statuses")
	}
	values := Vec_IntArray(store)
	if len(values) != 5 || values[0] != 1 || values[1] != 0 || values[2] != 3 || values[3] != 1 {
		t.Fatal("source store encoding", values)
	}
}
