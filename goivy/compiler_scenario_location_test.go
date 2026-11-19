package goivy

import (
	"os"
	"path/filepath"
	"testing"
)

// Python ARGSetup.scenario builds AssumeAction(find_symbol(place)) with no
// lineno, so check prints "(internal) assumption" for it (test/scen2.ivy).
func TestScenarioPlaceAssumesHaveNoLocation(t *testing.T) {
	fn := filepath.Join(examplesDir(), "test", "scen2.ivy")
	if _, err := os.Stat(fn); err != nil {
		t.Skip(err)
	}
	mod, err := IvyLoadFile(fn)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, act := range mod.Actions.All() {
		for _, sub := range act.(ActionsAction).IterSubactions() {
			as, ok := sub.(*LogicAssumeAction)
			if !ok {
				continue
			}
			c, ok := as.Formula.(*Const)
			if !ok || (c.Name != "s0" && c.Name != "s1") {
				continue
			}
			n++
			if as.HasLineno() {
				t.Errorf("assume %s has location %v", c.Name, as.GetLineno())
			}
		}
	}
	if n == 0 {
		t.Fatal("no scenario place assumes")
	}
}
