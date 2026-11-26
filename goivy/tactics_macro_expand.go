package goivy

import "fmt"

type proofMacro struct {
	parameters []*LogicVariable
	body       Expr
}

type macroExpansion struct {
	definitions map[NodeKey]proofMacro
	active      map[NodeKey]bool
	err         error
}

func proofMacroDefinition(premise *LabeledFormula) (*Const, proofMacro, bool) {
	formula, ok := premise.Formula.(Expr)
	if !ok {
		return nil, proofMacro{}, false
	}
	definition, ok := dropUniversals(formula).(*IvyDefinition)
	if !ok {
		return nil, proofMacro{}, false
	}
	macro := proofMacro{body: definition.Rhs}
	if constant, ok := definition.Lhs.(*Const); ok {
		return constant, macro, true
	}
	if app, ok := definition.Lhs.(*Apply); ok {
		symbol, isConstant := app.Func.(*Const)
		if !isConstant {
			return nil, proofMacro{}, false
		}
		seen := make(map[string]bool)
		for _, arg := range app.Terms {
			variable, ok := arg.(*LogicVariable)
			if !ok || seen[variable.Name] {
				return nil, proofMacro{}, false
			}
			seen[variable.Name] = true
			macro.parameters = append(macro.parameters, variable)
		}
		return symbol, macro, true
	}
	return nil, proofMacro{}, false
}

func (expansion *macroExpansion) rewrite(expr Expr) Expr {
	if expr == nil || expansion.err != nil {
		return expr
	}
	if action, ok := expr.(ActionsAction); ok {
		var args []Expr
		for _, arg := range action.ActionArgs() {
			args = append(args, expansion.rewrite(arg))
		}
		return action.ActionClone(args)
	}
	symbol, _ := expr.(*Const)
	children := expr.Children()
	if app, ok := expr.(*Apply); ok {
		symbol, _ = app.Func.(*Const)
		children = app.Terms
	}
	args := make([]Expr, len(children))
	for i, child := range children {
		args[i] = expansion.rewrite(child)
	}
	if symbol != nil {
		key := Key(symbol)
		if macro, ok := expansion.definitions[key]; ok && len(args) == len(macro.parameters) {
			if expansion.active[key] {
				expansion.err = fmt.Errorf("recursive proof macro %s", symbol.Name)
				return expr
			}
			expansion.active[key] = true
			body := expansion.rewrite(macro.body)
			delete(expansion.active, key)
			var variables []*LogicVariable
			substitutions := make(map[string]Expr)
			for i, arg := range args {
				variables = append(variables, FreeVariablesList(arg)...)
				substitutions[macro.parameters[i].Name] = arg
			}
			return SubstituteByName(AlphaAvoid(body, variables), substitutions)
		}
	}
	if len(args) == 0 {
		return expr
	}
	return CloneNode(expr, args)
}

func (expansion *macroExpansion) labeled(formulas []*LabeledFormula) []*LabeledFormula {
	result := make([]*LabeledFormula, len(formulas))
	for i, formula := range formulas {
		result[i] = formula.Clone([]Node{formula.Label, expansion.rewrite(formula.Formula.(Expr))}).(*LabeledFormula)
	}
	return result
}

func MacroExpandTactic(pc ProofCheckerInterface, goals []*LabeledFormula, proof Node) ([]*LabeledFormula, error) {
	if len(goals) == 0 {
		return goals, nil
	}
	goal := goals[0]
	model, ok := GoalConc(goal).(*TemporalModels)
	if !ok {
		return nil, NewIvyError(proof, "macro_expand applies only to invariance or temporal properties")
	}
	original, ok := model.Model.(*NormalProgram)
	if !ok {
		return nil, NewIvyError(proof, "macro_expand requires a normal program")
	}
	expansion := macroExpansion{definitions: make(map[NodeKey]proofMacro), active: make(map[NodeKey]bool)}
	var prems []Node
	for _, premise := range GoalPrems(goal) {
		if property, ok := premise.(*LabeledFormula); ok && GoalIsProperty(property) && property.IsDefinition {
			if symbol, macro, ok := proofMacroDefinition(property); ok {
				expansion.definitions[Key(symbol)] = macro
				continue
			}
		}
		prems = append(prems, premise)
	}
	for i, premise := range prems {
		if property, ok := premise.(*LabeledFormula); ok && GoalIsProperty(property) {
			prems[i] = expansion.labeled([]*LabeledFormula{property})[0]
		}
	}
	program := NormalProgramClone(original)
	program.Bindings = make([]*ActionTermBinding, 0, len(original.Bindings))
	for _, binding := range original.Bindings {
		copy := *binding
		copy.Action = binding.Action.CloneStmt(expansion.rewrite(binding.Action.Stmt).(ActionsAction))
		program.Bindings = append(program.Bindings, &copy)
	}
	if original.Init != nil {
		program.Init = expansion.rewrite(original.Init).(ActionsAction)
	}
	program.Invars = expansion.labeled(original.Invars)
	program.Asms = expansion.labeled(original.Asms)
	if original.Postconds != nil {
		program.Postconds = make(map[string][]*LabeledFormula)
		for name, postconditions := range original.Postconds {
			program.Postconds[name] = expansion.labeled(postconditions)
		}
	}
	conclusion := expansion.rewrite(model.Fmla.(Expr))
	if expansion.err != nil {
		return nil, expansion.err
	}
	cfg := pc.GetModule().Cfg.AstCfg
	result := CloneGoal(cfg, goal, prems, cfg.NewTemporalModels(program, conclusion))
	return append([]*LabeledFormula{result}, goals[1:]...), nil
}
