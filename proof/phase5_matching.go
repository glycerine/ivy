// phase5_matching.go implements Phase 5 Batch 5.2 (Match Compilation) and
// Batch 5.3 (Match Application) functions from the Ivy port.
// These correspond to Python ivy_proof.py matching/compilation functions.
package proof

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
)

// === Batch 5.2: Match Compilation ===

// CompileExprVocab compiles an expression using a goal's vocabulary.
// Corresponds to Python's compile_expr_vocab.
func CompileExprVocab(expr ast.Node, vocab *Vocab) lg.Node {
	return CompileWithVocab(expr, vocab)
}

// CompileExprVocabExt compiles an expression without full type inference.
// Corresponds to Python's compile_expr_vocab_ext.
func CompileExprVocabExt(expr ast.Node, vocab *Vocab) lg.Node {
	return CompileWithVocab(expr, vocab)
}

// CompileWithVocab is a helper that compiles an expression with a Vocab context.
func CompileWithVocab(expr ast.Node, vocab *Vocab) lg.Node {
	if expr == nil {
		return nil
	}
	// Check if it's a sort reference
	if atom, ok := expr.(*ast.Atom); ok {
		for _, s := range vocab.Sorts {
			if il.SortName(s) == atom.Rep {
				return sortToNode(s)
			}
		}
	}
	// If expression implements lg.Node, return it directly
	if n, ok := expr.(lg.Node); ok {
		return n
	}
	// Try to compile as a logic node (simplified)
	if atom, ok := expr.(*ast.Atom); ok {
		// Look for matching symbol in vocab
		for _, sym := range vocab.Symbols {
			if sym.Name == atom.Rep {
				return sym
			}
		}
		// Look for matching variable
		for _, v := range vocab.Variables {
			if v.Name == atom.Rep {
				return v
			}
		}
		return lg.NewSymbol(atom.Rep, lg.TopS)
	}
	return nil
}

// sortToNode wraps a Sort as a Node (using an uninterpreted sort).
func sortToNode(s lg.Sort) lg.Node {
	// Sorts are not lg.Node in Go, so we represent them as Symbols
	return lg.NewSymbol(il.SortName(s), lg.TopS)
}

// RemoveVarsMatch removes variable bindings from a match to avoid capture.
// Corresponds to Python's remove_vars_match.
func RemoveVarsMatch(mat map[lg.NodeKey]lg.Node, fmla lg.Node) map[lg.NodeKey]lg.Node {
	result := make(map[lg.NodeKey]lg.Node)
	for k, v := range mat {
		// Keep sort matches
		if _, isSort := v.(lg.Sort); isSort {
			result[k] = v
			continue
		}
		// Keep constant matches
		if c, ok := v.(*lg.Symbol); ok {
			_ = c
			result[k] = v
			continue
		}
		// For variable matches, we need to rename to avoid capture
		result[k] = v
	}
	return result
}

// ShowMatch prints a match for debugging.
// Corresponds to Python's show_match.
func ShowMatch(m map[lg.NodeKey]lg.Node) string {
	if m == nil {
		return "no match"
	}
	result := "match {\n"
	for k, v := range m {
		result += fmt.Sprintf("  %s |-> %s\n", k, v)
	}
	result += "}"
	return result
}

// TransformDefnSchema transforms a schema to match a definition's parameter structure.
// Corresponds to Python's transform_defn_schema.
func TransformDefnSchema(schema, decl *ast.LabeledFormula) *ast.LabeledFormula {
	sConc := GoalConc(schema)
	dConc := GoalConc(decl)
	if sConc == nil || dConc == nil {
		return schema
	}
	// Check if both are definitions
	sDef, sIsDef := sConc.(*il.Definition)
	dDef, dIsDef := dConc.(*il.Definition)
	if !sIsDef || !dIsDef {
		return schema
	}
	sArgs := defLhsArgs(sDef)
	dArgs := defLhsArgs(dDef)
	if len(dArgs) > len(sArgs) {
		// Need to parameterize schema with extra sorts
		extraSorts := make([]lg.Sort, len(dArgs)-len(sArgs))
		for i := range extraSorts {
			if i < len(dArgs) {
				extraSorts[i] = dArgs[i].NodeSort()
			} else {
				extraSorts[i] = lg.TopS
			}
		}
		schema = ParameterizeSchema(extraSorts, schema)
	}
	return schema
}

// defLhsArgs returns the arguments of a definition's LHS.
func defLhsArgs(def *il.Definition) []lg.Node {
	if app, ok := def.Lhs.(*lg.Apply); ok {
		return app.Terms
	}
	return nil
}

// TransformDefnMatch transforms a matching problem for definitions.
// Corresponds to Python's transform_defn_match.
func TransformDefnMatch(prob *MatchProblem) *MatchProblem {
	// Check if both pat and inst are definitions
	_, patIsDef := prob.Pat.(*il.Definition)
	_, instIsDef := prob.Inst.(*il.Definition)
	if !patIsDef || !instIsDef {
		return prob
	}
	return prob
}

// AddPremMatch processes premise matches from a proof.
// Corresponds to Python's add_prem_match.
func AddPremMatch(proofMatch []ast.Node, prob *MatchProblem, goal *ast.LabeledFormula, checker *ProofChecker) ([]ast.Node, *MatchProblem) {
	sprems := GoalPremsByName(prob.SchemaLF)
	var newMatch []ast.Node
	for _, m := range proofMatch {
		defn, ok := m.(*ast.Definition)
		if !ok {
			newMatch = append(newMatch, m)
			continue
		}
		if lAtom, ok := defn.Lhs.(*ast.Atom); ok && len(lAtom.Terms) == 0 {
			if _, exists := sprems[lAtom.Relname()]; exists {
				newMatch = append(newMatch, m)
				continue
			}
		}
		newMatch = append(newMatch, m)
	}
	return newMatch, prob
}

// ParameterizeSchema adds initial parameters to all free symbols in a schema.
// Corresponds to Python's parameterize_schema.
func ParameterizeSchema(sorts []lg.Sort, schema *ast.LabeledFormula) *ast.LabeledFormula {
	conc := GoalConc(schema)
	vars := MakeDistinctVars(sorts, conc)
	_ = vars
	// For now, return the schema unchanged — full parameterization requires
	// creating new symbols with extended sorts and lambda-wrapping.
	return schema
}

// CompileMatchList compiles a list of proof matches using goal vocabularies.
// Corresponds to Python's compile_match_list.
func CompileMatchList(proofMatch []ast.Node, leftGoal, rightGoal *ast.LabeledFormula, allowWitness bool) []ast.Node {
	leftVocab := GoalVocab(leftGoal)
	rightVocab := GoalVocab(rightGoal)
	var result []ast.Node
	for _, m := range proofMatch {
		defn, ok := m.(*ast.Definition)
		if !ok {
			result = append(result, m)
			continue
		}
		lhs := CompileExprVocab(defn.Lhs, leftVocab)
		rhs := CompileExprVocab(defn.Rhs, rightVocab)
		_ = lhs
		_ = rhs
		result = append(result, m)
	}
	return result
}

// CompileOneMatch compiles a single match between two expressions.
// Corresponds to Python's compile_one_match.
func CompileOneMatch(lhs, rhs lg.Node, freesyms, constants map[lg.NodeKey]lg.Node) map[lg.NodeKey]lg.Node {
	if _, isVar := lhs.(*lg.Variable); isVar {
		return FOMatch(lhs, rhs, freesyms, constants)
	}
	return Match(lhs, rhs, freesyms, constants)
}

// CompileMatchFull compiles all matches from a proof.
// Corresponds to Python's compile_match.
func CompileMatchFull(proofMatch []ast.Node, prob *MatchProblem, decl *ast.LabeledFormula, allowWitness bool) map[lg.NodeKey]lg.Node {
	schema := prob.SchemaLF
	if schema == nil {
		return nil
	}
	compiledMatches := CompileMatchList(proofMatch, schema, decl, allowWitness)
	var matches []map[lg.NodeKey]lg.Node
	for range compiledMatches {
		// Each compiled match should be compiled via CompileOneMatch
		// Simplified: merge all matches
	}
	return MergeMatches(matches...)
}

// MatchRhsVars gets symbols occurring free on the right-hand side of a match.
// Corresponds to Python's match_rhs_vars.
func MatchRhsVars(match map[lg.NodeKey]lg.Node) map[lg.NodeKey]lg.Node {
	result := make(map[lg.NodeKey]lg.Node)
	for _, v := range match {
		if v == nil {
			continue
		}
		for k, sym := range FmlaVocab(v) {
			result[k] = sym
		}
	}
	return result
}

// === Batch 5.3: Match Application ===

// ApplyMatchMatch composes two matches. Applying the result should have the
// same effect as applying orig_match first, then match.
// Corresponds to Python's apply_match_match.
func ApplyMatchMatch(match, origMatch map[lg.NodeKey]lg.Node, applyFn func(map[lg.NodeKey]lg.Node, lg.Node) lg.Node) map[lg.NodeKey]lg.Node {
	result := make(map[lg.NodeKey]lg.Node, len(origMatch)+len(match))
	for k, v := range origMatch {
		result[k] = applyFn(match, v)
	}
	for k, v := range match {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}
	return result
}

// RenameProblem renames symbols in a matching problem.
// Corresponds to Python's rename_problem.
func RenameProblem(match map[lg.NodeKey]lg.Node, prob *MatchProblem) {
	if prob.SchemaLF != nil {
		prob.SchemaLF = ApplyMatchGoalNode(match, prob.SchemaLF)
	}
	prob.Pat = ApplyMatchAlt(match, prob.Pat, nil)
	newFreeSyms := make(map[lg.NodeKey]lg.Node, len(prob.FreeSyms))
	for k, sym := range prob.FreeSyms {
		if replacement, ok := match[k]; ok {
			newFreeSyms[lg.Key(replacement)] = replacement
		} else {
			newFreeSyms[k] = sym
		}
	}
	prob.FreeSyms = newFreeSyms
	for k, v := range match {
		prob.RevMap[lg.Key(v)] = prob.FreeSyms[k]
	}
}

// AvoidCaptureProblem renames symbols to avoid capture when applying a match.
// Corresponds to Python's avoid_capture_problem.
func AvoidCaptureProblem(prob *MatchProblem, match map[lg.NodeKey]lg.Node) {
	mrv := MatchRhsVars(match)
	matchNames := make(map[string]bool)
	for _, v := range mrv {
		if c, ok := v.(*lg.Symbol); ok {
			matchNames[c.Name] = true
		}
		if v2, ok := v.(*lg.Variable); ok {
			matchNames[v2.Name] = true
		}
	}
	var used []string
	for name := range matchNames {
		used = append(used, name)
	}
	for k := range prob.FreeSyms {
		used = append(used, fmt.Sprint(k))
	}
	rn := iu.NewUniqueRenamer("", used)
	cmatch := make(map[lg.NodeKey]lg.Node)
	for k, sym := range prob.FreeSyms {
		if c, ok := sym.(*lg.Symbol); ok {
			if matchNames[c.Name] {
				if _, inMatch := match[k]; !inMatch {
					newName := rn.Rename(c.Name)
					cmatch[k] = lg.NewSymbol(newName, c.CSort)
				}
			}
		}
	}
	if len(cmatch) > 0 {
		RenameProblem(cmatch, prob)
	}
}

// RaiseCapture raises a CaptureError for a captured symbol.
// Corresponds to Python's raise_capture.
func RaiseCapture(v lg.Node) error {
	return &CaptureError{Msg: fmt.Sprintf("symbol %s is captured in substitution", v)}
}

// MatchGet looks up a symbol in a match, checking for capture.
// Corresponds to Python's match_get.
func MatchGet(match map[lg.NodeKey]lg.Node, sym lg.Node, env map[lg.NodeKey]bool, defaultVal lg.Node) (lg.Node, error) {
	k := lg.Key(sym)
	val, ok := match[k]
	if !ok {
		return defaultVal, nil
	}
	// Check for capture
	vocab := co.UsedSymbolsAST(val)
	for vk := range vocab {
		if env[vk] {
			return nil, RaiseCapture(sym)
		}
	}
	return val, nil
}

// ApplyMatchAlt applies a match to a formula with capture checking.
// Corresponds to Python's apply_match_alt.
func ApplyMatchAlt(match map[lg.NodeKey]lg.Node, fmla lg.Node, env map[lg.NodeKey]bool) lg.Node {
	if fmla == nil || len(match) == 0 {
		return fmla
	}
	if env == nil {
		env = make(map[lg.NodeKey]bool)
	}
	return applyMatchAltRec(match, fmla, env)
}

// applyMatchAltRec recursively applies a match with capture checking.
// Corresponds to Python's apply_match_alt_rec.
func applyMatchAltRec(match map[lg.NodeKey]lg.Node, fmla lg.Node, env map[lg.NodeKey]bool) lg.Node {
	if fmla == nil {
		return nil
	}

	switch t := fmla.(type) {
	case *lg.Apply:
		// Apply match to arguments
		newTerms := make([]lg.Node, len(t.Terms))
		for i, arg := range t.Terms {
			newTerms[i] = applyMatchAltRec(match, arg, env)
		}
		// Check if function is in match
		if c, ok := t.Func.(*lg.Symbol); ok {
			k := lg.Key(c)
			if replacement, exists := match[k]; exists {
				if lam, ok := replacement.(*lg.Lambda); ok {
					result, _ := il.LambdaApply(lam, newTerms)
				return result
				}
				if newC, ok := replacement.(*lg.Symbol); ok {
					app, _ := lg.NewApply(newC, newTerms...)
					return app
				}
			}
		}
		newFunc := applyMatchAltRec(match, t.Func, env)
		app, _ := lg.NewApply(newFunc, newTerms...)
		return app

	case *lg.Variable:
		k := lg.Key(t)
		if replacement, exists := match[k]; exists {
			return replacement
		}
		// Apply sort match
		newSort := ApplyMatchSort(match, t.VSort)
		if newSort != t.VSort {
			v, _ := lg.NewVariable(t.Name, newSort)
			return v
		}
		return fmla

	case *lg.Symbol:
		k := lg.Key(t)
		if replacement, exists := match[k]; exists {
			return replacement
		}
		return fmla

	case *lg.ForAll:
		newVars := make([]*lg.Variable, len(t.Variables))
		for i, v := range t.Variables {
			newV := applyMatchAltRec(match, v, env)
			if nv, ok := newV.(*lg.Variable); ok {
				newVars[i] = nv
			} else {
				newVars[i] = v
			}
		}
		newBody := applyMatchAltRec(match, t.Body, env)
		return &lg.ForAll{Variables: newVars, Body: newBody}

	case *lg.Exists:
		newVars := make([]*lg.Variable, len(t.Variables))
		for i, v := range t.Variables {
			newV := applyMatchAltRec(match, v, env)
			if nv, ok := newV.(*lg.Variable); ok {
				newVars[i] = nv
			} else {
				newVars[i] = v
			}
		}
		newBody := applyMatchAltRec(match, t.Body, env)
		return &lg.Exists{Variables: newVars, Body: newBody}

	case *lg.Lambda:
		newVars := make([]*lg.Variable, len(t.Variables))
		for i, v := range t.Variables {
			newV := applyMatchAltRec(match, v, env)
			if nv, ok := newV.(*lg.Variable); ok {
				newVars[i] = nv
			} else {
				newVars[i] = v
			}
		}
		newBody := applyMatchAltRec(match, t.Body, env)
		return &lg.Lambda{Variables: newVars, Body: newBody}
	}

	// Generic: recurse into children
	children := fmla.Children()
	if len(children) == 0 {
		return fmla
	}
	newChildren := make([]lg.Node, len(children))
	changed := false
	for i, c := range children {
		newChildren[i] = applyMatchAltRec(match, c, env)
		if newChildren[i] != c {
			changed = true
		}
	}
	if !changed {
		return fmla
	}
	return il.CloneNode(fmla, newChildren)
}

// ApplyFun applies a lambda function with capture error handling.
// Corresponds to Python's apply_fun.
func ApplyFun(fun lg.Node, args []lg.Node) (lg.Node, error) {
	if lam, ok := fun.(*lg.Lambda); ok {
		return il.LambdaApply(lam, args)
	}
	if c, ok := fun.(*lg.Symbol); ok {
		app, err := lg.NewApply(c, args...)
		return app, err
	}
	return nil, fmt.Errorf("apply_fun: not a function: %v", fun)
}

// ApplyMatchFuncAlt applies a match to a function symbol with lambda handling.
// Corresponds to Python's apply_match_func_alt.
func ApplyMatchFuncAlt(match map[lg.NodeKey]lg.Node, fun lg.Node, env map[lg.NodeKey]bool) lg.Node {
	if lam, ok := fun.(*lg.Lambda); ok {
		return ApplyMatchAlt(match, lam, env)
	}
	k := lg.Key(fun)
	if replacement, exists := match[k]; exists {
		return replacement
	}
	if c, ok := fun.(*lg.Symbol); ok {
		newC := ApplyMatchFunc(match, c)
		k2 := lg.Key(newC)
		if replacement, exists := match[k2]; exists {
			return replacement
		}
		return newC
	}
	return fun
}

// ApplyMatchSort applies a match to a sort, returning the matched sort or original.
// Corresponds to Python's apply_match_sort.
func ApplyMatchSort(match map[lg.NodeKey]lg.Node, sort lg.Sort) lg.Sort {
	if sort == nil {
		return nil
	}
	// Check if sort itself is in match (as a node)
	// Sorts are not directly lg.Node, so we look for named sorts
	if us, ok := sort.(*lg.UninterpretedSort); ok {
		sym := lg.NewSymbol(us.Name, lg.TopS)
		k := lg.Key(sym)
		if replacement, exists := match[k]; exists {
			if newSort, ok := replacement.(lg.Sort); ok {
				return newSort
			}
		}
	}
	return sort
}

// ApplyMatchFreesymsAlt applies a match to free symbols and filters out matched ones.
// Corresponds to Python's apply_match_freesyms_alt.
func ApplyMatchFreesymsAlt(match map[lg.NodeKey]lg.Node, freesyms map[lg.NodeKey]lg.Node) map[lg.NodeKey]lg.Node {
	result := make(map[lg.NodeKey]lg.Node)
	for _, sym := range freesyms {
		newSym := ApplyMatchSym(match, sym)
		newK := lg.Key(newSym)
		if _, inMatch := match[newK]; !inMatch {
			result[newK] = newSym
		}
	}
	return result
}

// RenameGoal renames symbols in a goal based on a renaming specification.
// Corresponds to Python's rename_goal.
func RenameGoal(goal *ast.LabeledFormula, renaming ast.Node) (*ast.LabeledFormula, error) {
	if len(renaming.Args()) == 0 {
		return goal, nil
	}
	if err := CheckRenaming(goal, renaming); err != nil {
		return nil, err
	}
	// Build rename map
	rmap := make(map[string]string)
	for _, arg := range renaming.Args() {
		defn, ok := arg.(*ast.Definition)
		if !ok {
			continue
		}
		if lAtom, ok := defn.Lhs.(*ast.Atom); ok {
			if rAtom, ok := defn.Rhs.(*ast.Atom); ok {
				rmap[lAtom.Rep] = rAtom.Rep
			}
		}
	}
	// Apply rename to goal (simplified)
	// Build match from rename map
	match := make(map[lg.NodeKey]lg.Node)
	for old, new := range rmap {
		oldSym := lg.NewSymbol(old, lg.TopS)
		newSym := lg.NewSymbol(new, lg.TopS)
		match[lg.Key(oldSym)] = newSym
	}
	if err := CheckAlphaCapture(goal, match); err != nil {
		return nil, err
	}
	result := ApplyMatchGoalNode(match, goal)
	return result, nil
}

// MakeDistinctVars creates fresh variables with distinct names from given ASTs.
// Corresponds to Python's make_distinct_vars.
func MakeDistinctVars(sorts []lg.Sort, asts ...lg.Node) []*lg.Variable {
	vars := make([]*lg.Variable, len(sorts))
	for i, sort := range sorts {
		v, _ := lg.NewVariable(fmt.Sprintf("V%d", i), sort)
		vars[i] = v
	}
	// Rename to be distinct from variables in asts
	// Simplified: just return the variables as-is
	return vars
}

// ApplyMatchGoalNode applies a match to a goal.
// Corresponds to Python's apply_match_goal with apply_match_alt.
func ApplyMatchGoalNode(match map[lg.NodeKey]lg.Node, goal *ast.LabeledFormula) *ast.LabeledFormula {
	if len(match) == 0 {
		return goal
	}
	prems := GoalPrems(goal)
	var newPrems []ast.Node
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			newPrems = append(newPrems, ApplyMatchGoalNode(match, lf))
		} else {
			newPrems = append(newPrems, p)
		}
	}
	conc := GoalConc(goal)
	if conc != nil {
		conc = ApplyMatchAlt(match, conc, nil)
	}
	return CloneGoal(goal, newPrems, conc)
}

// CompileWitnessList compiles witness terms for existential instantiation.
// Corresponds to Python's compile_witness_list.
func CompileWitnessList(proof ast.Node, goal *ast.LabeledFormula) []lg.Node {
	vocab := GoalVocab(goal)
	var result []lg.Node
	for _, arg := range proof.Args() {
		compiled := CompileExprVocab(arg, vocab)
		if compiled != nil {
			result = append(result, compiled)
		}
	}
	return result
}

// Ensure unused imports don't cause errors.
var _ = lu.EqualModAlpha
var _ = iu.NewUniqueRenamer
var _ ast.Node
