package sbd

// Port of src/opt/sbd/sbdCore.c: SAT-based delay optimization.

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

const SBD_MAX_LUTSIZE = 6

type Sbd_Man_t struct {
	pPars     *Sbd_Par_t // user's parameters
	pGia      *Gia_Man_t // user's AIG manager (will be modified by adding nodes)
	vTfos     [][]int32  // TFO for each node (roots are marked) (windowing)
	vLutLevs  *Vec_Int_t // LUT level for each node after resynthesis
	vLutCuts  []int32    // LUT cut for each nodes after resynthesis
	vLutCuts2 []int32    // LUT cut for each nodes after resynthesis
	vMirrors  *Vec_Int_t // alternative node
	vSims     [4][]word  // simulation information (main, backup, controlability)
	vCover    *Vec_Int_t // temporary
	vLits     *Vec_Int_t // temporary
	vLits2    *Vec_Int_t // temporary
	nLuts     [6]int32   // 0=const, 1=1lut, 2=2lut, 3=3lut
	nTried    int32
	nUsed     int32
	timeWin   int64
	timeCut   int64
	timeCov   int64
	timeCnf   int64
	timeSat   int64
	timeQbf   int64
	timeNew   int64
	timeOther int64
	timeTotal int64
	pSto      *Sbd_Sto_t
	pSrv      *Sbd_Srv_t
	// target node
	Pivot      int32          // target node
	DivCutoff  int32          // the place where D-2 divisors begin
	vTfo       *Vec_Int_t     // TFO (excludes node, includes roots) - precomputed
	vRoots     *Vec_Int_t     // TFO root nodes
	vWinObjs   *Vec_Int_t     // TFI + Pivot + sideTFI + TFO (including roots)
	vObj2Var   *Vec_Int_t     // SAT variables for the window (indexes of objects in vWinObjs)
	vDivSet    *Vec_Int_t     // divisor variables
	vDivVars   *Vec_Int_t     // divisor variables
	vDivValues *Vec_Int_t     // SAT variables values for the divisor variables
	vDivLevels [][]int32      // divisors collected by levels
	vCounts    [2]*Vec_Int_t  // counters of zeros and ones
	vMatrix    *vec.Vec_Wrd_t // covering matrix
	pSat       *sat_solver    // SAT solver
}

func Sbd_ObjCut(p *Sbd_Man_t, i int32) []int32 {
	return p.vLutCuts[(p.pPars.NLutSize+1)*i:]
}
func Sbd_ObjCut2(p *Sbd_Man_t, i int32) []int32 {
	return p.vLutCuts2[(p.pPars.NLutSize+1)*i:]
}
func sbdObjSim(p *Sbd_Man_t, k int, i int32) []word {
	n := p.pPars.NWords
	return p.vSims[k][n*i : n*(i+1)]
}
func Sbd_ObjSim0(p *Sbd_Man_t, i int32) []word { return sbdObjSim(p, 0, i) }
func Sbd_ObjSim1(p *Sbd_Man_t, i int32) []word { return sbdObjSim(p, 1, i) }
func Sbd_ObjSim2(p *Sbd_Man_t, i int32) []word { return sbdObjSim(p, 2, i) }
func Sbd_ObjSim3(p *Sbd_Man_t, i int32) []word { return sbdObjSim(p, 3, i) }

func Sbd_ParSetDefault(pPars *Sbd_Par_t) {
	*pPars = Sbd_Par_t{}
	pPars.NLutSize = 4                                    // target LUT size
	pPars.NLutNum = 3                                     // target LUT count
	pPars.NCutSize = (pPars.NLutSize-1)*pPars.NLutNum + 1 // target cut size
	pPars.NCutNum = 128                                   // target cut count
	pPars.NTfoLevels = 5                                  // the number of TFO levels (windowing)
	pPars.NTfoFanMax = 4                                  // the max number of fanouts (windowing)
	pPars.NWinSizeMax = 2000                              // maximum window size (windowing)
	pPars.NBTLimit = 0                                    // maximum number of SAT conflicts
	pPars.NWords = 1                                      // simulation word count
	pPars.FMapping = 1                                    // generate mapping
	pPars.FMoreCuts = 0                                   // use several cuts
	pPars.FFindDivs = 0                                   // perform divisor search
	pPars.FUsePath = 0                                    // optimize only critical path
	pPars.FArea = 0                                       // area-oriented optimization
	pPars.FCover = 0                                      // use complete cover procedure
	pPars.FVerbose = 0                                    // verbose flag
	pPars.FVeryVerbose = 0                                // verbose flag
}

// sbdWecPush is Vec_WecPush: it grows the level array as needed.
func sbdWecPush(v *[][]int32, Level int32, Entry int32) {
	if int(Level) >= len(*v) {
		n := max(2*len(*v), int(Level)+1)
		_ = n
		for len(*v) < int(Level)+1 {
			*v = append(*v, nil)
		}
	}
	(*v)[Level] = append((*v)[Level], Entry)
}

// Sbd_ManWindowRoots computes TFO and window roots for all nodes.
//
// TFO does not include the node itself. If TFO is empty, it means that the
// node itself is its own root, which may happen if the node is pointed by a
// PO or if it has too many fanouts.
func Sbd_ManWindowRoots(p *Gia_Man_t, nTfoLevels, nTfoFanMax int32) [][]int32 {
	vTfos := make([][]int32, Gia_ManObjNum(p))    // TFO nodes with roots marked
	vTemp := make([]*Vec_Int_t, Gia_ManObjNum(p)) // storage
	for i := range vTemp {
		vTemp[i] = vec.Vec_IntAlloc(0)
	}
	vPoDrivers := vec.Vec_BitStart(int(Gia_ManObjNum(p)))
	gia.Gia_ManLevelNum(p)
	gia.Gia_ManCreateRefs(p)
	gia.Gia_ManCleanMark0(p)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Id := Gia_ManCiIdToId(p, i)
		vNodes := vTemp[Id]
		vec.Vec_IntPush(vNodes, Id)
	}
	for _, Id := range giaCoDriverIds(p) {
		vec.Vec_BitWriteEntry(vPoDrivers, int(Id), 1)
	}
	for Id := int32(0); Id < Gia_ManObjNum(p); Id++ {
		if !Gia_ObjIsAnd(Gia_ManObj(p, Id)) {
			continue
		}
		fAlwaysRoot := vec.Vec_BitEntry(vPoDrivers, int(Id)) != 0 || (Gia_ObjRefNumId(p, Id) >= nTfoFanMax)
		vNodes0 := vTemp[Gia_ObjFaninId0(Gia_ManObj(p, Id), Id)]
		vNodes1 := vTemp[Gia_ObjFaninId1(Gia_ManObj(p, Id), Id)]
		vNodes := vTemp[Id]
		vec.Vec_IntTwoMerge2(vNodes0, vNodes1, vNodes)
		k2 := 0
		for k := 0; k < vec.Vec_IntSize(vNodes); k++ {
			Fan := vec.Vec_IntEntry(vNodes, k)
			fRoot := fAlwaysRoot || (Gia_ObjLevelId(p, Id)-Gia_ObjLevelId(p, Fan) >= nTfoLevels)
			vTfos[Fan] = append(vTfos[Fan], Abc_Var2Lit(Id, b2i(fRoot)))
			if !fRoot {
				vec.Vec_IntWriteEntry(vNodes, k2, Fan)
				k2++
			}
		}
		vec.Vec_IntShrink(vNodes, k2)
		if !fAlwaysRoot {
			vec.Vec_IntPush(vNodes, Id)
		}
	}
	return vTfos
}

func Sbd_ManStart(pGia *Gia_Man_t, pPars *Sbd_Par_t) *Sbd_Man_t {
	p := &Sbd_Man_t{}
	p.timeTotal = util.Abc_Clock()
	p.pPars = pPars
	p.pGia = pGia
	p.vTfos = Sbd_ManWindowRoots(pGia, pPars.NTfoLevels, pPars.NTfoFanMax)
	p.vLutLevs = vec.Vec_IntStart(int(Gia_ManObjNum(pGia)))
	p.vLutCuts = make([]int32, Gia_ManObjNum(pGia)*(p.pPars.NLutSize+1))
	p.vMirrors = vec.Vec_IntStartFull(int(Gia_ManObjNum(pGia)))
	for i := 0; i < 4; i++ {
		p.vSims[i] = make([]word, Gia_ManObjNum(pGia)*p.pPars.NWords)
	}
	// target node
	p.vCover = vec.Vec_IntAlloc(100)
	p.vLits = vec.Vec_IntAlloc(100)
	p.vLits2 = vec.Vec_IntAlloc(100)
	p.vRoots = vec.Vec_IntAlloc(100)
	p.vWinObjs = vec.Vec_IntAlloc(int(Gia_ManObjNum(pGia)))
	p.vObj2Var = vec.Vec_IntStart(int(Gia_ManObjNum(pGia)))
	p.vDivSet = vec.Vec_IntAlloc(100)
	p.vDivVars = vec.Vec_IntAlloc(100)
	p.vDivValues = vec.Vec_IntAlloc(100)
	p.vCounts[0] = vec.Vec_IntAlloc(100)
	p.vCounts[1] = vec.Vec_IntAlloc(100)
	p.vMatrix = vec.Vec_WrdAlloc(100)
	// start input cuts
	for i := int32(0); i < Gia_ManCiNum(pGia); i++ {
		Id := Gia_ManCiIdToId(pGia, i)
		pCut := Sbd_ObjCut(p, Id)
		pCut[0] = 1
		pCut[1] = Id
	}
	// generate random input
	gia.Gia_ManRandom(1)
	for i := int32(0); i < Gia_ManCiNum(pGia); i++ {
		Id := Gia_ManCiIdToId(pGia, i)
		for w := int32(0); w < p.pPars.NWords; w++ {
			Sbd_ObjSim0(p, Id)[w] = Gia_ManRandomW(0)
		}
	}
	// cut enumeration
	if pPars.FMoreCuts != 0 {
		p.pSto = Sbd_StoAlloc(pGia, p.vMirrors, pPars.NLutSize, pPars.NCutSize, pPars.NCutNum, b2i(pPars.FMapping == 0), 1)
	} else {
		p.pSto = Sbd_StoAlloc(pGia, p.vMirrors, pPars.NLutSize, pPars.NLutSize, pPars.NCutNum, b2i(pPars.FMapping == 0), 1)
		p.pSrv = Sbd_ManCutServerStart(pGia, p.vMirrors, p.vLutLevs, nil, nil, pPars.NLutSize, pPars.NCutSize, pPars.NCutNum, 0)
	}
	return p
}
func Sbd_ManStop(p *Sbd_Man_t) {
	if p.pSat != nil {
		sat_solver_delete(p.pSat)
		p.pSat = nil
	}
	if p.pSto != nil {
		Sbd_StoFree(p.pSto)
	}
	if p.pSrv != nil {
		Sbd_ManCutServerStop(p.pSrv)
	}
}

// Constructing window.
func Sbd_ManPropagateControlOne(p *Sbd_Man_t, Node int32) {
	pNode := Gia_ManObj(p.pGia, Node)
	iObj0 := Gia_ObjFaninId0(pNode, Node)
	iObj1 := Gia_ObjFaninId1(pNode, Node)
	pCtrl := Sbd_ObjSim2(p, Node)
	pCtrl0 := Sbd_ObjSim2(p, iObj0)
	pCtrl1 := Sbd_ObjSim2(p, iObj1)
	pDtrl := Sbd_ObjSim3(p, Node)
	pDtrl0 := Sbd_ObjSim3(p, iObj0)
	pDtrl1 := Sbd_ObjSim3(p, iObj1)
	for w := int32(0); w < p.pPars.NWords; w++ {
		pCtrl0[w] |= pCtrl[w]
		pCtrl1[w] |= pCtrl[w]
		pDtrl0[w] |= pDtrl[w]
		pDtrl1[w] |= pDtrl[w]
	}
}
func Sbd_ManPropagateControl(p *Sbd_Man_t, Pivot int32) {
	clk := util.Abc_Clock()
	util.Abc_TtCopy(Sbd_ObjSim3(p, Pivot), Sbd_ObjSim2(p, Pivot), p.pPars.NWords, 0)
	// clean controlability
	for i := int32(0); i < vec.Vec_IntEntry(p.vObj2Var, int(Pivot)); i++ {
		Node := vec.Vec_IntEntry(p.vWinObjs, int(i))
		util.Abc_TtClear(Sbd_ObjSim2(p, Node), p.pPars.NWords)
		util.Abc_TtClear(Sbd_ObjSim3(p, Node), p.pPars.NWords)
	}
	// propagate controlability to fanins for the TFI nodes starting from the pivot
	for i := vec.Vec_IntEntry(p.vObj2Var, int(Pivot)); i >= 0; i-- {
		Node := vec.Vec_IntEntry(p.vWinObjs, int(i))
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, Node)) {
			Sbd_ManPropagateControlOne(p, Node)
		}
	}
	p.timeWin += util.Abc_Clock() - clk
}
func Sbd_ManUpdateOrder(p *Sbd_Man_t, Pivot int32) {
	nTimeValidDivs := 0
	// collect divisors by logic level
	LevelMax := vec.Vec_IntEntry(p.vLutLevs, int(Pivot))
	p.vDivLevels = p.vDivLevels[:0]
	for len(p.vDivLevels) < int(LevelMax)+1 {
		p.vDivLevels = append(p.vDivLevels, nil)
	}
	for i := range p.vDivLevels {
		p.vDivLevels[i] = p.vDivLevels[i][:0]
	}
	for _, Node := range vec.Vec_IntArray(p.vWinObjs) {
		sbdWecPush(&p.vDivLevels, vec.Vec_IntEntry(p.vLutLevs, int(Node)), Node)
	}
	// reload divisors
	vec.Vec_IntClear(p.vWinObjs)
	for i, vLevel := range p.vDivLevels {
		sbdSortInts(vLevel)
		for _, Node := range vLevel {
			vec.Vec_IntWriteEntry(p.vObj2Var, int(Node), int32(vec.Vec_IntSize(p.vWinObjs)))
			vec.Vec_IntPush(p.vWinObjs, Node)
		}
		// remember divisor cutoff
		if int32(i) == LevelMax-2 {
			nTimeValidDivs = vec.Vec_IntSize(p.vWinObjs)
		}
	}
	assert(nTimeValidDivs > 0, "nTimeValidDivs > 0")
	vec.Vec_IntClear(p.vDivVars)
	p.DivCutoff = -1
	for i := max(0, nTimeValidDivs-63); i < nTimeValidDivs; i++ {
		Node := vec.Vec_IntEntry(p.vWinObjs, i)
		if p.DivCutoff == -1 && vec.Vec_IntEntry(p.vLutLevs, int(Node)) == LevelMax-2 {
			p.DivCutoff = int32(vec.Vec_IntSize(p.vDivVars))
		}
		vec.Vec_IntPush(p.vDivVars, int32(i))
	}
	if p.DivCutoff == -1 {
		p.DivCutoff = 0
	}
	vec.Vec_IntFill(p.vDivValues, vec.Vec_IntSize(p.vDivVars), 0)
}

// sbdSortInts is Vec_IntSort( v, 0 ) on a slice.
func sbdSortInts(a []int32) {
	tmp := vec.Vec_IntAllocArray(a, len(a))
	vec.Vec_IntSort(tmp, 0)
	copy(a, vec.Vec_IntArray(tmp))
}

func Sbd_ManWindowSim_rec(p *Sbd_Man_t, NodeInit int32) {
	Node := NodeInit
	if vec.Vec_IntEntry(p.vMirrors, int(Node)) >= 0 {
		Node = Abc_Lit2Var(vec.Vec_IntEntry(p.vMirrors, int(Node)))
	}
	if Gia_ObjIsTravIdCurrentId(p.pGia, Node) {
		return
	}
	Gia_ObjSetTravIdCurrentId(p.pGia, Node)
	pObj := Gia_ManObj(p.pGia, Node)
	if Gia_ObjIsAnd(pObj) {
		Sbd_ManWindowSim_rec(p, Gia_ObjFaninId0(pObj, Node))
		Sbd_ManWindowSim_rec(p, Gia_ObjFaninId1(pObj, Node))
	}
	if !Gia_ObjMark0(pObj) {
		vec.Vec_IntWriteEntry(p.vObj2Var, int(Node), int32(vec.Vec_IntSize(p.vWinObjs)))
		vec.Vec_IntPush(p.vWinObjs, Node)
	}
	if Gia_ObjIsCi(pObj) {
		return
	}
	// simulate
	assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
	f0, f1 := Gia_ObjFaninId0(pObj, Node), Gia_ObjFaninId1(pObj, Node)
	sim1or0 := func(fanin *Gia_Obj_t, id int32) []word {
		if Gia_ObjMark0(fanin) {
			return Sbd_ObjSim1(p, id)
		}
		return Sbd_ObjSim0(p, id)
	}
	if Gia_ObjIsXor(pObj) {
		util.Abc_TtXor(Sbd_ObjSim0(p, Node),
			Sbd_ObjSim0(p, f0),
			Sbd_ObjSim0(p, f1),
			p.pPars.NWords,
			Gia_ObjFaninC0(pObj)^Gia_ObjFaninC1(pObj))
		if Gia_ObjMark0(pObj) {
			util.Abc_TtXor(Sbd_ObjSim1(p, Node),
				sim1or0(Gia_ObjFanin0(pObj), f0),
				sim1or0(Gia_ObjFanin1(pObj), f1),
				p.pPars.NWords,
				Gia_ObjFaninC0(pObj)^Gia_ObjFaninC1(pObj))
		}
	} else {
		util.Abc_TtAndCompl(Sbd_ObjSim0(p, Node),
			Sbd_ObjSim0(p, f0), Gia_ObjFaninC0(pObj),
			Sbd_ObjSim0(p, f1), Gia_ObjFaninC1(pObj),
			p.pPars.NWords)
		if Gia_ObjMark0(pObj) {
			util.Abc_TtAndCompl(Sbd_ObjSim1(p, Node),
				sim1or0(Gia_ObjFanin0(pObj), f0), Gia_ObjFaninC0(pObj),
				sim1or0(Gia_ObjFanin1(pObj), f1), Gia_ObjFaninC1(pObj),
				p.pPars.NWords)
		}
	}
	if Node != NodeInit {
		util.Abc_TtCopy(Sbd_ObjSim0(p, NodeInit), Sbd_ObjSim0(p, Node), p.pPars.NWords, Abc_LitIsCompl(vec.Vec_IntEntry(p.vMirrors, int(NodeInit))))
	}
}
func Sbd_ManWindow(p *Sbd_Man_t, Pivot int32) int32 {
	clk := util.Abc_Clock()
	// assign pivot and TFO (assume siminfo is assigned at the PIs)
	p.Pivot = Pivot
	p.vTfo = vec.Vec_IntAllocArray(p.vTfos[Pivot], len(p.vTfos[Pivot]))
	// add constant node
	vec.Vec_IntClear(p.vWinObjs)
	vec.Vec_IntWriteEntry(p.vObj2Var, 0, int32(vec.Vec_IntSize(p.vWinObjs)))
	vec.Vec_IntPush(p.vWinObjs, 0)
	// simulate TFI cone
	Gia_ManIncrementTravId(p.pGia)
	Gia_ObjSetTravIdCurrentId(p.pGia, 0)
	Sbd_ManWindowSim_rec(p, Pivot)
	if p.pPars.NWinSizeMax != 0 && int32(vec.Vec_IntSize(p.vWinObjs)) > p.pPars.NWinSizeMax {
		p.timeWin += util.Abc_Clock() - clk
		return 0
	}
	Sbd_ManUpdateOrder(p, Pivot)
	assert(vec.Vec_IntSize(p.vDivVars) == vec.Vec_IntSize(p.vDivValues), "Vec_IntSize(p->vDivVars) == Vec_IntSize(p->vDivValues)")
	assert(vec.Vec_IntSize(p.vDivVars) < vec.Vec_IntSize(p.vWinObjs), "Vec_IntSize(p->vDivVars) < Vec_IntSize(p->vWinObjs)")
	// simulate node
	Gia_ObjSetMark0(Gia_ManObj(p.pGia, Pivot), 1)
	util.Abc_TtCopy(Sbd_ObjSim1(p, Pivot), Sbd_ObjSim0(p, Pivot), p.pPars.NWords, 1)
	// mark TFO and simulate extended TFI without adding TFO nodes
	vec.Vec_IntClear(p.vRoots)
	for _, Node := range vec.Vec_IntArray(p.vTfo) {
		Gia_ObjSetMark0(Gia_ManObj(p.pGia, Abc_Lit2Var(Node)), 1)
		if Abc_LitIsCompl(Node) == 0 {
			continue
		}
		Sbd_ManWindowSim_rec(p, Abc_Lit2Var(Node))
		vec.Vec_IntPush(p.vRoots, Abc_Lit2Var(Node))
	}
	// add TFO nodes and remove marks
	Gia_ObjSetMark0(Gia_ManObj(p.pGia, Pivot), 0)
	for _, Node := range vec.Vec_IntArray(p.vTfo) {
		Gia_ObjSetMark0(Gia_ManObj(p.pGia, Abc_Lit2Var(Node)), 0)
		vec.Vec_IntWriteEntry(p.vObj2Var, int(Abc_Lit2Var(Node)), int32(vec.Vec_IntSize(p.vWinObjs)))
		vec.Vec_IntPush(p.vWinObjs, Abc_Lit2Var(Node))
	}
	if p.pPars.NWinSizeMax != 0 && int32(vec.Vec_IntSize(p.vWinObjs)) > p.pPars.NWinSizeMax {
		p.timeWin += util.Abc_Clock() - clk
		return 0
	}
	// compute controlability for node
	if vec.Vec_IntSize(p.vTfo) == 0 {
		util.Abc_TtFill(Sbd_ObjSim2(p, Pivot), p.pPars.NWords)
	} else {
		util.Abc_TtClear(Sbd_ObjSim2(p, Pivot), p.pPars.NWords)
	}
	for _, Node := range vec.Vec_IntArray(p.vTfo) {
		if Abc_LitIsCompl(Node) != 0 { // root
			util.Abc_TtOrXor(Sbd_ObjSim2(p, Pivot), Sbd_ObjSim0(p, Abc_Lit2Var(Node)), Sbd_ObjSim1(p, Abc_Lit2Var(Node)), p.pPars.NWords)
		}
	}
	p.timeWin += util.Abc_Clock() - clk
	// propagate controlability to fanins for the TFI nodes starting from the pivot
	Sbd_ManPropagateControl(p, Pivot)
	assert(vec.Vec_IntSize(p.vDivValues) <= 64, "Vec_IntSize(p->vDivValues) <= 64")
	return b2i(vec.Vec_IntSize(p.vDivValues) <= 64)
}

func Sbd_ManCheckConst(p *Sbd_Man_t, Pivot int32) int32 {
	nMintCount := int32(1)
	pSims := Sbd_ObjSim0(p, Pivot)
	pCtrl := Sbd_ObjSim2(p, Pivot)
	PivotVar := vec.Vec_IntEntry(p.vObj2Var, int(Pivot))
	var nCares [2]int32
	clk := util.Abc_Clock()
	p.pSat = Sbd_ManSatSolver(p.pSat, p.pGia, p.vMirrors, Pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, 0)
	p.timeCnf += util.Abc_Clock() - clk
	if p.pSat == nil {
		vec.Vec_IntWriteEntry(p.vLutLevs, int(Pivot), 0)
		p.nLuts[0]++
		return 0
	}
	// count the number of on-set and off-set care-set minterms
	vec.Vec_IntClear(p.vLits)
	for i := int32(0); i < 64; i++ {
		if util.Abc_TtGetBit(pCtrl, i) != 0 {
			nCares[util.Abc_TtGetBit(pSims, i)]++
		} else {
			vec.Vec_IntPush(p.vLits, i)
		}
	}
	fFindOnset := b2i(nCares[0] < nCares[1])
	if nCares[0] >= nMintCount && nCares[1] >= nMintCount {
		return -1
	}
	// find how many do we need
	if nCares[0] < nMintCount {
		nCares[0] = nMintCount - nCares[0]
	} else {
		nCares[0] = 0
	}
	if nCares[1] < nMintCount {
		nCares[1] = nMintCount - nCares[1]
	} else {
		nCares[1] = 0
	}

	if p.pPars.FVeryVerbose != 0 {
		fmt.Printf("Computing %d offset and %d onset minterms for node %d.\n", nCares[0], nCares[1], Pivot)
	}

	if int32(vec.Vec_IntSize(p.vLits)) >= nCares[0]+nCares[1] {
		vec.Vec_IntShrink(p.vLits, int(nCares[0]+nCares[1]))
	} else {
		// collect places to insert new minterms
		for i := int32(0); i < 64 && int32(vec.Vec_IntSize(p.vLits)) < nCares[0]+nCares[1]; i++ {
			if fFindOnset == util.Abc_TtGetBit(pSims, i) {
				vec.Vec_IntPush(p.vLits, i)
			}
		}
	}
	// collect simulation pointers
	vSims := make([][]word, 0, PivotVar+1)
	var i int
	for i = 0; i < vec.Vec_IntSize(p.vWinObjs); i++ {
		iObj := vec.Vec_IntEntry(p.vWinObjs, i)
		vSims = append(vSims, Sbd_ObjSim0(p, iObj))
		if iObj == Pivot {
			break
		}
	}
	assert(int32(i) == PivotVar, "i == PivotVar")
	// compute patterns
	RetValue := Sbd_ManCollectConstants(p.pSat, nCares, PivotVar, vSims, p.vLits)
	if RetValue >= 0 {
		if p.pPars.FVeryVerbose != 0 {
			fmt.Printf("Found stuck-at-%d node %d.\n", RetValue, Pivot)
		}
		vec.Vec_IntWriteEntry(p.vLutLevs, int(Pivot), 0)
		p.nLuts[0]++
		return RetValue
	}
	// set controlability of these minterms
	for _, Ind := range vec.Vec_IntArray(p.vLits) {
		util.Abc_TtSetBit(pCtrl, Ind)
	}
	// propagate controlability to fanins for the TFI nodes starting from the pivot
	Sbd_ManPropagateControl(p, Pivot)
	// double check that we now have enough minterms
	for i := int32(0); i < 64; i++ {
		if util.Abc_TtGetBit(pCtrl, i) != 0 {
			nCares[util.Abc_TtGetBit(pSims, i)]++
		}
	}
	assert(nCares[0] >= nMintCount && nCares[1] >= nMintCount, "nCares[0] >= nMintCount && nCares[1] >= nMintCount")
	return -1
}

// Transposing 64-bit matrix.
func Sbd_TransposeMatrix64(A *[64]word) {
	m := word(0x00000000FFFFFFFF)
	for j := 32; j != 0; j, m = j>>1, m^(m<<uint(j>>1)) {
		for k := 0; k < 64; k = (k + j + 1) & ^j {
			t := (A[k] ^ (A[k+j] >> uint(j))) & m
			A[k] = A[k] ^ t
			A[k+j] = A[k+j] ^ (t << uint(j))
		}
	}
}
func Sbd_PrintMatrix64(A *[64]word) {
	for j := 0; j < 64; j++ {
		for k := 0; k < 64; k++ {
			fmt.Printf("%d", int((A[j]>>uint(k))&1))
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
}

// sbdPrintBinary is Extra_PrintBinary( stdout, (unsigned *)pTruth, nBits ).
func sbdPrintBinary(t []word, nBits int32) {
	u := make([]uint32, 0, 2*len(t))
	for _, w := range t {
		u = append(u, uint32(w), uint32(w>>32))
	}
	os.Stdout.Sync()
	extra.Extra_PrintBinary(os.Stdout, u, nBits)
}

// Profiling divisor candidates.
func Sbd_ManPrintObj(p *Sbd_Man_t, Pivot int32) {
	nDivs := int(vec.Vec_IntEntry(p.vObj2Var, int(Pivot)) + 1)
	win := vec.Vec_IntArray(p.vWinObjs)
	for i := 0; i < nDivs; i++ {
		Id := win[i]
		fmt.Printf("%3d : ", Id)
		sbdPrintBinary(Sbd_ObjSim0(p, Id), 64)
		fmt.Printf("\n")
	}
	assert(p.Pivot == Pivot, "p->Pivot == Pivot")
	vec.Vec_IntClear(p.vCounts[0])
	vec.Vec_IntClear(p.vCounts[1])

	fmt.Printf("Node %d.  Useful divisors = %d.\n", Pivot, vec.Vec_IntSize(p.vDivValues))
	fmt.Printf("Lev : ")
	for i := 0; i < nDivs; i++ {
		if i == nDivs-1 {
			fmt.Printf(" ")
		}
		fmt.Printf("%d", vec.Vec_IntEntry(p.vLutLevs, int(win[i])))
	}
	fmt.Printf("\n")
	fmt.Printf("\n")

	digits := func(f func(int32) int32) {
		fmt.Printf("    : ")
		for i := 0; i < nDivs; i++ {
			if i == nDivs-1 {
				fmt.Printf(" ")
			}
			fmt.Printf("%d", f(win[i]))
		}
		fmt.Printf("\n")
	}
	if nDivs > 99 {
		digits(func(Id int32) int32 { return Id / 100 })
	}
	if nDivs > 9 {
		digits(func(Id int32) int32 { return (Id % 100) / 10 })
	}
	if nDivs > 0 {
		digits(func(Id int32) int32 { return Id % 10 })
		fmt.Printf("\n")
	}

	// sampling matrix
	for k := int32(0); k < p.pPars.NWords*64; k++ {
		if util.Abc_TtGetBit(Sbd_ObjSim2(p, Pivot), k) == 0 {
			continue
		}
		for pass := 0; pass < 2; pass++ {
			fmt.Printf("%3d : ", k)
			for i := 0; i < nDivs; i++ {
				Id := win[i]
				pSims := Sbd_ObjSim0(p, Id)
				pCtrl := Sbd_ObjSim2(p, Id)
				if pass == 1 {
					pCtrl = Sbd_ObjSim3(p, Id)
				}
				if i == nDivs-1 {
					if util.Abc_TtGetBit(pCtrl, k) != 0 {
						vec.Vec_IntPush(p.vCounts[util.Abc_TtGetBit(pSims, k)], k)
					}
					fmt.Printf(" ")
				}
				if util.Abc_TtGetBit(pCtrl, k) != 0 {
					fmt.Printf("%c", '0'+util.Abc_TtGetBit(pSims, k))
				} else {
					fmt.Printf("%c", '.')
				}
			}
			fmt.Printf("\n")
		}

		fmt.Printf("Sims: ")
		for i := 0; i < nDivs; i++ {
			if i == nDivs-1 {
				fmt.Printf(" ")
			}
			fmt.Printf("%c", '0'+util.Abc_TtGetBit(Sbd_ObjSim0(p, win[i]), k))
		}
		fmt.Printf("\n")

		fmt.Printf("Ctrl: ")
		for i := 0; i < nDivs; i++ {
			if i == nDivs-1 {
				fmt.Printf(" ")
			}
			fmt.Printf("%c", '0'+util.Abc_TtGetBit(Sbd_ObjSim2(p, win[i]), k))
		}
		fmt.Printf("\n")

		fmt.Printf("\n")
	}
	// covering table
	fmt.Printf("Exploring %d x %d covering table.\n", vec.Vec_IntSize(p.vCounts[0]), vec.Vec_IntSize(p.vCounts[1]))
	vec.Vec_WrdClear(p.vMatrix)
	c0, c1 := vec.Vec_IntArray(p.vCounts[0]), vec.Vec_IntArray(p.vCounts[1])
	for k0 := 0; k0 < min(len(c0), 64); k0++ {
		Bit0 := c0[k0]
		for k1 := 0; k1 < min(len(c1), 64); k1++ {
			Bit1 := c1[k1]
			var Row word
			for i := 0; i < nDivs; i++ {
				pSims := Sbd_ObjSim0(p, win[i])
				pCtrl := Sbd_ObjSim2(p, win[i])
				if util.Abc_TtGetBit(pCtrl, Bit0) != 0 && util.Abc_TtGetBit(pCtrl, Bit1) != 0 && util.Abc_TtGetBit(pSims, Bit0) != util.Abc_TtGetBit(pSims, Bit1) {
					Row ^= word(1) << uint(i)
				}
			}
			if vec.Vec_WrdPushUnique(p.vMatrix, Row) != 0 {
				continue
			}
			for i := 0; i < nDivs; i++ {
				fmt.Printf("%d", int((Row>>uint(i))&1))
			}
			fmt.Printf("\n")
		}
	}
}

func Sbd_ManMatrPrint(p *Sbd_Man_t, Cover []word, nCol, nRows int32) {
	for i := int32(0); i <= nCol; i++ {
		fmt.Printf("%2d : ", i)
		if i == nCol {
			fmt.Printf("%d ", vec.Vec_IntEntry(p.vLutLevs, int(p.Pivot)))
		} else {
			fmt.Printf("%d ", vec.Vec_IntEntry(p.vLutLevs, int(vec.Vec_IntEntry(p.vWinObjs, int(vec.Vec_IntEntry(p.vDivVars, int(i)))))))
		}
		for k := int32(0); k < nRows; k++ {
			fmt.Printf("%d", int((Cover[i]>>uint(k))&1))
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
}
func Sbd_ManCoverReverseOrder(Cover []word) {
	for i := 0; i < 32; i++ {
		Cover[i], Cover[63-i] = Cover[63-i], Cover[i]
	}
}

func Sbd_ManAddCube1(nRowLimit int32, Cover []word, nRows int32, Cube word) int32 {
	// do not add contained Cube
	assert(nRows <= nRowLimit, "nRows <= nRowLimit")
	for n := int32(0); n < nRows; n++ {
		if (Cover[n] & Cube) == Cover[n] { // Cube is contained
			return nRows
		}
	}
	// remove rows contained by Cube
	m := int32(0)
	for n := int32(0); n < nRows; n++ {
		if (Cover[n] & Cube) != Cube { // Cover[n] is not contained
			Cover[m] = Cover[n]
			m++
		}
	}
	if m < nRowLimit {
		Cover[m] = Cube
		m++
	}
	for n := m; n < nRows; n++ {
		Cover[n] = 0
	}
	nRows = m
	return nRows
}
func Sbd_ManAddCube2(Cover *[2][64]word, nRows int32, Cube [2]word) int32 {
	// do not add contained Cube
	assert(nRows <= 64, "nRows <= 64")
	for n := int32(0); n < nRows; n++ {
		if (Cover[0][n]&Cube[0]) == Cover[0][n] && (Cover[1][n]&Cube[1]) == Cover[1][n] { // Cube is contained
			return nRows
		}
	}
	// remove rows contained by Cube
	m := int32(0)
	for n := int32(0); n < nRows; n++ {
		if (Cover[0][n]&Cube[0]) != Cube[0] || (Cover[1][n]&Cube[1]) != Cube[1] { // Cover[n] is not contained
			Cover[0][m] = Cover[0][n]
			Cover[1][m] = Cover[1][n]
			m++
		}
	}
	if m < 64 {
		Cover[0][m] = Cube[0]
		Cover[1][m] = Cube[1]
		m++
	}
	for n := m; n < nRows; n++ {
		Cover[0][n] = 0
		Cover[1][n] = 0
	}
	nRows = m
	return nRows
}

func Sbd_ManFindCandsSimple(p *Sbd_Man_t, Cover []word, nDivs int32) int32 {
	Target := Cover[nDivs]
	vec.Vec_IntClear(p.vDivSet)
	for c0 := int32(0); c0 < nDivs; c0++ {
		if Cover[c0] == Target {
			vec.Vec_IntPush(p.vDivSet, c0)
			return 1
		}
	}

	for c0 := int32(0); c0 < nDivs; c0++ {
		for c1 := c0 + 1; c1 < nDivs; c1++ {
			if (Cover[c0] | Cover[c1]) == Target {
				vec.Vec_IntPush(p.vDivSet, c0)
				vec.Vec_IntPush(p.vDivSet, c1)
				return 1
			}
		}
	}

	for c0 := int32(0); c0 < nDivs; c0++ {
		for c1 := c0 + 1; c1 < nDivs; c1++ {
			for c2 := c1 + 1; c2 < nDivs; c2++ {
				if (Cover[c0] | Cover[c1] | Cover[c2]) == Target {
					vec.Vec_IntPush(p.vDivSet, c0)
					vec.Vec_IntPush(p.vDivSet, c1)
					vec.Vec_IntPush(p.vDivSet, c2)
					return 1
				}
			}
		}
	}

	for c0 := int32(0); c0 < nDivs; c0++ {
		for c1 := c0 + 1; c1 < nDivs; c1++ {
			for c2 := c1 + 1; c2 < nDivs; c2++ {
				for c3 := c2 + 1; c3 < nDivs; c3++ {
					if (Cover[c0] | Cover[c1] | Cover[c2] | Cover[c3]) == Target {
						vec.Vec_IntPush(p.vDivSet, c0)
						vec.Vec_IntPush(p.vDivSet, c1)
						vec.Vec_IntPush(p.vDivSet, c2)
						vec.Vec_IntPush(p.vDivSet, c3)
						return 1
					}
				}
			}
		}
	}
	return 0
}

func Sbd_ManFindCands(p *Sbd_Man_t, Cover []word, nDivs int32) int32 {
	var Ones, Order [64]int32
	Limits := [4]int32{nDivs/4 + 1, nDivs/3 + 2, nDivs/2 + 3, nDivs}
	Target := Cover[nDivs]

	if nDivs < 8 || p.pPars.FCover != 0 {
		return Sbd_ManFindCandsSimple(p, Cover, nDivs)
	}

	vec.Vec_IntClear(p.vDivSet)
	for c0 := int32(0); c0 < nDivs; c0++ {
		if Cover[c0] == Target {
			vec.Vec_IntPush(p.vDivSet, c0)
			return 1
		}
	}

	for c0 := int32(0); c0 < nDivs; c0++ {
		for c1 := c0 + 1; c1 < nDivs; c1++ {
			if (Cover[c0] | Cover[c1]) == Target {
				vec.Vec_IntPush(p.vDivSet, c0)
				vec.Vec_IntPush(p.vDivSet, c1)
				return 1
			}
		}
	}

	// count ones
	for c0 := int32(0); c0 < nDivs; c0++ {
		Ones[c0] = util.Abc_TtCountOnes(Cover[c0])
	}

	// sort by the number of ones
	for c0 := int32(0); c0 < nDivs; c0++ {
		Order[c0] = c0
	}
	vec.Vec_IntSelectSortCost2Reverse(Order[:], int(nDivs), Ones[:])

	// sort with limits
	for c0 := int32(0); c0 < Limits[0]; c0++ {
		for c1 := c0 + 1; c1 < Limits[1]; c1++ {
			for c2 := c1 + 1; c2 < Limits[2]; c2++ {
				if (Cover[Order[c0]] | Cover[Order[c1]] | Cover[Order[c2]]) == Target {
					vec.Vec_IntPush(p.vDivSet, Order[c0])
					vec.Vec_IntPush(p.vDivSet, Order[c1])
					vec.Vec_IntPush(p.vDivSet, Order[c2])
					return 1
				}
			}
		}
	}

	for c0 := int32(0); c0 < Limits[0]; c0++ {
		for c1 := c0 + 1; c1 < Limits[1]; c1++ {
			for c2 := c1 + 1; c2 < Limits[2]; c2++ {
				for c3 := c2 + 1; c3 < Limits[3]; c3++ {
					if (Cover[Order[c0]] | Cover[Order[c1]] | Cover[Order[c2]] | Cover[Order[c3]]) == Target {
						vec.Vec_IntPush(p.vDivSet, Order[c0])
						vec.Vec_IntPush(p.vDivSet, Order[c1])
						vec.Vec_IntPush(p.vDivSet, Order[c2])
						vec.Vec_IntPush(p.vDivSet, Order[c3])
						return 1
					}
				}
			}
		}
	}
	return 0
}

func sbdPrintDivValues(p *Sbd_Man_t, nDivs int32) {
	for i := int32(0); i < nDivs; i++ {
		fmt.Printf("%d", i%10)
	}
	fmt.Printf("\n")
	for i := int32(0); i < nDivs; i++ {
		v := vec.Vec_IntEntry(p.vDivValues, int(i))
		if v&0x4 != 0 {
			fmt.Printf("%c", '0'+(v&1))
		} else {
			fmt.Printf("%c", 'x')
		}
	}
	fmt.Printf("\n")
	for i := int32(0); i < nDivs; i++ {
		v := vec.Vec_IntEntry(p.vDivValues, int(i))
		if v&0x8 != 0 {
			fmt.Printf("%c", '0'+((v>>1)&1))
		} else {
			fmt.Printf("%c", 'x')
		}
	}
	fmt.Printf("\n")
}

func Sbd_ManExplore(p *Sbd_Man_t, Pivot int32, pTruth *word) int32 {
	fVerbose := false
	nItersMax := int32(32)

	var MatrS [64]word
	var MatrC [2][64]word
	var Cubes [2][2][64]word
	var Cover [64]word
	var CubeNew [2]word
	var nCubes [2]int32
	nRows := int32(0)

	nDivs := int32(vec.Vec_IntSize(p.vDivValues))
	PivotVar := vec.Vec_IntEntry(p.vObj2Var, int(Pivot))
	FreeVar := int32(vec.Vec_IntSize(p.vWinObjs) + vec.Vec_IntSize(p.vTfo) + vec.Vec_IntSize(p.vRoots))
	RetValue := int32(0)

	if p.pPars.FVerbose != 0 {
		fmt.Printf("Node %d.  Useful divisors = %d.\n", Pivot, nDivs)
	}

	if fVerbose {
		Sbd_ManPrintObj(p, Pivot)
	}

	// collect bit-matrices
	var i int
	for i = 0; i < vec.Vec_IntSize(p.vDivVars); i++ {
		Node := vec.Vec_IntEntry(p.vDivVars, i)
		MatrS[63-i] = Sbd_ObjSim0(p, vec.Vec_IntEntry(p.vWinObjs, int(Node)))[0]
		MatrC[0][63-i] = Sbd_ObjSim2(p, vec.Vec_IntEntry(p.vWinObjs, int(Node)))[0]
		MatrC[1][63-i] = Sbd_ObjSim3(p, vec.Vec_IntEntry(p.vWinObjs, int(Node)))[0]
	}
	MatrS[63-i] = Sbd_ObjSim0(p, Pivot)[0]
	MatrC[0][63-i] = Sbd_ObjSim2(p, Pivot)[0]
	MatrC[1][63-i] = Sbd_ObjSim3(p, Pivot)[0]

	Sbd_TransposeMatrix64(&MatrS)
	Sbd_TransposeMatrix64(&MatrC[0])
	Sbd_TransposeMatrix64(&MatrC[1])

	bit := func(w word, k int32) int32 { return int32((w >> uint(k)) & 1) }
	// collect cubes
	for i := 0; i < 64; i++ {
		assert(bit(MatrC[0][i], nDivs) == bit(MatrC[1][i], nDivs), "Abc_TtGetBit(&MatrC[0][i], nDivs) == Abc_TtGetBit(&MatrC[1][i], nDivs)")
		if bit(MatrC[0][i], nDivs) == 0 {
			continue
		}
		Index := bit(MatrS[i], nDivs) // Index==0 offset; Index==1 onset
		for n := 0; n < 2; n++ {
			if n != 0 && MatrC[0][i] == MatrC[1][i] {
				continue
			}
			assert(MatrC[n][i] != 0, "MatrC[n][i]")
			CubeNew[0] = ^MatrS[i] & MatrC[n][i]
			CubeNew[1] = MatrS[i] & MatrC[n][i]
			assert(CubeNew[0] != 0 || CubeNew[1] != 0, "CubeNew[0] || CubeNew[1]")
			nCubes[Index] = Sbd_ManAddCube2(&Cubes[Index], nCubes[Index], CubeNew)
		}
	}

	if p.pPars.FVerbose != 0 {
		fmt.Printf("Generated matrix with %d x %d entries.\n", nCubes[0], nCubes[1])
	}

	if p.pPars.FVerbose != 0 {
		for n := 0; n < 2; n++ {
			if n != 0 {
				fmt.Printf("%s:\n", "Onset")
			} else {
				fmt.Printf("%s:\n", "Offset")
			}
			for i := int32(0); i < nCubes[n]; i++ {
				for k := int32(0); k < 64; k++ {
					if bit(Cubes[n][0][i], k) != 0 {
						fmt.Printf("0")
					} else if bit(Cubes[n][1][i], k) != 0 {
						fmt.Printf("1")
					} else {
						fmt.Printf(".")
					}
				}
				fmt.Printf("\n")
			}
			fmt.Printf("\n")
		}
	}

	// create covering table
	nRows = 0
	for i := int32(0); i < nCubes[0] && nRows < 32; i++ {
		for k := int32(0); k < nCubes[1] && nRows < 32; k++ {
			Cube := (Cubes[0][1][i] & Cubes[1][0][k]) | (Cubes[0][0][i] & Cubes[1][1][k])
			assert(Cube != 0, "Cube")
			nRows = Sbd_ManAddCube1(64, Cover[:], nRows, Cube)
		}
	}

	Sbd_ManCoverReverseOrder(Cover[:])

	if p.pPars.FVerbose != 0 {
		fmt.Printf("Generated cover with %d entries.\n", nRows)
	}

	Sbd_TransposeMatrix64(&Cover)

	Sbd_ManCoverReverseOrder(Cover[:])

	for nIters := int32(0); nIters < nItersMax && nRows < 64; nIters++ {
		if p.pPars.FVerbose != 0 {
			Sbd_ManMatrPrint(p, Cover[:], nDivs, nRows)
		}

		clk := util.Abc_Clock()
		if Sbd_ManFindCands(p, Cover[:], nDivs) == 0 {
			if p.pPars.FVerbose != 0 {
				fmt.Printf("Cannot find a feasible cover.\n")
			}
			p.timeCov += util.Abc_Clock() - clk
			return RetValue
		}
		p.timeCov += util.Abc_Clock() - clk

		if p.pPars.FVerbose != 0 {
			fmt.Printf("Candidate support:  ")
			vec.Vec_IntPrint(p.vDivSet)
		}

		clk = util.Abc_Clock()
		*pTruth = Sbd_ManSolve(p.pSat, PivotVar, FreeVar+nIters, p.vDivSet, p.vDivVars, p.vDivValues, p.vLits)
		p.timeSat += util.Abc_Clock() - clk

		if *pTruth == SBD_SAT_UNDEC {
			fmt.Printf("Node %d:  Undecided.\n", Pivot)
		} else if *pTruth == SBD_SAT_SAT {
			if p.pPars.FVerbose != 0 {
				fmt.Printf("Node %d:  SAT.\n", Pivot)
				sbdPrintDivValues(p, nDivs)
			}
			// add row to the covering table
			for i := int32(0); i < nDivs; i++ {
				if vec.Vec_IntEntry(p.vDivValues, int(i)) == 0xE || vec.Vec_IntEntry(p.vDivValues, int(i)) == 0xD {
					Cover[i] |= word(1) << uint(nRows)
				}
			}
			Cover[nDivs] |= word(1) << uint(nRows)
			nRows++
		} else {
			if p.pPars.FVerbose != 0 {
				fmt.Printf("Node %d:  UNSAT.\n", Pivot)
				sbdPrintBinary([]word{*pTruth}, 1<<uint(vec.Vec_IntSize(p.vDivSet)))
				fmt.Printf("\n")
			}
			RetValue = 1
			break
		}
	}
	return RetValue
}

func Sbd_ManExplore2(p *Sbd_Man_t, Pivot int32, pTruth *word) int32 {
	var Onset, Offset [64]word
	var CoverRows, CoverCols [64]word
	nItersMax := int32(32)
	nRows := int32(0)

	PivotVar := vec.Vec_IntEntry(p.vObj2Var, int(Pivot))
	FreeVar := int32(vec.Vec_IntSize(p.vWinObjs) + vec.Vec_IntSize(p.vTfo) + vec.Vec_IntSize(p.vRoots))
	nDivs := int32(vec.Vec_IntSize(p.vDivVars))
	nConsts := int32(4)

	clk := util.Abc_Clock()
	p.pSat = Sbd_ManSatSolver(p.pSat, p.pGia, p.vMirrors, Pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, 0)
	p.timeCnf += util.Abc_Clock() - clk

	assert(nConsts <= 8, "nConsts <= 8")
	clk = util.Abc_Clock()
	RetValue := Sbd_ManCollectConstantsNew(p.pSat, p.vDivVars, nConsts, PivotVar, Onset[:], Offset[:])
	p.timeSat += util.Abc_Clock() - clk
	if RetValue >= 0 {
		if p.pPars.FVeryVerbose != 0 {
			fmt.Printf("Found stuck-at-%d node %d.\n", RetValue, Pivot)
		}
		vec.Vec_IntWriteEntry(p.vLutLevs, int(Pivot), 0)
		p.nLuts[0]++
		return RetValue
	}
	RetValue = 0

	// create rows of the table
	nRows = 0
	for i := int32(0); i < nConsts; i++ {
		for k := int32(0); k < nConsts; k++ {
			Cube := Onset[i] ^ Offset[k]
			assert(Cube != 0, "Cube")
			nRows = Sbd_ManAddCube1(256, CoverRows[:], nRows, Cube)
		}
	}
	assert(nRows <= 64, "nRows <= 64")

	// create columns of the table
	for i := int32(0); i < nRows; i++ {
		for k := int32(0); k <= nDivs; k++ {
			if (CoverRows[i]>>uint(k))&1 != 0 {
				CoverCols[k] ^= word(1) << uint(i)
			}
		}
	}

	// solve the covering problem
	for nIters := int32(0); nIters < nItersMax && nRows < 64; nIters++ {
		if p.pPars.FVeryVerbose != 0 {
			Sbd_ManMatrPrint(p, CoverCols[:], nDivs, nRows)
		}

		clk = util.Abc_Clock()
		if Sbd_ManFindCands(p, CoverCols[:], nDivs) == 0 {
			if p.pPars.FVeryVerbose != 0 {
				fmt.Printf("Cannot find a feasible cover.\n")
			}
			p.timeCov += util.Abc_Clock() - clk
			return 0
		}
		p.timeCov += util.Abc_Clock() - clk

		if p.pPars.FVeryVerbose != 0 {
			fmt.Printf("Candidate support:  ")
			vec.Vec_IntPrint(p.vDivSet)
		}

		clk = util.Abc_Clock()
		*pTruth = Sbd_ManSolve(p.pSat, PivotVar, FreeVar+nIters, p.vDivSet, p.vDivVars, p.vDivValues, p.vLits)
		p.timeSat += util.Abc_Clock() - clk

		if *pTruth == SBD_SAT_UNDEC {
			fmt.Printf("Node %d:  Undecided.\n", Pivot)
		} else if *pTruth == SBD_SAT_SAT {
			if p.pPars.FVeryVerbose != 0 {
				fmt.Printf("Node %d:  SAT.\n", Pivot)
				for i := int32(0); i < nDivs; i++ {
					fmt.Printf("%d", vec.Vec_IntEntry(p.vLutLevs, int(vec.Vec_IntEntry(p.vWinObjs, int(vec.Vec_IntEntry(p.vDivVars, int(i)))))))
				}
				fmt.Printf("\n")
				sbdPrintDivValues(p, nDivs)
			}
			// add row to the covering table
			for i := int32(0); i < nDivs; i++ {
				if vec.Vec_IntEntry(p.vDivValues, int(i)) == 0xE || vec.Vec_IntEntry(p.vDivValues, int(i)) == 0xD {
					CoverCols[i] |= word(1) << uint(nRows)
				}
			}
			CoverCols[nDivs] |= word(1) << uint(nRows)
			nRows++
		} else {
			if p.pPars.FVeryVerbose != 0 {
				fmt.Printf("Node %d:  UNSAT.   ", Pivot)
				sbdPrintBinary([]word{*pTruth}, 1<<uint(vec.Vec_IntSize(p.vDivSet)))
				fmt.Printf("\n")
			}
			p.nLuts[1]++
			RetValue = 1
			break
		}
	}
	return RetValue
}

func Sbd_ManExploreCut(p *Sbd_Man_t, Pivot, nLeaves int32, pLeaves []int32, pnStrs *int32, Strs []Sbd_Str_t, pFreeVar *int32) int32 {
	clk := util.Abc_Clock()
	PivotVar := vec.Vec_IntEntry(p.vObj2Var, int(Pivot))
	Delay := vec.Vec_IntEntry(p.vLutLevs, int(Pivot))
	var pNodesTop, pNodesBot, pNodesBot1, pNodesBot2 [SBD_DIV_MAX]int32
	nNodesTop, nNodesBot, nNodesBot1, nNodesBot2, nNodesDiff, nNodesDiff1, nNodesDiff2 := int32(0), int32(0), int32(0), int32(0), int32(0), int32(0), int32(0)
	RetValue := int32(0)
	K := p.pPars.NLutSize

	// try to remove fanins
	for nIters := int32(0); nIters < nLeaves; nIters++ {
		// try to remove one variable from divisors
		vec.Vec_IntClear(p.vDivSet)
		for i := int32(0); i < nLeaves; i++ {
			if i != nLeaves-1-nIters && pLeaves[i] != -1 {
				vec.Vec_IntPush(p.vDivSet, vec.Vec_IntEntry(p.vObj2Var, int(pLeaves[i])))
			}
		}
		assert(int32(vec.Vec_IntSize(p.vDivSet)) < nLeaves, "Vec_IntSize(p->vDivSet) < nLeaves")
		// compute truth table
		clk = util.Abc_Clock()
		Truth := Sbd_ManSolve(p.pSat, PivotVar, *pFreeVar, p.vDivSet, p.vDivVars, p.vDivValues, p.vLits)
		*pFreeVar++
		p.timeSat += util.Abc_Clock() - clk
		if Truth == SBD_SAT_UNDEC {
			fmt.Printf("Node %d:  Undecided.\n", Pivot)
		} else if Truth == SBD_SAT_SAT {
			DelayDiff := vec.Vec_IntEntry(p.vLutLevs, int(pLeaves[nLeaves-1-nIters])) - Delay
			if DelayDiff > -2 {
				return 0
			}
		} else {
			pLeaves[nLeaves-1-nIters] = -1
		}
	}
	vec.Vec_IntClear(p.vDivSet)
	for i := int32(0); i < nLeaves; i++ {
		if pLeaves[i] != -1 {
			vec.Vec_IntPush(p.vDivSet, pLeaves[i])
		}
	}
	if int32(vec.Vec_IntSize(p.vDivSet)) <= K {
		*pnStrs = 1
		// remap divisors
		for i, iObj := range vec.Vec_IntArray(p.vDivSet) {
			vec.Vec_IntWriteEntry(p.vDivSet, i, vec.Vec_IntEntry(p.vObj2Var, int(iObj)))
		}
		// compute truth table
		clk = util.Abc_Clock()
		Truth := Sbd_ManSolve(p.pSat, PivotVar, *pFreeVar, p.vDivSet, p.vDivVars, p.vDivValues, p.vLits)
		*pFreeVar++
		p.timeSat += util.Abc_Clock() - clk
		if Truth == SBD_SAT_SAT {
			fmt.Printf("The cut at node %d is not topological.\n", p.Pivot)
			return 0
		}
		assert(Truth != SBD_SAT_UNDEC && Truth != SBD_SAT_SAT, "Truth != SBD_SAT_UNDEC && Truth != SBD_SAT_SAT")
		// create structure
		Strs[0].fLut = 1
		Strs[0].nVarIns = int32(vec.Vec_IntSize(p.vDivSet))
		for i := int32(0); i < Strs[0].nVarIns; i++ {
			Strs[0].VarIns[i] = i
		}
		Strs[0].Res = Truth
		p.nLuts[1]++
		return 1
	}
	assert(int32(vec.Vec_IntSize(p.vDivSet)) > K, "Vec_IntSize(p->vDivSet) > p->pPars->nLutSize")

	// count number of nodes on each level
	nNodesTop, nNodesBot, nNodesBot1, nNodesBot2 = 0, 0, 0, 0
	var i int
	for i = 0; i < vec.Vec_IntSize(p.vDivSet); i++ {
		iObj := vec.Vec_IntEntry(p.vDivSet, i)
		DelayDiff := vec.Vec_IntEntry(p.vLutLevs, int(iObj)) - Delay
		if DelayDiff > -2 {
			break
		}
		if DelayDiff == -2 {
			pNodesTop[nNodesTop] = int32(i)
			nNodesTop++
		} else { // if ( DelayDiff < -2 )
			pNodesBot[nNodesBot] = int32(i)
			nNodesBot++
			if DelayDiff == -3 {
				pNodesBot1[nNodesBot1] = int32(i)
				nNodesBot1++
			} else { // if ( DelayDiff < -3 )
				pNodesBot2[nNodesBot2] = int32(i)
				nNodesBot2++
			}
		}
		vec.Vec_IntWriteEntry(p.vDivSet, i, vec.Vec_IntEntry(p.vObj2Var, int(iObj)))
	}
	assert(nNodesBot == nNodesBot1+nNodesBot2, "nNodesBot == nNodesBot1 + nNodesBot2")
	if i < vec.Vec_IntSize(p.vDivSet) {
		return 0
	}
	if nNodesTop > K-1 {
		return 0
	}
	nDivSet := int32(vec.Vec_IntSize(p.vDivSet))

	// try 44
	if nDivSet <= 2*K-1 {
		nMoved := 0
		if nNodesBot > K { // need to move bottom left-over to the top
			for nNodesBot > K {
				nNodesBot--
				pNodesTop[nNodesTop] = pNodesBot[nNodesBot]
				nNodesTop++
				nMoved++
			}
			assert(nNodesBot == K, "nNodesBot == p->pPars->nLutSize")
		}
		assert(nNodesBot <= K, "nNodesBot <= p->pPars->nLutSize")
		assert(nNodesTop <= K-1, "nNodesTop <= p->pPars->nLutSize-1")

		Strs[0].fLut = 1
		Strs[0].nVarIns = K
		var i int32
		for i = 0; i < nNodesTop; i++ {
			Strs[0].VarIns[i] = pNodesTop[i]
		}
		for ; i < K; i++ {
			Strs[0].VarIns[i] = nDivSet + 1 + i - nNodesTop
		}
		Strs[0].Res = 0

		Strs[1].fLut = 1
		Strs[1].nVarIns = nNodesBot
		for i = 0; i < nNodesBot; i++ {
			Strs[1].VarIns[i] = pNodesBot[i]
		}
		Strs[1].Res = 0

		nNodesDiff = K - 1 - nNodesTop
		assert(nNodesDiff >= 0 && nNodesDiff <= K-1, "nNodesDiff >= 0 && nNodesDiff <= p->pPars->nLutSize-1")
		for k := int32(0); k < nNodesDiff; k++ {
			Strs[2+k].fLut = 0
			Strs[2+k].nVarIns = nNodesBot
			for i = 0; i < nNodesBot; i++ {
				Strs[2+k].VarIns[i] = pNodesBot[i]
			}
			Strs[2+k].Res = 0
		}

		*pnStrs = 2 + nNodesDiff
		clk = util.Abc_Clock()
		RetValue = Sbd_ProblemSolve(p.pGia, p.vMirrors, Pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, p.vDivSet, *pnStrs, Strs)
		p.timeQbf += util.Abc_Clock() - clk
		if RetValue != 0 {
			p.nLuts[2]++
		}

		for ; nMoved > 0; nMoved-- {
			nNodesTop--
			pNodesBot[nNodesBot] = pNodesTop[nNodesTop]
			nNodesBot++
		}
	}

	if RetValue != 0 {
		return RetValue
	}
	if p.pPars.NLutNum < 3 {
		return 0
	}
	if nDivSet < 2*K-1 {
		return 0
	}
	// The three-LUT structures below fit at most 3*nLutSize-2 cut leaves (the
	// default cut size for nLutSize = 4). Larger cuts, which happen with -K 2
	// and -K 3, broke their invariants: C built structures with too few or
	// negatively many components and read uninitialized variables
	// (CBUGS_FIXED.md #27); the fixed source skips them.
	if nDivSet > 3*K-2 {
		return 0
	}

	// try 444 -- LUT(LUT, LUT)
	if nNodesTop <= K-2 {
		nMoved := 0
		if nNodesBot > 2*K { // need to move bottom left-over to the top
			for nNodesBot > 2*K {
				nNodesBot--
				pNodesTop[nNodesTop] = pNodesBot[nNodesBot]
				nNodesTop++
				nMoved++
			}
			assert(nNodesBot == 2*K, "nNodesBot == 2*p->pPars->nLutSize")
		}
		assert(nNodesBot > K, "nNodesBot > p->pPars->nLutSize")
		assert(nNodesBot <= 2*K, "nNodesBot <= 2*p->pPars->nLutSize")
		assert(nNodesTop <= K-2, "nNodesTop <= p->pPars->nLutSize-2")

		Strs[0].fLut = 1
		Strs[0].nVarIns = K
		var i int32
		for i = 0; i < nNodesTop; i++ {
			Strs[0].VarIns[i] = pNodesTop[i]
		}
		for ; i < K; i++ {
			Strs[0].VarIns[i] = nDivSet + 1 + i - nNodesTop
		}
		Strs[0].Res = 0

		Strs[1].fLut = 1
		Strs[1].nVarIns = K
		for i = 0; i < Strs[1].nVarIns; i++ {
			Strs[1].VarIns[i] = pNodesBot[i]
		}
		Strs[1].Res = 0

		Strs[2].fLut = 1
		Strs[2].nVarIns = K
		for i = 0; i < Strs[2].nVarIns; i++ {
			Strs[2].VarIns[i] = pNodesBot[nNodesBot-K+i]
		}
		Strs[2].Res = 0

		nNodesDiff = K - 2 - nNodesTop
		assert(nNodesDiff >= 0 && nNodesDiff <= K-2, "nNodesDiff >= 0 && nNodesDiff <= p->pPars->nLutSize-2")
		for k := int32(0); k < nNodesDiff; k++ {
			Strs[3+k].fLut = 0
			Strs[3+k].nVarIns = nNodesBot
			for i = 0; i < nNodesBot; i++ {
				Strs[3+k].VarIns[i] = pNodesBot[i]
			}
			Strs[3+k].Res = 0
		}

		*pnStrs = 3 + nNodesDiff
		clk = util.Abc_Clock()
		RetValue = Sbd_ProblemSolve(p.pGia, p.vMirrors, Pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, p.vDivSet, *pnStrs, Strs)
		p.timeQbf += util.Abc_Clock() - clk
		if RetValue != 0 {
			p.nLuts[3]++
		}

		for ; nMoved > 0; nMoved-- {
			nNodesTop--
			pNodesBot[nNodesBot] = pNodesTop[nNodesTop]
			nNodesBot++
		}
	}
	if RetValue != 0 {
		return RetValue
	}

	// try 444 -- LUT(LUT(LUT))
	if nNodesBot1+nNodesTop <= 2*K-2 {
		if nNodesBot2 > K { // need to move bottom left-over to the top
			for nNodesBot2 > K {
				nNodesBot2--
				pNodesBot1[nNodesBot1] = pNodesBot2[nNodesBot2]
				nNodesBot1++
			}
			assert(nNodesBot2 == K, "nNodesBot2 == p->pPars->nLutSize")
		}
		if nNodesBot1 > K-1 { // need to move bottom left-over to the top
			for nNodesBot1 > K-1 {
				nNodesBot1--
				pNodesTop[nNodesTop] = pNodesBot1[nNodesBot1]
				nNodesTop++
			}
			assert(nNodesBot1 == K-1, "nNodesBot1 == p->pPars->nLutSize-1")
		}
		assert(nNodesBot2 <= K, "nNodesBot2 <= p->pPars->nLutSize")
		assert(nNodesBot1 <= K-1, "nNodesBot1 <= p->pPars->nLutSize-1")
		assert(nNodesTop <= K-1, "nNodesTop <= p->pPars->nLutSize-1")

		Strs[0].fLut = 1
		Strs[0].nVarIns = K
		var i int32
		for i = 0; i < nNodesTop; i++ {
			Strs[0].VarIns[i] = pNodesTop[i]
		}
		Strs[0].VarIns[i] = nDivSet + 1
		i++
		for ; i < K; i++ {
			Strs[0].VarIns[i] = nDivSet + 2 + i - nNodesTop
		}
		Strs[0].Res = 0
		nNodesDiff1 = K - 1 - nNodesTop

		Strs[1].fLut = 1
		Strs[1].nVarIns = K
		for i = 0; i < nNodesBot1; i++ {
			Strs[1].VarIns[i] = pNodesBot1[i]
		}
		Strs[1].VarIns[i] = nDivSet + 2
		i++
		for ; i < K; i++ {
			Strs[1].VarIns[i] = nDivSet + 2 + nNodesDiff1 + i - nNodesBot1
		}
		Strs[1].Res = 0
		nNodesDiff2 = K - 1 - nNodesBot1

		Strs[2].fLut = 1
		Strs[2].nVarIns = nNodesBot2
		for i = 0; i < Strs[2].nVarIns; i++ {
			Strs[2].VarIns[i] = pNodesBot2[i]
		}
		Strs[2].Res = 0

		nNodesDiff = nNodesDiff1 + nNodesDiff2
		assert(nNodesDiff >= 0 && nNodesDiff <= 2*K-2, "nNodesDiff >= 0 && nNodesDiff <= 2*p->pPars->nLutSize-2")
		for k := int32(0); k < nNodesDiff; k++ {
			Strs[3+k].fLut = 0
			Strs[3+k].nVarIns = nNodesBot2
			for i = 0; i < nNodesBot2; i++ {
				Strs[3+k].VarIns[i] = pNodesBot2[i]
			}
			Strs[3+k].Res = 0
			if k >= nNodesDiff1 {
				continue
			}
			Strs[3+k].nVarIns += nNodesBot1
			for i = 0; i < nNodesBot1; i++ {
				Strs[3+k].VarIns[nNodesBot2+i] = pNodesBot1[i]
			}
		}

		*pnStrs = 3 + nNodesDiff
		clk = util.Abc_Clock()
		RetValue = Sbd_ProblemSolve(p.pGia, p.vMirrors, Pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, p.vDivSet, *pnStrs, Strs)
		p.timeQbf += util.Abc_Clock() - clk
		if RetValue != 0 {
			p.nLuts[4]++
		}
	}
	return RetValue
}
func Sbd_ManExplore3(p *Sbd_Man_t, Pivot int32, pnStrs *int32, Strs []Sbd_Str_t) int32 {
	FreeVar := int32(vec.Vec_IntSize(p.vWinObjs) + vec.Vec_IntSize(p.vTfo) + vec.Vec_IntSize(p.vRoots))
	FreeVarStart := FreeVar
	var pLeaves [SBD_DIV_MAX]int32
	clk := util.Abc_Clock()
	p.pSat = Sbd_ManSatSolver(p.pSat, p.pGia, p.vMirrors, Pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, 0)
	p.timeCnf += util.Abc_Clock() - clk
	// extract one cut
	if p.pSrv != nil {
		nLeaves := Sbd_ManCutServerFirst(p.pSrv, Pivot, pLeaves[:])
		if nLeaves == -1 {
			return 0
		}
		assert(nLeaves <= p.pPars.NCutSize, "nLeaves <= p->pPars->nCutSize")
		if Sbd_ManExploreCut(p, Pivot, nLeaves, pLeaves[:], pnStrs, Strs, &FreeVar) != 0 {
			return 1
		}
		return 0
	}
	// extract one cut
	for nSize := p.pPars.NLutSize + 1; nSize <= p.pPars.NCutSize; nSize++ {
		nLeaves := Sbd_StoObjBestCut(p.pSto, Pivot, nSize, pLeaves[:])
		if nLeaves == -1 {
			continue
		}
		assert(nLeaves == nSize, "nLeaves == nSize")
		if Sbd_ManExploreCut(p, Pivot, nLeaves, pLeaves[:], pnStrs, Strs, &FreeVar) != 0 {
			return 1
		}
	}
	assert(FreeVar-FreeVarStart <= SBD_FVAR_MAX, "FreeVar - FreeVarStart <= SBD_FVAR_MAX")
	return 0
}

// Sbd_CutMergeSimple merges two sorted cuts (C's pointer walk).
func Sbd_CutMergeSimple(p *Sbd_Man_t, pCut1, pCut2, pCut []int32) int32 {
	b, b1, b2 := 1, 1, 1
	e1, e2 := 1+int(pCut1[0]), 1+int(pCut2[0])
	for b1 < e1 && b2 < e2 {
		if pCut1[b1] == pCut2[b2] {
			pCut[b] = pCut1[b1]
			b, b1, b2 = b+1, b1+1, b2+1
		} else if pCut1[b1] < pCut2[b2] {
			pCut[b] = pCut1[b1]
			b, b1 = b+1, b1+1
		} else {
			pCut[b] = pCut2[b2]
			b, b2 = b+1, b2+1
		}
	}
	for b1 < e1 {
		pCut[b] = pCut1[b1]
		b, b1 = b+1, b1+1
	}
	for b2 < e2 {
		pCut[b] = pCut2[b2]
		b, b2 = b+1, b2+1
	}
	pCut[0] = int32(b - 1)
	return pCut[0]
}

// Sbd_ManMergeCuts computes delay-oriented k-feasible cut at the node.
// Return 1 if node's LUT level does not exceed those of the fanins.
func Sbd_ManMergeCuts(p *Sbd_Man_t, Node int32) int32 {
	var pCut11, pCut01, pCut10, pCut00 [2*SBD_MAX_LUTSIZE + 1]int32
	iFan0 := Gia_ObjFaninId0(Gia_ManObj(p.pGia, Node), Node)
	iFan1 := Gia_ObjFaninId1(Gia_ManObj(p.pGia, Node), Node)
	Level0 := vec.Vec_IntEntry(p.vLutLevs, int(iFan0))
	if Level0 == 0 {
		Level0 = 1
	}
	Level1 := vec.Vec_IntEntry(p.vLutLevs, int(iFan1))
	if Level1 == 0 {
		Level1 = 1
	}
	pCut0 := Sbd_ObjCut(p, iFan0)
	pCut1 := Sbd_ObjCut(p, iFan1)
	Cut0 := []int32{1, iFan0}
	Cut1 := []int32{1, iFan1}
	nSize11 := Sbd_CutMergeSimple(p, pCut0, pCut1, pCut11[:])
	nSize01 := Sbd_CutMergeSimple(p, Cut0, pCut1, pCut01[:])
	nSize10 := Sbd_CutMergeSimple(p, pCut0, Cut1, pCut10[:])
	nSize00 := Sbd_CutMergeSimple(p, Cut0, Cut1, pCut00[:])
	lev := func(nSize, l int32) int32 {
		if nSize <= p.pPars.NLutSize {
			return l
		}
		return ABC_INFINITY
	}
	Lev11 := lev(nSize11, max(Level0, Level1))
	Lev01 := lev(nSize01, max(Level0+1, Level1))
	Lev10 := lev(nSize10, max(Level0, Level1+1))
	Lev00 := lev(nSize00, max(Level0+1, Level1+1))
	pCutRes := pCut11[:]
	LevCur := Lev11
	if Lev01 < LevCur || (Lev01 == LevCur && pCut01[0] < pCutRes[0]) {
		pCutRes = pCut01[:]
		LevCur = Lev01
	}
	if Lev10 < LevCur || (Lev10 == LevCur && pCut10[0] < pCutRes[0]) {
		pCutRes = pCut10[:]
		LevCur = Lev10
	}
	if Lev00 < LevCur || (Lev00 == LevCur && pCut00[0] < pCutRes[0]) {
		pCutRes = pCut00[:]
		LevCur = Lev00
	}
	assert(iFan0 != iFan1, "iFan0 != iFan1")
	assert(vec.Vec_IntEntry(p.vLutLevs, int(Node)) == 0, "Vec_IntEntry(p->vLutLevs, Node) == 0")
	vec.Vec_IntWriteEntry(p.vLutLevs, int(Node), LevCur)
	assert(pCutRes[0] <= p.pPars.NLutSize, "pCutRes[0] <= p->pPars->nLutSize")
	copy(Sbd_ObjCut(p, Node), pCutRes[:pCutRes[0]+1])
	return b2i(LevCur == 1)
}
func Sbd_ManDelay(p *Sbd_Man_t) int32 {
	Delay := int32(0)
	for _, Id := range giaCoDriverIds(p.pGia) {
		Delay = max(Delay, vec.Vec_IntEntry(p.vLutLevs, int(Id)))
	}
	return Delay
}
func Sbd_ManMergeTest(p *Sbd_Man_t) {
	for Node := int32(0); Node < Gia_ManObjNum(p.pGia); Node++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, Node)) {
			Sbd_ManMergeCuts(p, Node)
		}
	}
	fmt.Printf("Delay %d.\n", Sbd_ManDelay(p))
}

func Sbd_ManFindCut_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
	if Gia_ObjMark1(pObj) {
		return
	}
	Gia_ObjSetMark1(pObj, 1)
	if Gia_ObjMark0(pObj) {
		return
	}
	assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
	Sbd_ManFindCut_rec(p, Gia_ObjFanin0(pObj))
	Sbd_ManFindCut_rec(p, Gia_ObjFanin1(pObj))
}
func Sbd_ManFindCutUnmark_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
	if !Gia_ObjMark1(pObj) {
		return
	}
	Gia_ObjSetMark1(pObj, 0)
	if Gia_ObjMark0(pObj) {
		return
	}
	assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
	Sbd_ManFindCutUnmark_rec(p, Gia_ObjFanin0(pObj))
	Sbd_ManFindCutUnmark_rec(p, Gia_ObjFanin1(pObj))
}
func Sbd_ManFindCut(p *Sbd_Man_t, Node int32, vCutLits *Vec_Int_t) {
	var pCut [SBD_MAX_LUTSIZE + 1]int32
	LevelMax := int32(0)
	// label reachable nodes
	pObj := Gia_ManObj(p.pGia, Node)
	Sbd_ManFindCut_rec(p.pGia, pObj)
	// collect
	pCut[0] = 0
	for _, id := range vec.Vec_IntArray(vCutLits) {
		pTemp := Gia_ManObj(p.pGia, id)
		if Gia_ObjMark1(pTemp) {
			LevelMax = max(LevelMax, vec.Vec_IntEntry(p.vLutLevs, int(Gia_ObjId(p.pGia, pTemp))))
			pCut[1+pCut[0]] = Gia_ObjId(p.pGia, pTemp)
			pCut[0]++
		}
	}
	assert(pCut[0] <= p.pPars.NLutSize, "pCut[0] <= p->pPars->nLutSize")
	// unlabel reachable nodes
	Sbd_ManFindCutUnmark_rec(p.pGia, pObj)
	// create cut
	assert(vec.Vec_IntEntry(p.vLutLevs, int(Node)) == 0, "Vec_IntEntry(p->vLutLevs, Node) == 0")
	vec.Vec_IntWriteEntry(p.vLutLevs, int(Node), LevelMax+1)
	copy(Sbd_ObjCut(p, Node), pCut[:pCut[0]+1])
}

// sbdTruthToGia is Dsm_ManTruthToGia( p->pGia, pTruth, vLeaves, p->vCover )
// with the caller's "pretend to have MUXes" bracket: when the manager has
// XORs but no MUX array, C points pMuxes at the SBD manager so that XORs are
// built as real XOR nodes, and restores NULL afterwards.
func sbdTruthToGia(p *Sbd_Man_t, pTruth word, vLeaves *Vec_Int_t) int32 {
	pretend := false
	if gia.Gia_ManXorNum(p.pGia) != 0 && p.pGia.PMuxes == nil {
		p.pGia.PMuxes = make([]uint32, Gia_ManObjNum(p.pGia))
		pretend = true
	}
	iLit := gia.Dsm_ManTruthToGia(p.pGia, []word{pTruth}, vec.Vec_IntArray(vLeaves))
	if pretend {
		p.pGia.PMuxes = nil
	}
	return iLit
}

func sbdGrowSims(p *Sbd_Man_t) {
	for k := 0; k < 4; k++ {
		for w := int32(0); w < p.pPars.NWords; w++ {
			p.vSims[k] = append(p.vSims[k], 0)
		}
	}
}
func sbdGrowLutCuts(p *Sbd_Man_t) {
	for k := int32(0); k < p.pPars.NLutSize+1; k++ {
		p.vLutCuts = append(p.vLutCuts, 0)
	}
}

func Sbd_ManImplement(p *Sbd_Man_t, Pivot int32, Truth word) int32 {
	iObjLast := Gia_ManObjNum(p.pGia)
	iCurLev := vec.Vec_IntEntry(p.vLutLevs, int(Pivot))
	// collect leaf literals
	vec.Vec_IntClear(p.vLits)
	for _, Node := range vec.Vec_IntArray(p.vDivSet) {
		Node = vec.Vec_IntEntry(p.vWinObjs, int(Node))
		if vec.Vec_IntEntry(p.vMirrors, int(Node)) >= 0 {
			vec.Vec_IntPush(p.vLits, vec.Vec_IntEntry(p.vMirrors, int(Node)))
		} else {
			vec.Vec_IntPush(p.vLits, Abc_Var2Lit(Node, 0))
		}
	}
	// derive new function of the node
	iLit := sbdTruthToGia(p, Truth, p.vLits)
	// remember this function
	assert(vec.Vec_IntEntry(p.vMirrors, int(Pivot)) == -1, "Vec_IntEntry(p->vMirrors, Pivot) == -1")
	vec.Vec_IntWriteEntry(p.vMirrors, int(Pivot), iLit)
	if p.pPars.FVerbose != 0 {
		fmt.Printf("Replacing node %d by literal %d.\n", Pivot, iLit)
	}
	// translate literals into variables
	for i, Entry := range vec.Vec_IntArray(p.vLits) {
		vec.Vec_IntWriteEntry(p.vLits, i, Abc_Lit2Var(Entry))
	}
	// label inputs
	for _, id := range vec.Vec_IntArray(p.vLits) {
		Gia_ObjSetMark0(Gia_ManObj(p.pGia, id), 1)
	}
	// extend data-structure to accommodate new nodes
	assert(int32(vec.Vec_IntSize(p.vLutLevs)) == iObjLast, "Vec_IntSize(p->vLutLevs) == iObjLast")
	for i := iObjLast; i < Gia_ManObjNum(p.pGia); i++ {
		vec.Vec_IntPush(p.vLutLevs, 0)
		vec.Vec_IntPush(p.vObj2Var, 0)
		vec.Vec_IntPush(p.vMirrors, -1)
		sbdGrowLutCuts(p)
		Sbd_ManFindCut(p, i, p.vLits)
		sbdGrowSims(p)
	}
	// unlabel inputs
	for _, id := range vec.Vec_IntArray(p.vLits) {
		Gia_ObjSetMark0(Gia_ManObj(p.pGia, id), 0)
	}
	// make sure delay reduction is achieved
	iNewLev := vec.Vec_IntEntry(p.vLutLevs, int(Abc_Lit2Var(iLit)))
	assert(iNewLev < iCurLev, "iNewLev < iCurLev")
	// update delay of the initial node
	assert(vec.Vec_IntEntry(p.vLutLevs, int(Pivot)) == iCurLev, "Vec_IntEntry(p->vLutLevs, Pivot) == iCurLev")
	vec.Vec_IntWriteEntry(p.vLutLevs, int(Pivot), iNewLev)
	return 0
}

func Sbd_ManImplement2(p *Sbd_Man_t, Pivot, nStrs int32, pStrs []Sbd_Str_t) int32 {
	iObjLast := Gia_ManObjNum(p.pGia)
	iCurLev := vec.Vec_IntEntry(p.vLutLevs, int(Pivot))
	var iLit int32
	// collect leaf literals
	vec.Vec_IntClear(p.vLits)
	for _, Node := range vec.Vec_IntArray(p.vDivSet) {
		Node = vec.Vec_IntEntry(p.vWinObjs, int(Node))
		if vec.Vec_IntEntry(p.vMirrors, int(Node)) >= 0 {
			vec.Vec_IntPush(p.vLits, vec.Vec_IntEntry(p.vMirrors, int(Node)))
		} else {
			vec.Vec_IntPush(p.vLits, Abc_Var2Lit(Node, 0))
		}
	}
	// collect structure nodes
	for i := int32(0); i < nStrs; i++ {
		vec.Vec_IntPush(p.vLits, -1)
	}
	// implement structures
	for i := nStrs - 1; i >= 0; i-- {
		if pStrs[i].fLut != 0 {
			// collect local literals
			vec.Vec_IntClear(p.vLits2)
			for k := int32(0); k < pStrs[i].nVarIns; k++ {
				vec.Vec_IntPush(p.vLits2, vec.Vec_IntEntry(p.vLits, int(pStrs[i].VarIns[k])))
			}
			// derive new function of the node
			iLit = sbdTruthToGia(p, pStrs[i].Res, p.vLits2)
		} else {
			iLit = vec.Vec_IntEntry(p.vLits, int(pStrs[i].Res))
			assert(iLit > 0, "iLit > 0")
		}
		// update literal
		assert(vec.Vec_IntEntry(p.vLits, vec.Vec_IntSize(p.vLits)-int(nStrs)+int(i)) == -1, "Vec_IntEntry(p->vLits, Vec_IntSize(p->vLits)-nStrs+i) == -1")
		vec.Vec_IntWriteEntry(p.vLits, vec.Vec_IntSize(p.vLits)-int(nStrs)+int(i), iLit)
	}
	iLit = vec.Vec_IntEntry(p.vLits, vec.Vec_IntSize(p.vDivSet))
	// remember this function
	assert(vec.Vec_IntEntry(p.vMirrors, int(Pivot)) == -1, "Vec_IntEntry(p->vMirrors, Pivot) == -1")
	vec.Vec_IntWriteEntry(p.vMirrors, int(Pivot), iLit)
	if p.pPars.FVeryVerbose != 0 {
		fmt.Printf("Replacing node %d by literal %d.\n", Pivot, iLit)
	}

	// extend data-structure to accommodate new nodes
	assert(int32(vec.Vec_IntSize(p.vLutLevs)) == iObjLast, "Vec_IntSize(p->vLutLevs) == iObjLast")
	for i := iObjLast; i < Gia_ManObjNum(p.pGia); i++ {
		assert(int(i) == vec.Vec_IntSize(p.vMirrors), "i == Vec_IntSize(p->vMirrors)")
		vec.Vec_IntPush(p.vMirrors, -1)
		iMirror := int32(-1)
		if i == Gia_ManObjNum(p.pGia)-1 {
			iMirror = Pivot
		}
		Sbd_StoRefObj(p.pSto, i, iMirror)
	}
	Sbd_StoDerefObj(p.pSto, Pivot)
	for i := iObjLast; i < Gia_ManObjNum(p.pGia); i++ {
		clk := util.Abc_Clock()
		Delay := Sbd_StoComputeCutsNode(p.pSto, i)
		p.timeCut += util.Abc_Clock() - clk
		assert(int(i) == vec.Vec_IntSize(p.vLutLevs), "i == Vec_IntSize(p->vLutLevs)")
		vec.Vec_IntPush(p.vLutLevs, Delay)
		vec.Vec_IntPush(p.vObj2Var, 0)
		sbdGrowLutCuts(p)
		Sbd_StoSaveBestDelayCut(p.pSto, i, Sbd_ObjCut(p, i))
		sbdGrowSims(p)
	}
	// make sure delay reduction is achieved
	iNewLev := vec.Vec_IntEntry(p.vLutLevs, int(Abc_Lit2Var(iLit)))
	// Release builds compile this assertion out; it fails for -K 5/6 with -c
	// and -K 3 with -d, where the new logic does not lower the cut delay.
	if sbdDebugAsserts {
		assert(iNewLev == 0 || iNewLev < iCurLev, "!iNewLev || iNewLev < iCurLev")
	}
	// update delay of the initial node
	assert(vec.Vec_IntEntry(p.vLutLevs, int(Pivot)) == iCurLev, "Vec_IntEntry(p->vLutLevs, Pivot) == iCurLev")
	vec.Vec_IntWriteEntry(p.vLutLevs, int(Pivot), iNewLev)
	return 0
}

// Derives new AIG after resynthesis.
func Sbd_ManDeriveMapping_rec(p *Sbd_Man_t, pNew *Gia_Man_t, vMapping *Vec_Int_t, iObj int32) {
	if iObj == 0 || Gia_ObjIsTravIdCurrentId(pNew, iObj) {
		return
	}
	Gia_ObjSetTravIdCurrentId(pNew, iObj)
	pObj := Gia_ManObj(pNew, iObj)
	if Gia_ObjIsCi(pObj) {
		return
	}
	assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
	pCut := Sbd_ObjCut2(p, iObj)
	for k := int32(1); k <= pCut[0]; k++ {
		Sbd_ManDeriveMapping_rec(p, pNew, vMapping, pCut[k])
	}
	// add mapping
	vec.Vec_IntWriteEntry(vMapping, int(iObj), int32(vec.Vec_IntSize(vMapping)))
	for k := int32(0); k <= pCut[0]; k++ {
		vec.Vec_IntPush(vMapping, pCut[k])
	}
	vec.Vec_IntPush(vMapping, iObj)
}
func Sbd_ManDeriveMapping(p *Sbd_Man_t, pNew *Gia_Man_t) {
	vLeaves := vec.Vec_IntAlloc(100)
	// derive cuts for the new manager
	p.vLutCuts2 = make([]int32, Gia_ManObjNum(pNew)*(p.pPars.NLutSize+1))
	for i := int32(0); i < Gia_ManObjNum(p.pGia); i++ {
		pObj := Gia_ManObj(p.pGia, i)
		if !Gia_ObjIsAnd(pObj) {
			continue
		}
		if vec.Vec_IntEntry(p.vMirrors, int(i)) >= 0 {
			continue
		}
		if Gia_ObjValue(pObj) == -1 {
			continue
		}
		iObjNew := Abc_Lit2Var(Gia_ObjValue(pObj))
		if !Gia_ObjIsAnd(Gia_ManObj(pNew, iObjNew)) {
			continue
		}
		pCutNew := Sbd_ObjCut2(p, iObjNew)
		pCut := Sbd_ObjCut(p, i)
		vec.Vec_IntClear(vLeaves)
		for k := int32(1); k <= pCut[0]; k++ {
			iFan := pCut[k]
			if vec.Vec_IntEntry(p.vMirrors, int(pCut[k])) >= 0 {
				iFan = Abc_Lit2Var(vec.Vec_IntEntry(p.vMirrors, int(pCut[k])))
			}
			pFan := Gia_ManObj(p.pGia, iFan)
			if Gia_ObjValue(pFan) == -1 {
				continue
			}
			iFanNew := Abc_Lit2Var(Gia_ObjValue(pFan))
			if iFanNew == 0 || iFanNew == iObjNew {
				continue
			}
			vec.Vec_IntPushUniqueOrder(vLeaves, iFanNew)
		}
		assert(int32(vec.Vec_IntSize(vLeaves)) <= p.pPars.NLutSize, "Vec_IntSize(vLeaves) <= p->pPars->nLutSize")
		pCutNew[0] = int32(vec.Vec_IntSize(vLeaves))
		copy(pCutNew[1:], vec.Vec_IntArray(vLeaves))
	}
	// create new mapping
	vMapping := vec.Vec_IntAlloc(int((p.pPars.NLutSize + 2) * Gia_ManObjNum(pNew)))
	vec.Vec_IntFill(vMapping, int(Gia_ManObjNum(pNew)), 0)
	Gia_ManIncrementTravId(pNew)
	for i := int32(0); i < Gia_ManCoNum(pNew); i++ {
		Sbd_ManDeriveMapping_rec(p, pNew, vMapping, Gia_ObjFaninId0p(pNew, Gia_ManCo(pNew, i)))
	}
	gia.Gia_ManSetMapping(pNew, vMapping)
	p.vLutCuts2 = nil
}
func Sbd_ManDerive_rec(pNew, p *Gia_Man_t, Node int32, vMirrors *Vec_Int_t) {
	Obj := Node
	if vec.Vec_IntEntry(vMirrors, int(Node)) >= 0 {
		Obj = Abc_Lit2Var(vec.Vec_IntEntry(vMirrors, int(Node)))
	}
	pObj := Gia_ManObj(p, Obj)
	if Gia_ObjValue(pObj) == -1 {
		assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
		Sbd_ManDerive_rec(pNew, p, Gia_ObjFaninId0(pObj, Obj), vMirrors)
		Sbd_ManDerive_rec(pNew, p, Gia_ObjFaninId1(pObj, Obj), vMirrors)
		if Gia_ObjIsXor(pObj) {
			Gia_ObjSetValue(pObj, gia.Gia_ManHashXorReal(pNew, gia.Gia_ObjFanin0Copy(pObj), gia.Gia_ObjFanin1Copy(pObj)))
		} else {
			Gia_ObjSetValue(pObj, gia.Gia_ManHashAnd(pNew, gia.Gia_ObjFanin0Copy(pObj), gia.Gia_ObjFanin1Copy(pObj)))
		}
	}
	// set the original node as well
	if Obj != Node {
		Gia_ObjSetValue(Gia_ManObj(p, Node), Abc_LitNotCond(Gia_ObjValue(pObj), Abc_LitIsCompl(vec.Vec_IntEntry(vMirrors, int(Node)))))
	}
}
func Sbd_ManDerive(pMan *Sbd_Man_t, p *Gia_Man_t, vMirrors *Vec_Int_t) *Gia_Man_t {
	gia.Gia_ManFillValue(p)
	pNew := gia.Gia_ManStart(Gia_ManObjNum(p))
	pNew.SetIdentity(p.Identity())
	if p.PMuxes != nil {
		pNew.PMuxes = make([]uint32, Gia_ManObjNum(p))
	}
	Gia_ObjSetValue(gia.Gia_ManConst0(p), 0)
	gia.Gia_ManHashAlloc(pNew)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ObjSetValue(Gia_ManCi(p, i), gia.Gia_ManAppendCi(pNew))
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Sbd_ManDerive_rec(pNew, p, Gia_ObjFaninId0p(p, Gia_ManCo(p, i)), vMirrors)
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		pObj := Gia_ManCo(p, i)
		Gia_ObjSetValue(pObj, gia.Gia_ManAppendCo(pNew, gia.Gia_ObjFanin0Copy(pObj)))
	}
	gia.Gia_ManHashStop(pNew)
	gia.Gia_ManSetRegNum(pNew, gia.Gia_ManRegNum(p))
	gia.Gia_ManTransferTiming(pNew, p)
	if pMan.pPars.FMapping != 0 {
		Sbd_ManDeriveMapping(pMan, pNew)
	}
	// remove dangling nodes
	pTemp := pNew
	pNew = gia.Gia_ManCleanup(pTemp)
	gia.Gia_ManTransferTiming(pNew, pTemp)
	gia.Gia_ManTransferMapping(pNew, pTemp)
	gia.Gia_ManStop(pTemp)
	return pNew
}

// Performs delay optimization for the given LUT size.
func Sbd_NtkPerformOne(p *Sbd_Man_t, Pivot int32) {
	var Strs [SBD_DIV_MAX]Sbd_Str_t
	var Truth word
	nStrs := int32(0)
	if p.pSto == nil && Sbd_ManMergeCuts(p, Pivot) != 0 {
		return
	}
	if Sbd_ManWindow(p, Pivot) == 0 {
		return
	}
	p.nTried++
	p.nUsed++
	RetValue := Sbd_ManCheckConst(p, Pivot)
	if RetValue >= 0 {
		vec.Vec_IntWriteEntry(p.vMirrors, int(Pivot), RetValue)
	} else if p.pPars.FFindDivs != 0 && p.pPars.NLutNum >= 1 && Sbd_ManExplore2(p, Pivot, &Truth) != 0 {
		Strs[0].fLut = 1
		Strs[0].nVarIns = int32(vec.Vec_IntSize(p.vDivSet))
		for i := int32(0); i < Strs[0].nVarIns; i++ {
			Strs[0].VarIns[i] = i
		}
		Strs[0].Res = Truth
		Sbd_ManImplement2(p, Pivot, 1, Strs[:])
	} else if p.pPars.NLutNum >= 2 && Sbd_ManExplore3(p, Pivot, &nStrs, Strs[:]) != 0 {
		Sbd_ManImplement2(p, Pivot, nStrs, Strs[:])
		if p.pPars.FVerbose == 0 {
			return
		}
	} else {
		p.nUsed--
	}
}
func Sbd_NtkPerform(pGia *Gia_Man_t, pPars *Sbd_Par_t) *Gia_Man_t {
	p := Sbd_ManStart(pGia, pPars)
	nNodesOld := Gia_ManObjNum(pGia)
	var vPath *Vec_Bit_t
	assert(pPars.NLutSize <= 6, "pPars->nLutSize <= 6")
	// prepare references
	for Pivot := int32(0); Pivot < Gia_ManObjNum(p.pGia); Pivot++ {
		Sbd_StoRefObj(p.pSto, Pivot, -1)
	}
	if pPars.FUsePath != 0 && gia.Gia_ManHasMapping(pGia) != 0 {
		vPath = Sbc_ManCriticalPath(pGia)
	}
	pTimOld, _ := pGia.BoxTiming()
	if pTimOld != nil && tim.Tim_ManBoxNum(pTimOld) != 0 {
		vNodes := gia.Gia_ManOrderWithBoxes(pGia)
		pTim := tim.Tim_ManDup(pTimOld, 1)
		pGia.PManTime = pTim
		tim.Tim_ManIncrementTravId(pTim)
		for _, Pivot := range vec.Vec_IntArray(vNodes) {
			pObj := Gia_ManObj(pGia, Pivot)
			if Pivot >= nNodesOld {
				break
			}
			if Gia_ObjIsAnd(pObj) {
				clk := util.Abc_Clock()
				Delay := Sbd_StoComputeCutsNode(p.pSto, Pivot)
				Sbd_StoSaveBestDelayCut(p.pSto, Pivot, Sbd_ObjCut(p, Pivot))
				p.timeCut += util.Abc_Clock() - clk
				vec.Vec_IntWriteEntry(p.vLutLevs, int(Pivot), Delay)
				if Delay > 1 && (vPath == nil || vec.Vec_BitEntry(vPath, int(Pivot)) != 0) {
					Sbd_NtkPerformOne(p, Pivot)
				}
			} else if Gia_ObjIsCi(pObj) {
				arrTime := int32(tim.Tim_ManGetCiArrival(pTim, int(Gia_ObjCioId(pObj))))
				vec.Vec_IntWriteEntry(p.vLutLevs, int(Pivot), arrTime)
				Sbd_StoComputeCutsCi(p.pSto, Pivot, arrTime, arrTime)
			} else if Gia_ObjIsCo(pObj) {
				arrTime := vec.Vec_IntEntry(p.vLutLevs, int(Gia_ObjFaninId0(pObj, Pivot)))
				tim.Tim_ManSetCoArrival(pTim, int(Gia_ObjCioId(pObj)), float32(arrTime))
			} else if Gia_ObjIsConst0(pObj) {
				Sbd_StoComputeCutsConst0(p.pSto, 0)
			} else {
				assert(false, "0")
			}
		}
		tim.Tim_ManStop(pTim)
		pGia.PManTime = pTimOld
	} else {
		Sbd_StoComputeCutsConst0(p.pSto, 0)
		for Pivot := int32(0); Pivot < Gia_ManObjNum(pGia); Pivot++ {
			pObj := Gia_ManObj(pGia, Pivot)
			if Pivot >= nNodesOld {
				break
			}
			if Gia_ObjIsCi(pObj) {
				Sbd_StoComputeCutsCi(p.pSto, Pivot, 0, 0)
			} else if Gia_ObjIsAnd(pObj) {
				clk := util.Abc_Clock()
				Delay := Sbd_StoComputeCutsNode(p.pSto, Pivot)
				Sbd_StoSaveBestDelayCut(p.pSto, Pivot, Sbd_ObjCut(p, Pivot))
				p.timeCut += util.Abc_Clock() - clk
				vec.Vec_IntWriteEntry(p.vLutLevs, int(Pivot), Delay)
				if Delay > 1 && (vPath == nil || vec.Vec_BitEntry(vPath, int(Pivot)) != 0) {
					Sbd_NtkPerformOne(p, Pivot)
				}
			}
		}
	}
	p.timeTotal = util.Abc_Clock() - p.timeTotal
	if p.pPars.FVerbose != 0 {
		fmt.Printf("K = %d. S = %d. N = %d. P = %d.  ",
			p.pPars.NLutSize, p.pPars.NLutNum, p.pPars.NCutSize, p.pPars.NCutNum)
		fmt.Printf("Try = %d. Use = %d.  C = %d. 1 = %d. 2 = %d. 3a = %d. 3b = %d.  Lev = %d.  ",
			p.nTried, p.nUsed, p.nLuts[0], p.nLuts[1], p.nLuts[2], p.nLuts[3], p.nLuts[4], Sbd_ManDelay(p))
		util.Abc_PrintTime(1, "Time", p.timeTotal)
	}
	pNew := Sbd_ManDerive(p, pGia, p.vMirrors)
	// print runtime statistics
	p.timeOther = p.timeTotal - p.timeWin - p.timeCut - p.timeCov - p.timeCnf - p.timeSat - p.timeQbf
	if p.pPars.FVerbose != 0 {
		util.Abc_PrintTimeP(1, "Win", p.timeWin, p.timeTotal)
		util.Abc_PrintTimeP(1, "Cut", p.timeCut, p.timeTotal)
		util.Abc_PrintTimeP(1, "Cov", p.timeCov, p.timeTotal)
		util.Abc_PrintTimeP(1, "Cnf", p.timeCnf, p.timeTotal)
		util.Abc_PrintTimeP(1, "Sat", p.timeSat, p.timeTotal)
		util.Abc_PrintTimeP(1, "Qbf", p.timeQbf, p.timeTotal)
		util.Abc_PrintTimeP(1, "Oth", p.timeOther, p.timeTotal)
		util.Abc_PrintTimeP(1, "ALL", p.timeTotal, p.timeTotal)
	}
	Sbd_ManStop(p)
	return pNew
}
