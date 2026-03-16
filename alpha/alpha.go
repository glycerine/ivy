// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_alpha.py.
//
// Package alpha implements predicate abstraction and concept-space
// enumeration for Ivy verification. It provides the Alpha and
// PredicateAlpha functions for computing abstract post-images, along
// with ProgressiveDomain for incremental concept-space exploration,
// and RelAlg1/RelAlg2/RelAlg3 for relational algebra on abstract states.
package alpha

import (
	"fmt"

	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
	"github.com/glycerine/goivy/solver"
	"github.com/glycerine/goivy/webui"
	"github.com/glycerine/goivy/z3bridge"
)

// TestBottom controls whether UNSAT checking is performed
// on the concrete state before abstraction.
// Corresponds to Python's test_bottom = True.
var TestBottom = true

// Log controls verbose logging of alpha operations.
// Corresponds to Python's log = False.
var Log = false

// --- Domain interface for Alpha ---

// AlphaDomain represents the domain information needed by Alpha.
// In Python this is state.domain with concept_spaces and background_theory.
type AlphaDomain struct {
	ConceptSpaces    []ConceptSpaceEntry
	BackgroundTheory func(inScope map[string]bool) *clauseops.Clauses
	AbstrPreds       []*clauseops.Clauses
}

// ConceptSpaceEntry pairs an atom with a concept space expression.
// Corresponds to Python's (atom, cs) tuples in concept_spaces.
type ConceptSpaceEntry struct {
	Atom  *webui.CSAtom
	Space webui.CSNode
}

// AlphaState represents the state to abstract.
// In Python this is the state argument to alpha() and predicate_alpha().
type AlphaState struct {
	Clauses *clauseops.Clauses
	Domain  *AlphaDomain
	InScope map[string]bool
}

// --- Alpha function ---

// Alpha computes the abstract post-image of a state using concept spaces.
// Corresponds to Python's alpha(state).
func Alpha(state *AlphaState) {
	d := NewProgressiveDomain(state.Domain.ConceptSpaces, false)
	var bgTheory *clauseops.Clauses
	if state.Domain.BackgroundTheory != nil {
		bgTheory = state.Domain.BackgroundTheory(state.InScope)
	} else {
		bgTheory = clauseops.TrueClauses(nil)
	}
	state.Clauses = d.Post(state.Clauses, bgTheory, nil, nil)
}

// --- ProgressiveDomain ---

// ProgressiveDomain implements incremental concept-space exploration.
// It maintains a solver with the concrete state and background theory,
// and enumerates concept spaces to compute an abstract post-image.
// Corresponds to Python's ProgressiveDomain class.
type ProgressiveDomain struct {
	conceptSpaces  []ConceptSpaceEntry
	verbose        bool
	slvr           *solver.Solver
	z3solver       *z3bridge.Solver
	cubeMemo       map[string]bool // Z3 expr ID -> sat result
	inhabitedCubes map[string]bool // Z3 expr ID -> inhabited
	z3Cubes        []z3bridge.Expr // prevent GC of Z3 cubes
	memo           map[string]webui.CSMemoEntry
	inferred       [][]lg.Node
	unsat          bool
	newSym         map[string]*lg.Const
}

// NewProgressiveDomain creates a new ProgressiveDomain.
// Corresponds to Python's ProgressiveDomain.__init__.
func NewProgressiveDomain(cs []ConceptSpaceEntry, verbose bool) *ProgressiveDomain {
	return &ProgressiveDomain{
		conceptSpaces: cs,
		verbose:       verbose,
	}
}

// AddConceptSpace appends a concept space entry.
// Corresponds to Python's add_concept_space.
func (pd *ProgressiveDomain) AddConceptSpace(atom *webui.CSAtom, space webui.CSNode) {
	pd.conceptSpaces = append(pd.conceptSpaces, ConceptSpaceEntry{Atom: atom, Space: space})
}

// cubeID computes a string identifier for a cube (list of literals).
// Corresponds to Python's cube_id.
func (pd *ProgressiveDomain) cubeID(cube []*il.Literal) (string, error) {
	z3cube, err := pd.slvr.CubeToZ3(cube)
	if err != nil {
		return "", err
	}
	pd.z3Cubes = append(pd.z3Cubes, z3cube) // prevent GC
	return z3cube.String(), nil
}

// inhabitedCube marks a cube as inhabited.
// Corresponds to Python's inhabited_cube.
func (pd *ProgressiveDomain) inhabitedCube(cube []*il.Literal, truth bool) {
	cube = canonizeClause(cube)
	id, err := pd.cubeID(cube)
	if err != nil {
		return
	}
	if _, exists := pd.inhabitedCubes[id]; !exists {
		if Log {
			fmt.Printf("inhabited: %v\n", cube)
		}
		pd.inhabitedCubes[id] = truth
	}
}

// inhabitedLit marks a single literal as inhabited.
// Corresponds to Python's inhabited_lit.
func (pd *ProgressiveDomain) inhabitedLit(lit *il.Literal) {
	pd.inhabitedCube([]*il.Literal{lit}, true)
}

// modelCheck performs model checking on the current solver state.
// In Python this is effectively a no-op (starts with 'return').
// Corresponds to Python's model_check.
func (pd *ProgressiveDomain) modelCheck() {
	// Python version begins with "return" making it effectively a no-op.
	return
}

// unfoldDefs adds definition instances for a cube to the solver.
// Corresponds to Python's unfold_defs.
func (pd *ProgressiveDomain) unfoldDefs(cube []*il.Literal) {
	// In Python: definition_instances(cube_to_formula(cube))
	// This is a placeholder; full implementation would compute definition
	// instances and add them to the solver.
	_ = cube
}

// testCube tests whether a cube is consistent with the solver state.
// Returns true if the cube is satisfiable (inhabited).
// Corresponds to Python's test_cube.
func (pd *ProgressiveDomain) testCube(cube []*il.Literal) bool {
	canonCube := canonizeClause(cube)
	if Log {
		strs := make([]string, len(canonCube))
		for i, c := range canonCube {
			strs[i] = c.String()
		}
		fmt.Printf("cube: %v\n", strs)
	}
	myID, err := pd.cubeID(canonCube)
	if err != nil {
		return false
	}
	if val, exists := pd.inhabitedCubes[myID]; exists {
		if Log {
			strs := make([]string, len(cube))
			for i, c := range cube {
				strs[i] = c.String()
			}
			fmt.Printf("cached: %v\n", strs)
		}
		return val
	}
	if Log {
		strs := make([]string, len(cube))
		for i, c := range cube {
			strs[i] = c.String()
		}
		fmt.Printf("test: %v\n", strs)
	}

	// Rename clause with new symbols
	renamedCube := renameClause(cube, pd.newSym)

	// Collect used variables and create Skolem substitution
	vs := usedVariablesClause(renamedCube)
	subs := make(map[lg.Node]lg.Node, len(vs))
	for _, v := range vs {
		subs[v] = varToSkolem("__c", v)
	}
	scube := substituteClause(renamedCube, subs)

	// Unfold definitions
	pd.unfoldDefs(scube)

	// Check cube satisfiability
	res, err := pd.slvr.CheckCube(pd.z3solver, scube)
	if err != nil {
		return false
	}

	if res {
		if len(cube) <= 4 {
			pd.modelCheck()
		}
	} else {
		// Cube is unsat - infer negation
		negated := make([]lg.Node, len(cube))
		for i, lit := range cube {
			negated[i] = negateLiteral(lit)
		}
		pd.inferred = append(pd.inferred, negated)
		if Log {
			strs := make([]string, len(canonCube))
			for i, c := range canonCube {
				strs[i] = c.String()
			}
			fmt.Printf("uninhabited: %v\n", strs)
		}
		pd.inhabitedCubes[myID] = false
	}
	return res
}

// postInit initializes the solver state for a post computation.
// Corresponds to Python's post_init.
func (pd *ProgressiveDomain) postInit(
	theory, backgroundTheory *clauseops.Clauses,
	newSym map[string]*lg.Const,
	toKeep []string,
) {
	pd.newSym = newSym
	pd.slvr = solver.New()
	pd.z3solver = pd.slvr.NewZ3Solver()
	pd.cubeMemo = make(map[string]bool)
	pd.inhabitedCubes = make(map[string]bool)
	pd.z3Cubes = nil
	pd.memo = make(map[string]webui.CSMemoEntry)

	if Log {
		fmt.Printf("concrete state: %s\n", theory)
		fmt.Printf("background: %s\n", backgroundTheory)
	}

	combined := clauseops.AndClausesTyped(theory, backgroundTheory)
	if err := pd.slvr.AddClauses(pd.z3solver, combined); err != nil {
		pd.unsat = true
		return
	}

	if TestBottom {
		result := pd.z3solver.Check()
		pd.unsat = (result == z3bridge.Unsat)
	} else {
		pd.unsat = false
	}

	if pd.unsat {
		// In Python: prints unsat_core. We just note it.
		fmt.Printf("core: <unsat state detected>\n")
	}
}

// postStep performs one step of concept space enumeration.
// Returns the inferred clauses.
// Corresponds to Python's post_step.
func (pd *ProgressiveDomain) postStep(conceptSpaces []ConceptSpaceEntry) *clauseops.Clauses {
	if pd.unsat {
		return clauseops.FalseClauses(nil)
	}
	pd.inferred = nil

	for _, entry := range conceptSpaces {
		if Log {
			fmt.Printf("concept space: %s\n", entry.Atom)
		}

		// Create the test function that wraps testCube
		testFn := func(csLits []*webui.CSLiteral) bool {
			ilLits := csLitsToILLits(csLits)
			return pd.testCube(ilLits)
		}

		concepts := entry.Space.Enumerate(pd.memo, testFn)
		if Log {
			fmt.Printf("result: %v\n", concepts)
		}

		// Store result in memo for subsequent concept spaces
		params := make([]webui.CSTerm, len(entry.Atom.Args))
		copy(params, entry.Atom.Args)
		pd.memo[entry.Atom.RelName] = webui.CSMemoEntry{
			Params: params,
			Value:  concepts,
		}
	}

	res := pd.inferred
	if Log {
		fmt.Printf("inferred: %v\n", res)
	}
	pd.inferred = nil

	// Convert inferred to Clauses
	fmlas := make([]lg.Node, len(res))
	for i, clause := range res {
		if len(clause) == 1 {
			fmlas[i] = clause[0]
		} else {
			fmlas[i] = &lg.Or{Terms: clause}
		}
	}
	return clauseops.NewClauses(fmlas, nil, nil)
}

// postQuit cleans up after a post computation.
// Corresponds to Python's post_quit.
func (pd *ProgressiveDomain) postQuit() {
	pd.newSym = nil
	pd.cubeMemo = nil
	pd.inhabitedCubes = nil
	pd.z3Cubes = nil
	pd.z3solver = nil
	pd.slvr = nil
	pd.memo = nil
}

// Post computes the abstract post-image given a concrete theory,
// background theory, new symbol map, and symbols to keep.
// Corresponds to Python's post.
func (pd *ProgressiveDomain) Post(
	theory, backgroundTheory *clauseops.Clauses,
	newSym map[string]*lg.Const,
	toKeep []string,
) *clauseops.Clauses {
	pd.postInit(theory, backgroundTheory, newSym, toKeep)
	res := pd.postStep(pd.conceptSpaces)
	pd.postQuit()
	return res
}

// --- Relational algebra ---

// VarCorr computes a variable correspondence between two term lists.
// Returns a list of (index1, index2) pairs where both positions hold
// the same variable.
// Corresponds to Python's var_corr.
func VarCorr(terms1, terms2 []lg.Node) [][2]int {
	d := make(map[string]int)
	for i, t := range terms2 {
		if v, ok := t.(*lg.Var); ok {
			d[v.Name] = i
		}
	}
	var result [][2]int
	for j, t := range terms1 {
		if v, ok := t.(*lg.Var); ok {
			if idx, exists := d[v.Name]; exists {
				result = append(result, [2]int{j, idx})
			}
		}
	}
	return result
}

// firstSeen returns true if elem is not yet in memo, and adds it.
// Corresponds to Python's first_seen.
func firstSeen(memo map[string]bool, elem string) bool {
	if memo[elem] {
		return false
	}
	memo[elem] = true
	return true
}

// cutRow extracts elements at the given column indices.
// Corresponds to Python's cut_row.
func cutRow(row []lg.Node, goodCols []int) []lg.Node {
	result := make([]lg.Node, len(goodCols))
	for i, col := range goodCols {
		result[i] = row[col]
	}
	return result
}

// RelTable represents a relational table: a tuple of (variables, rows).
// Corresponds to Python's (v, rows) tuples.
type RelTable struct {
	Vars []lg.Node
	Rows [][]lg.Node
}

// compactTable removes redundant columns from a relation table.
// Corresponds to Python's compact_table.
func compactTable(tab *RelTable) *RelTable {
	memo := make(map[string]bool)
	var goodCols []int
	for i, t := range tab.Vars {
		if v, ok := t.(*lg.Var); ok && firstSeen(memo, v.Name) {
			goodCols = append(goodCols, i)
		}
	}
	newVars := cutRow(tab.Vars, goodCols)
	newRows := make([][]lg.Node, len(tab.Rows))
	for i, row := range tab.Rows {
		newRows[i] = cutRow(row, goodCols)
	}
	return &RelTable{Vars: newVars, Rows: newRows}
}

// --- RelAlg1 ---

// RelAlg1 implements relational algebra using model instances.
// Corresponds to Python's RelAlg1 class.
type RelAlg1 struct {
	Slvr   *solver.Solver
	Z3Slvr *z3bridge.Solver
	Model  *solver.HerbrandModel
	Parent *ProgressiveDomain
}

// NewRelAlg1 creates a new RelAlg1.
// Corresponds to Python's RelAlg1.__init__.
func NewRelAlg1(slvr *solver.Solver, z3slvr *z3bridge.Solver, parent *ProgressiveDomain) *RelAlg1 {
	return &RelAlg1{
		Slvr:   slvr,
		Z3Slvr: z3slvr,
		Parent: parent,
	}
}

// Prim evaluates a primitive literal against the model.
// Returns a RelTable of matching instances.
// Corresponds to Python's RelAlg1.prim.
func (ra *RelAlg1) Prim(lit *il.Literal) *RelTable {
	// Get model instances
	app, ok := lit.Atom.(*lg.Apply)
	if !ok {
		return &RelTable{}
	}

	// Check the literal against the model
	posLit := il.NewLiteral(lit.Polarity, lit.Atom)
	_, rows := ra.Model.Check(posLit)

	// Build result table
	resultRows := make([][]lg.Node, len(rows))
	for i, row := range rows {
		resultRow := make([]lg.Node, len(row))
		for j, c := range row {
			resultRow[j] = c
		}
		resultRows[i] = resultRow
	}

	tab := compactTable(&RelTable{Vars: nodeSlice(app.Terms), Rows: resultRows})

	if ra.Parent != nil && len(tab.Rows) > 0 {
		ra.Parent.inhabitedLit(lit)
	}
	return tab
}

// Prod computes the product (join) of two relation tables.
// Corresponds to Python's RelAlg1.prod.
func (ra *RelAlg1) Prod(x, y *RelTable) *RelTable {
	corr := VarCorr(x.Vars, y.Vars)
	var rows [][]lg.Node
	if len(corr) > 0 {
		xc, yc := corr[0][0], corr[0][1]
		index := make(map[string][]int) // variable rep -> row indices in y
		for i, ry := range y.Rows {
			key := ry[yc].String()
			index[key] = append(index[key], i)
		}
		for _, xr := range x.Rows {
			key := xr[xc].String()
			for _, yi := range index[key] {
				yr := y.Rows[yi]
				if allMatch(xr, yr, corr) {
					combined := make([]lg.Node, 0, len(xr)+len(yr))
					combined = append(combined, xr...)
					combined = append(combined, yr...)
					rows = append(rows, combined)
				}
			}
		}
	} else {
		// Cross product
		for _, xr := range x.Rows {
			for _, yr := range y.Rows {
				if allMatch(xr, yr, corr) {
					combined := make([]lg.Node, 0, len(xr)+len(yr))
					combined = append(combined, xr...)
					combined = append(combined, yr...)
					rows = append(rows, combined)
				}
			}
		}
	}
	combinedVars := make([]lg.Node, 0, len(x.Vars)+len(y.Vars))
	combinedVars = append(combinedVars, x.Vars...)
	combinedVars = append(combinedVars, y.Vars...)
	return compactTable(&RelTable{Vars: combinedVars, Rows: rows})
}

// Subst applies a substitution to a table's variable names.
// Corresponds to Python's RelAlg1.subst.
func (ra *RelAlg1) Subst(tab *RelTable, subst map[string]lg.Node) *RelTable {
	newVars := make([]lg.Node, len(tab.Vars))
	for i, v := range tab.Vars {
		if vv, ok := v.(*lg.Var); ok {
			if repl, exists := subst[vv.Name]; exists {
				newVars[i] = repl
			} else {
				newVars[i] = v
			}
		} else {
			newVars[i] = v
		}
	}
	return &RelTable{Vars: newVars, Rows: tab.Rows}
}

// Empty returns true if the table has no rows.
// Corresponds to Python's RelAlg1.empty.
func (ra *RelAlg1) Empty(tab *RelTable) bool {
	return len(tab.Rows) == 0
}

// --- RelAlg2 ---

// RelAlg2 implements relational algebra using Z3 cubes.
// Corresponds to Python's RelAlg2 class.
type RelAlg2 struct {
	Slvr       *solver.Solver
	Z3Slvr     *z3bridge.Solver
	TempSolver *z3bridge.Solver
	Model      *solver.HerbrandModel
	Parent     *ProgressiveDomain
	Numbering  map[string]int
	NextNumber int
	PrimCache  map[string][]z3bridge.Expr
	PrimList   []z3bridge.Expr // prevent GC
	NewSym     map[string]*lg.Const
	Hm         *solver.HerbrandModel
}

// NewRelAlg2 creates a new RelAlg2.
// Corresponds to Python's RelAlg2.__init__.
func NewRelAlg2(
	slvr *solver.Solver,
	z3slvr *z3bridge.Solver,
	newSym map[string]*lg.Const,
	parent *ProgressiveDomain,
) *RelAlg2 {
	return &RelAlg2{
		Slvr:       slvr,
		Z3Slvr:     z3slvr,
		TempSolver: slvr.NewZ3Solver(),
		Parent:     parent,
		Numbering:  make(map[string]int),
		NextNumber: 0,
		PrimCache:  make(map[string][]z3bridge.Expr),
		NewSym:     newSym,
	}
}

// IsSat checks if a formula is satisfiable in the temp solver.
// Corresponds to Python's is_sat.
func (ra *RelAlg2) IsSat(f z3bridge.Expr) bool {
	ra.TempSolver.Push()
	defer ra.TempSolver.Pop()
	ra.TempSolver.Assert(f)
	result := ra.TempSolver.Check()
	return result != z3bridge.Unsat
}

// Prim evaluates a primitive literal.
// Returns a list of Z3 cube expressions.
// Corresponds to Python's RelAlg2.prim.
func (ra *RelAlg2) Prim(lit *il.Literal) []z3bridge.Expr {
	z3lit, err := ra.Slvr.LiteralToZ3(lit)
	if err != nil {
		return nil
	}
	id := z3lit.String()
	if cached, ok := ra.PrimCache[id]; ok {
		return cached
	}

	var cubes []z3bridge.Expr

	// Rename literal and get ground instances from Herbrand model
	renamedLit := renameLit(lit, ra.NewSym)
	if ra.Hm != nil {
		vs, rows := ra.Hm.Check(renamedLit)
		ctx := ra.Slvr.Context()
		for _, row := range rows {
			eqs := make([]z3bridge.Expr, 0, len(vs))
			for j, v := range vs {
				origTerms := getAtomArgs(lit.Atom)
				if j < len(origTerms) {
					zt, err1 := ra.Slvr.FormulaToZ3(origTerms[j])
					zc, err2 := ra.Slvr.FormulaToZ3(row[j])
					if err1 == nil && err2 == nil {
						eqs = append(eqs, ctx.Eq(zt, zc))
					}
				} else {
					zv, err1 := ra.Slvr.FormulaToZ3(v)
					zc, err2 := ra.Slvr.FormulaToZ3(row[j])
					if err1 == nil && err2 == nil {
						eqs = append(eqs, ctx.Eq(zv, zc))
					}
				}
			}
			if len(eqs) > 0 {
				cube := ctx.And(eqs...)
				cubes = append(cubes, cube)
			}
		}
	}

	if ra.Parent != nil && len(cubes) > 0 {
		ra.Parent.inhabitedLit(lit)
	}

	ra.PrimCache[id] = cubes
	ra.PrimList = append(ra.PrimList, z3lit) // prevent GC
	return cubes
}

// Top returns a list with a single true cube.
// Corresponds to Python's RelAlg2.top.
func (ra *RelAlg2) Top() []z3bridge.Expr {
	ctx := ra.Slvr.Context()
	return []z3bridge.Expr{ctx.BoolVal(true)}
}

// Prod computes the product of two cube lists.
// Corresponds to Python's RelAlg2.prod.
func (ra *RelAlg2) Prod(x, y []z3bridge.Expr) []z3bridge.Expr {
	ctx := ra.Slvr.Context()
	var cubes []z3bridge.Expr
	for _, xr := range x {
		for _, yr := range y {
			combined := ctx.And(xr, yr)
			if ra.IsSat(combined) {
				cubes = append(cubes, combined)
			}
		}
	}
	return cubes
}

// Subst applies a substitution to a cube list.
// Corresponds to Python's RelAlg2.subst.
func (ra *RelAlg2) Subst(tab []z3bridge.Expr, subst map[string]lg.Node) []z3bridge.Expr {
	// Build Z3-level substitution
	ctx := ra.Slvr.Context()
	var fromExprs, toExprs []z3bridge.Expr
	for name, node := range subst {
		c := lg.NewConst(name, node.NodeSort())
		zFrom, err1 := ra.Slvr.FormulaToZ3(c)
		zTo, err2 := ra.Slvr.FormulaToZ3(node)
		if err1 == nil && err2 == nil {
			fromExprs = append(fromExprs, zFrom)
			toExprs = append(toExprs, zTo)
		}
	}
	if len(fromExprs) == 0 {
		return tab
	}
	result := make([]z3bridge.Expr, len(tab))
	for i, expr := range tab {
		result[i] = ctx.Substitute(expr, fromExprs, toExprs)
	}
	return result
}

// Empty returns true if the cube list is empty.
// Corresponds to Python's RelAlg2.empty.
func (ra *RelAlg2) Empty(tab []z3bridge.Expr) bool {
	return len(tab) == 0
}

// --- RelAlg3 ---

// RelAlg3 extends RelAlg2 with a different prim method that uses
// HerbrandModel.Check instead of ground_instances.
// Corresponds to Python's RelAlg3 class.
type RelAlg3 struct {
	*RelAlg2
}

// NewRelAlg3 creates a new RelAlg3.
func NewRelAlg3(
	slvr *solver.Solver,
	z3slvr *z3bridge.Solver,
	newSym map[string]*lg.Const,
	parent *ProgressiveDomain,
) *RelAlg3 {
	return &RelAlg3{
		RelAlg2: NewRelAlg2(slvr, z3slvr, newSym, parent),
	}
}

// Prim evaluates a primitive literal using HerbrandModel.Check.
// Corresponds to Python's RelAlg3.prim.
func (ra *RelAlg3) Prim(lit *il.Literal) []z3bridge.Expr {
	if Log {
		fmt.Printf("prim: %s\n", lit)
	}
	z3lit, err := ra.Slvr.LiteralToZ3(lit)
	if err != nil {
		return nil
	}
	id := z3lit.String()
	if cached, ok := ra.PrimCache[id]; ok {
		return cached
	}

	renamedLit := renameLit(lit, ra.NewSym)
	var cubes []z3bridge.Expr

	if ra.Hm != nil {
		vs, rows := ra.Hm.Check(renamedLit)
		ctx := ra.Slvr.Context()
		for _, row := range rows {
			eqs := make([]z3bridge.Expr, 0, len(vs))
			for j, v := range vs {
				zv, err1 := ra.Slvr.FormulaToZ3(v)
				zc, err2 := ra.Slvr.FormulaToZ3(row[j])
				if err1 == nil && err2 == nil {
					eqs = append(eqs, ctx.Eq(zv, zc))
				}
			}
			if len(eqs) > 0 {
				cube := ctx.And(eqs...)
				cubes = append(cubes, cube)
			}
		}
	}

	if ra.Parent != nil && len(cubes) > 0 {
		ra.Parent.inhabitedLit(lit)
	}

	ra.PrimCache[id] = cubes
	ra.PrimList = append(ra.PrimList, z3lit)
	return cubes
}

// --- PredicateAlpha ---

// PredicateAlpha computes the predicate abstraction of a state.
// For each abstraction predicate, checks whether it is implied by
// the concrete state; the result is the conjunction of implied predicates.
// Corresponds to Python's predicate_alpha.
func PredicateAlpha(state *AlphaState) {
	fmt.Println("running predicate alpha")
	slvr := solver.New()
	z3slvr := slvr.NewZ3Solver()

	var bgTheory *clauseops.Clauses
	if state.Domain.BackgroundTheory != nil {
		bgTheory = state.Domain.BackgroundTheory(nil)
	} else {
		bgTheory = clauseops.TrueClauses(nil)
	}

	combined := clauseops.AndClausesTyped(state.Clauses, bgTheory)
	if err := slvr.AddClauses(z3slvr, combined); err != nil {
		return
	}

	res := clauseops.TrueClauses(nil)
	for _, pred := range state.Domain.AbstrPreds {
		z3slvr.Push()
		dual := clauseops.NegateClauses(pred)
		if err := slvr.AddClauses(z3slvr, dual); err != nil {
			z3slvr.Pop()
			continue
		}
		cr := z3slvr.Check()
		if Log {
			fmt.Printf("predicate: %s result %v\n", pred, cr)
		}
		if cr == z3bridge.Unsat {
			res = clauseops.AndClausesTyped(res, pred)
		}
		z3slvr.Pop()
	}
	state.Clauses = res
}

// --- Helper functions ---

// canonizeClause sorts and deduplicates literals in a clause.
// Corresponds to Python's canonize_clause.
func canonizeClause(cube []*il.Literal) []*il.Literal {
	if len(cube) == 0 {
		return cube
	}
	// Deduplicate by string representation
	seen := make(map[string]bool)
	var result []*il.Literal
	for _, lit := range cube {
		key := lit.String()
		if !seen[key] {
			seen[key] = true
			result = append(result, lit)
		}
	}
	return result
}

// renameClause renames symbols in a clause using the newSym map.
// Corresponds to Python's rename_clause.
func renameClause(cube []*il.Literal, newSym map[string]*lg.Const) []*il.Literal {
	if len(newSym) == 0 {
		return cube
	}
	result := make([]*il.Literal, len(cube))
	for i, lit := range cube {
		result[i] = renameLit(lit, newSym)
	}
	return result
}

// renameLit renames symbols in a literal using the newSym map.
// Corresponds to Python's rename_lit.
func renameLit(lit *il.Literal, newSym map[string]*lg.Const) *il.Literal {
	if len(newSym) == 0 {
		return lit
	}
	newAtom := renameNode(lit.Atom, newSym)
	return il.NewLiteral(lit.Polarity, newAtom)
}

// renameNode renames constant symbols in a node using the newSym map.
func renameNode(node lg.Node, newSym map[string]*lg.Const) lg.Node {
	if len(newSym) == 0 {
		return node
	}
	switch t := node.(type) {
	case *lg.Const:
		if repl, ok := newSym[t.Name]; ok {
			return repl
		}
		return t
	case *lg.Apply:
		newFunc := renameNode(t.Func, newSym)
		newArgs := make([]lg.Node, len(t.Terms))
		for i, arg := range t.Terms {
			newArgs[i] = renameNode(arg, newSym)
		}
		return &lg.Apply{Func: newFunc, Terms: newArgs}
	case *lg.Eq:
		return &lg.Eq{T1: renameNode(t.T1, newSym), T2: renameNode(t.T2, newSym)}
	case *lg.Not:
		return &lg.Not{Body: renameNode(t.Body, newSym)}
	default:
		return node
	}
}

// usedVariablesClause collects all variables used in a clause.
// Corresponds to Python's used_variables_clause.
func usedVariablesClause(cube []*il.Literal) []*lg.Var {
	seen := make(map[string]bool)
	var result []*lg.Var
	for _, lit := range cube {
		fvs := lu.FreeVariablesList(lit.Atom)
		for _, v := range fvs {
			if !seen[v.Name] {
				seen[v.Name] = true
				result = append(result, v)
			}
		}
	}
	return result
}

// varToSkolem creates a Skolem constant for a variable.
// Corresponds to Python's var_to_skolem.
func varToSkolem(prefix string, v *lg.Var) *lg.Const {
	return lg.NewConst(prefix+v.Name, v.VSort)
}

// substituteClause applies a substitution to all literals in a clause.
// Corresponds to Python's substitute_clause.
func substituteClause(cube []*il.Literal, subs map[lg.Node]lg.Node) []*il.Literal {
	if len(subs) == 0 {
		return cube
	}
	result := make([]*il.Literal, len(cube))
	for i, lit := range cube {
		newAtom, err := lu.Substitute(lit.Atom, subs)
		if err != nil {
			result[i] = lit
		} else {
			result[i] = il.NewLiteral(lit.Polarity, newAtom)
		}
	}
	return result
}

// negateLiteral returns the formula representing the negation of a literal.
// Corresponds to Python's ~lit.
func negateLiteral(lit *il.Literal) lg.Node {
	if lit.Polarity == 0 {
		return lit.Atom // double negation
	}
	return &lg.Not{Body: lit.Atom}
}

// nodeSlice converts []lg.Node to a new copy.
func nodeSlice(nodes []lg.Node) []lg.Node {
	result := make([]lg.Node, len(nodes))
	copy(result, nodes)
	return result
}

// allMatch checks that all correlated positions match between two rows.
func allMatch(xr, yr []lg.Node, corr [][2]int) bool {
	for _, pair := range corr {
		if !xr[pair[0]].Equal(yr[pair[1]]) {
			return false
		}
	}
	return true
}

// getAtomArgs extracts the argument list from an atom node.
func getAtomArgs(atom lg.Node) []lg.Node {
	switch t := atom.(type) {
	case *lg.Apply:
		return t.Terms
	case *lg.Eq:
		return []lg.Node{t.T1, t.T2}
	default:
		return nil
	}
}

// csLitsToILLits converts concept-space literals to ivylogic literals.
// This bridges the webui.CSLiteral type to il.Literal.
func csLitsToILLits(csLits []*webui.CSLiteral) []*il.Literal {
	result := make([]*il.Literal, len(csLits))
	for i, csl := range csLits {
		atom := csAtomToNode(csl.Atom)
		result[i] = il.NewLiteral(csl.Polarity, atom)
	}
	return result
}

// csAtomToNode converts a concept-space atom to an Ivy logic node.
func csAtomToNode(atom *webui.CSAtom) lg.Node {
	if atom.RelName == "=" && len(atom.Args) == 2 {
		return &lg.Eq{
			T1: csTermToNode(atom.Args[0]),
			T2: csTermToNode(atom.Args[1]),
		}
	}
	if len(atom.Args) == 0 {
		return lg.NewConst(atom.RelName, lg.Boolean)
	}
	args := make([]lg.Node, len(atom.Args))
	sorts := make([]lg.Sort, len(atom.Args)+1)
	for i, a := range atom.Args {
		args[i] = csTermToNode(a)
		sorts[i] = args[i].NodeSort()
	}
	sorts[len(atom.Args)] = lg.Boolean
	funcSort, _ := lg.NewFunctionSort(sorts...)
	fn := lg.NewConst(atom.RelName, funcSort)
	return &lg.Apply{Func: fn, Terms: args}
}

// csTermToNode converts a concept-space term to an Ivy logic node.
func csTermToNode(t webui.CSTerm) lg.Node {
	if t.IsVariable {
		v, _ := lg.NewVar(t.Name, &lg.TopSort{Name: "alpha"})
		return v
	}
	return lg.NewConst(t.Name, &lg.TopSort{Name: "alpha"})
}
