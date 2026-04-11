// l2s_hooks.go implements trace hooks for L2S diagnostics.
// Moved from l2s/hooks.go to break the import cycle l2s → check → l2s.
// Ported from Python ivy_l2s.py: trace_hook, renaming_hook, auto_hook,
// temporal_and_l2s, ls2_g_to_globally.
package check

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/l2s"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/trace"
)

// L2STraceHook is the L2S trace hook. It finds the loop start point by
// searching for l2s_saved=true in the trace states.
// Corresponds to Python trace_hook (ivy_l2s.py:97-106).
func L2STraceHook(tr *trace.TraceBase, fcs []Checker) *trace.TraceBase {
	for idx, state := range tr.TraceStates {
		if state.State == nil || state.State.Clauses == nil {
			continue
		}
		for _, f := range state.State.Clauses.Fmlas {
			eq, ok := f.(*lg.Eq)
			if !ok {
				continue
			}
			s1 := fmt.Sprint(eq.T1)
			s2 := fmt.Sprint(eq.T2)
			if s1 == "l2s_saved" && s2 == "true" {
				loopIdx := idx
				if idx > 0 {
					loopIdx = idx - 1
				}
				tr.TraceStates[loopIdx].LoopStart = true
				return tr
			}
		}
	}
	fmt.Println("failed to find loop start!")
	return tr
}

// L2SRenamingHook applies the reverse substitution map to a trace for
// readable symbol names.
// Corresponds to Python renaming_hook (ivy_l2s.py:1318-1319).
func L2SRenamingHook(subs map[string]string, tr *trace.TraceBase, fcs []Checker) *trace.TraceBase {
	// Build reverse map
	rsubs := make(map[string]string)
	for k, v := range subs {
		rsubs[v] = k
	}
	return tr.Rename(rsubs)
}

// L2SAutoHookConfig holds the configuration for the auto_hook diagnostic.
type L2SAutoHookConfig struct {
	Tasks    map[string]map[string]*lg.Eq // sfx -> name -> definition
	Triggers map[string]map[string]*lg.Eq // sfx -> name -> definition
	Subs     map[string]string            // symbol renaming map
}

// L2SAutoHook is the l2s_auto5 diagnostic trace hook. It identifies which
// invariant failed and prints diagnostic information.
// Corresponds to Python auto_hook (ivy_l2s.py:1333-1506).
func L2SAutoHook(cfg *L2SAutoHookConfig, tr *trace.TraceBase, fcs []Checker) *trace.TraceBase {
	if cfg == nil {
		return tr
	}

	// Apply renaming
	tr = L2SRenamingHook(cfg.Subs, tr, fcs)
	tr.PP = l2s.L2SGToGlobally

	// Figure out which property failed
	var failedFC Checker
	for _, fc := range fcs {
		if fc.Failed() {
			failedFC = fc
			break
		}
	}
	if failedFC == nil {
		return tr
	}
	lf := failedFC.GetLF()
	if lf == nil {
		return tr
	}

	name := l2sLfName(lf)
	l2sDiagnoseAutoFailure(name, cfg.Tasks, cfg.Triggers, lf, tr)

	return tr
}

// l2sDiagnoseAutoFailure prints diagnostic information based on the failed invariant name.
func l2sDiagnoseAutoFailure(name string, tasks, triggers map[string]map[string]*lg.Eq,
	lf *ast.LabeledFormula, tr *trace.TraceBase) {

	switch {
	case strings.HasPrefix(name, "l2s_created"):
		sfx := name[len("l2s_created"):]
		fmt.Printf("\n\nFailed to prove that work_created%s is finite by induction.\n", sfx)
		if task, ok := tasks[sfx]; ok {
			if wc := task["work_created"]; wc != nil {
				fmt.Printf("work_created%s definition: %v\n", sfx, wc)
			}
		}
		if tr != nil {
			tr.HiddenSymbols = L2STemporalAndL2SFilter
		}

	case strings.HasPrefix(name, "l2s_needed_when_start"):
		sfx := name[len("l2s_needed_when_start"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is a subset of work_created%s when the start condition has occurred.\n", sfx, sfx)
		if tr != nil {
			tr.HiddenSymbols = L2STemporalAndL2SFilter
		}

	case strings.HasPrefix(name, "l2s_work_preserved"):
		sfx := name[len("l2s_work_preserved"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is preserved.\n", sfx)
		if tr != nil {
			tr.HiddenSymbols = L2STemporalAndL2SFilter
		}

	case strings.HasPrefix(name, "l2s_needed_are_frozen"):
		sfx := name[len("l2s_needed_are_frozen"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is preserved.\n", sfx)
		if tr != nil {
			tr.HiddenSymbols = L2STemporalAndL2SFilter
		}

	case strings.HasPrefix(name, "l2s_progress_made"):
		sfx := name[len("l2s_progress_made"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s decreases when a helpful transition occurs\n", sfx)
		if tr != nil {
			tr.HiddenSymbols = L2STemporalAndL2SFilter
		}

	case strings.HasPrefix(name, "l2s_sched_stable"):
		sfx := name[len("l2s_sched_stable"):]
		fmt.Printf("\n\nFailed to prove that work_helpful%s is stable until helpful transition occurs\n", sfx)
		if tr != nil {
			tr.HiddenSymbols = L2STemporalAndL2SFilter
		}

	case strings.HasPrefix(name, "l2s_not_all_done"):
		var rankNames []string
		for sfx, task := range tasks {
			if task["work_needed"] != nil {
				rankNames = append(rankNames, "work_needed"+sfx)
			}
		}
		fmt.Printf("The ranking(s) %s have become empty, but termination has not occurred.\n",
			strings.Join(rankNames, " and "))
		if tr != nil {
			tr.HiddenSymbols = L2STemporalAndL2SFilter
		}

	case strings.HasPrefix(name, "l2s_sched_exists"):
		var rankNames []string
		for sfx, task := range tasks {
			if task["work_helpful"] != nil {
				rankNames = append(rankNames, "work_helpful"+sfx)
			}
		}
		fmt.Printf("The helpful set(s) %s have become empty, but termination has not occurred.\n",
			strings.Join(rankNames, " and "))
		if tr != nil {
			tr.HiddenSymbols = L2STemporalAndL2SFilter
		}
	}
}

// l2sLfName extracts the name from a labeled formula's label.
func l2sLfName(lf *ast.LabeledFormula) string {
	if lf == nil || lf.Label == nil {
		return ""
	}
	if c, ok := lf.Label.(*lg.Const); ok {
		return c.Name
	}
	return fmt.Sprint(lf.Label)
}

// L2STemporalAndL2SFilter is a HiddenSymbols filter function that hides
// L2S monitor symbols from trace display.
// Corresponds to Python temporal_and_l2s (ivy_l2s.py:1321-1323).
func L2STemporalAndL2SFilter(name string) bool {
	return (strings.HasPrefix(name, "l2s") && !strings.HasPrefix(name, "l2s_g")) ||
		strings.HasPrefix(name, "_old_l2s")
}

// applyL2SRenamingToHandler applies the inverse of subs to the
// MatchHandler's Lines. subs maps {fresh-const-name → original-binder-key};
// we build the reverse and rewrite occurrences in each line. This is the
// MatchHandler counterpart of L2SRenamingHook (which operates on
// *trace.TraceBase).
//
// C5 / Python ivy_l2s.py:1318-1319 renaming_hook.
func applyL2SRenamingToHandler(handler *MatchHandler, subs map[string]string) {
	if handler == nil || len(subs) == 0 {
		return
	}
	rsubs := make(map[string]string, len(subs))
	for k, v := range subs {
		rsubs[k] = v
	}
	for i, line := range handler.Lines {
		out := line
		for fresh, orig := range rsubs {
			out = strings.ReplaceAll(out, fresh, orig)
		}
		handler.Lines[i] = out
	}
}

// applyL2SAutoDiagnostics dispatches to the auto-failure diagnostic
// printer based on which checker failed. This is the MatchHandler
// counterpart of L2SAutoHook (which operates on *trace.TraceBase).
//
// C5 / Python ivy_l2s.py:1333-1506 auto_hook.
func applyL2SAutoDiagnostics(handler *MatchHandler, fcs []Checker, data *l2s.L2STraceHookData) {
	if handler == nil || data == nil {
		return
	}
	// Find the failing checker.
	var failedFC Checker
	for _, fc := range fcs {
		if fc != nil && fc.Failed() {
			failedFC = fc
			break
		}
	}
	if failedFC == nil {
		return
	}
	lf := failedFC.GetLF()
	if lf == nil {
		return
	}
	name := l2sLfName(lf)
	// l2sDiagnoseAutoFailure expects a *trace.TraceBase for setting
	// HiddenSymbols; passing nil suppresses that side effect. The
	// MatchHandler<->TraceBase bridge is a follow-up to Phase 7.
	l2sDiagnoseAutoFailure(name, data.Tasks, data.Triggers, lf, nil)
}
