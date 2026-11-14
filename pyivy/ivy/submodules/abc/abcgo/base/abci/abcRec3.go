package abci

import "fmt"

import "os"
import "sort"
import "encoding/binary"

import "github.com/glycerine/abcgo/misc/extra"
import "github.com/glycerine/abcgo/aig/gia"

func wordsToUint32(words []word) []uint32 {
	result := make([]uint32, len(words)*2)
	for i, value := range words {
		result[2*i] = uint32(value)
		result[2*i+1] = uint32(value >> 32)
	}
	return result
}

func lmsStretch5(words []word, nVars, nVarsOut int32) {
	parts := wordsToUint32(words)
	Abc_TtStretch5(parts, nVars, nVarsOut)
	for i := range words {
		words[i] = word(parts[2*i]) | word(parts[2*i+1])<<32
	}
}

func lmsCompareTruthWords(a, b []word) int32 {
	for i := 0; i < len(a) && i < len(b); i++ {
		var aa, bb [8]byte
		binary.NativeEndian.PutUint64(aa[:], uint64(a[i]))
		binary.NativeEndian.PutUint64(bb[:], uint64(b[i]))
		for j := range aa {
			if aa[j] != bb[j] {
				return int32(aa[j]) - int32(bb[j])
			}
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func lmsBoolInt32(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

/**CFile****************************************************************

  FileName    [abcRec2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Record of semi-canonical AIG subgraphs.]

  Author      [Allan Yang, Alan Mishchenko]

  Affiliation [Fudan University in Shanghai, UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcRec.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "map/if/if.h"
//#include "bool/kit/kit.h"
//#include "aig/gia/giaAig.h"
//#include "misc/vec/vecMem.h"
//#include "opt/dau/dau.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

const LMS_VAR_MAX = 16 // LMS_VAR_MAX >= 6
const LMS_MAX_WORD = 1 << (LMS_VAR_MAX - 6)

//#define LMS_USE_OLD_FORM

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/*
   This LMS manager can be used in two modes:
       - library constuction
       - AIG level minimization
   To switch from library construction to AIG level minimization
   LSM manager should be restarted by dumping GIA (rec_dump3 <file>.aig)
   and starting LMS manager again (rec_start3 <file>.aig).
*/

type Lms_Man_t struct {
	// parameters
	nVars      int32 // the number of variables
	nWords     int32 // the number of TT words
	nCuts      int32 // the max number of cuts to use
	fFuncOnly  int32 // record only functions
	fLibConstr int32 // this manager is used for library construction
	// internal data for library construction
	pGia   *Gia_Man_t // the record
	vTtMem *Vec_Mem_t // truth table memory and hash table
	//    Vec_Mem_t *       vTtMem2;      // truth table memory and hash table
	vTruthIds *Vec_Int_t // truth table IDs of each PO
	// internal data for AIG level minimization (allocated the first time it is called)
	vTruthPo    *Vec_Int_t // first PO where this canonicized truth table was seen
	vDelays     *Vec_Wrd_t // pin-to-pin delays of each PO
	vAreas      *Vec_Str_t // number of AND gates in each PO
	vFreqs      *Vec_Int_t // subgraph usage frequencies
	vTruthFreqs *Vec_Int_t // truth table usage frequencies
	// temporaries
	vNodes   *Vec_Ptr_t           // the temporary nodes
	vLabelsP *Vec_Ptr_t           // temporary storage for HOP node labels
	vLabels  *Vec_Int_t           // temporary storage for AIG node labels
	vSupps   *Vec_Str_t           // used temporarily by TT dumping
	word     pTemp1[LMS_MAX_WORD] // copy of the truth table
	word     pTemp2[LMS_MAX_WORD] // copy of the truth table
	// statistics
	nTried         int32
	nFilterSize    int32
	nFilterRedund  int32
	nFilterVolume  int32
	nFilterTruth   int32
	nFilterError   int32
	nFilterSame    int32
	nAdded         int32
	nAddedFuncs    int32
	nHoleInTheWall int32
	// runtime
	timeTruth  abctime
	timeCanon  abctime
	timeBuild  abctime
	timeCheck  abctime
	timeInsert abctime
	timeOther  abctime
	timeTotal  abctime
}

var s_pMan3 *Lms_Man_t = nil

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Compute delay/area profiles of POs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lms_DelayGet(D word, v int32) int32 {
	assert(v >= 0 && v < LMS_VAR_MAX, "v >= 0 && v < LMS_VAR_MAX")
	return int32((D >> (v << 2)) & 0xF)
}
func Lms_DelaySet(pD *word, v int32, d int32) {
	assert(v >= 0 && v < LMS_VAR_MAX, "v >= 0 && v < LMS_VAR_MAX")
	assert(d >= 0 && d < LMS_VAR_MAX, "d >= 0 && d < LMS_VAR_MAX")
	*pD |= word(d) << (v << 2)
}
func Lms_DelayInit(v int32) word {
	assert(v >= 0 && v < LMS_VAR_MAX, "v >= 0 && v < LMS_VAR_MAX")
	return word(1) << (v << 2)
}
func Lms_DelayMax(D1 word, D2 word, nVars int32) word {
	var v, Max int32
	var D word
	for v = 0; v < nVars; v++ {
		if Max = Abc_MaxInt(Lms_DelayGet(D1, v), Lms_DelayGet(D2, v)); Max != 0 {
			Lms_DelaySet(&D, v, Abc_MinInt(Max+1, 15))
		}
	}
	return D
}
func lmsChooseInt(cond bool, yes, no int32) int32 {
	if cond {
		return yes
	}
	return no
}
func lmsChooseFloat(cond bool, yes, no float64) float64 {
	if cond {
		return yes
	}
	return no
}
func Lms_DelayDecrement(D1 word, nVars int32) word {
	var v int32
	var D word
	for v = 0; v < nVars; v++ {
		if Lms_DelayGet(D1, v) != 0 {
			Lms_DelaySet(&D, v, Lms_DelayGet(D1, v)-1)
		}
	}
	return D
}
func Lms_DelayEqual(D1 word, D2 word, nVars int32) int32 { // returns 1 if D1 has the same delays than D2
	var v int32
	for v = 0; v < nVars; v++ {
		if Lms_DelayGet(D1, v) != Lms_DelayGet(D2, v) {
			return 0
		}
	}
	return 1
}
func Lms_DelayDom(D1 word, D2 word, nVars int32) int32 { // returns 1 if D1 has the same or smaller delays than D2
	var v int32
	for v = 0; v < nVars; v++ {
		if Lms_DelayGet(D1, v) > Lms_DelayGet(D2, v) {
			return 0
		}
	}
	return 1
}
func Lms_DelayPrint(D word, nVars int32) {
	var v int32
	fmt.Printf("Delay profile = {")
	for v = 0; v < nVars; v++ {
		fmt.Printf(" %d", Lms_DelayGet(D, v))
	}
	fmt.Printf(" }\n")
}
func Lms_GiaDelays(p *Gia_Man_t) *Vec_Wrd_t {
	var vDelays, vResult *Vec_Wrd_t
	var pObj *Gia_Obj_t
	var i int32
	// compute delay profiles of all objects
	vDelays = Vec_WrdAlloc(Gia_ManObjNum(p))
	Vec_WrdPush(vDelays, 0) // const 0
	for i = 1; i < Gia_ManObjNum(p); i++ {
		pObj = Gia_ManObj(p, i)
		if Gia_ObjIsAnd(pObj) != 0 {
			Vec_WrdPush(vDelays, Lms_DelayMax(Vec_WrdEntry(vDelays, Gia_ObjFaninId0(pObj, i)), Vec_WrdEntry(vDelays, Gia_ObjFaninId1(pObj, i)), Gia_ManCiNum(p)))
		} else if Gia_ObjIsCo(pObj) != 0 {
			Vec_WrdPush(vDelays, Lms_DelayDecrement(Vec_WrdEntry(vDelays, Gia_ObjFaninId0(pObj, i)), Gia_ManCiNum(p)))
		} else if Gia_ObjIsCi(pObj) != 0 {
			Vec_WrdPush(vDelays, Lms_DelayInit(Gia_ObjCioId(pObj)))
		} else {
			assert(false, "0")
		}
	}
	// collect delay profiles of COs only
	vResult = Vec_WrdAlloc(Gia_ManCoNum(p))
	for i = 0; i < Gia_ManCoNum(p); i++ {
		pObj = Gia_ManCo(p, i)
		Vec_WrdPush(vResult, Vec_WrdEntry(vDelays, Gia_ObjId(p, pObj)))
	}
	Vec_WrdFree(vDelays)
	return vResult
}
func Lms_ObjAreaMark_rec(pObj *Gia_Obj_t) {
	if gia.Gia_ObjMark0(pObj) != 0 || Gia_ObjIsCi(pObj) != 0 {
		return
	}
	gia.Gia_ObjSetMark0(pObj, 1)
	Lms_ObjAreaMark_rec(Gia_ObjFanin0(pObj))
	Lms_ObjAreaMark_rec(Gia_ObjFanin1(pObj))
}
func Lms_ObjAreaUnmark_rec(pObj *Gia_Obj_t) int32 {
	if gia.Gia_ObjMark0(pObj) == 0 || Gia_ObjIsCi(pObj) != 0 {
		return 0
	}
	gia.Gia_ObjSetMark0(pObj, 0)
	return 1 + Lms_ObjAreaUnmark_rec(Gia_ObjFanin0(pObj))
	+Lms_ObjAreaUnmark_rec(Gia_ObjFanin1(pObj))
}
func Lms_ObjArea(pObj *Gia_Obj_t) int32 {
	assert(Gia_ObjIsAnd(pObj) != 0, "Gia_ObjIsAnd(pObj)")
	Lms_ObjAreaMark_rec(pObj)
	return Lms_ObjAreaUnmark_rec(pObj)
}
func Lms_GiaAreas(p *Gia_Man_t) *Vec_Str_t {
	var vAreas *Vec_Str_t
	var pObj *Gia_Obj_t
	var i int32
	vAreas = Vec_StrAlloc(Gia_ManCoNum(p))
	for i = 0; i < Gia_ManCoNum(p); i++ {
		pObj = Gia_ManCo(p, i)
		Vec_StrPush(vAreas, byte(lmsChooseInt(Gia_ObjIsAnd(Gia_ObjFanin0(pObj)) != 0, Lms_ObjArea(Gia_ObjFanin0(pObj)), 0)))
	}
	return vAreas
}
func Lms_GiaSuppSizes(p *Gia_Man_t) *Vec_Str_t {
	var vResult *Vec_Str_t
	var vSupps *Vec_Str_t
	var pObj *Gia_Obj_t
	var i int32
	vSupps = Vec_StrAlloc(Gia_ManObjNum(p))
	Vec_StrPush(vSupps, 0)
	for i = 1; i < Gia_ManObjNum(p); i++ {
		pObj = Gia_ManObj(p, i)
		if Gia_ObjIsAnd(pObj) != 0 {
			Vec_StrPush(vSupps, byte(Abc_MaxInt(Vec_StrEntry(vSupps, Gia_ObjFaninId0(pObj, i)), Vec_StrEntry(vSupps, Gia_ObjFaninId1(pObj, i)))))
		} else if Gia_ObjIsCo(pObj) != 0 {
			Vec_StrPush(vSupps, Vec_StrEntry(vSupps, Gia_ObjFaninId0(pObj, i)))
		} else if Gia_ObjIsCi(pObj) != 0 {
			Vec_StrPush(vSupps, byte((Gia_ObjCioId(pObj) + 1)))
		} else {
			assert(false, "0")
		}
	}
	assert(Vec_StrSize(vSupps) == Gia_ManObjNum(p), "Vec_StrSize(vSupps) == Gia_ManObjNum(p)")
	vResult = Vec_StrAlloc(Gia_ManCoNum(p))
	for i = 0; i < Gia_ManCoNum(p); i++ {
		pObj = Gia_ManCo(p, i)
		Vec_StrPush(vResult, Vec_StrEntry(vSupps, Gia_ObjId(p, pObj)))
	}
	Vec_StrFree(vSupps)
	return vResult
}
func Lms_GiaProfilesPrint(p *Gia_Man_t) {
	var pObj *Gia_Obj_t
	var i int32
	var vDelays *Vec_Wrd_t
	var vAreas *Vec_Str_t
	vDelays = Lms_GiaDelays(p)
	vAreas = Lms_GiaAreas(p)
	for i = 0; i < Gia_ManPoNum(p); i++ {
		pObj = Gia_ManPo(p, i)
		fmt.Printf("%6d : ", i)
		fmt.Printf("A = %2d  ", Vec_StrEntry(vAreas, i))
		Lms_DelayPrint(Vec_WrdEntry(vDelays, i), Gia_ManPiNum(p))
		//        Lms_GiaPrintSubgraph( p, pObj );
		//        printf( "\n" );
	}
	Vec_WrdFree(vDelays)
	Vec_StrFree(vAreas)
}

/*
*Function*************************************************************

	Synopsis    [Prints one GIA subgraph.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lms_GiaPrintSubgraph_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
	if gia.Gia_ObjMark0(pObj) == 0 || Gia_ObjIsCi(pObj) != 0 {
		return
	}
	gia.Gia_ObjSetMark0(pObj, 0)
	assert(Gia_ObjIsAnd(pObj) != 0, "Gia_ObjIsAnd(pObj)")
	Lms_GiaPrintSubgraph_rec(p, Gia_ObjFanin0(pObj))
	Lms_GiaPrintSubgraph_rec(p, Gia_ObjFanin1(pObj))
	Gia_ObjPrint(p, pObj)
}
func Lms_GiaPrintSubgraph(p *Gia_Man_t, pObj *Gia_Obj_t) {
	assert(Gia_ObjIsCo(pObj) != 0, "Gia_ObjIsCo(pObj)")
	if Gia_ObjIsAnd(Gia_ObjFanin0(pObj)) != 0 {
		Lms_ObjAreaMark_rec(Gia_ObjFanin0(pObj))
		Lms_GiaPrintSubgraph_rec(p, Gia_ObjFanin0(pObj))

	} else {
		Gia_ObjPrint(p, Gia_ObjFanin0(pObj))
	}
	Gia_ObjPrint(p, pObj)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lms_ManStart(pGia *Gia_Man_t, nVars int32, nCuts int32, fFuncOnly int32, fVerbose int32) *Lms_Man_t {
	var p *Lms_Man_t
	var clk abctime
	var clk2 abctime = Abc_Clock()
	// if GIA is given, use the number of variables from GIA
	if pGia != nil {
		nVars = Gia_ManCiNum(pGia)
	} else {
		nVars = nVars
	}
	assert(nVars >= 6 && nVars <= LMS_VAR_MAX, "nVars >= 6 && nVars <= LMS_VAR_MAX")
	// allocate manager
	p = new(Lms_Man_t)
	// parameters
	p.nVars = nVars
	p.nCuts = nCuts
	p.nWords = Abc_Truth6WordNum(nVars)
	p.fFuncOnly = fFuncOnly
	// internal data for library construction
	p.vTtMem = Vec_MemAlloc(p.nWords, 12) // 32 KB/page for 6-var functions
	//    p->vTtMem2 = Vec_MemAlloc( p->nWords, 12 ); // 32 KB/page for 6-var functions
	Vec_MemHashAlloc(p.vTtMem, 10000)
	//    Vec_MemHashAlloc( p->vTtMem2, 10000 );
	if fFuncOnly != 0 {
		return p
	}
	p.vTruthIds = Vec_IntAlloc(10000)
	if pGia == nil {
		var i int32
		p.pGia = Gia_ManStart(10000)
		gia.Gia_ManSetNameSpec(p.pGia, "record", "")
		for i = 0; i < nVars; i++ {
			Gia_ManAppendCi(p.pGia)
		}

	} else {
		var pObj *Gia_Obj_t
		var pTruth []word
		var i int32
		var Index int32
		var Prev int32 = -1
		p.pGia = pGia
		// populate the manager with subgraphs present in GIA
		p.nAdded = Gia_ManCoNum(p.pGia)
		for i = 0; i < Gia_ManCoNum(p.pGia); i++ {
			pObj = Gia_ManCo(p.pGia, i)
			clk = Abc_Clock()
			pTruth = Gia_ObjComputeTruthTable(p.pGia, pObj)
			p.timeTruth += Abc_Clock() - clk
			clk = Abc_Clock()
			Index = Vec_MemHashInsert(p.vTtMem, pTruth)
			p.timeInsert += Abc_Clock() - clk
			assert(Index == Prev || Index == Prev+1, "Index == Prev || Index == Prev + 1") // GIA subgraphs should be ordered
			Vec_IntPush(p.vTruthIds, Index)
			Prev = Index
		}
	}
	// temporaries
	p.vNodes = Vec_PtrAlloc(1000)
	p.vLabelsP = Vec_PtrAlloc(1000)
	p.vLabels = Vec_IntAlloc(1000)
	p.timeTotal += Abc_Clock() - clk2
	return p
}
func Lms_ManStop(p *Lms_Man_t) {
	// temporaries
	Vec_IntFreeP(&p.vLabels)
	Vec_PtrFreeP(&p.vLabelsP)
	Vec_PtrFreeP(&p.vNodes)
	// internal data for AIG level minimization
	Vec_IntFreeP(&p.vTruthPo)
	Vec_WrdFreeP(&p.vDelays)
	Vec_StrFreeP(&p.vAreas)
	Vec_IntFreeP(&p.vFreqs)
	Vec_IntFreeP(&p.vTruthFreqs)
	// internal data for library construction
	Vec_IntFreeP(&p.vTruthIds)
	Vec_MemHashFree(p.vTtMem)
	//    Vec_MemHashFree( p->vTtMem2 );
	Vec_MemFree(p.vTtMem)
	//    Vec_MemFree( p->vTtMem2 );
	Gia_ManStopP(&p.pGia)
	ABC_FREE(p)
}
func Lms_ManPrepare(p *Lms_Man_t) {
	// compute the first PO for each semi-canonical form
	var i, Entry int32
	assert(p.fLibConstr == 0, "!p.fLibConstr")
	assert(p.vTruthPo == nil, "p.vTruthPo == nil")
	p.vTruthPo = Vec_IntStartFull(Vec_MemEntryNum(p.vTtMem) + 1)
	assert(Vec_IntFindMin(p.vTruthIds) >= 0, "Vec_IntFindMin(p.vTruthIds) >= 0")
	assert(Vec_IntFindMax(p.vTruthIds) < Vec_MemEntryNum(p.vTtMem), "Vec_IntFindMax(p.vTruthIds) < Vec_MemEntryNum(p.vTtMem)")
	for i = 0; i < Vec_IntSize(p.vTruthIds); i++ {
		Entry = Vec_IntEntry(p.vTruthIds, i)
		if Vec_IntEntry(p.vTruthPo, Entry) == -1 {
			Vec_IntWriteEntry(p.vTruthPo, Entry, i)
		}
	}
	Vec_IntWriteEntry(p.vTruthPo, Vec_MemEntryNum(p.vTtMem), Gia_ManCoNum(p.pGia))
	// compute delay/area and init frequency
	assert(p.vDelays == nil, "p.vDelays == nil")
	assert(p.vAreas == nil, "p.vAreas == nil")
	assert(p.vFreqs == nil, "p.vFreqs == nil")
	p.vDelays = Lms_GiaDelays(p.pGia)
	p.vAreas = Lms_GiaAreas(p.pGia)
	p.vFreqs = Vec_IntStart(Gia_ManCoNum(p.pGia))
}
func Lms_ManPrintFuncStats(p *Lms_Man_t) {
	var vSupps *Vec_Str_t
	var Counters [LMS_VAR_MAX + 1]int32
	var CountersS [LMS_VAR_MAX + 1]int32
	var i, Entry, Next int32
	if p.pGia == nil {
		return
	}
	if p.fLibConstr != 0 {
		return
	}
	if p.vTruthPo == nil {
		Lms_ManPrepare(p)
	}
	vSupps = Lms_GiaSuppSizes(p.pGia)
	for i = 0; i < Vec_IntSize(p.vTruthPo); i++ {
		Entry = Vec_IntEntry(p.vTruthPo, i)
		if i == Vec_IntSize(p.vTruthPo)-1 {
			break
		}
		Next = Vec_IntEntry(p.vTruthPo, i+1)
		Counters[int32(Vec_StrEntry(vSupps, Entry))]++
		CountersS[int32(Vec_StrEntry(vSupps, Entry))] += Next - Entry
	}
	for i = 0; i <= LMS_VAR_MAX; i++ {
		if Counters[i] != 0 {
			fmt.Printf("Inputs = %2d.  Funcs = %8d.  Subgrs = %8d.  Ratio = %6.2f.\n", i, Counters[i], CountersS[i], 1.0*CountersS[i]/Counters[i])
		}
	}
	Vec_StrFree(vSupps)
}
func Lms_ManPrintFreqStats(p *Lms_Man_t) {
	var CountDsdNpn [3]int32  // full/part/none
	var CountDsdAll [3]int32  // full/part/none
	var CountStepNpn [3]int32 // full/1step/complex
	var CountStepAll [3]int32 // full/1step/complex
	var pBuffer [1000]byte
	var nSuppSize int32
	var nNonDecSize int32
	var pTruth []word
	var i, Freq, Status int32
	fmt.Printf("Cuts  = %10d. ", p.nTried)
	//    printf( "Funcs = %10d (%6.2f %%). ", Vec_MemEntryNum(p->vTtMem2), 100.0*Vec_MemEntryNum(p->vTtMem2)/p->nTried );
	fmt.Printf("Class = %10d (%6.2f %%). ", Vec_MemEntryNum(p.vTtMem), 100.0*Vec_MemEntryNum(p.vTtMem)/p.nTried)
	fmt.Printf("\n")
	//    return;

	for i = 0; i < Vec_IntSize(p.vTruthFreqs); i++ {
		Freq = Vec_IntEntry(p.vTruthFreqs, i)
		pTruth = Vec_MemReadEntry(p.vTtMem, i)
		/*
		   printf( "%6d -- %6d : ", i, Freq );
		   Kit_DsdWriteFromTruth( pBuffer, (unsigned *)pTruth, p->nVars );
		   printf( "%s\n", pBuffer );
		*/
		nSuppSize = Abc_TtSupportSize(pTruth, p.nVars)
		nNonDecSize = Dau_DsdDecompose(pTruth, p.nVars, 0, 0, &pBuffer[:])
		if nNonDecSize == 0 {
			CountDsdNpn[0]++
			CountDsdAll[0] += Freq

		} else if nNonDecSize < nSuppSize {
			CountDsdNpn[1]++
			CountDsdAll[1] += Freq

		} else { // non-dec

			CountDsdNpn[2]++
			CountDsdAll[2] += Freq
		}

		if nNonDecSize == 0 {
			CountStepNpn[0]++
			CountStepAll[0] += Freq
			continue
		}

		// check the non dec core
		Status = Dau_DsdCheck1Step(nil, pTruth, nNonDecSize, nil)
		if Status >= 0 {
			CountStepNpn[1]++
			CountStepAll[1] += Freq

		} else {
			assert(Status == -2, "Status == -2")
			CountStepNpn[2]++
			CountStepAll[2] += Freq
		}
	}

	// print the results
	fmt.Printf("NPN: ")
	fmt.Printf("Full = %6.2f %%  ", 100.0*CountDsdNpn[0]/Vec_MemEntryNum(p.vTtMem))
	fmt.Printf("Part = %6.2f %%  ", 100.0*CountDsdNpn[1]/Vec_MemEntryNum(p.vTtMem))
	fmt.Printf("None = %6.2f %%  ", 100.0*CountDsdNpn[2]/Vec_MemEntryNum(p.vTtMem))
	//    printf( "\n" );
	fmt.Printf("   ")
	// print the results
	fmt.Printf("All: ")
	fmt.Printf("Full = %6.2f %%  ", 100.0*CountDsdAll[0]/p.nTried)
	fmt.Printf("Part = %6.2f %%  ", 100.0*CountDsdAll[1]/p.nTried)
	fmt.Printf("None = %6.2f %%  ", 100.0*CountDsdAll[2]/p.nTried)
	fmt.Printf("\n")

	// print the results
	fmt.Printf("NPN: ")
	fmt.Printf("Full = %6.2f %%  ", 100.0*CountStepNpn[0]/Vec_MemEntryNum(p.vTtMem))
	fmt.Printf("1stp = %6.2f %%  ", 100.0*CountStepNpn[1]/Vec_MemEntryNum(p.vTtMem))
	fmt.Printf("Comp = %6.2f %%  ", 100.0*CountStepNpn[2]/Vec_MemEntryNum(p.vTtMem))
	//    printf( "\n" );
	fmt.Printf("   ")
	// print the results
	fmt.Printf("All: ")
	fmt.Printf("Full = %6.2f %%  ", 100.0*CountStepAll[0]/p.nTried)
	fmt.Printf("1stp = %6.2f %%  ", 100.0*CountStepAll[1]/p.nTried)
	fmt.Printf("Comp = %6.2f %%  ", 100.0*CountStepAll[2]/p.nTried)
	fmt.Printf("\n")

}
func Lms_ManPrint(p *Lms_Man_t) {
	//    Gia_ManPrintStats( p->pGia, 0, 0 );
	fmt.Printf("Library with %d vars has %d classes and %d AIG subgraphs with %d AND nodes.\n", p.nVars, Vec_MemEntryNum(p.vTtMem), p.nAdded, lmsChooseInt(p.pGia != nil, Gia_ManAndNum(p.pGia), 0))

	//    Lms_ManPrintFreqStats( p );
	Lms_ManPrintFuncStats(p)

	p.nAddedFuncs = Vec_MemEntryNum(p.vTtMem)
	fmt.Printf("Subgraphs tried                             = %10d. (%6.2f %%)\n", p.nTried, lmsChooseFloat(p.nTried != 0, 100.0*float64(p.nTried)/float64(p.nTried), 0))
	fmt.Printf("Subgraphs filtered by support size          = %10d. (%6.2f %%)\n", p.nFilterSize, lmsChooseFloat(p.nTried != 0, 100.0*float64(p.nFilterSize)/float64(p.nTried), 0))
	fmt.Printf("Subgraphs filtered by structural redundancy = %10d. (%6.2f %%)\n", p.nFilterRedund, lmsChooseFloat(p.nTried != 0, 100.0*float64(p.nFilterRedund)/float64(p.nTried), 0))
	fmt.Printf("Subgraphs filtered by volume                = %10d. (%6.2f %%)\n", p.nFilterVolume, lmsChooseFloat(p.nTried != 0, 100.0*float64(p.nFilterVolume)/float64(p.nTried), 0))
	fmt.Printf("Subgraphs filtered by TT redundancy         = %10d. (%6.2f %%)\n", p.nFilterTruth, lmsChooseFloat(p.nTried != 0, 100.0*float64(p.nFilterTruth)/float64(p.nTried), 0))
	fmt.Printf("Subgraphs filtered by error                 = %10d. (%6.2f %%)\n", p.nFilterError, lmsChooseFloat(p.nTried != 0, 100.0*float64(p.nFilterError)/float64(p.nTried), 0))
	fmt.Printf("Subgraphs filtered by isomorphism           = %10d. (%6.2f %%)\n", p.nFilterSame, lmsChooseFloat(p.nTried != 0, 100.0*float64(p.nFilterSame)/float64(p.nTried), 0))
	fmt.Printf("Subgraphs added                             = %10d. (%6.2f %%)\n", p.nAdded, lmsChooseFloat(p.nTried != 0, 100.0*float64(p.nAdded)/float64(p.nTried), 0))
	fmt.Printf("Functions added                             = %10d. (%6.2f %%)\n", p.nAddedFuncs, lmsChooseFloat(p.nTried != 0, 100.0*float64(p.nAddedFuncs)/float64(p.nTried), 0))
	if p.nHoleInTheWall != 0 {
		fmt.Printf("Cuts whose logic structure has a hole       = %10d. (%6.2f %%)\n", p.nHoleInTheWall, lmsChooseFloat(p.nTried != 0, 100.0*float64(p.nHoleInTheWall)/float64(p.nTried), 0))
	}

	p.timeOther = p.timeTotal - p.timeTruth - p.timeCanon - p.timeBuild - p.timeCheck - p.timeInsert
	ABC_PRTP("Runtime: Truth ", p.timeTruth, p.timeTotal)
	ABC_PRTP("Runtime: Canon ", p.timeCanon, p.timeTotal)
	ABC_PRTP("Runtime: Build ", p.timeBuild, p.timeTotal)
	ABC_PRTP("Runtime: Check ", p.timeCheck, p.timeTotal)
	ABC_PRTP("Runtime: Insert", p.timeInsert, p.timeTotal)
	ABC_PRTP("Runtime: Other ", p.timeOther, p.timeTotal)
	ABC_PRTP("Runtime: TOTAL ", p.timeTotal, p.timeTotal)
}

/*
*Function*************************************************************

	Synopsis    [Recanonicizes the library and add it to the current library.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkRecLibMerge3(pLib *Gia_Man_t) {
	var fCheck int32
	var p *Lms_Man_t = s_pMan3
	var pGia *Gia_Man_t = p.pGia
	var vSupps *Vec_Str_t
	var pCanonPerm [LMS_VAR_MAX]byte
	var uCanonPhase uint32
	var pTruth []word
	var i, k, Index, iFanin0, iFanin1, nLeaves int32
	var pObjPo *Gia_Obj_t
	var pDriver *Gia_Obj_t
	var pTemp *Gia_Obj_t = nil
	var clk abctime
	var clk2 abctime = Abc_Clock()

	if Gia_ManCiNum(pLib) != Gia_ManCiNum(pGia) {
		fmt.Printf("The number of Library inputs (%d) differs from the number of Gia inputs (%d).\n", Gia_ManCiNum(pLib), Gia_ManCiNum(pGia))
		return
	}
	assert(Gia_ManCiNum(pLib) == Gia_ManCiNum(pGia), "Gia_ManCiNum(pLib) == Gia_ManCiNum(pGia)")

	// create hash table if not available
	if gia.Gia_ManHashStarted(pGia) == 0 {
		Gia_ManHashStart(pGia)
	}

	// add AIG subgraphs
	vSupps = Lms_GiaSuppSizes(pLib)
	for k = 0; k < Gia_ManCoNum(pLib); k++ {
		pObjPo = Gia_ManCo(pLib, k)
		// get support size
		nLeaves = Vec_StrEntry(vSupps, k)
		assert(nLeaves > 1, "nLeaves > 1")

		// compute the truth table
		clk = Abc_Clock()
		pTruth = Gia_ObjComputeTruthTable(pLib, Gia_ObjFanin0(pObjPo))
		p.timeTruth += Abc_Clock() - clk
		// semi-canonicize
		clk = Abc_Clock()
		copy(p.pTemp1[:], pTruth)
		uCanonPhase = Abc_TtCanonicize(p.pTemp1[:], nLeaves, pCanonPerm[:])
		lmsStretch5(p.pTemp1[:], nLeaves, p.nVars)
		p.timeCanon += Abc_Clock() - clk
		// pCanonPerm and uCanonPhase show what was the variable corresponding to each var in the current truth
		if nLeaves == 2 && Abc_TtSupportSize(pTruth, 2) != 2 {
			continue
		}

		clk = Abc_Clock()
		// map cut leaves into elementary variables of GIA
		for i = 0; i < nLeaves; i++ {
			Gia_ManCi(pLib, pCanonPerm[i]).Value = Abc_Var2Lit(Gia_ObjId(pGia, Gia_ManPi(pGia, i)), (uCanonPhase>>i)&1)
		}
		// build internal nodes
		vTtNodes := gia.Gia_ManTruthTableNodeIDs(pLib)
		assert(len(vTtNodes) > 0, "truth-table nodes are available")
		for i = 0; i < int32(len(vTtNodes)); i++ {
			pTemp = Gia_ManObj(pLib, vTtNodes[i])
			iFanin0 = Abc_LitNotCond(Gia_ObjFanin0(pTemp).Value, Gia_ObjFaninC0(pTemp))
			iFanin1 = Abc_LitNotCond(Gia_ObjFanin1(pTemp).Value, Gia_ObjFaninC1(pTemp))
			pTemp.Value = Gia_ManHashAnd(pGia, iFanin0, iFanin1)
		}
		p.timeBuild += Abc_Clock() - clk

		// check if this node is already driving a PO
		assert(Gia_ObjIsAnd(pTemp) != 0, "Gia_ObjIsAnd(pTemp)")
		pDriver = Gia_ManObj(pGia, Abc_Lit2Var(pTemp.Value))
		if gia.Gia_ObjMark1(pDriver) != 0 {
			p.nFilterSame++
			continue
		}
		gia.Gia_ObjSetMark1(pDriver, 1)
		// create output
		Gia_ManAppendCo(pGia, Abc_LitNotCond(pTemp.Value, (uCanonPhase>>nLeaves)&1))

		// verify truth table
		if fCheck != 0 {
			clk = Abc_Clock()
			pTemp = Gia_ManCo(pGia, Gia_ManCoNum(pGia)-1)
			pTruth = Gia_ObjComputeTruthTable(pGia, Gia_ManCo(pGia, Gia_ManCoNum(pGia)-1))
			p.timeCheck += Abc_Clock() - clk
			if lmsCompareTruthWords(p.pTemp1[:p.nWords], pTruth[:p.nWords]) != 0 {

				Kit_DsdPrintFromTruth(wordsToUint32(pTruth), nLeaves)
				fmt.Printf("\n")
				Kit_DsdPrintFromTruth(wordsToUint32(p.pTemp1[:]), nLeaves)
				fmt.Printf("\n")
				fmt.Printf("Truth table verification has failed.\n")

				// drive PO with constant
				Gia_ManPatchCoDriver(pGia, Gia_ManCoNum(pGia)-1, 0)
				// save truth table ID
				Vec_IntPush(p.vTruthIds, -1)
				p.nFilterTruth++
				continue
			}
		}

		clk = Abc_Clock()
		// add the resulting truth table to the hash table
		Index = Vec_MemHashInsert(p.vTtMem, p.pTemp1)
		// save truth table ID
		Vec_IntPush(p.vTruthIds, Index)
		assert(Gia_ManCoNum(pGia) == Vec_IntSize(p.vTruthIds), "Gia_ManCoNum(pGia) == Vec_IntSize(p.vTruthIds)")
		p.nAdded++
		p.timeInsert += Abc_Clock() - clk
	}
	Vec_StrFree(vSupps)
	p.timeTotal += Abc_Clock() - clk2
}

/*
*Function*************************************************************

	Synopsis    [Evaluates one cut during library construction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkRecAddCut3(pIfMan *If_Man_t, pRoot *If_Obj_t, pCut *If_Cut_t) int32 {
	var p *Lms_Man_t = s_pMan3
	var pCanonPerm [LMS_VAR_MAX]byte
	var uCanonPhase uint32
	var i, Index, iFanin0, iFanin1, fHole int32
	var nLeaves int32 = If_CutLeaveNum(pCut)
	var vNodes *Vec_Ptr_t = p.vNodes
	var pGia *Gia_Man_t = p.pGia
	var pDriver *Gia_Obj_t
	var pIfObj *If_Obj_t = nil
	var pTruth []word
	var clk abctime
	p.nTried++

	// skip small cuts
	assert(p.nVars == int32(pCut.nLimit), "p.nVars == int32(pCut.nLimit)")
	if nLeaves < 2 || (nLeaves == 2 && Abc_TtSupportSize(If_CutTruthW(pIfMan, pCut), 2) != 2) {
		p.nFilterSize++
		return 1
	}

	//    if ( p->vTtMem2 )
	//        Vec_MemHashInsert( p->vTtMem2, If_CutTruthW(pCut) );

	// semi-canonicize truth table
	clk = Abc_Clock()
	copy(p.pTemp1[:], If_CutTruthW(pIfMan, pCut))
	uCanonPhase = Abc_TtCanonicize(p.pTemp1[:], nLeaves, pCanonPerm[:])
	lmsStretch5(p.pTemp1[:], nLeaves, p.nVars)
	p.timeCanon += Abc_Clock() - clk
	// pCanonPerm and uCanonPhase show what was the variable corresponding to each var in the current truth

	if p.pGia == nil {
		clk = Abc_Clock()
		// add the resulting truth table to the hash table
		Index = Vec_MemHashInsert(p.vTtMem, p.pTemp1)
		/*
		   if ( p->vTruthFreqs == NULL )
		       p->vTruthFreqs = Vec_IntAlloc( 1000 );
		   assert( Index <= Vec_IntSize(p->vTruthFreqs)  );
		   if ( Index < Vec_IntSize(p->vTruthFreqs) )
		       Vec_IntAddToEntry( p->vTruthFreqs, Index, 1 );
		   else
		       Vec_IntPush( p->vTruthFreqs, 1 );
		*/
		p.nAdded++
		p.timeInsert += Abc_Clock() - clk
		return 1
	}

	// collect internal nodes and skip redundant cuts
	clk = Abc_Clock()
	If_CutTraverse(pIfMan, pRoot, pCut, vNodes)
	p.timeTruth += Abc_Clock() - clk
	if Vec_PtrSize(vNodes) > 253 {
		p.nFilterSize++
		return 1
	}

	clk = Abc_Clock()
	// map cut leaves into elementary variables of GIA
	for i = 0; i < nLeaves; i++ {
		If_ManObj(pIfMan, pCut.pLeaves[int32(pCanonPerm[i])]).iCopy = Abc_Var2Lit(Gia_ObjId(pGia, Gia_ManPi(pGia, i)), (uCanonPhase>>i)&1)
	}
	// build internal nodes
	fHole = 0
	assert(Vec_PtrSize(vNodes) > 0, "Vec_PtrSize(vNodes) > 0")
	for i = nLeaves; i < int32(Vec_PtrSize(vNodes)); i++ {
		pIfObj = Vec_PtrEntry(vNodes, int(i)).(*If_Obj_t)
		if If_ObjIsCi(pIfObj) != 0 {
			pIfObj.iCopy = 0
			fHole = 1
			continue
		}
		iFanin0 = Abc_LitNotCond(If_ObjFanin0(pIfObj).iCopy, If_ObjFaninC0(pIfObj))
		iFanin1 = Abc_LitNotCond(If_ObjFanin1(pIfObj).iCopy, If_ObjFaninC1(pIfObj))
		pIfObj.iCopy = Gia_ManHashAnd(pGia, iFanin0, iFanin1)
	}
	p.nHoleInTheWall += fHole
	p.timeBuild += Abc_Clock() - clk

	// check if this node is already driving a PO
	assert(If_ObjIsAnd(pIfObj) != 0, "If_ObjIsAnd(pIfObj)")
	pDriver = Gia_ManObj(pGia, Abc_Lit2Var(pIfObj.iCopy))
	if gia.Gia_ObjMark1(pDriver) != 0 {
		p.nFilterSame++
		return 1
	}
	gia.Gia_ObjSetMark1(pDriver, 1)
	// create output
	Gia_ManAppendCo(pGia, Abc_LitNotCond(pIfObj.iCopy, (uCanonPhase>>nLeaves)&1))

	// verify truth table
	clk = Abc_Clock()
	pTruth = Gia_ObjComputeTruthTable(pGia, Gia_ManCo(pGia, Gia_ManCoNum(pGia)-1))
	p.timeCheck += Abc_Clock() - clk
	if lmsCompareTruthWords(p.pTemp1[:p.nWords], pTruth[:p.nWords]) != 0 {
		/*
		   Kit_DsdPrintFromTruth( pTruth, nLeaves ); printf( "\n" );
		   Kit_DsdPrintFromTruth( (unsigned *)p->pTemp1, nLeaves ); printf( "\n" );
		   printf( "Truth table verification has failed.\n" );
		*/
		// drive PO with constant
		Gia_ManPatchCoDriver(pGia, Gia_ManCoNum(pGia)-1, 0)
		// save truth table ID
		Vec_IntPush(p.vTruthIds, -1)
		p.nFilterTruth++
		return 1
	}

	clk = Abc_Clock()
	// add the resulting truth table to the hash table
	Index = Vec_MemHashInsert(p.vTtMem, p.pTemp1)
	// save truth table ID
	Vec_IntPush(p.vTruthIds, Index)
	assert(Gia_ManCoNum(pGia) == Vec_IntSize(p.vTruthIds), "Gia_ManCoNum(pGia) == Vec_IntSize(p.vTruthIds)")
	p.nAdded++
	p.timeInsert += Abc_Clock() - clk
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Top level procedure for library construction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkRecAdd3(pNtk *Abc_Ntk_t, fUseSOPB int32) {
	var Pars If_Par_t
	var pPars *If_Par_t = &Pars
	var pNtkNew *Abc_Ntk_t
	var clk int32 = Abc_Clock()
	if Abc_NtkGetChoiceNum(pNtk) != 0 {
		fmt.Printf("Performing recoding structures with choices.\n")
	}
	// remember that the manager was used for library construction
	s_pMan3.fLibConstr = 1
	// create hash table if not available
	if s_pMan3.pGia != nil && gia.Gia_ManHashStarted(s_pMan3.pGia) == 0 {
		Gia_ManHashStart(s_pMan3.pGia)
	}

	// set defaults
	// Pars is zero-initialized, matching the C memset defaults.
	// user-controlable paramters
	pPars.nLutSize = s_pMan3.nVars
	pPars.nCutsMax = s_pMan3.nCuts
	pPars.DelayTarget = -1
	pPars.Epsilon = float32(0.005)
	pPars.fArea = 1
	// internal parameters
	if fUseSOPB != 0 {
		pPars.fTruth = 1
		pPars.fCutMin = 0
		pPars.fUsePerm = 1
		pPars.fDelayOpt = 1

	} else {
		pPars.fTruth = 1
		pPars.fCutMin = 1
		pPars.fUsePerm = 0
		pPars.fDelayOpt = 0
	}
	pPars.fSkipCutFilter = 0
	pPars.pFuncCost = nil
	pPars.pFuncUser = Abc_NtkRecAddCut3
	// perform recording
	pNtkNew = Abc_NtkIf(pNtk, pPars)
	Abc_NtkDelete(pNtkNew)
	s_pMan3.timeTotal += Abc_Clock() - clk
}

/*
*Function*************************************************************

	Synopsis    [Returns min AIG level at the output fo the cut using the library.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_CutComputeDelay(p *If_Man_t, pCut *If_Cut_t, pCanonPerm []byte, DelayProfile word) int32 {
	var pLeaf *If_Obj_t
	var nLeaves int32 = If_CutLeaveNum(pCut)
	var i int32
	var delayTemp int32
	var delayMax int32 = -ABC_INFINITY
	for i = 0; i < nLeaves; i++ {
		pLeaf = If_ManObj(p, (pCut).pLeaves[int32(pCanonPerm[i])])
		delayTemp = If_ObjCutBest(pLeaf).Delay + Lms_DelayGet(DelayProfile, i)
		delayMax = Abc_MaxInt(delayMax, delayTemp)
	}
	return delayMax
}
func If_CutFindBestStruct(pIfMan *If_Man_t, pCut *If_Cut_t, pCanonPerm []byte, puCanonPhase *uint32, pBestPo *int32) int32 {
	var p *Lms_Man_t = s_pMan3
	var i int32
	var pTruthId *int32
	var iFirstPo int32
	var iFirstPoNext int32
	var iBestPo int32
	var BestDelay int32 = ABC_INFINITY
	var BestArea int32 = ABC_INFINITY
	var Delay int32
	var Area int32
	var uSupport int32
	var nLeaves int32 = If_CutLeaveNum(pCut)
	var pPerm []byte = If_CutPerm(pCut)
	var DelayProfile word
	var clk abctime
	pCut.fUser = 1
	// compute support
	uSupport = Abc_TtSupport(If_CutTruthW(pIfMan, pCut), nLeaves)
	if uSupport == 0 {
		pCut.Cost = 1
		for i = 0; i < nLeaves; i++ {
			pPerm[i] = IF_BIG_CHAR
		}
		return 0
	}
	if Abc_TtSuppIsMinBase(uSupport) == 0 || uSupport == 1 {
		assert(Abc_TtSuppOnlyOne(uSupport) != 0, "Abc_TtSuppOnlyOne(uSupport)")
		pCut.Cost = 1
		for i = 0; i < nLeaves; i++ {
			pPerm[i] = IF_BIG_CHAR
		}
		pPerm[Abc_TtSuppFindFirst(uSupport)] = 0
		return If_ObjCutBest(If_ManObj(pIfMan, pCut.pLeaves[Abc_TtSuppFindFirst(uSupport)])).Delay
	}
	assert(Gia_WordCountOnes(uSupport) == nLeaves, "Gia_WordCountOnes(uSupport) == nLeaves")

	// semicanonicize the function
	clk = Abc_Clock()
	copy(p.pTemp1[:], If_CutTruthW(pIfMan, pCut))
	*puCanonPhase = Abc_TtCanonicize(p.pTemp1[:], nLeaves, pCanonPerm[:])
	lmsStretch5(p.pTemp1[:], nLeaves, p.nVars)
	p.timeCanon += Abc_Clock() - clk

	// get TT ID for the given class
	pTruthId = Vec_MemHashLookup(p.vTtMem, p.pTemp1)
	if *pTruthId == -1 {
		pCut.Cost = IF_COST_MAX
		pCut.fUseless = 1
		return ABC_INFINITY
	}

	// note that array p->vTruthPo contains the first PO for the given truth table
	// other POs belonging to the same equivalence class follow immediately after this one
	// to iterate through the POs, we need to perform the following steps

	// find the first PO of this class
	iFirstPo = Vec_IntEntry(p.vTruthPo, *pTruthId)
	// find the first PO of the next class
	iFirstPoNext = Vec_IntEntry(p.vTruthPo, *pTruthId+1)
	// iterate through the subgraphs of this class
	iBestPo = -1
	for i = iFirstPo; i < iFirstPoNext; i++ {
		Delay = If_CutComputeDelay(pIfMan, pCut, pCanonPerm, Vec_WrdEntry(p.vDelays, i))
		Area = Vec_StrEntry(p.vAreas, i)
		if iBestPo == -1 || BestDelay > Delay || (BestDelay == Delay && BestArea > Area) {
			iBestPo = i
			BestDelay = Delay
			BestArea = Area
		}
	}
	if pBestPo != nil {
		*pBestPo = iBestPo
	}

	// mark as user cut.
	DelayProfile = Vec_WrdEntry(p.vDelays, iBestPo)
	pCut.Cost = Vec_StrEntry(p.vAreas, iBestPo)
	for i = 0; i < nLeaves; i++ {
		pPerm[int32(pCanonPerm[i])] = Lms_DelayGet(DelayProfile, i)
	}
	if false {
		var Max int32
		var Two int32
		var pTimes [16]int32
		for i = 0; i < nLeaves; i++ {
			pTimes[i] = int32(If_ObjCutBest(If_CutLeaf(pIfMan, pCut, i)).Delay)
		}
		for i = 0; i < If_CutLeaveNum(pCut); i++ {
			Max = Abc_MaxInt(Max, pTimes[i])
		}
		for i = 0; i < If_CutLeaveNum(pCut); i++ {
			if pTimes[i] != Max {
				Two = Abc_MaxInt(Two, pTimes[i])
			}
		}
		if Two+2 < Max && Max+3 < BestDelay {
			for i = 0; i < If_CutLeaveNum(pCut); i++ {
				fmt.Printf("%3d ", pTimes[i])
			}
			for ; i < pIfMan.pPars.nLutSize; i++ {
				fmt.Printf("    ")
			}
			fmt.Printf("-> %3d   ", BestDelay)
			Dau_DsdPrintFromTruth(If_CutTruthW(pIfMan, pCut), If_CutLeaveNum(pCut))
		}
	}
	return BestDelay
}
func If_CutDelayRecCost3(pIfMan *If_Man_t, pCut *If_Cut_t, pObj *If_Obj_t) int32 {
	var p *Lms_Man_t = s_pMan3
	var pCanonPerm [LMS_VAR_MAX]byte
	var uCanonPhase uint32
	// make sure the cut functions match the library
	assert(p.nVars == int32(pCut.nLimit), "p.nVars == int32(pCut.nLimit)")
	// if this assertion fires, it means that LMS manager was used for library construction
	// in this case, GIA has to be written out and the manager restarted as described above
	assert(p.fLibConstr == 0, "!p.fLibConstr")
	if p.vTruthPo == nil {
		Lms_ManPrepare(p)
	}
	// return the delay of the best structure
	return If_CutFindBestStruct(pIfMan, pCut, pCanonPerm[:], &uCanonPhase, nil)
}

/*
*Function*************************************************************

	Synopsis    [Reexpresses the best structure of the cut in the HOP manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_RecToHop3(pMan *Hop_Man_t, pIfMan *If_Man_t, pCut *If_Cut_t, pIfObj *If_Obj_t) *Hop_Obj_t {
	var p *Lms_Man_t = s_pMan3
	var pCanonPerm [LMS_VAR_MAX]byte
	var uCanonPhase uint32
	var pFan0, pFan1, pHopObj *Hop_Obj_t
	var pGia *Gia_Man_t = p.pGia
	var pGiaPo *Gia_Obj_t
	var pGiaTemp *Gia_Obj_t = nil
	var i int32
	var uSupport int32
	var BestPo int32 = -1
	var nLeaves int32 = If_CutLeaveNum(pCut)
	assert(pIfMan.pPars.fCutMin == 1, "pIfMan.pPars.fCutMin == 1")

	// compute support
	uSupport = Abc_TtSupport(If_CutTruthW(pIfMan, pCut), nLeaves)
	if uSupport == 0 {
		return Hop_NotCond(Hop_ManConst0(pMan), If_CutTruthIsCompl(pCut))
	}
	if Abc_TtSuppIsMinBase(uSupport) == 0 || uSupport == 1 {
		assert(Abc_TtSuppOnlyOne(uSupport) != 0, "Abc_TtSuppOnlyOne(uSupport)")
		return Hop_NotCond(Hop_IthVar(pMan, Abc_TtSuppFindFirst(uSupport)), If_CutTruthIsCompl(pCut))
	}
	assert(Gia_WordCountOnes(uSupport) == nLeaves, "Gia_WordCountOnes(uSupport) == nLeaves")

	// get the best output for this node
	If_CutFindBestStruct(pIfMan, pCut, pCanonPerm[:], &uCanonPhase, &BestPo)
	assert(BestPo >= 0, "BestPo >= 0")
	pGiaPo = Gia_ManCo(pGia, BestPo)
	// collect internal nodes into pGia->vTtNodes
	assert(Gia_ObjIsAnd(Gia_ObjFanin0(pGiaPo)) != 0, "Gia_ObjIsAnd( Gia_ObjFanin0(pGiaPo) )")
	Gia_ObjCollectInternal(pGia, Gia_ObjFanin0(pGiaPo))
	vTtNodes := gia.Gia_ManTruthTableNodeIDs(pGia)
	assert(len(vTtNodes) > 0, "truth-table nodes are available")

	// collect HOP nodes for leaves
	Vec_PtrClear(p.vLabelsP)
	for i = 0; i < nLeaves; i++ {
		Vec_PtrPush(p.vLabelsP, Hop_NotCond(Hop_IthVar(pMan, pCanonPerm[i]), (uCanonPhase>>i)&1))
	}

	// compute HOP nodes for internal nodes
	for i = 0; i < int32(len(vTtNodes)); i++ {
		pGiaTemp = Gia_ManObj(pGia, vTtNodes[i])
		gia.Gia_ObjSetMark0(pGiaTemp, 0) // unmark node marked by Gia_ObjCollectInternal()
		if Gia_ObjIsAnd(Gia_ObjFanin0(pGiaTemp)) != 0 {
			pFan0 = Vec_PtrEntry(p.vLabelsP, int(Gia_ObjNum(pGia, Gia_ObjFanin0(pGiaTemp))+nLeaves)).(*Hop_Obj_t)
		} else {
			pFan0 = Vec_PtrEntry(p.vLabelsP, int(Gia_ObjCioId(Gia_ObjFanin0(pGiaTemp)))).(*Hop_Obj_t)
		}
		pFan0 = Hop_NotCond(pFan0, Gia_ObjFaninC0(pGiaTemp))
		if Gia_ObjIsAnd(Gia_ObjFanin1(pGiaTemp)) != 0 {
			pFan1 = Vec_PtrEntry(p.vLabelsP, int(Gia_ObjNum(pGia, Gia_ObjFanin1(pGiaTemp))+nLeaves)).(*Hop_Obj_t)
		} else {
			pFan1 = Vec_PtrEntry(p.vLabelsP, int(Gia_ObjCioId(Gia_ObjFanin1(pGiaTemp)))).(*Hop_Obj_t)
		}
		pFan1 = Hop_NotCond(pFan1, Gia_ObjFaninC1(pGiaTemp))

		pHopObj = Hop_And(pMan, pFan0, pFan1)
		Vec_PtrPush(p.vLabelsP, pHopObj)
	}
	// get the final result
	assert(Gia_ObjIsAnd(pGiaTemp) != 0, "Gia_ObjIsAnd(pGiaTemp)")
	pHopObj = Vec_PtrEntry(p.vLabelsP, int(Gia_ObjNum(pGia, pGiaTemp)+nLeaves)).(*Hop_Obj_t)
	// complement the result if needed
	return Hop_NotCond(pHopObj, Gia_ObjFaninC0(pGiaPo)^((uCanonPhase>>nLeaves)&1))
}

/*
*Function*************************************************************

	Synopsis    [Reexpresses the best structure of the cut in the GIA manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_RecToGia3(pMan *Gia_Man_t, pIfMan *If_Man_t, pCut *If_Cut_t, vLeaves *Vec_Int_t, fHash int32) int32 {
	var p *Lms_Man_t = s_pMan3
	var pCanonPerm [LMS_VAR_MAX]byte
	var uCanonPhase uint32
	var iFan0, iFan1, iGiaObj int32
	var pGia *Gia_Man_t = p.pGia
	var pGiaPo *Gia_Obj_t
	var pGiaTemp *Gia_Obj_t = nil
	var i int32
	var uSupport int32
	var BestPo int32 = -1
	var nLeaves int32 = If_CutLeaveNum(pCut)
	assert(pIfMan.pPars.fCutMin == 1, "pIfMan.pPars.fCutMin == 1")
	assert(nLeaves == Vec_IntSize(vLeaves), "nLeaves == Vec_IntSize(vLeaves)")

	// compute support
	uSupport = Abc_TtSupport(If_CutTruthW(pIfMan, pCut), nLeaves)
	if uSupport == 0 {
		return Abc_LitNotCond(0, If_CutTruthIsCompl(pCut))
	}
	if Abc_TtSuppIsMinBase(uSupport) == 0 || uSupport == 1 {
		assert(Abc_TtSuppOnlyOne(uSupport) != 0, "Abc_TtSuppOnlyOne(uSupport)")
		return Abc_LitNotCond(Vec_IntEntry(vLeaves, Abc_TtSuppFindFirst(uSupport)), If_CutTruthIsCompl(pCut))
	}
	assert(Gia_WordCountOnes(uSupport) == nLeaves, "Gia_WordCountOnes(uSupport) == nLeaves")

	// get the best output for this node
	If_CutFindBestStruct(pIfMan, pCut, pCanonPerm[:], &uCanonPhase, &BestPo)
	assert(BestPo >= 0, "BestPo >= 0")
	pGiaPo = Gia_ManCo(pGia, BestPo)

	// collect internal nodes into pGia->vTtNodes
	assert(Gia_ObjIsAnd(Gia_ObjFanin0(pGiaPo)) != 0, "Gia_ObjIsAnd( Gia_ObjFanin0(pGiaPo) )")
	Gia_ObjCollectInternal(pGia, Gia_ObjFanin0(pGiaPo))
	vTtNodes := gia.Gia_ManTruthTableNodeIDs(pGia)
	assert(len(vTtNodes) > 0, "truth-table nodes are available")

	// collect GIA nodes for leaves
	Vec_IntClear(p.vLabels)
	for i = 0; i < nLeaves; i++ {
		Vec_IntPush(p.vLabels, Abc_LitNotCond(Vec_IntEntry(vLeaves, pCanonPerm[i]), (uCanonPhase>>i)&1))
	}

	// compute HOP nodes for internal nodes
	for i = 0; i < int32(len(vTtNodes)); i++ {
		pGiaTemp = Gia_ManObj(pGia, vTtNodes[i])
		gia.Gia_ObjSetMark0(pGiaTemp, 0) // unmark node marked by Gia_ObjCollectInternal()
		if Gia_ObjIsAnd(Gia_ObjFanin0(pGiaTemp)) != 0 {
			iFan0 = Vec_IntEntry(p.vLabels, Gia_ObjNum(pGia, Gia_ObjFanin0(pGiaTemp))+nLeaves)
		} else {
			iFan0 = Vec_IntEntry(p.vLabels, Gia_ObjCioId(Gia_ObjFanin0(pGiaTemp)))
		}
		iFan0 = Abc_LitNotCond(iFan0, Gia_ObjFaninC0(pGiaTemp))
		if Gia_ObjIsAnd(Gia_ObjFanin1(pGiaTemp)) != 0 {
			iFan1 = Vec_IntEntry(p.vLabels, Gia_ObjNum(pGia, Gia_ObjFanin1(pGiaTemp))+nLeaves)
		} else {
			iFan1 = Vec_IntEntry(p.vLabels, Gia_ObjCioId(Gia_ObjFanin1(pGiaTemp)))
		}
		iFan1 = Abc_LitNotCond(iFan1, Gia_ObjFaninC1(pGiaTemp))
		if fHash != 0 {
			iGiaObj = Gia_ManHashAnd(pMan, iFan0, iFan1)
		} else {
			iGiaObj = Gia_ManAppendAnd(pMan, iFan0, iFan1)
		}
		Vec_IntPush(p.vLabels, iGiaObj)
	}
	// get the final result
	assert(Gia_ObjIsAnd(pGiaTemp) != 0, "Gia_ObjIsAnd(pGiaTemp)")
	iGiaObj = Vec_IntEntry(p.vLabels, Gia_ObjNum(pGia, pGiaTemp)+nLeaves)
	// complement the result if needed
	return Abc_LitNotCond(iGiaObj, Gia_ObjFaninC0(pGiaPo)^((uCanonPhase>>nLeaves)&1)^pCut.fCompl)
}

/**Function*************************************************************

  Synopsis    [Reduces GIA to contain only useful COs and internal nodes.]

  Description [During library construction, redundant nodes are added.
  Some COs are found to be useless because their TT does not match the
  (semi-canonicized TT) of the cut, etc.  This procedure reduces GIA
  to contains only useful (non-redundant, non-dominated) COs and the
  corresponding internal nodes. This procedure replaces GIA by a new GIA
  and creates new vTruthIds. The COs with the same truth table have
  adjacent IDs. This procedure does not change the truth tables.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
// count how many times TT occurs
func Lms_GiaCountTruths(p *Lms_Man_t) *Vec_Int_t {
	var vCounts *Vec_Int_t = Vec_IntStart(Vec_MemEntryNum(p.vTtMem))
	var i, Entry int32
	for i = 0; i < Vec_IntSize(p.vTruthIds); i++ {
		Entry = Vec_IntEntry(p.vTruthIds, i)
		if Entry >= 0 {
			Vec_IntAddToEntry(vCounts, Entry, 1)
		}
	}
	return vCounts
}

// collect PO indexes worth visiting
func Lms_GiaCollectUsefulCos(p *Lms_Man_t) *Vec_Int_t {
	var vBegins *Vec_Int_t = Vec_IntAlloc(Vec_MemEntryNum(p.vTtMem))
	var vUseful *Vec_Int_t = Vec_IntStartFull(Gia_ManCoNum(p.pGia) + Vec_MemEntryNum(p.vTtMem))
	var vCounts *Vec_Int_t = Lms_GiaCountTruths(p)
	var i int32
	var Entry int32
	var pPlace *int32
	var SumTotal int32
	// mark up the place for POs
	for i = 0; i < Vec_IntSize(vCounts); i++ {
		Entry = Vec_IntEntry(vCounts, i)
		assert(Entry > 0, "Entry > 0")
		Vec_IntPush(vBegins, SumTotal)
		SumTotal += Entry + 1
		//        printf( "%d ", Entry );
	}
	Vec_IntPush(vBegins, SumTotal)
	// fill out POs in their places
	Vec_IntFill(vCounts, Vec_IntSize(vCounts), 0)
	for i = 0; i < Vec_IntSize(p.vTruthIds); i++ {
		Entry = Vec_IntEntry(p.vTruthIds, i)
		if Entry < 0 {
			continue
		}
		pPlace = Vec_IntEntryP(vUseful, Vec_IntEntry(vBegins, Entry)+Vec_IntEntry(vCounts, Entry))
		assert(*pPlace == -1, "*pPlace == -1")
		*pPlace = i
		Vec_IntAddToEntry(vCounts, Entry, 1)
	}
	Vec_IntFree(vBegins)
	Vec_IntFree(vCounts)
	return vUseful
}

// collect non-dominated COs
func Lms_GiaFindNonRedundantCos(p *Lms_Man_t) *Vec_Int_t {
	var vRemain *Vec_Int_t
	var vUseful *Vec_Int_t
	var vDelays *Vec_Wrd_t
	var i, k, EntryI, EntryK int32
	var D1, D2 word
	vDelays = Lms_GiaDelays(p.pGia)
	vUseful = Lms_GiaCollectUsefulCos(p)
	for i = 0; i < Vec_IntSize(vUseful); i++ {
		EntryI = Vec_IntEntry(vUseful, i)
		if EntryI < 0 {
			continue
		}
		D1 = Vec_WrdEntry(vDelays, EntryI)
		assert(D1 > 0, "D1 > 0")
		for k = i + 1; k < Vec_IntSize(vUseful); k++ {
			EntryK = Vec_IntEntry(vUseful, k)
			if EntryK == -1 {
				break
			}
			if EntryK == -2 {
				continue
			}
			D2 = Vec_WrdEntry(vDelays, EntryK)
			assert(D2 > 0, "D2 > 0")
			if Lms_DelayDom(D1, D2, Gia_ManCiNum(p.pGia)) != 0 { // D1 dominate D2

				Vec_IntWriteEntry(vUseful, k, -2)
				continue
			}
			if Lms_DelayDom(D2, D1, Gia_ManCiNum(p.pGia)) != 0 { // D2 dominate D1

				Vec_IntWriteEntry(vUseful, i, -2)
				break
			}
		}
	}

	vRemain = Vec_IntAlloc(1000)
	for i = 0; i < Vec_IntSize(vUseful); i++ {
		EntryI = Vec_IntEntry(vUseful, i)
		if EntryI >= 0 {
			Vec_IntPush(vRemain, EntryI)
		}
	}
	Vec_IntFree(vUseful)
	Vec_WrdFree(vDelays)
	return vRemain
}

// replace GIA and vTruthIds by filtered ones
func Lms_GiaNormalize(p *Lms_Man_t) {
	var pGiaNew *Gia_Man_t
	var pObj *Gia_Obj_t
	var vRemain *Vec_Int_t
	var vTruthIdsNew *Vec_Int_t
	var i int32
	var Entry int32
	var Prev int32 = -1
	var Next int32
	// collect non-redundant COs
	vRemain = Lms_GiaFindNonRedundantCos(p)
	// change these to be useful literals
	vTruthIdsNew = Vec_IntAlloc(Vec_IntSize(vRemain))
	for i = 0; i < Vec_IntSize(vRemain); i++ {
		Entry = Vec_IntEntry(vRemain, i)
		pObj = Gia_ManCo(p.pGia, Entry)
		assert(Gia_ObjIsAnd(Gia_ObjFanin0(pObj)) != 0, "Gia_ObjIsAnd(Gia_ObjFanin0(pObj))")
		Vec_IntWriteEntry(vRemain, i, Gia_ObjFaninLit0p(p.pGia, pObj))
		// create new truth IDs
		Next = Vec_IntEntry(p.vTruthIds, Gia_ObjCioId(pObj))
		assert(Prev <= Next, "Prev <= Next")
		Vec_IntPush(vTruthIdsNew, Next)
		Prev = Next
	}
	// create a new GIA
	for i = 0; i < Gia_ManObjNum(p.pGia); i++ {
		pObj = Gia_ManObj(p.pGia, i)
		assert(gia.Gia_ObjMark0(pObj) == 0, "pObj.fMark0 == 0")
	}
	for i = 0; i < Gia_ManCoNum(p.pGia); i++ {
		Gia_ManPatchCoDriver(p.pGia, i, 0)
	}
	for i = 0; i < Vec_IntSize(vRemain); i++ {
		Entry = Vec_IntEntry(vRemain, i)
		Gia_ManAppendCo(p.pGia, Entry)
	}
	//    pGiaNew = Gia_ManCleanup( p->pGia );
	pGiaNew = Gia_ManCleanupOutputs(p.pGia, Gia_ManCoNum(p.pGia)-Vec_IntSize(vRemain))
	Gia_ManStop(p.pGia)
	p.pGia = pGiaNew
	Vec_IntFree(vRemain)
	// update truth IDs
	Vec_IntFree(p.vTruthIds)
	p.vTruthIds = vTruthIdsNew
	// Vec_IntPrint( vTruthIdsNew );
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkRecTruthCompare(p1 *int32, p2 *int32) int32 {
	var Diff int32 = Vec_StrEntry(s_pMan3.vSupps, *p1) - Vec_StrEntry(s_pMan3.vSupps, *p2)
	if Diff != 0 {
		return Diff
	}
	return lmsCompareTruthWords(Vec_MemReadEntry(s_pMan3.vTtMem, int(*p1)), Vec_MemReadEntry(s_pMan3.vTtMem, int(*p2)))
}
func Abc_NtkRecDumpTt3(pFileName string, fBinary int32) {
	var pFile *os.File
	var pBuffer [1000]byte
	var p *Lms_Man_t = s_pMan3
	var vEntries *Vec_Int_t
	var pTruth []word
	var i int32
	var Entry int32
	var nVars int32 = p.nVars
	var nEntries int32 = Vec_MemEntryNum(p.vTtMem)
	if nEntries == 0 {
		fmt.Printf("There is not truth tables.\n")
		return
	}
	var err error
	pFile, err = os.OpenFile(pFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("The file cannot be opened.\n")
		return
	}
	p.vSupps = Vec_StrAlloc(nEntries)
	for i = 0; i < Vec_MemEntryNum(p.vTtMem); i++ {
		pTruth = Vec_MemReadEntry(p.vTtMem, int(i))
		Vec_StrPush(p.vSupps, byte(Abc_TtSupportSize(pTruth, nVars)))
	}
	vEntries = Vec_IntStartNatural(nEntries)
	sort.Slice(Vec_IntArray(vEntries), func(i, j int) bool {
		a, b := Vec_IntEntry(vEntries, int32(i)), Vec_IntEntry(vEntries, int32(j))
		return Abc_NtkRecTruthCompare(&a, &b) < 0
	})
	Vec_StrFreeP(&p.vSupps)
	// write the file
	for i = 0; i < Vec_IntSize(vEntries); i++ {
		Entry = Vec_IntEntry(vEntries, i)
		pTruth = Vec_MemReadEntry(p.vTtMem, Entry)
		if fBinary != 0 {
			if err := binary.Write(pFile, binary.NativeEndian, pTruth[:p.nWords]); err != nil {
				return
			}
			continue
		}
		extra.Extra_PrintHex(pFile, wordsToUint32(pTruth), nVars)
		fmt.Fprintf(pFile, "  ")
		//        Kit_DsdWriteFromTruth( pBuffer, (unsigned *)pTruth, nVars );
		Dau_DsdDecompose(pTruth, p.nVars, 0, lmsBoolInt32(nVars <= 10), &pBuffer[:])
		fmt.Fprintf(pFile, "%s\n", pBuffer[:])
	}
	_ = (pFile).Close()
	Vec_IntFree(vEntries)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkRecInputNum3() int32 {
	return Gia_ManCiNum(s_pMan3.pGia)
}
func Abc_NtkRecIsRunning3() int32 {
	if s_pMan3 != nil {
		return 1
	}
	return 0
}
func Abc_NtkRecGetGia3() *Gia_Man_t {
	var clk abctime = Abc_Clock()
	fmt.Printf("Before normalizing: Library has %d classes and %d AIG subgraphs with %d AND nodes.\n", Vec_MemEntryNum(s_pMan3.vTtMem), Gia_ManPoNum(s_pMan3.pGia), Gia_ManAndNum(s_pMan3.pGia))
	Lms_GiaNormalize(s_pMan3)
	fmt.Printf("After normalizing:  Library has %d classes and %d AIG subgraphs with %d AND nodes.\n", Vec_MemEntryNum(s_pMan3.vTtMem), Gia_ManPoNum(s_pMan3.pGia), Gia_ManAndNum(s_pMan3.pGia))
	Abc_PrintTime(1, "Normalization runtime", Abc_Clock()-clk)
	s_pMan3.fLibConstr = 0
	return s_pMan3.pGia
}
func Abc_NtkRecPs3(fPrintLib int32) {
	Lms_ManPrint(s_pMan3)
}
func Abc_NtkRecStart3(p *Gia_Man_t, nVars int32, nCuts int32, fFuncOnly int32, fVerbose int32) {
	assert(s_pMan3 == nil, "s_pMan3 == nil")
	s_pMan3 = Lms_ManStart(p, nVars, nCuts, fFuncOnly, fVerbose)
}

func Abc_NtkRecStop3() {
	assert(s_pMan3 != nil, "s_pMan3 != nil")
	Lms_ManStop(s_pMan3)
	s_pMan3 = nil
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
