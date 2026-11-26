package goivy

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
)

var mypyvyNames = strings.NewReplacer("_", "__", ".", "_d_", ":", "_c_", "@", "_at_", "#", "_h_", "$", "_s_", "-", "_m_",
	"+", "_plus_", "*", "_mul_", "/", "_div_", "<", "_lt_", ">", "_gt_", "=", "_eq_", "[", "_lb_", "]", "_rb_")

func mypyvyName(name string) string {
	return mypyvyNames.Replace(name)
}

type mypyvyForm struct {
	numeric map[NodeKey]string
	err     error
}

func (p *mypyvyForm) render(expr Expr) string {
	switch f := expr.(type) {
	case *Const:
		if name, ok := p.numeric[Key(f)]; ok {
			return name
		}
		if IsNew(f.Name) && !mypyvyLocal(f) {
			return "new(" + mypyvyName(NewOf(f.Name)) + ")"
		}
		return mypyvyName(f.Name)
	case *LogicVariable:
		return mypyvyName(f.Name)
	case *Apply:
		if f.Func == nil {
			break
		}
		var args []string
		for _, term := range f.Terms {
			args = append(args, p.render(term))
		}
		if symbol, ok := f.Func.(*Const); ok && IsNew(symbol.Name) && !mypyvyLocal(symbol) {
			return "new(" + mypyvyName(NewOf(symbol.Name)) + "(" + strings.Join(args, ", ") + "))"
		}
		if len(args) == 0 {
			return p.render(f.Func)
		}
		return p.render(f.Func) + "(" + strings.Join(args, ", ") + ")"
	case *LogicAnd:
		return p.join(f.Terms, " & ", "true")
	case *LogicOr:
		return p.join(f.Terms, " | ", "false")
	case *LogicNot:
		return "!(" + p.render(f.Body) + ")"
	case *Eq:
		return p.join([]Expr{f.T1, f.T2}, " = ", "true")
	case *LogicIff:
		return p.join([]Expr{f.T1, f.T2}, " <-> ", "true")
	case *LogicImplies:
		return p.join([]Expr{f.T1, f.T2}, " -> ", "true")
	case *LogicIte:
		return "(if " + p.render(f.Cond) + " then " + p.render(f.Then) + " else " + p.render(f.Else) + ")"
	case *ForAll:
		return p.quantifier("forall", f.Variables, f.Body)
	case *LogicExists:
		return p.quantifier("exists", f.Variables, f.Body)
	case *LogicDefinition:
		return p.join([]Expr{f.Lhs, f.Rhs}, " = ", "true")
	}
	if p.err == nil {
		p.err = fmt.Errorf("cannot translate %T to mypyvy", expr)
	}
	return ""
}

func (p *mypyvyForm) join(terms []Expr, op, empty string) string {
	if len(terms) == 0 {
		return empty
	}
	result := make([]string, len(terms))
	for i, term := range terms {
		result[i] = p.render(term)
	}
	return "(" + strings.Join(result, op) + ")"
}

func (p *mypyvyForm) quantifier(kind string, variables []*LogicVariable, body Expr) string {
	if len(variables) == 0 {
		return p.render(body)
	}
	args := make([]string, len(variables))
	for i, variable := range variables {
		args[i] = mypyvyName(variable.Name) + ":" + mypyvyName(IvySortName(variable.VSort))
	}
	return "(" + kind + " " + strings.Join(args, ", ") + ". " + p.render(body) + ")"
}

func mypyvyLocal(symbol *Const) bool {
	name := symbol.Name
	if IsNew(name) {
		name = NewOf(name)
	}
	for _, prefix := range []string{"fml:", "loc:", "prm:", "__", "old_", "@"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// MypyvyExport writes the selected isolate as a mypyvy model.
func MypyvyExport(mod *Module) error {
	leaveTheory := mod.TheoryContext()
	defer leaveTheory()
	var forms []Expr
	var actions []vmtTransition
	background := mod.BackgroundTheory(nil)
	axioms := append([]Expr(nil), background.Fmlas...)
	for _, def := range background.Defs {
		axioms = append(axioms, def)
	}
	var initializer []Expr
	for _, action := range mod.Initializers {
		initializer = append(initializer, action.Action)
	}
	initialUpdate := computeUpdate(mod, NewSequence(initializer...))
	initCond := mod.InitCond
	if initCond == nil {
		initCond = TrueClauses(EmptyAnnotation{})
	}
	initial := ForwardImage(initCond, background, initialUpdate).ToOpenFormula()
	forms = append(forms, initial)
	mutable := make(map[string]bool)
	for _, symbol := range initialUpdate.Modified {
		mutable[symbol.Name] = true
	}
	var exported []string
	for name := range mod.PublicActions.All() {
		exported = append(exported, name)
	}
	sort.Strings(exported)
	for _, name := range exported {
		action := mod.Actions.Get(name)
		if action == nil {
			continue
		}
		update := computeUpdate(mod, action)
		transition := vmtTransition{name: name, formula: update.TR.ToOpenFormula(), modified: update.Modified}
		for _, symbol := range transition.modified {
			mutable[symbol.Name] = true
		}
		actions = append(actions, transition)
		forms = append(forms, transition.formula)
	}
	for _, invariant := range mod.LabeledConjs {
		forms = append(forms, invariant.Formula.(Expr))
	}
	forms = append(forms, axioms...)
	var text bytes.Buffer
	for name, sort := range mod.Sig.Sorts.All() {
		if SortEqual(sort, Boolean) {
			continue
		}
		if _, uninterpreted := sort.(*UninterpretedSort); !uninterpreted || mod.Sig.Interp.Get(name) != nil {
			return fmt.Errorf("mypyvy requires uninterpreted sort %s", name)
		}
		fmt.Fprintf(&text, "sort %s\n", mypyvyName(name))
	}
	translation := mypyvyForm{numeric: make(map[NodeKey]string)}
	declared := make(map[string]bool)
	for _, symbol := range collectAllSymbols(forms) {
		if IsNew(symbol.Name) && !mypyvyLocal(symbol) {
			symbol = NewConst(NewOf(symbol.Name), symbol.CSort)
		}
		name := mypyvyName(symbol.Name)
		if IsNumeral(symbol) {
			name = "ivy_numeral_" + name + "_" + mypyvyName(IvySortName(symbol.CSort))
			translation.numeric[Key(symbol)] = name
		}
		if declared[name] || symbol.Name == "true" || symbol.Name == "false" || (mypyvyLocal(symbol) && !IsFunctionSort(symbol.CSort)) {
			continue
		}
		declared[name] = true
		kind := "immutable"
		if mutable[symbol.Name] || mypyvyLocal(symbol) {
			kind = "mutable"
		}
		if fs, ok := symbol.CSort.(*LogicFunctionSort); ok {
			var domain []string
			for _, sort := range fs.Domain() {
				domain = append(domain, mypyvyName(IvySortName(sort)))
			}
			if SortEqual(fs.Range(), Boolean) {
				fmt.Fprintf(&text, "%s relation %s(%s)\n", kind, name, strings.Join(domain, ", "))
			} else {
				fmt.Fprintf(&text, "%s function %s(%s): %s\n", kind, name, strings.Join(domain, ", "), mypyvyName(IvySortName(fs.Range())))
			}
		} else if SortEqual(symbol.CSort, Boolean) {
			fmt.Fprintf(&text, "%s relation %s\n", kind, name)
		} else {
			fmt.Fprintf(&text, "%s constant %s: %s\n", kind, name, mypyvyName(IvySortName(symbol.CSort)))
		}
	}
	for _, axiom := range axioms {
		fmt.Fprintf(&text, "axiom %s\n", translation.render(CloseFormula(axiom)))
	}
	initial = mypyvyBindInitializerLocals(initial)
	fmt.Fprintf(&text, "init %s\n", translation.render(CloseFormula(initial)))
	for _, action := range actions {
		var parameters, modifies []string
		for _, symbol := range collectAllSymbols([]Expr{action.formula}) {
			if mypyvyLocal(symbol) && !IsFunctionSort(symbol.CSort) {
				parameters = append(parameters, mypyvyName(symbol.Name)+":"+mypyvyName(IvySortName(symbol.CSort)))
			}
		}
		for _, symbol := range action.modified {
			modifies = append(modifies, mypyvyName(symbol.Name))
		}
		fmt.Fprintf(&text, "transition %s(%s)\n", mypyvyName(action.name), strings.Join(parameters, ", "))
		if len(modifies) > 0 {
			fmt.Fprintf(&text, "  modifies %s\n", strings.Join(modifies, ", "))
		}
		fmt.Fprintf(&text, "  %s\n", translation.render(CloseFormula(action.formula)))
	}
	for _, invariant := range mod.LabeledConjs {
		label := ""
		if invariant.Label != nil {
			label = " [" + mypyvyName(lfLabelName(invariant)) + "]"
		}
		fmt.Fprintf(&text, "invariant%s %s\n", label, translation.render(CloseFormula(invariant.Formula.(Expr))))
	}
	if translation.err != nil {
		return translation.err
	}
	path := mod.Name + ".pyv"
	if mod.Name == "" {
		path = "ivy.pyv"
	}
	if err := os.WriteFile(path, text.Bytes(), 0o644); err != nil {
		return err
	}
	fmt.Println("output written to " + path)
	return nil
}

func mypyvyBindInitializerLocals(initial Expr) Expr {
	var variables []*LogicVariable
	bindings := make(map[NodeKey]Expr)
	for _, symbol := range collectAllSymbols([]Expr{initial}) {
		if !mypyvyLocal(symbol) || IsFunctionSort(symbol.CSort) {
			continue
		}
		variable := &LogicVariable{Name: fmt.Sprintf("IVYINIT%d", len(variables)), VSort: symbol.CSort}
		variables = append(variables, variable)
		bindings[Key(symbol)] = variable
	}
	if len(variables) == 0 {
		return initial
	}
	return &LogicExists{Variables: variables, Body: SubstituteConstantsAST(initial, bindings).(Expr)}
}
