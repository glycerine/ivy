package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bmc"
	"time"
)

// Saig_ManCbaPerform checks the abstraction and returns registers to restore.
// A nil result denotes a real counterexample; an empty vector means the
// bounded search stopped without a counterexample to refine.
func Saig_ManCbaPerform(g *aig.Graph, nInputs int32, params *bmc.Saig_ParBmc_t) *vec.Vec_Int_t {
	started := time.Now()
	status := bmc.Saig_ManBmcScalable(g, params)
	if status == -1 {
		fmt.Println("Resource limit is reached during BMC.")
		if g.SequentialCounterexample() != nil {
			panic("BMC retained a counterexample after resource limit")
		}
		return vec.Vec_IntAlloc(0)
	}
	cex := g.SequentialCounterexample()
	if cex == nil {
		fmt.Println("BMC did not detect a CEX with the given depth.")
		return vec.Vec_IntAlloc(0)
	}
	result := Saig_ManCbaFilterInputs(g, nInputs, cex, params.AbstractionVerbose())
	if vec.Vec_IntSize(result) == 0 {
		return nil
	}
	if params.AbstractionVerbose() != 0 {
		fmt.Printf("Adding %d registers to the abstraction (total = %d). Time = %s\n", vec.Vec_IntSize(result), int(saig.Saig_ManRegNum(g))+vec.Vec_IntSize(result), time.Since(started))
	}
	return result
}
