package goivy

import (
	"fmt"
	"strings"
)

// --- Additional classification functions ---

// segVarPat returns the variable pattern for an Apply node's arguments.
// Each position is the NodeKey of the Var if it's a variable, or "" otherwise.
// Uses structural keys (not pointers) to match Python's structural equality.
func segVarPat(t *Apply) []NodeKey {
	result := make([]NodeKey, len(t.Terms))
	for i, arg := range t.Terms {
		if v, ok := arg.(*Variable); ok {
			result[i] = Key(v)
		}
	}
	return result
}

// IsSegregated returns true if the formula is "segregated", meaning
// that in each function application, the same variables appear in the
// same positions across all occurrences of that function.
// Corresponds to Python's is_segregated.
func IsSegregated(fmla Expr) bool {
	fmla = DropExistentials(fmla)
	vs := UsedVariables(fmla)

	// Collect all Apply nodes with variables
	var apps []*Apply
	for _, sub := range Subterms(fmla) {
		if app, ok := sub.(*Apply); ok {
			appVars := UsedVariables(app)
			if len(appVars) > 0 {
				apps = append(apps, app)
			}
		}
	}

	// Partition by function name
	byName := make(map[string][]*Apply)
	for _, app := range apps {
		var name string
		if c, ok := app.Func.(*Const); ok {
			name = c.Name
		} else {
			name = app.Func.String()
		}
		byName[name] = append(byName[name], app)
	}

	for _, terms := range byName {
		pat := segVarPat(terms[0])
		// Check that all variables appear in the pattern
		pvs := make(map[NodeKey]bool)
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
func isEPRRec(term Expr, uvars map[NodeKey]Expr) bool {
	if fa, ok := term.(*ForAll); ok {
		newUvars := make(map[NodeKey]Expr, len(uvars)+len(fa.Variables))
		for k := range uvars {
			newUvars[k] = uvars[k]
		}
		for _, v := range fa.Variables {
			newUvars[Key(v)] = v
		}
		return isEPRRec(fa.Body, newUvars)
	}
	if ex, ok := term.(*Exists); ok {
		// Check if any free variable of the exists is in uvars
		fvs := FreeVariables(ex)
		for _, v := range fvs.All() {
			if _, inUvars := uvars[Key(v)]; inUvars {
				return false
			}
		}
		return isEPRRec(ex.Body, make(map[NodeKey]Expr))
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
func IsEPR(term Expr) bool {
	fvs := FreeVariables(term)
	fvsKeyed := make(map[NodeKey]Expr, fvs.Len())
	for _, v := range fvs.All() {
		fvsKeyed[Key(v)] = v
	}
	return isEPRRec(term, fvsKeyed)
}

// checkEssentiallyUninterpreted checks that no variable occurs under
// an interpreted function symbol. Returns true if the term has no
// variables, false otherwise.
func checkEssentiallyUninterpreted(sig *Sig, fmla Expr) (bool, error) {
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
		var sym *Const
		if app, ok := fmla.(*Apply); ok {
			if c, ok := app.Func.(*Const); ok {
				sym = c
			}
		} else if c, ok := fmla.(*Const); ok {
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
func IsInLogic(sig *Sig, term Expr, logic string) bool {
	switch logic {
	case LogicEPR:
		_, err := checkEssentiallyUninterpreted(sig, term)
		if err != nil {
			return false
		}
		cs := UsedSymbolsAst(term)
		for _, c := range cs.All() {
			if _, ok := sig.Interp[ExprName(c)]; ok {
				return false
			}
		}
		return true
	case LogicQF:
		return IsQF(term)
	case LogicFO:
		cs := UsedSymbolsAst(term)
		for _, c := range cs.All() {
			if _, ok := sig.Interp[ExprName(c)]; ok {
				return false
			}
		}
		return true
	}
	return false
}

// symbolsOverUniversalsRec is the recursive helper for SymbolsOverUniversals.
func symbolsOverUniversalsRec(fmla Expr, syms map[string]*Const, pos bool, univs map[NodeKey]Expr) bool {
	if IsVariable(fmla) {
		_, inUnivs := univs[Key(fmla)]
		return !inUnivs
	}
	if IsQuantifier(fmla) {
		isFA := IsForall(fmla)
		if pos == isFA || len(univs) > 0 {
			vars := BinderVars(fmla)
			for _, v := range vars {
				univs[Key(v)] = v
			}
			body := BinderBody(fmla)
			res := symbolsOverUniversalsRec(body, syms, pos, univs)
			for _, v := range vars {
				delete(univs, Key(v))
			}
			return res
		}
	}
	if _, ok := fmla.(*Not); ok {
		pos = !pos
	}
	argres := true
	for _, a := range NodeArgs(fmla) {
		if !symbolsOverUniversalsRec(a, syms, pos, univs) {
			argres = false
		}
	}
	if IsApp(fmla) && !IsEq(fmla) && !argres {
		if app, ok := fmla.(*Apply); ok {
			if c, ok := app.Func.(*Const); ok {
				syms[c.Name] = c
			}
		} else if c, ok := fmla.(*Const); ok {
			syms[c.Name] = c
		}
	}
	return argres
}

// SymbolsOverUniversals returns the set of function symbols that occur
// over universally quantified variables after skolemization.
// Corresponds to Python's symbols_over_universals.
func SymbolsOverUniversals(fmlas []Expr) []*Const {
	syms := make(map[string]*Const)
	for _, fmla := range fmlas {
		symbolsOverUniversalsRec(fmla, syms, true, make(map[NodeKey]Expr))
	}
	result := make([]*Const, 0, len(syms))
	for _, c := range syms {
		result = append(result, c)
	}
	return result
}

// universalVariablesRec is the recursive helper for UniversalVariables.
func universalVariablesRec(fmla Expr, pos bool, univs map[NodeKey]Expr) {
	if IsQuantifier(fmla) {
		isFA := IsForall(fmla)
		if pos == isFA {
			vars := BinderVars(fmla)
			for _, v := range vars {
				univs[Key(v)] = v
			}
			return
		}
	}
	if _, ok := fmla.(*Implies); ok {
		args := NodeArgs(fmla)
		if len(args) == 2 {
			universalVariablesRec(args[0], !pos, univs)
			universalVariablesRec(args[1], pos, univs)
			return
		}
	}
	if _, ok := fmla.(*Not); ok {
		pos = !pos
	}
	for _, a := range NodeArgs(fmla) {
		universalVariablesRec(a, pos, univs)
	}
}

// UniversalVariables returns the variables that are universally quantified
// after skolemization.
// Corresponds to Python's universal_variables.
func UniversalVariables(fmlas []Expr) []*Variable {
	univs := make(map[NodeKey]Expr)
	for _, fmla := range fmlas {
		universalVariablesRec(fmla, true, univs)
	}
	result := make([]*Variable, 0, len(univs))
	for _, node := range univs {
		if vv, ok := node.(*Variable); ok {
			result = append(result, vv)
		}
	}
	return result
}

// --- Macro expansion ---

// macroExpansions maps operator names to their expansion functions.
// These correspond to Python's macros_expansions dict.
var macroExpansions = map[string]func(*Apply) Expr{
	"<=": func(t *Apply) Expr {
		if len(t.Terms) != 2 {
			return t
		}
		ltSym := NewConst("<", t.Func.NodeSort())
		ltApp := MustApply(ltSym, t.Terms...)
		eq := &Eq{T1: t.Terms[0], T2: t.Terms[1]}
		return &Or{Terms: []Expr{ltApp, eq}}
	},
	">": func(t *Apply) Expr {
		if len(t.Terms) != 2 {
			return t
		}
		ltSym := NewConst("<", t.Func.NodeSort())
		swapped := []Expr{t.Terms[1], t.Terms[0]}
		return MustApply(ltSym, swapped...)
	},
	">=": func(t *Apply) Expr {
		if len(t.Terms) != 2 {
			return t
		}
		ltSym := NewConst("<", t.Func.NodeSort())
		swapped := []Expr{t.Terms[1], t.Terms[0]}
		ltApp := MustApply(ltSym, swapped...)
		eq := &Eq{T1: t.Terms[0], T2: t.Terms[1]}
		return &Or{Terms: []Expr{ltApp, eq}}
	},
}

// IsMacro returns true if the term is a macro application that can be expanded.
// Corresponds to Python's is_macro.
func IsMacro(term Expr, iuCfg *IvyUtilsConfig) bool {
	if iuCfg == nil || !iuCfg.UsePolymorphicMacros {
		return false
	}
	app, ok := term.(*Apply)
	if !ok {
		return false
	}
	c, ok := app.Func.(*Const)
	if !ok {
		return false
	}
	_, isMacro := macroExpansions[c.Name]
	return isMacro
}

// ExpandMacro expands a macro application.
// Corresponds to Python's expand_macro.
func ExpandMacro(term Expr) Expr {
	app, ok := term.(*Apply)
	if !ok {
		return term
	}
	c, ok := app.Func.(*Const)
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
func IvyExclusivity(sort Sort, variants []Sort) Expr {
	pto := func(s Sort) *Const {
		return NewConst("*>", RelationSort([]Sort{sort, s}))
	}

	var conjuncts []Expr

	// Partial function for each variant
	for _, s := range variants {
		conjuncts = append(conjuncts, PartialFunction(pto(s)))
	}

	// Extensionality for variants
	for _, s := range variants {
		x, _ := NewVariable("X", sort)
		y, _ := NewVariable("Y", sort)
		z, _ := NewVariable("Z", s)
		ptoXZ := MustApply(pto(s), x, z)
		ptoYZ := MustApply(pto(s), y, z)
		premise := &And{Terms: []Expr{ptoXZ, ptoYZ}}
		conclusion := &Eq{T1: x, T2: y}
		conjuncts = append(conjuncts, &Implies{T1: premise, T2: conclusion})
	}

	// Mutual exclusion between different variants
	for i1 := 1; i1 < len(variants); i1++ {
		s1 := variants[i1]
		for _, s2 := range variants[:i1] {
			x, _ := NewVariable("X", sort)
			y, _ := NewVariable("Y", s1)
			z, _ := NewVariable("Z", s2)
			pto1 := MustApply(pto(s1), x, y)
			pto2 := MustApply(pto(s2), x, z)
			conjuncts = append(conjuncts, &Not{Body: &And{Terms: []Expr{pto1, pto2}}})
		}
	}

	return &And{Terms: conjuncts}
}

// Variables generates a list of variables, one for each sort in the list.
// Variable names are V0, V1, V2, ...
// Corresponds to Python's variables.
func Variables(sorts []Sort) []*Variable {
	vars := make([]*Variable, len(sorts))
	for i, s := range sorts {
		v, _ := NewVariable(fmt.Sprintf("V%d", i), s)
		vars[i] = v
	}
	return vars
}

// NaryRepr returns a string representation of an n-ary operation
// with the given operator and arguments.
// Corresponds to Python's nary_repr.
func IvyNaryRepr(op string, args []Expr) string {
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
func IsDefinitional(defn Expr) bool {
	for {
		fa, ok := defn.(*ForAll)
		if !ok {
			break
		}
		defn = fa.Body
	}

	var lhs Expr
	if eq, ok := defn.(*Eq); ok {
		lhs = eq.T1
	} else if iff, ok := defn.(*Iff); ok {
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
		v, ok := a.(*Variable)
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
