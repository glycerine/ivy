package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
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
		return fmt.Sprintf("line %d: ", lf.Lineno)
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
		for name, act := range mod.Actions {
			if a, ok := act.(actions.Action); ok {
				actionMap[name] = a
			}
		}
		actionNames = actions.CallSet(actionName, actionMap)
	} else {
		for name := range mod.Actions {
			actionNames = append(actionNames, name)
		}
		sort.Strings(actionNames)
	}

	// Search each action for assert/ranking subactions
	// Python: isinstance(sub, (act.AssertAction, act.Ranking))
	for _, name := range actionNames {
		action, ok := mod.Actions[name]
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
	Clauses interface{}
	// Model holds the satisfying assignment.
	Model interface{}
	// Vocab contains the vocabulary symbols.
	Vocab []*lg.Symbol
	// Current tracks current symbol valuations (lhs → rhs).
	Current map[string]string
	// Eqs maps symbol names to their equality formulas.
	Eqs map[string][]lg.Expr
	// Renaming tracks symbol renamings (sym → renamed_sym).
	Renaming map[string]*lg.Symbol
	// Started is true after the initial state is printed.
	Started bool
	// Lines collects output lines.
	Lines []string
	// IsCti is set after trace is built; holds the failing conjecture clauses.
	IsCti interface{}
}

// NewMatchHandler creates a MatchHandler. Corresponds to Python's
// MatchHandler.__init__ (lines 282-310) which takes clauses, model, and vocab,
// then builds an equation map (eqs) from clauses_model_to_clauses.
func NewMatchHandler(clauses interface{}, model interface{}, vocab []*lg.Symbol) *MatchHandler {
	h := &MatchHandler{
		Clauses:  clauses,
		Model:    model,
		Vocab:    vocab,
		Current:  make(map[string]string),
		Eqs:      make(map[string][]lg.Expr),
		Renaming: make(map[string]*lg.Symbol),
	}
	// TODO: When solver.ClausesModelToClauses is available, extract ground
	// equalities from the model and populate h.Eqs. For each formula in
	// mod_clauses.fmlas:
	//   if is_eq: eqs[lhs.rep].append(fmla)
	//   elif is_not: eqs[app.rep].append(Equals(app, Or()))
	//   elif is_app: eqs[fmla.rep].append(Equals(fmla, And()))
	fmt.Println()
	fmt.Println("Trace follows...")
	fmt.Println(strings.Repeat("*", 80))
	return h
}

// ShowSym displays a symbol's value, applying renaming.
// Corresponds to Python's MatchHandler.show_sym (lines 312-324).
func (h *MatchHandler) ShowSym(sym, renamedSym *lg.Symbol) {
	if prev, ok := h.Renaming[sym.Name]; ok && prev.Name == renamedSym.Name {
		return
	}
	h.Renaming[sym.Name] = renamedSym
	// Display equations for this symbol
	for _, fmla := range h.Eqs[renamedSym.Name] {
		s := fmt.Sprintf("    %s", fmla)
		h.Lines = append(h.Lines, s)
		fmt.Println(s)
	}
}

// Eval evaluates a condition against the model.
// Corresponds to Python's MatchHandler.eval (lines 326-332).
func (h *MatchHandler) Eval(cond lg.Expr) bool {
	// TODO: implement using model.eval_to_constant when model supports it
	return true
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
func (h *MatchHandler) Handle(action interface{}, env map[string]*lg.Symbol) {
	if !h.Started {
		// Show initial values for vocab symbols not in env
		for _, sym := range h.Vocab {
			if _, inEnv := env[sym.Name]; !inEnv {
				if !tr.IsNew(sym.Name) && !h.IsSkolem(sym) {
					h.ShowSym(sym, sym)
				}
			}
		}
		h.Started = true
	}
	for symName, renamedSym := range env {
		sym := lg.NewSymbol(symName, nil)
		if !tr.IsNew(symName) && !h.IsSkolem(sym) {
			h.ShowSym(sym, renamedSym)
		}
	}
	line := fmt.Sprint(action)
	h.Lines = append(h.Lines, line)
	fmt.Println(line)
}

// DoReturn handles a return from an action. No-op in Python.
func (h *MatchHandler) DoReturn(action interface{}, env map[string]*lg.Symbol) {}

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
	for actname, action := range mod.Actions {
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
