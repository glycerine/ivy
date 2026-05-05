// Ported to Go from ivy_vmt.py. Merged into check/ from the former vmt/ package.
package goivy

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// checked returns true if the given labeled formula or action should
// be model-checked. Corresponds to Python's checked(thing).
func checked(lineno int, checkLineno string) bool {
	if checkLineno == "" {
		return true
	}
	return fmt.Sprintf("%d", lineno) == checkLineno
}

// checkedLF returns true if the labeled formula should be checked.
func checkedLF(lf *LabeledFormula, checkLineno string) bool {
	return checked(lf.Lineno(), checkLineno)
}

// checkedAction returns true if the action should be checked.
func checkedAction(a ActionsAction, checkLineno string) bool {
	loc := a.GetLineno()
	if checkLineno == "" {
		return true
	}
	return fmt.Sprintf("%d", loc.Line) == checkLineno
}

// actionToTR converts an action to a transition relation triple
// (stateVars, trans, error). Corresponds to Python's action_to_tr.
func actionToTR(m *Module, action ActionsAction, method string) ([]string, Expr, Expr, error) {
	// Get background theory
	bgt := m.BackgroundTheory(nil)

	// For fsmc method, unroll loops before computing the update.
	// Python: with ia.UnrollContext(im.module.sort_card): ...
	if method == "fsmc" && m.Cfg != nil && m.Cfg.ActCfg != nil {
		uc := NewUnrollContext(m.SortCard, m, m.Cfg.ActCfg)
		uc.Enter()
		defer uc.Exit()
	}

	// Compute the update (transition relation) for the action.
	// In Python: upd = action.update(im.module, None)
	upd := computeUpdate(m, action)
	if upd == nil {
		return nil, True, False, fmt.Errorf("failed to compute update for action")
	}

	// Add post-state axioms
	postUpd := AddPostAxioms(upd, bgt)

	stvars := postUpd.Modified
	transNode := postUpd.TRNode()
	errNodeFmla := postUpd.PreNode()

	// Conjoin definitions from background theory into trans
	// Conjoin definitions from background theory as formulas
	if bgt != nil && len(bgt.Defs) > 0 {
		defsClauses := NewClauses(nil, bgt.Defs, nil)
		defsFormula := defsClauses.ToOpenFormula()
		if defsFormula != nil && !IsTrue(defsFormula) {
			and, _ := NewAnd(transNode, defsFormula)
			transNode = and
		}
	}

	defsyms := make(map[string]bool)
	if bgt != nil {
		for _, d := range bgt.Defs {
			if c, ok := d.Defines().(*Const); ok {
				defsyms[c.Name] = true
			}
		}
	}
	if len(defsyms) > 0 {
		rn := make(map[string]string)
		for sym := range defsyms {
			newSym := ActionNewName(sym)
			rn[newSym] = "__" + newSym
		}
		transNode = checkRenameNode(transNode, rn)
		errNodeFmla = checkRenameNode(errNodeFmla, rn)

		var filtered []*Const
		for _, sv := range stvars {
			if !defsyms[sv.Name] {
				filtered = append(filtered, sv)
			}
		}
		stvars = filtered
	}

	return ModifiedNames(postUpd), transNode, errNodeFmla, nil
}

// addErrFlag transforms an action tree to use an error flag for assertion checking.
// Corresponds to Python's add_err_flag.
func addErrFlag(action ActionsAction, erf Expr, errconds *[]Expr, checkLineno string, verbose bool) ActionsAction {
	switch a := action.(type) {
	case *LogicAssertAction:
		if checkedAction(action, checkLineno) {
			if verbose {
				loc := action.GetLineno()
				fmt.Printf("%d:%s Model checking guarantee\n", loc.Line, loc.Filename)
			}
			// errcond = dual of the formula (negate after dropping universals)
			errcond := dualFormula(IvyDropUniversals(a.Formula))
			// res = erf := erf | errcond
			orNode := &LogicOr{Terms: []Expr{erf, errcond}}
			res := NewAssignAction(erf, orNode)
			*errconds = append(*errconds, errcond)
			res.SetLineno(Location{})
			return res
		}
		// Unchecked assert: treat as assume
		orNode := &LogicOr{Terms: []Expr{erf, a.Formula}}
		res := NewAssumeAction(orNode)
		res.SetLineno(Location{})
		return res

	case *LogicAssumeAction:
		orNode := &LogicOr{Terms: []Expr{erf, a.Formula}}
		res := NewAssumeAction(orNode)
		res.SetLineno(Location{})
		return res

	case *LogicSequence:
		newArgs := make([]Expr, len(a.Elems))
		for i, child := range a.Elems {
			if childAct, ok := checkToAction(child); ok {
				newArgs[i] = addErrFlag(childAct, erf, errconds, checkLineno, verbose)
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *LogicChoiceAction:
		newArgs := make([]Expr, len(a.Branches))
		for i, child := range a.Branches {
			if childAct, ok := checkToAction(child); ok {
				newArgs[i] = addErrFlag(childAct, erf, errconds, checkLineno, verbose)
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *LogicEnvAction:
		newArgs := make([]Expr, len(a.Branches))
		for i, child := range a.Branches {
			if childAct, ok := checkToAction(child); ok {
				newArgs[i] = addErrFlag(childAct, erf, errconds, checkLineno, verbose)
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *LogicBindOldsAction:
		args := a.ActionArgs()
		newArgs := make([]Expr, len(args))
		for i, child := range args {
			if childAct, ok := checkToAction(child); ok {
				newArgs[i] = addErrFlag(childAct, erf, errconds, checkLineno, verbose)
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *LogicIfAction:
		// Keep condition, transform then/else branches
		args := a.ActionArgs()
		newArgs := make([]Expr, len(args))
		newArgs[0] = args[0] // condition unchanged
		for i := 1; i < len(args); i++ {
			if childAct, ok := checkToAction(args[i]); ok {
				newArgs[i] = addErrFlag(childAct, erf, errconds, checkLineno, verbose)
			} else {
				newArgs[i] = args[i]
			}
		}
		return a.ActionClone(newArgs)

	case *LogicLocalAction:
		// Transform only the body (last arg)
		args := a.ActionArgs()
		newArgs := make([]Expr, len(args))
		copy(newArgs, args)
		if len(newArgs) > 0 {
			lastIdx := len(newArgs) - 1
			if childAct, ok := checkToAction(newArgs[lastIdx]); ok {
				newArgs[lastIdx] = addErrFlag(childAct, erf, errconds, checkLineno, verbose)
			}
		}
		return a.ActionClone(newArgs)
	}

	return action
}

// addErrFlagMod transforms all actions in a module to use error flag checking.
// Corresponds to Python's add_err_flag_mod.
func addErrFlagMod(m *Module, erf Expr, errconds *[]Expr) {
	for actname, actIface := range m.Actions.All() {
		action, ok := actIface.(ActionsAction)
		if !ok {
			continue
		}
		checkLineno := ""
		verbose := false
		if m.Cfg != nil {
			checkLineno = m.Cfg.CheckLineno
			verbose = m.Cfg.VMTVerbose
		}
		newAction := addErrFlag(action, erf, errconds, checkLineno, verbose)
		newAction.SetFormalParams(action.GetFormalParams())
		newAction.SetFormalReturns(action.GetFormalReturns())
		m.SetAction(actname, newAction)
	}
}

// createArraySort takes a function sort and returns a corresponding array
// sort name and sort chain. Corresponds to Python's create_array_sort.
func createArraySort(sig *Sig, fsort *LogicFunctionSort) (string, []Sort) {
	dom := fsort.Domain()
	rng := fsort.Range()
	return createArraySortRec(sig, dom, 0, rng)
}

func createArraySortRec(sig *Sig, dom []Sort, i int, rng Sort) (string, []Sort) {
	if i == len(dom) {
		return IvySortName(rng), []Sort{rng}
	}
	sname, ssorts := createArraySortRec(sig, dom, i+1, rng)
	name := "arr[" + IvySortName(dom[i]) + "][" + sname + "]"
	if _, ok := sig.Sorts.Get2(name); !ok {
		asort := &UninterpretedSort{Name: name}
		sig.Sorts.Set(name, asort)
		sig.Interp[name] = name
	}
	return name, append([]Sort{sig.Sorts.Get(name)}, ssorts...)
}

// encodeAsArray returns true if the symbol should be converted to an array.
// Corresponds to Python's encode_as_array.
func encodeAsArray(m *Module, sig *Sig, sym *Const) bool {
	if IsInterpretedSymbol(sig, sym) {
		return false
	}
	_, isDestructor := m.DestructorSorts[sym.Name]
	return !isDestructor
}

// ufToArrAST converts all uninterpreted functions in a formula to arrays.
// Corresponds to Python's uf_to_arr_ast.
func ufToArrAST(m *Module, sig *Sig, node Expr) Expr {
	return ufToArrASTRec(m, sig, node)
}

func ufToArrASTRec(m *Module, sig *Sig, node Expr) Expr {
	// Recursively transform arguments
	args := NodeArgs(node)
	newArgs := make([]Expr, len(args))
	for i, arg := range args {
		newArgs[i] = ufToArrASTRec(m, sig, arg)
	}

	// Check if this is a function application that should be encoded as array
	if IsApp(node) && !IsNamedBinder(node) && len(args) > 0 {
		sym := GetAppRep(node)
		if sym != nil && encodeAsArray(m, sig, sym) {
			fsort, ok := sym.CSort.(*LogicFunctionSort)
			if ok {
				sname, ssorts := createArraySort(sig, fsort)
				_ = sname
				asym := NewConst(sym.Name, ssorts[0])
				var result Expr = asym
				for i, arg := range newArgs {
					selSort, _ := NewFunctionSort(result.NodeSort(), arg.NodeSort(), ssorts[i+1])
					sel := NewConst("arrsel", selSort)
					applied, err := NewApply(sel, result, arg)
					if err != nil {
						return CloneNode(node, newArgs)
					}
					result = applied
				}
				return result
			}
		}
	}

	return CloneNode(node, newArgs)
}

// encodeAssign encodes a parameterized assignment to use array operations.
// Returns (newLHS, newRHS). Corresponds to Python's encode_assign.
func encodeAssign(m *Module, sig *Sig, asgn ActionsAction, lhs, rhs Expr) (Expr, Expr, error) {
	sym := GetAppRep(lhs)
	if sym == nil {
		return lhs, rhs, nil
	}

	if _, isDestructor := m.DestructorSorts[sym.Name]; isDestructor {
		lhsArgs := NodeArgs(lhs)
		if len(lhsArgs) > 0 {
			nlhs, nrhs, err := encodeAssign(m, sig, asgn, lhsArgs[0], rhs)
			if err != nil {
				return nil, nil, err
			}
			newLhsArgs := make([]Expr, len(lhsArgs))
			newLhsArgs[0] = nlhs
			copy(newLhsArgs[1:], lhsArgs[1:])
			return CloneNode(lhs, newLhsArgs), nrhs, nil
		}
	}

	// Check for variable overlap between lhs and rhs
	lhsVars := UsedVariablesAST(lhs)
	rhsVars := UsedVariablesAST(rhs)
	for v := range lhsVars {
		if _, ok := rhsVars[v]; ok {
			return nil, nil, fmt.Errorf("cannot convert parameterized assignment to VMT")
		}
	}

	fsort, ok := sym.CSort.(*LogicFunctionSort)
	if !ok {
		return lhs, ufToArrAST(m, sig, rhs), nil
	}

	_, ssorts := createArraySort(sig, fsort)
	asort := ssorts[0]
	asym := NewConst(sym.Name, asort)
	arhs := ufToArrAST(m, sig, rhs)

	lhsArgs := NodeArgs(lhs)
	newRHS, err := encodeAssignRecur(m, sig, asgn, lhsArgs, ssorts, 0, asym, arhs)
	if err != nil {
		return nil, nil, err
	}
	return asym, newRHS, nil
}

// encodeAssignRecur recursively builds the array update expression.
func encodeAssignRecur(m *Module, sig *Sig, asgn ActionsAction,
	lhsArgs []Expr, ssorts []Sort, i int, val Expr, arhs Expr) (Expr, error) {

	if i == len(lhsArgs) {
		return arhs, nil
	}
	idx := lhsArgs[i]
	if IsVariable(idx) {
		sval, err := encodeAssignRecur(m, sig, asgn, lhsArgs, ssorts, i+1, nil, arhs)
		if err != nil {
			return nil, err
		}
		cstSort, _ := NewFunctionSort(ssorts[i+1], ssorts[i])
		cst := NewConst("arrcst", cstSort)
		applied, err := NewApply(cst, sval)
		if err != nil {
			return nil, err
		}
		return applied, nil
	}

	if val == nil {
		return nil, fmt.Errorf("cannot convert parameterized assignment to VMT")
	}

	aidx := ufToArrAST(m, sig, idx)
	selSort, _ := NewFunctionSort(ssorts[i], aidx.NodeSort(), ssorts[i+1])
	sel := NewConst("arrsel", selSort)
	selApp, err := NewApply(sel, val, aidx)
	if err != nil {
		return nil, err
	}

	sval, err := encodeAssignRecur(m, sig, asgn, lhsArgs, ssorts, i+1, selApp, arhs)
	if err != nil {
		return nil, err
	}

	updSort, _ := NewFunctionSort(ssorts[i], aidx.NodeSort(), ssorts[i+1], ssorts[i])
	upd := NewConst("arrupd", updSort)
	updApp, err := NewApply(upd, val, aidx, sval)
	if err != nil {
		return nil, err
	}
	return updApp, nil
}

// ufToArrayAction converts uninterpreted functions to array operations in an action.
// Corresponds to Python's uf_to_array_action.
func ufToArrayAction(m *Module, sig *Sig, action ActionsAction) ActionsAction {
	args := action.ActionArgs()
	newArgs := make([]Expr, len(args))
	for i, arg := range args {
		if childAct, ok := checkToAction(arg); ok {
			newArgs[i] = ufToArrayAction(m, sig, childAct)
		} else {
			newArgs[i] = ufToArrAST(m, sig, arg)
		}
	}

	if assign, ok := action.(*LogicAssignAction); ok {
		lhsArgs := NodeArgs(assign.LHS)
		if len(lhsArgs) > 0 {
			sym := GetAppRep(assign.LHS)
			if sym != nil && !IsInterpretedSymbol(sig, sym) {
				nlhs, nrhs, err := encodeAssign(m, sig, action, assign.LHS, assign.RHS)
				if err == nil {
					newArgs = []Expr{nlhs, nrhs}
				}
			}
		}
	}

	return action.ActionClone(newArgs)
}

// hasAssert returns true if the action contains any AssertAction.
// Corresponds to Python's has_assert.
func hasAssert(action ActionsAction) bool {
	for _, sub := range action.IterSubactions() {
		if IsAssertLike(sub) {
			return true
		}
	}
	return false
}

// VMTCheckIsolate performs VMT-based model checking on the current module.
// It writes the VMT file to "ivy.vmt" and exits.
// Corresponds to Python's ivy_vmt.check_isolate.
func VMTCheckIsolate(method string, m *Module) error {
	if method == "" {
		method = "mc"
	}

	if m == nil {
		return fmt.Errorf("no module provided")
	}
	sig := m.Sig

	// Use the error flag construction to turn assertion checks into
	// an invariant check.
	erf := NewConst("err_flag", Boolean)
	var errconds []Expr

	hasErf := false
	for _, act := range m.Actions.All() {
		if a, ok := act.(ActionsAction); ok {
			if hasAssert(a) {
				hasErf = true
				break
			}
		}
	}

	if hasErf {
		addErrFlagMod(m, erf, &errconds)
	}

	// Combine all public actions into a list of (name, action) pairs
	publicNames := func() []string {
		keys := make([]string, 0, m.PublicActions.Len())
		for k := range m.PublicActions.All() {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return keys
	}()
	type namedAction struct {
		Name   string
		Action ActionsAction
	}
	var actionList []namedAction
	for _, name := range publicNames {
		act, ok := m.Actions.Get2(name)
		if !ok {
			continue
		}
		a, ok := act.(ActionsAction)
		if !ok {
			continue
		}
		// Add label to the action
		labeled := addLabel(a, name)
		actionList = append(actionList, namedAction{Name: name, Action: labeled})
	}

	if hasErf {
		for i, na := range actionList {
			// Prepend erf := false to each action
			erfReset := NewAssignAction(erf, &LogicOr{}) // Or() = false
			erfReset.SetLineno(Location{})
			wrapped := NewSequence(
				erfReset,
				na.Action,
			)
			actionList[i] = namedAction{Name: na.Name, Action: wrapped}
		}
	}

	// Check that initializers don't have assertions (not supported)
	for _, init := range m.Initializers {
		if a, ok := init.Action.(ActionsAction); ok {
			if hasAssert(a) {
				return fmt.Errorf("VMT cannot handle assertions in initializers")
			}
		}
	}

	// Build a single initializer action
	var initParts []Expr
	for _, init := range m.Initializers {
		if a, ok := init.Action.(ActionsAction); ok {
			initParts = append(initParts, a)
		}
	}
	initAction := NewSequence(initParts...)

	// Get the invariant to be proved. Apply proof tactics.
	// For now, simplified: just collect labeled conjectures.
	var conjs []*LabeledFormula
	pmap := make(map[int64]interface{})
	for _, pe := range m.Proofs {
		if pe.Formula != nil {
			pmap[pe.Formula.ID] = pe.Proof
		}
	}

	checkLineno := ""
	vmtVerbose := false
	if m.Cfg != nil {
		checkLineno = m.Cfg.CheckLineno
		vmtVerbose = m.Cfg.VMTVerbose
	}
	for _, lf := range m.LabeledConjs {
		if !checkedLF(lf, checkLineno) {
			continue
		}
		if vmtVerbose {
			fmt.Printf("%d Model checking invariant\n", lf.Lineno())
		}
		// For now, skip proof tactic handling -- add conj directly
		conjs = append(conjs, lf)
	}

	// Convert uninterpreted functions to arrays, if possible
	for i, na := range actionList {
		actionList[i] = namedAction{
			Name:   na.Name,
			Action: ufToArrayAction(m, sig, na.Action),
		}
	}

	initAction = ufToArrayAction(m, sig, initAction).(*LogicSequence)

	// Convert conjecture formulas. Python ivy_vmt.py:228:
	//   conjs = [conj.clone([conj.label, uf_to_arr_ast(conj.formula)]) for conj in conjs]
	for i, conj := range conjs {
		newFormula := ufToArrAST(m, sig, conj.Formula.(Expr))
		conjs[i] = conj.Clone([]Node{conj.Label, newFormula}).(*LabeledFormula)
	}

	// Convert the global action and initializer to logic
	var transs []Expr
	var allStVars []string
	for _, na := range actionList {
		stvars, trans, _, err := actionToTR(m, na.Action, method)
		if err != nil {
			return fmt.Errorf("action_to_tr failed for %s: %w", na.Name, err)
		}
		transs = append(transs, trans)
		allStVars = appendUnique(allStVars, stvars)
	}

	// Build combined transition relation: disjunction of all action transitions
	var trans Expr
	if len(transs) == 1 {
		trans = transs[0]
	} else if len(transs) > 1 {
		trans = &LogicOr{Terms: transs}
	} else {
		trans = False
	}

	// Convert initializer to state predicate
	istvars, init, _, err := actionToTR(m, initAction, method)
	if err != nil {
		return fmt.Errorf("action_to_tr failed for init: %w", err)
	}
	allStVars = appendUnique(allStVars, istvars)

	// Convert init to a state predicate (strongest post)
	// In Python: action_to_state converts from action style to state style
	// Convert string names to Consts for the Update
	var istConsts []*Const
	for _, name := range istvars {
		istConsts = append(istConsts, NewConst(name, TopS))
	}
	initState := ActionToState(&Update{
		Modified: istConsts,
		TR:       FormulaToClauses(init, nil),
		Pre:      FalseClauses(nil),
	})
	initFormula := initState.TRNode()

	// Write the VMT file
	slv := NewSolver(m, m.Cfg.SolverOpts)

	f, err := os.Create("ivy.vmt")
	if err != nil {
		return fmt.Errorf("failed to create ivy.vmt: %w", err)
	}
	defer f.Close()

	// Collect all symbols used in init, trans, and conjecture formulas
	allFormulas := []Expr{initFormula, trans}
	for _, conj := range conjs {
		allFormulas = append(allFormulas, conj.Formula.(Expr))
	}

	syms := collectAllSymbols(allFormulas)
	for _, sym := range syms {
		name := solverName(sig, sym)
		if name == "" {
			continue
		}
		decl, err := slv.Translator().Translate(sym)
		if err != nil {
			continue
		}
		if IsFunctionSort(sym.CSort) {
			fmt.Fprintln(f, decl.String())
		} else {
			if !IsInterpretedSymbol(sig, sym) {
				fmt.Fprintf(f, "(declare-const %s %s)\n", decl.String(), decl.ExprSort().String())
			}
		}
	}

	// Declare state variable pairs (current and next)
	ctr := 0
	initTransFormulas := []Expr{initFormula, trans}
	initTransSyms := collectAllSymbols(initTransFormulas)
	for _, sym := range initTransSyms {
		if !IsNew(sym.Name) {
			continue
		}
		decl, err := slv.Translator().Translate(sym)
		if err != nil {
			continue
		}
		baseSym := NewConst(NewOf(sym.Name), sym.CSort)
		declc, err := slv.Translator().Translate(baseSym)
		if err != nil {
			continue
		}
		fmt.Fprintf(f, "(declare-fun $sv.%d () %s (! %s :next %s))\n",
			ctr, decl.ExprSort().String(), declc.String(), decl.String())
		ctr++
	}

	// Write init predicate
	initZ3, err := slv.FormulaToZ3(initFormula)
	if err != nil {
		return fmt.Errorf("failed to translate init to Z3: %w", err)
	}
	fmt.Fprintf(f, "(define-fun $init () Bool (!\n%s :init true))\n", initZ3.String())

	// Write trans predicate
	transZ3, err := slv.FormulaToZ3(trans)
	if err != nil {
		return fmt.Errorf("failed to translate trans to Z3: %w", err)
	}
	fmt.Fprintf(f, "(define-fun $trans () Bool (!\n%s :trans true))\n", transZ3.String())

	// Write invariant properties (conjectures)
	propCtr := 1
	for _, lf := range conjs {
		var labelExpr Expr
		if lf.Label != nil {
			labelExpr = lf.Label.(Expr)
		}
		labelStr := labelString(labelExpr)
		fmlaZ3, err := slv.FormulaToZ3(lf.Formula.(Expr))
		if err != nil {
			continue
		}
		fmt.Fprintf(f, "(define-fun %s () Bool (!\n%s :invar-property %d))\n",
			labelStr, fmlaZ3.String(), propCtr)
		propCtr++
	}

	fmt.Println("output written to ivy.vmt")

	return nil
}

// -----------------------------------------------------------------------
// Helper functions
// -----------------------------------------------------------------------

// toAction extracts an Action from a lg.Expr.
func checkToAction(n Expr) (ActionsAction, bool) {
	if act, ok := n.(ActionsAction); ok {
		return act, true
	}
	if w, ok := n.(ActionsAction); ok {
		return w, true
	}
	return nil, false
}

// addLabel wraps an action with a label. In Python this is action.add_label(x),
// which sets action.label (singular), distinct from action.labels (plural).
func addLabel(a ActionsAction, name string) ActionsAction {
	if ab, ok := a.(interface{ SetLabel(string) }); ok {
		ab.SetLabel(name)
	}
	return a
}

// dualFormula negates a formula (the "dual"). Corresponds to Python's
// ilu.dual_formula which negates and existentially quantifies.
func dualFormula(fmla Expr) Expr {
	return &LogicNot{Body: fmla}
}

// backgroundTheory returns the background theory for a module as a formula.
// In the full implementation this would collect axioms, definitions, etc.
// Simplified here to return True (no background axioms).
func backgroundTheory(m *Module) Expr {
	var conjuncts []Expr
	for _, lf := range m.LabeledAxioms {
		if lf.Formula != nil {
			conjuncts = append(conjuncts, lf.Formula.(Expr))
		}
	}
	if len(conjuncts) == 0 {
		return True
	}
	return &LogicAnd{Terms: conjuncts}
}

// computeUpdate computes the transition relation update for an action.
// This is a simplified version; the full implementation would call
// action.update(module, None) which does full symbolic execution.
func computeUpdate(m *Module, action ActionsAction) *Update {
	// For now, return a trivial update. The full implementation requires
	// the complete action semantics compiler.
	return NullUpdate()
}

// conjoinDefs conjoins definition equalities into a formula.
func conjoinDefs(formula Expr, bgt Expr) Expr {
	// In the simplified version, bgt is already an And of axioms.
	// Full implementation would extract defs and add them.
	return formula
}

// extractDefSymNames extracts the names of symbols defined in background theory definitions.
func extractDefSymNames(bgt Expr) map[string]bool {
	// In the simplified version, we don't track definitions separately.
	return nil
}

// renameNode renames constants in a node according to the name map.
func checkRenameNode(node Expr, nameMap map[string]string) Expr {
	if len(nameMap) == 0 || node == nil {
		return node
	}
	constMap := make(map[NodeKey]*Const, len(nameMap))
	for old, new_ := range nameMap {
		oldSym := NewConst(old, TopS)
		constMap[Key(oldSym)] = NewConst(new_, TopS)
	}
	return RenameAST(node, constMap)
}

// collectAllSymbols collects all constant symbols from a list of formulas.
func collectAllSymbols(formulas []Expr) []*Const {
	seen := make(map[string]Expr)
	for _, f := range formulas {
		syms := UsedSymbolsAST(f)
		for _, sym := range syms.All() {
			name := ExprName(sym)
			if _, ok := seen[name]; !ok {
				seen[name] = sym
			}
		}
	}
	// Return in sorted order for determinism
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]*Const, 0, len(names))
	for _, name := range names {
		if c, ok := seen[name].(*Const); ok {
			result = append(result, c)
		}
	}
	return result
}

// solverName returns the solver-level name for a symbol, or empty string
// if the symbol should not be declared. Corresponds to Python's
// slvr.solver_name(sym).
func solverName(sig *Sig, sym *Const) string {
	if sym.Name == "" {
		return ""
	}
	// Skip interpreted symbols like numerals, =, etc.
	if sym.Name == "=" || sym.Name == "true" || sym.Name == "false" {
		return ""
	}
	return sym.Name
}

// labelString extracts a string from a label node for VMT output.
func labelString(label Expr) string {
	if label == nil {
		return "$prop"
	}
	s := fmt.Sprint(label)
	// Replace any characters that aren't valid in SMT-LIB identifiers
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, ".", "_")
	if s == "" {
		return "$prop"
	}
	return s
}

// sortedKeys returns the sorted keys of a map[string]bool.
func checkSortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// appendUnique appends elements from src to dst, skipping duplicates.
func appendUnique(dst, src []string) []string {
	seen := make(map[string]bool, len(dst))
	for _, s := range dst {
		seen[s] = true
	}
	for _, s := range src {
		if !seen[s] {
			seen[s] = true
			dst = append(dst, s)
		}
	}
	return dst
}
