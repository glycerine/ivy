package module

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// SymbolsAST yields all constant symbols in a node (the function symbol
// of applications, plus any bare constants). This corresponds to Python's
// symbols_ast which yields the "rep" of app nodes.
func SymbolsAST(node lg.Expr) []*lg.Const {
	result := usedSymbolsAST(node)
	out := make([]*lg.Const, 0, len(result))
	for _, sym := range result {
		out = append(out, sym)
	}
	return out
}

// UsedSymbolsAST returns the set of used constant symbols in a node.
func UsedSymbolsAST(node lg.Expr) map[lg.NodeKey]*lg.Const {
	return usedSymbolsAST(node)
}

// VariablesAST yields free variables in a node (not bound variables).
// This matches Python's variables_ast which skips bound variables.
func VariablesAST(node lg.Expr) []*lg.Variable {
	result := make(map[lg.NodeKey]lg.Expr)
	variablesASTRec(node, result, nil)
	out := make([]*lg.Variable, 0, len(result))
	for _, node := range result {
		if v, ok := node.(*lg.Variable); ok {
			out = append(out, v)
		}
	}
	return out
}

// UsedVariablesAST returns the set of free variables in a node.
func UsedVariablesAST(node lg.Expr) map[lg.NodeKey]lg.Expr {
	result := make(map[lg.NodeKey]lg.Expr)
	variablesASTRec(node, result, nil)
	return result
}

func variablesASTRec(node lg.Expr, result map[lg.NodeKey]lg.Expr, bound map[string]struct{}) {
	switch t := node.(type) {
	case *lg.Variable:
		if bound != nil {
			if _, ok := bound[t.Name]; ok {
				return
			}
		}
		result[lg.Key(t)] = t
		return
	case *lg.ForAll:
		newBound := copyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		variablesASTRec(t.Body, result, newBound)
		return
	case *lg.Exists:
		newBound := copyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		variablesASTRec(t.Body, result, newBound)
		return
	case *lg.Lambda:
		newBound := copyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		variablesASTRec(t.Body, result, newBound)
		return
	case *lg.NamedBinder:
		newBound := copyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		variablesASTRec(t.Body, result, newBound)
		return
	}
	for _, c := range node.Children() {
		variablesASTRec(c, result, bound)
	}
}

func copyStringSet(s map[string]struct{}) map[string]struct{} {
	r := make(map[string]struct{}, len(s))
	for k := range s {
		r[k] = struct{}{}
	}
	return r
}

// SubstituteConstantsAST substitutes terms for constants in an AST node.
// Matches Python's substitute_constants_ast (ivy_logic_utils.py:173).
// The map keys are lg.NodeKey (via lg.Key(sym)) for structural equality
// matching Python's recstruct-based Symbol lookup: subs.get(ast.rep, ast).
func SubstituteConstantsAST(node ast.Node, subs map[lg.NodeKey]lg.Expr) ast.Node {
	// Python: if is_constant(ast): return subs.get(ast.rep, ast)
	if sym, ok := node.(*lg.Const); ok {
		if rep, found := subs[lg.Key(sym)]; found {
			return rep
		}
		return sym
	}

	args := node.Args()
	xtracer.Trace("actions.substitute_constants_action ENTER type=%s nargs=%d",
		iu.ShortTypeName(node), len(args))

	if len(args) == 0 {
		// Leaf non-constant (Variable, Atom label, etc.).
		// Python traces then clones with empty args.
		return node.Clone(args)
	}

	newArgs := make([]ast.Node, len(args))
	for i, arg := range args {
		newArgs[i] = SubstituteConstantsAST(arg, subs)
	}
	return node.Clone(newArgs)
}

// SubstituteConstantsExpr is SubstituteConstantsAST for lg.Expr callers.
func SubstituteConstantsExpr(node lg.Expr, subs map[lg.NodeKey]lg.Expr) lg.Expr {
	return SubstituteConstantsAST(node, subs).(lg.Expr)
}

// RenameAST renames symbols in an AST by structural identity.
// The map keys are lg.NodeKey (via lg.Key(sym)) for structural equality
// matching Python's recstruct-based Symbol lookup: subs.get(ast.rep, ast.rep).
// Variables are not renamed.
func RenameAST(node lg.Expr, subs map[lg.NodeKey]*lg.Const) lg.Expr {
	if len(subs) == 0 {
		return node
	}
	return renameASTRec(node, subs)
}

func renameASTRec(node lg.Expr, subs map[lg.NodeKey]*lg.Const) lg.Expr {
	switch t := node.(type) {
	case *lg.Const:
		if r, ok := subs[lg.Key(t)]; ok {
			// Preserve original sort if replacement has TopSort.
			// This matches Python where rename_ast substitutions carry
			// the original sort via sym.prefix('new_') etc.
			if _, isTop := r.CSort.(*lg.TopSort); isTop && t.CSort != nil {
				if _, origIsTop := t.CSort.(*lg.TopSort); !origIsTop {
					return lg.NewConst(r.Name, t.CSort)
				}
			}
			return r
		}
		return node
	case *lg.Variable:
		return node // variables not renamed
	case *lg.Apply:
		newFunc := renameASTRec(t.Func, subs)
		newTerms := make([]lg.Expr, len(t.Terms))
		changed := newFunc != t.Func
		for i, arg := range t.Terms {
			newTerms[i] = renameASTRec(arg, subs)
			if newTerms[i] != arg {
				changed = true
			}
		}
		if !changed {
			return node
		}
		return &lg.Apply{Func: newFunc, Terms: newTerms}
	case *il.Definition:
		lhs := renameASTRec(t.Lhs, subs)
		rhs := renameASTRec(t.Rhs, subs)
		if lhs == t.Lhs && rhs == t.Rhs {
			return node
		}
		return il.NewDefinition(lhs, rhs)
	}

	children := node.Children()
	if len(children) == 0 {
		return node
	}
	newChildren := make([]lg.Expr, len(children))
	changed := false
	for i, c := range children {
		newChildren[i] = renameASTRec(c, subs)
		if newChildren[i] != c {
			changed = true
		}
	}
	if !changed {
		return node
	}
	return il.CloneNode(node, newChildren)
}

// collectConstsByName walks an expression tree and collects all *lg.Const
// nodes whose names appear in the nameSubs map. For each found constant,
// it adds a (name,sort)-keyed entry to the output map with the renamed
// constant preserving the original sort. This matches Python's recstruct
// __eq__ which compares (name, sort) tuples.
func collectConstsByName(node lg.Expr, nameSubs map[string]string, out map[lg.NodeKey]*lg.Const) {
	if node == nil {
		return
	}
	switch t := node.(type) {
	case *lg.Const:
		if newName, ok := nameSubs[t.Name]; ok {
			out[lg.Key(t)] = lg.NewConst(newName, t.CSort)
		}
	case *lg.Variable:
		// variables not collected
	case *lg.Apply:
		collectConstsByName(t.Func, nameSubs, out)
		for _, arg := range t.Terms {
			collectConstsByName(arg, nameSubs, out)
		}
	case *il.Definition:
		collectConstsByName(t.Lhs, nameSubs, out)
		collectConstsByName(t.Rhs, nameSubs, out)
	default:
		for _, c := range node.Children() {
			collectConstsByName(c, nameSubs, out)
		}
	}
}

// RenameASTByName renames constants in a formula using a name→name map.
// It first scans the formula to discover actual (name, sort) keys, then
// applies RenameAST with correctly-keyed entries. This matches Python's
// rename_ast where subs.get(ast.rep, ast.rep) uses recstruct equality
// on (name, sort) pairs.
func RenameASTByName(node lg.Expr, subs map[string]string) lg.Expr {
	if len(subs) == 0 || node == nil {
		return node
	}
	constMap := make(map[lg.NodeKey]*lg.Const)
	collectConstsByName(node, subs, constMap)
	if len(constMap) == 0 {
		return node
	}
	return RenameAST(node, constMap)
}

// RenameClausesByName renames symbols in clauses using a name→name map.
// Scans the clauses to discover actual (name, sort) keys, then applies
// RenameClauses with correctly-keyed entries. This matches Python where
// rename_clauses → rename_ast uses recstruct (name, sort) equality.
func RenameClausesByName(clauses *Clauses, subs map[string]string) *Clauses {
	if len(subs) == 0 {
		return clauses
	}
	constMap := make(map[lg.NodeKey]*lg.Const)
	for _, f := range clauses.Fmlas {
		collectConstsByName(f, subs, constMap)
	}
	for _, d := range clauses.Defs {
		collectConstsByName(d, subs, constMap)
	}
	if len(constMap) == 0 {
		return clauses
	}
	return RenameClauses(clauses, constMap)
}

// FreeVariablesAST is an alias for logicutil.FreeVariables.
func FreeVariablesAST(node lg.Expr) map[lg.NodeKey]lg.Expr {
	return lu.FreeVariables(node)
}

// UsedAllVariablesAST returns all variables used in the node
// (both free and bound). This delegates to logicutil.UsedVariables.
func UsedAllVariablesAST(node lg.Expr) map[lg.NodeKey]lg.Expr {
	return lu.UsedVariables(node)
}

// IsGroundAST returns true if the formula has no free variables.
func IsGroundAST(node lg.Expr) bool {
	vars := VariablesAST(node)
	return len(vars) == 0
}

// --- Normalize / collect helpers ---

// CollectAndList flattens nested And formulas into a flat list.
func CollectAndList(fmlas []lg.Expr) []lg.Expr {
	return collectAndList(fmlas)
}

// CollectOr flattens nested Or formulas into a flat list.
func CollectOr(fmla lg.Expr) []lg.Expr {
	if o, ok := fmla.(*lg.Or); ok {
		var result []lg.Expr
		for _, t := range o.Terms {
			result = append(result, CollectOr(t)...)
		}
		return result
	}
	return []lg.Expr{fmla}
}

// DropUniversals strips leading ForAll quantifiers from a formula.
func DropUniversals(f lg.Expr) lg.Expr {
	return dropUniversals(f)
}

// NormalizeFreeVariables transforms a formula so free variables are renamed
// to V0, V1, ... in order of first occurrence.
// Returns (oldVars, newVars, normalizedFormula).
func NormalizeFreeVariables(node lg.Expr) ([]*lg.Variable, []*lg.Variable, lg.Expr) {
	subs := make(map[lg.NodeKey]lg.Expr)
	var oldVars, newVars []*lg.Variable
	seen := make(map[string]bool)

	// Collect free variables in order
	collectFreeVarsOrdered(node, nil, seen, &oldVars)

	for i, v := range oldVars {
		nv, _ := lg.NewVariable(fmt.Sprintf("V%d", i), v.VSort)
		newVars = append(newVars, nv)
		subs[lg.Key(v)] = nv
	}

	result, err := lu.Substitute(node, subs)
	if err != nil {
		return oldVars, newVars, node
	}
	return oldVars, newVars, result
}

func collectFreeVarsOrdered(node lg.Expr, bound map[string]struct{}, seen map[string]bool, result *[]*lg.Variable) {
	switch t := node.(type) {
	case *lg.Variable:
		if bound != nil {
			if _, ok := bound[t.Name]; ok {
				return
			}
		}
		if !seen[t.Name] {
			seen[t.Name] = true
			*result = append(*result, t)
		}
		return
	case *lg.ForAll:
		newBound := copyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		collectFreeVarsOrdered(t.Body, newBound, seen, result)
		return
	case *lg.Exists:
		newBound := copyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		collectFreeVarsOrdered(t.Body, newBound, seen, result)
		return
	case *lg.Lambda:
		newBound := copyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		collectFreeVarsOrdered(t.Body, newBound, seen, result)
		return
	case *lg.NamedBinder:
		newBound := copyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		collectFreeVarsOrdered(t.Body, newBound, seen, result)
		return
	}
	for _, c := range node.Children() {
		collectFreeVarsOrdered(c, bound, seen, result)
	}
}
