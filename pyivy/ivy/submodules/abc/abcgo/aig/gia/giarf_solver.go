package gia

import (
	"fmt"

	ceccore "github.com/glycerine/abcgo/proof/cec/core"
)

// giarfSolver checks conjunctions of one or two literals using either the
// native TAS solver or the CEC SAT manager, returning ABC's status convention.
type giarfSolver struct {
	graph *Gia_Man_t
	limit int32
	mini  bool
	tas   *Tas_Man_t
	cec   *ceccore.Manager
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
	if s.cec == nil {
		options := ceccore.DefaultSATOptions()
		options.ConflictLimit = int64(s.limit)
		options.PolarFlip = false
		options.TreatRegistersAsInputs = true
		var err error
		s.cec, err = ceccore.NewManager(s.graph, options)
		if err != nil {
			return -1, nil, err
		}
	}
	var result ceccore.Result
	var err error
	if b == nil {
		result, err = s.cec.CheckLiteral(a)
	} else {
		result, err = s.cec.CheckConjunction(a, *b)
	}
	if err != nil {
		return -1, nil, err
	}
	if result.Status != ceccore.Different {
		return int32(result.Status), nil, nil
	}
	if result.Counterexample == nil {
		return -1, nil, fmt.Errorf("gia: SAT model missing")
	}
	cex := Vec_IntAlloc(len(result.Counterexample.Inputs))
	for i, value := range result.Counterexample.Inputs {
		literal := int32(i) << 1
		if !value {
			literal |= 1
		}
		Vec_IntPush(cex, literal)
	}
	return 0, cex, nil
}
