package if_

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/bool/kit"
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

// If_DsdManAlloc creates constant and variable records up front, matching the
// source manager's literal numbering (object ID is the literal's variable).
func If_DsdManAlloc(nVars, lutSize int32) *If_DsdMan_t {
	if nVars < 0 || nVars > 16 {
		panic("if: DSD manager supports 0..16 variables")
	}
	p := &If_DsdMan_t{nVars: nVars, LutSize: lutSize, vTemp1: vec.Vec_IntAlloc(16), vTemp2: vec.Vec_IntAlloc(16)}
	Vec_PtrPush(&p.vObjs, &If_DsdObj_t{Id: 0, Type: IF_DSD_CONST0, truth: []uint64{0}})
	for v := int32(0); v < nVars; v++ {
		truth := make([]uint64, dsdWordCount(v+1))
		for x := 0; x < 1<<uint(v+1); x++ {
			if x&(1<<uint(v)) != 0 {
				truth[x>>6] |= uint64(1) << uint(x&63)
			}
		}
		Vec_PtrPush(&p.vObjs, &If_DsdObj_t{Id: v + 1, Type: IF_DSD_VAR, nSupp: 1, truth: truth, text: string(rune('a' + v))})
	}
	return p
}

func If_DsdManFileName(p *If_DsdMan_t) string           { return p.pStore }
func If_DsdManVarNum(p *If_DsdMan_t) int32              { return p.nVars }
func If_DsdManObjNum(p *If_DsdMan_t) int32              { return int32(Vec_PtrSize(&p.vObjs)) }
func If_DsdManLutSize(p *If_DsdMan_t) int32             { return p.LutSize }
func If_DsdManTtBitNum(p *If_DsdMan_t) int32            { return 1 << uint(p.nVars) }
func If_DsdManPermBitNum(p *If_DsdMan_t) int32          { return p.nVars }
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

// If_DsdManCompute records a canonical DSD spelling and exact truth table.
// Returned IDs follow the original even-literal convention; caller's pin map
// is initialized to identity because the truth table is retained in source order.
func If_DsdManCompute(p *If_DsdMan_t, truth []uint64, nLeaves int32, permutation []byte, lutStruct string) int32 {
	if p == nil || nLeaves < 0 || nLeaves > 16 || len(truth) < dsdWordCount(nLeaves) || len(permutation) < int(nLeaves) {
		panic("if: invalid DSD compute arguments")
	}
	words := dsdWordCount(nLeaves)
	copyTruth := append([]uint64(nil), truth[:words]...)
	var text []byte
	decomposeTruth := append([]uint64(nil), copyTruth...)
	prime := dau.Dau_DsdDecompose(decomposeTruth, nLeaves, 0, 1, &text)
	support := dsdSupport(copyTruth, nLeaves)
	for i := range permutation {
		permutation[i] = byte(i << 1)
	}
	obj := &If_DsdObj_t{Id: int32(Vec_PtrSize(&p.vObjs)), Type: IF_DSD_PRIME, nSupp: int32(len(support)), nFans: nLeaves, truth: copyTruth, supportVars: support, text: string(text), network: dau.Dss_NtkCreate(string(text), nLeaves, nil), permutation: append([]byte(nil), permutation[:nLeaves]...)}
	if prime == 0 && len(support) <= 1 {
		obj.Type = IF_DSD_VAR
	}
	if p.fNewAsUseless != 0 {
		obj.fMark = 1
		p.fNewAsUseless = 0
	}
	Vec_PtrPush(&p.vObjs, obj)
	p.pCellStr = lutStruct
	return obj.Id << 1
}

// If_DsdManCheckXY searches exact two-block functional decompositions. A
// returned set uses DAU's two-bit-per-input encoding, preserving shared pins.
func If_DsdManCheckXY(p *If_DsdMan_t, lit, lutSize, derive int32, maskNot uint32, highEffort, verbose int32) uint32 {
	obj := dsdObject(p, lit)
	if obj.nSupp <= lutSize {
		return ^uint32(0)
	}
	if obj.nSupp < 3 || obj.nSupp > 16 {
		return 0
	}
	baseTruth := append([]uint64(nil), obj.truth...)
	dau.Dau_DsdMinBase(baseTruth, obj.nFans, make([]int32, obj.nFans))
	sets := dau.Dau_DecFindSets(baseTruth, obj.nSupp)
	defer vec.Vec_IntFree(sets)
	for _, raw := range vec.Vec_IntArray(sets) {
		set := uint32(raw)
		var mapped uint32
		for i, variable := range obj.supportVars {
			mapped |= ((set >> uint(2*i)) & 3) << uint(2*variable)
		}
		if mapped&maskNot != 0 {
			continue
		}
		var unique, shared, free int32
		dau.Dau_DecSortSet(set, obj.nSupp, &unique, &shared, &free)
		if unique+shared <= lutSize && free+shared+1 <= lutSize {
			return mapped
		}
	}
	return 0
}
func If_DsdManCheckXYZ(p *If_DsdMan_t, lit, lutSize, derive, verbose int32) uint32 { return ^uint32(0) }

func If_DsdManFree(p *If_DsdMan_t, verbose int32) {
	if p == nil {
		return
	}
	p.vObjs = Vec_Ptr_t{}
	p.vTemp1 = nil
	p.vTemp2 = nil
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
func If_DsdManAllocIsops(p *If_DsdMan_t, nLutSize int32) { p.LutSize = nLutSize }
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
func If_DsdManGetFuncConfig(p *If_DsdMan_t, lit int32) *word {
	o := dsdObject(p, lit)
	if len(o.truth) == 0 {
		return nil
	}
	return &o.truth[0]
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
	delay, profile := dsdDelayRec(obj.network, obj.network.Root, obj.permutation, times, &support, int32(cut.nLeaves))
	assert(support == int32(cut.nLeaves), "support == cut.nLeaves")
	If_CutPinDelayTranslate(profile, int32(cut.nLeaves), permutation)
	return delay
}

func If_CutDsdBalanceEval(p *If_Man_t, cut *If_Cut_t, _ *Vec_Int_t) float32 {
	if cut.nLeaves == 0 {
		return 0
	}
	if cut.nLeaves == 1 {
		return If_ObjCutBest(If_CutLeaf(p, cut, 0)).Delay
	}
	permutation := make([]byte, cut.nLeaves)
	return float32(If_CutDsdBalancePinDelays(p, cut, permutation))
}
