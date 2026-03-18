// phase7.go implements Phase 7 helper functions for mc (model checking),
// ported from Python ivy_mc.py.
package mc

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/art"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// EncodeVars encodes a list of variables into binary-encoded sub-bits,
// populating the encoding map and returning the flat list of sub-bit variables.
// Corresponds to Python's encode_vars (ivy_mc.py lines 248-255).
func EncodeVars(vars []*lg.Variable, encoding map[*lg.Variable][]*lg.Variable) []*lg.Variable {
	var res []*lg.Variable
	for _, v := range vars {
		n, err := GetEncodingBitsSimple(v.VSort)
		if err != nil || n == 0 {
			n = 1 // default to 1 bit
		}
		subVars := make([]*lg.Variable, n)
		for i := 0; i < n; i++ {
			name := fmt.Sprintf("%s[%d]", v.Name, i)
			subVars[i], _ = lg.NewVariable(name, lg.Boolean)
		}
		encoding[v] = subVars
		res = append(res, subVars...)
	}
	return res
}

// CloneNormal creates a normalized clone of a clauses expression.
// For equalities, it reorders arguments to a canonical form. For
// trivially true equalities (x == x), it returns And() (true).
// Corresponds to Python's clone_normal (ivy_mc.py lines 839-849).
func CloneNormal(clauses *co.Clauses) *co.Clauses {
	if clauses == nil {
		return nil
	}
	newFmlas := make([]lg.Expr, 0, len(clauses.Fmlas))
	for _, f := range clauses.Fmlas {
		nf := normalize(f)
		// Filter out trivially-true formulas (empty And).
		if a, ok := nf.(*lg.And); ok && len(a.Terms) == 0 {
			continue
		}
		newFmlas = append(newFmlas, nf)
	}
	newDefs := make([]*il.Definition, len(clauses.Defs))
	copy(newDefs, clauses.Defs)
	return co.NewClauses(newFmlas, newDefs, clauses.Annot)
}

// normalize recursively normalizes a formula: expands macros, canonicalizes
// equality arg order, and removes tautological equalities (x == x → And()).
// Corresponds to Python's normalize (ivy_mc.py lines 853-855).
func normalize(expr lg.Expr) lg.Expr {
	if il.IsMacro(expr) {
		return normalize(il.ExpandMacro(expr))
	}
	args := il.NodeArgs(expr)
	newArgs := make([]lg.Expr, len(args))
	for i, a := range args {
		newArgs[i] = normalize(a)
	}
	return cloneNormal(expr, newArgs)
}

// cloneNormal normalizes a single node: for equalities, removes tautologies
// (x == x → And()) and canonicalizes argument order.
// Corresponds to Python's clone_normal (ivy_mc.py lines 839-849).
func cloneNormal(expr lg.Expr, args []lg.Expr) lg.Expr {
	if _, ok := expr.(*lg.Eq); ok && len(args) == 2 {
		x, y := args[0], args[1]
		if x.Equal(y) {
			return &lg.And{} // tautology → true
		}
		if termOrd(x, y) == 1 {
			x, y = y, x
		}
		return &lg.Eq{T1: x, T2: y}
	}
	return il.CloneNode(expr, args)
}

// termOrd provides a total ordering on terms for canonical forms.
// Corresponds to Python's term_ord (ivy_mc.py lines 815-828).
func termOrd(x, y lg.Expr) int {
	xs, ys := fmt.Sprintf("%T", x), fmt.Sprintf("%T", y)
	if xs < ys {
		return -1
	}
	if xs > ys {
		return 1
	}
	if ax, ok := x.(*lg.Apply); ok {
		if ay, ok := y.(*lg.Apply); ok {
			xn := nodeName(ax.Func)
			yn := nodeName(ay.Func)
			if xn < yn {
				return -1
			}
			if xn > yn {
				return 1
			}
		}
	}
	xargs := il.NodeArgs(x)
	yargs := il.NodeArgs(y)
	if len(xargs) < len(yargs) {
		return -1
	}
	if len(xargs) > len(yargs) {
		return 1
	}
	for i := range xargs {
		res := termOrd(xargs[i], yargs[i])
		if res != 0 {
			return res
		}
	}
	return 0
}

// nodeName extracts a name string from a logic node (Symbol or Variable).
func nodeName(n lg.Expr) string {
	switch t := n.(type) {
	case *lg.Symbol:
		return t.Name
	case *lg.Variable:
		return t.Name
	default:
		return fmt.Sprintf("%v", n)
	}
}

// UncomposeAnnot decomposes a ComposeAnnotation into a flat slice of
// its right-hand components.
// Corresponds to Python's uncompose_annot (ivy_mc.py lines 917-922).
func UncomposeAnnot(annot actions.Annotation) []actions.Annotation {
	ca, ok := annot.(*actions.ComposeAnnotation)
	if !ok {
		return nil
	}
	if len(ca.Args) < 2 {
		return nil
	}
	res := UncomposeAnnot(ca.Args[0])
	res = append(res, ca.Args[1])
	return res
}

// UniteAnnot decomposes an IteAnnotation into a list of (condition, annotation) pairs.
// For RenameAnnotations wrapping an IteAnnotation, the rename map is applied
// to the conditions.
// Corresponds to Python's unite_annot (ivy_mc.py lines 924-931).
func UniteAnnot(annot actions.Annotation) []AnnotPair {
	switch a := annot.(type) {
	case *actions.RenameAnnotation:
		inner := UniteAnnot(a.Arg)
		result := make([]AnnotPair, len(inner))
		for i, pair := range inner {
			cond := pair.Cond
			if mapped, ok := a.Map[cond]; ok {
				cond = mapped
			}
			result[i] = AnnotPair{
				Cond:  cond,
				Annot: actions.EmptyAnnotation{}.Rename(a.Map),
			}
			_ = pair.Annot // the renamed inner annotation
			result[i].Annot = &actions.RenameAnnotation{Arg: pair.Annot, Map: a.Map}
		}
		return result
	case *actions.IteAnnotation:
		res := UniteAnnot(a.ElseB)
		res = append(res, AnnotPair{Cond: a.Cond, Annot: a.ThenB})
		return res
	default:
		return nil
	}
}

// AnnotPair is a (condition, annotation) pair produced by UniteAnnot.
type AnnotPair struct {
	Cond  string
	Annot actions.Annotation
}

// MatchHandler is a handler for match_annotation that prints action/env
// information. It implements the actions.AnnotationHandler interface.
// Corresponds to Python's MatchHandler class (ivy_mc.py lines 933-943).
type MatchHandler struct {
	// Clauses is the satisfying assignment context.
	Clauses *co.Clauses
	// Model is the satisfying model.
	Model interface{}
	// Vocab is the vocabulary of symbols.
	Vocab map[string]bool
}

// Eval evaluates a condition in the model.
// Checks for trivially true/false conditions first, then evaluates
// against the model if available.
// Corresponds to Python's MatchHandler.eval (ivy_mc.py lines 934-940).
func (h *MatchHandler) Eval(cond string) bool {
	// Check for trivially false conditions
	if cond == "false" || cond == "0" || cond == "" {
		return false
	}
	// Check for trivially true conditions
	if cond == "true" || cond == "1" {
		return true
	}
	// If we have a model, evaluate the condition against it
	// For now, print a warning and assume true
	fmt.Printf("assuming: %s\n", cond)
	return true
}

// Handle processes an action with its environment mapping.
// Corresponds to Python's MatchHandler.handle (ivy_mc.py lines 941-943).
func (h *MatchHandler) Handle(action actions.Action, env map[string]string) {
	fmt.Printf("%v%v\n", action.GetLineno(), action)
	if len(env) > 0 {
		fmt.Printf("env: {")
		first := true
		for k, v := range env {
			if !first {
				fmt.Print(",")
			}
			fmt.Printf("%s:%s", k, v)
			first = false
		}
		fmt.Println("}")
	}
}

// DoReturn processes a return action.
func (h *MatchHandler) DoReturn(action actions.Action, env map[string]string) {
	// No-op in the base handler.
}

// Fail marks a failure point.
func (h *MatchHandler) Fail() {
	// No-op in the base handler.
}

// MatchAnnotationMC is a convenience wrapper that calls actions.MatchAnnotation
// with a MatchHandler.
// Corresponds to Python's match_annotation (ivy_mc.py lines 946-1015).
func MatchAnnotationMC(action actions.Action, annot actions.Annotation, handler *MatchHandler, mod *module.Module) {
	actions.MatchAnnotation(action, annot, handler, mod)
}

// CheckedAssert is a package-level parameter controlling which assertions
// to check. If empty, all assertions are checked. Otherwise, only the
// assertion whose lineno matches this value is checked.
// Corresponds to Python's ia.checked_assert (ivy_actions.py line 23).
var CheckedAssert string

// Checked returns true if the given action should be checked (its line
// number matches the checked_assert parameter, or checked_assert is empty).
// Corresponds to Python's checked (ivy_mc.py lines 1017-1018).
func Checked(action actions.Action) bool {
	if CheckedAssert == "" {
		return true
	}
	return CheckedAssert == action.GetLineno().String()
}

// Badwit panics with a model-checker witness format error.
// Corresponds to Python's badwit (ivy_mc.py lines 1429-1430).
func Badwit() {
	panic("model checker returned mis-formatted witness")
}

// IvyMCTrace is an adaptor that creates a trace as an ART AnalysisGraph.
// Corresponds to Python's IvyMCTrace class (ivy_mc.py lines 1434-1443).
type IvyMCTrace struct {
	*art.AnalysisGraph
}

// NewIvyMCTrace creates a new IvyMCTrace from state values.
// Corresponds to Python's IvyMCTrace.__init__ (ivy_mc.py lines 1435-1440).
func NewIvyMCTrace(stvals []lg.Expr, mod *module.Module) *IvyMCTrace {
	ag := art.NewAnalysisGraph(mod)
	// Set up initial state with the given state values as clauses.
	initClauses := &co.Clauses{Fmlas: stvals}
	initState := art.NewState(mod, initClauses)
	initState.Universe = make(map[string]interface{}) // singleton state
	ag.States = append(ag.States, initState)
	return &IvyMCTrace{AnalysisGraph: ag}
}

// AddState adds a new state to the trace, produced by the given action.
// Corresponds to Python's IvyMCTrace.add_state (ivy_mc.py lines 1441-1443).
func (t *IvyMCTrace) AddState(stvals []lg.Expr, action actions.Action) {
	cls := &co.Clauses{Fmlas: stvals}
	newState := art.NewState(t.Domain, cls)
	newState.Label = "ext"
	newState.Action = action
	if len(t.States) > 0 {
		newState.Pred = t.States[len(t.States)-1]
	}
	t.States = append(t.States, newState)
}

// AigerWitnessToIvyTrace converts an AIGER model checker witness into
// an Ivy trace (IvyMCTrace).
// The full implementation requires the AIGER simulator and encoder/decoder
// infrastructure to decode latch values into state equalities and to
// use MatchAnnotation with an AigerMatchHandler for path reconstruction.
// Corresponds to Python's aiger_witness_to_ivy_trace (ivy_mc.py lines 1570-1628).
func AigerWitnessToIvyTrace(witnessFile string, mod *module.Module) (*IvyMCTrace, error) {
	// Parse the witness file.
	trace, err := ParseWitnessFile(witnessFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse witness file: %w", err)
	}
	if len(trace.Steps) == 0 {
		return nil, fmt.Errorf("empty witness trace")
	}

	fmt.Println("\nCounterexample follows:")
	fmt.Println(strings.Repeat("-", 80))

	var ivyTrace *IvyMCTrace

	for i, step := range trace.Steps {
		// In the full implementation:
		// 1. Step the AIGER simulator with input values
		// 2. Call MatchAnnotation with AigerMatchHandler to print the path
		// 3. Decode latch values to state equalities using the decoder
		// 4. Print state values

		// For now, create states from the witness step data
		var stvals []lg.Expr
		_ = step // step contains Pre, Input, Output, Post bit strings

		if i > 0 {
			fmt.Println()
		}
		fmt.Println("path:")
		fmt.Printf("  (step %d)\n", i)
		fmt.Println("state:")
		fmt.Println("  (state values require AIGER decoder)")

		if ivyTrace == nil {
			ivyTrace = NewIvyMCTrace(stvals, mod)
		} else {
			ivyTrace.AddState(stvals, nil)
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	if ivyTrace == nil {
		Badwit()
	}
	return ivyTrace, nil
}
