package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/ivy"
)

// Abc_NtkBmc converts the network to Ivy, unrolls nFrames timeframes, and
// fraigs the resulting combinational network. This follows abcBmc.c; fVerbose
// is intentionally unused because the original report call is commented out.
func Abc_NtkBmc(pNtk *Abc_Ntk_t, nFrames, fInit, fVerbose int32) {
	pMan := Abc_NtkIvyBefore(pNtk, 0, 0)
	if pMan == nil {
		return
	}
	var mapping []*ivy.Ivy_Obj_t
	pFrames := ivy.Ivy_ManFrames(pMan, Abc_NtkLatchNum(pNtk), nFrames, fInit, &mapping)
	params := ivy.Ivy_FraigParams_t{}
	ivy.Ivy_FraigParamsDefault(&params)
	ivy.Ivy_FraigParamsSetBmc(&params, 1_000_000_000, 0, 0)
	pFraig := ivy.Ivy_FraigPerform(pFrames, &params)
	fmt.Printf("Frames have %6d nodes.  ", ivy.Ivy_ManNodeNum(pFrames))
	fmt.Printf("Fraig has %6d nodes.\n", ivy.Ivy_ManNodeNum(pFraig))
	ivy.Ivy_ManStop(pFraig)
	ivy.Ivy_ManStop(pFrames)
	ivy.Ivy_ManStop(pMan)
}

// Abc_NtkBmcReport reports per-frame node counts using the mapping produced
// by Ivy_ManFrames. It mirrors the (currently disabled) C reporting helper.
func Abc_NtkBmcReport(pMan, pFrames, pFraig *ivy.Ivy_Man_t, mapping []*ivy.Ivy_Obj_t, nFrames int32) {
	nIdMax := ivy.Ivy_ManObjIdMax(pMan)
	var prevStrash, prevFraig int32
	for frame := int32(0); frame < nFrames; frame++ {
		var firstStrash, firstFraig *ivy.Ivy_Obj_t
		for id := int32(0); id <= nIdMax; id++ {
			node := ivy.Ivy_ManObj(pMan, id)
			if node == nil {
				continue
			}
			if ivy.Ivy_ObjIsNode(node) == 0 {
				continue
			}
			index := int(frame*nIdMax + ivy.Ivy_ObjId(node))
			if index < 0 || index >= len(mapping) {
				continue
			}
			mapped := ivy.Ivy_Regular(mapping[index])
			if mapped == nil || ivy.Ivy_ObjIsConst1(mapped) != 0 || ivy.Ivy_ObjIsNone(mapped) != 0 {
				continue
			}
			equiv := ivy.Ivy_Regular(ivy.Ivy_ObjEquiv(mapped))
			if equiv == nil || ivy.Ivy_ObjIsConst1(equiv) != 0 || ivy.Ivy_ObjIsNone(equiv) != 0 {
				continue
			}
			firstStrash, firstFraig = mapped, equiv
			break
		}
		if firstStrash == nil || firstFraig == nil {
			continue
		}
		if frame != 0 {
			fmt.Printf("Frame %3d :  Strash = %5d  Fraig = %5d\n", frame,
				ivy.Ivy_ObjId(firstStrash)-prevStrash, ivy.Ivy_ObjId(firstFraig)-prevFraig)
		}
		prevStrash, prevFraig = ivy.Ivy_ObjId(firstStrash), ivy.Ivy_ObjId(firstFraig)
	}
}
