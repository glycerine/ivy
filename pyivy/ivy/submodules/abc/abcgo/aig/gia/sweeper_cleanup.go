package gia

import (
	"time"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Gia_SweeperCleanup rebuilds all live probes, retaining their IDs and
// conditions while discarding unused nodes and the incremental SAT database.
func Gia_SweeperCleanup(p *Gia_Man_t, command string) *Gia_Man_t {
	s := sweeper(p)
	ids := Gia_SweeperCollectValidProbeIds(p)
	// Keep probe cones as temporary outputs so graph cleanup and optional
	// commands cannot discard the logic they represent.
	tmp := Gia_SweeperExtractUserLogic(p, ids, nil, nil)
	if command != "" {
		frame := abcmain.Abc_FrameGetGlobalFrame()
		abcmain.Abc_FrameUpdateGia(frame, tmp)
		if abcmain.Cmd_CommandExecute(frame, command) != 0 {
			panic("gia: sweeper cleanup command failed")
		}
		var ok bool
		tmp, ok = abcmain.Abc_FrameGetGia(frame).(*Gia_Man_t)
		if !ok || tmp == nil {
			panic("gia: sweeper cleanup command removed graph")
		}
	}
	if Gia_ManPiNum(tmp) != Gia_ManPiNum(p) || Gia_ManPoNum(tmp) != int32(Vec_IntSize(ids)) {
		panic("gia: cleanup command changed probe interface")
	}
	result := Gia_ManStart(Gia_ManObjNum(tmp) + 100)
	result.pName, result.PSpec = tmp.pName, tmp.PSpec
	Gia_ManFillValue(tmp)
	Gia_ManConst0(tmp).Value = 0
	for i := int32(0); i < Gia_ManPiNum(tmp); i++ {
		Gia_ManPi(tmp, i).Value = uint32(Gia_ManAppendCi(result))
	}
	Gia_ManHashStart(result)
	for _, obj := range tmp.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			obj.Value = uint32(Gia_ManHashAnd(result, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		}
	}
	for i, id := range Vec_IntArray(ids) {
		Vec_IntWriteEntry(s.vProbes, int(id), Gia_ObjFanin0Copy(Gia_ManPo(tmp, int32(i))))
	}
	Gia_ManStop(tmp)
	Vec_IntClear(s.vId2Lit)
	Vec_IntClear(s.vCondAssump)
	Vec_IntClear(s.vFront)
	Vec_IntClear(s.vFanins)
	Vec_IntClear(s.vCexSwp)
	s.vCexUser = nil
	s.pSat = bsat.NewSolver()
	s.pSat.NewVariable()
	s.pSat.NewVariable()
	s.nSatVars = 2
	Swp_ManSetObj2Lit(s, 0, 2)
	if !s.pSat.AddClause(3) {
		panic("gia: cannot reset SAT constant")
	}
	s.timeStart = time.Now()
	s.pGia = result
	p.pData = nil
	result.pData = s
	result.fSweeper = 1
	Gia_ManStop(p)
	return result
}
