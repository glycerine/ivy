package sim

import "github.com/glycerine/abcgo/aig/aig"

// SequentialSimulation allows explicit initial 0/1/unknown latch values.
// Initial nil means all zero (the native Graph convention); -1 means random.
// Info rows contain consecutive words for each frame, indexed by node ID.
func SequentialSimulation(g *aig.Graph, frames, words int, model []bool, initial []int8) Info {
	if frames < 0 || words <= 0 {
		panic("invalid sequential simulation dimensions")
	}
	order := graphOrder(g)
	inputs, outputs := g.Inputs(), g.Outputs()
	regs := int(aig.Aig_ManRegNum(g))
	pis := len(inputs) - regs
	if model != nil && (words != 1 || len(model) != frames*pis) {
		panic("sequential model dimensions mismatch")
	}
	if initial != nil && len(initial) != regs {
		panic("initial state dimensions mismatch")
	}
	info := Sim_UtilInfoAlloc(int(aig.Aig_ManObjNumMax(g)), frames*words, true)
	rng := newRandom()
	Sim_UtilSetConst(info[0], frames*words, true)
	for i, e := range inputs[:pis] {
		if model == nil {
			fillRandom(info[e.NodeID()], rng)
		} else {
			for f := 0; f < frames; f++ {
				Sim_UtilSetConst(info[e.NodeID()][f:f+1], 1, model[f*pis+i])
			}
		}
	}
	if frames == 0 {
		return info
	}
	for i, e := range inputs[pis:] {
		init := int8(0)
		if initial != nil {
			init = initial[i]
		}
		switch init {
		case 0:
		case 1:
			Sim_UtilSetConst(info[e.NodeID()], words, true)
		case -1:
			fillRandom(info[e.NodeID()][:words], rng)
		default:
			panic("invalid initial state")
		}
	}
	for f := 0; f < frames; f++ {
		simulateSeqFrame(info, g, order, inputs, outputs, f, words, f+1 < frames)
	}
	return info
}
func Sim_SimulateSeqRandom(g *aig.Graph, frames, words int) Info {
	return SequentialSimulation(g, frames, words, nil, nil)
}
func Sim_SimulateSeqModel(g *aig.Graph, frames int, model []bool) Info {
	if model == nil && frames > 0 && g.InputCount() > int(aig.Aig_ManRegNum(g)) {
		panic("missing sequential model")
	}
	return SequentialSimulation(g, frames, 1, model, nil)
}
func Sim_SimulateSeqFrame(info Info, g *aig.Graph, frame, words int, transfer bool) {
	simulateSeqFrame(info, g, graphOrder(g), g.Inputs(), g.Outputs(), frame, words, transfer)
}
func simulateSeqFrame(info Info, g *aig.Graph, order, inputs, outputs []aig.Edge, frame, words int, transfer bool) {
	off := frame * words
	for _, e := range order {
		Sim_UtilSimulateNodeOne(g, e, info, words, off)
	}
	if !transfer {
		return
	}
	regs := int(aig.Aig_ManRegNum(g))
	for i := 0; i < regs; i++ {
		src := info[outputs[len(outputs)-regs+i].NodeID()][off : off+words]
		dst := info[inputs[len(inputs)-regs+i].NodeID()][off+words : off+2*words]
		copy(dst, src)
	}
}
