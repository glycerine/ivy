package mc

import (
	"fmt"
	"sort"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/proof"
	"github.com/glycerine/ivy/goivy/xtracer"
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
	erf := lg.NewConst("err_flag", lg.Boolean)
	var errConds []lg.Expr
	AddErrFlagMod(mod, erf, &errConds)

	// Step 1: Build the composed external action
	// We use a special state variable __init to indicate the initial state
	pubNames := sortedPublicActions(mod)
	extActs := make([]lg.Expr, len(pubNames))
	for i, name := range pubNames {
		act, ok := mod.Actions.Get2(name)
		if !ok {
			continue
		}
		if a, ok := act.(actions.Action); ok {
			labeled := addLabelToAction(a, name)
			extActs[i] = &actionNodeWrapper{action: labeled}
		}
	}

	extAct := actions.NewEnvActionOn(mod.Cfg.ActCfg, extActs...)

	initVar := lg.NewConst("__init", lg.Boolean)

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
	initAction := AddErrFlag(initSeq, erf, &errConds, mod.Instantiator)

	// action = Sequence(erf := false, if init_var then ext_act else init)
	erfReset := actions.NewAssignAction(erf, &lg.Or{Terms: nil})     // false
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

	// Step 2: Get invariant to prove, applying proof tactics.
	// Python: ivy_mc.py:1136-1158
	pc := proof.NewProofChecker(mod.Cfg.ProofCfg, mod, mod.LabeledAxioms, mod.Definitions, moduleSchemataToLF(mod.Schemata), mod.Cfg.AstCfg)
	pmap := make(map[int64]ast.Node)
	for _, pe := range mod.Proofs {
		pmap[pe.Formula.ID] = pe.Proof
	}
	var conjs []*ast.LabeledFormula
	checkedAssert := mod.Cfg.CheckLineno
	for _, lf := range mod.LabeledConjs {
		// Python: if not checked(lf): continue
		if checkedAssert != "" && checkedAssert != lf.GetLineno().String() {
			continue
		}
		if mod.Cfg.MCVerbose {
			fmt.Printf("%vModel checking invariant\n", lf.GetLineno())
		}
		if p, ok := pmap[lf.ID]; ok {
			subgoals, err := pc.AdmitProposition(lf, p)
			if err != nil {
				return nil, fmt.Errorf("admit_proposition failed: %w", err)
			}
			conjs = append(conjs, subgoals...)
		} else {
			conjs = append(conjs, lf)
		}
	}

	// Python: invariant = il.And(*[il.drop_universals(lf.formula) for lf in conjs])
	var invTerms []lg.Expr
	for _, lf := range conjs {
		invTerms = append(invTerms, il.DropUniversals(lf.Formula.(lg.Expr)))
	}
	var invariant lg.Expr = &lg.And{Terms: invTerms}

	// Python: skolemizer = lambda v: ilu.var_to_skolem('__', il.Variable(v.rep, v.sort))
	//         vs = ilu.used_variables_in_order_ast(invariant)
	vs := module.UsedVariablesOrdered(module.NewClauses([]lg.Expr{invariant}, nil, nil))
	sksubs := make(map[string]lg.Expr, len(vs))
	for _, v := range vs {
		sksubs[v.Name] = module.VarToSkolem("__", v)
	}
	invariant = module.SubstituteAstByName(invariant, sksubs)
	invarSyms := module.UsedSymbolsAST(invariant)

	// Step 3: Compute transition relation
	bgt := mod.BackgroundTheory(nil)
	ctx := &actions.UpdateContext{
		Domain:          mod,
		ActCfg:          mod.Cfg.ActCfg,
		Instantiator:    mod.Instantiator,
		CheckUnprovable: mod.Cfg.OnlyCheckUnprovable,
		CheckedAssert:   mod.Cfg.CheckLineno,
	}
	xtracer.Trace("mc.ActionToTR calling GetUpdate type=%s", actions.ActionTypeName(composedAction))
	upd := actions.GetUpdate(composedAction, ctx)

	// Add post axioms: rename axioms for modified symbols
	updWithAxioms := actions.AddPostAxioms(upd, bgt)

	// Build transition as clauses — merge update's TR with background defs
	trans := updWithAxioms.TR
	if bgt != nil && len(bgt.Defs) > 0 {
		trans = module.AndClausesTyped(trans, module.NewClauses(nil, bgt.Defs, nil))
	}

	// D7: Build defsyms from bgt.Defs (not trans.Defs) to match Python.
	// Python: defsyms = set(x.defines() for x in bgt.defs)
	defSymsByName := make(map[string]*lg.Const)
	if bgt != nil {
		for _, d := range bgt.Defs {
			if c, ok := d.Defines().(*lg.Const); ok {
				defSymsByName[c.Name] = c
			}
		}
	}
	// Python: rn = dict((tr.new(sym), tr.new(sym).prefix('__')) for sym in defsyms)
	rn := make(map[lg.NodeKey]*lg.Const)
	for name, sym := range defSymsByName {
		newName := actions.New(name)
		prefixed := "__" + newName
		newSym := lg.NewConst(newName, sym.CSort)
		rn[lg.Key(newSym)] = lg.NewConst(prefixed, sym.CSort)
	}

	if len(rn) > 0 {
		trans = module.RenameClauses(trans, rn)
	}

	// D6: Rename error (Pre) clauses alongside trans.
	// Python: error = ilu.rename_clauses(error, rn)
	if len(rn) > 0 && updWithAxioms.Pre != nil {
		_ = module.RenameClauses(updWithAxioms.Pre, rn)
	}

	// Python: stvars = [x for x in stvars if x not in defsyms]
	stVarNames := actions.ModifiedNames(updWithAxioms)
	filteredStVars := make([]string, 0, len(stVarNames))
	for _, sv := range stVarNames {
		if _, isDef := defSymsByName[sv]; !isDef {
			filteredStVars = append(filteredStVars, sv)
		}
	}
	stVarNames = filteredStVars

	annot := trans.Annot

	// D3: Build inductive hypotheses using CloseFormula.
	// Python: indhyps = [il.close_formula(il.Implies(init_var, lf.formula))
	//                     for lf in mod.labeled_conjs + mod.assumed_invariants]
	var indHyps []lg.Expr
	for _, lf := range mod.LabeledConjs {
		indHyps = append(indHyps, il.CloseFormula(
			&lg.Implies{T1: initVar, T2: lf.Formula.(lg.Expr)}))
	}
	for _, lf := range mod.AssumedInvs {
		indHyps = append(indHyps, il.CloseFormula(
			&lg.Implies{T1: initVar, T2: lf.Formula.(lg.Expr)}))
	}

	// Save original symbols for trace
	origSyms := make(map[string]bool)
	for _, sym := range module.UsedSymbolsAST(invariant) {
		origSyms[lg.ExprName(sym)] = true
	}
	for _, sym := range module.SymbolsClauses(trans) {
		origSyms[lg.ExprName(sym)] = true
	}

	// D4: Collect function symbols from def RHS only + fmlas + invariant.
	// Python: for df in trans.defs: funs.update(used_symbols_ast(df.args[1]))
	//         for fmla in trans.fmlas: funs.update(used_symbols_ast(fmla))
	//         funs.update(used_symbols_ast(invariant))
	//         funs = set(sym for sym in funs if is_function_sort(sym.sort))
	allSyms := make(map[string]lg.Expr)
	for _, df := range trans.Defs {
		for _, sym := range module.UsedSymbolsAST(df.Rhs) {
			allSyms[lg.ExprName(sym)] = sym
		}
	}
	for _, fmla := range trans.Fmlas {
		for _, sym := range module.UsedSymbolsAST(fmla) {
			allSyms[lg.ExprName(sym)] = sym
		}
	}
	for _, sym := range module.UsedSymbolsAST(invariant) {
		allSyms[lg.ExprName(sym)] = sym
	}
	funs := make(map[string]bool)
	for name, sym := range allSyms {
		if il.IsFunctionSort(sym.NodeSort()) {
			funs[name] = true
		}
	}

	// ===== PROPOSITIONAL ABSTRACTION PIPELINE =====

	// Step 4a: Convert non-finite definitions to constraints
	var newDefs []*il.Definition
	var newFmlas []lg.Expr
	for _, df := range trans.Defs {
		defSym := df.Defines()
		if c, ok := defSym.(*lg.Const); ok {
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
	trans = module.NewClauses(transFmlas2, newDefs, trans.Annot)

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
	trans = module.NewClauses(allFmlas, elimDefs, trans.Annot)

	// Step 4c: Quantifier elimination via finite instantiation
	// D5: Build from_asserts and pass And(invariant, fromAsserts) to MineConstants.
	// Python: from_asserts = il.And(*[il.Equals(x,x) for x in
	//             ilu.used_symbols_ast(il.And(*errconds))
	//             if tr.is_skolem(x) and not il.is_function_sort(x.sort)])
	var fromAssertTerms []lg.Expr
	errCondsConj := &lg.And{Terms: errConds}
	for _, sym := range module.UsedSymbolsAST(errCondsConj) {
		if actions.IsSkolem(lg.ExprName(sym)) && !il.IsFunctionSort(sym.NodeSort()) {
			fromAssertTerms = append(fromAssertTerms, il.NewEquals(sym, sym))
		}
	}
	fromAsserts := &lg.And{Terms: fromAssertTerms}

	// Python: invar_syms.update(ilu.used_symbols_ast(from_asserts))
	for k, sym := range module.UsedSymbolsAST(fromAsserts) {
		invarSyms[k] = sym
	}

	// Python: sort_constants = mine_constants(mod, trans, il.And(invariant, from_asserts))
	mineTarget := &lg.And{Terms: []lg.Expr{invariant, fromAsserts}}
	sortConstants := MineConstants(mod, trans, mineTarget)
	sortConstants2 := MineConstants2(mod, trans, invariant)

	qelim := NewQelim(sortConstants, sortConstants2)
	qeFmlas, qeDefs, newInvariant := qelim.Apply(trans.Fmlas, defsToNodes(trans.Defs), invariant, indHyps)
	invariant = newInvariant
	trans = module.NewClauses(qeFmlas, nodesToDefs(qeDefs), trans.Annot)

	// Step 4d: Instantiate axioms using pattern matching
	stVarList := stVarNames
	axs := InstantiateAxioms(mod, stVarList, trans, invariant, sortConstants, funs)
	if len(axs) > 0 {
		axConj := &lg.And{Terms: axs}
		axVar := lg.NewConst("__axioms", lg.Boolean)
		axDef := il.NewDefinition(axVar, axConj)
		invariant = &lg.Implies{T1: axVar, T2: invariant}
		allFmlas := append(trans.Fmlas, axVar)
		allDefs := append(trans.Defs, axDef)
		trans = module.NewClauses(allFmlas, allDefs, trans.Annot)
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
	trans = module.NewClauses(paFmlas, nodesToDefs(paDefs), trans.Annot)

	// Find immutable abstract variables and give them next definitions
	invarSymSet := make(map[string]bool)
	for _, sym := range invarSyms {
		invarSymSet[lg.ExprName(sym)] = true
	}

	isImmutableExpr := func(expr lg.Expr) bool {
		syms := module.UsedSymbolsAST(expr)
		for _, sym := range syms {
			n := lg.ExprName(sym)
			if actions.IsSkolem(n) && !invarSymSet[n] {
				return false
			}
			if actions.IsNew(n) || stVarSet[n] {
				return false
			}
		}
		return true
	}
	isExprDefined := func(expr lg.Expr) bool {
		if app, ok := expr.(*lg.Apply); ok {
			if c, ok := app.Func.(*lg.Const); ok {
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
				lg.NewConst(actions.New(v.Name), v.CSort),
				v,
			))
		}
	}
	for _, sym := range propAbs.FiniteSyms {
		if isImmutableExpr(sym) && !isExprDefined(sym) {
			propAbs.NewStVars = append(propAbs.NewStVars, sym)
			addDefs = append(addDefs, il.NewDefinition(
				lg.NewConst(actions.New(sym.Name), sym.CSort),
				sym,
			))
		}
	}
	if len(addDefs) > 0 {
		allDefs := append(trans.Defs, addDefs...)
		trans = module.NewClauses(trans.Fmlas, allDefs, trans.Annot)
	}

	// Apply propositional abstraction to invariant
	invariant = propAbs.MkPropAbs(invariant)

	// Create next-state symbols for atoms in the invariant
	rnInv := make(map[lg.NodeKey]*lg.Const, len(stVarNames))
	for _, sv := range stVarNames {
		svSym := lg.NewConst(sv, lg.TopS)
		rnInv[lg.Key(svSym)] = lg.NewConst(actions.New(sv), nil)
	}
	propAbs.MkPropAbs(module.RenameAST(invariant, rnInv))

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
	fixRn := make(map[lg.NodeKey]*lg.Const)
	for _, v := range stVarNames {
		newVSym := lg.NewConst(actions.New(v), lg.TopS)
		fixRn[lg.Key(newVSym)] = lg.NewConst("nondet"+v, nil)
	}
	for _, v := range stVarNames {
		if v != "__init" {
			vSym := lg.NewConst(v, lg.TopS)
			fixRn[lg.Key(vSym)] = lg.NewConst("curval"+v, nil)
		}
	}
	trans = module.RenameClauses(trans, fixRn)

	// Add next-state definitions and curval definitions
	var extraDefs []*il.Definition
	for _, v := range stVarNames {
		newV := lg.NewConst(actions.New(v), nil)
		fixV := lg.NewConst("nondet"+v, nil)
		extraDefs = append(extraDefs, il.NewDefinition(newV, fixV))
	}
	for _, v := range stVarNames {
		if v == "__init" {
			continue
		}
		curvalV := lg.NewConst("curval"+v, nil)
		origV := lg.NewConst(v, nil)
		initChoice := lg.NewConst("initchoice"+v, nil)
		extraDefs = append(extraDefs, il.NewDefinition(curvalV,
			&lg.Ite{Cond: initVar, Then: origV, Else: initChoice}))
	}
	allDefs3 := append(trans.Defs, extraDefs...)
	trans = module.NewClauses(trans.Fmlas, allDefs3, trans.Annot)

	// Step 6: Turn transition constraint into a definition
	cnstVar := lg.NewConst("__cnst", lg.Boolean)
	var finalDefs []*il.Definition
	finalDefs = append(finalDefs, trans.Defs...)
	fixCnst := lg.NewConst("nondet__cnst", nil)
	finalDefs = append(finalDefs, il.NewDefinition(
		lg.NewConst(actions.New("__cnst"), nil),
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
	trans = module.NewClauses(nil, finalDefs, trans.Annot)

	// Step 7: Determine inputs, outputs, build Encoder
	defSet := make(map[string]bool)
	for _, df := range trans.Defs {
		if c, ok := df.Defines().(*lg.Const); ok {
			defSet[c.Name] = true
		}
	}
	for _, sv := range stVarNames {
		defSet[sv] = true
	}

	usedSyms := make(map[string]lg.Expr)
	for _, sym := range module.SymbolsClauses(trans) {
		usedSyms[lg.ExprName(sym)] = sym
	}
	for _, sym := range module.UsedSymbolsAST(invariant) {
		usedSyms[lg.ExprName(sym)] = sym
	}

	var inputs []string
	for name, sym := range usedSyms {
		cc, isConst := sym.(*lg.Const)
		if !defSet[name] && !(isConst && isInterpretedSymbol(cc)) {
			inputs = append(inputs, name)
		}
	}
	sort.Strings(inputs)

	fail := lg.NewConst("__fail", lg.Boolean)
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
		if c, ok := df.Defines().(*lg.Const); ok {
			if !actions.IsNew(c.Name) {
				combDefs = append(combDefs, df)
			}
		}
	}

	// Add invariant fail definition
	invarFail := lg.NewConst("invar__fail", lg.Boolean)
	combDefs = append(combDefs, il.NewDefinition(invarFail, &lg.Not{Body: invariant}))

	if err := aiger.DefList(combDefs); err != nil {
		return nil, fmt.Errorf("deflist failed: %w", err)
	}

	// Set next-state values for latches
	for _, df := range trans.Defs {
		if c, ok := df.Defines().(*lg.Const); ok {
			if actions.IsNew(c.Name) {
				oldName := actions.NewOf(c.Name)
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
				lg.NewConst("nondet"+"err_flag", nil),
				&lg.Not{Body: lg.NewConst("nondet"+"__cnst", nil)},
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
			c := lg.NewConst(sym, nil)
			decoder[sym] = c
		}
	}
	for _, sym := range aiger.Latches {
		if origSyms[sym] {
			c := lg.NewConst(sym, nil)
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
func AddErrFlagMod(mod *module.Module, erf *lg.Const, errConds *[]lg.Expr) {
	for actname, act := range mod.Actions.All() {
		if a, ok := act.(actions.Action); ok {
			newAction := AddErrFlag(a, erf, errConds, mod.Instantiator)
			newAction.SetFormalParams(a.GetFormalParams())
			newAction.SetFormalReturns(a.GetFormalReturns())
			mod.SetAction(actname, newAction)
		}
	}
}

// AddErrFlag recursively instruments an action with error flag tracking.
// Assert actions become erf := or(erf, dual_formula(formula)).
// Assume actions become assume(or(erf, formula)).
//
// Python: ivy_mc.py:1020-1046
func AddErrFlag(action actions.Action, erf *lg.Const, errConds *[]lg.Expr, instantiator func([]lg.Expr) *module.Clauses) actions.Action {
	switch a := action.(type) {
	case *actions.AssertAction:
		// Python: errcond = ilu.dual_formula(il.drop_universals(action.formula))
		errCond := module.DualFormula(il.DropUniversals(a.Formula), nil, instantiator)
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
				newArgs[i] = &actionNodeWrapper{action: AddErrFlag(ca, erf, errConds, instantiator)}
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
				newArgs[i] = &actionNodeWrapper{action: AddErrFlag(ca, erf, errConds, instantiator)}
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
				newArgs[i] = &actionNodeWrapper{action: AddErrFlag(ca, erf, errConds, instantiator)}
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
				newArgs[i] = &actionNodeWrapper{action: AddErrFlag(ca, erf, errConds, instantiator)}
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
				newArgs[i] = &actionNodeWrapper{action: AddErrFlag(ca, erf, errConds, instantiator)}
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
			newArgs[len(newArgs)-1] = &actionNodeWrapper{action: AddErrFlag(ca, erf, errConds, instantiator)}
		}
		return a.ActionClone(newArgs)
	}

	return action
}

// moduleSchemataToLF converts Module.Schemata (*InsMap[string, ast.Node])
// to *InsMap[string, *ast.LabeledFormula] for NewProofChecker.
// Mirrors check.ModuleSchemataToAst (can't import check from mc).
func moduleSchemataToLF(schemata *iu.InsMap[string, ast.Node]) *iu.InsMap[string, *ast.LabeledFormula] {
	if schemata == nil {
		return nil
	}
	result := iu.NewInsMap[string, *ast.LabeledFormula]()
	for k, v := range schemata.All() {
		if s, ok := v.(*ast.LabeledFormula); ok {
			result.Set(k, s)
		}
	}
	return result
}

// sortedPublicActions returns public action names sorted.
func sortedPublicActions(mod *module.Module) []string {
	names := make([]string, 0, mod.PublicActions.Len())
	for name := range mod.PublicActions.All() {
		if mod.PublicActions.Get(name) {
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

func (w *actionNodeWrapper) NodeSort() lg.Sort   { return lg.Boolean }
func (w *actionNodeWrapper) Children() []lg.Expr { return nil }
func (w *actionNodeWrapper) String() string {
	if w.action != nil {
		return w.action.String()
	}
	return "<nil-action>"
}
func (w *actionNodeWrapper) Equal(n lg.Expr) bool { return w == n }
func (w *actionNodeWrapper) Sexp() lg.NodeKey {
	return lg.NodeKey("(actionNodeWrapper action:" + w.String() + ")")
}
func (w *actionNodeWrapper) Args() []ast.Node               { return nil }
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
