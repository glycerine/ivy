// Ported to Go from ivy_vmt.py. Merged into check/ from the former vmt/ package.
package goivy

import (
	"fmt"
	"sort"

	"github.com/glycerine/ivy/goivy/xtracer"
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
	modified, transition, failure, err := vmtActionTransition(m, action, method, m.BackgroundTheory(nil))
	var names []string
	for _, symbol := range modified {
		names = append(names, symbol.Name)
	}
	return names, transition, failure, err
}

func vmtActionTransition(m *Module, action ActionsAction, method string, bgt *Clauses) ([]*Const, Expr, Expr, error) {
	// For fsmc method, unroll loops before computing the update.
	// Python: with ia.UnrollContext(im.module.sort_card): ...
	if method == "fsmc" && m.Cfg != nil && m.Cfg.ActCfg != nil {
		uc := NewUnrollContext(m.SortCard, m, m.Cfg.ActCfg)
		uc.Enter()
		defer uc.Exit()
		xtracer.Trace("vmt.ActionToTR calling GetUpdate (fsmc) type=%s", ActionTypeName(action))
	} else {
		xtracer.Trace("vmt.ActionToTR calling GetUpdate type=%s", ActionTypeName(action))
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
		transNode = RenameASTByName(transNode, rn)
		errNodeFmla = RenameASTByName(errNodeFmla, rn)

		var filtered []*Const
		for _, sv := range stvars {
			if !defsyms[sv.Name] {
				filtered = append(filtered, sv)
			}
		}
		stvars = filtered
	}

	return stvars, transNode, errNodeFmla, nil
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
		sig.Interp.Set(name, name)
	}
	return name, append([]Sort{sig.Sorts.Get(name)}, ssorts...)
}

// vmtArrayEncoding lowers reads of mutable functions in symbolic transitions.
type vmtArrayEncoding map[NodeKey][]Sort

func (arrays vmtArrayEncoding) rewrite(node Expr) Expr {
	children := NodeArgs(node)
	if len(children) == 0 {
		return node
	}
	args := make([]Expr, len(children))
	for i, arg := range children {
		args[i] = arrays.rewrite(arg)
	}
	if IsApp(node) && !IsNamedBinder(node) {
		if symbol := GetAppRep(node); symbol != nil {
			if sorts := arrays[Key(symbol)]; len(sorts) != 0 {
				var value Expr = NewConst(symbol.Name, sorts[0])
				for i, arg := range args {
					sort, err := NewFunctionSort(sorts[i], arg.NodeSort(), sorts[i+1])
					panicOn(err)
					value = NewApplyUnchecked(NewConst("arrsel", sort), value, arg)
				}
				return value
			}
		}
	}
	return CloneNode(node, args)
}

// vmtArrayState converts state pairs and their reads after computing updates.
// Ivy's symbolic update already captures point writes, full assignments and
// possible aliasing between indices, including those dependent on a binder.
func vmtArrayState(m *Module, state *InsMap[NodeKey, *Const], transitions []vmtTransition) (*InsMap[NodeKey, *Const], vmtArrayEncoding) {
	arrays := make(vmtArrayEncoding)
	rewrite := func(symbol *Const) *Const {
		if sorts := arrays[Key(symbol)]; len(sorts) != 0 {
			return NewConst(symbol.Name, sorts[0])
		}
		return symbol
	}
	arrayState := NewInsMap[NodeKey, *Const]()
	for _, symbol := range state.All() {
		fs, function := symbol.CSort.(*LogicFunctionSort)
		if function && !IsInterpretedSymbol(m.Sig, symbol) {
			if _, destructor := m.DestructorSorts[symbol.Name]; !destructor {
				_, sorts := createArraySort(m.Sig, fs)
				arrays[Key(symbol)] = sorts
				arrays[Key(NewConst(ActionNewName(symbol.Name), symbol.CSort))] = sorts
			}
		}
		result := rewrite(symbol)
		arrayState.Set(Key(result), result)
	}
	for i := range transitions {
		transitions[i].formula = arrays.rewrite(transitions[i].formula)
		transitions[i].modified = append([]*Const(nil), transitions[i].modified...)
		for j, symbol := range transitions[i].modified {
			transitions[i].modified[j] = rewrite(symbol)
		}
	}
	return arrayState, arrays
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
	return exportVMTModel(method, m, true, nil)
}

func exportVMTModel(method string, m *Module, encodeArrays bool, predicates []vmtPredicate) error {
	if method == "" {
		method = "mc"
	}

	if m == nil {
		return fmt.Errorf("no module provided")
	}
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

	checkLineno := ""
	vmtVerbose := false
	if m.Cfg != nil {
		checkLineno = m.Cfg.CheckLineno
		vmtVerbose = m.Cfg.VMTVerbose
	}
	conjs, err := vmtCollectConjectures(m, checkLineno, vmtVerbose)
	if err != nil {
		return err
	}

	var transitions []vmtTransition
	state := NewInsMap[NodeKey, *Const]()
	for _, na := range actionList {
		modified, transition, _, err := vmtActionTransition(m, na.Action, method, m.BackgroundTheory(nil))
		if err != nil {
			return fmt.Errorf("action_to_tr failed for %s: %w", na.Name, err)
		}
		for _, symbol := range modified {
			state.Set(Key(symbol), symbol)
		}
		transitions = append(transitions, vmtTransition{name: na.Name, formula: transition, modified: modified})
	}
	modified, init, _, err := vmtActionTransition(m, initAction, method, m.BackgroundTheory(nil))
	if err != nil {
		return fmt.Errorf("action_to_tr failed for init: %w", err)
	}
	for _, symbol := range modified {
		state.Set(Key(symbol), symbol)
	}
	initUpdate := &Update{
		Modified: modified,
		TR:       FormulaToClauses(init, EmptyAnnotation{}),
		Pre:      FalseClauses(EmptyAnnotation{}),
	}
	initial := ActionToState(initUpdate).TRNode()
	if m.InitCond != nil {
		initial = ForwardImage(m.InitCond, m.BackgroundTheory(nil), initUpdate).ToOpenFormula()
	}
	var axioms []Expr
	for _, formula := range m.BackgroundTheory(nil).Fmlas {
		if !vmtMentionsState(formula, state) {
			axioms = append(axioms, formula)
		}
	}
	if encodeArrays {
		var arrays vmtArrayEncoding
		state, arrays = vmtArrayState(m, state, transitions)
		initial = arrays.rewrite(initial)
		for i, conj := range conjs {
			formula := arrays.rewrite(conj.Formula.(Expr))
			conjs[i] = conj.Clone([]Node{conj.Label, formula}).(*LabeledFormula)
		}
	}

	addVMTImmutableSymbols(m, initial, transitions, state, axioms)
	return writeVMT(m, initial, transitions, state, axioms, conjs, predicates)
}

// -----------------------------------------------------------------------
// Helper functions
// -----------------------------------------------------------------------

func vmtCollectConjectures(m *Module, checkLineno string, verbose bool) ([]*LabeledFormula, error) {
	proofCfg := TacticNewConfig()
	astCfg := NewAstConfig()
	if m != nil && m.Cfg != nil {
		if m.Cfg.ProofCfg != nil {
			proofCfg = m.Cfg.ProofCfg
		}
		if m.Cfg.AstCfg != nil {
			astCfg = m.Cfg.AstCfg
		}
	}
	pc := NewProofChecker(proofCfg, m, m.LabeledAxioms, m.Definitions, ModuleSchemataToAst(m.Schemata), astCfg)
	pmap := make(map[int64]Node)
	for _, pe := range m.Proofs {
		if pe.Formula != nil && pe.Proof != nil {
			pmap[pe.Formula.ID] = pe.Proof
		}
	}

	var conjs []*LabeledFormula
	for _, lf := range m.LabeledConjs {
		if !checkedLF(lf, checkLineno) {
			continue
		}
		if verbose {
			fmt.Printf("%d Model checking invariant\n", lf.Lineno())
		}
		if proof, ok := pmap[lf.ID]; ok {
			subgoals, err := pc.AdmitProposition(lf, proof)
			if err != nil {
				return nil, err
			}
			conjs = append(conjs, subgoals...)
			continue
		}
		conjs = append(conjs, lf)
	}
	return conjs, nil
}

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
// Python: action.update(im.module, None).
func computeUpdate(m *Module, action ActionsAction) *Update {
	if m == nil || action == nil {
		return nil
	}
	if m.Cfg == nil {
		m.Cfg = NewConfig()
	}
	if m.Cfg.ActCfg == nil {
		m.Cfg.ActCfg = NewActionsConfig()
	}
	ctx := &UpdateContext{
		Domain:          m,
		PVars:           nil,
		ActCfg:          m.Cfg.ActCfg,
		Instantiator:    m.Instantiator,
		CheckUnprovable: m.Cfg.OnlyCheckUnprovable,
		CheckedAssert:   m.Cfg.CheckLineno,
		GetAction: func(name string) ActionsAction {
			if found, ok := m.FindAction(name); ok {
				return found
			}
			return nil
		},
	}
	return GetUpdate(action, ctx)
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

// collectAllSymbols collects all constant symbols from a list of formulas.
func collectAllSymbols(formulas []Expr) []*Const {
	seen := make(map[NodeKey]*Const)
	for _, f := range formulas {
		for key, sym := range UsedSymbolsAST(f).All() {
			if symbol, ok := sym.(*Const); ok {
				seen[key] = symbol
			}
		}
	}
	result := make([]*Const, 0, len(seen))
	for _, symbol := range seen {
		result = append(result, symbol)
	}
	sort.Slice(result, func(i, j int) bool { return Key(result[i]) < Key(result[j]) })
	return result
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
