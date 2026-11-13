package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSpecialInitializationAndReinitialization(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	li := g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_SmlStart(g, 0, 2, 2)
	init := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(init, 1)
	vec.Vec_IntPush(init, 0)
	Ssw_SmlInitializeSpecial(p, init)
	state := Ssw_SmlSimInfo(p, lo)
	if state[0] != ^uint32(0) || state[1] != 0 {
		t.Fatalf("initial state %x", state)
	}
	out := Ssw_SmlSimInfo(p, li)
	out[2], out[3] = 123, 456
	Ssw_SmlReinitialize(p)
	if state[0] != 123 || state[1] != 456 {
		t.Fatal("final state not copied to initial state")
	}
}
