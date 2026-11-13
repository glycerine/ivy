package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// SetSplitCofactors stores the literal-indexed PI assignments used by CEC's
// cofactoring solver. The slice is copied; nil clears the metadata.
func (p *Gia_Man_t) SetSplitCofactors(lits []int32) {
	if lits == nil {
		p.vCofVars = nil
		return
	}
	p.vCofVars = vec.Vec_IntAllocArrayCopy(lits, len(lits))
}
func (p *Gia_Man_t) SplitCofactors() []int32 {
	if p.vCofVars == nil {
		return nil
	}
	return append([]int32(nil), vec.Vec_IntArray(p.vCofVars)...)
}

// SetSplitCounterexample transfers CEC split's combinational PI-only model,
// including when the source manager has register-output CIs.
func (p *Gia_Man_t) SetSplitCounterexample(c *util.Abc_Cex_t) error {
	if p == nil || c == nil || c.NRegs != 0 || c.NPis < 0 || c.NBits != c.NPis || c.IFrame != 0 || c.IPo < 0 || c.IPo >= Gia_ManPoNum(p) || int64(len(c.PData))*32 < int64(c.NBits) {
		return fmt.Errorf("gia: invalid split counterexample")
	}
	p.pCexComb = c
	return nil
}
