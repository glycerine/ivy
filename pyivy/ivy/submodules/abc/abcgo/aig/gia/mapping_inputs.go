package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

// MappingInputs returns borrowed timing and output-attribute data.
func (p *Gia_Man_t) MappingInputs() (arrivals, required, attrs []int32, timing *tim.Tim_Man_t) {
	if p.vCiArrs != nil {
		arrivals = vec.Vec_IntArray(p.vCiArrs)
	}
	if p.vCoReqs != nil {
		required = vec.Vec_IntArray(p.vCoReqs)
	}
	if p.vCoAttrs != nil {
		attrs = vec.Vec_IntArray(p.vCoAttrs)
	}
	if p.PManTime != nil {
		var ok bool
		timing, ok = p.PManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("gia: invalid timing manager")
		}
	}
	return
}

// SetMappingOutputArrivals replaces the output arrival times with owned data.
func (p *Gia_Man_t) SetMappingOutputArrivals(arrivals []int32) {
	if len(arrivals) != int(Gia_ManCoNum(p)) {
		panic("gia: output arrival count mismatch")
	}
	p.vCoArrs = vec.Vec_IntAllocArrayCopy(arrivals, len(arrivals))
}

// ClearCellConfiguration drops cell-specific configuration on remapping.
func (p *Gia_Man_t) ClearCellConfiguration() { p.VConfigs = nil; p.PCellStr = "" }

// ReplacePacking installs a validated owned packing vector.
func (p *Gia_Man_t) ReplacePacking(packing *Vec_Int_t) error {
	if packing != nil {
		if err := aigerValidatePacking(vec.Vec_IntArray(packing), p.nObjs); err != nil {
			return err
		}
	}
	p.VPacking = packing
	return nil
}
