package if_

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/extra"
	"os"
	"slices"

	"github.com/glycerine/abcgo/bool/kit"
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

// If_DsdObj_t stores the exact truth table as the canonical semantic value.
// The DSD spelling is cached for diagnostics; algorithms never infer behavior
// from parsing that spelling back into a truth table.
type If_DsdObj_t struct {
	Id                  int32
	Type                If_DsdType_t
	nSupp, fMark, Count int32
	nFans               int32
	pFans               []int32
	truth               []uint64
	supportVars         []int32
	text                string
	network             *dau.Dss_Ntk_t
	permutation         []byte
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

func dsdSupport(truth []uint64, n int32) []int32 {
	vars := make([]int32, n)
	base := append([]uint64(nil), truth...)
	return vars[:dau.Dau_DsdMinBase(base, n, vars)]
}

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
func If_DsdManCheckDec(p *If_DsdMan_t, lit int32) int32 {
	if dsdObject(p, lit).nSupp > p.LutSize {
		return 1
	}
	return 0
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
	o := dsdObject(p, lit)
	// Transitional diagnostic/balancing caches are derived from the canonical
	// structure, never used for interning or decomposition decisions.
	if o.network == nil {
		o.text = ifDsdText(p, lit&^1, nil)
		o.truth = append([]word(nil), If_DsdManComputeTruth(p, lit&^1, nil)[:dsdWordCount(o.nSupp)]...)
		o.network = dau.Dss_NtkCreate(o.text, o.nSupp, nil)
		o.permutation = make([]byte, o.nSupp)
		o.supportVars = make([]int32, o.nSupp)
		for i := range o.permutation {
			o.permutation[i] = byte(2 * i)
			o.supportVars[i] = int32(i)
		}
	}
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
	if verbose != 0 {
		If_DsdManPrint(p, "", 0, 0, 0, 0, 0)
	}
	If_ManSatUnbuild(p.pSat)
	// All remaining allocations are Go-owned vectors, slices and node records.
	// Clear the owner as a whole so all caches become collectible together.
	*p = If_DsdMan_t{}
}

func If_DsdManCleanMarks(p *If_DsdMan_t, verbose int32) {
	for _, e := range Vec_PtrArray(&p.vObjs) {
		e.(*If_DsdObj_t).fMark = 0
	}
}
func If_DsdManInvertMarks(p *If_DsdMan_t, verbose int32) {
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
func If_DsdManMerge(p, other *If_DsdMan_t) {
	if p == nil || other == nil {
		return
	}
	for i := 1; i < Vec_PtrSize(&other.vObjs); i++ {
		o := *If_DsdVecObj(&other.vObjs, int32(i))
		o.Id = int32(Vec_PtrSize(&p.vObjs))
		Vec_PtrPush(&p.vObjs, &o)
	}
}
func If_DsdManFilter(p *If_DsdMan_t, limit int32) *If_DsdMan_t {
	if p == nil {
		return nil
	}
	q := If_DsdManAlloc(p.nVars, p.LutSize)
	for i := 0; i < Vec_PtrSize(&p.vObjs); i++ {
		o := If_DsdVecObj(&p.vObjs, int32(i))
		if i < int(p.nVars+1) || limit <= 0 || o.nSupp <= limit {
			clone := *o
			clone.Id = int32(Vec_PtrSize(&q.vObjs))
			Vec_PtrPush(&q.vObjs, &clone)
		}
	}
	return q
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
func If_DsdManPrint(p *If_DsdMan_t, fileName string, number, support, occurs, dump, verbose int32) {
	if fileName == "" {
		for i, e := range Vec_PtrArray(&p.vObjs) {
			o := e.(*If_DsdObj_t)
			if support == 0 || o.nSupp == support {
				fmt.Printf("%d %s\n", i, o.text)
			}
		}
		return
	}
	f, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer f.Close()
	for i, e := range Vec_PtrArray(&p.vObjs) {
		o := e.(*If_DsdObj_t)
		if support == 0 || o.nSupp == support {
			fmt.Fprintf(f, "%d %s\n", i, o.text)
		}
	}
}
func If_DsdManSave(p *If_DsdMan_t, fileName string) { If_DsdManPrint(p, fileName, 0, 0, 0, 0, 0) }
func If_DsdManLoad(fileName string) *If_DsdMan_t    { return nil }

func dsdSopCover(obj *dau.Dss_Obj_t) *Vec_Int_t {
	truth32 := make([]uint32, (len(obj.Truth) * 2))
	for i, word := range obj.Truth {
		truth32[2*i], truth32[2*i+1] = uint32(word), uint32(word>>32)
	}
	var cover []uint32
	kit.Kit_TruthIsop(truth32, int32(len(obj.Fans)), &cover, 0)
	result := vec.Vec_IntAlloc(len(cover))
	for _, cube := range cover {
		vec.Vec_IntPush(result, int32(cube))
	}
	return result
}

// dsdDelayRec mirrors the source recursion: variables inherit arrival times,
// muxes cost two stages, prime nodes use their ISOP, and associative nodes use
// the logarithmic counter scheduler.
func dsdDelayRec(network *dau.Dss_Ntk_t, obj *dau.Dss_Obj_t, pinMap []byte, times []int32, support *int32, all int32) (int32, uint64) {
	obj = dau.Dss_Regular(obj)
	switch obj.Type {
	case dau.DAU_DSD_VAR:
		if int(obj.Var) >= len(pinMap) {
			panic("if: DSD variable outside permutation")
		}
		pin := int(Abc_Lit2Var(int32(pinMap[obj.Var])))
		if pin >= len(times) {
			panic("if: DSD variable outside cut")
		}
		*support++
		return times[pin], If_CutPinDelayInit(int32(pin))
	case dau.DAU_DSD_MUX:
		if len(obj.Fans) != 3 {
			panic("if: malformed DSD mux")
		}
		delays, profiles := make([]int32, 3), make([]uint64, 3)
		for i, lit := range obj.Fans {
			delays[i], profiles[i] = dsdDelayRec(network, dau.Dss_Lit2Obj(network.Objects, lit), pinMap, times, support, all)
		}
		profile := If_CutPinDelayMax(If_CutPinDelayMax(profiles[0], profiles[1], all, 1), If_CutPinDelayMax(profiles[0], profiles[2], all, 1), all, 1)
		return 2 + max(delays[0], max(delays[1], delays[2])), profile
	case dau.DAU_DSD_PRIME:
		delays, profiles := make([]int32, len(obj.Fans)), make([]uint64, len(obj.Fans))
		for i, lit := range obj.Fans {
			delays[i], profiles[i] = dsdDelayRec(network, dau.Dss_Lit2Obj(network.Objects, lit), pinMap, times, support, all)
		}
		cover := dsdSopCover(obj)
		defer vec.Vec_IntFree(cover)
		var profile uint64
		delay := If_CutSopBalancePinDelaysInt(cover, delays, profiles, int32(len(obj.Fans)), &profile)
		return delay, profile
	case dau.DAU_DSD_AND, dau.DAU_DSD_XOR:
		var count int32
		delays, profiles, counters := make([]int32, len(obj.Fans)), make([]uint64, len(obj.Fans)), make([]int32, len(obj.Fans))
		var delay int32
		for i, lit := range obj.Fans {
			delays[i], profiles[i] = dsdDelayRec(network, dau.Dss_Lit2Obj(network.Objects, lit), pinMap, times, support, all)
			delay = If_LogCounterPinDelays(counters, &count, profiles, delays[i], profiles[i], all, 0)
		}
		return delay, If_LogPinDelaysMulti(profiles, count, all, 0)
	default:
		panic("if: unsupported DSD node type")
	}
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
	obj := dsdObject(p.pIfDsdMan, If_CutDsdLit(p, cut))
	if obj.network == nil {
		return -1
	}
	times := make([]int32, cut.nLeaves)
	for i := range times {
		times[i] = int32(If_ObjCutBest(If_CutLeaf(p, cut, int32(i))).Delay)
	}
	var support int32
	delay, profile := dsdDelayRec(obj.network, obj.network.Root, If_CutDsdPerm(p, cut), times, &support, int32(cut.nLeaves))
	assert(support == int32(cut.nLeaves), "support == cut.nLeaves")
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
	obj := dsdObject(p.pIfDsdMan, lit)
	if obj.network == nil {
		return -1
	}
	times := ifDelayTimes(p, cut)
	var area int32
	delay, root := ifDsdBalance(obj.network, obj.network.Root, If_CutDsdPerm(p, cut), times, aig, int32(cut.nLeaves), &area)
	if delay < 0 {
		return -1
	}
	if aig != nil {
		area = int32(vec.Vec_IntSize(aig) / 2)
		if root>>1 != int32(cut.nLeaves)+area-1 {
			panic("if: DSD AIG root is not final")
		}
		vec.Vec_IntPush(aig, root&1^lit&1^int32(cut.fCompl))
	}
	cut.Cost = uint32(area)
	return float32(delay)
}

// ifDsdBalance builds the compact AND graph while preserving DSD edge phases.
// area is accumulated for timing-only callers; graph callers count emitted ANDs.
func ifDsdBalance(network *dau.Dss_Ntk_t, edge *dau.Dss_Obj_t, pins []byte, times []int32, aig *Vec_Int_t, support int32, area *int32) (int32, int32) {
	phase := dau.Dss_IsComplement(edge)
	obj := dau.Dss_Regular(edge)
	if obj.Type == dau.DAU_DSD_VAR {
		if obj.Var < 0 || int(obj.Var) >= len(pins) {
			panic("if: invalid DSD variable")
		}
		lit := int32(pins[obj.Var])
		pin := lit >> 1
		if pin < 0 || int(pin) >= len(times) {
			panic("if: invalid DSD pin")
		}
		return times[pin], lit ^ phase
	}
	delays := make([]int32, len(obj.Fans))
	lits := make([]int32, len(obj.Fans))
	for i, fan := range obj.Fans {
		delays[i], lits[i] = ifDsdBalance(network, dau.Dss_Lit2Obj(network.Objects, fan), pins, times, aig, support, area)
		if delays[i] < 0 {
			return -1, 0
		}
	}
	switch obj.Type {
	case dau.DAU_DSD_MUX:
		if len(lits) != 3 {
			panic("if: invalid DSD mux")
		}
		var root int32
		if aig == nil {
			*area += 3
		} else {
			root = If_LogCreateMux(aig, lits[0], lits[1], lits[2], support)
		}
		return 2 + max(delays[0], delays[1], delays[2]), root ^ phase
	case dau.DAU_DSD_PRIME:
		cover := dsdSopCover(obj)
		defer vec.Vec_IntFree(cover)
		var root int32
		delay := ifCutSopBalanceEvalInt(cover, delays, lits, aig, &root, int32(len(lits)), support, area)
		return delay, root ^ phase
	case dau.DAU_DSD_AND, dau.DAU_DSD_XOR:
		if len(lits) == 0 {
			panic("if: empty DSD operator")
		}
		var count, delay int32
		queue, queuedLits := make([]int32, len(lits)), make([]int32, len(lits))
		xor := int32(0)
		if obj.Type == dau.DAU_DSD_XOR {
			xor = 1
		}
		for i := range lits {
			delay = If_LogCounterAddAig(queue, &count, queuedLits, delays[i], lits[i], aig, support, 0, xor)
		}
		var root int32
		if aig == nil {
			*area += int32(len(lits) - 1)
		} else {
			root = If_LogCreateAndXorMulti(aig, queuedLits, count, support, xor)
		}
		return delay, root ^ phase
	default:
		panic("if: unsupported DSD node")
	}
}
