package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManCexAbstractionRefine(g *gia.Gia_Man_t, cex *util.Abc_Cex_t, limit, tryFour, sensePath, verbose int32) int32 {
	classes := g.AbstractionFlopClasses()
	if classes == nil {
		fmt.Println("Gia_ManCexAbstractionRefine(): Abstraction latch map is missing.")
		return -1
	}
	classVec := vec.Vec_IntAllocArrayCopy(classes, len(classes))
	flops := Gia_ManClasses2Flops(classVec)
	graph := gia.Gia_ManToAig(g, 0)
	if Saig_ManCexRefineStep(graph, flops, classVec, cex, limit, tryFour, sensePath, verbose) == 0 {
		model := graph.SequentialCounterexample()
		if model == nil {
			g.ClearSequentialCounterexample()
		} else if err := g.SetSequentialCounterexample(model); err != nil {
			panic(err)
		}
		return 0
	}
	replacement := Gia_ManFlops2Classes(g, flops)
	if err := g.SetAbstractionFlopClasses(vec.Vec_IntArray(replacement)); err != nil {
		panic(err)
	}
	return -1
}
