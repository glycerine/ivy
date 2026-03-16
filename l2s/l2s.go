// Package l2s implements liveness-to-safety reduction for temporal property verification.
//
// This is a port of Python's ivy_l2s.py. It transforms temporal (liveness)
// properties into safety properties that can be checked with standard
// IC3/UPDR verification.
//
// The transformation works by:
// 1. Constructing a monitor that tracks whether the temporal property holds
// 2. Adding saved-state copies of all relevant state
// 3. Adding Skolem constants/functions for existentially quantified variables
// 4. Adding fairness constraints
// 5. Proving that the resulting safety property implies the temporal one
//
// Key entry points:
//   - L2STactic: The main tactic for "l2s" proof goals
//   - L2STacticFull: Full version including auxiliary state
//   - L2STacticAuto: Automatic version with trigger inference
package l2s

import (
	"fmt"

	lg "github.com/glycerine/goivy/logic"
	il "github.com/glycerine/goivy/ivylogic"
	mod "github.com/glycerine/goivy/module"
)

// Debug controls l2s debug output.
var Debug bool

// --- Named constants used by the L2S transformation ---

// L2SWaiting is the "l2s_waiting" boolean flag.
func L2SWaiting() *lg.Const {
	return lg.NewConst("l2s_waiting", lg.Boolean)
}

// L2SFrozen is the "l2s_frozen" boolean flag.
func L2SFrozen() *lg.Const {
	return lg.NewConst("l2s_frozen", lg.Boolean)
}

// L2SSaved is the "l2s_saved" boolean flag.
func L2SSaved() *lg.Const {
	return lg.NewConst("l2s_saved", lg.Boolean)
}

// L2SD creates the l2s_d predicate for a sort (domain tracking).
func L2SD(sort lg.Sort) *lg.Const {
	return lg.NewConst("l2s_d", il.RelationSort([]lg.Sort{sort}))
}

// L2SA creates the l2s_a predicate for a sort (abstract domain).
func L2SA(sort lg.Sort) *lg.Const {
	return lg.NewConst("l2s_a", il.RelationSort([]lg.Sort{sort}))
}

// L2SW creates an l2s_w (waited) named binder.
func L2SW(vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{
		Name:      "l2s_w",
		Variables: vs,
		Environ:   &label,
		Body:      t,
	}
}

// L2SS creates an l2s_s (saved) named binder.
func L2SS(vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{
		Name:      "l2s_s",
		Variables: vs,
		Environ:   &label,
		Body:      t,
	}
}

// L2SG creates an l2s_g (globally/safety) named binder.
func L2SG(vs []*lg.Var, t lg.Node, environ string) *lg.NamedBinder {
	return &lg.NamedBinder{
		Name:      "l2s_g",
		Variables: vs,
		Environ:   &environ,
		Body:      t,
	}
}

// L2SInit creates an l2s_init named binder.
func L2SInit(vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{
		Name:      "l2s_init",
		Variables: vs,
		Environ:   &label,
		Body:      t,
	}
}

// L2SOld creates an l2s_old named binder.
func L2SOld(vs []*lg.Var, t lg.Node, label string) *lg.NamedBinder {
	return &lg.NamedBinder{
		Name:      "l2s_old",
		Variables: vs,
		Environ:   &label,
		Body:      t,
	}
}

// --- Transformation state ---

// L2SState holds the state accumulated during L2S transformation.
type L2SState struct {
	Module     *mod.Module
	ProofLabel string

	// Finite sorts that need domain tracking
	FiniteSorts map[string]lg.Sort

	// New symbols added by the transformation
	NewSymbols []*lg.Const

	// New axioms added
	NewAxioms []lg.Node

	// New conjectures added
	NewConjs []lg.Node

	// New actions (monitors)
	NewActions map[string]interface{}

	// Saved state relations
	SavedSymbols map[string]*lg.Const

	// Skolem constants from negation of temporal formula
	SkolemConsts []*lg.Const
}

// NewL2SState creates a new L2S transformation state.
func NewL2SState(m *mod.Module, label string) *L2SState {
	return &L2SState{
		Module:       m,
		ProofLabel:   label,
		FiniteSorts:  make(map[string]lg.Sort),
		NewActions:   make(map[string]interface{}),
		SavedSymbols: make(map[string]*lg.Const),
	}
}

// --- Tactic entry points ---

// L2STactic is the main entry point for the "l2s" proof tactic.
// It takes a prover, goals, and proof object and transforms temporal
// properties into safety properties.
//
// NOTE: This is a skeleton implementation. The full implementation requires:
// - proof infrastructure (goal_vocab, goal_conc, goal_prems)
// - compiler integration (compile_with_goal_vocab)
// - temporal logic normalization
// - monitor construction
// These will be filled in as the dependent modules are completed.
func L2STactic(m *mod.Module, goals []interface{}, proof interface{}) error {
	return l2sTacticInt(m, goals, proof, "l2s")
}

// L2STacticFull includes all auxiliary state in the transformation.
func L2STacticFull(m *mod.Module, goals []interface{}, proof interface{}) error {
	return l2sTacticInt(m, goals, proof, "l2s_full")
}

// L2STacticAuto uses automatic trigger inference.
func L2STacticAuto(m *mod.Module, goals []interface{}, proof interface{}) error {
	return l2sTacticInt(m, goals, proof, "l2s_auto")
}

// l2sTacticInt is the internal implementation of the L2S tactic.
func l2sTacticInt(m *mod.Module, goals []interface{}, proof interface{}, tacticName string) error {
	if len(goals) == 0 {
		return fmt.Errorf("l2s: no proof goals")
	}

	state := NewL2SState(m, "")

	// The full L2S transformation involves:
	//
	// 1. Extract the temporal formula from the proof goal
	// 2. Negate the temporal formula to get the safety property
	// 3. Normalize temporal operators (G, F, W) into named binders
	// 4. Create saved-state copies of all relevant symbols
	// 5. Create Skolem constants for existentially quantified variables
	// 6. Build the monitor automaton:
	//    a. l2s_waiting: initially true, becomes false when trigger fires
	//    b. l2s_frozen: becomes true when waiting becomes false
	//    c. l2s_saved: copies of all state at the freeze point
	// 7. Add invariant conjectures from the tactic
	// 8. Add fairness constraints
	// 9. Compile the monitor into actions that execute at each step
	// 10. Add the safety property assertion

	_ = state
	_ = tacticName

	// TODO: Full implementation requires proof infrastructure types.
	// The following is a structural skeleton showing the transformation steps.

	return fmt.Errorf("l2s: not yet fully implemented (requires proof infrastructure)")
}

// --- Helper functions ---

// NormalizeTemporalFormula normalizes a temporal formula by replacing
// temporal operators with named binders. This is a preprocessing step
// for the L2S transformation.
func NormalizeTemporalFormula(fmla lg.Node, label string) lg.Node {
	// Replace G(phi) with l2s_g(phi)
	// Replace F(phi) with ~l2s_g(~phi)
	// Replace phi W psi with l2s_w(phi, psi)
	return normalizeTemporalRec(fmla, label)
}

func normalizeTemporalRec(fmla lg.Node, label string) lg.Node {
	switch t := fmla.(type) {
	case *lg.Globally:
		body := normalizeTemporalRec(t.Body, label)
		return L2SG(nil, body, label)
	case *lg.Eventually:
		body := normalizeTemporalRec(t.Body, label)
		// F(phi) = ~G(~phi)
		return &lg.Not{Body: L2SG(nil, &lg.Not{Body: body}, label)}
	case *lg.Not:
		return &lg.Not{Body: normalizeTemporalRec(t.Body, label)}
	case *lg.And:
		terms := make([]lg.Node, len(t.Terms))
		for i, term := range t.Terms {
			terms[i] = normalizeTemporalRec(term, label)
		}
		return &lg.And{Terms: terms}
	case *lg.Or:
		terms := make([]lg.Node, len(t.Terms))
		for i, term := range t.Terms {
			terms[i] = normalizeTemporalRec(term, label)
		}
		return &lg.Or{Terms: terms}
	case *lg.Implies:
		return &lg.Implies{
			T1: normalizeTemporalRec(t.T1, label),
			T2: normalizeTemporalRec(t.T2, label),
		}
	case *lg.ForAll:
		return &lg.ForAll{
			Variables: t.Variables,
			Body:      normalizeTemporalRec(t.Body, label),
		}
	case *lg.Exists:
		return &lg.Exists{
			Variables: t.Variables,
			Body:      normalizeTemporalRec(t.Body, label),
		}
	}
	return fmla
}

// CreateSavedCopy creates a "saved" version of a symbol name.
func CreateSavedCopy(name string) string {
	return "l2s_saved_" + name
}

// IsSavedSymbol returns true if a symbol name is a saved copy.
func IsSavedSymbol(name string) bool {
	return len(name) > 10 && name[:10] == "l2s_saved_"
}

// IsL2SSymbol returns true if a symbol name is an L2S auxiliary symbol.
func IsL2SSymbol(name string) bool {
	return len(name) > 4 && name[:4] == "l2s_"
}

// TemporalAndL2S checks if a symbol is temporal-related or L2S-related.
func TemporalAndL2S(sym *lg.Const) bool {
	return IsL2SSymbol(sym.Name)
}

// L2SGToGlobally converts l2s_g named binders back to Globally operators.
func L2SGToGlobally(ast lg.Node) lg.Node {
	switch t := ast.(type) {
	case *lg.NamedBinder:
		if t.Name == "l2s_g" {
			return &lg.Globally{
				Environ: t.Environ,
				Body:    L2SGToGlobally(t.Body),
			}
		}
	}
	// Recurse into children
	children := ast.Children()
	if len(children) == 0 {
		return ast
	}
	newChildren := make([]lg.Node, len(children))
	changed := false
	for i, c := range children {
		nc := L2SGToGlobally(c)
		newChildren[i] = nc
		if nc != c {
			changed = true
		}
	}
	if !changed {
		return ast
	}
	return il.CloneNode(ast, newChildren)
}
