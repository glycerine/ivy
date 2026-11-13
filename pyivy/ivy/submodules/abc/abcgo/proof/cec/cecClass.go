package cec

// Equivalence-class refinement and recyclable word-parallel simulation from
// cecClass.c. A simulation record holds its reference count followed by words.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

type Cec_ManSim_t struct {
	pAig                                *Gia_Man_t
	pPars                               *Cec_ParSim_t
	nWords                              int32
	pSimInfo                            []int32
	pMems                               []uint32
	nMems, nMemsMax, MemFree, nWordsOld int32
	vCiSimInfo, vCoSimInfo              *Vec_Ptr_t
	pCexes                              []bool
	iOut, nOuts                         int32
	pCexComb, pBestState                *Abc_Cex_t
	pScores                             []int32
	vClassOld, vClassNew, vRefinedC     *Vec_Int_t
}

func Cec_ManSimStart(aig *Gia_Man_t, pars *Cec_ParSim_t) *Cec_ManSim_t {
	if pars == nil || pars.nWords <= 0 {
		panic("invalid CEC simulation parameters")
	}
	p := &Cec_ManSim_t{pAig: aig, pPars: pars, nWords: pars.nWords, pSimInfo: make([]int32, gia.Gia_ManObjNum(aig)), iOut: -1, nMems: 1, vClassOld: vec.Vec_IntAlloc(1000), vClassNew: vec.Vec_IntAlloc(1000), vRefinedC: vec.Vec_IntAlloc(10000)}
	p.vCiSimInfo = vec.Vec_PtrAllocSimInfo(int(gia.Gia_ManCiNum(aig)), int(pars.nWords))
	if pars.fCheckMiter != 0 || gia.Gia_ManRegNum(aig) != 0 {
		p.vCoSimInfo = vec.Vec_PtrAllocSimInfo(int(gia.Gia_ManCoNum(aig)), int(pars.nWords))
	}
	return p
}
func Cec_ManSimStop(p *Cec_ManSim_t) {
	if p != nil {
		*p = Cec_ManSim_t{}
	}
}
func Cec_ObjSim(p *Cec_ManSim_t, id int32) []uint32 {
	off := p.pSimInfo[id]
	if off == 0 {
		panic("CEC object has no simulation record")
	}
	return p.pMems[off+1 : off+1+p.nWords]
}
func Cec_ObjSetSim(p *Cec_ManSim_t, id, n int32) { p.pSimInfo[id] = n }
func Cec_MemUsage(p *Cec_ManSim_t) float32 {
	return 4 * float32(p.nMemsMax) * float32(p.nWords+1) / (1 << 20)
}
func cecClass(p *Gia_Man_t, id int32) bool {
	r := gia.Gia_ObjRepr(p, id)
	return r > 0 && r != gia.GIA_VOID || p.Next(id) > 0
}
func cecHead(p *Gia_Man_t, id int32) bool {
	return gia.Gia_ObjRepr(p, id) == gia.GIA_VOID && p.Next(id) > 0
}
func cecTail(p *Gia_Man_t, id int32) bool {
	r := gia.Gia_ObjRepr(p, id)
	return r > 0 && r != gia.GIA_VOID && p.Next(id) <= 0
}
func cecClassMembers(p *Gia_Man_t, head int32) []int32 {
	a := []int32{}
	for id := head; id > 0; id = p.Next(id) {
		a = append(a, id)
	}
	return a
}
func Cec_ManSimClassCreate(p *Gia_Man_t, class *Vec_Int_t) {
	a := vec.Vec_IntArray(class)
	if len(a) == 0 {
		panic("cannot create empty CEC class")
	}
	head := a[0]
	gia.Gia_ObjSetRepr(p, head, gia.GIA_VOID)
	for i, id := range a {
		if i > 0 {
			if a[i-1] >= id {
				panic("CEC classes must be sorted")
			}
			gia.Gia_ObjSetRepr(p, id, head)
			p.SetNext(a[i-1], id)
		}
	}
	p.SetNext(a[len(a)-1], 0)
}
func Cec_ManSimClassRefineOne(p *Cec_ManSim_t, head int32) int32 {
	vec.Vec_IntClear(p.vClassOld)
	vec.Vec_IntClear(p.vClassNew)
	vec.Vec_IntPush(p.vClassOld, head)
	base := Cec_ObjSim(p, head)
	for id := p.pAig.Next(head); id > 0; id = p.pAig.Next(id) {
		other := Cec_ObjSim(p, id)
		if Cec_ManSimCompareEqual(base, other, p.nWords) != 0 {
			vec.Vec_IntPush(p.vClassOld, id)
		} else {
			vec.Vec_IntPush(p.vClassNew, id)
			if p.pBestState != nil {
				Cec_ManSimCompareEqualScore(base, other, p.nWords, p.pScores)
			}
		}
	}
	if vec.Vec_IntSize(p.vClassNew) == 0 {
		return 0
	}
	Cec_ManSimClassCreate(p.pAig, p.vClassOld)
	Cec_ManSimClassCreate(p.pAig, p.vClassNew)
	if vec.Vec_IntSize(p.vClassNew) > 1 {
		return 1 + Cec_ManSimClassRefineOne(p, vec.Vec_IntEntry(p.vClassNew, 0))
	}
	return 1
}
func Cec_ManSimClassRemoveOne(p *Cec_ManSim_t, id int32) int32 {
	aig := p.pAig
	if gia.Gia_ObjRepr(aig, id) == 0 {
		gia.Gia_ObjSetRepr(aig, id, gia.GIA_VOID)
		return 1
	}
	if !cecClass(aig, id) {
		return 0
	}
	head := gia.Gia_ObjRepr(aig, id)
	if head == gia.GIA_VOID {
		head = id
	}
	vec.Vec_IntClear(p.vClassOld)
	vec.Vec_IntClear(p.vClassNew)
	for _, ent := range cecClassMembers(aig, head) {
		if ent == id {
			vec.Vec_IntPush(p.vClassNew, ent)
		} else {
			vec.Vec_IntPush(p.vClassOld, ent)
		}
	}
	Cec_ManSimClassCreate(aig, p.vClassOld)
	Cec_ManSimClassCreate(aig, p.vClassNew)
	return 1
}
func Cec_ManSimMemRelink(p *Cec_ManSim_t) {
	if p.nMems > 1 && p.nWordsOld != p.nWords {
		panic("cannot relink live CEC simulation records")
	}
	stride := p.nWords + 1
	p.MemFree = 0
	prev := int32(0)
	for off := p.nMems * stride; off+stride < int32(len(p.pMems)); off += stride {
		if prev == 0 {
			p.MemFree = off
		} else {
			p.pMems[prev] = uint32(off)
		}
		prev = off
	}
	if prev != 0 {
		p.pMems[prev] = 0
	}
	p.nWordsOld = p.nWords
}
func Cec_ManSimSimRef(p *Cec_ManSim_t, id int32) []uint32 {
	if p.pSimInfo[id] != 0 {
		panic("CEC simulation record already referenced")
	}
	if p.MemFree == 0 {
		size := len(p.pMems)
		if size == 0 {
			size = 1 << 17
			p.nMems = 1
		}
		memory := make([]uint32, size*2)
		copy(memory, p.pMems)
		p.pMems = memory
		Cec_ManSimMemRelink(p)
	}
	off := p.MemFree
	p.MemFree = int32(p.pMems[off])
	p.pSimInfo[id] = off
	p.pMems[off] = uint32(gia.Gia_ObjValue(gia.Gia_ManObj(p.pAig, id)))
	p.nMems++
	p.nMemsMax = max(p.nMemsMax, p.nMems)
	return p.pMems[off : off+p.nWords+1]
}
func cecSimFree(p *Cec_ManSim_t, id int32) {
	off := p.pSimInfo[id]
	p.pMems[off] = uint32(p.MemFree)
	p.MemFree = off
	p.pSimInfo[id] = 0
	p.nMems--
}
func Cec_ManSimSimDeref(p *Cec_ManSim_t, id int32) []uint32 {
	off := p.pSimInfo[id]
	if off == 0 || p.pMems[off] == 0 {
		panic("CEC simulation reference underflow")
	}
	r := p.pMems[off : off+p.nWords+1]
	r[0]--
	if r[0] == 0 {
		cecSimFree(p, id)
	}
	return r
}
func Cec_ManSimProcessRefined(p *Cec_ManSim_t, refined *Vec_Int_t) {
	ids := vec.Vec_IntArray(refined)
	if len(ids) == 0 {
		return
	}
	size := util.Abc_PrimeCudd(uint32(100 + len(ids)/3))
	table := make([]int32, size)
	for _, id := range ids {
		key := Cec_ManSimHashKey(Cec_ObjSim(p, id), p.nWords, size)
		tail := table[key]
		if tail == 0 {
			gia.Gia_ObjSetRepr(p.pAig, id, gia.GIA_VOID)
		} else {
			p.pAig.SetNext(tail, id)
			r := gia.Gia_ObjRepr(p.pAig, tail)
			if r == gia.GIA_VOID {
				r = tail
			}
			gia.Gia_ObjSetRepr(p.pAig, id, r)
		}
		table[key] = id
	}
	for _, id := range ids {
		if cecHead(p.pAig, id) {
			Cec_ManSimClassRefineOne(p, id)
		}
	}
	for _, id := range ids {
		Cec_ManSimSimDeref(p, id)
	}
}
func cecInfoBit(a []uint32, bit int32) bool { return a[bit>>5]&(1<<uint(bit&31)) != 0 }
func Cec_ManSimSavePattern(p *Cec_ManSim_t, bit int32) {
	if p.pCexComb != nil || bit < 0 || bit >= 32*p.nWords {
		panic("invalid CEC counterexample pattern")
	}
	p.pCexComb = util.Abc_CexAlloc(0, gia.Gia_ManCiNum(p.pAig), 1)
	p.pCexComb.IPo = p.iOut
	for i := int32(0); i < gia.Gia_ManCiNum(p.pAig); i++ {
		row := vec.Vec_PtrEntry(p.vCiSimInfo, int(i)).([]uint32)
		if cecInfoBit(row, bit) {
			p.pCexComb.PData[i>>5] |= 1 << uint(i&31)
		}
	}
}
func Cec_ManSimFindBestPattern(p *Cec_ManSim_t) {
	score, bit := int32(0), int32(1)
	for i, s := range p.pScores {
		if s > score {
			score = s
			bit = int32(i)
		}
	}
	if p.pBestState.IPo > score {
		return
	}
	if p.pBestState.NRegs != gia.Gia_ManRegNum(p.pAig) {
		panic("CEC best-state register mismatch")
	}
	for i := int32(0); i < p.pBestState.NRegs; i++ {
		row := vec.Vec_PtrEntry(p.vCiSimInfo, int(gia.Gia_ManPiNum(p.pAig)+i)).([]uint32)
		mask := uint32(1) << uint(i&31)
		if cecInfoBit(row, bit) {
			p.pBestState.PData[i>>5] |= mask
		} else {
			p.pBestState.PData[i>>5] &^= mask
		}
	}
	p.pBestState.IPo = score
}
func Cec_ManSimAnalyzeOutputs(p *Cec_ManSim_t) int32 {
	if p.pPars.fCheckMiter == 0 {
		return 0
	}
	n := gia.Gia_ManPoNum(p.pAig)
	step := int32(1)
	if p.pPars.fDualOut != 0 {
		if n&1 != 0 {
			panic("dual-output miter must have pairs")
		}
		step = 2
	}
	for i := int32(0); i < n; i += step {
		a := vec.Vec_PtrEntry(p.vCoSimInfo, int(i)).([]uint32)
		var bit int32
		if step == 2 {
			b := vec.Vec_PtrEntry(p.vCoSimInfo, int(i+1)).([]uint32)
			bit = Cec_ManSimCompareEqualFirstBit(a, b, p.nWords)
		} else {
			bit = Cec_ManSimCompareConstFirstBit(a, p.nWords)
		}
		if bit < 0 {
			continue
		}
		out := i / step
		if p.iOut < 0 {
			p.iOut = out
			Cec_ManSimSavePattern(p, bit)
		}
		if p.pCexes == nil {
			p.pCexes = make([]bool, n/step)
		}
		if !p.pCexes[out] {
			p.pCexes[out] = true
			p.nOuts++
		}
	}
	if p.pCexes != nil {
		return 1
	}
	return 0
}
func cecSimRefs(p *Gia_Man_t) {
	n := gia.Gia_ManObjNum(p)
	for i := int32(0); i < n; i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManObj(p, i), 0)
	}
	for i := int32(1); i < n; i++ {
		o := gia.Gia_ManObj(p, i)
		if gia.Gia_ObjIsAnd(o) != 0 || gia.Gia_ObjIsCo(o) != 0 {
			a := gia.Gia_ObjFanin0(o)
			gia.Gia_ObjSetValue(a, gia.Gia_ObjValue(a)+1)
		}

		if gia.Gia_ObjIsMux(p, o) != 0 {
			a := gia.Gia_ObjFanin2(p, o)
			gia.Gia_ObjSetValue(a, gia.Gia_ObjValue(a)+1)
		}
		if gia.Gia_ObjIsAnd(o) != 0 && gia.Gia_ObjIsBuf(o) == 0 {
			a := gia.Gia_ObjFanin1(o)
			gia.Gia_ObjSetValue(a, gia.Gia_ObjValue(a)+1)
		}
	}
}
func cecSimClassPrint(p *Cec_ManSim_t) {
	constants, classes := 0, 0
	for i := int32(1); i < gia.Gia_ManObjNum(p.pAig); i++ {
		if gia.Gia_ObjRepr(p.pAig, i) == 0 {
			constants++
		}
		if cecHead(p.pAig, i) {
			classes++
		}
	}
	fmt.Printf("Constant candidates = %d. Classes = %d. Memory = %.2f MB\n", constants, classes, Cec_MemUsage(p))
}

func Cec_ManSimSimulateRound(p *Cec_ManSim_t, infoCis, infoCos *Vec_Ptr_t) int32 {
	if p.nWordsOld != p.nWords {
		Cec_ManSimMemRelink(p)
	}
	p.nMemsMax = 0
	p.pScores = nil
	if p.pBestState != nil {
		p.pScores = make([]int32, 32*p.nWords)
	}
	vec.Vec_IntClear(p.vRefinedC)
	if gia.Gia_ObjValue(gia.Gia_ManConst0(p.pAig)) != 0 {
		r := Cec_ManSimSimRef(p, 0)
		clear(r[1:])
	}
	ci, co := 0, 0
	for id := int32(1); id < gia.Gia_ManObjNum(p.pAig); id++ {
		if id&255 == 0 && !p.pPars.deadline.IsZero() && !time.Now().Before(p.pPars.deadline) {
			return 0
		}
		o := gia.Gia_ManObj(p.pAig, id)
		if gia.Gia_ObjIsCo(o) != 0 {
			r := Cec_ManSimSimDeref(p, gia.Gia_ObjFaninId0(o, id))
			if infoCos != nil {
				dst := vec.Vec_PtrEntry(infoCos, co).([]uint32)
				co++
				for w := int32(0); w < p.nWords; w++ {
					dst[w] = r[w+1]
					if gia.Gia_ObjFaninC0(o) != 0 {
						dst[w] = ^dst[w]
					}
				}
			}
			continue
		}
		if gia.Gia_ObjIsCi(o) != 0 && gia.Gia_ObjValue(o) == 0 {
			ci++
			continue
		}
		r := Cec_ManSimSimRef(p, id)
		if gia.Gia_ObjIsCi(o) != 0 {
			if infoCis != nil {
				src := vec.Vec_PtrEntry(infoCis, ci).([]uint32)
				copy(r[1:], src[:p.nWords])
			} else {
				for w := int32(1); w <= p.nWords; w++ {
					r[w] = gia.Gia_ManRandom(0)
				}
			}
			ci++
			r[1] &^= 1
		} else {
			a := Cec_ManSimSimDeref(p, gia.Gia_ObjFaninId0(o, id))
			var control []uint32
			if gia.Gia_ObjIsMux(p.pAig, o) != 0 {
				control = Cec_ManSimSimDeref(p, gia.Gia_ObjId(p.pAig, gia.Gia_ObjFanin2(p.pAig, o)))
			}
			if gia.Gia_ObjIsBuf(o) != 0 {
				for w := int32(1); w <= p.nWords; w++ {
					r[w] = a[w]
					if gia.Gia_ObjFaninC0(o) != 0 {
						r[w] = ^r[w]
					}
				}
			} else {
				b := Cec_ManSimSimDeref(p, gia.Gia_ObjFaninId1(o, id))
				for w := int32(1); w <= p.nWords; w++ {
					x, y := a[w], b[w]
					if gia.Gia_ObjFaninC0(o) != 0 {
						x = ^x
					}
					if gia.Gia_ObjFaninC1(o) != 0 {
						y = ^y
					}
					if control != nil {
						c := control[w]
						if gia.Gia_ObjFaninC2(p.pAig, o) != 0 {
							c = ^c
						}
						r[w] = c&y | ^c&x
					} else if gia.Gia_ObjIsXor(o) != 0 {
						r[w] = x ^ y
					} else {
						r[w] = x & y
					}
				}
			}
		}
		if gia.Gia_ObjRepr(p.pAig, id) == 0 && Cec_ManSimCompareConst(r[1:], p.nWords) == 0 {
			r[0]++
			vec.Vec_IntPush(p.vRefinedC, id)
			if p.pBestState != nil {
				Cec_ManSimCompareConstScore(r[1:], p.nWords, p.pScores)
			}
		}
		if cecClass(p.pAig, id) {
			r[0]++
		}
		if cecTail(p.pAig, id) {
			head := gia.Gia_ObjRepr(p.pAig, id)
			members := cecClassMembers(p.pAig, head)
			Cec_ManSimClassRefineOne(p, head)
			for _, ent := range members {
				Cec_ManSimSimDeref(p, ent)
			}
		}
		// Dangling nodes can occur in an untrimmed Go input. A zero-use signature
		// still participates in candidate refinement, then returns to the pool.
		if p.pSimInfo[id] != 0 && p.pMems[p.pSimInfo[id]] == 0 {
			cecSimFree(p, id)
		}
	}
	if p.pPars.fConstCorr != 0 {
		for _, id := range vec.Vec_IntArray(p.vRefinedC) {
			gia.Gia_ObjSetRepr(p.pAig, id, gia.GIA_VOID)
			Cec_ManSimSimDeref(p, id)
		}
		vec.Vec_IntClear(p.vRefinedC)
	}
	Cec_ManSimProcessRefined(p, p.vRefinedC)
	if infoCis != nil && ci != int(gia.Gia_ManCiNum(p.pAig)) || infoCos != nil && co != int(gia.Gia_ManCoNum(p.pAig)) {
		panic("CEC simulation interface count mismatch")
	}
	if p.nMems != 1 {
		panic("CEC simulation record leak")
	}
	if p.pPars.fVeryVerbose != 0 {
		cecSimClassPrint(p)
	}
	if p.pBestState != nil {
		Cec_ManSimFindBestPattern(p)
	}
	return Cec_ManSimAnalyzeOutputs(p)
}
func Cec_ManSimCreateInfo(p *Cec_ManSim_t, infoCis, infoCos *Vec_Ptr_t) {
	cis := gia.Gia_ManCiNum(p.pAig)
	pis := gia.Gia_ManPiNum(p.pAig)
	seq := p.pPars.fSeqSimulate != 0 && gia.Gia_ManRegNum(p.pAig) > 0
	for i := int32(0); i < cis; i++ {
		dst := vec.Vec_PtrEntry(infoCis, int(i)).([]uint32)
		if seq && i >= pis {
			src := vec.Vec_PtrEntry(infoCos, int(gia.Gia_ManPoNum(p.pAig)+i-pis)).([]uint32)
			copy(dst[:p.nWords], src[:p.nWords])
		} else {
			for w := int32(0); w < p.nWords; w++ {
				dst[w] = gia.Gia_ManRandom(0)
			}
		}
	}
}
func cecSimLevels(p *Gia_Man_t) []int32 {
	gia.Gia_ManLevelNum(p)
	levels := make([]int32, gia.Gia_ManObjNum(p))
	for id := range levels {
		levels[id] = gia.Gia_ObjLevelId(p, int32(id))
	}
	return levels
}
func cecSimEnsureInfo(v *Vec_Ptr_t, nWords int32) {
	if v == nil {
		return
	}
	for i, row := range vec.Vec_PtrArray(v) {
		a := row.([]uint32)
		if len(a) < int(nWords) {
			b := make([]uint32, nWords)
			copy(b, a)
			vec.Vec_PtrWriteEntry(v, i, b)
		}
	}
}
func Cec_ManSimClassesPrepare(p *Cec_ManSim_t, levelMax int32) int32 {
	if p.pAig.HasRepresentatives() {
		panic("CEC classes already initialized")
	}
	p.pAig.StartRepresentatives()
	cecSimRefs(p.pAig)
	var levels []int32
	if levelMax >= 0 {
		levels = cecSimLevels(p.pAig)
	}
	for id := int32(0); id < gia.Gia_ManObjNum(p.pAig); id++ {
		o := gia.Gia_ManObj(p.pAig, id)
		r := int32(gia.GIA_VOID)
		if p.pPars.fLatchCorr == 0 && gia.Gia_ObjIsAnd(o) != 0 && (levelMax < 0 || levels[id] <= levelMax) {
			r = 0
		}
		gia.Gia_ObjSetRepr(p.pAig, id, r)
	}
	if p.pPars.fSeqSimulate != 0 {
		for i := int32(0); i < gia.Gia_ManRegNum(p.pAig); i++ {
			o := gia.Gia_ManRo(p.pAig, i)
			if gia.Gia_ObjValue(o) != 0 {
				gia.Gia_ObjSetRepr(p.pAig, gia.Gia_ObjId(p.pAig, o), 0)
			}
		}
	}
	if count := p.pAig.SimulationWordCount(); count > 0 {
		p.nWords = 2 * count
		cecSimEnsureInfo(p.vCiSimInfo, p.nWords)
		cecSimEnsureInfo(p.vCoSimInfo, p.nWords)
		src := p.pAig.SimulationInputs()
		if len(src) != int(gia.Gia_ManCiNum(p.pAig)*count) {
			panic("CEC initial simulation size mismatch")
		}
		for i := int32(0); i < gia.Gia_ManCiNum(p.pAig); i++ {
			dst := vec.Vec_PtrEntry(p.vCiSimInfo, int(i)).([]uint32)
			for j := int32(0); j < count; j++ {
				w := src[i*count+j]
				dst[2*j] = uint32(w)
				dst[2*j+1] = uint32(w >> 32)
			}
		}
		return Cec_ManSimSimulateRound(p, p.vCiSimInfo, p.vCoSimInfo)
	}
	for p.nWords = 1; p.nWords <= p.pPars.nWords; p.nWords = 2*p.nWords + 1 {
		if !p.pPars.deadline.IsZero() && !time.Now().Before(p.pPars.deadline) {
			return 0
		}
		if p.pPars.fVerbose != 0 {
			cecSimClassPrint(p)
		}
		for i := 0; i < 4; i++ {
			if !p.pPars.deadline.IsZero() && !time.Now().Before(p.pPars.deadline) {
				return 0
			}
			Cec_ManSimCreateInfo(p, p.vCiSimInfo, p.vCoSimInfo)
			if Cec_ManSimSimulateRound(p, p.vCiSimInfo, p.vCoSimInfo) != 0 {
				return 1
			}
		}
	}
	return 0
}
func Cec_ManSimClassesRefine(p *Cec_ManSim_t) int32 {
	cecSimRefs(p.pAig)
	p.nWords = p.pPars.nWords
	cecSimEnsureInfo(p.vCiSimInfo, p.nWords)
	cecSimEnsureInfo(p.vCoSimInfo, p.nWords)
	for i := int32(0); i < p.pPars.nRounds; i++ {
		if !p.pPars.deadline.IsZero() && !time.Now().Before(p.pPars.deadline) {
			return 0
		}
		if p.pPars.fVerbose != 0 && i%max(1, p.pPars.nRounds/5) == 0 {
			cecSimClassPrint(p)
		}
		Cec_ManSimCreateInfo(p, p.vCiSimInfo, p.vCoSimInfo)
		if Cec_ManSimSimulateRound(p, p.vCiSimInfo, p.vCoSimInfo) != 0 {
			return 1
		}
	}
	if p.pPars.fVerbose != 0 {
		cecSimClassPrint(p)
	}
	return 0
}
