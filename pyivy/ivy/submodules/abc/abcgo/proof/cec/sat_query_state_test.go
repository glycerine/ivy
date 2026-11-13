package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// Cec_ManSatCheckNode checks constant zero before changing SAT state, and
// neither node-check function clears the pattern vector (cecSolve.c:470-655).
func TestSourceNodeQueryPreservesSavedPattern(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	p := Cec_ManSatCreate(g, nil)
	obj := gia.Gia_ManObj(g, a>>1)
	if Cec_ManSatCheckNode(p, obj) != 0 {
		t.Fatal("expected SAT")
	}
	vec.Vec_IntPush(p.vCex, 17)
	if Cec_ManSatCheckNode(p, gia.Gia_ManConst0(g)) != 1 {
		t.Fatal("constant zero")
	}
	if len(p.values) == 0 || Cec_ObjSatVarValue(p, obj) != 1 || vec.Vec_IntSize(p.vCex) != 1 {
		t.Fatal("constant query changed saved witness or pattern")
	}
	if Cec_ManSatCheckNode(p, gia.Gia_Not(obj)) != 0 {
		t.Fatal("expected second SAT")
	}
	if vec.Vec_IntSize(p.vCex) != 1 || vec.Vec_IntEntry(p.vCex, 0) != 17 {
		t.Fatal("SAT query cleared saved pattern")
	}
	if Cec_ManSatCheckNodeTwo(p, obj, obj) != 0 {
		t.Fatal("expected SAT pair")
	}
	if vec.Vec_IntSize(p.vCex) != 1 {
		t.Fatal("pair query cleared saved pattern")
	}
}
