package abci

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unsafe"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/aig/saig"
	baseabc "github.com/glycerine/abcgo/base/abc"
	abcio "github.com/glycerine/abcgo/base/io"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab"
	"github.com/glycerine/abcgo/bool/dec"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

import (
	"os"
	"strconv"
)

/**CFile****************************************************************

  FileName    [abcPrint.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Printing statistics.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcPrint.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <math.h>
//#include "base/abc/abc.h"
//#include "bool/dec/dec.h"
//#include "base/main/main.h"
//#include "map/mio/mio.h"
//#include "aig/aig/aig.h"
//#include "map/if/if.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

//extern int s_TotalNodes = 0;
//extern int s_TotalChanges = 0;

var s_MappingTime abctime
var s_MappingMem int32
var s_ResubTime abctime
var s_ResynTime abctime

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [If the network is best, saves it in "best.blif" and returns 1.]

	Description [If the networks are incomparable, saves the new network,
	returns its parameters in the internal parameter structure, and returns 1.
	If the new network is not a logic network, quits without saving and returns 0.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
type abcBestStats struct {
	name                                 string
	depth, flops, nodes, edges, pis, pos int32
}

var abcParsBest abcBestStats

func Abc_NtkCompareAndSaveBest(pNtk *Abc_Ntk_t) int32 {
	if pNtk == nil {
		abcParsBest = abcBestStats{}
		return 0
	}
	if baseabc.Abc_NtkIsLogic(pNtk) == 0 {
		return 0
	}
	current := abcBestStats{name: pNtk.Name, depth: baseabc.Abc_NtkLevel(pNtk), flops: baseabc.Abc_NtkLatchNum(pNtk), nodes: baseabc.Abc_NtkNodeNum(pNtk), edges: baseabc.Abc_NtkGetTotalFanins(pNtk), pis: baseabc.Abc_NtkPiNum(pNtk), pos: baseabc.Abc_NtkPoNum(pNtk)}
	if abcParsBest.name != "" && abcParsBest.name == current.name && !(abcParsBest.depth > current.depth || (abcParsBest.depth == current.depth && abcParsBest.flops > current.flops) || (abcParsBest.depth == current.depth && abcParsBest.flops == current.flops && abcParsBest.edges > current.edges)) {
		return 0
	}
	abcParsBest = current
	out := pNtk.Spec
	if !strings.HasSuffix(out, "_best.blif") {
		out += "_best.blif"
	}
	_ = abcio.Io_Write(pNtk, out, abcio.IO_FILE_BLIF)
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Collects memory usage.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMemory(p *Abc_Ntk_t) float64 {
	if p == nil {
		return 0
	}
	return float64(unsafe.Sizeof(*p)) +
		float64(unsafe.Sizeof(Abc_Obj_t{}))*float64(baseabc.Abc_NtkObjNum(p)) +
		baseabc.Abc_NtkMemoryVectorBytes(p)
}

/*
*Function*************************************************************

	Synopsis    [Marks nodes for power-optimization.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMfsTotalSwitching(pNtk *Abc_Ntk_t) float32 {
	if pNtk == nil {
		return 0
	}
	strashed := baseabc.Abc_NtkStrash(pNtk, 0, 1, 0)
	graph := baseabc.Abc_NtkToDar(strashed, 0, boolInt(baseabc.Abc_NtkLatchNum(pNtk) > 0))
	values := saig.Saig_ManComputeSwitchProb4s(graph, 48, 16, 0)
	defer vec.Vec_IntFree(values)
	var total float32
	for _, obj := range pNtk.Objects() {
		strashObj := baseabc.Abc_ObjCopyEdge(obj).Object
		if strashObj == nil || baseabc.Abc_ObjIsNode(obj) == 0 {
			continue
		}
		id := strashObj.Temp
		if id < 0 || int(id) >= vec.Vec_IntSize(values) {
			continue
		}
		activity := math.Float32frombits(uint32(vec.Vec_IntEntry(values, int(id))))
		total += float32(baseabc.Abc_ObjFanoutNum(obj)) * activity
	}
	baseabc.Abc_NtkDelete(strashed)
	return total
}
func boolInt(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Compute area using LUT library.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkGetArea(pNtk *Abc_Ntk_t) float32 {
	assert(baseabc.Abc_NtkIsLogic(pNtk) != 0, "Abc_NtkIsLogic(pNtk)")
	library := ifpkg.Abc_FrameReadLibLut()
	if library == nil {
		return 0
	}
	var area float32
	for _, node := range pNtk.ObjectsOfKind(baseabc.ObjectNode) {
		fanins := baseabc.Abc_ObjFaninNum(node)
		if fanins > library.LutMax {
			return 0
		}
		area += library.Area(fanins)
	}
	return area
}

/*
*Function*************************************************************

	Synopsis    [Print the vital stats of the network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintStats(pNtk *Abc_Ntk_t, fFactored, fSaveBest, fDumpResult, fUseLutLib, fPrintMuxes, fPower, fGlitch, fSkipBuf, fSkipSmall, fPrintMem int32) {
	if pNtk == nil {
		return
	}
	nSingles := int32(0)
	if fSkipBuf != 0 {
		nSingles = baseabc.Abc_NtkGetBufNum(pNtk)
	}
	if fPrintMuxes != 0 && baseabc.Abc_NtkIsStrash(pNtk) != 0 {
		nXors := baseabc.Abc_NtkGetExorNum(pNtk)
		nMuxes := baseabc.Abc_NtkCountMuxes(pNtk) - nXors
		nAnds := baseabc.Abc_NtkNodeNum(pNtk) - (nMuxes+nXors)*3 - nSingles
		denom := float64(baseabc.Abc_NtkNodeNum(pNtk))
		util.Abc_Print(1, "XMA stats:  ")
		util.Abc_Print(1, "Xor =%7d (%6.2f %%)  ", nXors, 300*float64(nXors)/denom)
		util.Abc_Print(1, "Mux =%7d (%6.2f %%)  ", nMuxes, 300*float64(nMuxes)/denom)
		util.Abc_Print(1, "And =%7d (%6.2f %%)  ", nAnds, 100*float64(nAnds)/denom)
		util.Abc_Print(1, "Total =%7d\n", nAnds+nXors+nMuxes)
		return
	}
	if fSaveBest != 0 {
		Abc_NtkCompareAndSaveBest(pNtk)
	}
	util.Abc_Print(1, "\033[1;37m%-30s:\033[0m", pNtk.Name)
	util.Abc_Print(1, " i/o =%5d/%5d", baseabc.Abc_NtkPiNum(pNtk), baseabc.Abc_NtkPoNum(pNtk))
	if pNtk.Constraints != 0 {
		util.Abc_Print(1, "(c=%d)", pNtk.Constraints)
	}
	util.Abc_Print(1, "  lat =%5d", baseabc.Abc_NtkLatchNum(pNtk))
	if pNtk.Barriers != 0 {
		util.Abc_Print(1, "(b=%d)", pNtk.Barriers)
	}
	switch {
	case baseabc.Abc_NtkIsNetlist(pNtk) != 0:
		util.Abc_Print(1, "  net =%5d", baseabc.Abc_NtkNetNum(pNtk))
		if fSkipSmall != 0 {
			util.Abc_Print(1, "  nd =%5d", baseabc.Abc_NtkGetLargeNodeNum(pNtk))
		} else {
			util.Abc_Print(1, "  nd =%5d", baseabc.Abc_NtkNodeNum(pNtk)-nSingles)
		}
		util.Abc_Print(1, "  wbox =%3d", baseabc.Abc_NtkWhiteboxNum(pNtk))
		util.Abc_Print(1, "  bbox =%3d", baseabc.Abc_NtkBlackboxNum(pNtk))
	case baseabc.Abc_NtkIsStrash(pNtk) != 0:
		util.Abc_Print(1, "  and =%7d", baseabc.Abc_NtkNodeNum(pNtk))
		if choices := baseabc.Abc_NtkGetChoiceNum(pNtk); choices != 0 {
			util.Abc_Print(1, " (choice = %d)", choices)
		}
	default:
		if fSkipSmall != 0 {
			util.Abc_Print(1, "  nd =%6d", baseabc.Abc_NtkGetLargeNodeNum(pNtk))
		} else {
			util.Abc_Print(1, "  nd =%6d", baseabc.Abc_NtkNodeNum(pNtk)-nSingles)
		}
		util.Abc_Print(1, "  edge =%7d", baseabc.Abc_NtkGetTotalFanins(pNtk)-nSingles)
	}
	if baseabc.Abc_NtkIsStrash(pNtk) == 0 && baseabc.Abc_NtkIsNetlist(pNtk) == 0 {
		switch {
		case baseabc.Abc_NtkHasSop(pNtk) != 0:
			util.Abc_Print(1, "  cube =%6d", baseabc.Abc_NtkGetCubeNum(pNtk)-nSingles)
			if fFactored != 0 {
				util.Abc_Print(1, "  lit(sop) =%6d", baseabc.Abc_NtkGetLitNum(pNtk)-nSingles)
				util.Abc_Print(1, "  lit(fac) =%6d", baseabc.Abc_NtkGetLitFactNum(pNtk)-nSingles)
			}
		case baseabc.Abc_NtkHasAig(pNtk) != 0:
			util.Abc_Print(1, "  aig  =%6d", baseabc.Abc_NtkGetAigNodeNum(pNtk)-nSingles)
		case baseabc.Abc_NtkHasBdd(pNtk) != 0:
			util.Abc_Print(1, "  bdd  =%6d", baseabc.Abc_NtkGetBddNodeNum(pNtk)-nSingles)
		case baseabc.Abc_NtkHasMapping(pNtk) != 0:
			util.Abc_Print(1, "  area =%5.2f", baseabc.Abc_NtkGetMappedArea(pNtk))
			util.Abc_Print(1, "  delay =%5.2f", Abc_NtkDelayTrace(pNtk, nil, nil, 0))
		case baseabc.Abc_NtkHasBlackbox(pNtk) == 0:
			panic("Abc_NtkPrintStats: unsupported network function")
		}
	}
	if baseabc.Abc_NtkIsStrash(pNtk) != 0 {
		util.Abc_Print(1, "  lev =%3d", baseabc.Abc_NtkLevel(pNtk))
	} else {
		util.Abc_Print(1, "  lev = %d", baseabc.Abc_NtkLevel(pNtk))
	}
	if buffers := baseabc.Abc_NtkBarBufs2(pNtk); buffers != 0 {
		util.Abc_Print(1, "  buf = %d", buffers)
	}
	if fUseLutLib != 0 && ifpkg.Abc_FrameReadLibLut() != nil {
		util.Abc_Print(1, "  delay =%5.2f", Abc_NtkDelayTraceLut(pNtk, 1))
		util.Abc_Print(1, "  area =%5.2f", Abc_NtkGetArea(pNtk))
	}
	if fPower != 0 {
		util.Abc_Print(1, "  power =%7.2f", Abc_NtkMfsTotalSwitching(pNtk))
	}
	if fGlitch != 0 {
		if baseabc.Abc_NtkIsLogic(pNtk) != 0 && baseabc.Abc_NtkGetFaninMax(pNtk) <= 6 {
			util.Abc_Print(1, "  glitch =%7.2f %%", Abc_NtkMfsTotalGlitching(pNtk, 4000, 8, 0))
		} else {
			fmt.Print("\nCurrently computes glitching only for K-LUT networks with K <= 6.")
		}
	}
	if fPrintMem != 0 {
		util.Abc_Print(1, "  mem =%5.2f MB", Abc_NtkMemory(pNtk)/float64(1<<20))
	}
	util.Abc_Print(1, "\n")
	if fDumpResult != 0 {
		if file, err := os.OpenFile("abcstats.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			_, _ = fmt.Fprintf(file, "%s %d %d %d %d %d \n", pNtk.Name, baseabc.Abc_NtkPiNum(pNtk), baseabc.Abc_NtkPoNum(pNtk), baseabc.Abc_NtkNodeNum(pNtk), baseabc.Abc_NtkGetTotalFanins(pNtk), baseabc.Abc_NtkLevel(pNtk))
			_ = file.Close()
		}
	}
	_ = os.Stdout.Sync()
	if pNtk.Exdc != nil {
		Abc_NtkPrintStats(pNtk.Exdc, fFactored, fSaveBest, fDumpResult, fUseLutLib, fPrintMuxes, fPower, fGlitch, fSkipBuf, fSkipSmall, fPrintMem)
	}
}

/*
*Function*************************************************************

	Synopsis    [Prints PIs/POs and LIs/LOs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintIo(pFile *os.File, pNtk *Abc_Ntk_t, fPrintFlops int32) {
	pis := pNtk.ObjectsOfKind(baseabc.ObjectPI)
	pos := pNtk.ObjectsOfKind(baseabc.ObjectPO)
	fmt.Fprintf(pFile, "Primary inputs (%d): ", len(pis))
	for i, obj := range pis {
		fmt.Fprintf(pFile, " %d=%s", i, obj.Name())
	}
	fmt.Fprintln(pFile)
	fmt.Fprintf(pFile, "Primary outputs (%d):", len(pos))
	for i, obj := range pos {
		fmt.Fprintf(pFile, " %d=%s", i, obj.Name())
	}
	fmt.Fprintln(pFile)
	if fPrintFlops == 0 {
		return
	}
	latches := pNtk.ObjectsOfKind(baseabc.ObjectLatch)
	fmt.Fprintf(pFile, "Latches (%d):  ", len(latches))
	for _, latch := range latches {
		fmt.Fprintf(pFile, " %s(%s=%s)", latch.Name(), baseabc.Abc_ObjFanout0(latch).Name(), baseabc.Abc_ObjFanin0(latch).Name())
	}
	fmt.Fprintln(pFile)
}
func Abc_NtkPrintLatch(pFile *os.File, pNtk *Abc_Ntk_t) {
	assert(baseabc.Abc_NtkIsNetlist(pNtk) == 0, "!Abc_NtkIsNetlist(pNtk)")
	latches := pNtk.ObjectsOfKind(baseabc.ObjectLatch)
	if len(latches) == 0 {
		fmt.Fprintln(pFile, "The network is combinational.")
		return
	}
	var initNums [4]int32
	var constants int32
	for _, latch := range latches {
		init := int32(0)
		if baseabc.Abc_LatchIsInit0(latch) != 0 {
			init = 1
		} else if baseabc.Abc_LatchIsInit1(latch) != 0 {
			init = 2
		} else if baseabc.Abc_LatchIsInitDc(latch) != 0 {
			init = 3
		}
		initNums[init]++
		fanin := baseabc.Abc_ObjFanin0(baseabc.Abc_ObjFanin0(latch))
		if baseabc.Abc_NtkIsLogic(pNtk) != 0 {
			if baseabc.Abc_NodeIsConst(fanin) == 0 {
				continue
			}
		} else if baseabc.Abc_NtkIsStrash(pNtk) != 0 {
			if baseabc.Abc_AigNodeIsConst(fanin) == 0 {
				continue
			}
		} else {
			panic("unsupported latch network")
		}
		constants++
	}
	fmt.Fprintf(pFile, "Total latches = %5d. Init0 = %d. Init1 = %d. InitDC = %d. Const data = %d.\n", len(latches), initNums[1], initNums[2], initNums[3], constants)
}

/**Function*************************************************************

  Synopsis    [Prints statistics about latches.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Prints the distribution of fanins/fanouts in the network.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFaninFanoutCounters(pNtk *Abc_Ntk_t, vFan, vFon, vFanR, vFonR *Vec_Int_t) {
	objects := pNtk.Objects()
	maxIn, maxOut := 0, 0
	for _, obj := range objects {
		if len(obj.Fanins()) > maxIn {
			maxIn = len(obj.Fanins())
		}
		if len(obj.Fanouts()) > maxOut {
			maxOut = len(obj.Fanouts())
		}
	}
	vec.Vec_IntFill(vFan, int(maxIn+1), 0)
	vec.Vec_IntFill(vFon, int(maxOut+1), 0)
	vec.Vec_IntFill(vFanR, int(maxIn+1), 0)
	vec.Vec_IntFill(vFonR, int(maxOut+1), 0)
	for i := len(objects) - 1; i >= 0; i-- {
		obj := objects[i]
		ni, no := len(obj.Fanins()), len(obj.Fanouts())
		vec.Vec_IntAddToEntry(vFan, int(ni), 1)
		vec.Vec_IntAddToEntry(vFon, int(no), 1)
		vec.Vec_IntWriteEntry(vFanR, ni, int32(i))
		vec.Vec_IntWriteEntry(vFonR, no, int32(i))
	}
}
func Abc_NtkInputOutputCounters(pNtk *Abc_Ntk_t, vFan, vFon, vFanR, vFonR *Vec_Int_t) {
	maxIn, maxOut := 0, 0
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		if n := baseabc.Abc_ObjFanoutNum(baseabc.Abc_NtkCi(pNtk, i)); int(n) > maxOut {
			maxOut = int(n)
		}
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		co := baseabc.Abc_NtkCo(pNtk, i)
		if len(baseabc.Abc_ObjFanin0(co).Fanins()) > maxIn {
			maxIn = len(baseabc.Abc_ObjFanin0(co).Fanins())
		}
	}
	vec.Vec_IntFill(vFan, int(maxIn+1), 0)
	vec.Vec_IntFill(vFon, int(maxOut+1), 0)
	vec.Vec_IntFill(vFanR, int(maxIn+1), 0)
	vec.Vec_IntFill(vFonR, int(maxOut+1), 0)
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		obj := baseabc.Abc_NtkCi(pNtk, i)
		k := int(baseabc.Abc_ObjFanoutNum(obj))
		vec.Vec_IntAddToEntry(vFon, int(k), 1)
		vec.Vec_IntWriteEntry(vFonR, k, baseabc.Abc_ObjId(obj))
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		obj := baseabc.Abc_NtkCo(pNtk, i)
		k := int(baseabc.Abc_ObjFaninNum(baseabc.Abc_ObjFanin0(obj)))
		vec.Vec_IntAddToEntry(vFan, int(k), 1)
		vec.Vec_IntWriteEntry(vFanR, k, baseabc.Abc_ObjId(obj))
	}
}
func abcMergeSupport(left, right []int32) []int32 {
	out := make([]int32, 0, len(left)+len(right))
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		if left[i] < right[j] {
			out = append(out, left[i])
			i++
		} else if left[i] > right[j] {
			out = append(out, right[j])
			j++
		} else {
			out = append(out, left[i])
			i++
			j++
		}
	}
	out = append(out, left[i:]...)
	out = append(out, right[j:]...)
	return out
}
func abcNodeSupports(pNtk *Abc_Ntk_t) [][]int32 {
	supports := make([][]int32, baseabc.Abc_NtkObjNumMax(pNtk))
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		obj := baseabc.Abc_NtkCi(pNtk, i)
		supports[baseabc.Abc_ObjId(obj)] = []int32{i}
	}
	for _, node := range baseabc.Abc_NtkDfsIter(pNtk, 0) {
		var current []int32
		for _, edge := range node.Fanins() {
			current = abcMergeSupport(current, supports[baseabc.Abc_ObjId(edge.Object)])
		}
		supports[baseabc.Abc_ObjId(node)] = current
	}
	return supports
}
func Abc_NtkCollectCoSupps(pNtk *Abc_Ntk_t, fVerbose int32) *Vec_Int_t {
	clk := util.Abc_Clock()
	supports := abcNodeSupports(pNtk)
	result := vec.Vec_IntAlloc(int(baseabc.Abc_NtkCoNum(pNtk)))
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		co := baseabc.Abc_NtkCo(pNtk, i)
		vec.Vec_IntPush(result, int32(len(supports[baseabc.Abc_ObjId(baseabc.Abc_ObjFanin0(co))])))
	}
	if fVerbose != 0 {
		util.Abc_PrintTime(1, "Input  support computation", util.Abc_Clock()-clk)
	}
	return result
}
func Abc_NtkCollectCiSupps(pNtk *Abc_Ntk_t, fVerbose int32) *Vec_Int_t {
	clk := util.Abc_Clock()
	supports := abcNodeSupports(pNtk)
	counts := make([]int32, baseabc.Abc_NtkCiNum(pNtk))
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		co := baseabc.Abc_NtkCo(pNtk, i)
		for _, ci := range supports[baseabc.Abc_ObjId(baseabc.Abc_ObjFanin0(co))] {
			counts[ci]++
		}
	}
	result := vec.Vec_IntAlloc(len(counts))
	for _, count := range counts {
		vec.Vec_IntPush(result, count)
	}
	if fVerbose != 0 {
		util.Abc_PrintTime(1, "Output support computation", util.Abc_Clock()-clk)
	}
	return result
}
func Abc_NtkInOutSupportCounters(pNtk *Abc_Ntk_t, vFan *Vec_Int_t, vFon *Vec_Int_t, vFanR *Vec_Int_t, vFonR *Vec_Int_t) {
	vSuppsCo := Abc_NtkCollectCoSupps(pNtk, 1)
	vSuppsCi := Abc_NtkCollectCiSupps(pNtk, 1)
	defer vec.Vec_IntFree(vSuppsCo)
	defer vec.Vec_IntFree(vSuppsCi)
	nFaninsMax, nFanoutsMax := vec.Vec_IntFindMax(vSuppsCo), vec.Vec_IntFindMax(vSuppsCi)
	vec.Vec_IntFill(vFan, int(nFaninsMax+1), 0)
	vec.Vec_IntFill(vFon, int(nFanoutsMax+1), 0)
	vec.Vec_IntFill(vFanR, int(nFaninsMax+1), 0)
	vec.Vec_IntFill(vFonR, int(nFanoutsMax+1), 0)
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		co := baseabc.Abc_NtkCo(pNtk, i)
		n := Vec_IntEntry(vSuppsCo, i)
		vec.Vec_IntAddToEntry(vFan, int(n), 1)
		Vec_IntWriteEntry(vFanR, n, baseabc.Abc_ObjId(co))
	}
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		ci := baseabc.Abc_NtkCi(pNtk, i)
		n := Vec_IntEntry(vSuppsCi, i)
		vec.Vec_IntAddToEntry(vFon, int(n), 1)
		Vec_IntWriteEntry(vFonR, n, baseabc.Abc_ObjId(ci))
	}
}

func Abc_NtkCollectCoCones(pNtk *Abc_Ntk_t, fVerbose int32) *Vec_Int_t {
	clk := util.Abc_Clock()
	nodes := baseabc.Abc_NtkDfsIter(pNtk, 0)
	supports := make([][]int32, baseabc.Abc_NtkObjNumMax(pNtk))
	for i, node := range nodes {
		var set []int32
		for _, edge := range node.Fanins() {
			set = abcMergeSupport(set, supports[baseabc.Abc_ObjId(edge.Object)])
		}
		supports[baseabc.Abc_ObjId(node)] = append(set, int32(i))
	}
	result := vec.Vec_IntAlloc(int(baseabc.Abc_NtkCoNum(pNtk)))
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		co := baseabc.Abc_NtkCo(pNtk, i)
		vec.Vec_IntPush(result, int32(len(supports[baseabc.Abc_ObjId(baseabc.Abc_ObjFanin0(co))])))
	}
	if fVerbose != 0 {
		util.Abc_PrintTime(1, "Input  cone computation", util.Abc_Clock()-clk)
	}
	return result
}
func Abc_NtkCollectCiCones(pNtk *Abc_Ntk_t, fVerbose int32) *Vec_Int_t {
	clk := util.Abc_Clock()
	nodes := baseabc.Abc_NtkDfsIter(pNtk, 0)
	counts := make([]int32, baseabc.Abc_NtkObjNumMax(pNtk))
	for i, node := range nodes {
		for _, edge := range node.Fanins() {
			counts[baseabc.Abc_ObjId(edge.Object)]++
		}
		_ = i
	}
	result := vec.Vec_IntAlloc(int(baseabc.Abc_NtkCiNum(pNtk)))
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		ci := baseabc.Abc_NtkCi(pNtk, i)
		seen := make(map[int32]bool)
		queue := []*Abc_Obj_t{ci}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			for _, fo := range current.Fanouts() {
				id := baseabc.Abc_ObjId(fo)
				if !seen[id] && baseabc.Abc_ObjIsNode(fo) != 0 {
					seen[id] = true
					queue = append(queue, fo)
				}
			}
		}
		vec.Vec_IntPush(result, int32(len(seen)))
	}
	if fVerbose != 0 {
		util.Abc_PrintTime(1, "Output cone computation", util.Abc_Clock()-clk)
	}
	return result
}
func Abc_NtkInOutConeCounters(pNtk *Abc_Ntk_t, vFan *Vec_Int_t, vFon *Vec_Int_t, vFanR *Vec_Int_t, vFonR *Vec_Int_t) {
	vSuppsCo := Abc_NtkCollectCoCones(pNtk, 1)
	vSuppsCi := Abc_NtkCollectCiCones(pNtk, 1)
	defer vec.Vec_IntFree(vSuppsCo)
	defer vec.Vec_IntFree(vSuppsCi)
	nFaninsMax, nFanoutsMax := vec.Vec_IntFindMax(vSuppsCo), vec.Vec_IntFindMax(vSuppsCi)
	vec.Vec_IntFill(vFan, int(nFaninsMax+1), 0)
	vec.Vec_IntFill(vFon, int(nFanoutsMax+1), 0)
	vec.Vec_IntFill(vFanR, int(nFaninsMax+1), 0)
	vec.Vec_IntFill(vFonR, int(nFanoutsMax+1), 0)
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		co := baseabc.Abc_NtkCo(pNtk, i)
		n := Vec_IntEntry(vSuppsCo, i)
		vec.Vec_IntAddToEntry(vFan, int(n), 1)
		Vec_IntWriteEntry(vFanR, n, baseabc.Abc_ObjId(co))
	}
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		ci := baseabc.Abc_NtkCi(pNtk, i)
		n := Vec_IntEntry(vSuppsCi, i)
		vec.Vec_IntAddToEntry(vFon, int(n), 1)
		Vec_IntWriteEntry(vFonR, n, baseabc.Abc_ObjId(ci))
	}
}

func Abc_NtkPrintDistribInternal(pFile *os.File, pNtk *Abc_Ntk_t, pFanins string, pFanouts string, pNode string, pFanin string, pFanout string, vFan *Vec_Int_t, vFon *Vec_Int_t, vFanR *Vec_Int_t, vFonR *Vec_Int_t) {
	var k int32
	var nSizeMax int32 = util.Abc_MaxInt(Vec_IntSize(vFan), Vec_IntSize(vFon))
	fmt.Fprintf(pFile, "The distribution of %s and %s in the network:\n", pFanins, pFanouts)
	fmt.Fprintf(pFile, "  Number   %s with %s  %s with %s          Repr1             Repr2\n", pNode, pFanin, pNode, pFanout)
	for k = 0; k < nSizeMax; k++ {
		var EntryFan int32
		if k < Vec_IntSize(vFan) {
			EntryFan = Vec_IntEntry(vFan, k)
		} else {
			EntryFan = 0
		}
		var EntryFon int32
		if k < Vec_IntSize(vFon) {
			EntryFon = Vec_IntEntry(vFon, k)
		} else {
			EntryFon = 0
		}
		if EntryFan == 0 && EntryFon == 0 {
			continue
		}

		fmt.Fprintf(pFile, "%5d : ", k)
		if EntryFan == 0 {
			fmt.Fprintf(pFile, "              ")
		} else {
			fmt.Fprintf(pFile, "%12d  ", EntryFan)
		}
		fmt.Fprintf(pFile, "    ")
		if EntryFon == 0 {
			fmt.Fprintf(pFile, "              ")
		} else {
			fmt.Fprintf(pFile, "%12d  ", EntryFon)
		}

		fmt.Fprintf(pFile, "        ")
		if EntryFan == 0 {
			fmt.Fprintf(pFile, "              ")
		} else {
			fmt.Fprintf(pFile, "%12s  ", baseabc.Abc_ObjName(baseabc.Abc_NtkObj(pNtk, Vec_IntEntry(vFanR, k))))
		}
		fmt.Fprintf(pFile, "    ")
		if EntryFon == 0 {
			fmt.Fprintf(pFile, "              ")
		} else {
			fmt.Fprintf(pFile, "%12s  ", baseabc.Abc_ObjName(baseabc.Abc_NtkObj(pNtk, Vec_IntEntry(vFonR, k))))
		}
		fmt.Fprintf(pFile, "\n")
	}
}
func Abc_NtkPrintFanio(pFile *os.File, pNtk *Abc_Ntk_t, fUseFanio int32, fUsePio int32, fUseSupp int32, fUseCone int32) {
	var vFan *Vec_Int_t = Vec_IntAlloc(0)
	var vFon *Vec_Int_t = Vec_IntAlloc(0)
	var vFanR *Vec_Int_t = Vec_IntAlloc(0)
	var vFonR *Vec_Int_t = Vec_IntAlloc(0)
	assert(fUseFanio+fUsePio+fUseSupp+fUseCone == 1, "fUseFanio + fUsePio + fUseSupp + fUseCone == 1")
	if fUseFanio != 0 {
		Abc_NtkFaninFanoutCounters(pNtk, vFan, vFon, vFanR, vFonR)
		Abc_NtkPrintDistribInternal(pFile, pNtk, "fanins", "fanouts", "Nodes", "fanin", "fanout", vFan, vFon, vFanR, vFonR)

	} else if fUsePio != 0 {
		Abc_NtkInputOutputCounters(pNtk, vFan, vFon, vFanR, vFonR)
		Abc_NtkPrintDistribInternal(pFile, pNtk, "fanins", "fanouts", "I/O", "fanin", "fanout", vFan, vFon, vFanR, vFonR)

	} else if fUseSupp != 0 {
		Abc_NtkInOutSupportCounters(pNtk, vFan, vFon, vFanR, vFonR)
		Abc_NtkPrintDistribInternal(pFile, pNtk, "input supports", "output supports", "I/O", "in-supp", "out-supp", vFan, vFon, vFanR, vFonR)

	} else if fUseCone != 0 {
		Abc_NtkInOutConeCounters(pNtk, vFan, vFon, vFanR, vFonR)
		Abc_NtkPrintDistribInternal(pFile, pNtk, "input cones", "output cones", "I/O", "in-cone", "out-cone", vFan, vFon, vFanR, vFonR)
	}
	Vec_IntFree(vFan)
	Vec_IntFree(vFon)
	Vec_IntFree(vFanR)
	Vec_IntFree(vFonR)
}

/*
*Function*************************************************************

	Synopsis    [Prints the distribution of fanins/fanouts in the network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintFanioNew(pFile *os.File, pNtk *Abc_Ntk_t, fMffc int32) {
	nodes := pNtk.ObjectsOfKind(baseabc.ObjectNode)
	inBins, outBins := make([]int32, 70), make([]int32, 70)
	var maxIn, maxOut, sumIn, sumOut int32
	for _, node := range nodes {
		if fMffc != 0 && baseabc.Abc_ObjFanoutNum(node) == 1 {
			continue
		}
		ni := baseabc.Abc_ObjFaninNum(node)
		no := baseabc.Abc_ObjFanoutNum(node)
		if baseabc.Abc_NtkIsNetlist(pNtk) != 0 && baseabc.Abc_ObjFanoutNum(node) > 0 {
			no = baseabc.Abc_ObjFanoutNum(baseabc.Abc_ObjFanout0(node))
		} else if fMffc != 0 {
			no = baseabc.Abc_NodeMffcSize(node)
		}
		if ni > maxIn {
			maxIn = ni
		}
		if no > maxOut {
			maxOut = no
		}
		sumIn += ni
		sumOut += no
		abcPrintBin(inBins, ni)
		abcPrintBin(outBins, no)
	}
	size := 10 * (base10Log(maxIn) + 1)
	if other := 10 * (base10Log(maxOut) + 1); other > size {
		size = other
	}
	fmt.Fprintln(pFile, "The distribution of fanins and fanouts in the network:")
	fmt.Fprintln(pFile, "         Number   Nodes with fanin  Nodes with fanout")
	for k := 0; k < size; k++ {
		if inBins[k] == 0 && outBins[k] == 0 {
			continue
		}
		if k < 10 {
			fmt.Fprintf(pFile, "%15d : ", k)
		} else {
			power := int32(math.Pow10(int(k / 10)))
			fmt.Fprintf(pFile, "%15s : ", fmt.Sprintf("%d - %d", power*(int32(k%10)), power*(int32(k%10)+1)-1))
		}
		if inBins[k] == 0 {
			fmt.Fprintf(pFile, "              ")
		} else {
			fmt.Fprintf(pFile, "%12d  ", inBins[k])
		}
		fmt.Fprint(pFile, "    ")
		if outBins[k] == 0 {
			fmt.Fprintf(pFile, "              ")
		} else {
			fmt.Fprintf(pFile, "%12d  ", outBins[k])
		}
		fmt.Fprintln(pFile)
	}
	nodeCount := float64(baseabc.Abc_NtkNodeNum(pNtk))
	fmt.Fprintf(pFile, "Fanins: Max = %d. Ave = %.2f.  Fanouts: Max = %d. Ave =  %.2f.\n", maxIn, float64(sumIn)/nodeCount, maxOut, float64(sumOut)/nodeCount)
}
func abcPrintBin(bins []int32, value int32) {
	var index int32
	switch {
	case value < 10:
		index = value
	case value < 100:
		index = 10 + value/10
	case value < 1000:
		index = 20 + value/100
	case value < 10000:
		index = 30 + value/1000
	case value < 100000:
		index = 40 + value/10000
	case value < 1000000:
		index = 50 + value/100000
	case value < 10000000:
		index = 60 + value/1000000
	default:
		return
	}
	if index >= 0 && int(index) < len(bins) {
		bins[index]++
	}
}
func base10Log(n int32) int32 {
	if n < 2 {
		return n
	}
	var result int32
	for n--; n != 0; n /= 10 {
		result++
	}
	return result
}

/*
*Function*************************************************************

	Synopsis    [Prints the fanins/fanouts of a node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodePrintFanio(pFile *os.File, pNode *Abc_Obj_t) {
	if baseabc.Abc_ObjIsPo(pNode) != 0 {
		pNode = baseabc.Abc_ObjFanin0(pNode)
	}
	fmt.Fprintf(pFile, "Node %s\n", pNode.Name())
	fmt.Fprintf(pFile, "Fanins (%d): ", baseabc.Abc_ObjFaninNum(pNode))
	for _, edge := range pNode.Fanins() {
		fmt.Fprintf(pFile, " %s", edge.Object.Name())
	}
	fmt.Fprintln(pFile)
	fmt.Fprintf(pFile, "Fanouts (%d): ", baseabc.Abc_ObjFanoutNum(pNode))
	for _, obj := range pNode.Fanouts() {
		fmt.Fprintf(pFile, " %s", obj.Name())
	}
	fmt.Fprintln(pFile)
}

/*
*Function*************************************************************

	Synopsis    [Prints the MFFCs of the nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintMffc(pFile *os.File, pNtk *Abc_Ntk_t) {
	for _, node := range pNtk.ObjectsOfKind(baseabc.ObjectNode) {
		if baseabc.Abc_ObjFanoutNum(node) > 1 {
			baseabc.Abc_NodeMffcConeSuppPrint(node)
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Prints the factored form of one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintFactor(pFile *os.File, pNtk *Abc_Ntk_t, fUseRealNames int32) {
	assert(baseabc.Abc_NtkIsSopLogic(pNtk) != 0, "Abc_NtkIsSopLogic(pNtk)")
	for _, node := range pNtk.ObjectsOfKind(baseabc.ObjectNode) {
		Abc_NodePrintFactor(pFile, node, fUseRealNames)
	}
}
func Abc_NodePrintFactor(pFile *os.File, pNode *Abc_Obj_t, fUseRealNames int32) {
	if baseabc.Abc_ObjIsCo(pNode) != 0 {
		pNode = baseabc.Abc_ObjFanin0(pNode)
	}
	if baseabc.Abc_ObjIsPi(pNode) != 0 {
		fmt.Fprintln(pFile, "Skipping the PI node.")
		return
	}
	if baseabc.Abc_ObjIsLatch(pNode) != 0 {
		fmt.Fprintln(pFile, "Skipping the latch.")
		return
	}
	assert(baseabc.Abc_ObjIsNode(pNode) != 0, "Abc_ObjIsNode(pNode)")
	sop, ok := pNode.Data.(string)
	if !ok {
		fmt.Fprintln(pFile, "Cannot factor node without SOP data.")
		return
	}
	graph := dec.Dec_Factor(sop)
	if graph == nil {
		return
	}
	defer dec.Dec_GraphFree(graph)
	var names []string
	if fUseRealNames != 0 {
		for _, edge := range pNode.Fanins() {
			names = append(names, edge.Object.Name())
		}
	}
	dec.Dec_GraphPrint(os.Stdout, graph, names, pNode.Name())
}

/**Function*************************************************************

  Synopsis    [Prints the factored form of one node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Prints the level stats of the PO node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPrintLevel(pFile *os.File, pNtk *Abc_Ntk_t, fProfile, fListNodes, fVerbose int32) {
	if fListNodes != 0 {
		max := baseabc.Abc_NtkLevel(pNtk)
		fmt.Printf("Nodes by level:\n")
		for level := int32(0); level <= max; level++ {
			fmt.Printf("%2d : ", level)
			for _, node := range pNtk.ObjectsOfKind(baseabc.ObjectNode) {
				if int32(node.Level) == level {
					fmt.Printf(" %s", node.Name())
				}
			}
			fmt.Printf("\n")
		}
		return
	}
	if fProfile != 0 && baseabc.Abc_NtkHasMapping(pNtk) != 0 {
		const intervals = 12
		delayMax := Abc_NtkDelayTrace(pNtk, nil, nil, 0)
		delta := delayMax / intervals
		counts := make([]int32, intervals)
		for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
			co := baseabc.Abc_NtkCo(pNtk, i)
			driver := baseabc.Abc_ObjFanin0(co)
			bin := 0
			if !(baseabc.Abc_ObjIsNode(driver) != 0 && baseabc.Abc_ObjFaninNum(driver) == 0) && delta > 0 {
				bin = int(Abc_NodeReadArrivalWorst(driver) / delta)
				if bin >= intervals {
					bin = intervals - 1
				}
				if bin < 0 {
					bin = 0
				}
			}
			counts[bin]++
		}
		total := baseabc.Abc_NtkCoNum(pNtk)
		sum := int32(0)
		for i, count := range counts {
			sum += count
			percent := float64(0)
			if total > 0 {
				percent = 100 * float64(sum) / float64(total)
			}
			fmt.Printf("[%8.2f - %8.2f] :   COs = %4d.   %5.1f %%\n", delta*float32(i), delta*float32(i+1), count, percent)
		}
		return
	}
	if fProfile != 0 {
		if baseabc.Abc_NtkIsStrash(pNtk) == 0 {
			baseabc.Abc_NtkLevel(pNtk)
		}
		max := int32(0)
		for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
			level := int32(baseabc.Abc_ObjFanin0(baseabc.Abc_NtkCo(pNtk, i)).Level)
			if level > max {
				max = level
			}
		}
		counts := make([]int32, max+1)
		for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
			counts[baseabc.Abc_ObjFanin0(baseabc.Abc_NtkCo(pNtk, i)).Level]++
		}
		sum, total := int32(0), baseabc.Abc_NtkCoNum(pNtk)
		for level, count := range counts {
			if count == 0 {
				continue
			}
			sum += count
			percent := float64(0)
			if total > 0 {
				percent = 100 * float64(sum) / float64(total)
			}
			fmt.Printf("Level = %4d.  COs = %4d.   %5.1f %%\n", level, count, percent)
		}
		return
	}
	assert(baseabc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	if fVerbose == 0 {
		return
	}
	length := 5
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		if n := len(baseabc.Abc_NtkCo(pNtk, i).Name()); n > length {
			length = n
		}
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		co := baseabc.Abc_NtkCo(pNtk, i)
		fmt.Fprintf(pFile, "CO %4d :  %*s    ", i, length, co.Name())
		Abc_NodePrintLevel(pFile, co)
	}
}

/*
*Function*************************************************************

	Synopsis    [Prints the factored form of one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodePrintLevel(pFile *os.File, pNode *Abc_Obj_t) {
	driver := pNode
	if baseabc.Abc_ObjIsCo(pNode) != 0 {
		driver = baseabc.Abc_ObjFanin0(pNode)
	}
	switch {
	case baseabc.Abc_ObjIsPi(driver) != 0:
		fmt.Fprintln(pFile, "Primary input.")
		return
	case baseabc.Abc_ObjIsLatch(driver) != 0:
		fmt.Fprintln(pFile, "Latch.")
		return
	case baseabc.Abc_NodeIsConst(driver) != 0:
		value := baseabc.Abc_NodeIsConst1(driver)
		if len(pNode.Fanins()) > 0 && pNode.Fanins()[0].Complement {
			value = 1 - value
		}
		fmt.Fprintf(pFile, "Constant %d.\n", value)
		return
	}
	fmt.Fprintf(pFile, "Level = %3d.  ", driver.Level)
	fmt.Fprintf(pFile, "Mffc = %5d.  ", baseabc.Abc_NodeMffcSize(driver))
	nodes := baseabc.Abc_NtkDfsNodes(pNode.Network(), []*Abc_Obj_t{driver}, 1)
	fmt.Fprintf(pFile, "Cone = %5d.  \n", len(nodes))
}

/*
*Function*************************************************************

	Synopsis    [Prints the factored form of one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodePrintKMap(pNode *Abc_Obj_t, fUseRealNames int32) {
	manager, ok := pNode.Network().FunctionManager.(*core.DdManager)
	if !ok {
		return
	}
	on, ok := pNode.Data.(*core.DdNode)
	if !ok || on == nil {
		return
	}
	var names []string
	if fUseRealNames != 0 {
		for _, edge := range pNode.Fanins() {
			names = append(names, edge.Object.Name())
		}
	}
	extrab.Extra_PrintKMap(os.Stdout, manager, on, core.Cudd_Not(on), int(baseabc.Abc_ObjFaninNum(pNode)), nil, 0, names)
}

/*
*Function*************************************************************

	Synopsis    [Prints statistics about gates used in the network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintGates(pNtk *Abc_Ntk_t, fUseLibrary, fUpdateProfile int32) {
	if fUseLibrary != 0 && baseabc.Abc_NtkHasMapping(pNtk) != 0 {
		lib, ok := pNtk.FunctionManager.(*mio.Mio_Library_t)
		if !ok || lib == nil {
			return
		}
		gates := mio.Mio_LibraryReadGateArray(lib)
		for _, gate := range gates {
			mio.Mio_GateSetValue(gate, 0)
			if fUpdateProfile != 0 {
				mio.Mio_GateSetProfile2(gate, 0)
			}
		}
		nodes := pNtk.ObjectsOfKind(baseabc.ObjectNode)
		total := 0
		for i := 0; i < len(nodes); i++ {
			node := nodes[i]
			if i == 0 || node.Barrier {
				continue
			}
			gate, _ := node.Data.(*mio.Mio_Gate_t)
			if gate == nil {
				continue
			}
			mio.Mio_GateSetValue(gate, mio.Mio_GateReadValue(gate)+1)
			if fUpdateProfile != 0 {
				mio.Mio_GateIncProfile2(gate)
			}
			total++
			if Abc_NtkFetchTwinNode(node) != nil {
				i++
			}
		}
		nameLen := 5
		for _, gate := range gates {
			if mio.Mio_GateReadValue(gate) > 0 && len(mio.Mio_GateReadName(gate)) > nameLen {
				nameLen = len(mio.Mio_GateReadName(gate))
			}
		}
		areaTotal := baseabc.Abc_NtkGetMappedArea(pNtk)
		diff := int32(0)
		for _, gate := range gates {
			count := mio.Mio_GateReadValue(gate)
			profile := mio.Mio_GateReadProfile(gate)
			profile2 := mio.Mio_GateReadProfile2(gate)
			if count == 0 && profile == 0 {
				continue
			}
			if mio.Mio_GateReadPinNum(gate) > 1 {
				diff += abs32(profile - profile2)
			}
			area := float64(count) * mio.Mio_GateReadArea(gate)
			pct := float64(0)
			if areaTotal != 0 {
				pct = 100 * area / areaTotal
			}
			fmt.Printf("%-*s   Fanin = %2d   Instance = %8d   Area = %10.2f   %6.2f %%   %8d  %8d   %s\n", nameLen, mio.Mio_GateReadName(gate), mio.Mio_GateReadPinNum(gate), count, area, pct, profile, profile2, mio.Mio_GateReadForm(gate))
		}
		fmt.Printf("%-*s                Instance = %8d   Area = %10.2f   %6.2f %%   AbsDiff = %d\n", nameLen, "TOTAL", total, areaTotal, 100.0, diff)
		return
	}
	if baseabc.Abc_NtkIsAigLogic(pNtk) != 0 {
		return
	}
	wasBDD := baseabc.Abc_NtkIsBddLogic(pNtk) != 0
	if wasBDD && baseabc.Abc_NtkBddToSop(pNtk, -1, 1_000_000_000) == 0 {
		fmt.Println("Abc_NtkPrintGates(): Converting to SOPs has failed.")
		return
	}
	var constN, bufN, invN, andN, orN, other, total int32
	for i, node := range pNtk.ObjectsOfKind(baseabc.ObjectNode) {
		if i == 0 || node.Barrier {
			continue
		}
		var sop string
		if baseabc.Abc_NtkHasMapping(pNtk) != 0 {
			if gate, ok := node.Data.(*mio.Mio_Gate_t); ok {
				sop = mio.Mio_GateReadSop(gate)
			}
		} else {
			sop, _ = node.Data.(string)
		}
		switch {
		case baseabc.Abc_SopIsConst0(sop) != 0 || baseabc.Abc_SopIsConst1(sop) != 0:
			constN++
		case baseabc.Abc_SopIsBuf(sop) != 0:
			bufN++
		case baseabc.Abc_SopIsInv(sop) != 0:
			invN++
		case baseabc.Abc_SopIsComplement(sop) == 0 && baseabc.Abc_SopIsAndType(sop) != 0 || baseabc.Abc_SopIsComplement(sop) != 0 && baseabc.Abc_SopIsOrType(sop) != 0:
			andN++
		case baseabc.Abc_SopIsComplement(sop) != 0 && baseabc.Abc_SopIsAndType(sop) != 0 || baseabc.Abc_SopIsComplement(sop) == 0 && baseabc.Abc_SopIsOrType(sop) != 0:
			orN++
		default:
			other++
		}
		total++
	}
	percent := func(n int32) float64 {
		if total == 0 {
			return 0
		}
		return 100 * float64(n) / float64(total)
	}
	fmt.Printf("Const        = %8d    %6.2f %%\n", constN, percent(constN))
	fmt.Printf("Buffer       = %8d    %6.2f %%\n", bufN, percent(bufN))
	fmt.Printf("Inverter     = %8d    %6.2f %%\n", invN, percent(invN))
	fmt.Printf("And          = %8d    %6.2f %%\n", andN, percent(andN))
	fmt.Printf("Or           = %8d    %6.2f %%\n", orN, percent(orN))
	fmt.Printf("Other        = %8d    %6.2f %%\n", other, percent(other))
	fmt.Printf("TOTAL        = %8d    %6.2f %%\n", total, percent(total))
	if wasBDD {
		baseabc.Abc_NtkSopToBdd(pNtk)
	}
}
func abs32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

/*
*Function*************************************************************

	Synopsis    [Prints statistics about gates used in the network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintSharing(pNtk *Abc_Ntk_t) {
	cos := make([]*Abc_Obj_t, baseabc.Abc_NtkCoNum(pNtk))
	for i := range cos {
		cos[i] = baseabc.Abc_NtkCo(pNtk, int32(i))
	}
	fmt.Println("Statistics about sharing of logic nodes among the CO pairs.")
	fmt.Print("(CO1,CO2)=NumShared : ")
	for i, co1 := range cos {
		nodes1 := baseabc.Abc_NtkDfsNodes(pNtk, []*Abc_Obj_t{co1}, 1)
		for _, node := range nodes1 {
			node.MarkA = true
		}
		for k := i + 1; k < len(cos); k++ {
			nodes2 := baseabc.Abc_NtkDfsNodes(pNtk, []*Abc_Obj_t{cos[k]}, 1)
			count := 0
			for _, node := range nodes2 {
				if node.MarkA {
					count++
				}
			}
			fmt.Printf("(%d,%d)=%d ", i, k, count)
		}
		for _, node := range nodes1 {
			node.MarkA = false
		}
	}
	fmt.Println()
}
func Abc_NtkCountPis(vSupp *Vec_Ptr_t) int32 {
	var count int32
	for _, value := range vec.Vec_PtrArray(vSupp) {
		if node, ok := value.(*Abc_Obj_t); ok && baseabc.Abc_ObjIsPi(node) != 0 {
			count++
		}
	}
	return count
}

/*
*Function*************************************************************

	Synopsis    [Prints info for each output cone.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintStrSupports(pNtk *Abc_Ntk_t, fMatrix int32) {
	fmt.Println("Structural support info:")
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		co := baseabc.Abc_NtkCo(pNtk, i)
		supp := baseabc.Abc_NtkNodeSupport(pNtk, []*Abc_Obj_t{co}, 1)
		nodes := baseabc.Abc_NtkDfsNodes(pNtk, []*Abc_Obj_t{co}, 1)
		pis := int32(0)
		for _, obj := range supp {
			if baseabc.Abc_ObjIsPi(obj) != 0 {
				pis++
			}
		}
		fmt.Printf("%5d  %20s :  Cone = %5d.  Supp = %5d. (PIs = %5d. FFs = %5d.)\n", i, co.Name(), len(nodes), len(supp), pis, int32(len(supp))-pis)
	}
	if fMatrix == 0 {
		baseabc.Abc_NtkCleanMarkA(pNtk)
		return
	}
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		baseabc.Abc_NtkCi(pNtk, i).MarkA = false
	}
	fmt.Println("Actual support info:")
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		co := baseabc.Abc_NtkCo(pNtk, i)
		for _, obj := range baseabc.Abc_NtkNodeSupport(pNtk, []*Abc_Obj_t{co}, 1) {
			obj.MarkA = true
		}
		for k := int32(0); k < baseabc.Abc_NtkCiNum(pNtk); k++ {
			if baseabc.Abc_NtkCi(pNtk, k).MarkA {
				fmt.Print("1")
			} else {
				fmt.Print("0")
			}
		}
		fmt.Println()
		for k := int32(0); k < baseabc.Abc_NtkCiNum(pNtk); k++ {
			baseabc.Abc_NtkCi(pNtk, k).MarkA = false
		}
	}
	baseabc.Abc_NtkCleanMarkA(pNtk)
}

/*
*Function*************************************************************

	Synopsis    [Prints information about the object.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_ObjPrint(pFile *os.File, pObj *Abc_Obj_t) {
	fmt.Fprintf(pFile, "Object %5d : ", baseabc.Abc_ObjId(pObj))
	kind := map[baseabc.ObjectKind]string{baseabc.ObjectNone: "NONE   ", baseabc.ObjectConst1: "Const1 ", baseabc.ObjectPI: "PI     ", baseabc.ObjectPO: "PO     ", baseabc.ObjectBI: "BI     ", baseabc.ObjectBO: "BO     ", baseabc.ObjectNet: "Net    ", baseabc.ObjectNode: "Node   ", baseabc.ObjectLatch: "Latch  ", baseabc.ObjectWhitebox: "Whitebox", baseabc.ObjectBlackbox: "Blackbox"}
	label, ok := kind[pObj.Kind]
	assert(ok, "known object kind")
	fmt.Fprint(pFile, label, " Fanins ( ")
	for _, edge := range pObj.Fanins() {
		fmt.Fprintf(pFile, "%d ", baseabc.Abc_ObjId(edge.Object))
	}
	fmt.Fprint(pFile, ") ")
	if baseabc.Abc_ObjIsNode(pObj) != 0 && baseabc.Abc_NtkHasSop(pObj.Network()) != 0 {
		fmt.Fprintf(pFile, " %s", pObj.Data)
	} else if baseabc.Abc_ObjIsNode(pObj) != 0 && baseabc.Abc_NtkHasMapping(pObj.Network()) != 0 {
		if gate, ok := pObj.Data.(*mio.Mio_Gate_t); ok {
			fmt.Fprintf(pFile, " %s\n", mio.Mio_GateReadName(gate))
		} else {
			fmt.Fprintln(pFile)
		}
	} else {
		fmt.Fprintln(pFile)
	}
}

/*
*Function*************************************************************

	Synopsis    [Checks the status of the miter.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintMiter(pNtk *Abc_Ntk_t) {
	start := util.Abc_Clock()
	pis := int32(0)
	for i := int32(0); i < baseabc.Abc_NtkPiNum(pNtk); i++ {
		if baseabc.Abc_ObjFanoutNum(baseabc.Abc_NtkPi(pNtk, i)) > 0 {
			pis++
		}
	}
	unsat, sat, undec := int32(0), int32(0), int32(0)
	first := -1
	one := pNtk.One()
	for i := int32(0); i < baseabc.Abc_NtkPoNum(pNtk); i++ {
		po := baseabc.Abc_NtkPo(pNtk, i)
		edge := po.Fanins()[0]
		if edge.Object == one && edge.Complement {
			unsat++
		} else if edge.Object == one && !edge.Complement {
			sat++
			if first < 0 {
				first = int(i)
			}
		} else if baseabc.Abc_ObjIsPi(edge.Object) != 0 {
			sat++
			if first < 0 {
				first = int(i)
			}
		} else if edge.Object.Phase != edge.Complement {
			sat++
			if first < 0 {
				first = int(i)
			}
		} else {
			undec++
		}
	}
	fmt.Printf("Miter:  I =%6d", pis)
	fmt.Printf("  N =%7d", baseabc.Abc_NtkNodeNum(pNtk))
	fmt.Printf("  ? =%7d", undec)
	fmt.Printf("  U =%6d", unsat)
	fmt.Printf("  S =%6d", sat)
	fmt.Printf(" %7.2f sec\n", float64(util.Abc_Clock()-start)/float64(time.Second))
	if first >= 0 {
		po := baseabc.Abc_NtkPo(pNtk, int32(first))
		fmt.Printf("The first satisfiable output is number %d (%s).\n", first, po.Name())
	}
}

/*
*Function*************************************************************

	Synopsis    [Checks the status of the miter.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintPoEquivs(pNtk *Abc_Ntk_t) {
	mapping := make([]int, baseabc.Abc_NtkObjNumMax(pNtk))
	for i := range mapping {
		mapping[i] = -1
	}
	for i := int32(0); i < baseabc.Abc_NtkPoNum(pNtk); i++ {
		po := baseabc.Abc_NtkPo(pNtk, i)
		edge := po.Fanins()[0]
		driver := edge.Object
		if baseabc.Abc_NtkIsStrash(pNtk) != 0 && driver == pNtk.One() {
			fmt.Printf("%s = Const%d\n", po.Name(), boolInt(!edge.Complement))
			continue
		}
		if baseabc.Abc_NtkIsStrash(pNtk) == 0 && baseabc.Abc_NodeIsConst(driver) != 0 {
			fmt.Printf("%s = Const%d\n", po.Name(), baseabc.Abc_NodeIsConst1(driver))
			continue
		}
		id := baseabc.Abc_ObjId(driver)
		if mapping[id] < 0 {
			mapping[id] = int(i)
			continue
		}
		repr := baseabc.Abc_NtkCo(pNtk, int32(mapping[id]))
		reprEdge := repr.Fanins()[0]
		prefix := ""
		if reprEdge.Complement != edge.Complement {
			prefix = "!"
		}
		fmt.Printf("%s = %s%s\n", po.Name(), prefix, repr.Name())
	}
}

/*
*Function*************************************************************

	Synopsis    [Returns the percentable of increased power due to glitching.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMfsTotalGlitchingLut(pNtk *Abc_Ntk_t, nPats int32, Prob int32, fVerbose int32) float32 {
	if baseabc.Abc_NtkIsLogic(pNtk) == 0 {
		panic("Abc_NtkMfsTotalGlitchingLut requires logic network")
	}
	if baseabc.Abc_NtkGetFaninMax(pNtk) > 6 {
		fmt.Println("Abc_NtkMfsTotalGlitching() This procedure works only for mapped networks with LUTs size up to 6 inputs.")
		return -1
	}
	if baseabc.Abc_NtkToAig(pNtk) == 0 {
		return -1
	}
	manager, ok := pNtk.FunctionManager.(*hop.Hop_Man_t)
	if !ok {
		return -1
	}
	nodes := baseabc.Abc_NtkDfs(pNtk, 0)
	fanins := vec.Vec_IntAlloc(6)
	truthScratch := vec.Vec_IntAlloc(1 << 12)
	p := gia.Gli_ManAlloc(int32(len(nodes))+baseabc.Abc_NtkCiNum(pNtk)+baseabc.Abc_NtkCoNum(pNtk), baseabc.Abc_NtkLatchNum(pNtk), baseabc.Abc_NtkGetTotalFanins(pNtk)+baseabc.Abc_NtkCoNum(pNtk))
	defer gia.Gli_ManStop(p)
	defer vec.Vec_IntFree(fanins)
	defer vec.Vec_IntFree(truthScratch)
	for _, obj := range pNtk.Objects() {
		obj.Temp = -1
	}
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		obj := baseabc.Abc_NtkCi(pNtk, i)
		obj.Temp = gia.Gli_ManCreateCi(p, baseabc.Abc_ObjFanoutNum(obj))
	}
	truths := make([][]uint64, baseabc.Abc_NtkObjNumMax(pNtk))
	for _, obj := range nodes {
		vec.Vec_IntClear(fanins)
		for _, edge := range obj.Fanins() {
			vec.Vec_IntPush(fanins, int32(edge.Object.Temp))
		}
		root, ok := obj.Data.(*hop.Hop_Obj_t)
		if !ok {
			return -1
		}
		words := hop.Hop_ManConvertAigToTruth(manager, root, baseabc.Abc_ObjFaninNum(obj), truthScratch, 0)
		lo, hi := uint32(0), uint32(0)
		if len(words) > 0 {
			lo = words[0]
		}
		if baseabc.Abc_ObjFaninNum(obj) == 6 && len(words) > 1 {
			hi = words[1]
		}
		truths[baseabc.Abc_ObjId(obj)] = []uint64{uint64(hi)<<32 | uint64(lo)}
		obj.Temp = gia.Gli_ManCreateNode(p, fanins, baseabc.Abc_ObjFanoutNum(obj), truths[baseabc.Abc_ObjId(obj)])
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		co := baseabc.Abc_NtkCo(pNtk, i)
		gia.Gli_ManCreateCo(p, int32(baseabc.Abc_ObjFanin0(co).Temp))
	}
	gia.Gli_ManSwitchesAndGlitches(p, 4000, 1.0/8.0, 0)
	return abcGlitchRatio(pNtk, p)
}

/*
*Function*************************************************************

	Synopsis    [Returns the percentable of increased power due to glitching.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMfsTotalGlitching(pNtk *Abc_Ntk_t, nPats int32, Prob int32, fVerbose int32) float32 {
	if baseabc.Abc_NtkIsMappedLogic(pNtk) == 0 {
		return Abc_NtkMfsTotalGlitchingLut(pNtk, nPats, Prob, fVerbose)
	}
	if baseabc.Abc_NtkGetFaninMax(pNtk) > 16 {
		fmt.Println("Abc_NtkMfsTotalGlitching() This procedure works only for mapped networks with LUTs size up to 6 inputs.")
		return -1
	}
	nodes := baseabc.Abc_NtkDfs(pNtk, 0)
	fanins := vec.Vec_IntAlloc(6)
	p := gia.Gli_ManAlloc(int32(len(nodes))+baseabc.Abc_NtkCiNum(pNtk)+baseabc.Abc_NtkCoNum(pNtk), baseabc.Abc_NtkLatchNum(pNtk), baseabc.Abc_NtkGetTotalFanins(pNtk)+baseabc.Abc_NtkCoNum(pNtk))
	defer gia.Gli_ManStop(p)
	defer vec.Vec_IntFree(fanins)
	for _, obj := range pNtk.Objects() {
		obj.Temp = -1
	}
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		obj := baseabc.Abc_NtkCi(pNtk, i)
		obj.Temp = gia.Gli_ManCreateCi(p, baseabc.Abc_ObjFanoutNum(obj))
	}
	for _, obj := range nodes {
		vec.Vec_IntClear(fanins)
		for _, edge := range obj.Fanins() {
			vec.Vec_IntPush(fanins, int32(edge.Object.Temp))
		}
		gate, ok := obj.Data.(*mio.Mio_Gate_t)
		if !ok {
			return -1
		}
		obj.Temp = gia.Gli_ManCreateNode(p, fanins, baseabc.Abc_ObjFanoutNum(obj), mio.Mio_GateReadTruthP(gate))
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		co := baseabc.Abc_NtkCo(pNtk, i)
		gia.Gli_ManCreateCo(p, int32(baseabc.Abc_ObjFanin0(co).Temp))
	}
	prob := float32(1) / float32(Prob)
	gia.Gli_ManSwitchesAndGlitches(p, nPats, prob, fVerbose)
	return abcGlitchRatio(pNtk, p)
}

func abcGlitchRatio(n *Abc_Ntk_t, p *gia.Gli_Man_t) float32 {
	var switches, glitches int32
	for _, obj := range n.Objects() {
		if obj.Temp < 0 {
			continue
		}
		fanouts := baseabc.Abc_ObjFanoutNum(obj)
		switches += fanouts * gia.Gli_ObjNumSwitches(p, int32(obj.Temp))
		glitches += fanouts * gia.Gli_ObjNumGlitches(p, int32(obj.Temp))
	}
	if switches == 0 {
		return 0
	}
	return 100 * float32(glitches-switches) / float32(switches)
}

/*
*Function*************************************************************

	Synopsis    [Prints K-map of 6-var function represented by truth table.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
type abcPrintWord = uint64

func Abc_Show6VarFunc(F0, F1 uint64) {
	cells := [8]int{0, 4, 6, 2, 3, 7, 5, 1}
	if F0&F1 != 0 {
		panic("truth tables must be disjoint")
	}
	fmt.Print("     ")
	for _, h := range cells {
		fmt.Printf("%d%d%d ", (h>>2)&1, (h>>1)&1, h&1)
	}
	fmt.Println()
	fmt.Print("    +")
	for range cells {
		fmt.Print("---+")
	}
	fmt.Println()
	for _, v := range cells {
		fmt.Printf("%d%d%d |", (v>>2)&1, (v>>1)&1, v&1)
		for _, h := range cells {
			bit := uint((v * 8) + h)
			fmt.Print(" ")
			if (F0>>bit)&1 != 0 {
				fmt.Print("0")
			} else if (F1>>bit)&1 != 0 {
				fmt.Print("1")
			} else {
				fmt.Print(" ")
			}
			fmt.Print(" |")
		}
		fmt.Println()
		fmt.Print("    +")
		for j := range cells {
			fmt.Print("---")
			if j == len(cells)-1 {
				fmt.Print("+")
			} else {
				fmt.Print("|")
			}
		}
		fmt.Println()
	}
}

func Abc_NtkShow6VarFunc(pF0, pF1 string) {
	if len(pF0) != 16 {
		fmt.Printf("Wrong length (%d) of 6-var truth table.\n", len(pF0))
		return
	}
	if len(pF1) != 16 {
		fmt.Printf("Wrong length (%d) of 6-var truth table.\n", len(pF1))
		return
	}
	f0, e0 := strconv.ParseUint(pF0, 16, 64)
	f1, e1 := strconv.ParseUint(pF1, 16, 64)
	if e0 != nil || e1 != nil {
		fmt.Println("Invalid hexadecimal truth table.")
		return
	}
	Abc_Show6VarFunc(f0, f1)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
