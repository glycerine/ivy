package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/solver"
	tr "github.com/glycerine/goivy/transrel"
)

// --- Pretty printing ---

// PrettyLabel formats a label for display. Returns "(no name)" for nil labels.
func PrettyLabel(label interface{}) string {
	if label == nil {
		return "(no name)"
	}
	s := fmt.Sprint(label)
	if s == "" || s == "<nil>" {
		return "(no name)"
	}
	return s
}

// PrettyLineno formats a line number from a labeled formula.
func PrettyLineno(lf *ast.LabeledFormula) string {
	if lf == nil {
		return "(internal) "
	}
	if lf.Lineno > 0 {
		// Python: return str(ast.lineno) — bare number, no "line" prefix.
		return fmt.Sprintf("%d", lf.Lineno)
	}
	return "(internal) "
}

// PrettyLF formats a labeled formula for display with the given indent.
func PrettyLF(lf *ast.LabeledFormula, indent int) string {
	if lf == nil {
		return strings.Repeat(" ", indent) + "(nil)"
	}
	return strings.Repeat(" ", indent) + PrettyLineno(lf) + PrettyLabel(lf.Label)
}

// --- Action finding ---

// FindAssertions finds all assert actions reachable from the given action name.
// If actionName is empty, searches all actions in the module.
func FindAssertions(actionName string, mod *module.Module) []actions.Action {
	var result []actions.Action

	// Determine which actions to search
	var actionNames []string
	if actionName != "" {
		// Use call_set to find reachable actions
		actionMap := make(map[string]actions.Action)
		for name, act := range mod.Actions.All() {
			if a, ok := act.(actions.Action); ok {
				actionMap[name] = a
			}
		}
		actionNames = actions.CallSet(actionName, actionMap)
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
		act, ok := action.(actions.Action)
		if !ok {
			continue
		}
		for _, sub := range act.IterSubactions() {
			_, isAssert := sub.(*actions.AssertAction)
			_, isRanking := sub.(*actions.Ranking)
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
	Clauses *clauseops.Clauses
	// Model holds the satisfying assignment.
	Model *solver.ModelResult
	// Slv is the solver used to extract ground equalities.
	Slv *solver.Solver
	// Vocab contains the vocabulary symbols.
	Vocab []*lg.Symbol
	// Current tracks current symbol valuations (lhs key → rhs string).
	Current map[lg.NodeKey]string
	// Eqs maps function symbol (by NodeKey) to their equality formulas.
	// Python: self.eqs = defaultdict(list); keyed by lhs.rep (the function symbol).
	Eqs map[lg.NodeKey][]lg.Expr
	// Renaming tracks symbol renamings (sym key → renamed_sym).
	Renaming map[lg.NodeKey]*lg.Symbol
	// Started is true after the initial state is printed.
	Started bool
	// Lines collects output lines.
	Lines []string
	// IsCti is set after trace is built; holds the failing conjecture clauses.
	IsCti interface{}
}

// NewMatchHandler creates a MatchHandler. Corresponds to Python's
// MatchHandler.__init__ (lines 282-310) which takes clauses, model, and vocab,
// then calls islv.clauses_model_to_clauses to extract ground equalities.
func NewMatchHandler(clauses *clauseops.Clauses, model *solver.ModelResult, vocab []*lg.Symbol, slv *solver.Solver) *MatchHandler {
	h := &MatchHandler{
		Clauses:  clauses,
		Model:    model,
		Slv:      slv,
		Vocab:    vocab,
		Current:  make(map[lg.NodeKey]string),
		Eqs:      make(map[lg.NodeKey][]lg.Expr),
		Renaming: make(map[lg.NodeKey]*lg.Symbol),
	}

	// Python: mod_clauses = islv.clauses_model_to_clauses(clauses, model=model, numerals=True)
	if slv != nil {
		modClauses, err := slv.ClausesModelToClausesWithModel(clauses, model, nil, true)
		if err == nil && modClauses != nil {
			// Python: for fmla in mod_clauses.fmlas:
			for _, fmla := range modClauses.Fmlas {
				switch f := fmla.(type) {
				case *lg.Eq:
					// Python: if lg.is_eq(fmla): lhs,rhs = fmla.args; if lg.is_app(lhs): eqs[lhs.rep].append(fmla)
					if app, ok := f.T1.(*lg.Apply); ok {
						key := lg.Key(app.Func)
						h.Eqs[key] = append(h.Eqs[key], fmla)
					}
				case *lg.Not:
					// Python: elif isinstance(fmla, lg.Not): app = fmla.args[0]; eqs[app.rep].append(Equals(app, Or()))
					if app, ok := f.Body.(*lg.Apply); ok {
						key := lg.Key(app.Func)
						h.Eqs[key] = append(h.Eqs[key], &lg.Eq{T1: app, T2: &lg.Or{}})
					}
				default:
					// Python: elif lg.is_app(fmla): eqs[fmla.rep].append(Equals(fmla, And()))
					if app, ok := fmla.(*lg.Apply); ok {
						key := lg.Key(app.Func)
						h.Eqs[key] = append(h.Eqs[key], &lg.Eq{T1: app, T2: &lg.And{}})
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
func (h *MatchHandler) ShowSym(sym, renamedSym *lg.Symbol) {
	symKey := lg.Key(sym)
	if prev, ok := h.Renaming[symKey]; ok && prev.Name == renamedSym.Name {
		return
	}
	h.Renaming[symKey] = renamedSym

	// Python: rmap = {renamed_sym: sym}
	// Python: for fmla in self.eqs[renamed_sym]: rfmla = lut.rename_ast(fmla, rmap)
	renamedKey := lg.Key(renamedSym)
	for _, fmla := range h.Eqs[renamedKey] {
		// Python: rfmla = lut.rename_ast(fmla, rmap); lhs,rhs = rfmla.args
		rfmla := clauseops.RenameAST(fmla, map[lg.NodeKey]*lg.Symbol{lg.Key(renamedSym): sym})
		// Python: if lhs in self.current and self.current[lhs] == rhs: continue
		if eq, ok := rfmla.(*lg.Eq); ok {
			lhsKey := lg.Key(eq.T1)
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
func (h *MatchHandler) Eval(cond lg.Expr) bool {
	if h.Model == nil || h.Slv == nil {
		return true
	}
	// Build a HerbrandModel to evaluate the condition
	hm := solver.NewHerbrandModel(h.Slv, h.Model.Solver, h.Model.Model, h.Vocab)
	truth := hm.EvalToConstant(cond)
	if lg.IsFalse(truth) {
		return false
	}
	if lg.IsTrue(truth) {
		return true
	}
	panic(fmt.Sprintf("unexpected truth value: %v", truth))
}

// IsSkolem checks if a symbol is a skolem (but not a __ prefixed uppercase one).
// Corresponds to Python's MatchHandler.is_skolem (lines 334-336).
func (h *MatchHandler) IsSkolem(sym *lg.Symbol) bool {
	if !tr.IsSkolem(sym.Name) {
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
func (h *MatchHandler) Handle(action actions.Action, env map[lg.NodeKey]lg.Expr) {
	// Python: if hasattr(action,'lineno'):
	lineno := action.GetLineno()
	if lineno.Line <= 0 {
		return
	}

	if !h.Started {
		// Show initial values for vocab symbols not in env
		// Python: for sym in self.vocab: if sym not in env and not itr.is_new(sym) ...
		for _, sym := range h.Vocab {
			if _, inEnv := env[lg.Key(sym)]; !inEnv {
				if !tr.IsNew(sym.Name) && !h.IsSkolem(sym) {
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
	vocabByKey := make(map[lg.NodeKey]*lg.Symbol, len(h.Vocab))
	for _, sym := range h.Vocab {
		vocabByKey[lg.Key(sym)] = sym
	}
	for symKey, renamedExpr := range env {
		renamedSym, ok := renamedExpr.(*lg.Symbol)
		if !ok {
			continue
		}
		// Use the original symbol (from the env key) for IsNew/IsSkolem checks.
		// Python checks is_new(sym) and is_skolem(sym) on the ORIGINAL sym.
		origSym := vocabByKey[symKey]
		if origSym == nil {
			// If original not in vocab, reconstruct from key info.
			// Fallback: use renamedSym for checks (may not be fully correct).
			origSym = renamedSym
		}
		if !tr.IsNew(origSym.Name) && !h.IsSkolem(origSym) {
			h.ShowSym(origSym, renamedSym)
		}
	}
	// Python: print('{}{}'.format(action.lineno, action))
	line := fmt.Sprintf("%d%v", lineno.Line, action)
	h.Lines = append(h.Lines, line)
	fmt.Println(line)
}

// DoReturn handles a return from an action. No-op in Python.
// Implements actions.AnnotationHandler.
func (h *MatchHandler) DoReturn(action actions.Action, env map[lg.NodeKey]lg.Expr) {}

// End finalizes the trace output.
// Corresponds to Python's MatchHandler.end (lines 358-361).
func (h *MatchHandler) End() {
	for _, sym := range h.Vocab {
		if !tr.IsNew(sym.Name) && !h.IsSkolem(sym) {
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
func BuildCallGraph(mod *module.Module) map[string][]string {
	callgraph := make(map[string][]string)
	for actname, action := range mod.Actions.All() {
		act, ok := action.(actions.Action)
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

// FilterCheckers filters checkers by line number.
// If checkLineno is empty, returns all checkers.
func FilterCheckers(checkers []Checker, checkLineno string) []Checker {
	if checkLineno == "" {
		return checkers
	}
	var result []Checker
	for _, fc := range checkers {
		if cc, ok := fc.(*ConjChecker); ok {
			if fmt.Sprintf("line %d", cc.LF.Lineno) == checkLineno ||
				fmt.Sprintf("%d", cc.LF.Lineno) == checkLineno {
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
	if n, ok := f.(lg.Expr); ok {
		return hasTemporalRec(n)
	}
	return false
}

// hasTemporalRec recursively checks for temporal operators and named binders.
func hasTemporalRec(n lg.Expr) bool {
	if n == nil {
		return false
	}
	switch n.(type) {
	case *lg.Globally, *lg.Eventually, *lg.WhenOperator:
		return true
	case *lg.NamedBinder:
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
func ModuleLFToAstLF(lf *ast.LabeledFormula) *ast.LabeledFormula {
	return lf
}

// AstLFToModuleLF is the identity function — both ast and module use *ast.LabeledFormula.
// Retained for call-site compatibility; simply returns its argument.
func AstLFToModuleLF(lf *ast.LabeledFormula) *ast.LabeledFormula {
	return lf
}

// ModuleSchemataToAst converts a module schemata map to ast schemata map.
// Module.Schemata is map[string]ast.Node — values may be *ast.LabeledFormula
// or other ast.Node types depending on how they were stored.
func ModuleSchemataToAst(schemata map[string]ast.Node) map[string]*ast.LabeledFormula {
	if schemata == nil {
		return nil
	}
	result := make(map[string]*ast.LabeledFormula, len(schemata))
	for k, v := range schemata {
		if s, ok := v.(*ast.LabeledFormula); ok {
			result[k] = s
		}
	}
	return result
}
