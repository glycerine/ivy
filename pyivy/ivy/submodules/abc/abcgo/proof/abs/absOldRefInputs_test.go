package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestFirstAbstractedFlopInput(t *testing.T) {
	g := aig.NewGraph()
	x, lo0, lo1 := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.OR(x, g.OR(lo0, lo1)))
	g.AddOutput(x)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 2)
	abstraction := saig.Saig_ManDupAbstraction(g, vec.Vec_IntAllocArrayCopy([]int32{0}, 1))
	if got := Saig_ManCexFirstFlopPi(g, abstraction); got != 1 {
		t.Fatalf("first flop PI = %d", got)
	}
}
