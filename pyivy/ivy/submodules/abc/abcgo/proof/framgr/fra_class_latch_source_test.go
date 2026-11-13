package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"reflect"
	"testing"
)

// C Fra_ClassesLatchCorr clears the constant-class vector and initializes
// latch candidates, but leaves existing classes and other repr fields intact.
func TestLatchCorrespondencePreservesExistingClasses(t *testing.T) {
	g := aig.NewGraph()
	a, b, old, reg := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(reg)
	aig.Aig_ManSetRegNum(g, 1)
	p := Fra_ClassesStart(g)
	Fra_ClassesTest(p, int32(a.NodeID()), int32(b.NodeID()))
	p.Constants = []int{old.NodeID()}
	p.Reprs[old.NodeID()] = 0
	Fra_ClassesLatchCorr(&Fra_Man_t{PManAig: g, PCla: p})
	if !reflect.DeepEqual(p.Classes, [][]int{{a.NodeID(), b.NodeID()}}) {
		t.Errorf("classes = %v", p.Classes)
	}
	if p.Reprs[b.NodeID()] != a.NodeID() || p.Reprs[old.NodeID()] != 0 {
		t.Errorf("existing reprs = %v", p.Reprs)
	}
	if !reflect.DeepEqual(p.Constants, []int{reg.NodeID()}) || p.Reprs[reg.NodeID()] != 0 {
		t.Errorf("latch candidates = %v reprs %v", p.Constants, p.Reprs)
	}
}
