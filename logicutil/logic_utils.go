// Additional logic utility functions.
// Ports of functions from Python's ivy_logic_utils.py that are used by
// multiple subsystems.
package logicutil

import (
	"github.com/glycerine/goivy/logic"
)

// --- AST collectors ---

// SortsAst collects all sorts used in an AST.
func SortsAst(ast logic.Node) map[logic.Sort]bool {
	result := make(map[logic.Sort]bool)
	sortsAstRec(ast, result)
	return result
}

func sortsAstRec(ast logic.Node, result map[logic.Sort]bool) {
	s := ast.NodeSort()
	if s != nil {
		result[s] = true
	}
	if v, ok := ast.(*logic.Var); ok {
		result[v.VSort] = true
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
	switch ast.(type) {
	case *logic.Globally, *logic.Eventually, *logic.WhenOperator:
		*result = append(*result, ast)
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
	if nb, ok := ast.(*logic.NamedBinder); ok {
		*result = append(*result, nb)
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
		if len(children) > 0 {
			return &logic.Apply{Func: children[0], Terms: children[1:]}
		}
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
	}
	return n
}
