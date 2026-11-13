package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
)

// SetCecCombinationalWitness stores the one-frame, zero-register witness used
// by ABC's CEC sweepers in the sequential witness slot. Every CI, including
// register outputs, is represented as a combinational input.
func (p *Gia_Man_t) SetCecCombinationalWitness(model *util.Abc_Cex_t) error {
	if p == nil || model == nil {
		return fmt.Errorf("gia: nil CEC witness input")
	}
	if model.NRegs != 0 || model.NPis != Gia_ManCiNum(p) || model.NBits != model.NPis || model.IFrame != 0 || model.IPo < 0 || model.IPo >= Gia_ManCoNum(p) || int64(len(model.PData))*32 < int64(model.NBits) {
		return fmt.Errorf("gia: invalid CEC witness dimensions")
	}
	clone := *model
	clone.PData = append([]uint32(nil), model.PData[:(int64(model.NBits)+31)/32]...)
	p.pCexSeq = &clone
	return nil
}
