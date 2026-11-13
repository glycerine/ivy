package aig

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// SetCombinationalCounterexample copies CI values, storing state before primary
// input bits in ABC's counterexample layout. A failed call preserves the model.
func (p *Graph) SetCombinationalCounterexample(inputs []bool, output int) error {
	if p.nRegs < 0 || int(p.nRegs) > len(p.inputs) || int(p.nRegs) > len(p.outputs) || len(inputs) != len(p.inputs) {
		return fmt.Errorf("aig: invalid counterexample input dimensions")
	}
	if output < -1 || output >= len(p.outputs) {
		return fmt.Errorf("aig: invalid counterexample output %d", output)
	}
	pis := len(p.inputs) - int(p.nRegs)
	model := util.Abc_CexAlloc(p.nRegs, int32(pis), 1)
	model.IPo = int32(output)
	for i, v := range inputs {
		if !v {
			continue
		}
		bit := int(p.nRegs) + i
		if i >= pis {
			bit = i - pis
		}
		model.PData[bit>>5] |= 1 << uint(bit&31)
	}
	p.combinationalCex = model
	return nil
}

// CombinationalCounterexample returns borrowed model storage.
func (p *Graph) CombinationalCounterexample() *util.Abc_Cex_t { return p.combinationalCex }
func (p *Graph) ClearCombinationalCounterexample()            { p.combinationalCex = nil }
func (p *Graph) TakeCombinationalCounterexample() *util.Abc_Cex_t {
	model := p.combinationalCex
	p.combinationalCex = nil
	return model
}

// SetSequentialCounterexample validates and copies a complete multi-frame
// witness. It shares the original pSeqModel storage with one-frame witnesses.
// A rejected model leaves the previous witness intact.
func (p *Graph) SetSequentialCounterexample(model *util.Abc_Cex_t) error {
	if p == nil {
		return fmt.Errorf("aig: nil graph")
	}
	if model == nil {
		return fmt.Errorf("aig: nil sequential counterexample")
	}
	pis := len(p.inputs) - int(p.nRegs)
	pos := len(p.outputs) - int(p.nRegs)
	if p.nRegs < 0 || pis < 0 || pos < 0 || model.NRegs != p.nRegs || int64(model.NPis) != int64(pis) || model.IFrame < 0 || model.IPo < 0 || int64(model.IPo) >= int64(pos) {
		return fmt.Errorf("aig: invalid sequential counterexample dimensions")
	}
	count := int64(model.NRegs) + (int64(model.IFrame)+1)*int64(model.NPis)
	if count < 0 || count != int64(model.NBits) || int64(len(model.PData))*32 < count {
		return fmt.Errorf("aig: truncated or inconsistent sequential counterexample")
	}
	copied := *model
	copied.PData = append([]uint32(nil), model.PData[:(count+31)/32]...)
	p.combinationalCex = &copied
	return nil
}

// SequentialCounterexample returns borrowed witness storage.
func (p *Graph) SequentialCounterexample() *util.Abc_Cex_t { return p.combinationalCex }
func (p *Graph) ClearSequentialCounterexample()            { p.combinationalCex = nil }
func (p *Graph) TakeSequentialCounterexample() *util.Abc_Cex_t {
	return p.TakeCombinationalCounterexample()
}

// SetSequentialCounterexamples transfers the per-output witness vector to
// the graph. Entries may be nil when an output has no counterexample.
func (p *Graph) SetSequentialCounterexamples(models *vec.Vec_Ptr_t) {
	p.seqModelVec = models
}

// SequentialCounterexamples returns the graph's borrowed per-output vector.
func (p *Graph) SequentialCounterexamples() *vec.Vec_Ptr_t { return p.seqModelVec }

// TakeSequentialCounterexamples transfers the vector out of the graph.
func (p *Graph) TakeSequentialCounterexamples() *vec.Vec_Ptr_t {
	models := p.seqModelVec
	p.seqModelVec = nil
	return models
}

// SetInductionCounterexample stores the witness produced by bounded
// induction. Its dimensions follow the induction algorithm's compact witness
// layout, so it is validated against its compact dimensions instead.
func (p *Graph) SetInductionCounterexample(model *util.Abc_Cex_t) error {
	if p == nil || model == nil || p.nRegs <= 0 {
		return fmt.Errorf("aig: invalid induction counterexample target")
	}
	wantRegs := p.nRegs - 1
	wantPis := int32(len(p.inputs)) - p.nRegs
	wantBits := wantRegs + wantPis
	if wantPis < 0 || model.NRegs != wantRegs || model.NPis != wantPis || model.IFrame != 0 || model.IPo != 0 || model.NBits != wantBits || int64(len(model.PData))*32 < int64(wantBits) {
		return fmt.Errorf("aig: invalid compact induction counterexample dimensions")
	}
	p.inductionCex = model
	return nil
}

// InductionCounterexample returns the witness saved by bounded induction.
func (p *Graph) InductionCounterexample() *util.Abc_Cex_t {
	if p == nil {
		return nil
	}
	return p.inductionCex
}

// TakeInductionCounterexample transfers the saved witness to the caller.
func (p *Graph) TakeInductionCounterexample() *util.Abc_Cex_t {
	if p == nil {
		return nil
	}
	model := p.inductionCex
	p.inductionCex = nil
	return model
}

// ClearInductionCounterexample discards the saved witness.
func (p *Graph) ClearInductionCounterexample() {
	if p != nil {
		p.inductionCex = nil
	}
}
