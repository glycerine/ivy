// hooks.go implements trace hooks for ranking-based L2S diagnostics.
// Ported from Python ivy_ranking.py: auto_hook (lines 1042-1151).
package ranking

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/check"
	"github.com/glycerine/goivy/l2s"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/trace"
)

// RankingAutoHookConfig holds the configuration for the ranking auto_hook.
type RankingAutoHookConfig struct {
	Tasks    map[string]*Task
	Triggers map[string]*Trigger
	Subs     map[string]string // symbol renaming map
}

// RankingAutoHook is the ranking-specific diagnostic trace hook.
// It identifies which invariant failed and prints diagnostic information.
// Corresponds to Python ivy_ranking.py auto_hook (lines 1042-1151).
func RankingAutoHook(cfg *RankingAutoHookConfig, tr *trace.TraceBase, fcs []check.Checker) *trace.TraceBase {
	if cfg == nil {
		return tr
	}

	// Apply renaming
	tr = l2s.RenamingHook(cfg.Subs, tr, fcs)
	tr.PP = l2s.L2SGToGlobally

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
	diagnoseRankingFailure(name, cfg.Tasks, cfg.Triggers, lf, tr)

	return tr
}

// diagnoseRankingFailure prints diagnostic information based on the failed invariant name.
func diagnoseRankingFailure(name string, tasks map[string]*Task, triggers map[string]*Trigger,
	lf *module.LabeledFormula, tr *trace.TraceBase) {

	switch {
	case strings.HasPrefix(name, "l2s_created"):
		sfx := name[len("l2s_created"):]
		fmt.Printf("\n\nFailed to prove that work_created%s is finite by induction.\n", sfx)
		if task, ok := tasks[sfx]; ok && task.WorkCreated != nil {
			fmt.Printf("work_created%s definition: %v\n", sfx, task.WorkCreated)
		}

	case strings.HasPrefix(name, "l2s_eventually_start"):
		sfx := name[len("l2s_eventually_start"):]
		fmt.Println("The eventually property fails to hold when liveness invariant is false.")
		fmt.Println("i.e. The following invariant does not hold:")
		fmt.Println("|/= ~work_invar -> \u25C7 (work_start)")
		for s, task := range tasks {
			fmt.Printf("work_invar%s : %v\n", s, task.WorkInvar)
			if trig, ok := triggers[s]; ok {
				fmt.Printf("work_start%s : %v\n", s, trig.WorkStart)
			}
		}
		_ = sfx
		fmt.Println("Reminder: auto-generated work_start is the negation of the globally condition.")

	case strings.HasPrefix(name, "l2s_needed_preserved"):
		sfx := name[len("l2s_needed_preserved"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s is preserved.\n", sfx)

	case strings.HasPrefix(name, "l2s_progress"):
		sfx := name[len("l2s_progress"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s decreases when a helpful transition occurs\n", sfx)

	case strings.HasPrefix(name, "l2s_needed_implies_created"):
		sfx := name[len("l2s_needed_implies_created"):]
		fmt.Printf("\n\nFailed to prove that work_needed%s implies work_created%s.\n", sfx, sfx)

	case strings.HasPrefix(name, "l2s_invar"):
		sfx := name[len("l2s_invar"):]
		fmt.Printf("\n\nFailed to establish work_invar%s.\n", sfx)

	case strings.HasPrefix(name, "l2s_sched_stable"):
		sfx := name[len("l2s_sched_stable"):]
		fmt.Printf("\n\nFailed to prove that work_helpful%s is stable until helpful transition occurs\n", sfx)

	case strings.HasPrefix(name, "l2s_not_all_done"):
		var rankNames []string
		for sfx, task := range tasks {
			if task.WorkNeeded != nil {
				rankNames = append(rankNames, "work_needed"+sfx)
			}
		}
		fmt.Printf("The ranking(s) %s have become empty, but termination has not occurred.\n",
			strings.Join(rankNames, " and "))

	case strings.HasPrefix(name, "l2s_sched_exists"):
		var rankNames []string
		for sfx, task := range tasks {
			if task.WorkHelpful != nil {
				rankNames = append(rankNames, "work_helpful"+sfx)
			}
		}
		fmt.Printf("The helpful set(s) %s have become empty, but termination has not occurred.\n",
			strings.Join(rankNames, " and "))
	}
}

// lfName extracts the name from a labeled formula's label.
func lfName(lf *module.LabeledFormula) string {
	if lf == nil || lf.Label == nil {
		return ""
	}
	if c, ok := lf.Label.(*lg.Const); ok {
		return c.Name
	}
	return fmt.Sprint(lf.Label)
}
