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
	copyOfModels := Vec_PtrAllocExact(len(models))
	for output, model := range models {
		if model == nil {
			Vec_PtrPush(copyOfModels, nil)
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
		Vec_PtrPush(copyOfModels, &owned)
	}
	p.vSeqModelVec = copyOfModels
	return nil
}

// SequentialCounterexamples returns an output-indexed copy of stored witnesses.
func (p *Gia_Man_t) SequentialCounterexamples() []*util.Abc_Cex_t {
	if p == nil || p.vSeqModelVec == nil {
		return nil
	}
	models := make([]*util.Abc_Cex_t, Vec_PtrSize(p.vSeqModelVec))
	for i := range models {
		if model, ok := Vec_PtrEntry(p.vSeqModelVec, i).(*util.Abc_Cex_t); ok && model != nil {
			copyOfModel := *model
			copyOfModel.PData = append([]uint32(nil), model.PData...)
			models[i] = &copyOfModel
		}
	}
	return models
}
