package abci

import (
	"fmt"
	"io"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bmc"
)

// Abc_NtkDumpOneCexSpecial writes the minimized counterexample in ABC's
// special CEX format. The writer corresponds to C's FILE* argument.
func Abc_NtkDumpOneCexSpecial(w io.Writer, network *Abc_Ntk_t, cex *Abc_Cex_t) {
	graph := baseabc.Abc_NtkToDar(network, 0, 1)
	care := bmc.Bmc_CexCareMinimize(graph, saig.Saig_ManPiNum(graph), cex, 4, 0, 0)
	aig.Aig_ManStop(graph)
	if care == nil {
		fmt.Println("Counter-example minimization has failed.")
		return
	}
	defer util.Abc_CexFreeP(&care)

	for _, latch := range network.ObjectsOfKind(baseabc.ObjectLatch) {
		value := 1 - baseabc.Abc_LatchIsInit0(latch)
		fmt.Fprintf(w, "CEX: %s@0=%c\n", baseabc.Abc_ObjName(baseabc.Abc_ObjFanout0(latch)), '0'+value)
	}
	pis := network.ObjectsOfKind(baseabc.ObjectPI)
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		for i, pi := range pis {
			index := cex.NRegs + cex.NPis*frame + int32(i)
			careIndex := care.NRegs + care.NPis*frame + int32(i)
			if !util.Abc_InfoHasBit(care.PData, careIndex) {
				continue
			}
			value := '0'
			if util.Abc_InfoHasBit(cex.PData, index) {
				value = '1'
			}
			fmt.Fprintf(w, "CEX: %s@%d=%c\n", baseabc.Abc_ObjName(pi), frame, value)
		}
	}
}
