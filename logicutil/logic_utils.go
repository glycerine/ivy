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
func SortsAst(ast logic.Node) map[logic.NodeKey]logic.Sort {
	result := make(map[logic.NodeKey]logic.Sort)
	sortsAstRec(ast, result)
	return result
}

func sortsAstRec(ast logic.Node, result map[logic.NodeKey]logic.Sort) {
	// Matches Python sorts_ast (ivy_logic_utils.py:570-583):
	// For Apply: yield rep.sort.rng and rep.sort.dom, or recurse into binder body.
	// For Var: yield sort. Then iterate ast.args (Terms only).
	if app, ok := ast.(*logic.Apply); ok {
		if nb, ok := app.Func.(*logic.NamedBinder); ok {
			sortsAstRec(nb.Body, result)
		} else if c, ok := app.Func.(*logic.Const); ok {
			if fs, ok := c.CSort.(*logic.FunctionSort); ok {
				rng := fs.Range()
				result[logic.SortKey(rng)] = rng
				for _, d := range fs.Domain() {
					result[logic.SortKey(d)] = d
				}
			}
		}
	} else if v, ok := ast.(*logic.Var); ok {
		result[logic.SortKey(v.VSort)] = v.VSort
	}
	for _, c := range ast.Children() {
		sortsAstRec(c, result)
	}
}

// RelationsAst collects all relation symbols from an AST.
func RelationsAst(ast logic.Node) []*logic.Const {
	seen := make(map[string]bool)
	var result []*logic.Const
	relationsAstRec(ast, &result, seen)
	return result
}

func relationsAstRec(ast logic.Node, result *[]*logic.Const, seen map[string]bool) {
	if app, ok := ast.(*logic.Apply); ok {
		if c, ok := app.Func.(*logic.Const); ok {
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
func FunctionsAst(ast logic.Node) []*logic.Const {
	seen := make(map[string]bool)
	var result []*logic.Const
	functionsAstRec(ast, &result, seen)
	return result
}

func functionsAstRec(ast logic.Node, result *[]*logic.Const, seen map[string]bool) {
	if app, ok := ast.(*logic.Apply); ok {
		if c, ok := app.Func.(*logic.Const); ok {
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
func EqsAst(ast logic.Node) []*logic.Eq {
	var result []*logic.Eq
	eqsAstRec(ast, &result)
	return result
}

func eqsAstRec(ast logic.Node, result *[]*logic.Eq) {
	if eq, ok := ast.(*logic.Eq); ok {
		*result = append(*result, eq)
	}
	for _, child := range ast.Children() {
		eqsAstRec(child, result)
	}
}

// GroundAppsAst collects all ground (variable-free) application subterms.
func GroundAppsAst(ast logic.Node) []logic.Node {
	var result []logic.Node
	groundAppsAstRec(ast, &result)
	return result
}

func groundAppsAstRec(ast logic.Node, result *[]logic.Node) bool {
	if _, ok := ast.(*logic.Var); ok {
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
func TemporalsAst(ast logic.Node) []logic.Node {
	var result []logic.Node
	temporalsAstRec(ast, &result)
	return result
}

func temporalsAstRec(ast logic.Node, result *[]logic.Node) {
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
func NamedBindersAst(ast logic.Node) []*logic.NamedBinder {
	var result []*logic.NamedBinder
	namedBindersAstRec(ast, &result)
	return result
}

func namedBindersAstRec(ast logic.Node, result *[]*logic.NamedBinder) {
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
func IsEqualityLit(atom logic.Node) bool {
	_, ok := atom.(*logic.Eq)
	return ok
}

// IsTautEqualityLit returns true if the literal is x=x.
func IsTautEqualityLit(atom logic.Node) bool {
	eq, ok := atom.(*logic.Eq)
	if !ok {
		return false
	}
	return eq.T1.Equal(eq.T2)
}

// IsDisequalityLit returns true if the literal is Not(x=y).
func IsDisequalityLit(atom logic.Node) bool {
	neg, ok := atom.(*logic.Not)
	if !ok {
		return false
	}
	_, isEq := neg.Body.(*logic.Eq)
	return isEq
}

// IsGroundLit returns true if the literal contains no variables.
func IsGroundLit(atom logic.Node) bool {
	fv := FreeVariables(atom)
	return len(fv) == 0
}

// IsGroundEqualityLit returns true if the equality literal is ground.
func IsGroundEqualityLit(atom logic.Node) bool {
	if !IsEqualityLit(atom) {
		return false
	}
	return IsGroundLit(atom)
}

// --- Formula / literal conversion ---

// EqLit creates an equality literal (positive).
func EqLit(x, y logic.Node) logic.Node {
	return &logic.Eq{T1: x, T2: y}
}

// NeqLit creates a disequality literal.
func NeqLit(x, y logic.Node) logic.Node {
	return &logic.Not{Body: &logic.Eq{T1: x, T2: y}}
}

// EqAtom creates an equality atom.
func EqAtom(x, y logic.Node) *logic.Eq {
	return &logic.Eq{T1: x, T2: y}
}

// SwapArgsLit swaps the arguments of an equality literal.
func SwapArgsLit(atom logic.Node) logic.Node {
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
func RelInst(rel *logic.Const) logic.Node {
	fs, ok := rel.CSort.(*logic.FunctionSort)
	if !ok {
		return rel
	}
	dom := fs.Domain()
	vars := make([]logic.Node, len(dom))
	for i, s := range dom {
		v, _ := logic.NewVar(varName(i), s)
		vars[i] = v
	}
	app, err := logic.NewApply(rel, vars...)
	if err != nil {
		return rel
	}
	return app
}

// FunInst creates a function instance: f(V0, V1, ...) for a function of given arity.
func FunInst(f *logic.Const) logic.Node {
	return RelInst(f) // same shape
}

// FunEqInst creates a function equality instance: Y = f(V0, V1, ...).
func FunEqInst(f *logic.Const) logic.Node {
	fs, ok := f.CSort.(*logic.FunctionSort)
	if !ok {
		return f
	}
	dom := fs.Domain()
	rng := fs.Range()
	vars := make([]logic.Node, len(dom))
	for i, s := range dom {
		v, _ := logic.NewVar(varName(i), s)
		vars[i] = v
	}
	y, _ := logic.NewVar("Y", rng)
	fapp, err := logic.NewApply(f, vars...)
	if err != nil {
		return f
	}
	return &logic.Eq{T1: y, T2: fapp}
}

// IsRelational returns true if the symbol has a relational sort (Boolean range).
func IsRelational(sym *logic.Const) bool {
	return isBoolSort(sym.CSort) || isBoolRange(sym.CSort)
}

// --- Formula transformations ---

// ExpandAbbrevs expands Iff and Implies into And/Or/Not.
func ExpandAbbrevs(f logic.Node) logic.Node {
	switch t := f.(type) {
	case *logic.Iff:
		a := ExpandAbbrevs(t.T1)
		b := ExpandAbbrevs(t.T2)
		return &logic.And{Terms: []logic.Node{
			&logic.Or{Terms: []logic.Node{&logic.Not{Body: a}, b}},
			&logic.Or{Terms: []logic.Node{a, &logic.Not{Body: b}}},
		}}
	case *logic.Implies:
		a := ExpandAbbrevs(t.T1)
		b := ExpandAbbrevs(t.T2)
		return &logic.Or{Terms: []logic.Node{&logic.Not{Body: a}, b}}
	default:
		children := f.Children()
		if len(children) == 0 {
			return f
		}
		newChildren := make([]logic.Node, len(children))
		changed := false
		for i, c := range children {
			nc := ExpandAbbrevs(c)
			newChildren[i] = nc
			if nc != c {
				changed = true
			}
		}
		if !changed {
			return f
		}
		return cloneNode(f, newChildren)
	}
}

// DeMorgan applies De Morgan's laws to push negation inward.
func DeMorgan(f logic.Node) logic.Node {
	neg, ok := f.(*logic.Not)
	if !ok {
		return f
	}
	switch t := neg.Body.(type) {
	case *logic.And:
		terms := make([]logic.Node, len(t.Terms))
		for i, term := range t.Terms {
			terms[i] = &logic.Not{Body: term}
		}
		return &logic.Or{Terms: terms}
	case *logic.Or:
		terms := make([]logic.Node, len(t.Terms))
		for i, term := range t.Terms {
			terms[i] = &logic.Not{Body: term}
		}
		return &logic.And{Terms: terms}
	case *logic.Not:
		return t.Body
	case *logic.ForAll:
		return &logic.Exists{Variables: t.Variables, Body: &logic.Not{Body: t.Body}}
	case *logic.Exists:
		return &logic.ForAll{Variables: t.Variables, Body: &logic.Not{Body: t.Body}}
	}
	return f
}

// BooleanConstant creates a boolean constant (true or false).
func BooleanConstant(val bool) logic.Node {
	if val {
		return &logic.And{} // empty And = true
	}
	return &logic.Or{} // empty Or = false
}

// CloseEPR converts formula to forall Y. fmla, where Y are the free variables.
// For conjunctions, applies recursively to each conjunct.
// Corresponds to Python's close_epr.
func CloseEPR(fmla logic.Node) logic.Node {
	if and, ok := fmla.(*logic.And); ok {
		terms := make([]logic.Node, len(and.Terms))
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
func ResortSort(s logic.Sort, subs map[string]logic.Sort) logic.Sort {
	name := s.String()
	if mapped, ok := subs[name]; ok {
		return mapped
	}
	return s
}

// ResortAst remaps all sorts in an AST through a substitution.
func ResortAst(ast logic.Node, subs map[string]logic.Sort) logic.Node {
	switch t := ast.(type) {
	case *logic.Var:
		newSort := ResortSort(t.VSort, subs)
		if newSort != t.VSort {
			v, _ := logic.NewVar(t.Name, newSort)
			return v
		}
		return t
	case *logic.Const:
		newSort := ResortSort(t.CSort, subs)
		if newSort != t.CSort {
			return logic.NewConst(t.Name, newSort)
		}
		return t
	default:
		children := ast.Children()
		if len(children) == 0 {
			return ast
		}
		newChildren := make([]logic.Node, len(children))
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
	return "V" + string(rune('0'+idx))
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

func isQuantifier(n logic.Node) bool {
	switch n.(type) {
	case *logic.ForAll, *logic.Exists:
		return true
	}
	return false
}

func isApp(n logic.Node) bool {
	switch n.(type) {
	case *logic.Apply, *logic.Const:
		return true
	}
	return false
}

// cloneNode creates a clone of a node with new children.
func cloneNode(n logic.Node, children []logic.Node) logic.Node {
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
func SubstituteByName(ast logic.Node, subs map[string]logic.Node) logic.Node {
	if len(subs) == 0 {
		return ast
	}
	return substituteByNameRec(ast, subs)
}

func substituteByNameRec(ast logic.Node, subs map[string]logic.Node) logic.Node {
	if v, ok := ast.(*logic.Var); ok {
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
	newChildren := make([]logic.Node, len(children))
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

func removeBoundNames(subs map[string]logic.Node, vars []*logic.Var) map[string]logic.Node {
	newsubs := make(map[string]logic.Node, len(subs))
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
func freeVariablesInOrder(ast logic.Node) []*logic.Var {
	seen := make(map[string]bool)
	var result []*logic.Var
	freeVariablesInOrderRec(ast, &result, seen, nil)
	return result
}

func freeVariablesInOrderRec(ast logic.Node, result *[]*logic.Var, seen map[string]bool, bound map[string]bool) {
	if v, ok := ast.(*logic.Var); ok {
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
func usedVariablesInOrder(ast logic.Node) []*logic.Var {
	seen := make(map[string]bool)
	var result []*logic.Var
	usedVariablesInOrderRec(ast, &result, seen)
	return result
}

func usedVariablesInOrderRec(ast logic.Node, result *[]*logic.Var, seen map[string]bool) {
	if v, ok := ast.(*logic.Var); ok {
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
func usedVariablesInOrderMulti(asts []logic.Node) []*logic.Var {
	seen := make(map[string]bool)
	var result []*logic.Var
	for _, ast := range asts {
		usedVariablesInOrderRec(ast, &result, seen)
	}
	return result
}

// --- NormalizeFreeVariables ---

// NormalizeFreeVariables normalizes free variables: renames them V0, V1, ...
// in the order they appear.
// Returns (old_vars, new_vars, normalized_ast).
func NormalizeFreeVariables(ast logic.Node) ([]*logic.Var, []*logic.Var, logic.Node) {
	subs := make(map[string]logic.Node)
	var vs []*logic.Var
	var nvs []*logic.Var
	for _, v := range freeVariablesInOrder(ast) {
		if _, exists := subs[v.Name]; !exists {
			nv, _ := logic.NewVar(fmt.Sprintf("V%d", len(subs)), v.VSort)
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
func NormalizeFreeVariablesTuple(asts ...logic.Node) ([]*logic.Var, []*logic.Var, []logic.Node) {
	subs := make(map[string]logic.Node)
	var vs []*logic.Var
	var nvs []*logic.Var
	for _, v := range usedVariablesInOrderMulti(asts) {
		if _, exists := subs[v.Name]; !exists {
			nv, _ := logic.NewVar(fmt.Sprintf("V%d", len(subs)), v.VSort)
			subs[v.Name] = nv
			vs = append(vs, v)
			nvs = append(nvs, nv)
		}
	}
	nasts := make([]logic.Node, len(asts))
	for i, ast := range asts {
		nasts[i] = SubstituteByName(ast, subs)
	}
	return vs, nvs, nasts
}

// --- NormalizeNamedBinders ---

// NormalizeNamedBinders normalizes the variables bound by named binders.
// If names is nil, all named binders are normalized; otherwise only those
// whose name is in the names set.
func NormalizeNamedBinders(ast logic.Node, names map[string]bool) logic.Node {
	if _, ok := ast.(*logic.Const); ok {
		return ast
	}
	if nb, ok := ast.(*logic.NamedBinder); ok {
		if names == nil || names[nb.Name] {
			nvs := make([]*logic.Var, len(nb.Variables))
			subs := make(map[string]logic.Node, len(nb.Variables))
			for i, v := range nb.Variables {
				nv, _ := logic.NewVar(fmt.Sprintf("V%d", i), v.VSort)
				nvs[i] = nv
				subs[v.Name] = nv
			}
			body := NormalizeNamedBinders(nb.Body, names)
			body = SubstituteByName(body, subs)
			return &logic.NamedBinder{Name: nb.Name, Variables: nvs, Environ: nb.Environ, Body: body}
		}
	}
	// For Apply nodes, normalize the func part too
	if app, ok := ast.(*logic.Apply); ok {
		newFunc := NormalizeNamedBinders(app.Func, names)
		newTerms := make([]logic.Node, len(app.Terms))
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
	newChildren := make([]logic.Node, len(children))
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
type GloballyBinderFunc func(vars []*logic.Var, body logic.Node, environ *string) *logic.NamedBinder

// WhenBinderFunc creates a named binder from a WhenOperator.
type WhenBinderFunc func(name string, vars []*logic.Var, body logic.Node) *logic.NamedBinder

// DefaultGloballyBinder creates a NamedBinder named "g" (or "g[env]" if env != nil).
func DefaultGloballyBinder(vars []*logic.Var, body logic.Node, environ *string) *logic.NamedBinder {
	name := "g"
	if environ != nil {
		name = "g[" + *environ + "]"
	}
	return &logic.NamedBinder{Name: name, Variables: vars, Environ: nil, Body: body}
}

// DefaultWhenBinder creates a NamedBinder named "l2s_when<name>".
func DefaultWhenBinder(name string, vars []*logic.Var, body logic.Node) *logic.NamedBinder {
	return &logic.NamedBinder{Name: "l2s_when" + name, Variables: vars, Environ: nil, Body: body}
}

// applyNamedBinder applies a named binder to arguments, returning Apply if
// there are arguments or the binder itself if there are none.
func applyNamedBinder(nb *logic.NamedBinder, args []logic.Node) logic.Node {
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
func ReplaceTemporalsByNamedBinder(ast logic.Node, g GloballyBinderFunc, when WhenBinderFunc) logic.Node {
	if g == nil {
		g = DefaultGloballyBinder
	}
	if when == nil {
		when = DefaultWhenBinder
	}
	return replaceTemporalsRec(ast, g, when)
}

func replaceTemporalsRec(ast logic.Node, g GloballyBinderFunc, when WhenBinderFunc) logic.Node {
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
			newArgs := make([]logic.Node, len(t.Terms))
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
		newTerms := make([]logic.Node, len(t.Terms))
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
	newChildren := make([]logic.Node, len(children))
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

func varsToNodes(vars []*logic.Var) []logic.Node {
	nodes := make([]logic.Node, len(vars))
	for i, v := range vars {
		nodes[i] = v
	}
	return nodes
}

// --- ReduceNamedBinders ---

// ReduceNamedBinders reduces subexpressions of the form ($b V1...Vn. t)(s1...sn)
// to ($b . t[Vi/si]).
func ReduceNamedBinders(ast logic.Node, g GloballyBinderFunc) logic.Node {
	if g == nil {
		g = DefaultGloballyBinder
	}
	return reduceNamedBindersRec(ast, g)
}

func reduceNamedBindersRec(ast logic.Node, g GloballyBinderFunc) logic.Node {
	if _, ok := ast.(*logic.Const); ok {
		return ast
	}
	if app, ok := ast.(*logic.Apply); ok {
		if nb, ok := app.Func.(*logic.NamedBinder); ok {
			subst := make(map[string]logic.Node, len(nb.Variables))
			for i, v := range nb.Variables {
				if i < len(app.Terms) {
					subst[v.Name] = app.Terms[i]
				}
			}
			body := reduceNamedBindersRec(SubstituteByName(nb.Body, subst), g)
			return &logic.NamedBinder{Name: nb.Name, Variables: nil, Environ: nil, Body: body}
		}
		newFunc := NormalizeNamedBinders(app.Func, nil)
		newTerms := make([]logic.Node, len(app.Terms))
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
	newChildren := make([]logic.Node, len(children))
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
func ReplaceNamedBindersAst(ast logic.Node, subs map[string]logic.Node) logic.Node {
	if nb, ok := ast.(*logic.NamedBinder); ok {
		key := nb.String()
		if rep, found := subs[key]; found {
			return rep
		}
		return ast
	}
	if app, ok := ast.(*logic.Apply); ok {
		newTerms := make([]logic.Node, len(app.Terms))
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
	newChildren := make([]logic.Node, len(children))
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
func ExpandNamedBindersAst(ast logic.Node, fun func(*logic.NamedBinder) logic.Node) logic.Node {
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
	newChildren := make([]logic.Node, len(children))
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
func IsTrue(n logic.Node) bool {
	a, ok := n.(*logic.And)
	return ok && len(a.Terms) == 0
}

// IsFalse returns true if the node is the logical false constant (empty Or).
func IsFalse(n logic.Node) bool {
	o, ok := n.(*logic.Or)
	return ok && len(o.Terms) == 0
}

// DenormalizeTemporal reverses temporal normalization:
//   - Not(Globally(Not(x))) => Eventually(x)
//   - Iff/Eq(Globally(Not(x)), true) => Iff/Eq(Eventually(x), Or())
//   - Iff/Eq(Globally(Not(x)), false) => Iff/Eq(Eventually(x), And())
func DenormalizeTemporal(ast logic.Node) logic.Node {
	children := ast.Children()
	newChildren := make([]logic.Node, len(children))
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
						return cloneNode(ast, []logic.Node{ev, &logic.Or{}})
					}
					if IsFalse(rhs) {
						return cloneNode(ast, []logic.Node{ev, &logic.And{}})
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
func IsVacEqualityLit(lit logic.Node) bool {
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
func IsTrueLit(lit logic.Node) bool {
	if neg, ok := lit.(*logic.Not); ok {
		return IsFalse(neg.Body)
	}
	return IsTrue(lit)
}

// IsFalseLit returns true if the literal evaluates to false.
// A positive literal that is Or() (false), or a Not(And()) (not true).
func IsFalseLit(lit logic.Node) bool {
	if neg, ok := lit.(*logic.Not); ok {
		return IsTrue(neg.Body)
	}
	return IsFalse(lit)
}

// IsTautLit returns true if the literal is a tautology (always true).
func IsTautLit(lit logic.Node) bool {
	return IsTrueLit(lit) || IsTautEqualityLit(lit)
}

// IsVacLit returns true if the literal is vacuously true (always false).
func IsVacLit(lit logic.Node) bool {
	return IsFalseLit(lit) || IsVacEqualityLit(lit)
}

// --- TermListsEq ---

// TermListsEq returns true if two node slices are structurally equal.
func TermListsEq(l1, l2 []logic.Node) bool {
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
	Clauses [][]logic.Node // accumulated definition clauses
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
func (tc *TseitinContext) AddDefs(cls [][]logic.Node) [][]logic.Node {
	result := make([][]logic.Node, 0, len(tc.Clauses)+len(cls))
	result = append(result, cls...)
	result = append(result, tc.Clauses...)
	return result
}

// TseitinEncoding performs Tseitin transformation on a formula.
// Returns a node representing the formula as a literal.
// Side effect: adds definition clauses to the TseitinContext.
func TseitinEncoding(tc *TseitinContext, f logic.Node) logic.Node {
	f = ExpandAbbrevs(f)

	if and, ok := f.(*logic.And); ok {
		args := make([]logic.Node, len(and.Terms))
		for i, t := range and.Terms {
			args[i] = TseitinEncoding(tc, t)
		}
		// Collect variables in order from args
		varsSeen := make(map[string]bool)
		var vs []*logic.Var
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
		fn := logic.NewConst(fname, fnSort)
		var res logic.Node
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
			tc.Clauses = append(tc.Clauses, []logic.Node{&logic.Not{Body: res}, arg})
		}
		lastClause := []logic.Node{res}
		for _, arg := range args {
			lastClause = append(lastClause, &logic.Not{Body: arg})
		}
		tc.Clauses = append(tc.Clauses, lastClause)
		return res
	}

	if or, ok := f.(*logic.Or); ok {
		negTerms := make([]logic.Node, len(or.Terms))
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
func FormulaToLit(tc *TseitinContext, f logic.Node) logic.Node {
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
func isAtomNode(n logic.Node) bool {
	if _, ok := n.(*logic.Eq); ok {
		return true
	}
	switch n.(type) {
	case *logic.Apply:
		return logic.SortEqual(n.NodeSort(), logic.Boolean)
	case *logic.Const:
		return logic.SortEqual(n.NodeSort(), logic.Boolean)
	}
	return false
}

// --- FormulaToClause ---

// FormulaToClause converts a formula to a clause (disjunction of literals).
// Returns a slice of literal nodes.
func FormulaToClause(tc *TseitinContext, f logic.Node) []logic.Node {
	f = ExpandAbbrevs(f)
	f = DeMorgan(f)
	if IsTrue(f) {
		return []logic.Node{f}
	}
	if or, ok := f.(*logic.Or); ok {
		var result []logic.Node
		for _, t := range or.Terms {
			result = append(result, FormulaToClause(tc, t)...)
		}
		return result
	}
	return []logic.Node{FormulaToLit(tc, f)}
}

// --- FormulaToClube ---

// FormulaToClube converts a formula to a cube (conjunction of literals).
// Returns a slice of literal nodes.
func FormulaToClube(tc *TseitinContext, f logic.Node) []logic.Node {
	f = ExpandAbbrevs(f)
	if not, ok := f.(*logic.Not); ok {
		clause := FormulaToClause(tc, not.Body)
		result := make([]logic.Node, len(clause))
		for i, lit := range clause {
			result[i] = &logic.Not{Body: lit}
		}
		return result
	}
	if and, ok := f.(*logic.And); ok {
		var result []logic.Node
		for _, t := range and.Terms {
			result = append(result, FormulaToLit(tc, t))
		}
		return result
	}
	return []logic.Node{FormulaToLit(tc, f)}
}

// --- ReduceNumerically ---

// ReduceNumerically evaluates numeric expressions where all arguments
// are numeral constants. Handles <, =, and .succ operators.
func ReduceNumerically(ast logic.Node) logic.Node {
	children := ast.Children()
	newChildren := make([]logic.Node, len(children))
	for i, c := range children {
		newChildren[i] = ReduceNumerically(c)
	}

	if app, ok := ast.(*logic.Apply); ok {
		allNumeral := len(app.Terms) > 0
		for i := range app.Terms {
			nc := newChildren[i] // Children() = Terms only (Func excluded)
			c, ok := nc.(*logic.Const)
			if !ok || !isAllDigits(c.Name) {
				allNumeral = false
				break
			}
		}
		if allNumeral {
			vals := make([]int, len(app.Terms))
			for i := range app.Terms {
				nc := newChildren[i]
				c := nc.(*logic.Const)
				vals[i], _ = strconv.Atoi(c.Name)
			}
			if fn, ok := app.Func.(*logic.Const); ok {
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
func NormalizeQuantifiers(t logic.Node) logic.Node {
	switch n := t.(type) {
	case *logic.Var, *logic.Const:
		return t

	case *logic.Apply:
		newTerms := make([]logic.Node, len(n.Terms))
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
		var terms []logic.Node
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
		var terms []logic.Node
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
			terms := make([]logic.Node, len(inner.Terms))
			for i, x := range inner.Terms {
				terms[i] = &logic.ForAll{Variables: n.Variables, Body: x}
			}
			return NormalizeQuantifiers(&logic.And{Terms: terms})
		}
		// Otherwise, restrict variables to those actually free in the body
		body := NormalizeQuantifiers(n.Body)
		fvs := FreeVariables(body)
		var vars []*logic.Var
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
			terms := make([]logic.Node, len(inner.Terms))
			for i, x := range inner.Terms {
				terms[i] = &logic.Exists{Variables: n.Variables, Body: x}
			}
			return NormalizeQuantifiers(&logic.Or{Terms: terms})
		}
		// Otherwise, restrict variables to those actually free in the body
		body := NormalizeQuantifiers(n.Body)
		fvs := FreeVariables(body)
		var vars []*logic.Var
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
type SubstituteApplyFunc func(terms []logic.Node) logic.Node

// SubstituteApply performs second-order substitution: for any key in subs,
// Apply(key, terms...) is replaced by subs[key](terms'...), where terms'
// are recursively substituted. Non-application occurrences of keys are
// NOT substituted.
// Corresponds to Python logic_util.substitute_apply.
func SubstituteApply(t logic.Node, subs map[logic.NodeKey]SubstituteApplyFunc) logic.Node {
	if len(subs) == 0 {
		return t
	}
	return substituteApplyRec(t, subs)
}

func substituteApplyRec(t logic.Node, subs map[logic.NodeKey]SubstituteApplyFunc) logic.Node {
	switch n := t.(type) {
	case *logic.Var, *logic.Const:
		return t

	case *logic.Apply:
		if fn, ok := subs[logic.Key(n.Func)]; ok {
			newTerms := make([]logic.Node, len(n.Terms))
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

func substituteApplyChildren(t logic.Node, subs map[logic.NodeKey]SubstituteApplyFunc) logic.Node {
	switch n := t.(type) {
	case *logic.Apply:
		// Python substitute_apply iterates ALL children (for x in t),
		// including Func, to handle nested Apply whose func is in subs.
		newFunc := substituteApplyRec(n.Func, subs)
		newTerms := make([]logic.Node, len(n.Terms))
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
		terms := make([]logic.Node, len(n.Terms))
		for i, term := range n.Terms {
			terms[i] = substituteApplyRec(term, subs)
		}
		return &logic.And{Terms: terms}

	case *logic.Or:
		terms := make([]logic.Node, len(n.Terms))
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

func filterSubs(subs map[logic.NodeKey]SubstituteApplyFunc, vars []*logic.Var) map[logic.NodeKey]SubstituteApplyFunc {
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
