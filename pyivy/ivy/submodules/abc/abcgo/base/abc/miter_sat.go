package abc

import (
	"fmt"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"os"
	"strings"
)

// satNodeVar stores the SAT variable assigned to an ABC object in CopyData.
func satNodeVar(o *Object) int32 {
	v, ok := o.CopyData.(int32)
	if !ok {
		panic("SAT variable missing from ABC object")
	}
	return v
}

func satEdgeLit(s *bsat.Solver2, edge Edge) bsat.Literal {
	v := satNodeVar(edge.Object)
	if edge.Complement {
		return bsat.Negative(v)
	}
	return bsat.Positive(v)
}

func satNeg(l bsat.Literal) bsat.Literal { return l ^ 1 }

func Abc_NtkGetCiSatVarNums(n *Abc_Ntk_t) *vec.Vec_Int_t {
	v := vec.Vec_IntAlloc(int(Abc_NtkCiNum(n)))
	for _, ci := range n.Inputs() {
		vec.Vec_IntPush(v, satNodeVar(ci))
	}
	return v
}

// Abc_NtkMiterSat solves the combinational miter. Return values follow ABC:
// -1 timeout, 0 satisfiable, 1 unsatisfiable.
func Abc_NtkMiterSat(n *Abc_Ntk_t, conflictLimit, inspectLimit int64, verbose int32, numConflicts, numInspects *int64) int32 {
	if numConflicts != nil {
		*numConflicts = 0
	}
	if numInspects != nil {
		*numInspects = 0
	}
	if Abc_NtkLatchNum(n) != 0 {
		panic("combinational SAT requires a network without latches")
	}
	solver := Abc_NtkMiterSatCreate(n, 0)
	if solver == nil {
		return 1
	}
	s := solver.(*bsat.Solver2)
	if !s.Simplify() {
		st := s.Statistics()
		if numConflicts != nil {
			*numConflicts = st.Conflicts
		}
		if numInspects != nil {
			*numInspects = st.Inspections
		}
		return 1
	}
	status := s.SolveWithSourceLimits(nil, int64(conflictLimit), int64(inspectLimit), 0, 0)
	st := s.Statistics()
	if numConflicts != nil {
		*numConflicts = st.Conflicts
	}
	if numInspects != nil {
		*numInspects = st.Inspections
	}
	if verbose != 0 {
		fmt.Fprintf(os.Stdout, "SAT variables=%d clauses=%d conflicts=%d inspections=%d\n", st.Variables, st.Clauses, st.Conflicts, st.Inspections)
	}
	switch status {
	case bsat.Unknown:
		return -1
	case bsat.Satisfiable:
		ids := Abc_NtkGetCiSatVarNums(n)
		n.Model = make([]int32, vec.Vec_IntSize(ids))
		model := s.Model()
		for i, v := range vec.Vec_IntArray(ids) {
			if int(v) < len(model) && model[v] {
				n.Model[i] = 1
			}
		}
		vec.Vec_IntFree(ids)
		return 0
	default:
		return 1
	}
}

func Abc_NtkMiterSatCreate(n *Abc_Ntk_t, allPrimes int32) any {
	if Abc_NtkIsBddLogic(n) != 0 {
		return Abc_NtkMiterSatCreateLogic(n, allPrimes)
	}
	if Abc_NtkIsStrash(n) == 0 {
		return nil
	}
	s := bsat.NewSolver2()
	if !Abc_NtkMiterSatCreateInt(s, n) {
		return nil
	}
	return s
}

func Abc_NtkMiterSatCreateInt(s *bsat.Solver2, n *Abc_Ntk_t) bool {
	// Keep object variable numbering deterministic: constant, CIs, then DFS nodes.
	constant := Abc_AigConst1(n).Object
	all := n.Objects()
	defer func() {
		for _, o := range all {
			o.MarkA = false
		}
	}()
	for _, o := range all {
		o.CopyData = nil
		o.MarkA = false
	}
	constVar := s.NewVariable()
	constant.CopyData = constVar
	if !s.AddClause(bsat.Positive(constVar)) {
		return false
	}
	for _, ci := range n.Inputs() {
		ci.CopyData = s.NewVariable()
	}
	roots := make([]*Object, 0, len(n.Outputs()))
	for _, po := range n.Outputs() {
		f := po.Fanins()
		if len(f) != 0 {
			roots = append(roots, f[0].Object)
		}
	}
	nodes := Abc_NtkDfsNodes(n, roots, int32(len(roots)))
	for _, node := range nodes {
		if node.Kind == ObjectNode {
			node.CopyData = s.NewVariable()
		}
	}
	for _, node := range nodes {
		if node.Kind != ObjectNode {
			continue
		}
		f := node.Fanins()
		if len(f) != 2 {
			continue
		}
		y := satEdgeLit(s, Edge{Object: node})
		a, b := satEdgeLit(s, f[0]), satEdgeLit(s, f[1])
		// Equivalence y = a AND b, represented by its three prime clauses.
		if !s.AddClause(satNeg(a), satNeg(b), y) || !s.AddClause(a, satNeg(y)) || !s.AddClause(b, satNeg(y)) {
			return false
		}
	}
	for _, po := range n.Outputs() {
		f := po.Fanins()
		if len(f) != 0 && !s.AddClause(satEdgeLit(s, f[0])) {
			return false
		}
	}
	return true
}

// The default Go build has the CUDD package available, but ABC's C encoder
// depends on allocator-owned SOP strings. This entry point is implemented
// below using the Go BDD-to-CNF bridge.
func Abc_NtkMiterSatCreateLogic(n *Abc_Ntk_t, allPrimes int32) any {
	s := bsat.NewSolver2()
	maxID := int32(0)
	for _, o := range n.Objects() {
		if int32(o.ID) > maxID {
			maxID = int32(o.ID)
		}
		o.CopyData = int32(o.ID)
	}
	for s.NumVars() <= maxID {
		s.NewVariable()
	}
	var cube []byte
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		root, ok := node.Data.(*core.DdNode)
		manager, managerOK := n.FunctionManager.(*core.DdManager)
		if !ok || !managerOK {
			panic("BDD SAT encoding requires BDD nodes and their manager")
		}
		if core.Cudd_Regular(root) == core.Cudd_ReadOne(manager) {
			lit := bsat.Positive(satNodeVar(node))
			if core.Cudd_IsComplement(root) {
				lit = satNeg(lit)
			}
			if !s.AddClause(lit) {
				return nil
			}
			continue
		}
		negative, positive := "", ""
		Abc_NodeBddToCnf(node, nil, &cube, allPrimes, &negative, &positive)
		if !satAddSOPPhase(s, node, negative, false) || !satAddSOPPhase(s, node, positive, true) {
			return nil
		}
	}
	for _, po := range n.Outputs() {
		f := po.Fanins()
		if len(f) == 0 || !s.AddClause(satEdgeLit(s, f[0])) {
			return nil
		}
	}
	return s
}

func satAddSOPPhase(s *bsat.Solver2, node *Object, sop string, positive bool) bool {
	rows := strings.Split(sop, "\n")
	for _, row := range rows {
		fields := strings.Fields(row)
		if len(fields) < 2 {
			continue
		}
		clause := make([]bsat.Literal, 0, len(fields))
		for i, c := range []byte(fields[0]) {
			if i >= len(node.Fanins()) {
				break
			}
			e := node.Fanins()[i]
			switch c {
			case '0':
				clause = append(clause, satEdgeLit(s, e))
			case '1':
				clause = append(clause, satNeg(satEdgeLit(s, e)))
			}
		}
		out := bsat.Positive(satNodeVar(node))
		if !positive {
			out = satNeg(out)
		}
		clause = append(clause, out)
		if !s.AddClause(clause...) {
			return false
		}
	}
	return true
}
