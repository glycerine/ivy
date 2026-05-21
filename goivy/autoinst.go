// autoinst.go — Automatic axiom instantiation for the auto_inst proof tactic.
// Port of Python ivy_auto_inst.py (trigger-based eager instantiation).
package goivy

import (
	"fmt"
)

// --- Pattern matching for trigger instantiation ---

// PatternMatch checks if a pattern matches an expression, filling in variable bindings.
// Corresponds to Python's pattern matching in ivy_auto_inst.py.
func PatternMatch(pat, expr Expr, mp map[string]Expr) bool {
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
			if !PatternMatch(app.Terms[i], eapp.Terms[i], mp) {
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
		save := make(map[string]Expr)
		for k, v := range mp {
			save[k] = v
		}
		if PatternMatch(peq.T1, eq.T1, mp) && PatternMatch(peq.T2, eq.T2, mp) {
			return true
		}
		for k := range mp {
			delete(mp, k)
		}
		for k, v := range save {
			mp[k] = v
		}
		return PatternMatch(peq.T1, eq.T2, mp) && PatternMatch(peq.T2, eq.T1, mp)
	}
	patArgs := NodeArgs(pat)
	exprArgs := NodeArgs(expr)
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
	if PatternMatch(trig, expr, mp) {
		*results = append(*results, mp)
	}
}

// --- Data structures ---

// TriggerAxiom pairs trigger patterns with an axiom to instantiate.
type TriggerAxiom struct {
	Triggers []Expr
	Axiom    *LabeledFormula
}

// InstResult pairs an axiom with its instantiated formula.
type InstResult struct {
	Axiom   *LabeledFormula
	Formula Expr
}

// --- Main instantiation function ---

// InstantiateAxiomsFromTriggers performs pattern-based eager instantiation of axioms.
// For each TriggerAxiom, it matches the trigger patterns against fmlas and
// substitutes variable bindings into the axiom formula.
// Returns a deduplicated list of (axiom, instantiated formula) pairs.
// Corresponds to Python's instantiate_axioms in ivy_auto_inst.py.
func InstantiateAxiomsFromTriggers(m *Module, fmlas []Expr, triggers []TriggerAxiom) []InstResult {
	instSet := make(map[string]bool)
	var results []InstResult

	for _, ta := range triggers {
		trigMatches := make([][]map[string]Expr, len(ta.Triggers))
		for i, trig := range ta.Triggers {
			trigMatches[i] = TriggerMatches(fmlas, trig)
		}

		merged := MergeMatchLists(trigMatches)

		for _, mp := range merged {
			subs := make(map[NodeKey]Expr)
			for k, v := range mp {
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
// merging compatible matches (bindings that agree on shared variables).
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
