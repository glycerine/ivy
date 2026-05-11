// Additional logic utility functions.
// Ports of functions from Python's ivy_logic_utils.py that are used by
// multiple subsystems.
package goivy

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// --- AST collectors ---

// SortsAst collects all sorts used in an AST.
// Returns map keyed by SortKey (structural identity) with Sort values.
func SortsAst(ast Expr) map[NodeKey]Sort {
	result := make(map[NodeKey]Sort)
	sortsAstRec(ast, result)
	return result
}

func sortsAstRec(ast Expr, result map[NodeKey]Sort) {
	// Matches Python sorts_ast (ivy_logic_utils.py:632-645):
	// For is_app(ast) [App OR bare Const OR 0-arity NamedBinder]:
	//   if is_binder(rep): recurse into rep.body
	//   else: yield rep.sort.rng and rep.sort.dom
	// For Variable: yield sort. Then iterate ast.args.
	//
	// Python is_app(Const) == True means a bare Symbol also yields
	// its sort's rng/dom. For non-FunctionSort (BooleanSort,
	// UninterpretedSort, EnumeratedSort, ConstantSort), Python's
	// rng property returns self and dom returns []; we mirror that
	// by yielding the sort itself.
	yieldFuncSort := func(s Sort) {
		if fs, ok := s.(*LogicFunctionSort); ok {
			rng := fs.Range()
			result[SortKey(rng)] = rng
			for _, d := range fs.Domain() {
				result[SortKey(d)] = d
			}
		} else {
			result[SortKey(s)] = s
		}
	}
	if app, ok := ast.(*Apply); ok {
		if nb, ok := app.Func.(*LogicNamedBinder); ok {
			sortsAstRec(nb.Body, result)
		} else if c, ok := app.Func.(*Const); ok {
			yieldFuncSort(c.CSort)
		}
	} else if c, ok := ast.(*Const); ok {
		// Python: is_app(Const) == True → yield rep.sort.rng/rep.sort.dom.
		yieldFuncSort(c.CSort)
	} else if nb, ok := ast.(*LogicNamedBinder); ok && len(nb.Variables) == 0 {
		// Python: is_app(NamedBinder with no vars) == True → is_binder path,
		// recurse into body (not the binder's .sort).
		sortsAstRec(nb.Body, result)
	} else if v, ok := ast.(*LogicVariable); ok {
		result[SortKey(v.VSort)] = v.VSort
	}
	for _, c := range ast.Children() {
		sortsAstRec(c, result)
	}
}

// RelationsAst collects all relation symbols from an AST.
func RelationsAst(ast Expr) []*Const {
	seen := make(map[string]bool)
	var result []*Const
	relationsAstRec(ast, &result, seen)
	return result
}

func relationsAstRec(ast Expr, result *[]*Const, seen map[string]bool) {
	if app, ok := ast.(*Apply); ok {
		if c, ok := app.Func.(*Const); ok {
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
func FunctionsAst(ast Expr) []*Const {
	seen := make(map[string]bool)
	var result []*Const
	functionsAstRec(ast, &result, seen)
	return result
}

func functionsAstRec(ast Expr, result *[]*Const, seen map[string]bool) {
	if app, ok := ast.(*Apply); ok {
		if c, ok := app.Func.(*Const); ok {
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
func EqsAst(ast Expr) []*Eq {
	var result []*Eq
	eqsAstRec(ast, &result)
	return result
}

func eqsAstRec(ast Expr, result *[]*Eq) {
	if eq, ok := ast.(*Eq); ok {
		*result = append(*result, eq)
	}
	for _, child := range ast.Children() {
		eqsAstRec(child, result)
	}
}

// GroundAppsAst collects all ground (variable-free) application subterms.
func GroundAppsAst(ast Expr) []Expr {
	var result []Expr
	groundAppsAstRec(ast, &result)
	return result
}

func groundAppsAstRec(ast Expr, result *[]Expr) bool {
	if _, ok := ast.(*LogicVariable); ok {
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
func TemporalsAst(ast Expr) []Expr {
	var result []Expr
	temporalsAstRec(ast, &result)
	return result
}

func temporalsAstRec(ast Expr, result *[]Expr) {
	// Matches Python temporals_ast (ivy_logic_utils.py:559-568):
	// yields temporal ops, or recurses into Apply rep body if NamedBinder.
	switch ast.(type) {
	case *LogicGlobally, *LogicEventually, *LogicWhenOperator:
		*result = append(*result, ast)
	default:
		if app, ok := ast.(*Apply); ok {
			if nb, ok := app.Func.(*LogicNamedBinder); ok {
				temporalsAstRec(nb.Body, result)
			}
		}
	}
	for _, c := range ast.Children() {
		temporalsAstRec(c, result)
	}
}

// NamedBindersAst collects all named binders from an AST.
func NamedBindersAst(ast Expr) []*LogicNamedBinder {
	var result []*LogicNamedBinder
	namedBindersAstRec(ast, &result)
	return result
}

func namedBindersAstRec(ast Expr, result *[]*LogicNamedBinder) {
	// Matches Python named_binders_ast (ivy_logic_utils.py:547-557):
	// yields standalone NamedBinder, or Apply with NamedBinder rep.
	if nb, ok := ast.(*LogicNamedBinder); ok {
		*result = append(*result, nb)
	} else if app, ok := ast.(*Apply); ok {
		if nb, ok := app.Func.(*LogicNamedBinder); ok {
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
func IsEqualityLit(atom Expr) bool {
	_, ok := atom.(*Eq)
	return ok
}

// IsTautEqualityLit returns true if the literal is x=x.
func IsTautEqualityLit(atom Expr) bool {
	eq, ok := atom.(*Eq)
	if !ok {
		return false
	}
	return eq.T1.Equal(eq.T2)
}

// IsDisequalityLit returns true if the literal is Not(x=y).
func IsDisequalityLit(atom Expr) bool {
	neg, ok := atom.(*LogicNot)
	if !ok {
		return false
	}
	_, isEq := neg.Body.(*Eq)
	return isEq
}

// IsGroundLit returns true if the literal contains no variables.
func IsGroundLit(atom Expr) bool {
	fv := FreeVariables(atom)
	return fv.Len() == 0
}

// IsGroundEqualityLit returns true if the equality literal is ground.
func IsGroundEqualityLit(atom Expr) bool {
	if !IsEqualityLit(atom) {
		return false
	}
	return IsGroundLit(atom)
}

// --- Formula / literal conversion ---

// EqLit creates an equality literal (positive).
func LogicUtilEqLit(x, y Expr) Expr {
	return &Eq{T1: x, T2: y}
}

// NeqLit creates a disequality literal.
func LogicUtilNeqLit(x, y Expr) Expr {
	return &LogicNot{Body: &Eq{T1: x, T2: y}}
}

// EqAtom creates an equality atom.
func LogicUtilEqAtom(x, y Expr) *Eq {
	return &Eq{T1: x, T2: y}
}

// SwapArgsLit swaps the arguments of an equality literal.
func SwapArgsLit(atom Expr) Expr {
	if eq, ok := atom.(*Eq); ok {
		return &Eq{T1: eq.T2, T2: eq.T1}
	}
	if neg, ok := atom.(*LogicNot); ok {
		if eq, ok := neg.Body.(*Eq); ok {
			return &LogicNot{Body: &Eq{T1: eq.T2, T2: eq.T1}}
		}
	}
	return atom
}

// RelInst creates a relational instance: r(V0, V1, ...) for a relation of given arity.
func RelInst(rel *Const) Expr {
	fs, ok := rel.CSort.(*LogicFunctionSort)
	if !ok {
		return rel
	}
	dom := fs.Domain()
	vars := make([]Expr, len(dom))
	for i, s := range dom {
		v, _ := NewVariable(logicutilVarName(i), s)
		vars[i] = v
	}
	app, err := NewApply(rel, vars...)
	if err != nil {
		return rel
	}
	return app
}

// FunInst creates a function instance: f(V0, V1, ...) for a function of given arity.
func FunInst(f *Const) Expr {
	return RelInst(f) // same shape
}

// FunEqInst creates a function equality instance: Y = f(V0, V1, ...).
func FunEqInst(f *Const) Expr {
	fs, ok := f.CSort.(*LogicFunctionSort)
	if !ok {
		return f
	}
	dom := fs.Domain()
	rng := fs.Range()
	vars := make([]Expr, len(dom))
	for i, s := range dom {
		v, _ := NewVariable(logicutilVarName(i), s)
		vars[i] = v
	}
	y, _ := NewVariable("Y", rng)
	fapp, err := NewApply(f, vars...)
	if err != nil {
		return f
	}
	return &Eq{T1: y, T2: fapp}
}

// IsRelational returns true if the symbol has a relational sort (Boolean range).
func IsRelational(sym *Const) bool {
	return isBoolSort(sym.CSort) || isBoolRange(sym.CSort)
}

// --- Formula transformations ---

// ExpandAbbrevs expands Iff, Implies, and Ite into And/Or/Not.
// Top-level only — does not recurse into children.
// Matches Python ivy_logic_utils.py expand_abbrevs (lines 881-900).
func ExpandAbbrevs(f Expr) Expr {
	switch t := f.(type) {
	case *LogicImplies:
		return &LogicOr{Terms: []Expr{negate(t.T1), t.T2}}
	case *LogicIff:
		// Special case: Iff(lhs, Ite(...)) — distribute Iff over Ite branches
		if ite, ok := t.T2.(*LogicIte); ok {
			thenp := ExpandAbbrevs(&LogicIff{T1: t.T1, T2: ite.Then})
			elsep := ExpandAbbrevs(&LogicIff{T1: t.T1, T2: ite.Else})
			return ExpandAbbrevs(&LogicIte{Cond: ite.Cond, Then: thenp, Else: elsep})
		}
		// Note: Python has is_true/is_false shortcuts here but they are
		// dead code (missing return statements). We faithfully omit them.
		return &LogicAnd{Terms: []Expr{
			&LogicOr{Terms: []Expr{negate(t.T1), t.T2}},
			&LogicOr{Terms: []Expr{negate(t.T2), t.T1}},
		}}
	case *LogicIte:
		thenps := conditionConj(negate(t.Cond), t.Then)
		elseps := conditionConj(t.Cond, t.Else)
		all := append(thenps, elseps...)
		return &LogicAnd{Terms: all}
	}
	return f
}

// negate returns Not(f), or unwraps double negation.
// Matches Python ivy_logic_utils.py negate (lines 926-929).
func negate(f Expr) Expr {
	if n, ok := f.(*LogicNot); ok {
		return n.Body
	}
	return &LogicNot{Body: f}
}

// conditionConj returns [Or(c,q) for q in p.args] if p is And, else [Or(c,p)].
// Matches Python ivy_logic_utils.py condition_conj (lines 877-879).
func conditionConj(c, p Expr) []Expr {
	var ps []Expr
	if a, ok := p.(*LogicAnd); ok {
		ps = a.Terms
	} else {
		ps = []Expr{p}
	}
	result := make([]Expr, len(ps))
	for i, q := range ps {
		result[i] = &LogicOr{Terms: []Expr{c, q}}
	}
	return result
}

// DeMorgan applies De Morgan's laws to push negation inward.
// Matches Python ivy_logic_utils.py de_morgan (lines 931-943).
func DeMorgan(f Expr) Expr {
	f = ExpandAbbrevs(f)
	neg, ok := f.(*LogicNot)
	if !ok {
		return f
	}
	g := DeMorgan(neg.Body)
	switch t := g.(type) {
	case *LogicAnd:
		if len(t.Terms) == 1 {
			return DeMorgan(negate(t.Terms[0]))
		}
		terms := make([]Expr, len(t.Terms))
		for i, term := range t.Terms {
			terms[i] = negate(term)
		}
		return &LogicOr{Terms: terms}
	case *LogicOr:
		if len(t.Terms) == 1 {
			return DeMorgan(negate(t.Terms[0]))
		}
		terms := make([]Expr, len(t.Terms))
		for i, term := range t.Terms {
			terms[i] = negate(term)
		}
		return &LogicAnd{Terms: terms}
	}
	return &LogicNot{Body: g}
}

// BooleanConstant creates a boolean constant (true or false).
func BooleanConstant(val bool) Expr {
	if val {
		return &LogicAnd{} // empty And = true
	}
	return &LogicOr{} // empty Or = false
}

// CloseEPR converts formula to forall Y. fmla, where Y are the free variables.
// For conjunctions, applies recursively to each conjunct.
//
// Put another way:
// CloseEPR distributes ForAll over And conjuncts, universally
// quantifying each conjunct's free variables independently.
//
// Corresponds to Python close_epr in ivy_logic_utils.py.
func CloseEPR(fmla Expr) Expr {
	xtraceParts("logicutil.CloseEPR HASH canon= fmla=", canonPart(fmla))
	if and, ok := fmla.(*LogicAnd); ok {
		terms := make([]Expr, len(and.Terms))
		for i, t := range and.Terms {
			terms[i] = CloseEPR(t)
		}
		return &LogicAnd{Terms: terms}
	}
	fvs := FreeVariablesList(fmla)
	if len(fvs) == 0 {
		return fmla
	}
	return &ForAll{Variables: fvs, Body: fmla}
}

// --- Resort functions ---

// ResortSort remaps a sort through a substitution.
// The subs map is keyed by SortKey (Sexp-based structural identity),
// matching Python's structural equality on Sort objects.
func LogicUtilResortSort(s Sort, subs map[NodeKey]Sort) Sort {
	// Direct substitution
	key := SortKey(s)
	if mapped, ok := subs[key]; ok {
		return mapped
	}
	// Recurse into FunctionSort domain/range
	// Matches Python ivy_logic_utils.py:398-410 resort_sort
	if fs, ok := s.(*LogicFunctionSort); ok {
		allSorts := make([]Sort, len(fs.Sorts))
		for i, sub := range fs.Sorts {
			allSorts[i] = LogicUtilResortSort(sub, subs)
		}
		result, err := NewFunctionSort(allSorts...)
		if err != nil {
			return &LogicFunctionSort{Sorts: allSorts}
		}
		return result
	}
	return s
}

// ResortSymbol returns a new Symbol with its sort remapped through subs.
// Matches Python ivy_logic_utils.py resort_symbol (lines 412-413).
func LogicUtilResortSymbol(sym *Const, subs map[NodeKey]Sort) *Const {
	newSort := LogicUtilResortSort(sym.CSort, subs)
	if newSort == sym.CSort {
		return sym
	}
	return NewConst(sym.Name, newSort)
}

// ResortAst remaps all sorts in an AST through a substitution.
func ResortAst(ast Expr, subs map[NodeKey]Sort) Expr {
	switch t := ast.(type) {
	case *LogicVariable:
		newSort := LogicUtilResortSort(t.VSort, subs)
		if newSort != t.VSort {
			v, _ := NewVariable(t.Name, newSort)
			return v
		}
		return t
	case *Const:
		return LogicUtilResortSymbol(t, subs)
	case *Apply:
		// Python: resort_symbol(ast.rep)(*args) — must resort the Func too
		newFunc := ResortAst(t.Func, subs)
		newTerms := make([]Expr, len(t.Terms))
		for i, term := range t.Terms {
			newTerms[i] = ResortAst(term, subs)
		}
		return MustApply(newFunc, newTerms...)
	default:
		children := ast.Children()
		if len(children) == 0 {
			return ast
		}
		newChildren := make([]Expr, len(children))
		for i, c := range children {
			newChildren[i] = ResortAst(c, subs)
		}
		return cloneNode(ast, newChildren)
	}
}

// --- helpers ---

func logicutilVarName(idx int) string {
	return fmt.Sprintf("V%d", idx)
}

func isBoolSort(s Sort) bool {
	return SortEqual(s, Boolean)
}

func isBoolRange(s Sort) bool {
	if fs, ok := s.(*LogicFunctionSort); ok {
		return SortEqual(fs.Range(), Boolean)
	}
	return false
}

func isQuantifier(n Expr) bool {
	switch n.(type) {
	case *ForAll, *LogicExists:
		return true
	}
	return false
}

func isApp(n Expr) bool {
	switch n.(type) {
	case *Apply, *Const:
		return true
	}
	return false
}

// cloneNode creates a clone of a node with new children.
func cloneNode(n Expr, children []Expr) Expr {
	switch t := n.(type) {
	case *LogicNot:
		if len(children) >= 1 {
			return &LogicNot{Body: children[0]}
		}
	case *LogicAnd:
		return &LogicAnd{Terms: children}
	case *LogicOr:
		return &LogicOr{Terms: children}
	case *LogicImplies:
		if len(children) >= 2 {
			return &LogicImplies{T1: children[0], T2: children[1]}
		}
	case *LogicIff:
		if len(children) >= 2 {
			return &LogicIff{T1: children[0], T2: children[1]}
		}
	case *Eq:
		if len(children) >= 2 {
			return &Eq{T1: children[0], T2: children[1]}
		}
	case *LogicIte:
		if len(children) >= 3 {
			return &LogicIte{ISort: t.ISort, Cond: children[0], Then: children[1], Else: children[2]}
		}
	case *Apply:
		// Matches Python Apply.clone(args) = Apply(self.func, *args).
		// Preserves Func and aSort; skips sort validation to match Python.
		return CloneApplyTerms(t, children)
	case *ForAll:
		if len(children) >= 1 {
			return &ForAll{Variables: t.Variables, Body: children[0]}
		}
	case *LogicExists:
		if len(children) >= 1 {
			return &LogicExists{Variables: t.Variables, Body: children[0]}
		}
	case *Lambda:
		if len(children) >= 1 {
			return &Lambda{Variables: t.Variables, Body: children[0]}
		}
	case *LogicGlobally:
		if len(children) >= 1 {
			return &LogicGlobally{Environ: t.Environ, Body: children[0]}
		}
	case *LogicEventually:
		if len(children) >= 1 {
			return &LogicEventually{Environ: t.Environ, Body: children[0]}
		}
	case *LogicNamedBinder:
		if len(children) >= 1 {
			return &LogicNamedBinder{Name: t.Name, Variables: t.Variables, Environ: t.Environ, Body: children[0]}
		}
	case *LogicWhenOperator:
		if len(children) >= 2 {
			return &LogicWhenOperator{WSort: t.WSort, Name: t.Name, T1: children[0], T2: children[1]}
		}
	case *Cond:
		if len(children) >= 2 {
			return &Cond{CSort: t.CSort, T1: children[0], T2: children[1]}
		}
	}
	return n
}

// --- Name-based substitution (needed by several functions below) ---

// SubstituteByName substitutes terms for variables by name (string key).
// This mirrors Python's substitute_ast(ast, subs) where subs maps
// variable names to replacement terms.
func SubstituteByName(ast Expr, subs map[string]Expr) Expr {
	// Python's substitute_ast has no early return for empty subs — always recurses.
	return substituteByNameRec(ast, subs)
}

func substituteByNameRec(ast Expr, subs map[string]Expr) Expr {
	if v, ok := ast.(*LogicVariable); ok {
		if rep, found := subs[v.Name]; found {
			return rep
		}
		return ast
	}
	// For quantifiers/binders, remove bound variables from subs
	switch t := ast.(type) {
	case *ForAll:
		newsubs := removeBoundNames(subs, t.Variables)
		body := substituteByNameRec(t.Body, newsubs)
		return &ForAll{Variables: t.Variables, Body: body}
	case *LogicExists:
		newsubs := removeBoundNames(subs, t.Variables)
		body := substituteByNameRec(t.Body, newsubs)
		return &LogicExists{Variables: t.Variables, Body: body}
	}
	children := ast.Children()
	if len(children) == 0 {
		return ast
	}
	newChildren := make([]Expr, len(children))
	for i, c := range children {
		newChildren[i] = substituteByNameRec(c, subs)
	}
	return cloneNode(ast, newChildren)
}

func removeBoundNames(subs map[string]Expr, vars []*LogicVariable) map[string]Expr {
	newsubs := make(map[string]Expr, len(subs))
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
func freeVariablesInOrder(ast Expr) []*LogicVariable {
	seen := make(map[string]bool)
	var result []*LogicVariable
	freeVariablesInOrderRec(ast, &result, seen, nil)
	return result
}

// freeVariablesInOrderMulti returns free variables across multiple ASTs,
// unique by name, in order of first appearance. Matches Python's
// used_variables_asts which uses variables_ast (free vars only).
func freeVariablesInOrderMulti(asts []Expr) []*LogicVariable {
	seen := make(map[string]bool)
	var result []*LogicVariable
	for _, ast := range asts {
		freeVariablesInOrderRec(ast, &result, seen, nil)
	}
	return result
}

func freeVariablesInOrderRec(ast Expr, result *[]*LogicVariable, seen map[string]bool, bound map[string]bool) {
	if v, ok := ast.(*LogicVariable); ok {
		if !bound[v.Name] && !seen[v.Name] {
			seen[v.Name] = true
			*result = append(*result, v)
		}
		return
	}
	switch t := ast.(type) {
	case *ForAll:
		newBound := copyBoundSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = true
		}
		freeVariablesInOrderRec(t.Body, result, seen, newBound)
		return
	case *LogicExists:
		newBound := copyBoundSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = true
		}
		freeVariablesInOrderRec(t.Body, result, seen, newBound)
		return
	case *Lambda:
		newBound := copyBoundSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = true
		}
		freeVariablesInOrderRec(t.Body, result, seen, newBound)
		return
	case *LogicNamedBinder:
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
func usedVariablesInOrder(ast Expr) []*LogicVariable {
	seen := make(map[string]bool)
	var result []*LogicVariable
	usedVariablesInOrderRec(ast, &result, seen)
	return result
}

func usedVariablesInOrderRec(ast Expr, result *[]*LogicVariable, seen map[string]bool) {
	if v, ok := ast.(*LogicVariable); ok {
		if !seen[v.Name] {
			seen[v.Name] = true
			*result = append(*result, v)
		}
		return
	}
	switch t := ast.(type) {
	case *ForAll:
		for _, v := range t.Variables {
			if !seen[v.Name] {
				seen[v.Name] = true
				*result = append(*result, v)
			}
		}
		usedVariablesInOrderRec(t.Body, result, seen)
		return
	case *LogicExists:
		for _, v := range t.Variables {
			if !seen[v.Name] {
				seen[v.Name] = true
				*result = append(*result, v)
			}
		}
		usedVariablesInOrderRec(t.Body, result, seen)
		return
	case *Lambda:
		for _, v := range t.Variables {
			if !seen[v.Name] {
				seen[v.Name] = true
				*result = append(*result, v)
			}
		}
		usedVariablesInOrderRec(t.Body, result, seen)
		return
	case *LogicNamedBinder:
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
func usedVariablesInOrderMulti(asts []Expr) []*LogicVariable {
	seen := make(map[string]bool)
	var result []*LogicVariable
	for _, ast := range asts {
		usedVariablesInOrderRec(ast, &result, seen)
	}
	return result
}

// --- NormalizeFreeVariables ---

// NormalizeFreeVariables normalizes free variables: renames them V0, V1, ...
// in the order they appear.
// Returns (old_vars, new_vars, normalized_ast).
func LogicUtilNormalizeFreeVariables(ast Expr) ([]*LogicVariable, []*LogicVariable, Expr) {
	subs := make(map[string]Expr)
	var vs []*LogicVariable
	var nvs []*LogicVariable
	for _, v := range freeVariablesInOrder(ast) {
		if _, exists := subs[v.Name]; !exists {
			nv, _ := NewVariable(fmt.Sprintf("V%d", len(subs)), v.VSort)
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
func NormalizeFreeVariablesTuple(asts ...Expr) ([]*LogicVariable, []*LogicVariable, []Expr) {
	subs := make(map[string]Expr)
	var vs []*LogicVariable
	var nvs []*LogicVariable
	for _, v := range freeVariablesInOrderMulti(asts) {
		if _, exists := subs[v.Name]; !exists {
			nv, _ := NewVariable(fmt.Sprintf("V%d", len(subs)), v.VSort)
			subs[v.Name] = nv
			vs = append(vs, v)
			nvs = append(nvs, nv)
		}
	}
	nasts := make([]Expr, len(asts))
	for i, ast := range asts {
		nasts[i] = SubstituteByName(ast, subs)
	}
	return vs, nvs, nasts
}

// --- NormalizeNamedBinders ---

// NormalizeNamedBinders normalizes the variables bound by named binders.
// If names is nil, all named binders are normalized; otherwise only those
// whose name is in the names set.
func NormalizeNamedBinders(n Node, names map[string]bool) Node {
	xtraceParts("ilu.normalizeNamedBinders ENTER type=", ShortTypeName(n), " HASH canon=", canonPart(n))
	if _, ok := n.(*Const); ok {
		xtracer.Trace("ilu.normalizeNamedBinders EXIT type=%s const", ShortTypeName(n))
		return n
	}
	if nb, ok := n.(*LogicNamedBinder); ok {
		if names == nil || names[nb.Name] {
			nvs := make([]*LogicVariable, len(nb.Variables))
			subs := make(map[string]Expr, len(nb.Variables))
			for i, v := range nb.Variables {
				nv, _ := NewVariable(fmt.Sprintf("V%d", i), v.VSort)
				nvs[i] = nv
				subs[v.Name] = nv
			}
			// Python assertion: generated V0,V1,... must not clash with free variables
			free := FreeVariablesSet(n.(Expr))
			for _, nv := range nvs {
				if free[nv.Name] {
					panic(fmt.Sprintf("NormalizeNamedBinders: generated variable %s clashes with free variable", nv.Name))
				}
			}
			body := NormalizeNamedBinders(nb.Body, names).(Expr)
			body = SubstituteByName(body, subs)
			result := &LogicNamedBinder{Name: nb.Name, Variables: nvs, Environ: nb.Environ, Body: body}
			xtraceParts("ilu.normalizeNamedBinders EXIT type=", ShortTypeName(result), " binder HASH canon=", canonPart(result))
			return result
		}
	}
	// For Apply nodes, normalize the func part too.
	// Python processes ast.args (terms) before ast.rep (func).
	if app, ok := n.(*Apply); ok {
		newTerms := make([]Expr, len(app.Terms))
		for i, t := range app.Terms {
			newTerms[i] = NormalizeNamedBinders(t, names).(Expr)
		}
		newFunc := NormalizeNamedBinders(app.Func, names).(Expr)
		result := MustApply(newFunc, newTerms...)
		xtraceParts("ilu.normalizeNamedBinders EXIT type=", ShortTypeName(result), " app HASH canon=", canonPart(result))
		return result
	}
	// General case: recurse into children and always clone (matching Python).
	children := n.Args()
	newChildren := make([]Node, len(children))
	for i, c := range children {
		newChildren[i] = NormalizeNamedBinders(c, names)
	}
	result := n.Clone(newChildren)
	xtraceParts("ilu.normalizeNamedBinders EXIT type=", ShortTypeName(result), " cloned HASH canon=", canonPart(result))
	return result
}

// --- Temporal -> NamedBinder conversion ---

// GloballyBinderFunc creates a named binder from a Globally operator.
type GloballyBinderFunc func(vars []*LogicVariable, body Expr, environ *string) *LogicNamedBinder

// WhenBinderFunc creates a named binder from a WhenOperator.
type WhenBinderFunc func(name string, vars []*LogicVariable, body Expr) *LogicNamedBinder

// DefaultGloballyBinder creates a NamedBinder named "g" (or "g[env]" if env != nil).
func DefaultGloballyBinder(vars []*LogicVariable, body Expr, environ *string) *LogicNamedBinder {
	name := "g"
	if environ != nil {
		name = "g[" + *environ + "]"
	}
	return &LogicNamedBinder{Name: name, Variables: vars, Environ: nil, Body: body}
}

// DefaultWhenBinder creates a NamedBinder named "l2s_when<name>".
func DefaultWhenBinder(name string, vars []*LogicVariable, body Expr) *LogicNamedBinder {
	return &LogicNamedBinder{Name: "l2s_when" + name, Variables: vars, Environ: nil, Body: body}
}

// applyNamedBinder applies a named binder to arguments, returning Apply if
// there are arguments or the binder itself if there are none.
func applyNamedBinder(nb *LogicNamedBinder, args []Expr) Expr {
	if len(args) == 0 {
		return nb
	}
	return MustApply(nb, args...)
}

// ReplaceTemporalsByNamedBinder replaces temporal operators (Globally,
// Eventually, WhenOperator) with named binders.
// If g is nil, DefaultGloballyBinder is used.
// If when is nil, DefaultWhenBinder is used.
func ReplaceTemporalsByNamedBinder(n Node, g GloballyBinderFunc, when WhenBinderFunc) Node {
	if g == nil {
		g = DefaultGloballyBinder
	}
	if when == nil {
		when = DefaultWhenBinder
	}
	return replaceTemporalsRec(n, g, when)
}

var rtrDepth int

// ResetRtrDepth resets the replaceTemporalsRec depth counter (for diagnostics).
func ResetRtrDepth() { rtrDepth = 0 }

func replaceTemporalsRec(n Node, g GloballyBinderFunc, when WhenBinderFunc) Node {
	rtrDepth++
	myDepth := rtrDepth
	xtraceParts("ilu.replaceTemporalsRec ENTER depth=", myDepth, " type=", ShortTypeName(n), " HASH canon=", canonPart(n))

	// Python outer if/elif: Globally, Eventually, WhenOperator (lines 302-319)
	switch t := n.(type) {
	case *LogicGlobally:
		xtraceParts("ilu.replaceTemporalsRec GLOBALLY_BODY HASH canon=", canonPart(t.Body))
		body := replaceTemporalsRec(t.Body, g, when).(Expr)
		vs, nvs, body := LogicUtilNormalizeFreeVariables(body)
		nb := g(nvs, body, t.Environ)
		result := applyNamedBinder(nb, logicutilVarsToNodes(vs))
		xtraceParts("ilu.replaceTemporalsRec EXIT type=", ShortTypeName(result), " globally HASH canon=", canonPart(result))
		return result

	case *LogicEventually:
		notBody := &LogicNot{Body: t.Body}
		glob := &LogicGlobally{Environ: t.Environ, Body: notBody}
		xtraceParts("ilu.replaceTemporalsRec EVENTUALLY_DESUGAR HASH canon=", canonPart(&LogicNot{Body: glob}))
		result := replaceTemporalsRec(&LogicNot{Body: glob}, g, when)
		xtraceParts("ilu.replaceTemporalsRec EXIT type=", ShortTypeName(result), " eventually HASH canon=", canonPart(result))
		return result

	case *LogicWhenOperator:
		xtraceParts("ilu.replaceTemporalsRec WHEN_T1 HASH canon=", canonPart(t.T1))
		val := replaceTemporalsRec(t.T1, g, when).(Expr)
		xtraceParts("ilu.replaceTemporalsRec WHEN_T2 HASH canon=", canonPart(t.T2))
		cond := replaceTemporalsRec(t.T2, g, when).(Expr)
		body := &Cond{CSort: val.NodeSort(), T1: cond, T2: val}
		vs, nvs, nbody := LogicUtilNormalizeFreeVariables(body)
		nb := when(t.Name, nvs, nbody)
		result := applyNamedBinder(nb, logicutilVarsToNodes(vs))
		xtraceParts("ilu.replaceTemporalsRec EXIT type=", ShortTypeName(result), " when HASH canon=", canonPart(result))
		return result
	}

	// === Python else branch (line 321) ===

	// Step 1: Common recursion of all args (Python line 322)
	// args = [replace_temporals_by_named_binder_g_ast(x, g, when) for x in ast.args]
	children := n.Args()
	newChildren := make([]Node, len(children))
	for i, c := range children {
		xtraceParts("ilu.replaceTemporalsRec CHILD parent=", ShortTypeName(n), " childIdx=", i, " nChildren=", len(children), " childType=", ShortTypeName(c), " HASH canon=", canonPart(c))
		newChildren[i] = replaceTemporalsRec(c, g, when)
	}

	// Step 2a: Apply (Python line 323)
	if t, ok := n.(*Apply); ok {
		// Python line 324: l2s_init sub-case
		if nb, ok := t.Func.(*LogicNamedBinder); ok && nb.Name == "l2s_init" {
			// Python line 325: recurse body AFTER terms (terms already done above)
			xtraceParts("ilu.replaceTemporalsRec L2S_INIT_BODY HASH canon=", canonPart(nb.Body))
			body := replaceTemporalsRec(nb.Body, g, when).(Expr)
			newArgs := logicutilNodesToExprs(newChildren)
			if notBody, ok := body.(*LogicNot); ok {
				// Python line 327: lg.Not(lg.Apply(ast.func.clone([body.body]), *args))
				clonedNB := nb.Clone([]Node{notBody.Body}).(Expr)
				inner := MustApply(clonedNB, newArgs...)
				result := &LogicNot{Body: inner}
				xtraceParts("ilu.replaceTemporalsRec EXIT type=", ShortTypeName(result), " l2s_init_not HASH canon=", canonPart(result))
				return result
			}
			// Python line 331: lg.Apply(ast.func.clone([body]), *args)
			clonedNB := nb.Clone([]Node{body}).(Expr)
			result := MustApply(clonedNB, newArgs...)
			xtraceParts("ilu.replaceTemporalsRec EXIT type=", ShortTypeName(result), " l2s_init HASH canon=", canonPart(result))
			return result
		}
		// Python line 334: general Apply — recurse func AFTER terms
		xtraceParts("ilu.replaceTemporalsRec APPLY_FUNC funcType=", ShortTypeName(t.Func), " HASH canon=", canonPart(t.Func))
		newFunc := replaceTemporalsRec(t.Func, g, when).(Expr)
		newArgs := logicutilNodesToExprs(newChildren)
		result := MustApply(newFunc, newArgs...)
		xtraceParts("ilu.replaceTemporalsRec EXIT type=", ShortTypeName(result), " app HASH canon=", canonPart(result))
		return result
	}

	// Step 2b: Not double negation (Python line 338)
	if _, ok := n.(*LogicNot); ok && len(newChildren) > 0 {
		if inner, ok := newChildren[0].(*LogicNot); ok {
			xtraceParts("ilu.replaceTemporalsRec EXIT type=", ShortTypeName(inner.Body), " doubleNeg HASH canon=", canonPart(inner.Body))
			return inner.Body
		}
	}

	// Step 2c: NamedBinder l2s_init with Not body (Python line 342)
	if t, ok := n.(*LogicNamedBinder); ok && t.Name == "l2s_init" && len(newChildren) > 0 {
		if notChild, ok := newChildren[0].(*LogicNot); ok {
			// Python line 343: lg.Not(ast.clone([args[0].args[0]]))
			cloned := n.Clone([]Node{notChild.Body})
			result := &LogicNot{Body: cloned.(Expr)}
			xtraceParts("ilu.replaceTemporalsRec EXIT type=", ShortTypeName(result), " nb_init_not HASH canon=", canonPart(result))
			return result
		}
	}

	// Step 2d: Default clone (Python line 346)
	xtracer.Trace("ilu.replaceTemporalsRec CLONE depth=%d type=%s nChildren=%d", myDepth, ShortTypeName(n), len(newChildren))
	result := n.Clone(newChildren)
	xtraceParts("ilu.replaceTemporalsRec EXIT depth=", myDepth, " type=", ShortTypeName(result), " cloned HASH canon=", canonPart(result))
	return result
}

// nodesToExprs converts a slice of ast.Node to a slice of logic.Expr.
func logicutilNodesToExprs(nodes []Node) []Expr {
	exprs := make([]Expr, len(nodes))
	for i, n := range nodes {
		exprs[i] = n.(Expr)
	}
	return exprs
}

func logicutilVarsToNodes(vars []*LogicVariable) []Expr {
	nodes := make([]Expr, len(vars))
	for i, v := range vars {
		nodes[i] = v
	}
	return nodes
}

// --- ReduceNamedBinders ---

// ReduceNamedBinders reduces subexpressions of the form ($b V1...Vn. t)(s1...sn)
// to ($b . t[Vi/si]).
func ReduceNamedBinders(ast Expr, g GloballyBinderFunc) Expr {
	if g == nil {
		g = DefaultGloballyBinder
	}
	return reduceNamedBindersRec(ast, g)
}

func reduceNamedBindersRec(ast Expr, g GloballyBinderFunc) Expr {
	if _, ok := ast.(*Const); ok {
		return ast
	}
	if app, ok := ast.(*Apply); ok {
		if nb, ok := app.Func.(*LogicNamedBinder); ok {
			subst := make(map[string]Expr, len(nb.Variables))
			for i, v := range nb.Variables {
				if i < len(app.Terms) {
					subst[v.Name] = app.Terms[i]
				}
			}
			body := reduceNamedBindersRec(SubstituteByName(nb.Body, subst), g)
			return &LogicNamedBinder{Name: nb.Name, Variables: nil, Environ: nil, Body: body}
		}
		newFunc := NormalizeNamedBinders(app.Func, nil).(Expr)
		newTerms := make([]Expr, len(app.Terms))
		for i, t := range app.Terms {
			newTerms[i] = reduceNamedBindersRec(t, g)
		}
		return MustApply(newFunc, newTerms...)
	}

	children := ast.Children()
	if len(children) == 0 {
		return ast
	}
	newChildren := make([]Expr, len(children))
	for i, c := range children {
		newChildren[i] = reduceNamedBindersRec(c, g)
	}
	return cloneNode(ast, newChildren)
}

// --- ReplaceNamedBindersAst ---

// ReplaceNamedBindersAst replaces named binders that are present in subs.
// subs maps NamedBinder Sexp (structural canonical form) keys to
// replacement nodes. Lookups must use the same key as the construction
// site (see check/l2s_shared.go:786) — Sexp, not String/PrettyFmla.
// Python's replace_named_binders_ast (ivy_logic_utils.py:1142+) uses
// `if ast in subs` which is struct-eq via NamedBinder.__hash__/__eq__.
// Named binders inside other named binders are NOT replaced.
func ReplaceNamedBindersAst(n Node, subs map[string]Expr) Node {
	xtraceParts("ilu.replaceNamedBindersAst ENTER type=", ShortTypeName(n), " HASH canon=", canonPart(n))
	if nb, ok := n.(*LogicNamedBinder); ok {
		key := string(nb.Sexp())
		if rep, found := subs[key]; found {
			xtraceParts("ilu.replaceNamedBindersAst EXIT type=", ShortTypeName(n), " found=True HASH canon=", canonPart(rep))
			return rep
		}
		xtraceParts("ilu.replaceNamedBindersAst EXIT type=", ShortTypeName(n), " found=False HASH canon=", canonPart(n))
		return n
	}
	// python: is_app returns True for *logic.Apply and *logic.Const as
	// well as *logic.NamedBinder with len(term.variables) == 0, but
	// all NamedBinders are taken care of above. So only 2 cases left:
	if app, ok := n.(*Apply); ok {
		newTerms := make([]Expr, len(app.Terms))
		for i, t := range app.Terms {
			newTerms[i] = ReplaceNamedBindersAst(t, subs).(Expr)
		}
		newFunc := app.Func
		if nb, ok := app.Func.(*LogicNamedBinder); ok {
			key := string(nb.Sexp())
			if rep, found := subs[key]; found {
				newFunc = rep
			}
		}
		result := MustApply(newFunc, newTerms...)
		xtraceParts("ilu.replaceNamedBindersAst EXIT type=", ShortTypeName(result), " app HASH canon=", canonPart(result))
		return result
	}
	if cnst, ok := n.(*Const); ok {
		// Python: is_app(Const) → True, enters is_app branch
		// subs.get(ast.rep, ast.rep)(*args) with empty args:
		// Symbol.__call__: returns App(self) if FunctionSort, else self
		var result Node = cnst
		if _, ok := cnst.CSort.(*LogicFunctionSort); ok {
			result = MustApply(cnst)
		}
		xtraceParts("ilu.replaceNamedBindersAst EXIT type=", ShortTypeName(result), " app HASH canon=", canonPart(result))
		return result
	}
	// Python: args = [replace_named_binders_ast(x, subs) for x in ast.args]
	// Python: return ast.clone(args)  — ALWAYS clones, even with empty args.
	children := n.Args()
	newChildren := make([]Node, len(children))
	for i, c := range children {
		newChildren[i] = ReplaceNamedBindersAst(c, subs)
	}
	result := n.Clone(newChildren)
	xtraceParts("ilu.replaceNamedBindersAst EXIT type=", ShortTypeName(result), " cloned HASH canon=", canonPart(result))
	return result
}

// --- ExpandNamedBindersAst ---

// ExpandNamedBindersAst expands nullary named binders by applying fun.
// If fun returns a non-nil node, the result is recursively expanded.
func ExpandNamedBindersAst(ast Expr, fun func(*LogicNamedBinder) Expr) Expr {
	if nb, ok := ast.(*LogicNamedBinder); ok {
		res := fun(nb)
		if res != nil {
			return ExpandNamedBindersAst(res, fun)
		}
	}
	children := ast.Children()
	if len(children) == 0 {
		return ast
	}
	newChildren := make([]Expr, len(children))
	for i, c := range children {
		newChildren[i] = ExpandNamedBindersAst(c, fun)
	}
	return cloneNode(ast, newChildren)
}

// --- DenormalizeTemporal ---

// IsTrue returns true if the node is the logical true constant (empty And).
func LogicUtilIsTrue(n Expr) bool {
	a, ok := n.(*LogicAnd)
	return ok && len(a.Terms) == 0
}

// IsFalse returns true if the node is the logical false constant (empty Or).
func LogicUtilIsFalse(n Expr) bool {
	o, ok := n.(*LogicOr)
	return ok && len(o.Terms) == 0
}

// DenormalizeTemporal reverses temporal normalization:
//   - Not(Globally(Not(x))) => Eventually(x)
//   - Iff/Eq(Globally(Not(x)), true) => Iff/Eq(Eventually(x), Or())
//   - Iff/Eq(Globally(Not(x)), false) => Iff/Eq(Eventually(x), And())
func DenormalizeTemporal(ast Expr) Expr {
	children := ast.Children()
	newChildren := make([]Expr, len(children))
	for i, c := range children {
		newChildren[i] = DenormalizeTemporal(c)
	}

	// Not(Globally(Not(x))) => Eventually(x)
	if _, ok := ast.(*LogicNot); ok && len(newChildren) == 1 {
		if glob, ok := newChildren[0].(*LogicGlobally); ok {
			if inner, ok := glob.Body.(*LogicNot); ok {
				return &LogicEventually{Environ: glob.Environ, Body: inner.Body}
			}
		}
	}

	// Iff/Eq with Globally(Not(x)) on the left
	switch ast.(type) {
	case *LogicIff, *Eq:
		if len(newChildren) >= 2 {
			lhs := newChildren[0]
			rhs := newChildren[1]
			if glob, ok := lhs.(*LogicGlobally); ok {
				if inner, ok := glob.Body.(*LogicNot); ok {
					ev := &LogicEventually{Environ: glob.Environ, Body: inner.Body}
					if LogicUtilIsTrue(rhs) {
						return cloneNode(ast, []Expr{ev, &LogicOr{}})
					}
					if LogicUtilIsFalse(rhs) {
						return cloneNode(ast, []Expr{ev, &LogicAnd{}})
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
func IsVacEqualityLit(lit Expr) bool {
	neg, ok := lit.(*LogicNot)
	if !ok {
		return false
	}
	eq, ok := neg.Body.(*Eq)
	if !ok {
		return false
	}
	return eq.T1.Equal(eq.T2)
}

// IsTrueLit returns true if the literal evaluates to true.
// A positive literal that is And() (true), or a Not(Or()) (not false).
func IsTrueLit(lit Expr) bool {
	if neg, ok := lit.(*LogicNot); ok {
		return LogicUtilIsFalse(neg.Body)
	}
	return LogicUtilIsTrue(lit)
}

// IsFalseLit returns true if the literal evaluates to false.
// A positive literal that is Or() (false), or a Not(And()) (not true).
func IsFalseLit(lit Expr) bool {
	if neg, ok := lit.(*LogicNot); ok {
		return LogicUtilIsTrue(neg.Body)
	}
	return LogicUtilIsFalse(lit)
}

// IsTautLit returns true if the literal is a tautology (always true).
func IsTautLit(lit Expr) bool {
	return IsTrueLit(lit) || IsTautEqualityLit(lit)
}

// IsVacLit returns true if the literal is vacuously true (always false).
func IsVacLit(lit Expr) bool {
	return IsFalseLit(lit) || IsVacEqualityLit(lit)
}

// --- TermListsEq ---

// TermListsEq returns true if two node slices are structurally equal.
func TermListsEq(l1, l2 []Expr) bool {
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
	Clauses [][]Expr // accumulated definition clauses
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
func (tc *TseitinContext) AddDefs(cls [][]Expr) [][]Expr {
	result := make([][]Expr, 0, len(tc.Clauses)+len(cls))
	result = append(result, cls...)
	result = append(result, tc.Clauses...)
	return result
}

// TseitinEncoding performs Tseitin transformation on a formula.
// Returns a node representing the formula as a literal.
// Side effect: adds definition clauses to the TseitinContext.
func TseitinEncoding(tc *TseitinContext, f Expr) Expr {
	f = ExpandAbbrevs(f)

	if and, ok := f.(*LogicAnd); ok {
		args := make([]Expr, len(and.Terms))
		for i, t := range and.Terms {
			args[i] = TseitinEncoding(tc, t)
		}
		// Collect variables in order from args
		varsSeen := make(map[string]bool)
		var vs []*LogicVariable
		for _, arg := range args {
			for _, v := range freeVariablesInOrder(arg) {
				if !varsSeen[v.Name] {
					varsSeen[v.Name] = true
					vs = append(vs, v)
				}
			}
		}
		fname := tc.fresh(strconv.Itoa(len(vs)))
		var fnSort Sort
		if len(vs) > 0 {
			sorts := make([]Sort, len(vs)+1)
			for i, v := range vs {
				sorts[i] = v.VSort
			}
			sorts[len(vs)] = Boolean
			fs, err := NewFunctionSort(sorts...)
			if err != nil {
				fnSort = Boolean
			} else {
				fnSort = fs
			}
		} else {
			fnSort = Boolean
		}
		fn := NewConst(fname, fnSort)
		var res Expr
		if len(vs) > 0 {
			res = MustApply(fn, logicutilVarsToNodes(vs)...)
		} else {
			res = fn
		}
		for _, arg := range args {
			tc.Clauses = append(tc.Clauses, []Expr{&LogicNot{Body: res}, arg})
		}
		lastClause := []Expr{res}
		for _, arg := range args {
			lastClause = append(lastClause, &LogicNot{Body: arg})
		}
		tc.Clauses = append(tc.Clauses, lastClause)
		return res
	}

	if or, ok := f.(*LogicOr); ok {
		negTerms := make([]Expr, len(or.Terms))
		for i, t := range or.Terms {
			negTerms[i] = &LogicNot{Body: t}
		}
		return &LogicNot{Body: TseitinEncoding(tc, &LogicAnd{Terms: negTerms})}
	}

	if not, ok := f.(*LogicNot); ok {
		return &LogicNot{Body: TseitinEncoding(tc, not.Body)}
	}

	// Atom: return as-is
	return f
}

// --- FormulaToLit ---

// FormulaToLit converts a formula to a literal (positive or negated atom).
// Uses Tseitin encoding for complex subformulas.
func FormulaToLit(tc *TseitinContext, f Expr) Expr {
	f = ExpandAbbrevs(f)
	if not, ok := f.(*LogicNot); ok {
		inner := FormulaToLit(tc, not.Body)
		return &LogicNot{Body: inner}
	}
	if isAtomNode(f) {
		return f
	}
	return TseitinEncoding(tc, f)
}

// isAtomNode checks if the node is an atomic formula.
func isAtomNode(n Expr) bool {
	if _, ok := n.(*Eq); ok {
		return true
	}
	switch n.(type) {
	case *Apply:
		return SortEqual(n.NodeSort(), Boolean)
	case *Const:
		return SortEqual(n.NodeSort(), Boolean)
	}
	return false
}

// --- FormulaToClause ---

// FormulaToClause converts a formula to a clause (disjunction of literals).
// Returns a slice of literal nodes.
func FormulaToClause(tc *TseitinContext, f Expr) []Expr {
	f = ExpandAbbrevs(f)
	f = DeMorgan(f)
	if LogicUtilIsTrue(f) {
		return []Expr{f}
	}
	if or, ok := f.(*LogicOr); ok {
		var result []Expr
		for _, t := range or.Terms {
			result = append(result, FormulaToClause(tc, t)...)
		}
		return result
	}
	return []Expr{FormulaToLit(tc, f)}
}

// --- FormulaToClube ---

// FormulaToClube converts a formula to a cube (conjunction of literals).
// Returns a slice of literal nodes.
func FormulaToClube(tc *TseitinContext, f Expr) []Expr {
	f = ExpandAbbrevs(f)
	if not, ok := f.(*LogicNot); ok {
		clause := FormulaToClause(tc, not.Body)
		result := make([]Expr, len(clause))
		for i, lit := range clause {
			result[i] = &LogicNot{Body: lit}
		}
		return result
	}
	if and, ok := f.(*LogicAnd); ok {
		var result []Expr
		for _, t := range and.Terms {
			result = append(result, FormulaToLit(tc, t))
		}
		return result
	}
	return []Expr{FormulaToLit(tc, f)}
}

// --- ReduceNumerically ---

// ReduceNumerically evaluates numeric expressions where all arguments
// are numeral constants. Handles <, =, and .succ operators.
func ReduceNumerically(ast Expr) Expr {
	children := ast.Children()
	newChildren := make([]Expr, len(children))
	for i, c := range children {
		newChildren[i] = ReduceNumerically(c)
	}

	if app, ok := ast.(*Apply); ok {
		allNumeral := len(app.Terms) > 0
		for i := range app.Terms {
			nc := newChildren[i] // ast.Children() = Terms only (Func excluded)
			c, ok := nc.(*Const)
			if !ok || !isAllDigits(c.Name) {
				allNumeral = false
				break
			}
		}
		if allNumeral {
			vals := make([]int, len(app.Terms))
			for i := range app.Terms {
				nc := newChildren[i]
				c := nc.(*Const)
				vals[i], _ = strconv.Atoi(c.Name)
			}
			if fn, ok := app.Func.(*Const); ok {
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
func NormalizeQuantifiers(t Expr) Expr {
	switch n := t.(type) {
	case *LogicVariable, *Const:
		return t

	case *Apply:
		newTerms := make([]Expr, len(n.Terms))
		for i, term := range n.Terms {
			newTerms[i] = NormalizeQuantifiers(term)
		}
		return MustApply(n.Func, newTerms...)

	case *Eq:
		return &Eq{T1: NormalizeQuantifiers(n.T1), T2: NormalizeQuantifiers(n.T2)}

	case *LogicIte:
		return &LogicIte{
			Cond: NormalizeQuantifiers(n.Cond),
			Then: NormalizeQuantifiers(n.Then),
			Else: NormalizeQuantifiers(n.Else),
		}

	case *LogicNot:
		return &LogicNot{Body: NormalizeQuantifiers(n.Body)}

	case *LogicImplies:
		return &LogicImplies{T1: NormalizeQuantifiers(n.T1), T2: NormalizeQuantifiers(n.T2)}

	case *LogicIff:
		return &LogicIff{T1: NormalizeQuantifiers(n.T1), T2: NormalizeQuantifiers(n.T2)}

	case *LogicAnd:
		// Flatten: And(x, And(a,b), y) -> And(x, a, b, y)
		var terms []Expr
		for _, x := range n.Terms {
			y := NormalizeQuantifiers(x)
			if inner, ok := y.(*LogicAnd); ok {
				terms = append(terms, inner.Terms...)
			} else {
				terms = append(terms, y)
			}
		}
		return &LogicAnd{Terms: terms}

	case *LogicOr:
		// Flatten: Or(x, Or(a,b), y) -> Or(x, a, b, y)
		var terms []Expr
		for _, x := range n.Terms {
			y := NormalizeQuantifiers(x)
			if inner, ok := y.(*LogicOr); ok {
				terms = append(terms, inner.Terms...)
			} else {
				terms = append(terms, y)
			}
		}
		return &LogicOr{Terms: terms}

	case *ForAll:
		// ForAll(vars, And(a,b)) -> And(ForAll(vars,a), ForAll(vars,b))
		if inner, ok := n.Body.(*LogicAnd); ok {
			terms := make([]Expr, len(inner.Terms))
			for i, x := range inner.Terms {
				terms[i] = &ForAll{Variables: n.Variables, Body: x}
			}
			return NormalizeQuantifiers(&LogicAnd{Terms: terms})
		}
		// Otherwise, restrict variables to those actually free in the body
		body := NormalizeQuantifiers(n.Body)
		fvs := FreeVariables(body)
		var vars []*LogicVariable
		for _, v := range n.Variables {
			if _, ok := fvs.Get2(Key(v)); ok {
				vars = append(vars, v)
			}
		}
		if len(vars) == 0 {
			return body
		}
		return &ForAll{Variables: vars, Body: body}

	case *LogicExists:
		// Exists(vars, Or(a,b)) -> Or(Exists(vars,a), Exists(vars,b))
		if inner, ok := n.Body.(*LogicOr); ok {
			terms := make([]Expr, len(inner.Terms))
			for i, x := range inner.Terms {
				terms[i] = &LogicExists{Variables: n.Variables, Body: x}
			}
			return NormalizeQuantifiers(&LogicOr{Terms: terms})
		}
		// Otherwise, restrict variables to those actually free in the body
		body := NormalizeQuantifiers(n.Body)
		fvs := FreeVariables(body)
		var vars []*LogicVariable
		for _, v := range n.Variables {
			if _, ok := fvs.Get2(Key(v)); ok {
				vars = append(vars, v)
			}
		}
		if len(vars) == 0 {
			return body
		}
		return &LogicExists{Variables: vars, Body: body}
	}

	// Python logic_util.py:271: assert False, type(t)
	panic(fmt.Sprintf("NormalizeQuantifiers: unexpected type %T", t))
}

// -----------------------------------------------------------------------
// SubstituteApply
// -----------------------------------------------------------------------

// SubstituteApplyFunc is the type for functions used in SubstituteApply.
// Given replacement terms, it produces a result node.
type SubstituteApplyFunc func(terms []Expr) Expr

// SubstituteApply performs second-order substitution: for any key in subs,
// Apply(key, terms...) is replaced by subs[key](terms'...), where terms'
// are recursively substituted. Non-application occurrences of keys are
// NOT substituted.
// Corresponds to Python logic_util.substitute_apply.
func SubstituteApply(t Expr, subs map[NodeKey]SubstituteApplyFunc) Expr {
	if len(subs) == 0 {
		return t
	}
	return substituteApplyRec(t, subs)
}

func substituteApplyRec(t Expr, subs map[NodeKey]SubstituteApplyFunc) Expr {
	switch n := t.(type) {
	case *LogicVariable, *Const:
		return t

	case *Apply:
		if fn, ok := subs[Key(n.Func)]; ok {
			newTerms := make([]Expr, len(n.Terms))
			for i, term := range n.Terms {
				newTerms[i] = substituteApplyRec(term, subs)
			}
			result := fn(newTerms)
			// Python logic_util.py:222-224: assert fvr <= fvt
			termFVs := make(map[NodeKey]bool)
			for _, term := range newTerms {
				for _, v := range FreeVariables(term).All() {
					termFVs[Key(v)] = true
				}
			}
			for _, v := range FreeVariables(result).All() {
				if !termFVs[Key(v)] {
					panic(fmt.Sprintf("SubstituteApply: new free variables introduced by substitution"))
				}
			}
			return result
		}
		return substituteApplyChildren(t, subs)

	default:
		return substituteApplyChildren(t, subs)
	}
}

func substituteApplyChildren(t Expr, subs map[NodeKey]SubstituteApplyFunc) Expr {
	switch n := t.(type) {
	case *Apply:
		// Python substitute_apply iterates ALL children (for x in t),
		// including Func, to handle nested Apply whose func is in subs.
		newFunc := substituteApplyRec(n.Func, subs)
		newTerms := make([]Expr, len(n.Terms))
		for i, term := range n.Terms {
			newTerms[i] = substituteApplyRec(term, subs)
		}
		result, err := NewApply(newFunc, newTerms...)
		if err != nil {
			return t
		}
		return result

	case *Eq:
		t1 := substituteApplyRec(n.T1, subs)
		t2 := substituteApplyRec(n.T2, subs)
		return &Eq{T1: t1, T2: t2}

	case *LogicIte:
		c := substituteApplyRec(n.Cond, subs)
		th := substituteApplyRec(n.Then, subs)
		el := substituteApplyRec(n.Else, subs)
		return &LogicIte{Cond: c, Then: th, Else: el}

	case *LogicNot:
		b := substituteApplyRec(n.Body, subs)
		return &LogicNot{Body: b}

	case *LogicAnd:
		terms := make([]Expr, len(n.Terms))
		for i, term := range n.Terms {
			terms[i] = substituteApplyRec(term, subs)
		}
		return &LogicAnd{Terms: terms}

	case *LogicOr:
		terms := make([]Expr, len(n.Terms))
		for i, term := range n.Terms {
			terms[i] = substituteApplyRec(term, subs)
		}
		return &LogicOr{Terms: terms}

	case *LogicImplies:
		t1 := substituteApplyRec(n.T1, subs)
		t2 := substituteApplyRec(n.T2, subs)
		return &LogicImplies{T1: t1, T2: t2}

	case *LogicIff:
		t1 := substituteApplyRec(n.T1, subs)
		t2 := substituteApplyRec(n.T2, subs)
		return &LogicIff{T1: t1, T2: t2}

	case *ForAll:
		// Remove bound vars from subs
		newSubs := filterSubs(subs, n.Variables)
		body := substituteApplyRec(n.Body, newSubs)
		return &ForAll{Variables: n.Variables, Body: body}

	case *LogicExists:
		newSubs := filterSubs(subs, n.Variables)
		body := substituteApplyRec(n.Body, newSubs)
		return &LogicExists{Variables: n.Variables, Body: body}

	case *Lambda:
		newSubs := filterSubs(subs, n.Variables)
		body := substituteApplyRec(n.Body, newSubs)
		return &Lambda{Variables: n.Variables, Body: body}

	case *LogicNamedBinder:
		newSubs := filterSubs(subs, n.Variables)
		body := substituteApplyRec(n.Body, newSubs)
		return &LogicNamedBinder{Name: n.Name, Variables: n.Variables, Environ: n.Environ, Body: body}
	}
	return t
}

func filterSubs(subs map[NodeKey]SubstituteApplyFunc, vars []*LogicVariable) map[NodeKey]SubstituteApplyFunc {
	newSubs := make(map[NodeKey]SubstituteApplyFunc, len(subs))
	varSet := make(map[NodeKey]struct{}, len(vars))
	for _, v := range vars {
		varSet[Key(v)] = struct{}{}
	}
	for k, v := range subs {
		if _, bound := varSet[k]; !bound {
			newSubs[k] = v
		}
	}
	return newSubs
}
