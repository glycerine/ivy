// Go port of Gia_ManCexRemap from absOut.c.
package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManCexRemap transfers abstract primary input bits to their original PIs.
// An invalid replay is rejected rather than returning an incorrect witness.
func Gia_ManCexRemap(p *gia.Gia_Man_t, abstract *util.Abc_Cex_t, pis *vec.Vec_Int_t) *util.Abc_Cex_t {
	if abstract.IPo != 0 {
		panic("abs: expected output zero")
	}
	result := util.Abc_CexAlloc(gia.Gia_ManRegNum(p), gia.Gia_ManPiNum(p), abstract.IFrame+1)
	result.IFrame, result.IPo = abstract.IFrame, abstract.IPo
	for frame := int32(0); frame <= abstract.IFrame; frame++ {
		for i, id := range vec.Vec_IntArray(pis) {
			if util.Abc_InfoHasBit(abstract.PData, abstract.NRegs+abstract.NPis*frame+int32(i)) {
				pi := gia.Gia_ObjCioId(gia.Gia_ManObj(p, id))
				util.Abc_InfoSetBit(result.PData, result.NRegs+result.NPis*frame+pi)
			}
		}
	}
	if gia.Gia_ManVerifyCex(p, result, 0) == 0 {
		fmt.Println("Gia_ManCexRemap(): Counter-example is invalid.")
		return nil
	}
	fmt.Println("Counter-example verification is successful.")
	return result
}
