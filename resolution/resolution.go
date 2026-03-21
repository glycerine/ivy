// Ported to Go from ivy_resolution.py.

// Package resolution implements MGU (most general unifier) for atoms and terms,
// following the algorithm in ivy_resolution.py.
package resolution

import (
	"github.com/glycerine/goivy/logic"
)

// Term is an alias for logic.Expr (variables or constants).
type Term = logic.Expr

// Env maps variable names to terms.
type Env map[string]Term

// EnvFind follows the variable chain in env until a non-variable or
// unmapped variable is found.
func EnvFind(env Env, t Term) Term {
	v := t
	for {
		vr, ok := v.(*logic.Variable)
		if !ok {
			return v
		}
		next, found := env[vr.Name]
		if !found {
			return v
		}
		v = next
	}
}

// IsConstant returns true if the node is a *logic.Symbol.
func IsConstant(n Term) bool {
	_, ok := n.(*logic.Symbol)
	return ok
}

// rep returns the name of a Var or Const.
func rep(n Term) string {
	switch t := n.(type) {
	case *logic.Variable:
		return t.Name
	case *logic.Symbol:
		return t.Name
	default:
		return n.String()
	}
}

// TermsMGU attempts to find a most general unifier for two lists of terms.
// Returns (true, substitution) on success, (false, nil) on failure.
func TermsMGU(terms1, terms2 []Term) (bool, Env) {
	if len(terms1) != len(terms2) {
		return false, nil
	}

	env := make(Env)
	for i := range terms1 {
		t1 := terms1[i]
		t2 := terms2[i]
		// Check sort compatibility.
		s1 := t1.NodeSort()
		s2 := t2.NodeSort()
		if !s1.Equal(s2) {
			return false, nil
		}

		v1 := EnvFind(env, t1)
		v2 := EnvFind(env, t2)

		if IsConstant(v1) {
			if IsConstant(v2) {
				if rep(v1) != rep(v2) {
					return false, nil
				}
			} else {
				// v2 must be a variable not already in env
				env[rep(v2)] = v1
			}
		} else {
			// v1 is a variable
			if IsConstant(v2) || rep(v1) != rep(v2) {
				env[rep(v1)] = v2
			}
		}
	}

	// Resolve the substitution: follow chains to get final bindings.
	subs := make(Env, len(env))
	for k := range env {
		subs[k] = EnvFind(env, env[k])
	}
	return true, subs
}

// Atom represents an atomic formula with a relation name and arguments.
// In the existing logic package, atoms are represented as *logic.Apply nodes
// where Func is a *logic.Symbol with the relation name. This type provides
// a lightweight wrapper for the resolution API.
type Atom struct {
	RelName string
	Args    []Term
}

// NewAtom creates an Atom with the given relation name and arguments.
func NewAtom(relname string, args ...Term) *Atom {
	return &Atom{RelName: relname, Args: args}
}

// AtomFromApply extracts an Atom from a *logic.Apply node.
// Returns nil if the node is not an Apply with a Const function.
func AtomFromApply(app *logic.Apply) *Atom {
	if c, ok := app.Func.(*logic.Symbol); ok {
		return &Atom{RelName: c.Name, Args: app.Terms}
	}
	return nil
}

// MGU attempts unification of two atoms. Returns (true, substitution) on
// success, (false, nil) on failure.
func MGU(atom1, atom2 *Atom) (bool, Env) {
	if atom1.RelName != atom2.RelName {
		return false, nil
	}
	return TermsMGU(atom1.Args, atom2.Args)
}

// EqualityAtom creates an equality atom (= t1 t2).
func EqualityAtom(t1, t2 Term) *Atom {
	return &Atom{RelName: "=", Args: []Term{t1, t2}}
}

// TermsMGUEq is like TermsMGU but instead of failing on mismatched constants,
// it collects equalities. Returns (true, substitution, equalities) on success,
// (false, nil, nil) on failure (only fails on sort mismatch or length mismatch).
func TermsMGUEq(terms1, terms2 []Term) (bool, Env, []*Atom) {
	if len(terms1) != len(terms2) {
		return false, nil, nil
	}

	var eqs []*Atom
	env := make(Env)
	for i := range terms1 {
		t1 := terms1[i]
		t2 := terms2[i]
		s1 := t1.NodeSort()
		s2 := t2.NodeSort()
		if !s1.Equal(s2) {
			return false, nil, nil
		}

		v1 := EnvFind(env, t1)
		v2 := EnvFind(env, t2)

		if IsConstant(v1) {
			if IsConstant(v2) {
				if rep(v1) != rep(v2) {
					eqs = append(eqs, EqualityAtom(v1, v2))
				}
			} else {
				env[rep(v2)] = v1
			}
		} else {
			if IsConstant(v2) || rep(v1) != rep(v2) {
				env[rep(v1)] = v2
			}
		}
	}

	subs := make(Env, len(env))
	for k := range env {
		subs[k] = EnvFind(env, env[k])
	}
	return true, subs, eqs
}

// MGUEq is like MGU but collects equalities instead of failing on constant
// mismatches.
func MGUEq(atom1, atom2 *Atom) (bool, Env, []*Atom) {
	if atom1.RelName != atom2.RelName {
		return false, nil, nil
	}
	return TermsMGUEq(atom1.Args, atom2.Args)
}
