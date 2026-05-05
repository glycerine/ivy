// symbols.go — mechanical port of Python ivy_logic_utils.py symbols_ilu_ast
// and derived helpers (symbols_asts, used_symbols_ast, used_symbols_asts, etc.).
package goivy

import (
	"iter"
)

// NodeRep returns the "rep" (function/operator head) of an app-like node.
// Mirrors Python's .rep property:
//   - Apply.rep  -> self.func   (Go: app.Func)
//   - Const.rep  -> self        (Go: the Const itself)
//   - NamedBinder(0 vars).rep -> self
//
// Returns nil for non-app nodes.
func IvyNodeRep(n Expr) Expr {
	switch t := n.(type) {
	case *Apply:
		return t.Func
	case *Const:
		return t
	case *LogicNamedBinder:
		if len(t.Variables) == 0 {
			return t
		}
	}
	return nil
}

// SymbolsIluAst yields all function symbols referenced in an AST node.
// Mechanical port of Python ivy_logic_utils.py:537 symbols_ilu_ast.
//
// When an app node's rep is a binder (ForAll, Exists, Lambda, NamedBinder, Some),
// it recurses into the binder's body instead of yielding the binder.
// Otherwise it yields the rep (the function symbol).
// Then it recurses into all args of the node.
func SymbolsIluAst(node Expr) iter.Seq[Expr] {
	return func(yield func(Expr) bool) {
		symbolsIluAstRec(node, yield)
	}
}

func symbolsIluAstRec(node Expr, yield func(Expr) bool) bool {
	if IsApp(node) {
		rep := IvyNodeRep(node)
		if IsBinder(rep) {
			if !symbolsIluAstRec(BinderBody(rep), yield) {
				return false
			}
		} else {
			if !yield(rep) {
				return false
			}
		}
	}
	for _, arg := range NodeArgs(node) {
		if !symbolsIluAstRec(arg, yield) {
			return false
		}
	}
	return true
}

// SymbolsAsts yields symbols from a list of ASTs.
// Matches Python: symbols_asts = apply_gen_to_list(symbols_ilu_ast)
func SymbolsAsts(nodes []Expr) iter.Seq[Expr] {
	return func(yield func(Expr) bool) {
		for _, n := range nodes {
			for sym := range SymbolsIluAst(n) {
				if !yield(sym) {
					return
				}
			}
		}
	}
}

// UsedSymbolsAst returns the set of symbols occurring in a single AST.
// Matches Python: used_symbols_ast = gen_to_ordered_dict(symbols_ilu_ast)
// Returns InsMap preserving DFS first-occurrence order.
func UsedSymbolsAst(node Expr) *InsMap[NodeKey, Expr] {
	result := NewInsMap[NodeKey, Expr]()
	for sym := range SymbolsIluAst(node) {
		result.Set(Key(sym), sym)
	}
	return result
}

// UsedSymbolsAsts returns the set of symbols occurring in a list of ASTs.
// Matches Python: used_symbols_asts = gen_to_ordered_dict(symbols_clause)
// Returns InsMap preserving DFS first-occurrence order.
func UsedSymbolsAsts(nodes []Expr) *InsMap[NodeKey, Expr] {
	result := NewInsMap[NodeKey, Expr]()
	for sym := range SymbolsAsts(nodes) {
		result.Set(Key(sym), sym)
	}
	return result
}

// UsedSymbolsInOrderAst returns unique symbols in order of first occurrence.
// Matches Python: used_symbols_in_order_ast = gen_unique(symbols_ilu_ast)
func UsedSymbolsInOrderAst(node Expr) []Expr {
	seen := make(map[NodeKey]bool)
	var result []Expr
	for sym := range SymbolsIluAst(node) {
		k := Key(sym)
		if !seen[k] {
			seen[k] = true
			result = append(result, sym)
		}
	}
	return result
}
