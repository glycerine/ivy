package cec

// Pattern compression, sensitizing-path minimization and compatible-pattern
// packing follow cecPat.c. Literals encode CI index<<1 | inverted assignment.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Cec_ManPat_t struct {
	vPattern1, vPattern2                                                           *Vec_Int_t
	vStorage                                                                       *Vec_Str_t
	iStart, nPats, nPatsAll, nPatLits, nPatLitsAll, nPatLitsMin, nPatLitsMinAll    int32
	nSeries, fVerbose                                                              int32
	timeFind, timeShrink, timeVerify, timeSort, timePack, timeTotal, timeTotalSave int64
}

func Cec_ManPatStart() *Cec_ManPat_t {
	return &Cec_ManPat_t{vPattern1: vec.Vec_IntAlloc(1000), vPattern2: vec.Vec_IntAlloc(1000), vStorage: vec.Vec_StrAlloc(1 << 20)}
}
func Cec_ManPatStop(p *Cec_ManPat_t) {
	if p != nil {
		p.vPattern1 = nil
		p.vPattern2 = nil
		p.vStorage = nil
	}
}
func Cec_ManPatPrintStats(p *Cec_ManPat_t) {
	ave := func(l, n int32) float64 {
		if n == 0 {
			return 0
		}
		return float64(l) / float64(n)
	}
	fmt.Printf("Latest: P = %8d. L = %10d. Lm = %10d. Ave = %6.1f. MEM =%6.2f MB\n", p.nPats, p.nPatLits, p.nPatLitsMin, ave(p.nPatLitsMin, p.nPats), float64(vec.Vec_StrSize(p.vStorage)-int(p.iStart))/(1<<20))
	fmt.Printf("Total: P = %8d. L = %10d. Lm = %10d. Ave = %6.1f. MEM =%6.2f MB\n", p.nPatsAll, p.nPatLitsAll, p.nPatLitsMinAll, ave(p.nPatLitsMinAll, p.nPatsAll), float64(vec.Vec_StrSize(p.vStorage))/(1<<20))
	for _, s := range []struct {
		name string
		ns   int64
	}{{"Finding", p.timeFind}, {"Shrinking", p.timeShrink}, {"Verifying", p.timeVerify}, {"Sorting", p.timeSort}, {"Packing", p.timePack}, {"TOTAL", p.timeTotal}} {
		fmt.Printf("%s = %.2f sec\n", s.name, float64(s.ns)/1e9)
	}
}
func Cec_ManPatStoreNum(p *Cec_ManPat_t, num int32) {
	if num < 0 {
		panic("CEC pattern number is negative")
	}
	x := uint32(num)
	for x > 127 {
		vec.Vec_StrPush(p.vStorage, byte(x&127)|128)
		x >>= 7
	}
	vec.Vec_StrPush(p.vStorage, byte(x))
}
func Cec_ManPatRestoreNum(p *Cec_ManPat_t) int32 {
	var x uint32
	for shift := uint(0); shift < 35; shift += 7 {
		if int(p.iStart) >= vec.Vec_StrSize(p.vStorage) {
			panic("truncated CEC pattern")
		}
		b := vec.Vec_StrEntry(p.vStorage, int(p.iStart))
		p.iStart++
		if shift == 28 && b > 7 {
			panic("CEC pattern number overflow")
		}
		x |= uint32(b&127) << shift
		if b&128 == 0 {
			return int32(x)
		}
	}
	panic("invalid CEC pattern number")
}
func Cec_ManPatStore(p *Cec_ManPat_t, pat *Vec_Int_t) {
	a := vec.Vec_IntArray(pat)
	if len(a) == 0 {
		panic("empty CEC pattern")
	}
	Cec_ManPatStoreNum(p, int32(len(a)))
	prev := int32(0)
	for i, x := range a {
		if i > 0 && x <= prev {
			panic("CEC pattern is not strictly increasing")
		}
		Cec_ManPatStoreNum(p, x-prev)
		prev = x
	}
}
func Cec_ManPatRestore(p *Cec_ManPat_t, pat *Vec_Int_t) {
	vec.Vec_IntClear(pat)
	n := Cec_ManPatRestoreNum(p)
	if n <= 0 || int(n) > vec.Vec_StrSize(p.vStorage)-int(p.iStart) {
		panic("invalid CEC pattern length")
	}
	value := int32(0)
	for i := int32(0); i < n; i++ {
		delta := Cec_ManPatRestoreNum(p)
		if i > 0 && delta == 0 || value > int32(0x7fffffff)-delta {
			panic("invalid CEC pattern delta")
		}
		value += delta
		vec.Vec_IntPush(pat, value)
	}
}

// The source justification algorithms consume normalized AND graphs. The
// forward class simulator separately supports specialized XOR/MUX objects.
func cecRequireAndGraph(p *Gia_Man_t, o *Gia_Obj_t) {
	if gia.Gia_ObjIsXor(o) != 0 || gia.Gia_ObjIsMux(p, o) != 0 {
		panic("normalize XOR/MUX gates before CEC backward justification")
	}
}
func Cec_ManPatComputePattern_rec(sat *Cec_ManSat_t, p *Gia_Man_t, o *Gia_Obj_t) int32 {
	if gia.Gia_ObjIsTravIdCurrent(p, o) != 0 {
		return 0
	}
	gia.Gia_ObjSetTravIdCurrent(p, o)
	if gia.Gia_ObjIsConst0(o) != 0 {
		gia.Gia_ObjSetMark1(o, 0)
		return 0
	}
	if gia.Gia_ObjIsCi(o) != 0 {
		gia.Gia_ObjSetMark1(o, Cec_ObjSatVarValue(sat, o))
		return 1
	}
	cecRequireAndGraph(p, o)
	a, b := gia.Gia_ObjFanin0(o), gia.Gia_ObjFanin1(o)
	n := Cec_ManPatComputePattern_rec(sat, p, a) + Cec_ManPatComputePattern_rec(sat, p, b)
	gia.Gia_ObjSetMark1(o, (gia.Gia_ObjMark1(a)^gia.Gia_ObjFaninC0(o))&(gia.Gia_ObjMark1(b)^gia.Gia_ObjFaninC1(o)))
	return n
}
func cecPatSensitize(p *Gia_Man_t, o *Gia_Obj_t, pat *Vec_Int_t, right bool) {
	if gia.Gia_ObjIsTravIdCurrent(p, o) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p, o)
	if gia.Gia_ObjIsConst0(o) != 0 {
		return
	}
	if gia.Gia_ObjIsCi(o) != 0 {
		v := int32(0)
		if gia.Gia_ObjMark1(o) == 0 {
			v = 1
		}
		vec.Vec_IntPush(pat, gia.Gia_ObjCioId(o)<<1|v)
		return
	}
	cecRequireAndGraph(p, o)
	a, b := gia.Gia_ObjFanin0(o), gia.Gia_ObjFanin1(o)
	if gia.Gia_ObjMark1(o) != 0 {
		cecPatSensitize(p, a, pat, right)
		cecPatSensitize(p, b, pat, right)
		return
	}
	if right {
		a, b = b, a
	}
	c := gia.Gia_ObjFaninC0(o)
	if right {
		c = gia.Gia_ObjFaninC1(o)
	}
	if gia.Gia_ObjMark1(a)^c == 0 {
		cecPatSensitize(p, a, pat, right)
	} else {
		cecPatSensitize(p, b, pat, right)
	}
}
func Cec_ManPatComputePattern1_rec(p *Gia_Man_t, o *Gia_Obj_t, pat *Vec_Int_t) {
	cecPatSensitize(p, o, pat, false)
}
func Cec_ManPatComputePattern2_rec(p *Gia_Man_t, o *Gia_Obj_t, pat *Vec_Int_t) {
	cecPatSensitize(p, o, pat, true)
}
func cecXNot(v, c int32) int32 {
	if c != 0 && v != 3 {
		return 3 ^ v
	}
	return v
}
func cecXAnd(a, b int32) int32 {
	if a == 1 || b == 1 {
		return 1
	}
	if a == 3 || b == 3 {
		return 3
	}
	return 2
}
func Cec_ManPatComputePattern3_rec(p *Gia_Man_t, o *Gia_Obj_t) int32 {
	if gia.Gia_ObjIsTravIdCurrent(p, o) != 0 {
		return gia.Gia_ObjMark1(o)<<1 | gia.Gia_ObjMark0(o)
	}
	gia.Gia_ObjSetTravIdCurrent(p, o)
	v := int32(3)
	if gia.Gia_ObjIsConst0(o) != 0 {
		v = 1
	} else if gia.Gia_ObjIsCi(o) == 0 {
		cecRequireAndGraph(p, o)
		a := Cec_ManPatComputePattern3_rec(p, gia.Gia_ObjFanin0(o))
		b := Cec_ManPatComputePattern3_rec(p, gia.Gia_ObjFanin1(o))
		v = cecXAnd(cecXNot(a, gia.Gia_ObjFaninC0(o)), cecXNot(b, gia.Gia_ObjFaninC1(o)))
	}
	gia.Gia_ObjSetMark0(o, v&1)
	gia.Gia_ObjSetMark1(o, (v>>1)&1)
	return v
}
func Cec_ManPatVerifyPattern(p *Gia_Man_t, o *Gia_Obj_t, pat *Vec_Int_t) {
	gia.Gia_ManIncrementTravId(p)
	for _, lit := range vec.Vec_IntArray(pat) {
		ci := gia.Gia_ManCi(p, lit>>1)
		v := int32(2)
		if lit&1 != 0 {
			v = 1
		}
		gia.Gia_ObjSetMark0(ci, v&1)
		gia.Gia_ObjSetMark1(ci, v>>1)
		gia.Gia_ObjSetTravIdCurrent(p, ci)
	}
	v := Cec_ManPatComputePattern3_rec(p, gia.Gia_ObjFanin0(o))
	if cecXNot(v, gia.Gia_ObjFaninC0(o)) != 2 {
		panic("Cec_ManPatVerifyPattern: verification failed")
	}
}
func Cec_ManPatComputePattern4_rec(p *Gia_Man_t, o *Gia_Obj_t) {
	if gia.Gia_ObjIsTravIdCurrent(p, o) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p, o)
	gia.Gia_ObjSetMark0(o, 0)
	if gia.Gia_ObjIsCi(o) != 0 || gia.Gia_ObjIsConst0(o) != 0 {
		return
	}
	Cec_ManPatComputePattern4_rec(p, gia.Gia_ObjFanin0(o))
	Cec_ManPatComputePattern4_rec(p, gia.Gia_ObjFanin1(o))
}
func Cec_ManPatCleanMark0(p *Gia_Man_t, o *Gia_Obj_t) {
	gia.Gia_ManIncrementTravId(p)
	Cec_ManPatComputePattern4_rec(p, gia.Gia_ObjFanin0(o))
}
func Cec_ManPatSavePattern(man *Cec_ManPat_t, sat *Cec_ManSat_t, o *Gia_Obj_t) {
	start := util.Abc_Clock()
	p := sat.pAig
	man.nPats++
	man.nPatsAll++
	gia.Gia_ManIncrementTravId(p)
	n := Cec_ManPatComputePattern_rec(sat, p, gia.Gia_ObjFanin0(o))
	man.nPatLits += n
	man.nPatLitsAll += n
	vec.Vec_IntClear(man.vPattern1)
	vec.Vec_IntClear(man.vPattern2)
	gia.Gia_ManIncrementTravId(p)
	Cec_ManPatComputePattern1_rec(p, gia.Gia_ObjFanin0(o), man.vPattern1)
	gia.Gia_ManIncrementTravId(p)
	Cec_ManPatComputePattern2_rec(p, gia.Gia_ObjFanin0(o), man.vPattern2)
	pat := man.vPattern2
	if vec.Vec_IntSize(man.vPattern1) < vec.Vec_IntSize(pat) {
		pat = man.vPattern1
	}
	man.nPatLitsMin += int32(vec.Vec_IntSize(pat))
	man.nPatLitsMinAll += int32(vec.Vec_IntSize(pat))
	vec.Vec_IntSort(pat, 0)
	Cec_ManPatStore(man, pat)
	man.timeTotal += util.Abc_Clock() - start
}
func Cec_ManPatSavePatternCSat(p *Cec_ManPat_t, pat *Vec_Int_t) {
	vec.Vec_IntSort(pat, 0)
	Cec_ManPatStore(p, pat)
}
func Cec_ManPatCollectTry(info, pres *Vec_Ptr_t, bit int32, lits []int32, nLits int32) int32 {
	for _, l := range lits[:nLits] {
		a := vec.Vec_PtrEntry(info, int(l>>1)).([]uint32)
		b := vec.Vec_PtrEntry(pres, int(l>>1)).([]uint32)
		mask := uint32(1) << uint(bit&31)
		if b[bit>>5]&mask != 0 && ((a[bit>>5]>>uint(bit&31))&1) == uint32(l&1) {
			return 0
		}
	}
	for _, l := range lits[:nLits] {
		a := vec.Vec_PtrEntry(info, int(l>>1)).([]uint32)
		b := vec.Vec_PtrEntry(pres, int(l>>1)).([]uint32)
		mask := uint32(1) << uint(bit&31)
		b[bit>>5] |= mask
		if ((a[bit>>5] >> uint(bit&31)) & 1) == uint32(l&1) {
			a[bit>>5] ^= mask
		}
	}
	return 1
}
func cecPatGrow(info, pres *Vec_Ptr_t, nInputs, nRegs, nWords int32) {
	for i := int32(0); i < nInputs; i++ {
		a := vec.Vec_PtrEntry(info, int(i)).([]uint32)
		b := vec.Vec_PtrEntry(pres, int(i)).([]uint32)
		aa := make([]uint32, 2*nWords)
		bb := make([]uint32, 2*nWords)
		copy(aa, a)
		copy(bb, b)
		if i >= nRegs {
			for w := nWords; w < 2*nWords; w++ {
				aa[w] = gia.Gia_ManRandom(0)
			}
		}
		vec.Vec_PtrWriteEntry(info, int(i), aa)
		vec.Vec_PtrWriteEntry(pres, int(i), bb)
	}
}
func cecPatInsert(info, pres *Vec_Ptr_t, pat []int32, nInputs, nRegs, chunk int32, nWords *int32) int32 {
	for {
		bits := 32 * (*nWords)
		for k := int32(1); k < bits; k++ {
			if k%(32*chunk) == 0 {
				continue
			}
			if Cec_ManPatCollectTry(info, pres, k, pat, int32(len(pat))) != 0 {
				if k == bits-1 {
					cecPatGrow(info, pres, nInputs, nRegs, *nWords)
					*nWords *= 2
				}
				return k
			}
		}
		cecPatGrow(info, pres, nInputs, nRegs, *nWords)
		*nWords *= 2
	}
}
func Cec_ManPatCollectPatterns(p *Cec_ManPat_t, nInputs, nWordsInit int32) *Vec_Ptr_t {
	if nWordsInit <= 0 {
		panic("CEC simulation word count must be positive")
	}
	start := util.Abc_Clock()
	old := p.iStart
	words := nWordsInit
	info := vec.Vec_PtrAllocSimInfo(int(nInputs), int(words))
	pres := vec.Vec_PtrAllocSimInfo(int(nInputs), int(words))
	gia.Gia_ManRandomInfo(info, 0, 0, words)
	for int(p.iStart) < vec.Vec_StrSize(p.vStorage) {
		Cec_ManPatRestore(p, p.vPattern1)
		cecPatInsert(info, pres, vec.Vec_IntArray(p.vPattern1), nInputs, 0, nWordsInit, &words)
	}
	p.nSeries = words / nWordsInit
	p.iStart = old
	elapsed := util.Abc_Clock() - start
	p.timePack += elapsed
	p.timeTotal += elapsed
	if p.fVerbose != 0 {
		Cec_ManPatPrintStats(p)
	}
	return info
}
func Cec_ManPatPackPatterns(store *Vec_Int_t, nInputs, nRegs, nWordsInit int32) *Vec_Ptr_t {
	if nRegs < 0 || nRegs > nInputs || nWordsInit <= 0 {
		panic("invalid CEC pattern packing dimensions")
	}
	words := nWordsInit
	info := vec.Vec_PtrAllocSimInfo(int(nInputs), int(words))
	pres := vec.Vec_PtrAllocSimInfo(int(nInputs), int(words))
	gia.Gia_ManRandomInfo(info, nRegs, 0, words)
	a := vec.Vec_IntArray(store)
	for i := 0; i < len(a); {
		if i+2 > len(a) {
			panic("truncated CEC counterexample")
		}
		n := a[i+1]
		i += 2
		if n <= 0 {
			continue
		}
		if int(n) > len(a)-i {
			panic("truncated CEC pattern")
		}
		cecPatInsert(info, pres, a[i:i+int(n)], nInputs, nRegs, nWordsInit, &words)
		i += int(n)
	}
	return info
}
