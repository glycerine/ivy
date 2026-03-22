// Additional logic utility functions.
// Ports of functions from Python's ivy_logic_utils.py that are used by
// multiple subsystems.
package logicutil

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/goivy/logic"
)

// --- AST collectors ---

// SortsAst collects all sorts used in an AST.
// Returns map keyed by SortKey (structural identity) with Sort values.
func SortsAst(ast logic.Expr) map[logic.NodeKey]logic.Sort {
	result := make(map[logic.NodeKey]logic.Sort)
	sortsAstRec(ast, result)
	return result
}

func sortsAstRec(ast logic.Expr, result map[logic.NodeKey]logic.Sort) {
	// Matches Python sorts_ast (ivy_logic_utils.py:570-583):
	// For Apply: yield rep.sort.rng and rep.sort.dom, or recurse into binder body.
	// For Var: yield sort. Then iterate ast.args (Terms only).
	if app, ok := ast.(*logic.Apply); ok {
		if nb, ok := app.Func.(*logic.NamedBinder); ok {
			sortsAstRec(nb.Body, result)
		} else if c, ok := app.Func.(*logic.Symbol); ok {
			if fs, ok := c.CSort.(*logic.FunctionSort); ok {
				rng := fs.Range()
				result[logic.SortKey(rng)] = rng
				for _, d := range fs.Domain() {
					result[logic.SortKey(d)] = d
				}
			}
		}
	} else if v, ok := ast.(*logic.Variable); ok {
		result[logic.SortKey(v.VSort)] = v.VSort
	}
	for _, c := range ast.Children() {
		sortsAstRec(c, result)
	}
}

// RelationsAst collects all relation symbols from an AST.
func RelationsAst(ast logic.Expr) []*logic.Symbol {
	seen := make(map[string]bool)
	var result []*logic.Symbol
	relationsAstRec(ast, &result, seen)
	return result
}

func relationsAstRec(ast logic.Expr, result *[]*logic.Symbol, seen map[string]bool) {
	if app, ok := ast.(*logic.Apply); ok {
		if c, ok := app.Func.(*logic.Symbol); ok {
			if isBoolSort(app.NodeSort()) && !seen[c.Name] {
				seen[c.Name] = true
				*result = append(*result, c)
			}
		}
	}
	for _, child := range ast.Children() {
		relationsAstRec(child, result, seen)
	}
}

// FunctionsAst collects all function symbols (non-relational) from an AST.
func FunctionsAst(ast logic.Expr) []*logic.Symbol {
	seen := make(map[string]bool)
	var result []*logic.Symbol
	functionsAstRec(ast, &result, seen)
	return result
}

func functionsAstRec(ast logic.Expr, result *[]*logic.Symbol, seen map[string]bool) {
	if app, ok := ast.(*logic.Apply); ok {
		if c, ok := app.Func.(*logic.Symbol); ok {
			if !isBoolSort(app.NodeSort()) && !seen[c.Name] {
				seen[c.Name] = true
				*result = append(*result, c)
			}
		}
	}
	for _, child := range ast.Children() {
		functionsAstRec(child, result, seen)
	}
}

// EqsAst collects all equality subterms from an AST.
func EqsAst(ast logic.Expr) []*logic.Eq {
	var result []*logic.Eq
	eqsAstRec(ast, &result)
	return result
}

func eqsAstRec(ast logic.Expr, result *[]*logic.Eq) {
	if eq, ok := ast.(*logic.Eq); ok {
		*result = append(*result, eq)
	}
	for _, child := range ast.Children() {
		eqsAstRec(child, result)
	}
}

// GroundAppsAst collects all ground (variable-free) application subterms.
func GroundAppsAst(ast logic.Expr) []logic.Expr {
	var result []logic.Expr
	groundAppsAstRec(ast, &result)
	return result
}

func groundAppsAstRec(ast logic.Expr, result *[]logic.Expr) bool {
	if _, ok := ast.(*logic.Variable); ok {
		return false
	}
	if isQuantifier(ast) {
		return false
	}
	ground := true
	for _, c := range ast.Children() {
		if !groundAppsAstRec(c, result) {
			ground = false
		}
	}
	if ground && isApp(ast) {
		*result = append(*result, ast)
	}
	return ground
}

// TemporalsAst collects all temporal operators from an AST.
func TemporalsAst(ast logic.Expr) []logic.Expr {
	var result []logic.Expr
	temporalsAstRec(ast, &result)
	return result
}

func temporalsAstRec(ast logic.Expr, result *[]logic.Expr) {
	// Matches Python temporals_ast (ivy_logic_utils.py:559-568):
	// yields temporal ops, or recurses into Apply rep body if NamedBinder.
	switch ast.(type) {
	case *logic.Globally, *logic.Eventually, *logic.WhenOperator:
		*result = append(*result, ast)
	default:
		if app, ok := ast.(*logic.Apply); ok {
			if nb, ok := app.Func.(*logic.NamedBinder); ok {
				temporalsAstRec(nb.Body, result)
			}
		}
	}
	for _, c := range ast.Children() {
		temporalsAstRec(c, result)
	}
}

// NamedBindersAst collects all named binders from an AST.
func NamedBindersAst(ast logic.Expr) []*logic.NamedBinder {
	var result []*logic.NamedBinder
	namedBindersAstRec(ast, &result)
	return result
}

func namedBindersAstRec(ast logic.Expr, result *[]*logic.NamedBinder) {
	// Matches Python named_binders_ast (ivy_logic_utils.py:547-557):
	// yields standalone NamedBinder, or Apply with NamedBinder rep.
	if nb, ok := ast.(*logic.NamedBinder); ok {
		*result = append(*result, nb)
	} else if app, ok := ast.(*logic.Apply); ok {
		if nb, ok := app.Func.(*logic.NamedBinder); ok {
			*result = append(*result, nb)
			namedBindersAstRec(nb.Body, result)
		}
	}
	for _, c := range ast.Children() {
		namedBindersAstRec(c, result)
	}
}

// --- Literal predicates ---

// IsEqualityLit returns true if the literal's atom is an equality.
func IsEqualityLit(atom logic.Expr) bool {
	_, ok := atom.(*logic.Eq)
	return ok
}

// IsTautEqualityLit returns true if the literal is x=x.
func IsTautEqualityLit(atom logic.Expr) bool {
	eq, ok := atom.(*logic.Eq)
	if !ok {
		return false
	}
	return eq.T1.Equal(eq.T2)
}

// IsDisequalityLit returns true if the literal is Not(x=y).
func IsDisequalityLit(atom logic.Expr) bool {
	neg, ok := atom.(*logic.Not)
	if !ok {
		return false
	}
	_, isEq := neg.Body.(*logic.Eq)
	return isEq
}

// IsGroundLit returns true if the literal contains no variables.
func IsGroundLit(atom logic.Expr) bool {
	fv := FreeVariables(atom)
	return len(fv) == 0
}

// IsGroundEqualityLit returns true if the equality literal is ground.
func IsGroundEqualityLit(atom logic.Expr) bool {
	if !IsEqualityLit(atom) {
		return false
	}
	return IsGroundLit(atom)
}

// --- Formula / literal conversion ---

// EqLit creates an equality literal (positive).
func EqLit(x, y logic.Expr) logic.Expr {
	return &logic.Eq{T1: x, T2: y}
}

// NeqLit creates a disequality literal.
func NeqLit(x, y logic.Expr) logic.Expr {
	return &logic.Not{Body: &logic.Eq{T1: x, T2: y}}
}

// EqAtom creates an equality atom.
func EqAtom(x, y logic.Expr) *logic.Eq {
	return &logic.Eq{T1: x, T2: y}
}

// SwapArgsLit swaps the arguments of an equality literal.
func SwapArgsLit(atom logic.Expr) logic.Expr {
	if eq, ok := atom.(*logic.Eq); ok {
		return &logic.Eq{T1: eq.T2, T2: eq.T1}
	}
	if neg, ok := atom.(*logic.Not); ok {
		if eq, ok := neg.Body.(*logic.Eq); ok {
			return &logic.Not{Body: &logic.Eq{T1: eq.T2, T2: eq.T1}}
		}
	}
	return atom
}

// RelInst creates a relational instance: r(V0, V1, ...) for a relation of given arity.
func RelInst(rel *logic.Symbol) logic.Expr {
	fs, ok := rel.CSort.(*logic.FunctionSort)
	if !ok {
		return rel
	}
	dom := fs.Domain()
	vars := make([]logic.Expr, len(dom))
	for i, s := range dom {
		v, _ := logic.NewVariable(varName(i), s)
		vars[i] = v
	}
	app, err := logic.NewApply(rel, vars...)
	if err != nil {
		return rel
	}
	return app
}

// FunInst creates a function instance: f(V0, V1, ...) for a function of given arity.
func FunInst(f *logic.Symbol) logic.Expr {
	return RelInst(f) // same shape
}

// FunEqInst creates a function equality instance: Y = f(V0, V1, ...).
func FunEqInst(f *logic.Symbol) logic.Expr {
	fs, ok := f.CSort.(*logic.FunctionSort)
	if !ok {
		return f
	}
	dom := fs.Domain()
	rng := fs.Range()
	vars := make([]logic.Expr, len(dom))
	for i, s := range dom {
		v, _ := logic.NewVariable(varName(i), s)
		vars[i] = v
	}
	y, _ := logic.NewVariable("Y", rng)
	fapp, err := logic.NewApply(f, vars...)
	if err != nil {
		return f
	}
	return &logic.Eq{T1: y, T2: fapp}
}

// IsRelational returns true if the symbol has a relational sort (Boolean range).
func IsRelational(sym *logic.Symbol) bool {
	return isBoolSort(sym.CSort) || isBoolRange(sym.CSort)
}

// --- Formula transformations ---

// ExpandAbbrevs expands Iff, Implies, and Ite into And/Or/Not.
// Top-level only — does not recurse into children.
// Matches Python ivy_logic_utils.py expand_abbrevs (lines 881-900).
func ExpandAbbrevs(f logic.Expr) logic.Expr {
	switch t := f.(type) {
	case *logic.Implies:
		return &logic.Or{Terms: []logic.Expr{negate(t.T1), t.T2}}
	case *logic.Iff:
		// Special case: Iff(lhs, Ite(...)) — distribute Iff over Ite branches
		if ite, ok := t.T2.(*logic.Ite); ok {
			thenp := ExpandAbbrevs(&logic.Iff{T1: t.T1, T2: ite.Then})
			elsep := ExpandAbbrevs(&logic.Iff{T1: t.T1, T2: ite.Else})
			return ExpandAbbrevs(&logic.Ite{Cond: ite.Cond, Then: thenp, Else: elsep})
		}
		// Note: Python has is_true/is_false shortcuts here but they are
		// dead code (missing return statements). We faithfully omit them.
		return &logic.And{Terms: []logic.Expr{
			&logic.Or{Terms: []logic.Expr{negate(t.T1), t.T2}},
			&logic.Or{Terms: []logic.Expr{negate(t.T2), t.T1}},
		}}
	case *logic.Ite:
		thenps := conditionConj(negate(t.Cond), t.Then)
		elseps := conditionConj(t.Cond, t.Else)
		all := append(thenps, elseps...)
		return &logic.And{Terms: all}
	}
	return f
}

// negate returns Not(f), or unwraps double negation.
// Matches Python ivy_logic_utils.py negate (lines 926-929).
func negate(f logic.Expr) logic.Expr {
	if n, ok := f.(*logic.Not); ok {
		return n.Body
	}
	return &logic.Not{Body: f}
}

// conditionConj returns [Or(c,q) for q in p.args] if p is And, else [Or(c,p)].
// Matches Python ivy_logic_utils.py condition_conj (lines 877-879).
func conditionConj(c, p logic.Expr) []logic.Expr {
	var ps []logic.Expr
	if a, ok := p.(*logic.And); ok {
		ps = a.Terms
	} else {
		ps = []logic.Expr{p}
	}
	result := make([]logic.Expr, len(ps))
	for i, q := range ps {
		result[i] = &logic.Or{Terms: []logic.Expr{c, q}}
	}
	return result
}

// DeMorgan applies De Morgan's laws to push negation inward.
// Matches Python ivy_logic_utils.py de_morgan (lines 931-943).
func DeMorgan(f logic.Expr) logic.Expr {
	f = ExpandAbbrevs(f)
	neg, ok := f.(*logic.Not)
	if !ok {
		return f
	}
	g := DeMorgan(neg.Body)
	switch t := g.(type) {
	case *logic.And:
		if len(t.Terms) == 1 {
			return DeMorgan(negate(t.Terms[0]))
		}
		terms := make([]logic.Expr, len(t.Terms))
		for i, term := range t.Terms {
			terms[i] = negate(term)
		}
		return &logic.Or{Terms: terms}
	case *logic.Or:
		if len(t.Terms) == 1 {
			return DeMorgan(negate(t.Terms[0]))
		}
		terms := make([]logic.Expr, len(t.Terms))
		for i, term := range t.Terms {
			terms[i] = negate(term)
		}
		return &logic.And{Terms: terms}
	}
	return &logic.Not{Body: g}
}

// BooleanConstant creates a boolean constant (true or false).
func BooleanConstant(val bool) logic.Expr {
	if val {
		return &logic.And{} // empty And = true
	}
	return &logic.Or{} // empty Or = false
}

// CloseEPR converts formula to forall Y. fmla, where Y are the free variables.
// For conjunctions, applies recursively to each conjunct.
// Corresponds to Python's close_epr.
func CloseEPR(fmla logic.Expr) logic.Expr {
	if and, ok := fmla.(*logic.And); ok {
		terms := make([]logic.Expr, len(and.Terms))
		for i, t := range and.Terms {
			terms[i] = CloseEPR(t)
		}
		return &logic.And{Terms: terms}
	}
	fvs := FreeVariablesList(fmla)
	if len(fvs) == 0 {
		return fmla
	}
	return &logic.ForAll{Variables: fvs, Body: fmla}
}

// --- Resort functions ---

// ResortSort remaps a sort through a substitution.
// The subs map is keyed by SortKey (Sexp-based structural identity),
// matching Python's structural equality on Sort objects.
func ResortSort(s logic.Sort, subs map[logic.NodeKey]logic.Sort) logic.Sort {
	// Direct substitution
	key := logic.SortKey(s)
	if mapped, ok := subs[key]; ok {
		return mapped
	}
	// Recurse into FunctionSort domain/range
	// Matches Python ivy_logic_utils.py:398-410 resort_sort
	if fs, ok := s.(*logic.FunctionSort); ok {
		allSorts := make([]logic.Sort, len(fs.Sorts))
		changed := false
		for i, sub := range fs.Sorts {
			ns := ResortSort(sub, subs)
			allSorts[i] = ns
			if ns != sub {
				changed = true
			}
		}
		if !changed {
			return s
		}
		result, err := logic.NewFunctionSort(allSorts...)
		if err != nil {
			return &logic.FunctionSort{Sorts: allSorts}
		}
		return result
	}
	return s
}

// ResortSymbol returns a new Symbol with its sort remapped through subs.
// Matches Python ivy_logic_utils.py resort_symbol (lines 412-413).
func ResortSymbol(sym *logic.Symbol, subs map[logic.NodeKey]logic.Sort) *logic.Symbol {
	newSort := ResortSort(sym.CSort, subs)
	if newSort == sym.CSort {
		return sym
	}
	return logic.NewSymbol(sym.Name, newSort)
}

// ResortAst remaps all sorts in an AST through a substitution.
func ResortAst(ast logic.Expr, subs map[logic.NodeKey]logic.Sort) logic.Expr {
	switch t := ast.(type) {
	case *logic.Variable:
		newSort := ResortSort(t.VSort, subs)
		if newSort != t.VSort {
			v, _ := logic.NewVariable(t.Name, newSort)
			return v
		}
		return t
	case *logic.Symbol:
		return ResortSymbol(t, subs)
	case *logic.Apply:
		// Python: resort_symbol(ast.rep)(*args) — must resort the Func too
		newFunc := ResortAst(t.Func, subs)
		newTerms := make([]logic.Expr, len(t.Terms))
		changed := newFunc != t.Func
		for i, term := range t.Terms {
			nt := ResortAst(term, subs)
			newTerms[i] = nt
			if nt != term {
				changed = true
			}
		}
		if !changed {
			return ast
		}
		result, err := logic.NewApply(newFunc, newTerms...)
		if err != nil {
			return &logic.Apply{Func: newFunc, Terms: newTerms}
		}
		return result
	default:
		children := ast.Children()
		if len(children) == 0 {
			return ast
		}
		newChildren := make([]logic.Expr, len(children))
		changed := false
		for i, c := range children {
			nc := ResortAst(c, subs)
			newChildren[i] = nc
			if nc != c {
				changed = true
			}
		}
		if !changed {
			return ast
		}
		return cloneNode(ast, newChildren)
	}
}

// --- helpers ---

func varName(idx int) string {
	return fmt.Sprintf("V%d", idx)
}

func isBoolSort(s logic.Sort) bool {
	return logic.SortEqual(s, logic.Boolean)
}

func isBoolRange(s logic.Sort) bool {
	if fs, ok := s.(*logic.FunctionSort); ok {
		return logic.SortEqual(fs.Range(), logic.Boolean)
	}
	return false
}

func isQuantifier(n logic.Expr) bool {
	switch n.(type) {
	case *logic.ForAll, *logic.Exists:
		return true
	}
	return false
}

func isApp(n logic.Expr) bool {
	switch n.(type) {
	case *logic.Apply, *logic.Symbol:
		return true
	}
	return false
}

// cloneNode creates a clone of a node with new children.
func cloneNode(n logic.Expr, children []logic.Expr) logic.Expr {
	switch t := n.(type) {
	case *logic.Not:
		if len(children) >= 1 {
			return &logic.Not{Body: children[0]}
		}
	case *logic.And:
		return &logic.And{Terms: children}
	case *logic.Or:
		return &logic.Or{Terms: children}
	case *logic.Implies:
		if len(children) >= 2 {
			return &logic.Implies{T1: children[0], T2: children[1]}
		}
	case *logic.Iff:
		if len(children) >= 2 {
			return &logic.Iff{T1: children[0], T2: children[1]}
		}
	case *logic.Eq:
		if len(children) >= 2 {
			return &logic.Eq{T1: children[0], T2: children[1]}
		}
	case *logic.Ite:
		if len(children) >= 3 {
			return &logic.Ite{ISort: t.ISort, Cond: children[0], Then: children[1], Else: children[2]}
		}
	case *logic.Apply:
		// Matches Python Apply.clone(args) = Apply(self.func, *args).
		// Children() returns Terms only, so children ARE the new terms.
		// Preserve the original Func.
		result, err := logic.NewApply(t.Func, children...)
		if err != nil {
			// Fallback: construct directly (may have TopSort issues)
			return n
		}
		return result
	case *logic.ForAll:
		if len(children) >= 1 {
			return &logic.ForAll{Variables: t.Variables, Body: children[0]}
		}
	case *logic.Exists:
		if len(children) >= 1 {
			return &logic.Exists{Variables: t.Variables, Body: children[0]}
		}
	case *logic.Lambda:
		if len(children) >= 1 {
			return &logic.Lambda{Variables: t.Variables, Body: children[0]}
		}
	case *logic.Globally:
		if len(children) >= 1 {
			return &logic.Globally{Environ: t.Environ, Body: children[0]}
		}
	case *logic.Eventually:
		if len(children) >= 1 {
			return &logic.Eventually{Environ: t.Environ, Body: children[0]}
		}
	case *logic.NamedBinder:
		if len(children) >= 1 {
			return &logic.NamedBinder{Name: t.Name, Variables: t.Variables, Environ: t.Environ, Body: children[0]}
		}
	case *logic.WhenOperator:
		if len(children) >= 2 {
			return &logic.WhenOperator{WSort: t.WSort, Name: t.Name, T1: children[0], T2: children[1]}
		}
	case *logic.Cond:
		if len(children) >= 2 {
			return &logic.Cond{CSort: t.CSort, T1: children[0], T2: children[1]}
		}
	}
	return n
}

// --- Name-based substitution (needed by several functions below) ---

// SubstituteByName substitutes terms for variables by name (string key).
// This mirrors Python's substitute_ast(ast, subs) where subs maps
// variable names to replacement terms.
func SubstituteByName(ast logic.Expr, subs map[string]logic.Expr) logic.Expr {
	if len(subs) == 0 {
		return ast
	}
	return substituteByNameRec(ast, subs)
}

func substituteByNameRec(ast logic.Expr, subs map[string]logic.Expr) logic.Expr {
	if v, ok := ast.(*logic.Variable); ok {
		if rep, found := subs[v.Name]; found {
			return rep
		}
		return ast
	}
	// For quantifiers/binders, remove bound variables from subs
	switch t := ast.(type) {
	case *logic.ForAll:
		newsubs := removeBoundNames(subs, t.Variables)
		body := substituteByNameRec(t.Body, newsubs)
		return &logic.ForAll{Variables: t.Variables, Body: body}
	case *logic.Exists:
		newsubs := removeBoundNames(subs, t.Variables)
		body := substituteByNameRec(t.Body, newsubs)
		return &logic.Exists{Variables: t.Variables, Body: body}
	case *logic.Lambda:
		newsubs := removeBoundNames(subs, t.Variables)
		body := substituteByNameRec(t.Body, newsubs)
		return &logic.Lambda{Variables: t.Variables, Body: body}
	case *logic.NamedBinder:
		newsubs := removeBoundNames(subs, t.Variables)
		body := substituteByNameRec(t.Body, newsubs)
		return &logic.NamedBinder{Name: t.Name, Variables: t.Variables, Environ: t.Environ, Body: body}
	}
	children := ast.Children()
	if len(children) == 0 {
		return ast
	}
	newChildren := make([]logic.Expr, len(children))
	changed := false
	for i, c := range children {
		nc := substituteByNameRec(c, subs)
		newChildren[i] = nc
		if nc != c {
			changed = true
		}
	}
	if !changed {
		return ast
	}
	return cloneNode(ast, newChildren)
}

func removeBoundNames(subs map[string]logic.Expr, vars []*logic.Variable) map[string]logic.Expr {
	newsubs := make(map[string]logic.Expr, len(subs))
	boundNames := make(map[string]bool, len(vars))
	for _, v := range vars {
		boundNames[v.Name] = true
	}
	for k, v := range subs {
		if !boundNames[k] {
			newsubs[k] = v
		}
	}
	return newsubs
}

// --- Free variable collection in order (for normalization) ---

// freeVariablesInOrder returns free variables in the order they first appear,
// with unique names only.
func freeVariablesInOrder(ast logic.Expr) []*logic.Variable {
	seen := make(map[string]bool)
	var result []*logic.Variable
	freeVariablesInOrderRec(ast, &result, seen, nil)
	return result
}

// freeVariablesInOrderMulti returns free variables across multiple ASTs,
// unique by name, in order of first appearance. Matches Python's
// used_variables_asts which uses variables_ast (free vars only).
func freeVariablesInOrderMulti(asts []logic.Expr) []*logic.Variable {
	seen := make(map[string]bool)
	var result []*logic.Variable
	for _, ast := range asts {
		freeVariablesInOrderRec(ast, &result, seen, nil)
	}
	return result
}

func freeVariablesInOrderRec(ast logic.Expr, result *[]*logic.Variable, seen map[string]bool, bound map[string]bool) {
	if v, ok := ast.(*logic.Variable); ok {
		if !bound[v.Name] && !seen[v.Name] {
			seen[v.Name] = true
			*result = append(*result, v)
		}
		return
	}
	switch t := ast.(type) {
	case *logic.ForAll:
		newBound := copyBoundSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = true
		}
		freeVariablesInOrderRec(t.Body, result, seen, newBound)
		return
	case *logic.Exists:
		newBound := copyBoundSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = true
		}
		freeVariablesInOrderRec(t.Body, result, seen, newBound)
		return
	case *logic.Lambda:
		newBound := copyBoundSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = true
		}
		freeVariablesInOrderRec(t.Body, result, seen, newBound)
		return
	case *logic.NamedBinder:
		newBound := copyBoundSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = true
		}
		freeVariablesInOrderRec(t.Body, result, seen, newBound)
		return
	}
	for _, c := range ast.Children() {
		freeVariablesInOrderRec(c, result, seen, bound)
	}
}

func copyBoundSet(s map[string]bool) map[string]bool {
	r := make(map[string]bool, len(s))
	for k, v := range s {
		r[k] = v
	}
	return r
}

// usedVariablesInOrder returns all variables used (free or bound) in order of
// first appearance, unique by name.
func usedVariablesInOrder(ast logic.Expr) []*logic.Variable {
	seen := make(map[string]bool)
	var result []*logic.Variable
	usedVariablesInOrderRec(ast, &result, seen)
	return result
}

func usedVariablesInOrderRec(ast logic.Expr, result *[]*logic.Variable, seen map[string]bool) {
	if v, ok := ast.(*logic.Variable); ok {
		if !seen[v.Name] {
			seen[v.Name] = true
			*result = append(*result, v)
		}
		return
	}
	switch t := ast.(type) {
	case *logic.ForAll:
		for _, v := range t.Variables {
			if !seen[v.Name] {
				seen[v.Name] = true
				*result = append(*result, v)
			}
		}
		usedVariablesInOrderRec(t.Body, result, seen)
		return
	case *logic.Exists:
		for _, v := range t.Variables {
			if !seen[v.Name] {
				seen[v.Name] = true
				*result = append(*result, v)
			}
		}
		usedVariablesInOrderRec(t.Body, result, seen)
		return
	case *logic.Lambda:
		for _, v := range t.Variables {
			if !seen[v.Name] {
				seen[v.Name] = true
				*result = append(*result, v)
			}
		}
		usedVariablesInOrderRec(t.Body, result, seen)
		return
	case *logic.NamedBinder:
		for _, v := range t.Variables {
			if !seen[v.Name] {
				seen[v.Name] = true
				*result = append(*result, v)
			}
		}
		usedVariablesInOrderRec(t.Body, result, seen)
		return
	}
	for _, c := range ast.Children() {
		usedVariablesInOrderRec(c, result, seen)
	}
}

// usedVariablesInOrderMulti returns all variables used across multiple ASTs,
// unique by name, in order of first appearance.
func usedVariablesInOrderMulti(asts []logic.Expr) []*logic.Variable {
	seen := make(map[string]bool)
	var result []*logic.Variable
	for _, ast := range asts {
		usedVariablesInOrderRec(ast, &result, seen)
	}
	return result
}

// --- NormalizeFreeVariables ---

// NormalizeFreeVariables normalizes free variables: renames them V0, V1, ...
// in the order they appear.
// Returns (old_vars, new_vars, normalized_ast).
func NormalizeFreeVariables(ast logic.Expr) ([]*logic.Variable, []*logic.Variable, logic.Expr) {
	subs := make(map[string]logic.Expr)
	var vs []*logic.Variable
	var nvs []*logic.Variable
	for _, v := range freeVariablesInOrder(ast) {
		if _, exists := subs[v.Name]; !exists {
			nv, _ := logic.NewVariable(fmt.Sprintf("V%d", len(subs)), v.VSort)
			subs[v.Name] = nv
			vs = append(vs, v)
			nvs = append(nvs, nv)
		}
	}
	nast := SubstituteByName(ast, subs)
	return vs, nvs, nast
}

// NormalizeFreeVariablesTuple normalizes free variables across a tuple of ASTs.
// Returns (old_vars, new_vars, normalized_asts).
func NormalizeFreeVariablesTuple(asts ...logic.Expr) ([]*logic.Variable, []*logic.Variable, []logic.Expr) {
	subs := make(map[string]logic.Expr)
	var vs []*logic.Variable
	var nvs []*logic.Variable
	for _, v := range freeVariablesInOrderMulti(asts) {
		if _, exists := subs[v.Name]; !exists {
			nv, _ := logic.NewVariable(fmt.Sprintf("V%d", len(subs)), v.VSort)
			subs[v.Name] = nv
			vs = append(vs, v)
			nvs = append(nvs, nv)
		}
	}
	nasts := make([]logic.Expr, len(asts))
	for i, ast := range asts {
		nasts[i] = SubstituteByName(ast, subs)
	}
	return vs, nvs, nasts
}

// --- NormalizeNamedBinders ---

// NormalizeNamedBinders normalizes the variables bound by named binders.
// If names is nil, all named binders are normalized; otherwise only those
// whose name is in the names set.
func NormalizeNamedBinders(ast logic.Expr, names map[string]bool) logic.Expr {
	if _, ok := ast.(*logic.Symbol); ok {
		return ast
	}
	if nb, ok := ast.(*logic.NamedBinder); ok {
		if names == nil || names[nb.Name] {
			nvs := make([]*logic.Variable, len(nb.Variables))
			subs := make(map[string]logic.Expr, len(nb.Variables))
			for i, v := range nb.Variables {
				nv, _ := logic.NewVariable(fmt.Sprintf("V%d", i), v.VSort)
				nvs[i] = nv
				subs[v.Name] = nv
			}
			// Python assertion: generated V0,V1,... must not clash with free variables
			free := FreeVariablesSet(ast)
			for _, nv := range nvs {
				if free[nv.Name] {
					panic(fmt.Sprintf("NormalizeNamedBinders: generated variable %s clashes with free variable", nv.Name))
				}
			}
			body := NormalizeNamedBinders(nb.Body, names)
			body = SubstituteByName(body, subs)
			return &logic.NamedBinder{Name: nb.Name, Variables: nvs, Environ: nb.Environ, Body: body}
		}
	}
	// For Apply nodes, normalize the func part too
	if app, ok := ast.(*logic.Apply); ok {
		newFunc := NormalizeNamedBinders(app.Func, names)
		newTerms := make([]logic.Expr, len(app.Terms))
		for i, t := range app.Terms {
			newTerms[i] = NormalizeNamedBinders(t, names)
		}
		result, err := logic.NewApply(newFunc, newTerms...)
		if err != nil {
			return &logic.Apply{Func: newFunc, Terms: newTerms}
		}
		return result
	}
	children := ast.Children()
	if len(children) == 0 {
		return ast
	}
	newChildren := make([]logic.Expr, len(children))
	changed := false
	for i, c := range children {
		nc := NormalizeNamedBinders(c, names)
		newChildren[i] = nc
		if nc != c {
			changed = true
		}
	}
	if !changed {
		return ast
	}
	return cloneNode(ast, newChildren)
}

// --- Temporal -> NamedBinder conversion ---

// GloballyBinderFunc creates a named binder from a Globally operator.
type GloballyBinderFunc func(vars []*logic.Variable, body logic.Expr, environ *string) *logic.NamedBinder

// WhenBinderFunc creates a named binder from a WhenOperator.
type WhenBinderFunc func(name string, vars []*logic.Variable, body logic.Expr) *logic.NamedBinder

// DefaultGloballyBinder creates a NamedBinder named "g" (or "g[env]" if env != nil).
func DefaultGloballyBinder(vars []*logic.Variable, body logic.Expr, environ *string) *logic.NamedBinder {
	name := "g"
	if environ != nil {
		name = "g[" + *environ + "]"
	}
	return &logic.NamedBinder{Name: name, Variables: vars, Environ: nil, Body: body}
}

// DefaultWhenBinder creates a NamedBinder named "l2s_when<name>".
func DefaultWhenBinder(name string, vars []*logic.Variable, body logic.Expr) *logic.NamedBinder {
	return &logic.NamedBinder{Name: "l2s_when" + name, Variables: vars, Environ: nil, Body: body}
}

// applyNamedBinder applies a named binder to arguments, returning Apply if
// there are arguments or the binder itself if there are none.
func applyNamedBinder(nb *logic.NamedBinder, args []logic.Expr) logic.Expr {
	if len(args) == 0 {
		return nb
	}
	result, err := logic.NewApply(nb, args...)
	if err != nil {
		return &logic.Apply{Func: nb, Terms: args}
	}
	return result
}

// ReplaceTemporalsByNamedBinder replaces temporal operators (Globally,
// Eventually, WhenOperator) with named binders.
// If g is nil, DefaultGloballyBinder is used.
// If when is nil, DefaultWhenBinder is used.
func ReplaceTemporalsByNamedBinder(ast logic.Expr, g GloballyBinderFunc, when WhenBinderFunc) logic.Expr {
	if g == nil {
		g = DefaultGloballyBinder
	}
	if when == nil {
		when = DefaultWhenBinder
	}
	return replaceTemporalsRec(ast, g, when)
}

func replaceTemporalsRec(ast logic.Expr, g GloballyBinderFunc, when WhenBinderFunc) logic.Expr {
	switch t := ast.(type) {
	case *logic.Globally:
		body := replaceTemporalsRec(t.Body, g, when)
		vs, nvs, body := NormalizeFreeVariables(body)
		nb := g(nvs, body, t.Environ)
		return applyNamedBinder(nb, varsToNodes(vs))

	case *logic.Eventually:
		notBody := &logic.Not{Body: t.Body}
		glob := &logic.Globally{Environ: t.Environ, Body: notBody}
		return replaceTemporalsRec(&logic.Not{Body: glob}, g, when)

	case *logic.WhenOperator:
		val := replaceTemporalsRec(t.T1, g, when)
		cond := replaceTemporalsRec(t.T2, g, when)
		body := &logic.Cond{CSort: val.NodeSort(), T1: cond, T2: val}
		vs, nvs, nbody := NormalizeFreeVariables(body)
		nb := when(t.Name, nvs, nbody)
		return applyNamedBinder(nb, varsToNodes(vs))

	case *logic.Apply:
		if nb, ok := t.Func.(*logic.NamedBinder); ok && nb.Name == "l2s_init" {
			body := replaceTemporalsRec(nb.Body, g, when)
			newArgs := make([]logic.Expr, len(t.Terms))
			for i, a := range t.Terms {
				newArgs[i] = replaceTemporalsRec(a, g, when)
			}
			if notBody, ok := body.(*logic.Not); ok {
				newNB := &logic.NamedBinder{Name: nb.Name, Variables: nb.Variables, Environ: nb.Environ, Body: notBody.Body}
				inner, err := logic.NewApply(newNB, newArgs...)
				if err != nil {
					inner = &logic.Apply{Func: newNB, Terms: newArgs}
				}
				return &logic.Not{Body: inner}
			}
			newNB := &logic.NamedBinder{Name: nb.Name, Variables: nb.Variables, Environ: nb.Environ, Body: body}
			result, err := logic.NewApply(newNB, newArgs...)
			if err != nil {
				return &logic.Apply{Func: newNB, Terms: newArgs}
			}
			return result
		}
		newFunc := replaceTemporalsRec(t.Func, g, when)
		newTerms := make([]logic.Expr, len(t.Terms))
		for i, a := range t.Terms {
			newTerms[i] = replaceTemporalsRec(a, g, when)
		}
		result, err := logic.NewApply(newFunc, newTerms...)
		if err != nil {
			return &logic.Apply{Func: newFunc, Terms: newTerms}
		}
		return result

	case *logic.Not:
		body := replaceTemporalsRec(t.Body, g, when)
		if inner, ok := body.(*logic.Not); ok {
			return inner.Body
		}
		return &logic.Not{Body: body}

	case *logic.NamedBinder:
		if t.Name == "l2s_init" {
			body := replaceTemporalsRec(t.Body, g, when)
			if notBody, ok := body.(*logic.Not); ok {
				newNB := &logic.NamedBinder{Name: t.Name, Variables: t.Variables, Environ: t.Environ, Body: notBody.Body}
				return &logic.Not{Body: newNB}
			}
			return &logic.NamedBinder{Name: t.Name, Variables: t.Variables, Environ: t.Environ, Body: body}
		}
	}

	children := ast.Children()
	if len(children) == 0 {
		return ast
	}
	newChildren := make([]logic.Expr, len(children))
	changed := false
	for i, c := range children {
		nc := replaceTemporalsRec(c, g, when)
		newChildren[i] = nc
		if nc != c {
			changed = true
		}
	}
	if !changed {
		return ast
	}
	return cloneNode(ast, newChildren)
}

func varsToNodes(vars []*logic.Variable) []logic.Expr {
	nodes := make([]logic.Expr, len(vars))
	for i, v := range vars {
		nodes[i] = v
	}
	return nodes
}

// --- ReduceNamedBinders ---

// ReduceNamedBinders reduces subexpressions of the form ($b V1...Vn. t)(s1...sn)
// to ($b . t[Vi/si]).
func ReduceNamedBinders(ast logic.Expr, g GloballyBinderFunc) logic.Expr {
	if g == nil {
		g = DefaultGloballyBinder
	}
	return reduceNamedBindersRec(ast, g)
}

func reduceNamedBindersRec(ast logic.Expr, g GloballyBinderFunc) logic.Expr {
	if _, ok := ast.(*logic.Symbol); ok {
		return ast
	}
	if app, ok := ast.(*logic.Apply); ok {
		if nb, ok := app.Func.(*logic.NamedBinder); ok {
			subst := make(map[string]logic.Expr, len(nb.Variables))
			for i, v := range nb.Variables {
				if i < len(app.Terms) {
					subst[v.Name] = app.Terms[i]
				}
			}
			body := reduceNamedBindersRec(SubstituteByName(nb.Body, subst), g)
			return &logic.NamedBinder{Name: nb.Name, Variables: nil, Environ: nil, Body: body}
		}
		newFunc := NormalizeNamedBinders(app.Func, nil)
		newTerms := make([]logic.Expr, len(app.Terms))
		for i, t := range app.Terms {
			newTerms[i] = reduceNamedBindersRec(t, g)
		}
		result, err := logic.NewApply(newFunc, newTerms...)
		if err != nil {
			return &logic.Apply{Func: newFunc, Terms: newTerms}
		}
		return result
	}

	children := ast.Children()
	if len(children) == 0 {
		return ast
	}
	newChildren := make([]logic.Expr, len(children))
	changed := false
	for i, c := range children {
		nc := reduceNamedBindersRec(c, g)
		newChildren[i] = nc
		if nc != c {
			changed = true
		}
	}
	if !changed {
		return ast
	}
	return cloneNode(ast, newChildren)
}

// --- ReplaceNamedBindersAst ---

// ReplaceNamedBindersAst replaces named binders that are present in subs.
// subs maps NamedBinder string keys to replacement nodes.
// Named binders inside other named binders are NOT replaced.
func ReplaceNamedBindersAst(ast logic.Expr, subs map[string]logic.Expr) logic.Expr {
	if nb, ok := ast.(*logic.NamedBinder); ok {
		key := nb.String()
		if rep, found := subs[key]; found {
			return rep
		}
		return ast
	}
	if app, ok := ast.(*logic.Apply); ok {
		newTerms := make([]logic.Expr, len(app.Terms))
		for i, t := range app.Terms {
			newTerms[i] = ReplaceNamedBindersAst(t, subs)
		}
		newFunc := app.Func
		if nb, ok := app.Func.(*logic.NamedBinder); ok {
			key := nb.String()
			if rep, found := subs[key]; found {
				newFunc = rep
			}
		}
		result, err := logic.NewApply(newFunc, newTerms...)
		if err != nil {
			return &logic.Apply{Func: newFunc, Terms: newTerms}
		}
		return result
	}
	children := ast.Children()
	if len(children) == 0 {
		return ast
	}
	newChildren := make([]logic.Expr, len(children))
	changed := false
	for i, c := range children {
		nc := ReplaceNamedBindersAst(c, subs)
		newChildren[i] = nc
		if nc != c {
			changed = true
		}
	}
	if !changed {
		return ast
	}
	return cloneNode(ast, newChildren)
}

// --- ExpandNamedBindersAst ---

// ExpandNamedBindersAst expands nullary named binders by applying fun.
// If fun returns a non-nil node, the result is recursively expanded.
func ExpandNamedBindersAst(ast logic.Expr, fun func(*logic.NamedBinder) logic.Expr) logic.Expr {
	if nb, ok := ast.(*logic.NamedBinder); ok {
		res := fun(nb)
		if res != nil {
			return ExpandNamedBindersAst(res, fun)
		}
	}
	children := ast.Children()
	if len(children) == 0 {
		return ast
	}
	newChildren := make([]logic.Expr, len(children))
	changed := false
	for i, c := range children {
		nc := ExpandNamedBindersAst(c, fun)
		newChildren[i] = nc
		if nc != c {
			changed = true
		}
	}
	if !changed {
		return ast
	}
	return cloneNode(ast, newChildren)
}

// --- DenormalizeTemporal ---

// IsTrue returns true if the node is the logical true constant (empty And).
func IsTrue(n logic.Expr) bool {
	a, ok := n.(*logic.And)
	return ok && len(a.Terms) == 0
}

// IsFalse returns true if the node is the logical false constant (empty Or).
func IsFalse(n logic.Expr) bool {
	o, ok := n.(*logic.Or)
	return ok && len(o.Terms) == 0
}

// DenormalizeTemporal reverses temporal normalization:
//   - Not(Globally(Not(x))) => Eventually(x)
//   - Iff/Eq(Globally(Not(x)), true) => Iff/Eq(Eventually(x), Or())
//   - Iff/Eq(Globally(Not(x)), false) => Iff/Eq(Eventually(x), And())
func DenormalizeTemporal(ast logic.Expr) logic.Expr {
	children := ast.Children()
	newChildren := make([]logic.Expr, len(children))
	for i, c := range children {
		newChildren[i] = DenormalizeTemporal(c)
	}

	// Not(Globally(Not(x))) => Eventually(x)
	if _, ok := ast.(*logic.Not); ok && len(newChildren) == 1 {
		if glob, ok := newChildren[0].(*logic.Globally); ok {
			if inner, ok := glob.Body.(*logic.Not); ok {
				return &logic.Eventually{Environ: glob.Environ, Body: inner.Body}
			}
		}
	}

	// Iff/Eq with Globally(Not(x)) on the left
	switch ast.(type) {
	case *logic.Iff, *logic.Eq:
		if len(newChildren) >= 2 {
			lhs := newChildren[0]
			rhs := newChildren[1]
			if glob, ok := lhs.(*logic.Globally); ok {
				if inner, ok := glob.Body.(*logic.Not); ok {
					ev := &logic.Eventually{Environ: glob.Environ, Body: inner.Body}
					if IsTrue(rhs) {
						return cloneNode(ast, []logic.Expr{ev, &logic.Or{}})
					}
					if IsFalse(rhs) {
						return cloneNode(ast, []logic.Expr{ev, &logic.And{}})
					}
				}
			}
		}
	}

	return cloneNode(ast, newChildren)
}

// --- RenameClausesAnnotFun ---

// Renamer is the interface for annotation objects that support Rename.
type Renamer interface {
	Rename(m map[string]string) interface{}
}

// RenameClausesAnnotFun renames within clause annotations.
// If annot is nil, returns nil. If annot implements Renamer, calls Rename.
func RenameClausesAnnotFun(annot interface{}, m map[string]string) interface{} {
	if annot == nil {
		return nil
	}
	if r, ok := annot.(Renamer); ok {
		return r.Rename(m)
	}
	return annot
}

// --- Literal truth value checks ---

// IsVacEqualityLit returns true if the literal is a negative equality x=x.
// (Vacuously true disequality: Not(x=x).)
func IsVacEqualityLit(lit logic.Expr) bool {
	neg, ok := lit.(*logic.Not)
	if !ok {
		return false
	}
	eq, ok := neg.Body.(*logic.Eq)
	if !ok {
		return false
	}
	return eq.T1.Equal(eq.T2)
}

// IsTrueLit returns true if the literal evaluates to true.
// A positive literal that is And() (true), or a Not(Or()) (not false).
func IsTrueLit(lit logic.Expr) bool {
	if neg, ok := lit.(*logic.Not); ok {
		return IsFalse(neg.Body)
	}
	return IsTrue(lit)
}

// IsFalseLit returns true if the literal evaluates to false.
// A positive literal that is Or() (false), or a Not(And()) (not true).
func IsFalseLit(lit logic.Expr) bool {
	if neg, ok := lit.(*logic.Not); ok {
		return IsTrue(neg.Body)
	}
	return IsFalse(lit)
}

// IsTautLit returns true if the literal is a tautology (always true).
func IsTautLit(lit logic.Expr) bool {
	return IsTrueLit(lit) || IsTautEqualityLit(lit)
}

// IsVacLit returns true if the literal is vacuously true (always false).
func IsVacLit(lit logic.Expr) bool {
	return IsFalseLit(lit) || IsVacEqualityLit(lit)
}

// --- TermListsEq ---

// TermListsEq returns true if two node slices are structurally equal.
func TermListsEq(l1, l2 []logic.Expr) bool {
	if len(l1) != len(l2) {
		return false
	}
	for i := range l1 {
		if !l1[i].Equal(l2[i]) {
			return false
		}
	}
	return true
}

// --- Tseitin encoding ---

// TseitinContext manages the Tseitin transformation state.
type TseitinContext struct {
	Clauses [][]logic.Expr // accumulated definition clauses
	used    map[string]bool
	counter int
}

// NewTseitinContext creates a new TseitinContext.
func NewTseitinContext(used map[string]bool) *TseitinContext {
	if used == nil {
		used = make(map[string]bool)
	}
	return &TseitinContext{used: used}
}

// fresh generates a unique Tseitin symbol name.
func (tc *TseitinContext) fresh(hint string) string {
	for {
		name := fmt.Sprintf("__ts%s:%d", hint, tc.counter)
		tc.counter++
		if !tc.used[name] {
			tc.used[name] = true
			return name
		}
	}
}

// AddDefs prepends the accumulated Tseitin clauses to the given clauses.
func (tc *TseitinContext) AddDefs(cls [][]logic.Expr) [][]logic.Expr {
	result := make([][]logic.Expr, 0, len(tc.Clauses)+len(cls))
	result = append(result, cls...)
	result = append(result, tc.Clauses...)
	return result
}

// TseitinEncoding performs Tseitin transformation on a formula.
// Returns a node representing the formula as a literal.
// Side effect: adds definition clauses to the TseitinContext.
func TseitinEncoding(tc *TseitinContext, f logic.Expr) logic.Expr {
	f = ExpandAbbrevs(f)

	if and, ok := f.(*logic.And); ok {
		args := make([]logic.Expr, len(and.Terms))
		for i, t := range and.Terms {
			args[i] = TseitinEncoding(tc, t)
		}
		// Collect variables in order from args
		varsSeen := make(map[string]bool)
		var vs []*logic.Variable
		for _, arg := range args {
			for _, v := range freeVariablesInOrder(arg) {
				if !varsSeen[v.Name] {
					varsSeen[v.Name] = true
					vs = append(vs, v)
				}
			}
		}
		fname := tc.fresh(strconv.Itoa(len(vs)))
		var fnSort logic.Sort
		if len(vs) > 0 {
			sorts := make([]logic.Sort, len(vs)+1)
			for i, v := range vs {
				sorts[i] = v.VSort
			}
			sorts[len(vs)] = logic.Boolean
			fs, err := logic.NewFunctionSort(sorts...)
			if err != nil {
				fnSort = logic.Boolean
			} else {
				fnSort = fs
			}
		} else {
			fnSort = logic.Boolean
		}
		fn := logic.NewSymbol(fname, fnSort)
		var res logic.Expr
		if len(vs) > 0 {
			r, err := logic.NewApply(fn, varsToNodes(vs)...)
			if err != nil {
				res = &logic.Apply{Func: fn, Terms: varsToNodes(vs)}
			} else {
				res = r
			}
		} else {
			res = fn
		}
		for _, arg := range args {
			tc.Clauses = append(tc.Clauses, []logic.Expr{&logic.Not{Body: res}, arg})
		}
		lastClause := []logic.Expr{res}
		for _, arg := range args {
			lastClause = append(lastClause, &logic.Not{Body: arg})
		}
		tc.Clauses = append(tc.Clauses, lastClause)
		return res
	}

	if or, ok := f.(*logic.Or); ok {
		negTerms := make([]logic.Expr, len(or.Terms))
		for i, t := range or.Terms {
			negTerms[i] = &logic.Not{Body: t}
		}
		return &logic.Not{Body: TseitinEncoding(tc, &logic.And{Terms: negTerms})}
	}

	if not, ok := f.(*logic.Not); ok {
		return &logic.Not{Body: TseitinEncoding(tc, not.Body)}
	}

	// Atom: return as-is
	return f
}

// --- FormulaToLit ---

// FormulaToLit converts a formula to a literal (positive or negated atom).
// Uses Tseitin encoding for complex subformulas.
func FormulaToLit(tc *TseitinContext, f logic.Expr) logic.Expr {
	f = ExpandAbbrevs(f)
	if not, ok := f.(*logic.Not); ok {
		inner := FormulaToLit(tc, not.Body)
		return &logic.Not{Body: inner}
	}
	if isAtomNode(f) {
		return f
	}
	return TseitinEncoding(tc, f)
}

// isAtomNode checks if the node is an atomic formula.
func isAtomNode(n logic.Expr) bool {
	if _, ok := n.(*logic.Eq); ok {
		return true
	}
	switch n.(type) {
	case *logic.Apply:
		return logic.SortEqual(n.NodeSort(), logic.Boolean)
	case *logic.Symbol:
		return logic.SortEqual(n.NodeSort(), logic.Boolean)
	}
	return false
}

// --- FormulaToClause ---

// FormulaToClause converts a formula to a clause (disjunction of literals).
// Returns a slice of literal nodes.
func FormulaToClause(tc *TseitinContext, f logic.Expr) []logic.Expr {
	f = ExpandAbbrevs(f)
	f = DeMorgan(f)
	if IsTrue(f) {
		return []logic.Expr{f}
	}
	if or, ok := f.(*logic.Or); ok {
		var result []logic.Expr
		for _, t := range or.Terms {
			result = append(result, FormulaToClause(tc, t)...)
		}
		return result
	}
	return []logic.Expr{FormulaToLit(tc, f)}
}

// --- FormulaToClube ---

// FormulaToClube converts a formula to a cube (conjunction of literals).
// Returns a slice of literal nodes.
func FormulaToClube(tc *TseitinContext, f logic.Expr) []logic.Expr {
	f = ExpandAbbrevs(f)
	if not, ok := f.(*logic.Not); ok {
		clause := FormulaToClause(tc, not.Body)
		result := make([]logic.Expr, len(clause))
		for i, lit := range clause {
			result[i] = &logic.Not{Body: lit}
		}
		return result
	}
	if and, ok := f.(*logic.And); ok {
		var result []logic.Expr
		for _, t := range and.Terms {
			result = append(result, FormulaToLit(tc, t))
		}
		return result
	}
	return []logic.Expr{FormulaToLit(tc, f)}
}

// --- ReduceNumerically ---

// ReduceNumerically evaluates numeric expressions where all arguments
// are numeral constants. Handles <, =, and .succ operators.
func ReduceNumerically(ast logic.Expr) logic.Expr {
	children := ast.Children()
	newChildren := make([]logic.Expr, len(children))
	for i, c := range children {
		newChildren[i] = ReduceNumerically(c)
	}

	if app, ok := ast.(*logic.Apply); ok {
		allNumeral := len(app.Terms) > 0
		for i := range app.Terms {
			nc := newChildren[i] // Children() = Terms only (Func excluded)
			c, ok := nc.(*logic.Symbol)
			if !ok || !isAllDigits(c.Name) {
				allNumeral = false
				break
			}
		}
		if allNumeral {
			vals := make([]int, len(app.Terms))
			for i := range app.Terms {
				nc := newChildren[i]
				c := nc.(*logic.Symbol)
				vals[i], _ = strconv.Atoi(c.Name)
			}
			if fn, ok := app.Func.(*logic.Symbol); ok {
				if fn.Name == "<" && len(vals) == 2 {
					return BooleanConstant(vals[0] < vals[1])
				}
				if fn.Name == "=" && len(vals) == 2 {
					return BooleanConstant(vals[0] == vals[1])
				}
				if strings.HasSuffix(fn.Name, ".succ") && len(vals) == 2 {
					return BooleanConstant(vals[0]+1 == vals[1])
				}
			}
		}
	}

	return cloneNode(ast, newChildren)
}

func isAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, b := range s {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// -----------------------------------------------------------------------
// NormalizeQuantifiers
// -----------------------------------------------------------------------

// NormalizeQuantifiers pushes universals inside conjunctions and
// existentials inside disjunctions, and flattens conjunctions and
// disjunctions. Corresponds to Python logic_util.normalize_quantifiers.
func NormalizeQuantifiers(t logic.Expr) logic.Expr {
	switch n := t.(type) {
	case *logic.Variable, *logic.Symbol:
		return t

	case *logic.Apply:
		newTerms := make([]logic.Expr, len(n.Terms))
		for i, term := range n.Terms {
			newTerms[i] = NormalizeQuantifiers(term)
		}
		return &logic.Apply{Func: n.Func, Terms: newTerms}

	case *logic.Eq:
		return &logic.Eq{T1: NormalizeQuantifiers(n.T1), T2: NormalizeQuantifiers(n.T2)}

	case *logic.Ite:
		return &logic.Ite{
			Cond: NormalizeQuantifiers(n.Cond),
			Then: NormalizeQuantifiers(n.Then),
			Else: NormalizeQuantifiers(n.Else),
		}

	case *logic.Not:
		return &logic.Not{Body: NormalizeQuantifiers(n.Body)}

	case *logic.Implies:
		return &logic.Implies{T1: NormalizeQuantifiers(n.T1), T2: NormalizeQuantifiers(n.T2)}

	case *logic.Iff:
		return &logic.Iff{T1: NormalizeQuantifiers(n.T1), T2: NormalizeQuantifiers(n.T2)}

	case *logic.And:
		// Flatten: And(x, And(a,b), y) -> And(x, a, b, y)
		var terms []logic.Expr
		for _, x := range n.Terms {
			y := NormalizeQuantifiers(x)
			if inner, ok := y.(*logic.And); ok {
				terms = append(terms, inner.Terms...)
			} else {
				terms = append(terms, y)
			}
		}
		return &logic.And{Terms: terms}

	case *logic.Or:
		// Flatten: Or(x, Or(a,b), y) -> Or(x, a, b, y)
		var terms []logic.Expr
		for _, x := range n.Terms {
			y := NormalizeQuantifiers(x)
			if inner, ok := y.(*logic.Or); ok {
				terms = append(terms, inner.Terms...)
			} else {
				terms = append(terms, y)
			}
		}
		return &logic.Or{Terms: terms}

	case *logic.ForAll:
		// ForAll(vars, And(a,b)) -> And(ForAll(vars,a), ForAll(vars,b))
		if inner, ok := n.Body.(*logic.And); ok {
			terms := make([]logic.Expr, len(inner.Terms))
			for i, x := range inner.Terms {
				terms[i] = &logic.ForAll{Variables: n.Variables, Body: x}
			}
			return NormalizeQuantifiers(&logic.And{Terms: terms})
		}
		// Otherwise, restrict variables to those actually free in the body
		body := NormalizeQuantifiers(n.Body)
		fvs := FreeVariables(body)
		var vars []*logic.Variable
		for _, v := range n.Variables {
			if _, ok := fvs[logic.Key(v)]; ok {
				vars = append(vars, v)
			}
		}
		if len(vars) == 0 {
			return body
		}
		return &logic.ForAll{Variables: vars, Body: body}

	case *logic.Exists:
		// Exists(vars, Or(a,b)) -> Or(Exists(vars,a), Exists(vars,b))
		if inner, ok := n.Body.(*logic.Or); ok {
			terms := make([]logic.Expr, len(inner.Terms))
			for i, x := range inner.Terms {
				terms[i] = &logic.Exists{Variables: n.Variables, Body: x}
			}
			return NormalizeQuantifiers(&logic.Or{Terms: terms})
		}
		// Otherwise, restrict variables to those actually free in the body
		body := NormalizeQuantifiers(n.Body)
		fvs := FreeVariables(body)
		var vars []*logic.Variable
		for _, v := range n.Variables {
			if _, ok := fvs[logic.Key(v)]; ok {
				vars = append(vars, v)
			}
		}
		if len(vars) == 0 {
			return body
		}
		return &logic.Exists{Variables: vars, Body: body}
	}

	// Fallback for other node types
	return t
}

// -----------------------------------------------------------------------
// SubstituteApply
// -----------------------------------------------------------------------

// SubstituteApplyFunc is the type for functions used in SubstituteApply.
// Given replacement terms, it produces a result node.
type SubstituteApplyFunc func(terms []logic.Expr) logic.Expr

// SubstituteApply performs second-order substitution: for any key in subs,
// Apply(key, terms...) is replaced by subs[key](terms'...), where terms'
// are recursively substituted. Non-application occurrences of keys are
// NOT substituted.
// Corresponds to Python logic_util.substitute_apply.
func SubstituteApply(t logic.Expr, subs map[logic.NodeKey]SubstituteApplyFunc) logic.Expr {
	if len(subs) == 0 {
		return t
	}
	return substituteApplyRec(t, subs)
}

func substituteApplyRec(t logic.Expr, subs map[logic.NodeKey]SubstituteApplyFunc) logic.Expr {
	switch n := t.(type) {
	case *logic.Variable, *logic.Symbol:
		return t

	case *logic.Apply:
		if fn, ok := subs[logic.Key(n.Func)]; ok {
			newTerms := make([]logic.Expr, len(n.Terms))
			for i, term := range n.Terms {
				newTerms[i] = substituteApplyRec(term, subs)
			}
			return fn(newTerms)
		}
		return substituteApplyChildren(t, subs)

	default:
		return substituteApplyChildren(t, subs)
	}
}

func substituteApplyChildren(t logic.Expr, subs map[logic.NodeKey]SubstituteApplyFunc) logic.Expr {
	switch n := t.(type) {
	case *logic.Apply:
		// Python substitute_apply iterates ALL children (for x in t),
		// including Func, to handle nested Apply whose func is in subs.
		newFunc := substituteApplyRec(n.Func, subs)
		newTerms := make([]logic.Expr, len(n.Terms))
		changed := newFunc != n.Func
		for i, term := range n.Terms {
			nt := substituteApplyRec(term, subs)
			newTerms[i] = nt
			if nt != term {
				changed = true
			}
		}
		if !changed {
			return t
		}
		result, err := logic.NewApply(newFunc, newTerms...)
		if err != nil {
			return t
		}
		return result

	case *logic.Eq:
		t1 := substituteApplyRec(n.T1, subs)
		t2 := substituteApplyRec(n.T2, subs)
		if t1 == n.T1 && t2 == n.T2 {
			return t
		}
		return &logic.Eq{T1: t1, T2: t2}

	case *logic.Ite:
		c := substituteApplyRec(n.Cond, subs)
		th := substituteApplyRec(n.Then, subs)
		el := substituteApplyRec(n.Else, subs)
		return &logic.Ite{Cond: c, Then: th, Else: el}

	case *logic.Not:
		b := substituteApplyRec(n.Body, subs)
		if b == n.Body {
			return t
		}
		return &logic.Not{Body: b}

	case *logic.And:
		terms := make([]logic.Expr, len(n.Terms))
		for i, term := range n.Terms {
			terms[i] = substituteApplyRec(term, subs)
		}
		return &logic.And{Terms: terms}

	case *logic.Or:
		terms := make([]logic.Expr, len(n.Terms))
		for i, term := range n.Terms {
			terms[i] = substituteApplyRec(term, subs)
		}
		return &logic.Or{Terms: terms}

	case *logic.Implies:
		t1 := substituteApplyRec(n.T1, subs)
		t2 := substituteApplyRec(n.T2, subs)
		return &logic.Implies{T1: t1, T2: t2}

	case *logic.Iff:
		t1 := substituteApplyRec(n.T1, subs)
		t2 := substituteApplyRec(n.T2, subs)
		return &logic.Iff{T1: t1, T2: t2}

	case *logic.ForAll:
		// Remove bound vars from subs
		newSubs := filterSubs(subs, n.Variables)
		body := substituteApplyRec(n.Body, newSubs)
		return &logic.ForAll{Variables: n.Variables, Body: body}

	case *logic.Exists:
		newSubs := filterSubs(subs, n.Variables)
		body := substituteApplyRec(n.Body, newSubs)
		return &logic.Exists{Variables: n.Variables, Body: body}

	case *logic.Lambda:
		newSubs := filterSubs(subs, n.Variables)
		body := substituteApplyRec(n.Body, newSubs)
		return &logic.Lambda{Variables: n.Variables, Body: body}

	case *logic.NamedBinder:
		newSubs := filterSubs(subs, n.Variables)
		body := substituteApplyRec(n.Body, newSubs)
		return &logic.NamedBinder{Name: n.Name, Variables: n.Variables, Environ: n.Environ, Body: body}
	}
	return t
}

func filterSubs(subs map[logic.NodeKey]SubstituteApplyFunc, vars []*logic.Variable) map[logic.NodeKey]SubstituteApplyFunc {
	newSubs := make(map[logic.NodeKey]SubstituteApplyFunc, len(subs))
	varSet := make(map[logic.NodeKey]struct{}, len(vars))
	for _, v := range vars {
		varSet[logic.Key(v)] = struct{}{}
	}
	for k, v := range subs {
		if _, bound := varSet[k]; !bound {
			newSubs[k] = v
		}
	}
	return newSubs
}
