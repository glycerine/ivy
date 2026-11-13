package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Bmc_PerformFindFlopOrder visits registers backward from primary outputs.
func Bmc_PerformFindFlopOrder(p *gia.Gia_Man_t, regs *vec.Vec_Int_t) {
	vec.Vec_IntClear(regs)
	inputs, outputs := p.InputLiterals(), p.OutputLiterals()
	npi, npo := int(gia.Gia_ManPiNum(p)), int(gia.Gia_ManPoNum(p))
	seen := make([]bool, gia.Gia_ManObjNum(p))
	regIndex := map[int32]int32{}
	for i, lit := range inputs[npi:] {
		regIndex[lit>>1] = int32(i)
	}
	var visit func(int32)
	visit = func(lit int32) {
		id := lit >> 1
		if seen[id] {
			return
		}
		seen[id] = true
		if i, ok := regIndex[id]; ok {
			vec.Vec_IntPush(regs, i)
			return
		}
		a, b, c, k := p.Gate(id)
		switch k {
		case 2, 3:
			visit(a)
			visit(b)
		case 4:
			visit(a)
			visit(b)
			visit(c)
		case 5, 6:
			visit(a)
		}
	}
	for _, lit := range outputs[:npo] {
		visit(lit)
	}
	for i := 0; i < vec.Vec_IntSize(regs); i++ {
		visit(outputs[npo+int(vec.Vec_IntEntry(regs, i))])
	}
	for i, lit := range inputs[npi:] {
		if !seen[lit>>1] {
			vec.Vec_IntPush(regs, int32(i))
		}
	}
}
