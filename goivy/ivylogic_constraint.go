package goivy

// DefinitionToConstraint converts a logic.Definition to its constraint form.
// This is the Go port of Python ivy_logic.py IvyDefinition.to_constraint (line 236-249).
//
// The conversion rules are:
//   - If RHS is Some with if_value: complex conditional constraint
//   - If RHS is Some without if_value: IvyForAll(x, Implies(phi, substitute(phi, {x: lhs})))
//   - If LHS is Apply (function definition): Eq(lhs, rhs)
//   - Otherwise (propositional): Iff(lhs, rhs)
func DefinitionToConstraint(d *Definition) Expr {
	// Check if RHS is a Some (conditional definition)
	if some, ok := d.Rhs.(*Some); ok {
		return someToConstraint(d.Lhs, some)
	}
	// If LHS is individual (non-Boolean sort): use Eq(lhs, rhs)
	// Python: is_individual(self.args[0]) checks term.sort != lg.Boolean
	if IsIndividual(d.Lhs) {
		return &Eq{T1: d.Lhs, T2: d.Rhs}
	}
	// Otherwise: use Iff(lhs, rhs) for propositional definitions
	return &Iff{T1: d.Lhs, T2: d.Rhs}
}

// someToConstraint converts a definition with a Some RHS to a constraint.
// Python ivy_logic.py:237-246.
func someToConstraint(lhs Expr, some *Some) Expr {
	if len(some.Params) == 0 {
		// Degenerate: no parameters, treat as simple definition
		return &Eq{T1: lhs, T2: some.Fmla}
	}
	x, ok := some.Params[0].(*Variable)
	if !ok {
		// Fallback: treat as equality
		return &Eq{T1: lhs, T2: some}
	}
	phi := some.Fmla

	if some.IfVal != nil {
		// Complex case: some X. phi in ifval else elseval
		//
		// Python:
		//   And(Implies(phi, IvyExists([x], And(phi, Eq(lhs, ifval)))),
		//       Or(IvyExists([x], phi), Eq(lhs, elseval)))
		ifVal := some.IfVal
		elseVal := some.ElseVal

		existsInner := &Exists{
			Variables: []*Variable{x},
			Body:      &And{Terms: []Expr{phi, &Eq{T1: lhs, T2: ifVal}}},
		}
		arm1 := &Implies{T1: phi, T2: existsInner}

		existsOuter := &Exists{Variables: []*Variable{x}, Body: phi}
		eqElse := &Eq{T1: lhs, T2: elseVal}
		arm2 := &Or{Terms: []Expr{existsOuter, eqElse}}

		return &And{Terms: []Expr{arm1, arm2}}
	}

	// Simple case: some X. phi (no if/else)
	//
	// Python:
	//   IvyForAll([x], Implies(phi, substitute(phi, {x: lhs})))
	subs := map[NodeKey]Expr{Key(x): lhs}
	substPhi, err := Substitute(phi, subs)
	if err != nil {
		// Fallback on substitution error: use Eq
		return &Eq{T1: lhs, T2: some}
	}
	return &ForAll{
		Variables: []*Variable{x},
		Body:      &Implies{T1: phi, T2: substPhi},
	}
}
