package fra

// One-hot candidate and constraint operations from ABC's fraHot.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Fra_OneHotCompute(p *Fra_Man_t, sim *Fra_Sml_t) *vec.Vec_Int_t {
	if sim.Graph != p.pManAig {
		panic("fra: wrong one-hot simulation graph")
	}
	inputs := sim.Graph.Inputs()
	first := len(inputs) - int(sim.Graph.RegisterCount())
	out := vec.Vec_IntAlloc(100)
	for i := first; i < len(inputs); i++ {
		a := inputs[i]
		if Fra_OneHotNodeIsConst(sim, a) != 0 {
			continue
		}
		for j := i + 1; j < len(inputs); j++ {
			b := inputs[j]
			if Fra_OneHotNodeIsConst(sim, b) != 0 || Fra_OneHotNodesAreEqual(sim, a, b) != 0 {
				continue
			}
			for _, signs := range [][2]int32{{1, 1}, {0, 1}, {1, 0}} {
				if Fra_OneHotNodesAreClause(sim, a, b, signs[0], signs[1]) != 0 {
					vec.Vec_IntPush(out, Fra_RegToLit(int32(i-first), signs[0]))
					vec.Vec_IntPush(out, Fra_RegToLit(int32(j-first), signs[1]))
					break
				}
			}
		}
	}
	return out
}
func Fra_OneHotAssume(p *Fra_Man_t, clauses *vec.Vec_Int_t) {
	if p.pPars.NFramesK != 1 {
		panic("fra: one-hot constraints require one frame")
	}
	inputs := p.pManFraig.Inputs()
	first := len(inputs) - int(p.pManFraig.RegisterCount())
	for i := 0; i < vec.Vec_IntSize(clauses); i += 2 {
		a, b := vec.Vec_IntEntry(clauses, i), vec.Vec_IntEntry(clauses, i+1)
		if a == 0 && b == 0 {
			continue
		}
		left, right := inputs[first+int(Fra_LitReg(a))], inputs[first+int(Fra_LitReg(b))]
		if !p.pSat.AddClause(fraLit(Fra_ObjSatNum(p, left), a < 0), fraLit(Fra_ObjSatNum(p, right), b < 0)) {
			fmt.Println("Fra_OneHotAssume(): Adding clause makes SAT solver unsat.")
			p.pSat = nil
			return
		}
	}
}
func Fra_OneHotCheck(p *Fra_Man_t, clauses *vec.Vec_Int_t) {
	outputs := p.pManFraig.Outputs()
	first := len(outputs) - int(p.pManFraig.RegisterCount())
	for i := 0; i < vec.Vec_IntSize(clauses); i += 2 {
		a, b := vec.Vec_IntEntry(clauses, i), vec.Vec_IntEntry(clauses, i+1)
		if a == 0 && b == 0 {
			continue
		}
		left, right := outputs[first+int(Fra_LitReg(a))], outputs[first+int(Fra_LitReg(b))]
		result := Fra_NodesAreClause(p, left, right, Fra_LitSign(a), Fra_LitSign(b))
		if result != 1 {
			p.pCla.FRefinement = 1
			if result == 0 {
				Fra_SmlResimulate(p)
			}
			if vec.Vec_IntEntry(clauses, i) != 0 {
				fmt.Println("Fra_OneHotCheck(): Clause is not refined!")
				panic("fra: one-hot clause was not refined")
			}
		}
	}
}
func Fra_OneHotCreateExdc(p *Fra_Man_t, clauses *vec.Vec_Int_t) *aig.Graph {
	graph := aig.NewGraph()
	for range p.pManAig.Inputs() {
		graph.NewInput()
	}
	inputs := graph.Inputs()
	first := len(inputs) - int(p.pManAig.RegisterCount())
	for i := 0; i < vec.Vec_IntSize(clauses); i += 2 {
		a, b := vec.Vec_IntEntry(clauses, i), vec.Vec_IntEntry(clauses, i+1)
		if a == 0 && b == 0 {
			continue
		}
		left, right := inputs[first+int(Fra_LitReg(a))], inputs[first+int(Fra_LitReg(b))]
		if a < 0 {
			left = left.Not()
		}
		if b < 0 {
			right = right.Not()
		}
		graph.AddOutput(graph.OR(left, right))
	}
	aig.Aig_ManCleanup(graph)
	return graph
}
func Fra_OneHotAddKnownConstraint(p *Fra_Man_t, groups *vec.Vec_Ptr_t) {
	inputs := p.pManFraig.Inputs()
	for k := 0; k < vec.Vec_PtrSize(groups); k++ {
		group, ok := vec.Vec_PtrEntry(groups, k).(*vec.Vec_Int_t)
		if !ok {
			panic("fra: invalid one-hot group")
		}
		values := vec.Vec_IntArray(group)
		for i, a := range values {
			for _, b := range values[i+1:] {
				if !p.pSat.AddClause(fraLit(Fra_ObjSatNum(p, inputs[a]), true), fraLit(Fra_ObjSatNum(p, inputs[b]), true)) {
					fmt.Println("Fra_OneHotAddKnownConstraint(): Adding clause makes SAT solver unsat.")
					p.pSat = nil
					return
				}
			}
		}
	}
}
