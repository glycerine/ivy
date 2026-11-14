package gia

import (
	"fmt"

	ceccore "github.com/glycerine/abcgo/proof/cec/core"
	"github.com/glycerine/abcgo/sat/bsat"
)

// giarfSolver checks conjunctions of one or two literals using either the
// native TAS solver or the source CEC CNF/BSAT engine, using ABC's status codes.
// Its graph is the ordinary AND-inverter graph from Gia_GenerateReducedLevel;
// real gates are expanded before this SAT boundary.
type giarfSolver struct {
	graph    *Gia_Man_t
	limit    int32
	mini     bool
	tas      *Tas_Man_t
	cec      *ceccore.SourceCNF
	cecSince int32
}

func newGiarfSolver(graph *Gia_Man_t, limit int32, mini bool) *giarfSolver {
	s := &giarfSolver{graph: graph, limit: limit, mini: mini}
	if !mini {
		s.tas = Tas_ManAlloc(graph, limit)
	}
	return s
}
func (s *giarfSolver) close() {
	if s.tas != nil {
		Tas_ManStop(s.tas)
	}
	s.cec = nil
}
func (s *giarfSolver) check(a int32, b *int32) (int32, *Vec_Int_t, error) {
	if a < 0 || a>>1 >= s.graph.nObjs || b != nil && (*b < 0 || *b>>1 >= s.graph.nObjs) {
		return -1, nil, fmt.Errorf("gia: invalid SAT assumption")
	}
	if !s.mini {
		var second *Gia_Obj_t
		if b != nil {
			second = Gia_ObjFromLit(s.graph, *b)
		}
		status := Tas_ManSolve(s.tas, Gia_ObjFromLit(s.graph, a), second)
		if status == 0 {
			return status, Vec_IntDup(Tas_ReadModel(s.tas)), nil
		}
		return status, nil, nil
	}
	if s.limit < 0 {
		return -1, nil, fmt.Errorf("gia: negative CEC conflict limit")
	}
	// These source CEC trivial cases do not allocate a SAT cone.
	if a == 0 || b != nil && (*b == 0 || a == *b^1) {
		return 1, nil, nil
	}
	if s.cec == nil {
		s.cec = ceccore.NewSourceCNF(s.graph.SourceCNFGraph(), nil)
	}
	s.cecSince++
	// Defaults from Cec_ManSatSetDefaultParams, with Giarf polarity flipping
	// disabled. Keep the source's strict variable/call recycling boundaries.
	if s.cec.Solver() == nil || s.cec.Solver().NumVars() > 2000 && s.cecSince > 200 {
		s.cec.Reset()
		s.cecSince = 0
	}
	s.cec.AddNode(a >> 1)
	assumptions := []int32{s.cec.Literal(a)}
	if b != nil {
		s.cec.AddNode(*b >> 1)
		assumptions = append(assumptions, s.cec.Literal(*b))
	}
	solver := s.cec.Solver()
	switch solver.SolveWithResourceLimits(assumptions, int64(s.limit), 0, 0, 0) {
	case bsat.Unsatisfiable:
		for i := range assumptions {
			assumptions[i] ^= 1
		}
		s.cec.AddClause(assumptions...)
		return 1, nil, nil
	case bsat.Unknown:
		return -1, nil, nil
	}
	inputs := make([]bool, Gia_ManCiNum(s.graph))
	for i, id := range Vec_IntArray(s.graph.vCis) {
		if variable := s.cec.IDs()[id]; variable != 0 {
			value, ok := solver.Value(variable)
			if !ok {
				return -1, nil, fmt.Errorf("gia: SAT model missing")
			}
			inputs[i] = value
		}
	}
	values, _, err := s.graph.evaluateObjects(inputs)
	if err != nil {
		return -1, nil, err
	}
	value := func(lit int32) bool { return values[lit>>1] != (lit&1 != 0) }
	if !value(a) || b != nil && !value(*b) {
		return -1, nil, fmt.Errorf("gia: source SAT witness failed independent simulation")
	}
	cex := Vec_IntAlloc(len(inputs))
	for i, value := range inputs {
		literal := int32(i) << 1
		if !value {
			literal |= 1
		}
		Vec_IntPush(cex, literal)
	}
	return 0, cex, nil
}
