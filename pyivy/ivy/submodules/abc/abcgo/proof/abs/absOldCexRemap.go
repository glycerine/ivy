package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
)

// Saig_ManCexRemap copies the original primary inputs from an abstraction's
// counterexample and verifies that they also produce a real counterexample.
func Saig_ManCexRemap(g, abstraction *aig.Graph, cex *util.Abc_Cex_t) *util.Abc_Cex_t {
	if saig.Saig_ManVerifyCex(abstraction, cex) == 0 {
		fmt.Println("Saig_ManCexRemap(): The initial counter-example is invalid.")
	}
	result := util.Abc_CexAlloc(saig.Saig_ManRegNum(g), saig.Saig_ManPiNum(g), cex.IFrame+1)
	result.IFrame, result.IPo = cex.IFrame, cex.IPo
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		for i := int32(0); i < saig.Saig_ManPiNum(g) && i < saig.Saig_ManPiNum(abstraction); i++ {
			if util.Abc_InfoHasBit(cex.PData, cex.NRegs+cex.NPis*frame+i) {
				util.Abc_InfoSetBit(result.PData, result.NRegs+result.NPis*frame+i)
			}
		}
	}
	if saig.Saig_ManVerifyCex(g, result) == 0 {
		fmt.Println("Saig_ManCexRemap(): Counter-example is invalid.")
		return nil
	}
	fmt.Printf("Counter-example verification is successful.\nOutput %d of miter %q was asserted in frame %d.\n", result.IPo, g.Name, result.IFrame)
	return result
}
