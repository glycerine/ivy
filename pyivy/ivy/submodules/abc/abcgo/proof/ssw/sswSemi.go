package ssw

// Semi-formal filtering, ported from sswSemi.c. The original manager start
// intentionally asserts because its frame-storage migration was unfinished.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

type Ssw_Sem_t struct {
	nConfMaxStart, nConfMax, nFramesSweep, fVerbose int32
	pMan                                            *Ssw_Man_t
	vTargets                                        *vec.Vec_Ptr_t
	nPatternsAlloc, nPatterns                       int32
	vPatterns                                       *vec.Vec_Ptr_t
	vHistory                                        *vec.Vec_Int_t
}

func Ssw_SemManStart(pMan *Ssw_Man_t, nConfMax, verbose int32) *Ssw_Sem_t {
	// The C implementation ends in assert(0) before installing the required
	// enlarged frame arrays. Preserve this failure rather than running it unsafely.
	panic("ssw: semi-formal manager frame storage is unfinished in ABC")
}
func Ssw_SemManStop(p *Ssw_Sem_t) {
	if p != nil {
		*p = Ssw_Sem_t{}
	}
}
func Ssw_SemCheckTargets(p *Ssw_Sem_t) int32 {
	for _, item := range vec.Vec_PtrArray(p.vTargets) {
		if Ssw_ObjIsConst1Cand(p.pMan.pAig, item.(aig.Edge)) == 0 {
			return 1
		}
	}
	return 0
}
func Ssw_ManFilterBmcSavePattern(p *Ssw_Sem_t) {
	if p.nPatterns >= p.nPatternsAlloc {
		return
	}
	g := p.pMan.pAig
	inputs := g.Inputs()
	regs := int(g.RegisterCount())
	pis := len(inputs) - regs
	for i := 0; i < regs; i++ {
		if sswPatternBit(p.pMan.pPatWords, pis+i) != 0 {
			util.Abc_InfoSetBit(vec.Vec_PtrEntry(p.vPatterns, i).([]uint32), p.nPatterns)
		}
	}
	p.nPatterns++
}
func Ssw_ManFilterBmc(semi *Ssw_Sem_t, pattern, checkTargets int32) int32 {
	start := time.Now()
	p := semi.pMan
	p.pFrames = aig.NewGraph()
	p.pMSat.pAig = p.pFrames
	inputs, outputs := p.pAig.Inputs(), p.pAig.Outputs()
	regs := int(p.pAig.RegisterCount())
	pis, pos := len(inputs)-regs, len(outputs)-regs
	for i, lo := range inputs[pis:] {
		value := sswPatternBit(vec.Vec_PtrEntry(semi.vPatterns, i).([]uint32), int(pattern))
		mapped := p.pFrames.Const0()
		if value != 0 {
			mapped = p.pFrames.Const1()
		}
		Ssw_ObjSetFrame(p, lo, 0, mapped)
	}
	result := semi.nFramesSweep
	first := false
	for f := int32(0); f < semi.nFramesSweep; f++ {
		Ssw_ObjSetFrame(p, p.pAig.Const1(), f, p.pFrames.Const1())
		for _, pi := range inputs[:pis] {
			Ssw_ObjSetFrame(p, pi, f, p.pFrames.NewInput())
		}
		for _, obj := range sswFrameNodes(p.pAig) {
			Ssw_ObjSetFrame(p, obj, f, sswMakeNode(p, p.pFrames, obj, f))
			if p.pAig.Node(obj).Kind != aig.Buffer && Ssw_ManSweepNode(p, obj, f, 1, nil) != 0 {
				Ssw_ManFilterBmcSavePattern(semi)
				if !first {
					first = true
					semi.nConfMax *= 10
				}
			}
			if f > 0 && p.pMSat.pSat.Statistics().Conflicts >= int64(semi.nConfMax) {
				result = -1
				break
			}
		}
		if p.pMSat.pSat.Statistics().Conflicts >= int64(semi.nConfMax) {
			result += f + 1
			break
		}
		if checkTargets != 0 && Ssw_SemCheckTargets(semi) != 0 {
			break
		}
		for i, lo := range inputs[pis:] {
			mapped := Ssw_ObjChild0Fra(p, outputs[pos+i], f)
			Ssw_ObjSetFrame(p, lo, f+1, mapped)
			Ssw_CnfNodeAddToSolver(p.pMSat, mapped&^1)
		}
	}
	if first {
		semi.nConfMax /= 10
	}
	Ssw_ClassesCheck(p.ppClasses)
	p.timeBmc += time.Since(start)
	return result
}
func Ssw_FilterUsingSemi(manager *Ssw_Man_t, checkTargets, nConfMax, verbose int32) int32 {
	p := Ssw_SemManStart(manager, nConfMax, verbose)
	if checkTargets != 0 && Ssw_SemCheckTargets(p) != 0 {
		panic("ssw: target already failed")
	}
	if verbose != 0 {
		fmt.Printf("AIG : C = %6d. Cl = %6d. Nodes = %6d. ConfMax = %6d. FramesMax = %6d.\n", Ssw_ClassesCand1Num(manager.ppClasses), Ssw_ClassesClassNum(manager.ppClasses), aig.Aig_ManNodeNum(manager.pAig), p.nConfMax, p.nFramesSweep)
	}
	var result int32
	for i := int32(0); i < p.nPatterns; i++ {
		manager.pMSat = Ssw_SatStart(0)
		frames := Ssw_ManFilterBmc(p, i, checkTargets)
		if verbose != 0 {
			fmt.Printf("%3d : C = %6d. Cl = %6d. F = %3d. P = %3d.\n", i, Ssw_ClassesCand1Num(manager.ppClasses), Ssw_ClassesClassNum(manager.ppClasses), frames, p.nPatterns)
		}
		Ssw_ManCleanup(manager)
		if checkTargets != 0 && Ssw_SemCheckTargets(p) != 0 {
			fmt.Println("Target is hit!!!")
			result = 1
		}
		if p.nPatterns >= p.nPatternsAlloc {
			break
		}
	}
	Ssw_SemManStop(p)
	manager.nStrangers = 0
	manager.nSatCalls = 0
	manager.nSatProof = 0
	manager.nSatFailsReal = 0
	manager.nSatCallsUnsat = 0
	manager.nSatCallsSat = 0
	manager.timeSimSat = 0
	manager.timeSat = 0
	manager.timeSatSat = 0
	manager.timeSatUnsat = 0
	manager.timeSatUndec = 0
	return result
}
