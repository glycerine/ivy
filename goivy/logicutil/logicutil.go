package logicutil

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/logic"
)

// CaptureError is raised when a substitution would create variable capture.
type CaptureError struct {
	Variables []*logic.Variable
}

func (e *CaptureError) Error() string {
	return fmt.Sprintf("variable capture: %v", e.Variables)
}

// UsedVariables returns all variables used in the given term (both free and bound).
func UsedVariables(t logic.Expr) map[logic.NodeKey]logic.Expr {
	result := make(map[logic.NodeKey]logic.Expr)
	usedVariablesRec(t, result)
	return result
}

func usedVariablesRec(t logic.Expr, result map[logic.NodeKey]logic.Expr) {
	switch n := t.(type) {
	case *logic.Variable:
		result[logic.Key(n)] = n
	case *logic.ForAll:
		for _, v := range n.Variables {
			result[logic.Key(v)] = v
		}
		usedVariablesRec(n.Body, result)
	case *logic.Exists:
		for _, v := range n.Variables {
			result[logic.Key(v)] = v
		}
		usedVariablesRec(n.Body, result)
	case *logic.Lambda:
		for _, v := range n.Variables {
			result[logic.Key(v)] = v
		}
		usedVariablesRec(n.Body, result)
	case *logic.NamedBinder:
		for _, v := range n.Variables {
			result[logic.Key(v)] = v
		}
		usedVariablesRec(n.Body, result)
	default:
		for _, c := range t.Children() {
			usedVariablesRec(c, result)
		}
	}
}

// FreeVariables returns the set of variables free in the given term.
// Variables are compared by object identity (pointer equality).
func FreeVariables(t logic.Expr) map[logic.NodeKey]logic.Expr {
	result := make(map[logic.NodeKey]logic.Expr)
	freeVariablesRec(t, result, nil)
	return result
}

// FreeVariablesSet returns a set of free variable names (map[string]bool).
func FreeVariablesSet(t logic.Expr) map[string]bool {
	byName := FreeVariablesByName(t)
	result := make(map[string]bool, len(byName))
	for name := range byName {
		result[name] = true
	}
	return result
}

// FreeVariablesByName returns the set of variable names free in the given term.
// Binding occurs by name (ForAll X:s1 binds X:s2).
func FreeVariablesByName(t logic.Expr) map[string]struct{} {
	result := make(map[string]struct{})
	freeVariablesByNameRec(t, result, nil)
	return result
}

func freeVariablesRec(t logic.Expr, result map[logic.NodeKey]logic.Expr, bound map[logic.NodeKey]logic.Expr) {
	switch n := t.(type) {
	case *logic.Variable:
		if _, isBound := bound[logic.Key(n)]; !isBound {
			result[logic.Key(n)] = n
		}
	case *logic.ForAll:
		newBound := copyVarSet(bound)
		for _, v := range n.Variables {
			newBound[logic.Key(v)] = v
		}
		freeVariablesRec(n.Body, result, newBound)
	case *logic.Exists:
		newBound := copyVarSet(bound)
		for _, v := range n.Variables {
			newBound[logic.Key(v)] = v
		}
		freeVariablesRec(n.Body, result, newBound)
	case *logic.Lambda:
		newBound := copyVarSet(bound)
		for _, v := range n.Variables {
			newBound[logic.Key(v)] = v
		}
		freeVariablesRec(n.Body, result, newBound)
	case *logic.NamedBinder:
		newBound := copyVarSet(bound)
		for _, v := range n.Variables {
			newBound[logic.Key(v)] = v
		}
		freeVariablesRec(n.Body, result, newBound)
	default:
		for _, c := range t.Children() {
			freeVariablesRec(c, result, bound)
		}
	}
}

func freeVariablesByNameRec(t logic.Expr, result map[string]struct{}, bound map[string]struct{}) {
	switch n := t.(type) {
	case *logic.Variable:
		if _, isBound := bound[n.Name]; !isBound {
			result[n.Name] = struct{}{}
		}
	case *logic.ForAll:
		newBound := copyStringSet(bound)
		for _, v := range n.Variables {
			newBound[v.Name] = struct{}{}
		}
		freeVariablesByNameRec(n.Body, result, newBound)
	case *logic.Exists:
		newBound := copyStringSet(bound)
		for _, v := range n.Variables {
			newBound[v.Name] = struct{}{}
		}
		freeVariablesByNameRec(n.Body, result, newBound)
	case *logic.Lambda:
		newBound := copyStringSet(bound)
		for _, v := range n.Variables {
			newBound[v.Name] = struct{}{}
		}
		freeVariablesByNameRec(n.Body, result, newBound)
	case *logic.NamedBinder:
		newBound := copyStringSet(bound)
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
func BoundVariables(t logic.Expr) map[logic.NodeKey]logic.Expr {
	result := make(map[logic.NodeKey]logic.Expr)
	boundVariablesRec(t, result)
	return result
}

func boundVariablesRec(t logic.Expr, result map[logic.NodeKey]logic.Expr) {
	switch n := t.(type) {
	case *logic.Variable:
		// leaf — no bound variables
	case *logic.ForAll:
		for _, v := range n.Variables {
			result[logic.Key(v)] = v
		}
		boundVariablesRec(n.Body, result)
	case *logic.Exists:
		for _, v := range n.Variables {
			result[logic.Key(v)] = v
		}
		boundVariablesRec(n.Body, result)
	case *logic.Lambda:
		for _, v := range n.Variables {
			result[logic.Key(v)] = v
		}
		boundVariablesRec(n.Body, result)
	case *logic.NamedBinder:
		for _, v := range n.Variables {
			result[logic.Key(v)] = v
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
func Substitute(t logic.Expr, subs map[logic.NodeKey]logic.Expr) (logic.Expr, error) {
	if len(subs) == 0 {
		return t, nil
	}
	return substituteRec(t, subs)
}

func substituteRec(t logic.Expr, subs map[logic.NodeKey]logic.Expr) (logic.Expr, error) {
	switch n := t.(type) {
	case *logic.Variable:
		if r, ok := subs[logic.Key(n)]; ok {
			return r, nil
		}
		return t, nil

	case *logic.Const:
		if r, ok := subs[logic.Key(n)]; ok {
			return r, nil
		}
		return t, nil

	case *logic.Apply:
		// Substitute in terms only, NOT in Func — matches Python substitute_ast
		// which iterates ast.args (terms only) and preserves ast.func via clone.
		newTerms := make([]logic.Expr, len(n.Terms))
		for i, term := range n.Terms {
			nt, err := substituteRec(term, subs)
			if err != nil {
				return nil, err
			}
			newTerms[i] = nt
		}
		return logic.NewApply(n.Func, newTerms...)

	case *logic.Eq:
		t1, err := substituteRec(n.T1, subs)
		if err != nil {
			return nil, err
		}
		t2, err := substituteRec(n.T2, subs)
		if err != nil {
			return nil, err
		}
		return logic.NewEq(t1, t2)

	case *logic.Ite:
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
		return logic.NewIte(c, th, el)

	case *logic.Not:
		b, err := substituteRec(n.Body, subs)
		if err != nil {
			return nil, err
		}
		return logic.NewNot(b)

	case *logic.And:
		terms := make([]logic.Expr, len(n.Terms))
		for i, term := range n.Terms {
			nt, err := substituteRec(term, subs)
			if err != nil {
				return nil, err
			}
			terms[i] = nt
		}
		return logic.NewAnd(terms...)

	case *logic.Or:
		terms := make([]logic.Expr, len(n.Terms))
		for i, term := range n.Terms {
			nt, err := substituteRec(term, subs)
			if err != nil {
				return nil, err
			}
			terms[i] = nt
		}
		return logic.NewOr(terms...)

	case *logic.Implies:
		t1, err := substituteRec(n.T1, subs)
		if err != nil {
			return nil, err
		}
		t2, err := substituteRec(n.T2, subs)
		if err != nil {
			return nil, err
		}
		return logic.NewImplies(t1, t2)

	case *logic.Iff:
		t1, err := substituteRec(n.T1, subs)
		if err != nil {
			return nil, err
		}
		t2, err := substituteRec(n.T2, subs)
		if err != nil {
			return nil, err
		}
		return logic.NewIff(t1, t2)

	case *logic.Globally:
		b, err := substituteRec(n.Body, subs)
		if err != nil {
			return nil, err
		}
		return logic.NewGlobally(n.Environ, b)

	case *logic.Eventually:
		b, err := substituteRec(n.Body, subs)
		if err != nil {
			return nil, err
		}
		return logic.NewEventually(n.Environ, b)

	case *logic.WhenOperator:
		t1, err := substituteRec(n.T1, subs)
		if err != nil {
			return nil, err
		}
		t2, err := substituteRec(n.T2, subs)
		if err != nil {
			return nil, err
		}
		return logic.NewWhenOperator(n.Name, t1, t2)

	case *logic.ForAll:
		return substituteBinder(n.Variables, n.Body, subs, func(vars []*logic.Variable, body logic.Expr) (logic.Expr, error) {
			return logic.NewForAll(vars, body)
		})

	case *logic.Exists:
		return substituteBinder(n.Variables, n.Body, subs, func(vars []*logic.Variable, body logic.Expr) (logic.Expr, error) {
			return logic.NewExists(vars, body)
		})

	case *logic.Lambda:
		return substituteBinder(n.Variables, n.Body, subs, func(vars []*logic.Variable, body logic.Expr) (logic.Expr, error) {
			return logic.NewLambda(vars, body)
		})

	case *logic.NamedBinder:
		return substituteNamedBinder(n, subs)

	default:
		return nil, fmt.Errorf("substitute: unsupported node type: %T", t)
	}
}

// substituteBinder handles substitution into ForAll, Exists, Lambda.
func substituteBinder(
	variables []*logic.Variable,
	body logic.Expr,
	subs map[logic.NodeKey]logic.Expr,
	construct func([]*logic.Variable, logic.Expr) (logic.Expr, error),
) (logic.Expr, error) {
	// Remove bound variables from substitution
	newsubs := make(map[logic.NodeKey]logic.Expr)
	varSet := make(map[logic.NodeKey]struct{})
	for _, v := range variables {
		varSet[logic.Key(v)] = struct{}{}
	}
	for k, v := range subs {
		if _, isBound := varSet[k]; !isBound {
			newsubs[k] = v
		}
	}

	// Check for variable capture
	forbidden := make(map[logic.NodeKey]logic.Expr)
	for _, v := range newsubs {
		fv := FreeVariables(v)
		for fvarKey, fvarNode := range fv {
			forbidden[fvarKey] = fvarNode
		}
	}
	for _, v := range variables {
		if _, captured := forbidden[logic.Key(v)]; captured {
			return nil, &CaptureError{Variables: []*logic.Variable{v}}
		}
	}

	newBody, err := substituteRec(body, newsubs)
	if err != nil {
		return nil, err
	}
	return construct(variables, newBody)
}

func substituteNamedBinder(nb *logic.NamedBinder, subs map[logic.NodeKey]logic.Expr) (logic.Expr, error) {
	newsubs := make(map[logic.NodeKey]logic.Expr)
	varSet := make(map[logic.NodeKey]struct{})
	for _, v := range nb.Variables {
		varSet[logic.Key(v)] = struct{}{}
	}
	for k, v := range subs {
		if _, isBound := varSet[k]; !isBound {
			newsubs[k] = v
		}
	}

	forbidden := make(map[logic.NodeKey]logic.Expr)
	for _, v := range newsubs {
		fv := FreeVariables(v)
		for fvarKey, fvarNode := range fv {
			forbidden[fvarKey] = fvarNode
		}
	}
	for _, v := range nb.Variables {
		if _, captured := forbidden[logic.Key(v)]; captured {
			return nil, &CaptureError{Variables: []*logic.Variable{v}}
		}
	}

	newBody, err := substituteRec(nb.Body, newsubs)
	if err != nil {
		return nil, err
	}
	return logic.NewNamedBinder(nb.Name, nb.Variables, nb.Environ, newBody)
}

// IsTautologyEquality returns true if t is Eq(x, x) for some x.
func IsTautologyEquality(t logic.Expr) bool {
	eq, ok := t.(*logic.Eq)
	if !ok {
		return false
	}
	return eq.T1.Equal(eq.T2)
}

// pushableMap is a map with push/pop semantics for alpha-conversion.
type pushableMap struct {
	stack []pushEntry
	m     map[*logic.Variable]int
}

type pushEntry struct {
	key *logic.Variable
	val int
	had bool
}

func newPushableMap() *pushableMap {
	return &pushableMap{m: make(map[*logic.Variable]int)}
}

func (pm *pushableMap) push(key *logic.Variable, val int) {
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

func (pm *pushableMap) get(key *logic.Variable) (int, bool) {
	v, ok := pm.m[key]
	return v, ok
}

// EqualModAlpha returns true if t and u are syntactically equal modulo
// alpha-conversion (renaming of bound variables).
func EqualModAlpha(t, u logic.Expr) bool {
	return equalModAlphaRec(t, u, newPushableMap(), newPushableMap(), 0)
}

func equalModAlphaRec(t, u logic.Expr, m1, m2 *pushableMap, n int) bool {
	tv, tIsVar := t.(*logic.Variable)
	uv, uIsVar := u.(*logic.Variable)
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
		logic.Expr
	}
	switch tb := t.(type) {
	case *logic.ForAll:
		ub, ok := u.(*logic.ForAll)
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

	case *logic.Exists:
		ub, ok := u.(*logic.Exists)
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

	case *logic.Lambda:
		ub, ok := u.(*logic.Lambda)
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

	case *logic.NamedBinder:
		ub, ok := u.(*logic.NamedBinder)
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
	ta, tIsApply := t.(*logic.Apply)
	ua, uIsApply := u.(*logic.Apply)
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
	tc, tIsConst := t.(*logic.Const)
	uc, uIsConst := u.(*logic.Const)
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

// FreeVariablesList returns free variables as a slice (convenience).
func FreeVariablesList(t logic.Expr) []*logic.Variable {
	fv := FreeVariables(t)
	result := make([]*logic.Variable, 0, len(fv))
	for _, node := range fv {
		if v, ok := node.(*logic.Variable); ok {
			result = append(result, v)
		}
	}
	return result
}

// --- helpers ---

func copyVarSet(s map[logic.NodeKey]logic.Expr) map[logic.NodeKey]logic.Expr {
	r := make(map[logic.NodeKey]logic.Expr, len(s))
	for k, v := range s {
		r[k] = v
	}
	return r
}

func copyStringSet(s map[string]struct{}) map[string]struct{} {
	r := make(map[string]struct{}, len(s))
	for k, v := range s {
		r[k] = v
	}
	return r
}
