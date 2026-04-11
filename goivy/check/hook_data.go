// hook_data.go defines the data structure that the l2s tactic attaches to
// LabeledFormula.TraceHook so the check trace-display code can route to
// the correct trace-formatting function.
//
// Mirrors Python's dynamically-attached `goal.trace_hook` lambda
// (ivy_l2s.py:88, 1311, 1313).
package check

import (
	lg "github.com/glycerine/ivy/goivy/logic"
)

// L2STraceHookKind selects which hook the consumer should run.
type L2STraceHookKind int

const (
	// HookKindNone is the zero value (no hook).
	HookKindNone L2STraceHookKind = iota
	// HookKindFull corresponds to Python's trace_hook (ivy_l2s.py:97-106),
	// used by l2s_full to mark the loop start.
	HookKindFull
	// HookKindRenaming corresponds to Python's renaming_hook (ivy_l2s.py:1318-1319),
	// used by l2s_auto/l2s_auto2/l2s_auto3/l2s_auto4 to apply the inverse
	// substitution map for readable trace symbols.
	HookKindRenaming
	// HookKindAuto corresponds to Python's auto_hook (ivy_l2s.py:1333-1506),
	// used by l2s_auto5 for full diagnostic output identifying the failed
	// invariant and printing relevant work_* state.
	HookKindAuto
)

// L2STraceHookData is the opaque payload stored on
// ast.LabeledFormula.TraceHook by the l2s tactic. The check package
// type-asserts this struct out of the interface{} field to dispatch to
// the appropriate hook function.
type L2STraceHookData struct {
	Kind     L2STraceHookKind
	Subs     map[string]string
	Tasks    map[string]map[string]*lg.Eq
	Triggers map[string]map[string]*lg.Eq
}
