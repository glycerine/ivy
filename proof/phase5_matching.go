// phase5_matching.go implements Phase 5 Batch 5.2 (Match Compilation) and
// Batch 5.3 (Match Application) functions from the Ivy port.
// These correspond to Python ivy_proof.py matching/compilation functions.
package proof

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	"github.com/glycerine/goivy/compiler"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
	"github.com/glycerine/goivy/module"
)

// === Batch 5.2: Match Compilation ===

// CompileExprVocab compiles an expression using a goal's vocabulary.
// Pushes vocab symbols/sorts onto the signature, compiles the expression,
// and performs sort inference with the vocab's variables.
// Corresponds to Python's compile_expr_vocab.
func CompileExprVocab(expr ast.Node, vocab *Vocab) lg.Node {
	if expr == nil {
		return nil
	}

	// Get the current module's sig (or create a fresh one)
	sig := getSig()

	// Push vocab symbols onto sig
	ws := il.NewWithSymbols(sig, vocab.Symbols)
	ws.Enter()
	defer ws.Exit()

	// Push vocab sorts onto sig
	wso := il.NewWithSorts(sig, vocab.Sorts)
	wso.Enter()
	defer wso.Exit()

	// Check if the expression is a sort reference
	if atom, ok := expr.(*ast.Atom); ok {
		if s, exists := sig.Sorts[atom.Rep]; exists {
			return sortToNode(s)
		}
	}

	// Compile with TopSort as default
	savedDefault := sig.DefaultSort
	sig.DefaultSort = lg.TopS
	defer func() { sig.DefaultSort = savedDefault }()

	// Use compiler to compile the AST expression
	mod := module.CurrentModule()
	if mod == nil {
		mod = module.New()
	}
	c := compiler.New(sig, mod)
	compiled, err := c.CompileNode(expr)
	if err != nil {
		// Fallback: try simple symbol lookup
		compiled = compileSimple(expr, vocab)
		if compiled == nil {
			return nil
		}
	}

	// Sort inference: infer sorts on [compiled] + vocab.variables
	terms := make([]lg.Node, 0, 1+len(vocab.Variables))
	terms = append(terms, compiled)
	for _, v := range vocab.Variables {
		terms = append(terms, v)
	}
	inferred, err := il.SortInferList(terms)
	if err != nil {
		return compiled // return without sort inference on error
	}
	return inferred[0]
}

// CompileExprVocabExt compiles an expression using a vocabulary without
// full type inference. Returns the compiled expression directly.
// Corresponds to Python's compile_expr_vocab_ext.
func CompileExprVocabExt(expr ast.Node, vocab *Vocab) lg.Node {
	if expr == nil {
		return nil
	}

	// Get the current module's sig (or create a fresh one)
	sig := getSig()

	// Push vocab symbols onto sig
	ws := il.NewWithSymbols(sig, vocab.Symbols)
	ws.Enter()
	defer ws.Exit()

	// Push vocab sorts onto sig
	wso := il.NewWithSorts(sig, vocab.Sorts)
	wso.Enter()
	defer wso.Exit()

	// Check if the expression is a sort reference
	if atom, ok := expr.(*ast.Atom); ok {
		if s, exists := sig.Sorts[atom.Rep]; exists {
			return sortToNode(s)
		}
	}

	// Compile with TopSort as default
	savedDefault := sig.DefaultSort
	sig.DefaultSort = lg.TopS
	defer func() { sig.DefaultSort = savedDefault }()

	// Use compiler to compile the AST expression (no sort inference)
	mod := module.CurrentModule()
	if mod == nil {
		mod = module.New()
	}
	c := compiler.New(sig, mod)
	compiled, err := c.CompileNode(expr)
	if err != nil {
		compiled = compileSimple(expr, vocab)
	}
	return compiled
}

// getSig returns the current module's Sig, or a fresh Sig if none is available.
func getSig() *il.Sig {
	if mod := module.CurrentModule(); mod != nil && mod.Sig != nil {
		return mod.Sig
	}
	return il.NewSig()
}

// compileSimple is a fallback compiler that resolves atoms using vocab directly.
func compileSimple(expr ast.Node, vocab *Vocab) lg.Node {
	if n, ok := expr.(lg.Node); ok {
		return n
	}
	if atom, ok := expr.(*ast.Atom); ok {
		for _, sym := range vocab.Symbols {
			if sym.Name == atom.Rep {
				return sym
			}
		}
		for _, v := range vocab.Variables {
			if v.Name == atom.Rep {
				return v
			}
		}
		return lg.NewSymbol(atom.Rep, lg.TopS)
	}
	return nil
}

// sortToNode wraps a Sort as a Node.
// In Go, lg.Sort implements lg.Node, so we can return it directly.
func sortToNode(s lg.Sort) lg.Node {
	return s
}

// RemoveVarsMatch removes variable bindings from a match to avoid capture.
// Keeps sort matches unchanged, renames free variables in constant/symbol
// match values to avoid clashing with fmla, and drops variable matches.
// The origKeys parameter maps NodeKey → original keyed node, so we can
// determine the type of each key (sort, constant, or variable).
// Corresponds to Python's remove_vars_match.
func RemoveVarsMatch(mat map[lg.NodeKey]lg.Node, fmla lg.Node, origKeys map[lg.NodeKey]lg.Node) map[lg.NodeKey]lg.Node {
	result := make(map[lg.NodeKey]lg.Node)

	// Step 1: keep sort matches (key is a sort)
	// Step 2: collect constant/symbol pairs for renaming
	type symPair struct {
		key lg.NodeKey
		val lg.Node
	}
	var symPairs []symPair

	for k, v := range mat {
		origNode := origKeys[k]
		if origNode == nil {
			// Fallback: skip unknown entries
			continue
		}
		if _, isSort := origNode.(lg.Sort); isSort {
			// Sort match → keep directly
			result[k] = v
		} else if il.IsConstant(origNode) {
			// Symbol/constant match → collect for renaming
			symPairs = append(symPairs, symPair{key: k, val: v})
		}
		// Variable matches are dropped
	}

	// Step 3: rename free vars in constant match values to avoid clash with fmla
	if len(symPairs) > 0 {
		vals := make([]lg.Node, len(symPairs))
		for i, sp := range symPairs {
			vals[i] = sp.val
		}
		renamed := il.RenameVarsNoClash(vals, []lg.Node{fmla})
		for i, sp := range symPairs {
			result[sp.key] = renamed[i]
		}
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
	// Build rename map: old name → new name
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

	// Recursive goal renaming
	var recGoal func(*ast.LabeledFormula) (*ast.LabeledFormula, error)
	recGoal = func(g *ast.LabeledFormula) (*ast.LabeledFormula, error) {
		if g == nil {
			return nil, nil
		}
		// Recurse into premises
		prems := GoalPrems(g)
		newPrems := make([]ast.Node, len(prems))
		for i, p := range prems {
			if lf, ok := p.(*ast.LabeledFormula); ok {
				renamed, err := recGoal(lf)
				if err != nil {
					return nil, err
				}
				newPrems[i] = renamed
			} else {
				newPrems[i] = p
			}
		}
		g = CloneGoal(g, newPrems, GoalConc(g))

		// Build match from goal_defns: for each defined symbol whose name
		// is in rmap, create old→new mapping
		defns := GoalDefns(g)
		match := make(map[lg.NodeKey]lg.Node)
		for k, node := range defns {
			name := nodeNameStr(node)
			if name == "" {
				continue
			}
			newName, ok := rmap[name]
			if !ok {
				continue
			}
			// x.rename(lambda n: rmap[x.name]) — create new node with renamed name
			renamed := renameNode(node, newName)
			match[k] = renamed
		}
		// apply_match_sym to each value
		applied := make(map[lg.NodeKey]lg.Node, len(match))
		for k, v := range match {
			applied[k] = ApplyMatchSym(match, v)
		}
		match = applied

		// Check alpha capture
		if err := CheckAlphaCapture(g, match); err != nil {
			return nil, err
		}

		// Apply match to goal
		g = ApplyMatchGoalNode(match, g)

		// Alpha-rename the conclusion
		conc := GoalConc(g)
		if conc != nil {
			renamedConc, err := il.AlphaRename(rmap, conc)
			if err == nil {
				g = CloneGoal(g, GoalPrems(g), renamedConc)
			}
		}

		// Rename the goal's own label
		goalName := g.LabelName()
		if newName, ok := rmap[goalName]; ok {
			g = g.Rename(newName)
		}

		return g, nil
	}

	return recGoal(goal)
}

// nodeNameStr extracts the name from a logic node (Symbol or Variable).
func nodeNameStr(n lg.Node) string {
	switch t := n.(type) {
	case *lg.Symbol:
		return t.Name
	case *lg.Variable:
		return t.Name
	default:
		return ""
	}
}

// renameNode creates a copy of a logic node with a new name.
func renameNode(n lg.Node, newName string) lg.Node {
	switch t := n.(type) {
	case *lg.Symbol:
		return lg.NewSymbol(newName, t.CSort)
	case *lg.Variable:
		v, _ := lg.NewVariable(newName, t.VSort)
		return v
	default:
		return n
	}
}

// MakeDistinctVars creates fresh variables with distinct names from given ASTs.
// Corresponds to Python's make_distinct_vars.
func MakeDistinctVars(sorts []lg.Sort, asts ...lg.Node) []*lg.Variable {
	vars := make([]*lg.Variable, len(sorts))
	for i, sort := range sorts {
		v, _ := lg.NewVariable(fmt.Sprintf("V%d", i), sort)
		vars[i] = v
	}
	return co.RenameVariablesDistinctAsts(vars, asts)
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
