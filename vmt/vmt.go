// Ported to Go from ivy_vmt.py.

// Package vmt provides VMT (Verification Modulo Theories) format export
// for model checking. It converts Ivy modules into VMT format files that
// can be consumed by VMT-compatible model checkers.
package vmt

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	mod "github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/solver"
	"github.com/glycerine/goivy/transrel"
	"github.com/glycerine/goivy/z3bridge"
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
func checkedLF(lf *ast.LabeledFormula, checkLineno string) bool {
	return checked(lf.Lineno, checkLineno)
}

// checkedAction returns true if the action should be checked.
func checkedAction(a actions.Action, checkLineno string) bool {
	loc := a.GetLineno()
	if checkLineno == "" {
		return true
	}
	return fmt.Sprintf("%d", loc.Line) == checkLineno
}

// actionToTR converts an action to a transition relation triple
// (stateVars, trans, error). Corresponds to Python's action_to_tr.
func actionToTR(m *mod.Module, action actions.Action, method string) ([]string, lg.Expr, lg.Expr, error) {
	// Get background theory
	bgt := m.BackgroundTheory(nil)

	// Compute the update (transition relation) for the action.
	// In Python: upd = action.update(im.module, None)
	// For fsmc method, loops would be unrolled first - not yet implemented.
	upd := computeUpdate(m, action)
	if upd == nil {
		return nil, lg.True, lg.False, fmt.Errorf("failed to compute update for action")
	}

	// Add post-state axioms
	postUpd := transrel.AddPostAxioms(upd, bgt)

	stvars := postUpd.Modified
	transNode := postUpd.TRNode()
	errNodeFmla := postUpd.PreNode()

	// Conjoin definitions from background theory into trans
	// Conjoin definitions from background theory as formulas
	if bgt != nil && len(bgt.Defs) > 0 {
		defsClauses := co.NewClauses(nil, bgt.Defs, nil)
		defsFormula := defsClauses.ToOpenFormula()
		if defsFormula != nil && !lg.IsTrue(defsFormula) {
			and, _ := lg.NewAnd(transNode, defsFormula)
			transNode = and
		}
	}

	defsyms := make(map[string]bool)
	if bgt != nil {
		for _, d := range bgt.Defs {
			if c, ok := d.Defines().(*lg.Symbol); ok {
				defsyms[c.Name] = true
			}
		}
	}
	if len(defsyms) > 0 {
		rn := make(map[string]string)
		for sym := range defsyms {
			newSym := transrel.New(sym)
			rn[newSym] = "__" + newSym
		}
		transNode = renameNode(transNode, rn)
		errNodeFmla = renameNode(errNodeFmla, rn)

		var filtered []*lg.Symbol
		for _, sv := range stvars {
			if !defsyms[sv.Name] {
				filtered = append(filtered, sv)
			}
		}
		stvars = filtered
	}

	return transrel.ModifiedNames(postUpd), transNode, errNodeFmla, nil
}

// addErrFlag transforms an action tree to use an error flag for assertion checking.
// Corresponds to Python's add_err_flag.
func addErrFlag(action actions.Action, erf lg.Expr, errconds *[]lg.Expr, checkLineno string, verbose bool) actions.Action {
	switch a := action.(type) {
	case *actions.AssertAction:
		if checkedAction(action, checkLineno) {
			if verbose {
				loc := action.GetLineno()
				fmt.Printf("%d:%s Model checking guarantee\n", loc.Line, loc.Filename)
			}
			// errcond = dual of the formula (negate after dropping universals)
			errcond := dualFormula(il.DropUniversals(a.Formula))
			// res = erf := erf | errcond
			orNode := &lg.Or{Terms: []lg.Expr{erf, errcond}}
			res := actions.NewAssignAction(erf, orNode)
			*errconds = append(*errconds, errcond)
			res.SetLineno(ast.Location{})
			return res
		}
		// Unchecked assert: treat as assume
		orNode := &lg.Or{Terms: []lg.Expr{erf, a.Formula}}
		res := actions.NewAssumeAction(orNode)
		res.SetLineno(ast.Location{})
		return res

	case *actions.AssumeAction:
		orNode := &lg.Or{Terms: []lg.Expr{erf, a.Formula}}
		res := actions.NewAssumeAction(orNode)
		res.SetLineno(ast.Location{})
		return res

	case *actions.Sequence:
		newArgs := make([]lg.Expr, len(a.Elems))
		for i, child := range a.Elems {
			if childAct, ok := toAction(child); ok {
				newArgs[i] = addErrFlag(childAct, erf, errconds, checkLineno, verbose)
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *actions.ChoiceAction:
		newArgs := make([]lg.Expr, len(a.Branches))
		for i, child := range a.Branches {
			if childAct, ok := toAction(child); ok {
				newArgs[i] = addErrFlag(childAct, erf, errconds, checkLineno, verbose)
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *actions.EnvAction:
		newArgs := make([]lg.Expr, len(a.Branches))
		for i, child := range a.Branches {
			if childAct, ok := toAction(child); ok {
				newArgs[i] = addErrFlag(childAct, erf, errconds, checkLineno, verbose)
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *actions.BindOldsAction:
		args := a.ActionArgs()
		newArgs := make([]lg.Expr, len(args))
		for i, child := range args {
			if childAct, ok := toAction(child); ok {
				newArgs[i] = addErrFlag(childAct, erf, errconds, checkLineno, verbose)
			} else {
				newArgs[i] = child
			}
		}
		return a.ActionClone(newArgs)

	case *actions.IfAction:
		// Keep condition, transform then/else branches
		args := a.ActionArgs()
		newArgs := make([]lg.Expr, len(args))
		newArgs[0] = args[0] // condition unchanged
		for i := 1; i < len(args); i++ {
			if childAct, ok := toAction(args[i]); ok {
				newArgs[i] = addErrFlag(childAct, erf, errconds, checkLineno, verbose)
			} else {
				newArgs[i] = args[i]
			}
		}
		return a.ActionClone(newArgs)

	case *actions.LocalAction:
		// Transform only the body (last arg)
		args := a.ActionArgs()
		newArgs := make([]lg.Expr, len(args))
		copy(newArgs, args)
		if len(newArgs) > 0 {
			lastIdx := len(newArgs) - 1
			if childAct, ok := toAction(newArgs[lastIdx]); ok {
				newArgs[lastIdx] = addErrFlag(childAct, erf, errconds, checkLineno, verbose)
			}
		}
		return a.ActionClone(newArgs)
	}

	return action
}

// addErrFlagMod transforms all actions in a module to use error flag checking.
// Corresponds to Python's add_err_flag_mod.
func addErrFlagMod(m *mod.Module, erf lg.Expr, errconds *[]lg.Expr) {
	for actname := range m.Actions {
		action, ok := m.Actions[actname].(actions.Action)
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
		m.Actions[actname] = newAction
	}
}

// createArraySort takes a function sort and returns a corresponding array
// sort name and sort chain. Corresponds to Python's create_array_sort.
func createArraySort(sig *il.Sig, fsort *lg.FunctionSort) (string, []lg.Sort) {
	dom := fsort.Domain()
	rng := fsort.Range()
	return createArraySortRec(sig, dom, 0, rng)
}

func createArraySortRec(sig *il.Sig, dom []lg.Sort, i int, rng lg.Sort) (string, []lg.Sort) {
	if i == len(dom) {
		return il.SortName(rng), []lg.Sort{rng}
	}
	sname, ssorts := createArraySortRec(sig, dom, i+1, rng)
	name := "arr[" + il.SortName(dom[i]) + "][" + sname + "]"
	if _, ok := sig.Sorts[name]; !ok {
		asort := &lg.UninterpretedSort{Name: name}
		sig.Sorts[name] = asort
		sig.Interp[name] = name
	}
	return name, append([]lg.Sort{sig.Sorts[name]}, ssorts...)
}

// encodeAsArray returns true if the symbol should be converted to an array.
// Corresponds to Python's encode_as_array.
func encodeAsArray(m *mod.Module, sig *il.Sig, sym *lg.Symbol) bool {
	if il.IsInterpretedSymbol(sig, sym) {
		return false
	}
	_, isDestructor := m.DestructorSorts[sym.Name]
	return !isDestructor
}

// ufToArrAST converts all uninterpreted functions in a formula to arrays.
// Corresponds to Python's uf_to_arr_ast.
func ufToArrAST(m *mod.Module, sig *il.Sig, node lg.Expr) lg.Expr {
	return ufToArrASTRec(m, sig, node)
}

func ufToArrASTRec(m *mod.Module, sig *il.Sig, node lg.Expr) lg.Expr {
	// Recursively transform arguments
	args := il.NodeArgs(node)
	newArgs := make([]lg.Expr, len(args))
	for i, arg := range args {
		newArgs[i] = ufToArrASTRec(m, sig, arg)
	}

	// Check if this is a function application that should be encoded as array
	if il.IsApp(node) && !il.IsNamedBinder(node) && len(args) > 0 {
		sym := il.GetAppRep(node)
		if sym != nil && encodeAsArray(m, sig, sym) {
			fsort, ok := sym.CSort.(*lg.FunctionSort)
			if ok {
				sname, ssorts := createArraySort(sig, fsort)
				_ = sname
				asym := lg.NewSymbol(sym.Name, ssorts[0])
				var result lg.Expr = asym
				for i, arg := range newArgs {
					selSort, _ := lg.NewFunctionSort(result.NodeSort(), arg.NodeSort(), ssorts[i+1])
					sel := lg.NewSymbol("arrsel", selSort)
					applied, err := lg.NewApply(sel, result, arg)
					if err != nil {
						return il.CloneNode(node, newArgs)
					}
					result = applied
				}
				return result
			}
		}
	}

	return il.CloneNode(node, newArgs)
}

// encodeAssign encodes a parameterized assignment to use array operations.
// Returns (newLHS, newRHS). Corresponds to Python's encode_assign.
func encodeAssign(m *mod.Module, sig *il.Sig, asgn actions.Action, lhs, rhs lg.Expr) (lg.Expr, lg.Expr, error) {
	sym := il.GetAppRep(lhs)
	if sym == nil {
		return lhs, rhs, nil
	}

	if _, isDestructor := m.DestructorSorts[sym.Name]; isDestructor {
		lhsArgs := il.NodeArgs(lhs)
		if len(lhsArgs) > 0 {
			nlhs, nrhs, err := encodeAssign(m, sig, asgn, lhsArgs[0], rhs)
			if err != nil {
				return nil, nil, err
			}
			newLhsArgs := make([]lg.Expr, len(lhsArgs))
			newLhsArgs[0] = nlhs
			copy(newLhsArgs[1:], lhsArgs[1:])
			return il.CloneNode(lhs, newLhsArgs), nrhs, nil
		}
	}

	// Check for variable overlap between lhs and rhs
	lhsVars := co.UsedVariablesAST(lhs)
	rhsVars := co.UsedVariablesAST(rhs)
	for v := range lhsVars {
		if _, ok := rhsVars[v]; ok {
			return nil, nil, fmt.Errorf("cannot convert parameterized assignment to VMT")
		}
	}

	fsort, ok := sym.CSort.(*lg.FunctionSort)
	if !ok {
		return lhs, ufToArrAST(m, sig, rhs), nil
	}

	_, ssorts := createArraySort(sig, fsort)
	asort := ssorts[0]
	asym := lg.NewSymbol(sym.Name, asort)
	arhs := ufToArrAST(m, sig, rhs)

	lhsArgs := il.NodeArgs(lhs)
	newRHS, err := encodeAssignRecur(m, sig, asgn, lhsArgs, ssorts, 0, asym, arhs)
	if err != nil {
		return nil, nil, err
	}
	return asym, newRHS, nil
}

// encodeAssignRecur recursively builds the array update expression.
func encodeAssignRecur(m *mod.Module, sig *il.Sig, asgn actions.Action,
	lhsArgs []lg.Expr, ssorts []lg.Sort, i int, val lg.Expr, arhs lg.Expr) (lg.Expr, error) {

	if i == len(lhsArgs) {
		return arhs, nil
	}
	idx := lhsArgs[i]
	if il.IsVariable(idx) {
		sval, err := encodeAssignRecur(m, sig, asgn, lhsArgs, ssorts, i+1, nil, arhs)
		if err != nil {
			return nil, err
		}
		cstSort, _ := lg.NewFunctionSort(ssorts[i+1], ssorts[i])
		cst := lg.NewSymbol("arrcst", cstSort)
		applied, err := lg.NewApply(cst, sval)
		if err != nil {
			return nil, err
		}
		return applied, nil
	}

	if val == nil {
		return nil, fmt.Errorf("cannot convert parameterized assignment to VMT")
	}

	aidx := ufToArrAST(m, sig, idx)
	selSort, _ := lg.NewFunctionSort(ssorts[i], aidx.NodeSort(), ssorts[i+1])
	sel := lg.NewSymbol("arrsel", selSort)
	selApp, err := lg.NewApply(sel, val, aidx)
	if err != nil {
		return nil, err
	}

	sval, err := encodeAssignRecur(m, sig, asgn, lhsArgs, ssorts, i+1, selApp, arhs)
	if err != nil {
		return nil, err
	}

	updSort, _ := lg.NewFunctionSort(ssorts[i], aidx.NodeSort(), ssorts[i+1], ssorts[i])
	upd := lg.NewSymbol("arrupd", updSort)
	updApp, err := lg.NewApply(upd, val, aidx, sval)
	if err != nil {
		return nil, err
	}
	return updApp, nil
}

// UFToArrayAction converts uninterpreted functions to array operations in an action.
// Corresponds to Python's uf_to_array_action.
func UFToArrayAction(m *mod.Module, sig *il.Sig, action actions.Action) actions.Action {
	args := action.ActionArgs()
	newArgs := make([]lg.Expr, len(args))
	for i, arg := range args {
		if childAct, ok := toAction(arg); ok {
			newArgs[i] = UFToArrayAction(m, sig, childAct)
		} else {
			newArgs[i] = ufToArrAST(m, sig, arg)
		}
	}

	if assign, ok := action.(*actions.AssignAction); ok {
		lhsArgs := il.NodeArgs(assign.LHS)
		if len(lhsArgs) > 0 {
			sym := il.GetAppRep(assign.LHS)
			if sym != nil && !il.IsInterpretedSymbol(sig, sym) {
				nlhs, nrhs, err := encodeAssign(m, sig, action, assign.LHS, assign.RHS)
				if err == nil {
					newArgs = []lg.Expr{nlhs, nrhs}
				}
			}
		}
	}

	return action.ActionClone(newArgs)
}

// hasAssert returns true if the action contains any AssertAction.
// Corresponds to Python's has_assert.
func hasAssert(action actions.Action) bool {
	for _, sub := range action.IterSubactions() {
		if _, ok := sub.(*actions.AssertAction); ok {
			return true
		}
	}
	return false
}

// CheckIsolate performs VMT-based model checking on the current module.
// It writes the VMT file to "ivy.vmt" and exits.
// Corresponds to Python's check_isolate.
func CheckIsolate(method string, m *mod.Module) error {
	if method == "" {
		method = "mc"
	}

	if m == nil {
		return fmt.Errorf("no module provided")
	}
	sig := m.Sig

	// Use the error flag construction to turn assertion checks into
	// an invariant check.
	erf := lg.NewSymbol("err_flag", lg.Boolean)
	var errconds []lg.Expr

	hasErf := false
	for _, act := range m.Actions {
		if a, ok := act.(actions.Action); ok {
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
	publicNames := sortedKeys(m.PublicActions)
	type namedAction struct {
		Name   string
		Action actions.Action
	}
	var actionList []namedAction
	for _, name := range publicNames {
		act, ok := m.Actions[name]
		if !ok {
			continue
		}
		a, ok := act.(actions.Action)
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
			erfReset := actions.NewAssignAction(erf, &lg.Or{}) // Or() = false
			erfReset.SetLineno(ast.Location{})
			wrapped := actions.NewSequence(
				erfReset,
				na.Action,
			)
			actionList[i] = namedAction{Name: na.Name, Action: wrapped}
		}
	}

	// Check that initializers don't have assertions (not supported)
	for _, init := range m.Initializers {
		if a, ok := init.Action.(actions.Action); ok {
			if hasAssert(a) {
				return fmt.Errorf("VMT cannot handle assertions in initializers")
			}
		}
	}

	// Build a single initializer action
	var initParts []lg.Expr
	for _, init := range m.Initializers {
		if a, ok := init.Action.(actions.Action); ok {
			initParts = append(initParts, a)
		}
	}
	initAction := actions.NewSequence(initParts...)

	// Get the invariant to be proved. Apply proof tactics.
	// For now, simplified: just collect labeled conjectures.
	var conjs []*ast.LabeledFormula
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
			fmt.Printf("%d Model checking invariant\n", lf.Lineno)
		}
		// For now, skip proof tactic handling -- add conj directly
		conjs = append(conjs, lf)
	}

	// Convert uninterpreted functions to arrays, if possible
	for i, na := range actionList {
		actionList[i] = namedAction{
			Name:   na.Name,
			Action: UFToArrayAction(m, sig, na.Action),
		}
	}

	initAction = UFToArrayAction(m, sig, initAction).(*actions.Sequence)

	// Convert conjecture formulas
	for i, conj := range conjs {
		newFormula := ufToArrAST(m, sig, conj.Formula.(lg.Expr))
		lf := m.Cfg.AstCfg.NewLabeledFormulaFrom(conj, newFormula)
		lf.ID = conj.ID
		conjs[i] = lf
	}

	// Convert the global action and initializer to logic
	var transs []lg.Expr
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
	var trans lg.Expr
	if len(transs) == 1 {
		trans = transs[0]
	} else if len(transs) > 1 {
		trans = &lg.Or{Terms: transs}
	} else {
		trans = lg.False
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
	var istConsts []*lg.Symbol
	for _, name := range istvars {
		istConsts = append(istConsts, lg.NewSymbol(name, lg.TopS))
	}
	initState := transrel.ActionToState(&transrel.Update{
		Modified: istConsts,
		TR:       co.FormulaToClauses(init, nil),
		Pre:      co.FalseClauses(nil),
	})
	initFormula := initState.TRNode()

	// Write the VMT file
	slv := solver.New()

	f, err := os.Create("ivy.vmt")
	if err != nil {
		return fmt.Errorf("failed to create ivy.vmt: %w", err)
	}
	defer f.Close()

	// Collect all symbols used in init, trans, and conjecture formulas
	allFormulas := []lg.Expr{initFormula, trans}
	for _, conj := range conjs {
		allFormulas = append(allFormulas, conj.Formula.(lg.Expr))
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
		if il.IsFunctionSort(sym.CSort) {
			fmt.Fprintln(f, decl.String())
		} else {
			if !il.IsInterpretedSymbol(sig, sym) {
				fmt.Fprintf(f, "(declare-const %s %s)\n", decl.String(), decl.ExprSort().String())
			}
		}
	}

	// Declare state variable pairs (current and next)
	ctr := 0
	initTransFormulas := []lg.Expr{initFormula, trans}
	initTransSyms := collectAllSymbols(initTransFormulas)
	for _, sym := range initTransSyms {
		if !transrel.IsNew(sym.Name) {
			continue
		}
		decl, err := slv.Translator().Translate(sym)
		if err != nil {
			continue
		}
		baseSym := lg.NewSymbol(transrel.NewOf(sym.Name), sym.CSort)
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
		var labelExpr lg.Expr
		if lf.Label != nil {
			labelExpr = lf.Label.(lg.Expr)
		}
		labelStr := labelString(labelExpr)
		fmlaZ3, err := slv.FormulaToZ3(lf.Formula.(lg.Expr))
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
func toAction(n lg.Expr) (actions.Action, bool) {
	if act, ok := n.(actions.Action); ok {
		return act, true
	}
	if w, ok := n.(actions.Action); ok {
		return w, true
	}
	return nil, false
}

// addLabel wraps an action with a label. In Python this is action.add_label(x).
// Since Go actions don't have an add_label method yet, we set labels if possible.
func addLabel(a actions.Action, name string) actions.Action {
	if ab, ok := a.(interface{ SetLabels([]string) }); ok {
		ab.SetLabels([]string{name})
	}
	return a
}

// dualFormula negates a formula (the "dual"). Corresponds to Python's
// ilu.dual_formula which negates and existentially quantifies.
func dualFormula(fmla lg.Expr) lg.Expr {
	return &lg.Not{Body: fmla}
}

// backgroundTheory returns the background theory for a module as a formula.
// In the full implementation this would collect axioms, definitions, etc.
// Simplified here to return True (no background axioms).
func backgroundTheory(m *mod.Module) lg.Expr {
	var conjuncts []lg.Expr
	for _, lf := range m.LabeledAxioms {
		if lf.Formula != nil {
			conjuncts = append(conjuncts, lf.Formula.(lg.Expr))
		}
	}
	if len(conjuncts) == 0 {
		return lg.True
	}
	return &lg.And{Terms: conjuncts}
}

// computeUpdate computes the transition relation update for an action.
// This is a simplified version; the full implementation would call
// action.update(module, None) which does full symbolic execution.
func computeUpdate(m *mod.Module, action actions.Action) *transrel.Update {
	// For now, return a trivial update. The full implementation requires
	// the complete action semantics compiler.
	return transrel.NullUpdate()
}

// conjoinDefs conjoins definition equalities into a formula.
func conjoinDefs(formula lg.Expr, bgt lg.Expr) lg.Expr {
	// In the simplified version, bgt is already an And of axioms.
	// Full implementation would extract defs and add them.
	return formula
}

// extractDefSymNames extracts the names of symbols defined in background theory definitions.
func extractDefSymNames(bgt lg.Expr) map[string]bool {
	// In the simplified version, we don't track definitions separately.
	return nil
}

// renameNode renames constants in a node according to the name map.
func renameNode(node lg.Expr, nameMap map[string]string) lg.Expr {
	if len(nameMap) == 0 || node == nil {
		return node
	}
	constMap := make(map[lg.NodeKey]*lg.Symbol, len(nameMap))
	for old, new_ := range nameMap {
		oldSym := lg.NewSymbol(old, lg.TopS)
		constMap[lg.Key(oldSym)] = lg.NewSymbol(new_, lg.TopS)
	}
	return co.RenameAST(node, constMap)
}

// collectAllSymbols collects all constant symbols from a list of formulas.
func collectAllSymbols(formulas []lg.Expr) []*lg.Symbol {
	seen := make(map[string]*lg.Symbol)
	for _, f := range formulas {
		syms := co.UsedSymbolsAST(f)
		for _, sym := range syms {
			if _, ok := seen[sym.Name]; !ok {
				seen[sym.Name] = sym
			}
		}
	}
	// Return in sorted order for determinism
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]*lg.Symbol, len(names))
	for i, name := range names {
		result[i] = seen[name]
	}
	return result
}

// solverName returns the solver-level name for a symbol, or empty string
// if the symbol should not be declared. Corresponds to Python's
// slvr.solver_name(sym).
func solverName(sig *il.Sig, sym *lg.Symbol) string {
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
func labelString(label lg.Expr) string {
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
func sortedKeys(m map[string]bool) []string {
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

// Ensure imports are used. These will be fully utilized once the complete
// action semantics compiler is in place.
var (
	_ = z3bridge.NewTranslator
)
