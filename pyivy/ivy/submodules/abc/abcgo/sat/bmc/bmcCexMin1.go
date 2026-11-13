package bmc

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_ManCexMinPerform derives register reasons and writes the intermediate
// cube AIG. The original routine does not yet construct a reduced CEX.
func Saig_ManCexMinPerform(g *aig.Graph, cex *util.Abc_Cex_t) *util.Abc_Cex_t {
	reasons := Saig_ManCexMinComputeReason(g, cex, 0)
	if reasons == nil {
		return nil
	}
	size := vec.Vec_VecSizeSize(reasons)
	fmt.Printf("Reason size = %d.  Ave = %d.\n", size, size/int(cex.IFrame+1))
	graph := Saig_ManCexMinDupWithCubes(g, reasons)
	if graph == nil {
		return nil
	}
	if err := ioa.Ioa_WriteAiger(graph, "aigcube.aig", 0, 0); err != nil {
		fmt.Printf("Could not write intermediate AIG: %v\n", err)
	} else {
		fmt.Println("Intermediate AIG is written into file \"aigcube.aig\".")
	}
	return nil
}
