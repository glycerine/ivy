package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManFaultUnfold creates two combinational timeframes and fault controls.
// Inputs are initial register values, the two sets of primary inputs, then
// one control per register or gate. Outputs are the second-frame outputs.
func Gia_ManFaultUnfold(p *gia.Gia_Man_t, useMuxes, ffOnly int32) *gia.Gia_Man_t {
	if p == nil {
		return nil
	}
	q := gia.Gia_ManStart(5 * gia.Gia_ManObjNum(p))
	gia.Gia_ManHashStart(q)
	first := make([]int32, gia.Gia_ManObjNum(p))
	second := make([]int32, len(first))
	inputs, outputs := p.InputLiterals(), p.OutputLiterals()
	npi, npo := int(gia.Gia_ManPiNum(p)), int(gia.Gia_ManPoNum(p))
	for _, lit := range inputs[npi:] {
		first[lit>>1] = gia.Gia_ManAppendCi(q)
	}
	for _, lit := range inputs[:npi] {
		first[lit>>1] = gia.Gia_ManAppendCi(q)
	}
	read := func(m []int32, lit int32) int32 { return m[lit>>1] ^ (lit & 1) }
	gate := func(m []int32, a, b, c int32, kind int) int32 {
		switch kind {
		case 2:
			return gia.Gia_ManHashAnd(q, read(m, a), read(m, b))
		case 3:
			return gia.Gia_ManHashXor(q, read(m, a), read(m, b))
		case 4:
			return gia.Gia_ManHashMux(q, read(m, a), read(m, b), read(m, c))
		case 5:
			return read(m, a)
		default:
			panic("bmc: unsupported GIA gate")
		}
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		a, b, c, kind := p.Gate(id)
		if kind >= 2 && kind <= 5 {
			first[id] = gate(first, a, b, c, kind)
		}
	}
	for i, lit := range inputs[npi:] {
		second[lit>>1] = read(first, outputs[npo+i])
	}
	for _, lit := range inputs[:npi] {
		second[lit>>1] = gia.Gia_ManAppendCi(q)
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		a, b, c, kind := p.Gate(id)
		if kind < 2 || kind > 5 {
			continue
		}
		value := gate(second, a, b, c, kind)
		if ffOnly == 0 {
			ctrl := gia.Gia_ManAppendCi(q)
			if useMuxes != 0 {
				value = gia.Gia_ManHashMux(q, ctrl, first[id], value)
			}
		}
		second[id] = value
	}
	for i, lit := range outputs {
		value := read(second, lit)
		if ffOnly != 0 && i >= npo {
			ctrl := gia.Gia_ManAppendCi(q)
			if useMuxes != 0 {
				value = gia.Gia_ManHashMux(q, ctrl, read(first, lit), value)
			}
		}
		gia.Gia_ManAppendCo(q, value)
	}
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}

// faultUnfold inserts controls after each internal node, preserving all source
// inputs and outputs in their original order.
func faultUnfold(p *gia.Gia_Man_t, enabled []int32, stuck bool) *gia.Gia_Man_t {
	if p == nil {
		return nil
	}
	count := 0
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		_, _, _, k := p.Gate(id)
		if k >= 2 && k <= 5 {
			count++
		}
	}
	per := 1
	if stuck {
		per = 2
	}
	if len(enabled) != count*per {
		panic("bmc: invalid fault map length")
	}
	q := gia.Gia_ManStart(4 * gia.Gia_ManObjNum(p))
	gia.Gia_ManHashStart(q)
	values := make([]int32, gia.Gia_ManObjNum(p))
	for _, lit := range p.InputLiterals() {
		values[lit>>1] = gia.Gia_ManAppendCi(q)
	}
	read := func(lit int32) int32 { return values[lit>>1] ^ (lit & 1) }
	index := 0
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		a, b, c, k := p.Gate(id)
		var value int32
		switch k {
		case 2:
			value = gia.Gia_ManHashAnd(q, read(a), read(b))
		case 3:
			value = gia.Gia_ManHashXor(q, read(a), read(b))
		case 4:
			value = gia.Gia_ManHashMux(q, read(a), read(b), read(c))
		case 5:
			value = read(a)
		default:
			continue
		}
		ctrl := gia.Gia_ManAppendCi(q)
		if stuck {
			if enabled[index] != 0 {
				value = gia.Gia_ManHashAnd(q, ctrl^1, value)
			}
			index++
			ctrl = gia.Gia_ManAppendCi(q)
			if enabled[index] != 0 {
				value = gia.Gia_ManHashOr(q, ctrl, value)
			}
		} else if enabled[index] != 0 {
			value = gia.Gia_ManHashXor(q, ctrl, value)
		}
		index++
		values[id] = value
	}
	for _, lit := range p.OutputLiterals() {
		gia.Gia_ManAppendCo(q, read(lit))
	}
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}

// Gia_ManStuckAtUnfold adds stuck-at-zero and stuck-at-one controls per gate.
func Gia_ManStuckAtUnfold(p *gia.Gia_Man_t, enabled *vec.Vec_Int_t) *gia.Gia_Man_t {
	return faultUnfold(p, vec.Vec_IntArray(enabled), true)
}

// Gia_ManFlipUnfold adds one output-inversion control per gate.
func Gia_ManFlipUnfold(p *gia.Gia_Man_t, enabled *vec.Vec_Int_t) *gia.Gia_Man_t {
	return faultUnfold(p, vec.Vec_IntArray(enabled), false)
}

// Gia_ManFOFUnfold adds truth-table fault controls for each AND node.
// The final control is reserved and forced to zero, as in ABC's model.
func Gia_ManFOFUnfold(p *gia.Gia_Man_t, enabled *vec.Vec_Int_t) *gia.Gia_Man_t {
	if p == nil {
		return nil
	}
	count := 0
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		_, _, _, kind := p.Gate(id)
		if kind == 2 {
			count++
		} else if kind >= 3 && kind <= 5 {
			panic("bmc: function faults require an AND-inverter graph")
		}
	}
	mapping := vec.Vec_IntArray(enabled)
	if len(mapping) != 4*count {
		panic("bmc: invalid function fault map length")
	}
	q := gia.Gia_ManStart(9 * gia.Gia_ManObjNum(p))
	gia.Gia_ManHashStart(q)
	values := make([]int32, gia.Gia_ManObjNum(p))
	for _, lit := range p.InputLiterals() {
		values[lit>>1] = gia.Gia_ManAppendCi(q)
	}
	index := 0
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		a, b, _, kind := p.Gate(id)
		if kind != 2 {
			continue
		}
		var controls [4]int32
		for i := range controls {
			ctrl := gia.Gia_ManAppendCi(q)
			if mapping[index] != 0 && index+1 < len(mapping) {
				controls[i] = ctrl
			}
			index++
		}
		// The selected row corresponds to the input values satisfying both edges.
		row := int((a&1)^1) + 2*int((b&1)^1)
		controls[row] ^= 1
		lo := gia.Gia_ManHashMux(q, values[a>>1], controls[1], controls[0])
		hi := gia.Gia_ManHashMux(q, values[a>>1], controls[3], controls[2])
		values[id] = gia.Gia_ManHashMux(q, values[b>>1], hi, lo)
	}
	for _, lit := range p.OutputLiterals() {
		gia.Gia_ManAppendCo(q, values[lit>>1]^(lit&1))
	}
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}

// Gia_ManFaultCofactor fixes the prefix of primary inputs to constant values
// while retaining the original input interface for subsequent comparisons.
func Gia_ManFaultCofactor(p *gia.Gia_Man_t, assignments *vec.Vec_Int_t) *gia.Gia_Man_t {
	if p == nil {
		return nil
	}
	if p.RegisterCount() != 0 {
		panic("bmc: fault cofactor requires a combinational graph")
	}
	fixed := vec.Vec_IntArray(assignments)
	inputs := p.InputLiterals()
	if len(fixed) > len(inputs) {
		panic("bmc: too many cofactor assignments")
	}
	q := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	gia.Gia_ManHashStart(q)
	values := make([]int32, gia.Gia_ManObjNum(p))
	for i, lit := range inputs {
		value := gia.Gia_ManAppendCi(q)
		if i < len(fixed) {
			if fixed[i] != 0 && fixed[i] != 1 {
				panic("bmc: cofactor assignment must be 0 or 1")
			}
			value = fixed[i]
		}
		values[lit>>1] = value
	}
	read := func(lit int32) int32 { return values[lit>>1] ^ (lit & 1) }
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		a, b, c, kind := p.Gate(id)
		switch kind {
		case 2:
			values[id] = gia.Gia_ManHashAnd(q, read(a), read(b))
		case 3:
			values[id] = gia.Gia_ManHashXor(q, read(a), read(b))
		case 4:
			values[id] = gia.Gia_ManHashMux(q, read(a), read(b), read(c))
		case 5:
			values[id] = read(a)
		}
	}
	for _, lit := range p.OutputLiterals() {
		gia.Gia_ManAppendCo(q, read(lit))
	}
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}
