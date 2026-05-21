// l2s_shared.go contains the L2S instrumentation pipeline steps used by
// the L2S and ranking tactics. Each SharedStep* function corresponds to a
// numbered step in l2sTacticInt. The InstrumentationConfig struct carries
// all state between steps.
package goivy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// InstrumentationConfig holds all state for the shared L2S instrumentation pipeline.
type InstrumentationConfig struct {
	ProofLabel         string
	Lineno             Location
	FiniteSorts        map[string]bool
	UninterpretedSorts []Sort
	Mod                *Module
	Fmla               Expr              // the temporal formula
	Postconds          []*LabeledFormula // nil for l2s

	// Common building blocks (populated by BuildCommonBlocks or caller)
	AddConstsToD     []ActionsAction
	ResetW           []ActionsAction
	AssumeGAxioms    []ActionsAction
	AssumeWhenAxioms []ActionsAction
	AssumeWAxioms    []ActionsAction
	AssumeInitAxioms []ActionsAction

	// Collected state (populated by shared steps)
	L2sGs             map[NodeKey]L2sGTriple
	L2sWhensSet       map[NodeKey]*LogicNamedBinder
	NamedBindersConjs map[string][]VarBodyPair
	ToWait            []VarBodyPair
	ToSave            []VarBodyPair
	SaveState         []ActionsAction
	DoneWaiting       []Expr
	NotLf             Expr

	// Functions built during step 1
	ReplaceTemporals func(Node) Node
	// Dependencies closure (built by caller from defnDeps)
	Dependencies func(map[NodeKey]bool) map[NodeKey]bool

	// IsRankingTactic suppresses the L2S-only TOPLEVEL_NotLf_START trace,
	// ResetRtrDepth(), and notLf HASH trace in SharedStep1_ConvertTemporals.
	// Python's ranking.SharedStep1 omits these; l2s.SharedStep1 has them.
	IsRankingTactic bool

	// C5 trace_hook plumbing: data populated by l2sAutoInvariants and
	// SharedStep11_ReplaceNamedBinders, used to attach a hook to the
	// result goal so that the check package can route diagnostics.
	// Subs is the {fresh-const-name → original-binder-key} renaming map
	// (from SharedStep11). Tasks/Triggers are the per-suffix definition
	// maps from l2sAutoInvariants (only set for l2s_auto* tactics).
	Subs     map[string]string
	Tasks    map[string]map[string]*Eq
	Triggers map[string]map[string]*Eq

	// RSubs maps nonce const name → original NamedBinder (inverse of subs).
	// Used by extractJusticePredMap to resolve nonces back to original
	// NamedBinders for navigating formula structure.
	// Python: rsubs = dict((x,y) for (y,x) in subs.items())
	RSubs map[string]*LogicNamedBinder
	// FullSubs maps binder.Sexp() → nonce Const (the SharedStep11 subs map).
	// Used by extractJusticePredMap to resolve NamedBinders to their nonces.
	// Python: subs[jfmla.rep] at ivy_l2s.py:1551
	FullSubs map[string]Expr
}

// L2sGTriple holds the vars, body, and environ of an l2s_g binder.
type L2sGTriple = l2sGTriple

// VarBodyPair holds a pair of variables and a body expression.
type VarBodyPair = varBodyPair

// sortNamedBinderMap extracts values from a map[lg.NodeKey]*lg.NamedBinder and
// returns them sorted by Canon() for deterministic cross-language ordering.
func sortNamedBinderMap(m map[NodeKey]*LogicNamedBinder) []*LogicNamedBinder {
	sorted := make([]*LogicNamedBinder, 0, len(m))
	for _, v := range m {
		sorted = append(sorted, v)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return string(sorted[i].Canon()) < string(sorted[j].Canon())
	})
	return sorted
}

// sortL2sGTriples extracts values from a map[lg.NodeKey]L2sGTriple and
// returns them sorted by the full triple canon (Vars + Body + Environ) for
// strongly deterministic cross-language ordering. Sorting by Body.Canon()
// alone would leave ties for two triples with the same body but different
// vars or environ — a latent landmine that fires whenever such triples
// appear. Python ivy_l2s.py:1051 and :1202 are kept in lockstep using
// _l2s_g_triple_canon producing the byte-identical key format.
func sortL2sGTriples(m map[NodeKey]L2sGTriple) []L2sGTriple {
	sorted := make([]L2sGTriple, 0, len(m))
	for _, v := range m {
		sorted = append(sorted, v)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return string(sorted[i].key()) < string(sorted[j].key())
	})
	return sorted
}

// SharedStep1_ConvertTemporals converts temporal operators to named binders.
// This is Step 1 of l2sTacticInt. It populates cfg.L2sGs, cfg.L2sWhensSet,
// cfg.ReplaceTemporals, and cfg.NotLf.
//
// modPass should apply a transform to the entire model (invars, asms, bindings, init, invars list, and postconds if applicable).
func SharedStep1_ConvertTemporals(cfg *InstrumentationConfig, model *NormalProgram, modPass func(string, func(Node) Node)) {
	cfg.L2sGs = make(map[NodeKey]L2sGTriple)
	cfg.L2sWhensSet = make(map[NodeKey]*LogicNamedBinder)

	_l2sG := func(vs []*LogicVariable, t Expr, env *string) *LogicNamedBinder {
		res := l2sG(vs, t, env)
		triple := L2sGTriple{vs, t, env}
		cfg.L2sGs[triple.key()] = triple
		return res
	}
	_l2sWhen := func(name string, vs []*LogicVariable, t Expr) *LogicNamedBinder {
		fmt.Printf("l2s._l2sWhen CALLED name=%s nVars=%d HASH canon=%s\n", name, len(vs), t.Canon())
		// Key by Sexp (structural canonical form) not String (PrettyFmla,
		// drops sort annotations). Mirrors Python's set() in
		// l2s_whens (ivy_l2s.py:806,818,822) which dedups by NamedBinder
		// struct equality.
		if name == "first" {
			res := l2sWhen("next", vs, t, cfg.ProofLabel)
			cfg.L2sWhensSet[res.Sexp()] = res
			return l2sInit(vs, applyNB(res, checkVarsToNodes(vs)...), cfg.ProofLabel)
		}
		res := l2sWhen(name, vs, t, cfg.ProofLabel)
		cfg.L2sWhensSet[res.Sexp()] = res
		return res
	}

	cfg.ReplaceTemporals = func(n Node) Node {
		return ReplaceTemporalsByNamedBinder(n,
			func(vs []*LogicVariable, body Expr, env *string) *LogicNamedBinder {
				return _l2sG(vs, body, env)
			},
			func(name string, vs []*LogicVariable, body Expr) *LogicNamedBinder {
				return _l2sWhen(name, vs, body)
			},
		)
	}

	modPass("ReplaceTemporals", cfg.ReplaceTemporals)
	if !cfg.IsRankingTactic {
		xtracer.Trace("l2s.SharedStep1 TOPLEVEL_NotLf_START fmla HASH canon=%s", cfg.Fmla.Canon())
		ResetRtrDepth()
	}
	cfg.NotLf = cfg.ReplaceTemporals(&LogicNot{Body: cfg.Fmla}).(Expr)
	if !cfg.IsRankingTactic {
		xtracer.Trace("l2s.SharedStep1 notLf HASH canon=%s", cfg.NotLf.Canon())
	}

	// Normalize named binders
	modPass("NormalizeNamedBinders", func(n Node) Node {
		return NormalizeNamedBinders(n, nil)
	})
}

// SharedStep3_CollectNamedBinders collects l2s_w and l2s_s binders from
// model.Invars (and postconds if cfg.Postconds != nil).
// This is Step 3 of l2sTacticInt.
//
// The 'full' parameter controls whether to add all state variables (l2s_full mode).
// It populates cfg.NamedBindersConjs, cfg.ToWait, cfg.ToSave.
func SharedStep3_CollectNamedBinders(cfg *InstrumentationConfig, model *NormalProgram, full bool) {
	cfg.NamedBindersConjs = make(map[string][]VarBodyPair)

	// Collect from model.Invars
	sources := make([]Expr, 0, len(model.Invars)+len(cfg.Postconds))
	for _, inv := range model.Invars {
		sources = append(sources, inv.Formula.(Expr))
	}
	// Ranking also collects from postconds
	for _, pc := range cfg.Postconds {
		sources = append(sources, pc.Formula.(Expr))
	}
	for srcIdx, src := range sources {
		for _, b := range NamedBindersAst(src) {
			if !cfg.IsRankingTactic {
				xtracer.Trace("l2s.SharedStep3 collecting binder name=%s fromSource=%d nVars=%d HASH canon=%s", b.Name, srcIdx, len(b.Variables), b.Body.Canon())
			}
			cfg.NamedBindersConjs[b.Name] = append(cfg.NamedBindersConjs[b.Name],
				VarBodyPair{b.Variables, b.Body})
		}
	}
	for k, v := range cfg.NamedBindersConjs {
		cfg.NamedBindersConjs[k] = dedupeVarBodyPairs(v)
	}
	// L2S traces named_binders_conjs after dedup; ranking does not.
	if !cfg.IsRankingTactic {
		keys := make([]string, 0, len(cfg.NamedBindersConjs))
		for k := range cfg.NamedBindersConjs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			xtracer.Trace("l2s.SharedStep3 namedBindersConjs key=%s nEntries=%d", k, len(cfg.NamedBindersConjs[k]))
		}
	}

	// In full mode, add all state variables to 'to_save'
	if full {
		// Match Python defaultdict auto-vivification: bracket-read creates entry
		if _, ok := cfg.NamedBindersConjs["l2s_s"]; !ok {
			cfg.NamedBindersConjs["l2s_s"] = nil
		}
		// Key by Sexp (structural canonical form) not fmt.Sprint (which
		// calls String = PrettyFmla, drops sort annotations). Mirrors
		// Python's set(t for (vs,t) in named_binders_conjs['l2s_s'])
		// (ivy_l2s.py:892) which dedups by Expr struct equality.
		seenSave := make(map[string]bool)
		for _, vb := range cfg.NamedBindersConjs["l2s_s"] {
			seenSave[string(vb.Body.Sexp())] = true
		}
		m := cfg.Mod
		for _, bnd := range model.Bindings {
			for _, act := range bnd.Action.Stmt.IterSubactions() {
				mods := Modifies(act)
				for _, modSym := range mods {
					symName := modSym.Name
					if m != nil && m.Sig != nil {
						if entry, ok := m.Sig.Symbols.Get2(symName); ok {
							vs := SymPlaceholders(NewConst(symName, entry.Sort))
							var expr Expr
							if len(vs) > 0 {
								expr = checkMustApply(NewConst(symName, entry.Sort), checkVarsToNodes(vs)...)
							} else {
								expr = NewConst(symName, entry.Sort)
							}
							key := string(expr.Sexp())
							if !seenSave[key] {
								cfg.NamedBindersConjs["l2s_s"] = append(cfg.NamedBindersConjs["l2s_s"],
									VarBodyPair{vs, expr})
							}
						}
					}
				}
			}
		}

		// Match Python defaultdict auto-vivification
		if _, ok := cfg.NamedBindersConjs["l2s_w"]; !ok {
			cfg.NamedBindersConjs["l2s_w"] = nil
		}
		// Key by Sexp (structural canonical form). Mirrors Python's
		// set(t for (vs,t) in named_binders_conjs['l2s_w']) at
		// ivy_l2s.py:900 — Expr struct equality.
		seenWait := make(map[string]bool)
		for _, vb := range cfg.NamedBindersConjs["l2s_w"] {
			seenWait[string(vb.Body.Sexp())] = true
		}
		normNotLf := NormalizeNamedBinders(cfg.NotLf, nil).(Expr)
		for _, b := range NamedBindersAst(normNotLf) {
			if b.Name == "l2s_g" {
				negBody := Negate(b.Body)
				key := string(negBody.Sexp())
				if !seenWait[key] {
					cfg.NamedBindersConjs["l2s_w"] = append(cfg.NamedBindersConjs["l2s_w"],
						VarBodyPair{b.Variables, negBody})
				}
			}
			if b.Name == "l2s_init" {
				// Match Python defaultdict auto-vivification
				if _, ok := cfg.NamedBindersConjs["l2s_init"]; !ok {
					cfg.NamedBindersConjs["l2s_init"] = nil
				}
				cfg.NamedBindersConjs["l2s_init"] = append(cfg.NamedBindersConjs["l2s_init"],
					VarBodyPair{b.Variables, b.Body})
			}
		}
		cfg.NamedBindersConjs["l2s_init"] = dedupeVarBodyPairs(cfg.NamedBindersConjs["l2s_init"])
	}

	// Match Python defaultdict auto-vivification for non-full mode
	if _, ok := cfg.NamedBindersConjs["l2s_w"]; !ok {
		cfg.NamedBindersConjs["l2s_w"] = nil
	}
	if _, ok := cfg.NamedBindersConjs["l2s_s"]; !ok {
		cfg.NamedBindersConjs["l2s_s"] = nil
	}
	cfg.ToWait = cfg.NamedBindersConjs["l2s_w"]
	cfg.ToSave = cfg.NamedBindersConjs["l2s_s"]
	if !cfg.IsRankingTactic {
		for i, vb := range cfg.ToWait {
			xtracer.Trace("l2s.SharedStep3 toWait[%d] nVars=%d HASH canon=%s", i, len(vb.Vars), vb.Body.Canon())
		}
		for i, vb := range cfg.ToSave {
			xtracer.Trace("l2s.SharedStep3 toSave[%d] nVars=%d HASH canon=%s", i, len(vb.Vars), vb.Body.Canon())
		}
	}
}

// SharedBuildSaveAndWait builds saveState, doneWaiting, and resetW from ToWait/ToSave.
// Call after SharedStep3 and before SharedStep6.
func SharedBuildSaveAndWait(cfg *InstrumentationConfig) {
	// save_state actions
	cfg.SaveState = nil
	for i, vb := range cfg.ToSave {
		if !cfg.IsRankingTactic {
			xtracer.Trace("l2s.SharedBuildSaveAndWait saveState[%d] nVars=%d HASH canon=%s", i, len(vb.Vars), vb.Body.Canon())
		}
		lhs := applyNB(l2sS(vb.Vars, vb.Body, cfg.ProofLabel), checkVarsToNodes(vb.Vars)...)
		cfg.SaveState = append(cfg.SaveState, setLineno(NewAssignAction(lhs, vb.Body), cfg.Lineno))
	}

	// done_waiting formulas
	cfg.DoneWaiting = nil
	for i, vb := range cfg.ToWait {
		inner := applyNB(l2sW(vb.Vars, vb.Body, cfg.ProofLabel), checkVarsToNodes(vb.Vars)...)
		if !cfg.IsRankingTactic {
			xtracer.Trace("l2s.SharedBuildSaveAndWait doneWaiting[%d] nVars=%d HASH canon=%s", i, len(vb.Vars), inner.Canon())
		}
		cfg.DoneWaiting = append(cfg.DoneWaiting, forall(vb.Vars, &LogicNot{Body: inner}))
	}

	// reset_w actions
	cfg.ResetW = nil
	for i, vb := range cfg.ToWait {
		if !cfg.IsRankingTactic {
			xtracer.Trace("l2s.SharedBuildSaveAndWait resetW[%d] nVars=%d body HASH canon=%s", i, len(vb.Vars), vb.Body.Canon())
		}
		lhs := applyNB(l2sW(vb.Vars, vb.Body, cfg.ProofLabel), checkVarsToNodes(vb.Vars)...)
		var conjuncts []Expr
		for _, v := range vb.Vars {
			if !cfg.FiniteSorts[SortName(v.VSort)] {
				conjuncts = append(conjuncts, checkMustApply(L2SD(v.VSort), v))
			}
		}
		conjuncts = append(conjuncts, &LogicNot{Body: vb.Body})
		negatedBody := Negate(vb.Body)
		if !cfg.IsRankingTactic {
			xtracer.Trace("l2s.SharedBuildSaveAndWait resetW[%d] negatedBody HASH canon=%s", i, negatedBody.Canon())
		}
		preReplaceInput := &LogicNot{Body: &LogicGlobally{Environ: strPtr(cfg.ProofLabel), Body: negatedBody}}
		if !cfg.IsRankingTactic {
			xtracer.Trace("l2s.SharedBuildSaveAndWait resetW[%d] preReplace HASH canon=%s", i, preReplaceInput.Canon())
			ResetRtrDepth()
		}
		negGlob := cfg.ReplaceTemporals(preReplaceInput).(Expr)
		if !cfg.IsRankingTactic {
			xtracer.Trace("l2s.SharedBuildSaveAndWait resetW[%d] postReplace HASH canon=%s", i, negGlob.Canon())
		}
		conjuncts = append(conjuncts, negGlob)
		cfg.ResetW = append(cfg.ResetW, setLineno(NewAssignAction(lhs, checkMakeAnd(conjuncts...)), cfg.Lineno))
	}
	if !cfg.IsRankingTactic {
		xtracer.Trace("l2s.SharedBuildSaveAndWait EXIT nSaveState=%d nDoneWaiting=%d nResetW=%d", len(cfg.SaveState), len(cfg.DoneWaiting), len(cfg.ResetW))
	}
}

// SharedStep6_BuildTableau builds the tableau axiom actions.
// Populates cfg.AssumeGAxioms, cfg.AssumeWhenAxioms, cfg.AssumeInitAxioms, cfg.AssumeWAxioms.
func SharedStep6_BuildTableau(cfg *InstrumentationConfig) {
	toG := make([]L2sGTriple, 0, len(cfg.L2sGs))
	for _, triple := range cfg.L2sGs {
		toG = append(toG, triple)
	}
	// Sort by the full triple canon (Vars + Body + Environ), not Body.Canon()
	// alone. Two triples with the same body but different vars/environ would
	// tie under body-only sort, leaving nondeterministic order for those ties.
	// triple.key() is byte-identical to Python's _l2s_g_triple_canon used at
	// ivy_l2s.py:1051 and :1202, so the sort orders agree across languages.
	sort.Slice(toG, func(i, j int) bool {
		return string(toG[i].key()) < string(toG[j].key())
	})

	// assume_g_axioms
	if !cfg.IsRankingTactic {
		for i, triple := range toG {
			xtracer.Trace("l2s.SharedStep6 toG[%d] nVars=%d HASH canon=%s", i, len(triple.Vars), triple.Body.Canon())
		}
	}
	cfg.AssumeGAxioms = nil
	for _, triple := range toG {
		inner := &LogicImplies{
			T1: applyNB(l2sG(triple.Vars, triple.Body, triple.Environ), checkVarsToNodes(triple.Vars)...),
			T2: triple.Body,
		}
		cfg.AssumeGAxioms = append(cfg.AssumeGAxioms,
			setLineno(NewAssumeAction(forall(triple.Vars, inner)), cfg.Lineno))
	}

	// assume_when_axioms
	// Python ivy_l2s.py:1005-1008:
	//   AssumeAction(forall(when.variables, lg.Implies(when.body.t1, lg.Eq(when(*when.variables), when.body.t2))))
	// when.body is a Cond(condition, value); decompose T1=condition, T2=value.
	cfg.AssumeWhenAxioms = nil
	sortedWhens := sortNamedBinderMap(cfg.L2sWhensSet)
	for _, when := range sortedWhens {
		cond, ok := when.Body.(*Cond)
		if !ok {
			panic(fmt.Sprintf("assume_when_axioms: when binder %s has non-Cond body type %T", when.Name, when.Body))
		}
		inner := forall(when.Variables, &LogicImplies{
			T1: cond.T1,
			T2: &Eq{T1: applyNB(when, checkVarsToNodes(when.Variables)...), T2: cond.T2},
		})
		cfg.AssumeWhenAxioms = append(cfg.AssumeWhenAxioms,
			setLineno(NewAssumeAction(inner), cfg.Lineno))
	}

	// assume_init_axioms
	cfg.AssumeInitAxioms = nil
	for _, vb := range cfg.NamedBindersConjs["l2s_init"] {
		applied := applyL2sInit(vb.Vars, vb.Body, cfg.ProofLabel)
		inner := forall(vb.Vars, &Eq{T1: applied, T2: vb.Body})
		cfg.AssumeInitAxioms = append(cfg.AssumeInitAxioms,
			setLineno(NewAssumeAction(inner), cfg.Lineno))
	}

	// assume_w_axioms
	cfg.AssumeWAxioms = nil
	for _, vb := range cfg.NamedBindersConjs["l2s_w"] {
		wApp := applyNB(l2sW(vb.Vars, vb.Body, cfg.ProofLabel), checkVarsToNodes(vb.Vars)...)
		inner := forall(vb.Vars, &LogicNot{Body: &LogicAnd{Terms: []Expr{vb.Body, wApp}}})
		cfg.AssumeWAxioms = append(cfg.AssumeWAxioms,
			setLineno(NewAssumeAction(inner), cfg.Lineno))
	}
	if !cfg.IsRankingTactic {
		xtracer.Trace("l2s.SharedStep6 EXIT nAssumeG=%d nAssumeWhen=%d nAssumeInit=%d nAssumeW=%d",
			len(cfg.AssumeGAxioms), len(cfg.AssumeWhenAxioms), len(cfg.AssumeInitAxioms), len(cfg.AssumeWAxioms))
	}
}

// SharedStep7_InstrumentActions instruments all binding actions with
// prop events, when events, and wait events.
func SharedStep7_InstrumentActions(cfg *InstrumentationConfig, model *NormalProgram) {
	symprops := make(map[NodeKey][]*LogicNamedBinder)
	symwaits := make(map[NodeKey][]*LogicNamedBinder)
	symwhens := make(map[NodeKey][]*LogicNamedBinder)
	sortedTriples := sortL2sGTriples(cfg.L2sGs)
	displayTi := 0
	for _, triple := range sortedTriples {
		prop := l2sG(triple.Vars, triple.Body, triple.Environ)
		si := 0
		for sym := range SymbolsIluAst(triple.Body) {
			if c, ok := sym.(*Const); ok {
				k := c.Sexp()
				if !cfg.IsRankingTactic {
					xtracer.Trace("l2s.SharedStep7 symprops triple[%d] sym[%d]=%s HASH canon=%s", displayTi, si, c.Name, triple.Body.Canon())
				}
				symprops[k] = append(symprops[k], prop)

				si++
			}
		}
		if si > 0 {
			displayTi++
		}
	}
	sortedWhens7 := sortNamedBinderMap(cfg.L2sWhensSet)
	for wi, when := range sortedWhens7 {
		si := 0
		for sym := range SymbolsIluAst(when.Body) {
			if c, ok := sym.(*Const); ok {
				k := c.Sexp()
				if !cfg.IsRankingTactic && xtracer.Enabled {
					xtracer.Trace("l2s.SharedStep7 symwhens when[%d] sym[%d]=%s HASH canon=%s", wi, si, c.Name, when.Body.Canon())
					fmt.Printf("l2s.SharedStep7 symwhens when[%d] sym[%d]=%s HASH canon=%s\n", wi, si, c.Name, when.Body.Canon())
				}
				symwhens[k] = append(symwhens[k], when)

				si++
			} else {
				if !cfg.IsRankingTactic {
					fmt.Printf("l2s.SharedStep7 not lgConst! type(sym)=%T; symwhens when[%d] when='%v' sym=%s HASH canon= when.Body=%s\n", sym, wi, when, sym, when.Body.Canon())
				}
			}
		}
	}
	for wi, vb := range cfg.ToWait {
		wait := l2sW(vb.Vars, vb.Body, cfg.ProofLabel)
		si := 0
		for sym := range SymbolsIluAst(vb.Body) {
			if c, ok := sym.(*Const); ok {
				k := c.Sexp()
				if !cfg.IsRankingTactic {
					xtracer.Trace("l2s.SharedStep7 symwaits toWait[%d] sym[%d]=%s HASH canon=%s", wi, si, c.Name, vb.Body.Canon())
				}
				symwaits[k] = append(symwaits[k], wait)

				si++
			} else {
				if !cfg.IsRankingTactic {
					fmt.Printf("l2s.SharedStep7 not lgConst! type(sym)=%T; symwaits toWait[%d] sym=%s HASH canon= vb.Body=%s\n", sym, wi, sym, vb.Body.Canon())
				}
			}
		}
	}

	if false {
		// Dump all keys in each map
		fmt.Printf("l2s.SharedStep7 symprops keys: %v\n", func() []string {
			keys := make([]string, 0, len(symprops))
			for k := range symprops {
				keys = append(keys, string(k))
			}
			sort.Strings(keys)
			return keys
		}())
		fmt.Printf("l2s.SharedStep7 symwhens keys: %v\n", func() []string {
			keys := make([]string, 0, len(symwhens))
			for k := range symwhens {
				keys = append(keys, string(k))
			}
			sort.Strings(keys)
			return keys
		}())
		fmt.Printf("l2s.SharedStep7 symwaits keys: %v\n", func() []string {
			keys := make([]string, 0, len(symwaits))
			for k := range symwaits {
				keys = append(keys, string(k))
			}
			sort.Strings(keys)
			return keys
		}())
	}

	lineno := cfg.Lineno

	propEventsFunc := func(gprops map[NodeKey]*LogicNamedBinder) ([]ActionsAction, []ActionsAction) {
		var pre, post []ActionsAction
		sortedProps := sortNamedBinderMap(gprops)
		for _, gprop := range sortedProps {
			vs, t, env := gprop.Variables, gprop.Body, gprop.Environ
			pre = append(pre,
				setLineno(NewAssignAction(
					applyNB(oldL2sG(vs, t, env), checkVarsToNodes(vs)...),
					applyNB(l2sG(vs, t, env), checkVarsToNodes(vs)...),
				), lineno))
			pre = append(pre,
				setLineno(NewHavocAction(
					applyNB(l2sG(vs, t, env), checkVarsToNodes(vs)...),
				), lineno))
		}
		for _, gprop := range sortedProps {
			vs, t, env := gprop.Variables, gprop.Body, gprop.Environ
			pre = append(pre,
				setLineno(NewAssumeAction(forall(vs,
					&LogicImplies{
						T1: applyNB(oldL2sG(vs, t, env), checkVarsToNodes(vs)...),
						T2: applyNB(l2sG(vs, t, env), checkVarsToNodes(vs)...),
					})), lineno))
			pre = append(pre,
				setLineno(NewAssumeAction(forall(vs,
					&LogicImplies{
						T1: &LogicAnd{Terms: []Expr{
							&LogicNot{Body: applyNB(oldL2sG(vs, t, env), checkVarsToNodes(vs)...)},
							t,
						}},
						T2: &LogicNot{Body: applyNB(l2sG(vs, t, env), checkVarsToNodes(vs)...)},
					})), lineno))
			post = append(post,
				setLineno(NewAssumeAction(forall(vs,
					&LogicImplies{
						T1: applyNB(l2sG(vs, t, env), checkVarsToNodes(vs)...),
						T2: t,
					})), lineno))
		}
		return pre, post
	}

	// Python ivy_l2s.py:1070-1085: when_events.
	// when.body is a Cond(condition, value); decompose T1=cond, T2=val.
	whenEventsFunc := func(whens map[NodeKey]*LogicNamedBinder) ([]ActionsAction, []ActionsAction) {
		var pre, post []ActionsAction
		sortedWhens := sortNamedBinderMap(whens)
		for _, when := range sortedWhens {
			condVal, ok := when.Body.(*Cond)
			if !ok {
				panic(fmt.Sprintf("whenEventsFunc: when binder %s has non-Cond body type %T", when.Name, when.Body))
			}
			vs := when.Variables
			cond := condVal.T1
			if when.Name == "l2s_whennext" {
				oldcond := applyNB(l2sOld(vs, cond, cfg.ProofLabel), checkVarsToNodes(vs)...)
				pre = append(pre, setLineno(NewAssignAction(oldcond, cond), lineno))
				post = append(post, setLineno(NewIfAction(
					oldcond,
					NewHavocAction(applyNB(when, checkVarsToNodes(vs)...)),
				), lineno))
			}
			if when.Name == "l2s_whenprev" {
				post = append(post, setLineno(NewIfAction(
					cond,
					NewHavocAction(applyNB(when, checkVarsToNodes(vs)...)),
				), lineno))
			}
		}
		for _, when := range sortedWhens {
			condVal, ok := when.Body.(*Cond)
			if !ok {
				panic(fmt.Sprintf("whenEventsFunc: when binder %s has non-Cond body type %T", when.Name, when.Body))
			}
			post = append(post,
				setLineno(NewAssumeAction(forall(when.Variables,
					&LogicImplies{
						T1: condVal.T1,
						T2: &Eq{T1: applyNB(when, checkVarsToNodes(when.Variables)...), T2: condVal.T2},
					})), lineno))
		}
		return pre, post
	}

	waitEventsFunc := func(waits map[NodeKey]*LogicNamedBinder) []ActionsAction {
		var res []ActionsAction
		sortedWaits := sortNamedBinderMap(waits)
		if !cfg.IsRankingTactic {
			xtracer.Trace("l2s.SharedStep7 waitEventsFunc nWaits=%d", len(sortedWaits))
		}
		for wi, wait := range sortedWaits {
			if !cfg.IsRankingTactic {
				xtracer.Trace("l2s.SharedStep7 waitEventsFunc wait[%d] HASH canon=%s", wi, wait.Canon())
			}
			vs, t := wait.Variables, wait.Body
			waitApp := applyNB(wait, checkVarsToNodes(vs)...)
			rhs := &LogicAnd{Terms: []Expr{
				waitApp,
				&LogicNot{Body: t},
				cfg.ReplaceTemporals(&LogicNot{Body: &LogicGlobally{
					Environ: strPtr(cfg.ProofLabel),
					Body:    Negate(t),
				}}).(Expr),
			}}
			res = append(res, setLineno(NewAssignAction(waitApp, rhs), lineno))
		}
		return res
	}

	var instrStmt func(stmt ActionsAction) ActionsAction
	instrStmt = func(stmt ActionsAction) ActionsAction {
		// H7 / Python ivy_l2s.py:1127-1131: if a CallAction's returns
		// include any monitored symbol (in symprops, symwhens, or symwaits),
		// split the call so the assignment is a separate statement.
		if call, ok := stmt.(*LogicCallAction); ok {
			callArgs := call.ActionArgs()
			if len(callArgs) > 1 {
				returns := callArgs[1:] // []lg.Expr of CallAction.ActualReturns
				monitored := false
				for ri, r := range returns {
					// Extract symbol from return, handling both bare
					// Const and Apply(Const, terms) forms. Python checks
					// `sym in symprops` using structural equality on
					// Const objects; Go uses NodeKey (Sexp) for matching.
					var k NodeKey
					var symName string
					switch v := r.(type) {
					case *Const:
						k = v.Sexp()
						symName = v.Name
					case *Apply:
						if c, ok := v.Func.(*Const); ok {
							k = c.Sexp()
							symName = c.Name
						}
					}
					_, inSP := symprops[k]
					_, inSWh := symwhens[k]
					_, inSWa := symwaits[k]
					if !cfg.IsRankingTactic && xtracer.Enabled {
						xtracer.Trace("l2s.SharedStep7 instrStmt.monitor return[%d] type=%s name=%s inSP=%s inSWh=%s inSWa=%s",
							ri, ShortTypeName(r), symName, checkPyBool(inSP), checkPyBool(inSWh), checkPyBool(inSWa))
					}
					if k != "" && (inSP || inSWh || inSWa) {
						monitored = true
						break
					}
				}
				if monitored {
					split := call.SplitReturns(call.ActCfg)
					return instrStmt(split)
				}
			}
		}

		args := stmt.ActionArgs()
		newArgs := make([]Expr, len(args))
		for i, a := range args {
			if sub, ok := a.(ActionsAction); ok {
				newArgs[i] = instrStmt(sub)
			} else {
				newArgs[i] = a
			}
		}
		res := stmt.ActionClone(newArgs)

		eventProps := make(map[NodeKey]*LogicNamedBinder)
		eventWhens := make(map[NodeKey]*LogicNamedBinder)
		eventWaits := make(map[NodeKey]*LogicNamedBinder)

		modifiedSyms := Modifies(stmt)
		modSet := make(map[NodeKey]bool, len(modifiedSyms))
		for _, sym := range modifiedSyms {
			modSet[sym.Sexp()] = true
		}
		allDeps := cfg.Dependencies(modSet)
		if !cfg.IsRankingTactic {
			sortedDeps := make([]string, 0, len(allDeps))
			for sym := range allDeps {
				sortedDeps = append(sortedDeps, string(sym))
			}
			sort.Strings(sortedDeps)
			sortedMods := make([]string, 0, len(modSet))
			for sym := range modSet {
				sortedMods = append(sortedMods, string(sym))
			}
			sort.Strings(sortedMods)
			xtracer.Trace("l2s.SharedStep7 instrStmt mods=[%s] deps=[%s]", strings.Join(sortedMods, ","), strings.Join(sortedDeps, ","))
		}
		for k := range allDeps {
			// Python uses defaultdict(list) for symprops/symwhens/symwaits.
			// Bracket access on defaultdict creates an empty-list entry for
			// missing keys. Later, the monitoring check uses `sym in symprops`
			// which finds these entries. Replicate by touching the Go maps.
			if _, ok := symprops[k]; !ok {
				symprops[k] = nil
			}
			if _, ok := symwhens[k]; !ok {
				symwhens[k] = nil
			}
			if _, ok := symwaits[k]; !ok {
				symwaits[k] = nil
			}
			for _, prop := range symprops[k] {
				eventProps[prop.Sexp()] = prop
			}
			for _, when := range symwhens[k] {
				eventWhens[when.Sexp()] = when
			}
			for _, wait := range symwaits[k] {
				eventWaits[wait.Sexp()] = wait
			}
		}

		preEvents, postEvents := propEventsFunc(eventProps)
		whenPre, whenPost := whenEventsFunc(eventWhens)
		preEvents = append(whenPre, preEvents...)
		postEvents = append(postEvents, whenPost...)
		postEvents = append(postEvents, waitEventsFunc(eventWaits)...)

		res = PrefixAction(res, preEvents)
		res = PostfixAction(res, postEvents)
		CopyFormalsTo(stmt, res)
		return res
	}

	for i, b := range model.Bindings {
		newStmt := instrStmt(b.Action.Stmt)
		model.Bindings[i] = b.CloneAction(b.Action.CloneStmt(newStmt))
	}
}

// SharedStep8_PatchExports patches exported actions with tableau axioms.
// If cfg.Postconds is non-nil (ranking mode), uses resetW instead of assumeWAxioms
// and attaches postconds to each exported action.
func SharedStep8_PatchExports(cfg *InstrumentationConfig, model *NormalProgram) {
	calls := make(map[string]bool)
	for _, c := range model.Calls {
		calls[c] = true
	}

	if model.Postconds == nil && cfg.Postconds != nil {
		model.Postconds = make(map[string][]*LabeledFormula)
	}

	for i, b := range model.Bindings {
		if !calls[b.Name] {
			continue
		}
		// Python ivy_l2s.py:1228-1231: add l2s_d for all non-finite-sort inputs.
		var addParamsToD []ActionsAction
		for _, p := range b.Action.Inputs {
			if p.CSort != nil && !cfg.FiniteSorts[SortName(p.CSort)] {
				addParamsToD = append(addParamsToD,
					setLineno(NewAssignAction(checkMustApply(L2SD(p.CSort), p), True), cfg.Lineno))
			}
		}

		var stmtParts []ActionsAction
		stmtParts = append(stmtParts, addParamsToD...)
		stmtParts = append(stmtParts, cfg.AssumeGAxioms...)
		stmtParts = append(stmtParts, cfg.AssumeWhenAxioms...)
		if cfg.Postconds != nil {
			// Ranking mode: use reset_w instead of assume_w_axioms
			stmtParts = append(stmtParts, cfg.ResetW...)
		} else {
			// L2S mode: use assume_w_axioms
			stmtParts = append(stmtParts, cfg.AssumeWAxioms...)
		}
		stmtParts = append(stmtParts, b.Action.Stmt)
		stmtParts = append(stmtParts, cfg.AddConstsToD...)

		newStmt := setLineno(ConcatActions(stmtParts...), cfg.Lineno)
		CopyFormalsTo(b.Action.Stmt, newStmt)
		model.Bindings[i] = b.CloneAction(b.Action.CloneStmt(newStmt))

		if cfg.Postconds != nil {
			model.Postconds[b.Name] = cfg.Postconds
		}
	}
}

// SharedStep11_ReplaceNamedBinders replaces named binders with fresh relation constants.
func SharedStep11_ReplaceNamedBinders(cfg *InstrumentationConfig, model *NormalProgram, modPass func(string, func(Node) Node)) {
	namedBinders := collectAllNamedBinders(model)
	if !cfg.IsRankingTactic {
		// Sorted trace for diagnostics only
		keys := make([]string, 0, namedBinders.Len())
		for k := range namedBinders.All() {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v, _ := namedBinders.Get2(k)
			xtracer.Trace("l2s.SharedStep11 namedBinders key=%s count=%d", k, len(v))
		}
	}

	// Ensure _old_l2s_g is consistent with l2s_g
	// Match Python defaultdict auto-vivification
	if _, ok := namedBinders.Get2("l2s_g"); !ok {
		namedBinders.Set("l2s_g", nil)
	}
	l2sGBinders, _ := namedBinders.Get2("l2s_g")
	var oldL2sG []*LogicNamedBinder
	for _, b := range l2sGBinders {
		oldL2sG = append(oldL2sG,
			&LogicNamedBinder{Name: "_old_l2s_g", Variables: b.Variables, Environ: b.Environ, Body: b.Body})
	}
	namedBinders.Set("_old_l2s_g", oldL2sG)

	subs := make(map[string]Expr)
	// C5: also build a string-keyed inverse map for the trace hook.
	// Maps fresh-const-name → original-binder-key string. The renaming
	// hook later builds the reverse for trace display.
	if cfg.Subs == nil {
		cfg.Subs = make(map[string]string)
	}
	// RSubs/FullSubs: used by extractJusticePredMap to navigate
	// l2s_progress_invar formulas through the nonce substitution.
	// Python: rsubs = dict((x,y) for (y,x) in subs.items())
	cfg.RSubs = make(map[string]*LogicNamedBinder)
	// Python iterates named_binders.items() in insertion order (Python 3.7+).
	// InsMap preserves insertion order, matching Python's dict.
	for k, binders := range namedBinders.All() {
		for i, b := range binders {
			freshName := fmt.Sprintf("%s_%d", k, i)
			// Key subs by Sexp (structural canonical form), matching the dedup
			// in collectAllNamedBinders. Python ivy_l2s.py:1432-1436 keys subs
			// by the NamedBinder OBJECT itself (struct eq via recstruct
			// __hash__/__eq__). String (PrettyFmla) drops variable sort
			// annotations and would collapse 4 sort-distinct entries into one,
			// causing ReplaceNamedBindersAst to substitute several distinct
			// binders with the same fresh constant.
			//
			// cfg.Subs (the inverse trace-display map) and the trace text
			// remain keyed/displayed by b.String() so trace output continues
			// to match Python's str(b) format.
			subs[string(b.Sexp())] = NewConst(freshName, b.NodeSort())
			cfg.Subs[freshName] = b.String()
			cfg.RSubs[freshName] = b
			if !cfg.IsRankingTactic {
				xtracer.Trace("l2s.SharedStep11 sub freshName=%s binderKey=%s", freshName, b.String())
			}
		}
	}
	// FullSubs: the binder.Sexp() → nonce Const map for forward resolution.
	cfg.FullSubs = subs

	modPass("ReplaceNamedBindersAst", func(n Node) Node {
		return ReplaceNamedBindersAst(n, subs)
	})

	// Reestablish formals invariant
	for _, b := range model.Bindings {
		b.Action.Stmt.SetFormalParams(b.Action.Inputs)
		b.Action.Stmt.SetFormalReturns(b.Action.Outputs)
	}
}

// SharedStep12_BuildGoal builds the new goal with M |= true as conclusion.
// Python ivy_l2s.py:1461: conc = ivy_ast.TemporalModels(model, lg.And())
// The model parameter is the l2s-modified NormalProgram, not tm.Model (the original).
func SharedStep12_BuildGoal(acfg *AstConfig, goal *LabeledFormula, goals []*LabeledFormula, prems []Node, model Node) ([]*LabeledFormula, error) {
	// Python ivy_l2s.py:1491: lg.And() — empty And is True but canons as "(and)".
	newConc := acfg.NewTemporalModels(model, &LogicAnd{})

	var nonTemporalPrems []Node
	for _, p := range prems {
		if lf, ok := p.(*LabeledFormula); ok && lf.IsTemporal() {
			continue
		}
		nonTemporalPrems = append(nonTemporalPrems, p)
	}

	newGoal := CloneGoal(acfg, goal, nonTemporalPrems, newConc)

	result := make([]*LabeledFormula, len(goals))
	result[0] = newGoal
	copy(result[1:], goals[1:])
	return result, nil
}

// BuildAddConstsToD builds the addConstsToD action list.
func BuildAddConstsToD(mod *Module, uninterpretedSorts []Sort, lineno Location) []ActionsAction {
	var addConstsToD []ActionsAction
	if mod != nil && mod.Sig != nil {
		for _, s := range uninterpretedSorts {
			for _, sym := range insertionOrderSymbols(mod) {
				if sym.CSort != nil && SortEqual(sym.CSort, s) {
					addConstsToD = append(addConstsToD,
						setLineno(NewAssignAction(checkMustApply(L2SD(s), sym), True), lineno))
				}
			}
		}
	}
	return addConstsToD
}

// insertionOrderSymbols returns module symbols in insertion order (matching
// Python's ilg.sig.symbols.values()). Sig.Symbols is an InsMap that
// preserves insertion order and deduplicates by key automatically.
func insertionOrderSymbols(mod *Module) []*Const {
	if mod == nil || mod.Sig == nil || mod.Sig.Symbols.Len() == 0 {
		return nil
	}
	result := make([]*Const, 0, mod.Sig.Symbols.Len())
	for name, entry := range mod.Sig.Symbols.All() {
		result = append(result, NewConst(name, entry.Sort))
	}
	return result
}

// BuildDefnDeps builds the definition dependency map from a module.
// H12 / Python ivy_l2s.py:159-166: also include premise definitions
// from the goal (premises with IsDefinition == true), not just the
// module-level definitions.
func BuildDefnDeps(mod *Module, goalPrems ...Node) map[NodeKey][]NodeKey {
	defnDeps := make(map[NodeKey][]NodeKey)
	addEq := func(formula Expr) {
		// DropUniversals already applied by caller (matching Python ivy_l2s.py:181).
		// Python Definition inherits from Eq, so isinstance(f, Eq) is True
		// for Definition. Go has separate types, so handle both.
		var t1, t2 Expr
		switch ff := formula.(type) {
		case *Eq:
			t1, t2 = ff.T1, ff.T2
		case *LogicDefinition:
			t1, t2 = ff.Lhs, ff.Rhs
		default:
			return
		}
		// Python: defn_deps[sym].append(fml.args[0].rep)
		// .rep works for both Apply(Const,args) and bare Const.
		var lhsKey NodeKey
		switch lhs := t1.(type) {
		case *Apply:
			if c, ok := lhs.Func.(*Const); ok {
				lhsKey = c.Sexp()
			}
		case *Const:
			lhsKey = lhs.Sexp()
		default:
			panicf("how to handle t1=%T here?; canon=%v", t1, t1.Canon())
		}
		if lhsKey != "" {
			// Python ivy_l2s.py:183: for sym in iu.unique(ilu.symbols_ilu_ast(fml.args[1])):
			for _, x := range UsedSymbolsInOrderAst(t2) {
				if sym, ok := x.(*Const); ok {
					defnDeps[sym.Sexp()] = append(defnDeps[sym.Sexp()], lhsKey)
				} else {
					panicf("how to handle x=%T here? x=%v", x, x.Canon())
				}
			}
		}
	}
	// Python ivy_l2s.py:179: _all_defns = list(prover.definitions.values()) + prem_defns
	// Merge module definitions + premise definitions into a single list with
	// unified counter and modDefn label, matching Python's trace output.
	var allDefnExprs []Expr
	if mod != nil {
		for _, defn := range mod.Definitions {
			if e, ok := defn.Formula.(Expr); ok {
				// Python reads from prover.definitions which were normalized
				// via normalize_goal (ivy_proof.py:53). Match that here.
				allDefnExprs = append(allDefnExprs, NormalizeOps(e))
			}
		}
	}
	// H12: include user-supplied definition premises from the goal.
	for _, p := range goalPrems {
		lf, ok := p.(*LabeledFormula)
		if !ok || !lf.IsDefinition {
			continue
		}
		if e, ok := lf.Formula.(Expr); ok {
			allDefnExprs = append(allDefnExprs, NormalizeOps(e))
		}
	}
	for di, e := range allDefnExprs {
		// Python ivy_l2s.py:181: fml = ilg.drop_universals(defn.formula)
		e = IvyDropUniversals(e)
		xtracer.Trace("l2s.BuildDefnDeps modDefn[%d] HASH canon=%s", di, e.Canon())
		addEq(e)
	}
	// Trace the resulting defnDeps map
	{
		keys := make([]string, 0, len(defnDeps))
		for k := range defnDeps {
			keys = append(keys, string(k))
		}
		sort.Strings(keys)
		for _, k := range keys {
			nk := NodeKey(k)
			vals := make([]string, len(defnDeps[nk]))
			for i, v := range defnDeps[nk] {
				vals[i] = string(v)
			}
			xtracer.Trace("l2s.BuildDefnDeps result dep[%s] -> [%s]", k, strings.Join(vals, ","))
		}
	}
	return defnDeps
}

// BuildDependenciesFunc builds a dependency closure function from defnDeps.
func BuildDependenciesFunc(defnDeps map[NodeKey][]NodeKey) func(map[NodeKey]bool) map[NodeKey]bool {
	return func(syms map[NodeKey]bool) map[NodeKey]bool {
		result := make(map[NodeKey]bool)
		var stack []NodeKey
		for s := range syms {
			stack = append(stack, s)
		}
		for len(stack) > 0 {
			s := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if result[s] {
				continue
			}
			result[s] = true
			stack = append(stack, defnDeps[s]...)
		}
		return result
	}
}

// FindTemporalModels looks through the goal formula for a TemporalModels node.
func FindTemporalModels(goal *LabeledFormula) *TemporalModels {
	return checkFindTemporalModels(goal)
}

// ExtractNormalProgram extracts a NormalProgram from a module.
func ExtractNormalProgram(m *Module) *NormalProgram {
	return extractNormalProgram(m)
}

// (CloneGoalWithASTConc was deleted — callers should use proof.CloneGoal
// directly, which now accepts ast.Node for conc.)
