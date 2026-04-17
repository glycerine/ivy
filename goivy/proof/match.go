package proof

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
)

// MatchProblem describes a schema-instantiation matching problem.
type MatchProblem struct {
	Schema      lg.Expr          // the schema being instantiated (as logic node)
	SchemaLF    *ast.LabeledFormula // the schema as a LabeledFormula (may be nil)
	Pat         lg.Expr          // pattern to match (conclusion of schema)
	Inst        lg.Expr          // instance to match against (conclusion of goal)
	FreeSyms    map[lg.NodeKey]lg.Expr // free symbols in the schema
	Constants   map[lg.NodeKey]lg.Expr // constants (free variables of the goal)
	PremMatches []lg.Expr        // premise match patterns (for Tuple matching)
	RevMap      map[lg.NodeKey]lg.Expr // reverse mapping for nonce symbols
	// TuplePats and TupleInsts are set by AddPremMatch when premise
	// matching creates combined premise+conclusion patterns.
	// nil means no tuple matching; non-nil activates per-element matching.
	// Python: when add_prem_match creates ia.Tuple patterns.
	TuplePats   []lg.Expr
	TupleInsts  []lg.Expr
}

// NewMatchProblem creates a MatchProblem.
func NewMatchProblem(schema, pat, inst lg.Expr, freesyms, constants map[lg.NodeKey]lg.Expr) *MatchProblem {
	fs := make(map[lg.NodeKey]lg.Expr, len(freesyms))
	for k, v := range freesyms {
		fs[k] = v
	}
	cs := make(map[lg.NodeKey]lg.Expr, len(constants))
	for k, v := range constants {
		cs[k] = v
	}
	return &MatchProblem{
		Schema:    schema,
		Pat:       pat,
		Inst:      inst,
		FreeSyms:  fs,
		Constants: cs,
		RevMap:    make(map[lg.NodeKey]lg.Expr),
	}
}

func (mp *MatchProblem) String() string {
	syms := make([]string, 0, len(mp.FreeSyms))
	for s := range mp.FreeSyms {
		syms = append(syms, fmt.Sprint(s))
	}
	return fmt.Sprintf("{pat:%s,inst:%s,freesyms:%v}", mp.Pat, mp.Inst, syms)
}

// --- Match functions ---

// FuncSorts returns the domain sorts followed by the range sort of a constant.
func FuncSorts(c *lg.Const) []lg.Sort {
	if fs, ok := c.CSort.(*lg.FunctionSort); ok {
		dom := fs.Domain()
		result := make([]lg.Sort, len(dom)+1)
		copy(result, dom)
		result[len(dom)] = fs.Range()
		return result
	}
	return []lg.Sort{c.CSort}
}

// FuncsMatch checks whether two constants match structurally:
// same name, same arity, and non-free sorts agree.
func FuncsMatch(pat, inst *lg.Const, freesyms map[lg.NodeKey]lg.Expr) bool {
	ps := FuncSorts(pat)
	is := FuncSorts(inst)
	if pat.Name != inst.Name || len(ps) != len(is) {
		return false
	}
	for i := range ps {
		if freesyms[lg.Key(ps[i])] == nil && !ps[i].Equal(is[i]) {
			return false
		}
	}
	return true
}

// HeadsMatch checks if the heads of two terms match.
// Same top-level operator and same number of arguments.
// Quantifiers do not match anything.
// A function symbol matches if it has the same name and non-free sorts agree.
func HeadsMatch(pat, inst lg.Expr, freesyms map[lg.NodeKey]lg.Expr) bool {
	if il.IsApp(pat) && il.IsApp(inst) {
		pc := appFunc(pat)
		ic := appFunc(inst)
		if pc != nil && ic != nil && freesyms[lg.Key(pc)] == nil {
			return FuncsMatch(pc, ic, freesyms) && len(il.NodeArgs(pat)) == len(il.NodeArgs(inst))
		}
		return false
	}
	if il.IsApp(pat) || il.IsQuantifier(pat) {
		return false
	}
	if il.IsApp(inst) || il.IsQuantifier(inst) {
		return false
	}
	return sameNodeType(pat, inst) && len(il.NodeArgs(pat)) == len(il.NodeArgs(inst))
}

// TermSorts returns the domain+range sorts of the head function of a term,
// or the single sort for a variable.
func TermSorts(term lg.Expr) []lg.Sort {
	if c := appFunc(term); c != nil {
		return FuncSorts(c)
	}
	if v, ok := term.(*lg.Variable); ok {
		return []lg.Sort{v.VSort}
	}
	return nil
}

// LambdaSorts returns the variable sorts followed by the body sort of a lambda.
func LambdaSorts(lam *lg.Lambda) []lg.Sort {
	result := make([]lg.Sort, len(lam.Variables)+1)
	for i, v := range lam.Variables {
		result[i] = v.VSort
	}
	result[len(lam.Variables)] = lam.Body.NodeSort()
	return result
}

// MatchSort matches a sort: if pat is free, map it to inst; otherwise require equality.
func MatchSort(pat, inst lg.Sort, freesyms map[lg.NodeKey]lg.Expr) map[lg.NodeKey]lg.Expr {
	if freesyms[lg.Key(pat)] != nil {
		return map[lg.NodeKey]lg.Expr{lg.Key(pat): inst}
	}
	if pat.Equal(inst) {
		return map[lg.NodeKey]lg.Expr{}
	}
	return nil
}

// MergeMatches merges multiple match maps. Returns nil if any match is nil
// or if there is a conflicting assignment.
func MergeMatches(matches ...map[lg.NodeKey]lg.Expr) map[lg.NodeKey]lg.Expr {
	if len(matches) == 0 {
		return map[lg.NodeKey]lg.Expr{}
	}
	for _, m := range matches {
		if m == nil {
			return nil
		}
	}
	res := make(map[lg.NodeKey]lg.Expr)
	for k, v := range matches[0] {
		res[k] = v
	}
	for _, m := range matches[1:] {
		for k, v := range m {
			if prev, ok := res[k]; ok {
				if !EquivAlpha(prev, v) {
					return nil
				}
			} else {
				res[k] = v
			}
		}
	}
	return res
}

// EquivAlpha checks if two closed terms are equivalent modulo alpha conversion.
func EquivAlpha(x, y lg.Expr) bool {
	if x.Equal(y) {
		return true
	}
	xLam, xOk := x.(*lg.Lambda)
	yLam, yOk := y.(*lg.Lambda)
	if xOk && yOk && len(xLam.Variables) == len(yLam.Variables) {
		// substitute y's variables with x's variables
		subs := make(map[lg.NodeKey]lg.Expr)
		for i, v := range yLam.Variables {
			subs[lg.Key(v)] = xLam.Variables[i]
		}
		subBody, err := lu.Substitute(yLam.Body, subs)
		if err != nil {
			return false
		}
		return xLam.Body.Equal(subBody)
	}
	return false
}

// Match matches an instance to a pattern.
// Returns an assignment sigma to freesyms such that sigma(pat) =_alpha inst.
// Returns nil on failure.
func Match(pat, inst lg.Expr, freesyms, constants map[lg.NodeKey]lg.Expr) map[lg.NodeKey]lg.Expr {
	if il.IsQuantifier(pat) {
		return MatchQuants(pat, inst, freesyms, constants)
	}
	if HeadsMatch(pat, inst, freesyms) {
		patArgs := il.NodeArgs(pat)
		instArgs := il.NodeArgs(inst)
		matches := make([]map[lg.NodeKey]lg.Expr, 0, len(patArgs)+4)
		for i := range patArgs {
			matches = append(matches, Match(patArgs[i], instArgs[i], freesyms, constants))
		}
		// match sorts
		ps := TermSorts(pat)
		is := TermSorts(inst)
		for i := range ps {
			if i < len(is) {
				matches = append(matches, MatchSort(ps[i], is[i], freesyms))
			}
		}
		if v, ok := pat.(*lg.Variable); ok {
			matches = append(matches, map[lg.NodeKey]lg.Expr{lg.Key(v): inst})
		}
		return MergeMatches(matches...)
	}

	// If pat is a free application, try to extract lambda
	if il.IsApp(pat) {
		c := appFunc(pat)
		if c != nil && freesyms[lg.Key(c)] != nil {
			patArgs := il.NodeArgs(pat)
			B := ExtractTerms(inst, patArgs, constants)
			if B != nil {
				matches := []map[lg.NodeKey]lg.Expr{{lg.Key(c): B}}
				ps := TermSorts(pat)
				ls := LambdaSorts(B)
				for i := range ps {
					if i < len(ls) {
						matches = append(matches, MatchSort(ps[i], ls[i], freesyms))
					}
				}
				return MergeMatches(matches...)
			}
		}
	}
	return nil
}

// MatchQuants matches quantified formulas.
func MatchQuants(pat, inst lg.Expr, freesyms, constants map[lg.NodeKey]lg.Expr) map[lg.NodeKey]lg.Expr {
	if sameNodeType(pat, inst) {
		patVars := il.BinderVars(pat)
		instVars := il.BinderVars(inst)
		if len(patVars) != len(instVars) {
			return nil
		}
		// temporarily add pat.variables to freesyms
		for _, v := range patVars {
			freesyms[lg.Key(v)] = v
		}
		defer func() {
			for _, v := range patVars {
				delete(freesyms, lg.Key(v))
			}
		}()

		matches := make([]map[lg.NodeKey]lg.Expr, 0, len(patVars)+1)
		for i := range patVars {
			matches = append(matches, Match(patVars[i], instVars[i], freesyms, constants))
		}
		mat := MergeMatches(matches...)
		if mat != nil {
			mbody := ApplyMatch(mat, il.BinderBody(pat))
			bodyFreesyms := ApplyMatchFreesyms(mat, freesyms)
			bodyMat := Match(mbody, il.BinderBody(inst), bodyFreesyms, constants)
			// Build quants map from patVars (Python: quants=pat.variables, checked with `sym not in quants`)
			patVarMap := make(map[lg.NodeKey]lg.Expr, len(patVars))
			for _, v := range patVars {
				patVarMap[lg.Key(v)] = v
			}
			bodyMat = ComposeMatches(freesyms, mat, bodyMat, patVarMap)
			mat = MergeMatches(mat, bodyMat)
		}
		if mat != nil {
			for _, v := range patVars {
				delete(mat, lg.Key(v))
			}
		}
		return mat
	}
	return nil
}

// FOMatch computes a partial first-order match.
// Matches free FO variables to ground terms, but ignores variable
// occurrences under free second-order symbols.
func FOMatch(pat, inst lg.Expr, freesyms, constants map[lg.NodeKey]lg.Expr) map[lg.NodeKey]lg.Expr {
	if v, ok := pat.(*lg.Variable); ok {
		if freesyms[lg.Key(v)] != nil && allVariablesAreConstants(inst, constants) {
			res := map[lg.NodeKey]lg.Expr{lg.Key(v): inst}
			if freesyms[lg.Key(v.VSort)] != nil {
				res[lg.Key(v.VSort)] = inst.NodeSort()
				return res
			}
			if v.VSort.Equal(inst.NodeSort()) {
				return res
			}
		}
	}
	if il.IsQuantifier(pat) && il.IsQuantifier(inst) && sameNodeType(pat, inst) {
		patVars := il.BinderVars(pat)
		// temporarily remove pat.variables from freesyms
		saved := make(map[lg.NodeKey]lg.Expr)
		for _, v := range patVars {
			if freesyms[lg.Key(v)] != nil {
				saved[lg.Key(v)] = v
				delete(freesyms, lg.Key(v))
			}
		}
		defer func() {
			for k := range saved {
				freesyms[k] = saved[k]
			}
		}()
		return FOMatch(il.BinderBody(pat), il.BinderBody(inst), freesyms, constants)
	}
	if HeadsMatch(pat, inst, freesyms) {
		patArgs := il.NodeArgs(pat)
		instArgs := il.NodeArgs(inst)
		matches := make([]map[lg.NodeKey]lg.Expr, len(patArgs))
		for i := range patArgs {
			matches[i] = FOMatch(patArgs[i], instArgs[i], freesyms, constants)
		}
		return MergeMatches(matches...)
	}
	return map[lg.NodeKey]lg.Expr{}
}

// ComposeMatches composes two matches: for each free symbol not in quants,
// if mat1 maps it to sym1 and sym1 is in mat2, the result maps the original
// to mat2[sym1].
// quants is a set of symbol keys to exclude (not-in-quants check).
// Python: compose_matches(freesyms, mat1, mat2, quants) where quants can be
// a list of variables or a dict — both checked with `sym not in quants`.
func ComposeMatches(freesyms map[lg.NodeKey]lg.Expr, mat1, mat2 map[lg.NodeKey]lg.Expr, quants map[lg.NodeKey]lg.Expr) map[lg.NodeKey]lg.Expr {
	if mat1 == nil || mat2 == nil {
		return nil
	}
	res := make(map[lg.NodeKey]lg.Expr)
	for symKey, symNode := range freesyms {
		if quants[symKey] != nil {
			continue
		}
		if symNode == nil {
			continue
		}
		sym1 := ApplyMatchSym(mat1, symNode)
		if v, ok := mat2[lg.Key(sym1)]; ok {
			res[symKey] = v
		}
	}
	return res
}

// ApplyMatch applies a match to a formula.
// Substitutes all symbols in the match with the corresponding lambda terms
// and performs beta reduction. Alpha-renames to avoid capture.
//
// Python: ivy_proof.py:1075-1086 (apply_match / apply_match_rec)
// Python first calls alpha_avoid to rename bound variables that would clash
// with free variables introduced by the substitution.
func ApplyMatch(match map[lg.NodeKey]lg.Expr, fmla lg.Expr) lg.Expr {
	if len(match) == 0 {
		return fmla
	}
	// Alpha-rename bound vars to avoid capture by match RHS free vars.
	// Python: freevars = match_rhs_vars(match); fmla = il.alpha_avoid(fmla, freevars)
	freeVars := MatchRhsVars(match)
	fmla = il.AlphaAvoidMap(fmla, freeVars)
	return applyMatchRec(match, fmla)
}

// applyMatchRec recursively applies a match to a formula with beta reduction.
func applyMatchRec(match map[lg.NodeKey]lg.Expr, fmla lg.Expr) lg.Expr {
	args := il.NodeArgs(fmla)
	newArgs := make([]lg.Expr, len(args))
	for i, a := range args {
		newArgs[i] = applyMatchRec(match, a)
	}

	// Application: check if the function is in the match
	if il.IsApp(fmla) {
		c := appFunc(fmla)
		if c != nil {
			if replacement, ok := match[lg.Key(c)]; ok {
				// Beta reduction: apply the lambda to the arguments
				if lam, ok := replacement.(*lg.Lambda); ok {
					return betaReduce(lam, newArgs)
				}
				// If replacement is a constant, build new application
				if rc, ok := replacement.(*lg.Const); ok {
					if len(newArgs) > 0 {
						return lg.MustApply(rc, newArgs...)
					}
					return rc
				}
				return replacement
			}
			// Apply sort mapping to the function
			newC := ApplyMatchFunc(match, c)
			if newC != c {
				if len(newArgs) > 0 {
					return lg.MustApply(newC, newArgs...)
				}
				return newC
			}
		}
	}

	// Variable: check if in match
	if v, ok := fmla.(*lg.Variable); ok {
		if replacement, ok := match[lg.Key(v)]; ok {
			return replacement
		}
		// Apply sort mapping
		newSort := matchGetSort(match, v.VSort)
		if newSort != v.VSort {
			nv, _ := lg.NewVariable(v.Name, newSort)
			return nv
		}
		return fmla
	}

	// Binder: apply match to bound variable sorts, then use CloneBinder.
	// Python: with il.BindSymbols(env, fmla.variables):
	//             fmla = fmla.clone_binder([apply_match_rec(match,v,env) for v in fmla.variables], args[0])
	if il.IsQuantifier(fmla) && len(newArgs) > 0 {
		vars := il.BinderVars(fmla)
		newVars := make([]*lg.Variable, len(vars))
		for i, v := range vars {
			processed := applyMatchRec(match, v)
			if nv, ok := processed.(*lg.Variable); ok {
				newVars[i] = nv
			} else {
				newVars[i] = v
			}
		}
		return il.CloneBinder(fmla, newVars, newArgs[0])
	}

	// Clone with new args
	if len(args) == 0 {
		return fmla
	}
	return il.CloneNode(fmla, newArgs)
}

// betaReduce applies a lambda to arguments, performing beta reduction.
func betaReduce(lam *lg.Lambda, args []lg.Expr) lg.Expr {
	if len(lam.Variables) != len(args) {
		// Arity mismatch — return the lambda applied to args as-is
		return lam.Body
	}
	subs := make(map[string]lg.Expr, len(lam.Variables))
	for i, v := range lam.Variables {
		subs[v.Name] = args[i]
	}
	return lu.SubstituteByName(lam.Body, subs)
}

// ApplyMatchSym applies a match to a single symbol (constant, variable, or sort).
func ApplyMatchSym(match map[lg.NodeKey]lg.Expr, sym lg.Expr) lg.Expr {
	if v, ok := match[lg.Key(sym)]; ok {
		return v
	}
	if v, ok := sym.(*lg.Variable); ok {
		newSort := matchGetSort(match, v.VSort)
		if newSort != v.VSort {
			nv, _ := lg.NewVariable(v.Name, newSort)
			return nv
		}
		return sym
	}
	if c, ok := sym.(*lg.Const); ok {
		return ApplyMatchFunc(match, c)
	}
	return sym
}

// ApplyMatchFunc applies sort mappings to a constant's sort.
func ApplyMatchFunc(match map[lg.NodeKey]lg.Expr, c *lg.Const) *lg.Const {
	sorts := FuncSorts(c)
	newSorts := make([]lg.Sort, len(sorts))
	for i, s := range sorts {
		if rep, ok := match[lg.Key(s)]; ok {
			if rs, ok := rep.(lg.Sort); ok {
				newSorts[i] = rs
				continue
			}
		}
		newSorts[i] = s
	}
	var newSort lg.Sort
	if len(newSorts) == 1 {
		newSort = newSorts[0]
	} else {
		fs, err := lg.NewFunctionSort(newSorts...)
		if err != nil {
			return c
		}
		newSort = fs
	}
	return lg.NewConst(c.Name, newSort)
}

// ApplyMatchFreesyms applies a match to the free symbols set, returning a new set
// without the matched symbols.
func ApplyMatchFreesyms(match map[lg.NodeKey]lg.Expr, freesyms map[lg.NodeKey]lg.Expr) map[lg.NodeKey]lg.Expr {
	result := make(map[lg.NodeKey]lg.Expr)
	for sym := range freesyms {
		if _, matched := match[sym]; matched {
			continue
		}
		symNode := freesyms[sym]
		result[lg.Key(ApplyMatchSym(match, symNode))] = ApplyMatchSym(match, symNode)
	}
	return result
}

// ExtractTerms returns a lambda term t such that t(terms) = inst and
// the given terms do not occur in t. Returns nil if the extraction
// would introduce non-constant variables.
func ExtractTerms(inst lg.Expr, terms []lg.Expr, constants map[lg.NodeKey]lg.Expr) *lg.Lambda {
	if len(terms) == 0 {
		return nil
	}
	vars := make([]*lg.Variable, len(terms))
	for i, t := range terms {
		name := fmt.Sprintf("V%d", i)
		v, _ := lg.NewVariable(name, t.NodeSort())
		vars[i] = v
	}
	body := extractRec(inst, terms, vars)

	// Build a set of the lambda's own variables to exclude from the check
	lamVarSet := make(map[lg.NodeKey]lg.Expr, len(vars))
	for _, v := range vars {
		lamVarSet[lg.Key(v)] = v
	}

	// Check that all free variables in the body (excluding lambda vars) are constants
	freeVars := lu.FreeVariables(body)
	for v := range freeVars.All() {
		if lamVarSet[v] == nil && constants[v] == nil {
			return nil
		}
	}

	lam, err := lg.NewLambda(vars, body)
	if err != nil {
		return nil
	}
	return lam
}

func extractRec(inst lg.Expr, terms []lg.Expr, vars []*lg.Variable) lg.Expr {
	for i, t := range terms {
		if t.Equal(inst) {
			return vars[i]
		}
	}
	args := il.NodeArgs(inst)
	if len(args) == 0 {
		return inst
	}
	newArgs := make([]lg.Expr, len(args))
	for i, a := range args {
		newArgs[i] = extractRec(a, terms, vars)
	}
	return il.CloneNode(inst, newArgs)
}

// --- AddSymbols / RemoveSymbols ---

// AddSymbols temporarily adds symbols to a set. Use with defer Restore().
type AddSymbols struct {
	symset  map[lg.NodeKey]lg.Expr
	added   []lg.Expr
	removed []lg.Expr
}

// NewAddSymbols adds symlist to symset, tracking what was changed.
func NewAddSymbols(symset map[lg.NodeKey]lg.Expr, symlist []lg.Expr) *AddSymbols {
	as := &AddSymbols{symset: symset}
	for _, sym := range symlist {
		if symset[lg.Key(sym)] != nil {
			as.removed = append(as.removed, sym)
		}
		symset[lg.Key(sym)] = sym
		as.added = append(as.added, sym)
	}
	return as
}

// Restore undoes the changes made by NewAddSymbols.
func (as *AddSymbols) Restore() {
	for _, sym := range as.added {
		delete(as.symset, lg.Key(sym))
	}
	for _, sym := range as.removed {
		as.symset[lg.Key(sym)] = sym
	}
}

// RemoveSymbols temporarily removes symbols from a set. Use with defer Restore().
type RemoveSymbols struct {
	symset map[lg.NodeKey]lg.Expr
	saved  []lg.Expr
}

// NewRemoveSymbols removes symlist from symset, tracking what was changed.
func NewRemoveSymbols(symset map[lg.NodeKey]lg.Expr, symlist []lg.Expr) *RemoveSymbols {
	rs := &RemoveSymbols{symset: symset}
	for _, sym := range symlist {
		if symset[lg.Key(sym)] != nil {
			rs.saved = append(rs.saved, sym)
			delete(symset, lg.Key(sym))
		}
	}
	return rs
}

// Restore undoes the changes made by NewRemoveSymbols.
func (rs *RemoveSymbols) Restore() {
	for _, sym := range rs.saved {
		rs.symset[lg.Key(sym)] = sym
	}
}

// --- helpers ---

// appFunc returns the Const func of an Apply, or nil.
func appFunc(n lg.Expr) *lg.Const {
	switch t := n.(type) {
	case *lg.Apply:
		if c, ok := t.Func.(*lg.Const); ok {
			return c
		}
	case *lg.Const:
		return t
	}
	return nil
}

// sameNodeType returns true if a and b are the same concrete type.
func sameNodeType(a, b lg.Expr) bool {
	return fmt.Sprintf("%T", a) == fmt.Sprintf("%T", b)
}

// allVariablesAreConstants checks that all variables in a term are in the constants set.
func allVariablesAreConstants(n lg.Expr, constants map[lg.NodeKey]lg.Expr) bool {
	if v, ok := n.(*lg.Variable); ok {
		return constants[lg.Key(v)] != nil
	}
	for _, c := range n.Children() {
		if !allVariablesAreConstants(c, constants) {
			return false
		}
	}
	return true
}

// matchGetSort looks up a sort in a match map.
func matchGetSort(match map[lg.NodeKey]lg.Expr, s lg.Sort) lg.Sort {
	if rep, ok := match[lg.Key(s)]; ok {
		if rs, ok := rep.(lg.Sort); ok {
			return rs
		}
	}
	return s
}
