// Sequential signal/latch correspondence: speculative time-frame reduction,
// incremental SAT witnesses, packed resimulation, and class refinement.
package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"slices"
	"time"
)

type corrBuilder struct {
	graph *Gia_Man_t
	ands  map[[2]int32]int32
}

func newCorrBuilder(source *Gia_Man_t) *corrBuilder {
	p := gia.Gia_ManStart(gia.Gia_ManObjNum(source))
	p.CopyIdentityFrom(source)
	return &corrBuilder{p, make(map[[2]int32]int32)}
}
func (b *corrBuilder) and(a, c int32) int32 {
	if a == 0 || c == 0 || a == c^1 {
		return 0
	}
	if a == 1 {
		return c
	}
	if c == 1 || a == c {
		return a
	}
	if a > c {
		a, c = c, a
	}
	key := [2]int32{a, c}
	if lit, ok := b.ands[key]; ok {
		return lit
	}
	lit := gia.Gia_ManAppendAnd(b.graph, a, c)
	b.ands[key] = lit
	return lit
}
func (b *corrBuilder) xor(a, c int32) int32    { return b.and(b.and(a, c^1)^1, b.and(a^1, c)^1) ^ 1 }
func (b *corrBuilder) mux(c, a, d int32) int32 { return b.and(b.and(c, a)^1, b.and(c^1, d)^1) ^ 1 }
func (b *corrBuilder) gate(kind int, a, c, d int32) int32 {
	switch kind {
	case 2:
		return b.and(a, c)
	case 3:
		return b.xor(a, c)
	case 4:
		return b.mux(a, c, d)
	case 5, 6:
		return a
	}
	panic("cec: invalid correspondence gate")
}

type corrUnroller struct {
	source  *Gia_Man_t
	builder *corrBuilder
	copies  [][]int32
	prefix  int32
	skip    map[int32]bool
}

func (u *corrUnroller) copy(id, frame int32) int32 {
	if lit := u.copies[frame][id]; lit >= 0 {
		return lit
	}
	if frame >= u.prefix && u.source.HasRepresentatives() && !u.skip[id] {
		repr := gia.Gia_ObjRepr(u.source, id)
		if repr != gia.GIA_VOID {
			lit := u.copy(repr, frame) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(u.source, repr)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(u.source, id))
			u.copies[frame][id] = lit
			return lit
		}
	}
	lit := u.real(id, frame)
	u.copies[frame][id] = lit
	return lit
}
func (u *corrUnroller) real(id, frame int32) int32 {
	obj := gia.Gia_ManObj(u.source, id)
	if gia.Gia_ObjIsCi(obj) != 0 {
		if frame == 0 {
			return u.copies[0][id]
		}
		reg := gia.Gia_ObjCioId(obj) - gia.Gia_ManPiNum(u.source)
		if reg < 0 {
			return u.copies[frame][id]
		}
		lit := u.source.OutputLiterals()[gia.Gia_ManPoNum(u.source)+reg]
		return u.copy(lit>>1, frame-1) ^ lit&1
	}
	a, b, c, kind := u.source.Gate(id)
	mapped := func(lit int32) int32 { return u.copy(lit>>1, frame) ^ lit&1 }
	av := mapped(a)
	bv, cv := int32(0), int32(0)
	if kind == 2 || kind == 3 || kind == 4 {
		bv = mapped(b)
	}
	if kind == 4 {
		cv = mapped(c)
	}
	return u.builder.gate(kind, av, bv, cv)
}
func cecCorrUnroller(p *Gia_Man_t, frames, prefix int32, zero bool, state *Abc_Cex_t) *corrUnroller {
	if frames <= 0 {
		panic("cec: nonpositive correspondence frame count")
	}
	u := &corrUnroller{source: p, builder: newCorrBuilder(p), prefix: prefix, copies: make([][]int32, frames)}
	gia.Gia_ManSetPhase(p)
	for f := range u.copies {
		u.copies[f] = make([]int32, gia.Gia_ManObjNum(p))
		for i := range u.copies[f] {
			u.copies[f][i] = -1
		}
		u.copies[f][0] = 0
	}
	for r := int32(0); r < gia.Gia_ManRegNum(p); r++ {
		id := gia.Gia_ObjId(p, gia.Gia_ManRo(p, r))
		lit := int32(0)
		if state == nil {
			lit = gia.Gia_ManAppendCi(u.builder.graph)
		}
		if zero {
			lit = 0
		}
		if state != nil {
			lit = int32(cecBit(state.PData, r))
		}
		u.copies[0][id] = lit
	}
	if !zero && state == nil && p.HasRepresentatives() {
		for r := int32(0); r < gia.Gia_ManRegNum(p); r++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManRo(p, r))
			repr := gia.Gia_ObjRepr(p, id)
			if repr != gia.GIA_VOID {
				u.copies[0][id] = u.copies[0][repr]
			}
		}
	}
	for f := int32(0); f < frames; f++ {
		for pi := int32(0); pi < gia.Gia_ManPiNum(p); pi++ {
			u.copies[f][gia.Gia_ObjId(p, gia.Gia_ManPi(p, pi))] = gia.Gia_ManAppendCi(u.builder.graph)
		}
	}
	return u
}
func Gia_ManCorrSpecReduce(p *Gia_Man_t, frames, scorr int32, outputs **Vec_Int_t, rings int32) *Gia_Man_t {
	if frames <= 0 || gia.Gia_ManRegNum(p) == 0 || !p.HasRepresentatives() {
		panic("cec: invalid correspondence input")
	}
	u := cecCorrUnroller(p, frames+scorr, 0, false, nil)
	pairs := vec.Vec_IntAlloc(1000)
	lits := []int32{}
	add := func(a, b, x, y int32, ring bool) {
		if x == y {
			return
		}
		lit := int32(0)
		if ring {
			if x == 0 || y == 1 {
				return
			}
			lit = u.builder.and(x, y^1)
		} else {
			lit = u.builder.xor(x, y)
		}
		vec.Vec_IntPush(pairs, a)
		vec.Vec_IntPush(pairs, b)
		lits = append(lits, lit)
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		repr := gia.Gia_ObjRepr(p, id)
		if rings != 0 {
			if repr == 0 {
				lit := u.real(id, frames) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p, id))
				if lit != 0 {
					vec.Vec_IntPush(pairs, 0)
					vec.Vec_IntPush(pairs, id)
					lits = append(lits, lit)
				}
			} else if gia.Gia_ObjIsHead(p, id) != 0 {
				previous := id
				seen := make(map[int32]bool)
				for other := p.Next(id); ; other = p.Next(other) {
					if other != 0 && seen[other] {
						panic("cec: cyclic correspondence class")
					}
					seen[other] = true
					if other == 0 {
						other = id
					}
					phase := gia.Gia_ObjPhase(gia.Gia_ManObj(p, id))
					x := u.real(previous, frames) ^ phase ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p, previous))
					y := u.real(other, frames) ^ phase ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p, other))
					add(previous, other, x, y, true)
					previous = other
					if other == id {
						break
					}
				}
			}
		} else if repr != gia.GIA_VOID {
			x := int32(0)
			if repr != 0 {
				x = u.real(repr, frames)
			}
			y := u.real(id, frames) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p, repr)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p, id))
			add(repr, id, x, y, false)
		}
	}
	for _, lit := range lits {
		gia.Gia_ManAppendCo(u.builder.graph, lit)
	}
	if outputs != nil {
		*outputs = pairs
	}
	result, _ := cecCorrDuplicate(u.builder.graph, false, false, nil)
	return result
}
func Gia_ManCorrSpecReduceInit(p *Gia_Man_t, frames, prefix, scorr int32, outputs **Vec_Int_t, rings int32) *Gia_Man_t {
	return cecCorrUnrollInit(p, frames, prefix, scorr, outputs, rings, nil)
}
func cecCorrUnrollInit(p *Gia_Man_t, frames, prefix, scorr int32, outputs **Vec_Int_t, rings int32, state *Abc_Cex_t) *Gia_Man_t {
	u := cecCorrUnroller(p, frames+prefix+scorr, prefix, true, state)
	if state != nil && rings != 0 {
		colors := make([]uint8, gia.Gia_ManObjNum(p))
		var mark func(int32, uint8)
		mark = func(id int32, color uint8) {
			if id == 0 || colors[id]&color != 0 {
				return
			}
			colors[id] |= color
			a, b, c, kind := p.Gate(id)
			if kind < 2 {
				return
			}
			mark(a>>1, color)
			if kind == 2 || kind == 3 || kind == 4 {
				mark(b>>1, color)
			}
			if kind == 4 {
				mark(c>>1, color)
			}
		}
		for i, lit := range p.OutputLiterals()[:gia.Gia_ManPoNum(p)] {
			mark(lit>>1, 1<<uint(i%2))
		}
		u.skip = make(map[int32]bool)
		for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
			if repr := gia.Gia_ObjRepr(p, id); repr != gia.GIA_VOID && colors[id] == colors[repr] {
				u.skip[id] = true
			}
		}
	}

	pairs := vec.Vec_IntAlloc(1000)
	for frame := prefix; frame < frames+prefix; frame++ {
		for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
			repr := gia.Gia_ObjRepr(p, id)
			if repr == gia.GIA_VOID || u.skip[id] {
				continue
			}
			x := int32(0)
			if repr != 0 {
				x = u.real(repr, frame)
			}
			y := u.real(id, frame) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p, repr)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p, id))
			if x == y || state != nil && gia.Gia_ObjProved(p, id) != 0 {
				continue
			}
			vec.Vec_IntPush(pairs, repr)
			vec.Vec_IntPush(pairs, id)
			gia.Gia_ManAppendCo(u.builder.graph, u.builder.xor(x, y))
		}
	}
	if state != nil && gia.Gia_ManCoNum(u.builder.graph) == 0 {
		gia.Gia_ManAppendCo(u.builder.graph, 0)
	}
	if outputs != nil {
		*outputs = pairs
	}
	result, _ := cecCorrDuplicate(u.builder.graph, false, false, nil)
	return result
}

func Cec_ManStartSimInfo(info *Vec_Ptr_t, flops int32) {
	if flops < 0 || int(flops) > vec.Vec_PtrSize(info) {
		panic("cec: invalid initial register count")
	}
	for k := int32(0); int(k) < vec.Vec_PtrSize(info); k++ {
		words := cecInfo(info, k)
		for w := range words {
			if k < flops {
				words[w] = 0
			} else {
				words[w] = gia.Gia_ManRandom(0)
			}
		}
	}
}
func Gia_ManCorrCreateRemapping(p *Gia_Man_t) *Vec_Int_t {
	pairs := vec.Vec_IntAlloc(100)
	for r := int32(0); r < gia.Gia_ManRegNum(p); r++ {
		id := gia.Gia_ObjId(p, gia.Gia_ManRo(p, r))
		repr := gia.Gia_ObjRepr(p, id)
		if repr == gia.GIA_VOID || repr == 0 || gia.Gia_ObjFailed(p, id) != 0 {
			continue
		}
		obj := gia.Gia_ManObj(p, repr)
		if gia.Gia_ObjIsCi(obj) == 0 || gia.Gia_ObjCioId(obj) < gia.Gia_ManPiNum(p) {
			panic("cec: register represented by non-register")
		}
		vec.Vec_IntPush(pairs, gia.Gia_ObjCioId(obj)-gia.Gia_ManPiNum(p))
		vec.Vec_IntPush(pairs, r)
	}
	return pairs
}
func Gia_ManCorrPerformRemapping(pairs *Vec_Int_t, info *Vec_Ptr_t) {
	data := vec.Vec_IntArray(pairs)
	if len(data)%2 != 0 {
		panic("cec: odd remapping pairs")
	}
	for i := 0; i < len(data); i += 2 {
		destination := cecInfo(info, data[i+1])
		for _, word := range destination {
			if word != 0 {
				panic("cec: remapping destination not zero")
			}
		}
		copy(destination, cecInfo(info, data[i]))
	}
}
func Gia_ManCorrRemapSimInfo(p *Gia_Man_t, info *Vec_Ptr_t) {
	Gia_ManCorrPerformRemapping(Gia_ManCorrCreateRemapping(p), info)
}
func Cec_ManLoadCounterExamplesTry(info, presence *Vec_Ptr_t, bit int32, lits []int32, count int32) int32 {
	for _, lit := range lits[:count] {
		if cecBit(cecInfo(presence, lit>>1), bit) != 0 && cecBit(cecInfo(info, lit>>1), bit) == uint32(lit&1) {
			return 0
		}
	}
	for _, lit := range lits[:count] {
		cecWriteBit(cecInfo(presence, lit>>1), bit, 1)
		cecWriteBit(cecInfo(info, lit>>1), bit, uint32(lit&1)^1)
	}
	return 1
}
func Cec_ManLoadCounterExamples(info *Vec_Ptr_t, store *Vec_Int_t, start int32) int32 {
	presence := vec.Vec_PtrAllocSimInfo(vec.Vec_PtrSize(info), vec.Vec_PtrReadWordsSimInfo(info))
	bits := int32(32 * vec.Vec_PtrReadWordsSimInfo(info))
	data := vec.Vec_IntArray(store)
	for int(start) < len(data) {
		if int(start)+2 > len(data) {
			panic("cec: truncated counterexample header")
		}
		count := data[start+1]
		start += 2
		if count <= 0 {
			continue
		}
		if count > int32(len(data))-start {
			panic("cec: truncated counterexample cube")
		}
		pattern := data[start : start+count]
		chosen := int32(1)
		for ; chosen < bits; chosen++ {
			if Cec_ManLoadCounterExamplesTry(info, presence, chosen, pattern, count) != 0 {
				break
			}
		}
		if chosen == bits {
			return start - 2
		} // retry this unconsumed cube in the next batch
		start += count
		if chosen == bits-1 {
			break
		}
	}
	return start
}
func Cec_ManLoadCounterExamples2(info *Vec_Ptr_t, store *Vec_Int_t, start int32) int32 {
	data := vec.Vec_IntArray(store)
	bits := int32(32 * vec.Vec_PtrReadWordsSimInfo(info))
	bit := int32(1)
	for int(start) < len(data) {
		if int(start)+2 > len(data) {
			panic("cec: truncated counterexample header")
		}
		count := data[start+1]
		start += 2
		if count <= 0 {
			continue
		}
		if count > int32(len(data))-start {
			panic("cec: truncated counterexample cube")
		}
		for _, lit := range data[start : start+count] {
			cecWriteBit(cecInfo(info, lit>>1), bit, uint32(lit&1)^1)
		}
		start += count
		bit++
		if bit == bits {
			break
		}
	}
	return start
}
func Cec_ManResimulateCounterExamples(sim *Cec_ManSim_t, store *Vec_Int_t, frames int32) int32 {
	pairs := Gia_ManCorrCreateRemapping(sim.pAig)
	cecSimRefs(sim.pAig)
	sim.pPars.NFrames = frames
	info := vec.Vec_PtrAllocSimInfo(int(gia.Gia_ManRegNum(sim.pAig)+gia.Gia_ManPiNum(sim.pAig)*frames), int(sim.pPars.NWords))
	result := int32(0)
	for start := int32(0); int(start) < vec.Vec_IntSize(store); {
		if !sim.pPars.deadline.IsZero() && !time.Now().Before(sim.pPars.deadline) {
			return result
		}
		Cec_ManStartSimInfo(info, gia.Gia_ManRegNum(sim.pAig))
		next := Cec_ManLoadCounterExamples(info, store, start)
		if next == start {
			panic("cec: counterexample does not fit batch")
		}
		start = next
		Gia_ManCorrPerformRemapping(pairs, info)
		result |= Cec_ManSeqResimulate(sim, info)
	}
	return result
}
func Cec_ManResimulateCounterExamplesComb(sim *Cec_ManSim_t, store *Vec_Int_t) int32 {
	cecSimRefs(sim.pAig)
	sim.pPars.NFrames = 1
	info := vec.Vec_PtrAllocSimInfo(int(gia.Gia_ManCiNum(sim.pAig)), int(sim.pPars.NWords))
	result := int32(0)
	for start := int32(0); int(start) < vec.Vec_IntSize(store); {
		Cec_ManStartSimInfo(info, 0)
		next := Cec_ManLoadCounterExamples(info, store, start)
		if next == start {
			panic("cec: counterexample does not fit batch")
		}
		start = next
		result |= Cec_ManSeqResimulate(sim, info)
	}
	return result
}
func Gia_ManCheckRefinements(p *Gia_Man_t, status *Vec_Str_t, outputs *Vec_Int_t, sim *Cec_ManSim_t, rings int32) int32 {
	if 2*vec.Vec_StrSize(status) != vec.Vec_IntSize(outputs) {
		panic("cec: refinement output map mismatch")
	}
	for i, value := range vec.Vec_StrArray(status) {
		if int8(value) == -1 {
			Cec_ManSimClassRemoveOne(sim, vec.Vec_IntEntry(outputs, 2*i+1))
		}
	}
	return 1
}

// Duplicate only output cones; optional sequential cleanup follows live latch
// next-state cones to a fixed point before retaining the latch interface.
func cecCorrDuplicate(p *Gia_Man_t, useClasses, cleanup bool, flip []uint32) (*Gia_Man_t, []int32) {
	pis, pos, regs := gia.Gia_ManPiNum(p), gia.Gia_ManPoNum(p), gia.Gia_ManRegNum(p)
	live := make([]bool, regs)
	if !cleanup {
		for i := range live {
			live[i] = true
		}
	}
	seen := make(map[int32]bool)
	var mark func(int32)
	mark = func(id int32) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsCi(o) != 0 {
			r := gia.Gia_ObjCioId(o) - pis
			if r >= 0 && !live[r] {
				live[r] = true
				mark(p.OutputLiterals()[pos+r] >> 1)
			}
			return
		}
		a, b, c, kind := p.Gate(id)
		mark(a >> 1)
		if kind == 2 || kind == 3 || kind == 4 {
			mark(b >> 1)
		}
		if kind == 4 {
			mark(c >> 1)
		}
	}
	if cleanup {
		for _, lit := range p.OutputLiterals()[:pos] {
			mark(lit >> 1)
		}
	}
	builder := newCorrBuilder(p)
	mapping := make([]int32, gia.Gia_ManObjNum(p))
	for i := range mapping {
		mapping[i] = -1
	}
	mapping[0] = 0
	inputs, outputs := []int32{}, []int32{}
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		if i >= pis && !live[i-pis] {
			continue
		}
		lit := gia.Gia_ManAppendCi(builder.graph)
		if flip != nil && i >= pis {
			lit ^= int32(cecBit(flip, i-pis))
		}
		mapping[gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))] = lit
		inputs = append(inputs, i)
	}
	if useClasses {
		gia.Gia_ManSetPhase(p)
	}
	active := make(map[int32]bool)
	var visit func(int32) int32
	visit = func(id int32) int32 {
		if useClasses && id != 0 && p.HasRepresentatives() {
			repr := gia.Gia_ObjRepr(p, id)
			if repr != gia.GIA_VOID {
				if active[id] {
					panic("cec: cyclic correspondence class")
				}
				active[id] = true
				lit := visit(repr) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p, repr)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p, id))
				delete(active, id)
				mapping[id] = lit
				return lit
			}
		}
		if mapping[id] >= 0 {
			return mapping[id]
		}
		a, b, c, kind := p.Gate(id)
		av := visit(a>>1) ^ a&1
		bv, cv := int32(0), int32(0)
		if kind == 2 || kind == 3 || kind == 4 {
			bv = visit(b>>1) ^ b&1
		}
		if kind == 4 {
			cv = visit(c>>1) ^ c&1
		}
		mapping[id] = builder.gate(kind, av, bv, cv)
		return mapping[id]
	}
	for i, lit := range p.OutputLiterals() {
		if int32(i) >= pos && !live[int32(i)-pos] {
			continue
		}
		out := visit(lit>>1) ^ lit&1
		if flip != nil && int32(i) >= pos {
			out ^= int32(cecBit(flip, int32(i)-pos))
		}
		gia.Gia_ManAppendCo(builder.graph, out)
		outputs = append(outputs, int32(i))
	}
	count := int32(0)
	for _, yes := range live {
		if yes {
			count++
		}
	}
	gia.Gia_ManSetRegNum(builder.graph, count)
	if err := builder.graph.CopyIONamesFrom(p, inputs, outputs); err != nil {
		panic(err)
	}
	return builder.graph, mapping
}
func Gia_ManCorrReduce(p *Gia_Man_t) *Gia_Man_t {
	result, mapping := cecCorrDuplicate(p, true, false, nil)
	for id, lit := range mapping {
		gia.Gia_ObjSetValue(gia.Gia_ManObj(p, int32(id)), lit)
	}
	return result
}
func Cec_ManRefinedClassPrintStats(p *Gia_Man_t, status *Vec_Str_t, iteration int32, elapsed abctime) {
	constants, classes, literals := 0, 0, 0
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		if gia.Gia_ObjIsConst(p, id) != 0 {
			constants++
		}
		if gia.Gia_ObjIsHead(p, id) != 0 {
			classes++
		}
		if gia.Gia_ObjIsCand(gia.Gia_ManObj(p, id)) != 0 && gia.Gia_ObjIsNone(p, id) == 0 {
			literals++
		}
	}
	proved, disproved, failed := 0, 0, 0
	if status != nil {
		for _, value := range vec.Vec_StrArray(status) {
			switch int8(value) {
			case 1:
				proved++
			case 0:
				disproved++
			case -1:
				failed++
			}
		}
	}
	fmt.Printf("%3d : c=%8d cl=%7d lit=%8d p=%6d d=%6d f=%6d T=%d\n", iteration, constants, classes, literals-classes, proved, disproved, failed, elapsed)
}
func cecCorrDeadlineExpired(pars *Cec_ParCor_t) bool {
	return !pars.deadline.IsZero() && !time.Now().Before(pars.deadline)
}
func cecCorrSim(p *Gia_Man_t, pars *Cec_ParCor_t) *Cec_ManSim_t {
	var sim Cec_ParSim_t
	Cec_ManSimSetDefaultParams(&sim)
	sim.NWords = pars.nWords
	sim.NFrames = pars.NFrames
	sim.FVerbose = pars.FVerbose
	sim.fLatchCorr = pars.FLatchCorr
	sim.fConstCorr = pars.FConstCorr
	sim.FSeqSimulate = 1
	sim.deadline = pars.deadline
	return Cec_ManSimStart(p, &sim)
}
func cecCorrCallback(pars *Cec_ParCor_t) {
	if pars.pFunc == nil {
		return
	}
	switch callback := pars.pFunc.(type) {
	case func(any) int:
		callback(pars.pData)
	case func(any) int32:
		callback(pars.pData)
	case func(any):
		callback(pars.pData)
	default:
		panic("cec: unsupported correspondence callback type")
	}
}
func Cec_ManLSCorrespondenceBmc(p *Gia_Man_t, pars *Cec_ParCor_t, prefix int32) {
	// This legacy entry point has no status result. On interruption, discard
	// candidates so a direct caller cannot mistake them for proved classes.
	defer func() {
		if cecCorrDeadlineExpired(pars) {
			p.ClearRepresentatives()
		}
	}()
	if cecCorrDeadlineExpired(pars) {
		return
	}
	sim := cecCorrSim(p, pars)
	defer Cec_ManSimStop(sim)
	var sat Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&sat)
	sat.NBTLimit = pars.NBTLimit
	sat.FVerbose = pars.FVerbose
	sat.deadline = pars.deadline
	for {
		if cecCorrDeadlineExpired(pars) {
			return
		}
		var outputs *Vec_Int_t
		model := Gia_ManCorrSpecReduceInit(p, pars.NFrames, prefix, 1-pars.FLatchCorr, &outputs, pars.FUseRings)
		if gia.Gia_ManCoNum(model) == 0 {
			break
		}
		if sat.NBTLimit > 0 {
			sat.NBTLimit = int32(min(int64(sat.NBTLimit)*10, int64(1<<30)))
		}
		var status *Vec_Str_t
		store := Cec_ManSatSolveMiter(model, &sat, &status)
		if cecCorrDeadlineExpired(pars) {
			return
		}
		if vec.Vec_IntSize(store) == 0 {
			break
		}
		before := cecCorrClassSignature(p)
		Cec_ManResimulateCounterExamples(sim, store, pars.NFrames+1+prefix)
		if cecCorrDeadlineExpired(pars) {
			return
		}
		Gia_ManCheckRefinements(p, status, outputs, sim, pars.FUseRings)
		if slices.Equal(before, cecCorrClassSignature(p)) {
			cecCorrDiscardDisproved(p, status, outputs, sim)
		}
		if slices.Equal(before, cecCorrClassSignature(p)) {
			break
		}
	}
}
func cecCorrClassSignature(p *Gia_Man_t) []int32 {
	data := make([]int32, gia.Gia_ManObjNum(p))
	for id := range data {
		data[id] = gia.Gia_ObjRepr(p, int32(id))
	}
	return data
}

// A validated SAT witness must refine its candidate. Sparse bit packing may
// combine independent cubes; fall back to deleting a still-disproved member.
func cecCorrDiscardDisproved(p *Gia_Man_t, status *Vec_Str_t, outputs *Vec_Int_t, sim *Cec_ManSim_t) {
	for i, value := range vec.Vec_StrArray(status) {
		if int8(value) == 0 {
			id := vec.Vec_IntEntry(outputs, 2*i+1)
			if gia.Gia_ObjHasRepr(p, id) != 0 || gia.Gia_ObjIsHead(p, id) != 0 {
				Cec_ManSimClassRemoveOne(sim, id)
			}
		}
	}
}
func Cec_ManLSCorrespondenceClasses(p *Gia_Man_t, pars *Cec_ParCor_t) (result int32) {
	if cecCorrDeadlineExpired(pars) {
		return -1
	}
	defer func() {
		if result < 0 {
			p.ClearRepresentatives()
		}
	}()
	if gia.Gia_ManRegNum(p) == 0 {
		return 0
	}
	if pars.NFrames <= 0 || pars.nWords <= 0 {
		panic("cec: invalid correspondence parameters")
	}
	gia.Gia_ManRandom(1)
	sim := cecCorrSim(p, pars)
	defer Cec_ManSimStop(sim)
	if !p.HasRepresentatives() {
		Cec_ManSimClassesPrepare(sim, pars.nLevelMax)
		if cecCorrDeadlineExpired(pars) {
			return -1
		}
		Cec_ManSimClassesRefine(sim)
		if cecCorrDeadlineExpired(pars) {
			return -1
		}
	}
	var sat Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&sat)
	sat.NBTLimit = pars.NBTLimit
	sat.FVerbose = pars.FVerbose
	sat.deadline = pars.deadline
	if pars.FUseCSat != 0 {
		sat.NBTLimit = min(sat.NBTLimit, 1000)
	}
	if pars.FLatchCorr == 0 || pars.NFrames > 1 {
		Cec_ManLSCorrespondenceBmc(p, pars, 0)
		if cecCorrDeadlineExpired(pars) {
			return -1
		}
	}
	cecCorrCallback(pars)
	if cecCorrDeadlineExpired(pars) {
		return -1
	}
	cecCorrCallback(pars)
	for round := int32(0); round < 100000; round++ {
		if cecCorrDeadlineExpired(pars) {
			return -1
		}
		if pars.nStepsMax == round {
			return 1
		}
		start := time.Now()
		var outputs *Vec_Int_t
		model := Gia_ManCorrSpecReduce(p, pars.NFrames, 1-pars.FLatchCorr, &outputs, pars.FUseRings)
		if gia.Gia_ManCoNum(model) == 0 {
			return 1
		}
		var status *Vec_Str_t
		store := Cec_ManSatSolveMiter(model, &sat, &status)
		if cecCorrDeadlineExpired(pars) {
			return -1
		}
		if vec.Vec_IntSize(store) == 0 {
			return 1
		}
		before := cecCorrClassSignature(p)
		Cec_ManResimulateCounterExamples(sim, store, pars.NFrames+2)
		if cecCorrDeadlineExpired(pars) {
			return -1
		}
		Gia_ManCheckRefinements(p, status, outputs, sim, pars.FUseRings)
		if slices.Equal(before, cecCorrClassSignature(p)) {
			cecCorrDiscardDisproved(p, status, outputs, sim)
		}
		if pars.FVerbose != 0 {
			Cec_ManRefinedClassPrintStats(p, status, round+1, time.Since(start).Nanoseconds())
		}
		cecCorrCallback(pars)
		if pars.FStopWhenGone != 0 && gia.Gia_ManPoNum(p) == 1 && gia.Gia_ObjIsConst(p, gia.Gia_ObjFaninId0p(p, gia.Gia_ManPo(p, 0))) == 0 {
			return 0
		}
		if slices.Equal(before, cecCorrClassSignature(p)) {
			return 1
		}
	}
	return 0
}
func Cec_ManLSCorrAnalyzeDependence(p *Gia_Man_t, equivs *Vec_Int_t, status *Vec_Str_t) int32 {
	if 2*vec.Vec_StrSize(status) != vec.Vec_IntSize(equivs) {
		panic("cec: dependence pair/status size mismatch")
	}
	marked := make([]bool, gia.Gia_ManObjNum(p))
	initial := 0
	for i, value := range vec.Vec_StrArray(status) {
		if int8(value) != 1 {
			marked[vec.Vec_IntEntry(equivs, 2*i+1)] = true
			initial++
		}
	}
	iteration := 0
	for ; iteration < 100; iteration++ {
		for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
			a, b, c, kind := p.Gate(id)
			if kind < 2 {
				continue
			}
			marked[id] = marked[id] || marked[a>>1]
			if kind == 2 || kind == 3 || kind == 4 {
				marked[id] = marked[id] || marked[b>>1]
			}
			if kind == 4 {
				marked[id] = marked[id] || marked[c>>1]
			}
		}
		changes := false
		for r := int32(0); r < gia.Gia_ManRegNum(p); r++ {
			ro := gia.Gia_ObjId(p, gia.Gia_ManRo(p, r))
			next := marked[gia.Gia_ObjId(p, gia.Gia_ManRi(p, r))]
			changes = changes || marked[ro] != next
			marked[ro] = next
		}
		if !changes {
			break
		}
	}
	final := 0
	for id, yes := range marked {
		if yes && gia.Gia_ObjHasRepr(p, int32(id)) != 0 {
			final++
		}
	}
	fmt.Printf("%5d -> %5d (%3d)  ", initial, final, iteration)
	return 0
}
func Cec_ManComputeInitState(p *Gia_Man_t, frames int32) []uint32 {
	gia.Gia_ManRandom(1)
	state := make([]uint32, (gia.Gia_ManRegNum(p)+31)/32)
	values := make([]uint32, gia.Gia_ManObjNum(p))
	for frame := int32(0); frame < frames; frame++ {
		for pi := int32(0); pi < gia.Gia_ManPiNum(p); pi++ {
			values[gia.Gia_ObjId(p, gia.Gia_ManPi(p, pi))] = gia.Gia_ManRandom(0) & 1
		}
		for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
			a, b, c, kind := p.Gate(id)
			x := values[a>>1] ^ uint32(a&1)
			y := values[b>>1] ^ uint32(b&1)
			switch kind {
			case 2:
				values[id] = x & y
			case 3:
				values[id] = x ^ y
			case 4:
				if x != 0 {
					values[id] = y
				} else {
					values[id] = values[c>>1] ^ uint32(c&1)
				}
			case 5, 6:
				values[id] = x
			}
		}
		for r := int32(0); r < gia.Gia_ManRegNum(p); r++ {
			values[gia.Gia_ObjId(p, gia.Gia_ManRo(p, r))] = values[gia.Gia_ObjId(p, gia.Gia_ManRi(p, r))]
		}
	}
	for r := int32(0); r < gia.Gia_ManRegNum(p); r++ {
		cecWriteBit(state, r, values[gia.Gia_ObjId(p, gia.Gia_ManRo(p, r))])
	}
	return state
}
func Cec_ManPrintFlopEquivs(p *Gia_Man_t) {
	names := p.InputNames()
	if names == nil {
		fmt.Println("Flop output names are not available. Use command &get -n.")
		return
	}
	for r := int32(0); r < gia.Gia_ManRegNum(p); r++ {
		obj := gia.Gia_ManRo(p, r)
		id := gia.Gia_ObjId(p, obj)
		repr := gia.Gia_ObjRepr(p, id)
		if repr == gia.GIA_VOID {
			continue
		}
		name := names[gia.Gia_ObjCioId(obj)]
		if repr == 0 {
			fmt.Printf("Original flop %s is proved equivalent to constant.\n", name)
		} else if other := gia.Gia_ManObj(p, repr); gia.Gia_ObjIsCi(other) != 0 {
			fmt.Printf("Original flop %s is proved equivalent to flop %s.\n", name, names[gia.Gia_ObjCioId(other)])
		} else {
			fmt.Printf("Original flop %s is proved equivalent to internal node %d.\n", name, repr)
		}
	}
}

func Cec_ManLSCorrespondence(p *Gia_Man_t, pars *Cec_ParCor_t) *Gia_Man_t {
	if cecCorrDeadlineExpired(pars) {
		return nil
	}
	p.ClearRepresentatives()
	if pars.NPrefix == 0 {
		status := Cec_ManLSCorrespondenceClasses(p, pars)
		if status < 0 || cecCorrDeadlineExpired(pars) {
			p.ClearRepresentatives()
			return nil
		}
		if status == 0 {
			copy, _ := cecCorrDuplicate(p, false, false, nil)
			return copy
		}
	} else {
		shifted, mapping := cecCorrDuplicate(p, false, false, Cec_ManComputeInitState(p, pars.NPrefix))
		if Cec_ManLSCorrespondenceClasses(shifted, pars) < 0 || cecCorrDeadlineExpired(pars) {
			p.ClearRepresentatives()
			return nil
		}
		inverse := make([]int32, gia.Gia_ManObjNum(shifted))
		for i := range inverse {
			inverse[i] = -1
		}
		for id, lit := range mapping {
			if lit >= 0 && inverse[lit>>1] < 0 {
				inverse[lit>>1] = int32(id)
			}
		}
		inverse[0] = 0
		if err := p.TransferRepresentativesFrom(shifted, inverse); err != nil {
			panic(err)
		}
		pars.FUseCSat = 0
		pars.NBTLimit = max(pars.NBTLimit, 1000)
		Cec_ManLSCorrespondenceBmc(p, pars, pars.NPrefix)
		if cecCorrDeadlineExpired(pars) {
			p.ClearRepresentatives()
			return nil
		}
	}
	if pars.FMakeChoices != 0 {
		return cecCorrChoices(p)
	}
	reduced := Gia_ManCorrReduce(p)
	result, _ := cecCorrDuplicate(reduced, false, true, nil)
	if pars.FVerboseFlops != 0 {
		Cec_ManPrintFlopEquivs(p)
	}
	return result
}

// Materialize proven internal alternatives without allowing a choice to depend
// on its own representative. Register equivalences are substituted, as in C.
func cecCorrChoices(p *Gia_Man_t) *Gia_Man_t {
	gia.Gia_ManSetPhase(p)
	builder := newCorrBuilder(p)
	copies := make([]int32, gia.Gia_ManObjNum(p))
	for i := range copies {
		copies[i] = -1
	}
	copies[0] = 0
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		copies[gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))] = gia.Gia_ManAppendCi(builder.graph)
	}
	for r := int32(0); r < gia.Gia_ManRegNum(p); r++ {
		id := gia.Gia_ObjId(p, gia.Gia_ManRo(p, r))
		if repr := gia.Gia_ObjRepr(p, id); repr != gia.GIA_VOID {
			copies[id] = copies[repr]
		}
	}
	classes := make(map[int32][]int32)
	representatives := make(map[int32]int32)
	var hasInCone func(int32, int32, map[int32]bool) bool
	hasInCone = func(target, id int32, seen map[int32]bool) bool {
		if id == target {
			return true
		}
		if id == 0 || seen[id] {
			return false
		}
		seen[id] = true
		a, b, c, kind := builder.graph.Gate(id)
		if kind < 2 {
			return false
		}
		if hasInCone(target, a>>1, seen) {
			return true
		}
		if kind == 2 || kind == 3 || kind == 4 {
			if hasInCone(target, b>>1, seen) {
				return true
			}
		}
		if kind == 4 {
			if hasInCone(target, c>>1, seen) {
				return true
			}
		}
		for _, other := range classes[id] {
			if hasInCone(target, other, seen) {
				return true
			}
		}
		return false
	}
	var visit func(int32) int32
	visit = func(id int32) int32 {
		if copies[id] >= 0 {
			return copies[id]
		}
		repr := gia.Gia_ObjRepr(p, id)
		canonical := int32(-1)
		if repr != gia.GIA_VOID {
			canonical = visit(repr) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p, repr)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p, id))
			if repr == 0 {
				copies[id] = canonical
				return canonical
			}
		}
		a, b, c, kind := p.Gate(id)
		av := visit(a>>1) ^ a&1
		bv, cv := int32(0), int32(0)
		if kind == 2 || kind == 3 || kind == 4 {
			bv = visit(b>>1) ^ b&1
		}
		if kind == 4 {
			cv = visit(c>>1) ^ c&1
		}
		lit := builder.gate(kind, av, bv, cv)
		copies[id] = lit
		if canonical >= 0 {
			head, alternative := canonical>>1, lit>>1
			if head != alternative && head < alternative && head != 0 {
				if _, exists := representatives[alternative]; !exists {
					if !hasInCone(head, alternative, make(map[int32]bool)) {
						classes[head] = append(classes[head], alternative)
						representatives[alternative] = head
					}
				}
			}
			copies[id] = canonical
		}
		return copies[id]
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		if gia.Gia_ObjIsCand(gia.Gia_ManObj(p, id)) != 0 {
			visit(id)
		}
	}
	for _, lit := range p.OutputLiterals() {
		gia.Gia_ManAppendCo(builder.graph, visit(lit>>1)^lit&1)
	}
	gia.Gia_ManSetRegNum(builder.graph, gia.Gia_ManRegNum(p))
	// Alternatives with ordinary fanouts cannot be choice-only nodes.
	fanouts := make(map[int32]bool)
	for id := int32(1); id < gia.Gia_ManObjNum(builder.graph); id++ {
		a, b, c, kind := builder.graph.Gate(id)
		if kind >= 2 {
			fanouts[a>>1] = true
		}
		if kind == 2 || kind == 3 || kind == 4 {
			fanouts[b>>1] = true
		}
		if kind == 4 {
			fanouts[c>>1] = true
		}
	}
	builder.graph.StartRepresentatives()
	for head, members := range classes {
		previous := head
		for _, id := range members {
			if fanouts[id] {
				continue
			}
			gia.Gia_ObjSetRepr(builder.graph, id, head)
			builder.graph.SetNext(previous, id)
			previous = id
		}
	}
	if err := builder.graph.CopyIONamesFrom(p, nil, nil); err != nil {
		panic(err)
	}
	return builder.graph
}
