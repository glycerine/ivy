package ivylogic

import (
	"fmt"
	"strings"

	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
)

// --- Additional classification functions ---

// segVarPat returns the variable pattern for an Apply node's arguments.
// Each position is the NodeKey of the Var if it's a variable, or "" otherwise.
// Uses structural keys (not pointers) to match Python's structural equality.
func segVarPat(t *lg.Apply) []lg.NodeKey {
	result := make([]lg.NodeKey, len(t.Terms))
	for i, arg := range t.Terms {
		if v, ok := arg.(*lg.Variable); ok {
			result[i] = lg.Key(v)
		}
	}
	return result
}

// IsSegregated returns true if the formula is "segregated", meaning
// that in each function application, the same variables appear in the
// same positions across all occurrences of that function.
// Corresponds to Python's is_segregated.
func IsSegregated(fmla lg.Expr) bool {
	fmla = DropExistentials(fmla)
	vs := lu.UsedVariables(fmla)

	// Collect all Apply nodes with variables
	var apps []*lg.Apply
	for _, sub := range Subterms(fmla) {
		if app, ok := sub.(*lg.Apply); ok {
			appVars := lu.UsedVariables(app)
			if len(appVars) > 0 {
				apps = append(apps, app)
			}
		}
	}

	// Partition by function name
	byName := make(map[string][]*lg.Apply)
	for _, app := range apps {
		var name string
		if c, ok := app.Func.(*lg.Const); ok {
			name = c.Name
		} else {
			name = app.Func.String()
		}
		byName[name] = append(byName[name], app)
	}

	for _, terms := range byName {
		pat := segVarPat(terms[0])
		// Check that all variables appear in the pattern
		pvs := make(map[lg.NodeKey]bool)
		for _, k := range pat {
			if k != "" {
				pvs[k] = true
			}
		}
		if len(pvs) != len(vs) {
			return false
		}
		// Check that all subsequent terms have the same pattern
		for _, t := range terms[1:] {
			pat2 := segVarPat(t)
			if len(pat2) != len(pat) {
				return false
			}
			for i := range pat {
				if pat[i] != pat2[i] {
					return false
				}
			}
		}
	}
	return true
}

// isEPRRec is the recursive helper for IsEPR.
func isEPRRec(term lg.Expr, uvars map[lg.NodeKey]lg.Expr) bool {
	if fa, ok := term.(*lg.ForAll); ok {
		newUvars := make(map[lg.NodeKey]lg.Expr, len(uvars)+len(fa.Variables))
		for k := range uvars {
			newUvars[k] = uvars[k]
		}
		for _, v := range fa.Variables {
			newUvars[lg.Key(v)] = v
		}
		return isEPRRec(fa.Body, newUvars)
	}
	if ex, ok := term.(*lg.Exists); ok {
		// Check if any free variable of the exists is in uvars
		fvs := lu.FreeVariables(ex)
		for _, v := range fvs.All() {
			if _, inUvars := uvars[lg.Key(v)]; inUvars {
				return false
			}
		}
		return isEPRRec(ex.Body, make(map[lg.NodeKey]lg.Expr))
	}
	for _, a := range NodeArgs(term) {
		if !isEPRRec(a, uvars) {
			return false
		}
	}
	return true
}

// IsEPR returns true if the term is in the Effectively Propositional logic.
// A formula is EPR if no existential quantifier is in the scope of
// a universal quantifier (after accounting for free variables).
func IsEPR(term lg.Expr) bool {
	fvs := lu.FreeVariables(term)
	fvsKeyed := make(map[lg.NodeKey]lg.Expr, fvs.Len())
	for _, v := range fvs.All() {
		fvsKeyed[lg.Key(v)] = v
	}
	return isEPRRec(term, fvsKeyed)
}

// checkEssentiallyUninterpreted checks that no variable occurs under
// an interpreted function symbol. Returns true if the term has no
// variables, false otherwise.
func checkEssentiallyUninterpreted(sig *Sig, fmla lg.Expr) (bool, error) {
	if IsVariable(fmla) {
		return false, nil
	}
	if IsBinder(fmla) {
		body := BinderBody(fmla)
		if body != nil {
			res, err := checkEssentiallyUninterpreted(sig, body)
			if err != nil {
				return false, err
			}
			vars := BinderVars(fmla)
			return res && len(vars) == 0, nil
		}
		return true, nil
	}
	argres := true
	for _, a := range NodeArgs(fmla) {
		res, err := checkEssentiallyUninterpreted(sig, a)
		if err != nil {
			return false, err
		}
		if !res {
			argres = false
		}
	}
	if IsApp(fmla) {
		var sym *lg.Const
		if app, ok := fmla.(*lg.Apply); ok {
			if c, ok := app.Func.(*lg.Const); ok {
				sym = c
			}
		} else if c, ok := fmla.(*lg.Const); ok {
			sym = c
		}
		if sym != nil && IsInterpretedSymbol(sig, sym) {
			if !argres {
				return false, fmt.Errorf("a variable occurs under an interpreted function symbol")
			}
		}
	}
	return argres, nil
}

// IsInLogic checks if a term is in the given logic ("epr", "qf", or "fo").
// For "epr", it checks essentially uninterpreted and no interpreted sorts.
// For "qf", it checks quantifier-free.
// For "fo", it checks no interpreted sorts.
func IsInLogic(sig *Sig, term lg.Expr, logic string) bool {
	switch logic {
	case LogicEPR:
		_, err := checkEssentiallyUninterpreted(sig, term)
		if err != nil {
			return false
		}
		cs := UsedSymbolsAst(term)
		for _, c := range cs.All() {
			if _, ok := sig.Interp[lg.ExprName(c)]; ok {
				return false
			}
		}
		return true
	case LogicQF:
		return IsQF(term)
	case LogicFO:
		cs := UsedSymbolsAst(term)
		for _, c := range cs.All() {
			if _, ok := sig.Interp[lg.ExprName(c)]; ok {
				return false
			}
		}
		return true
	}
	return false
}

// symbolsOverUniversalsRec is the recursive helper for SymbolsOverUniversals.
func symbolsOverUniversalsRec(fmla lg.Expr, syms map[string]*lg.Const, pos bool, univs map[lg.NodeKey]lg.Expr) bool {
	if IsVariable(fmla) {
		_, inUnivs := univs[lg.Key(fmla)]
		return !inUnivs
	}
	if IsQuantifier(fmla) {
		isFA := IsForall(fmla)
		if pos == isFA || len(univs) > 0 {
			vars := BinderVars(fmla)
			for _, v := range vars {
				univs[lg.Key(v)] = v
			}
			body := BinderBody(fmla)
			res := symbolsOverUniversalsRec(body, syms, pos, univs)
			for _, v := range vars {
				delete(univs, lg.Key(v))
			}
			return res
		}
	}
	if _, ok := fmla.(*lg.Not); ok {
		pos = !pos
	}
	argres := true
	for _, a := range NodeArgs(fmla) {
		if !symbolsOverUniversalsRec(a, syms, pos, univs) {
			argres = false
		}
	}
	if IsApp(fmla) && !IsEq(fmla) && !argres {
		if app, ok := fmla.(*lg.Apply); ok {
			if c, ok := app.Func.(*lg.Const); ok {
				syms[c.Name] = c
			}
		} else if c, ok := fmla.(*lg.Const); ok {
			syms[c.Name] = c
		}
	}
	return argres
}

// SymbolsOverUniversals returns the set of function symbols that occur
// over universally quantified variables after skolemization.
// Corresponds to Python's symbols_over_universals.
func SymbolsOverUniversals(fmlas []lg.Expr) []*lg.Const {
	syms := make(map[string]*lg.Const)
	for _, fmla := range fmlas {
		symbolsOverUniversalsRec(fmla, syms, true, make(map[lg.NodeKey]lg.Expr))
	}
	result := make([]*lg.Const, 0, len(syms))
	for _, c := range syms {
		result = append(result, c)
	}
	return result
}

// universalVariablesRec is the recursive helper for UniversalVariables.
func universalVariablesRec(fmla lg.Expr, pos bool, univs map[lg.NodeKey]lg.Expr) {
	if IsQuantifier(fmla) {
		isFA := IsForall(fmla)
		if pos == isFA {
			vars := BinderVars(fmla)
			for _, v := range vars {
				univs[lg.Key(v)] = v
			}
			return
		}
	}
	if _, ok := fmla.(*lg.Implies); ok {
		args := NodeArgs(fmla)
		if len(args) == 2 {
			universalVariablesRec(args[0], !pos, univs)
			universalVariablesRec(args[1], pos, univs)
			return
		}
	}
	if _, ok := fmla.(*lg.Not); ok {
		pos = !pos
	}
	for _, a := range NodeArgs(fmla) {
		universalVariablesRec(a, pos, univs)
	}
}

// UniversalVariables returns the variables that are universally quantified
// after skolemization.
// Corresponds to Python's universal_variables.
func UniversalVariables(fmlas []lg.Expr) []*lg.Variable {
	univs := make(map[lg.NodeKey]lg.Expr)
	for _, fmla := range fmlas {
		universalVariablesRec(fmla, true, univs)
	}
	result := make([]*lg.Variable, 0, len(univs))
	for _, node := range univs {
		if vv, ok := node.(*lg.Variable); ok {
			result = append(result, vv)
		}
	}
	return result
}

// --- Macro expansion ---

// macroExpansions maps operator names to their expansion functions.
// These correspond to Python's macros_expansions dict.
var macroExpansions = map[string]func(*lg.Apply) lg.Expr{
	"<=": func(t *lg.Apply) lg.Expr {
		if len(t.Terms) != 2 {
			return t
		}
		ltSym := lg.NewConst("<", t.Func.NodeSort())
		ltApp := lg.MustApply(ltSym, t.Terms...)
		eq := &lg.Eq{T1: t.Terms[0], T2: t.Terms[1]}
		return &lg.Or{Terms: []lg.Expr{ltApp, eq}}
	},
	">": func(t *lg.Apply) lg.Expr {
		if len(t.Terms) != 2 {
			return t
		}
		ltSym := lg.NewConst("<", t.Func.NodeSort())
		swapped := []lg.Expr{t.Terms[1], t.Terms[0]}
		return lg.MustApply(ltSym, swapped...)
	},
	">=": func(t *lg.Apply) lg.Expr {
		if len(t.Terms) != 2 {
			return t
		}
		ltSym := lg.NewConst("<", t.Func.NodeSort())
		swapped := []lg.Expr{t.Terms[1], t.Terms[0]}
		ltApp := lg.MustApply(ltSym, swapped...)
		eq := &lg.Eq{T1: t.Terms[0], T2: t.Terms[1]}
		return &lg.Or{Terms: []lg.Expr{ltApp, eq}}
	},
}

// IsMacro returns true if the term is a macro application that can be expanded.
// Corresponds to Python's is_macro.
func IsMacro(term lg.Expr, iuCfg *iu.IvyUtilsConfig) bool {
	if iuCfg == nil || !iuCfg.UsePolymorphicMacros {
		return false
	}
	app, ok := term.(*lg.Apply)
	if !ok {
		return false
	}
	c, ok := app.Func.(*lg.Const)
	if !ok {
		return false
	}
	_, isMacro := macroExpansions[c.Name]
	return isMacro
}

// ExpandMacro expands a macro application.
// Corresponds to Python's expand_macro.
func ExpandMacro(term lg.Expr) lg.Expr {
	app, ok := term.(*lg.Apply)
	if !ok {
		return term
	}
	c, ok := app.Func.(*lg.Const)
	if !ok {
		return term
	}
	fn, ok := macroExpansions[c.Name]
	if !ok {
		return term
	}
	return fn(app)
}

// --- Additional utility functions ---

// Exclusivity generates an exclusivity axiom for an enumerated sort.
// Given a sort and variant sorts, it asserts that the pointer-to (*>)
// relation is a partial function for each variant, and that variants
// are mutually exclusive.
// Corresponds to Python's exclusivity.
func Exclusivity(sort lg.Sort, variants []lg.Sort) lg.Expr {
	pto := func(s lg.Sort) *lg.Const {
		return lg.NewConst("*>", RelationSort([]lg.Sort{sort, s}))
	}

	var conjuncts []lg.Expr

	// Partial function for each variant
	for _, s := range variants {
		conjuncts = append(conjuncts, PartialFunction(pto(s)))
	}

	// Extensionality for variants
	for _, s := range variants {
		x, _ := lg.NewVariable("X", sort)
		y, _ := lg.NewVariable("Y", sort)
		z, _ := lg.NewVariable("Z", s)
		ptoXZ := lg.MustApply(pto(s), x, z)
		ptoYZ := lg.MustApply(pto(s), y, z)
		premise := &lg.And{Terms: []lg.Expr{ptoXZ, ptoYZ}}
		conclusion := &lg.Eq{T1: x, T2: y}
		conjuncts = append(conjuncts, &lg.Implies{T1: premise, T2: conclusion})
	}

	// Mutual exclusion between different variants
	for i1 := 1; i1 < len(variants); i1++ {
		s1 := variants[i1]
		for _, s2 := range variants[:i1] {
			x, _ := lg.NewVariable("X", sort)
			y, _ := lg.NewVariable("Y", s1)
			z, _ := lg.NewVariable("Z", s2)
			pto1 := lg.MustApply(pto(s1), x, y)
			pto2 := lg.MustApply(pto(s2), x, z)
			conjuncts = append(conjuncts, &lg.Not{Body: &lg.And{Terms: []lg.Expr{pto1, pto2}}})
		}
	}

	return &lg.And{Terms: conjuncts}
}

// Variables generates a list of variables, one for each sort in the list.
// Variable names are V0, V1, V2, ...
// Corresponds to Python's variables.
func Variables(sorts []lg.Sort) []*lg.Variable {
	vars := make([]*lg.Variable, len(sorts))
	for i, s := range sorts {
		v, _ := lg.NewVariable(fmt.Sprintf("V%d", i), s)
		vars[i] = v
	}
	return vars
}

// NaryRepr returns a string representation of an n-ary operation
// with the given operator and arguments.
// Corresponds to Python's nary_repr.
func NaryRepr(op string, args []lg.Expr) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = a.String()
	}
	res := strings.Join(parts, " "+op+" ")
	if len(args) > 1 {
		return "(" + res + ")"
	}
	return res
}

// IsDefinitional returns true if the formula is a definitional axiom.
// A definitional formula has the form: forall vars. lhs = rhs (or lhs <-> rhs)
// where lhs is an application with distinct variable arguments.
// Corresponds to Python's is_definitional.
func IsDefinitional(defn lg.Expr) bool {
	for {
		fa, ok := defn.(*lg.ForAll)
		if !ok {
			break
		}
		defn = fa.Body
	}

	var lhs lg.Expr
	if eq, ok := defn.(*lg.Eq); ok {
		lhs = eq.T1
	} else if iff, ok := defn.(*lg.Iff); ok {
		lhs = iff.T1
	} else {
		return false
	}

	if !IsApp(lhs) {
		return false
	}

	// Check that all arguments are distinct variables
	args := NodeArgs(lhs)
	seen := make(map[string]bool)
	for _, a := range args {
		v, ok := a.(*lg.Variable)
		if !ok {
			continue
		}
		if seen[v.Name] {
			return false
		}
		seen[v.Name] = true
	}
	return true
}
