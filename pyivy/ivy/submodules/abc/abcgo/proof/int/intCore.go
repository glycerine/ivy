package int

// Top-level interpolation loop from proof/int/intCore.c.
import (
	"context"
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/sat/bmc"
	"github.com/glycerine/abcgo/sat/cnf"
)

// rewriteForInterpolation applies DAR's original rwsat sequence under the
// interpolation deadline. A nil result signals a stopped or failed rewrite.
func rewriteForInterpolation(graph *aig.Graph, deadline int64) *aig.Graph {
	ctx := context.Background()
	cancel := func() {}
	if deadline != 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(deadline-util.Abc_Clock()))
	}
	defer cancel()
	result, err := dar.RewriteForSAT(ctx, graph, true, nil)
	if err != nil {
		return nil
	}
	return result
}

func Inter_ManSetDefaultParams(p *Inter_ManParams_t) {
	*p = Inter_ManParams_t{nFramesK: 1, fCheckKstep: 1, iFrameMax: -1}
}

// Inter_ManPerformInterpolation returns 1 for a proof, 0 for a failure,
// and -1 when a resource limit prevents a decision.
func Inter_ManPerformInterpolation(graph *aig.Graph, pars *Inter_ManParams_t, frame *int32) int32 {
	if graph == nil || pars == nil || frame == nil {
		panic("invalid interpolation arguments")
	}
	start := util.Abc_Clock()
	var deadline int64
	if pars.nSecLimit != 0 {
		deadline = start + int64(time.Duration(pars.nSecLimit)*time.Second)
	}
	if pars.nFramesK > 1 {
		pars.fTransLoop = 1
	}
	regs := int(graph.RegisterCount())
	if regs <= 0 || graph.InputCount() <= regs || graph.OutputCount()-regs-int(aig.Aig_ManConstrNum(graph)) != 1 {
		panic("invalid interpolation graph")
	}
	*frame = -1
	if Inter_ManCheckInitialState(graph) != 0 {
		fmt.Println("Property trivially fails in the initial state.")
		return 0
	}
	p := Inter_ManCreate(graph, pars)
	var check *Inter_Check_t
	finish := func(result, proved int32) int32 {
		p.timeTotal = util.Abc_Clock() - start
		Inter_ManStop(p, proved)
		Inter_CheckStop(check)
		return result
	}
	if pars.fTransLoop != 0 {
		p.pAigTrans = Inter_ManStartOneOutput(graph, 0)
	} else {
		p.pAigTrans = Inter_ManStartDuplicated(graph)
	}
	p.pCnfAig = cnf.Cnf_Derive(p.pAigTrans, p.pAigTrans.RegisterCount())
	p.nFrames = 1
	for {
		if pars.fUseBackward != 0 {
			p.pInter = Inter_ManStartOneOutput(graph, 1)
		} else {
			p.pInter = Inter_ManStartInitState(graph.RegisterCount())
		}
		p.pCnfInter = cnf.Cnf_Derive(p.pInter, 0)
		p.pFrames = Inter_ManFramesInter(graph, p.nFrames, pars.fUseBackward, pars.fUseTwoFrames)
		if pars.fRewrite != 0 {
			p.pFrames = rewriteForInterpolation(p.pFrames, deadline)
			if p.pFrames == nil {
				return finish(-1, 0)
			}
		}
		if pars.fUseBackward != 0 {
			p.pCnfFrames = cnf.Cnf_Derive(p.pFrames, int32(p.pFrames.OutputCount()))
		} else {
			p.pCnfFrames = cnf.Cnf_DeriveSimple(p.pFrames, 0)
		}
		if pars.fTransLoop == 0 && pars.fUseBackward == 0 && pars.nFramesK <= 1 {
			check = Inter_CheckStart(p.pAigTrans, pars.nFramesK)
			initial := cnf.Cnf_Derive(p.pInter, 1)
			Inter_CheckPerform(check, initial, deadline)
			cnf.Cnf_DataFree(initial)
			if p.vInters != nil {
				vec.Vec_PtrPush(p.vInters, aig.Aig_ManDupSimple(p.pInter))
			}
		}
		for i := int32(0); ; i++ {
			if pars.nFramesMax != 0 && p.nFrames+i >= pars.nFramesMax {
				return finish(-1, 0)
			}
			result := Inter_ManPerformOneStep(p, pars.fUseBias, pars.fUseBackward, deadline)
			pars.iFrameMax = i - 1 + p.nFrames
			if result == 0 {
				if i == 0 {
					*frame = p.nFrames
					var opts bmc.Saig_ParBmc_t
					bmc.Saig_ParBmcSetDefaultParams(&opts)
					opts.SetInterpolationOptions(p.nFrames, 100000000, pars.fVerbose)
					status := bmc.Saig_ManBmcScalable(graph, &opts)
					if status == 1 {
						fmt.Println("Error: The problem should be SAT but it is UNSAT.")
					} else if status == -1 {
						fmt.Println("Error: The problem timed out.")
					}
					return finish(0, 0)
				}
				p.nFrames += i
				Inter_ManClean(p)
				break
			}
			if result == -1 {
				return finish(-1, 0)
			}
			if result != 1 {
				panic("invalid interpolation result")
			}
			if p.pInterNew != nil {
				p.pInterNew = rewriteForInterpolation(p.pInterNew, deadline)
				if p.pInterNew == nil {
					return finish(-1, 1)
				}
			}
			if p.pInterNew == nil || p.pInterNew.Node(p.pInterNew.Outputs()[0]).Fanins[0] == p.pInterNew.Const0() {
				return finish(1, 1)
			}
			status := int32(0)
			if p.pInterNew.InputCount() == p.pInter.InputCount() {
				if pars.fCheckKstep == 0 {
					status = Inter_ManCheckContainment(p.pInterNew, p.pInter)
				} else if pars.fTransLoop != 0 || pars.fUseBackward != 0 || pars.nFramesK > 1 {
					steps := i + 1
					if steps > pars.nFramesK {
						steps = pars.nFramesK
					}
					status = Inter_ManCheckInductiveContainment(p.pAigTrans, p.pInterNew, steps, pars.fUseBackward)
				} else {
					data := cnf.Cnf_Derive(p.pInterNew, 1)
					status = Inter_CheckPerform(check, data, deadline)
					cnf.Cnf_DataFree(data)
					if p.vInters != nil {
						vec.Vec_PtrPush(p.vInters, aig.Aig_ManDupSimple(p.pInterNew))
					}
				}
			}
			if status == 1 {
				return finish(1, 1)
			}
			if status < 0 || deadline != 0 && util.Abc_Clock() > deadline {
				return finish(-1, 1)
			}
			if pars.fUseBackward != 0 && pars.fTransLoop == 0 {
				p.pInter = aig.Aig_ManCreateMiter(p.pInter, p.pInterNew, 2)
				p.pInter = rewriteForInterpolation(p.pInter, deadline)
				if p.pInter == nil {
					return finish(-1, 1)
				}
			} else {
				p.pInter = p.pInterNew
			}
			p.pInterNew = nil
			cnf.Cnf_DataFree(p.pCnfInter)
			p.pCnfInter = cnf.Cnf_Derive(p.pInter, 0)
		}
		Inter_CheckStop(check)
		check = nil
	}
}
