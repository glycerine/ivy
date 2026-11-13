package fra

// Sequential equivalence orchestration from fraSec.c.
import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/aig/saig"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/bdd/bbr"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/opt/dar"
	interp "github.com/glycerine/abcgo/proof/int"
	"github.com/glycerine/abcgo/proof/pdr"
	"github.com/glycerine/abcgo/proof/ssw"
)

var secUnsolvedNumber atomic.Int32

func secSaveUnsolved(graph *aig.Graph) {
	abcmain.Abc_FrameSetSave1(aig.Aig_ManDupSimple(graph))
}

func Fra_SecSetDefaultParams(p *Fra_Sec_t) {
	*p = Fra_Sec_t{fTryComb: 1, fTryBmc: 1, nFramesMax: 4, nBTLimit: 1000, nBTLimitGlobal: 5000000, nBTLimitInter: 10000, nBddVarsMax: 150, nBddMax: 50000, nBddIterMax: 1000000, fRetimeFirst: 1, fRetimeRegs: 1, fFraiging: 1, fInduction: 1, fInterpolation: 1, fReachability: 1, fReorderImage: 1, fStopOnFirstFail: 1, fUsePdr: 1, nPdrTimeout: 60}
}

// secTransferWitness copies only a counterexample that replays on the original
// graph. Transformations that alter inputs or state may invalidate a trace.
func secTransferWitness(original, reduced *aig.Graph) bool {
	pis := original.InputCount() - int(original.RegisterCount())
	if pis != reduced.InputCount()-int(reduced.RegisterCount()) {
		return false
	}
	cex := reduced.SequentialCounterexample()
	if cex == nil && reduced.RegisterCount() == 0 {
		model, ok := reduced.Data.([]int32)
		if !ok || len(model) != pis {
			return false
		}
		copy := util.Abc_CexAlloc(original.RegisterCount(), int32(pis), 1)
		for i, value := range model {
			if value != 0 {
				util.Abc_InfoSetBit(copy.PData, original.RegisterCount()+int32(i))
			}
		}
		output := saig.Saig_ManFindFailedPoCex(original, copy)
		if output < 0 {
			return false
		}
		copy.IPo = output
		return original.SetSequentialCounterexample(copy) == nil
	}
	if cex == nil {
		return false
	}
	return secTransferCex(original, cex)
}

// Replaying the input trace on the original also recovers the original PO
// index after output compaction or single-output interpolation.
func secTransferCex(original *aig.Graph, cex *util.Abc_Cex_t) bool {
	if cex == nil || cex.NPis != int32(original.InputCount())-original.RegisterCount() {
		return false
	}
	copy := util.Abc_CexDup(cex, original.RegisterCount())
	if copy == nil {
		return false
	}
	output := saig.Saig_ManFindFailedPoCex(original, copy)
	if output < 0 {
		return false
	}
	copy.IPo = output
	return original.SetSequentialCounterexample(copy) == nil
}
func secTrivialWitness(g *aig.Graph) bool {
	inputs := int32(g.InputCount()) - g.RegisterCount()
	for allOne := 0; allOne < 2; allOne++ {
		cex := util.Abc_CexAlloc(g.RegisterCount(), inputs, 1)
		if allOne != 0 {
			for i := int32(0); i < inputs; i++ {
				util.Abc_InfoSetBit(cex.PData, g.RegisterCount()+i)
			}
		}
		if output := saig.Saig_ManFindFailedPoCex(g, cex); output >= 0 {
			cex.IPo = output
			if g.SetSequentialCounterexample(cex) == nil {
				return true
			}
		}
	}
	return false
}

// Fra_FraigSec returns 1 for equivalent, 0 for a verified counterexample,
// and -1 when the configured procedures cannot decide the miter.
func Fra_FraigSec(original *aig.Graph, pars *Fra_Sec_t, result **aig.Graph) int32 {
	if original == nil || pars == nil {
		panic("invalid sequential equivalence arguments")
	}
	pars.nSMnumber = -1
	graph := aig.Aig_ManDupSimple(original)
	status := Fra_FraigMiterStatus(graph)
	timedOut := false
	finish := func() int32 {
		if status == -1 && timedOut && original.SequentialCounterexample() != nil &&
			secTransferCex(original, original.SequentialCounterexample()) {
			status = 0
		}
		if status == 0 {
			verified := secTransferWitness(original, graph)
			if !verified && original.SequentialCounterexample() != nil {
				verified = secTransferCex(original, original.SequentialCounterexample())
			}
			if !verified && !secTrivialWitness(original) {
				status = -1
			}
		}
		if status == -1 {
			secSaveUnsolved(graph)
		}
		if pars.fReportSolution != 0 && pars.fRecursive == 0 {
			switch status {
			case 1:
				fmt.Println("SOLUTION: PASS")
			case 0:
				fmt.Println("SOLUTION: FAIL")
			default:
				fmt.Println("SOLUTION: UNDECIDED")
			}
		}
		if status == -1 && !timedOut && pars.fSilent == 0 {
			number := secUnsolvedNumber.Add(1)
			pars.nSMnumber = number
			name := fmt.Sprintf("sm%02d.aig", number)
			if err := ioa.Ioa_WriteAiger(graph, name, 0, 0); err != nil {
				fmt.Printf("Could not write unsolved reduced miter %q: %v\n", name, err)
			} else {
				fmt.Printf("The unsolved reduced miter is written into file %q.\n", name)
			}
		}
		if result != nil {
			*result = aig.Aig_ManDupSimpleDfs(graph)
		}
		if pars.fSilent == 0 {
			switch status {
			case 1:
				fmt.Println("Networks are equivalent.")
			case 0:
				fmt.Println("Networks are NOT EQUIVALENT.")
			default:
				fmt.Println("Networks are UNDECIDED.")
			}
		}
		return status
	}
	if status >= 0 {
		return finish()
	}
	if graph.RegisterCount() > 0 {
		graph = aig.Aig_ManReduceLaches(graph, 0)
	}
	if graph.RegisterCount() > 0 {
		graph = aig.Aig_ManConstReduce(graph, 0, -1, -1, 0, 0)
	}
	status = Fra_FraigMiterStatus(graph)
	if status >= 0 {
		return finish()
	}
	if pars.fPhaseAbstract != 0 {
		aig.Aig_ManSetRegNum(graph, graph.RegisterCount())
		graph = saig.Saig_ManPhaseAbstractAuto(graph, 0)
	}
	if pars.fRetimeFirst != 0 && graph.RegisterCount() > 0 {
		graph = saig.Saig_ManRetimeForward(graph, 100, 0)
	}
	var sw ssw.Ssw_Pars_t
	if graph.RegisterCount() > 0 {
		graph = aig.Aig_ManDupOrdered(graph)
		ssw.Ssw_ManSetDefaultParamsLcorr(&sw)
		previous := graph
		graph = ssw.Ssw_LatchCorrespondence(previous, &sw)
		if previous.SequentialCounterexample() != nil {
			secTransferWitness(original, previous)
		}
		if graph == nil {
			graph = previous
			if secTransferWitness(original, previous) {
				status = 0
			} else {
				status = -1
				timedOut = true
			}
			return finish()
		}
	}
	ssw.Ssw_ManSetDefaultParams(&sw)
	if pars.fFraiging != 0 {
		graph = Fra_FraigEquivence(graph, 100, 0)
	}
	if graph.RegisterCount() == 0 {
		Fra_FraigCec(&graph, 100000, 0)
	}
	status = Fra_FraigMiterStatus(graph)
	if status >= 0 {
		return finish()
	}
	if pars.fRetimeRegs != 0 && graph.RegisterCount() > 0 {
		aig.Aig_ManSetRegNum(graph, graph.RegisterCount())
		graph = saig.Saig_ManRetimeMinArea(graph, 1000, 0, 0, 1, 0)
		graph = aig.Aig_ManDupOrdered(graph)
	}
	status = Fra_FraigMiterStatus(graph)
	if status == -1 && pars.fInduction != 0 {
		for frames := int32(1); frames <= pars.nFramesMax && frames > 0; frames *= 2 {
			ssw.Ssw_ParsSetFramesK(&sw, frames)
			ssw.Ssw_ParsSetBTLimit(&sw, pars.nBTLimit)
			ssw.Ssw_ParsSetBTLimitGlobal(&sw, pars.nBTLimitGlobal)
			if ssw.Ssw_ParsConflicts(&sw) > pars.nBTLimitGlobal {
				timedOut = true
				break
			}
			previous := graph
			if graph.RegisterCount() > 0 {
				graph = ssw.Ssw_SignalCorrespondence(graph, &sw)
			} else {
				graph = aig.Aig_ManDupSimpleDfs(graph)
			}
			if graph == nil {
				graph = previous
				if secTransferWitness(original, previous) {
					status = 0
					return finish()
				}
				timedOut = true
				break
			}
			status = Fra_FraigMiterStatus(graph)
			if status != -1 {
				break
			}
			if graph.RegisterCount() > 0 {
				aig.Aig_ManSetRegNum(graph, graph.RegisterCount())
				graph = saig.Saig_ManRetimeMinArea(graph, 1000, 0, 0, 1, 0)
				graph = aig.Aig_ManDupOrdered(graph)
			}
			if graph.RegisterCount() > 0 {
				graph = aig.Aig_ManConstReduce(graph, 0, -1, -1, 0, 0)
			}
			var err error
			graph, err = dar.Dar_ManCompress2(aig.Aig_ManDupOrdered(graph), true, false, true, false, false)
			if err != nil || graph == nil {
				graph = previous
				timedOut = true
				break
			}
			if graph.RegisterCount() > 0 {
				sim := Fra_SmlSimulateSeq(graph, 0, 128*frames, 1+16/(1+aig.Aig_ManNodeNum(graph)/1000), 1)
				if sim.NonconstantOutput {
					if cex := Fra_SmlGetCounterExample(sim); cex != nil {
						_ = graph.SetSequentialCounterexample(cex)
					}
					status = 0
					return finish()
				}
			}
		}
	}
	if timedOut {
		status = -1
		return finish()
	}
	status = Fra_FraigMiterStatus(graph)
	if pars.fInterpolation != 0 && status == -1 && graph.RegisterCount() > 0 {
		var options interp.Inter_ManParams_t
		interp.Inter_ManSetDefaultParams(&options)
		options.NBTLimit = pars.nBTLimitInter
		options.FVerbose = pars.fVeryVerbose
		var depth int32
		outputs := graph.OutputCount() - int(graph.RegisterCount())
		if outputs == 1 {
			status = interp.Inter_ManPerformInterpolation(graph, &options, &depth)
		} else if pars.fInterSeparate != 0 {
			remaining := 0
			status = 1
			for i, out := range graph.Outputs()[:outputs] {
				if graph.Node(out).Fanins[0].NodeID() == 0 {
					continue
				}
				single := aig.Aig_ManDupOneOutput(graph, int32(i), 1)
				single = aig.Aig_ManScl(single, 1, 1, 0, -1, -1, 0, 0)
				if single.RegisterCount() == 0 {
					remaining++
					continue
				}
				one := interp.Inter_ManPerformInterpolation(single, &options, &depth)
				if cex := single.SequentialCounterexample(); cex != nil {
					secTransferCex(original, cex)
					cex.IPo = int32(i)
					_ = graph.SetSequentialCounterexample(cex)
					status = 0
					break
				}
				if one == 1 {
					aig.Aig_ObjPatchFanin0(graph, out, graph.Const0())
				} else {
					remaining++
				}
			}
			if status != 0 && remaining > 0 {
				status = -1
			}
			if timedOut {
				return finish()
			}
			if status != 0 {
				graph = aig.Aig_ManDupUnsolvedOutputs(graph, 1)
				graph = aig.Aig_ManScl(graph, 1, 1, 0, -1, -1, 0, 0)
			}
		} else {
			merged := saig.Saig_ManDupOrpos(graph)
			status = interp.Inter_ManPerformInterpolation(merged, &options, &depth)
			if cex := merged.SequentialCounterexample(); cex != nil {
				cex.IPo = saig.Saig_ManFindFailedPoCex(graph, cex)
				_ = graph.SetSequentialCounterexample(cex)
			}
		}
	}
	if pars.fReachability != 0 && status == -1 && graph.RegisterCount() > 0 && graph.RegisterCount() < pars.nBddVarsMax {
		var options bbr.Saig_ParBbr_t
		bbr.Bbr_ManSetDefaultParams(&options)
		options.NBddMax = pars.nBddMax
		options.NIterMax = pars.nBddIterMax
		options.FPartition = 1
		options.FReorder = 1
		options.FReorderImage = 1
		options.FSilent = pars.fSilent
		aig.Aig_ManSetRegNum(graph, graph.RegisterCount())
		status = bbr.Aig_ManVerifyUsingBdds(graph, &options)
	}
	if pars.fUsePdr != 0 && status == -1 && graph.RegisterCount() > 0 {
		options := pdr.DefaultOptions()
		options.Timeout = time.Duration(pars.nPdrTimeout) * time.Second
		options.Verbose = pars.fVerbose != 0
		answer, err := pdr.Solve(graph, &options)
		if err == nil {
			status = int32(answer.Status)
		}
	}
	return finish()
}
