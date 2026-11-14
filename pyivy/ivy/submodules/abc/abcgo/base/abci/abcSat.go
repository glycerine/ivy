package abci

import (
	"fmt"
	"os"
	"strings"

	abc "github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

var nMuxes int32

// satNodeVar stores the SAT variable assigned to an ABC object in CopyData.
func satNodeVar(o *abc.Object) int32 {
	v, ok := o.CopyData.(int32)
	if !ok {
		panic("SAT variable missing from ABC object")
	}
	return v
}

func satEdgeLit(s *bsat.Solver2, edge abc.Edge) bsat.Literal {
	v := satNodeVar(edge.Object)
	if edge.Complement {
		return bsat.Negative(v)
	}
	return bsat.Positive(v)
}

func satNeg(l bsat.Literal) bsat.Literal { return l ^ 1 }

func Abc_NtkGetCiSatVarNums(n *Abc_Ntk_t) *Vec_Int_t {
	v := vec.Vec_IntAlloc(int(Abc_NtkCiNum(n)))
	for _, ci := range n.Inputs() {
		vec.Vec_IntPush(v, satNodeVar(ci))
	}
	return v
}

// Abc_NtkMiterSat solves the combinational miter. Return values follow ABC:
// -1 timeout, 0 satisfiable, 1 unsatisfiable.
func Abc_NtkMiterSat(n *Abc_Ntk_t, conflictLimit, inspectLimit ABC_INT64_T, verbose int32, numConflicts, numInspects *ABC_INT64_T) int32 {
	if numConflicts != nil {
		*numConflicts = 0
	}
	if numInspects != nil {
		*numInspects = 0
	}
	assert(Abc_NtkLatchNum(n) == 0, "Abc_NtkLatchNum(n) == 0")
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
	if abc.Abc_NtkIsBddLogic(n) != 0 {
		return Abc_NtkMiterSatCreateLogic(n, allPrimes)
	}
	if abc.Abc_NtkIsStrash(n) == 0 {
		return nil
	}
	nMuxes = 0
	s := bsat.NewSolver2()
	if !Abc_NtkMiterSatCreateInt(s, n) {
		return nil
	}
	return s
}

func Abc_NtkMiterSatCreateInt(s *bsat.Solver2, n *Abc_Ntk_t) bool {
	// Keep object variable numbering deterministic: constant, CIs, then DFS nodes.
	constant := abc.Abc_AigConst1(n).Object
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
	roots := make([]*abc.Object, 0, len(n.Outputs()))
	for _, po := range n.Outputs() {
		f := po.Fanins()
		if len(f) != 0 {
			roots = append(roots, f[0].Object)
		}
	}
	nodes := abc.Abc_NtkDfsNodes(n, roots, int32(len(roots)))
	for _, node := range nodes {
		if node.Kind == abc.ObjectNode {
			node.CopyData = s.NewVariable()
		}
	}
	for _, node := range nodes {
		if node.Kind != abc.ObjectNode {
			continue
		}
		f := node.Fanins()
		if len(f) != 2 {
			continue
		}
		y := satEdgeLit(s, abc.Edge{Object: node})
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
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
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
		abc.Abc_NodeBddToCnf(node, nil, &cube, allPrimes, &negative, &positive)
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

func satAddSOPPhase(s *bsat.Solver2, node *abc.Object, sop string, positive bool) bool {
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

func Abc_NtkNodeFactor(pObj *Abc_Obj_t, nLevelMax int32) int32 {
	assert(int32(pObj.Level) <= nLevelMax, "object level is within requested maximum")
	return int32(100000000.0 * (1 + 0.01*float64(pObj.Level)))
}

// Abc_NtkWriteSorterCnf creates the sorter through ABC's registered command
// path and writes a CNF for outputs Q and Q+1.
func Abc_NtkWriteSorterCnf(fileName string, nVars, nQueens int32) {
	if nQueens <= 0 || nQueens >= nVars {
		fmt.Printf("The number of queens (Q = %d) should belong to the interval: 0 < Q < %d.\n", nQueens, nVars)
		return
	}
	frame := abcmain.Abc_FrameGetGlobalFrame()
	command := fmt.Sprintf("gen -s -N %d sorter%d.blif", nVars, nVars)
	if abcmain.Cmd_CommandExecute(frame, command) != 0 {
		fmt.Fprintf(os.Stdout, "Cannot execute command %q.\n", command)
		return
	}
	command = fmt.Sprintf("read sorter%d.blif; strash", nVars)
	if abcmain.Cmd_CommandExecute(frame, command) != 0 {
		fmt.Fprintf(os.Stdout, "Cannot execute command %q.\n", command)
		return
	}
	n := Abc_FrameReadNtk(frame)
	if n == nil || int32(len(n.Outputs())) <= nVars-nQueens {
		return
	}
	// The full writer is kept behind the source network's two selected PO roots.
	// Build a focused CNF via the same node equivalence clauses as the SAT encoder.
	if err := writeSorterCNF(fileName, n, nVars, nQueens); err != nil {
		fmt.Fprintf(os.Stdout, "Cannot write CNF %q: %v\n", fileName, err)
	}
}

func writeSorterCNF(fileName string, n *Abc_Ntk_t, nVars, nQueens int32) error {
	po := n.Outputs()
	first, second := po[nVars-nQueens-1], po[nVars-nQueens]
	roots := []*abc.Object{first.Fanins()[0].Object, second.Fanins()[0].Object}
	nodes := abc.Abc_NtkDfsNodes(n, roots, 2)
	counter := int32(0)
	for _, pi := range n.Inputs() {
		pi.CopyData = counter
		counter++
	}
	for _, node := range nodes {
		if node.Kind == abc.ObjectNode {
			node.CopyData = counter
			counter++
		}
	}
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = fmt.Fprintf(f, "c CNF for %d-bit sorter with %d bits set to 1 generated by ABC.\np cnf %d %d\n", nVars, nQueens, counter, 3*len(nodes)+2); err != nil {
		return err
	}
	litText := func(e abc.Edge) string {
		v := satNodeVar(e.Object) + 1
		if !e.Complement {
			return fmt.Sprintf("%d", v)
		}
		return fmt.Sprintf("-%d", v)
	}
	for _, node := range nodes {
		if node.Kind != abc.ObjectNode {
			continue
		}
		fins := node.Fanins()
		if len(fins) != 2 {
			continue
		}
		y := satNodeVar(node) + 1
		if _, err = fmt.Fprintf(f, "%d %s %s 0\n-%d %s 0\n-%d %s 0\n", y, litText(fins[0].Not()), litText(fins[1].Not()), y, litText(fins[0]), y, litText(fins[1])); err != nil {
			return err
		}
	}
	if _, err = fmt.Fprintf(f, "%s 0\n%s 0\n", litText(first.Fanins()[0].Not()), litText(second.Fanins()[0])); err != nil {
		return err
	}
	return f.Close()
}
