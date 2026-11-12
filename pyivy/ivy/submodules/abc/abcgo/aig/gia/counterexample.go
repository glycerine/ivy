package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	ceccore "github.com/glycerine/abcgo/proof/cec/core"
)

func (p *Gia_Man_t) SetCombinationalCounterexample(inputs []bool, output int) error {
	if p.nRegs < 0 || p.nRegs > Gia_ManCiNum(p) || p.nRegs > Gia_ManCoNum(p) || len(inputs) != int(Gia_ManCiNum(p)) {
		return fmt.Errorf("gia: invalid counterexample input dimensions")
	}
	if output < -1 || output >= int(Gia_ManCoNum(p)) {
		return fmt.Errorf("gia: invalid counterexample output %d", output)
	}
	giaSaveCounterexample(p, &ceccore.Counterexample{Inputs: inputs, Output: output})
	return nil
}

// CombinationalCounterexample returns borrowed model storage.
func (p *Gia_Man_t) CombinationalCounterexample() *util.Abc_Cex_t { return p.pCexComb }
func (p *Gia_Man_t) ClearCombinationalCounterexample()            { p.pCexComb = nil }
func (p *Gia_Man_t) TakeCombinationalCounterexample() *util.Abc_Cex_t {
	model := p.pCexComb
	p.pCexComb = nil
	return model
}
