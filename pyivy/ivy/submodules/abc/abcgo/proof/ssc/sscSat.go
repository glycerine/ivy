package ssc

// SAT frontier and class queries ported from ABC's sscSat.c.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"github.com/glycerine/abcgo/sat/satoko"
)

func sscStatus(status bsat.Status) int {
	switch status {
	case bsat.Satisfiable:
		return satoko.SATOKO_SAT
	case bsat.Unsatisfiable:
		return satoko.SATOKO_UNSAT
	default:
		return satoko.SATOKO_UNDEC
	}
}
func Ssc_ManStartSolver(p *Ssc_Man_t) {
	graph := gia.Gia_ManToAigSimple(p.Aig)
	formula := cnf.Cnf_Derive(graph, 0)
	p.sat = bsat.NewSolver()
	p.inconsistent = false
	p.satVars[0] = make([]int32, p.Aig.ObjectCount())
	p.satNext = int32(formula.Variables) + 1
	for p.sat.NumVars() < p.satNext {
		p.sat.NewVariable()
	}
	// SSC reserves variable zero as the unmapped-object sentinel.
	for _, row := range formula.Clauses {
		lits := make([]bsat.Literal, len(row))
		for i, lit := range row {
			lits[i] = lit + 2
		}
		if !p.sat.AddClause(lits...) {
			p.inconsistent = true
			return
		}
	}
	// The Go CNF encoder folds constants rather than assigning a variable;
	// SSC needs the explicit constant-one SAT variable used by the C solver.
	p.satVars[0][0] = p.satNext
	p.satNext++
	for p.sat.NumVars() < p.satNext {
		p.sat.NewVariable()
	}
	if !p.sat.AddClause(bsat.Positive(p.satVars[0][0])) {
		p.inconsistent = true
		return
	}
	inputs := p.Aig.InputLiterals()
	p.satInputs = make([]int32, len(inputs))
	for i, lit := range inputs {
		copy := gia.Gia_ObjValue(gia.Gia_ManObj(p.Aig, lit>>1))
		variable := formula.VarNums[copy>>1] + 1
		if variable <= 0 {
			panic("ssc: missing CNF input variable")
		}
		p.satVars[0][lit>>1] = variable
		p.satInputs[i] = variable
	}
}
func (p *Ssc_Man_t) satVar(id int32) int32 {
	if id < 0 {
		panic("ssc: negative node ID")
	}
	for int(id) >= len(p.satVars[0]) {
		p.satVars[0] = append(p.satVars[0], 0)
	}
	return p.satVars[0][id]
}
func (p *Ssc_Man_t) satLit(lit int32) bsat.Literal {
	variable := p.satVar(lit >> 1)
	if variable == 0 {
		panic("ssc: unmapped SAT literal")
	}
	sign := lit & 1
	if lit>>1 == 0 {
		sign ^= 1
	}
	return 2*variable + sign
}
func (p *Ssc_Man_t) addFrontier(id int32, front *[]int32) {
	if id <= 0 {
		panic("ssc: invalid frontier node")
	}
	if p.satVar(id) != 0 {
		return
	}
	p.satVars[0][id] = p.satNext
	p.satNext++
	for p.sat.NumVars() < p.satNext+100 {
		p.sat.NewVariable()
	}
	_, _, _, kind := p.Aig.Gate(id)
	switch kind {
	case 1:
	case 2:
		*front = append(*front, id)
	default:
		panic("ssc: frontier requires AND-only GIA")
	}
}
func (p *Ssc_Man_t) addStructuralClause(lits ...bsat.Literal) {
	if !p.sat.AddClause(lits...) {
		panic("ssc: structural CNF is inconsistent")
	}
}
func (p *Ssc_Man_t) addMuxClauses(id int32) {
	obj := gia.Gia_ManObj(p.Aig, id)
	var yes, no int32
	control := gia.Gia_ObjRecognizeMuxLits(p.Aig, obj, &yes, &no)
	f, i, t, e := p.satLit(id<<1), p.satLit(control), p.satLit(yes), p.satLit(no)
	p.addStructuralClause(i^1, t^1, f)
	p.addStructuralClause(i^1, t, f^1)
	p.addStructuralClause(i, e^1, f)
	p.addStructuralClause(i, e, f^1)
	if t != e {
		p.addStructuralClause(t, e, f^1)
		p.addStructuralClause(t^1, e^1, f)
	}
}
func (p *Ssc_Man_t) collectSuper(id int32) []int32 {
	var leaves []int32
	seen := make(map[int32]bool)
	var visit func(int32)
	visit = func(lit int32) {
		node := lit >> 1
		_, _, _, kind := p.Aig.Gate(node)
		if lit&1 != 0 || kind == 1 || (kind == 2 && gia.Gia_ObjIsMuxType(gia.Gia_ManObj(p.Aig, node)) != 0) {
			if !seen[lit] {
				seen[lit] = true
				leaves = append(leaves, lit)
			}
			return
		}
		if kind != 2 {
			panic("ssc: supergate requires AND-only GIA")
		}
		a, b, _, _ := p.Aig.Gate(node)
		visit(a)
		visit(b)
	}
	a, b, _, kind := p.Aig.Gate(id)
	if kind != 2 {
		panic("ssc: invalid supergate root")
	}
	visit(a)
	visit(b)
	return leaves
}
func (p *Ssc_Man_t) addSuperClauses(id int32, leaves []int32) {
	f := p.satLit(id << 1)
	long := make([]bsat.Literal, 0, len(leaves)+1)
	for _, leaf := range leaves {
		v := p.satLit(leaf)
		p.addStructuralClause(v, f^1)
		long = append(long, v^1)
	}
	p.addStructuralClause(append(long, f)...)
}
func (p *Ssc_Man_t) cnfNode(id int32) {
	if id <= 0 {
		panic("ssc: invalid CNF node")
	}
	if p.satVar(id) != 0 {
		return
	}
	var front []int32
	p.addFrontier(id, &front)
	for i := 0; i < len(front); i++ {
		node := front[i]
		obj := gia.Gia_ManObj(p.Aig, node)
		if gia.Gia_ObjIsMuxType(obj) != 0 {
			a, b, _, _ := p.Aig.Gate(node)
			a0, a1, _, _ := p.Aig.Gate(a >> 1)
			b0, b1, _, _ := p.Aig.Gate(b >> 1)
			seen := map[int32]bool{}
			for _, lit := range []int32{a0, b0, a1, b1} {
				id := lit >> 1
				if !seen[id] {
					seen[id] = true
					p.addFrontier(id, &front)
				}
			}
			p.addMuxClauses(node)
		} else {
			leaves := p.collectSuper(node)
			for _, lit := range leaves {
				p.addFrontier(lit>>1, &front)
			}
			p.addSuperClauses(node, leaves)
		}
	}
}
func (p *Ssc_Man_t) encode(side int, lit int32) int {
	if side != 0 {
		panic("ssc: SAT literals belong to the working FRAIG")
	}
	if lit>>1 != 0 {
		p.cnfNode(lit >> 1)
	}
	return int(p.satLit(lit))
}
func (p *Ssc_Man_t) pattern() []bool {
	result := make([]bool, len(p.satInputs))
	for i, v := range p.satInputs {
		value, ok := p.sat.Value(v)
		if ok {
			result[i] = value
		}
	}
	return result
}
func Ssc_ManCollectSatPattern(p *Ssc_Man_t) []bool { return p.pattern() }
func Ssc_ManFindPivotSat(p *Ssc_Man_t) ([]bool, int) {
	if p.inconsistent {
		return nil, satoko.SATOKO_UNSAT
	}
	p.sat.SetLimits(int64(p.Params.BTLimit), 0)
	status := sscStatus(p.sat.Solve())
	if status != satoko.SATOKO_SAT {
		return nil, status
	}
	return p.pattern(), status
}
func Ssc_ManCheckEquivalence(p *Ssc_Man_t, repr, node, complement int32) int {
	p.Stats.Calls++
	if repr == node {
		panic("ssc: identical equivalence nodes")
	}
	if repr > node {
		p.Stats.Undecided++
		return satoko.SATOKO_UNDEC
	}
	if repr != 0 {
		p.cnfNode(repr)
	}
	p.cnfNode(node)
	if !p.sat.Compress() {
		panic("ssc: SAT solver is inconsistent")
	}
	sign := complement
	if repr > 0 {
		sign ^= 1
	}
	literals := [2]bsat.Literal{2 * p.satVar(repr), 2*p.satVar(node) + sign}
	p.sat.SetLimits(int64(p.Params.BTLimit), 0)
	for phase := 0; phase < 2; phase++ {
		status := sscStatus(p.sat.Solve(literals[:]...))
		switch status {
		case satoko.SATOKO_SAT:
			p.Pattern = p.pattern()
			p.Stats.Counterexamples++
			return status
		case satoko.SATOKO_UNDEC:
			p.Stats.Undecided++
			return status
		case satoko.SATOKO_UNSAT:
			literals[0] ^= 1
			literals[1] ^= 1
			if !p.sat.AddClause(literals[:]...) {
				panic("ssc: learned implication contradicts solver")
			}
			if repr == 0 {
				p.Stats.Proofs++
				return status
			}
		}
	}
	p.Stats.Proofs++
	return satoko.SATOKO_UNSAT
}
