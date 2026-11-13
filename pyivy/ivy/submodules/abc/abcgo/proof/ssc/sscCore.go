package ssc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec"
	"github.com/glycerine/abcgo/proof/cec/core"
	"github.com/glycerine/abcgo/sat/satoko"
)

// StartStatus distinguishes an empty care set from an exhausted SAT budget.
// This replaces C's invalid pointer sentinel without collapsing the outcomes.
func Ssc_ManStart(subject, care *gia.Gia_Man_t, params *Ssc_Pars_t) (*Ssc_Man_t, int) {
	if subject == nil || care == nil || len(subject.InputLiterals()) != len(care.InputLiterals()) || care.RegisterCount() != 0 {
		panic("ssc: incompatible care interface")
	}
	for _, network := range []*gia.Gia_Man_t{subject, care} {
		if _, err := core.NewModel(network); err != nil {
			panic(err)
		}
	}
	v := DefaultParams()
	if params != nil {
		v = *params
	}
	if v.NWords < 1 || v.BTLimit < 0 || v.SatVarMax < 0 || v.CallsRecycle < 0 {
		panic("ssc: invalid parameters")
	}
	fraig := newBuilder(len(subject.InputLiterals()), subject.ObjectCount()+care.ObjectCount())
	for _, lit := range copyInto(fraig, care, nil) {
		gia.Gia_ManAppendCo(fraig.graph, lit^1)
	}
	p := &Ssc_Man_t{Params: v, Aig: fraig.graph, Original: subject, Care: care, fraig: fraig}
	Ssc_ManStartSolver(p)
	pivot, status := Ssc_ManFindPivotSat(p)
	if status != satoko.SATOKO_SAT {
		return p, status
	}
	p.Pivot = pivot
	p.Sim = NewSimulation(subject, int(v.NWords))
	p.Classes = Ssc_GiaClassesInit(subject, pivot)
	return p, status
}
func Ssc_ManStop(p *Ssc_Man_t) {
	if p != nil {
		*p = Ssc_Man_t{}
	}
}
func Ssc_ManPrintStats(p *Ssc_Man_t) {
	fmt.Printf("SSC: %d queries, %d proofs, %d counterexamples, %d undecided, %d recycles, %d simulation rounds\n", p.Stats.Calls, p.Stats.Proofs, p.Stats.Counterexamples, p.Stats.Undecided, p.Stats.Recycles, p.Stats.SimulationRounds)
}
func Ssc_GiaResimulateOneClass(p *Ssc_Man_t, head, node int32) int32 {
	if node < 0 || int(node) >= len(p.Classes.Repr) || p.Classes.Repr[node] != head {
		panic("ssc: resimulation pair is not in the same class")
	}
	s := NewSimulation(p.Original, 1)
	Ssc_GiaSavePiPattern(s, p.Pattern)
	Ssc_GiaSimRound(s)
	values := make([]bool, len(s.Values))
	for i := range values {
		values[i] = s.Values[i][0]&1 != 0
	}
	if p.mapping != nil {
		f := NewSimulation(p.Aig, 1)
		Ssc_GiaSavePiPattern(f, p.Pattern)
		Ssc_GiaSimRound(f)
		done := make([]bool, len(values))
		var visit func(int32) bool
		visit = func(id int32) bool {
			if done[id] {
				return values[id]
			}
			done[id] = true
			if lit := p.mapping[id]; lit >= 0 {
				values[id] = (f.Values[lit>>1][0]&1 != 0) != (lit&1 != 0)
				return values[id]
			}
			a, b, c, kind := p.Original.Gate(id)
			read := func(lit int32) bool { return visit(lit>>1) != (lit&1 != 0) }
			switch kind {
			case 2:
				values[id] = read(a) && read(b)
			case 3:
				values[id] = read(a) != read(b)
			case 4:
				if read(a) {
					values[id] = read(b)
				} else {
					values[id] = read(c)
				}
			case 5, 6:
				values[id] = read(a)
			}
			return values[id]
		}
		for id := range values {
			visit(int32(id))
		}
	}
	refined := Ssc_GiaSimClassRefineOneBit(p.Classes, values, head)
	if p.Classes.Repr[node] == head {
		panic("ssc: SAT witness did not separate candidate pair")
	}
	return refined
}
func carePatterns(p *Ssc_Man_t) {
	care := NewSimulation(p.Care, 5)
	candidateCount := gia.Gia_ManCiNum(p.Original) + gia.Gia_ManAndNum(p.Original)
	for {
		Ssc_GiaRandomPiPattern(care, 5, nil)
		Ssc_GiaSimRound(care)
		mask := make([]uint64, 5)
		for i := range mask {
			mask[i] = ^uint64(0)
		}
		for _, lit := range p.Care.OutputLiterals() {
			for w := range mask {
				x := care.Values[lit>>1][w]
				if lit&1 != 0 {
					x = ^x
				}
				mask[w] &= ^x
			}
		}
		hits := Ssc_SimCountBits(mask, 5)
		if hits == 0 {
			break
		}
		Ssc_GiaResetPiPattern(p.Sim, 5)
		for i := range p.Sim.Inputs {
			for w := range mask {
				x := care.Inputs[i][w] & mask[w]
				if p.Pivot[i] {
					x |= ^mask[w]
				}
				p.Sim.Inputs[i][w] = x
			}
		}
		Ssc_GiaSimRound(p.Sim)
		p.Stats.SimulationRounds++
		if Ssc_GiaClassesRefine(p.Classes, p.Sim) <= candidateCount/100 {
			break
		}
	}
}
func Ssc_PerformSweepingInt(subject, care *gia.Gia_Man_t, params *Ssc_Pars_t) *gia.Gia_Man_t {
	// C SSC operates on ordinary AND-inverter graphs. Lower the extended GIA
	// gate kinds at the boundary so the source reducer sees the same contract.
	if hasNativeGates(subject) {
		subject = gia.Gia_ManDupNoMuxes(subject, 1)
	}
	if hasNativeGates(care) {
		care = gia.Gia_ManDupNoMuxes(care, 1)
	}
	gia.Gia_ManRandom(1)
	p, status := Ssc_ManStart(subject, care, params)
	if status == satoko.SATOKO_UNSAT {
		return copyOutputs(subject, nil, true)
	}
	if status != satoko.SATOKO_SAT {
		return copyOutputs(subject, nil, false)
	}
	carePatterns(p)
	// The original solver's working graph starts with the care cone and grows
	// as subject candidates are visited. Later SAT queries must see earlier
	// proved substitutions, not the unreduced subject graph.
	fraig := p.fraig
	mapping := make([]int32, subject.ObjectCount())
	for i := range mapping {
		mapping[i] = -1
	}
	mapping[0] = 0
	for i, lit := range subject.InputLiterals() {
		mapping[lit>>1] = fraig.graph.InputLiterals()[i]
	}
	p.mapping = mapping
	// The C sweep saves a pivot and batches counterexamples until a simulation
	// word is full, refining any remaining candidates at the end of the pass.
	Ssc_GiaResetPiPattern(p.Sim, int(p.Params.NWords))
	Ssc_GiaSavePiPattern(p.Sim, p.Pivot)
	flushPatterns := func() {
		for p.Sim.Patterns < 64*p.Sim.Words {
			Ssc_GiaSavePiPattern(p.Sim, p.Pivot)
		}
		Ssc_GiaSimRound(p.Sim)
		Ssc_GiaClassesRefine(p.Classes, p.Sim)
		Ssc_GiaResetPiPattern(p.Sim, int(p.Params.NWords))
		Ssc_GiaSavePiPattern(p.Sim, p.Pivot)
	}
	subject.StartRepresentatives()
	for id := int32(1); int(id) < subject.ObjectCount(); id++ {
		a, b, c, kind := subject.Gate(id)
		if kind < 1 || kind > 5 {
			continue
		}
		read := func(lit int32) int32 { return mapping[lit>>1] ^ (lit & 1) }
		switch kind {
		case 2:
			mapping[id] = fraig.and(read(a), read(b))
		case 3:
			mapping[id] = fraig.xor(read(a), read(b))
		case 4:
			mapping[id] = fraig.mux(read(a), read(b), read(c))
		case 5:
			mapping[id] = read(a)
		}
		if p.Sim.Patterns == 64*p.Sim.Words {
			flushPatterns()
		}
		if p.Classes.Repr[id] >= 0 {
			repr := p.Classes.Repr[id]
			complement := int32(0)
			if p.Classes.Phase[id] != p.Classes.Phase[repr] {
				complement = 1
			}
			mappedRepr, mappedNode := mapping[repr], mapping[id]
			if mappedRepr == mappedNode^complement {
				gia.Gia_ObjSetRepr(subject, id, repr)
				gia.Gia_ObjSetProved(subject, id)
				continue
			}
			if mappedRepr>>1 > mappedNode>>1 || mappedRepr>>1 == mappedNode>>1 {
				continue
			}
			fraigComplement := complement ^ (mappedRepr & 1) ^ (mappedNode & 1)
			result := Ssc_ManCheckEquivalence(p, mappedRepr>>1, mappedNode>>1, fraigComplement)
			if result == satoko.SATOKO_UNSAT {
				mapping[id] = mapping[repr] ^ complement
				gia.Gia_ObjSetRepr(subject, id, repr)
				gia.Gia_ObjSetProved(subject, id)
			} else if result == satoko.SATOKO_SAT {
				Ssc_GiaSavePiPattern(p.Sim, p.Pattern)
				if repr > 0 {
					Ssc_GiaResimulateOneClass(p, repr, id)
				}
			}
		}
	}
	if p.Sim.Patterns > 1 {
		flushPatterns()
	}
	// The C class links include candidates that remain unproved; proved flags
	// separately control which links the equivalence reducer may use. Constant
	// candidates have representative zero but are not a linked class.
	for id := int32(1); int(id) < subject.ObjectCount(); id++ {
		gia.Gia_ObjUnsetRepr(subject, id)
		gia.Gia_ObjSetNext(subject, id, 0)
	}
	for _, group := range p.Classes.Groups {
		for i := 1; i < len(group); i++ {
			gia.Gia_ObjSetRepr(subject, group[i], group[0])
			if group[0] != 0 {
				gia.Gia_ObjSetNext(subject, group[i-1], group[i])
			}
		}
	}
	// C derives the final graph from the proved representative classes and
	// then removes dangling nodes, rather than rebuilding directly from the SAT
	// query mappings.
	result := gia.Gia_ManEquivReduce(subject, 0, 0, 1, 0)
	if result == nil {
		result = gia.Gia_ManDup(subject)
	}
	result = gia.Gia_ManCleanup(result)
	if p.Params.Verbose {
		Ssc_ManPrintStats(p)
	}
	return result
}

func hasNativeGates(g *gia.Gia_Man_t) bool {
	for id := int32(0); int(id) < g.ObjectCount(); id++ {
		_, _, _, kind := g.Gate(id)
		if kind >= 3 && kind <= 5 {
			return true
		}
	}
	return false
}
func Ssc_PerformSweeping(subject, care *gia.Gia_Man_t, params *Ssc_Pars_t) *gia.Gia_Man_t {
	result := Ssc_PerformSweepingInt(subject, care, params)
	if params != nil && params.Verify && Ssc_PerformVerification(subject, result, care) == 0 {
		panic("ssc: conditional equivalence verification failed")
	}
	return result
}
func Ssc_PerformVerification(a, b, care *gia.Gia_Man_t) int32 {
	if len(a.OutputLiterals()) != len(b.OutputLiterals()) || len(a.InputLiterals()) != len(b.InputLiterals()) {
		panic("ssc: verification interface mismatch")
	}
	condition := gia.Gia_ManDupAndOr(care, gia.Gia_ManPoNum(a), 1, 0)
	left := gia.Gia_ManMiter(a, condition, 0, 0, 0, 1, 0)
	if left == nil {
		return -1
	}
	right := gia.Gia_ManMiter(b, condition, 0, 0, 0, 1, 0)
	if right == nil {
		return -1
	}
	miter := gia.Gia_ManMiter(left, right, 0, 1, 0, 0, 0)
	if miter == nil {
		return -1
	}
	params := new(cec.Cec_ParCec_t)
	cec.Cec_ManCecSetDefaultParams(params)
	return cec.Cec_ManVerify(miter, params)
}
func emptyCare(p *gia.Gia_Man_t) *gia.Gia_Man_t {
	q := gia.Gia_ManStart(int32(len(p.InputLiterals()) + 2))
	for range p.InputLiterals() {
		gia.Gia_ManAppendCi(q)
	}
	gia.Gia_ManAppendCo(q, 0)
	return q
}
func Ssc_PerformSweepingConstr(p *gia.Gia_Man_t, params *Ssc_Pars_t) *gia.Gia_Man_t {
	pos := len(p.OutputLiterals()) - p.RegisterCount()
	constraints := int(gia.Gia_ManConstrNum(p))
	if constraints < 0 || constraints > pos {
		panic("ssc: invalid constraint count")
	}
	ordinary := append([]int32(nil), p.OutputLiterals()[:pos-constraints]...)
	ordinary = append(ordinary, p.OutputLiterals()[pos:]...)
	subject := copyOutputs(p, ordinary, false)
	care := copyWithRegs(p, p.OutputLiterals()[pos-constraints:pos], nil, 0)
	if constraints == 0 {
		care = emptyCare(subject)
	}
	result := Ssc_PerformSweeping(subject, care, params)
	if params != nil && params.Append {
		outputs := result.OutputLiterals()
		q := newBuilder(len(result.InputLiterals()), result.ObjectCount()+care.ObjectCount())
		r := copyInto(q, result, nil)
		c := copyInto(q, care, nil)
		n := len(outputs) - result.RegisterCount()
		for _, v := range r[:n] {
			gia.Gia_ManAppendCo(q.graph, v)
		}
		for _, v := range c {
			gia.Gia_ManAppendCo(q.graph, v)
		}
		for _, v := range r[n:] {
			gia.Gia_ManAppendCo(q.graph, v)
		}
		gia.Gia_ManSetRegNum(q.graph, int32(result.RegisterCount()))
		gia.Gia_ManSetConstrNum(q.graph, int32(len(c)))
		q.graph.CopyIdentityFrom(p)
		result = q.graph
	}
	return result
}
