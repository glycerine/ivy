package glucose2

/**CFile****************************************************************

  FileName    [AbcGlucose.cpp]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver Glucose 3.0 by Gilles Audemard and Laurent Simon.]

  Synopsis    [Interface to Glucose.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 6, 2017.]

  Revision    [$Id: AbcGlucose.cpp,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"io"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

func boolInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func solverStatus(v LBool) int32 {
	if v == LTrue {
		return GLUCOSE_SAT
	}
	if v == LFalse {
		return GLUCOSE_UNSAT
	}
	return GLUCOSE_UNDEC
}
func checkedLits(ps []int32, n int32) []Lit {
	if n < 0 || int(n) > len(ps) {
		panic("invalid literal count")
	}
	out := make([]Lit, n)
	for i, p := range ps[:n] {
		if p < 0 {
			panic("negative SAT literal")
		}
		out[i] = Lit(p)
	}
	return out
}
func glucose2_solver_start() *SimpSolver { return NewABCSolver() }
func glucose2_solver_stop(s *SimpSolver) {
	if s != nil {
		s.Reset()
	}
}
func glucose2_solver_reset(s *SimpSolver) { s.Reset() }
func glucose2_solver_addclause(s *SimpSolver, ps []int32, n int32) int32 {
	lits := checkedLits(ps, n)
	for _, p := range lits {
		s.AddVar(p.Var())
	}
	return boolInt(s.AddClause(lits))
}
func glucose2_solver_setcallback(s *SimpSolver, manager any, callback func(any, int32, []int32) int32) {
	s.CnfManager = manager
	if callback == nil {
		s.CnfCallback = nil
		return
	}
	s.CnfCallback = func(m any, status int, model []int) int {
		var values []int32
		if model != nil {
			values = make([]int32, len(model))
			for i, v := range model {
				values[i] = int32(v)
			}
		}
		return int(callback(m, int32(status), values))
	}
}
func glucose2_solver_solve(s *SimpSolver, ps []int32, n int32) int32 {
	// The C wrapper passes the explicit `do_simp = 0` argument. Preprocessing
	// here would eliminate variables between incremental assumption queries.
	return solverStatus(s.SolveLimited(checkedLits(ps, n), false))
}
func glucose2_solver_addvar(s *SimpSolver) int32 { return int32(s.NewVar(true, true)) }
func glucose2_solver_read_cex(s *SimpSolver) []int32 {
	out := make([]int32, len(s.Model))
	for i, v := range s.Model {
		out[i] = boolInt(v == LTrue)
	}
	return out
}
func glucose2_solver_read_cex_varvalue(s *SimpSolver, v int32) int32 {
	return boolInt(s.Model[v] == LTrue)
}
func glucose2_solver_setstop(s *SimpSolver, stop *int32) {
	if stop == nil {
		s.Stop = nil
	} else {
		s.Stop = func() bool { return atomic.LoadInt32(stop) != 0 }
	}
}
func bmcg2_sat_solver_start() *SimpSolver  { return glucose2_solver_start() }
func bmcg2_sat_solver_stop(s *SimpSolver)  { glucose2_solver_stop(s) }
func bmcg2_sat_solver_reset(s *SimpSolver) { s.Reset() }
func bmcg2_sat_solver_addclause(s *SimpSolver, ps []int32, n int32) int32 {
	return glucose2_solver_addclause(s, ps, n)
}
func bmcg2_sat_solver_setcallback(s *SimpSolver, m any, callback func(any, int32, []int32) int32) {
	glucose2_solver_setcallback(s, m, callback)
}
func bmcg2_sat_solver_solve(s *SimpSolver, ps []int32, n int32) int32 {
	return glucose2_solver_solve(s, ps, n)
}
func bmcg2_sat_solver_final(s *SimpSolver, out *[]int32) int32 {
	lits := make([]int32, len(s.Conflict))
	for i, p := range s.Conflict {
		lits[i] = int32(p)
	}
	*out = lits
	return int32(len(lits))
}
func bmcg2_sat_solver_addvar(s *SimpSolver) int32 { return glucose2_solver_addvar(s) }
func bmcg2_sat_solver_set_nvars(s *SimpSolver, n int32) {
	if n > 0 {
		s.AddVar(Var(n - 1))
	}
}

func bmcg2_sat_solver_eliminate(s *SimpSolver, turnOff int32) int32 {
	return boolInt(s.Eliminate(turnOff != 0))
}
func bmcg2_sat_solver_var_is_elim(s *SimpSolver, v int32) int32 {
	if v < 0 || int(v) >= len(s.eliminated) {
		return 0
	}
	return boolInt(s.IsEliminated(Var(v)))
}
func bmcg2_sat_solver_var_set_frozen(s *SimpSolver, v, freeze int32) {
	if v < 0 || int(v) >= s.NVars() {
		panic("glucose2: frozen variable out of range")
	}
	s.SetFrozen(Var(v), freeze != 0)
}
func bmcg2_sat_solver_elim_varnum(s *SimpSolver) int32 { return int32(s.EliminatedVars) }
func bmcg2_sat_solver_read_cex(s *SimpSolver) []int32  { return glucose2_solver_read_cex(s) }
func bmcg2_sat_solver_read_cex_varvalue(s *SimpSolver, v int32) int32 {
	return glucose2_solver_read_cex_varvalue(s, v)
}
func bmcg2_sat_solver_set_stop(s *SimpSolver, stop *int32) { glucose2_solver_setstop(s, stop) }
func bmcg2_sat_solver_set_runtime_limit(s *SimpSolver, limit int64) int64 {
	old := s.runtimeLimit
	s.runtimeLimit = limit
	if limit == 0 {
		s.Deadline = time.Time{}
	} else {
		s.Deadline = time.Now().Add(time.Duration(limit - util.Abc_Clock()))
	}
	return old
}
func bmcg2_sat_solver_set_conflict_budget(s *SimpSolver, limit int32) {
	if limit > 0 {
		s.SetConfBudget(int64(limit))
	} else {
		s.BudgetOff()
	}
}
func bmcg2_sat_solver_varnum(s *SimpSolver) int32      { return int32(s.NVars()) }
func bmcg2_sat_solver_clausenum(s *SimpSolver) int32   { return int32(s.NClauses()) }
func bmcg2_sat_solver_learntnum(s *SimpSolver) int32   { return int32(s.NLearnts()) }
func bmcg2_sat_solver_conflictnum(s *SimpSolver) int32 { return int32(s.Conflicts) }
func bmcg2_sat_solver_minimize_assumptions(s *SimpSolver, ps []int32, n, pivot int32) int32 {
	checkedLits(ps, n)
	if pivot < 0 || pivot >= n {
		panic("invalid minimization pivot")
	}
	if n-pivot == 1 {
		return boolInt(bmcg2_sat_solver_solve(s, ps, pivot) != GLUCOSE_UNSAT)
	}
	left := (n - pivot) / 2
	right := n - pivot - left
	if bmcg2_sat_solver_solve(s, ps, pivot+left) == GLUCOSE_UNSAT {
		return bmcg2_sat_solver_minimize_assumptions(s, ps, pivot+left, pivot)
	}
	keptRight := int32(1)
	if right != 1 {
		keptRight = bmcg2_sat_solver_minimize_assumptions(s, ps, n, pivot+left)
	}
	saved := append([]int32(nil), ps[pivot:pivot+left]...)
	copy(ps[pivot:pivot+keptRight], ps[pivot+left:pivot+left+keptRight])
	copy(ps[pivot+keptRight:pivot+keptRight+left], saved)
	if bmcg2_sat_solver_solve(s, ps, pivot+keptRight) == GLUCOSE_UNSAT {
		return keptRight
	}
	keptLeft := int32(1)
	if left != 1 {
		keptLeft = bmcg2_sat_solver_minimize_assumptions(s, ps, pivot+keptRight+left, pivot+keptRight)
	}
	return keptRight + keptLeft
}
func bmcg2_sat_solver_add_and(s *SimpSolver, v, a, b, ca, cb, co int32) int32 {
	out := int32(MkLit(Var(v), co != 0))
	x, y := int32(MkLit(Var(a), ca != 0)), int32(MkLit(Var(b), cb != 0))
	for _, row := range [][]int32{{out ^ 1, x}, {out ^ 1, y}, {out, x ^ 1, y ^ 1}} {
		if bmcg2_sat_solver_addclause(s, row, int32(len(row))) == 0 {
			return 0
		}
	}
	return 1
}
func bmcg2_sat_solver_add_xor(s *SimpSolver, v, a, b, complement int32) int32 {
	o := int32(MkLit(Var(v), complement != 0))
	x, y := a<<1, b<<1
	for _, row := range [][]int32{{o ^ 1, x ^ 1, y ^ 1}, {o ^ 1, x, y}, {o, x ^ 1, y}, {o, x, y ^ 1}} {
		if bmcg2_sat_solver_addclause(s, row, 3) == 0 {
			return 0
		}
	}
	return 4
}
func bmcg2_solver_add_xor(s *SimpSolver, v, a, b, complement int32) int32 {
	return bmcg2_sat_solver_add_xor(s, v, a, b, complement)
}
func bmcg2_sat_solver_jftr(s *SimpSolver) int32        { return int32(s.JustificationMode) }
func bmcg2_sat_solver_set_jftr(s *SimpSolver, v int32) { s.JustificationMode = int(v) }

// Solver.h:466-468 intentionally defines these circuit hooks as empty in the
// legacy implementation. They do not install Glucose2 circuit semantics.
func bmcg2_sat_solver_set_var_fanin_lit(s *SimpSolver, v, a, b int32) {
	s.SetVarFaninLits(Var(v), Lit(a), Lit(b))
}
func bmcg2_sat_solver_start_new_round(s *SimpSolver)    { s.SetNewRound() }
func bmcg2_sat_solver_mark_cone(s *SimpSolver, v int32) { s.MarkCone(Var(v)) }

func Glucose2_ReadDimacs(name string, s *SimpSolver) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = ParseDIMACS(f, s)
	return err
}
func glucose2_print_stats(s *SimpSolver, w io.Writer) {
	fmt.Fprintf(w, "c restarts: %d\nc conflicts: %d\nc decisions: %d\nc propagations: %d\n", s.Starts, s.Conflicts, s.Decisions, s.Propagations)
}
func Glucose2_SolveCnf(name string, pars *Glucose2_Pars) (int32, []int32, error) {
	s := NewSimpSolver()
	s.Verbosity = int(boolInt(pars.Verbose))
	if pars.ConflictLimit > 0 {
		s.SetConfBudget(int64(pars.ConflictLimit))
	}
	if err := Glucose2_ReadDimacs(name, s); err != nil {
		return 0, nil, err
	}
	if pars.Preprocess && !s.Eliminate(true) {
		return 20, nil, nil
	}
	status := s.SolveLimited(nil, false)
	if pars.Verbose {
		glucose2_print_stats(s, os.Stdout)
	}
	result := int32(0)
	if status == LTrue {
		result = 10
	} else if status == LFalse {
		result = 20
	}
	return result, glucose2_solver_read_cex(s), nil
}

// graphCNF preserves stable graph object IDs as SAT variables. AIG constants
// use constant one, hence the positive unit clause for variable zero.
func graphCNF(p *aig.Graph, s *SimpSolver, roots []aig.Edge, assertRoots bool) []int32 {
	nodes := p.Nodes()
	s.AddVar(Var(len(nodes) - 1))
	ids := make([]int32, len(nodes))
	for i := range ids {
		ids[i] = int32(i)
	}
	s.AddClause([]Lit{0})
	var visit func(aig.Edge)
	seen := make(map[int]bool)
	visit = func(edge aig.Edge) {
		id := edge.NodeID()
		if seen[id] {
			return
		}
		seen[id] = true
		n := p.Node(edge)
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			visit(n.Fanins[0])
			visit(n.Fanins[1])
		case aig.Buffer, aig.Output:
			visit(n.Fanins[0])
		}
		v := int32(id)
		switch n.Kind {
		case aig.And:
			bmcg2_sat_solver_add_and(s, v, int32(n.Fanins[0].NodeID()), int32(n.Fanins[1].NodeID()), boolInt(n.Fanins[0].Complemented()), boolInt(n.Fanins[1].Complemented()), 0)
			a, b := Lit(n.Fanins[0]), Lit(n.Fanins[1])
			if a.Var() > b.Var() {
				a, b = b, a
			}
			s.SetVarFaninLits(Var(v), a, b)
		case aig.ExclusiveOr:
			bmcg2_sat_solver_add_xor(s, v, int32(n.Fanins[0].NodeID()), int32(n.Fanins[1].NodeID()), boolInt(n.Fanins[0].Complemented() != n.Fanins[1].Complemented()))
			a, b := Lit(n.Fanins[0]), Lit(n.Fanins[1])
			if a.Var() < b.Var() {
				a, b = b, a
			}
			s.SetVarFaninLits(Var(v), a, b)
		case aig.Buffer, aig.Output:
			a := Lit(n.Fanins[0])
			s.AddClause([]Lit{MkLit(Var(v), true), a})
			s.AddClause([]Lit{MkLit(Var(v), false), a.Not()})
		}
	}
	for _, root := range roots {
		visit(root)
	}
	if assertRoots {
		row := make([]Lit, len(roots))
		for i, e := range roots {
			row[i] = Lit(e)
		}
		s.AddClause(row)
	}
	return ids
}
func Glucose2_SolverFromAig(p *aig.Graph, s *SimpSolver) []int32 {
	ids := graphCNF(p, s, p.Outputs(), true)
	return ids
}
func Glucose2_SolverFromAig2(p *aig.Graph, s *SimpSolver) []int32 {
	return Glucose2_SolverFromAig(p, s)
}

func Glucose2_GenerateCubes(solvers [2]*SimpSolver, ciVars, varToIndex []int32, limit int32) string {
	var out strings.Builder
	for count := int32(0); limit == 0 || count < limit; count++ {
		status := bmcg2_sat_solver_solve(solvers[1], nil, 0)
		if status == GLUCOSE_UNSAT {
			break
		}
		if status != GLUCOSE_SAT {
			panic("cube enumeration exhausted solver budget")
		}
		assumptions := make([]int32, len(ciVars))
		for i, v := range ciVars {
			assumptions[i] = int32(MkLit(Var(v), solvers[1].Model[v] != LTrue))
		}
		kept := int32(0)
		offsetStatus := bmcg2_sat_solver_solve(solvers[0], nil, 0)
		if offsetStatus == GLUCOSE_UNDEC {
			panic("cube enumeration exhausted solver budget")
		}
		if offsetStatus == GLUCOSE_SAT {
			status = bmcg2_sat_solver_solve(solvers[0], assumptions, int32(len(assumptions)))
			if status == GLUCOSE_UNDEC {
				panic("cube enumeration exhausted solver budget")
			}
			if status == GLUCOSE_SAT {
				panic("overlapping cube on/off sets")
			}
			if len(assumptions) > 0 {
				kept = bmcg2_sat_solver_minimize_assumptions(solvers[0], assumptions, int32(len(assumptions)), 0)
			}
		}
		cube := []byte(strings.Repeat("-", len(ciVars)))
		var block []int32
		for _, p := range assumptions[:kept] {
			p ^= 1
			index := varToIndex[p>>1]
			if index < 0 {
				continue
			}
			if int(index) >= len(cube) {
				panic("cube index out of range")
			}
			block = append(block, p)
			cube[index] = '0' + byte(p&1)
		}
		out.Write(cube)
		out.WriteString(" 1\n")
		if bmcg2_sat_solver_addclause(solvers[1], block, int32(len(block))) == 0 {
			break
		}
	}
	return out.String()
}
func bmcg2_sat_solver_sop(p *aig.Graph, limit int32) string {
	if p.OutputCount() != 1 {
		panic("SOP extraction requires one output")
	}
	root := p.Outputs()[0]
	solvers := [2]*SimpSolver{bmcg2_sat_solver_start(), bmcg2_sat_solver_start()}
	for i, s := range solvers {
		graphCNF(p, s, []aig.Edge{root}, false)
		lit := int32(root)
		if i == 0 {
			lit ^= 1
		}
		if bmcg2_sat_solver_addclause(s, []int32{lit}, 1) == 0 {
			if i == 0 {
				return " 1\n"
			}
			return " 0\n"
		}
	}
	vars := make([]int32, p.InputCount())
	mapping := make([]int32, len(p.Nodes()))
	for i := range mapping {
		mapping[i] = -1
	}
	for i, e := range p.Inputs() {
		vars[i] = int32(e.NodeID())
		mapping[e.NodeID()] = int32(i)
	}
	return Glucose2_GenerateCubes(solvers, vars, mapping, limit)
}
func bmcg2_sat_solver_print_sop(p *aig.Graph) { fmt.Print(bmcg2_sat_solver_sop(p, 0)) }
func bmcg2_sat_solver_print_sop_lit(p *aig.Graph, lit int32) string {
	return graphSOPForLiteral(p, aig.Edge(lit), 0)
}
func graphSOPForLiteral(p *aig.Graph, root aig.Edge, limit int32) string {
	solvers := [2]*SimpSolver{bmcg2_sat_solver_start(), bmcg2_sat_solver_start()}
	for i, s := range solvers {
		graphCNF(p, s, []aig.Edge{root}, false)
		lit := int32(root)
		if i == 0 {
			lit ^= 1
		}
		s.AddClause([]Lit{Lit(lit)})
	}
	inputs := p.Inputs()
	vars := make([]int32, len(inputs))
	mapping := make([]int32, len(p.Nodes()))
	for i := range mapping {
		mapping[i] = -1
	}
	for i, e := range inputs {
		vars[i] = int32(e.NodeID())
		mapping[e.NodeID()] = int32(i)
	}
	return Glucose2_GenerateCubes(solvers, vars, mapping, limit)
}
func bmcg2_sat_generate_dvars(ciVars []int32, sop string) []int32 {
	counts := make([]int, 2*len(ciVars))
	cubes := 0
	for _, line := range strings.Split(strings.TrimSuffix(sop, "\x00"), "\n") {
		if line == "" {
			continue
		}
		if len(line) < len(ciVars)+2 {
			continue
		}
		cubes++
		for i := range ciVars {
			if line[i] == '1' {
				counts[2*i]++
			} else if line[i] == '0' {
				counts[2*i+1]++
			}
		}
	}
	var literals []int32
	for lit, count := range counts {
		if cubes > 0 && count == cubes {
			literals = append(literals, ciVars[lit/2]*2+int32(lit&1))
		}
	}
	sort.Slice(literals, func(i, j int) bool { return literals[i] < literals[j] })
	return literals
}
func graphSupport(p *aig.Graph, root aig.Edge) (objects []int32, inputs []int32) {
	seen := make(map[int]bool)
	var visit func(aig.Edge)
	visit = func(e aig.Edge) {
		id := e.NodeID()
		if seen[id] {
			return
		}
		seen[id] = true
		n := p.Node(e)
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			visit(n.Fanins[0])
			visit(n.Fanins[1])
		case aig.Buffer, aig.Output:
			visit(n.Fanins[0])
		}
		objects = append(objects, int32(id))
		if n.Kind == aig.Input {
			inputs = append(inputs, int32(id))
		}
	}
	visit(root)
	return
}
func Gia_ManSatAndCollect2_rec(p *aig.Graph, obj int32, objects, ciVars *[]int32) int32 {
	objs, inputs := graphSupport(p, aig.Edge(obj<<1))
	index := make(map[int32]int32, len(*objects))
	for i, v := range *objects {
		index[v] = int32(i)
	}
	isCI := make(map[int32]bool)
	for _, v := range inputs {
		isCI[v] = true
	}
	for _, v := range objs {
		if _, ok := index[v]; ok {
			continue
		}
		index[v] = int32(len(*objects))
		*objects = append(*objects, v)
		if ciVars != nil && isCI[v] {
			*ciVars = append(*ciVars, index[v])
		}
	}
	return index[obj]
}
func Gia_ManQuantLoadCnf2(p *aig.Graph, objects []int32, solvers [2]*SimpSolver) {
	roots := make([]aig.Edge, len(objects))
	for i, id := range objects {
		roots[i] = aig.Edge(id << 1)
	}
	for _, s := range solvers {
		if s != nil {
			s.Reset()
			s.SetIncrementalMode()
			graphCNF(p, s, roots, false)
		}
	}
}
func Gia_ManFactorSop2(p *aig.Graph, ciObjects []int32, sop string, hash int32) int32 {
	result := p.Const0()
	for _, line := range strings.Split(strings.TrimSuffix(sop, "\x00"), "\n") {
		if line == "" {
			continue
		}
		if strings.TrimSpace(line) == "0" {
			continue
		}
		if strings.TrimSpace(line) == "1" {
			return int32(p.Const1())
		}
		if len(line) < len(ciObjects)+2 {
			if strings.TrimSpace(line) == "1" {
				return int32(p.Const1())
			}
			continue
		}
		cube := p.Const1()
		for i, id := range ciObjects {
			switch line[i] {
			case '0':
				cube = p.AND(cube, aig.Edge(id<<1).Not())
			case '1':
				cube = p.AND(cube, aig.Edge(id<<1))
			}
		}
		result = p.AND(result.Not(), cube.Not()).Not()
	}
	return int32(result)
}
func bmcg2_sat_solver_quantify(solvers [2]*SimpSolver, p *aig.Graph, lit, hash int32, keep func(any, int32) int32, data any, divisors *[]int32) int32 {
	root := aig.Edge(lit)
	if divisors != nil {
		*divisors = nil
	}
	if root.NodeID() == 0 {
		return lit
	}
	_, support := graphSupport(p, root)
	ciIndices := make(map[int32]int32)
	for i, e := range p.Inputs() {
		ciIndices[int32(e.NodeID())] = int32(i)
	}
	mapping := make([]int32, len(p.Nodes()))
	for i := range mapping {
		mapping[i] = -1
	}
	kept := 0
	for i, v := range support {
		if keep(data, ciIndices[v]) != 0 {
			mapping[v] = int32(i)
			kept++
		}
	}
	for i, s := range solvers {
		if s == nil {
			panic("quantification needs two solvers")
		}
		s.Reset()
		s.SetIncrementalMode()
		graphCNF(p, s, []aig.Edge{root}, false)
		unit := lit
		if i == 0 {
			unit ^= 1
		}
		s.AddClause([]Lit{Lit(unit)})
	}
	for i, solver := range solvers {
		switch bmcg2_sat_solver_solve(solver, nil, 0) {
		case GLUCOSE_UNSAT:
			if i == 0 {
				return int32(p.Const1())
			}
			return int32(p.Const0())
		case GLUCOSE_UNDEC:
			// The SAT route cannot establish the constant or compute divisors.
			// Direct cofactoring is exact when no divisor list is requested.
			if divisors == nil {
				return bmcg2_sat_solver_quantify2(p, lit, hash, keep, data, nil)
			}
			return -1
		}
	}
	if kept == 0 {
		return int32(p.Const1())
	}
	if kept == len(support) {
		return lit
	}
	sop := Glucose2_GenerateCubes(solvers, support, mapping, 0)
	if divisors != nil {
		*divisors = bmcg2_sat_generate_dvars(support, sop)
	}
	return Gia_ManFactorSop2(p, support, sop, hash)
}
func graphCofactor(p *aig.Graph, root aig.Edge, variable int, value bool, memo map[aig.Edge]aig.Edge) aig.Edge {
	if r, ok := memo[root]; ok {
		return r
	}
	if root.NodeID() == variable {
		r := p.Const0()
		if value {
			r = p.Const1()
		}
		if root.Complemented() {
			r = r.Not()
		}
		return r
	}
	n := p.Node(root)
	regular := aig.Edge(root.NodeID() << 1)
	result := regular
	switch n.Kind {
	case aig.And:
		result = p.AND(graphCofactor(p, n.Fanins[0], variable, value, memo), graphCofactor(p, n.Fanins[1], variable, value, memo))
	case aig.ExclusiveOr:
		a, b := graphCofactor(p, n.Fanins[0], variable, value, memo), graphCofactor(p, n.Fanins[1], variable, value, memo)
		result = p.AND(p.AND(a, b.Not()).Not(), p.AND(a.Not(), b).Not()).Not()
	case aig.Buffer, aig.Output:
		result = graphCofactor(p, n.Fanins[0], variable, value, memo)
	}
	if root.Complemented() {
		result = result.Not()
	}
	memo[root] = result
	return result
}
func bmcg2_sat_solver_quantify2(p *aig.Graph, lit, hash int32, keep func(any, int32) int32, data any, divisors *[]int32) int32 {
	if divisors != nil {
		*divisors = nil
	}
	root := aig.Edge(lit)
	_, support := graphSupport(p, root)
	indices := make(map[int]int32)
	for i, e := range p.Inputs() {
		indices[e.NodeID()] = int32(i)
	}
	for _, id := range support {
		if keep(data, indices[int(id)]) == 0 {
			a := graphCofactor(p, root, int(id), false, make(map[aig.Edge]aig.Edge))
			b := graphCofactor(p, root, int(id), true, make(map[aig.Edge]aig.Edge))
			root = p.AND(a.Not(), b.Not()).Not()
		}
	}
	return int32(root)
}
func bmcg2_sat_solver_quantify_test(solvers [2]*SimpSolver, p *aig.Graph, lit, hash int32, keep func(any, int32) int32, data any, divisors *[]int32) int32 {
	return bmcg2_sat_solver_quantify(solvers, p, lit, hash, keep, data, divisors)
}
func Gia_ManCiIsToKeep2(data any, i int32) int32 { return boolInt(i%5 != 0) }
func bmcg2_sat_solver_equiv_overlap_check(s *SimpSolver, p *aig.Graph, a, b, equivalence int32) int32 {
	s.Reset()
	s.SetIncrementalMode()
	graphCNF(p, s, []aig.Edge{aig.Edge(a), aig.Edge(b)}, false)
	if equivalence != 0 {
		for _, as := range [][]int32{{a, int32(Lit(b).Not())}, {int32(Lit(a).Not()), b}} {
			if bmcg2_sat_solver_solve(s, as, int32(len(as))) != GLUCOSE_UNSAT {
				return 0
			}
		}
		return 1
	}
	return boolInt(bmcg2_sat_solver_solve(s, []int32{a, b}, 2) == GLUCOSE_SAT)
}
func Glucose2_QuantifyAigTest(p *aig.Graph) []int32 {
	results := make([]int32, p.OutputCount())
	for i, root := range p.Outputs() {
		results[i] = bmcg2_sat_solver_quantify2(p, int32(root), 1, Gia_ManCiIsToKeep2, nil, nil)
	}
	return results
}
func Glucose2_CheckTwoNodesTest(p *aig.Graph) [][2]int32 {
	outputs := p.Outputs()
	s := NewSimpSolver()
	var results [][2]int32
	for i := 0; i+1 < len(outputs); i += 2 {
		results = append(results, [2]int32{bmcg2_sat_solver_equiv_overlap_check(s, p, int32(outputs[i]), int32(outputs[i+1]), 0), bmcg2_sat_solver_equiv_overlap_check(s, p, int32(outputs[i]), int32(outputs[i+1]), 1)})
	}
	return results
}
func Glucose2_SolveAig(p *aig.Graph, pars *Glucose2_Pars) (int32, []int32) {
	s := NewSimpSolver()
	s.Verbosity = int(boolInt(pars.Verbose))
	if pars.ConflictLimit > 0 {
		s.SetConfBudget(int64(pars.ConflictLimit))
	}
	Glucose2_SolverFromAig(p, s)
	if pars.Preprocess && !s.Eliminate(true) {
		return 20, nil
	}
	status := s.SolveLimited(nil, false)
	if status == LFalse {
		return 20, nil
	}
	if status.IsUndef() {
		return 0, nil
	}
	model := make([]int32, p.InputCount())
	for i, e := range p.Inputs() {
		model[i] = boolInt(s.Model[e.NodeID()] == LTrue)
	}
	return 10, model
}
