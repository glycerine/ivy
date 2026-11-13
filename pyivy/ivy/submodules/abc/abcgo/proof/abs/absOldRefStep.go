package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// Saig_ManCexRefineStep adds registers needed by an abstract counterexample.
func Saig_ManCexRefineStep(g *aig.Graph, flops, classes *vec.Vec_Int_t, cex *util.Abc_Cex_t, limit, tryFour, sensePath, verbose int32) int32 {
	started := time.Now()
	abstract := saig.Saig_ManDupAbstraction(g, flops)
	first := Saig_ManCexFirstFlopPi(g, abstract)
	var selected *vec.Vec_Int_t
	if sensePath != 0 {
		selected = Saig_ManExtendCounterExampleTest2(abstract, first, cex, verbose)
	} else {
		selected = Saig_ManExtendCounterExampleTest3(abstract, first, cex, verbose)
	}
	if selected == nil {
		return 0
	}
	if vec.Vec_IntSize(selected) == 0 {
		fmt.Println("Refinement did not happen. Discovered a true counter-example.")
		fmt.Printf("Remapping counter-example from %d to %d primary inputs.\n", len(abstract.Inputs()), len(g.Inputs()))
		model := Saig_ManCexRemap(g, abstract, cex)
		if model != nil {
			if err := g.SetSequentialCounterexample(model); err != nil {
				panic(err)
			}
		}
		return 0
	}
	if verbose != 0 {
		fmt.Printf("Adding %d registers to the abstraction (total = %d). Time = %s\n", vec.Vec_IntSize(selected), int(saig.Saig_ManRegNum(g))+vec.Vec_IntSize(selected), time.Since(started))
	}
	piCount := saig.Saig_ManPiNum(g)
	if limit > 0 && vec.Vec_IntSize(selected) > int(limit) {
		for i := range vec.Vec_IntArray(selected) {
			vec.Vec_IntAddToEntry(selected, i, -piCount)
		}
		before := vec.Vec_IntSize(selected)
		selected = Saig_ManCbaFilterFlops(g, cex, classes, selected, limit)
		fmt.Printf("Filtering flops based on cost (%d -> %d).\n", before, vec.Vec_IntSize(selected))
		for i := range vec.Vec_IntArray(selected) {
			vec.Vec_IntAddToEntry(selected, i, piCount)
		}
	}
	for _, input := range vec.Vec_IntArray(selected) {
		if input < 0 || int(input) >= len(abstract.InputNumbers) {
			panic("invalid abstract input")
		}
		original := abstract.InputNumbers[input]
		if original < piCount || int(original) >= len(g.Inputs()) {
			panic("abstract input is not a latch")
		}
		vec.Vec_IntPush(flops, original-piCount)
	}
	return 1
}
