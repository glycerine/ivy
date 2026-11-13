package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDuplicateWithNextStateCubes(t *testing.T) {
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	cubes := vec.Vec_VecAlloc(2)
	vec.Vec_VecPushInt(cubes, 0, 0)
	vec.Vec_VecPushInt(cubes, 1, 1)
	dup := Saig_ManDupWithCubes(g, cubes)
	for input := int32(0); input < 2; input++ {
		for output := int32(0); output < 2; output++ {
			cex := util.Abc_CexAlloc(1, 1, 1)
			cex.IPo = output
			if input != 0 {
				util.Abc_InfoSetBit(cex.PData, 1)
			}
			want := (input == 1 && output == 0) || (input == 0 && output == 1)
			if got := saig.Saig_ManVerifyCex(dup, cex) != 0; got != want {
				t.Fatalf("input %d output %d = %v", input, output, got)
			}
		}
	}
}
