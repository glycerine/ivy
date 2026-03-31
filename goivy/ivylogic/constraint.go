package ivylogic

import (
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
)

// DefinitionToConstraint converts a logic.Definition to its constraint form.
// This is the Go port of Python ivy_logic.py Definition.to_constraint (line 236-249).
//
// The conversion rules are:
//   - If RHS is Some with if_value: complex conditional constraint
//   - If RHS is Some without if_value: ForAll(x, Implies(phi, substitute(phi, {x: lhs})))
//   - If LHS is Apply (function definition): Eq(lhs, rhs)
//   - Otherwise (propositional): Iff(lhs, rhs)
func DefinitionToConstraint(d *lg.Definition) lg.Expr {
	// Check if RHS is a Some (conditional definition)
	if some, ok := d.Rhs.(*Some); ok {
		return someToConstraint(d.Lhs, some)
	}
	// If LHS is individual (non-Boolean sort): use Eq(lhs, rhs)
	// Python: is_individual(self.args[0]) checks term.sort != lg.Boolean
	if IsIndividual(d.Lhs) {
		return &lg.Eq{T1: d.Lhs, T2: d.Rhs}
	}
	// Otherwise: use Iff(lhs, rhs) for propositional definitions
	return &lg.Iff{T1: d.Lhs, T2: d.Rhs}
}

// someToConstraint converts a definition with a Some RHS to a constraint.
// Python ivy_logic.py:237-246.
func someToConstraint(lhs lg.Expr, some *Some) lg.Expr {
	if len(some.Params) == 0 {
		// Degenerate: no parameters, treat as simple definition
		return &lg.Eq{T1: lhs, T2: some.Fmla}
	}
	x, ok := some.Params[0].(*lg.Variable)
	if !ok {
		// Fallback: treat as equality
		return &lg.Eq{T1: lhs, T2: some}
	}
	phi := some.Fmla

	if some.IfVal != nil {
		// Complex case: some X. phi in ifval else elseval
		//
		// Python:
		//   And(Implies(phi, Exists([x], And(phi, Eq(lhs, ifval)))),
		//       Or(Exists([x], phi), Eq(lhs, elseval)))
		ifVal := some.IfVal
		elseVal := some.ElseVal

		existsInner := &lg.Exists{
			Variables: []*lg.Variable{x},
			Body:      &lg.And{Terms: []lg.Expr{phi, &lg.Eq{T1: lhs, T2: ifVal}}},
		}
		arm1 := &lg.Implies{T1: phi, T2: existsInner}

		existsOuter := &lg.Exists{Variables: []*lg.Variable{x}, Body: phi}
		eqElse := &lg.Eq{T1: lhs, T2: elseVal}
		arm2 := &lg.Or{Terms: []lg.Expr{existsOuter, eqElse}}

		return &lg.And{Terms: []lg.Expr{arm1, arm2}}
	}

	// Simple case: some X. phi (no if/else)
	//
	// Python:
	//   ForAll([x], Implies(phi, substitute(phi, {x: lhs})))
	subs := map[lg.NodeKey]lg.Expr{lg.Key(x): lhs}
	substPhi, err := lu.Substitute(phi, subs)
	if err != nil {
		// Fallback on substitution error: use Eq
		return &lg.Eq{T1: lhs, T2: some}
	}
	return &lg.ForAll{
		Variables: []*lg.Variable{x},
		Body:      &lg.Implies{T1: phi, T2: substPhi},
	}
}
