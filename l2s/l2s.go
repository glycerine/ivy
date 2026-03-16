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

	// ---------------------------------------------------------------
	// L2S Transformation Implementation
	// ---------------------------------------------------------------
	//
	// Step 1: Get the temporal formula from the proof goal
	// In the full integration, we'd extract this from the goal object.
	// For now, we accept the module's temporal properties directly.

	// Step 2: Collect temporal properties from the module's conjectures
	var temporalFmlas []lg.Node
	for _, lf := range m.LabeledConjs {
		if lf.Formula != nil && il.HasTemporal(lf.Formula) {
			temporalFmlas = append(temporalFmlas, lf.Formula)
		}
	}
	if len(temporalFmlas) == 0 {
		return fmt.Errorf("l2s: no temporal properties to verify")
	}

	// Step 3: Normalize temporal operators into named binders
	for i, fmla := range temporalFmlas {
		temporalFmlas[i] = NormalizeTemporalFormula(fmla, state.ProofLabel)
	}

	// Step 4: Create the monitor symbols
	waiting := L2SWaiting()
	frozen := L2SFrozen()
	saved := L2SSaved()
	state.NewSymbols = append(state.NewSymbols, waiting, frozen, saved)

	// Step 5: Create saved-state copies of all module symbols
	if m.Sig != nil {
		for symName := range m.Sig.Symbols {
			savedName := CreateSavedCopy(symName)
			entry := m.Sig.Symbols[symName]
			if entry.Sort != nil {
				savedSym := lg.NewConst(savedName, entry.Sort)
				state.SavedSymbols[symName] = savedSym
				state.NewSymbols = append(state.NewSymbols, savedSym)
			}
		}
	}

	// Step 6: Build monitor automaton axioms
	//
	// l2s_waiting initially true
	state.NewAxioms = append(state.NewAxioms, waiting)

	// ~l2s_frozen initially
	state.NewAxioms = append(state.NewAxioms, &lg.Not{Body: frozen})

	// ~l2s_saved initially
	state.NewAxioms = append(state.NewAxioms, &lg.Not{Body: saved})

	// Step 7: Build the safety property from temporal formulas
	// The safety property is: if l2s_frozen, then the negation of the
	// temporal formula (in terms of saved state) must be false.
	// This is the key L2S reduction.
	for _, fmla := range temporalFmlas {
		// Build: l2s_frozen => fmla_in_saved_state
		savedFmla := substituteSavedState(fmla, state.SavedSymbols)
		safetyProp := &lg.Implies{T1: frozen, T2: savedFmla}
		state.NewConjs = append(state.NewConjs, safetyProp)
	}

	// Step 8: Add the new symbols and axioms to the module
	for _, sym := range state.NewSymbols {
		if m.Sig != nil {
			m.Sig.AddSymbol(sym.Name, sym.CSort)
		}
	}

	for _, axiom := range state.NewAxioms {
		m.LabeledInits = append(m.LabeledInits, &mod.LabeledFormula{
			Formula: axiom,
		})
	}

	for _, conj := range state.NewConjs {
		m.LabeledConjs = append(m.LabeledConjs, &mod.LabeledFormula{
			Formula: conj,
		})
	}

	_ = tacticName
	return nil
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

// substituteSavedState replaces module symbol references with their
// saved-state counterparts. This is used when building the safety
// property from the temporal formula.
func substituteSavedState(fmla lg.Node, savedSymbols map[string]*lg.Const) lg.Node {
	if len(savedSymbols) == 0 {
		return fmla
	}
	subs := make(map[string]lg.Node, len(savedSymbols))
	for name, savedSym := range savedSymbols {
		subs[name] = savedSym
	}
	return substituteConstsRec(fmla, subs)
}

func substituteConstsRec(node lg.Node, subs map[string]lg.Node) lg.Node {
	switch t := node.(type) {
	case *lg.Const:
		if r, ok := subs[t.Name]; ok {
			return r
		}
		return node
	case *lg.Var:
		return node
	case *lg.Apply:
		newFunc := substituteConstsRec(t.Func, subs)
		newTerms := make([]lg.Node, len(t.Terms))
		changed := newFunc != t.Func
		for i, arg := range t.Terms {
			newTerms[i] = substituteConstsRec(arg, subs)
			if newTerms[i] != arg {
				changed = true
			}
		}
		if !changed {
			return node
		}
		return &lg.Apply{Func: newFunc, Terms: newTerms}
	case *lg.Not:
		body := substituteConstsRec(t.Body, subs)
		if body == t.Body {
			return node
		}
		return &lg.Not{Body: body}
	case *lg.And:
		newTerms := make([]lg.Node, len(t.Terms))
		changed := false
		for i, term := range t.Terms {
			newTerms[i] = substituteConstsRec(term, subs)
			if newTerms[i] != term {
				changed = true
			}
		}
		if !changed {
			return node
		}
		return &lg.And{Terms: newTerms}
	case *lg.Or:
		newTerms := make([]lg.Node, len(t.Terms))
		changed := false
		for i, term := range t.Terms {
			newTerms[i] = substituteConstsRec(term, subs)
			if newTerms[i] != term {
				changed = true
			}
		}
		if !changed {
			return node
		}
		return &lg.Or{Terms: newTerms}
	case *lg.Implies:
		t1 := substituteConstsRec(t.T1, subs)
		t2 := substituteConstsRec(t.T2, subs)
		if t1 == t.T1 && t2 == t.T2 {
			return node
		}
		return &lg.Implies{T1: t1, T2: t2}
	case *lg.Eq:
		t1 := substituteConstsRec(t.T1, subs)
		t2 := substituteConstsRec(t.T2, subs)
		if t1 == t.T1 && t2 == t.T2 {
			return node
		}
		return &lg.Eq{T1: t1, T2: t2}
	case *lg.ForAll:
		body := substituteConstsRec(t.Body, subs)
		if body == t.Body {
			return node
		}
		return &lg.ForAll{Variables: t.Variables, Body: body}
	case *lg.Exists:
		body := substituteConstsRec(t.Body, subs)
		if body == t.Body {
			return node
		}
		return &lg.Exists{Variables: t.Variables, Body: body}
	case *lg.NamedBinder:
		body := substituteConstsRec(t.Body, subs)
		if body == t.Body {
			return node
		}
		return &lg.NamedBinder{Name: t.Name, Variables: t.Variables, Environ: t.Environ, Body: body}
	}
	return node
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
