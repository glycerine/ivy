package goivy

import (
	"fmt"
	"sort"
	"strings"
)

// --- Pretty printing ---

// PrettyLabel formats a label for display. Returns "(no name)" for nil labels.
// Uses Repr() on Atom labels to include sort qualifiers on args,
// matching Python where str(label) falls to Atom.__repr__ which
// produces sort-qualified Variable args (e.g. M:mem_type).
func PrettyLabel(label interface{}) string {
	if label == nil {
		return "(no name)"
	}
	if a, ok := label.(*Atom); ok {
		return a.Repr()
	}
	s := fmt.Sprint(label)
	if s == "" || s == "<nil>" {
		return "(no name)"
	}
	return s
}

// PrettyLineno formats a line number from a labeled formula.
// Uses the full Location (with filename) from Base.Loc, matching
// Python's pretty_lineno which calls str(ast.lineno) where lineno
// is a LocationTuple that includes filename.
func CheckPrettyLineno(lf *LabeledFormula) string {
	if lf == nil {
		return "(internal) "
	}
	loc := lf.GetLineno()
	if loc.Filename != "" || loc.Line > 0 {
		return loc.String()
	}
	return "(internal) "
}

// PrettyLF formats a labeled formula for display with the given indent.
func PrettyLF(lf *LabeledFormula, indent int) string {
	if lf == nil {
		return strings.Repeat(" ", indent) + "(nil)"
	}
	return strings.Repeat(" ", indent) + CheckPrettyLineno(lf) + PrettyLabel(lf.Label)
}

// PrettyActionLineno formats an action's location for display.
// Mirrors PrettyLineno (for *ast.LabeledFormula) and matches Python's
// pretty_lineno which calls str(ast.lineno) when present.
// Returns "(internal) " when the action has no Location set.
// Uses Location.String() so the Reference chain (set during module
// instantiation by LinenoAddRef) is followed correctly.
func PrettyActionLineno(a ActionsAction) string {
	if a == nil {
		return "(internal) "
	}
	loc := a.GetLineno()
	if loc.Filename != "" || loc.Line > 0 {
		return loc.String()
	}
	return "(internal) "
}

// --- Action finding ---

// FindAssertions finds all assert actions reachable from the given action name.
// If actionName is empty, searches all actions in the module.
func FindAssertions(actionName string, mod *Module) []ActionsAction {
	var result []ActionsAction

	// Determine which actions to search
	var actionNames []string
	if actionName != "" {
		// Use call_set to find reachable actions
		actionMap := make(map[string]ActionsAction)
		for name, act := range mod.Actions.All() {
			if a, ok := act.(ActionsAction); ok {
				actionMap[name] = a
			}
		}
		actionNames = CallSet(actionName, actionMap)
	} else {
		for name := range mod.Actions.All() {
			actionNames = append(actionNames, name)
		}
		sort.Strings(actionNames)
	}

	// Search each action for assert/ranking subactions
	// Python: isinstance(sub, (act.AssertAction, act.Ranking))
	for _, name := range actionNames {
		action, ok := mod.Actions.Get2(name)
		if !ok {
			continue
		}
		act, ok := action.(ActionsAction)
		if !ok {
			continue
		}
		for _, sub := range act.IterSubactions() {
			isAssert := IsAssertLike(sub)
			_, isRanking := sub.(*LogicRanking)
			if isAssert || isRanking {
				result = append(result, sub)
			}
		}
	}

	return result
}

// --- MatchHandler ---

// MatchHandler reconstructs execution traces from satisfying assignments.
// This corresponds to Python's MatchHandler class (ivy_check.py lines 281-364).
type MatchHandler struct {
	// Clauses is the clause set used to build the model.
	Clauses *Clauses
	// Model holds the satisfying assignment.
	Model *ModelResult
	// Slv is the solver used to extract ground equalities.
	Slv *Solver
	// Vocab contains the vocabulary symbols.
	Vocab []*Const
	// Current tracks current symbol valuations (lhs key → rhs string).
	Current map[NodeKey]string
	// Eqs maps function symbol (by NodeKey) to their equality formulas.
	// Python: self.eqs = defaultdict(list); keyed by lhs.rep (the function symbol).
	Eqs map[NodeKey][]Expr
	// Renaming tracks symbol renamings (sym key → renamed_sym).
	Renaming map[NodeKey]*Const
	// Started is true after the initial state is printed.
	Started bool
	// Lines collects output lines.
	Lines []string
	// IsCti holds the failing conjecture clauses, if any. Set by the trace
	// formatter and consumed by the GUI analysis graph (currently a stub in
	// phase7.go GuiArt). Mirrors Python handler.is_cti (ivy_check.py:401-403).
	IsCti *Clauses

	// HiddenSymbols, when non-nil, returns true for symbol names that
	// should be hidden in trace output. Mirrors Python tr.hidden_symbols
	// (set by l2s diagnostic hooks to filter l2s_* auxiliary symbols).
	HiddenSymbols func(name string) bool

	// PP is an optional pretty-printer applied to formulas before display.
	// Mirrors Python tr.pp (set by auto_hook to ls2_g_to_globally).
	PP func(Expr) Expr

	// LoopStart is the index of the state that starts the fairness loop.
	// Set by the l2s_full trace hook. -1 means not set.
	// Mirrors Python tr.states[idx-1].loop_start = True (ivy_l2s.py:119).
	LoopStart int
}

// NewMatchHandler creates a MatchHandler. Corresponds to Python's
// MatchHandler.__init__ (lines 282-310) which takes clauses, model, and vocab,
// then calls islv.clauses_model_to_clauses to extract ground equalities.
func NewMatchHandler(clauses *Clauses, model *ModelResult, vocab []*Const, slv *Solver) *MatchHandler {
	h := &MatchHandler{
		Clauses:   clauses,
		Model:     model,
		Slv:       slv,
		Vocab:     vocab,
		Current:   make(map[NodeKey]string),
		Eqs:       make(map[NodeKey][]Expr),
		Renaming:  make(map[NodeKey]*Const),
		LoopStart: -1,
	}

	// Python: mod_clauses = islv.clauses_model_to_clauses(clauses, model=model, numerals=True)
	if slv != nil {
		modClauses, err := slv.ClausesModelToClausesWithModel(clauses, model, nil, true)
		if err == nil && modClauses != nil {
			// Python: for fmla in mod_clauses.fmlas:
			for _, fmla := range modClauses.Fmlas {
				switch f := fmla.(type) {
				case *Eq:
					// Python: if lg.is_eq(fmla): lhs,rhs = fmla.args; if lg.is_app(lhs): eqs[lhs.rep].append(fmla)
					// Python is_app returns True for Apply, Const, and nullary NamedBinder.
					if app, ok := f.T1.(*Apply); ok {
						key := Key(app.Func)
						h.Eqs[key] = append(h.Eqs[key], fmla)
					} else if cnst, ok := f.T1.(*Const); ok {
						// Python: is_app(Symbol) is True; eqs[lhs.rep].append(fmla)
						key := Key(cnst)
						h.Eqs[key] = append(h.Eqs[key], fmla)
					}
				case *LogicNot:
					// Python: elif isinstance(fmla, lg.Not): app = fmla.args[0]; eqs[app.rep].append(Equals(app, Or()))
					if app, ok := f.Body.(*Apply); ok {
						key := Key(app.Func)
						h.Eqs[key] = append(h.Eqs[key], &Eq{T1: app, T2: &LogicOr{}})
					} else if cnst, ok := f.Body.(*Const); ok {
						key := Key(cnst)
						h.Eqs[key] = append(h.Eqs[key], &Eq{T1: cnst, T2: &LogicOr{}})
					}
				default:
					// Python: elif lg.is_app(fmla): eqs[fmla.rep].append(Equals(fmla, And()))
					if app, ok := fmla.(*Apply); ok {
						key := Key(app.Func)
						h.Eqs[key] = append(h.Eqs[key], &Eq{T1: app, T2: &LogicAnd{}})
					} else if cnst, ok := fmla.(*Const); ok {
						key := Key(cnst)
						h.Eqs[key] = append(h.Eqs[key], &Eq{T1: cnst, T2: &LogicAnd{}})
					}
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("Trace follows...")
	fmt.Println(strings.Repeat("*", 80))
	return h
}

// ShowSym displays a symbol's value, applying renaming.
// Corresponds to Python's MatchHandler.show_sym (lines 312-324).
func (h *MatchHandler) ShowSym(sym, renamedSym *Const) {
	symKey := Key(sym)
	if prev, ok := h.Renaming[symKey]; ok && prev.Name == renamedSym.Name {
		return
	}
	h.Renaming[symKey] = renamedSym

	// Python: rmap = {renamed_sym: sym}
	// Python: for fmla in self.eqs[renamed_sym]: rfmla = lut.rename_ast(fmla, rmap)
	renamedKey := Key(renamedSym)
	// Match Python defaultdict auto-vivification
	if _, ok := h.Eqs[renamedKey]; !ok {
		h.Eqs[renamedKey] = nil
	}
	for _, fmla := range h.Eqs[renamedKey] {
		// Python: rfmla = lut.rename_ast(fmla, rmap); lhs,rhs = rfmla.args
		rfmla := RenameAST(fmla, map[NodeKey]*Const{Key(renamedSym): sym})
		// Python: if lhs in self.current and self.current[lhs] == rhs: continue
		if eq, ok := rfmla.(*Eq); ok {
			lhsKey := Key(eq.T1)
			rhsStr := fmt.Sprint(eq.T2)
			if cur, exists := h.Current[lhsKey]; exists && cur == rhsStr {
				continue
			}
			h.Current[lhsKey] = rhsStr
		}
		s := fmt.Sprintf("    %s", rfmla)
		h.Lines = append(h.Lines, s)
		fmt.Println(s)
	}
}

// Eval evaluates a condition against the model.
// Corresponds to Python's MatchHandler.eval (lines 326-332).
func (h *MatchHandler) Eval(cond Expr) bool {
	if h.Model == nil || h.Slv == nil {
		return true
	}
	// Build a HerbrandModel to evaluate the condition
	hm := NewHerbrandModel(h.Slv, h.Model.Solver, h.Model.Model, h.Vocab)
	truth := hm.EvalToConstant(cond)
	if IsFalse(truth) {
		return false
	}
	if IsTrue(truth) {
		return true
	}
	panic(fmt.Sprintf("unexpected truth value: %v", truth))
}

// IsSkolem checks if a symbol is a skolem (but not a __ prefixed uppercase one).
// Corresponds to Python's MatchHandler.is_skolem (lines 334-336).
func (h *MatchHandler) IsSkolem(sym *Const) bool {
	if !IsSkolem(sym.Name) {
		return false
	}
	// Python: not (sym.name.startswith('__') and sym.name[2:3].isupper())
	if len(sym.Name) > 2 && sym.Name[:2] == "__" {
		ch := sym.Name[2]
		if ch >= 'A' && ch <= 'Z' {
			return false
		}
	}
	return true
}

// Handle processes an action in the trace.
// Corresponds to Python's MatchHandler.handle (lines 338-353).
// Implements actions.AnnotationHandler.
func (h *MatchHandler) Handle(action ActionsAction, env map[NodeKey]Expr) {
	// Python: if hasattr(action,'lineno'):
	if !action.HasLineno() {
		return
	}
	lineno := action.GetLineno()

	if !h.Started {
		// Show initial values for vocab symbols not in env
		// Python: for sym in self.vocab: if sym not in env and not itr.is_new(sym) ...
		for _, sym := range h.Vocab {
			if _, inEnv := env[Key(sym)]; !inEnv {
				if !IsNew(sym.Name) && !h.IsSkolem(sym) {
					h.ShowSym(sym, sym)
				}
			}
		}
		h.Started = true
	}
	// Python: for sym, renamed_sym in env.items():
	//             if not itr.is_new(sym) and not self.is_skolem(sym):
	//                 self.show_sym(sym, renamed_sym)
	// Build a lookup from NodeKey → original *Symbol so we can check
	// IsNew/IsSkolem on the ORIGINAL symbol, not the renamed one.
	vocabByKey := make(map[NodeKey]*Const, len(h.Vocab))
	for _, sym := range h.Vocab {
		vocabByKey[Key(sym)] = sym
	}
	for symKey, renamedExpr := range env {
		renamedSym, ok := renamedExpr.(*Const)
		if !ok {
			continue
		}
		// Use the original symbol (from the env key) for IsNew/IsSkolem checks.
		// Python checks is_new(sym) and is_skolem(sym) on the ORIGINAL sym.
		origSym := vocabByKey[symKey]
		if origSym == nil {
			// Python's env only contains symbols from vocab, so skip
			// any that aren't in vocab. Using renamedSym as fallback
			// would check IsNew/IsSkolem on the wrong symbol name.
			continue
		}
		if !IsNew(origSym.Name) && !h.IsSkolem(origSym) {
			h.ShowSym(origSym, renamedSym)
		}
	}
	// Python: print('{}{}'.format(action.lineno, action))
	line := fmt.Sprintf("%s%v", lineno, action)
	h.Lines = append(h.Lines, line)
	fmt.Println(line)
}

// DoReturn handles a return from an action. No-op in Python.
// Implements actions.AnnotationHandler.
func (h *MatchHandler) DoReturn(action ActionsAction, env map[NodeKey]Expr) {}

// End finalizes the trace output.
// Corresponds to Python's MatchHandler.end (lines 358-361).
func (h *MatchHandler) End() {
	for _, sym := range h.Vocab {
		if !IsNew(sym.Name) && !h.IsSkolem(sym) {
			h.ShowSym(sym, sym)
		}
	}
}

// Fail handles trace failure. No-op in Python.
func (h *MatchHandler) Fail() {}

// String returns the collected trace as a string.
func (h *MatchHandler) String() string {
	return strings.Join(h.Lines, "\n")
}

// --- Call graph helpers ---

// BuildCallGraph builds the action call graph: called -> callers.
func BuildCallGraph(mod *Module) map[string][]string {
	callgraph := make(map[string][]string)
	for actname, action := range mod.Actions.All() {
		act, ok := action.(ActionsAction)
		if !ok {
			continue
		}
		for _, calledName := range act.IterCalls() {
			callgraph[calledName] = append(callgraph[calledName], actname)
		}
	}
	return callgraph
}

// PrettyActionName strips the "ext:" prefix from action names for display.
func PrettyActionName(name string) string {
	if strings.HasPrefix(name, "ext:") {
		return name[4:]
	}
	return name
}

// FilterCheckers filters checkers by location in "file:line" format.
// If checkLineno is empty, returns all checkers.
// Non-ConjChecker items always pass through (matching Python filter_fcs).
func FilterCheckers(checkers []Checker, checkLineno string) []Checker {
	if checkLineno == "" {
		return checkers
	}
	var result []Checker
	for _, fc := range checkers {
		if cc, ok := fc.(*ConjChecker); ok {
			if cc.LF.GetLineno().FileLineKey() == checkLineno {
				result = append(result, fc)
			}
		} else {
			result = append(result, fc)
		}
	}
	return result
}

// HasTemporalStuff returns true if the formula contains temporal operators
// or named binders. This corresponds to Python's has_temporal_stuff.
func HasTemporalStuff(f interface{}) bool {
	if f == nil {
		return false
	}
	// Try as logic.Expr
	if n, ok := f.(Expr); ok {
		return hasTemporalRec(n)
	}
	return false
}

// hasTemporalRec recursively checks for temporal operators and named binders.
func hasTemporalRec(n Expr) bool {
	if n == nil {
		return false
	}
	switch n.(type) {
	case *LogicGlobally, *LogicEventually, *LogicWhenOperator:
		return true
	case *LogicNamedBinder:
		return true
	}
	for _, c := range n.Children() {
		if hasTemporalRec(c) {
			return true
		}
	}
	return false
}

// --- LabeledFormula conversion helpers ---
// The proof package uses ast.LabeledFormula while check/ uses ast.LabeledFormula.
// These are structurally similar but distinct types. These helpers convert between them.

// ModuleLFToAstLF is the identity function — both module and ast use *ast.LabeledFormula.
// Retained for call-site compatibility; simply returns its argument.
func ModuleLFToAstLF(lf *LabeledFormula) *LabeledFormula {
	return lf
}

// AstLFToModuleLF is the identity function — both ast and module use *ast.LabeledFormula.
// Retained for call-site compatibility; simply returns its argument.
func AstLFToModuleLF(lf *LabeledFormula) *LabeledFormula {
	return lf
}

// ModuleSchemataToAst converts a module schemata InsMap to an ast schemata InsMap.
// Module.Schemata is *iu.InsMap[string, ast.Node] — values may be *ast.LabeledFormula
// or other ast.Node types depending on how they were stored. Preserves insertion
// order so ProofChecker receives entries in the same order Python sees them.
func ModuleSchemataToAst(schemata *InsMap[string, Node]) *InsMap[string, *LabeledFormula] {
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
