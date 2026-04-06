// Package fragment implements decidable fragment checking
// for Ivy verification conditions.
// This is a port of Python's ivy_fragment.py.
//
// It checks whether VCs are in the FEU (Finite Essentially
// Uninterpreted) fragment, which guarantees that Z3 can
// decide them. The check builds a stratification graph
// as described in:
//
// Yeting Ge and Leonardo de Moura, "Complete instantiation
// for quantified formulas in Satisfiability Modulo Theories"
//
// If the stratification graph is cyclic, the VC may
// generate an infinite sequence of instantiations,
// and an error is raised.
package fragment

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/module"
	thy "github.com/glycerine/ivy/goivy/theory"
	uf "github.com/glycerine/ivy/goivy/unionfind"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// FragmentError is raised when a VC is not in the FAU fragment.
type FragmentError struct {
	Message string
}

func (e *FragmentError) Error() string {
	return e.Message
}

// --- Stratification graph types ---

// stratEntry holds metadata associated with a stratification key, for error reporting.
type stratEntry struct {
	sym    *lg.Const    // non-nil for appKey entries
	idx    int          // argument index for appKey entries
	v      *lg.Variable // non-nil for varKey entries
	isSort bool
}

func varKey(v *lg.Variable) lg.NodeKey {
	return lg.NodeKey("v:" + string(lg.Key(v)))
}

func appKey(sym *lg.Const, idx int) lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("a:%s:%d", lg.Key(sym), idx))
}

// eqExprKey returns a strat_map key for an equality node keyed on the expression,
// matching Python: il.Symbol('=', fmla.args[0]) which creates Const('=', <expression>)
// and uses it as a dict key via recstruct hashing on (name, sort=expression).
// Uses (StratNode ...) — a unique s-expression tag not used by any AST type —
// to avoid collisions with real Const.Sexp() entries that use (Symbol ...).
func eqExprKey(expr lg.Expr) lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(StratNode eq:%v)", lg.Key(expr)))
}

// arc represents a directed edge in the stratification graph.
type arc struct {
	from   *uf.UFNode
	to     *uf.UFNode
	fmla   lg.Expr
	lineno int
	argIdx int  // -1 if not applicable
	hasIdx bool // true if argIdx is valid
}

// --- Fragment checker state ---

// checker holds the state for a single fragment check.
type checker struct {
	sig    *il.Sig
	interp map[string]interface{} // sort interpretations

	universallyQuantifiedVars map[varID]*lg.Variable // var → lineno origin info
	universalVarLineno        map[varID]int          // var → lineno

	stratMap  map[lg.NodeKey]*uf.UFNode // maps node key to UFNode
	stratInfo map[lg.NodeKey]stratEntry // metadata for error reporting
	arcs      []arc

	// Macro maps
	macroMap      map[lg.NodeKey]macroDef       // symbol key → (definition, labeled formula)
	macroValueMap map[lg.NodeKey]mapFmlaRes     // symbol key → memoized result
	macroVarMap   map[varID]*uf.UFNode          // macro param var → strat node
	macroDepMap   map[varID]map[*uf.UFNode]bool // macro param → dep nodes

	// Skolem map
	skolemMap map[varID]skolemEntry // existential var → (formula, origin)

	// Variable uniquifier (for undo in error messages)
	varUniq *il.VariableUniqifier
}

type varID struct {
	name string
	sort string
}

func makeVarID(v *lg.Variable) varID {
	return varID{name: v.Name, sort: v.VSort.String()}
}

type macroDef struct {
	def *il.Definition
	lf  *ast.LabeledFormula
}

type mapFmlaRes struct {
	node *uf.UFNode
	uvs  map[*uf.UFNode]bool
}

type skolemEntry struct {
	fmla lg.Expr
	ast  ast.Node
}

type SomeString string

func (s SomeString) Canon() iu.Canonical {
	return iu.Canonical(s)
}

// fmlaPair is a (formula, source) pair used throughout the checker.
type fmlaPair struct {
	fmla   lg.Expr
	source ast.Node // *ast.LabeledFormula or actions.Action
	lineno int
}

func newChecker(sig *il.Sig, interp map[string]interface{}) *checker {
	return &checker{
		sig:                       sig,
		interp:                    interp,
		universallyQuantifiedVars: make(map[varID]*lg.Variable),
		universalVarLineno:        make(map[varID]int),
		stratMap:                  make(map[lg.NodeKey]*uf.UFNode),
		stratInfo:                 make(map[lg.NodeKey]stratEntry),
		arcs:                      nil,
		macroMap:                  make(map[lg.NodeKey]macroDef),
		macroValueMap:             make(map[lg.NodeKey]mapFmlaRes),
		macroVarMap:               make(map[varID]*uf.UFNode),
		macroDepMap:               make(map[varID]map[*uf.UFNode]bool),
		skolemMap:                 make(map[varID]skolemEntry),
	}
}

// getStratNode gets or creates a UFNode for the given key.
func (c *checker) getStratNode(key lg.NodeKey) *uf.UFNode {
	if n, ok := c.stratMap[key]; ok {
		return n
	}
	n := uf.NewUFNode()
	c.stratMap[key] = n
	return n
}

// getStratNodeWith gets or creates a UFNode and stores metadata for error reporting.
func (c *checker) getStratNodeWith(key lg.NodeKey, entry stratEntry) *uf.UFNode {
	n := c.getStratNode(key)
	if _, exists := c.stratInfo[key]; !exists {
		c.stratInfo[key] = entry
	}
	return n
}

// isUnivVar checks if a variable is in the universally quantified set.
func (c *checker) isUnivVar(v *lg.Variable) bool {
	_, ok := c.universallyQuantifiedVars[makeVarID(v)]
	return ok
}

// getUnivNode gets the strat_map node for a universally quantified variable.
func (c *checker) getUnivNode(v *lg.Variable) *uf.UFNode {
	key := varKey(v)
	n := c.getStratNodeWith(key, stratEntry{v: v})
	n.Var = v
	return n
}

// --- mapFmla: build stratification graph ---

// mapFmla adds all subterms of fmla to the stratification graph.
// Returns (node, uvs) where node is the S_v if fmla is a universal variable,
// and uvs is the set of universal variable nodes occurring *under* the formula.
func (c *checker) mapFmla(lineno int, fmla lg.Expr, pol int) (*uf.UFNode, map[*uf.UFNode]bool) {
	if il.IsBinder(fmla) {
		body := il.BinderBody(fmla)
		if body != nil {
			return c.mapFmla(lineno, body, pol)
		}
		return nil, make(map[*uf.UFNode]bool)
	}

	if v, ok := fmla.(*lg.Variable); ok {
		vid := makeVarID(v)
		if c.isUnivVar(v) {
			node := c.getUnivNode(v)
			return node, make(map[*uf.UFNode]bool)
		}
		// Check macro maps
		node := c.macroVarMap[vid]
		deps := make(map[*uf.UFNode]bool)
		if d, ok := c.macroDepMap[vid]; ok {
			for k, v := range d {
				deps[k] = v
			}
		}
		return node, deps
	}

	args := il.NodeArgs(fmla)
	type argRes struct {
		node *uf.UFNode
		uvs  map[*uf.UFNode]bool
	}
	reses := make([]argRes, len(args))
	for i, arg := range args {
		n, uvs := c.mapFmla(lineno, arg, il.Polar(fmla, i, pol))
		reses[i] = argRes{n, uvs}
	}

	// Compute all_uvs: union of all uvs + all non-nil nodes
	allUvs := make(map[*uf.UFNode]bool)
	for _, r := range reses {
		for k := range r.uvs {
			allUvs[k] = true
		}
		if r.node != nil {
			allUvs[r.node] = true
		}
	}

	// Handle equality
	if il.IsEq(fmla) {
		eq := fmla.(*lg.Eq)
		sort := eq.T1.NodeSort()
		if !il.IsInterpretedSort(c.sig, sort) {
			sSigma := c.getStratNodeWith(eqExprKey(eq.T1), stratEntry{sym: lg.NewConst("=", sort), isSort: true})
			for i, r := range reses {
				if r.node != nil {
					uf.Unify(r.node, sSigma)
				}
				for v := range reses[i].uvs {
					c.arcs = append(c.arcs, arc{from: v, to: sSigma, fmla: fmla, lineno: lineno, argIdx: -1})
				}
			}
		} else {
			nodes := make([]*uf.UFNode, len(reses))
			uvss := make([]map[*uf.UFNode]bool, len(reses))
			for i, r := range reses {
				nodes[i] = r.node
				uvss[i] = r.uvs
			}
			c.checkInterpreted(fmla, nodes, uvss, lineno, pol)
		}
		return nil, allUvs
	}

	// Handle ite
	if il.IsIte(fmla) {
		if len(reses) >= 3 {
			if reses[1].node != nil && reses[2].node != nil {
				uf.Unify(reses[1].node, reses[2].node)
			}
			var resultNode *uf.UFNode
			if reses[1].node != nil {
				resultNode = reses[1].node
			} else {
				resultNode = reses[2].node
			}
			return resultNode, allUvs
		}
	}

	// Handle application
	if il.IsApp(fmla) {
		rep := il.GetAppRep(fmla)
		if rep != nil {
			if !il.IsInterpretedSymbol(c.sig, rep) {
				// Check macro maps
				repKey := lg.Key(rep)
				if res, ok := c.macroValueMap[repKey]; ok {
					return res.node, res.uvs
				}
				if md, ok := c.macroMap[repKey]; ok {
					resNode, resUvs := c.mapFmla(md.lf.Lineno, md.def.Rhs, -1)
					c.macroValueMap[repKey] = mapFmlaRes{node: resNode, uvs: resUvs}
					return resNode, resUvs
				}
				// Regular function application
				for i, r := range reses {
					anode := c.getStratNodeWith(appKey(rep, i), stratEntry{sym: rep, idx: i})
					if r.node != nil {
						uf.Unify(anode, r.node)
					}
					for v := range reses[i].uvs {
						c.arcs = append(c.arcs, arc{from: v, to: anode, fmla: fmla, lineno: lineno, argIdx: i, hasIdx: true})
					}
				}
			} else {
				nodes := make([]*uf.UFNode, len(reses))
				uvss := make([]map[*uf.UFNode]bool, len(reses))
				for i, r := range reses {
					nodes[i] = r.node
					uvss[i] = r.uvs
				}
				c.checkInterpreted(fmla, nodes, uvss, lineno, pol)
			}
		}
		return nil, allUvs
	}

	return nil, allUvs
}

// checkInterpreted checks that an interpreted symbol application satisfies the
// FAU arithmetic literal conditions.
func (c *checker) checkInterpreted(app lg.Expr, nodes []*uf.UFNode, uvs []map[*uf.UFNode]bool, lineno int, pol int) {
	for idx := range nodes {
		if nodes[idx] != nil {
			if !c.isArithmeticLiteral(app, idx, nodes, uvs, pol) {
				c.reportInterpOverVar(app, lineno, nodes[idx])
			}
		}
	}
}

// isArithmeticLiteral checks if an interpreted symbol application is an
// arithmetic literal (X = t, X < Y, X < t, t < X where t is ground).
func (c *checker) isArithmeticLiteral(app lg.Expr, pos int, nodes []*uf.UFNode, uvs []map[*uf.UFNode]bool, pol int) bool {
	rep := il.GetAppRep(app)
	if rep == nil {
		return false
	}
	args := il.NodeArgs(app)
	if len(args) < 2 {
		return false
	}

	isIneqOrEq := il.IsInequalitySymbol(rep.Name) || il.IsEq(app)
	if !isIneqOrEq {
		return false
	}

	sort := args[0].NodeSort()
	if !thy.HasIntegerInterp(sort, c.flatInterp()) {
		return false
	}

	if il.IsStrictInequalitySymbol(rep.Name, pol) {
		otherIdx := 1 - pos
		if otherIdx >= 0 && otherIdx < len(nodes) && nodes[otherIdx] != nil {
			uf.Unify(nodes[0], nodes[1])
			return true
		}
	}

	// If app is an integer theory literal and the other argument is ground, OK
	otherIdx := 1 - pos
	if otherIdx >= 0 && otherIdx < len(nodes) {
		if nodes[otherIdx] == nil && len(uvs[otherIdx]) == 0 {
			return true
		}
	}

	return false
}

// flatInterp converts the sig's interp map to the format expected by theory.HasIntegerInterp.
func (c *checker) flatInterp() map[string]interface{} {
	return c.interp
}

// --- Macro handling ---

// createMacroMaps sets up macro_map, macro_var_map, macro_dep_map, and macro_value_map.
func (c *checker) createMacroMaps(assumes, asserts []fmlaPair, macros []fmlaPair) {
	// Build macro_map
	for _, pair := range macros {
		if def, ok := pair.fmla.(*il.Definition); ok {
			defining := def.Defines()
			if defining != nil {
				if cst, ok := defining.(*lg.Const); ok {
					c.macroMap[lg.Key(cst)] = macroDef{
						def: def,
						lf:  pair.source.(*ast.LabeledFormula),
					}
				}
			}
		}
	}

	// Propagate variable maps through macro calls
	allPairs := make([]fmlaPair, 0, len(assumes)+len(asserts)+len(macros))
	allPairs = append(allPairs, assumes...)
	allPairs = append(allPairs, asserts...)
	// Add macros in reverse order
	for i := len(macros) - 1; i >= 0; i-- {
		allPairs = append(allPairs, macros[i])
	}

	for _, pair := range allPairs {
		for _, app := range il.AppsAst(pair.fmla) {
			rep := il.GetAppRep(app)
			if rep == nil {
				continue
			}
			md, isMacro := c.macroMap[lg.Key(rep)]
			if !isMacro {
				continue
			}

			// Get macro formal parameters
			lhsArgs := il.NodeArgs(md.def.Lhs)
			appArgs := il.NodeArgs(app)

			for i := 0; i < len(appArgs) && i < len(lhsArgs); i++ {
				w, wIsVar := lhsArgs[i].(*lg.Variable)
				if !wIsVar {
					continue
				}
				wid := makeVarID(w)

				v, vIsVar := appArgs[i].(*lg.Variable)
				if vIsVar {
					vid := makeVarID(v)
					if c.isUnivVar(v) {
						node := c.getUnivNode(v)
						c.varMapAdd(wid, node)
					}
					if mvNode, ok := c.macroVarMap[vid]; ok {
						c.varMapAdd(wid, mvNode)
					}
					if deps, ok := c.macroDepMap[vid]; ok {
						if c.macroDepMap[wid] == nil {
							c.macroDepMap[wid] = make(map[*uf.UFNode]bool)
						}
						for k, v := range deps {
							c.macroDepMap[wid][k] = v
						}
					}
				} else {
					// Non-variable argument: free variables contribute to deps.
					// Python uses ilu.used_variables_ast (ivy_logic_utils.variables_ast)
					// which collects free variables (excludes bound vars from binders).
					// Note: ivy_ast.py has a different used_variables_ast that collects ALL
					// variables, but ivy_fragment.py imports from ivy_logic_utils, not ivy_ast.
					fvs := lu.FreeVariables(appArgs[i])
					for _, uNode := range fvs {
						u := uNode.(*lg.Variable)
						uid := makeVarID(u)
						if c.isUnivVar(u) {
							node := c.getUnivNode(u)
							if c.macroDepMap[wid] == nil {
								c.macroDepMap[wid] = make(map[*uf.UFNode]bool)
							}
							c.macroDepMap[wid][node] = true
						}
						if mvNode, ok := c.macroVarMap[uid]; ok {
							if c.macroDepMap[wid] == nil {
								c.macroDepMap[wid] = make(map[*uf.UFNode]bool)
							}
							c.macroDepMap[wid][mvNode] = true
						}
						if deps, ok := c.macroDepMap[uid]; ok {
							if c.macroDepMap[wid] == nil {
								c.macroDepMap[wid] = make(map[*uf.UFNode]bool)
							}
							for k, v := range deps {
								c.macroDepMap[wid][k] = v
							}
						}
					}
				}
			}
		}
	}
}

// varMapAdd adds or unifies a macro variable mapping.
func (c *checker) varMapAdd(wid varID, vn *uf.UFNode) {
	if existing, ok := c.macroVarMap[wid]; ok {
		uf.Unify(existing, vn)
	} else {
		c.macroVarMap[wid] = vn
	}
}

// --- Skolem handling ---

// makeSkolems simulates Skolem functions for AE alternations.
func (c *checker) makeSkolems(fmla lg.Expr, source ast.Node, pol bool, univs []*lg.Variable) {
	switch t := fmla.(type) {
	case *lg.Not:
		c.makeSkolems(t.Body, source, !pol, univs)
		return
	case *lg.Implies:
		c.makeSkolems(t.T1, source, !pol, univs)
		c.makeSkolems(t.T2, source, pol, univs)
		return
	}

	isE := il.IsExists(fmla)
	isA := il.IsForall(fmla)

	if (isE && pol) || (isA && !pol) {
		fvs := lu.FreeVariables(fmla)
		for _, u := range univs {
			if _, ok := fvs[lg.Key(u)]; ok {
				qvars := il.QuantifierVars(fmla)
				for _, e := range qvars {
					eid := makeVarID(e)
					c.skolemMap[eid] = skolemEntry{fmla: fmla, ast: source}
					uNode := c.getUnivNode(u)
					if c.macroDepMap[eid] == nil {
						c.macroDepMap[eid] = make(map[*uf.UFNode]bool)
					}
					c.macroDepMap[eid][uNode] = true
				}
			}
		}
	}

	if (isE && !pol) || (isA && pol) {
		qvars := il.QuantifierVars(fmla)
		newUnivs := make([]*lg.Variable, len(univs)+len(qvars))
		copy(newUnivs, univs)
		copy(newUnivs[len(univs):], qvars)
		body := il.BinderBody(fmla)
		if body != nil {
			c.makeSkolems(body, source, pol, newUnivs)
		}
	}

	for _, arg := range il.NodeArgs(fmla) {
		c.makeSkolems(arg, source, pol, univs)
	}

	if _, ok := fmla.(*lg.Ite); ok {
		args := il.NodeArgs(fmla)
		if len(args) > 0 {
			c.makeSkolems(args[0], source, !pol, univs)
		}
	}

	if _, ok := fmla.(*lg.Iff); ok {
		args := il.NodeArgs(fmla)
		if len(args) >= 2 {
			c.makeSkolems(args[0], source, !pol, univs)
			c.makeSkolems(args[1], source, !pol, univs)
		}
	}

	// Handle boolean equality like iff
	if eq, ok := fmla.(*lg.Eq); ok {
		if lg.SortEqual(eq.T1.NodeSort(), lg.Boolean) {
			c.makeSkolems(eq.T1, source, !pol, univs)
			c.makeSkolems(eq.T2, source, !pol, univs)
		}
	}
}

// --- Stratification graph construction ---

// createStratMap builds the full stratification graph.
func (c *checker) createStratMap(assumes, asserts, macros []fmlaPair) {
	// Gather all formulas
	allFmlas := make([]fmlaPair, 0, len(assumes)+len(asserts)+len(macros))
	for _, a := range assumes {
		closed := il.CloseFormula(a.fmla)
		allFmlas = append(allFmlas, fmlaPair{fmla: closed, source: a.source, lineno: a.lineno})
	}
	for _, a := range asserts {
		negated := &lg.Not{Body: a.fmla}
		allFmlas = append(allFmlas, fmlaPair{fmla: negated, source: a.source, lineno: a.lineno})
	}
	allFmlas = append(allFmlas, macros...)

	// Get universally quantified variables
	for _, fp := range allFmlas {
		uvars := il.UniversalVariables([]lg.Expr{fp.fmla})
		for _, v := range uvars {
			vid := makeVarID(v)
			if il.IsUninterpretedSort(c.sig, v.VSort) ||
				il.HasInfiniteInterpretation(c.sig, v.VSort) {
				c.universallyQuantifiedVars[vid] = v
				c.universalVarLineno[vid] = fp.lineno
			}
		}
	}

	// Create macro maps
	c.createMacroMaps(assumes, asserts, macros)

	// Simulate Skolem functions
	for _, fp := range allFmlas {
		c.makeSkolems(fp.fmla, fp.source, true, nil)
	}

	// Build graph by mapping all assumes and asserts
	for _, pair := range append(assumes, asserts...) {
		c.mapFmla(pair.lineno, pair.fmla, 0)
	}
}

// --- Cycle detection and error reporting ---

func (c *checker) reportFEUError(text string) error {
	return &FragmentError{
		Message: "The verification condition is not in the fragment FAU.\n\n" + text,
	}
}

func (c *checker) getNodeSort(n *uf.UFNode) lg.Sort {
	for key, node := range c.stratMap {
		if node == n {
			if info, ok := c.stratInfo[key]; ok {
				if info.sym != nil && info.idx >= 0 && !info.isSort {
					// appKey
					dom := il.SortDomain(info.sym.CSort)
					if info.idx < len(dom) {
						return dom[info.idx]
					}
				}
				if info.v != nil {
					return info.v.VSort
				}
			}
		}
	}
	return lg.TopS
}

func (c *checker) reportArc(a arc) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n%d: %s", a.lineno, a.fmla)
	if a.hasIdx {
		args := il.NodeArgs(a.fmla)
		if a.argIdx >= 0 && a.argIdx < len(args) {
			term := args[a.argIdx]
			fmt.Fprintf(&b, "\n    (position %d is a function from %s to %s)",
				a.argIdx, c.getNodeSort(a.from), term.NodeSort())
			// Divergence 11 fix: check skolemMap for skolem origin info,
			// matching Python report_arc lines 418-420.
			if v, ok := term.(*lg.Variable); ok {
				vid := makeVarID(v)
				if se, found := c.skolemMap[vid]; found {
					fmt.Fprintf(&b, "\n    %sskolem function defined by:\n         %s",
						se.ast.GetLineno(), se.fmla)
				}
			}
		}
	}
	return b.String()
}

func (c *checker) reportCycle(cycle []arc) error {
	xtracer.Trace("fragment/checker.reportCycle ENTER")
	defer xtracer.Trace("fragment/checker.reportCycle EXIT")
	if len(cycle) == 0 {
		return nil
	}
	xtracer.Trace("fragment/checker.reportCycle report cycle error\n stack: %v", stack())
	var parts []string
	for _, a := range cycle {
		parts = append(parts, "  "+c.reportArc(a))
	}
	return c.reportFEUError(
		"The following terms may generate an infinite sequence of instantiations:\n" +
			strings.Join(parts, "\n"))
}

func (c *checker) reportInterpOverVar(fmla lg.Expr, lineno int, node *uf.UFNode) {
	varMsg := ""
	for key, n := range c.stratMap {
		if n == node {
			if info, ok := c.stratInfo[key]; ok && info.v != nil {
				vid := makeVarID(info.v)
				if origLn, exists := c.universalVarLineno[vid]; exists {
					varMsg = fmt.Sprintf("\n%d: The quantified variable is %s", origLn, c.varUniq.Undo(info.v))
				}
			}
		}
	}
	msg := fmt.Sprintf("An interpreted symbol is applied to a universally quantified variable:\n%d: %s%s",
		lineno, c.varUniq.Undo(fmla), varMsg)
	// In Python this raises; we panic-wrap it via the caller's error handling
	panic(&FragmentError{Message: "The verification condition is not in the fragment FAU.\n\n" + msg})
}

// --- Public API ---

// CheckFEU takes lists of assumes, asserts, and macros, and checks whether
// they are collectively in the FEU fragment. Returns an error if not.
//
// Each argument is a slice of (formula, source) pairs where source provides
// line number info.
func CheckFEU(
	sig *il.Sig,
	interp map[string]interface{},
	assumes, asserts, macros []fmlaPair,
) (err error) {

	xtracer.Trace("fragment CheckFEU ENTER")
	defer xtracer.Trace("fragment CheckFEU EXIT")

	c := newChecker(sig, interp)

	// Alpha convert so all variables have unique names
	c.varUniq = il.NewVariableUniqifier(nil)

	vupair := func(p fmlaPair) fmlaPair {
		return fmlaPair{
			fmla:   c.varUniq.Uniquify(p.fmla),
			source: p.source,
			lineno: p.lineno,
		}
	}

	newAssumes := make([]fmlaPair, len(assumes))
	for i, a := range assumes {
		newAssumes[i] = vupair(a)
	}
	newAsserts := make([]fmlaPair, len(asserts))
	for i, a := range asserts {
		newAsserts[i] = vupair(a)
	}
	newMacros := make([]fmlaPair, len(macros))
	for i, m := range macros {
		newMacros[i] = vupair(m)
	}

	// Catch panics from reportInterpOverVar
	defer func() {
		if r := recover(); r != nil {
			if fe, ok := r.(*FragmentError); ok {
				err = fe
			} else {
				panic(r) // re-panic for unexpected errors
			}
			panic(r) // re-throw anyway!
		}
	}()

	if xtracer.Enabled {
		xtracer.Trace("fragment CheckFEU input counts: assumes=%d asserts=%d macros=%d", len(newAssumes), len(newAsserts), len(newMacros))

		assumesB := &strings.Builder{}
		assumesB.WriteString("[")
		for i, a := range newAssumes {
			if i > 0 {
				assumesB.WriteString(", ")
			}
			assumesB.WriteString(string(a.Canon()))
		}
		assumesB.WriteString("]")
		xtracer.Trace("fragment CheckFEU input HASH canon= assumes=%v", assumesB.String())

	}

	// Build stratification graph
	c.createStratMap(newAssumes, newAsserts, newMacros)

	if xtracer.Enabled {
		//xtracer.Trace("fragment/fragment.go CheckFEU after createStratMap, newAssumes is: HASH canon= %s", newAssumes.Canon())

		// Dump sig.Interp keys for cross-language comparison
		if c.sig != nil {
			interpKeys := make([]string, 0, len(c.sig.Interp))
			for k := range c.sig.Interp {
				interpKeys = append(interpKeys, k)
			}
			sort.Strings(interpKeys)
			xtracer.Trace("fragment sig.Interp keys: %v", interpKeys)
		} else {
			xtracer.Trace("fragment sig.Interp keys: nil (no sig)")
		}

		// Dump universally quantified variables
		uqvNames := make([]string, 0, len(c.universallyQuantifiedVars))
		for vid := range c.universallyQuantifiedVars {
			uqvNames = append(uqvNames, fmt.Sprintf("'%s:%s'", vid.name, vid.sort))
		}
		sort.Strings(uqvNames)
		xtracer.Trace("fragment HASH canon= univQuantVars (%d): [%v]", len(c.universallyQuantifiedVars), strings.Join(uqvNames, ", "))

		// Dump strat_map size and keys
		smKeys := make([]string, 0, len(c.stratMap))
		for k := range c.stratMap {
			smKeys = append(smKeys, string(k))
		}
		sort.Strings(smKeys)
		xtracer.Trace("fragment HASH canon= stratMap (%d): %v", len(c.stratMap), smKeys)

		// Dump arcs count and simplified form
		xtracer.Trace("fragment HASH canon= arcs (%d):", len(c.arcs))
		for i, a := range c.arcs {
			fromRoot := uf.Find(a.from).ID
			toRoot := uf.Find(a.to).ID
			xtracer.Trace("  arc[%d]: from_id=%d(root=%d) to_id=%d(root=%d) fmla=%s argIdx=%d", // lineno=%d
				i, a.from.ID, fromRoot, a.to.ID, toRoot, exprSexp(a.fmla), a.argIdx) // a.lineno,
		}

		xtracer.Trace("fragment/fragment.go CheckFEU after createStratMap HASH canon= %s", c.Canon())
	}

	// Check for cycles
	cycle := c.findCycle()
	if len(cycle) > 0 {
		//vv("TODO REMOVE DEBUG ONLY: skip reporing cycle!")
		return c.reportCycle(cycle)
	}

	return nil
}

// findCycle looks for a cycle in the stratification graph's arcs.
func (c *checker) findCycle() []arc {
	if len(c.arcs) == 0 {
		return nil
	}

	// Build adjacency map using UFNode IDs (after find)
	type nodeID = int64
	adj := make(map[nodeID][]arc)
	for _, a := range c.arcs {
		fromID := uf.Find(a.from).ID
		adj[fromID] = append(adj[fromID], a)
	}

	heap := make(map[nodeID]bool)  // permanently visited
	stack := make(map[nodeID]bool) // currently on DFS stack
	var path []arc

	var dfs func(nodeID) bool
	dfs = func(node nodeID) bool {
		if heap[node] {
			return false
		}
		if stack[node] {
			return true
		}
		stack[node] = true
		for _, a := range adj[node] {
			toID := uf.Find(a.to).ID
			if dfs(toID) {
				path = append(path, a)
				return true
			}
		}
		delete(stack, node)
		heap[node] = true
		return false
	}

	for _, a := range c.arcs {
		fromID := uf.Find(a.from).ID
		if dfs(fromID) {
			if len(path) == 0 {
				return nil
			}
			// Extract actual cycle
			endID := uf.Find(path[0].to).ID
			var cycle []arc
			for _, pa := range path {
				cycle = append(cycle, pa)
				if uf.Find(pa.from).ID == endID {
					// Reverse
					for i, j := 0, len(cycle)-1; i < j; i, j = i+1, j-1 {
						cycle[i], cycle[j] = cycle[j], cycle[i]
					}
					return cycle
				}
			}
			return path
		}
	}
	return nil
}

// --- Module-level API ---

// GetAssumesAndAsserts extracts all assumes, asserts, and macros from the
// current module that might end up in a prover context.
// Corresponds to Python's get_assumes_and_asserts.
func GetAssumesAndAsserts(m *module.Module, precondsOnly bool) (assumes, asserts, macros []fmlaPair) {
	xtracer.Trace("fragment: GetAssumesAndAsserts ENTER")
	defer xtracer.Trace("fragment: GetAssumesAndAsserts EXIT")

	if precondsOnly {
		for name, action := range m.BeforeExport.All() {
			_ = name
			xtracer.Trace("fragment/fragment.go:908 precondsOnly holds, calling CloseEpr() name='%v' type=%s", name, actions.ActionTypeName(action))
			fps := makeFmlaPairsFromAction(action, m, precondsOnly)
			assumes = append(assumes, fps...)
		}
	} else {
		for name := range m.PublicActions.All() {
			action, ok := m.Actions.Get2(name)
			if !ok {
				continue
			}
			xtracer.Trace("fragment/fragment.go:918 not-precondsOnly, calling CloseEpr() name='%v' type=%s", name, actions.ActionTypeName(action))

			fps := makeFmlaPairsFromAction(action, m, precondsOnly)
			assumes = append(assumes, fps...)
		}
	}

	// Definitions: non-recursive become macros, recursive become axioms
	for _, ldf := range m.Definitions {
		def, isDef := ldf.Formula.(*il.Definition)
		if !isDef {
			continue
		}
		if _, isSchema := ldf.Formula.(*il.DefinitionSchema); isSchema {
			continue
		}

		// Check if recursive (defining symbol appears in RHS)
		defSym := def.Defines()
		symsInRHS := il.SymbolsAst(def.Rhs)
		isRecursive := false
		for _, s := range symsInRHS {
			if defSym != nil && lg.Key(s) == lg.Key(defSym) {
				isRecursive = true
				break
			}
		}

		// Check if RHS is a Some
		_, isSome := def.Rhs.(*il.Some)

		if !isRecursive && !isSome {
			macros = append(macros, fmlaPair{fmla: ldf.Formula.(lg.Expr), source: ldf, lineno: ldf.Lineno})
		} else {
			// Convert to constraint
			constraint := defToConstraint(def)
			assumes = append(assumes, fmlaPair{fmla: constraint, source: ldf, lineno: ldf.Lineno})
		}
	}

	// Axioms
	for _, ldf := range m.LabeledAxioms {
		if !ldf.IsTemporal() {
			assumes = append(assumes, fmlaPair{fmla: ldf.Formula.(lg.Expr), source: ldf, lineno: ldf.Lineno})
		}
	}

	// Properties
	proofIDs := make(map[int64]bool)
	for _, pe := range m.Proofs {
		proofIDs[pe.Formula.ID] = true
	}
	subgoalIDs := make(map[int64]bool)
	for _, se := range m.Subgoals {
		subgoalIDs[se.Formula.ID] = true
	}

	for _, ldf := range m.LabeledProps {
		if !ldf.IsTemporal() {
			if !proofIDs[ldf.ID] {
				asserts = append(asserts, fmlaPair{fmla: ldf.Formula.(lg.Expr), source: ldf, lineno: ldf.Lineno})
			} else if subgoalIDs[ldf.ID] && !ldf.Explicit {
				assumes = append(assumes, fmlaPair{fmla: ldf.Formula.(lg.Expr), source: ldf, lineno: ldf.Lineno})
			}
		}
	}

	// Conjectures (both assumed and asserted)
	for _, ldf := range m.LabeledConjs {
		asserts = append(asserts, fmlaPair{fmla: ldf.Formula.(lg.Expr), source: ldf, lineno: ldf.Lineno})
		assumes = append(assumes, fmlaPair{fmla: ldf.Formula.(lg.Expr), source: ldf, lineno: ldf.Lineno})
	}

	// Assumed invariants
	for _, ldf := range m.AssumedInvs {
		if !ldf.Explicit {
			assumes = append(assumes, fmlaPair{fmla: ldf.Formula.(lg.Expr), source: ldf, lineno: ldf.Lineno})
		}
	}

	return
}

// CheckFragment checks that the current module's VCs are in the decidable fragment.
// Corresponds to Python's check_fragment.
func CheckFragment(m *module.Module, precondsOnly bool) error {
	m.CanonSnapshot("fragment/fragment.go:1001 CheckFragment()")

	logics := m.GetLogics()
	for _, l := range logics {
		if l == "fo" {
			return nil // "fo" logic skips fragment checking
		}
	}

	assumes, asserts, macros := GetAssumesAndAsserts(m, precondsOnly)

	if xtracer.Enabled {
		xtracer.Trace("fragment.CheckFragment HASH canon= assumes count=%d", len(assumes))
		for i, a := range assumes {
			xtracer.Trace("fragment.CheckFragment HASH canon= assume[%d]=%s", i, a.fmla.Canon())
		}
		xtracer.Trace("fragment.CheckFragment HASH canon= asserts count=%d", len(asserts))
		for i, a := range asserts {
			xtracer.Trace("fragment.CheckFragment HASH canon= assert[%d]=%s", i, a.fmla.Canon())
		}
		xtracer.Trace("fragment.CheckFragment HASH canon= macros count=%d", len(macros))
		for i, a := range macros {
			xtracer.Trace("fragment.CheckFragment HASH canon= macro[%d]=%s", i, a.fmla.Canon())
		}
	}

	interp := make(map[string]interface{})
	if m.Sig != nil {
		for k, v := range m.Sig.Interp {
			interp[k] = v
		}
	}

	return CheckFEU(m.Sig, interp, assumes, asserts, macros)
}

// --- helpers ---

// defToConstraint converts a Definition to a constraint formula.
func defToConstraint(d *il.Definition) lg.Expr {
	lhs := d.Lhs
	rhs := d.Rhs
	if lg.SortEqual(rhs.NodeSort(), lg.Boolean) {
		return &lg.Iff{T1: lhs, T2: rhs}
	}
	return &lg.Eq{T1: lhs, T2: rhs}
}

// makeFmlaPairsFromAction returns fmlaPairs for an action's update.
// When precondsOnly is false, it returns two pairs: TR (triple[1]) and
// Pre (triple[2]), each wrapped with CloseEpr — matching Python's
// normal mode.
// When precondsOnly is true, it returns only the TR pair (triple[1]),
// matching Python's preconds_only=True which omits triple[2].
func makeFmlaPairsFromAction(action actions.Action, m *module.Module, precondsOnly bool) []fmlaPair {

	// Compute the action's transition relation
	ctx := &actions.UpdateContext{Domain: m, ActCfg: m.Cfg.ActCfg}
	xtracer.Trace("fragment calling GetUpdate type=%s", actions.ActionTypeName(action))
	upd := actions.GetUpdate(action, ctx)
	if upd == nil {
		return nil
	}

	// Extract TR formula (triple[1]), applying close_epr.
	// Python: foo = ilu.close_epr(ilu.clauses_to_formula(triple[1]))
	//         assumes.append((foo,action))
	tr := lu.CloseEpr(upd.TRNode())
	result := []fmlaPair{
		{fmla: tr, source: action},
	}

	// When not precondsOnly, also add Pre formula (triple[2]).
	// Python: foo = ilu.close_epr(ilu.clauses_to_formula(triple[2]))
	//         assumes.append((foo,action))
	if !precondsOnly {
		pre := lu.CloseEpr(upd.PreNode())
		result = append(result, fmlaPair{fmla: pre, source: action})
	}

	return result
}
