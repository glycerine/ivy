package core

import "time"

import "reflect"

func sameHook(a, b DD_HFP) bool {
	return a != nil && b != nil && reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

// Cudd_AddHook appends a callback to the selected manager hook list.
func Cudd_AddHook(dd *DdManager, hook DD_HFP, where Cudd_HookType) int32 {
	if dd == nil || dd.closed || hook == nil || where < CUDD_PRE_GC_HOOK || where > CUDD_POST_REORDERING_HOOK {
		return 0
	}
	if dd.hooks == nil {
		dd.hooks = make(map[Cudd_HookType][]DD_HFP)
	}
	for _, existing := range dd.hooks[where] {
		if sameHook(existing, hook) {
			return 0
		}
	}
	dd.hooks[where] = append(dd.hooks[where], hook)
	return 1
}

// Cudd_RemoveHook removes a registered callback.
func Cudd_RemoveHook(dd *DdManager, hook DD_HFP, where Cudd_HookType) int32 {
	if dd == nil || dd.closed || hook == nil {
		return 0
	}
	for i, existing := range dd.hooks[where] {
		if sameHook(existing, hook) {
			list := dd.hooks[where]
			dd.hooks[where] = append(list[:i], list[i+1:]...)
			return 1
		}
	}
	return 0
}

func Cudd_IsInHook(dd *DdManager, hook DD_HFP, where Cudd_HookType) int32 {
	if dd == nil || dd.closed || hook == nil {
		return 0
	}
	for _, existing := range dd.hooks[where] {
		if sameHook(existing, hook) {
			return 1
		}
	}
	return 0
}

func (dd *DdManager) runHooks(where Cudd_HookType, graph string, data any) bool {
	for _, hook := range dd.hooks[where] {
		if hook(dd, graph, data) == 0 {
			return false
		}
	}
	return true
}

// Cudd_BeginBDDReordering starts one BDD reorder lifecycle: it increments the
// call counter and runs the pre-reordering hooks. A false result means a hook
// cancelled the operation; elapsed time is still recorded.
func Cudd_BeginBDDReordering(dd *DdManager, method Cudd_ReorderingType) (time.Time, bool) {
	started := time.Now()
	if dd == nil || dd.closed {
		return started, false
	}
	dd.reorderings++
	if !dd.runHooks(CUDD_PRE_REORDERING_HOOK, "BDD", method) {
		dd.reorderingTime += time.Since(started)
		return started, false
	}
	return started, true
}

// Cudd_EndBDDReordering runs post hooks after a successful BDD reorder and
// records elapsed time whether the algorithm succeeded or failed.
func Cudd_EndBDDReordering(dd *DdManager, started time.Time, success bool) bool {
	if dd == nil || dd.closed {
		return false
	}
	if success {
		success = dd.runHooks(CUDD_POST_REORDERING_HOOK, "BDD", started.UnixNano())
	}
	dd.reorderingTime += time.Since(started)
	return success
}
