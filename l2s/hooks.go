// hooks.go implements trace hooks for L2S diagnostics.
// Ported from Python ivy_l2s.py: trace_hook, renaming_hook, auto_hook,
// temporal_and_l2s, ls2_g_to_globally.
package l2s

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/check"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/trace"
)

// TraceHook is the L2S trace hook. It finds the loop start point by
// searching for l2s_saved=true in the trace states.
// Corresponds to Python trace_hook (ivy_l2s.py:97-106).
func TraceHook(tr *trace.TraceBase, fcs []check.Checker) *trace.TraceBase {
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

// RenamingHook applies the reverse substitution map to a trace for
// readable symbol names.
// Corresponds to Python renaming_hook (ivy_l2s.py:1318-1319).
func RenamingHook(subs map[string]string, tr *trace.TraceBase, fcs []check.Checker) *trace.TraceBase {
	// Build reverse map
	rsubs := make(map[string]string)
	for k, v := range subs {
		rsubs[v] = k
	}
	return tr.Rename(rsubs)
}

// AutoHookConfig holds the configuration for the auto_hook diagnostic.
type AutoHookConfig struct {
	Tasks    map[string]map[string]*lg.Eq // sfx -> name -> definition
	Triggers map[string]map[string]*lg.Eq // sfx -> name -> definition
	Subs     map[string]string            // symbol renaming map
}

// AutoHook is the l2s_auto5 diagnostic trace hook. It identifies which
// invariant failed and prints diagnostic information.
// Corresponds to Python auto_hook (ivy_l2s.py:1333-1506).
func AutoHook(cfg *AutoHookConfig, tr *trace.TraceBase, fcs []check.Checker) *trace.TraceBase {
	if cfg == nil {
		return tr
	}

	// Apply renaming
	tr = RenamingHook(cfg.Subs, tr, fcs)
	tr.PP = L2SGToGlobally

	// Figure out which property failed
	var failedFC check.Checker
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

	name := lfName(lf)
	diagnoseAutoFailure(name, cfg.Tasks, cfg.Triggers, lf, tr)

	return tr
}

// diagnoseAutoFailure prints diagnostic information based on the failed invariant name.
func diagnoseAutoFailure(name string, tasks, triggers map[string]map[string]*lg.Eq,
	lf *module.LabeledFormula, tr *trace.TraceBase) {

	switch {
	case strings.HasPrefix(name, "l2s_created"):
		sfx := name[len("l2s_created"):]
		fmt.Printf("\n\nFailed to prove that work_created%s is finite by induction.\n", sfx)
		if task, ok := tasks[sfx]; ok {
			if wc := task["work_created"]; wc != nil {
				fmt.Printf("work_created%s definition: %v\n", sfx, wc)
			}
		}
		tr.HiddenSymbols = TemporalAndL2SFilter

	case strings.HasPrefix(name, "l2s_needed_when_start"):
		sfx := name[len("l2s_needed_when_start"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is a subset of work_created%s when the start condition has occurred.\n", sfx, sfx)
		tr.HiddenSymbols = TemporalAndL2SFilter

	case strings.HasPrefix(name, "l2s_work_preserved"):
		sfx := name[len("l2s_work_preserved"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is preserved.\n", sfx)
		tr.HiddenSymbols = TemporalAndL2SFilter

	case strings.HasPrefix(name, "l2s_needed_are_frozen"):
		sfx := name[len("l2s_needed_are_frozen"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is preserved.\n", sfx)
		tr.HiddenSymbols = TemporalAndL2SFilter

	case strings.HasPrefix(name, "l2s_progress_made"):
		sfx := name[len("l2s_progress_made"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s decreases when a helpful transition occurs\n", sfx)
		// In the full version, this would extract helpful predicates from
		// the trace states and display which work_helpful elements are true
		// but work_needed didn't decrease.
		tr.HiddenSymbols = TemporalAndL2SFilter

	case strings.HasPrefix(name, "l2s_sched_stable"):
		sfx := name[len("l2s_sched_stable"):]
		fmt.Printf("\n\nFailed to prove that work_helpful%s is stable until helpful transition occurs\n", sfx)
		tr.HiddenSymbols = TemporalAndL2SFilter

	case strings.HasPrefix(name, "l2s_not_all_done"):
		var rankNames []string
		for sfx, task := range tasks {
			if task["work_needed"] != nil {
				rankNames = append(rankNames, "work_needed"+sfx)
			}
		}
		fmt.Printf("The ranking(s) %s have become empty, but termination has not occurred.\n",
			strings.Join(rankNames, " and "))
		tr.HiddenSymbols = TemporalAndL2SFilter

	case strings.HasPrefix(name, "l2s_sched_exists"):
		var rankNames []string
		for sfx, task := range tasks {
			if task["work_helpful"] != nil {
				rankNames = append(rankNames, "work_helpful"+sfx)
			}
		}
		fmt.Printf("The helpful set(s) %s have become empty, but termination has not occurred.\n",
			strings.Join(rankNames, " and "))
		tr.HiddenSymbols = TemporalAndL2SFilter
	}
}

// TemporalAndL2SFilter is a HiddenSymbols filter function that hides
// L2S monitor symbols from trace display.
// Corresponds to Python temporal_and_l2s (ivy_l2s.py:1321-1323).
func TemporalAndL2SFilter(name string) bool {
	return (strings.HasPrefix(name, "l2s") && !strings.HasPrefix(name, "l2s_g")) ||
		strings.HasPrefix(name, "_old_l2s")
}

// lfName extracts the name from a labeled formula's label.
func lfName(lf *module.LabeledFormula) string {
	if lf == nil || lf.Label == nil {
		return ""
	}
	if c, ok := lf.Label.(*lg.Symbol); ok {
		return c.Name
	}
	return fmt.Sprint(lf.Label)
}
