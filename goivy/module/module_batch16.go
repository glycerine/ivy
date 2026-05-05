package module

// Batch 1.6: Substitution and AST Traversal functions from ivy_logic_utils.py.
// Most functions already exist in logicutil or clauseops. This file fills the gaps.

import (
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
)

// SubstituteAstByName substitutes terms for variables in an AST, keyed by
// variable name (string). Bound variables shadow the substitution.
// Corresponds to Python's substitute_ast (ivy_logic_utils.py:160-170).
// Note: The ast package has SubstituteAst for ast.Node; this is for lg.Expr.
func SubstituteAstByName(ast lg.Expr, subs map[string]lg.Expr) lg.Expr {
	if len(subs) == 0 {
		return ast
	}
	if v, ok := ast.(*lg.Variable); ok {
		if r, ok := subs[v.Name]; ok {
			return r
		}
		return ast
	}
	if il.IsQuantifier(ast) {
		vars := il.BinderVars(ast)
		newSubs := make(map[string]lg.Expr, len(subs))
		bounds := make(map[string]bool, len(vars))
		for _, v := range vars {
			bounds[v.Name] = true
		}
		for k, v := range subs {
			if !bounds[k] {
				newSubs[k] = v
			}
		}
		if len(newSubs) == 0 {
			return ast
		}
		body := il.BinderBody(ast)
		newBody := SubstituteAstByName(body, newSubs)
		return il.CloneBinder(ast, vars, newBody)
	}
	args := il.NodeArgs(ast)
	if len(args) == 0 {
		return ast
	}
	newArgs := make([]lg.Expr, len(args))
	for i, a := range args {
		newArgs[i] = SubstituteAstByName(a, subs)
	}
	return il.CloneNode(ast, newArgs)
}

// ConstantsAst yields all constant symbols used in an AST.
// Unlike UsedConstants (which uses NodeKey), this yields the actual
// symbol objects. Corresponds to Python's constants_ast
// (ivy_logic_utils.py:501-507).
func ConstantsAst(ast lg.Expr) []*lg.Const {
	var result []*lg.Const
	constantsAstRec(ast, &result, make(map[string]bool))
	return result
}

func constantsAstRec(ast lg.Expr, result *[]*lg.Const, seen map[string]bool) {
	if c, ok := ast.(*lg.Const); ok {
		if !seen[c.Name] {
			seen[c.Name] = true
			*result = append(*result, c)
		}
		return
	}
	for _, arg := range il.NodeArgs(ast) {
		constantsAstRec(arg, result, seen)
	}
}

// IsGroundClause returns true if a literal clause has no free variables.
// Corresponds to Python's is_ground_clause (ivy_logic_utils.py:763-765).
func IsGroundClause(clause []*il.Literal) bool {
	for _, lit := range clause {
		vars := VariablesAST(lit.Atom)
		if len(vars) > 0 {
			return false
		}
	}
	return true
}

// TermEq returns true if two terms are structurally equal.
// Corresponds to Python's term_eq (ivy_logic_utils.py:773-775).
func TermEq(t1, t2 lg.Expr) bool {
	return t1.Equal(t2)
}

// AtomEq returns true if two atoms are structurally equal.
// Corresponds to Python's atom_eq (ivy_logic_utils.py:781-783).
func AtomEq(at1, at2 lg.Expr) bool {
	return at1.Equal(at2)
}

// LitEq returns true if two literals are structurally equal.
// Corresponds to Python's lit_eq (ivy_logic_utils.py:785-787).
func LitEq(lit1, lit2 *il.Literal) bool {
	return lit1.Polarity == lit2.Polarity && lit1.Atom.Equal(lit2.Atom)
}

// UsedVariableNamesAst returns the names of all used variables in an AST.
// Corresponds to Python's used_variable_names_ast
// (ivy_logic_utils.py:1206-1208).
func UsedVariableNamesAst(ast lg.Expr) []string {
	vars := lu.UsedVariables(ast)
	result := make([]string, 0, len(vars))
	for _, v := range vars {
		if vv, ok := v.(*lg.Variable); ok {
			result = append(result, vv.Name)
		}
	}
	return result
}

// VariablesDistinctAst renames variables in ast1 so they don't occur in ast2.
// Corresponds to Python's variables_distinct_ast
// (ivy_logic_utils.py:1210-1214).
func ModuleVariablesDistinctAst(ast1, ast2 lg.Expr) lg.Expr {
	vars1 := lu.UsedVariables(ast1)
	vars2 := lu.UsedVariables(ast2)
	renaming := ModuleDistinctVariableRenaming(vars1, vars2)
	if len(renaming) == 0 {
		return ast1
	}
	return SubstituteAstByName(ast1, renaming)
}

// DistinctVariableRenaming creates a renaming map from variable names to
// new variables, ensuring variables from vars1 don't clash with vars2.
// Corresponds to Python's distinct_variable_renaming
// (ivy_logic_utils.py, Batch 1.7).
func ModuleDistinctVariableRenaming(vars1, vars2 map[lg.NodeKey]lg.Expr) map[string]lg.Expr {
	// Collect all used names from vars2
	used := make(map[string]bool)
	for _, v := range vars2 {
		if vv, ok := v.(*lg.Variable); ok {
			used[vv.Name] = true
		}
	}
	// Map ALL vars1 variables (not just clashing ones), matching Python's
	// UniqueRenamer which always returns a mapping for every variable.
	result := make(map[string]lg.Expr)
	for _, v := range vars1 {
		vv, ok := v.(*lg.Variable)
		if !ok {
			continue
		}
		newName := vv.Name
		if used[newName] {
			for used[newName] {
				newName = newName + "'"
			}
		}
		used[newName] = true
		nv, _ := lg.NewVariable(newName, vv.VSort)
		result[vv.Name] = nv
	}
	return result
}

// RenameVariablesDistinctAsts renames variables so they don't occur in any
// of the given asts. Returns the renamed variables.
// Corresponds to Python's rename_variables_distinct_asts
// (ivy_logic_utils.py:1216-1220).
func RenameVariablesDistinctAsts(vars []*lg.Variable, asts []lg.Expr) []*lg.Variable {
	// Collect all used variables from all asts
	allUsed := make(map[lg.NodeKey]lg.Expr)
	for _, ast := range asts {
		for k, v := range lu.UsedVariables(ast) {
			allUsed[k] = v
		}
	}
	// Build vars1 map from the input variables
	vars1 := make(map[lg.NodeKey]lg.Expr, len(vars))
	for _, v := range vars {
		vars1[lg.Key(v)] = v
	}
	renaming := ModuleDistinctVariableRenaming(vars1, allUsed)
	result := make([]*lg.Variable, len(vars))
	for i, v := range vars {
		if r, ok := renaming[v.Name]; ok {
			if rv, ok := r.(*lg.Variable); ok {
				result[i] = rv
			} else {
				result[i] = v
			}
		} else {
			result[i] = v
		}
	}
	return result
}

// VariablesDistinctListAst renames variables in a list of ASTs so they
// don't occur in ast2.
// Corresponds to Python's variables_distinct_list_ast
// (ivy_logic_utils.py:1222-1225).
func VariablesDistinctListAst(astList []lg.Expr, ast2 lg.Expr) []lg.Expr {
	// Python: variables_distinct_ast(Atom('#',ast_list),ast2).args
	// We collect all variables from all ASTs and rename
	allVars := make(map[lg.NodeKey]lg.Expr)
	for _, ast := range astList {
		for k, v := range lu.UsedVariables(ast) {
			allVars[k] = v
		}
	}
	vars2 := lu.UsedVariables(ast2)
	renaming := ModuleDistinctVariableRenaming(allVars, vars2)
	if len(renaming) == 0 {
		return astList
	}
	result := make([]lg.Expr, len(astList))
	for i, ast := range astList {
		result[i] = SubstituteAstByName(ast, renaming)
	}
	return result
}

// ResortVar resorts a variable using a sort substitution map.
// Corresponds to Python's resort_var (ivy_logic_utils.py:415-416).
func ResortVar(v *lg.Variable, subs map[lg.NodeKey]lg.Sort) *lg.Variable {
	newSort := resortSortBySort(v.VSort, subs)
	if lg.SortEqual(newSort, v.VSort) {
		return v
	}
	nv, _ := lg.NewVariable(v.Name, newSort)
	return nv
}

// resortSortBySort resorts a sort using a substitution map.
func resortSortBySort(s lg.Sort, subs map[lg.NodeKey]lg.Sort) lg.Sort {
	k := lg.SortKey(s)
	if newSort, ok := subs[k]; ok {
		return newSort
	}
	if fs, ok := s.(*lg.FunctionSort); ok {
		dom := fs.Domain()
		rng := fs.Range()
		newDom := make([]lg.Sort, len(dom))
		for i, d := range dom {
			newDom[i] = resortSortBySort(d, subs)
		}
		newRng := resortSortBySort(rng, subs)
		all := append(newDom, newRng)
		ns, err := lg.NewFunctionSort(all...)
		if err != nil {
			return s
		}
		return ns
	}
	return s
}
