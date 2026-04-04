// Package bmc implements bounded model checking (BMC) for Ivy.
//
// It provides a small BMC orchestrator that uses the analysis graph (art)
// and interpreter (interp) packages to unroll the system transitions
// up to a given bound, checking invariants at each step.
//
// Ported from ivy_bmc.py.
package bmc

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/art"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	mod "github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/trace"
)

const checkPrecondTrue = true
const checkPrecondFalse = false

// BMCResult holds the outcome of a BMC check.
type BMCResult struct {
	// Found is true if a counterexample was found.
	Found bool
	// Depth is the depth at which the counterexample was found (-1 if not found).
	Depth int
	// Trace is the counterexample trace (nil if not found).
	Trace *trace.TraceBase
	// Message describes the result.
	Message string
}

// Config holds the configuration for a BMC run.
type Config struct {
	// NSteps is the number of BMC steps (bound).
	NSteps int
	// NUnroll is the loop unroll count (nil for no unrolling).
	NUnroll *int
	// Module is the Ivy module to check.
	Module *mod.Module
	// Logger receives diagnostic messages. If nil, messages are discarded.
	Logger func(string)
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(mod *mod.Module, nSteps int) *Config {
	return &Config{
		NSteps: nSteps,
		Module: mod,
	}
}

func (c *Config) log(format string, args ...interface{}) {
	if c.Logger != nil {
		c.Logger(fmt.Sprintf(format, args...))
	}
}

// CheckIsolate performs bounded model checking on an Ivy mod.
//
// It unrolls the system up to nSteps transitions and checks whether
// the module's conjectures hold at each step. If a counterexample is
// found, it returns a BMCResult with Found=true.
//
// Parameters:
//   - nSteps: the number of BMC steps (depth bound)
//   - nUnroll: optional loop unrolling count (nil to skip)
//
// This corresponds to ivy_bmc.check_isolate.
func CheckIsolate(cfg *Config) *BMCResult {
	if cfg == nil {
		return &BMCResult{Message: "nil config"}
	}
	mod := cfg.Module
	if mod == nil {
		return &BMCResult{Message: "nil module"}
	}
	nSteps := cfg.NSteps

	// If unrolling is requested, duplicate the actions with unrolled loops.
	var oldActions *iu.InsMap[string, mod.Action]
	if cfg.NUnroll != nil {
		oldActions = mod.Actions
		mod.Actions = iu.NewInsMap[string, mod.Action]()
		for name, act := range oldActions.All() {
			mod.Actions.Set(name, UnrollAction(act, *cfg.NUnroll))
		}
	}
	defer func() {
		if oldActions != nil {
			mod.Actions = oldActions
		}
	}()

	// Build the step action (env action / nondeterministic choice).
	stepAction := EnvAction(mod)

	// Build the conjecture condition.
	conj := BuildConjecture(mod)

	// Build the dual (negated) conjecture for checking.
	dualConj := DualClauses(conj)

	// Create the analysis graph.
	ag := art.NewAnalysisGraph(mod)

	// Add initial state.
	initClauses := mod.TrueClauses(actions.EmptyAnnotation{})
	initState := art.NewState(mod, initClauses)
	ag.Add(initState, nil)
	post := initState

	// Execute the initialize action if present.
	if initAct, ok := mod.Actions.Get2("initialize"); ok {
		if act, ok2 := initAct.(actions.Action); ok2 {
			initPost, err := ag.Execute(checkPrecondTrue, act, nil, nil, "initialize")
			if err != nil {
				return &BMCResult{
					Found:   false,
					Message: fmt.Sprintf("initialize action failed: %v", err),
				}
			}
			if initPost != nil {
				post = initPost
			}
		}
	}

	// BMC loop.
	for n := 0; n <= nSteps; n++ {
		cfg.log("Checking invariants at depth %d...", n)

		// Check conjectures at this depth.
		result := trace.CheckFinalCond(ag, post, dualConj, nil, true)
		if result != nil {
			msg := fmt.Sprintf("BMC with bound %d found a counter-example", n)
			cfg.log("%s", msg)
			return &BMCResult{
				Found:   true,
				Depth:   n,
				Trace:   result,
				Message: msg,
			}
		}

		// Execute one step.
		stepPost, err := ag.Execute(checkPrecondFalse, stepAction, nil, nil, "")
		if err != nil {
			return &BMCResult{
				Found:   false,
				Message: fmt.Sprintf("step execution failed at depth %d: %v", n, err),
			}
		}
		post = stepPost

		// Safety check (assertion failures in the step).
		// The fail_expr extracts precondition-violation conditions from the
		// step action's update. ActionFailure swaps TR and Pre, so we check
		// whether the precondition violation (now the TR) is reachable.
		if post != nil && stepAction != nil {
			failUpdate := computeFailUpdate(stepAction, mod)
			if failUpdate != nil && failUpdate.TR != nil && !failUpdate.TR.IsFalse() {
				failClauses := failUpdate.TR
				failState := art.NewState(mod, failClauses)
				failState.Pred = post.Pred
				safetyResult := trace.CheckFinalCond(ag, failState, mod.TrueClauses(nil), nil, true)
				if safetyResult != nil {
					msg := fmt.Sprintf("BMC with bound %d found an assertion failure", n)
					cfg.log("%s", msg)
					return &BMCResult{
						Found:   true,
						Depth:   n,
						Trace:   safetyResult,
						Message: msg,
					}
				}
			}
		}
	}

	return &BMCResult{
		Found:   false,
		Depth:   -1,
		Message: fmt.Sprintf("BMC with bound %d found no counter-example", nSteps),
	}
}

// EnvAction creates the environment step action from a module's public actions.
// It produces a nondeterministic choice among all public actions.
func EnvAction(mod *mod.Module) actions.Action {
	if mod == nil {
		return actions.NewSequence()
	}
	var branches []lg.Expr
	for name := range mod.PublicActions.All() {
		act, ok := mod.Actions.Get2(name)
		if !ok {
			continue
		}
		action, ok := act.(actions.Action)
		if !ok {
			continue
		}
		branches = append(branches, action)
	}
	if len(branches) == 0 {
		return actions.NewSequence()
	}
	return actions.NewEnvAction(branches...)
}

// BuildConjecture combines a module's conjectures into a single Clauses.
func BuildConjecture(mod *mod.Module) *mod.Clauses {
	if mod == nil || len(mod.LabeledConjs) == 0 {
		return mod.TrueClauses(nil)
	}
	var fmlas []lg.Expr
	for _, lc := range mod.LabeledConjs {
		if lc.Formula != nil {
			fmlas = append(fmlas, lc.Formula.(lg.Expr))
		}
	}
	if len(fmlas) == 0 {
		return mod.TrueClauses(nil)
	}
	return mod.NewClauses(fmlas, nil, nil)
}

// DualClauses computes the negation of a conjecture for checking.
// The dual conjecture is satisfiable iff the original is not valid.
func DualClauses(conj *mod.Clauses) *mod.Clauses {
	if conj == nil || len(conj.Fmlas) == 0 {
		return mod.TrueClauses(nil)
	}
	// Negate: each conjunct becomes a disjunct of its negation.
	var negFmlas []lg.Expr
	for _, f := range conj.Fmlas {
		neg := &lg.Not{Body: f}
		negFmlas = append(negFmlas, neg)
	}
	or, err := lg.NewOr(negFmlas...)
	if err != nil {
		// Fallback
		return mod.NewClauses(negFmlas, nil, nil)
	}
	return mod.NewClauses([]lg.Expr{or}, nil, nil)
}

// UnrollAction is a placeholder for loop unrolling.
// The full implementation would traverse the action AST and unroll
// While loops up to n iterations.
func UnrollAction(act actions.Action, n int) actions.Action {
	// Stub: return the action unchanged.
	// Full implementation would unroll loops.
	return act
}

// computeFailUpdate computes the failure update for an action.
// This corresponds to Python's fail_expr/fail_action which extracts
// assertion-violation conditions by computing the action's update and
// then calling action_failure (which swaps TR and Pre).
//
// If the action implements the Updater interface, we compute its update
// and return action_failure(update). Otherwise returns nil.
func computeFailUpdate(action actions.Action, mod *mod.Module) *actions.Update {
	update := actions.GetUpdateForArt(action, mod, nil)
	if update == nil {
		return nil
	}
	// action_failure swaps TR and Pre: the precondition (failure condition)
	// becomes the new TR, and Pre becomes True (always satisfiable).
	return actions.ActionFailure(update)
}
