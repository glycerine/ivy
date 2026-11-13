package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
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

// SetSequentialCounterexample validates and copies a sequential BMC witness.
func (p *Gia_Man_t) SetSequentialCounterexample(model *util.Abc_Cex_t) error {
	if p == nil || model == nil {
		return fmt.Errorf("gia: nil sequential counterexample input")
	}
	pis := Gia_ManPiNum(p)
	pos := Gia_ManPoNum(p)
	if model.NRegs != p.nRegs || model.NPis != pis || model.IFrame < 0 || model.IPo < 0 || model.IPo >= pos {
		return fmt.Errorf("gia: invalid sequential counterexample dimensions")
	}
	bits := int64(model.NRegs) + (int64(model.IFrame)+1)*int64(model.NPis)
	if bits != int64(model.NBits) || bits < 0 || int64(len(model.PData))*32 < bits {
		return fmt.Errorf("gia: truncated or inconsistent sequential counterexample")
	}
	copyOfModel := *model
	copyOfModel.PData = append([]uint32(nil), model.PData[:(bits+31)/32]...)
	p.pCexSeq = &copyOfModel
	return nil
}

// SequentialCounterexample returns borrowed sequential witness storage.
func (p *Gia_Man_t) SequentialCounterexample() *util.Abc_Cex_t { return p.pCexSeq }

func (p *Gia_Man_t) ClearSequentialCounterexample() { p.pCexSeq = nil }

// TransferSequentialCounterexampleSource transfers ownership as ABC's direct
// pCexSeq assignment does, without checking dimensions against this manager.
// Some ABC callers transfer witnesses from transformed graphs whose register
// counts differ from the destination; SetSequentialCounterexample remains the
// validated copying API for ordinary callers.
func (p *Gia_Man_t) TransferSequentialCounterexampleSource(model *util.Abc_Cex_t) {
	p.pCexSeq = model
}

func (p *Gia_Man_t) TakeSequentialCounterexample() *util.Abc_Cex_t {
	model := p.pCexSeq
	p.pCexSeq = nil
	return model
}

// SetSequentialCounterexamples stores a copied output-indexed CEX list. Nil
// entries mark properties that were not disproved within the requested bound.
func (p *Gia_Man_t) SetSequentialCounterexamples(models []*util.Abc_Cex_t) error {
	if p == nil || len(models) != int(Gia_ManPoNum(p)) {
		return fmt.Errorf("gia: sequential CEX list has invalid output count")
	}
	copyOfModels := vec.Vec_PtrAllocExact(len(models))
	for output, model := range models {
		if model == nil {
			vec.Vec_PtrPush(copyOfModels, nil)
			continue
		}
		if model.NRegs != p.nRegs || model.NPis != Gia_ManPiNum(p) || model.IPo != int32(output) || model.IFrame < 0 {
			return fmt.Errorf("gia: sequential CEX %d has invalid dimensions", output)
		}
		bits := int64(model.NRegs) + (int64(model.IFrame)+1)*int64(model.NPis)
		if bits != int64(model.NBits) || bits < 0 || int64(len(model.PData))*32 < bits {
			return fmt.Errorf("gia: sequential CEX %d is truncated or inconsistent", output)
		}
		owned := *model
		owned.PData = append([]uint32(nil), model.PData[:(bits+31)/32]...)
		vec.Vec_PtrPush(copyOfModels, &owned)
	}
	p.vSeqModelVec = copyOfModels
	return nil
}

// SequentialCounterexamples returns an output-indexed copy of stored witnesses.
func (p *Gia_Man_t) SequentialCounterexamples() []*util.Abc_Cex_t {
	if p == nil || p.vSeqModelVec == nil {
		return nil
	}
	models := make([]*util.Abc_Cex_t, vec.Vec_PtrSize(p.vSeqModelVec))
	for i := range models {
		if model, ok := vec.Vec_PtrEntry(p.vSeqModelVec, i).(*util.Abc_Cex_t); ok && model != nil {
			copyOfModel := *model
			copyOfModel.PData = append([]uint32(nil), model.PData...)
			models[i] = &copyOfModel
		}
	}
	return models
}

// ValidateSequentialCounterexample independently replays the declared transition
// relation. checkResets additionally enforces explicit AIGER resets (or ordinary
// zero initialization when reset metadata is absent). The manager is unchanged.
func (p *Gia_Man_t) ValidateSequentialCounterexample(cex *util.Abc_Cex_t, checkResets bool) error {
	if p == nil || cex == nil {
		return fmt.Errorf("gia: nil sequential counterexample input")
	}
	if cex.NRegs != p.nRegs || cex.NPis != Gia_ManPiNum(p) || cex.IFrame < 0 || cex.IPo < 0 || cex.IPo >= Gia_ManPoNum(p) {
		return fmt.Errorf("gia: invalid sequential counterexample dimensions")
	}
	required := int64(cex.NRegs) + int64(cex.NPis)*(int64(cex.IFrame)+1)
	if required > int64(^uint32(0)>>1) || int64(cex.NBits) < required || int64(len(cex.PData))*32 < int64(cex.NBits) {
		return fmt.Errorf("gia: short or oversized sequential counterexample")
	}
	bit := func(index int32) bool { return cex.PData[index>>5]>>uint(index&31)&1 != 0 }
	state := make([]bool, p.nRegs)
	for i := range state {
		state[i] = bit(int32(i))
		reset := int32(0)
		if checkResets && len(p.aigerLatchInits) != 0 {
			if len(p.aigerLatchInits) != len(state) {
				return fmt.Errorf("gia: invalid latch reset metadata")
			}
			reset = p.aigerLatchInits[i]
		}
		if checkResets && (reset < -1 || reset > 1 || reset >= 0 && state[i] != (reset == 1)) {
			return fmt.Errorf("gia: counterexample violates reset of register %d", i)
		}
	}
	input := make([]bool, Gia_ManCiNum(p))
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		for i := int32(0); i < cex.NPis; i++ {
			input[i] = bit(cex.NRegs + frame*cex.NPis + i)
		}
		copy(input[cex.NPis:], state)
		output, err := p.Evaluate(input)
		if err != nil {
			return err
		}
		if frame == cex.IFrame {
			if !output[cex.IPo] {
				return fmt.Errorf("gia: counterexample does not fail output %d", cex.IPo)
			}
			break
		}
		copy(state, output[Gia_ManPoNum(p):])
	}
	return nil
}
