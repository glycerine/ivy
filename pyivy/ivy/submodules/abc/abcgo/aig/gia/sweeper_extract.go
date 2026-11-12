package gia

import "fmt"

func Gia_ManExtract_rec(p *Gia_Man_t, obj *Gia_Obj_t, ids *Vec_Int_t) {
	if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	Gia_ManExtract_rec(p, Gia_ObjFanin0(obj), ids)
	Gia_ManExtract_rec(p, Gia_ObjFanin1(obj), ids)
	Vec_IntPush(ids, obj.id)
}
func Gia_SweeperExtractUserLogic(p *Gia_Man_t, probes *Vec_Int_t, inputNames, outputNames *Vec_Ptr_t) *Gia_Man_t {
	if inputNames != nil && Vec_PtrSize(inputNames) != int(Gia_ManPiNum(p)) || outputNames != nil && Vec_PtrSize(outputNames) != Vec_IntSize(probes) {
		panic("gia: incorrect sweeper names")
	}
	Gia_ManIncrementTravId(p)
	ids := Vec_IntAlloc(1000)
	for _, id := range Vec_IntArray(probes) {
		Gia_ManExtract_rec(p, Gia_Regular(Gia_Lit2Obj(p, Gia_SweeperProbeLit(p, id))), ids)
	}
	dst := Gia_ManStart(1 + Gia_ManPiNum(p) + int32(Vec_IntSize(ids)+Vec_IntSize(probes)) + 100)
	dst.pName, dst.pSpec = p.pName, p.pSpec
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		Gia_ManPi(p, i).Value = uint32(Gia_ManAppendCi(dst))
	}
	Gia_ManHashStart(dst)
	old := make([]uint32, 0, Vec_IntSize(ids))
	for _, id := range Vec_IntArray(ids) {
		obj := Gia_ManObj(p, id)
		old = append(old, obj.Value)
		obj.Value = uint32(Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
	}
	Gia_ManHashStop(dst)
	for _, id := range Vec_IntArray(probes) {
		lit := Gia_SweeperProbeLit(p, id)
		Gia_ManAppendCo(dst, int32(Gia_ManObj(p, lit>>1).Value)^lit&1)
	}
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		Gia_ManPi(p, i).Value = 0
	}
	for i, id := range Vec_IntArray(ids) {
		Gia_ManObj(p, id).Value = old[i]
	}
	if Gia_ManHasDangling(dst) != 0 {
		previous := dst
		dst = Gia_ManCleanup(previous)
		Gia_ManStop(previous)
	}
	if inputNames != nil {
		dst.vNamesIn = Vec_PtrDupStr(inputNames)
	}
	if outputNames != nil {
		dst.vNamesOut = Vec_PtrDupStr(outputNames)
	}
	return dst
}
func Gia_SweeperLogicDump(p *Gia_Man_t, probes *Vec_Int_t, dumpConditions int32, path string) {
	out := Gia_SweeperExtractUserLogic(p, probes, nil, nil)
	conditions := Gia_SweeperCondVector(p)
	fmt.Print("Dumping logic cones")
	if dumpConditions != 0 && Vec_IntSize(conditions) > 0 {
		cond := Gia_SweeperExtractUserLogic(p, conditions, nil, nil)
		Gia_ManDupAppendShare(out, cond)
		out.nConstrs = Gia_ManPoNum(cond)
		Gia_ManHashStop(out)
		Gia_ManStop(cond)
		fmt.Print(" and conditions")
	}
	if err := Gia_AigerWrite(out, path, 0, 0, 0); err != nil {
		fmt.Printf(" failed: %v\n", err)
	} else {
		fmt.Printf(" into file %q.\n", path)
	}
	Gia_ManStop(out)
}
func Gia_SweeperGraft(dst *Gia_Man_t, probes *Vec_Int_t, src *Gia_Man_t) *Vec_Int_t {
	if Gia_SweeperIsRunning(dst) == 0 {
		panic("gia: sweeper is not running")
	}
	if probes != nil && Vec_IntSize(probes) != int(Gia_ManPiNum(src)) || probes == nil && Gia_ManPiNum(dst) != Gia_ManPiNum(src) {
		panic("gia: graft input count mismatch")
	}
	Gia_ManConst0(src).Value = 0
	for i := int32(0); i < Gia_ManPiNum(src); i++ {
		lit := int32(0)
		if probes != nil {
			lit = Gia_SweeperProbeLit(dst, Vec_IntEntry(probes, int(i)))
		} else {
			lit = Gia_Obj2Lit(dst, Gia_ManPi(dst, i))
		}
		Gia_ManPi(src, i).Value = uint32(lit)
	}
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			obj.Value = uint32(Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		}
	}
	out := Vec_IntAlloc(int(Gia_ManPoNum(src)))
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		Vec_IntPush(out, Gia_ObjFanin0Copy(Gia_ManPo(src, i)))
	}
	return out
}
