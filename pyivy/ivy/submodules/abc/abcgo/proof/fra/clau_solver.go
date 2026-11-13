package fra

// SAT instances and inductive-clause checks from fraClau.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

type clauManager struct {
	main, test, bmc                               *bsat.Solver
	mainCS, testCS, testNS, bmcNS                 *vec.Vec_Int_t
	mainToTest, testToMain, testToNext, testToBMC []int32
	helperStart, helperNext                       int32
}

func clauStart(graph *aig.Graph) *clauManager {
	if graph == nil || graph.RegisterCount() <= 0 || graph.OutputCount()-int(graph.RegisterCount()) != 1 {
		panic("clause induction requires one property and registers")
	}
	p := &clauManager{}
	main := aig.Aig_ManFrames(graph, 2, 0, 1, 0, 0, nil)
	aig.Aig_ManFlipFirstPo(main)
	mc := cnf.Cnf_DeriveSimple(main, 0)
	p.main = cnf.Cnf_DataWriteIntoSolver(mc, 1, 0)
	test := aig.Aig_ManFrames(graph, 1, 0, 0, 1, 0, nil)
	tc := cnf.Cnf_DeriveSimple(test, graph.RegisterCount())
	p.test = cnf.Cnf_DataWriteIntoSolver(tc, 1, 0)
	bmc := aig.Aig_ManFrames(graph, 1, 1, 0, 1, 0, nil)
	bc := cnf.Cnf_DeriveSimple(bmc, graph.RegisterCount())
	p.bmc = cnf.Cnf_DataWriteIntoSolver(bc, 1, 0)
	if p.main == nil || p.test == nil || p.bmc == nil {
		return nil
	}
	p.helperStart = p.test.NumVars()
	p.helperNext = p.helperStart
	p.mainCS = clauInputVars(main, mc, 2*(graph.InputCount()-int(graph.RegisterCount())))
	p.testCS = clauLatchVars(test, tc, true)
	p.testNS = clauLatchVars(test, tc, false)
	p.bmcNS = clauOutputVars(bmc, bc)
	p.mainToTest = clauMapping(p.mainCS, p.testCS, mc.Variables)
	p.testToMain = clauMapping(p.testCS, p.mainCS, tc.Variables)
	p.testToNext = clauMapping(p.testCS, p.testNS, tc.Variables)
	p.testToBMC = clauMapping(p.testCS, p.bmcNS, tc.Variables)
	return p
}

// modelCube records each variable's actual satisfying literal.
func modelCube(s *bsat.Solver, vars *vec.Vec_Int_t) *vec.Vec_Int_t {
	result := vec.Vec_IntAlloc(vec.Vec_IntSize(vars))
	for _, v := range vec.Vec_IntArray(vars) {
		value, ok := s.Value(v)
		if !ok {
			panic("missing SAT model variable")
		}
		lit := bsat.Positive(v)
		if !value {
			lit = bsat.Negative(v)
		}
		vec.Vec_IntPush(result, lit)
	}
	return result
}
func (p *clauManager) checkProperty() (int32, *vec.Vec_Int_t) {
	switch p.main.Solve() {
	case bsat.Unsatisfiable:
		return 1, nil
	case bsat.Satisfiable:
		return 0, modelCube(p.main, p.mainCS)
	default:
		return -1, nil
	}
}
func (p *clauManager) checkBMC(clause *vec.Vec_Int_t) int32 {
	switch p.bmc.Solve(vec.Vec_IntArray(clause)...) {
	case bsat.Unsatisfiable:
		return 1
	case bsat.Satisfiable:
		return 0
	default:
		return -1
	}
}

// checkClause tests relative induction. Each prior helper is disabled and the
// newly added clause is activated only for this query.
func (p *clauManager) checkClause(clause *vec.Vec_Int_t) (int32, *vec.Vec_Int_t) {
	helper := p.helperNext
	p.helperNext++
	learned := make([]int32, 0, vec.Vec_IntSize(clause)+1)
	for _, lit := range vec.Vec_IntArray(clause) {
		learned = append(learned, lit^1)
	}
	learned = append(learned, bsat.Negative(helper))
	if !p.test.AddClause(learned...) {
		panic("could not add conditional induction clause")
	}
	next := clauRemap(p.testToNext, clause, false)
	assumptions := append([]int32(nil), vec.Vec_IntArray(next)...)
	for i := p.helperStart; i < helper; i++ {
		assumptions = append(assumptions, bsat.Negative(i))
	}
	assumptions = append(assumptions, bsat.Positive(helper))
	switch p.test.Solve(assumptions...) {
	case bsat.Unsatisfiable:
		return 1, nil
	case bsat.Satisfiable:
		return 0, modelCube(p.test, p.testCS)
	default:
		return -1, nil
	}
}
