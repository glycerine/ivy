package fret

import (
	hop "github.com/glycerine/abcgo/aig/hop"
	abc "github.com/glycerine/abcgo/base/abc"
	mio "github.com/glycerine/abcgo/map/mio"
	mem "github.com/glycerine/abcgo/misc/mem"
	vec "github.com/glycerine/abcgo/misc/vec"
)

type Abc_Aig_t = abc.Abc_Aig_t
type Abc_Ntk_t = abc.Abc_Ntk_t
type Abc_Obj_t = abc.Abc_Obj_t
type Abc_ObjType_t = abc.Abc_ObjType_t
type Hop_Man_t = hop.Hop_Man_t
type Hop_Obj_t = hop.Hop_Obj_t
type Mem_Flex_t = mem.Mem_Flex_t
type Mio_Gate_t = mio.Mio_Gate_t
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t

// Flow-based retiming state shared by the retiming implementation files.
// The C header used anonymous unions and bit-fields. Separate pointer fields
// make those alternatives explicit in Go; the mark bits remain in one word.

const (
	MAX_DIST       int32 = 30000
	VISITED_E      int32 = 0x001
	VISITED_R      int32 = 0x002
	VISITED        int32 = VISITED_E | VISITED_R
	FLOW           int32 = 0x004
	CROSS_BOUNDARY int32 = 0x008
	BLOCK          int32 = 0x010
	INIT_0         int32 = 0x020
	INIT_1         int32 = 0x040
	INIT_CARE      int32 = INIT_0 | INIT_1
	CONSERVATIVE   int32 = 0x080
	BLOCK_OR_CONS  int32 = BLOCK | CONSERVATIVE
	BIAS_NODE      int32 = 0x100
)

type Flow_Data_t struct {
	mark                  int32
	pred, pInitObj, pCopy *Abc_Obj_t
	vNodes                *Vec_Ptr_t
	e_dist, r_dist        int32
}

type NodeLag_t struct {
	id, lag int32
}

type InitConstraint_t struct {
	pBiasNode *Abc_Obj_t
	vNodes    *Vec_Int_t
	vLags     *Vec_Int_t
}

type MinRegMan_t struct {
	maxDelay, fComputeInitState, fGuaranteeInitState, fBlockConst int32
	nNodes, nLatches, fForwardOnly, fBackwardOnly                 int32
	fConservTimingOnly, nMaxIters, fVerbose                       int32
	pNtk                                                          *Abc_Ntk_t
	nPreRefine                                                    int32
	fIsForward, fSinkDistTerminate                                int32
	nExactConstraints, nConservConstraints                        int32
	fSolutionIsDc, constraintMask, iteration, subIteration        int32
	vLags, vSinkDistHist                                          *Vec_Int_t
	pDataArray                                                    []Flow_Data_t
	vTimeEdges                                                    []Vec_Ptr_t
	vExactNodes, vInitConstraints, vNodes                         *Vec_Ptr_t
	pInitNtk                                                      *Abc_Ntk_t
	pInitToOrig                                                   []NodeLag_t
}

func flowData(obj *Abc_Obj_t) *Flow_Data_t {
	return &pManMR.pDataArray[Abc_ObjId(obj)]
}

func ensureInitToOrig(id int32) {
	if id < 0 {
		panic("negative init-network object ID")
	}
	need := int(id) + 1
	if need <= len(pManMR.pInitToOrig) {
		return
	}
	nextLen := max(need, max(10, 2*len(pManMR.pInitToOrig)))
	next := make([]NodeLag_t, nextLen)
	copy(next, pManMR.pInitToOrig)
	for i := len(pManMR.pInitToOrig); i < len(next); i++ {
		next[i].id = -1
	}
	pManMR.pInitToOrig = next
}

func FSET(obj *Abc_Obj_t, mask int32)        { flowData(obj).mark |= mask }
func FUNSET(obj *Abc_Obj_t, mask int32)      { flowData(obj).mark &^= mask }
func FTEST(obj *Abc_Obj_t, mask int32) int32 { return flowData(obj).mark & mask }
func FSETPRED(obj, pred *Abc_Obj_t) {
	if abc.Abc_ObjIsLatch(obj) != 0 {
		panic("FSETPRED: latch data is reserved for init-state mapping")
	}
	flowData(obj).pred = pred
}
func FGETPRED(obj *Abc_Obj_t) *Abc_Obj_t { return flowData(obj).pred }
func FTIMEEDGES(obj *Abc_Obj_t) *Vec_Ptr_t {
	return &pManMR.vTimeEdges[Abc_ObjId(obj)]
}
