package proof

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
)

// MatchProblem describes a schema-instantiation matching problem.
type MatchProblem struct {
	Schema      lg.Node          // the schema being instantiated (as logic node)
	SchemaLF    *ast.LabeledFormula // the schema as a LabeledFormula (may be nil)
	Pat         lg.Node          // pattern to match (conclusion of schema)
	Inst        lg.Node          // instance to match against (conclusion of goal)
	FreeSyms    map[lg.Node]bool // free symbols in the schema
	Constants   map[lg.Node]bool // constants (free variables of the goal)
	PremMatches []lg.Node        // premise match patterns (for Tuple matching)
	RevMap      map[lg.Node]lg.Node // reverse mapping for nonce symbols
}

// NewMatchProblem creates a MatchProblem.
func NewMatchProblem(schema, pat, inst lg.Node, freesyms, constants map[lg.Node]bool) *MatchProblem {
	fs := make(map[lg.Node]bool, len(freesyms))
	for k, v := range freesyms {
		fs[k] = v
	}
	cs := make(map[lg.Node]bool, len(constants))
	for k, v := range constants {
		cs[k] = v
	}
	return &MatchProblem{
		Schema:    schema,
		Pat:       pat,
		Inst:      inst,
		FreeSyms:  fs,
		Constants: cs,
		RevMap:    make(map[lg.Node]lg.Node),
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
func FuncsMatch(pat, inst *lg.Const, freesyms map[lg.Node]bool) bool {
	ps := FuncSorts(pat)
	is := FuncSorts(inst)
	if pat.Name != inst.Name || len(ps) != len(is) {
		return false
	}
	for i := range ps {
		if !freesyms[ps[i]] && !ps[i].Equal(is[i]) {
			return false
		}
	}
	return true
}

// HeadsMatch checks if the heads of two terms match.
// Same top-level operator and same number of arguments.
// Quantifiers do not match anything.
// A function symbol matches if it has the same name and non-free sorts agree.
func HeadsMatch(pat, inst lg.Node, freesyms map[lg.Node]bool) bool {
	if il.IsApp(pat) && il.IsApp(inst) {
		pc := appFunc(pat)
		ic := appFunc(inst)
		if pc != nil && ic != nil && !freesyms[pc] {
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
func TermSorts(term lg.Node) []lg.Sort {
	if c := appFunc(term); c != nil {
		return FuncSorts(c)
	}
	if v, ok := term.(*lg.Var); ok {
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
func MatchSort(pat, inst lg.Sort, freesyms map[lg.Node]bool) map[lg.Node]lg.Node {
	if freesyms[pat] {
		return map[lg.Node]lg.Node{pat: inst}
	}
	if pat.Equal(inst) {
		return map[lg.Node]lg.Node{}
	}
	return nil
}

// MergeMatches merges multiple match maps. Returns nil if any match is nil
// or if there is a conflicting assignment.
func MergeMatches(matches ...map[lg.Node]lg.Node) map[lg.Node]lg.Node {
	if len(matches) == 0 {
		return map[lg.Node]lg.Node{}
	}
	for _, m := range matches {
		if m == nil {
			return nil
		}
	}
	res := make(map[lg.Node]lg.Node)
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
func EquivAlpha(x, y lg.Node) bool {
	if x.Equal(y) {
		return true
	}
	xLam, xOk := x.(*lg.Lambda)
	yLam, yOk := y.(*lg.Lambda)
	if xOk && yOk && len(xLam.Variables) == len(yLam.Variables) {
		// substitute y's variables with x's variables
		subs := make(map[lg.Node]lg.Node)
		for i, v := range yLam.Variables {
			subs[v] = xLam.Variables[i]
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
func Match(pat, inst lg.Node, freesyms, constants map[lg.Node]bool) map[lg.Node]lg.Node {
	if il.IsQuantifier(pat) {
		return MatchQuants(pat, inst, freesyms, constants)
	}
	if HeadsMatch(pat, inst, freesyms) {
		patArgs := il.NodeArgs(pat)
		instArgs := il.NodeArgs(inst)
		matches := make([]map[lg.Node]lg.Node, 0, len(patArgs)+4)
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
		if v, ok := pat.(*lg.Var); ok {
			matches = append(matches, map[lg.Node]lg.Node{v: inst})
		}
		return MergeMatches(matches...)
	}

	// If pat is a free application, try to extract lambda
	if il.IsApp(pat) {
		c := appFunc(pat)
		if c != nil && freesyms[c] {
			patArgs := il.NodeArgs(pat)
			B := ExtractTerms(inst, patArgs, constants)
			if B != nil {
				matches := []map[lg.Node]lg.Node{{c: B}}
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
func MatchQuants(pat, inst lg.Node, freesyms, constants map[lg.Node]bool) map[lg.Node]lg.Node {
	if sameNodeType(pat, inst) {
		patVars := il.BinderVars(pat)
		instVars := il.BinderVars(inst)
		if len(patVars) != len(instVars) {
			return nil
		}
		// temporarily add pat.variables to freesyms
		for _, v := range patVars {
			freesyms[v] = true
		}
		defer func() {
			for _, v := range patVars {
				delete(freesyms, v)
			}
		}()

		matches := make([]map[lg.Node]lg.Node, 0, len(patVars)+1)
		for i := range patVars {
			matches = append(matches, Match(patVars[i], instVars[i], freesyms, constants))
		}
		mat := MergeMatches(matches...)
		if mat != nil {
			mbody := ApplyMatch(mat, il.BinderBody(pat))
			bodyFreesyms := ApplyMatchFreesyms(mat, freesyms)
			bodyMat := Match(mbody, il.BinderBody(inst), bodyFreesyms, constants)
			bodyMat = ComposeMatches(freesyms, mat, bodyMat, patVars)
			mat = MergeMatches(mat, bodyMat)
		}
		if mat != nil {
			for _, v := range patVars {
				delete(mat, v)
			}
		}
		return mat
	}
	return nil
}

// FOMatch computes a partial first-order match.
// Matches free FO variables to ground terms, but ignores variable
// occurrences under free second-order symbols.
func FOMatch(pat, inst lg.Node, freesyms, constants map[lg.Node]bool) map[lg.Node]lg.Node {
	if v, ok := pat.(*lg.Var); ok {
		if freesyms[v] && allVariablesAreConstants(inst, constants) {
			res := map[lg.Node]lg.Node{v: inst}
			if freesyms[v.VSort] {
				res[v.VSort] = inst.NodeSort()
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
		saved := make(map[lg.Node]bool)
		for _, v := range patVars {
			if freesyms[v] {
				saved[v] = true
				delete(freesyms, v)
			}
		}
		defer func() {
			for k := range saved {
				freesyms[k] = true
			}
		}()
		return FOMatch(il.BinderBody(pat), il.BinderBody(inst), freesyms, constants)
	}
	if HeadsMatch(pat, inst, freesyms) {
		patArgs := il.NodeArgs(pat)
		instArgs := il.NodeArgs(inst)
		matches := make([]map[lg.Node]lg.Node, len(patArgs))
		for i := range patArgs {
			matches[i] = FOMatch(patArgs[i], instArgs[i], freesyms, constants)
		}
		return MergeMatches(matches...)
	}
	return map[lg.Node]lg.Node{}
}

// ComposeMatches composes two matches: for each free symbol not in quants,
// if mat1 maps it to sym1 and sym1 is in mat2, the result maps the original
// to mat2[sym1].
func ComposeMatches(freesyms map[lg.Node]bool, mat1, mat2 map[lg.Node]lg.Node, quants []*lg.Var) map[lg.Node]lg.Node {
	if mat1 == nil || mat2 == nil {
		return nil
	}
	quantSet := make(map[lg.Node]bool, len(quants))
	for _, q := range quants {
		quantSet[q] = true
	}
	res := make(map[lg.Node]lg.Node)
	for sym := range freesyms {
		if quantSet[sym] {
			continue
		}
		sym1 := ApplyMatchSym(mat1, sym)
		if v, ok := mat2[sym1]; ok {
			res[sym] = v
		}
	}
	return res
}

// ApplyMatch applies a match to a formula.
// Substitutes all symbols in the match with the corresponding lambda terms
// and performs beta reduction. Alpha-renames to avoid capture.
//
// Python: ivy_proof.py:1117-1131, 1140-1158 (apply_match_alt / apply_match_alt_rec)
func ApplyMatch(match map[lg.Node]lg.Node, fmla lg.Node) lg.Node {
	if len(match) == 0 {
		return fmla
	}
	return applyMatchRec(match, fmla)
}

// applyMatchRec recursively applies a match to a formula with beta reduction.
func applyMatchRec(match map[lg.Node]lg.Node, fmla lg.Node) lg.Node {
	args := il.NodeArgs(fmla)
	newArgs := make([]lg.Node, len(args))
	for i, a := range args {
		newArgs[i] = applyMatchRec(match, a)
	}

	// Application: check if the function is in the match
	if il.IsApp(fmla) {
		c := appFunc(fmla)
		if c != nil {
			if replacement, ok := match[c]; ok {
				// Beta reduction: apply the lambda to the arguments
				if lam, ok := replacement.(*lg.Lambda); ok {
					return betaReduce(lam, newArgs)
				}
				// If replacement is a constant, build new application
				if rc, ok := replacement.(*lg.Const); ok {
					if len(newArgs) > 0 {
						return &lg.Apply{Func: rc, Terms: newArgs}
					}
					return rc
				}
				return replacement
			}
			// Apply sort mapping to the function
			newC := ApplyMatchFunc(match, c)
			if newC != c {
				if len(newArgs) > 0 {
					return &lg.Apply{Func: newC, Terms: newArgs}
				}
				return newC
			}
		}
	}

	// Variable: check if in match
	if v, ok := fmla.(*lg.Var); ok {
		if replacement, ok := match[v]; ok {
			return replacement
		}
		// Apply sort mapping
		newSort := matchGetSort(match, v.VSort)
		if newSort != v.VSort {
			nv, _ := lg.NewVar(v.Name, newSort)
			return nv
		}
		return fmla
	}

	// Binder: avoid capture by excluding bound variables
	if il.IsQuantifier(fmla) {
		// Already handled by recursion into body
	}

	// Clone with new args
	if len(args) == 0 {
		return fmla
	}
	changed := false
	for i := range args {
		if newArgs[i] != args[i] {
			changed = true
			break
		}
	}
	if !changed {
		return fmla
	}
	return il.CloneNode(fmla, newArgs)
}

// betaReduce applies a lambda to arguments, performing beta reduction.
func betaReduce(lam *lg.Lambda, args []lg.Node) lg.Node {
	if len(lam.Variables) != len(args) {
		// Arity mismatch — return the lambda applied to args as-is
		return lam.Body
	}
	subs := make(map[string]lg.Node, len(lam.Variables))
	for i, v := range lam.Variables {
		subs[v.Name] = args[i]
	}
	return lu.SubstituteByName(lam.Body, subs)
}

// ApplyMatchSym applies a match to a single symbol (constant, variable, or sort).
func ApplyMatchSym(match map[lg.Node]lg.Node, sym lg.Node) lg.Node {
	if v, ok := match[sym]; ok {
		return v
	}
	if v, ok := sym.(*lg.Var); ok {
		newSort := matchGetSort(match, v.VSort)
		if newSort != v.VSort {
			nv, _ := lg.NewVar(v.Name, newSort)
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
func ApplyMatchFunc(match map[lg.Node]lg.Node, c *lg.Const) *lg.Const {
	sorts := FuncSorts(c)
	changed := false
	newSorts := make([]lg.Sort, len(sorts))
	for i, s := range sorts {
		if rep, ok := match[s]; ok {
			if rs, ok := rep.(lg.Sort); ok {
				newSorts[i] = rs
				changed = true
				continue
			}
		}
		newSorts[i] = s
	}
	if !changed {
		return c
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
func ApplyMatchFreesyms(match map[lg.Node]lg.Node, freesyms map[lg.Node]bool) map[lg.Node]bool {
	result := make(map[lg.Node]bool)
	for sym := range freesyms {
		if _, matched := match[sym]; matched {
			continue
		}
		result[ApplyMatchSym(match, sym)] = true
	}
	return result
}

// ExtractTerms returns a lambda term t such that t(terms) = inst and
// the given terms do not occur in t. Returns nil if the extraction
// would introduce non-constant variables.
func ExtractTerms(inst lg.Node, terms []lg.Node, constants map[lg.Node]bool) *lg.Lambda {
	if len(terms) == 0 {
		return nil
	}
	vars := make([]*lg.Var, len(terms))
	for i, t := range terms {
		name := fmt.Sprintf("V%d", i)
		v, _ := lg.NewVar(name, t.NodeSort())
		vars[i] = v
	}
	body := extractRec(inst, terms, vars)

	// Build a set of the lambda's own variables to exclude from the check
	lamVarSet := make(map[lg.Node]bool, len(vars))
	for _, v := range vars {
		lamVarSet[v] = true
	}

	// Check that all free variables in the body (excluding lambda vars) are constants
	freeVars := lu.FreeVariables(body)
	for v := range freeVars {
		if !lamVarSet[v] && !constants[v] {
			return nil
		}
	}

	lam, err := lg.NewLambda(vars, body)
	if err != nil {
		return nil
	}
	return lam
}

func extractRec(inst lg.Node, terms []lg.Node, vars []*lg.Var) lg.Node {
	for i, t := range terms {
		if t.Equal(inst) {
			return vars[i]
		}
	}
	args := il.NodeArgs(inst)
	if len(args) == 0 {
		return inst
	}
	newArgs := make([]lg.Node, len(args))
	for i, a := range args {
		newArgs[i] = extractRec(a, terms, vars)
	}
	return il.CloneNode(inst, newArgs)
}

// --- AddSymbols / RemoveSymbols ---

// AddSymbols temporarily adds symbols to a set. Use with defer Restore().
type AddSymbols struct {
	symset  map[lg.Node]bool
	added   []lg.Node
	removed []lg.Node
}

// NewAddSymbols adds symlist to symset, tracking what was changed.
func NewAddSymbols(symset map[lg.Node]bool, symlist []lg.Node) *AddSymbols {
	as := &AddSymbols{symset: symset}
	for _, sym := range symlist {
		if symset[sym] {
			as.removed = append(as.removed, sym)
		}
		symset[sym] = true
		as.added = append(as.added, sym)
	}
	return as
}

// Restore undoes the changes made by NewAddSymbols.
func (as *AddSymbols) Restore() {
	for _, sym := range as.added {
		delete(as.symset, sym)
	}
	for _, sym := range as.removed {
		as.symset[sym] = true
	}
}

// RemoveSymbols temporarily removes symbols from a set. Use with defer Restore().
type RemoveSymbols struct {
	symset map[lg.Node]bool
	saved  []lg.Node
}

// NewRemoveSymbols removes symlist from symset, tracking what was changed.
func NewRemoveSymbols(symset map[lg.Node]bool, symlist []lg.Node) *RemoveSymbols {
	rs := &RemoveSymbols{symset: symset}
	for _, sym := range symlist {
		if symset[sym] {
			rs.saved = append(rs.saved, sym)
			delete(symset, sym)
		}
	}
	return rs
}

// Restore undoes the changes made by NewRemoveSymbols.
func (rs *RemoveSymbols) Restore() {
	for _, sym := range rs.saved {
		rs.symset[sym] = true
	}
}

// --- helpers ---

// appFunc returns the Const func of an Apply, or nil.
func appFunc(n lg.Node) *lg.Const {
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
func sameNodeType(a, b lg.Node) bool {
	return fmt.Sprintf("%T", a) == fmt.Sprintf("%T", b)
}

// allVariablesAreConstants checks that all variables in a term are in the constants set.
func allVariablesAreConstants(n lg.Node, constants map[lg.Node]bool) bool {
	if v, ok := n.(*lg.Var); ok {
		return constants[v]
	}
	for _, c := range n.Children() {
		if !allVariablesAreConstants(c, constants) {
			return false
		}
	}
	return true
}

// matchGetSort looks up a sort in a match map.
func matchGetSort(match map[lg.Node]lg.Node, s lg.Sort) lg.Sort {
	if rep, ok := match[s]; ok {
		if rs, ok := rep.(lg.Sort); ok {
			return rs
		}
	}
	return s
}
