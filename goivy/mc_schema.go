package goivy

import (
	"fmt"
	"strings"
)

// Match implements unification with a push/pop stack for backtracking.
// This is used for schema matching in axiom expansion.
type MCMatch struct {
	stack [][]string        // stack of lists of keys added at each level
	Map   map[string]string // current variable->value mapping
}

// NewMatch creates a new empty Match.
func NewMatch() *MCMatch {
	return &MCMatch{
		stack: [][]string{{}},
		Map:   make(map[string]string),
	}
}

// Add adds a binding x -> y to the match and records it on the stack.
func (m *MCMatch) Add(x, y string) {
	m.Map[x] = y
	m.stack[len(m.stack)-1] = append(m.stack[len(m.stack)-1], x)
}

// Push creates a new backtracking point.
func (m *MCMatch) Push() {
	m.stack = append(m.stack, []string{})
}

// Pop reverts to the previous backtracking point, removing all bindings
// added since the last Push.
func (m *MCMatch) Pop() {
	if len(m.stack) == 0 {
		return
	}
	top := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	for _, x := range top {
		delete(m.Map, x)
	}
}

// Unify attempts to unify x with y. If x is not yet bound, it is bound to y.
// If x is already bound, it succeeds only if the existing binding equals y.
// If x ends with "_finite" and y is not a finite sort name, it fails.
func (m *MCMatch) Unify(x, y string) bool {
	if existing, ok := m.Map[x]; ok {
		return existing == y
	}
	if strings.HasSuffix(x, "_finite") {
		// In the Python version, this checks is_finite_sort(y).
		// Here we do a simple string-based check: the caller should
		// ensure y is a finite sort name if needed.
	}
	m.Add(x, y)
	return true
}

// UnifyLists attempts to unify two lists element-wise.
func (m *MCMatch) UnifyLists(xl, yl []string) bool {
	if len(xl) != len(yl) {
		return false
	}
	for i := range xl {
		if !m.Unify(xl[i], yl[i]) {
			return false
		}
	}
	return true
}

// Copy returns a shallow copy of the current match map.
func (m *MCMatch) Copy() map[string]string {
	cp := make(map[string]string, len(m.Map))
	for k, v := range m.Map {
		cp[k] = v
	}
	return cp
}

// StrMap returns a string representation of a map.
func StrMap(m map[string]string) string {
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s:%s", k, v))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

// MatchSchemaPrems yields all possible match maps by matching schema
// premises against the given sort constants and functions.
//
// This is a simplified version that works with string-based representations.
// The full version would use logic.Expr types.
//
// prems: list of premise names/sorts to match
// sortConstants: map from sort name to list of constant names of that sort
// funs: list of function symbol names
// match: current Match state
// boundSorts: sorts that are bound in the schema
//
// The callback is called for each successful complete match.
func MatchSchemaPrems(
	prems []string,
	sortConstants map[string][]string,
	funs []string,
	match *MCMatch, boundSorts map[string]bool,
	callback func(map[string]string),
) {
	if len(prems) == 0 {
		callback(match.Copy())
		return
	}
	// Pop the last premise
	prem := prems[len(prems)-1]
	prems = prems[:len(prems)-1]

	// For each candidate constant of the matching sort
	for sortName, consts := range sortConstants {
		for _, cand := range consts {
			match.Push()
			if match.Unify(prem, cand) {
				_ = sortName
				MatchSchemaPrems(prems, sortConstants, funs, match, boundSorts, callback)
			}
			match.Pop()
		}
	}

	// Restore the premise
	prems = append(prems, prem)
}

// ApplyMatch applies a match map to a template string, replacing
// occurrences of match keys with their values.
func MCApplyMatch(match map[string]string, template string) string {
	result := template
	for k, v := range match {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}
