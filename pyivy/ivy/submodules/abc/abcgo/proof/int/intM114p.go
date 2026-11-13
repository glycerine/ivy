package int

// The optional MiniSat 1.14p proof interface in intM114p.c. ABC's ordinary
// build excludes it; abcgo does not yet provide the MiniSat 1.14p backend.
import (
	"errors"
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/cnf"
)

// ErrMiniSat114pUnavailable means that no original MiniSat 1.14p proof
// backend has been supplied. No alternative solver is selected implicitly.
var ErrMiniSat114pUnavailable = errors.New("MiniSat 1.14p proof backend is unavailable")

// M114pChain represents one resolution chain; Clauses has one more entry
// than Pivots, and indexes roots followed by earlier chains.
type M114pChain struct{ Clauses, Pivots []int32 }

// M114pBackend is the proof-producing interface required by the optional
// ABC path. Literals use ABC's var*2+sign encoding. Solve returns 0 for
// UNSAT, 1 for SAT and -1 for unknown, as in m114p.h.
type M114pBackend interface {
	SetVarNum(int32)
	AddClause([]int32) bool
	Solve() int32
	ConflictNum() int32
	Roots() [][]int32
	Chains() []M114pChain
	Close()
}

// Inter_ManDeriveSatSolverM114p assembles the original A and B partitions.
// A backend must be supplied explicitly; nil reports the unavailable port.
func Inter_ManDeriveSatSolverM114p(s M114pBackend, inter *aig.Graph, interCNF *cnf.Cnf_Dat_t, trans *aig.Graph, transCNF *cnf.Cnf_Dat_t, frames *aig.Graph, framesCNF *cnf.Cnf_Dat_t) ([]int32, []int32, error) {
	if s == nil {
		return nil, nil, ErrMiniSat114pUnavailable
	}
	if inter == nil || trans == nil || frames == nil || interCNF == nil || transCNF == nil || framesCNF == nil {
		return nil, nil, errors.New("invalid MiniSat interpolation input")
	}
	regs := int(trans.RegisterCount())
	if regs == 0 || inter.RegisterCount() != 0 || frames.RegisterCount() != 0 || inter.OutputCount() != 1 || frames.OutputCount() != 1 || inter.InputCount() != regs || frames.InputCount() < regs {
		return nil, nil, errors.New("incompatible MiniSat interpolation graphs")
	}
	fv, tv := int32(framesCNF.Variables), int32(transCNF.Variables)
	cnf.Cnf_DataLift(transCNF, fv)
	cnf.Cnf_DataLift(interCNF, fv+tv)
	defer cnf.Cnf_DataLift(transCNF, -fv)
	defer cnf.Cnf_DataLift(interCNF, -fv-tv)
	vars := make([]int32, framesCNF.Variables+transCNF.Variables+interCNF.Variables)
	for i := range vars {
		vars[i] = -1
		if i < framesCNF.Variables {
			vars[i] = -2
		}
	}
	s.SetVarNum(int32(len(vars)))
	var roots []int32
	add := func(clause []int32, part int32) bool { roots = append(roots, part); return s.AddClause(clause) }
	for _, clause := range interCNF.Clauses {
		if !add(clause, 0) {
			return nil, nil, errors.New("inconsistent A interpolant clauses")
		}
	}
	ti, to := trans.Inputs(), trans.Outputs()
	lo, li := ti[len(ti)-regs:], to[len(to)-regs:]
	equal := func(a, b int32) bool { return add([]int32{a * 2, b*2 + 1}, 0) && add([]int32{a*2 + 1, b * 2}, 0) }
	for i, input := range inter.Inputs() {
		if !equal(interCNF.VarNums[input.NodeID()], transCNF.VarNums[lo[i].NodeID()]) {
			return nil, nil, errors.New("inconsistent A connector")
		}
	}
	for _, clause := range transCNF.Clauses {
		if !add(clause, 0) {
			return nil, nil, errors.New("inconsistent A transition clauses")
		}
	}
	for i, input := range frames.Inputs()[:regs] {
		v := framesCNF.VarNums[input.NodeID()]
		vars[v] = int32(i)
		if !equal(v, transCNF.VarNums[li[i].NodeID()]) {
			return nil, nil, errors.New("inconsistent partition connector")
		}
	}
	for _, clause := range framesCNF.Clauses {
		if !add(clause, 1) {
			break
		}
	}
	return roots, vars, nil
}

// Inter_ManResolveM114p resolves a clause against a pivot, as in ABC.
func Inter_ManResolveM114p(resolvent, clause []int32, pivot int32) ([]int32, error) {
	index := -1
	for i, lit := range resolvent {
		if lit>>1 == pivot {
			index = i
			break
		}
	}
	if index < 0 {
		return nil, errors.New("resolution pivot absent")
	}
	old := resolvent[index]
	resolvent[index] = resolvent[len(resolvent)-1]
	resolvent = resolvent[:len(resolvent)-1]
	found := false
	for _, lit := range clause {
		if lit>>1 == pivot {
			if lit != old^1 {
				return nil, errors.New("resolution pivot has wrong polarity")
			}
			found = true
			continue
		}
		exists := false
		for _, x := range resolvent {
			if x == lit {
				exists = true
				break
			}
		}
		if !exists {
			resolvent = append(resolvent, lit)
		}
	}
	if !found {
		return nil, errors.New("resolution pivot absent from second clause")
	}
	return resolvent, nil
}

// InterpolateM114p reconstructs either original proof interpolation algorithm.
func InterpolateM114p(s M114pBackend, rootParts, varClasses []int32, pudlak bool) (*aig.Graph, error) {
	if s == nil {
		return nil, ErrMiniSat114pUnavailable
	}
	roots, chains := s.Roots(), s.Chains()
	if len(roots) == 0 || len(roots) != len(rootParts) {
		return nil, errors.New("invalid proof roots")
	}
	g := aig.NewGraph()
	max := int32(-1)
	for _, class := range varClasses {
		if class > max {
			max = class
		}
	}
	inputs := make([]aig.Edge, max+1)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	outputs := make([]aig.Edge, 0, len(roots)+len(chains))
	clauses := make([][]int32, 0, len(roots)+len(chains))
	classOf := func(lit int32) (int32, error) {
		v := lit >> 1
		if lit < 0 || int(v) >= len(varClasses) {
			return 0, errors.New("proof literal out of bounds")
		}
		return varClasses[v], nil
	}
	for i, clause := range roots {
		edge := g.Const0()
		if rootParts[i] == 1 {
			edge = g.Const1()
		} else if rootParts[i] != 0 {
			return nil, errors.New("invalid root partition")
		} else if !pudlak {
			for _, lit := range clause {
				c, err := classOf(lit)
				if err != nil {
					return nil, err
				}
				if c >= 0 {
					e := inputs[c]
					if lit&1 != 0 {
						e = e.Not()
					}
					edge = g.OR(edge, e)
				}
			}
		}
		outputs = append(outputs, edge)
		clauses = append(clauses, append([]int32(nil), clause...))
	}
	for _, chain := range chains {
		if len(chain.Clauses) != len(chain.Pivots)+1 || len(chain.Clauses) == 0 {
			return nil, errors.New("invalid proof chain")
		}
		first := int(chain.Clauses[0])
		if first < 0 || first >= len(outputs) {
			return nil, errors.New("invalid proof antecedent")
		}
		edge := outputs[first]
		resolvent := append([]int32(nil), clauses[first]...)
		for j, pivot := range chain.Pivots {
			next := int(chain.Clauses[j+1])
			if next < 0 || next >= len(outputs) || pivot < 0 || int(pivot) >= len(varClasses) {
				return nil, errors.New("invalid proof pivot or antecedent")
			}
			other := outputs[next]
			class := varClasses[pivot]
			if pudlak {
				var err error
				resolvent, err = Inter_ManResolveM114p(resolvent, clauses[next], pivot)
				if err != nil {
					return nil, err
				}
			}
			if class == -1 {
				edge = g.OR(edge, other)
			} else if !pudlak || class == -2 {
				edge = g.AND(edge, other)
			} else if class >= 0 {
				var lit int32 = -1
				for _, x := range clauses[next] {
					if x>>1 == pivot {
						lit = x
						break
					}
				}
				if lit < 0 {
					return nil, errors.New("proof pivot absent")
				}
				control := inputs[class]
				if lit&1 != 0 {
					control = control.Not()
				}
				edge = g.Mux(control, edge, other)
			} else {
				return nil, fmt.Errorf("invalid variable class %d", class)
			}
		}
		outputs = append(outputs, edge)
		clauses = append(clauses, resolvent)
	}
	if pudlak && len(clauses[len(clauses)-1]) != 0 {
		return nil, errors.New("proof does not end with empty clause")
	}
	g.AddOutput(outputs[len(outputs)-1])
	return g, nil
}

// Inter_ManPerformOneStepM114p runs the optional solver supplied by the caller.
func Inter_ManPerformOneStepM114p(p *Inter_Man_t, s M114pBackend, pudlak bool) (int32, error) {
	if s == nil {
		return -1, ErrMiniSat114pUnavailable
	}
	defer s.Close()
	roots, vars, err := Inter_ManDeriveSatSolverM114p(s, p.pInter, p.pCnfInter, p.pAigTrans, p.pCnfAig, p.pFrames, p.pCnfFrames)
	if err != nil {
		return -1, err
	}
	start := time.Now()
	status := s.Solve()
	p.timeSat += int64(time.Since(start))
	p.nConfCur = s.ConflictNum()
	switch status {
	case 0:
		start = time.Now()
		p.pInterNew, err = InterpolateM114p(s, roots, vars, pudlak)
		p.timeInt += int64(time.Since(start))
		if err != nil {
			return -1, err
		}
		return 1, nil
	case 1:
		return 0, nil
	default:
		return -1, nil
	}
}
