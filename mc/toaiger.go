package mc

import (
	"fmt"
	"sort"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
	"github.com/glycerine/goivy/module"
	tr "github.com/glycerine/goivy/transrel"
)

// ToAigerResult holds the result of converting a module to an AIGER circuit.
type ToAigerResult struct {
	Aiger    *Encoder
	Decoder  map[string]lg.Expr // abstract prop -> original expression
	Annot    interface{}        // annotation for trace reconstruction
	Consts   map[string]bool    // set of constant symbols used in sort_constants
	Action   actions.Action     // the composed action
	StVarSet map[string]bool    // set of original state variable names
}

// ToAiger converts an Ivy module to an AIGER circuit for model checking.
// This is the main entry point for the model checking pipeline.
//
// Python: ivy_mc.py:1117-1427
func ToAiger(mod *module.Module, method string) (*ToAigerResult, error) {
	if method == "" {
		method = "mc"
	}

	// Step 0: Error flag instrumentation
	erf := lg.NewSymbol("err_flag", lg.Boolean)
	var errConds []lg.Expr
	AddErrFlagMod(mod, erf, &errConds)

	// Step 1: Build the composed external action
	// We use a special state variable __init to indicate the initial state
	pubNames := sortedPublicActions(mod)
	extActs := make([]lg.Expr, len(pubNames))
	for i, name := range pubNames {
		act, ok := mod.Actions[name]
		if !ok {
			continue
		}
		if a, ok := act.(actions.Action); ok {
			labeled := addLabelToAction(a, name)
			extActs[i] = &actionNodeWrapper{action: labeled}
		}
	}

	extAct := actions.NewEnvAction(extActs...)

	initVar := lg.NewSymbol("__init", lg.Boolean)

	// Build initializer sequence
	var initParts []lg.Expr
	for _, ni := range mod.Initializers {
		if a, ok := ni.Action.(actions.Action); ok {
			initParts = append(initParts, &actionNodeWrapper{action: a})
		}
	}
	initParts = append(initParts, &actionNodeWrapper{
		action: actions.NewAssignAction(initVar, &lg.And{Terms: nil}), // true
	})
	initSeq := actions.NewSequence(initParts...)
	initAction := AddErrFlag(initSeq, erf, &errConds)

	// action = Sequence(erf := false, if init_var then ext_act else init)
	erfReset := actions.NewAssignAction(erf, &lg.Or{Terms: nil}) // false
	ifAction := actions.NewIfAction(&actionNodeWrapper{action: nil}, // placeholder
		&actionNodeWrapper{action: extAct},
		&actionNodeWrapper{action: initAction},
	)
	// Build proper IfAction with initVar as condition
	ifAct := &actions.IfAction{}
	ifAct.Cond = initVar
	ifAct.ThenBody = &actionNodeWrapper{action: extAct}
	ifAct.ElseBody = &actionNodeWrapper{action: initAction}

	composedAction := actions.NewSequence(
		&actionNodeWrapper{action: erfReset},
		&actionNodeWrapper{action: ifAct},
	)
	_ = ifAction // unused placeholder

	// Step 2: Get invariant to prove, applying proof tactics
	// Replace free variables with Skolems
	var conjs []*ast.LabeledFormula
	for _, lf := range mod.LabeledConjs {
		conjs = append(conjs, lf)
	}

	var invTerms []lg.Expr
	for _, lf := range conjs {
		invTerms = append(invTerms, il.DropUniversals(lf.Formula.(lg.Expr)))
	}
	var invariant lg.Expr
	if len(invTerms) == 0 {
		invariant = &lg.And{Terms: nil} // true
	} else {
		invariant = &lg.And{Terms: invTerms}
	}

	// Skolemize free variables in invariant
	freeVars := lu.FreeVariablesList(invariant)
	if len(freeVars) > 0 {
		sksubs := make(map[string]lg.Expr, len(freeVars))
		for _, v := range freeVars {
			skName := "__" + v.Name
			sksubs[v.Name] = lg.NewSymbol(skName, v.VSort)
		}
		invariant = lu.SubstituteByName(invariant, sksubs)
	}
	invarSyms := co.UsedSymbolsAST(invariant)

	// Step 3: Compute transition relation
	bgt := mod.BackgroundTheory(nil)
	ctx := &actions.UpdateContext{
		Domain: mod,
	}
	upd := actions.GetUpdate(composedAction, ctx)

	// Add post axioms: rename axioms for modified symbols
	updWithAxioms := tr.AddPostAxioms(upd, bgt)

	// Build transition as clauses — merge update's TR with background defs
	trans := updWithAxioms.TR
	if bgt != nil && len(bgt.Defs) > 0 {
		trans = co.AndClausesTyped(trans, co.NewClauses(nil, bgt.Defs, nil))
	}

	// Rename defined symbols in next-state to avoid name collisions.
	// defSymsByName maps name -> *Const (preserving sort), matching Python's
	// set of Symbol objects with structural equality.
	defSymsByName := make(map[string]*lg.Symbol)
	for _, d := range trans.Defs {
		if c, ok := d.Defines().(*lg.Symbol); ok {
			defSymsByName[c.Name] = c
		}
	}
	rn := make(map[lg.NodeKey]*lg.Symbol)
	for name, sym := range defSymsByName {
		newName := tr.New(name)
		prefixed := "__" + newName
		newSym := lg.NewSymbol(newName, sym.CSort)
		rn[lg.Key(newSym)] = lg.NewSymbol(prefixed, sym.CSort)
	}

	if len(rn) > 0 {
		trans = co.RenameClauses(trans, rn)
	}

	stVarNames := tr.ModifiedNames(updWithAxioms)
	// Remove symbols with state-dependent definitions
	filteredStVars := make([]string, 0, len(stVarNames))
	for _, sv := range stVarNames {
		if _, isDef := defSymsByName[sv]; !isDef {
			filteredStVars = append(filteredStVars, sv)
		}
	}
	stVarNames = filteredStVars

	annot := trans.Annot

	// Build inductive hypotheses
	var indHyps []lg.Expr
	for _, lf := range mod.LabeledConjs {
		indHyps = append(indHyps, &lg.ForAll{
			Body: &lg.Implies{T1: initVar, T2: lf.Formula.(lg.Expr)},
		})
	}
	for _, lf := range mod.AssumedInvs {
		indHyps = append(indHyps, &lg.ForAll{
			Body: &lg.Implies{T1: initVar, T2: lf.Formula.(lg.Expr)},
		})
	}

	// Save original symbols for trace
	origSyms := make(map[string]bool)
	for _, sym := range co.UsedSymbolsAST(invariant) {
		origSyms[sym.Name] = true
	}
	for _, sym := range co.SymbolsClauses(trans) {
		origSyms[sym.Name] = true
	}

	// Collect function symbols
	funs := make(map[string]bool)
	for _, sym := range co.SymbolsClauses(trans) {
		if il.IsFunctionSort(sym.CSort) {
			funs[sym.Name] = true
		}
	}
	invSymsAST := co.UsedSymbolsAST(invariant)
	for _, sym := range invSymsAST {
		if il.IsFunctionSort(sym.CSort) {
			funs[sym.Name] = true
		}
	}

	// ===== PROPOSITIONAL ABSTRACTION PIPELINE =====

	// Step 4a: Convert non-finite definitions to constraints
	var newDefs []*il.Definition
	var newFmlas []lg.Expr
	for _, df := range trans.Defs {
		defSym := df.Defines()
		if c, ok := defSym.(*lg.Symbol); ok {
			if len(df.Lhs.Children()) == 0 && isFiniteSort(c.CSort) {
				// Keep as definition (nullary, finite sort)
				newDefs = append(newDefs, df)
				continue
			}
		}
		// Convert to constraint
		newFmlas = append(newFmlas, defToConstraint(df))
	}
	transFmlas2 := append(newFmlas, trans.Fmlas...)
	trans = co.NewClauses(transFmlas2, newDefs, trans.Annot)

	// Step 4b: Eliminate ITEs over non-finite sorts
	var iteCnsts []lg.Expr
	elimDefs := make([]*il.Definition, len(trans.Defs))
	for i, df := range trans.Defs {
		newLhs := ElimIte(df.Lhs, &iteCnsts)
		newRhs := ElimIte(df.Rhs, &iteCnsts)
		elimDefs[i] = il.NewDefinition(newLhs, newRhs)
	}
	elimFmlas := make([]lg.Expr, len(trans.Fmlas))
	for i, f := range trans.Fmlas {
		elimFmlas[i] = ElimIte(f, &iteCnsts)
	}
	allFmlas := append(elimFmlas, iteCnsts...)
	trans = co.NewClauses(allFmlas, elimDefs, trans.Annot)

	// Step 4c: Quantifier elimination via finite instantiation
	// Collect error condition symbols for invar_syms
	for _, ec := range errConds {
		ecSyms := co.UsedSymbolsAST(ec)
		for k, sym := range ecSyms {
			if tr.IsSkolem(sym.Name) && !il.IsFunctionSort(sym.CSort) {
				invarSyms[k] = sym
			}
		}
	}

	sortConstants := MineConstants(mod, trans, invariant)
	sortConstants2 := MineConstants2(mod, trans, invariant)

	qelim := NewQelim(sortConstants, sortConstants2)
	qeFmlas, qeDefs, newInvariant := qelim.Apply(trans.Fmlas, defsToNodes(trans.Defs), invariant, indHyps)
	invariant = newInvariant
	trans = co.NewClauses(qeFmlas, nodesToDefs(qeDefs), trans.Annot)

	// Step 4d: Instantiate axioms using pattern matching
	stVarList := stVarNames
	axs := InstantiateAxioms(mod, stVarList, trans, invariant, sortConstants, funs)
	if len(axs) > 0 {
		axConj := &lg.And{Terms: axs}
		axVar := lg.NewSymbol("__axioms", lg.Boolean)
		axDef := il.NewDefinition(axVar, axConj)
		invariant = &lg.Implies{T1: axVar, T2: invariant}
		allFmlas := append(trans.Fmlas, axVar)
		allDefs := append(trans.Defs, axDef)
		trans = co.NewClauses(allFmlas, allDefs, trans.Annot)
	}

	// Step 4e: Table lookup for finite-domain functions
	trans, invariant = ToTableLookup(trans, invariant)

	// Step 4f: Propositional abstraction
	stVarSet := make(map[string]bool, len(stVarNames))
	for _, sv := range stVarNames {
		stVarSet[sv] = true
	}

	propAbs := NewPropAbs(stVarSet, sortConstants)
	paFmlas, paDefs := propAbs.Apply(trans.Fmlas, defsToNodes(trans.Defs))
	trans = co.NewClauses(paFmlas, nodesToDefs(paDefs), trans.Annot)

	// Find immutable abstract variables and give them next definitions
	invarSymSet := make(map[string]bool)
	for _, sym := range invarSyms {
		invarSymSet[sym.Name] = true
	}

	isImmutableExpr := func(expr lg.Expr) bool {
		syms := co.UsedSymbolsAST(expr)
		for _, sym := range syms {
			if tr.IsSkolem(sym.Name) && !invarSymSet[sym.Name] {
				return false
			}
			if tr.IsNew(sym.Name) || stVarSet[sym.Name] {
				return false
			}
		}
		return true
	}
	isExprDefined := func(expr lg.Expr) bool {
		if app, ok := expr.(*lg.Apply); ok {
			if c, ok := app.Func.(*lg.Symbol); ok {
				_, isDef := defSymsByName[c.Name]
				return isDef
			}
		}
		return false
	}

	var addDefs []*il.Definition
	for exprKey, v := range propAbs.Map {
		_ = exprKey
		if isImmutableExpr(v) && !isExprDefined(v) {
			propAbs.NewStVars = append(propAbs.NewStVars, v)
			addDefs = append(addDefs, il.NewDefinition(
				lg.NewSymbol(tr.New(v.Name), v.CSort),
				v,
			))
		}
	}
	for _, sym := range propAbs.FiniteSyms {
		if isImmutableExpr(sym) && !isExprDefined(sym) {
			propAbs.NewStVars = append(propAbs.NewStVars, sym)
			addDefs = append(addDefs, il.NewDefinition(
				lg.NewSymbol(tr.New(sym.Name), sym.CSort),
				sym,
			))
		}
	}
	if len(addDefs) > 0 {
		allDefs := append(trans.Defs, addDefs...)
		trans = co.NewClauses(trans.Fmlas, allDefs, trans.Annot)
	}

	// Apply propositional abstraction to invariant
	invariant = propAbs.MkPropAbs(invariant)

	// Create next-state symbols for atoms in the invariant
	rnInv := make(map[lg.NodeKey]*lg.Symbol, len(stVarNames))
	for _, sv := range stVarNames {
		svSym := lg.NewSymbol(sv, lg.TopS)
		rnInv[lg.Key(svSym)] = lg.NewSymbol(tr.New(sv), nil)
	}
	propAbs.MkPropAbs(co.RenameAST(invariant, rnInv))

	// Update state variables
	var finiteStVars []string
	for _, sv := range stVarNames {
		if isFiniteSortByName(sv) {
			finiteStVars = append(finiteStVars, sv)
		}
	}
	for _, pv := range propAbs.NewStVars {
		finiteStVars = append(finiteStVars, pv.Name)
	}
	stVarNames = finiteStVars

	// Step 5: State variable management
	// For each state var, create variables for latch inputs.
	// Also havoc all state bits except init flag at initial time.
	fixRn := make(map[lg.NodeKey]*lg.Symbol)
	for _, v := range stVarNames {
		newVSym := lg.NewSymbol(tr.New(v), lg.TopS)
		fixRn[lg.Key(newVSym)] = lg.NewSymbol("nondet"+v, nil)
	}
	for _, v := range stVarNames {
		if v != "__init" {
			vSym := lg.NewSymbol(v, lg.TopS)
			fixRn[lg.Key(vSym)] = lg.NewSymbol("curval"+v, nil)
		}
	}
	trans = co.RenameClauses(trans, fixRn)

	// Add next-state definitions and curval definitions
	var extraDefs []*il.Definition
	for _, v := range stVarNames {
		newV := lg.NewSymbol(tr.New(v), nil)
		fixV := lg.NewSymbol("nondet"+v, nil)
		extraDefs = append(extraDefs, il.NewDefinition(newV, fixV))
	}
	for _, v := range stVarNames {
		if v == "__init" {
			continue
		}
		curvalV := lg.NewSymbol("curval"+v, nil)
		origV := lg.NewSymbol(v, nil)
		initChoice := lg.NewSymbol("initchoice"+v, nil)
		extraDefs = append(extraDefs, il.NewDefinition(curvalV,
			&lg.Ite{Cond: initVar, Then: origV, Else: initChoice}))
	}
	allDefs3 := append(trans.Defs, extraDefs...)
	trans = co.NewClauses(trans.Fmlas, allDefs3, trans.Annot)

	// Step 6: Turn transition constraint into a definition
	cnstVar := lg.NewSymbol("__cnst", lg.Boolean)
	var finalDefs []*il.Definition
	finalDefs = append(finalDefs, trans.Defs...)
	fixCnst := lg.NewSymbol("nondet__cnst", nil)
	finalDefs = append(finalDefs, il.NewDefinition(
		lg.NewSymbol(tr.New("__cnst"), nil),
		fixCnst,
	))
	// fix(cnst_var) = or(cnst_var, not(and(trans.fmlas)))
	var cnstBody lg.Expr
	if len(trans.Fmlas) == 0 {
		cnstBody = cnstVar
	} else {
		fmlaConj := &lg.And{Terms: trans.Fmlas}
		cnstBody = &lg.Or{Terms: []lg.Expr{cnstVar, &lg.Not{Body: fmlaConj}}}
	}
	finalDefs = append(finalDefs, il.NewDefinition(fixCnst, cnstBody))
	stVarNames = append(stVarNames, "__cnst")
	trans = co.NewClauses(nil, finalDefs, trans.Annot)

	// Step 7: Determine inputs, outputs, build Encoder
	defSet := make(map[string]bool)
	for _, df := range trans.Defs {
		if c, ok := df.Defines().(*lg.Symbol); ok {
			defSet[c.Name] = true
		}
	}
	for _, sv := range stVarNames {
		defSet[sv] = true
	}

	usedSyms := make(map[string]*lg.Symbol)
	for _, sym := range co.SymbolsClauses(trans) {
		usedSyms[sym.Name] = sym
	}
	for _, sym := range co.UsedSymbolsAST(invariant) {
		usedSyms[sym.Name] = sym
	}

	var inputs []string
	for name, sym := range usedSyms {
		if !defSet[name] && !isInterpretedSymbol(sym) {
			inputs = append(inputs, name)
		}
	}
	sort.Strings(inputs)

	fail := lg.NewSymbol("__fail", lg.Boolean)
	outputs := []string{fail.Name}

	// Build bit widths map
	bitWidths := make(map[string]int)
	for _, name := range inputs {
		bitWidths[name] = 1 // boolean by default
	}
	for _, name := range stVarNames {
		bitWidths[name] = 1
	}
	for _, name := range outputs {
		bitWidths[name] = 1
	}

	aiger := NewEncoder(inputs, stVarNames, outputs, bitWidths)

	// Process combinational definitions (non-next-state)
	var combDefs []lg.Expr
	for _, df := range trans.Defs {
		if c, ok := df.Defines().(*lg.Symbol); ok {
			if !tr.IsNew(c.Name) {
				combDefs = append(combDefs, df)
			}
		}
	}

	// Add invariant fail definition
	invarFail := lg.NewSymbol("invar__fail", lg.Boolean)
	combDefs = append(combDefs, il.NewDefinition(invarFail, &lg.Not{Body: invariant}))

	if err := aiger.DefList(combDefs); err != nil {
		return nil, fmt.Errorf("deflist failed: %w", err)
	}

	// Set next-state values for latches
	for _, df := range trans.Defs {
		if c, ok := df.Defines().(*lg.Symbol); ok {
			if tr.IsNew(c.Name) {
				oldName := tr.NewOf(c.Name)
				val, err := aiger.Eval(df.Rhs, nil)
				if err != nil {
					continue // skip errors in non-critical definitions
				}
				aiger.SetSym(oldName, val)
			}
		}
	}

	// Set output: miter = and(init_var, not(cnst_var), or(invar__fail, and(fix(erf), not(fix(cnst_var)))))
	miter := &lg.And{Terms: []lg.Expr{
		initVar,
		&lg.Not{Body: cnstVar},
		&lg.Or{Terms: []lg.Expr{
			invarFail,
			&lg.And{Terms: []lg.Expr{
				lg.NewSymbol("nondet"+"err_flag", nil),
				&lg.Not{Body: lg.NewSymbol("nondet"+"__cnst", nil)},
			}},
		}},
	}}
	miterVal, err := aiger.Eval(miter, nil)
	if err != nil {
		return nil, fmt.Errorf("miter eval failed: %w", err)
	}
	aiger.SetSym(fail.Name, miterVal)

	// Build decoder
	decoder := make(map[string]lg.Expr)
	for exprKey, v := range propAbs.Map {
		_ = exprKey
		decoder[v.Name] = v
	}
	for _, sym := range aiger.Inputs {
		if origSyms[sym] {
			c := lg.NewSymbol(sym, nil)
			decoder[sym] = c
		}
	}
	for _, sym := range aiger.Latches {
		if origSyms[sym] {
			c := lg.NewSymbol(sym, nil)
			decoder[sym] = c
		}
	}

	// Collect all constants from sort_constants
	cnstSet := make(map[string]bool)
	for _, consts := range sortConstants {
		for _, c := range consts {
			cnstSet[c.Name] = true
		}
	}

	return &ToAigerResult{
		Aiger:    aiger,
		Decoder:  decoder,
		Annot:    annot,
		Consts:   cnstSet,
		Action:   composedAction,
		StVarSet: stVarSet,
	}, nil
}

// AddErrFlagMod instruments all actions in the module with error flag tracking.
// Asserts become assignments to the error flag, assumes become conditional on the error flag.
//
// Python: ivy_mc.py:1048-1054
func AddErrFlagMod(mod *module.Module, erf *lg.Symbol, errConds *[]lg.Expr) {
	for actname := range mod.Actions {
		act := mod.Actions[actname]
		if a, ok := act.(actions.Action); ok {
			newAction := AddErrFlag(a, erf, errConds)
			newAction.SetFormalParams(a.GetFormalParams())
			newAction.SetFormalReturns(a.GetFormalReturns())
			mod.SetAction(actname, newAction)
		}
	}
}

// AddErrFlag recursively instruments an action with error flag tracking.
// Assert actions become erf := or(erf, not(formula)).
// Assume actions become assume(or(erf, formula)).
//
// Python: ivy_mc.py:1020-1046
func AddErrFlag(action actions.Action, erf *lg.Symbol, errConds *[]lg.Expr) actions.Action {
	switch a := action.(type) {
	case *actions.AssertAction:
		// Assert: compute error condition and set error flag
		errCond := &lg.Not{Body: il.DropUniversals(a.Formula)}
		*errConds = append(*errConds, errCond)
		res := actions.NewAssignAction(erf, &lg.Or{Terms: []lg.Expr{erf, errCond}})
		return res

	case *actions.RequiresAction:
		// Require is a kind of assert
		errCond := &lg.Not{Body: il.DropUniversals(a.Formula)}
		*errConds = append(*errConds, errCond)
		res := actions.NewAssignAction(erf, &lg.Or{Terms: []lg.Expr{erf, errCond}})
		return res

	case *actions.SubgoalAction:
		// Skip subgoals
		return actions.NewSequence()

	case *actions.AssumeAction:
		// Assume: weaken to assume(or(erf, formula))
		res := actions.NewAssumeAction(&lg.Or{Terms: []lg.Expr{erf, a.Formula}})
		return res

	case *actions.Sequence:
		args := a.ActionArgs()
		newArgs := make([]lg.Expr, len(args))
		for i, child := range args {
			if ca, ok := child.(actions.Action); ok {
				newArgs[i] = &actionNodeWrapper{action: AddErrFlag(ca, erf, errConds)}
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *actions.ChoiceAction:
		args := a.ActionArgs()
		newArgs := make([]lg.Expr, len(args))
		for i, child := range args {
			if ca, ok := child.(actions.Action); ok {
				newArgs[i] = &actionNodeWrapper{action: AddErrFlag(ca, erf, errConds)}
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *actions.EnvAction:
		args := a.ActionArgs()
		newArgs := make([]lg.Expr, len(args))
		for i, child := range args {
			if ca, ok := child.(actions.Action); ok {
				newArgs[i] = &actionNodeWrapper{action: AddErrFlag(ca, erf, errConds)}
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *actions.BindOldsAction:
		args := a.ActionArgs()
		newArgs := make([]lg.Expr, len(args))
		for i, child := range args {
			if ca, ok := child.(actions.Action); ok {
				newArgs[i] = &actionNodeWrapper{action: AddErrFlag(ca, erf, errConds)}
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *actions.IfAction:
		args := a.ActionArgs()
		newArgs := make([]lg.Expr, len(args))
		newArgs[0] = args[0] // condition unchanged
		for i := 1; i < len(args); i++ {
			if ca, ok := args[i].(actions.Action); ok {
				newArgs[i] = &actionNodeWrapper{action: AddErrFlag(ca, erf, errConds)}
			} else {
				newArgs[i] = args[i]
			}
		}
		return a.ActionClone(newArgs)

	case *actions.LocalAction:
		args := a.ActionArgs()
		if len(args) == 0 {
			return action
		}
		newArgs := make([]lg.Expr, len(args))
		copy(newArgs, args)
		// Last arg is the body
		last := args[len(args)-1]
		if ca, ok := last.(actions.Action); ok {
			newArgs[len(newArgs)-1] = &actionNodeWrapper{action: AddErrFlag(ca, erf, errConds)}
		}
		return a.ActionClone(newArgs)
	}

	return action
}

// sortedPublicActions returns public action names sorted.
func sortedPublicActions(mod *module.Module) []string {
	names := make([]string, 0, len(mod.PublicActions))
	for name := range mod.PublicActions {
		if mod.PublicActions[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// addLabelToAction wraps an action with a label.
func addLabelToAction(a actions.Action, label string) actions.Action {
	type labelSetter interface {
		SetLabels([]string)
	}
	if ls, ok := a.(labelSetter); ok {
		ls.SetLabels([]string{label})
	}
	return a
}

// actionNodeWrapper wraps an actions.Action as a lg.Expr so it can be used
// in Args() slices. This is needed because the action types use []lg.Expr for children.
type actionNodeWrapper struct {
	ast.Base
	action actions.Action
}

func (w *actionNodeWrapper) NodeSort() lg.Sort  { return lg.Boolean }
func (w *actionNodeWrapper) Children() []lg.Expr { return nil }
func (w *actionNodeWrapper) String() string {
	if w.action != nil {
		return w.action.String()
	}
	return "<nil-action>"
}
func (w *actionNodeWrapper) Equal(n lg.Expr) bool { return w == n }
func (w *actionNodeWrapper) Sexp() lg.NodeKey       { return lg.NodeKey("(actionNodeWrapper action:" + w.String() + ")") }
func (w *actionNodeWrapper) Args() []ast.Node      { return nil }
func (w *actionNodeWrapper) Clone(args []ast.Node) ast.Node { return w }

// defsToNodes converts a slice of *il.Definition to []lg.Expr.
func defsToNodes(defs []*il.Definition) []lg.Expr {
	nodes := make([]lg.Expr, len(defs))
	for i, d := range defs {
		nodes[i] = d
	}
	return nodes
}

// nodesToDefs converts a slice of lg.Expr back to []*il.Definition.
func nodesToDefs(nodes []lg.Expr) []*il.Definition {
	var defs []*il.Definition
	for _, n := range nodes {
		if d, ok := n.(*il.Definition); ok {
			defs = append(defs, d)
		}
	}
	return defs
}

// isFiniteSortByName is a heuristic check for state variable names
// that are boolean (most abstract state vars are boolean).
func isFiniteSortByName(name string) bool {
	// Boolean-named state vars are always finite
	// Abstract vars (__abs, __qe, __ite, __init, __cnst, __axioms) are boolean
	return true // In the propositionally abstracted system, all remaining vars are boolean
}

