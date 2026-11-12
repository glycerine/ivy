package gia

import (
	"math/rand"
	"testing"
)

func TestCircuitSAT2ExhaustiveBothPropagationModes(t *testing.T) {
	rng := rand.New(rand.NewSource(863))
	for run := 0; run < 20; run++ {
		graph := Gia_ManStart(100)
		Gia_ManHashStart(graph)
		literals := []int32{0, 1}
		for i := 0; i < 5; i++ {
			literals = append(literals, Gia_ManAppendCi(graph))
		}
		for i := 0; i < 30; i++ {
			a, b := literals[rng.Intn(len(literals))]^int32(rng.Intn(2)), literals[rng.Intn(len(literals))]^int32(rng.Intn(2))
			lit := Gia_ManHashAnd(graph, a, b)
			if i%3 == 0 {
				lit = Gia_ManHashXor(graph, a, b)
			}
			literals = append(literals, lit)
		}
		Gia_ManHashStop(graph)
		p := Cbs2_ManAlloc(graph)
		for mode := int32(0); mode < 2; mode++ {
			p.Pars.fUseFanout = mode
			for query := 0; query < 60; query++ {
				a, b := literals[rng.Intn(len(literals))]^int32(rng.Intn(2)), literals[rng.Intn(len(literals))]^int32(rng.Intn(2))
				pair := query%2 != 0
				sat := false
				for mask := 0; mask < 32; mask++ {
					inputs := make([]bool, 5)
					for i := range inputs {
						inputs[i] = mask&(1<<uint(i)) != 0
					}
					values, _ := giaEvaluateNodes(graph, inputs)
					if values[a>>1] != (a&1 != 0) && (!pair || b == 0 || values[b>>1] != (b&1 != 0)) {
						sat = true
						break
					}
				}
				var status int32
				if pair {
					status = Cbs2_ManSolve2(p, a, b)
				} else {
					status = Cbs2_ManSolve(p, a)
				}
				if status < 0 || (status == 0) != sat {
					t.Fatal("CBS2/exhaustive mismatch", run, mode, query, a, b, pair, status, sat)
				}
				if status == 0 && !pair {
					inputs := make([]bool, 5)
					for _, lit := range Vec_IntArray(Cbs2_ReadModel(p)) {
						inputs[lit>>1] = lit&1 == 0
					}
					values, _ := giaEvaluateNodes(graph, inputs)
					if values[a>>1] == (a&1 != 0) {
						t.Fatal("invalid CBS2 model", run, mode, a, Vec_IntArray(p.vModel))
					}
				}
				for id, value := range p.vAssign {
					if value < 2 || p.vMark[id] != 0 || p.vLevReason[3*id] != -1 {
						t.Fatal("CBS2 leaked assignment", id)
					}
				}
				Cbs2_ManCheckFanouts(p)
			}
		}
	}
}
