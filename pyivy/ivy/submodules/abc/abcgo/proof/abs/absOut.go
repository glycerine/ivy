package abs

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/nwk"
)

// Gia_ManGlaRefine grows a gate abstraction by one layer or records a real CEX.
func Gia_ManGlaRefine(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, minCut, verbose int32) int32 {
	start := time.Now()
	classes := gia.Gia_ManGateClasses(p)
	if classes == nil {
		fmt.Println("Gia_ManGlaRefine(): Abstraction gate map is missing.")
		return -1
	}
	// Cleanup while duplicating can update classes; the original repeats it.
	abstract := Gia_ManDupAbsGates(p, classes)
	gia.Gia_ManStop(abstract)
	abstract = Gia_ManDupAbsGates(p, classes)
	defer gia.Gia_ManStop(abstract)
	if gia.Gia_ManPiNum(abstract) != cex.NPis {
		fmt.Println("Gia_ManGlaRefine(): The PI counts in GLA and in CEX do not match.")
		return -1
	}
	if gia.Gia_ManVerifyCex(abstract, cex, 0) == 0 {
		fmt.Println("Gia_ManGlaRefine(): The initial counter-example is invalid.")
	}
	var pis, ppis *vec.Vec_Int_t
	Gia_ManGlaCollect(p, classes, &pis, &ppis, nil, nil)
	if vec.Vec_IntSize(pis)+vec.Vec_IntSize(ppis) != int(gia.Gia_ManPiNum(abstract)) {
		panic("abstraction input count mismatch")
	}
	gia.Gia_ObjTerSimSet0(gia.Gia_ManConst0(abstract))
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		for i := int32(0); i < gia.Gia_ManPiNum(abstract); i++ {
			obj := gia.Gia_ManPi(abstract, i)
			if int(i) >= vec.Vec_IntSize(pis) {
				gia.Gia_ObjTerSimSetX(obj)
			} else if util.Abc_InfoHasBit(cex.PData, cex.NRegs+cex.NPis*frame+i) {
				gia.Gia_ObjTerSimSet1(obj)
			} else {
				gia.Gia_ObjTerSimSet0(obj)
			}
		}
		for i := int32(0); i < gia.Gia_ManRegNum(abstract); i++ {
			obj := gia.Gia_ManRo(abstract, i)
			if frame == 0 {
				gia.Gia_ObjTerSimSet0(obj)
			} else {
				gia.Gia_ObjTerSimRo(abstract, obj)
			}
		}
		for i := int32(0); i < gia.Gia_ManObjNum(abstract); i++ {
			obj := gia.Gia_ManObj(abstract, i)
			if gia.Gia_ObjIsAnd(obj) != 0 {
				gia.Gia_ObjTerSimAnd(obj)
			}
		}
		for i := int32(0); i < gia.Gia_ManCoNum(abstract); i++ {
			gia.Gia_ObjTerSimCo(gia.Gia_ManCo(abstract, i))
		}
	}
	var result *util.Abc_Cex_t
	if gia.Gia_ObjTerSimGet1(gia.Gia_ManPo(abstract, 0)) != 0 {
		result = Gia_ManCexRemap(p, cex, pis)
		if result != nil {
			fmt.Printf("Procedure &gla_refine found a real counter-example in frame %d.\n", result.IFrame)
		}
	}
	for i := int32(0); i < gia.Gia_ManObjNum(abstract); i++ {
		gia.Gia_ObjTerSimSetC(gia.Gia_ManObj(abstract, i))
	}
	if result != nil {
		if err := p.SetSequentialCounterexample(result); err != nil {
			panic(err)
		}
		return 0
	}
	for _, id := range vec.Vec_IntArray(ppis) {
		if vec.Vec_IntEntry(classes, int(id)) != 0 {
			panic("PPI is already in abstraction")
		}
		vec.Vec_IntWriteEntry(classes, int(id), 1)
	}
	if verbose != 0 {
		fmt.Printf("Additional objects = %d.  Time: %v\n", vec.Vec_IntSize(ppis), time.Since(start))
	}
	if minCut != 0 {
		nwk.Nwk_ManDeriveMinCut(p, verbose)
	}
	return -1
}

// Gia_ManNewRefine reruns GLA from the counterexample's selected starting state.
func Gia_ManNewRefine(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, frameStart, frameExtra, verbose int32) int32 {
	p.ClearSequentialCounterexample()
	classes := gia.Gia_ManGateClasses(p)
	if classes == nil {
		fmt.Println("Gia_ManNewRefine(): Abstraction gate map is missing.")
		return -1
	}
	saved := vec.Vec_IntDup(classes)
	fmt.Printf("Refining with %d-frame CEX, starting in frame %d, with %d extra frames.\n", cex.IFrame, frameStart, frameExtra)
	abstract := Gia_ManDupAbsGates(p, classes)
	gia.Gia_ManStop(abstract)
	abstract = Gia_ManDupAbsGates(p, classes)
	if gia.Gia_ManPiNum(abstract) != cex.NPis {
		fmt.Println("Gia_ManNewRefine(): The PI counts in GLA and in CEX do not match.")
		gia.Gia_ManStop(abstract)
		return -1
	}
	initial := Gia_ManGetStateAndCheckCex(abstract, cex, frameStart)
	if initial == nil {
		fmt.Println("Gia_ManNewRefine(): The initial counter-example is invalid.")
		gia.Gia_ManStop(abstract)
		return -1
	}
	if verbose != 0 {
		fmt.Println("Gia_ManNewRefine(): The initial counter-example is correct.")
	}
	var flops *vec.Vec_Int_t
	Gia_ManGlaCollect(p, classes, nil, nil, &flops, nil)
	gia.Gia_ManStop(abstract)
	next := Gia_ManTransformFlops(p, flops, initial)
	check := Gia_ManDupAbsGates(next, classes)
	Gia_ManCheckCex(check, cex, frameStart)
	gia.Gia_ManStop(check)
	if gia.Gia_ManGateClasses(next) != nil {
		panic("new manager already has abstraction classes")
	}
	gia.Gia_ManSetGateClasses(next, vec.Vec_IntDup(classes))
	var pars Abs_Par_t
	Abs_ParSetDefaults(&pars)
	pars.nFramesMax = cex.IFrame - frameStart + 1 + frameExtra
	pars.fVerbose = verbose
	if Gia_ManPerformGla(next, &pars) == 0 {
		gia.Gia_ManSetGateClasses(next, saved)
	}
	gia.Gia_ManSetGateClasses(p, gia.Gia_ManGateClasses(next))
	gia.Gia_ManSetGateClasses(next, nil)
	gia.Gia_ManStop(next)
	return -1
}
