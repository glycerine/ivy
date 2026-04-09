package module

// Batch 1.7: Subsumption and Advanced Clause Operations.
// Corresponds to ivy_logic_utils.py subsumption, renaming, and clause manipulation.

import (
	"fmt"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
)

// TermSubsume tries to make term1 match term2 (env only operates on term1).
// Returns true on success, updating env with variable bindings.
// Corresponds to Python's term_subsume (ivy_logic_utils.py:1119-1131).
func TermSubsume(term1, term2 lg.Expr, env map[string]lg.Expr) bool {
	if c, ok := term1.(*lg.Const); ok {
		c2, ok2 := term2.(*lg.Const)
		if !ok2 || c.Name != c2.Name {
			return false
		}
		args1 := il.NodeArgs(term1)
		args2 := il.NodeArgs(term2)
		if len(args1) != len(args2) {
			return false
		}
		for i := range args1 {
			if !TermSubsume(args1[i], args2[i], env) {
				return false
			}
		}
		return true
	}
	if v, ok := term1.(*lg.Variable); ok {
		if existing, exists := env[v.Name]; exists {
			return existing.Equal(term2)
		}
		env[v.Name] = term2
		return true
	}
	// Apply or other compound: check rep and args
	if app1, ok := term1.(*lg.Apply); ok {
		app2, ok2 := term2.(*lg.Apply)
		if !ok2 {
			return false
		}
		rep1 := il.GetAppRep(app1)
		rep2 := il.GetAppRep(app2)
		if rep1 == nil || rep2 == nil || rep1.Name != rep2.Name {
			return false
		}
		if len(app1.Terms) != len(app2.Terms) {
			return false
		}
		for i := range app1.Terms {
			if !TermSubsume(app1.Terms[i], app2.Terms[i], env) {
				return false
			}
		}
		return true
	}
	return term1.Equal(term2)
}

// LitSubsume tries to make lit1 match lit2.
// Corresponds to Python's lit_subsume (ivy_logic_utils.py:1133-1142).
func LitSubsume(lit1, lit2 *il.Literal, env map[string]lg.Expr) bool {
	if lit1.Polarity != lit2.Polarity {
		return false
	}
	rep1 := il.GetAppRep(lit1.Atom)
	rep2 := il.GetAppRep(lit2.Atom)
	if rep1 == nil || rep2 == nil || rep1.Name != rep2.Name {
		return false
	}
	args1 := il.NodeArgs(lit1.Atom)
	args2 := il.NodeArgs(lit2.Atom)
	if len(args1) != len(args2) {
		return false
	}
	for i := range args1 {
		if !TermSubsume(args1[i], args2[i], env) {
			return false
		}
	}
	return true
}

// AtomSubsume tries to make atom at1 match at2.
// Corresponds to Python's atom_subsume (ivy_logic_utils.py:1144-1154).
func AtomSubsume(at1, at2 lg.Expr) bool {
	env := make(map[string]lg.Expr)
	rep1 := il.GetAppRep(at1)
	rep2 := il.GetAppRep(at2)
	if rep1 == nil || rep2 == nil || rep1.Name != rep2.Name {
		return false
	}
	args1 := il.NodeArgs(at1)
	args2 := il.NodeArgs(at2)
	if len(args1) != len(args2) {
		return false
	}
	for i := range args1 {
		if !TermSubsume(args1[i], args2[i], env) {
			return false
		}
	}
	return true
}

// CommuteLit swaps the arguments of an equality literal.
// Corresponds to Python's commute_lit (ivy_logic_utils.py:1156-1157).
func CommuteLit(lit *il.Literal) *il.Literal {
	if eq, ok := lit.Atom.(*lg.Eq); ok {
		return il.NewLiteral(lit.Polarity, &lg.Eq{T1: eq.T2, T2: eq.T1})
	}
	args := il.NodeArgs(lit.Atom)
	if len(args) == 2 {
		rep := il.GetAppRep(lit.Atom)
		if rep != nil {
			return il.NewLiteral(lit.Polarity, il.Atom(rep, []lg.Expr{args[1], args[0]}))
		}
	}
	return lit
}

// ClauseSubsumeRecur recursively tries all possibilities of making
// cl1 subsume cl2. Returns true on success.
// Corresponds to Python's clause_subsume_recur (ivy_logic_utils.py:1159-1180).
func ClauseSubsumeRecur(cl1, cl2 []*il.Literal, env map[string]lg.Expr) bool {
	if len(cl1) == 0 {
		return true
	}
	lit := cl1[0]
	rest := cl1[1:]
	for i, lit2 := range cl2 {
		envCopy := copyEnv(env)
		if LitSubsume(lit, lit2, env) {
			// Remove lit2 from cl2
			remaining := make([]*il.Literal, 0, len(cl2)-1)
			remaining = append(remaining, cl2[:i]...)
			remaining = append(remaining, cl2[i+1:]...)
			if ClauseSubsumeRecur(rest, remaining, env) {
				return true
			}
		}
		restoreEnv(env, envCopy)

		// Try commuted equality
		rep := il.GetAppRep(lit.Atom)
		rep2 := il.GetAppRep(lit2.Atom)
		if lit.Polarity == lit2.Polarity && rep != nil && rep2 != nil &&
			rep.Name == "=" && rep2.Name == "=" {
			args2 := il.NodeArgs(lit2.Atom)
			if len(args2) == 2 {
				envCopy2 := copyEnv(env)
				if LitSubsume(lit, CommuteLit(lit2), env) {
					remaining := make([]*il.Literal, 0, len(cl2)-1)
					remaining = append(remaining, cl2[:i]...)
					remaining = append(remaining, cl2[i+1:]...)
					if ClauseSubsumeRecur(rest, remaining, env) {
						return true
					}
				}
				restoreEnv(env, envCopy2)
			}
		}
	}
	return false
}

func copyEnv(env map[string]lg.Expr) map[string]lg.Expr {
	c := make(map[string]lg.Expr, len(env))
	for k, v := range env {
		c[k] = v
	}
	return c
}

func restoreEnv(env, saved map[string]lg.Expr) {
	for k := range env {
		delete(env, k)
	}
	for k, v := range saved {
		env[k] = v
	}
}

// ClauseSubsume returns true iff cl2 is subsumed by cl1 (i.e., cl1 => cl2).
// Corresponds to Python's clause_subsume (ivy_logic_utils.py:1183-1188).
func ClauseSubsume(cl1, cl2 []*il.Literal) bool {
	env := make(map[string]lg.Expr)
	return ClauseSubsumeRecur(cl1, cl2, env)
}

// Subsume returns true iff cl is subsumed by a clause in clauses.
// Corresponds to Python's subsume (ivy_logic_utils.py:1190-1197).
func Subsume(clauses [][]*il.Literal, cl []*il.Literal) bool {
	for _, clp := range clauses {
		if ClauseSubsume(clp, cl) {
			return true
		}
	}
	return false
}

// RenameVariable renames a variable to a new name, preserving its sort.
// Corresponds to Python's rename_variable (ivy_logic_utils.py:1203-1204).
func RenameVariable(v *lg.Variable, name string) *lg.Variable {
	nv, _ := lg.NewVariable(name, v.VSort)
	return nv
}

// IsIndividualAst returns true if the AST has an individual (non-Boolean) sort.
// Corresponds to Python's is_individual_ast (ivy_logic_utils.py:1227-1228).
func IsIndividualAst(ast lg.Expr) bool {
	return il.IsIndividual(ast)
}

// OrClauses2 takes the logical or of two literal clause sets using
// a fresh Tseitin variable.
// Corresponds to Python's or_clauses2 (ivy_logic_utils.py:1232-1244).
func OrClauses2(clauses1, clauses2 [][]*il.Literal) [][]*il.Literal {
	if len(clauses1) == 0 || len(clauses2) == 0 {
		return nil
	}
	// Check for empty clause (false)
	for _, c := range clauses1 {
		if len(c) == 0 {
			return clauses2
		}
	}
	for _, c := range clauses2 {
		if len(c) == 0 {
			return clauses1
		}
	}
	// Create a fresh Tseitin variable
	used := collectUsedSymbolNames(clauses1, clauses2)
	rn := iu.NewUniqueRenamer("__ts", used)
	vName := rn.Rename("")
	v := lg.NewConst(vName, il.RelationSort(nil))
	posLit := il.NewLiteral(1, v)
	negLit := il.NewLiteral(0, v)

	var result [][]*il.Literal
	for _, c := range clauses1 {
		newC := make([]*il.Literal, 0, len(c)+1)
		newC = append(newC, posLit)
		newC = append(newC, c...)
		result = append(result, newC)
	}
	for _, c := range clauses2 {
		newC := make([]*il.Literal, 0, len(c)+1)
		newC = append(newC, negLit)
		newC = append(newC, c...)
		result = append(result, newC)
	}
	return result
}

func collectUsedSymbolNames(c1, c2 [][]*il.Literal) []string {
	names := make(map[string]bool)
	for _, cls := range [][]*il.Literal{} {
		for _, lit := range cls {
			syms := UsedSymbolsAST(lit.Atom)
			for _, s := range syms {
				names[lg.ExprName(s)] = true
			}
		}
	}
	for _, clauses := range [][][]*il.Literal{c1, c2} {
		for _, cls := range clauses {
			for _, lit := range cls {
				syms := UsedSymbolsAST(lit.Atom)
				for _, s := range syms {
					names[lg.ExprName(s)] = true
				}
			}
		}
	}
	result := make([]string, 0, len(names))
	for n := range names {
		result = append(result, n)
	}
	return result
}

// FixOrAnnot fixes the annotation of an or-clauses result.
// Corresponds to Python's fix_or_annot (ivy_logic_utils.py:1249-1256).
func FixOrAnnot(res *Clauses, vs []lg.Expr, args []*Clauses) *Clauses {
	if len(args) == 0 {
		return res
	}
	// Annotation handling is simplified: preserve first annotation
	annot := args[0].Annot
	return NewClauses(res.Fmlas, res.Defs, annot)
}

// ElimDefinitions eliminates definitions for the given symbols from clauses,
// converting them to constraints.
// Corresponds to Python's elim_definitions (ivy_logic_utils.py:1301-1310).
func ElimDefinitions(clauses *Clauses, dead []*lg.Const) *Clauses {
	fmlas := make([]lg.Expr, len(clauses.Fmlas))
	copy(fmlas, clauses.Fmlas)

	deadSet := make(map[lg.NodeKey]bool, len(dead))
	for _, sym := range dead {
		deadSet[lg.Key(sym)] = true
		if idx, ok := clauses.DefIdx[lg.Key(sym)]; ok && idx < len(clauses.Defs) {
			fmlas = append(fmlas, defToConstraint(clauses.Defs[idx]))
		}
	}

	var defs []*il.Definition
	for _, d := range clauses.Defs {
		rep := il.GetAppRep(d.Lhs)
		if rep == nil || !deadSet[lg.Key(rep)] {
			defs = append(defs, d)
		}
	}
	return NewClauses(fmlas, defs, clauses.Annot)
}

// RenameSymbols renames symbols in clauses using a renamer.
// Corresponds to Python's rename_symbols (ivy_logic_utils.py:1312-1314).
func RenameSymbols(rn *iu.UniqueRenamer, clauses *Clauses, toRename []*lg.Const) *Clauses {
	nameMap := make(map[lg.NodeKey]*lg.Const, len(toRename))
	for _, s := range toRename {
		newName := rn.Rename(s.Name)
		nameMap[lg.Key(s)] = lg.NewConst(newName, s.CSort)
	}
	return RenameClauses(clauses, nameMap)
}

// DebugClausesList prints clauses for debugging.
// Corresponds to Python's debug_clauses_list (ivy_logic_utils.py:1350-1357).
func DebugClausesList(cl []*Clauses) {
	for _, clauses := range cl {
		fmt.Println("definitions:")
		for _, df := range clauses.Defs {
			fmt.Println(df)
		}
		fmt.Println("fmlas:")
		for _, fmla := range clauses.Fmlas {
			fmt.Println(il.CloseFormula(fmla))
		}
	}
}

// TaggedOrClauses takes the logical or of clause sets, giving each disjunct
// a unique tag predicate that can be used to determine which disjunct is true
// in a model. See FindTrueDisjunct.
// Corresponds to Python's tagged_or_clauses (ivy_logic_utils.py:1391-1398).
// Unlike OrClausesTyped, this does NOT filter false branches (preserving
// disjunct indices) and uses an empty used-names set with prefix __to0.
func TaggedOrClauses(prefix string, args ...*Clauses) *Clauses {
	if len(args) == 0 {
		return TrueClauses(nil)
	}
	// Python: or_clauses_int(UniqueRenamer('__to0',dict()),args)
	rn := iu.NewUniqueRenamer("__to0", nil)
	res, vs, processedArgs := orClausesIntWithVs(rn, args)
	return fixOrAnnot(res, vs, processedArgs)
}

// FindTrueDisjunct finds the index of a true disjunct in a tagged disjunction.
// Corresponds to Python's find_true_disjunct (ivy_logic_utils.py:1400-1413).
func FindTrueDisjunct(clauses *Clauses, evalFun func(lg.Expr) bool) int {
	if len(clauses.Fmlas) == 0 {
		return -1
	}
	// The first formula should be a disjunction; check each disjunct
	fmla := clauses.Fmlas[0]
	if a, ok := fmla.(*lg.And); ok {
		// Actually Python checks fmlas[0].args (And terms)
		for idx, atom := range a.Terms {
			if evalFun(atom) {
				return idx
			}
		}
	} else if o, ok := fmla.(*lg.Or); ok {
		for idx, atom := range o.Terms {
			if evalFun(atom) {
				return idx
			}
		}
	}
	return -1
}

// EqcmUpd updates an equality class map. If lhs is in symset and rhs is
// a constant, removes lhs from symset and merges equivalence classes.
// Corresponds to Python's eqcm_upd (ivy_logic_utils.py:1447-1455).
func EqcmUpd(lhs, rhs lg.Expr, symset map[lg.NodeKey]bool, map2 map[lg.NodeKey][]lg.Expr) bool {
	lhsKey := lg.Key(lhs)
	if !symset[lhsKey] {
		return false
	}
	if !il.IsConstant(rhs) {
		return false
	}
	delete(symset, lhsKey)
	rhsKey := lg.Key(rhs)
	r := map2[rhsKey]
	l := map2[lhsKey]
	r = append(r, lhs)
	r = append(r, l...)
	map2[rhsKey] = r
	map2[lhsKey] = nil
	return true
}

// ExistsQuantClausesMap quantifies out symbols from clauses by renaming.
// Returns a substitution map and the renamed clauses.
// Corresponds to Python's exists_quant_clauses_map
// (ivy_logic_utils.py:1457-1475).
func ExistsQuantClausesMap(syms []*lg.Const, clauses *Clauses) (map[lg.NodeKey]*lg.Const, *Clauses) {
	used := collectAllUsedNames(clauses)
	symset := make(map[lg.NodeKey]bool, len(syms))
	for _, s := range syms {
		symset[lg.Key(s)] = true
	}
	map1 := make(map[lg.NodeKey]*lg.Const)
	map2 := make(map[lg.NodeKey][]lg.Expr)

	var defs []*il.Definition
	for _, df := range clauses.Defs {
		if !EqcmUpd(df.Lhs, df.Rhs, symset, map2) {
			if !EqcmUpd(df.Rhs, df.Lhs, symset, map2) {
				defs = append(defs, df)
			}
		}
	}

	// Build key→Const lookup for map2 keys (from definition args and syms)
	keyToConst := make(map[lg.NodeKey]*lg.Const)
	for _, df := range clauses.Defs {
		if c, ok := df.Lhs.(*lg.Const); ok {
			keyToConst[lg.Key(c)] = c
		}
		if c, ok := df.Rhs.(*lg.Const); ok {
			keyToConst[lg.Key(c)] = c
		}
	}
	for _, s := range syms {
		keyToConst[lg.Key(s)] = s
	}

	// Python: for v,w in map2.items(): for x in w: map1[x] = v
	// Map each element x in the equivalence class to its representative key v
	for vKey, w := range map2 {
		vConst := keyToConst[vKey]
		if vConst == nil {
			continue
		}
		for _, x := range w {
			if c, ok := x.(*lg.Const); ok {
				map1[lg.Key(c)] = vConst
			}
		}
	}

	newClauses := NewClauses(clauses.Fmlas, defs, nil)
	rn := iu.NewUniqueRenamer("__", used)
	// Python: unconditionally overwrite ALL syms with fresh names
	for _, s := range syms {
		newName := rn.Rename(s.Name)
		map1[lg.Key(s)] = lg.NewConst(newName, s.CSort)
	}
	return map1, RenameClauses(newClauses, map1)
}

func collectAllUsedNames(clauses *Clauses) []string {
	names := make(map[string]bool)
	for _, f := range clauses.Fmlas {
		for _, s := range UsedSymbolsAST(f) {
			names[lg.ExprName(s)] = true
		}
	}
	for _, d := range clauses.Defs {
		for _, s := range UsedSymbolsAST(d.Lhs) {
			names[lg.ExprName(s)] = true
		}
		for _, s := range UsedSymbolsAST(d.Rhs) {
			names[lg.ExprName(s)] = true
		}
	}
	result := make([]string, 0, len(names))
	for n := range names {
		result = append(result, n)
	}
	return result
}

// HasEnumeratedSort returns true if the symbol has an enumerated sort
// (either directly or as the range of a function sort).
// Corresponds to Python's has_enumerated_sort (ivy_logic_utils.py:1478-1481).
func HasEnumeratedSort(sig *il.Sig, sym *lg.Const) bool {
	sort := sym.CSort
	if il.IsEnumeratedSort(sort) {
		return true
	}
	if fs, ok := sort.(*lg.FunctionSort); ok {
		return il.IsEnumeratedSort(fs.Range())
	}
	return false
}

// Update ReduceClauses to use proper subsumption now that we have it
// (replaces the simplified version from litclause.go).
// This is the version that uses full clause subsumption.
func ReduceClausesFull(clauses [][]*il.Literal) [][]*il.Literal {
	var used [][]*il.Literal
	unexplored := make([][]*il.Literal, len(clauses))
	copy(unexplored, clauses)
	for len(unexplored) > 0 {
		cl := unexplored[0]
		unexplored = unexplored[1:]
		if !Subsume(used, cl) && !Subsume(unexplored, cl) {
			used = append(used, cl)
		}
	}
	return used
}

// Ensure lu import is used
var _ = lu.FreeVariables
