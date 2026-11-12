package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRetimeClassesAig(t *testing.T) {
	g := aig.NewGraph()
	pi := g.NewInput()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(pi)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 2)
	classes := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(classes, 3)
	vec.Vec_IntPush(classes, 3)
	result, out := Gia_ManRetimeClassesAig(g, classes, 0)
	if result.RegisterCount() != 1 || vec.Vec_IntSize(out) != 1 || vec.Vec_IntEntry(out, 0) != 3 || vec.Vec_IntSize(classes) != 2 {
		t.Fatal("retiming class mismatch")
	}
}
