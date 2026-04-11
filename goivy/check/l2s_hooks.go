// l2s_hooks.go implements the L2S diagnostic trace hooks invoked by the
// trace formatter in check.go after a checker fails. These mirror Python's
// renaming_hook (ivy_l2s.py:1318-1319) and auto_hook (ivy_l2s.py:1333-1506).
package check

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// TraceHookFn is the function type stored in ast.LabeledFormula.TraceHook
// (and propagated to module.Module.TraceHook) by L2S tactics. It is invoked
// by the trace formatter in check.go after constructing a MatchHandler from
// the failing checker. Mirrors Python's goal.trace_hook closure
// (ivy_l2s.py:1311-1313).
type TraceHookFn func(handler *MatchHandler, fcs []Checker)

// applyRenamingToHandler applies the inverse of subs to the MatchHandler's
// Lines. subs maps {fresh-const-name → original-binder-key}; we rewrite
// occurrences in each line. Mirrors Python ivy_l2s.py:1318-1319 renaming_hook.
func applyRenamingToHandler(handler *MatchHandler, subs map[string]string) {
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

// applyAutoDiagnosticsToHandler dispatches to the auto-failure diagnostic
// printer based on which checker failed. Mirrors Python ivy_l2s.py:1333-1506
// auto_hook.
func applyAutoDiagnosticsToHandler(
	handler *MatchHandler,
	fcs []Checker,
	tasks map[string]map[string]*lg.Eq,
	triggers map[string]map[string]*lg.Eq,
) {
	if handler == nil {
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
	name := lfName(lf)
	diagnoseAutoFailure(name, tasks, triggers, lf)
}

// diagnoseAutoFailure prints diagnostic information based on the failed
// invariant name. Mirrors the dispatch table in Python's auto_hook
// (ivy_l2s.py:1333-1506).
func diagnoseAutoFailure(name string, tasks, triggers map[string]map[string]*lg.Eq,
	lf *ast.LabeledFormula) {

	switch {
	case strings.HasPrefix(name, "l2s_created"):
		sfx := name[len("l2s_created"):]
		fmt.Printf("\n\nFailed to prove that work_created%s is finite by induction.\n", sfx)
		if task, ok := tasks[sfx]; ok {
			if wc := task["work_created"]; wc != nil {
				fmt.Printf("work_created%s definition: %v\n", sfx, wc)
			}
		}

	case strings.HasPrefix(name, "l2s_needed_when_start"):
		sfx := name[len("l2s_needed_when_start"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is a subset of work_created%s when the start condition has occurred.\n", sfx, sfx)

	case strings.HasPrefix(name, "l2s_work_preserved"):
		sfx := name[len("l2s_work_preserved"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is preserved.\n", sfx)

	case strings.HasPrefix(name, "l2s_needed_are_frozen"):
		sfx := name[len("l2s_needed_are_frozen"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is preserved.\n", sfx)

	case strings.HasPrefix(name, "l2s_progress_made"):
		sfx := name[len("l2s_progress_made"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s decreases when a helpful transition occurs\n", sfx)

	case strings.HasPrefix(name, "l2s_sched_stable"):
		sfx := name[len("l2s_sched_stable"):]
		fmt.Printf("\n\nFailed to prove that work_helpful%s is stable until helpful transition occurs\n", sfx)

	case strings.HasPrefix(name, "l2s_not_all_done"):
		var rankNames []string
		for sfx, task := range tasks {
			if task["work_needed"] != nil {
				rankNames = append(rankNames, "work_needed"+sfx)
			}
		}
		fmt.Printf("The ranking(s) %s have become empty, but termination has not occurred.\n",
			strings.Join(rankNames, " and "))

	case strings.HasPrefix(name, "l2s_sched_exists"):
		var rankNames []string
		for sfx, task := range tasks {
			if task["work_helpful"] != nil {
				rankNames = append(rankNames, "work_helpful"+sfx)
			}
		}
		fmt.Printf("The helpful set(s) %s have become empty, but termination has not occurred.\n",
			strings.Join(rankNames, " and "))
	}
}

// lfName extracts the name from a labeled formula's label.
func lfName(lf *ast.LabeledFormula) string {
	if lf == nil || lf.Label == nil {
		return ""
	}
	if c, ok := lf.Label.(*lg.Const); ok {
		return c.Name
	}
	return fmt.Sprint(lf.Label)
}
