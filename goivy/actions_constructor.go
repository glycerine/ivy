package goivy

import "fmt"

// eliminateStructDestructors gives each ground structure value a shared tuple
// of constructor arguments. A bound selector keeps its lexical dependencies.
func eliminateStructDestructors(mod *Module, clauses *Clauses) *Clauses {
	if clauses == nil || clauses.IsFalse() || mod.SortConstructors.Len() == 0 {
		return clauses
	}
	memo := make(map[NodeKey]Expr)
	var constraints []Expr
	var rewrite func(Expr) Expr
	rewrite = func(expr Expr) Expr {
		if app, ok := expr.(*Apply); ok && len(app.Terms) == 1 {
			if selector, ok := app.Func.(*Const); ok && mod.DestructorSorts[selector.Name] != nil && len(FreeVariablesList(app)) == 0 {
				if value := memo[Key(app)]; value != nil {
					return value
				}
				structure := app.Terms[0]
				name := IvySortName(mod.DestructorSorts[selector.Name])
				fields := mod.SortDestructors.Get(name)
				constructors := mod.SortConstructors.Get(name)
				if len(constructors) > 0 && canBuildConstructorTuple(constructors[0], fields) {
					term := rewrite(structure)
					id := mod.Cfg.IuCfg.ConstructorArgsCtr
					mod.Cfg.IuCfg.ConstructorArgsCtr++
					components := make([]Expr, len(fields))
					for i, field := range fields {
						components[i] = NewConst(fmt.Sprintf("__%s$%s$%d", name, field.Name, id), SortRange(field.CSort))
						memo[Key(NewApplyUnchecked(field, structure))] = components[i]
						if field.Name == selector.Name {
							memo[Key(app)] = components[i]
						}
					}
					constraints = append(constraints, &Eq{T1: term, T2: NewApplyUnchecked(constructors[0], components...)})
					if value := memo[Key(app)]; value != nil {
						return value
					}
				}
			}
		}
		children := expr.Children()
		if len(children) == 0 {
			return expr
		}
		if app, ok := expr.(*Apply); ok {
			children = app.Terms
		}
		args := make([]Expr, len(children))
		for i, child := range children {
			args[i] = rewrite(child)
		}
		return CloneNode(expr, args)
	}
	var formulas []Expr
	for _, formula := range clauses.Fmlas {
		formulas = append(formulas, rewrite(formula))
	}
	var definitions []*IvyDefinition
	for _, definition := range clauses.Defs {
		definitions = append(definitions, &IvyDefinition{Lhs: definition.Lhs, Rhs: rewrite(definition.Rhs)})
	}
	if len(constraints) == 0 {
		return clauses
	}
	return clausesFromConjuncts(append(formulas, constraints...), definitions, clauses.Annot)
}

func canBuildConstructorTuple(constructor *Const, fields []*Const) bool {
	domain := SortDomain(constructor.CSort)
	if len(fields) == 0 || len(domain) != len(fields) {
		return false
	}
	for i, field := range fields {
		if len(SortDomain(field.CSort)) != 1 || !SortEqual(domain[i], SortRange(field.CSort)) {
			return false
		}
	}
	return true
}
