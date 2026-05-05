// Package autoinst implements automatic schema/axiom instantiation.
// This is a port of Python's ivy_auto_inst.py.
//
// It provides pattern-based eager instantiation of axiom schemata,
// which is required for automatic proof construction.
package autoinst

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
)

// --- Match class ---

// Match is a backtrackable unification-style matching context.
// Corresponds to Python's Match class.
type Match struct {
	stack [][]goivy.Expr         // stack of frames, each listing keys added
	m     map[string]interface{} // current mapping (sort or symbol)
}

// NewMatch creates a new Match context.
func NewMatch() *Match {
	return &Match{
		stack: [][]goivy.Expr{nil},
		m:     make(map[string]interface{}),
	}
}

// Add records a mapping from x to y in the current frame.
func (m *Match) Add(key string, val interface{}) {
	m.m[key] = val
	frame := len(m.stack) - 1
	// Track key for rollback (using a nil node as marker)
	m.stack[frame] = append(m.stack[frame], goivy.NewConst(key, goivy.TopS))
}

// Push creates a new backtracking frame.
func (m *Match) Push() {
	m.stack = append(m.stack, nil)
}

// Pop removes the top frame and rolls back any bindings made in it.
func (m *Match) Pop() {
	frame := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	for _, marker := range frame {
		if c, ok := marker.(*goivy.Const); ok {
			delete(m.m, c.Name)
		}
	}
}

// Unify attempts to bind x to y. Returns true on success.
func (m *Match) Unify(x string, y interface{}) bool {
	if existing, ok := m.m[x]; ok {
		return fmt.Sprint(existing) == fmt.Sprint(y)
	}
	m.Add(x, y)
	return true
}

// UnifyLists unifies two lists element-wise.
func (m *Match) UnifyLists(xl, yl []interface{}) bool {
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
func (m *Match) Get(key string) interface{} {
	return m.m[key]
}

// CopyMap returns a copy of the current mapping.
func (m *Match) CopyMap() map[string]interface{} {
	result := make(map[string]interface{}, len(m.m))
	for k, v := range m.m {
		result[k] = v
	}
	return result
}

// --- Schema matching ---

// ApplyMatch applies a match mapping to a formula by substituting matched symbols.
// Corresponds to Python's apply_match.
func ApplyMatch(matchMap map[string]interface{}, fmla goivy.Expr) goivy.Expr {
	return applyMatchRec(matchMap, fmla)
}

func applyMatchRec(matchMap map[string]interface{}, fmla goivy.Expr) goivy.Expr {
	if v, ok := fmla.(*goivy.LogicVariable); ok {
		// Check if the variable's sort should be remapped
		sortStr := v.VSort.String()
		if newSort, exists := matchMap[sortStr]; exists {
			if s, ok := newSort.(goivy.Sort); ok {
				nv, _ := goivy.NewVariable(v.Name, s)
				return nv
			}
		}
		return v
	}

	if c, ok := fmla.(*goivy.Const); ok {
		// Check if the constant is in the match
		if replacement, exists := matchMap[c.Name]; exists {
			if rc, ok := replacement.(*goivy.Const); ok {
				return rc
			}
		}
		return c
	}

	if goivy.IsBinder(fmla) {
		vars := goivy.BinderVars(fmla)
		body := goivy.BinderBody(fmla)
		newVars := make([]*goivy.LogicVariable, len(vars))
		for i, v := range vars {
			nv := applyMatchRec(matchMap, v)
			if rv, ok := nv.(*goivy.LogicVariable); ok {
				newVars[i] = rv
			} else {
				newVars[i] = v
			}
		}
		newBody := applyMatchRec(matchMap, body)
		return goivy.CloneBinder(fmla, newVars, newBody)
	}

	args := goivy.NodeArgs(fmla)
	newArgs := make([]goivy.Expr, len(args))
	for i, arg := range args {
		newArgs[i] = applyMatchRec(matchMap, arg)
	}

	// Check if this is an application with a matched function
	if app, ok := fmla.(*goivy.Apply); ok {
		if c, ok := app.Func.(*goivy.Const); ok {
			if replacement, exists := matchMap[c.Name]; exists {
				if rc, ok := replacement.(*goivy.Const); ok {
					return goivy.MustApply(rc, newArgs...)
				}
			}
		}
	}

	return goivy.CloneNode(fmla, newArgs)
}

// --- Normalization ---

// TermOrd provides a total ordering on terms for canonical forms.
// Corresponds to Python's term_ord (ivy_mc.py lines 815-828).
func TermOrd(x, y goivy.Expr) int {
	xs, ys := fmt.Sprintf("%T", x), fmt.Sprintf("%T", y)
	if xs < ys {
		return -1
	}
	if xs > ys {
		return 1
	}
	// For Apply nodes, compare function names.
	if ax, ok := x.(*goivy.Apply); ok {
		if ay, ok := y.(*goivy.Apply); ok {
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
	xargs := goivy.NodeArgs(x)
	yargs := goivy.NodeArgs(y)
	if len(xargs) < len(yargs) {
		return -1
	}
	if len(xargs) > len(yargs) {
		return 1
	}
	// Recursively compare args.
	for i := range xargs {
		res := TermOrd(xargs[i], yargs[i])
		if res != 0 {
			return res
		}
	}
	return 0
}

// Normalize normalizes a formula by ordering equalities and simplifying x=x to true.
func Normalize(expr goivy.Expr, iuCfg *goivy.IvyUtilsConfig) goivy.Expr {
	if goivy.IsMacro(expr, iuCfg) {
		return Normalize(goivy.ExpandMacro(expr), iuCfg)
	}
	args := goivy.NodeArgs(expr)
	newArgs := make([]goivy.Expr, len(args))
	for i, a := range args {
		newArgs[i] = Normalize(a, iuCfg)
	}
	return cloneNormal(expr, newArgs)
}

func cloneNormal(expr goivy.Expr, args []goivy.Expr) goivy.Expr {
	if _, ok := expr.(*goivy.Eq); ok && len(args) == 2 {
		x, y := args[0], args[1]
		if x.Equal(y) {
			return &goivy.LogicAnd{} // true
		}
		if TermOrd(x, y) == 1 {
			x, y = y, x
		}
		return &goivy.Eq{T1: x, T2: y}
	}
	return goivy.CloneNode(expr, args)
}

// --- Pattern matching for trigger instantiation ---

// PatternMatch checks if a pattern matches an expression, filling in variable bindings.
func PatternMatch(pat, expr goivy.Expr, mp map[string]goivy.Expr) bool {
	if v, ok := pat.(*goivy.LogicVariable); ok {
		if existing, found := mp[v.Name]; found {
			return expr.Equal(existing)
		}
		mp[v.Name] = expr
		return true
	}
	if app, ok := pat.(*goivy.Apply); ok {
		eapp, ok := expr.(*goivy.Apply)
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
			if !PatternMatch(app.Terms[i], eapp.Terms[i], mp) {
				return false
			}
		}
		return true
	}
	if goivy.IsQuantifier(pat) {
		return false
	}
	if fmt.Sprintf("%T", pat) != fmt.Sprintf("%T", expr) {
		return false
	}
	if eq, ok := expr.(*goivy.Eq); ok {
		peq := pat.(*goivy.Eq)
		// Try both orderings
		save := make(map[string]goivy.Expr)
		for k, v := range mp {
			save[k] = v
		}
		if PatternMatch(peq.T1, eq.T1, mp) && PatternMatch(peq.T2, eq.T2, mp) {
			return true
		}
		// Restore and try swapped
		for k := range mp {
			delete(mp, k)
		}
		for k, v := range save {
			mp[k] = v
		}
		return PatternMatch(peq.T1, eq.T2, mp) && PatternMatch(peq.T2, eq.T1, mp)
	}
	patArgs := goivy.NodeArgs(pat)
	exprArgs := goivy.NodeArgs(expr)
	if len(patArgs) != len(exprArgs) {
		return false
	}
	for i := range patArgs {
		if !PatternMatch(patArgs[i], exprArgs[i], mp) {
			return false
		}
	}
	return true
}

// TriggerMatches finds all matches of a trigger pattern against a set of formulas.
func TriggerMatches(fmlas []goivy.Expr, trig goivy.Expr) []map[string]goivy.Expr {
	var results []map[string]goivy.Expr
	for _, f := range fmlas {
		triggerMatchRec(f, trig, &results)
	}
	return results
}

func triggerMatchRec(expr, trig goivy.Expr, results *[]map[string]goivy.Expr) {
	for _, child := range goivy.NodeArgs(expr) {
		triggerMatchRec(child, trig, results)
	}
	mp := make(map[string]goivy.Expr)
	if PatternMatch(trig, expr, mp) {
		*results = append(*results, mp)
	}
}

// --- Main instantiation function ---

// TriggerAxiom pairs triggers with an axiom formula.
type TriggerAxiom struct {
	Triggers []goivy.Expr
	Axiom    *goivy.LabeledFormula
}

// InstResult pairs an axiom with its instantiated formula.
type InstResult struct {
	Axiom   *goivy.LabeledFormula
	Formula goivy.Expr
}

// InstantiateAxioms performs pattern-based eager instantiation of axioms.
// Returns a list of (axiom, instantiated formula) pairs.
// Corresponds to Python's instantiate_axioms.
func InstantiateAxioms(m *goivy.Module, fmlas []goivy.Expr, triggers []TriggerAxiom) []InstResult {
	// Collect all symbols used in formulas
	symbolSet := make(map[string]*goivy.Const)
	for _, f := range fmlas {
		//for _, sym := range il.SymbolsAst(f) { // I suspect this porting choice is buggy.
		for _, expr := range goivy.UsedSymbolsAst(f).All() {
			switch sym := expr.(type) {
			case *goivy.Const:
				symbolSet[sym.Name] = sym
			default:
				panic(fmt.Sprintf("what should I do for type %T here?", expr))
			}
		}
	}

	// Categorize into sort constants and function symbols
	sortConstants := make(map[string][]*goivy.Const) // sort name → constants of that sort
	var funs []*goivy.Const
	for _, sym := range symbolSet {
		if goivy.IsFunctionSort(sym.CSort) {
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
		trigMatches := make([][]map[string]goivy.Expr, len(ta.Triggers))
		for i, trig := range ta.Triggers {
			trigMatches[i] = TriggerMatches(fmlas, trig)
		}

		// Merge match lists
		merged := MergeMatchLists(trigMatches)

		for _, mp := range merged {
			// Apply match to axiom formula
			subs := make(map[goivy.NodeKey]goivy.Expr)
			for k, v := range mp {
				// Create a variable with this name to use as key
				vKey, _ := goivy.NewVariable(k, v.NodeSort())
				subs[goivy.Key(vKey)] = v
			}
			inst, err := goivy.Substitute(ta.Axiom.Formula.(goivy.Expr), subs)
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
func MergeMatchLists(matchLists [][]map[string]goivy.Expr) []map[string]goivy.Expr {
	if len(matchLists) == 0 {
		return []map[string]goivy.Expr{{}}
	}
	var results []map[string]goivy.Expr
	mergeMatchListsRec(matchLists, 0, make(map[string]goivy.Expr), &results)
	return results
}

func mergeMatchListsRec(matchLists [][]map[string]goivy.Expr, idx int, current map[string]goivy.Expr, results *[]map[string]goivy.Expr) {
	if idx == len(matchLists) {
		cp := make(map[string]goivy.Expr, len(current))
		for k, v := range current {
			cp[k] = v
		}
		*results = append(*results, cp)
		return
	}
	for _, mp := range matchLists[idx] {
		merged := make(map[string]goivy.Expr, len(current))
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
