package gia

import (
	"github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
)

const nfInfinity int32 = 0x3fffffff

// nfMatch records the selected library cell and its input phase/permutation.
// cutOffset is relative to the object's stored cut set, as in the C mapper.
type nfMatch struct {
	gate, cutOffset int32
	cfg             Nf_Cfg_t
	delay           int32
	flow            float32
	inverted        bool
}
type nfObjectMatches [2][2]nfMatch // output phase, then delay/area

type nfMatcher struct {
	cuts     *nfCutManager
	objects  []nfObjectMatches
	required [][2]int32
	flowRefs [][2]float32
	invDelay int32
	invArea  float32
}

func newNfMatcher(cuts *nfCutManager) *nfMatcher {
	if len(cuts.library.cells) < 4 {
		panic("gia: Nf library lacks constant, buffer or inverter cells")
	}
	p := &nfMatcher{cuts: cuts, objects: make([]nfObjectMatches, len(cuts.sets)), required: make([][2]int32, len(cuts.sets)), flowRefs: make([][2]float32, len(cuts.sets)), invDelay: cuts.library.cells[3].Delay(0), invArea: cuts.library.cells[3].AreaF}
	for i := range p.required {
		p.required[i] = [2]int32{nfInfinity, nfInfinity}
		p.flowRefs[i] = [2]float32{cuts.refs[i] / 2, cuts.refs[i] / 2}
	}
	for phase := 0; phase < 2; phase++ {
		for kind := 0; kind < 2; kind++ {
			p.objects[0][phase][kind] = nfMatch{gate: int32(phase), flow: cuts.library.cells[phase].AreaF}
		}
	}
	return p
}

func (p *nfMatcher) prepareInput(id, arrival int32) {
	p.objects[id] = nfObjectMatches{}
	for kind := 0; kind < 2; kind++ {
		p.objects[id][0][kind] = nfMatch{delay: arrival}
		p.objects[id][1][kind] = nfMatch{delay: arrival + p.invDelay, flow: p.invArea, inverted: true}
	}
}

func (p *nfMatcher) matchCut(id, offset int32, cut []int32) {
	count, function := Nf_CutSize(cut), Nf_CutFunc(cut)
	if count == 0 {
		for phase := 0; phase < 2; phase++ {
			gate := int32(phase) ^ (function & 1)
			m := nfMatch{gate: gate, cutOffset: offset, flow: p.cuts.library.cells[gate].AreaF}
			p.objects[id][phase] = [2]nfMatch{m, m}
		}
		return
	}
	if int(function>>1) >= len(p.cuts.library.matches) {
		return
	}
	leaves := Nf_CutLeaves(cut)
	for _, candidate := range p.cuts.library.matches[function>>1] {
		cell := &p.cuts.library.cells[candidate.gate]
		if cell.Fanins() != count {
			panic("gia: inconsistent Nf cell match")
		}
		phase := int(candidate.cfg.fCompl) ^ int(function&1)
		required := p.required[id][phase]
		bestD, bestA := &p.objects[id][phase][0], &p.objects[id][phase][1]
		delay, flow := int32(0), cell.AreaF
		valid := true
		for pin := int32(0); pin < count; pin++ {
			variable, compl := Nf_CfgVar(candidate.cfg, pin), Nf_CfgCompl(candidate.cfg, pin)
			if variable < 0 || variable >= count {
				panic("gia: invalid Nf match permutation")
			}
			fanin := &p.objects[leaves[variable]][compl]
			cellDelay := cell.Delay(int(pin))
			selected := fanin[0]
			if required != nfInfinity && fanin[1].delay+cellDelay <= required {
				selected = fanin[1]
			} else if bestD.delay < nfInfinity && bestA.delay < nfInfinity && selected.delay+cellDelay > required {
				valid = false
				break
			}
			delay = max(delay, selected.delay+cellDelay)
			if flow >= 1e32 || selected.flow >= 1e32 {
				flow = 1e32
			} else {
				flow += selected.flow
			}
		}
		if !valid {
			continue
		}
		cfg := candidate.cfg
		cfg.fCompl = 0
		result := nfMatch{gate: candidate.gate, cutOffset: offset, cfg: cfg, delay: delay, flow: flow}
		if bestD.delay > delay {
			*bestD = result
		}
		if bestA.flow > flow+NF_EPSILON {
			*bestA = result
		}
	}
}

func (p *nfMatcher) matchObject(id int32) {
	p.objects[id] = nfObjectMatches{}
	for phase := 0; phase < 2; phase++ {
		for kind := 0; kind < 2; kind++ {
			p.objects[id][phase][kind] = nfMatch{delay: nfInfinity, flow: math.MaxFloat32}
		}
	}
	data := p.cuts.storage.at(p.cuts.sets[id])
	offset := int32(1)
	for i := int32(0); i < data[0]; i++ {
		cut := data[offset:]
		p.matchCut(id, offset, cut)
		offset += Nf_CutSize(cut) + 1
	}
	for phase := 0; phase < 2; phase++ {
		for kind := 0; kind < 2; kind++ {
			p.objects[id][phase][kind].flow /= p.flowRefs[id][phase]
		}
	}
	// An inverter can provide an output phase absent from the library or improve
	// the fastest implementation of that phase.
	for phase := 0; phase < 2; phase++ {
		own, other := &p.objects[id][phase][0], p.objects[id][phase^1][0]
		if other.delay < nfInfinity && own.delay > other.delay+p.invDelay {
			*own = other
			own.delay += p.invDelay
			own.flow += p.invArea
			own.inverted = true
			if p.objects[id][phase][1].delay == nfInfinity {
				p.objects[id][phase][1] = *own
			}
		}
	}
}

// computeMapping performs the first timing/area-flow match pass over the graph.
// Subsequent required-time and exact-area rounds are handled separately.
func (p *nfMatcher) computeMapping() {
	ciIndex := 0
	for _, obj := range p.cuts.graph.pObjs {
		id := obj.id
		if Gia_ObjIsCi(obj) != 0 {
			var arrival int32
			if scl.Scl_ConIsRunning() {
				arrival = scl.Scl_ConGetInArr(ciIndex)
			} else if p.cuts.graph.VInArrs != nil {
				arrival = scl.Scl_Flt2Int(vec.Vec_FltEntry(p.cuts.graph.VInArrs, ciIndex))
			}
			ciIndex++
			p.prepareInput(id, arrival)
			continue
		}
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		if Gia_ObjIsBuf(obj) != 0 {
			fanin := Gia_ObjFaninId0(obj, id)
			base := p.objects[fanin][Gia_ObjFaninC0(obj)][0]
			direct := nfMatch{delay: base.delay, flow: base.flow}
			inverted := nfMatch{delay: base.delay + p.invDelay, flow: base.flow + p.invArea, inverted: true}
			p.objects[id][0] = [2]nfMatch{direct, direct}
			p.objects[id][1] = [2]nfMatch{inverted, inverted}
		} else {
			p.matchObject(id)
		}
	}
}
