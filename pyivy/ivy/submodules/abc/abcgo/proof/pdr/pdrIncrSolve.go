// Incremental PDR solve drivers, ported from pdrIncr.c.
package pdr

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"time"
)

func IPdr_ManSolveInt(p *Pdr_Man_t, checkClauses, pushClauses int32) int32 {
	defer pdrStartTimedClock(p)()
	pars := p.pPars
	graph := p.pAig
	outputs := graph.TruePoCount()
	inputs := graph.TruePiCount()
	started := util.Abc_Clock()
	if pars.NTimeOut != 0 {
		p.timeToStop = pdrNow(p) + int64(pars.NTimeOut)*int64(time.Second)
	} else {
		p.timeToStop = 0
	}

	outputEdges := graph.Outputs()[:outputs]
	if pars.FSolveAll != 0 {
		for i, output := range outputEdges {
			if aig.Aig_ObjChild0(graph, output) == graph.Const0() {
				vec.Vec_IntWriteEntry(pars.VOutMap, i, 1)
				pars.NProveOuts++
				if pars.FUseBridge != 0 {
					util.Gia_ManToBridgeResult(os.Stdout, 1, nil, int32(i))
				}
			}
		}
	}
	pars.timeLastSolved = pdrNow(p)
	frame := int32(0)
	if vec.Vec_VecSize(p.vClauses) == 0 {
		Pdr_ManCreateSolver(p, frame)
	} else {
		frame = int32(vec.Vec_VecSize(p.vClauses) - 1)
		if checkClauses != 0 {
			IPdr_ManCheckClauses(p)
		}
		if pushClauses != 0 {
			p.iUseFrame = max(frame, 1)
			status := Pdr_ManPushClauses(p)
			if status != 0 {
				Pdr_ManReportInvariant(p)
				Pdr_ManVerifyInvariant(p)
				return 1
			}
		}
		if pars.FUseAbs != 0 && p.vAbsFlops == nil && frame >= 1 {
			regs := int(aig.Aig_ManRegNum(graph))
			p.vAbsFlops = vec.Vec_IntStart(regs)
			p.vMapFf2Ppi = vec.Vec_IntStartFull(regs)
			p.vMapPpi2Ff = vec.Vec_IntAlloc(100)
			IPdr_ManRestoreAbsFlops(p)
		}
	}
	for {
		refined := false
		if pars.FUseAbs != 0 && p.vAbsFlops == nil && frame == 1 {
			regs := int(aig.Aig_ManRegNum(graph))
			p.vAbsFlops = vec.Vec_IntStart(regs)
			p.vMapFf2Ppi = vec.Vec_IntStartFull(regs)
			p.vMapPpi2Ff = vec.Vec_IntAlloc(100)
		}
		p.nFrames = frame
		p.iUseFrame = max(frame, 1)
		for index, output := range outputEdges {
			p.iOutCur = int32(index)
			if p.vCexes != nil && vec.Vec_PtrEntry(p.vCexes, index) != nil {
				continue
			}
			if p.pTime4Outs != nil && p.pTime4Outs[index] == 0 {
				continue
			}
			driver := aig.Aig_ObjChild0(graph, output)
			if driver == graph.Const0() {
				continue
			}
			if driver == graph.Const1() {
				var witness *util.Abc_Cex_t
				if pars.FSolveAll == 0 || pars.FUseBridge != 0 || pars.FStoreCex != 0 {
					witness = util.Abc_CexMakeTriv(aig.Aig_ManRegNum(graph), inputs, outputs, frame*outputs+int32(index))
				}
				if pars.FSolveAll == 0 {
					if err := graph.SetSequentialCounterexample(witness); err != nil {
						panic(err)
					}
					return 0
				}
				pars.NFailOuts++
				if pars.VOutMap != nil {
					vec.Vec_IntWriteEntry(pars.VOutMap, index, 0)
				}
				if pars.FNotVerbose == 0 {
					fmt.Printf("Output %d was trivially asserted in frame %d (solved %d out of %d outputs).\n", index, frame, pars.NFailOuts, outputs)
				}
				if pars.FUseBridge != 0 {
					util.Gia_ManToBridgeResult(os.Stdout, 0, witness, int32(index))
				}
				var stored any = witness
				if witness == nil {
					stored = pdrNoCounterexample{}
				}
				vec.Vec_PtrWriteEntry(p.vCexes, index, stored)
				if pars.pFuncOnFail != nil {
					callbackWitness := witness
					if pars.FStoreCex == 0 {
						callbackWitness = nil
					}
					if pars.pFuncOnFail(int32(index), callbackWitness) != 0 {
						pars.IFrame = frame
						return -1
					}
				}
				if pars.NFailOuts+pars.NDropOuts == outputs {
					if pars.NFailOuts != 0 {
						return 0
					}
					return -1
				}
				pars.timeLastSolved = pdrNow(p)
				continue
			}
			oneStarted := int64(0)
			if p.pTime4Outs != nil {
				oneStarted = pdrNow(p)
				p.timeToStopOne = oneStarted + p.pTime4Outs[index]
			}
			for {
				if pars.NTimeOutGap != 0 && pars.timeLastSolved != 0 && pdrNow(p) > pars.timeLastSolved+int64(pars.NTimeOutGap)*int64(time.Second) {
					pars.IFrame = frame
					return -1
				}
				var cube *Pdr_Set_t
				status := Pdr_ManCheckCube(p, frame, nil, &cube, pars.NConfLimit, 0, 1)
				if status == 1 {
					break
				}
				if status == 0 {
					status = Pdr_ManBlockCube(p, cube)
				}
				if status == -1 {
					if pdrOnlyOutputTimedOut(p, pdrNow(p)) {
						Pdr_QueueClean(p)
						break
					}
					pars.IFrame = frame
					return -1
				}
				if status == 0 {
					pars.IFrame = frame
					if pars.FSolveAll == 0 {
						before := util.Abc_Clock()
						witness := Pdr_ManDeriveCexAbs(p)
						p.tAbs += util.Abc_Clock() - before
						if witness == nil {
							Pdr_QueueClean(p)
							refined = true
							break
						}
						if err := graph.SetSequentialCounterexample(witness); err != nil {
							panic(err)
						}
						return 0
					}
					pars.NFailOuts++
					var witness *util.Abc_Cex_t
					if pars.FUseBridge != 0 || pars.FStoreCex != 0 {
						witness = Pdr_ManDeriveCex(p)
					}
					if pars.VOutMap != nil {
						vec.Vec_IntWriteEntry(pars.VOutMap, index, 0)
					}
					if pars.FUseBridge != 0 {
						util.Gia_ManToBridgeResult(os.Stdout, 0, witness, int32(index))
					}
					var stored any = witness
					if witness == nil {
						stored = pdrNoCounterexample{}
					}
					vec.Vec_PtrWriteEntry(p.vCexes, index, stored)
					if pars.pFuncOnFail != nil {
						callbackWitness := witness
						if pars.FStoreCex == 0 {
							callbackWitness = nil
						}
						if pars.pFuncOnFail(int32(index), callbackWitness) != 0 {
							pars.IFrame = frame
							return -1
						}
					}
					if pars.FNotVerbose == 0 {
						fmt.Printf("Output %d was asserted in frame %d (solved %d out of %d outputs).\n", index, frame, pars.NFailOuts, outputs)
					}
					if pars.NFailOuts == outputs {
						return 0
					}
					Pdr_QueueClean(p)
					break
				}
				if pars.FVerbose != 0 {
					Pdr_ManPrintProgress(p, 0, util.Abc_Clock()-started)
				}
			}
			if refined {
				break
			}
			if p.pTime4Outs != nil {
				elapsed := pdrNow(p) - oneStarted
				if p.pTime4Outs[index] > elapsed {
					p.pTime4Outs[index] -= elapsed
				} else {
					p.pTime4Outs[index] = 0
				}
				if p.pTime4Outs[index] == 0 && vec.Vec_PtrEntry(p.vCexes, index) == nil {
					pars.NDropOuts++
					if pars.VOutMap != nil {
						vec.Vec_IntWriteEntry(pars.VOutMap, index, -1)
					}
					if pars.FNotVerbose == 0 {
						fmt.Printf("Timing out on output %d in frame %d.\n", index, frame)
					}
				}
				p.timeToStopOne = 0
			}
		}
		if pars.FUseAbs != 0 && p.vAbsFlops != nil && !refined {
			vec.Vec_IntFill(p.vAbsFlops, vec.Vec_IntSize(p.vAbsFlops), 0)
			for k := 0; k < vec.Vec_VecSize(p.vClauses); k++ {
				for _, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(p.vClauses, k)) {
					cube := item.(*Pdr_Set_t)
					for _, lit := range cube.Lits[:cube.nLits] {
						vec.Vec_IntWriteEntry(p.vAbsFlops, int(lit>>1), 1)
					}
				}
			}
		}
		if pars.FVerbose != 0 {
			var closeLine int32 = 1
			if refined {
				closeLine = 0
			}
			Pdr_ManPrintProgress(p, closeLine, util.Abc_Clock()-started)
		}
		if refined {
			continue
		}
		p.nQueLim = pars.NRestLimit
		Pdr_ManSetPropertyOutput(p, frame)
		frame++
		Pdr_ManCreateSolver(p, frame)
		status := Pdr_ManPushClauses(p)
		if status == -1 {
			pars.IFrame = frame
			return -1
		}
		if status != 0 {
			if pars.FVerbose != 0 {
				Pdr_ManPrintProgress(p, 1, util.Abc_Clock()-started)
			}
			if pars.FSilent == 0 {
				Pdr_ManReportInvariant(p)
				Pdr_ManVerifyInvariant(p)
			}
			pars.IFrame = frame
			pars.NProveOuts = outputs - pars.NFailOuts - pars.NDropOuts
			if pars.VOutMap != nil {
				for i := int32(0); i < outputs; i++ {
					if vec.Vec_IntEntry(pars.VOutMap, int(i)) == -2 {
						vec.Vec_IntWriteEntry(pars.VOutMap, int(i), 1)
						if pars.FUseBridge != 0 {
							util.Gia_ManToBridgeResult(os.Stdout, 1, nil, i)
						}
					}
				}
			}
			if pars.NProveOuts == outputs {
				return 1
			}
			if pars.NFailOuts > 0 {
				return 0
			}
			return -1
		}
		if pars.FVerbose != 0 {
			Pdr_ManPrintProgress(p, 0, util.Abc_Clock()-started)
		}
		now := pdrNow(p)
		if pars.pFuncStop != nil && pars.pFuncStop(pars.RunId) != 0 || p.timeToStop != 0 && now > p.timeToStop || pars.NTimeOutGap != 0 && pars.timeLastSolved != 0 && now > pars.timeLastSolved+int64(pars.NTimeOutGap)*int64(time.Second) || pars.NFrameMax != 0 && frame >= pars.NFrameMax {
			pars.IFrame = frame
			return -1
		}
	}
}

func IPdr_ManSolve(graph *aig.Graph, pars *Pdr_Par_t) int32 {
	if pars.FNewXSim != 0 && pars.FMonoCnf != 0 {
		return -1
	}
	started := util.Abc_Clock()
	outputs := graph.TruePoCount()
	if pars.NTimeOutOne != 0 && pars.FSolveAll == 0 {
		pars.NTimeOutOne = 0
	}
	if pars.NTimeOutOne != 0 && pars.NTimeOut == 0 {
		pars.NTimeOut = (pars.NTimeOutOne*outputs + 999) / 1000
	}
	if pars.FVerbose != 0 {
		fmt.Printf("VarMax = %d. FrameMax = %d. QueMax = %d. TimeMax = %d. MonoCNF = %t. SkipGen = %t. SolveAll = %t.\n", pars.NRecycle, pars.NFrameMax, pars.NRestLimit, pars.NTimeOut, pars.FMonoCnf != 0, pars.FSkipGeneral != 0, pars.FSolveAll != 0)
	}
	graph.ClearSequentialCounterexample()
	p := Pdr_ManStart(graph, pars, nil)
	var result int32
	for {
		result = IPdr_ManSolveInt(p, 1, 0)
		if result != -1 || pars.IFrame != pars.NFrameMax {
			break
		}
		saved := IPdr_ManSaveClauses(p, 1)
		Pdr_ManStop(p)
		p = Pdr_ManStart(graph, pars, nil)
		IPdr_ManRestoreClauses(p, saved, nil)
		pars.NFrameMax <<= 1
	}
	if p.vCexes != nil {
		if graph.SequentialCounterexamples() != nil {
			panic("pdr: graph already has per-output counterexamples")
		}
		graph.SetSequentialCounterexamples(p.vCexes)
		p.vCexes = nil
	}
	if pars.FDumpInv != 0 || result == 1 {
		var reduce int32
		if result != 1 {
			reduce = 1
		}
		abcmain.Abc_FrameSetInv(vec.Vec_IntArray(Pdr_ManDeriveInfinityClauses(p, reduce)))
		if pars.FDumpInv != 0 {
			filename := pars.PInvFileName
			if filename == "" {
				filename = aig.Aig_FileNameGenericAppend(graph.Name, "_inv.pla")
			}
			var proved int32
			if result == 1 {
				proved = 1
			}
			Pdr_ManDumpClauses(p, filename, proved)
		}
	}
	p.tTotal += util.Abc_Clock() - started
	Pdr_ManStop(p)
	pars.IFrame--
	if pars.VOutMap != nil {
		for i := int32(0); i < outputs; i++ {
			if vec.Vec_IntEntry(pars.VOutMap, int(i)) == -2 {
				vec.Vec_IntWriteEntry(pars.VOutMap, int(i), -1)
			}
		}
	}
	if pars.FUseBridge != 0 {
		util.Gia_ManToBridgeAbort(os.Stdout, 7, []byte("timeout"))
	}
	return result
}
