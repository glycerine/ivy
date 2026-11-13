package abs

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManPerformGlaOld performs ABC's earlier gate-level abstraction loop.
func Gia_ManPerformGlaOld(graph *gia.Gia_Man_t, pars *Abs_Par_t, startVta int32) int32 {
	start := time.Now()
	if graph == nil || pars == nil || gia.Gia_ManPoNum(graph) != 1 || pars.nFramesMax != 0 && pars.nFramesStart > pars.nFramesMax {
		panic("invalid old GLA arguments")
	}
	po := gia.Gia_ManPo(graph, 0)
	if gia.Gia_ObjIsConst0(gia.Gia_ObjFanin0(po)) != 0 {
		if gia.Gia_ObjFaninC0(po) == 0 {
			fmt.Println("Sequential miter is trivially UNSAT.")
			return 1
		}
		cex := util.Abc_CexAlloc(gia.Gia_ManRegNum(graph), gia.Gia_ManPiNum(graph), 1)
		if err := graph.SetSequentialCounterexample(cex); err != nil {
			panic(err)
		}
		fmt.Println("Sequential miter is trivially SAT.")
		return 0
	}
	ret := int32(-1)
	if gia.Gia_ManGateClasses(graph) == nil {
		if startVta != 0 {
			oldMax, oldStart, oldTimeout, oldDump := pars.nFramesMax, pars.nFramesStart, pars.nTimeOut, pars.fDumpVabs
			pars.nFramesMax = pars.nFramesStart
			pars.nFramesStart = min(pars.nFramesStart/2+1, 3)
			pars.nTimeOut = 20
			pars.fDumpVabs = 0
			ret = Gia_VtaPerformInt(graph, pars)
			pars.nFramesMax, pars.nFramesStart, pars.nTimeOut, pars.fDumpVabs = oldMax, oldStart, oldTimeout, oldDump
			gia.Gia_ManSetGateClasses(graph, nil)
			if classes := gia.Gia_ManObjClasses(graph); classes != nil {
				gia.Gia_ManSetGateClasses(graph, Gia_VtaConvertToGla(graph, classes))
			}
			gia.Gia_ManSetObjClasses(graph, nil)
			if ret == 0 || gia.Gia_ManGateClasses(graph) == nil {
				return ret
			}
		} else {
			classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(graph)))
			vec.Vec_IntWriteEntry(classes, 0, 1)
			vec.Vec_IntWriteEntry(classes, int(gia.Gia_ObjId(graph, gia.Gia_ObjFanin0(po))), 1)
			gia.Gia_ManSetGateClasses(graph, classes)
		}
	}
	p := Gla_ManStart(graph, pars)
	p.timeInit = time.Since(start)
	var deadline time.Time
	if pars.nTimeOut != 0 {
		deadline = time.Now().Add(time.Duration(pars.nTimeOut) * time.Second)
		p.pSat.SetDeadline(deadline)
	}
	if pars.fVerbose != 0 {
		fmt.Printf("Running gate-level abstraction: FrameMax = %d ConfMax = %d Timeout = %d RatioMin = %d%%.\n", pars.nFramesMax, pars.nConfLimit, pars.nTimeOut, pars.nRatioMin)
	}
	var witness *util.Abc_Cex_t
	var frame, i, previous, status int32
	var sent bool
	stopped := false
	for frame = 0; pars.nFramesMax == 0 || frame < pars.nFramesMax; frame++ {
		conflictStart := p.pSat.Statistics().Conflicts
		pars.iFrame = frame
		Gia_GlaAddTimeFrame(p, frame)
		var core *vec.Vec_Int_t
		var savedVars, conflicts int32
		for i = 0; ; i++ {
			clock := time.Now()
			core = Gla_ManUnsatCore(p, frame, p.pSat, pars.nConfLimit, pars.fVerbose, &status, &conflicts)
			if status == -1 || !deadline.IsZero() && time.Now().After(deadline) {
				stopped = true
				break
			}
			if status == 1 {
				p.timeUnsat += time.Since(clock)
				break
			}
			p.timeSat += time.Since(clock)
			p.nCexes++
			if abcmain.Abc_FrameIsBridgeMode() != 0 && sent {
				Gia_GlaSendCancel(p, pars.fVerbose)
				sent = false
			}
			clock = time.Now()
			var ppis *vec.Vec_Int_t
			if pars.fAddLayer != 0 {
				ppis = Gla_ManCollectPPis(p, nil)
				// No pseudo-inputs means another layer cannot make progress.
				if vec.Vec_IntSize(ppis) == 0 {
					ppis = Gla_ManRefinement(p)
				}
			} else {
				ppis = Gla_ManRefinement(p)
			}
			if ppis == nil {
				witness = p.pGia.TakeSequentialCounterexample()
				break
			}
			p.timeCex += time.Since(clock)
			if i == 0 {
				p.pSat.Bookmark()
				vec.Vec_IntClear(p.vAddedNew)
				p.nAbsOld = int32(vec.Vec_IntSize(p.vAbs))
				savedVars = p.nSatVars
			}
			Gia_GlaAddToAbs(p, ppis, 1)
			Gia_GlaAddOneSlice(p, frame, ppis)
			if pars.fVerbose != 0 {
				Gla_ManAbsPrintFrame(p, -1, frame+1, int32(p.pSat.Statistics().Conflicts-conflictStart), i, time.Since(start))
			}
		}
		if stopped || witness != nil {
			break
		}
		if status != 1 {
			panic("old GLA expected an UNSAT frame")
		}
		coreSize := int32(1)
		if core != nil {
			coreSize += int32(vec.Vec_IntSize(core))
			Gia_GlaAddToCounters(p, core)
		}
		if i == 0 {
			pars.nFramesNoChange++
		} else {
			pars.nFramesNoChange = 0
			p.pSat.Rollback()
			Gla_ManRollBack(p)
			p.nSatVars = savedVars
			Gia_GlaAddToAbs(p, core, 0)
			Gia_GlaAddOneSlice(p, frame, core)
			clock := time.Now()
			Gla_ManUnsatCore(p, frame, p.pSat, pars.nConfLimit, pars.fVerbose, &status, &conflicts)
			p.timeUnsat += time.Since(clock)
			if status == -1 {
				stopped = true
				break
			}
			if status == 0 {
				panic("old GLA proof core did not retain UNSAT")
			}
		}
		if pars.fVerbose != 0 {
			Gla_ManAbsPrintFrame(p, coreSize, frame+1, int32(p.pSat.Statistics().Conflicts-conflictStart), i, time.Since(start))
		}
		if frame > 2 && previous > 0 && i == 0 {
			if abcmain.Abc_FrameIsBridgeMode() != 0 {
				if sent {
					Gia_GlaSendCancel(p, pars.fVerbose)
				}
				Gia_GlaSendAbsracted(p, pars.fVerbose)
				sent = true
			}
			if pars.fDumpVabs != 0 {
				abcmain.Abc_FrameSetStatus(-1)
				abcmain.Abc_FrameSetCex(nil)
				abcmain.Abc_FrameSetNFrames(frame + 1)
				name := pars.pFileVabs
				if name == "" {
					name = "glabs.aig"
				}
				abcmain.Cmd_CommandExecute(abcmain.Abc_FrameGetGlobalFrame(), fmt.Sprintf("write_status %s", extra.Extra_FileNameGenericAppend(name, ".status")))
				Gia_GlaDumpAbsracted(p, pars.fVerbose)
			}
		}
		previous = i
		if Gia_GlaAbsCount(p, 0, 0) >= (p.nObjs-1)*(100-pars.nRatioMin)/100 {
			status = -1
			stopped = true
			break
		}
	}
	if witness == nil {
		gia.Gia_ManSetGateClasses(graph, Gla_ManTranslate(p))
		if status == -1 || stopped {
			fmt.Printf("Abstraction stopped in frame %d with a %d-stable abstraction.  ", frame, pars.nFramesNoChange)
		} else {
			pars.iFrame++
			fmt.Printf("GLA completed %d frames with a %d-stable abstraction.  ", frame, pars.nFramesNoChange)
		}
	} else {
		if err := graph.SetSequentialCounterexample(witness); err != nil {
			panic(err)
		}
		if gia.Gia_ManVerifyCex(graph, witness, 0) == 0 {
			fmt.Println("Gia_ManPerformGlaOld(): CEX verification has failed!")
		}
		fmt.Printf("Counter-example detected in frame %d.  ", frame)
		pars.iFrame = witness.IFrame - 1
		gia.Gia_ManSetGateClasses(graph, nil)
		ret = 0
	}
	fmt.Printf("Time: %v\n", time.Since(start))
	if pars.fVerbose != 0 {
		p.timeOther = time.Since(start) - p.timeUnsat - p.timeSat - p.timeCex - p.timeInit
		Gla_ManReportMemory(p)
	}
	Gla_ManStop(p)
	return ret
}
