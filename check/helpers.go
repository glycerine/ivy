package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
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
func PrettyLineno(lf *module.LabeledFormula) string {
	if lf == nil {
		return "(internal) "
	}
	if lf.Lineno > 0 {
		return fmt.Sprintf("line %d: ", lf.Lineno)
	}
	return "(internal) "
}

// PrettyLF formats a labeled formula for display with the given indent.
func PrettyLF(lf *module.LabeledFormula, indent int) string {
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

	// Search each action for assert subactions
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
			if _, isAssert := sub.(*actions.AssertAction); isAssert {
				result = append(result, sub)
			}
		}
	}

	return result
}

// --- MatchHandler ---

// MatchHandler reconstructs execution traces from satisfying assignments.
// This corresponds to Python's MatchHandler class.
type MatchHandler struct {
	// Model holds the satisfying assignment (stub: typed as interface).
	Model interface{}
	// Vocab contains the vocabulary symbols.
	Vocab interface{}
	// Current tracks current symbol valuations.
	Current map[string]string
	// Eqs maps symbol names to their equalities.
	Eqs map[string][]interface{}
	// Renaming tracks symbol renamings.
	Renaming map[string]string
	// Started is true after the initial state is printed.
	Started bool
	// Lines collects output lines.
	Lines []string
}

// NewMatchHandler creates a MatchHandler. Corresponds to Python's
// MatchHandler.__init__ which takes clauses, model, and vocab, then
// builds an equation map (eqs) from the model's clauses. In the full
// implementation, islv.clauses_model_to_clauses is called to extract
// ground equalities from the model. Until the solver interface is
// ported, we initialize the data structures but leave eqs empty.
func NewMatchHandler(model, vocab interface{}) *MatchHandler {
	h := &MatchHandler{
		Model:    model,
		Vocab:    vocab,
		Current:  make(map[string]string),
		Eqs:      make(map[string][]interface{}),
		Renaming: make(map[string]string),
	}
	// TODO: once solver is ported, call:
	//   modClauses := islv.ClausesModelToClauses(clauses, model, true)
	//   for _, fmla := range modClauses.Fmlas {
	//     populate h.Eqs from equalities in modClauses
	//   }
	fmt.Println()
	fmt.Println("Trace follows...")
	fmt.Println(strings.Repeat("*", 80))
	return h
}

// Handle processes an action in the trace.
func (h *MatchHandler) Handle(action interface{}, env map[string]string) {
	if !h.Started {
		h.Started = true
	}
	h.Lines = append(h.Lines, fmt.Sprint(action))
}

// End finalizes the trace output.
func (h *MatchHandler) End() {
	// Stub: print final state
}

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
	// Try as logic.Node
	if n, ok := f.(lg.Node); ok {
		return hasTemporalRec(n)
	}
	return false
}

// hasTemporalRec recursively checks for temporal operators and named binders.
func hasTemporalRec(n lg.Node) bool {
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
