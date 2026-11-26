package goivy

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/smt"
)

type vmtTransition struct {
	name     string
	formula  Expr
	modified []*Const
}

func vmtMentionsState(formula Expr, state *InsMap[NodeKey, *Const]) bool {
	for key := range UsedSymbolsAst(formula).All() {
		if state.Get(key) != nil {
			return true
		}
	}
	return false
}

func addVMTImmutableSymbols(mod *Module, initial Expr, transitions []vmtTransition, state *InsMap[NodeKey, *Const], axioms []Expr) {
	background := make(map[NodeKey]bool)
	for _, symbol := range collectAllSymbols(axioms) {
		background[Key(symbol)] = true
	}
	for _, def := range mod.BackgroundTheory(nil).Defs {
		background[Key(def.Defines())] = true
	}
	formulas := []Expr{initial}
	for _, transition := range transitions {
		formulas = append(formulas, transition.formula)
	}
	for _, symbol := range collectAllSymbols(formulas) {
		key := Key(symbol)
		if state.Get(key) != nil || background[key] || IsFunctionSort(symbol.CSort) ||
			IsNumeral(symbol) || IsInterpretedSymbol(mod.Sig, symbol) || mod.Sig.Constructors[symbol.Name] ||
			IsNew(symbol.Name) || IsSkolem(symbol.Name) || strings.Contains(symbol.Name, ":") {
			continue
		}
		state.Set(key, symbol)
	}
}

func vmtFrame(symbol *Const) Expr {
	next := NewConst(ActionNewName(symbol.Name), symbol.CSort)
	fs, isFunction := symbol.CSort.(*LogicFunctionSort)
	if !isFunction {
		return &Eq{T1: symbol, T2: next}
	}
	vars := make([]*LogicVariable, fs.Arity())
	args := make([]Expr, fs.Arity())
	for i, dom := range fs.Domain() {
		vars[i] = &LogicVariable{Name: fmt.Sprintf("X%d", i), VSort: dom}
		args[i] = vars[i]
	}
	body := &Eq{T1: NewApplyUnchecked(symbol, args...), T2: NewApplyUnchecked(next, args...)}
	return &ForAll{Variables: vars, Body: body}
}

func vmtName(name string) string {
	return "|" + name + "|"
}

type vmtDeclarations struct {
	solver *Solver
	sorts  map[string]string
}

func (decls *vmtDeclarations) sortName(zsort smt.Z3Sort, sort Sort) string {
	switch zsort.Kind() {
	case smt.SortBool:
		return "Bool"
	case smt.SortInt:
		return "Int"
	case smt.SortReal:
		return "Real"
	case smt.SortBV:
		return fmt.Sprintf("(_ BitVec %d)", decls.solver.Context().BvSortSize(zsort))
	case smt.SortArray:
		domain := zsort.ArrayDomain()
		rng := zsort.ArrayRange()
		return fmt.Sprintf("(Array %s %s)", decls.sortName(domain, nil), decls.sortName(rng, nil))
	case smt.SortSeq:
		return zsort.String()
	}
	name := vmtName(zsort.String())
	if e, ok := sort.(*LogicEnumeratedSort); ok && zsort.Kind() != smt.SortUninterpreted {
		constructors := make([]string, len(e.Extension))
		for i, constructor := range e.Extension {
			constructors[i] = vmtName(constructor)
		}
		decls.sorts[name] = fmt.Sprintf("(declare-datatypes () ((%s %s)))", name, strings.Join(constructors, " "))
	} else if decls.sorts[name] == "" {
		decls.sorts[name] = fmt.Sprintf("(declare-sort %s 0)", name)
	}
	return name
}

func (decls *vmtDeclarations) signature(symbol *Const) (name string, domain []string, result string, err error) {
	translator := decls.solver.Translator()
	if fs, isFunction := symbol.CSort.(*LogicFunctionSort); isFunction {
		fd, err := translator.getFuncDecl(symbol)
		if err != nil {
			return "", nil, "", err
		}
		for i, sort := range fs.Domain() {
			domain = append(domain, decls.sortName(fd.DomainSort(i), sort))
		}
		return vmtName(fd.Name()), domain, decls.sortName(fd.RangeSort(), fs.Range()), nil
	}
	term, err := translator.Translate(symbol)
	if err != nil {
		return "", nil, "", err
	}
	return term.String(), nil, decls.sortName(term.ExprSort(), symbol.CSort), nil
}

func writeVMT(mod *Module, initial Expr, transitions []vmtTransition, state *InsMap[NodeKey, *Const], axioms []Expr, invariants []*LabeledFormula) error {
	solver := NewSolver(mod, mod.Cfg.SolverOpts)
	defer solver.Close()
	decls := vmtDeclarations{solver: solver, sorts: make(map[string]string)}
	formulas := append([]Expr{initial}, axioms...)
	for i, transition := range transitions {
		modified := make(map[NodeKey]bool)
		for _, symbol := range transition.modified {
			modified[Key(symbol)] = true
		}
		terms := []Expr{transition.formula}
		for key, symbol := range state.All() {
			if !modified[key] {
				terms = append(terms, vmtFrame(symbol))
			}
		}
		transitions[i].formula = &LogicAnd{Terms: terms}
		formulas = append(formulas, transitions[i].formula)
	}
	for _, invariant := range invariants {
		formulas = append(formulas, invariant.Formula.(Expr))
	}
	for _, formula := range formulas {
		for _, variable := range VariablesAST(formula) {
			zsort, err := solver.Translator().TranslateSort(variable.VSort)
			if err != nil {
				return err
			}
			decls.sortName(zsort, variable.VSort)
		}
	}
	symbols := NewInsMap[NodeKey, *Const]()
	for _, symbol := range collectAllSymbols(formulas) {
		symbols.Set(Key(symbol), symbol)
	}
	for _, symbol := range state.All() {
		symbols.Set(Key(symbol), symbol)
		prime := NewConst(ActionNewName(symbol.Name), symbol.CSort)
		symbols.Set(Key(prime), prime)
	}
	var body bytes.Buffer
	for _, symbol := range symbols.All() {
		if solver.SolverName(symbol) == "" || IsNumeral(symbol) || IsInterpretedSymbol(mod.Sig, symbol) || mod.Sig.Constructors[symbol.Name] {
			continue
		}
		name, domain, result, err := decls.signature(symbol)
		if err != nil {
			return fmt.Errorf("VMT symbol %s: %w", symbol.Name, err)
		}
		fmt.Fprintf(&body, "(declare-fun %s (%s) %s)\n", name, strings.Join(domain, " "), result)
	}
	index := 0
	for _, symbol := range state.All() {
		name, domain, result, err := decls.signature(symbol)
		if err != nil {
			return err
		}
		prime, _, _, err := decls.signature(NewConst(ActionNewName(symbol.Name), symbol.CSort))
		if err != nil {
			return err
		}
		var parameters, arguments []string
		for j, typ := range domain {
			arg := fmt.Sprintf("arg%d", j)
			parameters = append(parameters, fmt.Sprintf("(%s %s)", arg, typ))
			arguments = append(arguments, arg)
		}
		if len(arguments) > 0 {
			name = fmt.Sprintf("(%s %s)", name, strings.Join(arguments, " "))
		}
		fmt.Fprintf(&body, "(define-fun $state.%d (%s) %s (! %s :next %s))\n", index, strings.Join(parameters, " "), result, name, prime)
		index++
	}
	init, err := solver.FormulaToZ3(initial)
	if err != nil {
		return fmt.Errorf("serialize VMT initial state: %w", err)
	}
	fmt.Fprintf(&body, "(define-fun $init () Bool (! %s :init true))\n", init.String())
	var guarded []string
	var choices []string
	for i, transition := range transitions {
		choice := vmtName(fmt.Sprintf("ivy.action.%d.%s", i, CanonAct(transition.name)))
		fmt.Fprintf(&body, "(declare-fun %s () Bool)\n", choice)
		fmt.Fprintf(&body, "(define-fun $action.%d () Bool (! %s :action 0))\n", i, choice)
		formula, err := solver.FormulaToZ3(transition.formula)
		if err != nil {
			return fmt.Errorf("serialize VMT action %s: %w", transition.name, err)
		}
		guarded = append(guarded, fmt.Sprintf("(=> %s %s)", choice, formula.String()))
		for _, earlier := range choices {
			guarded = append(guarded, fmt.Sprintf("(not (and %s %s))", earlier, choice))
		}
		choices = append(choices, choice)
	}
	transition := "false"
	if len(choices) > 0 {
		transition = fmt.Sprintf("(and (or %s) %s)", strings.Join(choices, " "), strings.Join(guarded, " "))
	}
	fmt.Fprintf(&body, "(define-fun $trans () Bool (! %s :trans true))\n", transition)
	for i, invariant := range invariants {
		prop, err := solver.FormulaToZ3(invariant.Formula.(Expr))
		if err != nil {
			return fmt.Errorf("serialize VMT invariant %s: %w", lfLabelName(invariant), err)
		}
		fmt.Fprintf(&body, "(define-fun $property.%d () Bool (! %s :invar-property %d))\n", i, prop.String(), i)
	}
	for _, axiom := range axioms {
		formula, err := solver.FormulaToZ3(axiom)
		if err != nil {
			return fmt.Errorf("serialize VMT axiom: %w", err)
		}
		fmt.Fprintf(&body, "(assert %s)\n", formula.String())
	}
	var declarations []string
	for _, decl := range decls.sorts {
		declarations = append(declarations, decl)
	}
	sort.Strings(declarations)
	path := "ivy.vmt"
	if mod.Name != "" {
		path = mod.Name + ".vmt"
	}
	text := body.String()
	if len(declarations) > 0 {
		text = strings.Join(declarations, "\n") + "\n" + text
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return fmt.Errorf("write VMT: %w", err)
	}
	fmt.Println("output written to " + path)
	return nil
}
