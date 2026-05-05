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
	goivy "github.com/glycerine/ivy/goivy"

	"github.com/glycerine/ivy/goivy/webui"
)

// --- Domain interface for Alpha ---

// AlphaDomain represents the domain information needed by Alpha.
// In Python this is state.domain with concept_spaces and background_theory.
type AlphaDomain struct {
	ConceptSpaces    []ConceptSpaceEntry
	BackgroundTheory func(inScope map[string]bool) *goivy.Clauses
	AbstrPreds       []*goivy.Clauses
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
	Clauses    *goivy.Clauses
	Domain     *AlphaDomain
	InScope    map[string]bool
	TestBottom bool // Whether to UNSAT-check concrete state before abstraction. Python: test_bottom.
	Log        bool // Verbose logging of alpha operations. Python: log.
}

// --- Alpha function ---

// Alpha computes the abstract post-image of a state using concept spaces.
// Corresponds to Python's alpha(state).
func Alpha(state *AlphaState) {
	d := NewProgressiveDomain(state.Domain.ConceptSpaces, false, state.TestBottom, state.Log)
	var bgTheory *goivy.Clauses
	if state.Domain.BackgroundTheory != nil {
		bgTheory = state.Domain.BackgroundTheory(state.InScope)
	} else {
		bgTheory = goivy.TrueClauses(nil)
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
	testBottom     bool // from AlphaState.TestBottom
	log            bool // from AlphaState.Log
	slvr           *goivy.Solver
	z3solver       *goivy.Z3Solver
	cubeMemo       map[uint]*goivy.CubeMemoEntry // Z3 AST ID -> cached result
	inhabitedCubes map[string]bool               // Z3 expr ID -> inhabited
	z3Cubes        []goivy.Z3Expr                // prevent GC of Z3 cubes
	memo           map[string]webui.CSMemoEntry
	inferred       [][]goivy.Expr
	unsat          bool
	newSym         map[string]*goivy.Const
}

// NewProgressiveDomain creates a new ProgressiveDomain.
// Corresponds to Python's ProgressiveDomain.__init__.
func NewProgressiveDomain(cs []ConceptSpaceEntry, verbose bool, testBottom bool, log bool) *ProgressiveDomain {
	return &ProgressiveDomain{
		conceptSpaces: cs,
		verbose:       verbose,
		testBottom:    testBottom,
		log:           log,
	}
}

// AddConceptSpace appends a concept space entry.
// Corresponds to Python's add_concept_space.
func (pd *ProgressiveDomain) AddConceptSpace(atom *webui.CSAtom, space webui.CSNode) {
	pd.conceptSpaces = append(pd.conceptSpaces, ConceptSpaceEntry{Atom: atom, Space: space})
}

// cubeID computes a string identifier for a cube (list of literals).
// Corresponds to Python's cube_id.
func (pd *ProgressiveDomain) cubeID(cube []*goivy.LogicLiteral) (string, error) {
	z3cube, err := pd.slvr.CubeToZ3(cube)
	if err != nil {
		return "", err
	}
	pd.z3Cubes = append(pd.z3Cubes, z3cube) // prevent GC
	return z3cube.String(), nil
}

// inhabitedCube marks a cube as inhabited.
// Corresponds to Python's inhabited_cube.
func (pd *ProgressiveDomain) inhabitedCube(cube []*goivy.LogicLiteral, truth bool) {
	cube = canonizeClause(cube)
	id, err := pd.cubeID(cube)
	if err != nil {
		return
	}
	if _, exists := pd.inhabitedCubes[id]; !exists {
		if pd.log {
			fmt.Printf("inhabited: %v\n", cube)
		}
		pd.inhabitedCubes[id] = truth
	}
}

// inhabitedLit marks a single literal as inhabited.
// Corresponds to Python's inhabited_lit.
func (pd *ProgressiveDomain) inhabitedLit(lit *goivy.LogicLiteral) {
	pd.inhabitedCube([]*goivy.LogicLiteral{lit}, true)
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
func (pd *ProgressiveDomain) unfoldDefs(cube []*goivy.LogicLiteral) {
	// In Python: definition_instances(cube_to_formula(cube))
	// This is a placeholder; full implementation would compute definition
	// instances and add them to the solver.
	_ = cube
}

// testCube tests whether a cube is consistent with the solver state.
// Returns true if the cube is satisfiable (inhabited).
// Corresponds to Python's test_cube.
func (pd *ProgressiveDomain) testCube(cube []*goivy.LogicLiteral) bool {
	canonCube := canonizeClause(cube)
	if pd.log {
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
		if pd.log {
			strs := make([]string, len(cube))
			for i, c := range cube {
				strs[i] = c.String()
			}
			fmt.Printf("cached: %v\n", strs)
		}
		return val
	}
	if pd.log {
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
	subs := make(map[goivy.NodeKey]goivy.Expr, len(vs))
	for _, v := range vs {
		subs[goivy.Key(v)] = varToSkolem("__c", v)
	}
	scube := substituteClause(renamedCube, subs)

	// Unfold definitions
	pd.unfoldDefs(scube)

	// Check cube satisfiability
	res, err := pd.slvr.CheckCube(pd.z3solver, scube, pd.cubeMemo, false)
	if err != nil {
		return false
	}

	if res {
		if len(cube) <= 4 {
			pd.modelCheck()
		}
	} else {
		// Cube is unsat - infer negation
		negated := make([]goivy.Expr, len(cube))
		for i, lit := range cube {
			negated[i] = negateLiteral(lit)
		}
		pd.inferred = append(pd.inferred, negated)
		if pd.log {
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
	theory, backgroundTheory *goivy.Clauses,
	newSym map[string]*goivy.Const,
	toKeep []string,
) {
	pd.newSym = newSym
	pd.slvr = goivy.NewSolver(nil, nil)
	pd.z3solver = pd.slvr.NewZ3Solver()
	pd.cubeMemo = make(map[uint]*goivy.CubeMemoEntry)
	pd.inhabitedCubes = make(map[string]bool)
	pd.z3Cubes = nil
	pd.memo = make(map[string]webui.CSMemoEntry)

	if pd.log {
		fmt.Printf("concrete state: %s\n", theory)
		fmt.Printf("background: %s\n", backgroundTheory)
	}

	combined := goivy.AndClausesTyped(theory, backgroundTheory)
	if err := pd.slvr.AddClauses(pd.z3solver, combined); err != nil {
		pd.unsat = true
		return
	}

	if pd.testBottom {
		result := pd.z3solver.Check()
		pd.unsat = (result == goivy.Unsat)
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
func (pd *ProgressiveDomain) postStep(conceptSpaces []ConceptSpaceEntry) *goivy.Clauses {
	if pd.unsat {
		return goivy.FalseClauses(nil)
	}
	pd.inferred = nil

	for _, entry := range conceptSpaces {
		if pd.log {
			fmt.Printf("concept space: %s\n", entry.Atom)
		}

		// Create the test function that wraps testCube
		testFn := func(csLits []*webui.CSLiteral) bool {
			ilLits := csLitsToILLits(csLits)
			return pd.testCube(ilLits)
		}

		concepts := entry.Space.Enumerate(pd.memo, testFn)
		if pd.log {
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
	if pd.log {
		fmt.Printf("inferred: %v\n", res)
	}
	pd.inferred = nil

	// Convert inferred to Clauses
	fmlas := make([]goivy.Expr, len(res))
	for i, clause := range res {
		if len(clause) == 1 {
			fmlas[i] = clause[0]
		} else {
			fmlas[i] = &goivy.LogicOr{Terms: clause}
		}
	}
	return goivy.NewClauses(fmlas, nil, nil)
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
	theory, backgroundTheory *goivy.Clauses,
	newSym map[string]*goivy.Const,
	toKeep []string,
) *goivy.Clauses {
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
func VarCorr(terms1, terms2 []goivy.Expr) [][2]int {
	d := make(map[string]int)
	for i, t := range terms2 {
		if v, ok := t.(*goivy.LogicVariable); ok {
			d[v.Name] = i
		}
	}
	var result [][2]int
	for j, t := range terms1 {
		if v, ok := t.(*goivy.LogicVariable); ok {
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
func cutRow(row []goivy.Expr, goodCols []int) []goivy.Expr {
	result := make([]goivy.Expr, len(goodCols))
	for i, col := range goodCols {
		result[i] = row[col]
	}
	return result
}

// RelTable represents a relational table: a tuple of (variables, rows).
// Corresponds to Python's (v, rows) tuples.
type RelTable struct {
	Vars []goivy.Expr
	Rows [][]goivy.Expr
}

// compactTable removes redundant columns from a relation table.
// Corresponds to Python's compact_table.
func compactTable(tab *RelTable) *RelTable {
	memo := make(map[string]bool)
	var goodCols []int
	for i, t := range tab.Vars {
		if v, ok := t.(*goivy.LogicVariable); ok && firstSeen(memo, v.Name) {
			goodCols = append(goodCols, i)
		}
	}
	newVars := cutRow(tab.Vars, goodCols)
	newRows := make([][]goivy.Expr, len(tab.Rows))
	for i, row := range tab.Rows {
		newRows[i] = cutRow(row, goodCols)
	}
	return &RelTable{Vars: newVars, Rows: newRows}
}

// --- RelAlg1 ---

// RelAlg1 implements relational algebra using model instances.
// Corresponds to Python's RelAlg1 class.
type RelAlg1 struct {
	Slvr   *goivy.Solver
	Z3Slvr *goivy.Solver
	Model  *goivy.HerbrandModel
	Parent *ProgressiveDomain
}

// NewRelAlg1 creates a new RelAlg1.
// Corresponds to Python's RelAlg1.__init__.
func NewRelAlg1(slvr *goivy.Solver, z3slvr *goivy.Solver, parent *ProgressiveDomain) *RelAlg1 {
	return &RelAlg1{
		Slvr:   slvr,
		Z3Slvr: z3slvr,
		Parent: parent,
	}
}

// Prim evaluates a primitive literal against the model.
// Returns a RelTable of matching instances.
// Corresponds to Python's RelAlg1.prim.
func (ra *RelAlg1) Prim(lit *goivy.LogicLiteral) *RelTable {
	// Get model instances
	app, ok := lit.Atom.(*goivy.Apply)
	if !ok {
		return &RelTable{}
	}

	// Check the literal against the model
	posLit := goivy.NewLiteral(lit.Polarity, lit.Atom)
	_, rows := ra.Model.Check(posLit)

	// Build result table
	resultRows := make([][]goivy.Expr, len(rows))
	for i, row := range rows {
		resultRow := make([]goivy.Expr, len(row))
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
	var rows [][]goivy.Expr
	if len(corr) > 0 {
		xc, yc := corr[0][0], corr[0][1]
		index := make(map[string][]int) // variable rep -> row indices in y
		for i, ry := range y.Rows {
			key := ry[yc].String()
			index[key] = append(index[key], i)
		}
		for _, xr := range x.Rows {
			key := xr[xc].String()
			// Match Python defaultdict auto-vivification
			if _, ok := index[key]; !ok {
				index[key] = nil
			}
			for _, yi := range index[key] {
				yr := y.Rows[yi]
				if allMatch(xr, yr, corr) {
					combined := make([]goivy.Expr, 0, len(xr)+len(yr))
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
					combined := make([]goivy.Expr, 0, len(xr)+len(yr))
					combined = append(combined, xr...)
					combined = append(combined, yr...)
					rows = append(rows, combined)
				}
			}
		}
	}
	combinedVars := make([]goivy.Expr, 0, len(x.Vars)+len(y.Vars))
	combinedVars = append(combinedVars, x.Vars...)
	combinedVars = append(combinedVars, y.Vars...)
	return compactTable(&RelTable{Vars: combinedVars, Rows: rows})
}

// Subst applies a substitution to a table's variable names.
// Corresponds to Python's RelAlg1.subst.
func (ra *RelAlg1) Subst(tab *RelTable, subst map[string]goivy.Expr) *RelTable {
	newVars := make([]goivy.Expr, len(tab.Vars))
	for i, v := range tab.Vars {
		if vv, ok := v.(*goivy.LogicVariable); ok {
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
	Slvr       *goivy.Solver
	Z3Slvr     *goivy.Z3Solver
	TempSolver *goivy.Z3Solver
	Model      *goivy.HerbrandModel
	Parent     *ProgressiveDomain
	Numbering  map[string]int
	NextNumber int
	PrimCache  map[string][]goivy.Z3Expr
	PrimList   []goivy.Z3Expr // prevent GC
	NewSym     map[string]*goivy.Const
	Hm         *goivy.HerbrandModel
}

// NewRelAlg2 creates a new RelAlg2.
// Corresponds to Python's RelAlg2.__init__.
func NewRelAlg2(
	slvr *goivy.Solver,
	z3slvr *goivy.Z3Solver,
	newSym map[string]*goivy.Const,
	parent *ProgressiveDomain,
) *RelAlg2 {
	return &RelAlg2{
		Slvr:       slvr,
		Z3Slvr:     z3slvr,
		TempSolver: slvr.NewZ3Solver(),
		Parent:     parent,
		Numbering:  make(map[string]int),
		NextNumber: 0,
		PrimCache:  make(map[string][]goivy.Z3Expr),
		NewSym:     newSym,
	}
}

// IsSat checks if a formula is satisfiable in the temp solver.
// Corresponds to Python's is_sat.
func (ra *RelAlg2) IsSat(f goivy.Z3Expr) bool {
	ra.TempSolver.Push()
	defer ra.TempSolver.Pop()
	ra.TempSolver.Assert(f)
	result := ra.TempSolver.Check()
	return result != goivy.Unsat
}

// Prim evaluates a primitive literal.
// Returns a list of Z3 cube expressions.
// Corresponds to Python's RelAlg2.prim.
func (ra *RelAlg2) Prim(lit *goivy.LogicLiteral) []goivy.Z3Expr {
	z3lit, err := ra.Slvr.LiteralToZ3(lit)
	if err != nil {
		return nil
	}
	id := z3lit.String()
	if cached, ok := ra.PrimCache[id]; ok {
		return cached
	}

	var cubes []goivy.Z3Expr

	// Rename literal and get ground instances from Herbrand model
	renamedLit := renameLit(lit, ra.NewSym)
	if ra.Hm != nil {
		vs, rows := ra.Hm.Check(renamedLit)
		ctx := ra.Slvr.Context()
		for _, row := range rows {
			eqs := make([]goivy.Z3Expr, 0, len(vs))
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
func (ra *RelAlg2) Top() []goivy.Z3Expr {
	ctx := ra.Slvr.Context()
	return []goivy.Z3Expr{ctx.BoolVal(true)}
}

// Prod computes the product of two cube lists.
// Corresponds to Python's RelAlg2.prod.
func (ra *RelAlg2) Prod(x, y []goivy.Z3Expr) []goivy.Z3Expr {
	ctx := ra.Slvr.Context()
	var cubes []goivy.Z3Expr
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
func (ra *RelAlg2) Subst(tab []goivy.Z3Expr, subst map[string]goivy.Expr) []goivy.Z3Expr {
	// Build Z3-level substitution
	ctx := ra.Slvr.Context()
	var fromExprs, toExprs []goivy.Z3Expr
	for name, node := range subst {
		c := goivy.NewConst(name, node.NodeSort())
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
	result := make([]goivy.Z3Expr, len(tab))
	for i, expr := range tab {
		result[i] = ctx.Substitute(expr, fromExprs, toExprs)
	}
	return result
}

// Empty returns true if the cube list is empty.
// Corresponds to Python's RelAlg2.empty.
func (ra *RelAlg2) Empty(tab []goivy.Z3Expr) bool {
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
	slvr *goivy.Solver,
	z3slvr *goivy.Z3Solver,
	newSym map[string]*goivy.Const,
	parent *ProgressiveDomain,
) *RelAlg3 {
	return &RelAlg3{
		RelAlg2: NewRelAlg2(slvr, z3slvr, newSym, parent),
	}
}

// Prim evaluates a primitive literal using HerbrandModel.Check.
// Corresponds to Python's RelAlg3.prim.
func (ra *RelAlg3) Prim(lit *goivy.LogicLiteral) []goivy.Z3Expr {
	if ra.Parent.log {
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
	var cubes []goivy.Z3Expr

	if ra.Hm != nil {
		vs, rows := ra.Hm.Check(renamedLit)
		ctx := ra.Slvr.Context()
		for _, row := range rows {
			eqs := make([]goivy.Z3Expr, 0, len(vs))
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
	slvr := goivy.NewSolver(nil, nil)
	z3slvr := slvr.NewZ3Solver()

	var bgTheory *goivy.Clauses
	if state.Domain.BackgroundTheory != nil {
		bgTheory = state.Domain.BackgroundTheory(nil)
	} else {
		bgTheory = goivy.TrueClauses(nil)
	}

	combined := goivy.AndClausesTyped(state.Clauses, bgTheory)
	if err := slvr.AddClauses(z3slvr, combined); err != nil {
		return
	}

	res := goivy.TrueClauses(nil)
	for _, pred := range state.Domain.AbstrPreds {
		z3slvr.Push()
		dual := goivy.NegateClauses(pred)
		if err := slvr.AddClauses(z3slvr, dual); err != nil {
			z3slvr.Pop()
			continue
		}
		cr := z3slvr.Check()
		if state.Log {
			fmt.Printf("predicate: %s result %v\n", pred, cr)
		}
		if cr == goivy.Unsat {
			res = goivy.AndClausesTyped(res, pred)
		}
		z3slvr.Pop()
	}
	state.Clauses = res
}

// --- Helper functions ---

// canonizeClause sorts and deduplicates literals in a clause.
// Corresponds to Python's canonize_clause.
func canonizeClause(cube []*goivy.LogicLiteral) []*goivy.LogicLiteral {
	if len(cube) == 0 {
		return cube
	}
	// Deduplicate by string representation
	seen := make(map[string]bool)
	var result []*goivy.LogicLiteral
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
func renameClause(cube []*goivy.LogicLiteral, newSym map[string]*goivy.Const) []*goivy.LogicLiteral {
	if len(newSym) == 0 {
		return cube
	}
	result := make([]*goivy.LogicLiteral, len(cube))
	for i, lit := range cube {
		result[i] = renameLit(lit, newSym)
	}
	return result
}

// renameLit renames symbols in a literal using the newSym map.
// Corresponds to Python's rename_lit.
func renameLit(lit *goivy.LogicLiteral, newSym map[string]*goivy.Const) *goivy.LogicLiteral {
	if len(newSym) == 0 {
		return lit
	}
	newAtom := renameNode(lit.Atom, newSym)
	return goivy.NewLiteral(lit.Polarity, newAtom)
}

// renameNode renames constant symbols in a node using the newSym map.
func renameNode(node goivy.Expr, newSym map[string]*goivy.Const) goivy.Expr {
	if len(newSym) == 0 {
		return node
	}
	switch t := node.(type) {
	case *goivy.Const:
		if repl, ok := newSym[t.Name]; ok {
			return repl
		}
		return t
	case *goivy.Apply:
		newFunc := renameNode(t.Func, newSym)
		newArgs := make([]goivy.Expr, len(t.Terms))
		for i, arg := range t.Terms {
			newArgs[i] = renameNode(arg, newSym)
		}
		return goivy.MustApply(newFunc, newArgs...)
	case *goivy.Eq:
		return &goivy.Eq{T1: renameNode(t.T1, newSym), T2: renameNode(t.T2, newSym)}
	case *goivy.LogicNot:
		return &goivy.LogicNot{Body: renameNode(t.Body, newSym)}
	default:
		return node
	}
}

// usedVariablesClause collects all variables used in a clause.
// Corresponds to Python's used_variables_clause.
func usedVariablesClause(cube []*goivy.LogicLiteral) []*goivy.LogicVariable {
	seen := make(map[string]bool)
	var result []*goivy.LogicVariable
	for _, lit := range cube {
		fvs := goivy.FreeVariablesList(lit.Atom)
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
func varToSkolem(prefix string, v *goivy.LogicVariable) *goivy.Const {
	return goivy.NewConst(prefix+v.Name, v.VSort)
}

// substituteClause applies a substitution to all literals in a clause.
// Corresponds to Python's substitute_clause.
func substituteClause(cube []*goivy.LogicLiteral, subs map[goivy.NodeKey]goivy.Expr) []*goivy.LogicLiteral {
	if len(subs) == 0 {
		return cube
	}
	result := make([]*goivy.LogicLiteral, len(cube))
	for i, lit := range cube {
		newAtom, err := goivy.Substitute(lit.Atom, subs)
		if err != nil {
			result[i] = lit
		} else {
			result[i] = goivy.NewLiteral(lit.Polarity, newAtom)
		}
	}
	return result
}

// negateLiteral returns the formula representing the negation of a literal.
// Corresponds to Python's ~lit.
func negateLiteral(lit *goivy.LogicLiteral) goivy.Expr {
	if lit.Polarity == 0 {
		return lit.Atom // double negation
	}
	return &goivy.LogicNot{Body: lit.Atom}
}

// nodeSlice converts []lg.Expr to a new copy.
func nodeSlice(nodes []goivy.Expr) []goivy.Expr {
	result := make([]goivy.Expr, len(nodes))
	copy(result, nodes)
	return result
}

// allMatch checks that all correlated positions match between two rows.
func allMatch(xr, yr []goivy.Expr, corr [][2]int) bool {
	for _, pair := range corr {
		if !xr[pair[0]].Equal(yr[pair[1]]) {
			return false
		}
	}
	return true
}

// getAtomArgs extracts the argument list from an atom node.
func getAtomArgs(atom goivy.Expr) []goivy.Expr {
	switch t := atom.(type) {
	case *goivy.Apply:
		return t.Terms
	case *goivy.Eq:
		return []goivy.Expr{t.T1, t.T2}
	default:
		return nil
	}
}

// csLitsToILLits converts concept-space literals to ivylogic literals.
// This bridges the webui.CSLiteral type to il.Literal.
func csLitsToILLits(csLits []*webui.CSLiteral) []*goivy.LogicLiteral {
	result := make([]*goivy.LogicLiteral, len(csLits))
	for i, csl := range csLits {
		atom := csAtomToNode(csl.Atom)
		result[i] = goivy.NewLiteral(csl.Polarity, atom)
	}
	return result
}

// csAtomToNode converts a concept-space atom to an Ivy logic node.
func csAtomToNode(atom *webui.CSAtom) goivy.Expr {
	if atom.RelName == "=" && len(atom.Args) == 2 {
		return &goivy.Eq{
			T1: csTermToNode(atom.Args[0]),
			T2: csTermToNode(atom.Args[1]),
		}
	}
	if len(atom.Args) == 0 {
		return goivy.NewConst(atom.RelName, goivy.Boolean)
	}
	args := make([]goivy.Expr, len(atom.Args))
	sorts := make([]goivy.Sort, len(atom.Args)+1)
	for i, a := range atom.Args {
		args[i] = csTermToNode(a)
		sorts[i] = args[i].NodeSort()
	}
	sorts[len(atom.Args)] = goivy.Boolean
	funcSort, _ := goivy.NewFunctionSort(sorts...)
	fn := goivy.NewConst(atom.RelName, funcSort)
	return goivy.MustApply(fn, args...)
}

// csTermToNode converts a concept-space term to an Ivy logic node.
func csTermToNode(t webui.CSTerm) goivy.Expr {
	if t.IsVariable {
		v, _ := goivy.NewVariable(t.Name, &goivy.TopSort{Name: "alpha"})
		return v
	}
	return goivy.NewConst(t.Name, &goivy.TopSort{Name: "alpha"})
}
