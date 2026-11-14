// Package aigpartsat owns the partitioned SAT driver from aigPartSat.c above
// the AIG, mapped-CNF and BSAT packages to avoid cyclic imports.
package aigpartsat

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// NodeVariables maps original objects to phase-normalized shared SAT variables.
// Variable zero remains unused, as in the source CNF writer.
type NodeVariables struct {
	IDs    []int32
	Phases []bool
}

func NewNodeVariables(nodes int) *NodeVariables {
	return &NodeVariables{make([]int32, nodes), make([]bool, nodes)}
}
func sharedLiteral(v *NodeVariables, id int) bsat.Literal {
	lit := bsat.Positive(v.IDs[id])
	if v.Phases[id] {
		lit ^= 1
	}
	return lit
}
func newPartVariable(s *bsat.Solver) int32 {
	if s.NumVars() == 0 {
		s.SetNumVars(1)
	}
	return s.NewVariable()
}

// Aig_ManAddNewCnfToSolver follows the original mapped-CNF, polarity, lift,
// terminal-connector and output/restriction clause sequence. Explicit phases
// also allow differently rewritten partitions to share the same logical node.
func Aig_ManAddNewCnfToSolver(s *bsat.Solver, part *aig.Graph, v *NodeVariables, pio []int32, outputs []aig.Edge, _ int32) int32 {
	inputs, cos := part.Inputs(), part.Outputs()
	if len(pio) != len(inputs)+len(cos) || len(v.IDs) == 0 || len(v.IDs) != len(v.Phases) {
		panic("aigpartsat: invalid terminal mapping")
	}
	data := cnf.Cnf_Derive(part, int32(len(cos)))
	if data == nil {
		panic("aigpartsat: mapped CNF derivation failed")
	}
	defer cnf.Cnf_DataFree(data)
	cnf.Cnf_DataTranformPolarity(data, 1)
	offset := s.NumVars()
	cnf.Cnf_DataLift(data, offset)
	s.SetNumVars(offset + int32(data.Variables))
	for _, clause := range data.Clauses {
		lits := make([]bsat.Literal, len(clause))
		for i, l := range clause {
			lits[i] = bsat.Literal(l)
		}
		if !s.AddClause(lits...) {
			return 1
		}
	}
	terminals := append(inputs, cos...)
	for i, e := range terminals {
		id := int(pio[i])
		if id < 0 || id >= len(v.IDs) {
			panic("aigpartsat: terminal mapping outside original graph")
		}
		variable := data.VarNums[e.NodeID()]
		if variable <= 0 {
			panic("aigpartsat: terminal has no mapped CNF variable")
		}
		phase := part.Node(e).Phase
		if v.IDs[id] == 0 {
			v.IDs[id], v.Phases[id] = variable, phase
			continue
		}
		old, new := bsat.Positive(v.IDs[id]), bsat.Positive(variable)
		if v.Phases[id] != phase {
			new ^= 1
		}
		if !s.AddClause(old, new^1) || !s.AddClause(old^1, new) {
			return 1
		}
	}
	if v.IDs[0] == 0 {
		v.IDs[0], v.Phases[0] = data.VarNums[0], part.Node(part.Const1()).Phase
	}
	// Typed graphs may use the original constant as a partition boundary CI.
	// Constrain that binding too; its own local CNF constant is already true.
	for _, id := range pio {
		if id == 0 && !s.AddClause(sharedLiteral(v, 0)) {
			return 1
		}
	}
	for _, driver := range outputs {
		id := driver.NodeID()
		if id >= len(v.IDs) || v.IDs[id] == 0 {
			panic("aigpartsat: output has no shared variable")
		}
		lit := sharedLiteral(v, id)
		if driver.Complemented() {
			lit ^= 1
		}
		if !s.AddClause(lit) {
			return 1
		}
	}
	for i, e := range inputs {
		n := part.Node(e)
		lit := sharedLiteral(v, int(pio[i]))
		if n.MarkA && !s.AddClause(lit^1) {
			return 1
		}
		if n.MarkB && !s.AddClause(lit) {
			return 1
		}
		aig.Aig_ObjClearMarkA(part, e)
		aig.Aig_ObjClearMarkB(part, e)
	}
	return 0
}

func Aig_ManDeriveCounterExample(g *aig.Graph, v *NodeVariables, s *bsat.Solver) {
	model := make([]int32, g.InputCount())
	for i, e := range g.Inputs() {
		id := e.NodeID()
		if v.IDs[id] == 0 {
			continue
		}
		value, ok := s.Value(v.IDs[id])
		if !ok {
			panic("aigpartsat: SAT model unavailable")
		}
		if value != v.Phases[id] {
			model[i] = 1
		}
	}
	g.Data = model
}
