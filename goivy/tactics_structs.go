package goivy

import (
	"fmt"
	"slices"
	"strings"
)

type structField struct {
	path string
	typ  Sort
}

type flatTerm struct {
	path string
	term Expr
}

type structFlattening struct {
	mod     *Module
	symbols *InsMap[NodeKey, *Const]
	locals  map[NodeKey]bool
	err     error
}

func (f *structFlattening) fieldTypes(typ Sort, active map[string]bool) []structField {
	name := IvySortName(typ)
	fields := f.mod.SortDestructors.Get(name)
	if len(fields) == 0 {
		return []structField{{typ: typ}}
	}
	if active[name] {
		f.err = fmt.Errorf("cannot flatten recursive structure %s", name)
		return nil
	}
	active[name] = true
	defer delete(active, name)
	var result []structField
	for _, field := range fields {
		if len(SortDomain(field.CSort)) != 1 {
			f.err = fmt.Errorf("cannot flatten indexed structure field %s", field.Name)
			return nil
		}
		for _, leaf := range f.fieldTypes(SortRange(field.CSort), active) {
			leaf.path = "." + field.Name + leaf.path
			result = append(result, leaf)
		}
	}
	return result
}

func (f *structFlattening) splitSymbol(symbol *Const, args []Expr, result Sort) []flatTerm {
	var terms []flatTerm
	domain := make([]Sort, len(args))
	for i, arg := range args {
		domain[i] = arg.NodeSort()
	}
	for _, field := range f.fieldTypes(result, make(map[string]bool)) {
		var sort Sort = field.typ
		name := symbol.Name + field.path
		if len(domain) > 0 {
			sig, err := NewFunctionSort(append(append([]Sort(nil), domain...), sort)...)
			if err != nil {
				f.err = err
				return nil
			}
			sort = sig
			if IsPolymorphicName(symbol.Name) && !SortEqual(symbol.CSort, sort) {
				name = "$" + name
			}
		}
		fieldSymbol := NewConst(name, sort)
		if !f.locals[Key(symbol)] && !fieldSymbol.Equal(symbol) {
			f.symbols.Set(Key(fieldSymbol), fieldSymbol)
		}
		var term Expr = fieldSymbol
		if len(args) > 0 {
			term = NewApplyUnchecked(fieldSymbol, args...)
		}
		terms = append(terms, flatTerm{field.path, term})
	}
	return terms
}

func (f *structFlattening) expr(expr Expr) []flatTerm {
	if f.err != nil || expr == nil {
		return nil
	}
	if IsMacro(expr, f.mod.Cfg.IuCfg) {
		return f.expr(ExpandMacro(expr))
	}
	switch node := expr.(type) {
	case *Const:
		if IsFunctionSort(node.CSort) {
			return []flatTerm{{term: expr}}
		}
		return f.splitSymbol(node, nil, node.CSort)
	case *Apply:
		if symbol, ok := node.Func.(*Const); ok {
			if f.mod.DestructorSorts[symbol.Name] != nil && len(node.Terms) == 1 {
				prefix := "." + symbol.Name
				var terms []flatTerm
				for _, field := range f.expr(node.Terms[0]) {
					if field.path == prefix || strings.HasPrefix(field.path, prefix+".") {
						terms = append(terms, flatTerm{strings.TrimPrefix(field.path, prefix), field.term})
					}
				}
				return terms
			}
			if sort := f.mod.ConstructorSorts[symbol.Name]; sort != nil {
				fields := f.mod.SortDestructors.Get(IvySortName(sort))
				if len(fields) != len(node.Terms) {
					f.err = fmt.Errorf("constructor %s has the wrong number of fields", symbol.Name)
					return nil
				}
				var terms []flatTerm
				for i, arg := range node.Terms {
					for _, field := range f.expr(arg) {
						terms = append(terms, flatTerm{"." + fields[i].Name + field.path, field.term})
					}
				}
				return terms
			}
			return f.splitSymbol(symbol, f.terms(node.Terms), node.NodeSort())
		}
		return []flatTerm{{term: CloneApplyTerms(node, f.terms(node.Terms))}}
	case *LogicVariable:
		var terms []flatTerm
		for _, field := range f.fieldTypes(node.VSort, make(map[string]bool)) {
			terms = append(terms, flatTerm{field.path, &LogicVariable{Name: node.Name + field.path, VSort: field.typ}})
		}
		return terms
	case *Eq:
		return []flatTerm{{term: &LogicAnd{Terms: f.zip(node.T1, node.T2, func(a, b Expr) Expr { return &Eq{T1: a, T2: b} })}}}
	case *LogicIte:
		condition := f.scalar(node.Cond)
		var terms []flatTerm
		then, otherwise := f.expr(node.Then), f.expr(node.Else)
		if len(then) != len(otherwise) {
			f.err = fmt.Errorf("incompatible structure branches")
			return nil
		}
		for i, branch := range then {
			terms = append(terms, flatTerm{branch.path, &LogicIte{ISort: branch.term.NodeSort(), Cond: condition, Then: branch.term, Else: otherwise[i].term}})
		}
		return terms
	case *ForAll:
		return []flatTerm{{term: CloneBinder(node, f.vars(node.Variables), f.scalar(node.Body))}}
	case *LogicExists:
		return []flatTerm{{term: CloneBinder(node, f.vars(node.Variables), f.scalar(node.Body))}}
	case *LogicNamedBinder:
		return []flatTerm{{term: CloneBinder(node, f.vars(node.Variables), f.scalar(node.Body))}}
	}
	children := expr.Children()
	if len(children) == 0 {
		return []flatTerm{{term: expr}}
	}
	var terms []Expr
	for _, child := range children {
		terms = append(terms, f.scalar(child))
	}
	return []flatTerm{{term: CloneNode(expr, terms)}}
}

func (f *structFlattening) scalar(expr Expr) Expr {
	terms := f.expr(expr)
	if len(terms) == 1 && terms[0].path == "" {
		return terms[0].term
	}
	if f.err == nil {
		f.err = fmt.Errorf("expected scalar expression when flattening %s", expr)
	}
	return expr
}

func (f *structFlattening) terms(exprs []Expr) []Expr {
	var result []Expr
	for _, expr := range exprs {
		for _, field := range f.expr(expr) {
			result = append(result, field.term)
		}
	}
	return result
}

func (f *structFlattening) vars(variables []*LogicVariable) []*LogicVariable {
	var result []*LogicVariable
	for _, variable := range variables {
		for _, field := range f.expr(variable) {
			result = append(result, field.term.(*LogicVariable))
		}
	}
	return result
}

func (f *structFlattening) zip(a, b Expr, combine func(Expr, Expr) Expr) []Expr {
	left, right := f.expr(a), f.expr(b)
	if len(left) != len(right) {
		f.err = fmt.Errorf("incompatible values when flattening structures")
		return nil
	}
	var terms []Expr
	for i, lhs := range left {
		terms = append(terms, combine(lhs.term, right[i].term))
	}
	return terms
}

func (f *structFlattening) scope(symbols []*Const) func() {
	var added []NodeKey
	for _, symbol := range symbols {
		key := Key(symbol)
		if !f.locals[key] {
			f.locals[key] = true
			added = append(added, key)
		}
	}
	return func() {
		for _, key := range added {
			delete(f.locals, key)
		}
	}
}

func (f *structFlattening) constants(symbols []*Const) []*Const {
	var flattened []*Const
	for _, symbol := range symbols {
		for _, field := range f.expr(symbol) {
			flattened = append(flattened, field.term.(*Const))
		}
	}
	return flattened
}

func (f *structFlattening) action(action ActionsAction) ActionsAction {
	if action == nil || f.err != nil {
		return action
	}
	switch node := action.(type) {
	case *LogicAssignAction:
		assignments := f.zip(node.LHS, node.RHS, func(lhs, rhs Expr) Expr {
			return node.ActionClone([]Expr{lhs, rhs})
		})
		if len(assignments) == 1 {
			return assignments[0].(ActionsAction)
		}
		return f.simultaneousAssignments(node, assignments)
	case *LogicHavocAction:
		var actions []Expr
		for _, arg := range f.terms(node.ActionArgs()) {
			actions = append(actions, node.ActionClone([]Expr{arg}))
		}
		result := NewSequence(actions...)
		result.SetLineno(action.GetLineno())
		return result
	case *LogicLocalAction:
		var symbols []*Const
		for _, local := range node.Locals {
			if symbol, ok := local.(*Const); ok {
				symbols = append(symbols, symbol)
			}
		}
		defer f.scope(symbols)()
		args := f.terms(node.Locals)
		args = append(args, f.action(node.Body.(ActionsAction)))
		return node.ActionClone(args)
	case *LogicCallAction:
		callee := node.Callee
		if app, ok := callee.(*Apply); ok {
			callee = NewApplyUnchecked(app.Func, f.terms(app.Terms)...)
		}
		return node.ActionClone(append([]Expr{callee}, f.terms(node.ActualReturns)...))
	}
	var args []Expr
	for _, arg := range action.ActionArgs() {
		if child, ok := arg.(ActionsAction); ok {
			args = append(args, f.action(child))
		} else {
			args = append(args, f.scalar(arg))
		}
	}
	return action.ActionClone(args)
}

func (f *structFlattening) simultaneousAssignments(original *LogicAssignAction, assignments []Expr) ActionsAction {
	var locals, reads, writes []Expr
	for _, expr := range assignments {
		assignment := expr.(*LogicAssignAction)
		variables := FreeVariablesList(assignment.LHS)
		var domain []Sort
		var arguments []Expr
		for _, variable := range variables {
			domain = append(domain, variable.VSort)
			arguments = append(arguments, variable)
		}
		var sort Sort = assignment.RHS.NodeSort()
		if len(domain) > 0 {
			var err error
			sort, err = NewFunctionSort(append(domain, sort)...)
			if err != nil {
				f.err = err
				return original
			}
		}
		serial := f.mod.Cfg.IuCfg.ConstructorArgsCtr
		f.mod.Cfg.IuCfg.ConstructorArgsCtr++
		local := NewConst(fmt.Sprintf("loc:flatten%d", serial), sort)
		locals = append(locals, local)
		var value Expr = local
		if len(arguments) > 0 {
			value = NewApplyUnchecked(local, arguments...)
		}
		reads = append(reads, original.ActionClone([]Expr{value, assignment.RHS}))
		writes = append(writes, original.ActionClone([]Expr{assignment.LHS, value}))
	}
	sequence := NewSequence(append(reads, writes...)...)
	sequence.SetLineno(original.GetLineno())
	result := NewLocalActionOn(f.mod.Cfg.ActCfg, "flatten_structs", append(locals, sequence)...)
	result.SetLineno(original.GetLineno())
	return result
}

func (f *structFlattening) labeled(formulas []*LabeledFormula) []*LabeledFormula {
	result := make([]*LabeledFormula, len(formulas))
	for i, formula := range formulas {
		result[i] = formula.Clone([]Node{formula.Label, f.scalar(formula.Formula.(Expr))}).(*LabeledFormula)
	}
	return result
}

func FlattenStructsTactic(pc ProofCheckerInterface, goals []*LabeledFormula, proof Node) ([]*LabeledFormula, error) {
	if len(goals) == 0 {
		return goals, nil
	}
	goal := goals[0]
	model, ok := GoalConc(goal).(*TemporalModels)
	if !ok {
		return nil, NewIvyError(proof, "flatten_structs applies only to invariance or temporal properties")
	}
	original, ok := model.Model.(*NormalProgram)
	if !ok {
		return nil, NewIvyError(proof, "flatten_structs requires a normal program")
	}
	mod := pc.GetModule()
	flat := &structFlattening{mod: mod, symbols: NewInsMap[NodeKey, *Const](), locals: make(map[NodeKey]bool)}
	program := NormalProgramClone(original)
	program.Bindings = make([]*ActionTermBinding, 0, len(original.Bindings))
	for _, binding := range original.Bindings {
		old := binding.Action
		restore := flat.scope(slices.Concat(old.Inputs, old.Outputs))
		inputs := flat.constants(old.Inputs)
		outputs := flat.constants(old.Outputs)
		body := flat.action(old.Stmt)
		restore()
		if flat.err != nil {
			return nil, flat.err
		}
		body = body.ActionClone(body.ActionArgs())
		body.SetFormalParams(inputs)
		body.SetFormalReturns(outputs)
		term := old.CloneStmt(body)
		term.Inputs, term.Outputs = inputs, outputs
		copy := *binding
		copy.Action = term
		program.Bindings = append(program.Bindings, &copy)
	}
	program.Init = flat.action(original.Init)
	program.Invars = flat.labeled(original.Invars)
	program.Asms = flat.labeled(original.Asms)
	if original.Postconds != nil {
		program.Postconds = make(map[string][]*LabeledFormula)
		for name, conditions := range original.Postconds {
			program.Postconds[name] = flat.labeled(conditions)
		}
	}
	prems := append([]Node(nil), GoalPrems(goal)...)
	for i, premise := range prems {
		if formula, ok := premise.(*LabeledFormula); ok && GoalIsProperty(formula) {
			prems[i] = flat.labeled([]*LabeledFormula{formula})[0]
		}
	}
	conclusion := flat.scalar(model.Fmla.(Expr))
	if flat.err != nil {
		return nil, flat.err
	}
	for _, symbol := range flat.symbols.All() {
		prems = append(prems, mod.Cfg.AstCfg.NewConstantDecl(symbol))
	}
	result := CloneGoal(mod.Cfg.AstCfg, goal, prems, mod.Cfg.AstCfg.NewTemporalModels(program, conclusion))
	return append([]*LabeledFormula{result}, goals[1:]...), nil
}
