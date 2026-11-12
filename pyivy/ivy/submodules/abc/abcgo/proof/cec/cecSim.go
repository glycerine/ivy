package cec

// Backward word-parallel implication simulation from cecSim.c. A bit denotes
// one independent attempt to justify opposite phases at a candidate pair.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

const SIM_RANDS = 113

type Cec_ManS_t struct {
	nWords, nLevelMax, nLevelMin, iRand int32
	pAig                                *Gia_Man_t
	vInputs                             *Vec_Int_t
	vLevels                             [][]int32
	levels                              []int32
	vSims                               *Vec_Wrd_t
	Rands                               [SIM_RANDS]uint64
	nSkipped, nVisited, nCexes          int32
	clkSat, clkUnsat                    int64
}

func Cec_ManSSim(p *Cec_ManS_t, node, value int32) []uint64 {
	off := p.nWords * (2*node + value)
	return vec.Vec_WrdArray(p.vSims)[off : off+p.nWords]
}
func Cec_ManSStart(aig *Gia_Man_t, nWords int32) *Cec_ManS_t {
	if nWords <= 0 || nWords > 1 && nWords&1 != 0 {
		panic("CEC implication simulation needs one or an even word count")
	}
	levels := cecSimLevels(aig)
	maxLevel := int32(0)
	for _, l := range levels {
		maxLevel = max(maxLevel, l)
	}
	p := &Cec_ManS_t{pAig: aig, nWords: nWords, levels: levels, vLevels: make([][]int32, maxLevel+1), vInputs: vec.Vec_IntAlloc(100), vSims: vec.Vec_WrdStart(int(gia.Gia_ManObjNum(aig) * nWords * 2))}
	for i := range p.Rands {
		p.Rands[i] = gia.Gia_ManRandomW(0)
	}
	return p
}
func Cec_ManSStop(p *Cec_ManS_t) {
	if p != nil {
		*p = Cec_ManS_t{}
	}
}
func Cec_ManSVerify_rec(p *Gia_Man_t, id int32) int32 {
	if id == 0 {
		return 0
	}
	o := gia.Gia_ManObj(p, id)
	if gia.Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return gia.Gia_ObjMark1(o)
	}
	gia.Gia_ObjSetTravIdCurrentId(p, id)
	if gia.Gia_ObjIsCi(o) != 0 {
		return gia.Gia_ObjMark1(o)
	}
	a := Cec_ManSVerify_rec(p, gia.Gia_ObjFaninId0(o, id)) ^ gia.Gia_ObjFaninC0(o)
	b := Cec_ManSVerify_rec(p, gia.Gia_ObjFaninId1(o, id)) ^ gia.Gia_ObjFaninC1(o)
	v := a & b
	if gia.Gia_ObjIsXor(o) != 0 {
		v = a ^ b
	}
	gia.Gia_ObjSetMark1(o, v)
	return v
}
func Cec_ManSVerifyTwo(p *Gia_Man_t, a, b int32) {
	gia.Gia_ManIncrementTravId(p)
	va := Cec_ManSVerify_rec(p, a)
	vb := Cec_ManSVerify_rec(p, b)
	phase := gia.Gia_ObjPhase(gia.Gia_ManObj(p, a)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p, b))
	if va^vb == phase {
		fmt.Printf("CEX verification FAILED for obj %d and obj %d.\n", a, b)
	}
}
func Cec_ManSVerify(p *Cec_ManS_t, a, b int32) {
	if vec.Vec_IntSize(p.vInputs) == 0 {
		fmt.Println("No primary inputs.")
		return
	}
	care := Cec_ManSSim(p, 0, 0)
	for _, id := range vec.Vec_IntArray(p.vInputs) {
		zero, one := Cec_ManSSim(p, id, 0), Cec_ManSSim(p, id, 1)
		for w := range care {
			care[w] |= zero[w] & one[w]
		}
	}
	n := 0
	for _, w := range care {
		n += bits.OnesCount64(w)
	}
	if n == int(64*p.nWords) {
		fmt.Println("No CEXes.")
		return
	}
	fmt.Printf("Considered %d CEXes of nodes %d and %d.\n", int(64*p.nWords)-n, a, b)
}
func Cec_ManSRunImply(p *Cec_ManS_t, id int32) {
	o := gia.Gia_ManObj(p.pAig, id)
	if gia.Gia_ObjIsAnd(o) == 0 {
		return
	}
	a, b := gia.Gia_ObjFaninId0(o, id), gia.Gia_ObjFaninId1(o, id)
	ca, cb := gia.Gia_ObjFaninC0(o), gia.Gia_ObjFaninC1(o)
	care := Cec_ManSSim(p, 0, 0)
	zero, one := Cec_ManSSim(p, id, 0), Cec_ManSSim(p, id, 1)
	az, ao := Cec_ManSSim(p, a, ca), Cec_ManSSim(p, a, ca^1)
	bz, bo := Cec_ManSSim(p, b, cb), Cec_ManSSim(p, b, cb^1)
	for w := range care {
		zero[w] |= az[w] | bz[w]
		one[w] |= ao[w] & bo[w]
		care[w] |= zero[w] & one[w]
		zero[w] &^= care[w]
		one[w] &^= care[w]
	}
}
func Cec_ManSRunPropagate(p *Cec_ManS_t, id int32) int32 {
	o := gia.Gia_ManObj(p.pAig, id)
	a, b := gia.Gia_ObjFaninId0(o, id), gia.Gia_ObjFaninId1(o, id)
	ca, cb := gia.Gia_ObjFaninC0(o), gia.Gia_ObjFaninC1(o)
	zero, one := Cec_ManSSim(p, id, 0), Cec_ManSSim(p, id, 1)
	nonzero := false
	for w := range zero {
		if zero[w]|one[w] != 0 {
			nonzero = true
			break
		}
	}
	if !nonzero {
		p.nSkipped++
		return 0
	}
	p.nVisited++
	care := Cec_ManSSim(p, 0, 0)
	az, ao := Cec_ManSSim(p, a, ca), Cec_ManSSim(p, a, ca^1)
	bz, bo := Cec_ManSSim(p, b, cb), Cec_ManSSim(p, b, cb^1)
	p.iRand = (p.iRand + 1) % SIM_RANDS
	for w := range care {
		r := p.Rands[(int(p.iRand)+w)%SIM_RANDS]
		ao[w] |= one[w]
		bo[w] |= one[w]
		az[w] |= zero[w] & (bo[w] | ^r)
		bz[w] |= zero[w] & (ao[w] | r)
		care[w] |= az[w]&ao[w] | bz[w]&bo[w]
		az[w] &^= care[w]
		ao[w] &^= care[w]
		bz[w] &^= care[w]
		bo[w] &^= care[w]
	}
	return 1
}
func Cec_ManSInsert(p *Cec_ManS_t, id int32) {
	if id == 0 {
		panic("normalize constant fanins before CEC implication simulation")
	}
	if gia.Gia_ObjIsTravIdCurrentId(p.pAig, id) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrentId(p.pAig, id)
	o := gia.Gia_ManObj(p.pAig, id)
	if gia.Gia_ObjIsCi(o) != 0 {
		vec.Vec_IntPush(p.vInputs, id)
		return
	}
	cecRequireAndGraph(p.pAig, o)
	level := p.levels[id]
	if level <= 0 {
		panic("CEC implication node must have positive level")
	}
	p.vLevels[level] = append(p.vLevels[level], id)
	p.nLevelMin = min(p.nLevelMin, level)
	p.nLevelMax = max(p.nLevelMax, level)
}
func Cec_ManSRunSimInt(p *Cec_ManS_t) int32 {
	solved := int32(0)
	for level := p.nLevelMax; level >= p.nLevelMin; level-- {
		for _, id := range p.vLevels[level] {
			if solved == 0 && Cec_ManSRunPropagate(p, id) != 0 {
				o := gia.Gia_ManObj(p.pAig, id)
				Cec_ManSInsert(p, gia.Gia_ObjFaninId0(o, id))
				Cec_ManSInsert(p, gia.Gia_ObjFaninId1(o, id))
				all := true
				for _, w := range Cec_ManSSim(p, 0, 0) {
					if w != ^uint64(0) {
						all = false
						break
					}
				}
				if all {
					solved = 1
				}
			}
			clear(Cec_ManSSim(p, id, 0))
			clear(Cec_ManSSim(p, id, 1))
		}
		p.vLevels[level] = p.vLevels[level][:0]
	}
	return solved
}
func Cec_ManSRunSim(p *Cec_ManS_t, a, b int32) int32 {
	start := util.Abc_Clock()
	if vec.Vec_IntSize(p.vInputs) != 0 {
		panic("previous CEC input signatures not released")
	}
	phase := gia.Gia_ObjPhase(gia.Gia_ManObj(p.pAig, a)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(p.pAig, b))
	az, ao := Cec_ManSSim(p, a, 0), Cec_ManSSim(p, a, 1)
	bz, bo := Cec_ManSSim(p, b, phase), Cec_ManSSim(p, b, phase^1)
	clear(Cec_ManSSim(p, 0, 0))
	if a == 0 {
		for w := range bo {
			bo[w] = ^uint64(0)
		}
	} else if p.nWords == 1 {
		az[0] = 0xffffffff
		bo[0] = 0xffffffff
		ao[0] = az[0] << 32
		bz[0] = bo[0] << 32
	} else {
		for w := int32(0); w < p.nWords/2; w++ {
			az[w] = ^uint64(0)
			bo[w] = ^uint64(0)
			ao[w+p.nWords/2] = ^uint64(0)
			bz[w+p.nWords/2] = ^uint64(0)
		}
	}
	p.nLevelMin = util.ABC_INFINITY
	p.nLevelMax = 0
	gia.Gia_ManIncrementTravId(p.pAig)
	if a != 0 {
		Cec_ManSInsert(p, a)
	}
	Cec_ManSInsert(p, b)
	p.nSkipped = 0
	p.nVisited = 0
	status := Cec_ManSRunSimInt(p)
	if status == 0 {
		p.clkSat += util.Abc_Clock() - start
		Cec_ManSVerify(p, a, b)
		p.nCexes++
	} else {
		p.clkUnsat += util.Abc_Clock() - start
	}
	for _, id := range vec.Vec_IntArray(p.vInputs) {
		clear(Cec_ManSSim(p, id, 0))
		clear(Cec_ManSSim(p, id, 1))
	}
	vec.Vec_IntClear(p.vInputs)
	return status
}
func Cec_ManSRunTest(aig *Gia_Man_t) {
	start := util.Abc_Clock()
	gia.Gia_ManRandomW(1)
	p := Cec_ManSStart(aig, 1)
	for id := int32(0); id < gia.Gia_ManObjNum(aig); id++ {
		if cecHead(aig, id) {
			for ent := aig.Next(id); ent > 0; ent = aig.Next(ent) {
				Cec_ManSRunSim(p, id, ent)
			}
		}
	}
	fmt.Printf("Detected %d CEXes. Time = %.2f sec. Sat = %.2f sec. Unsat = %.2f sec.\n", p.nCexes, float64(util.Abc_Clock()-start)/1e9, float64(p.clkSat)/1e9, float64(p.clkUnsat)/1e9)
	Cec_ManSStop(p)
}
