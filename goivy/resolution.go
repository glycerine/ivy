// Ported to Go from ivy_resolution.py.

// Package resolution implements MGU (most general unifier) for atoms and terms,
// following the algorithm in ivy_resolution.py.
package goivy

// Term is an alias for logic.Expr (variables or constants).
type ResolutionTerm = Expr

// Env maps variable names to terms.
type Env map[string]ResolutionTerm

// EnvFind follows the variable chain in env until a non-variable or
// unmapped variable is found.
func EnvFind(env Env, t ResolutionTerm) ResolutionTerm {
	v := t
	for {
		vr, ok := v.(*LogicVariable)
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

// IsConstant returns true if the node is a *logic.Const.
func ResolutionIsConstant(n ResolutionTerm) bool {
	_, ok := n.(*Const)
	return ok
}

// rep returns the name of a Var or Const.
func rep(n ResolutionTerm) string {
	switch t := n.(type) {
	case *LogicVariable:
		return t.Name
	case *Const:
		return t.Name
	default:
		return n.String()
	}
}

// TermsMGU attempts to find a most general unifier for two lists of terms.
// Returns (true, substitution) on success, (false, nil) on failure.
func TermsMGU(terms1, terms2 []ResolutionTerm) (bool, Env) {
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

		if ResolutionIsConstant(v1) {
			if ResolutionIsConstant(v2) {
				if rep(v1) != rep(v2) {
					return false, nil
				}
			} else {
				// v2 must be a variable not already in env
				env[rep(v2)] = v1
			}
		} else {
			// v1 is a variable
			if ResolutionIsConstant(v2) || rep(v1) != rep(v2) {
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
// where Func is a *logic.Const with the relation name. This type provides
// a lightweight wrapper for the resolution API.
type ResolutionAtom struct {
	RelName string
	Args    []ResolutionTerm
}

// NewAtom creates an Atom with the given relation name and arguments.
func NewAtom(relname string, args ...ResolutionTerm) *ResolutionAtom {
	return &ResolutionAtom{RelName: relname, Args: args}
}

// AtomFromApply extracts an Atom from a *logic.Apply node.
// Returns nil if the node is not an Apply with a Const function.
func AtomFromApply(app *Apply) *ResolutionAtom {
	if c, ok := app.Func.(*Const); ok {
		return &ResolutionAtom{RelName: c.Name, Args: app.Terms}
	}
	return nil
}

// MGU attempts unification of two atoms. Returns (true, substitution) on
// success, (false, nil) on failure.
func MGU(atom1, atom2 *ResolutionAtom) (bool, Env) {
	if atom1.RelName != atom2.RelName {
		return false, nil
	}
	return TermsMGU(atom1.Args, atom2.Args)
}

// EqualityAtom creates an equality atom (= t1 t2).
func EqualityAtom(t1, t2 ResolutionTerm) *ResolutionAtom {
	return &ResolutionAtom{RelName: "=", Args: []ResolutionTerm{t1, t2}}
}

// TermsMGUEq is like TermsMGU but instead of failing on mismatched constants,
// it collects equalities. Returns (true, substitution, equalities) on success,
// (false, nil, nil) on failure (only fails on sort mismatch or length mismatch).
func TermsMGUEq(terms1, terms2 []ResolutionTerm) (bool, Env, []*ResolutionAtom) {
	if len(terms1) != len(terms2) {
		return false, nil, nil
	}

	var eqs []*ResolutionAtom
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

		if ResolutionIsConstant(v1) {
			if ResolutionIsConstant(v2) {
				if rep(v1) != rep(v2) {
					eqs = append(eqs, EqualityAtom(v1, v2))
				}
			} else {
				env[rep(v2)] = v1
			}
		} else {
			if ResolutionIsConstant(v2) || rep(v1) != rep(v2) {
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
func MGUEq(atom1, atom2 *ResolutionAtom) (bool, Env, []*ResolutionAtom) {
	if atom1.RelName != atom2.RelName {
		return false, nil, nil
	}
	return TermsMGUEq(atom1.Args, atom2.Args)
}
