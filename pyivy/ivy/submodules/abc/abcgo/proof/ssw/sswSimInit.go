package ssw

import "github.com/glycerine/abcgo/misc/vec"

func Ssw_SmlInitialize(p *Ssw_Sml_t, init int32) {
	inputs := p.pAig.Inputs()
	nRegs := int(p.pAig.RegisterCount())
	if init != 0 {
		if nRegs <= 0 || nRegs > len(inputs) {
			panic("ssw: expected sequential graph")
		}
		for _, obj := range inputs[:len(inputs)-nRegs] {
			Ssw_SmlAssignRandom(p, obj)
		}
		for _, obj := range inputs[len(inputs)-nRegs:] {
			Ssw_SmlObjAssignConst(p, obj, 0, 0)
		}
	} else {
		for _, obj := range inputs {
			Ssw_SmlAssignRandom(p, obj)
		}
	}
}
func Ssw_SmlInitializeSpecial(p *Ssw_Sml_t, init *vec.Vec_Int_t) {
	inputs := p.pAig.Inputs()
	nRegs := int(p.pAig.RegisterCount())
	if nRegs <= 0 || nRegs > len(inputs) || vec.Vec_IntSize(init) != nRegs*int(p.nWordsFrame) {
		panic("ssw: invalid initial state")
	}
	for _, obj := range inputs[:len(inputs)-nRegs] {
		Ssw_SmlAssignRandom(p, obj)
	}
	for i := 0; i < vec.Vec_IntSize(init); i++ {
		Ssw_SmlObjAssignConstWord(p, inputs[len(inputs)-nRegs+i%nRegs], vec.Vec_IntEntry(init, i), 0, int32(i/nRegs))
	}
}
func Ssw_SmlReinitialize(p *Ssw_Sml_t) {
	inputs, outputs := p.pAig.Inputs(), p.pAig.Outputs()
	nRegs := int(p.pAig.RegisterCount())
	if nRegs <= 0 || nRegs >= len(inputs) {
		panic("ssw: invalid sequential graph")
	}
	for _, obj := range inputs[:len(inputs)-nRegs] {
		Ssw_SmlAssignRandom(p, obj)
	}
	for i := 0; i < nRegs; i++ {
		Ssw_SmlNodeTransferFirst(p, outputs[len(outputs)-nRegs+i], inputs[len(inputs)-nRegs+i])
	}
}
