package goivy

import (
	"fmt"
	"sort"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// ToAigerResult holds the result of converting a module to an AIGER circuit.
type ToAigerResult struct {
	Aiger    *Encoder
	Decoder  map[string]Expr // abstract prop -> original expression
	Annot    interface{}     // annotation for trace reconstruction
	Consts   map[string]bool // set of constant symbols used in sort_constants
	Action   ActionsAction   // the composed action
	StVarSet map[string]bool // set of original state variable names
}

// ToAiger converts an Ivy module to an AIGER circuit for model checking.
// This is the main entry point for the model checking pipeline.
//
// Python: ivy_mc.py:1117-1427
func ToAiger(mod *Module, method string) (*ToAigerResult, error) {
	if method == "" {
		method = "mc"
	}

	// Step 0: Error flag instrumentation
	erf := NewConst("err_flag", Boolean)
	var errConds []Expr
	AddErrFlagMod(mod, erf, &errConds)

	// Step 1: Build the composed external action
	// Python: ext_acts = [mod.actions[x].add_label(x) for x in sorted(mod.public_actions)]
	//         ext_act = ia.EnvAction(*ext_acts)
	pubNames := sortedPublicActions(mod)
	extActs := make([]Expr, 0, len(pubNames))
	for _, name := range pubNames {
		act, ok := mod.Actions.Get2(name)
		if !ok {
			continue
		}
		if a, ok := act.(ActionsAction); ok {
			extActs = append(extActs, addLabelToAction(a, name))
		}
	}

	extAct := NewEnvActionOn(mod.Cfg.ActCfg, extActs...)

	initVar := NewConst("__init", Boolean)

	// Python: init = add_err_flag(Sequence(*([a for n,a in mod.initializers]+[AssignAction(init_var,And())])), erf, errconds)
	var initParts []Expr
	for _, ni := range mod.Initializers {
		if a, ok := ni.Action.(ActionsAction); ok {
			initParts = append(initParts, a)
		}
	}
	initParts = append(initParts, NewAssignAction(initVar, &And{Terms: nil}))
	initSeq := NewSequence(initParts...)
	initAction := AddErrFlag(initSeq, erf, &errConds, mod.Instantiator)

	// Python: action = Sequence(AssignAction(erf, Or()), IfAction(init_var, ext_act, init))
	erfReset := NewAssignAction(erf, &Or{Terms: nil})
	ifAct := NewIfAction(initVar, extAct, initAction)
	composedAction := NewSequence(erfReset, ifAct)

	// Step 2: Get invariant to prove, applying proof tactics.
	// Python: ivy_mc.py:1136-1158
	pc := NewProofChecker(mod.Cfg.ProofCfg, mod, mod.LabeledAxioms, mod.Definitions, moduleSchemataToLF(mod.Schemata), mod.Cfg.AstCfg)
	pmap := make(map[int64]Node)
	for _, pe := range mod.Proofs {
		pmap[pe.Formula.ID] = pe.Proof
	}
	var conjs []*LabeledFormula
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
	var invTerms []Expr
	for _, lf := range conjs {
		invTerms = append(invTerms, IvyDropUniversals(lf.Formula.(Expr)))
	}
	var invariant Expr = &And{Terms: invTerms}

	// Python: skolemizer = lambda v: ilu.var_to_skolem('__', il.Variable(v.rep, v.sort))
	//         vs = ilu.used_variables_in_order_ast(invariant)
	vs := UsedVariablesOrdered(NewClauses([]Expr{invariant}, nil, nil))
	sksubs := make(map[string]Expr, len(vs))
	for _, v := range vs {
		sksubs[v.Name] = VarToSkolem("__", v)
	}
	invariant = SubstituteAstByName(invariant, sksubs)
	invarSyms := UsedSymbolsAST(invariant)
	xtracer.Trace("mc.ToAiger postSkolemize HASH canon= invariant=%s", invariant.Sexp())

	// Step 3: Compute transition relation
	bgt := mod.BackgroundTheory(nil)
	ctx := &UpdateContext{
		Domain:          mod,
		ActCfg:          mod.Cfg.ActCfg,
		Instantiator:    mod.Instantiator,
		CheckUnprovable: mod.Cfg.OnlyCheckUnprovable,
		CheckedAssert:   mod.Cfg.CheckLineno,
	}
	xtracer.Trace("mc.ActionToTR calling GetUpdate type=%s", ActionTypeName(composedAction))
	upd := GetUpdate(composedAction, ctx)

	// Add post axioms: rename axioms for modified symbols
	updWithAxioms := AddPostAxioms(upd, bgt)

	// Build transition as clauses — merge update's TR with background defs
	trans := updWithAxioms.TR
	if bgt != nil && len(bgt.Defs) > 0 {
		trans = AndClausesTyped(trans, NewClauses(nil, bgt.Defs, nil))
	}

	// D7: Build defsyms from bgt.Defs (not trans.Defs) to match Python.
	// Python: defsyms = set(x.defines() for x in bgt.defs)
	defSymsByName := make(map[string]*Const)
	if bgt != nil {
		for _, d := range bgt.Defs {
			if c, ok := d.Defines().(*Const); ok {
				defSymsByName[c.Name] = c
			}
		}
	}
	// Python: rn = dict((tr.new(sym), tr.new(sym).prefix('__')) for sym in defsyms)
	rn := make(map[NodeKey]*Const)
	for name, sym := range defSymsByName {
		newName := ActionNewName(name)
		prefixed := "__" + newName
		newSym := NewConst(newName, sym.CSort)
		rn[Key(newSym)] = NewConst(prefixed, sym.CSort)
	}

	if len(rn) > 0 {
		trans = RenameClauses(trans, rn)
	}

	// D6: Rename error (Pre) clauses alongside trans.
	// Python: error = ilu.rename_clauses(error, rn)
	if len(rn) > 0 && updWithAxioms.Pre != nil {
		_ = RenameClauses(updWithAxioms.Pre, rn)
	}

	xtracer.Trace("mc.ToAiger postAddPostAxioms nStVars=%d nTRfmlas=%d nTRdefs=%d HASH canon= trans=%s", len(updWithAxioms.Modified), len(trans.Fmlas), len(trans.Defs), trans.Canon())

	// Python: stvars = [x for x in stvars if x not in defsyms]
	// Use []*lg.Const (matching Python's stvars [Symbol]) to preserve sort info.
	stVars := make([]*Const, 0, len(updWithAxioms.Modified))
	for _, sv := range updWithAxioms.Modified {
		if _, isDef := defSymsByName[sv.Name]; !isDef {
			stVars = append(stVars, sv)
		}
	}

	annot := trans.Annot

	// D3: Build inductive hypotheses using CloseFormula.
	// Python: indhyps = [il.close_formula(il.Implies(init_var, lf.formula))
	//                     for lf in mod.labeled_conjs + mod.assumed_invariants]
	var indHyps []Expr
	for _, lf := range mod.LabeledConjs {
		indHyps = append(indHyps, CloseFormula(
			&Implies{T1: initVar, T2: lf.Formula.(Expr)}))
	}
	for _, lf := range mod.AssumedInvs {
		indHyps = append(indHyps, CloseFormula(
			&Implies{T1: initVar, T2: lf.Formula.(Expr)}))
	}

	// Save original symbols for trace
	origSyms := make(map[string]bool)
	for _, sym := range UsedSymbolsAST(invariant).All() {
		origSyms[ExprName(sym)] = true
	}
	for _, sym := range SymbolsClauses(trans) {
		origSyms[ExprName(sym)] = true
	}

	// D4: Collect function symbols from def RHS only + fmlas + invariant.
	// Python: for df in trans.defs: funs.update(used_symbols_ast(df.args[1]))
	//         for fmla in trans.fmlas: funs.update(used_symbols_ast(fmla))
	//         funs.update(used_symbols_ast(invariant))
	//         funs = set(sym for sym in funs if is_function_sort(sym.sort))
	allSyms := NewInsMap[string, Expr]()
	for _, df := range trans.Defs {
		for _, sym := range UsedSymbolsAST(df.Rhs).All() {
			allSyms.Set(ExprName(sym), sym)
		}
	}
	for _, fmla := range trans.Fmlas {
		for _, sym := range UsedSymbolsAST(fmla).All() {
			allSyms.Set(ExprName(sym), sym)
		}
	}
	for _, sym := range UsedSymbolsAST(invariant).All() {
		allSyms.Set(ExprName(sym), sym)
	}
	funs := NewInsMap[string, *Const]()
	for name, sym := range allSyms.All() {
		if IsFunctionSort(sym.NodeSort()) {
			if c, ok := sym.(*Const); ok {
				funs.Set(name, c)
			}
		}
	}

	// ===== PROPOSITIONAL ABSTRACTION PIPELINE =====

	// Step 4a: Convert non-finite definitions to constraints
	var newDefs []*IvyDefinition
	var newFmlas []Expr
	for _, df := range trans.Defs {
		defSym := df.Defines()
		if c, ok := defSym.(*Const); ok {
			if len(df.Lhs.Children()) == 0 && isFiniteSort(c.CSort) {
				// Keep as definition (nullary, finite sort)
				newDefs = append(newDefs, df)
				continue
			}
		}
		// Convert to constraint
		newFmlas = append(newFmlas, mcDefToConstraint(df))
	}
	transFmlas2 := append(newFmlas, trans.Fmlas...)
	trans = NewClauses(transFmlas2, newDefs, trans.Annot)

	// Step 4b: Eliminate ITEs over non-finite sorts
	var iteCnsts []Expr
	mcIteCtr := &mod.Cfg.McIteCtr
	elimDefs := make([]*IvyDefinition, len(trans.Defs))
	for i, df := range trans.Defs {
		newLhs := ElimIte(df.Lhs, &iteCnsts, mcIteCtr)
		newRhs := ElimIte(df.Rhs, &iteCnsts, mcIteCtr)
		elimDefs[i] = NewIvyDefinition(newLhs, newRhs)
	}
	elimFmlas := make([]Expr, len(trans.Fmlas))
	for i, f := range trans.Fmlas {
		elimFmlas[i] = ElimIte(f, &iteCnsts, mcIteCtr)
	}
	allFmlas := append(elimFmlas, iteCnsts...)
	trans = NewClauses(allFmlas, elimDefs, trans.Annot)

	// Step 4c: Quantifier elimination via finite instantiation
	// D5: Build from_asserts and pass And(invariant, fromAsserts) to MineConstants.
	// Python: from_asserts = il.And(*[il.Equals(x,x) for x in
	//             ilu.used_symbols_ast(il.And(*errconds))
	//             if tr.is_skolem(x) and not il.is_function_sort(x.sort)])
	var fromAssertTerms []Expr
	errCondsConj := &And{Terms: errConds}
	for _, sym := range UsedSymbolsAST(errCondsConj).All() {
		if IsSkolem(ExprName(sym)) && !IsFunctionSort(sym.NodeSort()) {
			fromAssertTerms = append(fromAssertTerms, NewEquals(sym, sym))
		}
	}
	fromAsserts := &And{Terms: fromAssertTerms}

	// Python: invar_syms.update(ilu.used_symbols_ast(from_asserts))
	for k, sym := range UsedSymbolsAST(fromAsserts).All() {
		invarSyms.Set(k, sym)
	}

	// Python: sort_constants = mine_constants(mod, trans, il.And(invariant, from_asserts))
	mineTarget := &And{Terms: []Expr{invariant, fromAsserts}}
	sortConstants := MineConstants(mod, trans, mineTarget)
	sortConstants2 := MineConstants2(mod, trans, invariant)

	qelim := NewQelim(sortConstants, sortConstants2, mod.Cfg.IuCfg, mod.Cfg.FullQI)
	qeFmlas, qeDefs, newInvariant := qelim.Apply(trans.Fmlas, defsToNodes(trans.Defs), invariant, indHyps)
	invariant = newInvariant
	trans = NewClauses(qeFmlas, nodesToDefs(qeDefs), trans.Annot)
	xtracer.Trace("mc.ToAiger postQelim nStVars=%d nTRfmlas=%d nTRdefs=%d HASH canon= trans=%s", len(stVars), len(trans.Fmlas), len(trans.Defs), trans.Canon())

	// Step 4d: Instantiate axioms using pattern matching
	stVarNameList := mcConstNames(stVars)
	axs := InstantiateAxioms(mod, stVarNameList, trans, invariant, sortConstants, funs, mod.Cfg.IuCfg)
	if len(axs) > 0 {
		axConj := &And{Terms: axs}
		axVar := NewConst("__axioms", Boolean)
		axDef := NewIvyDefinition(axVar, axConj)
		invariant = &Implies{T1: axVar, T2: invariant}
		allFmlas := append(trans.Fmlas, axVar)
		allDefs := append(trans.Defs, axDef)
		trans = NewClauses(allFmlas, allDefs, trans.Annot)
	}
	xtracer.Trace("mc.ToAiger postAxiomInst nAxioms=%d nTRfmlas=%d nTRdefs=%d HASH canon= trans=%s", len(axs), len(trans.Fmlas), len(trans.Defs), trans.Canon())

	// Step 4e: Table lookup for finite-domain functions
	trans, invariant = ToTableLookup(trans, invariant)

	// Step 4f: Propositional abstraction
	stVarSet := make(map[string]bool, len(stVars))
	for _, sv := range stVars {
		stVarSet[sv.Name] = true
	}

	propAbs := NewPropAbs(stVarSet, sortConstants)
	paFmlas, paDefs := propAbs.Apply(trans.Fmlas, defsToNodes(trans.Defs))
	trans = NewClauses(paFmlas, nodesToDefs(paDefs), trans.Annot)

	// Find immutable abstract variables and give them next definitions
	invarSymSet := make(map[string]bool)
	for _, sym := range invarSyms.All() {
		invarSymSet[ExprName(sym)] = true
	}

	isImmutableExpr := func(expr Expr) bool {
		syms := UsedSymbolsAST(expr)
		for _, sym := range syms.All() {
			n := ExprName(sym)
			if IsSkolem(n) && !invarSymSet[n] {
				return false
			}
			if IsNew(n) || stVarSet[n] {
				return false
			}
		}
		return true
	}
	isExprDefined := func(expr Expr) bool {
		if app, ok := expr.(*Apply); ok {
			if c, ok := app.Func.(*Const); ok {
				_, isDef := defSymsByName[c.Name]
				return isDef
			}
		}
		return false
	}

	var addDefs []*IvyDefinition
	for exprKey, v := range propAbs.Map.All() {
		origExpr, _ := propAbs.OrigExprs.Get2(exprKey)
		if origExpr != nil && isImmutableExpr(origExpr) && !isExprDefined(origExpr) {
			propAbs.NewStVars = append(propAbs.NewStVars, v)
			addDefs = append(addDefs, NewIvyDefinition(
				NewConst(ActionNewName(v.Name), v.CSort),
				v,
			))
		}
	}
	for _, sym := range propAbs.FiniteSyms {
		if isImmutableExpr(sym) && !isExprDefined(sym) {
			propAbs.NewStVars = append(propAbs.NewStVars, sym)
			addDefs = append(addDefs, NewIvyDefinition(
				NewConst(ActionNewName(sym.Name), sym.CSort),
				sym,
			))
		}
	}
	if len(addDefs) > 0 {
		allDefs := append(trans.Defs, addDefs...)
		trans = NewClauses(trans.Fmlas, allDefs, trans.Annot)
	}

	xtracer.Trace("mc.ToAiger postPropAbs nStVars=%d nNewStVars=%d nTRfmlas=%d nTRdefs=%d HASH canon= trans=%s", len(stVars), len(propAbs.NewStVars), len(trans.Fmlas), len(trans.Defs), trans.Canon())

	// Apply propositional abstraction to invariant
	invariant = propAbs.MkPropAbs(invariant)

	// Create next-state symbols for atoms in the invariant
	rnInv := make(map[NodeKey]*Const, len(stVars))
	for _, sv := range stVars {
		rnInv[Key(sv)] = NewConst(ActionNewName(sv.Name), sv.CSort)
	}
	propAbs.MkPropAbs(RenameAST(invariant, rnInv))

	// Update state variables: filter to finite sorts + add new prop-abs vars
	// Python: stvars = [sym for sym in stvars if is_finite_sort(sym.sort)] + new_stvars
	var finiteStVars []*Const
	for _, sv := range stVars {
		if isFiniteSort(sv.CSort) {
			finiteStVars = append(finiteStVars, sv)
		}
	}
	for _, pv := range propAbs.NewStVars {
		finiteStVars = append(finiteStVars, pv)
	}
	stVars = finiteStVars
	xtracer.Trace("mc.ToAiger invariant HASH canon= %s", invariant.Sexp())

	// Step 5: State variable management
	// For each state var, create variables for latch inputs.
	// Also havoc all state bits except init flag at initial time.
	// Python: stvars_fix_map = dict((tr.new(v),fix(v)) for v in stvars)
	//         stvars_fix_map.update((v,curval(v)) for v in stvars if v != init_var)
	fixRn := make(map[NodeKey]*Const)
	for _, v := range stVars {
		newV := NewConst(ActionNewName(v.Name), v.CSort)
		fixRn[Key(newV)] = NewConst("nondet"+v.Name, v.CSort)
	}
	for _, v := range stVars {
		if v.Name != "__init" {
			fixRn[Key(v)] = NewConst("curval"+v.Name, v.CSort)
		}
	}
	trans = RenameClauses(trans, fixRn)

	// Add next-state definitions and curval definitions
	// Python: new_defs = trans.defs + [il.IvyDefinition(sym_inst(tr.new(v)),sym_inst(fix(v))) for v in stvars]
	var extraDefs []*IvyDefinition
	for _, v := range stVars {
		newV := NewConst(ActionNewName(v.Name), v.CSort)
		fixV := NewConst("nondet"+v.Name, v.CSort)
		extraDefs = append(extraDefs, NewIvyDefinition(newV, fixV))
	}
	for _, v := range stVars {
		if v.Name == "__init" {
			continue
		}
		curvalV := NewConst("curval"+v.Name, v.CSort)
		initChoice := NewConst("initchoice"+v.Name, v.CSort)
		extraDefs = append(extraDefs, NewIvyDefinition(curvalV,
			&Ite{Cond: initVar, Then: v, Else: initChoice}))
	}
	allDefs3 := append(trans.Defs, extraDefs...)
	trans = NewClauses(trans.Fmlas, allDefs3, trans.Annot)
	xtracer.Trace("mc.ToAiger postRename nStVars=%d nTRfmlas=%d nTRdefs=%d HASH canon= trans=%s", len(stVars), len(trans.Fmlas), len(trans.Defs), trans.Canon())

	// Step 6: Turn transition constraint into a definition
	cnstVar := NewConst("__cnst", Boolean)
	var finalDefs []*IvyDefinition
	finalDefs = append(finalDefs, trans.Defs...)
	fixCnst := NewConst("nondet__cnst", Boolean)
	finalDefs = append(finalDefs, NewIvyDefinition(
		NewConst(ActionNewName("__cnst"), Boolean),
		fixCnst,
	))
	// fix(cnst_var) = or(cnst_var, not(and(trans.fmlas)))
	var cnstBody Expr
	if len(trans.Fmlas) == 0 {
		cnstBody = cnstVar
	} else {
		fmlaConj := &And{Terms: trans.Fmlas}
		cnstBody = &Or{Terms: []Expr{cnstVar, &Not{Body: fmlaConj}}}
	}
	finalDefs = append(finalDefs, NewIvyDefinition(fixCnst, cnstBody))
	stVars = append(stVars, NewConst("__cnst", Boolean))
	trans = NewClauses(nil, finalDefs, trans.Annot)

	xtracer.Trace("mc.ToAiger finalTrans nStVars=%d nTRdefs=%d HASH canon= trans=%s", len(stVars), len(trans.Defs), trans.Canon())

	// Step 7: Determine inputs, outputs, build Encoder
	defSet := make(map[string]bool)
	for _, df := range trans.Defs {
		if c, ok := df.Defines().(*Const); ok {
			defSet[c.Name] = true
		}
	}
	for _, sv := range stVars {
		defSet[sv.Name] = true
	}

	usedSyms := trans.Symbols()
	for k, v := range UsedSymbolsAST(invariant).All() {
		usedSyms.Set(k, v)
	}

	var inputs []*Const
	for _, sym := range usedSyms.All() {
		name := ExprName(sym)
		cc, isConst := sym.(*Const)
		if !defSet[name] && !(isConst && isInterpretedSymbol(cc)) {
			if isConst {
				inputs = append(inputs, cc)
			}
		}
	}

	fail := NewConst("__fail", Boolean)
	outputs := []*Const{fail}

	aiger := NewEncoder(inputs, stVars, outputs)
	if mod.Sig != nil {
		aiger.Interp = mod.Sig.Interp
	}

	// Process combinational definitions (non-next-state)
	var combDefs []Expr
	for _, df := range trans.Defs {
		if c, ok := df.Defines().(*Const); ok {
			if !IsNew(c.Name) {
				combDefs = append(combDefs, df)
			}
		}
	}

	// Add invariant fail definition
	invarFail := NewConst("invar__fail", Boolean)
	combDefs = append(combDefs, NewIvyDefinition(invarFail, &Not{Body: invariant}))

	if err := aiger.DefList(combDefs); err != nil {
		return nil, fmt.Errorf("deflist failed: %w", err)
	}

	// Set next-state values for latches
	// Python: aiger.set(tr.new_of(df.defines()), aiger.eval(df.args[1]))
	for _, df := range trans.Defs {
		if c, ok := df.Defines().(*Const); ok {
			if IsNew(c.Name) {
				oldSym := NewConst(NewOf(c.Name), c.CSort)
				val, err := aiger.Eval(df.Rhs, nil)
				if err != nil {
					continue // skip errors in non-critical definitions
				}
				aiger.SetSym(oldSym, val)
			}
		}
	}

	// Set output: miter = and(init_var, not(cnst_var), or(invar__fail, and(fix(erf), not(fix(cnst_var)))))
	miter := &And{Terms: []Expr{
		initVar,
		&Not{Body: cnstVar},
		&Or{Terms: []Expr{
			invarFail,
			&And{Terms: []Expr{
				NewConst("nondet"+"err_flag", Boolean),
				&Not{Body: NewConst("nondet"+"__cnst", Boolean)},
			}},
		}},
	}}
	miterVal, err := aiger.Eval(miter, nil)
	if err != nil {
		return nil, fmt.Errorf("miter eval failed: %w", err)
	}
	aiger.SetSym(fail, miterVal)

	xtracer.Trace("mc.ToAiger aigerDone nInputs=%d nLatches=%d nOutputs=%d", len(aiger.Inputs), len(aiger.Latches), len(aiger.Outputs))

	// Build decoder
	decoder := make(map[string]Expr)
	for exprKey, v := range propAbs.Map.All() {
		_ = exprKey
		decoder[v.Name] = v
	}
	for _, sym := range aiger.Inputs {
		if origSyms[sym.Name] {
			decoder[sym.Name] = sym
		}
	}
	for _, sym := range aiger.Latches {
		if origSyms[sym.Name] {
			decoder[sym.Name] = sym
		}
	}

	// Collect all constants from sort_constants
	cnstSet := make(map[string]bool)
	for _, consts := range sortConstants.All() {
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

// constNames extracts string names from a slice of typed Const symbols.
func mcConstNames(syms []*Const) []string {
	names := make([]string, len(syms))
	for i, s := range syms {
		names[i] = s.Name
	}
	return names
}

// AddErrFlagMod instruments all actions in the module with error flag tracking.
// Asserts become assignments to the error flag, assumes become conditional on the error flag.
//
// Python: ivy_mc.py:1048-1054
func AddErrFlagMod(mod *Module, erf *Const, errConds *[]Expr) {
	for actname, act := range mod.Actions.All() {
		if a, ok := act.(ActionsAction); ok {
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
func AddErrFlag(action ActionsAction, erf *Const, errConds *[]Expr, instantiator func([]Expr) *Clauses) ActionsAction {
	switch a := action.(type) {
	case *AssertAction:
		// Python: errcond = ilu.dual_formula(il.drop_universals(action.formula))
		errCond := DualFormula(IvyDropUniversals(a.Formula), nil, instantiator)
		*errConds = append(*errConds, errCond)
		res := NewAssignAction(erf, &Or{Terms: []Expr{erf, errCond}})
		return res

	case *RequiresAction:
		// Require is a kind of assert
		errCond := &Not{Body: IvyDropUniversals(a.Formula)}
		*errConds = append(*errConds, errCond)
		res := NewAssignAction(erf, &Or{Terms: []Expr{erf, errCond}})
		return res

	case *SubgoalAction:
		// Skip subgoals
		return NewSequence()

	case *AssumeAction:
		// Assume: weaken to assume(or(erf, formula))
		res := NewAssumeAction(&Or{Terms: []Expr{erf, a.Formula}})
		return res

	case *Sequence:
		args := a.ActionArgs()
		newArgs := make([]Expr, len(args))
		for i, child := range args {
			if ca, ok := child.(ActionsAction); ok {
				newArgs[i] = AddErrFlag(ca, erf, errConds, instantiator)
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *ChoiceAction:
		args := a.ActionArgs()
		newArgs := make([]Expr, len(args))
		for i, child := range args {
			if ca, ok := child.(ActionsAction); ok {
				newArgs[i] = AddErrFlag(ca, erf, errConds, instantiator)
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *EnvAction:
		args := a.ActionArgs()
		newArgs := make([]Expr, len(args))
		for i, child := range args {
			if ca, ok := child.(ActionsAction); ok {
				newArgs[i] = AddErrFlag(ca, erf, errConds, instantiator)
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *BindOldsAction:
		args := a.ActionArgs()
		newArgs := make([]Expr, len(args))
		for i, child := range args {
			if ca, ok := child.(ActionsAction); ok {
				newArgs[i] = AddErrFlag(ca, erf, errConds, instantiator)
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *IfAction:
		args := a.ActionArgs()
		newArgs := make([]Expr, len(args))
		newArgs[0] = args[0] // condition unchanged
		for i := 1; i < len(args); i++ {
			if ca, ok := args[i].(ActionsAction); ok {
				newArgs[i] = AddErrFlag(ca, erf, errConds, instantiator)
			} else {
				newArgs[i] = args[i]
			}
		}
		return a.ActionClone(newArgs)

	case *LocalAction:
		args := a.ActionArgs()
		if len(args) == 0 {
			return action
		}
		newArgs := make([]Expr, len(args))
		copy(newArgs, args)
		// Last arg is the body
		last := args[len(args)-1]
		if ca, ok := last.(ActionsAction); ok {
			newArgs[len(newArgs)-1] = AddErrFlag(ca, erf, errConds, instantiator)
		}
		return a.ActionClone(newArgs)
	}

	return action
}

// moduleSchemataToLF converts Module.Schemata (*InsMap[string, ast.Node])
// to *InsMap[string, *ast.LabeledFormula] for NewProofChecker.
// Mirrors check.ModuleSchemataToAst (can't import check from mc).
func moduleSchemataToLF(schemata *InsMap[string, Node]) *InsMap[string, *LabeledFormula] {
	if schemata == nil {
		return nil
	}
	result := NewInsMap[string, *LabeledFormula]()
	for k, v := range schemata.All() {
		if s, ok := v.(*LabeledFormula); ok {
			result.Set(k, s)
		}
	}
	return result
}

// sortedPublicActions returns public action names sorted.
func sortedPublicActions(mod *Module) []string {
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
func addLabelToAction(a ActionsAction, label string) ActionsAction {
	type labelSetter interface {
		SetLabels([]string)
	}
	if ls, ok := a.(labelSetter); ok {
		ls.SetLabels([]string{label})
	}
	return a
}

// actionNodeWrapper wraps an actions.ActionsAction as a lg.Expr so it can be used
// in Args() slices. This is needed because the action types use []lg.Expr for children.
type actionNodeWrapper struct {
	Base
	action ActionsAction
}

func (w *actionNodeWrapper) NodeSort() Sort   { return Boolean }
func (w *actionNodeWrapper) Children() []Expr { return nil }
func (w *actionNodeWrapper) String() string {
	if w.action != nil {
		return w.action.String()
	}
	return "<nil-action>"
}
func (w *actionNodeWrapper) Equal(n Expr) bool { return w == n }
func (w *actionNodeWrapper) Sexp() NodeKey {
	return NodeKey("(actionNodeWrapper action:" + w.String() + ")")
}
func (w *actionNodeWrapper) Args() []Node           { return nil }
func (w *actionNodeWrapper) Clone(args []Node) Node { return w }

// defsToNodes converts a slice of *il.IvyDefinition to []lg.Expr.
func defsToNodes(defs []*IvyDefinition) []Expr {
	nodes := make([]Expr, len(defs))
	for i, d := range defs {
		nodes[i] = d
	}
	return nodes
}

// nodesToDefs converts a slice of lg.Expr back to []*il.IvyDefinition.
func nodesToDefs(nodes []Expr) []*IvyDefinition {
	var defs []*IvyDefinition
	for _, n := range nodes {
		if d, ok := n.(*IvyDefinition); ok {
			defs = append(defs, d)
		}
	}
	return defs
}

// isFiniteSortByName is a heuristic check for state variable names
// that are boolean (most abstract state vars are boolean).
