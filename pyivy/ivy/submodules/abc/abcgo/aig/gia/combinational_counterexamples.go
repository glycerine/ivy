package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// SetCombinationalCounterexamples stores the CO-indexed result vector used by
// combinational CEC, including register inputs. A nil entry means unknown and
// util.AbcCexSentinel means proved; other entries are copied witnesses.
// This replaces the same source vector used by sequential counterexamples.
func (p *Gia_Man_t) SetCombinationalCounterexamples(models []*util.Abc_Cex_t) error {
	if p == nil || len(models) != int(Gia_ManCoNum(p)) {
		return fmt.Errorf("gia: combinational CEX list has invalid output count")
	}
	result := vec.Vec_PtrAllocExact(len(models))
	for output, model := range models {
		if model == nil || model == util.AbcCexSentinel {
			vec.Vec_PtrPush(result, model)
			continue
		}
		n := Gia_ManCiNum(p)
		if model.NRegs != 0 || model.NPis != n || model.NBits != n || model.IFrame != 0 || model.IPo != int32(output) || int64(len(model.PData))*32 < int64(n) {
			return fmt.Errorf("gia: combinational CEX %d has invalid dimensions", output)
		}
		owned := *model
		owned.PData = append([]uint32(nil), model.PData[:(int64(n)+31)/32]...)
		vec.Vec_PtrPush(result, &owned)
	}
	p.vSeqModelVec = result
	return nil
}

// CombinationalCounterexamples returns owned copies and preserves the proved
// sentinel identity. The returned vector includes every CO, not just POs.
func (p *Gia_Man_t) CombinationalCounterexamples() []*util.Abc_Cex_t {
	if p == nil || p.vSeqModelVec == nil {
		return nil
	}
	models := make([]*util.Abc_Cex_t, vec.Vec_PtrSize(p.vSeqModelVec))
	for i := range models {
		model, ok := vec.Vec_PtrEntry(p.vSeqModelVec, i).(*util.Abc_Cex_t)
		if !ok || model == nil {
			continue
		}
		if model == util.AbcCexSentinel {
			models[i] = model
			continue
		}
		owned := *model
		owned.PData = append([]uint32(nil), model.PData...)
		models[i] = &owned
	}
	return models
}

// SetCombinationalCounterexampleInSequentialSlot stores the single
// combinational witness used by ABC's Cec4 sweep in its sequential-witness
// slot. Register outputs are treated as CIs and register inputs as COs.
func (p *Gia_Man_t) SetCombinationalCounterexampleInSequentialSlot(model *util.Abc_Cex_t) error {
	if p == nil || model == nil {
		return fmt.Errorf("gia: nil combinational CEX")
	}
	n := Gia_ManCiNum(p)
	if model.NRegs != 0 || model.NPis != n || model.NBits != n || model.IFrame != 0 || model.IPo < 0 || model.IPo >= Gia_ManCoNum(p) || int64(len(model.PData))*32 < int64(n) {
		return fmt.Errorf("gia: invalid combinational CEX dimensions")
	}
	owned := *model
	owned.PData = append([]uint32(nil), model.PData[:(int64(n)+31)/32]...)
	p.pCexSeq = &owned
	return nil
}
