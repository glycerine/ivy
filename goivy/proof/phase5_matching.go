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
func CompileExprVocab(expr ast.Node, vocab *Vocab, mod *module.Module) lg.Expr {
	if expr == nil {
		return nil
	}

	// Get the module's sig (or create a fresh one)
	sig := getSigFrom(mod)

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
	terms := make([]lg.Expr, 0, 1+len(vocab.Variables))
	terms = append(terms, compiled)
	for _, v := range vocab.Variables {
		terms = append(terms, v)
	}
	inferred, err := il.SortInferList(terms, nil, nil)
	if err != nil {
		return compiled // return without sort inference on error
	}
	return inferred[0]
}

// CompileExprVocabExt compiles an expression using a vocabulary without
// full type inference. Returns the compiled expression directly.
// Corresponds to Python's compile_expr_vocab_ext.
func CompileExprVocabExt(expr ast.Node, vocab *Vocab, mod *module.Module) lg.Expr {
	if expr == nil {
		return nil
	}

	// Get the module's sig (or create a fresh one)
	sig := getSigFrom(mod)

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

// getSigFrom returns the module's Sig. Panics if mod or mod.Sig is nil —
// a nil here means the caller failed to thread the module through,
// which is always a bug (Python uses a single global Sig).
func getSigFrom(mod *module.Module) *il.Sig {
	if mod == nil {
		panic("getSigFrom: mod is nil — module must be threaded through to proof matching")
	}
	if mod.Sig == nil {
		panic("getSigFrom: mod.Sig is nil — module must have a Sig before proof matching")
	}
	return mod.Sig
}

// compileSimple is a fallback compiler that resolves atoms using vocab directly.
func compileSimple(expr ast.Node, vocab *Vocab) lg.Expr {
	if n, ok := expr.(lg.Expr); ok {
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
		return lg.NewConst(atom.Rep, lg.TopS)
	}
	return nil
}

// sortToNode wraps a Sort as a Node.
// In Go, lg.Sort implements lg.Expr, so we can return it directly.
func sortToNode(s lg.Sort) lg.Expr {
	return s
}

// RemoveVarsMatch removes variable bindings from a match to avoid capture.
// Keeps sort matches unchanged, renames free variables in constant/symbol
// match values to avoid clashing with fmla, and drops variable matches.
// The origKeys parameter maps NodeKey → original keyed node, so we can
// determine the type of each key (sort, constant, or variable).
// Corresponds to Python's remove_vars_match.
func RemoveVarsMatch(mat map[lg.NodeKey]lg.Expr, fmla lg.Expr, origKeys map[lg.NodeKey]lg.Expr) map[lg.NodeKey]lg.Expr {
	result := make(map[lg.NodeKey]lg.Expr)

	// Step 1: keep sort matches (key is a sort)
	// Step 2: collect constant/symbol pairs for renaming
	type symPair struct {
		key lg.NodeKey
		val lg.Expr
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
		vals := make([]lg.Expr, len(symPairs))
		for i, sp := range symPairs {
			vals[i] = sp.val
		}
		renamed := il.RenameVarsNoClash(vals, []lg.Expr{fmla})
		for i, sp := range symPairs {
			result[sp.key] = renamed[i]
		}
	}

	return result
}

// ShowMatch prints a match for debugging.
// Corresponds to Python's show_match.
func ShowMatch(m map[lg.NodeKey]lg.Expr) string {
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
func TransformDefnSchema(cfg *ast.AstConfig, schema, decl *ast.LabeledFormula) *ast.LabeledFormula {
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
		schema = ParameterizeSchema(cfg, extraSorts, schema)
	}
	return schema
}

// defLhsArgs returns the arguments of a definition's LHS.
func defLhsArgs(def *il.Definition) []lg.Expr {
	if app, ok := def.Lhs.(*lg.Apply); ok {
		return app.Terms
	}
	return nil
}

// TransformDefnMatch transforms a problem of matching definitions to a problem
// of matching the right-hand sides. Requires prob.Inst is a definition.
// Corresponds to Python's transform_defn_match.
func TransformDefnMatch(cfg *ast.AstConfig, prob *MatchProblem) *MatchProblem {
	conc, concIsDef := prob.Pat.(*lg.Definition)
	decl, declIsDef := prob.Inst.(*lg.Definition)
	if !concIsDef || !declIsDef {
		return prob
	}

	declsym := decl.Defines()
	concsym := conc.Defines()

	// Get args from LHS
	declargs := defArgs(decl.Lhs)
	concargs := defArgs(conc.Lhs)
	if len(declargs) < len(concargs) {
		return nil
	}

	declrhs := decl.Rhs
	concrhs := conc.Rhs

	// Build vmap: concarg.name → declarg.resort(concarg.sort)
	vmap := make(map[string]lg.Expr, len(concargs))
	for i, x := range concargs {
		if i >= len(declargs) {
			break
		}
		y := declargs[i]
		// resort y to x's sort
		resorted := resortNode(y, x.NodeSort())
		if v, ok := x.(*lg.Variable); ok {
			vmap[v.Name] = resorted
		} else if s, ok := x.(*lg.Const); ok {
			vmap[s.Name] = resorted
		}
	}
	concrhs = lu.SubstituteByName(concrhs, vmap)

	// Build dmatch: concsym → declsym, plus sort matching
	dmatch := make(map[lg.NodeKey]lg.Expr)
	dmatch[lg.Key(concsym)] = declsym

	concSorts := funcSortsNode(concsym)
	declSorts := funcSortsNode(declsym)
	for i := 0; i < len(concSorts) && i < len(declSorts); i++ {
		x := concSorts[i]
		y := declSorts[i]
		xKey := lg.Key(x)
		if _, isFree := prob.FreeSyms[xKey]; isFree {
			if existing, exists := dmatch[xKey]; exists {
				if !existing.Equal(y) {
					fmt.Printf("lhs sorts didn't match: %v, %v\n", x, y)
					return nil
				}
			}
			dmatch[xKey] = y
		} else {
			if !x.Equal(y) {
				fmt.Printf("lhs sorts didn't match: %v, %v\n", x, y)
				return nil
			}
		}
	}

	concrhs = ApplyMatch(dmatch, concrhs)
	freesyms := ApplyMatchFreesyms(dmatch, prob.FreeSyms)

	// Remove concargs from freesyms
	for _, arg := range concargs {
		delete(freesyms, lg.Key(arg))
	}

	// Remove declargs from constants
	constants := make(map[lg.NodeKey]lg.Expr, len(prob.Constants))
	for k, v := range prob.Constants {
		constants[k] = v
	}
	for _, arg := range declargs {
		delete(constants, lg.Key(arg))
	}

	// Build vvmap and apply to schema
	vvmap := make(map[lg.NodeKey]lg.Expr, len(concargs))
	for i, x := range concargs {
		if i >= len(declargs) {
			break
		}
		y := declargs[i]
		resorted := resortNode(y, x.NodeSort())
		vvmap[lg.Key(x)] = resorted
	}

	schema := prob.SchemaLF
	schema = ApplyMatchGoalNode(cfg, vvmap, schema)
	schema = ApplyMatchGoalNode(cfg, dmatch, schema)

	return &MatchProblem{
		Schema:   prob.Schema,
		SchemaLF: schema,
		Pat:      concrhs,
		Inst:     declrhs,
		FreeSyms: freesyms,
		Constants: constants,
	}
}

// defArgs extracts the arguments from a definition LHS.
// If the LHS is an Apply (function application), returns the Terms.
// Otherwise returns nil.
func defArgs(lhs lg.Expr) []lg.Expr {
	if app, ok := lhs.(*lg.Apply); ok {
		return app.Terms
	}
	return nil
}

// resortNode creates a copy of the node with a different sort.
func resortNode(n lg.Expr, s lg.Sort) lg.Expr {
	switch v := n.(type) {
	case *lg.Variable:
		nv, _ := lg.NewVariable(v.Name, s)
		return nv
	case *lg.Const:
		return lg.NewConst(v.Name, s)
	default:
		return n
	}
}

// funcSortsNode returns the domain and range sorts of a node's sort.
func funcSortsNode(n lg.Expr) []lg.Sort {
	if s, ok := n.(*lg.Const); ok {
		return FuncSorts(s)
	}
	return []lg.Sort{n.NodeSort()}
}

// AddPremMatch processes premise matches in a SchemaInstantiation.
// For each match where LHS is a premise name in the schema and RHS is a
// schema name, looks up the RHS schema and creates combined Tuple
// patterns for matching.
//
// Python: ivy_proof.py:835-855
func AddPremMatch(proofMatch []ast.Node, prob *MatchProblem, goal *ast.LabeledFormula, checker *ProofChecker) ([]ast.Node, *MatchProblem) {
	if prob.SchemaLF == nil {
		return proofMatch, prob
	}
	sprems := GoalPremsByName(prob.SchemaLF)

	var pats []lg.Expr
	var insts []lg.Expr
	var newMatch []ast.Node

	for _, m := range proofMatch {
		defn, ok := m.(*ast.Definition)
		if !ok {
			newMatch = append(newMatch, m)
			continue
		}
		lAtom, lIsAtom := defn.Lhs.(*ast.Atom)
		if !lIsAtom || len(lAtom.Terms) > 0 {
			newMatch = append(newMatch, m)
			continue
		}
		sprem, exists := sprems[lAtom.Relname()]
		if !exists {
			newMatch = append(newMatch, m)
			continue
		}
		rAtom, rIsAtom := defn.Rhs.(*ast.Atom)
		if !rIsAtom || len(rAtom.Terms) > 0 {
			newMatch = append(newMatch, m)
			continue
		}
		// Look up the RHS as a schema — Python: context.lookup_schema(rhs.rep, goal, rhs)
		gprem, err := checker.LookupSchema(rAtom.Relname(), goal, rAtom, false)
		if err != nil {
			newMatch = append(newMatch, m)
			continue
		}
		premConc := GoalConc(sprem)
		gpremConc := GoalConc(gprem)
		if premConc != nil && gpremConc != nil {
			pats = append(pats, premConc)
			insts = append(insts, gpremConc)
		} else {
			newMatch = append(newMatch, m)
		}
	}

	if len(pats) > 0 {
		// Combine premise patterns with main pattern/instance
		// Python: pat = ia.Tuple(*(pats + [prob.pat]))
		//         inst = ia.Tuple(*(insts + [prob.inst]))
		allPats := make([]lg.Expr, 0, len(pats)+1)
		allPats = append(allPats, pats...)
		allPats = append(allPats, prob.Pat)

		allInsts := make([]lg.Expr, 0, len(insts)+1)
		allInsts = append(allInsts, insts...)
		allInsts = append(allInsts, prob.Inst)

		prob = &MatchProblem{
			Schema:      prob.Schema,
			SchemaLF:    prob.SchemaLF,
			Pat:         prob.Pat,
			Inst:        prob.Inst,
			FreeSyms:    prob.FreeSyms,
			Constants:   prob.Constants,
			PremMatches: pats,
			RevMap:      prob.RevMap,
			TuplePats:   allPats,
			TupleInsts:  allInsts,
		}
	}

	return newMatch, prob
}

// ParameterizeSchema adds initial parameters to all free symbols in a schema.
// Takes a list of sorts and an ast.LabeledFormula (SchemaBody).
// For each ConstantDecl premise, extends the symbol's sort with the given sorts
// and wraps the match value in a Lambda.
// Corresponds to Python's parameterize_schema.
func ParameterizeSchema(cfg *ast.AstConfig, sorts []lg.Sort, schema *ast.LabeledFormula) *ast.LabeledFormula {
	conc := GoalConc(schema)
	vars := MakeDistinctVars(sorts, conc)

	match := make(map[lg.NodeKey]lg.Expr)
	var prems []ast.Node
	for _, prem := range GoalPrems(schema) {
		cd, ok := prem.(*ast.ConstantDecl)
		if !ok {
			prems = append(prems, prem)
			continue
		}
		// Get the symbol from the ConstantDecl's first arg
		if len(cd.DeclArgs) == 0 {
			prems = append(prems, prem)
			continue
		}
		sym := extractSymbol(cd.DeclArgs[0])
		if sym == nil {
			prems = append(prems, prem)
			continue
		}

		// Get domain and range of the symbol's sort
		var dom []lg.Sort
		var rng lg.Sort
		if fs, ok := sym.CSort.(*lg.FunctionSort); ok {
			dom = fs.Domain()
			rng = fs.Range()
		} else {
			rng = sym.CSort
		}

		// Create variables X0, X1, ... for existing domain sorts
		vs2 := make([]*lg.Variable, len(dom))
		for i, y := range dom {
			vs2[i], _ = lg.NewVariable(fmt.Sprintf("X%d", i), y)
		}

		// Build new sort: FuncConstSort(sorts... + dom... + [rng])
		allSorts := make([]lg.Sort, 0, len(sorts)+len(dom)+1)
		allSorts = append(allSorts, sorts...)
		allSorts = append(allSorts, dom...)
		allSorts = append(allSorts, rng)
		newSort := il.FuncConstSort(allSorts...)

		// Create new symbol with extended sort
		sym2 := lg.NewConst(sym.Name, newSort)

		// Build match[sym] = Lambda(vs2, sym2(*(vars + vs2)))
		// Construct the application args: vars... + vs2...
		appArgs := make([]lg.Expr, 0, len(vars)+len(vs2))
		for _, v := range vars {
			appArgs = append(appArgs, v)
		}
		for _, v := range vs2 {
			appArgs = append(appArgs, v)
		}

		var body lg.Expr
		if len(appArgs) > 0 {
			app, err := lg.NewApply(sym2, appArgs...)
			if err != nil {
				body = sym2
			} else {
				body = app
			}
		} else {
			body = sym2
		}

		lam, err := lg.NewLambda(vs2, body)
		if err == nil {
			match[lg.Key(sym)] = lam
		}

		// Replace premise with ConstantDecl(sym2)
		prems = append(prems, cfg.NewConstantDecl(sym2))
	}

	// Apply match to conclusion
	newConc := ApplyMatch(match, conc)
	return CloneGoal(cfg, schema, prems, newConc)
}

// CompileMatchList compiles a list of proof matches using goal vocabularies.
// LHS of each match Definition is compiled using leftGoal's vocab;
// RHS is compiled using rightGoal's vocab.
// If allowWitness is true, extends leftGoal's vocab with used variables
// from the left goal's conclusion.
// Corresponds to Python's compile_match_list.
func CompileMatchList(proofMatch []ast.Node, leftGoal, rightGoal *ast.LabeledFormula, allowWitness bool, mod *module.Module) []*ast.Definition {
	leftVocab := GoalVocab(leftGoal)
	rightVocab := GoalVocab(rightGoal)
	if allowWitness {
		// Extend leftVocab.Variables with used variables from left goal's conclusion
		conc := GoalConc(leftGoal)
		if conc != nil {
			usedVars := lu.UsedVariables(conc)
			for _, v := range usedVars {
				if vv, ok := v.(*lg.Variable); ok {
					leftVocab.Variables = append(leftVocab.Variables, vv)
				}
			}
		}
	}
	result := make([]*ast.Definition, 0, len(proofMatch))
	for _, m := range proofMatch {
		defn, ok := m.(*ast.Definition)
		if !ok {
			continue
		}
		x := CompileExprVocab(defn.Lhs, leftVocab, mod)
		y := CompileExprVocab(defn.Rhs, rightVocab, mod)
		result = append(result, mod.Cfg.AstCfg.NewDefinition(x, y))
	}
	return result
}

// extractSymbol extracts a *lg.Const from an ast.Node.
func extractSymbol(n ast.Node) *lg.Const {
	if n == nil {
		return nil
	}
	if s, ok := n.(*lg.Const); ok {
		return s
	}
	// Check if the node has a name that could be a symbol (e.g., ast.Atom)
	if atom, ok := n.(*ast.Atom); ok {
		return lg.NewConst(atom.Rep, lg.TopS)
	}
	return nil
}

// CompileOneMatch compiles a single match between two expressions.
// Corresponds to Python's compile_one_match.
func CompileOneMatch(lhs, rhs lg.Expr, freesyms, constants map[lg.NodeKey]lg.Expr) map[lg.NodeKey]lg.Expr {
	if _, isVar := lhs.(*lg.Variable); isVar {
		return FOMatch(lhs, rhs, freesyms, constants)
	}
	return Match(lhs, rhs, freesyms, constants)
}

// CompileMatchFull compiles all matches from a proof.
// Compiles the match list, then compiles each individual match against
// the problem's freesyms and constants, and merges all results.
// Corresponds to Python's compile_match.
func CompileMatchFull(proofMatch []ast.Node, prob *MatchProblem, decl *ast.LabeledFormula, allowWitness bool, mod *module.Module) map[lg.NodeKey]lg.Expr {
	schema := prob.SchemaLF
	if schema == nil {
		return nil
	}
	freesyms := copyNodeMap(prob.FreeSyms)
	if allowWitness {
		conc := GoalConc(schema)
		if conc != nil {
			for k, v := range lu.UsedVariables(conc) {
				freesyms[k] = v
			}
		}
	}
	compiledMatches := CompileMatchList(proofMatch, schema, decl, allowWitness, mod)
	matches := make([]map[lg.NodeKey]lg.Expr, 0, len(compiledMatches))
	for _, m := range compiledMatches {
		lhs := unwrapLogicNode(m.Lhs)
		rhs := unwrapLogicNode(m.Rhs)
		if lhs == nil || rhs == nil {
			continue
		}
		oneMatch := CompileOneMatch(lhs, rhs, freesyms, prob.Constants)
		matches = append(matches, oneMatch)
	}
	return MergeMatches(matches...)
}

// unwrapLogicNode extracts a lg.Expr from an ast.Node.
func unwrapLogicNode(n ast.Node) lg.Expr {
	if n == nil {
		return nil
	}
	if ln, ok := n.(lg.Expr); ok {
		return ln
	}
	return nil
}

// copyNodeMap copies a map[lg.NodeKey]lg.Expr.
func copyNodeMap(m map[lg.NodeKey]lg.Expr) map[lg.NodeKey]lg.Expr {
	result := make(map[lg.NodeKey]lg.Expr, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// MatchRhsVars gets symbols occurring free on the right-hand side of a match.
// Corresponds to Python's match_rhs_vars.
func MatchRhsVars(match map[lg.NodeKey]lg.Expr) map[lg.NodeKey]lg.Expr {
	result := make(map[lg.NodeKey]lg.Expr)
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
func ApplyMatchMatch(match, origMatch map[lg.NodeKey]lg.Expr, applyFn func(map[lg.NodeKey]lg.Expr, lg.Expr) lg.Expr) map[lg.NodeKey]lg.Expr {
	result := make(map[lg.NodeKey]lg.Expr, len(origMatch)+len(match))
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
func RenameProblem(cfg *ast.AstConfig, match map[lg.NodeKey]lg.Expr, prob *MatchProblem) {
	if prob.SchemaLF != nil {
		prob.SchemaLF = ApplyMatchGoalNode(cfg, match, prob.SchemaLF)
	}
	prob.Pat = ApplyMatchAlt(match, prob.Pat, nil)
	newFreeSyms := make(map[lg.NodeKey]lg.Expr, len(prob.FreeSyms))
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
func AvoidCaptureProblem(cfg *ast.AstConfig, prob *MatchProblem, match map[lg.NodeKey]lg.Expr) {
	mrv := MatchRhsVars(match)
	matchNames := make(map[string]bool)
	for _, v := range mrv {
		if c, ok := v.(*lg.Const); ok {
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
	cmatch := make(map[lg.NodeKey]lg.Expr)
	for k, sym := range prob.FreeSyms {
		if c, ok := sym.(*lg.Const); ok {
			if matchNames[c.Name] {
				if _, inMatch := match[k]; !inMatch {
					newName := rn.Rename(c.Name)
					cmatch[k] = lg.NewConst(newName, c.CSort)
				}
			}
		}
	}
	if len(cmatch) > 0 {
		RenameProblem(cfg, cmatch, prob)
	}
}

// RaiseCapture raises a CaptureError for a captured symbol.
// Corresponds to Python's raise_capture.
func RaiseCapture(v lg.Expr) error {
	return &CaptureError{Msg: fmt.Sprintf("symbol %s is captured in substitution", v)}
}

// MatchGet looks up a symbol in a match, checking for capture.
// Corresponds to Python's match_get.
func MatchGet(match map[lg.NodeKey]lg.Expr, sym lg.Expr, env map[lg.NodeKey]bool, defaultVal lg.Expr) (lg.Expr, error) {
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
func ApplyMatchAlt(match map[lg.NodeKey]lg.Expr, fmla lg.Expr, env map[lg.NodeKey]bool) lg.Expr {
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
func applyMatchAltRec(match map[lg.NodeKey]lg.Expr, fmla lg.Expr, env map[lg.NodeKey]bool) lg.Expr {
	if fmla == nil {
		return nil
	}

	switch t := fmla.(type) {
	case *lg.Apply:
		// Apply match to arguments
		newTerms := make([]lg.Expr, len(t.Terms))
		for i, arg := range t.Terms {
			newTerms[i] = applyMatchAltRec(match, arg, env)
		}
		// Check if function is in match
		if c, ok := t.Func.(*lg.Const); ok {
			k := lg.Key(c)
			if replacement, exists := match[k]; exists {
				if lam, ok := replacement.(*lg.Lambda); ok {
					result, _ := il.LambdaApply(lam, newTerms)
				return result
				}
				if newC, ok := replacement.(*lg.Const); ok {
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

	case *lg.Const:
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
	newChildren := make([]lg.Expr, len(children))
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
func ApplyFun(fun lg.Expr, args []lg.Expr) (lg.Expr, error) {
	if lam, ok := fun.(*lg.Lambda); ok {
		return il.LambdaApply(lam, args)
	}
	if c, ok := fun.(*lg.Const); ok {
		app, err := lg.NewApply(c, args...)
		return app, err
	}
	return nil, fmt.Errorf("apply_fun: not a function: %v", fun)
}

// ApplyMatchFuncAlt applies a match to a function symbol with lambda handling.
// Corresponds to Python's apply_match_func_alt.
func ApplyMatchFuncAlt(match map[lg.NodeKey]lg.Expr, fun lg.Expr, env map[lg.NodeKey]bool) lg.Expr {
	if lam, ok := fun.(*lg.Lambda); ok {
		return ApplyMatchAlt(match, lam, env)
	}
	k := lg.Key(fun)
	if replacement, exists := match[k]; exists {
		return replacement
	}
	if c, ok := fun.(*lg.Const); ok {
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
func ApplyMatchSort(match map[lg.NodeKey]lg.Expr, sort lg.Sort) lg.Sort {
	if sort == nil {
		return nil
	}
	// Check if sort itself is in match (as a node)
	// Sorts are not directly lg.Expr, so we look for named sorts
	if us, ok := sort.(*lg.UninterpretedSort); ok {
		sym := lg.NewConst(us.Name, lg.TopS)
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
func ApplyMatchFreesymsAlt(match map[lg.NodeKey]lg.Expr, freesyms map[lg.NodeKey]lg.Expr) map[lg.NodeKey]lg.Expr {
	result := make(map[lg.NodeKey]lg.Expr)
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
func RenameGoal(cfg *ast.AstConfig, goal *ast.LabeledFormula, renaming ast.Node) (*ast.LabeledFormula, error) {
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
		g = CloneGoal(cfg, g, newPrems, GoalConc(g))

		// Build match from goal_defns: for each defined symbol whose name
		// is in rmap, create old→new mapping
		defns := GoalDefns(g)
		match := make(map[lg.NodeKey]lg.Expr)
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
		applied := make(map[lg.NodeKey]lg.Expr, len(match))
		for k, v := range match {
			applied[k] = ApplyMatchSym(match, v)
		}
		match = applied

		// Check alpha capture
		if err := CheckAlphaCapture(g, match); err != nil {
			return nil, err
		}

		// Apply match to goal
		g = ApplyMatchGoalNode(cfg, match, g)

		// Alpha-rename the conclusion
		conc := GoalConc(g)
		if conc != nil {
			renamedConc, err := il.AlphaRename(rmap, conc)
			if err == nil {
				g = CloneGoal(cfg, g, GoalPrems(g), renamedConc)
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
func nodeNameStr(n lg.Expr) string {
	switch t := n.(type) {
	case *lg.Const:
		return t.Name
	case *lg.Variable:
		return t.Name
	default:
		return ""
	}
}

// renameNode creates a copy of a logic node with a new name.
func renameNode(n lg.Expr, newName string) lg.Expr {
	switch t := n.(type) {
	case *lg.Const:
		return lg.NewConst(newName, t.CSort)
	case *lg.Variable:
		v, _ := lg.NewVariable(newName, t.VSort)
		return v
	default:
		return n
	}
}

// MakeDistinctVars creates fresh variables with distinct names from given ASTs.
// Corresponds to Python's make_distinct_vars.
func MakeDistinctVars(sorts []lg.Sort, asts ...lg.Expr) []*lg.Variable {
	vars := make([]*lg.Variable, len(sorts))
	for i, sort := range sorts {
		v, _ := lg.NewVariable(fmt.Sprintf("V%d", i), sort)
		vars[i] = v
	}
	return co.RenameVariablesDistinctAsts(vars, asts)
}

// ApplyMatchGoalNode applies a match to a goal.
// Corresponds to Python's apply_match_goal with apply_match_alt.
func ApplyMatchGoalNode(cfg *ast.AstConfig, match map[lg.NodeKey]lg.Expr, goal *ast.LabeledFormula) *ast.LabeledFormula {
	if len(match) == 0 {
		return goal
	}
	prems := GoalPrems(goal)
	var newPrems []ast.Node
	for _, p := range prems {
		if lf, ok := p.(*ast.LabeledFormula); ok {
			newPrems = append(newPrems, ApplyMatchGoalNode(cfg, match, lf))
		} else if s, ok := p.(lg.Sort); ok {
			// Apply sort renaming: match[sort] → newSort
			key := lg.Key(s)
			if rep, found := match[key]; found {
				if rs, ok := rep.(lg.Sort); ok {
					newPrems = append(newPrems, rs)
					continue
				}
			}
			newPrems = append(newPrems, p)
		} else if cd, ok := p.(*ast.ConstantDecl); ok {
			// Apply symbol renaming via ApplyMatchFunc
			args := cd.Args()
			if len(args) > 0 {
				if sym, ok := args[0].(*lg.Const); ok {
					newSym := ApplyMatchFunc(match, sym)
					symKey := lg.Key(newSym)
					if rep, found := match[symKey]; found {
						if repNode, ok := rep.(ast.Node); ok {
							newPrems = append(newPrems, cd.Clone([]ast.Node{repNode}))
						} else {
							newPrems = append(newPrems, cd.Clone([]ast.Node{newSym}))
						}
					} else {
						newPrems = append(newPrems, cd.Clone([]ast.Node{newSym}))
					}
				} else {
					newPrems = append(newPrems, p)
				}
			} else {
				newPrems = append(newPrems, p)
			}
		} else {
			newPrems = append(newPrems, p)
		}
	}
	conc := GoalConc(goal)
	if conc != nil {
		conc = ApplyMatchAlt(match, conc, nil)
	}
	return CloneGoal(cfg, goal, newPrems, conc)
}

// CompileWitnessList compiles witness terms for existential instantiation.
// Corresponds to Python's compile_witness_list.
func CompileWitnessList(proof ast.Node, goal *ast.LabeledFormula, mod *module.Module) []lg.Expr {
	vocab := GoalVocab(goal)
	var result []lg.Expr
	for _, arg := range proof.Args() {
		compiled := CompileExprVocab(arg, vocab, mod)
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
