package goivy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// traceInitAfter is a debug trace that follows the after-init actions
// through CreateIsolate (Python: trace_init_after in ivy_isolate.py).
// For each action whose name contains "init[after]", in sorted name
// order, it records the formal parameter count and the argument count of
// each call atom in the body.
func traceInitAfter(mod *Module, stage string) {
	var names []string
	for name := range mod.Actions.All() {
		if strings.Contains(name, "init[after]") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		a := mod.Actions.Get(name)
		var calls []string
		for _, s := range a.IterSubactions() {
			if ShortTypeName(s) == "CallAction" {
				calls = append(calls, fmt.Sprint(len(s.Args()[0].Args())))
			}
		}
		xtracer.Trace("isolate.trace_init_after stage=%s name=%s nparams=%d calls=[%s]",
			stage, name, len(a.GetFormalParams()), strings.Join(calls, ","))
	}
}
