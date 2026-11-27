package if_

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/extra"
	"slices"

	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

const DSD_VERSION = "dsd1"

type If_DsdType_t int32

const (
	IF_DSD_NONE If_DsdType_t = iota
	IF_DSD_CONST0
	IF_DSD_VAR
	IF_DSD_AND
	IF_DSD_XOR
	IF_DSD_MUX
	IF_DSD_PRIME
)

// If_DsdObj_t records shared canonical structure. Prime truth tables live in
// the manager; occurrence permutations belong to the cuts that use a node.
type If_DsdObj_t struct {
	Id                  int32
	Type                If_DsdType_t
	nSupp, fMark, Count int32
	nFans               int32
	pFans               []int32
}

func If_DsdObjClean(p *If_DsdObj_t)      { *p = If_DsdObj_t{} }
func If_DsdObjId(p *If_DsdObj_t) int32   { return p.Id }
func If_DsdObjType(p *If_DsdObj_t) int32 { return int32(p.Type) }
func If_DsdObjIsVar(p *If_DsdObj_t) int32 {
	if p.Type == IF_DSD_VAR {
		return 1
	}
	return 0
}
func If_DsdObjSuppSize(p *If_DsdObj_t) int32          { return p.nSupp }
func If_DsdObjFaninNum(p *If_DsdObj_t) int32          { return p.nFans }
func If_DsdObjFaninC(p *If_DsdObj_t, i int32) int32   { return Abc_LitIsCompl(p.pFans[i]) }
func If_DsdObjFaninLit(p *If_DsdObj_t, i int32) int32 { return p.pFans[i] }
func If_DsdVecObj(v *Vec_Ptr_t, id int32) *If_DsdObj_t {
	return Vec_PtrEntry(v, int(id)).(*If_DsdObj_t)
}
func If_DsdVecConst0(v *Vec_Ptr_t) *If_DsdObj_t              { return If_DsdVecObj(v, 0) }
func If_DsdVecVar(v *Vec_Ptr_t, variable int32) *If_DsdObj_t { return If_DsdVecObj(v, variable+1) }
func If_DsdVecObjSuppSize(v *Vec_Ptr_t, id int32) int32      { return If_DsdVecObj(v, id).nSupp }
func If_DsdVecLitSuppSize(v *Vec_Ptr_t, lit int32) int32 {
	return If_DsdVecObjSuppSize(v, Abc_Lit2Var(lit))
}
func If_DsdVecObjRef(v *Vec_Ptr_t, id int32) int32 { return If_DsdVecObj(v, id).Count }
func If_DsdVecObjIncRef(v *Vec_Ptr_t, id int32) {
	o := If_DsdVecObj(v, id)
	if o.Count < 0x3ffff {
		o.Count++
	}
}
func If_DsdObjFanin(v *Vec_Ptr_t, p *If_DsdObj_t, i int32) *If_DsdObj_t {
	return If_DsdVecObj(v, Abc_Lit2Var(p.pFans[i]))
}
func If_DsdVecObjMark(v *Vec_Ptr_t, id int32) int32 { return If_DsdVecObj(v, id).fMark }
func If_DsdVecObjSetMark(v *Vec_Ptr_t, id int32)    { If_DsdVecObj(v, id).fMark = 1 }
func If_DsdVecObjClearMark(v *Vec_Ptr_t, id int32)  { If_DsdVecObj(v, id).fMark = 0 }

func dsdObject(p *If_DsdMan_t, lit int32) *If_DsdObj_t {
	if p == nil || Abc_Lit2Var(lit) < 0 || int(Abc_Lit2Var(lit)) >= Vec_PtrSize(&p.vObjs) {
		panic("if: invalid DSD literal")
	}
	return If_DsdVecObj(&p.vObjs, Abc_Lit2Var(lit))
}

func dsdWordCount(n int32) int { return (1<<uint(n) + 63) / 64 }

// If_DsdManAlloc creates the source constant and one generic variable.
func If_DsdManAlloc(nVars, lutSize int32) *If_DsdMan_t {
	if nVars < 0 || nVars > 16 {
		panic("if: DSD manager supports 0..16 variables")
	}
	p := &If_DsdMan_t{pStore: fmt.Sprintf("%02d.dsd", nVars), nVars: nVars, LutSize: lutSize, nWords: int32(dsdWordCount(nVars)), nBins: abcutil.Abc_PrimeCudd(100000), nConfigWords: 1, vTemp1: vec.Vec_IntAlloc(32), vTemp2: vec.Vec_IntAlloc(32), vCover: vec.Vec_IntAlloc(0)}
	p.pBins = make([]int32, p.nBins)
	p.pTtElems = make([][]word, nVars+1)
	for v := int32(0); v <= nVars; v++ {
		p.pTtElems[v] = make([]word, p.nWords)
		if v == nVars {
			continue
		}
		for i := range p.pTtElems[v] {
			if v < 6 {
				p.pTtElems[v][i] = truth6Variable(v)
			} else if i>>uint(v-6)&1 != 0 {
				p.pTtElems[v][i] = ^word(0)
			}
		}
	}
	If_DsdObjAlloc(p, int32(IF_DSD_CONST0), 0)
	If_DsdObjAlloc(p, int32(IF_DSD_VAR), 0).nSupp = 1
	for v := int32(3); v <= nVars; v++ {
		p.vTtMem[v] = vec.Vec_MemAlloc(dsdWordCount(v), 12)
		vec.Vec_MemHashAlloc(p.vTtMem[v], 10000)
		p.vTtDecs[v] = vec.Vec_PtrAlloc(1000)
	}
	for v := int32(2); v < nVars; v++ {
		p.pSched[v] = extra.Extra_GreyCodeSchedule(v)
	}
	if lutSize != 0 {
		p.pSat = If_ManSatBuildXY(lutSize)
	}
	return p
}

func If_DsdManFileName(p *If_DsdMan_t) string { return p.pStore }
func If_DsdManVarNum(p *If_DsdMan_t) int32    { return p.nVars }
func If_DsdManObjNum(p *If_DsdMan_t) int32    { return int32(Vec_PtrSize(&p.vObjs)) }
func If_DsdManLutSize(p *If_DsdMan_t) int32   { return p.LutSize }
func If_DsdManTtBitNum(p *If_DsdMan_t) int32  { return p.nTtBits }
func If_DsdManPermBitNum(p *If_DsdMan_t) int32 {
	return (abcutil.Abc_Base2Log(uint32(p.nVars+1)) + 1) * p.nVars
}
func If_DsdManSetLutSize(p *If_DsdMan_t, n int32)       { p.LutSize = n }
func If_DsdManSuppSize(p *If_DsdMan_t, lit int32) int32 { return dsdObject(p, lit).nSupp }

// If_DsdManCheckDec reports the stored unusable mark, as in ifDsd.c.
// Decomposability and cell tuning cannot be inferred from support size alone.
func If_DsdManCheckDec(p *If_DsdMan_t, lit int32) int32 {
	return If_DsdManReadMark(p, lit)
}
func If_DsdManReadMark(p *If_DsdMan_t, lit int32) int32 { return dsdObject(p, lit).fMark }
func If_DsdManSetNewAsUseless(p *If_DsdMan_t)           { p.fNewAsUseless = 1 }
func If_DsdManGetCellStr(p *If_DsdMan_t) string         { return p.pCellStr }

// If_DsdManCompute decomposes, canonicalizes and interns the source DSD tree.
func If_DsdManCompute(p *If_DsdMan_t, truth []word, nLeaves int32, permutation []byte, lutStruct string) int32 {
	if p == nil || nLeaves < 0 || nLeaves > p.nVars || len(truth) < dsdWordCount(nLeaves) || len(permutation) < int(nLeaves) {
		panic("if: invalid DSD compute arguments")
	}
	copyTruth := make([]word, p.nWords)
	copy(copyTruth, truth[:dsdWordCount(nLeaves)])
	var text []byte
	prime := dau.Dau_DsdDecompose(copyTruth, nLeaves, 0, 1, &text)
	if prime > 0 {
		abcutil.Abc_TtStretch6(copyTruth, prime, p.nVars)
	}
	for i := int32(0); i < nLeaves; i++ {
		permutation[i] = 255
	}
	var support int32
	lit := If_DsdManAddDsd(p, string(text), copyTruth, permutation, &support)
	assert(support == nLeaves, "canonical DSD covers all cut leaves")
	result := If_DsdManComputeTruth(p, lit, permutation)
	if !slices.Equal(result[:dsdWordCount(nLeaves)], truth[:dsdWordCount(nLeaves)]) {
		panic("if: canonical DSD verification failed")
	}
	If_DsdVecObjIncRef(&p.vObjs, lit>>1)
	return lit
}

func If_DsdManCheckXY(p *If_DsdMan_t, lit, lutSize, derive int32, maskNot uint32, highEffort, verbose int32) uint32 {
	set := If_DsdManCheckXY_int(p, lit, lutSize, derive, maskNot, verbose)
	if set == 0 && highEffort != 0 {
		truth := If_DsdManComputeTruth(p, lit, nil)
		set = If_ManSatCheckXYall(p.pSat, lutSize, truth, dsdObject(p, lit).nSupp, p.vTemp1)
	}
	return set
}

func If_DsdManCheckXYZ(p *If_DsdMan_t, lit, lutSize, derive, verbose int32) uint32 { return ^uint32(0) }

func If_DsdManFree(p *If_DsdMan_t, verbose int32) {
	if p == nil {
		return
	}
	// Release the owner even if the existing vector dump API panics on I/O.
	defer func() {
		If_ManSatUnbuild(p.pSat)
		// All remaining caches, vectors and records are Go-owned.
		*p = If_DsdMan_t{}
	}()
	if verbose != 0 {
		If_DsdManPrint(p, "", 0, 0, 0, 0, 0)
		for v := int32(3); v <= p.nVars; v++ {
			vec.Vec_MemDumpTruthTables(p.vTtMem[v], fmt.Sprintf("dumpdsd%02d", v), int(v))
		}
	}
}

func If_DsdManCleanMarks(p *If_DsdMan_t, verbose int32) {
	p.pCellStr = ""
	vec.Vec_WrdFree(p.vConfigs)
	p.vConfigs = nil
	for _, e := range Vec_PtrArray(&p.vObjs) {
		e.(*If_DsdObj_t).fMark = 0
	}
}
func If_DsdManInvertMarks(p *If_DsdMan_t, verbose int32) {
	p.pCellStr = ""
	for _, e := range Vec_PtrArray(&p.vObjs) {
		o := e.(*If_DsdObj_t)
		if o.fMark == 0 {
			o.fMark = 1
		} else {
			o.fMark = 0
		}
	}
}
func If_DsdManCleanOccur(p *If_DsdMan_t, verbose int32) {
	for _, e := range Vec_PtrArray(&p.vObjs) {
		e.(*If_DsdObj_t).Count = 0
	}
}
func If_DsdManAllocIsops(p *If_DsdMan_t, nLutSize int32) {
	if p.vIsops[3] != nil {
		return
	}
	if Vec_PtrSize(&p.vObjs) > 2 {
		fmt.Println("Warning: DSD manager is already started without ISOPs.")
	}
	for v := int32(3); v <= nLutSize; v++ {
		p.vIsops[v] = vec.Vec_WecAlloc(100)
		for i := 0; i < vec.Vec_MemEntryNum(p.vTtMem[v]); i++ {
			level := vec.Vec_WecPushLevel(p.vIsops[v])
			cover, complement := ifDsdPrimeCover(vec.Vec_MemReadEntry(p.vTtMem[v], i), v)
			if complement >= 0 && len(cover) <= 8 {
				for _, cube := range cover {
					Vec_IntPush(level, cube)
				}
				p.vIsopComplement[v] = append(p.vIsopComplement[v], complement != 0)
			} else {
				p.vIsopComplement[v] = append(p.vIsopComplement[v], false)
			}
		}
	}
}
func If_DsdManHasMarks(p *If_DsdMan_t) int32 {
	for _, entry := range Vec_PtrArray(&p.vObjs) {
		if entry.(*If_DsdObj_t).fMark != 0 {
			return 1
		}
	}
	return 0
}
func If_DsdManMerge(p, other *If_DsdMan_t) {
	if p == nil || other == nil {
		return
	}
	if p.nVars < other.nVars {
		fmt.Println("The number of variables should be the same or smaller.")
		return
	}
	if p.LutSize != other.LutSize {
		fmt.Println("LUT size should be the same.")
		return
	}
	assert(p.nTtBits == other.nTtBits && p.nConfigWords == other.nConfigWords, "merged DSD configuration dimensions agree")
	if If_DsdManHasMarks(p) != If_DsdManHasMarks(other) {
		old, new := "", ""
		if If_DsdManHasMarks(p) == 0 {
			old = "no "
		}
		if If_DsdManHasMarks(other) == 0 {
			new = "no "
		}
		fmt.Printf("Warning! Old manager has %smarks while new manager has %smarks.\n", old, new)
	}
	mapping := make([]int32, If_DsdManObjNum(other))
	mapping[1] = 1
	configurations := p.vConfigs != nil && other.vConfigs != nil
	if configurations {
		vec.Vec_WrdFillExtra(p.vConfigs, int(p.nConfigWords*(If_DsdManObjNum(p)+If_DsdManObjNum(other))), 0)
	}
	var fans [16]int32
	for i := int32(2); i < If_DsdManObjNum(other); i++ {
		o := If_DsdVecObj(&other.vObjs, i)
		for k, fan := range o.pFans {
			fans[k] = mapping[fan>>1]<<1 | fan&1
		}
		var truth []word
		if o.Type == IF_DSD_PRIME {
			truth = If_DsdObjTruth(other, o)
		}
		id := If_DsdObjFindOrAdd(p, int32(o.Type), fans[:], o.nFans, truth)
		if o.fMark != 0 {
			If_DsdVecObj(&p.vObjs, id).fMark = 1
		}
		if configurations && int(p.nConfigWords*i) < len(vec.Vec_WrdArray(other.vConfigs)) {
			from, to := p.nConfigWords*i, p.nConfigWords*id
			copy(vec.Vec_WrdArray(p.vConfigs)[to:to+p.nConfigWords], vec.Vec_WrdArray(other.vConfigs)[from:from+p.nConfigWords])
		}
		mapping[i] = id
	}
	if configurations {
		vec.Vec_WrdShrink(p.vConfigs, int(p.nConfigWords*If_DsdManObjNum(p)))
	}
}
func If_DsdManFilter_rec(dst, src *If_DsdMan_t, id int32, mapping *Vec_Int_t) {
	if Vec_IntEntry(mapping, int(id)) >= 0 {
		return
	}
	o := If_DsdVecObj(&src.vObjs, id)
	for _, fan := range o.pFans {
		If_DsdManFilter_rec(dst, src, fan>>1, mapping)
	}
	var fans [16]int32
	for k, fan := range o.pFans {
		fans[k] = Vec_IntEntry(mapping, int(fan>>1))<<1 | fan&1
	}
	var truth []word
	if o.Type == IF_DSD_PRIME {
		truth = If_DsdObjTruth(src, o)
	}
	copied := If_DsdObjFindOrAdd(dst, int32(o.Type), fans[:], o.nFans, truth)
	target := If_DsdVecObj(&dst.vObjs, copied)
	if o.fMark != 0 {
		target.fMark = 1
	}
	target.Count = o.Count
	vec.Vec_IntWriteEntry(mapping, int(id), copied)
}
func If_DsdManFilter(p *If_DsdMan_t, limit int32) *If_DsdMan_t {
	if p == nil {
		return nil
	}
	result := If_DsdManAlloc(p.nVars, p.LutSize)
	mapping := vec.Vec_IntStartFull(Vec_PtrSize(&p.vObjs))
	defer vec.Vec_IntFree(mapping)
	vec.Vec_IntWriteEntry(mapping, 0, 0)
	vec.Vec_IntWriteEntry(mapping, 1, 1)
	for id := int32(2); id < If_DsdManObjNum(p); id++ {
		if If_DsdVecObj(&p.vObjs, id).Count >= limit {
			If_DsdManFilter_rec(result, p, id, mapping)
		}
	}
	return result
}

// If_DsdManGetFuncConfig returns the stored cell-input permutation and
// programmable bits, or nil when the function has not been configured.
func If_DsdManGetFuncConfig(p *If_DsdMan_t, lit int32) []word {
	if p == nil || p.vConfigs == nil || p.nConfigWords <= 0 {
		return nil
	}
	id := Abc_Lit2Var(lit)
	start := int(id * p.nConfigWords)
	end := start + int(p.nConfigWords)
	data := vec.Vec_WrdArray(p.vConfigs)
	if id < 0 || start < 0 || end > len(data) {
		return nil
	}
	return data[start:end:end]
}

func If_CutDsdBalancePinDelays(p *If_Man_t, cut *If_Cut_t, permutation []byte) int32 {
	if len(permutation) < int(cut.nLeaves) {
		return -1
	}
	if cut.nLeaves == 0 {
		return 0
	}
	if cut.nLeaves == 1 {
		permutation[0] = 0
		return int32(If_ObjCutBest(If_CutLeaf(p, cut, 0)).Delay)
	}
	times := ifDelayTimes(p, cut)
	var support int32
	var profile word
	delay := If_CutDsdBalancePinDelays_rec(p.pIfDsdMan, If_CutDsdLit(p, cut)>>1, times, &profile, &support, int32(cut.nLeaves), If_CutDsdPerm(p, cut))
	assert(support == int32(cut.nLeaves), "DSD profile visits complete support")
	If_CutPinDelayTranslate(profile, int32(cut.nLeaves), permutation)
	return delay
}

func If_CutDsdBalanceEval(p *If_Man_t, cut *If_Cut_t, aig *Vec_Int_t) float32 {
	cut.fUser = 1
	if aig != nil {
		vec.Vec_IntClear(aig)
	}
	lit := If_CutDsdLit(p, cut)
	if cut.nLeaves == 0 {
		if aig != nil {
			vec.Vec_IntPush(aig, lit&1)
		}
		cut.Cost = 0
		return 0
	}
	if cut.nLeaves == 1 {
		if aig != nil {
			vec.Vec_IntPush(aig, 0)
			vec.Vec_IntPush(aig, lit&1)
		}
		cut.Cost = 0
		return If_ObjCutBest(If_CutLeaf(p, cut, 0)).Delay
	}
	var timesBuf [IF_MAX_LUTSIZE]int32
	times := ifDelayTimesInto(timesBuf[:cut.nLeaves], p, cut)
	var area int32
	delay := If_CutDsdBalanceEvalInt(p.pIfDsdMan, lit^int32(cut.fCompl), times, aig, &area, If_CutDsdPerm(p, cut))
	cut.Cost = uint32(area)
	return float32(delay)
}
