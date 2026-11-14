package gia

import "fmt"

// These dependency-neutral methods expose the combinational region to CEC and
// other algorithms without making them import the GIA implementation package.
func (p *Gia_Man_t) InputLiterals() []int32 {
	if p.vCis == nil {
		return nil
	}
	ids := Vec_IntArray(p.vCis)
	out := make([]int32, len(ids))
	for i, id := range ids {
		out[i] = id << 1
	}
	return out
}
func (p *Gia_Man_t) OutputLiterals() []int32 {
	if p.vCos == nil {
		return nil
	}
	ids := Vec_IntArray(p.vCos)
	out := make([]int32, len(ids))
	for i, id := range ids {
		if id < 0 || id >= p.nObjs || int(id) >= len(p.pObjs) || p.pObjs[id] == nil || Gia_ObjIsCo(p.pObjs[id]) == 0 {
			out[i] = -1
			continue
		}
		out[i] = Gia_ObjFaninLit0(p.pObjs[id], id)
	}
	return out
}
func (p *Gia_Man_t) RegisterCount() int { return int(p.nRegs) }

// Gate kinds: -1 absent, 0 false, 1 input, 2 AND, 3 XOR, 4 MUX,
// 5 buffer, 6 output. MUX operands are control, then, else.
func (p *Gia_Man_t) Gate(id int32) (a, b, c int32, kind int) {
	if id < 0 || id >= p.nObjs || int(id) >= len(p.pObjs) || p.pObjs[id] == nil {
		return 0, 0, 0, -1
	}
	obj := p.pObjs[id]
	if id == 0 {
		return 0, 0, 0, 0
	}
	if Gia_ObjIsCi(obj) != 0 {
		return 0, 0, 0, 1
	}
	a = Gia_ObjFaninLit0(obj, id)
	if Gia_ObjIsCo(obj) != 0 {
		return a, 0, 0, 6
	}
	b = Gia_ObjFaninLit1(obj, id)
	if int(id) < len(p.PMuxes) && p.PMuxes[id] != 0 {
		return int32(p.PMuxes[id]), b, a, 4
	}
	if Gia_ObjIsXor(obj) != 0 {
		return a, b, 0, 3
	}
	if Gia_ObjIsBuf(obj) != 0 && obj.fCompl0 == obj.fCompl1 {
		return a, 0, 0, 5
	}
	return a, b, 0, 2
}

// Evaluate simulates the combinational region: register outputs are additional
// inputs, and register inputs are additional outputs, both in source order.
func (p *Gia_Man_t) Evaluate(inputs []bool) ([]bool, error) {
	_, outputs, err := p.evaluateObjects(inputs)
	return outputs, err
}

func (p *Gia_Man_t) evaluateObjects(inputs []bool) ([]bool, []bool, error) {
	literals := p.InputLiterals()
	if len(inputs) != len(literals) {
		return nil, nil, fmt.Errorf("gia: expected %d inputs, got %d", len(literals), len(inputs))
	}
	if p.nObjs < 1 || len(p.pObjs) != int(p.nObjs) || p.pObjs[0] == nil || Gia_ObjIsConst0(p.pObjs[0]) == 0 || p.pObjs[0].fTerm != 0 {
		return nil, nil, fmt.Errorf("gia: invalid object storage")
	}
	values := make([]bool, p.nObjs)
	seeded := make([]bool, p.nObjs)
	for i, lit := range literals {
		if lit < 0 || lit&1 != 0 || lit>>1 >= p.nObjs || seeded[lit>>1] {
			return nil, nil, fmt.Errorf("gia: invalid or duplicate input literal")
		}
		if _, _, _, kind := p.Gate(lit >> 1); kind != 1 {
			return nil, nil, fmt.Errorf("gia: input vector references non-input object")
		}
		values[lit>>1] = inputs[i]
		seeded[lit>>1] = true
	}
	read := func(lit, id int32) (bool, error) {
		if lit < 0 || lit>>1 >= id {
			return false, fmt.Errorf("gia: invalid backward edge at object %d", id)
		}
		return values[lit>>1] != (lit&1 != 0), nil
	}
	for id := int32(1); id < p.nObjs; id++ {
		a, b, c, kind := p.Gate(id)
		if kind == 1 {
			if !seeded[id] {
				return nil, nil, fmt.Errorf("gia: unbound input object %d", id)
			}
			continue
		}
		if kind < 2 || kind > 6 {
			return nil, nil, fmt.Errorf("gia: invalid gate kind at object %d", id)
		}
		x, err := read(a, id)
		if err != nil {
			return nil, nil, err
		}
		switch kind {
		case 5, 6:
			values[id] = x
		case 2, 3:
			y, err := read(b, id)
			if err != nil {
				return nil, nil, err
			}
			if kind == 2 {
				values[id] = x && y
			} else {
				values[id] = x != y
			}
		case 4:
			yes, err := read(b, id)
			if err != nil {
				return nil, nil, err
			}
			no, err := read(c, id)
			if err != nil {
				return nil, nil, err
			}
			if x {
				values[id] = yes
			} else {
				values[id] = no
			}
		}
	}
	outputs := p.OutputLiterals()
	result := make([]bool, len(outputs))
	for i, lit := range outputs {
		value, err := read(lit, p.nObjs)
		if err != nil {
			return nil, nil, err
		}
		result[i] = value
	}
	return values, result, nil
}

// AndFaninLiterals exposes ordinary AND/buffer dependencies to adapters that
// cannot represent specialized XOR/MUX nodes. Complements remain in low bits.
func (p *Gia_Man_t) AndFaninLiterals(id int32) (a, b int32, ok bool) {
	a, b, _, kind := p.Gate(id)
	if kind == 5 {
		return a, a, true
	}
	return a, b, kind == 2
}
func (p *Gia_Man_t) OutputFaninLiterals() []int32 { return p.OutputLiterals() }
