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
	var network ceccore.Network = s.graph
	root := a
	if b != nil {
		virtual := &allOutputConjunction{base: s.graph, ands: [][2]int32{{a, *b}}, root: s.graph.nObjs << 1}
		network = virtual
		root = virtual.root
	}
	options := ceccore.DefaultSATOptions()
	options.ConflictLimit = int64(s.limit)
	options.PolarFlip = false
	options.TreatRegistersAsInputs = true
	manager, err := ceccore.NewManager(network, options)
	if err != nil {
		return -1, nil, err
	}
	result, err := manager.CheckLiteral(root)
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
