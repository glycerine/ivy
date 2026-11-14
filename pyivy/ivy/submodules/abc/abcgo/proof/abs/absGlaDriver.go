package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

// Gia_ManPerformGla performs gate abstraction, bounded checking and refinement.
// Return 0 for a verified counterexample, 1 for a proved abstraction and -1
// when the configured bounds or resources are exhausted.
func Gia_ManPerformGla(g *gia.Gia_Man_t, pars *Abs_Par_t) int32 {
	if g == nil || pars == nil || gia.Gia_ManPoNum(g) != 1 {
		panic("abs: GLA requires one output")
	}
	if pars.NTimeOut != 0 {
		defer absPinClockThread()()
	}
	if glaOldHasSpecialGates(g) {
		return glaPerformLowered(g, pars, false, func(lowered *gia.Gia_Man_t) int32 {
			return Gia_ManPerformGla(lowered, pars)
		})
	}
	start := time.Now()
	pars.IFrame = -1
	g.ClearSequentialCounterexample()
	po := gia.Gia_ManPo(g, 0)
	driver := gia.Gia_ObjFanin0(po)
	invert := gia.Gia_ObjFaninC0(po)
	if gia.Gia_ObjIsConst0(driver) != 0 {
		if invert == 0 {
			fmt.Println("Sequential miter is trivially UNSAT.")
			return 1
		}
		c := util.Abc_CexAlloc(gia.Gia_ManRegNum(g), gia.Gia_ManPiNum(g), 1)
		c.IFrame = 0
		if err := g.SetSequentialCounterexample(c); err != nil {
			panic(err)
		}
		fmt.Println("Sequential miter is trivially SAT.")
		return 0
	}
	// A directly connected PI has no cut to mark, and its first-frame witness
	// can be constructed without invoking the abstraction manager.
	if gia.Gia_ObjIsPi(g, driver) != 0 {
		c := util.Abc_CexAlloc(gia.Gia_ManRegNum(g), gia.Gia_ManPiNum(g), 1)
		c.IFrame = 0
		if invert == 0 {
			util.Abc_InfoSetBit(c.PData, c.NRegs+gia.Gia_ObjCioId(driver))
		}
		if err := g.SetSequentialCounterexample(c); err != nil {
			panic(err)
		}
		return 0
	}
	if gia.Gia_ManGateClasses(g) == nil {
		classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
		vec.Vec_IntWriteEntry(classes, 0, 1)
		vec.Vec_IntWriteEntry(classes, int(gia.Gia_ObjId(g, driver)), 1)
		gia.Gia_ManSetGateClasses(g, classes)
	}
	p := Ga2_ManStart(g, pars)
	p.timeInit = time.Since(start).Nanoseconds()
	result := int32(-1)
	sent := false
	proving := int32(-1)
	stopped := false
	defer func() {
		if proving >= 0 {
			Gia_GlaProveCancel(pars.FVerbose)
		}
		Ga2_ManStop(p)
	}()
	if pars.FVerbose != 0 {
		fmt.Printf("Running gate-level abstraction: FrameMax = %d ConfMax = %d Timeout = %d\n", pars.NFramesMax, pars.NConfLimit, pars.NTimeOut)
		if pars.FDumpMabs != 0 {
			Ga2_ManWriteStatus(p, -1)
			Ga2_ManDumpInitialMap(p)
		}
	}
	for !stopped {
		if p.cpuDeadline != 0 && absCPUClock() > p.cpuDeadline {
			break
		}
		pars.IFrame = -1
		Ga2_ManRestart(p)
		oldAbs := vec.Vec_IntSize(p.vAbs)
		restart := false
		for frame := int32(0); pars.NFramesMax == 0 || frame < pars.NFramesMax; frame++ {
			if p.cpuDeadline != 0 && absCPUClock() > p.cpuDeadline {
				stopped = true
				break
			}
			conflictsBegin := p.pSat.Statistics().Conflicts
			nAbs := int32(vec.Vec_IntSize(p.vAbs))
			nValues := int32(vec.Vec_IntSize(p.vValues))
			pars.IFrame = frame
			if int(frame) == len(p.vId2Lit) {
				p.vId2Lit = append(p.vId2Lit, vec.Vec_IntAlloc(0))
			}
			vec.Vec_IntFillExtra(Ga2_MapFrameMap(p, frame), vec.Vec_IntSize(p.vValues), -1)
			Ga2_ManAddAbsClauses(p, frame)
			if pars.FUseSkip != 0 && frame <= pars.iFrameProved || pars.NFramesStart != 0 && frame <= pars.NFramesStart {
				continue
			}
			lit := Ga2_ObjFindLit(p, driver, frame)
			if lit < 0 {
				panic("abs: missing GLA output literal")
			}
			lit ^= invert
			if lit == 0 {
				continue
			}
			if lit == 1 {
				panic("abs: GLA target literal must exceed one")
			}

			for p.pSat.NumVars() < p.nSatVars {
				p.pSat.NewVariable()
			}
			oldVars := p.nSatVars
			refinements := int32(0)
			for {
				// The C driver stops for either polarity once the target variable
				// is assigned at the root, even when the literal is true.
				if p.pSat.RootAssigned(lit >> 1) {
					break
				}
				p.pSat.SetLimits(int64(pars.NConfLimit), 0)
				began := time.Now()
				status := p.pSat.Solve(bsat.Literal(lit))
				if status == bsat.Satisfiable {
					p.nCexes++
					p.timeSat += time.Since(began).Nanoseconds()
					if sent && abcmain.Abc_FrameIsBridgeMode() != 0 {
						Gia_Ga2SendCancel(p, pars.FVerbose)
						sent = false
					}
					if proving >= 0 {
						Gia_GlaProveCancel(pars.FVerbose)
						proving = -1
					}
					began = time.Now()
					Rnm_ManSetRefId(p.pRnm, refinements)
					add := Ga2_ManRefine(p)
					p.timeCex += time.Since(began).Nanoseconds()
					if add == nil {
						c := g.SequentialCounterexample()
						if c == nil || gia.Gia_ManVerifyCex(g, c, 0) == 0 {
							g.ClearSequentialCounterexample()
							stopped = true
							break
						}
						result = 0
						stopped = true
						break
					}
					if vec.Vec_IntSize(add) == 0 {
						var abstractCex *util.Abc_Cex_t
						var mapping *vec.Vec_Int_t
						Ga2_GlaPrepareCexAndMap(p, &abstractCex, &mapping)
						actual := Ga2_ManDeriveCex(p, mapping)
						if gia.Gia_ManVerifyCex(g, actual, 0) != 0 {
							if err := g.SetSequentialCounterexample(actual); err != nil {
								panic(err)
							}
							result = 0
						}
						stopped = true
						break
					}
					if refinements == 0 {
						if !p.pSat.Simplify() {
							stopped = true
							break
						}
						p.pSat.Bookmark()
					}
					Ga2_ManAddToAbs(p, add)
					refinements++
					if pars.NRatioMin2 != 0 && int32(vec.Vec_IntSize(p.vAbs)) >= p.nMarked*pars.NRatioMin2/100 {
						stopped = true
						break
					}
					continue
				}
				p.timeUnsat += time.Since(began).Nanoseconds()
				if status == bsat.Unknown {
					stopped = true
					break
				}
				if refinements == 0 {
					if frame > pars.iFrameProved {
						pars.nFramesNoChange++
					}
					break
				}
				if frame > pars.iFrameProved {
					pars.nFramesNoChange = 0
				}
				coreIDs := p.pSat.ProofCore()
				p.pSat.Rollback()
				Ga2_ManShrinkAbs(p, nAbs, nValues, oldVars)
				core := vec.Vec_IntAlloc(len(coreIDs))
				for _, id := range coreIDs {
					vec.Vec_IntPush(core, id)
				}
				if !p.pSat.Simplify() {
					stopped = true
					break
				}
				p.pSat.Bookmark()
				Ga2_ManAddToAbs(p, core)
				status = p.pSat.Solve(bsat.Literal(lit))
				if status == bsat.Unknown {
					stopped = true
					break
				}
				if status != bsat.Unsatisfiable {
					panic("abs: GLA core does not preserve UNSAT")
				}
				coreIDs = p.pSat.ProofCore()
				p.pSat.Rollback()
				Ga2_ManShrinkAbs(p, nAbs, nValues, oldVars)
				core = vec.Vec_IntAlloc(len(coreIDs))
				for _, id := range coreIDs {
					vec.Vec_IntPush(core, id)
				}
				Ga2_ManAddToAbs(p, core)
				break
			}
			if stopped {
				break
			}
			if pars.iFrameProved < frame {
				pars.iFrameProved = frame
			}
			if pars.FVerbose != 0 {
				Ga2_ManAbsPrintFrame(p, frame, int32(p.pSat.Statistics().Conflicts-conflictsBegin), refinements, time.Since(start).Nanoseconds(), 1)
			}
			if Gia_GlaProveCheck(pars.FVerbose) != 0 {
				result = 1
				stopped = true
				break
			}
			if refinements > 0 {
				gia.Gia_ManSetGateClasses(g, Ga2_ManAbsTranslate(p))
				if pars.NRatioMin != 0 && int32(vec.Vec_IntSize(p.vAbs)) >= p.nMarked*pars.NRatioMin/100 {
					stopped = true
					break
				}
			}
			if pars.nFramesNoChange == pars.NFramesNoChangeLim {
				if pars.FDumpVabs != 0 || pars.FDumpMabs != 0 {
					Ga2_ManWriteStatus(p, frame)
					Ga2_GlaDumpAbsracted(p, pars.FVerbose)
				}
				if pars.FCallProver != 0 {
					if proving >= 0 {
						Gia_GlaProveCancel(pars.FVerbose)
					}
					Gia_GlaProveAbsracted(g, pars.FSimpProver, pars.FVeryVerbose)
					proving = frame
					p.nPdrCalls++
				}
				if abcmain.Abc_FrameIsBridgeMode() != 0 {
					if sent {
						Gia_Ga2SendCancel(p, pars.FVerbose)
					}
					Gia_Ga2SendAbsracted(p, pars.FVerbose)
					sent = true
				}
			}
			if pars.NRatioMax != 0 && refinements > 0 && (frame > 20 || vec.Vec_IntSize(p.vAbs) > 100) && vec.Vec_IntSize(p.vAbs)-oldAbs >= oldAbs*int(pars.NRatioMax)/100 {
				restart = true
				break
			}
		}
		if !restart {
			break
		}
	}
	if result == 0 {
		pars.IFrame--
		gia.Gia_ManSetGateClasses(g, nil)
	} else if result < 0 {
		gia.Gia_ManSetGateClasses(g, Ga2_ManAbsTranslate(p))
		pars.IFrame = pars.iFrameProved
	}
	return result
}
