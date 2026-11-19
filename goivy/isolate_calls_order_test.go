package goivy

import "testing"

// Python get_calls_mods (ivy_isolate.py) now collects a summarized
// action's unsummarized callees in an insertion-ordered dict (it was a
// set); check_interference prints them with ','.join(callbacks). foo calls
// c2 then c1, so calls["foo"] is c2, c1. goivy used a Go map.
func TestGetCallsModsCalleesInCallOrder(t *testing.T) {
	mkCall := func(name string) *LogicCallAction {
		c := NewCallActionOn(NewActionsConfig(), NewConst(name, TopS))
		c.AstCallee = NewAstConfig().NewAtom(name)
		return c
	}
	for i := 0; i < 30; i++ {
		m := mkModuleWithActions(map[string]ActionsAction{
			"foo": NewSequence(mkCall("c2"), mkCall("c1")),
		})
		calls := make(map[string]*InsMap[string, bool])
		mods := make(map[string]map[string]bool)
		GetCallsModsRec(m, map[string]bool{"foo": true}, "foo", calls, mods)
		var got []string
		for k := range calls["foo"].All() {
			got = append(got, k)
		}
		if len(got) != 2 || got[0] != "c2" || got[1] != "c1" {
			t.Fatalf("run %d: calls[foo] = %v, want [c2 c1]", i, got)
		}
	}
}
