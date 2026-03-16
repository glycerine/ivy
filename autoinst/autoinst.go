// Package autoinst implements automatic schema/axiom instantiation.
// This is a port of Python's ivy_auto_inst.py.
//
// It provides pattern-based eager instantiation of axiom schemata,
// which is required for automatic proof construction.
package autoinst

import (
	"fmt"

	lg "github.com/glycerine/goivy/logic"
	il "github.com/glycerine/goivy/ivylogic"
	lu "github.com/glycerine/goivy/logicutil"
	mod "github.com/glycerine/goivy/module"
)

// Verbose controls auto_inst output.
var Verbose = true

// --- Match class ---

// Match is a backtrackable unification-style matching context.
// Corresponds to Python's Match class.
type Match struct {
	stack [][]lg.Node // stack of frames, each listing keys added
	m     map[string]interface{} // current mapping (sort or symbol)
}

// NewMatch creates a new Match context.
func NewMatch() *Match {
	return &Match{
		stack: [][]lg.Node{nil},
		m:     make(map[string]interface{}),
	}
}

// Add records a mapping from x to y in the current frame.
func (m *Match) Add(key string, val interface{}) {
	m.m[key] = val
	frame := len(m.stack) - 1
	// Track key for rollback (using a nil node as marker)
	m.stack[frame] = append(m.stack[frame], lg.NewConst(key, lg.TopS))
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
		if c, ok := marker.(*lg.Const); ok {
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
func ApplyMatch(matchMap map[string]interface{}, fmla lg.Node) lg.Node {
	return applyMatchRec(matchMap, fmla)
}

func applyMatchRec(matchMap map[string]interface{}, fmla lg.Node) lg.Node {
	if v, ok := fmla.(*lg.Var); ok {
		// Check if the variable's sort should be remapped
		sortStr := v.VSort.String()
		if newSort, exists := matchMap[sortStr]; exists {
			if s, ok := newSort.(lg.Sort); ok {
				nv, _ := lg.NewVar(v.Name, s)
				return nv
			}
		}
		return v
	}

	if c, ok := fmla.(*lg.Const); ok {
		// Check if the constant is in the match
		if replacement, exists := matchMap[c.Name]; exists {
			if rc, ok := replacement.(*lg.Const); ok {
				return rc
			}
		}
		return c
	}

	if il.IsBinder(fmla) {
		vars := il.BinderVars(fmla)
		body := il.BinderBody(fmla)
		newVars := make([]*lg.Var, len(vars))
		for i, v := range vars {
			nv := applyMatchRec(matchMap, v)
			if rv, ok := nv.(*lg.Var); ok {
				newVars[i] = rv
			} else {
				newVars[i] = v
			}
		}
		newBody := applyMatchRec(matchMap, body)
		return il.CloneBinder(fmla, newVars, newBody)
	}

	args := il.NodeArgs(fmla)
	newArgs := make([]lg.Node, len(args))
	for i, arg := range args {
		newArgs[i] = applyMatchRec(matchMap, arg)
	}

	// Check if this is an application with a matched function
	if app, ok := fmla.(*lg.Apply); ok {
		if c, ok := app.Func.(*lg.Const); ok {
			if replacement, exists := matchMap[c.Name]; exists {
				if rc, ok := replacement.(*lg.Const); ok {
					return &lg.Apply{Func: rc, Terms: newArgs}
				}
			}
		}
	}

	return il.CloneNode(fmla, newArgs)
}

// --- Normalization ---

// TermOrd provides a total ordering on terms for canonical forms.
func TermOrd(x, y lg.Node) int {
	xs, ys := fmt.Sprintf("%T", x), fmt.Sprintf("%T", y)
	if xs < ys {
		return -1
	}
	if xs > ys {
		return 1
	}
	return 0
}

// Normalize normalizes a formula by ordering equalities and simplifying x=x to true.
func Normalize(expr lg.Node) lg.Node {
	if il.IsMacro(expr) {
		return Normalize(il.ExpandMacro(expr))
	}
	args := il.NodeArgs(expr)
	newArgs := make([]lg.Node, len(args))
	for i, a := range args {
		newArgs[i] = Normalize(a)
	}
	return cloneNormal(expr, newArgs)
}

func cloneNormal(expr lg.Node, args []lg.Node) lg.Node {
	if eq, ok := expr.(*lg.Eq); ok && len(args) == 2 {
		x, y := args[0], args[1]
		if x.Equal(y) {
			return &lg.And{} // true
		}
		if TermOrd(x, y) == 1 {
			x, y = y, x
		}
		return &lg.Eq{T1: x, T2: y}
		_ = eq
	}
	return il.CloneNode(expr, args)
}

// --- Pattern matching for trigger instantiation ---

// PatternMatch checks if a pattern matches an expression, filling in variable bindings.
func PatternMatch(pat, expr lg.Node, mp map[string]lg.Node) bool {
	if v, ok := pat.(*lg.Var); ok {
		if existing, found := mp[v.Name]; found {
			return expr.Equal(existing)
		}
		mp[v.Name] = expr
		return true
	}
	if app, ok := pat.(*lg.Apply); ok {
		eapp, ok := expr.(*lg.Apply)
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
	if il.IsQuantifier(pat) {
		return false
	}
	if fmt.Sprintf("%T", pat) != fmt.Sprintf("%T", expr) {
		return false
	}
	if eq, ok := expr.(*lg.Eq); ok {
		peq := pat.(*lg.Eq)
		// Try both orderings
		save := make(map[string]lg.Node)
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
	patArgs := il.NodeArgs(pat)
	exprArgs := il.NodeArgs(expr)
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
func TriggerMatches(fmlas []lg.Node, trig lg.Node) []map[string]lg.Node {
	var results []map[string]lg.Node
	for _, f := range fmlas {
		triggerMatchRec(f, trig, &results)
	}
	return results
}

func triggerMatchRec(expr, trig lg.Node, results *[]map[string]lg.Node) {
	for _, child := range il.NodeArgs(expr) {
		triggerMatchRec(child, trig, results)
	}
	mp := make(map[string]lg.Node)
	if PatternMatch(trig, expr, mp) {
		*results = append(*results, mp)
	}
}

// --- Main instantiation function ---

// TriggerAxiom pairs triggers with an axiom formula.
type TriggerAxiom struct {
	Triggers []lg.Node
	Axiom    *mod.LabeledFormula
}

// InstResult pairs an axiom with its instantiated formula.
type InstResult struct {
	Axiom   *mod.LabeledFormula
	Formula lg.Node
}

// InstantiateAxioms performs pattern-based eager instantiation of axioms.
// Returns a list of (axiom, instantiated formula) pairs.
// Corresponds to Python's instantiate_axioms.
func InstantiateAxioms(m *mod.Module, fmlas []lg.Node, triggers []TriggerAxiom) []InstResult {
	// Collect all symbols used in formulas
	symbolSet := make(map[string]*lg.Const)
	for _, f := range fmlas {
		for _, sym := range il.SymbolsAst(f) {
			symbolSet[sym.Name] = sym
		}
	}

	// Categorize into sort constants and function symbols
	sortConstants := make(map[string][]*lg.Const) // sort name → constants of that sort
	var funs []*lg.Const
	for _, sym := range symbolSet {
		if il.IsFunctionSort(sym.CSort) {
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
		trigMatches := make([][]map[string]lg.Node, len(ta.Triggers))
		for i, trig := range ta.Triggers {
			trigMatches[i] = TriggerMatches(fmlas, trig)
		}

		// Merge match lists
		merged := MergeMatchLists(trigMatches)

		for _, mp := range merged {
			// Apply match to axiom formula
			subs := make(map[lg.Node]lg.Node)
			for k, v := range mp {
				// Create a variable with this name to use as key
				subs[&lg.Var{Name: k, VSort: v.NodeSort()}] = v
			}
			inst, err := lu.Substitute(ta.Axiom.Formula, subs)
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
func MergeMatchLists(matchLists [][]map[string]lg.Node) []map[string]lg.Node {
	if len(matchLists) == 0 {
		return []map[string]lg.Node{{}}
	}
	var results []map[string]lg.Node
	mergeMatchListsRec(matchLists, 0, make(map[string]lg.Node), &results)
	return results
}

func mergeMatchListsRec(matchLists [][]map[string]lg.Node, idx int, current map[string]lg.Node, results *[]map[string]lg.Node) {
	if idx == len(matchLists) {
		cp := make(map[string]lg.Node, len(current))
		for k, v := range current {
			cp[k] = v
		}
		*results = append(*results, cp)
		return
	}
	for _, mp := range matchLists[idx] {
		merged := make(map[string]lg.Node, len(current))
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
