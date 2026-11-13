package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestFraigNodeRejectsComplementedParentBeforeClassLookup(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	p := &Fra_Man_t{PManAig: g, PCla: Fra_ClassesStart(g), PPars: &Fra_Par_t{}}
	defer func() {
		if recover() == nil {
			t.Error("source rejects complemented parent even when it has no representative")
		}
	}()
	Fra_FraigNode(p, x.Not())
}

func TestFraigNodeRejectsConstantMappedNodeBeforeSAT(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	p := satTestManager(g)
	p.PCla.Reprs[b.NodeID()] = a.NodeID()
	Fra_ObjSetFraig(p, a, 0, a)
	Fra_ObjSetFraig(p, b, 0, g.Const1())
	defer func() {
		if recover() == nil {
			t.Error("source rejects a constant mapped node in combinational mode")
		}
		if p.NSatCalls != 0 {
			t.Errorf("source asserts before SAT; got %d calls", p.NSatCalls)
		}
	}()
	Fra_FraigNode(p, b)
}
