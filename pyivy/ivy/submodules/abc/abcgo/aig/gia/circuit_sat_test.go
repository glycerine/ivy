package gia

import (
	"math/rand"
	"testing"
)

func TestCircuitSATMatchesExhaustiveModelsAndRestoresScratch(t *testing.T) {
	rng := rand.New(rand.NewSource(313))
	for run := 0; run < 30; run++ {
		graph := Gia_ManStart(100)
		Gia_ManHashStart(graph)
		literals := []int32{0, 1}
		for i := 0; i < 5; i++ {
			literals = append(literals, Gia_ManAppendCi(graph))
		}
		for i := 0; i < 25; i++ {
			a, b := literals[rng.Intn(len(literals))]^int32(rng.Intn(2)), literals[rng.Intn(len(literals))]^int32(rng.Intn(2))
			literals = append(literals, Gia_ManHashAnd(graph, a, b))
			if i%3 == 0 {
				literals = append(literals, Gia_ManHashXor(graph, a, b))
			}
		}
		Gia_ManHashStop(graph)
		solver := Cbs_ManAlloc(graph)
		for heuristic := 0; heuristic < 3; heuristic++ {
			solver.Pars.fUseHighest, solver.Pars.fUseLowest, solver.Pars.fUseMaxFF = 0, 0, 0
			switch heuristic {
			case 0:
				solver.Pars.fUseHighest = 1
			case 1:
				solver.Pars.fUseLowest = 1
			case 2:
				solver.Pars.fUseMaxFF = 1
			}
			for query := 0; query < 30; query++ {
				a := literals[rng.Intn(len(literals))] ^ int32(rng.Intn(2))
				b := literals[rng.Intn(len(literals))] ^ int32(rng.Intn(2))
				pair := query%2 != 0
				satisfiable := false
				for mask := 0; mask < 32; mask++ {
					inputs := make([]bool, 5)
					for i := range inputs {
						inputs[i] = mask&(1<<uint(i)) != 0
					}
					values, _ := giaEvaluateNodes(graph, inputs)
					if values[a>>1] != (a&1 != 0) && (!pair || values[b>>1] != (b&1 != 0)) {
						satisfiable = true
						break
					}
				}
				var status int32
				if pair {
					status = Cbs_ManSolve2(solver, Gia_ObjFromLit(graph, a), Gia_ObjFromLit(graph, b))
				} else {
					status = Cbs_ManSolve(solver, Gia_ObjFromLit(graph, a))
				}
				if status < 0 || ((status == 0) != satisfiable) {
					t.Fatal("CBS/exhaustive mismatch", run, heuristic, a, b, pair, status, satisfiable)
				}
				if status == 0 && !pair {
					inputs := make([]bool, 5)
					for _, lit := range Vec_IntArray(Cbs_ReadModel(solver)) {
						inputs[lit>>1] = lit&1 == 0
					}
					values, _ := giaEvaluateNodes(graph, inputs)
					if values[a>>1] == (a&1 != 0) {
						t.Fatal("invalid CBS model", run, a, Vec_IntArray(Cbs_ReadModel(solver)))
					}
				}
				for _, obj := range graph.pObjs {
					if obj.fMark0 != 0 || obj.fMark1 != 0 || obj.Value != ^uint32(0) {
						t.Fatal("CBS leaked scratch", obj.id)
					}
				}
			}
		}
	}
}
func TestCircuitSATQueueSnapshotAndLimits(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	and := Gia_ManAppendAnd(p, a, b)
	solver := Cbs_ManAlloc(p)
	solver.Pars.nJustLimit = 0
	if status := Cbs_ManSolve(solver, Gia_ObjFromLit(p, and^1)); status != -1 {
		t.Fatal("ignored justification limit", status)
	}
	solver.Pars.nJustLimit = 100
	if status := Cbs_ManSolve2(solver, Gia_ObjFromLit(p, a), Gia_ObjFromLit(p, a^1)); status != 1 {
		t.Fatal("opposite root assumptions", status)
	}
	var queue Cbs_Que_t
	Cbs_QuePush(&queue, p.pObjs[a>>1])
	Cbs_QuePush(&queue, p.pObjs[b>>1])
	var head, tail int32
	Cbs_QueStore(&queue, &head, &tail)
	Cbs_QueRestore(&queue, head, tail)
	if queue.iHead != 0 || queue.iTail != 2 || queue.pData[0] != p.pObjs[a>>1] || queue.pData[1] != p.pObjs[b>>1] {
		t.Fatal("lost saved queue")
	}
}
