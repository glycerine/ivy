package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_ManCbaShrink maps visited frame nodes into per-frame state cubes.
func Saig_ManCbaShrink(p *Saig_ManCba_t) {
	if p.vReg2Value == nil {
		p.vReg2Value = vec.Vec_VecAlloc(vec.Vec_VecSize(p.vReg2Frame))
	}
	regs := saig.Saig_ManRegNum(p.pAig)
	for frame := 0; frame < vec.Vec_VecSize(p.vReg2Frame); frame++ {
		vec.Vec_VecExpandInt(p.vReg2Value, frame)
		values := vec.Vec_IntArray(vec.Vec_VecEntryInt(p.vReg2Frame, frame))
		for k := 0; k+1 < len(values); k += 2 {
			node, lit := values[k], values[k+1]
			id := int(lit >> 1)
			if id >= len(p.visited) || (id != p.pFrames.Const1().NodeID() && !p.visited[id]) {
				continue
			}
			reg := int32(-1)
			for i := int32(0); i < regs; i++ {
				if int32(saig.Saig_ManLi(p.pAig, i).NodeID()) == node {
					reg = i
					break
				}
			}
			if reg < 0 {
				panic("CBA frame entry is not a latch input")
			}
			sign := lit & 1
			if !p.phase[id] {
				sign ^= 1
			}
			vec.Vec_VecPushInt(p.vReg2Value, frame, 2*reg+sign)
		}
		care := vec.Vec_IntSize(vec.Vec_VecEntryInt(p.vReg2Value, frame))
		fmt.Printf("Level = %4d   StateBits = %4d (%6.2f %%)  CareBits = %4d (%6.2f %%)\n", frame, len(values)/2, 100*float64(len(values)/2)/float64(regs), care, 100*float64(care)/float64(regs))
	}
	_ = Saig_ManDupWithCubes(p.pAig, p.vReg2Value)
}
