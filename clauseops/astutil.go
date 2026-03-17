package clauseops

import (
	"fmt"

	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
)

// SymbolsAST yields all constant symbols in a node (the function symbol
// of applications, plus any bare constants). This corresponds to Python's
// symbols_ast which yields the "rep" of app nodes.
func SymbolsAST(node lg.Node) []*lg.Const {
	result := make(map[*lg.Const]struct{})
	symbolsASTRec(node, result)
	out := make([]*lg.Const, 0, len(result))
	for c := range result {
		out = append(out, c)
	}
	return out
}

// UsedSymbolsAST returns the set of used constant symbols in a node.
func UsedSymbolsAST(node lg.Node) map[*lg.Const]struct{} {
	return usedSymbolsAST(node)
}

// VariablesAST yields free variables in a node (not bound variables).
// This matches Python's variables_ast which skips bound variables.
func VariablesAST(node lg.Node) []*lg.Var {
	result := make(map[*lg.Var]struct{})
	variablesASTRec(node, result, nil)
	out := make([]*lg.Var, 0, len(result))
	for v := range result {
		out = append(out, v)
	}
	return out
}

// UsedVariablesAST returns the set of free variables in a node.
func UsedVariablesAST(node lg.Node) map[*lg.Var]struct{} {
	result := make(map[*lg.Var]struct{})
	variablesASTRec(node, result, nil)
	return result
}

func variablesASTRec(node lg.Node, result map[*lg.Var]struct{}, bound map[string]struct{}) {
	switch t := node.(type) {
	case *lg.Var:
		if bound != nil {
			if _, ok := bound[t.Name]; ok {
				return
			}
		}
		result[t] = struct{}{}
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

// SubstituteConstantsAST substitutes constants by name. The map keys are
// constant name strings, values are replacement nodes.
// Variables are not affected.
func SubstituteConstantsAST(node lg.Node, subs map[string]lg.Node) lg.Node {
	if len(subs) == 0 {
		return node
	}
	return substituteConstantsRec(node, subs)
}

func substituteConstantsRec(node lg.Node, subs map[string]lg.Node) lg.Node {
	switch t := node.(type) {
	case *lg.Const:
		if r, ok := subs[t.Name]; ok {
			return r
		}
		return node
	case *lg.Var:
		return node
	case *lg.Apply:
		// Python substitute_constants_ast iterates ast.args (Terms only)
		// and calls ast.clone(new_args) which preserves Func unchanged.
		// The function head is NOT substituted.
		newTerms := make([]lg.Node, len(t.Terms))
		changed := false
		for i, arg := range t.Terms {
			newTerms[i] = substituteConstantsRec(arg, subs)
			if newTerms[i] != arg {
				changed = true
			}
		}
		if !changed {
			return node
		}
		result, err := lg.NewApply(t.Func, newTerms...)
		if err != nil {
			return node
		}
		return result
	case *il.Definition:
		lhs := substituteConstantsRec(t.Lhs, subs)
		rhs := substituteConstantsRec(t.Rhs, subs)
		if lhs == t.Lhs && rhs == t.Rhs {
			return node
		}
		return il.NewDefinition(lhs, rhs)
	}

	// Generic: recurse into children via CloneNode
	children := node.Children()
	if len(children) == 0 {
		return node
	}
	newChildren := make([]lg.Node, len(children))
	changed := false
	for i, c := range children {
		newChildren[i] = substituteConstantsRec(c, subs)
		if newChildren[i] != c {
			changed = true
		}
	}
	if !changed {
		return node
	}
	return il.CloneNode(node, newChildren)
}

// RenameAST renames symbol names in an AST. The map keys are old constant
// name strings, values are replacement Consts. Variables are not renamed.
func RenameAST(node lg.Node, subs map[string]*lg.Const) lg.Node {
	if len(subs) == 0 {
		return node
	}
	return renameASTRec(node, subs)
}

func renameASTRec(node lg.Node, subs map[string]*lg.Const) lg.Node {
	switch t := node.(type) {
	case *lg.Const:
		if r, ok := subs[t.Name]; ok {
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
	case *lg.Var:
		return node // variables not renamed
	case *lg.Apply:
		newFunc := renameASTRec(t.Func, subs)
		newTerms := make([]lg.Node, len(t.Terms))
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
	newChildren := make([]lg.Node, len(children))
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

// FreeVariablesAST is an alias for logicutil.FreeVariables.
func FreeVariablesAST(node lg.Node) map[*lg.Var]struct{} {
	return lu.FreeVariables(node)
}

// UsedAllVariablesAST returns all variables used in the node
// (both free and bound). This delegates to logicutil.UsedVariables.
func UsedAllVariablesAST(node lg.Node) map[*lg.Var]struct{} {
	return lu.UsedVariables(node)
}

// IsGroundAST returns true if the formula has no free variables.
func IsGroundAST(node lg.Node) bool {
	vars := VariablesAST(node)
	return len(vars) == 0
}

// --- Normalize / collect helpers ---

// CollectAndList flattens nested And formulas into a flat list.
func CollectAndList(fmlas []lg.Node) []lg.Node {
	return collectAndList(fmlas)
}

// CollectOr flattens nested Or formulas into a flat list.
func CollectOr(fmla lg.Node) []lg.Node {
	if o, ok := fmla.(*lg.Or); ok {
		var result []lg.Node
		for _, t := range o.Terms {
			result = append(result, CollectOr(t)...)
		}
		return result
	}
	return []lg.Node{fmla}
}

// DropUniversals strips leading ForAll quantifiers from a formula.
func DropUniversals(f lg.Node) lg.Node {
	return dropUniversals(f)
}

// NormalizeFreeVariables transforms a formula so free variables are renamed
// to V0, V1, ... in order of first occurrence.
// Returns (oldVars, newVars, normalizedFormula).
func NormalizeFreeVariables(node lg.Node) ([]*lg.Var, []*lg.Var, lg.Node) {
	subs := make(map[lg.Node]lg.Node)
	var oldVars, newVars []*lg.Var
	seen := make(map[string]bool)

	// Collect free variables in order
	collectFreeVarsOrdered(node, nil, seen, &oldVars)

	for i, v := range oldVars {
		nv, _ := lg.NewVar(fmt.Sprintf("V%d", i), v.VSort)
		newVars = append(newVars, nv)
		subs[v] = nv
	}

	result, err := lu.Substitute(node, subs)
	if err != nil {
		return oldVars, newVars, node
	}
	return oldVars, newVars, result
}

func collectFreeVarsOrdered(node lg.Node, bound map[string]struct{}, seen map[string]bool, result *[]*lg.Var) {
	switch t := node.(type) {
	case *lg.Var:
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
