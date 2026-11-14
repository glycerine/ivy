package pdr

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func tsimVec(xs ...int32) *vec.Vec_Int_t { return vec.Vec_IntAllocArrayCopy(xs, len(xs)) }
func TestPdrTsimTruthTables(t *testing.T) {
	vals := []int32{PDR_ZER, PDR_ONE, PDR_UND}
	not := []int32{PDR_ONE, PDR_ZER, PDR_UND}
	and := [][]int32{{PDR_ZER, PDR_ZER, PDR_ZER}, {PDR_ZER, PDR_ONE, PDR_UND}, {PDR_ZER, PDR_UND, PDR_UND}}
	for i, a := range vals {
		if got := Pdr_ManSimInfoNot(a); got != not[i] {
			t.Fatalf("not %d = %d", a, got)
		}
		for j, b := range vals {
			if got := Pdr_ManSimInfoAnd(a, b); got != and[i][j] {
				t.Fatalf("%d and %d = %d", a, b, got)
			}
		}
	}
}
func TestPdrTsimPropagationAndUndo(t *testing.T) {
	g := aig.NewGraph()
	pi := g.NewInput()
	lo := g.NewInput()
	gate := g.AND(pi, lo)
	po := g.AddOutput(gate)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	cos := tsimVec(int32(po.NodeID()))
	cis, nodes := tsimVec(), tsimVec()
	Pdr_ManCollectCone(g, cos, cis, nodes)
	values := tsimVec()
	for _, id := range vec.Vec_IntArray(cis) {
		if id == int32(pi.NodeID()) {
			vec.Vec_IntPush(values, 0)
		} else {
			vec.Vec_IntPush(values, 1)
		}
	}
	undo, vis := tsimVec(), tsimVec()
	if Pdr_ManSimDataInit(g, cis, values, nodes, cos, tsimVec(0), nil) != 1 {
		t.Fatal("initial simulation failed")
	}
	if Pdr_ManExtendOne(g, lo, undo, vis) != 1 {
		t.Fatal("controlling zero should make latch removable")
	}
	if Pdr_ManSimInfoGet(g, po) != PDR_ZER {
		t.Fatal("output changed")
	}
	Pdr_ManExtendUndo(g, undo)
	for i, id := range vec.Vec_IntArray(cis) {
		if id == int32(pi.NodeID()) {
			vec.Vec_IntWriteEntry(values, i, 1)
		}
	}
	if Pdr_ManSimDataInit(g, cis, values, nodes, cos, tsimVec(1), nil) != 1 {
		t.Fatal("initial one simulation failed")
	}
	vec.Vec_IntClear(undo)
	if Pdr_ManExtendOne(g, lo, undo, vis) != 0 {
		t.Fatal("latch must be retained")
	}
	Pdr_ManExtendUndo(g, undo)
	if Pdr_ManSimInfoGet(g, lo) != PDR_ONE || Pdr_ManSimInfoGet(g, po) != PDR_ONE {
		t.Fatal("undo did not restore values")
	}
}

// Exercise generalization against actual frame SAT models and both CNF modes.
func TestPdrTsimSATModelGeneralization(t *testing.T) {
	for _, mono := range []int32{0, 1} {
		for _, priority := range []int32{0, 1} {
			t.Run(fmt.Sprintf("mono=%d/priority=%d", mono, priority), func(t *testing.T) {
				g := aig.NewGraph()
				input := g.NewInput()
				required := g.NewInput()
				removable := g.NewInput()
				property := g.AddOutput(g.AND(required, g.OR(input, removable)))
				g.AddOutput(required)
				g.AddOutput(removable)
				aig.Aig_ManSetRegNum(g, 2)
				p := Pdr_ManStart(g, &Pdr_Par_t{FMonoCnf: mono, FFlopPrio: priority}, nil)
				defer Pdr_ManStop(p)
				Pdr_ManCreateSolver(p, 0)
				solver := Pdr_ManCreateSolver(p, 1)
				pdrPrepareModel(p, 1, nil)
				outvar := Pdr_ObjSatVar(p, 1, 3, property)
				ivar := Pdr_ObjSatVar(p, 1, 3, input)
				rvar := Pdr_ObjSatVar(p, 1, 3, required)
				xvar := Pdr_ObjSatVar(p, 1, 3, removable)
				if got := solver.Solve(bsat.Positive(outvar), bsat.Positive(ivar), bsat.Positive(rvar), bsat.Negative(xvar)); got != bsat.Satisfiable {
					t.Fatalf("SAT solver returned %v", got)
				}
				cube := Pdr_ManTernarySim(p, 1, nil)
				if cube.nLits != 1 || cube.Lits[0] != 0 {
					t.Fatalf("flop literals %v, want required latch 0 only", cube.Lits[:cube.nLits])
				}
				if cube.nTotal != 2 || cube.Lits[1] != 0 {
					t.Fatalf("input literals %v, want positive input 0", cube.Lits[cube.nLits:cube.nTotal])
				}
			})
		}
	}
}
