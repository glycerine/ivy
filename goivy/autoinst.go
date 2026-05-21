// Package autoinst implements automatic schema/axiom instantiation.
// This is a port of Python's ivy_auto_inst.py.
//
// It provides pattern-based eager instantiation of axiom schemata,
// which is required for automatic proof construction.
package goivy

import (
	"fmt"
)

// TODO: now that we have merged goivy/autoinst/ up into goivy/ package,
// we need to deduplicate the duplicated structs and func. For example,
// this Match2 is the same as MCMatch in mc_schema.go.

// --- Match2 class ---

// Match2 is a backtrackable unification-style matching context.
// Corresponds to Python's Match class.
type Match2 struct {
	stack [][]Expr               // stack of frames, each listing keys added
	m     map[string]interface{} // current mapping (sort or symbol)
}

// NewMatch2 creates a new Match2 context.
func NewMatch2() *Match2 {
	return &Match2{
		stack: [][]Expr{nil},
		m:     make(map[string]interface{}),
	}
}

// Add records a mapping from x to y in the current frame.
func (m *Match2) Add(key string, val interface{}) {
	m.m[key] = val
	frame := len(m.stack) - 1
	// Track key for rollback (using a nil node as marker)
	m.stack[frame] = append(m.stack[frame], NewConst(key, TopS))
}

// Push creates a new backtracking frame.
func (m *Match2) Push() {
	m.stack = append(m.stack, nil)
}

// Pop removes the top frame and rolls back any bindings made in it.
func (m *Match2) Pop() {
	frame := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	for _, marker := range frame {
		if c, ok := marker.(*Const); ok {
			delete(m.m, c.Name)
		}
	}
}

// Unify attempts to bind x to y. Returns true on success.
func (m *Match2) Unify(x string, y interface{}) bool {
	if existing, ok := m.m[x]; ok {
		return fmt.Sprint(existing) == fmt.Sprint(y)
	}
	m.Add(x, y)
	return true
}

// UnifyLists unifies two lists element-wise.
func (m *Match2) UnifyLists(xl, yl []interface{}) bool {
	if len(xl) != len(yl) {
		return false
	}
	for i := range xl {
		if !m.Unify(fmt.Sprint(xl[i]), yl[i]) {
			return false
		}
	}
	return true
}

// Get returns the bound value for a key, or nil if unbound.
func (m *Match2) Get(key string) interface{} {
	return m.m[key]
}

// CopyMap returns a copy of the current mapping.
func (m *Match2) CopyMap() map[string]interface{} {
	result := make(map[string]interface{}, len(m.m))
	for k, v := range m.m {
		result[k] = v
	}
	return result
}

// --- Schema matching ---

// ApplyMatch2 applies a match mapping to a formula by substituting matched symbols.
// Corresponds to Python's apply_match.
func ApplyMatch2(matchMap map[string]interface{}, fmla Expr) Expr {
	return applyMatchRec2(matchMap, fmla)
}

func applyMatchRec2(matchMap map[string]interface{}, fmla Expr) Expr {
	if v, ok := fmla.(*LogicVariable); ok {
		// Check if the variable's sort should be remapped
		sortStr := v.VSort.String()
		if newSort, exists := matchMap[sortStr]; exists {
			if s, ok := newSort.(Sort); ok {
				nv, _ := NewVariable(v.Name, s)
				return nv
			}
		}
		return v
	}

	if c, ok := fmla.(*Const); ok {
		// Check if the constant is in the match
		if replacement, exists := matchMap[c.Name]; exists {
			if rc, ok := replacement.(*Const); ok {
				return rc
			}
		}
		return c
	}

	if IsBinder(fmla) {
		vars := BinderVars(fmla)
		body := BinderBody(fmla)
		newVars := make([]*LogicVariable, len(vars))
		for i, v := range vars {
			nv := applyMatchRec2(matchMap, v)
			if rv, ok := nv.(*LogicVariable); ok {
				newVars[i] = rv
			} else {
				newVars[i] = v
			}
		}
		newBody := applyMatchRec2(matchMap, body)
		return CloneBinder(fmla, newVars, newBody)
	}

	args := NodeArgs(fmla)
	newArgs := make([]Expr, len(args))
	for i, arg := range args {
		newArgs[i] = applyMatchRec2(matchMap, arg)
	}

	// Check if this is an application with a matched function
	if app, ok := fmla.(*Apply); ok {
		if c, ok := app.Func.(*Const); ok {
			if replacement, exists := matchMap[c.Name]; exists {
				if rc, ok := replacement.(*Const); ok {
					return MustApply(rc, newArgs...)
				}
			}
		}
	}

	return CloneNode(fmla, newArgs)
}

// --- Normalization ---

// TermOrd provides a total ordering on terms for canonical forms.
// Corresponds to Python's term_ord (ivy_mc.py lines 815-828).
func TermOrdExpr(x, y Expr) int {
	xs, ys := fmt.Sprintf("%T", x), fmt.Sprintf("%T", y)
	if xs < ys {
		return -1
	}
	if xs > ys {
		return 1
	}
	// For Apply nodes, compare function names.
	if ax, ok := x.(*Apply); ok {
		if ay, ok := y.(*Apply); ok {
			xn := fmt.Sprintf("%v", ax.Func)
			yn := fmt.Sprintf("%v", ay.Func)
			if xn < yn {
				return -1
			}
			if xn > yn {
				return 1
			}
		}
	}
	// Compare arg counts.
	xargs := NodeArgs(x)
	yargs := NodeArgs(y)
	if len(xargs) < len(yargs) {
		return -1
	}
	if len(xargs) > len(yargs) {
		return 1
	}
	// Recursively compare args.
	for i := range xargs {
		res := TermOrdExpr(xargs[i], yargs[i])
		if res != 0 {
			return res
		}
	}
	return 0
}

// Normalize2 normalizes a formula by ordering equalities and simplifying x=x to true.
func Normalize2(expr Expr, iuCfg *IvyUtilsConfig) Expr {
	if IsMacro(expr, iuCfg) {
		return Normalize2(ExpandMacro(expr), iuCfg)
	}
	args := NodeArgs(expr)
	newArgs := make([]Expr, len(args))
	for i, a := range args {
		newArgs[i] = Normalize2(a, iuCfg)
	}
	return cloneNormal2(expr, newArgs)
}

func cloneNormal2(expr Expr, args []Expr) Expr {
	if _, ok := expr.(*Eq); ok && len(args) == 2 {
		x, y := args[0], args[1]
		if x.Equal(y) {
			return &LogicAnd{} // true
		}
		if TermOrdExpr(x, y) == 1 {
			x, y = y, x
		}
		return &Eq{T1: x, T2: y}
	}
	return CloneNode(expr, args)
}

// --- Pattern matching for trigger instantiation ---

// PatternMatch2 checks if a pattern matches an expression, filling in variable bindings.
func PatternMatch2(pat, expr Expr, mp map[string]Expr) bool {
	if v, ok := pat.(*LogicVariable); ok {
		if existing, found := mp[v.Name]; found {
			return expr.Equal(existing)
		}
		mp[v.Name] = expr
		return true
	}
	if app, ok := pat.(*Apply); ok {
		eapp, ok := expr.(*Apply)
		if !ok {
			return false
		}
		if !app.Func.Equal(eapp.Func) {
			return false
		}
		if len(app.Terms) != len(eapp.Terms) {
			return false
		}
		for i := range app.Terms {
			if !PatternMatch2(app.Terms[i], eapp.Terms[i], mp) {
				return false
			}
		}
		return true
	}
	if IsQuantifier(pat) {
		return false
	}
	if fmt.Sprintf("%T", pat) != fmt.Sprintf("%T", expr) {
		return false
	}
	if eq, ok := expr.(*Eq); ok {
		peq := pat.(*Eq)
		// Try both orderings
		save := make(map[string]Expr)
		for k, v := range mp {
			save[k] = v
		}
		if PatternMatch2(peq.T1, eq.T1, mp) && PatternMatch2(peq.T2, eq.T2, mp) {
			return true
		}
		// Restore and try swapped
		for k := range mp {
			delete(mp, k)
		}
		for k, v := range save {
			mp[k] = v
		}
		return PatternMatch2(peq.T1, eq.T2, mp) && PatternMatch2(peq.T2, eq.T1, mp)
	}
	patArgs := NodeArgs(pat)
	exprArgs := NodeArgs(expr)
	if len(patArgs) != len(exprArgs) {
		return false
	}
	for i := range patArgs {
		if !PatternMatch2(patArgs[i], exprArgs[i], mp) {
			return false
		}
	}
	return true
}

// TriggerMatches finds all matches of a trigger pattern against a set of formulas.
func TriggerMatches(fmlas []Expr, trig Expr) []map[string]Expr {
	var results []map[string]Expr
	for _, f := range fmlas {
		triggerMatchRec(f, trig, &results)
	}
	return results
}

func triggerMatchRec(expr, trig Expr, results *[]map[string]Expr) {
	for _, child := range NodeArgs(expr) {
		triggerMatchRec(child, trig, results)
	}
	mp := make(map[string]Expr)
	if PatternMatch2(trig, expr, mp) {
		*results = append(*results, mp)
	}
}

// --- Main instantiation function ---

// TriggerAxiom pairs triggers with an axiom formula.
type TriggerAxiom struct {
	Triggers []Expr
	Axiom    *LabeledFormula
}

// InstResult pairs an axiom with its instantiated formula.
type InstResult struct {
	Axiom   *LabeledFormula
	Formula Expr
}

// InstantiateAxioms performs pattern-based eager instantiation of axioms.
// Returns a list of (axiom, instantiated formula) pairs.
// Corresponds to Python's instantiate_axioms.
func InstantiateAxioms2(m *Module, fmlas []Expr, triggers []TriggerAxiom) []InstResult {
	// Collect all symbols used in formulas
	symbolSet := make(map[string]*Const)
	for _, f := range fmlas {
		//for _, sym := range il.SymbolsAst(f) { // I suspect this porting choice is buggy.
		for _, expr := range UsedSymbolsAst(f).All() {
			switch sym := expr.(type) {
			case *Const:
				symbolSet[sym.Name] = sym
			default:
				panic(fmt.Sprintf("what should I do for type %T here?", expr))
			}
		}
	}

	// Categorize into sort constants and function symbols
	sortConstants := make(map[string][]*Const) // sort name → constants of that sort
	var funs []*Const
	for _, sym := range symbolSet {
		if IsFunctionSort(sym.CSort) {
			funs = append(funs, sym)
		} else {
			sortName := sym.CSort.String()
			sortConstants[sortName] = append(sortConstants[sortName], sym)
		}
	}

	// Collect existing axioms
	// Expand schemata would go here (requires schema infrastructure)

	// Match triggers against formulas
	instSet := make(map[string]bool)
	var results []InstResult

	for _, ta := range triggers {
		trigMatches := make([][]map[string]Expr, len(ta.Triggers))
		for i, trig := range ta.Triggers {
			trigMatches[i] = TriggerMatches(fmlas, trig)
		}

		// Merge match lists
		merged := MergeMatchLists(trigMatches)

		for _, mp := range merged {
			// Apply match to axiom formula
			subs := make(map[NodeKey]Expr)
			for k, v := range mp {
				// Create a variable with this name to use as key
				vKey, _ := NewVariable(k, v.NodeSort())
				subs[Key(vKey)] = v
			}
			inst, err := Substitute(ta.Axiom.Formula.(Expr), subs)
			if err != nil {
				continue
			}
			instStr := inst.String()
			if !instSet[instStr] {
				instSet[instStr] = true
				results = append(results, InstResult{
					Axiom:   ta.Axiom,
					Formula: inst,
				})
			}
		}
	}

	return results
}

// MergeMatchLists computes the Cartesian product of match lists,
// merging compatible matches.
func MergeMatchLists(matchLists [][]map[string]Expr) []map[string]Expr {
	if len(matchLists) == 0 {
		return []map[string]Expr{{}}
	}
	var results []map[string]Expr
	mergeMatchListsRec(matchLists, 0, make(map[string]Expr), &results)
	return results
}

func mergeMatchListsRec(matchLists [][]map[string]Expr, idx int, current map[string]Expr, results *[]map[string]Expr) {
	if idx == len(matchLists) {
		cp := make(map[string]Expr, len(current))
		for k, v := range current {
			cp[k] = v
		}
		*results = append(*results, cp)
		return
	}
	for _, mp := range matchLists[idx] {
		merged := make(map[string]Expr, len(current))
		for k, v := range current {
			merged[k] = v
		}
		compatible := true
		for k, v := range mp {
			if existing, ok := merged[k]; ok {
				if !existing.Equal(v) {
					compatible = false
					break
				}
			} else {
				merged[k] = v
			}
		}
		if compatible {
			mergeMatchListsRec(matchLists, idx+1, merged, results)
		}
	}
}

// Python: ivy_auto_inst.py:100-200 (uses match_schema_prems generator)
func ExpandSchemata2(m *Module, sortConstants map[string][]*Const, funs map[string]bool) []*LabeledFormula {
	var result []*LabeledFormula

	if m.Schemata == nil {
		return result
	}

	// Build initial match with all known sorts
	match := NewMatch2()
	if m.Sig != nil {
		for sortName, _ := range m.Sig.Sorts.All() {
			match.Add(sortName, sortName)
		}
	}

	// For each schema, try matching premises
	for name, lf := range m.Schemata.All() {
		// Skip recursive/inductive schemata
		if len(name) >= 4 && (name[:4] == "rec[" || name[:4] == "lep[" || name[:4] == "ind[") {
			continue
		}

		schema, ok := extractSchemaNode(lf)
		if !ok || schema == nil {
			continue
		}

		children := schema.Children()
		if len(children) < 2 {
			continue
		}

		conc := children[len(children)-1]
		prems := children[:len(children)-1]

		// Collect bound sorts
		boundSorts := make(map[string]bool)
		for _, prem := range prems {
			if us, ok := prem.(*UninterpretedSort); ok {
				boundSorts[us.Name] = true
			}
		}

		// Match premises and instantiate conclusion
		MatchSchemaPrems2(prems, sortConstants, funs, match, boundSorts, func(mp map[string]interface{}) {
			// Build substitution map
			subs := make(map[string]Expr)
			for k, v := range mp {
				switch val := v.(type) {
				case *Const:
					subs[k] = val
				case Expr:
					subs[k] = val
				case string:
					subs[k] = NewConst(val, nil)
				}
			}
			inst := SubstituteByName(conc, subs)
			result = append(result, m.Cfg.AstCfg.NewLabeledFormula(nil, inst))
		})
	}

	return result
}

// MatchSchemaPrems recursively matches schema premises against available
// constants and functions. Calls callback for each successful complete match.
//
// Python: ivy_auto_inst.py match_schema_prems generator
func MatchSchemaPrems2(
	prems []Expr,
	sortConstants map[string][]*Const,
	funs map[string]bool,
	match *Match2,
	boundSorts map[string]bool,
	callback func(map[string]interface{}),
) {
	if len(prems) == 0 {
		callback(match.CopyMap())
		return
	}

	// Process the last premise
	prem := prems[len(prems)-1]
	remainingPrems := prems[:len(prems)-1]

	switch p := prem.(type) {
	case *UninterpretedSort:
		// Match to known sorts
		for sortName := range sortConstants {
			match.Push()
			if match.Unify(p.Name, sortName) {
				MatchSchemaPrems2(remainingPrems, sortConstants, funs, match, boundSorts, callback)
			}
			match.Pop()
		}

	case *LogicVariable:
		// Match to constants of the appropriate sort
		sortKey := p.VSort.String()
		// Match Python defaultdict auto-vivification
		if _, ok := sortConstants[sortKey]; !ok {
			sortConstants[sortKey] = nil
		}
		consts := sortConstants[sortKey]
		for _, c := range consts {
			match.Push()
			if match.Unify(p.Name, c) {
				MatchSchemaPrems2(remainingPrems, sortConstants, funs, match, boundSorts, callback)
			}
			match.Pop()
		}

	case *Const:
		if IsFunctionSort(p.CSort) {
			// Match to function symbols
			for funName := range funs {
				match.Push()
				if match.Unify(p.Name, NewConst(funName, p.CSort)) {
					MatchSchemaPrems2(remainingPrems, sortConstants, funs, match, boundSorts, callback)
				}
				match.Pop()
			}
		} else {
			// Match to constants
			sortKey := ""
			if p.CSort != nil {
				sortKey = p.CSort.String()
			}
			// Match Python defaultdict auto-vivification
			if _, ok := sortConstants[sortKey]; !ok {
				sortConstants[sortKey] = nil
			}
			consts := sortConstants[sortKey]
			for _, c := range consts {
				match.Push()
				if match.Unify(p.Name, c) {
					MatchSchemaPrems2(remainingPrems, sortConstants, funs, match, boundSorts, callback)
				}
				match.Pop()
			}
		}

	default:
		// Skip unrecognized premise types
		MatchSchemaPrems2(remainingPrems, sortConstants, funs, match, boundSorts, callback)
	}
}

// extractSchemaNode tries to get a lg.Expr from a schema interface{}.
func extractSchemaNode(lf interface{}) (Expr, bool) {
	switch t := lf.(type) {
	case *LabeledFormula:
		if n, ok := t.Formula.(Expr); ok {
			return n, true
		}
		return nil, false
	case Expr:
		return t, true
	}
	return nil, false
}

// GetTrigger finds a trigger expression in a formula that covers all bound variables.
//
// Python: ivy_mc.py:674-684, also used in ivy_auto_inst.py
func GetTrigger(expr Expr, vars []*LogicVariable) Expr {
	if IsQuantifier(expr) || IsVariable(expr) {
		return nil
	}

	for _, child := range NodeArgs(expr) {
		r := GetTrigger(child, vars)
		if r != nil {
			return r
		}
	}

	if IsApp(expr) || isEqNode(expr) {
		exprVars := FreeVariablesList(expr)
		if containsAllVars2(exprVars, vars) {
			return expr
		}
	}
	return nil
}

func isEqNode(n Expr) bool {
	_, ok := n.(*Eq)
	return ok
}

func containsAllVars2(have []*LogicVariable, need []*LogicVariable) bool {
	haveSet := make(map[string]bool, len(have))
	for _, v := range have {
		haveSet[v.Name] = true
	}
	for _, v := range need {
		if !haveSet[v.Name] {
			return false
		}
	}
	return true
}
