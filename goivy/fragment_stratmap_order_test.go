package goivy

import (
	"strings"
	"testing"
)

// Python report_interp_over_var walks strat_map (an insertion-ordered
// dict) and the last variable sharing the node wins. With many variables
// on one node, goivy must name the same (last-inserted) variable every
// run, not whichever a Go map visits last.
func TestReportInterpOverVarUsesLastInserted(t *testing.T) {
	S := proofMkSort("S")
	node := NewUFNode()
	var names []string
	for i := 0; i < 12; i++ {
		names = append(names, "V"+string(rune('a'+i)))
	}
	for run := 0; run < 20; run++ {
		sig := NewSig()
		sig.AddSort(S)
		c := newChecker(sig, nil)
		c.varUniq = NewVariableUniqifier(nil)
		var last *LogicVariable
		for i, nm := range names {
			v, err := NewVariable(nm, S)
			if err != nil {
				t.Fatal(err)
			}
			vid := makeVarID(v)
			c.universallyQuantifiedVars[vid] = v
			c.universalVarLoc[vid] = Location{Filename: "f.ivy", Line: i + 1}
			vk := varKey(v)
			c.stratMap.Set(vk, node)
			c.stratInfo[vk] = stratEntry{v: v}
			last = v
		}
		var msg string
		func() {
			defer func() {
				if r := recover(); r != nil {
					msg = r.(*FragmentError).Message
				}
			}()
			c.reportInterpOverVar(last, Location{Filename: "u.ivy", Line: 50}, node)
		}()
		if !strings.Contains(msg, "The quantified variable is "+last.Name+":") {
			t.Fatalf("run %d: message should name %s:\n%s", run, last.Name, msg)
		}
	}
}
