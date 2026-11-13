package pdr

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestIPdrReduceUnsatCore(t *testing.T) {
	for _, mono := range []int32{0, 1} {
		t.Run(fmt.Sprintf("mono=%d", mono), func(t *testing.T) {
			g := aig.NewGraph()
			necessary := g.NewInput()
			irrelevant := g.NewInput()
			g.AddOutput(necessary)
			g.AddOutput(necessary)
			g.AddOutput(irrelevant)
			aig.Aig_ManSetRegNum(g, 2)
			p := Pdr_ManStart(g, &Pdr_Par_t{fMonoCnf: mono}, nil)
			defer Pdr_ManStop(p)
			Pdr_ManCreateSolver(p, 0)
			Pdr_ManCreateSolver(p, 1)
			first := Pdr_SetCreate(tsimVec(0), tsimVec())
			extra := Pdr_SetCreate(tsimVec(2), tsimVec())
			clauses := vec.Vec_PtrAlloc(2)
			vec.Vec_PtrPush(clauses, first)
			vec.Vec_PtrPush(clauses, extra)
			if got := IPdr_ManCheckCubeReduce(p, clauses, nil, 0); got != 1 {
				t.Fatalf("result %d", got)
			}
			if vec.Vec_PtrSize(clauses) != 1 || vec.Vec_PtrEntry(clauses, 0) != first || first.nRefs != 1 || extra.nRefs != 0 {
				t.Fatalf("wrong reduced core: %v, refs %d/%d", vec.Vec_PtrSize(clauses), first.nRefs, extra.nRefs)
			}
		})
	}
}
func TestIPdrReduceSatisfiable(t *testing.T) {
	g := aig.NewGraph()
	latch := g.NewInput()
	g.AddOutput(latch)
	g.AddOutput(latch)
	aig.Aig_ManSetRegNum(g, 1)
	p := Pdr_ManStart(g, &Pdr_Par_t{}, nil)
	defer Pdr_ManStop(p)
	Pdr_ManCreateSolver(p, 0)
	Pdr_ManCreateSolver(p, 1)
	clause := Pdr_SetCreate(tsimVec(1), tsimVec())
	clauses := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(clauses, clause)
	if got := IPdr_ManCheckCubeReduce(p, clauses, nil, 0); got != 0 || vec.Vec_PtrSize(clauses) != 1 || clause.nRefs != 1 {
		t.Fatalf("SAT result=%d, clauses=%d refs=%d", got, vec.Vec_PtrSize(clauses), clause.nRefs)
	}
}

func TestAbcNtkDarIPdrTransfersCounterexample(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	input := abc.Abc_NtkCreatePi(n)
	latch := abc.Abc_NtkAddLatch(n, nil, abc.ABC_INIT_ZERO)
	abc.Abc_ObjAddFanin(abc.Abc_ObjFanin0(abc.Abc_ObjFanin0(latch)), input)
	output := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(output, input)
	abc.Abc_NtkOrderCisCos(n)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	if got := Abc_NtkDarIPdr(n, &pars); got != 0 {
		t.Fatalf("result %d, want counterexample", got)
	}
	if n.SequenceModel == nil || n.SequenceModel.IFrame != 0 {
		t.Fatalf("counterexample not transferred: %+v", n.SequenceModel)
	}
}
