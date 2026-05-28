package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// SymbolsAST yields all symbols in a node (the function symbol
// of applications, plus any bare constants/variables). This corresponds to
// Python's symbols_ast which yields the "rep" of app nodes.
// Delegates to the faithful port il.UsedSymbolsAst (via symbols_ilu_ast).
func SymbolsAST(node Expr) []Expr {
	m := UsedSymbolsAst(node)
	result := make([]Expr, 0, m.Len())
	for _, sym := range m.All() {
		result = append(result, sym)
	}
	return result
}

// UsedSymbolsAST returns the set of used symbols in a node.
// Delegates to the faithful port il.UsedSymbolsAst (via symbols_ilu_ast).
// Returns InsMap preserving DFS first-occurrence order.
func UsedSymbolsAST(node Expr) *InsMap[NodeKey, Expr] {
	return UsedSymbolsAst(node)
}

// VariablesAST yields free variables in a node (not bound variables),
// in left-to-right traversal order with first-occurrence dedup.
// H6 / Python ivy_logic_utils.py:559-568 + iu.unique(): preserves traversal
// order so that downstream binder construction is deterministic across runs.
func VariablesAST(node Expr) []*LogicVariable {
	var out []*LogicVariable
	seen := make(map[string]bool)
	collectFreeVarsOrdered(node, nil, seen, &out)
	return out
}

// UsedVariablesAST returns the set of free variables in a node.
func UsedVariablesAST(node Expr) map[NodeKey]Expr {
	result := make(map[NodeKey]Expr)
	variablesASTRec(node, result, nil)
	return result
}

func variablesASTRec(node Expr, result map[NodeKey]Expr, bound map[string]struct{}) {
	switch t := node.(type) {
	case *LogicVariable:
		if bound != nil {
			if _, ok := bound[t.Name]; ok {
				return
			}
		}
		result[Key(t)] = t
		return
	case *ForAll:
		newBound := moduleCopyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		variablesASTRec(t.Body, result, newBound)
		return
	case *LogicExists:
		newBound := moduleCopyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		variablesASTRec(t.Body, result, newBound)
		return
	case *Lambda:
		newBound := moduleCopyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		variablesASTRec(t.Body, result, newBound)
		return
	case *LogicNamedBinder:
		newBound := moduleCopyStringSet(bound)
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

func moduleCopyStringSet(s map[string]struct{}) map[string]struct{} {
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
func SubstituteConstantsAST(node Node, subs map[NodeKey]Expr) Node {
	// Python: if ast is None: return None. This is required for optional
	// AST slots such as an unlabeled LabeledFormula's label.
	if node == nil {
		return nil
	}

	// Python: if is_constant(ast): return subs.get(ast.rep, ast)
	if sym, ok := node.(*Const); ok {
		if rep, found := subs[Key(sym)]; found {
			return rep
		}
		return sym
	}

	args := node.Args()
	xtracer.Trace("actions.substitute_constants_action ENTER type=%s nargs=%d",
		ShortTypeName(node), len(args))

	if len(args) == 0 {
		// Leaf non-constant (Variable, Atom label, etc.).
		// Python traces then clones with empty args.
		return node.Clone(args)
	}

	newArgs := make([]Node, len(args))
	for i, arg := range args {
		newArgs[i] = SubstituteConstantsAST(arg, subs)
	}
	return node.Clone(newArgs)
}

// SubstituteConstantsExpr is SubstituteConstantsAST for lg.Expr callers.
func SubstituteConstantsExpr(node Expr, subs map[NodeKey]Expr) Expr {
	return SubstituteConstantsAST(node, subs).(Expr)
}

// RenameAST renames symbols in an AST by structural identity.
// The map keys are lg.NodeKey (via lg.Key(sym)) for structural equality
// matching Python's recstruct-based Symbol lookup: subs.get(ast.rep, ast.rep).
// Variables are not renamed.
func RenameAST(node Expr, subs map[NodeKey]*Const) Expr {
	if len(subs) == 0 {
		return node
	}
	return renameASTRec(node, subs)
}

func renameASTRec(node Expr, subs map[NodeKey]*Const) Expr {
	switch t := node.(type) {
	case *Const:
		if r, ok := subs[Key(t)]; ok {
			// Preserve original sort if replacement has TopSort.
			// This matches Python where rename_ast substitutions carry
			// the original sort via sym.prefix('new_') etc.
			if _, isTop := r.CSort.(*TopSort); isTop && t.CSort != nil {
				if _, origIsTop := t.CSort.(*TopSort); !origIsTop {
					return NewConst(r.Name, t.CSort)
				}
			}
			return r
		}
		return node
	case *LogicVariable:
		return node // variables not renamed
	case *Apply:
		newFunc := renameASTRec(t.Func, subs)
		newTerms := make([]Expr, len(t.Terms))
		for i, arg := range t.Terms {
			newTerms[i] = renameASTRec(arg, subs)
		}
		return TryApply(newFunc, newTerms...)
	case *IvyDefinition:
		lhs := renameASTRec(t.Lhs, subs)
		rhs := renameASTRec(t.Rhs, subs)
		return NewIvyDefinition(lhs, rhs)
	}

	children := node.Children()
	if len(children) == 0 {
		return node
	}
	newChildren := make([]Expr, len(children))
	for i, c := range children {
		newChildren[i] = renameASTRec(c, subs)
	}
	return CloneNode(node, newChildren)
}

// collectConstsByName walks an expression tree and collects all *lg.Const
// nodes whose names appear in the nameSubs map. For each found constant,
// it adds a (name,sort)-keyed entry to the output map with the renamed
// constant preserving the original sort. This matches Python's recstruct
// __eq__ which compares (name, sort) tuples.
func collectConstsByName(node Expr, nameSubs map[string]string, out map[NodeKey]*Const) {
	if node == nil {
		return
	}
	switch t := node.(type) {
	case *Const:
		if newName, ok := nameSubs[t.Name]; ok {
			out[Key(t)] = NewConst(newName, t.CSort)
		}
	case *LogicVariable:
		// variables not collected
	case *Apply:
		collectConstsByName(t.Func, nameSubs, out)
		for _, arg := range t.Terms {
			collectConstsByName(arg, nameSubs, out)
		}
	case *IvyDefinition:
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
func RenameASTByName(node Expr, subs map[string]string) Expr {
	if len(subs) == 0 || node == nil {
		return node
	}
	constMap := make(map[NodeKey]*Const)
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
	constMap := make(map[NodeKey]*Const)
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
func FreeVariablesAST(node Expr) *Omap[NodeKey, Expr] {
	return FreeVariables(node)
}

// UsedAllVariablesAST returns all variables used in the node
// (both free and bound). This delegates to logicutil.UsedVariables.
func UsedAllVariablesAST(node Expr) map[NodeKey]Expr {
	return UsedVariables(node)
}

// IsGroundAST returns true if the formula has no free variables.
func IsGroundAST(node Expr) bool {
	vars := VariablesAST(node)
	return len(vars) == 0
}

// --- Normalize / collect helpers ---

// CollectAndList flattens nested And formulas into a flat list.
func CollectAndList(fmlas []Expr) []Expr {
	return collectAndList(fmlas)
}

// CollectOr flattens nested Or formulas into a flat list.
func CollectOr(fmla Expr) []Expr {
	if o, ok := fmla.(*LogicOr); ok {
		var result []Expr
		for _, t := range o.Terms {
			result = append(result, CollectOr(t)...)
		}
		return result
	}
	return []Expr{fmla}
}

// DropUniversals strips leading ForAll quantifiers from a formula.
func DropUniversals(f Expr) Expr {
	return dropUniversals(f)
}

// NormalizeFreeVariables transforms a formula so free variables are renamed
// to V0, V1, ... in order of first occurrence.
// Returns (oldVars, newVars, normalizedFormula).
func NormalizeFreeVariables(node Expr) ([]*LogicVariable, []*LogicVariable, Expr) {
	subs := make(map[NodeKey]Expr)
	var oldVars, newVars []*LogicVariable
	seen := make(map[string]bool)

	// Collect free variables in order
	collectFreeVarsOrdered(node, nil, seen, &oldVars)

	for i, v := range oldVars {
		nv, _ := NewVariable(fmt.Sprintf("V%d", i), v.VSort)
		newVars = append(newVars, nv)
		subs[Key(v)] = nv
	}

	result, err := Substitute(node, subs)
	if err != nil {
		return oldVars, newVars, node
	}
	return oldVars, newVars, result
}

func collectFreeVarsOrdered(node Expr, bound map[string]struct{}, seen map[string]bool, result *[]*LogicVariable) {
	switch t := node.(type) {
	case *LogicVariable:
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
	case *ForAll:
		newBound := moduleCopyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		collectFreeVarsOrdered(t.Body, newBound, seen, result)
		return
	case *LogicExists:
		newBound := moduleCopyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		collectFreeVarsOrdered(t.Body, newBound, seen, result)
		return
	case *Lambda:
		newBound := moduleCopyStringSet(bound)
		for _, v := range t.Variables {
			newBound[v.Name] = struct{}{}
		}
		collectFreeVarsOrdered(t.Body, newBound, seen, result)
		return
	case *LogicNamedBinder:
		newBound := moduleCopyStringSet(bound)
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
