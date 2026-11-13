package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"testing"
)

func TestClauseRefinementDoesNotStoreUncheckedCandidates(t *testing.T) {
	graph := aig.NewGraph()
	input := graph.NewInput()
	graph.NewInput()
	graph.AddOutput(input)
	aig.Aig_ManSetRegNum(graph, 1)
	p := Fra_ClausAlloc(graph, 1, 0, 10, 4, 0, 0, 1, 0, 0, 0, 0)
	p.pCnf = cnf.Cnf_DeriveSimple(graph, int32(graph.OutputCount()))
	variable := p.pCnf.VarNums[input.NodeID()]
	vec.Vec_IntPush(p.vLits, bsat.Positive(variable))
	vec.Vec_IntPush(p.vClauses, 1)
	vec.Vec_IntPush(p.vCosts, 1)
	vec.Vec_IntPush(p.vLits, bsat.Negative(variable))
	vec.Vec_IntPush(p.vClauses, 2)
	vec.Vec_IntPush(p.vCosts, 1)
	p.nClauses = 2
	if clausRefineAndStore(p) {
		t.Fatal("contradictory assumptions unexpectedly refined")
	}
	if vec.Vec_IntSize(p.vClausesProven) != 0 || vec.Vec_IntSize(p.vLitsProven) != 0 {
		t.Fatal("unchecked candidates were stored as proven")
	}
}
