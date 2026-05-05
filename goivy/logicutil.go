package goivy

import (
	"fmt"
)

// CaptureError is raised when a substitution would create variable capture.
type LogicUtilCaptureError struct {
	Variables []*Variable
}

func (e *LogicUtilCaptureError) Error() string {
	return fmt.Sprintf("variable capture: %v", e.Variables)
}

// UsedVariables returns all variables used in the given term (both free and bound).
func UsedVariables(t Expr) map[NodeKey]Expr {
	result := make(map[NodeKey]Expr)
	logicutilUsedVariablesRec(t, result)
	return result
}

func logicutilUsedVariablesRec(t Expr, result map[NodeKey]Expr) {
	switch n := t.(type) {
	case *Variable:
		result[Key(n)] = n
	case *ForAll:
		for _, v := range n.Variables {
			result[Key(v)] = v
		}
		logicutilUsedVariablesRec(n.Body, result)
	case *Exists:
		for _, v := range n.Variables {
			result[Key(v)] = v
		}
		logicutilUsedVariablesRec(n.Body, result)
	case *Lambda:
		for _, v := range n.Variables {
			result[Key(v)] = v
		}
		logicutilUsedVariablesRec(n.Body, result)
	case *NamedBinder:
		for _, v := range n.Variables {
			result[Key(v)] = v
		}
		logicutilUsedVariablesRec(n.Body, result)
	default:
		for _, c := range t.Children() {
			logicutilUsedVariablesRec(c, result)
		}
	}
}

// FreeVariables returns the set of variables free in the given term.
// Variables are compared by structural identity (Sexp key).
// The returned Omap iterates in sorted key (NodeKey) order, giving
// alphabetical variable ordering that matches Python's sorted tuple.
func FreeVariables(t Expr) *Omap[NodeKey, Expr] {
	result := NewOmap[NodeKey, Expr]()
	freeVariablesRec(t, result, nil)
	return result
}

// FreeVariablesSet returns a set of free variable names (map[string]bool).
func FreeVariablesSet(t Expr) map[string]bool {
	byName := FreeVariablesByName(t)
	result := make(map[string]bool, len(byName))
	for name := range byName {
		result[name] = true
	}
	return result
}

// FreeVariablesByName returns the set of variable names free in the given term.
// Binding occurs by name (ForAll X:s1 binds X:s2).
func FreeVariablesByName(t Expr) map[string]struct{} {
	result := make(map[string]struct{})
	freeVariablesByNameRec(t, result, nil)
	return result
}

func freeVariablesRec(t Expr, result *Omap[NodeKey, Expr], bound map[NodeKey]Expr) {
	switch n := t.(type) {
	case *Variable:
		if _, isBound := bound[Key(n)]; !isBound {
			result.Set(Key(n), n)
		}
	case *ForAll:
		newBound := copyVarSet(bound)
		for _, v := range n.Variables {
			newBound[Key(v)] = v
		}
		freeVariablesRec(n.Body, result, newBound)
	case *Exists:
		newBound := copyVarSet(bound)
		for _, v := range n.Variables {
			newBound[Key(v)] = v
		}
		freeVariablesRec(n.Body, result, newBound)
	case *Lambda:
		newBound := copyVarSet(bound)
		for _, v := range n.Variables {
			newBound[Key(v)] = v
		}
		freeVariablesRec(n.Body, result, newBound)
	case *NamedBinder:
		newBound := copyVarSet(bound)
		for _, v := range n.Variables {
			newBound[Key(v)] = v
		}
		freeVariablesRec(n.Body, result, newBound)
	default:
		for _, c := range t.Children() {
			freeVariablesRec(c, result, bound)
		}
	}
}

func freeVariablesByNameRec(t Expr, result map[string]struct{}, bound map[string]struct{}) {
	switch n := t.(type) {
	case *Variable:
		if _, isBound := bound[n.Name]; !isBound {
			result[n.Name] = struct{}{}
		}
	case *ForAll:
		newBound := logicutilCopyStringSet(bound)
		for _, v := range n.Variables {
			newBound[v.Name] = struct{}{}
		}
		freeVariablesByNameRec(n.Body, result, newBound)
	case *Exists:
		newBound := logicutilCopyStringSet(bound)
		for _, v := range n.Variables {
			newBound[v.Name] = struct{}{}
		}
		freeVariablesByNameRec(n.Body, result, newBound)
	case *Lambda:
		newBound := logicutilCopyStringSet(bound)
		for _, v := range n.Variables {
			newBound[v.Name] = struct{}{}
		}
		freeVariablesByNameRec(n.Body, result, newBound)
	case *NamedBinder:
		newBound := logicutilCopyStringSet(bound)
		for _, v := range n.Variables {
			newBound[v.Name] = struct{}{}
		}
		freeVariablesByNameRec(n.Body, result, newBound)
	default:
		for _, c := range t.Children() {
			freeVariablesByNameRec(c, result, bound)
		}
	}
}

// BoundVariables returns the set of variables bound in the given term.
func BoundVariables(t Expr) map[NodeKey]Expr {
	result := make(map[NodeKey]Expr)
	boundVariablesRec(t, result)
	return result
}

func boundVariablesRec(t Expr, result map[NodeKey]Expr) {
	switch n := t.(type) {
	case *Variable:
		// leaf — no bound variables
	case *ForAll:
		for _, v := range n.Variables {
			result[Key(v)] = v
		}
		boundVariablesRec(n.Body, result)
	case *Exists:
		for _, v := range n.Variables {
			result[Key(v)] = v
		}
		boundVariablesRec(n.Body, result)
	case *Lambda:
		for _, v := range n.Variables {
			result[Key(v)] = v
		}
		boundVariablesRec(n.Body, result)
	case *NamedBinder:
		for _, v := range n.Variables {
			result[Key(v)] = v
		}
		boundVariablesRec(n.Body, result)
	default:
		for _, c := range t.Children() {
			boundVariablesRec(c, result)
		}
	}
}

// Substitute returns the term obtained from t by simultaneous substitution.
// subs maps Nodes (Var or Const) to replacement Nodes.
// Only free occurrences of variables are substituted.
// Returns CaptureError if substitution would create variable capture.
func Substitute(t Expr, subs map[NodeKey]Expr) (Expr, error) {
	if len(subs) == 0 {
		return t, nil
	}
	return substituteRec(t, subs)
}

func substituteRec(t Expr, subs map[NodeKey]Expr) (Expr, error) {
	switch n := t.(type) {
	case *Variable:
		if r, ok := subs[Key(n)]; ok {
			return r, nil
		}
		return t, nil

	case *Const:
		if r, ok := subs[Key(n)]; ok {
			return r, nil
		}
		return t, nil

	case *Apply:
		// Substitute in terms only, NOT in Func — matches Python substitute_ast
		// which iterates ast.args (terms only) and preserves ast.func via clone.
		newTerms := make([]Expr, len(n.Terms))
		for i, term := range n.Terms {
			nt, err := substituteRec(term, subs)
			if err != nil {
				return nil, err
			}
			newTerms[i] = nt
		}
		return NewApply(n.Func, newTerms...)

	case *Eq:
		t1, err := substituteRec(n.T1, subs)
		if err != nil {
			return nil, err
		}
		t2, err := substituteRec(n.T2, subs)
		if err != nil {
			return nil, err
		}
		return NewEq(t1, t2)

	case *Ite:
		c, err := substituteRec(n.Cond, subs)
		if err != nil {
			return nil, err
		}
		th, err := substituteRec(n.Then, subs)
		if err != nil {
			return nil, err
		}
		el, err := substituteRec(n.Else, subs)
		if err != nil {
			return nil, err
		}
		return NewIte(c, th, el)

	case *Not:
		b, err := substituteRec(n.Body, subs)
		if err != nil {
			return nil, err
		}
		return NewNot(b)

	case *And:
		terms := make([]Expr, len(n.Terms))
		for i, term := range n.Terms {
			nt, err := substituteRec(term, subs)
			if err != nil {
				return nil, err
			}
			terms[i] = nt
		}
		return NewAnd(terms...)

	case *Or:
		terms := make([]Expr, len(n.Terms))
		for i, term := range n.Terms {
			nt, err := substituteRec(term, subs)
			if err != nil {
				return nil, err
			}
			terms[i] = nt
		}
		return NewOr(terms...)

	case *Implies:
		t1, err := substituteRec(n.T1, subs)
		if err != nil {
			return nil, err
		}
		t2, err := substituteRec(n.T2, subs)
		if err != nil {
			return nil, err
		}
		return NewImplies(t1, t2)

	case *Iff:
		t1, err := substituteRec(n.T1, subs)
		if err != nil {
			return nil, err
		}
		t2, err := substituteRec(n.T2, subs)
		if err != nil {
			return nil, err
		}
		return NewIff(t1, t2)

	case *Globally:
		b, err := substituteRec(n.Body, subs)
		if err != nil {
			return nil, err
		}
		return NewGlobally(n.Environ, b)

	case *Eventually:
		b, err := substituteRec(n.Body, subs)
		if err != nil {
			return nil, err
		}
		return NewEventually(n.Environ, b)

	case *WhenOperator:
		t1, err := substituteRec(n.T1, subs)
		if err != nil {
			return nil, err
		}
		t2, err := substituteRec(n.T2, subs)
		if err != nil {
			return nil, err
		}
		return NewWhenOperator(n.Name, t1, t2)

	case *ForAll:
		return substituteBinder(n.Variables, n.Body, subs, func(vars []*Variable, body Expr) (Expr, error) {
			return NewForAll(vars, body)
		})

	case *Exists:
		return substituteBinder(n.Variables, n.Body, subs, func(vars []*Variable, body Expr) (Expr, error) {
			return NewExists(vars, body)
		})

	case *Lambda:
		return substituteBinder(n.Variables, n.Body, subs, func(vars []*Variable, body Expr) (Expr, error) {
			return NewLambda(vars, body)
		})

	case *NamedBinder:
		return substituteNamedBinder(n, subs)

	default:
		return nil, fmt.Errorf("substitute: unsupported node type: %T", t)
	}
}

// substituteBinder handles substitution into ForAll, Exists, Lambda.
func substituteBinder(
	variables []*Variable,
	body Expr,
	subs map[NodeKey]Expr,
	construct func([]*Variable, Expr) (Expr, error),
) (Expr, error) {
	// Remove bound variables from substitution
	newsubs := make(map[NodeKey]Expr)
	varSet := make(map[NodeKey]struct{})
	for _, v := range variables {
		varSet[Key(v)] = struct{}{}
	}
	for k, v := range subs {
		if _, isBound := varSet[k]; !isBound {
			newsubs[k] = v
		}
	}

	// Check for variable capture
	forbidden := make(map[NodeKey]Expr)
	for _, v := range newsubs {
		fv := FreeVariables(v)
		for fvarKey, fvarNode := range fv.All() {
			forbidden[fvarKey] = fvarNode
		}
	}
	for _, v := range variables {
		if _, captured := forbidden[Key(v)]; captured {
			return nil, &LogicUtilCaptureError{Variables: []*Variable{v}}
		}
	}

	newBody, err := substituteRec(body, newsubs)
	if err != nil {
		return nil, err
	}
	return construct(variables, newBody)
}

func substituteNamedBinder(nb *NamedBinder, subs map[NodeKey]Expr) (Expr, error) {
	newsubs := make(map[NodeKey]Expr)
	varSet := make(map[NodeKey]struct{})
	for _, v := range nb.Variables {
		varSet[Key(v)] = struct{}{}
	}
	for k, v := range subs {
		if _, isBound := varSet[k]; !isBound {
			newsubs[k] = v
		}
	}

	forbidden := make(map[NodeKey]Expr)
	for _, v := range newsubs {
		fv := FreeVariables(v)
		for fvarKey, fvarNode := range fv.All() {
			forbidden[fvarKey] = fvarNode
		}
	}
	for _, v := range nb.Variables {
		if _, captured := forbidden[Key(v)]; captured {
			return nil, &LogicUtilCaptureError{Variables: []*Variable{v}}
		}
	}

	newBody, err := substituteRec(nb.Body, newsubs)
	if err != nil {
		return nil, err
	}
	return NewNamedBinder(nb.Name, nb.Variables, nb.Environ, newBody)
}

// IsTautologyEquality returns true if t is Eq(x, x) for some x.
func IsTautologyEquality(t Expr) bool {
	eq, ok := t.(*Eq)
	if !ok {
		return false
	}
	return eq.T1.Equal(eq.T2)
}

// pushableMap is a map with push/pop semantics for alpha-conversion.
type pushableMap struct {
	stack []pushEntry
	m     map[*Variable]int
}

type pushEntry struct {
	key *Variable
	val int
	had bool
}

func newPushableMap() *pushableMap {
	return &pushableMap{m: make(map[*Variable]int)}
}

func (pm *pushableMap) push(key *Variable, val int) {
	old, had := pm.m[key]
	pm.stack = append(pm.stack, pushEntry{key, old, had})
	pm.m[key] = val
}

func (pm *pushableMap) pop() {
	e := pm.stack[len(pm.stack)-1]
	pm.stack = pm.stack[:len(pm.stack)-1]
	if e.had {
		pm.m[e.key] = e.val
	} else {
		delete(pm.m, e.key)
	}
}

func (pm *pushableMap) get(key *Variable) (int, bool) {
	v, ok := pm.m[key]
	return v, ok
}

// EqualModAlpha returns true if t and u are syntactically equal modulo
// alpha-conversion (renaming of bound variables).
func EqualModAlpha(t, u Expr) bool {
	return equalModAlphaRec(t, u, newPushableMap(), newPushableMap(), 0)
}

func equalModAlphaRec(t, u Expr, m1, m2 *pushableMap, n int) bool {
	tv, tIsVar := t.(*Variable)
	uv, uIsVar := u.(*Variable)
	if tIsVar && uIsVar {
		tn, tok := m1.get(tv)
		un, uok := m2.get(uv)
		if tok && uok {
			return tn == un
		}
		if !tok && !uok {
			return tv.Equal(uv)
		}
		return false
	}

	// Binders: ForAll, Exists, Lambda, NamedBinder
	type binder interface {
		Expr
	}
	switch tb := t.(type) {
	case *ForAll:
		ub, ok := u.(*ForAll)
		if !ok || len(tb.Variables) != len(ub.Variables) {
			return false
		}
		for i := range tb.Variables {
			m1.push(tb.Variables[i], n)
			m2.push(ub.Variables[i], n)
			n++
		}
		res := equalModAlphaRec(tb.Body, ub.Body, m1, m2, n)
		for range tb.Variables {
			m1.pop()
			m2.pop()
		}
		return res

	case *Exists:
		ub, ok := u.(*Exists)
		if !ok || len(tb.Variables) != len(ub.Variables) {
			return false
		}
		for i := range tb.Variables {
			m1.push(tb.Variables[i], n)
			m2.push(ub.Variables[i], n)
			n++
		}
		res := equalModAlphaRec(tb.Body, ub.Body, m1, m2, n)
		for range tb.Variables {
			m1.pop()
			m2.pop()
		}
		return res

	case *Lambda:
		ub, ok := u.(*Lambda)
		if !ok || len(tb.Variables) != len(ub.Variables) {
			return false
		}
		for i := range tb.Variables {
			m1.push(tb.Variables[i], n)
			m2.push(ub.Variables[i], n)
			n++
		}
		res := equalModAlphaRec(tb.Body, ub.Body, m1, m2, n)
		for range tb.Variables {
			m1.pop()
			m2.pop()
		}
		return res

	case *NamedBinder:
		ub, ok := u.(*NamedBinder)
		if !ok || tb.Name != ub.Name || len(tb.Variables) != len(ub.Variables) {
			return false
		}
		for i := range tb.Variables {
			m1.push(tb.Variables[i], n)
			m2.push(ub.Variables[i], n)
			n++
		}
		res := equalModAlphaRec(tb.Body, ub.Body, m1, m2, n)
		for range tb.Variables {
			m1.pop()
			m2.pop()
		}
		return res
	}

	// Apply: compare Func recursively (not just structurally) and terms pairwise.
	// Children() returns only Terms, so Func must be compared explicitly.
	ta, tIsApply := t.(*Apply)
	ua, uIsApply := u.(*Apply)
	if tIsApply && uIsApply {
		if !equalModAlphaRec(ta.Func, ua.Func, m1, m2, n) {
			return false
		}
		if len(ta.Terms) != len(ua.Terms) {
			return false
		}
		for i := range ta.Terms {
			if !equalModAlphaRec(ta.Terms[i], ua.Terms[i], m1, m2, n) {
				return false
			}
		}
		return true
	}

	// Const
	tc, tIsConst := t.(*Const)
	uc, uIsConst := u.(*Const)
	if tIsConst && uIsConst {
		return tc.Equal(uc)
	}

	// Generic: same type, same number of children, children match
	tChildren := t.Children()
	uChildren := u.Children()
	if fmt.Sprintf("%T", t) != fmt.Sprintf("%T", u) {
		return false
	}
	if len(tChildren) != len(uChildren) {
		return false
	}
	for i := range tChildren {
		if !equalModAlphaRec(tChildren[i], uChildren[i], m1, m2, n) {
			return false
		}
	}
	return true
}

// FreeVariablesList returns free variables as a slice in sorted NodeKey order
// (alphabetical). Use VariablesAstList when DFS first-occurrence order is needed
// (matching Python's list(iu.unique(ilu.variables_ast(t)))).
func FreeVariablesList(t Expr) []*Variable {
	fv := FreeVariables(t)
	result := make([]*Variable, 0, fv.Len())
	for _, node := range fv.All() {
		if v, ok := node.(*Variable); ok {
			result = append(result, v)
		}
	}
	return result
}

type binderLike interface {
	BinderVars() []*Variable
	BinderBody() Expr
}

// VariablesAstList returns the free variables of t in DFS first-occurrence
// order, matching Python's list(iu.unique(ilu.variables_ast(t))).
// This is the Go equivalent of Python ivy_logic_utils.py:474-486 variables_ast
// wrapped with iu.unique (ivy_utils.py:48-55).
func VariablesAstList(t Expr) []*Variable {
	var result []*Variable
	seen := make(map[NodeKey]bool)
	variablesAstRec(t, &result, seen, nil)
	return result
}

func variablesAstRec(t Expr, result *[]*Variable, seen map[NodeKey]bool, bound map[NodeKey]bool) {
	switch n := t.(type) {
	case *Variable:
		k := Key(n)
		if !bound[k] && !seen[k] {
			seen[k] = true
			*result = append(*result, n)
		}
	case *ForAll:
		newBound := copyBoolKeySet(bound)
		for _, v := range n.Variables {
			newBound[Key(v)] = true
		}
		variablesAstRec(n.Body, result, seen, newBound)
	case *Exists:
		newBound := copyBoolKeySet(bound)
		for _, v := range n.Variables {
			newBound[Key(v)] = true
		}
		variablesAstRec(n.Body, result, seen, newBound)
	case *Lambda:
		newBound := copyBoolKeySet(bound)
		for _, v := range n.Variables {
			newBound[Key(v)] = true
		}
		variablesAstRec(n.Body, result, seen, newBound)
	case *NamedBinder:
		newBound := copyBoolKeySet(bound)
		for _, v := range n.Variables {
			newBound[Key(v)] = true
		}
		variablesAstRec(n.Body, result, seen, newBound)
	default:
		if b, ok := t.(binderLike); ok {
			newBound := copyBoolKeySet(bound)
			for _, v := range b.BinderVars() {
				newBound[Key(v)] = true
			}
			variablesAstRec(b.BinderBody(), result, seen, newBound)
			return
		}
		for _, c := range t.Children() {
			variablesAstRec(c, result, seen, bound)
		}
	}
}

func copyBoolKeySet(s map[NodeKey]bool) map[NodeKey]bool {
	r := make(map[NodeKey]bool, len(s))
	for k, v := range s {
		r[k] = v
	}
	return r
}

// UsedVariablesAsts mirrors Python ivy_logic_utils.used_variables_asts
// = apply_gen_to_list(variables_ast). Returns free-variable occurrences
// (NOT deduped) across every formula in fmlas, in DFS order. Matches
// Python's `list(lu.used_variables_asts(fmlas))` used by goal_vocab.
func UsedVariablesAsts(fmlas []Expr) []*Variable {
	var result []*Variable
	for _, f := range fmlas {
		variablesAstOccurrencesRec(f, &result, nil)
	}
	return result
}

func variablesAstOccurrencesRec(t Expr, result *[]*Variable, bound map[NodeKey]bool) {
	switch n := t.(type) {
	case *Variable:
		if _, isBound := bound[Key(n)]; !isBound {
			*result = append(*result, n)
		}
	case *ForAll:
		newBound := copyBoolKeySet(bound)
		for _, v := range n.Variables {
			newBound[Key(v)] = true
		}
		variablesAstOccurrencesRec(n.Body, result, newBound)
	case *Exists:
		newBound := copyBoolKeySet(bound)
		for _, v := range n.Variables {
			newBound[Key(v)] = true
		}
		variablesAstOccurrencesRec(n.Body, result, newBound)
	case *Lambda:
		newBound := copyBoolKeySet(bound)
		for _, v := range n.Variables {
			newBound[Key(v)] = true
		}
		variablesAstOccurrencesRec(n.Body, result, newBound)
	case *NamedBinder:
		newBound := copyBoolKeySet(bound)
		for _, v := range n.Variables {
			newBound[Key(v)] = true
		}
		variablesAstOccurrencesRec(n.Body, result, newBound)
	default:
		for _, c := range t.Children() {
			variablesAstOccurrencesRec(c, result, bound)
		}
	}
}

// --- helpers ---

func copyVarSet(s map[NodeKey]Expr) map[NodeKey]Expr {
	r := make(map[NodeKey]Expr, len(s))
	for k, v := range s {
		r[k] = v
	}
	return r
}

func logicutilCopyStringSet(s map[string]struct{}) map[string]struct{} {
	r := make(map[string]struct{}, len(s))
	for k, v := range s {
		r[k] = v
	}
	return r
}
