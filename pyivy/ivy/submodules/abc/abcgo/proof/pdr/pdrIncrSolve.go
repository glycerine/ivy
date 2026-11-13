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
	pars := p.pPars
	graph := p.pAig
	outputs := graph.TruePoCount()
	inputs := graph.TruePiCount()
	started := util.Abc_Clock()
	if pars.nTimeOut != 0 {
		p.timeToStop = started + int64(pars.nTimeOut)*int64(time.Second)
	} else {
		p.timeToStop = 0
	}

	outputEdges := graph.Outputs()[:outputs]
	if pars.fSolveAll != 0 {
		for i, output := range outputEdges {
			if aig.Aig_ObjChild0(graph, output) == graph.Const0() {
				vec.Vec_IntWriteEntry(pars.vOutMap, i, 1)
				pars.nProveOuts++
				if pars.fUseBridge != 0 {
					util.Gia_ManToBridgeResult(os.Stdout, 1, nil, int32(i))
				}
			}
		}
	}
	pars.timeLastSolved = util.Abc_Clock()
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
		if pars.fUseAbs != 0 && p.vAbsFlops == nil && frame >= 1 {
			regs := int(aig.Aig_ManRegNum(graph))
			p.vAbsFlops = vec.Vec_IntStart(regs)
			p.vMapFf2Ppi = vec.Vec_IntStartFull(regs)
			p.vMapPpi2Ff = vec.Vec_IntAlloc(100)
			IPdr_ManRestoreAbsFlops(p)
		}
	}
	for {
		refined := false
		if pars.fUseAbs != 0 && p.vAbsFlops == nil && frame == 1 {
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
				if pars.fSolveAll == 0 || pars.fUseBridge != 0 || pars.fStoreCex != 0 {
					witness = util.Abc_CexMakeTriv(aig.Aig_ManRegNum(graph), inputs, outputs, frame*outputs+int32(index))
				}
				if pars.fSolveAll == 0 {
					if err := graph.SetSequentialCounterexample(witness); err != nil {
						panic(err)
					}
					return 0
				}
				pars.nFailOuts++
				if pars.vOutMap != nil {
					vec.Vec_IntWriteEntry(pars.vOutMap, index, 0)
				}
				if pars.fNotVerbose == 0 {
					fmt.Printf("Output %d was trivially asserted in frame %d (solved %d out of %d outputs).\n", index, frame, pars.nFailOuts, outputs)
				}
				if pars.fUseBridge != 0 {
					util.Gia_ManToBridgeResult(os.Stdout, 0, witness, int32(index))
				}
				var stored any = witness
				if witness == nil {
					stored = pdrNoCounterexample{}
				}
				vec.Vec_PtrWriteEntry(p.vCexes, index, stored)
				if pars.pFuncOnFail != nil {
					callbackWitness := witness
					if pars.fStoreCex == 0 {
						callbackWitness = nil
					}
					if pars.pFuncOnFail(int32(index), callbackWitness) != 0 {
						pars.iFrame = frame
						return -1
					}
				}
				if pars.nFailOuts+pars.nDropOuts == outputs {
					if pars.nFailOuts != 0 {
						return 0
					}
					return -1
				}
				pars.timeLastSolved = util.Abc_Clock()
				continue
			}
			oneStarted := int64(0)
			if p.pTime4Outs != nil {
				oneStarted = util.Abc_Clock()
				p.timeToStopOne = oneStarted + p.pTime4Outs[index]
			}
			for {
				if pars.nTimeOutGap != 0 && pars.timeLastSolved != 0 && util.Abc_Clock() > pars.timeLastSolved+int64(pars.nTimeOutGap)*int64(time.Second) {
					pars.iFrame = frame
					return -1
				}
				var cube *Pdr_Set_t
				status := Pdr_ManCheckCube(p, frame, nil, &cube, pars.nConfLimit, 0, 1)
				if status == 1 {
					break
				}
				if status == 0 {
					status = Pdr_ManBlockCube(p, cube)
				}
				if status == -1 {
					if p.timeToStopOne != 0 && util.Abc_Clock() > p.timeToStopOne && (p.timeToStop == 0 || util.Abc_Clock() <= p.timeToStop) {
						Pdr_QueueClean(p)
						break
					}
					pars.iFrame = frame
					return -1
				}
				if status == 0 {
					pars.iFrame = frame
					if pars.fSolveAll == 0 {
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
					pars.nFailOuts++
					var witness *util.Abc_Cex_t
					if pars.fUseBridge != 0 || pars.fStoreCex != 0 {
						witness = Pdr_ManDeriveCex(p)
					}
					if pars.vOutMap != nil {
						vec.Vec_IntWriteEntry(pars.vOutMap, index, 0)
					}
					if pars.fUseBridge != 0 {
						util.Gia_ManToBridgeResult(os.Stdout, 0, witness, int32(index))
					}
					var stored any = witness
					if witness == nil {
						stored = pdrNoCounterexample{}
					}
					vec.Vec_PtrWriteEntry(p.vCexes, index, stored)
					if pars.pFuncOnFail != nil {
						callbackWitness := witness
						if pars.fStoreCex == 0 {
							callbackWitness = nil
						}
						if pars.pFuncOnFail(int32(index), callbackWitness) != 0 {
							pars.iFrame = frame
							return -1
						}
					}
					if pars.fNotVerbose == 0 {
						fmt.Printf("Output %d was asserted in frame %d (solved %d out of %d outputs).\n", index, frame, pars.nFailOuts, outputs)
					}
					if pars.nFailOuts == outputs {
						return 0
					}
					Pdr_QueueClean(p)
					break
				}
				if pars.fVerbose != 0 {
					Pdr_ManPrintProgress(p, 0, util.Abc_Clock()-started)
				}
			}
			if refined {
				break
			}
			if p.pTime4Outs != nil {
				elapsed := util.Abc_Clock() - oneStarted
				if p.pTime4Outs[index] > elapsed {
					p.pTime4Outs[index] -= elapsed
				} else {
					p.pTime4Outs[index] = 0
				}
				if p.pTime4Outs[index] == 0 && vec.Vec_PtrEntry(p.vCexes, index) == nil {
					pars.nDropOuts++
					if pars.vOutMap != nil {
						vec.Vec_IntWriteEntry(pars.vOutMap, index, -1)
					}
					if pars.fNotVerbose == 0 {
						fmt.Printf("Timing out on output %d in frame %d.\n", index, frame)
					}
				}
				p.timeToStopOne = 0
			}
		}
		if pars.fUseAbs != 0 && p.vAbsFlops != nil && !refined {
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
		if pars.fVerbose != 0 {
			var closeLine int32 = 1
			if refined {
				closeLine = 0
			}
			Pdr_ManPrintProgress(p, closeLine, util.Abc_Clock()-started)
		}
		if refined {
			continue
		}
		p.nQueLim = pars.nRestLimit
		Pdr_ManSetPropertyOutput(p, frame)
		frame++
		Pdr_ManCreateSolver(p, frame)
		status := Pdr_ManPushClauses(p)
		if status == -1 {
			pars.iFrame = frame
			return -1
		}
		if status != 0 {
			if pars.fVerbose != 0 {
				Pdr_ManPrintProgress(p, 1, util.Abc_Clock()-started)
			}
			if pars.fSilent == 0 {
				Pdr_ManReportInvariant(p)
				Pdr_ManVerifyInvariant(p)
			}
			pars.iFrame = frame
			pars.nProveOuts = outputs - pars.nFailOuts - pars.nDropOuts
			if pars.vOutMap != nil {
				for i := int32(0); i < outputs; i++ {
					if vec.Vec_IntEntry(pars.vOutMap, int(i)) == -2 {
						vec.Vec_IntWriteEntry(pars.vOutMap, int(i), 1)
						if pars.fUseBridge != 0 {
							util.Gia_ManToBridgeResult(os.Stdout, 1, nil, i)
						}
					}
				}
			}
			if pars.nProveOuts == outputs {
				return 1
			}
			if pars.nFailOuts > 0 {
				return 0
			}
			return -1
		}
		if pars.fVerbose != 0 {
			Pdr_ManPrintProgress(p, 0, util.Abc_Clock()-started)
		}
		now := util.Abc_Clock()
		if pars.pFuncStop != nil && pars.pFuncStop(pars.RunId) != 0 || p.timeToStop != 0 && now > p.timeToStop || pars.nTimeOutGap != 0 && pars.timeLastSolved != 0 && now > pars.timeLastSolved+int64(pars.nTimeOutGap)*int64(time.Second) || pars.nFrameMax != 0 && frame >= pars.nFrameMax {
			pars.iFrame = frame
			return -1
		}
	}
}

func IPdr_ManSolve(graph *aig.Graph, pars *Pdr_Par_t) int32 {
	if pars.fNewXSim != 0 && pars.fMonoCnf != 0 {
		return -1
	}
	started := util.Abc_Clock()
	outputs := graph.TruePoCount()
	if pars.nTimeOutOne != 0 && pars.fSolveAll == 0 {
		pars.nTimeOutOne = 0
	}
	if pars.nTimeOutOne != 0 && pars.nTimeOut == 0 {
		pars.nTimeOut = (pars.nTimeOutOne*outputs + 999) / 1000
	}
	if pars.fVerbose != 0 {
		fmt.Printf("VarMax = %d. FrameMax = %d. QueMax = %d. TimeMax = %d. MonoCNF = %t. SkipGen = %t. SolveAll = %t.\n", pars.nRecycle, pars.nFrameMax, pars.nRestLimit, pars.nTimeOut, pars.fMonoCnf != 0, pars.fSkipGeneral != 0, pars.fSolveAll != 0)
	}
	graph.ClearSequentialCounterexample()
	p := Pdr_ManStart(graph, pars, nil)
	var result int32
	for {
		result = IPdr_ManSolveInt(p, 1, 0)
		if result != -1 || pars.iFrame != pars.nFrameMax {
			break
		}
		saved := IPdr_ManSaveClauses(p, 1)
		Pdr_ManStop(p)
		p = Pdr_ManStart(graph, pars, nil)
		IPdr_ManRestoreClauses(p, saved, nil)
		pars.nFrameMax <<= 1
	}
	if p.vCexes != nil {
		if graph.SequentialCounterexamples() != nil {
			panic("pdr: graph already has per-output counterexamples")
		}
		graph.SetSequentialCounterexamples(p.vCexes)
		p.vCexes = nil
	}
	if pars.fDumpInv != 0 || result == 1 {
		var reduce int32
		if result != 1 {
			reduce = 1
		}
		abcmain.Abc_FrameSetInv(vec.Vec_IntArray(Pdr_ManDeriveInfinityClauses(p, reduce)))
		if pars.fDumpInv != 0 {
			filename := pars.pInvFileName
			if filename == "" {
				filename = aig.Aig_FileNameGenericAppend(graph.Name, "_inv.pla")
			}
			var proved int32
			if result == 1 {
				proved = 1
			}
			Pdr_ManDumpClauses(p, filename, proved)
			fmt.Printf("Dumped inductive invariant in file %q.\n", filename)
		}
	}
	p.tTotal += util.Abc_Clock() - started
	Pdr_ManStop(p)
	pars.iFrame--
	if pars.vOutMap != nil {
		for i := int32(0); i < outputs; i++ {
			if vec.Vec_IntEntry(pars.vOutMap, int(i)) == -2 {
				vec.Vec_IntWriteEntry(pars.vOutMap, int(i), -1)
			}
		}
	}
	if pars.fUseBridge != 0 {
		util.Gia_ManToBridgeAbort(os.Stdout, 7, []byte("timeout"))
	}
	return result
}
