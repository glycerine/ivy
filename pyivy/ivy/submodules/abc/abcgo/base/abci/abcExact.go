package abci

import (
	"encoding/binary"
	spFmt "fmt"
	"io"
	"os"
	"time"
	"unsafe"

	"github.com/glycerine/abcgo/aig/gia"
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec"
	"github.com/glycerine/abcgo/sat/bsat"
)

/**CFile****************************************************************

  FileName    [abcExact.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Find minimum size networks with a SAT solver.]

  Author      [Mathias Soeken]

  Affiliation [EPFL]

  Date        [Ver. 1.0. Started - July 15, 2016.]

  Revision    [$Id: abcFanio.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

/* This implementation is based on Exercises 477 and 478 in
 * Donald E. Knuth TAOCP Fascicle 6 (Satisfiability) Section 7.2.2.2
 */

//#include "base/abc/abc.h"

//#include "aig/gia/gia.h"
//#include "misc/util/utilTruth.h"
//#include "misc/vec/vecInt.h"
//#include "misc/vec/vecPtr.h"
//#include "proof/cec/cec.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/***********************************************************************

  Synopsis    [Some truth table helper functions.]

***********************************************************************/

var s_Truths8 [32]word = [32]word{
	word(0xAAAAAAAAAAAAAAAA), word(0xAAAAAAAAAAAAAAAA), word(0xAAAAAAAAAAAAAAAA), word(0xAAAAAAAAAAAAAAAA),
	word(0xCCCCCCCCCCCCCCCC), word(0xCCCCCCCCCCCCCCCC), word(0xCCCCCCCCCCCCCCCC), word(0xCCCCCCCCCCCCCCCC),
	word(0xF0F0F0F0F0F0F0F0), word(0xF0F0F0F0F0F0F0F0), word(0xF0F0F0F0F0F0F0F0), word(0xF0F0F0F0F0F0F0F0),
	word(0xFF00FF00FF00FF00), word(0xFF00FF00FF00FF00), word(0xFF00FF00FF00FF00), word(0xFF00FF00FF00FF00),
	word(0xFFFF0000FFFF0000), word(0xFFFF0000FFFF0000), word(0xFFFF0000FFFF0000), word(0xFFFF0000FFFF0000),
	word(0xFFFFFFFF00000000), word(0xFFFFFFFF00000000), word(0xFFFFFFFF00000000), word(0xFFFFFFFF00000000),
	word(0x0000000000000000), word(0xFFFFFFFFFFFFFFFF), word(0x0000000000000000), word(0xFFFFFFFFFFFFFFFF),
	word(0x0000000000000000), word(0x0000000000000000), word(0xFFFFFFFFFFFFFFFF), word(0xFFFFFFFFFFFFFFFF),
}

var s_Truths8Neg [32]word = [32]word{
	word(0x5555555555555555), word(0x5555555555555555), word(0x5555555555555555), word(0x5555555555555555),
	word(0x3333333333333333), word(0x3333333333333333), word(0x3333333333333333), word(0x3333333333333333),
	word(0x0F0F0F0F0F0F0F0F), word(0x0F0F0F0F0F0F0F0F), word(0x0F0F0F0F0F0F0F0F), word(0x0F0F0F0F0F0F0F0F),
	word(0x00FF00FF00FF00FF), word(0x00FF00FF00FF00FF), word(0x00FF00FF00FF00FF), word(0x00FF00FF00FF00FF),
	word(0x0000FFFF0000FFFF), word(0x0000FFFF0000FFFF), word(0x0000FFFF0000FFFF), word(0x0000FFFF0000FFFF),
	word(0x00000000FFFFFFFF), word(0x00000000FFFFFFFF), word(0x00000000FFFFFFFF), word(0x00000000FFFFFFFF),
	word(0xFFFFFFFFFFFFFFFF), word(0x0000000000000000), word(0xFFFFFFFFFFFFFFFF), word(0x0000000000000000),
	word(0xFFFFFFFFFFFFFFFF), word(0xFFFFFFFFFFFFFFFF), word(0x0000000000000000), word(0x0000000000000000),
}

var s_Truths6 = func() [6]word {
	var r [6]word
	for v := 0; v < 6; v++ {
		period := uint(1 << v)
		for bit := uint(0); bit < 64; bit++ {
			if (bit/period)&1 != 0 {
				r[v] |= word(1) << bit
			}
		}
	}
	return r
}()

func Abc_TtWordNum(nVars int32) int32 {
	if nVars < 0 || nVars > 8 {
		return 0
	}
	if nVars <= 6 {
		return 1
	}
	return 1 << uint(nVars-6)
}

func Abc_TtPrintHexRev(w io.Writer, truth []word, nVars int32) {
	if w == nil || nVars < 2 {
		return
	}
	nWords := int(Abc_TtWordNum(nVars))
	if nWords > len(truth) {
		nWords = len(truth)
	}
	start := 1 << (nVars - 2)
	if nVars >= 6 {
		start = 16
	}
	for wi := nWords - 1; wi >= 0; wi-- {
		for k := start - 1; k >= 0; k-- {
			d := (truth[wi] >> uint(k*4)) & 15
			if d < 10 {
				_, _ = spFmt.Fprintf(w, "%c", '0'+rune(d))
			} else {
				_, _ = spFmt.Fprintf(w, "%c", 'A'+rune(d-10))
			}
		}
	}
}
func abcExactWords(p *word, n int32) []word {
	if p == nil || n < 0 || n > 4 {
		return nil
	}
	return unsafe.Slice(p, int(n))
}
func Abc_TtCopy(dst, src *word, nWords, invert int32) {
	d, s := abcExactWords(dst, nWords), abcExactWords(src, nWords)
	if d == nil || s == nil {
		return
	}
	for i := range d {
		d[i] = s[i]
		if invert != 0 {
			d[i] = ^d[i]
		}
	}
}
func Abc_TtMask(dst *word, nWords, nBits int32) {
	d := abcExactWords(dst, nWords)
	if d == nil {
		return
	}
	for i := range d {
		d[i] = ^word(0)
	}
	if nBits < 0 {
		nBits = 0
	}
	for bit := nBits; bit < nWords*64; bit++ {
		d[bit/64] &^= word(1) << uint(bit%64)
	}
}
func Abc_TtAnd(dst, a, b *word, nWords, invert int32) {
	d, x, y := abcExactWords(dst, nWords), abcExactWords(a, nWords), abcExactWords(b, nWords)
	if d == nil || x == nil || y == nil {
		return
	}
	for i := range d {
		d[i] = x[i] & y[i]
		if invert != 0 {
			d[i] = ^d[i]
		}
	}
}
func Abc_TtXor(dst, a, b *word, nWords, invert int32) {
	d, x, y := abcExactWords(dst, nWords), abcExactWords(a, nWords), abcExactWords(b, nWords)
	if d == nil || x == nil || y == nil {
		return
	}
	for i := range d {
		d[i] = x[i] ^ y[i]
		if invert != 0 {
			d[i] = ^d[i]
		}
	}
}

func Abc_TtIsSubsetWithMask(pSmall, pLarge, pMask *word, nWords int32) int32 {
	if nWords < 0 || nWords > 4 || pSmall == nil || pLarge == nil || pMask == nil {
		return 0
	}
	small, large, mask := abcExactWords(pSmall, nWords), abcExactWords(pLarge, nWords), abcExactWords(pMask, nWords)
	for w := int32(0); w < nWords; w++ {
		if small[w]&large[w]&mask[w] != small[w]&mask[w] {
			return 0
		}
	}
	return 1
}
func Abc_TtCofsOppositeWithMask(pTruth, pMask *word, nWords, iVar int32) int32 {
	if pTruth == nil || pMask == nil || nWords < 0 || nWords > 4 || iVar < 0 || iVar >= 8 {
		return 0
	}
	truth, mask := abcExactWords(pTruth, nWords), abcExactWords(pMask, nWords)
	if iVar < 6 {
		shift := uint(1 << uint(iVar))
		for w := int32(0); w < nWords; w++ {
			if (truth[w]<<shift)&s_Truths6[iVar]&mask[w] != (^truth[w])&s_Truths6[iVar]&mask[w] {
				return 0
			}
		}
		return 1
	}
	step := int32(1) << uint(iVar-6)
	for base := int32(0); base < nWords; base += 2 * step {
		for w := int32(0); w < step && base+w+step < nWords; w++ {
			if truth[base+w]&mask[base+w] != (^truth[base+w+step])&mask[base+w+step] {
				return 0
			}
		}
	}
	return 1
}
func Abc_TtIsTopDecomposable(pTruth, pMask *word, nWords, iVar int32) int32 {
	if iVar < 0 || iVar >= 8 || nWords < 0 || nWords > 4 {
		return 0
	}
	if Abc_TtIsSubsetWithMask(pTruth, &s_Truths8[iVar<<2], pMask, nWords) != 0 {
		return 1
	}
	if Abc_TtIsSubsetWithMask(pTruth, &s_Truths8Neg[iVar<<2], pMask, nWords) != 0 {
		return 2
	}
	if Abc_TtIsSubsetWithMask(&s_Truths8[iVar<<2], pTruth, pMask, nWords) != 0 {
		return 3
	}
	if Abc_TtIsSubsetWithMask(&s_Truths8Neg[iVar<<2], pTruth, pMask, nWords) != 0 {
		return 4
	}
	if Abc_TtCofsOppositeWithMask(pTruth, pMask, nWords, iVar) != 0 {
		return 5
	}
	return 0
}
func Abc_TtIsStairDecomposable(pTruth *word, nWords int32, pVars *int32, nSize int32, pStairFunc *int32) int32 {
	if nWords < 0 || nWords > 4 || nSize < 0 || nSize > 8 || pTruth == nil || pVars == nil || pStairFunc == nil {
		return 0
	}
	vars, stair := unsafe.Slice(pVars, int(nSize)), unsafe.Slice(pStairFunc, int(nSize))
	var mask, copy [4]word
	Abc_TtCopy(&copy[0], pTruth, nWords, 0)
	Abc_TtMask(&mask[0], nWords, nWords*64)
	for i := int32(0); i < nSize; i++ {
		v := vars[i]
		if v < 0 || v >= 8 {
			return 0
		}
		d := Abc_TtIsTopDecomposable(&copy[0], &mask[0], nWords, v)
		if d == 0 {
			return 0
		}
		stair[i] = d
		switch d {
		case 1, 4:
			Abc_TtAnd(&mask[0], &mask[0], &s_Truths8[v<<2], nWords, 0)
		case 2, 3:
			Abc_TtAnd(&mask[0], &mask[0], &s_Truths8Neg[v<<2], nWords, 0)
		case 5:
			Abc_TtXor(&copy[0], &copy[0], &s_Truths8[v<<2], nWords, 0)
		}
	}
	return 1
}

/***********************************************************************

  Synopsis    [Some printing utilities.]

***********************************************************************/

func Abc_DebugPrint(str string, fCond int32) {
	if fCond != 0 {
		spFmt.Print(str)
		_ = os.Stdout.Sync()
	}
}
func Abc_DebugPrintInt(format string, n, fCond int32) {
	if fCond != 0 {
		spFmt.Printf(format, n)
		_ = os.Stdout.Sync()
	}
}
func Abc_DebugPrintIntInt(format string, n1, n2, fCond int32) {
	if fCond != 0 {
		spFmt.Printf(format, n1, n2)
		_ = os.Stdout.Sync()
	}
}
func Abc_DebugErase(n, fCond int32) {
	if fCond != 0 {
		for i := int32(0); i < n; i++ {
			spFmt.Print("\b")
		}
		_ = os.Stdout.Sync()
	}
}

/***********************************************************************

  Synopsis    [BMS.]

***********************************************************************/

const ABC_EXACT_SOL_NVARS = 0
const ABC_EXACT_SOL_NFUNC = 1
const ABC_EXACT_SOL_NGATES = 2

const ANSI_COLOR_RED = "\x1b[31m"
const ANSI_COLOR_GREEN = "\x1b[32m"
const ANSI_COLOR_YELLOW = "\x1b[33m"
const ANSI_COLOR_BLUE = "\x1b[34m"
const ANSI_COLOR_MAGENTA = "\x1b[35m"
const ANSI_COLOR_CYAN = "\x1b[36m"
const ANSI_COLOR_RESET = "\x1b[0m"

type Ses_Man_t struct {
	pSat                                                                           *bsat.Solver
	pSpec                                                                          *word
	bSpecInv, nSpecVars, nSpecFunc, nSpecWords, nRows                              int32
	nMaxDepth, nMaxDepthTmp                                                        int32
	pArrTimeProfile                                                                *int32
	pArrTimeProfileTmp                                                             [8]int32
	nArrTimeDelta, nArrTimeMax, nBTLimit                                           int32
	fMakeAIG, fVerbose, fVeryVerbose, fExtractVerbose, fSatVerbose, fReasonVerbose int32
	pTtValues                                                                      [4]word
	vPolar, vAssump                                                                *Vec_Int_t
	nRandRowAssigns, fKeepRowAssigns                                               int32
	nGates, nStartGates, nMaxGates, fDecStructure, pDecVars                        int32
	vStairDecVars                                                                  *Vec_Int_t
	pStairDecFunc                                                                  [8]int32
	pTtObjs                                                                        [100]word
	nSimVars, nOutputVars, nGateVars, nSelectVars, nDepthVars                      int32
	nSimOffset, nOutputOffset, nGateOffset, nSelectOffset, nDepthOffset            int32
	fHitResLimit                                                                   int32
	timeSat, timeSatSat, timeSatUnsat, timeSatUndef, timeInstance, timeTotal       int64
	nSatCalls, nUnsatCalls, nUndefCalls, nDebugOffset                              int32
	rng                                                                            *glibcRand
}

type Ses_TimesEntry_t struct {
	pArrTimeProfile [8]int32
	fResLimit       int32
	next            *Ses_TimesEntry_t
	pNetwork        string
}
type Ses_TruthEntry_t struct {
	pTruth [4]word
	nVars  int32
	next   *Ses_TruthEntry_t
	head   *Ses_TimesEntry_t
}

const SES_STORE_TABLE_SIZE = 1024

type Ses_Store_t struct {
	fMakeAIG, fVerbose, fVeryVerbose, nBTLimit                                          int32
	nEntriesCount, nValidEntriesCount                                                   int32
	pEntries                                                                            [SES_STORE_TABLE_SIZE]*Ses_TruthEntry_t
	pSat                                                                                *bsat.Solver
	pDebugEntries                                                                       *os.File
	szDBName                                                                            string
	nCutCount                                                                           uint64
	pCutCount                                                                           [9]uint64
	nUnsynthesizedImp                                                                   uint64
	pUnsynthesizedImp                                                                   [9]uint64
	nUnsynthesizedRL                                                                    uint64
	pUnsynthesizedRL                                                                    [9]uint64
	nSynthesizedTrivial                                                                 uint64
	pSynthesizedTrivial                                                                 [9]uint64
	nSynthesizedImp                                                                     uint64
	pSynthesizedImp                                                                     [9]uint64
	nSynthesizedRL                                                                      uint64
	pSynthesizedRL                                                                      [9]uint64
	nCacheHits                                                                          uint64
	pCacheHits                                                                          [9]uint64
	nSatCalls, nUnsatCalls, nUndefCalls                                                 uint64
	timeExact, timeSat, timeSatSat, timeSatUnsat, timeSatUndef, timeInstance, timeTotal int64
}

var s_pSesStore *Ses_Store_t = nil

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

func Abc_NormalizeArrivalTimes(pArrTimeProfile *int32, nVars int32, maxNormalized *int32) int32 {
	if pArrTimeProfile == nil || maxNormalized == nil || nVars <= 0 {
		return 0
	}
	values := unsafe.Slice(pArrTimeProfile, int(nVars))
	delta := values[0]
	for _, v := range values[1:] {
		if v < delta {
			delta = v
		}
	}
	*maxNormalized = 0
	for i := range values {
		values[i] -= delta
		if values[i] > *maxNormalized {
			*maxNormalized = values[i]
		}
	}
	*maxNormalized = *maxNormalized + 1
	return delta
}

func Ses_StoreAlloc(nBTLimit int32, fMakeAIG int32, fVerbose int32) *Ses_Store_t {
	store := &Ses_Store_t{fMakeAIG: fMakeAIG, fVerbose: fVerbose, nBTLimit: nBTLimit}
	store.pSat = bsat.NewSolver()
	return store
}

func Ses_StoreClean(store *Ses_Store_t) {
	if store == nil {
		return
	}
	for bucket := range store.pEntries {
		truth := store.pEntries[bucket]
		for truth != nil {
			times := truth.head
			for times != nil {
				next := times.next
				times.pNetwork = ""
				times.next = nil
				times = next
			}
			next := truth.next
			truth.head = nil
			truth.next = nil
			truth = next
		}
		store.pEntries[bucket] = nil
	}
	if store.pSat != nil {
		store.pSat.Close()
		store.pSat = nil
	}
	store.szDBName = ""
	*store = Ses_Store_t{}
}

var spStatic_Ses_StoreTableHash_s_Primes = [4]uint32{1291, 1699, 1999, 2357}

func sesStoreTruthWords(pTruth *word, nVars int32) []word {
	if pTruth == nil || nVars < 0 || nVars > 8 {
		return nil
	}
	nWords := int(Abc_TtWordNum(nVars))
	if nWords < 0 || nWords > 4 {
		return nil
	}
	return unsafe.Slice(pTruth, nWords)
}

func sesStoreTimes(pTimes *int32, nVars int32) []int32 {
	if pTimes == nil || nVars < 0 || nVars > 8 {
		return nil
	}
	return unsafe.Slice(pTimes, int(nVars))
}

func Ses_StoreTableHash(pTruth *word, nVars int32) int32 {
	words := sesStoreTruthWords(pTruth, nVars)
	var hash uint32
	for i, truth := range words {
		// C multiplies in word-width arithmetic and truncates into unsigned.
		hash ^= uint32(truth * uint64(spStatic_Ses_StoreTableHash_s_Primes[i&0xf]))
	}
	return int32(hash % SES_STORE_TABLE_SIZE)
}

func Ses_StoreTruthEqual(entry *Ses_TruthEntry_t, pTruth *word, nVars int32) int32 {
	if entry == nil || entry.nVars != nVars {
		return 0
	}
	words := sesStoreTruthWords(pTruth, nVars)
	if words == nil {
		return 0
	}
	for i, truth := range words {
		if entry.pTruth[i] != truth {
			return 0
		}
	}
	return 1
}

func Ses_StoreTruthCopy(entry *Ses_TruthEntry_t, pTruthSrc *word, nVars int32) {
	if entry == nil {
		return
	}
	words := sesStoreTruthWords(pTruthSrc, nVars)
	entry.nVars = nVars
	copy(entry.pTruth[:], words)
}

func Ses_StoreTimesEqual(pTimes1, pTimes2 *int32, nVars int32) int32 {
	if nVars < 0 || nVars > 8 || pTimes1 == nil || pTimes2 == nil {
		return 0
	}
	a, b := sesStoreTimes(pTimes1, nVars), sesStoreTimes(pTimes2, nVars)
	if len(a) != len(b) {
		return 0
	}
	for i := range a {
		if a[i] != b[i] {
			return 0
		}
	}
	return 1
}

func Ses_StoreTimesCopy(pTimesDest, pTimesSrc *int32, nVars int32) {
	if nVars < 0 || nVars > 8 || pTimesDest == nil || pTimesSrc == nil {
		return
	}
	copy(sesStoreTimes(pTimesDest, nVars), sesStoreTimes(pTimesSrc, nVars))
}

func Ses_StorePrintEntry(entry *Ses_TruthEntry_t, times *Ses_TimesEntry_t) {
	if entry == nil || times == nil {
		return
	}
	spFmt.Printf("f = ")
	Abc_TtPrintHexRev(os.Stdout, entry.pTruth[:], entry.nVars)
	spFmt.Printf(", n = %d", entry.nVars)
	spFmt.Printf(", arrival =")
	for i := int32(0); i < entry.nVars; i++ {
		spFmt.Printf(" %d", times.pArrTimeProfile[i])
	}
	spFmt.Printf("\n")
}

func Ses_StorePrintDebugEntry(store *Ses_Store_t, pTruth *word, nVars int32, pNormalArrTime *int32, nMaxDepth int32, pSol string, nStartGates int32) {
	if store == nil || store.pDebugEntries == nil {
		return
	}
	spFmt.Fprintf(store.pDebugEntries, "abc -c \"exact -v -C %d", store.nBTLimit)
	if s_pSesStore != nil && s_pSesStore.fMakeAIG != 0 {
		spFmt.Fprintf(store.pDebugEntries, " -a")
	}
	spFmt.Fprintf(store.pDebugEntries, " -S %d -D %d -A", nStartGates+1, nMaxDepth)
	for i, arrival := range sesStoreTimes(pNormalArrTime, nVars) {
		separator := ','
		if i == 0 {
			separator = ' '
		}
		spFmt.Fprintf(store.pDebugEntries, "%c%d", separator, arrival)
	}
	spFmt.Fprintf(store.pDebugEntries, " ")
	Abc_TtPrintHexRev(store.pDebugEntries, sesStoreTruthWords(pTruth, nVars), nVars)
	spFmt.Fprintf(store.pDebugEntries, "\" # ")
	if pSol == "" {
		spFmt.Fprintf(store.pDebugEntries, "no ")
	}
	spFmt.Fprintf(store.pDebugEntries, "solution found before\n")
}

func Abc_ExactNormalizeArrivalTimesForNetwork(nVars int32, pArrTimeProfile *int32, pSol string) {
	if nVars <= 0 || nVars > 8 || pArrTimeProfile == nil || len(pSol) <= ABC_EXACT_SOL_NGATES {
		return
	}
	nGates := int32(pSol[ABC_EXACT_SOL_NGATES])
	if nGates < 0 || nGates > 100 {
		return
	}
	need := 3 + int(nGates)*4
	if len(pSol) < need || nVars+nGates > 100 {
		return
	}
	levels := make([]int32, int(nVars+nGates))
	copy(levels, unsafe.Slice(pArrTimeProfile, int(nVars)))
	for i := int32(0); i < nGates; i++ {
		j, k := int32(pSol[3+i*4+2]), int32(pSol[3+i*4+3])
		if j < 0 || k < 0 || j >= nVars+i || k >= nVars+i {
			return
		}
		levels[nVars+i] = max32(levels[j], levels[k]) + 1
	}
	top := levels[nVars+nGates-1]
	for i := int32(0); i < nVars+nGates-1; i++ {
		levels[i] = top
	}
	for i := nGates - 1; i >= 0; i-- {
		j, k := int32(pSol[3+i*4+2]), int32(pSol[3+i*4+3])
		l := levels[nVars+i] - 1
		if levels[j] > l {
			levels[j] = l
		}
		if levels[k] > l {
			levels[k] = l
		}
	}
	maxNormalized := int32(0)
	Abc_NormalizeArrivalTimes(&levels[0], nVars, &maxNormalized)
	copy(unsafe.Slice(pArrTimeProfile, int(nVars)), levels[:nVars])
}
func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
func sesStoreWriteInt32(w io.Writer, value int32) error {
	var data [4]byte
	binary.NativeEndian.PutUint32(data[:], uint32(value))
	_, err := w.Write(data[:])
	return err
}

func sesStoreWriteWord(w io.Writer, value word) error {
	var data [8]byte
	binary.NativeEndian.PutUint64(data[:], uint64(value))
	_, err := w.Write(data[:])
	return err
}

func sesStoreReadInt32(r io.Reader, value *int32) error {
	var data [4]byte
	if _, err := io.ReadFull(r, data[:]); err != nil {
		return err
	}
	*value = int32(binary.NativeEndian.Uint32(data[:]))
	return nil
}

func sesStoreReadWord(r io.Reader, value *word) error {
	var data [8]byte
	if _, err := io.ReadFull(r, data[:]); err != nil {
		return err
	}
	*value = word(binary.NativeEndian.Uint64(data[:]))
	return nil
}

// sesStoreNetworkSize is the byte count used by the serialized C solution
// representation: three header bytes, four bytes per gate, two trailer bytes,
// and one output literal per variable.
func sesStoreNetworkSize(network string) (int, bool) {
	if len(network) < 3 {
		return 0, false
	}
	nVars, nGates := int(network[ABC_EXACT_SOL_NVARS]), int(network[ABC_EXACT_SOL_NGATES])
	size := 3 + 4*nGates + 2 + nVars
	return size, size >= 3 && size <= len(network)
}

func Ses_StoreWrite(store *Ses_Store_t, filename string, fSynthImp, fSynthRL, fUnsynthImp, fUnsynthRL int32) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		spFmt.Printf("cannot open file \"%s\" for writing\n", filename)
		return
	}
	defer file.Close()
	if store == nil {
		return
	}

	var nEntries uint64
	if fSynthImp != 0 {
		nEntries += store.nSynthesizedImp
	}
	if fSynthRL != 0 {
		nEntries += store.nSynthesizedRL
	}
	if fUnsynthImp != 0 {
		nEntries += store.nUnsynthesizedImp
	}
	if fUnsynthRL != 0 {
		nEntries += store.nUnsynthesizedRL
	}
	var header [8]byte
	binary.NativeEndian.PutUint64(header[:], nEntries)
	if _, err = file.Write(header[:]); err != nil {
		return
	}

	writeEntry := func(truth *Ses_TruthEntry_t, times *Ses_TimesEntry_t) bool {
		for _, wordValue := range truth.pTruth {
			if sesStoreWriteWord(file, wordValue) != nil {
				return false
			}
		}
		if sesStoreWriteInt32(file, truth.nVars) != nil {
			return false
		}
		for _, arrival := range times.pArrTimeProfile {
			if sesStoreWriteInt32(file, arrival) != nil {
				return false
			}
		}
		if sesStoreWriteInt32(file, times.fResLimit) != nil {
			return false
		}
		if times.pNetwork != "" {
			size, ok := sesStoreNetworkSize(times.pNetwork)
			if !ok {
				return false
			}
			_, err = file.Write([]byte(times.pNetwork[:size]))
			return err == nil
		}
		_, err = file.Write([]byte{0, 0, 0})
		return err == nil
	}

	for _, truth := range store.pEntries {
		for ; truth != nil; truth = truth.next {
			for times := truth.head; times != nil; times = times.next {
				hasNetwork := times.pNetwork != ""
				if !hasNetwork && times.fResLimit == 0 && fUnsynthImp == 0 || !hasNetwork && times.fResLimit != 0 && fUnsynthRL == 0 || hasNetwork && times.fResLimit == 0 && fSynthImp == 0 || hasNetwork && times.fResLimit != 0 && fSynthRL == 0 {
					continue
				}
				if !writeEntry(truth, times) {
					return
				}
			}
		}
	}
}

// pArrTimeProfile is normalized. Return 1 only when a new timing entry is added.
func Ses_StoreAddEntry(store *Ses_Store_t, pTruth *word, nVars int32, pArrTimeProfile *int32, pSol string, fResLimit int32) int32 {
	if store == nil || nVars < 0 || nVars > 8 || pTruth == nil || pArrTimeProfile == nil {
		return 0
	}
	if pSol != "" {
		Abc_ExactNormalizeArrivalTimesForNetwork(nVars, pArrTimeProfile, pSol)
	}
	key := Ses_StoreTableHash(pTruth, nVars)
	truthEntry := store.pEntries[key]
	for truthEntry != nil && Ses_StoreTruthEqual(truthEntry, pTruth, nVars) == 0 {
		truthEntry = truthEntry.next
	}
	if truthEntry == nil {
		truthEntry = &Ses_TruthEntry_t{}
		Ses_StoreTruthCopy(truthEntry, pTruth, nVars)
		truthEntry.next = store.pEntries[key]
		store.pEntries[key] = truthEntry
	}

	timesEntry := truthEntry.head
	for timesEntry != nil && Ses_StoreTimesEqual(pArrTimeProfile, &timesEntry.pArrTimeProfile[0], nVars) == 0 {
		timesEntry = timesEntry.next
	}
	added := int32(0)
	if timesEntry == nil {
		timesEntry = &Ses_TimesEntry_t{pNetwork: pSol, fResLimit: fResLimit}
		Ses_StoreTimesCopy(&timesEntry.pArrTimeProfile[0], pArrTimeProfile, nVars)
		timesEntry.next = truthEntry.head
		truthEntry.head = timesEntry
		added = 1
		store.nEntriesCount++
		if pSol != "" {
			store.nValidEntriesCount++
		}
	}

	if pSol != "" {
		if fResLimit != 0 {
			store.nSynthesizedRL++
			store.pSynthesizedRL[nVars]++
		} else {
			store.nSynthesizedImp++
			store.pSynthesizedImp[nVars]++
		}
	} else if fResLimit != 0 {
		store.nUnsynthesizedRL++
		store.pUnsynthesizedRL[nVars]++
	} else {
		store.nUnsynthesizedImp++
		store.pUnsynthesizedImp[nVars]++
	}
	if added != 0 && store.szDBName != "" {
		Ses_StoreWrite(store, store.szDBName, 1, 0, 0, 0)
	}
	return added
}

// pArrTimeProfile is normalized. A successful lookup may still return an empty solution.
func Ses_StoreGetEntrySimple(store *Ses_Store_t, pTruth *word, nVars int32, pArrTimeProfile *int32, pSol *string) int32 {
	if store == nil || pTruth == nil || pArrTimeProfile == nil || pSol == nil {
		return 0
	}
	key := Ses_StoreTableHash(pTruth, nVars)
	truthEntry := store.pEntries[key]
	for truthEntry != nil && Ses_StoreTruthEqual(truthEntry, pTruth, nVars) == 0 {
		truthEntry = truthEntry.next
	}
	if truthEntry == nil {
		return 0
	}
	for times := truthEntry.head; times != nil; times = times.next {
		if Ses_StoreTimesEqual(pArrTimeProfile, &times.pArrTimeProfile[0], nVars) != 0 {
			*pSol = times.pNetwork
			return 1
		}
	}
	return 0
}

func Ses_StoreGetEntry(store *Ses_Store_t, pTruth *word, nVars int32, pArrTimeProfile *int32, pSol *string) int32 {
	if store == nil || pTruth == nil || pArrTimeProfile == nil || pSol == nil || nVars < 0 || nVars > 8 {
		return 0
	}
	key := Ses_StoreTableHash(pTruth, nVars)
	truthEntry := store.pEntries[key]
	for truthEntry != nil && Ses_StoreTruthEqual(truthEntry, pTruth, nVars) == 0 {
		truthEntry = truthEntry.next
	}
	if truthEntry == nil {
		return 0
	}
	for times := truthEntry.head; times != nil; times = times.next {
		if times.pNetwork != "" {
			current := append([]int32(nil), sesStoreTimes(pArrTimeProfile, nVars)...)
			Abc_ExactNormalizeArrivalTimesForNetwork(nVars, &current[0], times.pNetwork)
			if Ses_StoreTimesEqual(&current[0], &times.pArrTimeProfile[0], nVars) != 0 {
				*pSol = times.pNetwork
				return 1
			}
		} else if Ses_StoreTimesEqual(pArrTimeProfile, &times.pArrTimeProfile[0], nVars) != 0 {
			*pSol = ""
			return 1
		}
	}
	return 0
}

func Ses_StoreRead(store *Ses_Store_t, filename string, fSynthImp, fSynthRL, fUnsynthImp, fUnsynthRL int32) {
	if store == nil {
		return
	}
	if store.szDBName != "" {
		spFmt.Printf("cannot read from database when szDBName is set")
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		spFmt.Printf("cannot open file \"%s\" for reading\n", filename)
		return
	}
	defer file.Close()
	var countBytes [8]byte
	if _, err := io.ReadFull(file, countBytes[:]); err != nil {
		return
	}
	count := binary.NativeEndian.Uint64(countBytes[:])
	for i := uint64(0); i < count; i++ {
		var truth [4]word
		for j := range truth {
			if err := sesStoreReadWord(file, &truth[j]); err != nil {
				return
			}
		}
		var nVars, resLimit int32
		if err := sesStoreReadInt32(file, &nVars); err != nil {
			return
		}
		var arrivals [8]int32
		for j := range arrivals {
			if err := sesStoreReadInt32(file, &arrivals[j]); err != nil {
				return
			}
		}
		if err := sesStoreReadInt32(file, &resLimit); err != nil {
			return
		}
		var header [3]byte
		if _, err := io.ReadFull(file, header[:]); err != nil {
			return
		}
		network := ""
		if header[0] != 0 {
			size := 3 + 4*int(header[ABC_EXACT_SOL_NGATES]) + 2 + int(header[ABC_EXACT_SOL_NVARS])
			if size < 3 || size > 1280 {
				return
			}
			data := make([]byte, size)
			copy(data, header[:])
			if _, err := io.ReadFull(file, data[3:]); err != nil {
				return
			}
			network = string(data)
		}
		if nVars < 0 || nVars > 8 {
			return
		}
		if network != "" && fSynthImp == 0 && resLimit == 0 || network != "" && fSynthRL == 0 && resLimit != 0 || network == "" && fUnsynthImp == 0 && resLimit == 0 || network == "" && fUnsynthRL == 0 && resLimit != 0 {
			continue
		}
		Ses_StoreAddEntry(store, &truth[0], nVars, &arrivals[0], network, resLimit)
	}
	spFmt.Printf("read %d entries from file\n", count)
}

// computes top decomposition of variables wrt. to AND and OR
func Ses_ManComputeTopDec(pSes *Ses_Man_t) {
	var mask [4]word
	Abc_TtMask(&mask[0], pSes.nSpecWords, pSes.nSpecWords*64)
	for level := int32(0); level < pSes.nSpecVars; level++ {
		if Abc_TtIsTopDecomposable(pSes.pSpec, &mask[0], pSes.nSpecWords, level) != 0 {
			pSes.pDecVars |= 1 << uint(level)
		}
	}
}

func Ses_ManAlloc(pTruth *word, nVars int32, nFunc int32, nMaxDepth int32, pArrTimeProfile *int32, fMakeAIG int32, nBTLimit int32, fVerbose int32) *Ses_Man_t {
	p := &Ses_Man_t{pSpec: pTruth, nSpecVars: nVars, nSpecFunc: nFunc, nMaxDepth: nMaxDepth, fMakeAIG: fMakeAIG, nBTLimit: nBTLimit, fVerbose: fVerbose}
	p.nSpecWords = Abc_TtWordNum(nVars)
	p.nRows = (int32(1) << uint(nVars)) - 1
	truths := unsafe.Slice(pTruth, int(nFunc*4))
	for h := int32(0); h < nFunc; h++ {
		if truths[h<<2]&1 != 0 {
			for i := int32(0); i < 4; i++ {
				truths[(h<<2)+i] = ^truths[(h<<2)+i]
			}
			p.bSpecInv |= 1 << uint(h)
		}
	}
	if nMaxDepth >= 0 {
		p.pArrTimeProfile = pArrTimeProfile
	}
	if p.pArrTimeProfile != nil {
		p.nArrTimeDelta = Abc_NormalizeArrivalTimes(p.pArrTimeProfile, nVars, &p.nArrTimeMax)
	}
	p.vPolar = Vec_IntAlloc(100)
	p.vAssump = Vec_IntAlloc(10)
	p.vStairDecVars = Vec_IntAlloc(nVars)
	p.nRandRowAssigns = 2 * nVars
	if nFunc == 1 {
		Ses_ManComputeTopDec(p)
	}
	p.rng = newGlibcRand(0xCAFE)
	return p
}

func Ses_ManCleanLight(pSes *Ses_Man_t) {
	if pSes == nil {
		return
	}
	truths := unsafe.Slice(pSes.pSpec, int(pSes.nSpecFunc*4))
	for h := int32(0); h < pSes.nSpecFunc; h++ {
		if (pSes.bSpecInv>>uint(h))&1 != 0 {
			for i := int32(0); i < 4; i++ {
				truths[(h<<2)+i] = ^truths[(h<<2)+i]
			}
		}
	}
	if pSes.pArrTimeProfile != nil {
		for i := int32(0); i < pSes.nSpecVars; i++ {
			unsafe.Slice(pSes.pArrTimeProfile, int(pSes.nSpecVars))[i] += pSes.nArrTimeDelta
		}
	}
	Vec_IntFree(pSes.vPolar)
	Vec_IntFree(pSes.vAssump)
	Vec_IntFree(pSes.vStairDecVars)
	pSes.vPolar, pSes.vAssump, pSes.vStairDecVars = nil, nil, nil
}

func Ses_ManClean(pSes *Ses_Man_t) {
	if pSes == nil {
		return
	}
	if pSes.pSat != nil {
		pSes.pSat.Close()
		pSes.pSat = nil
	}
	Ses_ManCleanLight(pSes)
}

/*
*Function*************************************************************

	Synopsis    [Access variables based on indexes.]

**********************************************************************
*/
func Ses_ManSimVar(pSes *Ses_Man_t, i int32, t int32) int32 {
	assert(i >= 0 && i < pSes.nGates, "i in gate range")
	assert(t >= 0 && t < pSes.nRows, "t in row range")
	return pSes.nSimOffset + pSes.nRows*i + t
}

func Ses_ManOutputVar(pSes *Ses_Man_t, h int32, i int32) int32 {
	assert(h >= 0 && h < pSes.nSpecFunc, "h in function range")
	assert(i >= 0 && i < pSes.nGates, "i in gate range")
	return pSes.nOutputOffset + pSes.nGates*h + i
}

func Ses_ManGateVar(pSes *Ses_Man_t, i int32, p int32, q int32) int32 {
	assert(i >= 0 && i < pSes.nGates, "i in gate range")
	assert(p >= 0 && p < 2 && q >= 0 && q < 2, "gate input values are bits")
	assert(p > 0 || q > 0, "gate truth index excludes 00")
	return pSes.nGateOffset + i*3 + (p << 1) + q - 1
}

func Ses_ManSelectVar(pSes *Ses_Man_t, i int32, j int32, k int32) int32 {
	assert(i >= 0 && i < pSes.nGates, "i in gate range")
	assert(j >= 0 && j < k && k < pSes.nSpecVars+i, "selection indices ordered and in range")
	offset := pSes.nSelectOffset
	for a := pSes.nSpecVars; a < pSes.nSpecVars+i; a++ {
		offset += a * (a - 1) / 2
	}
	return offset + (-j*(1+j-2*(pSes.nSpecVars+i)))/2 + (k - j - 1)
}

func Ses_ManDepthVar(pSes *Ses_Man_t, i int32, j int32) int32 {
	assert(i >= 0 && i < pSes.nGates, "i in gate range")
	assert(j >= 0 && j <= pSes.nArrTimeMax+i, "depth index in range")
	return pSes.nDepthOffset + i*pSes.nArrTimeMax + i*(i+1)/2 + j
}

/*
*Function*************************************************************

	Synopsis    [Compute truth table from a solution.]

**********************************************************************
*/
func Ses_ManDeriveTruth(pSes *Ses_Man_t, pSol string, fInvert int32) *word {
	assert(len(pSol) >= 3 && int32(pSol[ABC_EXACT_SOL_NFUNC]) == 1, "one-function solution")
	nGates := int32(pSol[ABC_EXACT_SOL_NGATES])
	assert(nGates > 0 && len(pSol) >= 4+int(nGates)*4, "complete solution encoding")
	for i := int32(0); i < 4*nGates; i++ {
		pSes.pTtObjs[i] = 0
	}
	for i := int32(0); i < nGates; i++ {
		pos := 3 + int(i)*4
		f, j, k := int32(pSol[pos]), int32(pSol[pos+2]), int32(pSol[pos+3])
		assert(pSol[pos+1] == 2, "binary gate encoding")
		assert(j >= 0 && j < pSes.nSpecVars+i && k > j && k < pSes.nSpecVars+i, "fanin indices ordered")
		var t0, t1 *word
		if j < pSes.nSpecVars {
			t0 = &s_Truths8[j<<2]
		} else {
			t0 = &pSes.pTtObjs[(j-pSes.nSpecVars)<<2]
		}
		if k < pSes.nSpecVars {
			t1 = &s_Truths8[k<<2]
		} else {
			t1 = &pSes.pTtObjs[(k-pSes.nSpecVars)<<2]
		}
		out := &pSes.pTtObjs[i<<2]
		for w := int32(0); w < pSes.nSpecWords; w++ {
			if f&1 != 0 {
				out[w] |= ^t0[w] & t1[w]
			}
			if f&2 != 0 {
				out[w] |= t0[w] & ^t1[w]
			}
			if f&4 != 0 {
				out[w] |= t0[w] & t1[w]
			}
		}
	}
	lit := int32(pSol[3+int(nGates)*4])
	assert(lit>>1 == nGates-1, "solution output refers to last gate")
	out := &pSes.pTtObjs[(nGates-1)<<2]
	if fInvert != 0 && lit&1 != 0 {
		for w := int32(0); w < pSes.nSpecWords; w++ {
			out[w] = ^out[w]
		}
	}
	return out
}

/*
*Function*************************************************************

	Synopsis    [Setup variables to find network with nGates gates.]

**********************************************************************
*/
func Ses_ManCreateVars(pSes *Ses_Man_t, nGates int32) {
	assert(nGates >= 0, "nonnegative gate count")
	if pSes.fSatVerbose != 0 {
		spFmt.Printf("create variables for network with %d functions over %d variables and %d/%d gates\n", pSes.nSpecFunc, pSes.nSpecVars, nGates, pSes.nMaxGates)
	}
	pSes.nGates = nGates
	pSes.nSimVars = nGates * pSes.nRows
	pSes.nOutputVars = pSes.nSpecFunc * nGates
	pSes.nGateVars = nGates * 3
	pSes.nSelectVars = 0
	for a := pSes.nSpecVars; a < pSes.nSpecVars+nGates; a++ {
		pSes.nSelectVars += a * (a - 1) / 2
	}
	pSes.nDepthVars = 0
	if pSes.nMaxDepth > 0 {
		pSes.nDepthVars = nGates*pSes.nArrTimeMax + nGates*(nGates+1)/2
	}
	pSes.nDepthOffset = 0
	pSes.nSelectOffset = pSes.nDepthVars
	pSes.nGateOffset = pSes.nDepthVars + pSes.nSelectVars
	pSes.nOutputOffset = pSes.nDepthVars + pSes.nSelectVars + pSes.nGateVars
	pSes.nSimOffset = pSes.nDepthVars + pSes.nSelectVars + pSes.nGateVars + pSes.nOutputVars
	if pSes.pSat == nil {
		pSes.pSat = bsat.NewSolver()
	} else {
		pSes.pSat.Reset()
	}
	pSes.pSat.SetNumVars(pSes.nSimVars + pSes.nOutputVars + pSes.nGateVars + pSes.nSelectVars + pSes.nDepthVars)
}

/*
*Function*************************************************************

	Synopsis    [Create clauses.]

**********************************************************************
*/
func Ses_ManGateCannotHaveChild(pSes *Ses_Man_t, i int32, j int32) {
	for k := int32(0); k < j; k++ {
		Vec_IntPush(pSes.vAssump, Ses_ManSelectVar(pSes, i, k, j)*2+1)
	}
	for k := j + 1; k < pSes.nSpecVars+i; k++ {
		Vec_IntPush(pSes.vAssump, Ses_ManSelectVar(pSes, i, j, k)*2+1)
	}
}

func Ses_ManCreateMainClause(pSes *Ses_Man_t, t int32, i int32, j int32, k int32, a int32, b int32, c int32) {
	var literals [5]bsat.Literal
	nLiterals := 0
	literals[nLiterals] = bsat.Literal(Ses_ManSelectVar(pSes, i, j, k)*2 + 1)
	nLiterals++
	literals[nLiterals] = bsat.Literal(Ses_ManSimVar(pSes, i, t)*2 + a)
	nLiterals++
	assignment := t + 1
	if j < pSes.nSpecVars {
		if int32((assignment>>uint(j))&1) != b {
			return
		}
	} else {
		literals[nLiterals] = bsat.Literal(Ses_ManSimVar(pSes, j-pSes.nSpecVars, t)*2 + b)
		nLiterals++
	}
	if k < pSes.nSpecVars {
		if int32((assignment>>uint(k))&1) != c {
			return
		}
	} else {
		literals[nLiterals] = bsat.Literal(Ses_ManSimVar(pSes, k-pSes.nSpecVars, t)*2 + c)
		nLiterals++
	}
	if b > 0 || c > 0 {
		literals[nLiterals] = bsat.Literal(Ses_ManGateVar(pSes, i, b, c)*2 + 1 - a)
		nLiterals++
	}
	pSes.pSat.AddClause(literals[:nLiterals]...)
}

func Ses_ManCreateTruthTableClause(pSes *Ses_Man_t, t int32) int32 {
	truths := unsafe.Slice(pSes.pSpec, int(pSes.nSpecFunc*4))
	for i := int32(0); i < pSes.nGates; i++ {
		for j := int32(0); j < pSes.nSpecVars+i; j++ {
			for k := j + 1; k < pSes.nSpecVars+i; k++ {
				Ses_ManCreateMainClause(pSes, t, i, j, k, 0, 0, 1)
				Ses_ManCreateMainClause(pSes, t, i, j, k, 0, 1, 0)
				Ses_ManCreateMainClause(pSes, t, i, j, k, 0, 1, 1)
				Ses_ManCreateMainClause(pSes, t, i, j, k, 1, 0, 0)
				Ses_ManCreateMainClause(pSes, t, i, j, k, 1, 0, 1)
				Ses_ManCreateMainClause(pSes, t, i, j, k, 1, 1, 0)
				Ses_ManCreateMainClause(pSes, t, i, j, k, 1, 1, 1)
			}
		}
		if pSes.nSpecFunc != 1 {
			for h := int32(0); h < pSes.nSpecFunc; h++ {
				truth := truths[h*4 : h*4+4]
				bit := int32((truth[(t+1)>>6] >> uint((t+1)&63)) & 1)
				literals := [2]bsat.Literal{bsat.Literal(Ses_ManOutputVar(pSes, h, i)*2 + 1), bsat.Literal(Ses_ManSimVar(pSes, i, t)*2 + 1 - bit)}
				if !pSes.pSat.AddClause(literals[:]...) {
					return 0
				}
			}
		}
	}
	if pSes.nSpecFunc == 1 {
		bit := int32((unsafe.Slice(pSes.pSpec, 4)[(t+1)>>6] >> uint((t+1)&63)) & 1)
		Vec_IntPush(pSes.vAssump, Ses_ManSimVar(pSes, pSes.nGates-1, t)*2+1-bit)
	}
	return 1
}

func Ses_ManCreateClauses(pSes *Ses_Man_t) int32 {
	lit := func(v, complement int32) bsat.Literal { return bsat.Literal(v*2 + complement) }
	add := func(clause ...bsat.Literal) bool { return pSes.pSat.AddClause(clause...) }
	for t := int32(0); t < pSes.nRows; t++ {
		if (pSes.pTtValues[t>>6]>>uint(t&63))&1 != 0 && Ses_ManCreateTruthTableClause(pSes, t) == 0 {
			return 0
		}
	}
	if pSes.nSpecFunc == 1 {
		for i := int32(0); i < pSes.nGates-1; i++ {
			Vec_IntPush(pSes.vAssump, int32(lit(Ses_ManOutputVar(pSes, 0, i), 1)))
		}
		Vec_IntPush(pSes.vAssump, int32(lit(Ses_ManOutputVar(pSes, 0, pSes.nGates-1), 0)))
	} else {
		for h := int32(0); h < pSes.nSpecFunc; h++ {
			clause := make([]bsat.Literal, pSes.nGates)
			for i := int32(0); i < pSes.nGates; i++ {
				clause[i] = lit(Ses_ManOutputVar(pSes, h, i), 0)
			}
			add(clause...)
		}
	}
	for i := int32(0); i < pSes.nGates; i++ {
		clause := make([]bsat.Literal, 0, (pSes.nSpecVars+i)*(pSes.nSpecVars+i-1)/2)
		for j := int32(0); j < pSes.nSpecVars+i; j++ {
			for k := j + 1; k < pSes.nSpecVars+i; k++ {
				clause = append(clause, lit(Ses_ManSelectVar(pSes, i, j, k), 0))
			}
		}
		add(clause...)
	}
	if pSes.fMakeAIG != 0 {
		for i := int32(0); i < pSes.nGates; i++ {
			x, y, z := Ses_ManGateVar(pSes, i, 0, 1), Ses_ManGateVar(pSes, i, 1, 0), Ses_ManGateVar(pSes, i, 1, 1)
			add(lit(x, 1), lit(y, 1), lit(z, 0))
			add(lit(x, 1), lit(y, 0), lit(z, 1))
			add(lit(x, 0), lit(y, 1), lit(z, 1))
		}
	}
	// Binary gate clauses and pairwise at-most-one constraints, in C source order.
	for i := int32(0); i < pSes.nGates; i++ {
		x, y, z := Ses_ManGateVar(pSes, i, 0, 1), Ses_ManGateVar(pSes, i, 1, 0), Ses_ManGateVar(pSes, i, 1, 1)
		add(lit(x, 0), lit(y, 0), lit(z, 0))
		add(lit(x, 1), lit(y, 0), lit(z, 1))
		add(lit(x, 0), lit(y, 1), lit(z, 1))
	}
	for i := int32(0); i < pSes.nGates; i++ {
		for k := int32(1); k < pSes.nSpecVars+i; k++ {
			for j := int32(0); j < k; j++ {
				first := lit(Ses_ManSelectVar(pSes, i, j, k), 1)
				for kk := int32(1); kk < pSes.nSpecVars+i; kk++ {
					for jj := int32(0); jj < kk; jj++ {
						if k == kk && j == jj {
							continue
						}
						second := lit(Ses_ManSelectVar(pSes, i, jj, kk), 1)
						if first < second {
							add(first, second)
						}
					}
				}
			}
		}
	}
	for i := int32(0); i < Vec_IntSize(pSes.vStairDecVars); i++ {
		j := Vec_IntEntry(pSes.vStairDecVars, i)
		if pSes.nGates-2-i < j {
			continue
		}
		gate := pSes.nGates - 1 - i
		Vec_IntPush(pSes.vAssump, int32(lit(Ses_ManSelectVar(pSes, gate, j, pSes.nSpecVars+pSes.nGates-2-i), 0)))
		switch pSes.pStairDecFunc[i] {
		case 1:
			Vec_IntPush(pSes.vAssump, int32(lit(Ses_ManGateVar(pSes, gate, 0, 1), 1)))
		case 2:
			Vec_IntPush(pSes.vAssump, int32(lit(Ses_ManGateVar(pSes, gate, 1, 0), 1)))
			Vec_IntPush(pSes.vAssump, int32(lit(Ses_ManGateVar(pSes, gate, 1, 1), 1)))
		case 3:
			Vec_IntPush(pSes.vAssump, int32(lit(Ses_ManGateVar(pSes, gate, 1, 0), 0)))
			Vec_IntPush(pSes.vAssump, int32(lit(Ses_ManGateVar(pSes, gate, 1, 1), 0)))
		case 4:
		case 5:
			Vec_IntPush(pSes.vAssump, int32(lit(Ses_ManGateVar(pSes, gate, 0, 1), 0)))
			Vec_IntPush(pSes.vAssump, int32(lit(Ses_ManGateVar(pSes, gate, 1, 0), 0)))
			Vec_IntPush(pSes.vAssump, int32(lit(Ses_ManGateVar(pSes, gate, 1, 1), 1)))
		default:
			panic("invalid stair decomposition function")
		}
	}
	for i := int32(0); i < pSes.nGates; i++ {
		clause := make([]bsat.Literal, 0)
		for h := int32(0); h < pSes.nSpecFunc; h++ {
			clause = append(clause, lit(Ses_ManOutputVar(pSes, h, i), 0))
		}
		for ii := i + 1; ii < pSes.nGates; ii++ {
			for j := int32(0); j < pSes.nSpecVars+i; j++ {
				clause = append(clause, lit(Ses_ManSelectVar(pSes, ii, j, pSes.nSpecVars+i), 0))
			}
			for j := pSes.nSpecVars + i + 1; j < pSes.nSpecVars+ii; j++ {
				clause = append(clause, lit(Ses_ManSelectVar(pSes, ii, pSes.nSpecVars+i, j), 0))
			}
		}
		add(clause...)
	}
	for i := int32(0); i < pSes.nGates-1; i++ {
		for ii := i + 1; ii < pSes.nGates; ii++ {
			for k := int32(1); k < pSes.nSpecVars+i; k++ {
				for j := int32(0); j < k; j++ {
					first := lit(Ses_ManSelectVar(pSes, i, j, k), 1)
					add(first, lit(Ses_ManSelectVar(pSes, ii, j, pSes.nSpecVars+i), 1))
					add(first, lit(Ses_ManSelectVar(pSes, ii, k, pSes.nSpecVars+i), 1))
				}
			}
		}
	}
	for i := int32(0); i < pSes.nGates-2; i++ {
		for ii := i + 1; ii < pSes.nGates-1; ii++ {
			for iii := ii + 1; iii < pSes.nGates; iii++ {
				for k := int32(1); k < pSes.nSpecVars+i; k++ {
					for j := int32(0); j < k; j++ {
						add(lit(Ses_ManSelectVar(pSes, i, j, k), 1), lit(Ses_ManSelectVar(pSes, ii, j, k), 1), lit(Ses_ManSelectVar(pSes, iii, i, ii), 1))
					}
				}
			}
		}
	}
	for i := int32(0); i < pSes.nGates-1; i++ {
		for k := int32(2); k < pSes.nSpecVars+i; k++ {
			for j := int32(1); j < k; j++ {
				for jj := int32(0); jj < j; jj++ {
					add(lit(Ses_ManSelectVar(pSes, i, j, k), 1), lit(Ses_ManSelectVar(pSes, i+1, jj, k), 1))
				}
			}
			for j := int32(0); j < k; j++ {
				for kk := int32(1); kk < k; kk++ {
					for jj := int32(0); jj < kk; jj++ {
						add(lit(Ses_ManSelectVar(pSes, i, j, k), 1), lit(Ses_ManSelectVar(pSes, i+1, jj, kk), 1))
					}
				}
			}
		}
	}
	// Symmetric input variables are ordered by forcing the earlier variable into earlier gates.
	if pSes.nSpecFunc == 1 {
		truth := unsafe.Slice(pSes.pSpec, int(pSes.nSpecWords))
		var arrivals []int32
		if pSes.pArrTimeProfile != nil {
			arrivals = unsafe.Slice(pSes.pArrTimeProfile, int(pSes.nSpecVars))
		}
		for q := int32(1); q < pSes.nSpecVars; q++ {
			for p := int32(0); p < q; p++ {
				symmetric := true
				for assignment := uint32(0); assignment < uint32(1)<<uint(pSes.nSpecVars); assignment++ {
					swapped := assignment
					bp, bq := (assignment>>uint(p))&1, (assignment>>uint(q))&1
					if bp != bq {
						swapped ^= (1 << uint(p)) | (1 << uint(q))
					}
					if ((truth[assignment>>6] >> uint(assignment&63)) & 1) != ((truth[swapped>>6] >> uint(swapped&63)) & 1) {
						symmetric = false
						break
					}
				}
				if !symmetric || (arrivals != nil && arrivals[p] > arrivals[q]) {
					continue
				}
				if pSes.fSatVerbose != 0 {
					spFmt.Printf("variables %d and %d are symmetric\n", p, q)
				}
				for i := int32(0); i < pSes.nGates; i++ {
					for j := int32(0); j < q; j++ {
						if j == p {
							continue
						}
						clause := []bsat.Literal{lit(Ses_ManSelectVar(pSes, i, j, q), 1)}
						for ii := int32(0); ii < i; ii++ {
							for kk := int32(1); kk < pSes.nSpecVars+ii; kk++ {
								for jj := int32(0); jj < kk; jj++ {
									if jj == p || kk == p {
										clause = append(clause, lit(Ses_ManSelectVar(pSes, ii, jj, kk), 0))
									}
								}
							}
						}
						add(clause...)
					}
				}
			}
		}
	}
	return 1
}

func Ses_ManCreateDepthClauses(pSes *Ses_Man_t) int32 {
	lit := func(v, complement int32) bsat.Literal { return bsat.Literal(v*2 + complement) }
	add := func(clause ...bsat.Literal) bool { return pSes.pSat.AddClause(clause...) }
	var arrivals []int32
	if pSes.pArrTimeProfile != nil {
		arrivals = unsafe.Slice(pSes.pArrTimeProfile, int(pSes.nSpecVars))
	}
	for i := int32(0); i < pSes.nGates; i++ {
		for k := int32(1); k < i; k++ {
			for j := int32(0); j < k; j++ {
				selected := lit(Ses_ManSelectVar(pSes, i, pSes.nSpecVars+j, pSes.nSpecVars+k), 1)
				for jj := int32(0); jj <= pSes.nArrTimeMax+j; jj++ {
					add(selected, lit(Ses_ManDepthVar(pSes, j, jj), 1), lit(Ses_ManDepthVar(pSes, i, jj+1), 0))
				}
			}
		}
		for k := int32(0); k < i; k++ {
			for j := int32(0); j < pSes.nSpecVars+k; j++ {
				selected := lit(Ses_ManSelectVar(pSes, i, j, pSes.nSpecVars+k), 1)
				for kk := int32(0); kk <= pSes.nArrTimeMax+k; kk++ {
					add(selected, lit(Ses_ManDepthVar(pSes, k, kk), 1), lit(Ses_ManDepthVar(pSes, i, kk+1), 0))
				}
			}
		}
		if arrivals != nil {
			for k := int32(1); k < pSes.nSpecVars+i; k++ {
				limit := k
				if limit > pSes.nSpecVars {
					limit = pSes.nSpecVars
				}
				for j := int32(0); j < limit; j++ {
					d := arrivals[j]
					if k < pSes.nSpecVars && arrivals[k] > d {
						d = arrivals[k]
					}
					add(lit(Ses_ManSelectVar(pSes, i, j, k), 1), lit(Ses_ManDepthVar(pSes, i, d), 0))
				}
			}
		} else {
			Vec_IntPush(pSes.vAssump, int32(lit(Ses_ManDepthVar(pSes, i, 0), 0)))
		}
		for j := int32(1); j <= pSes.nArrTimeMax+i; j++ {
			add(lit(Ses_ManDepthVar(pSes, i, j), 1), lit(Ses_ManDepthVar(pSes, i, j-1), 0))
		}
		if pSes.nMaxDepth < pSes.nArrTimeMax+i {
			for h := int32(0); h < pSes.nSpecFunc; h++ {
				if !add(lit(Ses_ManOutputVar(pSes, h, i), 1), lit(Ses_ManDepthVar(pSes, i, pSes.nMaxDepth), 1)) {
					return 0
				}
			}
		}
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Solve.]

**********************************************************************
*/
func Sat_Wrd2Dbl(w word) float64 {
	return float64(uint32((w & 0x3FFFFFFF))) + float64((1<<30))*uint32((w>>30))
}

func Ses_ManSolve(pSes *Ses_Man_t) int32 {
	statsBefore := pSes.pSat.RunStats()
	if pSes.fSatVerbose != 0 {
		spFmt.Printf("SAT   CL: %7d   VA: %5d", statsBefore.Clauses, pSes.pSat.NumVars())
		_ = os.Stdout.Sync()
	}

	start := Abc_Clock()
	status := pSes.pSat.SolveWithResourceLimits(
		vec.Vec_IntArray(pSes.vAssump)[:vec.Vec_IntSize(pSes.vAssump)],
		int64(pSes.nBTLimit), 0, 0, 0,
	)
	delta := Abc_Clock() - start
	if pSes.fSatVerbose != 0 {
		stats := pSes.pSat.RunStats()
		spFmt.Printf("   RE:   %2d   ST: %4.0f   CO: %7.0f   DE: %6.0f    PR: %6.0f\n",
			int32(status), Sat_Wrd2Dbl(word(stats.Starts)), Sat_Wrd2Dbl(word(stats.Conflicts)),
			Sat_Wrd2Dbl(word(stats.Decisions)), Sat_Wrd2Dbl(word(stats.Propagations)))
	}

	pSes.timeSat += delta
	switch status {
	case bsat.Satisfiable:
		pSes.nSatCalls++
		pSes.timeSatSat += delta
		return 1
	case bsat.Unsatisfiable:
		pSes.nUnsatCalls++
		pSes.timeSatUnsat += delta
		return 0
	default:
		pSes.nUndefCalls++
		pSes.timeSatUndef += delta
		if pSes.fSatVerbose != 0 {
			spFmt.Printf("resource limit reached\n")
		}
		return 2
	}
}

/**Function*************************************************************

  Synopsis    [Extract solution.]

***********************************************************************/
// char is an array of short integers that stores the synthesized network
// using the following format
// | nvars | nfunc | ngates | gate[1] | ... | gate[ngates] | func[1] | .. | func[nfunc] |
// nvars:       integer with number of variables
// nfunc:       integer with number of functions
// ngates:      integer with number of gates
// gate[i]:     | op | nfanin | fanin[1] | ... | fanin[nfanin] |
//   op:        integer of gate's truth table (divided by 2, because gate is normal)
//   nfanin:    integer with number of fanins
//   fanin[i]:  integer to primary input or other gate
// func[i]:     | fanin | delay | pin[1] | ... | pin[nvars] |
//   fanin:     integer as literal to some gate (not primary input), can be complemented
//   delay:     maximum delay to output (taking arrival times into account, not normalized) or 0 if not specified
//   pin[i]:    pin to pin delay to input i or 0 if not specified or if there is no connection to input i
// NOTE: since outputs can only point to gates, delay and pin-to-pin times cannot be 0
func Ses_ManExtractSolution(pSes *Ses_Man_t) string {
	value := func(v int32) int32 {
		b, _ := pSes.pSat.Value(v)
		if b {
			return 1
		}
		return 0
	}
	nSol := 3 + pSes.nGates*4 + pSes.nSpecFunc*(2+pSes.nSpecVars)
	buf := make([]byte, 0, int(nSol))
	buf = append(buf, byte(pSes.nSpecVars), byte(pSes.nSpecFunc), byte(pSes.nGates))
	for i := int32(0); i < pSes.nGates; i++ {
		op := value(Ses_ManGateVar(pSes, i, 0, 1)) | value(Ses_ManGateVar(pSes, i, 1, 0))<<1 | value(Ses_ManGateVar(pSes, i, 1, 1))<<2
		buf = append(buf, byte(op), 2)
		if pSes.fExtractVerbose != 0 {
			spFmt.Printf("add gate %d with operation %d", pSes.nSpecVars+i, op)
		}
		found := false
		for k := int32(0); k < pSes.nSpecVars+i && !found; k++ {
			for j := int32(0); j < k; j++ {
				if value(Ses_ManSelectVar(pSes, i, j, k)) != 0 {
					if pSes.fExtractVerbose != 0 {
						spFmt.Printf(" and operands %d and %d", j, k)
					}
					buf = append(buf, byte(j), byte(k))
					found = true
					break
				}
			}
		}
		assert(found, "each gate has exactly one selected operand pair")
		if pSes.fExtractVerbose != 0 {
			if pSes.nMaxDepth > 0 {
				spFmt.Print(" and depth vector ")
				for j := int32(0); j <= pSes.nArrTimeMax+i; j++ {
					spFmt.Print(value(Ses_ManDepthVar(pSes, i, j)))
				}
			}
			spFmt.Print("\n")
		}
	}
	var perm []int32
	var arrivals []int32
	if pSes.nMaxDepth != -1 {
		perm = make([]int32, int(pSes.nGates*pSes.nSpecVars))
		if pSes.pArrTimeProfile != nil {
			arrivals = unsafe.Slice(pSes.pArrTimeProfile, int(pSes.nSpecVars))
		}
		for i := int32(0); i < pSes.nGates; i++ {
			pos := 3 + int(i)*4
			j, k := int32(buf[pos+2]), int32(buf[pos+3])
			for l := int32(0); l < pSes.nSpecVars; l++ {
				aj, ak := int32(0), int32(0)
				if j >= pSes.nSpecVars {
					aj = perm[(j-pSes.nSpecVars)*pSes.nSpecVars+l]
				}
				if k >= pSes.nSpecVars {
					ak = perm[(k-pSes.nSpecVars)*pSes.nSpecVars+l]
				}
				if aj == 0 && ak == 0 {
					if l == j || l == k {
						perm[i*pSes.nSpecVars+l] = 1
					}
				} else {
					perm[i*pSes.nSpecVars+l] = max32(aj, ak) + 1
				}
			}
		}
	}
	for h := int32(0); h < pSes.nSpecFunc; h++ {
		for i := int32(0); i < pSes.nGates; i++ {
			if value(Ses_ManOutputVar(pSes, h, i)) == 0 {
				continue
			}
			buf = append(buf, byte(i*2+((pSes.bSpecInv>>uint(h))&1)))
			d := int32(0)
			if pSes.nMaxDepth != -1 {
				for l := int32(0); l < pSes.nSpecVars; l++ {
					pin := perm[i*pSes.nSpecVars+l]
					if arrivals != nil {
						pin += arrivals[l]
					}
					d = max32(d, pin)
				}
			}
			buf = append(buf, byte(d))
			if arrivals != nil && pSes.fExtractVerbose != 0 {
				spFmt.Printf("output %d points to gate %d and has normalized delay %d (nArrTimeDelta = %d)\n", h, pSes.nSpecVars+i, d, pSes.nArrTimeDelta)
			}
			for l := int32(0); l < pSes.nSpecVars; l++ {
				pin := int32(0)
				if pSes.nMaxDepth != -1 {
					pin = perm[i*pSes.nSpecVars+l]
				}
				if arrivals != nil && pSes.fExtractVerbose != 0 {
					spFmt.Printf("  pin-to-pin arrival time from input %d is %d (pArrTimeProfile = %d)\n", l, pin, arrivals[l])
				}
				buf = append(buf, byte(pin))
			}
		}
	}
	assert(len(buf) == int(nSol), "solution fields match allocated encoding")
	return string(buf)
}

/*
*Function*************************************************************

	Synopsis    [Debug.]

**********************************************************************
*/
func Ses_ManExtractNtk(pSol string) *baseabc.Network {
	nVars, nFuncs, nGates := int32(pSol[ABC_EXACT_SOL_NVARS]), int32(pSol[ABC_EXACT_SOL_NFUNC]), int32(pSol[ABC_EXACT_SOL_NGATES])
	ntk := baseabc.Abc_NtkAlloc(baseabc.ABC_NTK_LOGIC, baseabc.ABC_FUNC_SOP, 1)
	ntk.SetName("exact")
	gates := vec.Vec_PtrAlloc(int(nVars + nGates))
	names := baseabc.Abc_NodeGetFakeNames(nVars + nFuncs)
	for i := int32(0); i < nVars; i++ {
		pi := baseabc.Abc_NtkCreatePi(ntk)
		baseabc.Abc_ObjAssignName(pi, vec.Vec_PtrEntry(names, int(i)).(string), "")
		vec.Vec_PtrPush(gates, pi)
	}
	for i := int32(0); i < nGates; i++ {
		pos := 3 + int(i)*4
		op := pSol[pos]
		truth := string([]byte{'0' + ((op >> 2) & 1), '0' + ((op >> 1) & 1), '0' + (op & 1), '0'})
		obj := baseabc.Abc_NtkCreateNode(ntk)
		obj.Data = baseabc.Abc_SopRegister(nil, baseabc.Abc_SopFromTruthBin(truth))
		vec.Vec_PtrPush(gates, obj)
		assert(pSol[pos+1] == 2, "binary gate encoding")
		baseabc.Abc_ObjAddFanin(obj, vec.Vec_PtrEntry(gates, int(pSol[pos+2])).(*baseabc.Object))
		baseabc.Abc_ObjAddFanin(obj, vec.Vec_PtrEntry(gates, int(pSol[pos+3])).(*baseabc.Object))
	}
	pos := 3 + int(nGates)*4
	for h := int32(0); h < nFuncs; h++ {
		po := baseabc.Abc_NtkCreatePo(ntk)
		baseabc.Abc_ObjAssignName(po, vec.Vec_PtrEntry(names, int(nVars+h)).(string), "")
		lit := int32(pSol[pos])
		target := vec.Vec_PtrEntry(gates, int(nVars+(lit>>1))).(*baseabc.Object)
		if lit&1 != 0 {
			target = baseabc.Abc_NtkCreateNodeInv(ntk, target)
		}
		baseabc.Abc_ObjAddFanin(po, target)
		pos += 2 + int(nVars)
	}
	baseabc.Abc_NodeFreeNames(names)
	vec.Vec_PtrFree(gates)
	if baseabc.Abc_NtkCheck(ntk) == 0 {
		spFmt.Printf("Ses_ManExtractSolution(): Network check has failed.\n")
	}
	return ntk
}

func Ses_ManExtractGia(pSol string) *gia.Gia_Man_t {
	nVars, nFuncs, nGates := int32(pSol[ABC_EXACT_SOL_NVARS]), int32(pSol[ABC_EXACT_SOL_NFUNC]), int32(pSol[ABC_EXACT_SOL_NGATES])
	g := gia.Gia_ManStart(nVars + nGates + nFuncs + 1)
	gia.Gia_ManSetConstrNum(g, 0)
	gia.Gia_ManSetNameSpec(g, "exact", "")
	gates := vec.Vec_IntAlloc(int(nVars + nGates))
	names := baseabc.Abc_NodeGetFakeNames(nVars + nFuncs)
	inputNames := vec.Vec_PtrStart(int(nVars))
	gia.Gia_ManSetInputNames(g, inputNames)
	for i := int32(0); i < nVars; i++ {
		id := gia.Gia_ManAppendCi(g)
		vec.Vec_IntPush(gates, id)
		vec.Vec_PtrSetEntry(inputNames, int(i), vec.Vec_PtrEntry(names, int(i)).(string))
	}
	for i := int32(0); i < nGates; i++ {
		pos := 3 + int(i)*4
		op := pSol[pos]
		assert(pSol[pos+1] == 2, "binary gate encoding")
		a, b := vec.Vec_IntEntry(gates, int(pSol[pos+2])), vec.Vec_IntEntry(gates, int(pSol[pos+3]))
		if op&1 != 0 {
			a ^= 1
		}
		if (op>>1)&1 != 0 {
			b ^= 1
		}
		out := gia.Gia_ManAppendAnd(g, a, b)
		if op&1 != 0 && (op>>1)&1 != 0 {
			assert((op>>2)&1 != 0, "normalized gate output phase")
			out ^= 1
		}
		vec.Vec_IntPush(gates, out)
	}
	pos := 3 + int(nGates)*4
	outputNames := vec.Vec_PtrStart(int(nFuncs))
	gia.Gia_ManSetOutputNames(g, outputNames)
	for h := int32(0); h < nFuncs; h++ {
		lit := int32(pSol[pos])
		out := vec.Vec_IntEntry(gates, int(nVars+(lit>>1)))
		if lit&1 != 0 {
			out ^= 1
		}
		gia.Gia_ManAppendCo(g, out)
		vec.Vec_PtrSetEntry(outputNames, int(h), vec.Vec_PtrEntry(names, int(nVars+h)).(string))
		pos += 2 + int(nVars)
	}
	baseabc.Abc_NodeFreeNames(names)
	vec.Vec_IntFree(gates)
	return g
}

func Ses_ManPrintRuntime(pSes *Ses_Man_t) {
	spFmt.Printf("Runtime breakdown:\n")
	print := func(label string, elapsed int64) {
		percent := 0.0
		if pSes.timeTotal != 0 {
			percent = 100 * float64(elapsed) / float64(pSes.timeTotal)
		}
		spFmt.Printf("%s =", label)
		spFmt.Printf("%9.2f sec (%6.2f %%)\n", float64(elapsed)/float64(time.Second), percent)
	}
	print("Sat     ", pSes.timeSat)
	print(" Sat    ", pSes.timeSatSat)
	print(" Unsat  ", pSes.timeSatUnsat)
	print(" Undef  ", pSes.timeSatUndef)
	print("Instance", pSes.timeInstance)
	print("ALL     ", pSes.timeTotal)
}

func Ses_ManPrintFuncs(pSes *Ses_Man_t) {
	var h int32
	truths := unsafe.Slice(pSes.pSpec, int(pSes.nSpecFunc*4))

	spFmt.Printf("find optimum circuit for %d %d-variable functions:\n", pSes.nSpecFunc, pSes.nSpecVars)
	for h = 0; h < pSes.nSpecFunc; h++ {
		spFmt.Printf("  func %d: ", h+1)
		Abc_TtPrintHexRev(os.Stdout, &truths[h<<2], pSes.nSpecVars)
		spFmt.Printf("\n")
	}

	if pSes.nMaxDepth != -1 {
		spFmt.Printf("  max depth = %d\n", pSes.nMaxDepth)
		if pSes.pArrTimeProfile != nil {
			arrivals := unsafe.Slice(pSes.pArrTimeProfile, int(pSes.nSpecVars))
			spFmt.Printf("  arrival times =")
			for h = 0; h < pSes.nSpecVars; h++ {
				spFmt.Printf(" %d", arrivals[h])
			}
			spFmt.Printf("\n")
		}
	}
}

func Ses_ManPrintVars(pSes *Ses_Man_t) {
	var h, i, j, k, p, q, t int32

	for i = 0; i < pSes.nGates; i++ {
		for t = 0; t < pSes.nRows; t++ {
			spFmt.Printf("x(%d, %d) : %d\n", i, t, Ses_ManSimVar(pSes, i, t))
		}
	}

	for h = 0; h < pSes.nSpecFunc; h++ {
		for i = 0; i < pSes.nGates; i++ {
			spFmt.Printf("h(%d, %d) : %d\n", h, i, Ses_ManOutputVar(pSes, h, i))
		}
	}

	for i = 0; i < pSes.nGates; i++ {
		for p = 0; p <= 1; p++ {
			for q = 0; q <= 1; q++ {
				if p == 0 && q == 0 {
					continue
				}
				spFmt.Printf("f(%d, %d, %d) : %d\n", i, p, q, Ses_ManGateVar(pSes, i, p, q))
			}
		}
	}

	for i = 0; i < pSes.nGates; i++ {
		for j = 0; j < pSes.nSpecVars+i; j++ {
			for k = j + 1; k < pSes.nSpecVars+i; k++ {
				spFmt.Printf("s(%d, %d, %d) : %d\n", i, j, k, Ses_ManSelectVar(pSes, i, j, k))
			}
		}
	}

	if pSes.nMaxDepth > 0 {
		for i = 0; i < pSes.nGates; i++ {
			for j = 0; j <= i; j++ {
				spFmt.Printf("d(%d, %d) : %d\n", i, j, Ses_ManDepthVar(pSes, i, j))
			}
		}
	}

}

/*
*Function*************************************************************

	Synopsis    [Synthesis algorithm.]

**********************************************************************
*/
func Ses_ManComputeMaxGates(pSes *Ses_Man_t) {
	arrivals := unsafe.Slice(pSes.pArrTimeProfile, int(pSes.nSpecVars))
	s, level, available := int32(1), pSes.nMaxDepth, pSes.nSpecVars
	pSes.nMaxGates = 0
	for s != 0 && level != 0 && available*2 > s {
		for l := int32(0); l < pSes.nSpecVars; l++ {
			if arrivals[l] == level {
				s--
				available--
			}
		}
		level--
		pSes.nMaxGates += s
		s *= 2
	}
}

func Ses_CheckDepthConsistency(pSes *Ses_Man_t) int32 {
	arrivals := unsafe.Slice(pSes.pArrTimeProfile, int(pSes.nSpecVars))
	Vec_IntClear(pSes.vStairDecVars)
	pSes.fDecStructure = 0
	for l := int32(0); l < pSes.nSpecVars; l++ {
		if arrivals[l] >= pSes.nMaxDepth {
			if pSes.fReasonVerbose != 0 {
				spFmt.Printf("give up due to impossible arrival time (depth = %d, input = %d, arrival time = %d)", pSes.nMaxDepth, l, arrivals[l])
			}
			return 0
		} else if pSes.nSpecFunc == 1 && arrivals[l]+1 == pSes.nMaxDepth {
			if (pSes.fDecStructure == 1 && pSes.nSpecVars > 2) || pSes.fDecStructure == 2 {
				if pSes.fReasonVerbose != 0 {
					spFmt.Printf("give up due to impossible decomposition (depth = %d, input = %d, arrival time = %d)", pSes.nMaxDepth, l, arrivals[l])
				}
				return 0
			}
			pSes.fDecStructure++
			if (pSes.pDecVars>>uint(l))&1 == 0 {
				if pSes.fReasonVerbose != 0 {
					spFmt.Printf("give up due to impossible decomposition (depth = %d, input = %d, arrival time = %d)", pSes.nMaxDepth, l, arrivals[l])
				}
				return 0
			}
		}
	}
	if pSes.fDecStructure != 0 {
		level := int32(1)
		for {
			added := int32(0)
			for l := int32(0); l < pSes.nSpecVars; l++ {
				if arrivals[l]+level == pSes.nMaxDepth {
					if added != 0 {
						assert(level == int32(Vec_IntSize(pSes.vStairDecVars)), "stair level sequence")
						if added > 1 || level+1 < pSes.nSpecVars {
							if pSes.fReasonVerbose != 0 {
								spFmt.Printf("give up due to impossible decomposition at level %d", level)
							}
							return 0
						}
					} else {
						Vec_IntPush(pSes.vStairDecVars, l)
					}
					added++
				}
			}
			if added == 0 {
				break
			}
			level++
		}
		if Vec_IntSize(pSes.vStairDecVars) > 0 {
			vars := vec.Vec_IntArray(pSes.vStairDecVars)
			if Abc_TtIsStairDecomposable(pSes.pSpec, pSes.nSpecWords, &vars[0], int32(len(vars)), &pSes.pStairDecFunc[0]) == 0 {
				if pSes.fReasonVerbose != 0 {
					spFmt.Printf("give up due to impossible stair decomposition at level %d: ", level)
					Vec_IntPrint(pSes.vStairDecVars)
				}
				return 0
			}
		}
	}
	maxLevel2 := int32(3)
	if pSes.fDecStructure != 0 {
		if pSes.nSpecVars == 3 {
			maxLevel2 = 2
		} else {
			maxLevel2 = 1
		}
	} else if pSes.nSpecVars == 4 {
		maxLevel2 = 4
	}
	count := int32(0)
	for l := int32(0); l < pSes.nSpecVars; l++ {
		if arrivals[l]+2 == pSes.nMaxDepth {
			count++
			if count > maxLevel2 {
				if pSes.fReasonVerbose != 0 {
					spFmt.Printf("give up due to impossible decomposition at second level (depth = %d, input = %d, arrival time = %d)", pSes.nMaxDepth, l, arrivals[l])
				}
				return 0
			}
		}
	}
	if pSes.nSpecVars > 4 && pSes.fDecStructure != 0 && count == 1 {
		count = 0
		for l := int32(0); l < pSes.nSpecVars; l++ {
			if arrivals[l]+3 == pSes.nMaxDepth {
				count++
				if count > 1 {
					if pSes.fReasonVerbose != 0 {
						spFmt.Printf("give up due to impossible decomposition at third level (depth = %d, input = %d, arrival time = %d)", pSes.nMaxDepth, l, arrivals[l])
					}
					return 0
				}
			}
		}
	}
	return 1
}

func Ses_CheckGatesConsistency(pSes *Ses_Man_t, nGates int32) int32 {
	if pSes.nMaxDepth != -1 && nGates >= int32(1)<<uint(pSes.nMaxDepth) {
		if pSes.fReasonVerbose != 0 {
			spFmt.Printf("give up due to impossible depth (depth = %d, gates = %d)", pSes.nMaxDepth, nGates)
		}
		return 0
	}
	if pSes.nMaxDepth != -1 && pSes.pArrTimeProfile != nil && nGates > pSes.nMaxGates {
		if pSes.fReasonVerbose != 0 {
			spFmt.Printf("give up due to impossible depth and arrival times (depth = %d, gates = %d)", pSes.nMaxDepth, nGates)
		}
		return 0
	}
	if pSes.fDecStructure != 0 && nGates >= (int32(1)<<uint(pSes.nMaxDepth-1))+1 {
		if pSes.fReasonVerbose != 0 {
			spFmt.Printf("give up due to impossible depth in AND-dec structure (depth = %d, gates = %d)", pSes.nMaxDepth, nGates)
		}
		return 0
	}
	if nGates >= int32(1)<<uint(pSes.nSpecVars) {
		if pSes.fReasonVerbose != 0 {
			spFmt.Printf("give up due to impossible number of gates")
		}
		return 0
	}
	return 1
}

func Abc_ExactCopyDepthAndArrivalTimes(nVars int32, nDepthFrom int32, nDepthTo *int32, pTimesFrom *int32, pTimesTo *int32) {
	*nDepthTo = nDepthFrom
	if nVars <= 0 {
		return
	}
	copy(unsafe.Slice(pTimesTo, int(nVars)), unsafe.Slice(pTimesFrom, int(nVars)))
}

func Ses_ManStoreDepthAndArrivalTimes(pSes *Ses_Man_t) {
	if pSes.nMaxDepth == -1 {
		return
	}
	Abc_ExactCopyDepthAndArrivalTimes(pSes.nSpecVars, pSes.nMaxDepth, &pSes.nMaxDepthTmp, pSes.pArrTimeProfile, &pSes.pArrTimeProfileTmp[0])
}

func Ses_ManRestoreDepthAndArrivalTimes(pSes *Ses_Man_t) {
	if pSes.nMaxDepth == -1 {
		return
	}
	Abc_ExactCopyDepthAndArrivalTimes(pSes.nSpecVars, pSes.nMaxDepthTmp, &pSes.nMaxDepth, pSes.pArrTimeProfileTmp[:], pSes.pArrTimeProfile)
}

func Abc_ExactAdjustDepthAndArrivalTimes(nVars int32, nGates int32, nDepthFrom int32, nDepthTo *int32, pTimesFrom *int32, pTimesTo *int32, nTimesMax int32) {
	if nDepthFrom < nGates {
		*nDepthTo = nDepthFrom
	} else {
		*nDepthTo = nGates
	}
	if nVars <= 0 {
		return
	}
	from, to := unsafe.Slice(pTimesFrom, int(nVars)), unsafe.Slice(pTimesTo, int(nVars))
	for l := int32(0); l < nVars; l++ {
		limit := nGates - 1 - nTimesMax + from[l]
		if limit < 0 {
			limit = 0
		}
		if from[l] < limit {
			to[l] = from[l]
		} else {
			to[l] = limit
		}
	}
}

func Ses_AdjustDepthAndArrivalTimes(pSes *Ses_Man_t, nGates int32) {
	Abc_ExactAdjustDepthAndArrivalTimes(pSes.nSpecVars, nGates, pSes.nMaxDepthTmp, &pSes.nMaxDepth, pSes.pArrTimeProfileTmp, pSes.pArrTimeProfile, pSes.nArrTimeMax-1)
}

/* return: (2: continue, 1: found, 0: gave up) */
func Ses_ManFindNetworkExact(pSes *Ses_Man_t, nGates int32) int32 {
	var f, fSat int32
	var timeStart abctime

	/* solve */
	timeStart = Abc_Clock()
	Vec_IntClear(pSes.vPolar)
	Vec_IntClear(pSes.vAssump)
	Ses_ManCreateVars(pSes, nGates)

	if pSes.nMaxDepth != -1 {
		f = Ses_ManCreateDepthClauses(pSes)
		pSes.timeInstance += (Abc_Clock() - timeStart)
		if !(f != 0) {
			return 2
		} /* proven UNSAT while creating clauses */
	}

	for variable := int32(0); variable < pSes.pSat.NumVars(); variable++ {
		pSes.pSat.SetPolarity(variable, false)
	}
	for _, variable := range vec.Vec_IntArray(pSes.vPolar)[:vec.Vec_IntSize(pSes.vPolar)] {
		pSes.pSat.SetPolarity(variable, true)
	}

	/* first solve */
	fSat = Ses_ManSolve(pSes)
	if fSat == 0 {
		return 2 /* UNSAT, continue with next # of gates */
	} else if fSat == 2 {
		pSes.fHitResLimit = 1
		return 0
	}

	timeStart = Abc_Clock()
	f = Ses_ManCreateClauses(pSes)
	pSes.timeInstance += (Abc_Clock() - timeStart)
	if !(f != 0) {
		return 2
	} /* proven UNSAT while creating clauses */

	fSat = Ses_ManSolve(pSes)
	if fSat == 1 {
		return 1
	} else if fSat == 2 {
		pSes.fHitResLimit = 1
		return 0
	}

	return 2 /* UNSAT continue */
}

// is there a network for a given number of gates
/* return: (3: impossible, 2: continue, 1: found, 0: gave up) */
func sesSetTruthBit(truth *[4]word, bit int32) {
	if bit >= 0 && bit < 256 {
		truth[bit>>6] |= word(1) << uint(bit&63)
	}
}

func Ses_ManFindNetworkExactCEGAR(pSes *Ses_Man_t, nGates int32, pSol *string) int32 {
	extra := int32(4)
	if nGates > 10 {
		extra = 5
	}
	Abc_DebugErase(pSes.nDebugOffset+extra, pSes.fVeryVerbose)
	Abc_DebugPrintIntInt(" (%d/%d)", nGates, pSes.nMaxGates, pSes.fVeryVerbose)
	if Ses_CheckGatesConsistency(pSes, nGates) == 0 {
		return 3
	}
	for i := int32(0); i < pSes.nRandRowAssigns; i++ {
		sesSetTruthBit(&pSes.pTtValues, int32(pSes.rng.next()%uint32(pSes.nRows)))
	}
	result := Ses_ManFindNetworkExact(pSes, nGates)
	if result != 1 {
		return result
	}
	for {
		*pSol = Ses_ManExtractSolution(pSes)
		derived := Ses_ManDeriveTruth(pSes, *pSol, 0)
		truth := unsafe.Slice(derived, int(pSes.nSpecWords))
		spec := unsafe.Slice(pSes.pSpec, int(pSes.nSpecWords))
		mint := int32(-1)
		for bit := int32(0); bit < pSes.nSpecWords*64; bit++ {
			if ((truth[bit>>6]^spec[bit>>6])>>uint(bit&63))&1 != 0 {
				mint = bit
				break
			}
		}
		if mint == -1 || (pSes.nSpecVars < 6 && mint > pSes.nRows) {
			assert(result == 1, "SAT model satisfies specification")
			return 1
		}
		*pSol = ""
		if pSes.fKeepRowAssigns != 0 {
			sesSetTruthBit(&pSes.pTtValues, mint-1)
		}
		if Ses_ManCreateTruthTableClause(pSes, mint-1) == 0 {
			return 2
		}
		switch Ses_ManSolve(pSes) {
		case 1:
			continue
		case 2:
			return 0
		default:
			return 2
		}
	}
}

func Ses_ManFindMinimumSizeBottomUp(pSes *Ses_Man_t) string {
	nGates := pSes.nStartGates
	pSes.fHitResLimit = 0
	if pSes.nMaxGates >= 10 {
		pSes.nDebugOffset = 3
	} else {
		pSes.nDebugOffset = 2
	}
	if Vec_IntSize(pSes.vStairDecVars) > 0 {
		nGates = max32(nGates, int32(Vec_IntSize(pSes.vStairDecVars))-1)
	}
	pSes.pTtValues = [4]word{}
	Abc_DebugPrintIntInt(" (%d/%d)", nGates, pSes.nMaxGates, pSes.fVeryVerbose)
	var solution string
	for {
		nGates++
		result := Ses_ManFindNetworkExactCEGAR(pSes, nGates, &solution)
		if result == 0 {
			pSes.fHitResLimit = 1
			break
		}
		if result == 1 || result == 3 {
			break
		}
	}
	extra := int32(4)
	if nGates >= 10 {
		extra = 5
	}
	Abc_DebugErase(pSes.nDebugOffset+extra, pSes.fVeryVerbose)
	return solution
}

func Ses_ManFindMinimumSizeTopDown(pSes *Ses_Man_t, nMinGates int32) string {
	nGates := pSes.nMaxGates
	pSes.fHitResLimit = 0
	Abc_DebugPrintIntInt(" (%d/%d)", nGates, pSes.nMaxGates, pSes.fVeryVerbose)
	var solution, candidate string
	for {
		result := Ses_ManFindNetworkExactCEGAR(pSes, nGates, &candidate)
		if result == 0 {
			pSes.fHitResLimit = 1
			break
		}
		if result == 2 || result == 3 {
			break
		}
		solution = candidate
		if nGates == nMinGates {
			break
		}
		nGates--
	}
	extra := int32(4)
	if nGates >= 10 {
		extra = 5
	}
	Abc_DebugErase(pSes.nDebugOffset+extra, pSes.fVeryVerbose)
	return solution
}

func Ses_ManFindMinimumSize(pSes *Ses_Man_t) string {
	if pSes.nSpecFunc > 1 {
		for gates := pSes.nStartGates + 1; ; gates++ {
			if pSes.fVerbose != 0 {
				spFmt.Printf("try with %d gates\n", gates)
			}
			for i := range pSes.pTtValues {
				pSes.pTtValues[i] = ^word(0)
			}
			result := Ses_ManFindNetworkExact(pSes, gates)
			if result == 2 {
				continue
			}
			if result == 0 {
				return ""
			}
			return Ses_ManExtractSolution(pSes)
		}
	}
	if pSes.nMaxDepth != -1 && pSes.pArrTimeProfile != nil {
		if Ses_CheckDepthConsistency(pSes) == 0 {
			return ""
		}
		Ses_ManComputeMaxGates(pSes)
	}
	solution := Ses_ManFindMinimumSizeBottomUp(pSes)
	if solution == "" && pSes.nMaxDepth != -1 && pSes.fHitResLimit != 0 && pSes.nGates != pSes.nMaxGates {
		return Ses_ManFindMinimumSizeTopDown(pSes, pSes.nGates+1)
	}
	return solution
}

/*
*Function*************************************************************

	Synopsis    [Find minimum size networks with a SAT solver.]

	Description [If nMaxDepth is -1, then depth constraints are ignored.
	             If nMaxDepth is not -1, one can set pArrTimeProfile which should have the length of nVars.
	             One can ignore pArrTimeProfile by setting it to NULL.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFindExact(pTruth *word, nVars, nFunc, nMaxDepth int32, pArrTimeProfile *int32, nBTLimit, nStartGates, fVerbose int32) *Abc_Ntk_t {
	assert(nVars >= 2 && nVars <= 8, "nVars >= 2 && nVars <= 8")
	started := Abc_Clock()
	pSes := Ses_ManAlloc(pTruth, nVars, nFunc, nMaxDepth, pArrTimeProfile, 0, nBTLimit, fVerbose)
	pSes.nStartGates, pSes.fReasonVerbose, pSes.fSatVerbose = nStartGates, 0, 0
	if fVerbose != 0 {
		Ses_ManPrintFuncs(pSes)
	}
	var ntk *Abc_Ntk_t
	if sol := Ses_ManFindMinimumSize(pSes); sol != "" {
		ntk = Ses_ManExtractNtk(sol)
	}
	pSes.timeTotal = Abc_Clock() - started
	if fVerbose != 0 {
		Ses_ManPrintRuntime(pSes)
	}
	Ses_ManClean(pSes)
	return ntk
}

func Gia_ManFindExact(pTruth *word, nVars, nFunc, nMaxDepth int32, pArrTimeProfile *int32, nBTLimit, nStartGates, fVerbose int32) *Gia_Man_t {
	assert(nVars >= 2 && nVars <= 8, "nVars >= 2 && nVars <= 8")
	started := Abc_Clock()
	pSes := Ses_ManAlloc(pTruth, nVars, nFunc, nMaxDepth, pArrTimeProfile, 1, nBTLimit, fVerbose)
	pSes.nStartGates = nStartGates
	pSes.fVeryVerbose, pSes.fExtractVerbose, pSes.fSatVerbose, pSes.fReasonVerbose = 1, 0, 0, 1
	if fVerbose != 0 {
		Ses_ManPrintFuncs(pSes)
	}
	var graph *Gia_Man_t
	if sol := Ses_ManFindMinimumSize(pSes); sol != "" {
		graph = Ses_ManExtractGia(sol)
	}
	pSes.timeTotal = Abc_Clock() - started
	if fVerbose != 0 {
		Ses_ManPrintRuntime(pSes)
	}
	Ses_ManClean(pSes)
	return graph
}

func Abc_NtkFromTruthTable(pTruth *word, nVars int32) *Abc_Ntk_t {
	truth := unsafeWordSlice(pTruth, int(((uint32(1)<<nVars)+31)/32))
	arena := mem.Mem_FlexStart()
	cover := baseabc.Abc_SopCreateFromTruth(arena, nVars, truth)
	ntk := baseabc.Abc_NtkCreateWithNode(cover)
	baseabc.Abc_NtkShortNames(ntk)
	mem.Mem_FlexStop(arena, 0)
	return ntk
}

func Abc_ExactTestSingleOutput(fVerbose int32) {
	truth := [4]word{0xcafe, 0, 0, 0}
	profile := [4]int32{6, 2, 8, 5}
	ref := Abc_NtkFromTruthTable(&truth[0], 4)
	for _, tc := range []struct {
		depth    int32
		arrivals *int32
		limit    int32
		nodes    int32
	}{
		{-1, nil, 0, 6}, {3, nil, 0, -2}, {9, &profile[0], 50000, -2},
	} {
		got := Abc_NtkFindExact(&truth[0], 4, 1, tc.depth, tc.arrivals, tc.limit, 0, fVerbose)
		assert(got != nil, "exact network exists")
		baseabc.Abc_NtkShortNames(got)
		Abc_NtkCecSat(ref, got, 10000, 0)
		if tc.nodes >= 0 {
			assert(baseabc.Abc_NtkNodeNum(got) == tc.nodes, "expected six-gate minimum")
		} else {
			assert(baseabc.Abc_NtkLevel(got) <= tc.depth, "network meets depth bound")
		}
		Abc_NtkDelete(got)
	}
	assert(Abc_NtkFindExact(&truth[0], 4, 1, 2, nil, 50000, 0, fVerbose) == nil, "depth 2 infeasible")
	assert(Abc_NtkFindExact(&truth[0], 4, 1, 8, &profile[0], 50000, 0, fVerbose) == nil, "arrival bound infeasible")
	Abc_NtkDelete(ref)
}

func Abc_ExactTestSingleOutputAIG(fVerbose int32) {
	truth := [4]word{0xcafe, 0, 0, 0}
	profile := [4]int32{6, 2, 8, 5}
	params := cec.Cec_ParCec_t{}
	cec.Cec_ManCecSetDefaultParams(&params)
	ntk := Abc_NtkFromTruthTable(&truth[0], 4)
	baseabc.Abc_NtkToAig(ntk)
	ref := baseabc.Abc_NtkAigToGia(ntk, 1)
	for _, tc := range []struct {
		depth    int32
		arrivals *int32
		limit    int32
		require  bool
	}{
		{-1, nil, 0, true}, {3, nil, 0, true}, {9, &profile[0], 50000, true},
		{2, nil, 50000, false}, {8, &profile[0], 50000, false},
	} {
		got := Gia_ManFindExact(&truth[0], 4, 1, tc.depth, tc.arrivals, tc.limit, 0, fVerbose)
		if !tc.require {
			assert(got == nil, "expected infeasible exact synthesis")
			continue
		}
		assert(got != nil, "exact Gia exists")
		miter := gia.Gia_ManMiter(ref, got, 0, 1, 0, 0, 1)
		assert(miter != nil, "equivalence miter exists")
		cec.Cec_ManVerify(miter, &params)
		gia.Gia_ManStop(miter)
		gia.Gia_ManStop(got)
	}
	gia.Gia_ManStop(ref)
	baseabc.Abc_NtkDelete(ntk)
}

func Abc_ExactTest(fVerbose int32) {
	Abc_ExactTestSingleOutput(fVerbose)
	Abc_ExactTestSingleOutputAIG(fVerbose)
	spFmt.Println()
}

func Abc_ExactIsRunning() int32 {
	if s_pSesStore != nil {
		return 1
	}
	return 0
}
func Abc_ExactInputNum() int32 { return 8 }

func Abc_ExactStart(nBTLimit, fMakeAIG, fVerbose, fVeryVerbose int32, filename string) {
	if s_pSesStore != nil {
		spFmt.Println("BMS manager already started")
		return
	}
	store := Ses_StoreAlloc(nBTLimit, fMakeAIG, fVerbose)
	store.fVeryVerbose = fVeryVerbose
	if filename != "" {
		Ses_StoreRead(store, filename, 1, 0, 0, 0)
		store.szDBName = filename
	}
	if store.fVeryVerbose != 0 {
		f, err := os.OpenFile("bms.debug", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err == nil {
			store.pDebugEntries = f
		}
	}
	s_pSesStore = store
}

func Abc_ExactStop(filename string) {
	if s_pSesStore == nil {
		spFmt.Println("BMS manager has not been started")
		return
	}
	store := s_pSesStore
	if filename != "" {
		Ses_StoreWrite(store, filename, 1, 0, 0, 0)
	}
	if store.pDebugEntries != nil {
		_ = store.pDebugEntries.Close()
		store.pDebugEntries = nil
	}
	Ses_StoreClean(store)
	s_pSesStore = nil
}

func Abc_ExactStats() {
	st := s_pSesStore
	if st == nil {
		spFmt.Println("BMS manager has not been started")
		return
	}
	spFmt.Println("-------------------------------------------------------------------------------------------------------------------------------")
	spFmt.Println("                                    0         1         2         3         4         5         6         7         8     total")
	spFmt.Println("-------------------------------------------------------------------------------------------------------------------------------")
	rows := []struct {
		name  string
		per   *[9]uint64
		total uint64
	}{
		{"number of considered cuts :", &st.pCutCount, st.nCutCount}, {" - trivial                :", &st.pSynthesizedTrivial, st.nSynthesizedTrivial},
		{" - synth (imp)            :", &st.pSynthesizedImp, st.nSynthesizedImp}, {" - synth (res)            :", &st.pSynthesizedRL, st.nSynthesizedRL},
		{" - not synth (imp)        :", &st.pUnsynthesizedImp, st.nUnsynthesizedImp}, {" - not synth (res)        :", &st.pUnsynthesizedRL, st.nUnsynthesizedRL},
		{" - cache hits             :", &st.pCacheHits, st.nCacheHits},
	}
	for _, row := range rows {
		spFmt.Print(row.name)
		for _, n := range row.per {
			spFmt.Printf("%10d", n)
		}
		spFmt.Printf("%10d\n", row.total)
	}
	spFmt.Println("-------------------------------------------------------------------------------------------------------------------------------")
	spFmt.Printf("number of entries         : %d\nnumber of valid entries   : %d\nnumber of invalid entries : %d\n", st.nEntriesCount, st.nValidEntriesCount, st.nEntriesCount-st.nValidEntriesCount)
	spFmt.Println("-------------------------------------------------------------------------------------------------------------------------------")
	spFmt.Printf("number of SAT calls       : %d\nnumber of UNSAT calls     : %d\nnumber of UNDEF calls     : %d\n", st.nSatCalls, st.nUnsatCalls, st.nUndefCalls)
	spFmt.Println("-------------------------------------------------------------------------------------------------------------------------------\nRuntime breakdown:")
	printExactRuntime := func(name string, d abctime) {
		pct := 0.0
		if st.timeTotal > 0 {
			pct = 100 * float64(d) / float64(st.timeTotal)
		}
		spFmt.Printf("%s = %9.2f sec (%6.2f %%)\n", name, float64(d)/float64(time.Second), pct)
	}
	printExactRuntime("Exact    ", st.timeExact)
	printExactRuntime(" Sat     ", st.timeSat)
	printExactRuntime("  Sat    ", st.timeSatSat)
	printExactRuntime("  Unsat  ", st.timeSatUnsat)
	printExactRuntime("  Undef  ", st.timeSatUndef)
	printExactRuntime(" Instance", st.timeInstance)
	printExactRuntime("Other    ", st.timeTotal-st.timeExact)
	printExactRuntime("ALL      ", st.timeTotal)
}

// Abc_ExactDelayCost returns the best output arrival time and the source-order
// pin delays for the optimal cached/synthesized network.
func Abc_ExactDelayCost(pTruth *word, nVars int32, pArrTimeProfile *int32, pPerm []int8, cost *int32, aigLevel int32) int32 {
	st := s_pSesStore
	assert(st != nil, "exact manager is running")
	if nVars < 0 || nVars > 8 {
		spFmt.Printf("invalid truth table size %d\n", nVars)
		assert(false, "valid truth table size")
		return ABC_INFINITY
	}
	st.nCutCount++
	st.pCutCount[nVars]++
	started := Abc_Clock()
	if nVars == 0 {
		st.nSynthesizedTrivial++
		st.pSynthesizedTrivial[0]++
		*cost = 0
		st.timeTotal += Abc_Clock() - started
		return 0
	}
	if nVars == 1 {
		st.nSynthesizedTrivial++
		st.pSynthesizedTrivial[1]++
		*cost = 0
		pPerm[0] = 0
		st.timeTotal += Abc_Clock() - started
		return *pArrTimeProfile
	}
	var normalized [8]int32
	for i := int32(0); i < nVars; i++ {
		normalized[i] = unsafeInt32At(pArrTimeProfile, i)
	}
	var maxArrival int32
	delta := Abc_NormalizeArrivalTimes(&normalized[0], nVars, &maxArrival)
	*cost = ABC_INFINITY
	var sol string
	if Ses_StoreGetEntry(st, pTruth, nVars, &normalized[0], &sol) != 0 {
		st.nCacheHits++
		st.pCacheHits[nVars]++
	} else {
		maxDepth := normalized[0]
		for i := int32(1); i < nVars; i++ {
			if normalized[i] > maxDepth {
				maxDepth = normalized[i]
			}
		}
		maxDepth += nVars + 1
		if aigLevel != -1 {
			bound := aigLevel - delta
			cap := maxDepth + nVars + 1
			if bound < cap {
				maxDepth = bound
			} else {
				maxDepth = cap
			}
		}
		exactStart := Abc_Clock()
		man := Ses_ManAlloc(pTruth, nVars, 1, maxDepth, &normalized[0], st.fMakeAIG, st.nBTLimit, st.fVerbose)
		man.fVeryVerbose = st.fVeryVerbose
		man.pSat = st.pSat
		man.nStartGates = nVars - 2
		for man.nMaxDepth > 0 {
			if st.fVeryVerbose != 0 {
				spFmt.Printf(" %d", man.nMaxDepth)
			}
			candidate := Ses_ManFindMinimumSize(man)
			if candidate == "" {
				if st.fVeryVerbose != 0 {
					spFmt.Printf("\b%s%d%s", ANSI_COLOR_YELLOW, man.nMaxDepth, ANSI_COLOR_RESET)
				}
				break
			}
			sol = candidate
			if st.fVeryVerbose != 0 {
				spFmt.Printf("\b%s%d%s", ANSI_COLOR_GREEN, man.nMaxDepth, ANSI_COLOR_RESET)
			}
			man.nMaxDepth--
		}
		if st.fVeryVerbose != 0 {
			spFmt.Println("        ")
		}
		if st.pDebugEntries != nil && man.fHitResLimit != 0 {
			Ses_StorePrintDebugEntry(st, pTruth, nVars, &normalized[0], man.nMaxDepth, sol, nVars-2)
		}
		man.timeTotal = Abc_Clock() - exactStart
		st.nSatCalls += uint64(man.nSatCalls)
		st.nUnsatCalls += uint64(man.nUnsatCalls)
		st.nUndefCalls += uint64(man.nUndefCalls)
		st.timeSat += man.timeSat
		st.timeSatSat += man.timeSatSat
		st.timeSatUnsat += man.timeSatUnsat
		st.timeSatUndef += man.timeSatUndef
		st.timeInstance += man.timeInstance
		st.timeExact += man.timeTotal
		resLimit := man.fHitResLimit
		Ses_ManCleanLight(man)
		Ses_StoreAddEntry(st, pTruth, nVars, &normalized[0], sol, resLimit)
	}
	if sol != "" {
		gateCount := int(sol[ABC_EXACT_SOL_NGATES])
		*cost = gateCount
		pos := 3 + 4*gateCount + 1
		delay := int32(sol[pos])
		_ = delay
		pos++
		best := int32(0)
		for i := int32(0); i < nVars; i++ {
			if pos >= len(sol) {
				break
			}
			pPerm[i] = int8(sol[pos])
			pos++
			arrival := unsafeInt32At(pArrTimeProfile, i) + int32(pPerm[i])
			if arrival > best {
				best = arrival
			}
		}
		st.timeTotal += Abc_Clock() - started
		return best
	}
	st.timeTotal += Abc_Clock() - started
	return ABC_INFINITY
}

func Abc_ExactBuildNode(pTruth *word, nVars int32, pArrTimeProfile *int32, fanins []*Abc_Obj_t, ntk *Abc_Ntk_t) *Abc_Obj_t {
	started := Abc_Clock()
	if nVars == 0 {
		s_pSesStore.timeTotal += Abc_Clock() - started
		if *pTruth&1 != 0 {
			return baseabc.Abc_NtkCreateNodeConst1(ntk)
		}
		return baseabc.Abc_NtkCreateNodeConst0(ntk)
	}
	if nVars == 1 {
		s_pSesStore.timeTotal += Abc_Clock() - started
		if *pTruth&1 != 0 {
			return baseabc.Abc_NtkCreateNodeInv(ntk, fanins[0])
		}
		return baseabc.Abc_NtkCreateNodeBuf(ntk, fanins[0])
	}
	var normalized [8]int32
	for i := int32(0); i < nVars; i++ {
		normalized[i] = unsafeInt32At(pArrTimeProfile, i)
	}
	var maxArrival int32
	Abc_NormalizeArrivalTimes(&normalized[0], nVars, &maxArrival)
	var sol string
	if Ses_StoreGetEntry(s_pSesStore, pTruth, nVars, &normalized[0], &sol) == 0 || sol == "" {
		s_pSesStore.timeTotal += Abc_Clock() - started
		return nil
	}
	assert(int32(sol[ABC_EXACT_SOL_NVARS]) == nVars, "solution variable count")
	assert(sol[ABC_EXACT_SOL_NFUNC] == 1, "single output solution")
	gates := make([]*Abc_Obj_t, 0, int(nVars)+int(sol[ABC_EXACT_SOL_NGATES]))
	gates = append(gates, fanins[:nVars]...)
	pos := 3
	gateCount := int(sol[ABC_EXACT_SOL_NGATES])
	for i := 0; i < gateCount; i++ {
		bits := sol[pos]
		pos++
		assert(sol[pos] == 2, "binary solution gate")
		pos++
		truthBin := []byte{'0' + byte((bits>>2)&1), '0' + byte((bits>>1)&1), '0' + byte(bits&1), '0', '\x00'}
		if i+1 == gateCount && (sol[pos+2]&1) != 0 {
			for j := 0; j < 4; j++ {
				if truthBin[j] == '0' {
					truthBin[j] = '1'
				} else {
					truthBin[j] = '0'
				}
			}
		}
		sop := baseabc.Abc_SopFromTruthBin(string(truthBin))
		obj := baseabc.Abc_NtkCreateNode(ntk)
		obj.Data = baseabc.Abc_SopRegister(ntk.FunctionManager.(*mem.Mem_Flex_t), sop)
		gates = append(gates, obj)
		left := int(sol[pos])
		pos++
		right := int(sol[pos])
		pos++
		baseabc.Abc_ObjAddFanin(obj, gates[left])
		baseabc.Abc_ObjAddFanin(obj, gates[right])
	}
	outputLit := int(sol[pos])
	result := gates[outputLit>>1]
	s_pSesStore.timeTotal += Abc_Clock() - started
	return result
}

func Abc_ExactStoreTest(fVerbose int32) {
	truth := [4]word{0xcafe, 0, 0, 0}
	arrivals := [4]int32{6, 2, 8, 5}
	ntk := baseabc.Abc_NtkAlloc(baseabc.ABC_NTK_LOGIC, baseabc.ABC_FUNC_SOP, 1)
	ntk.Name = "exact"
	ntk.Spec = "exact"
	fanins := [4]*Abc_Obj_t{}
	names := baseabc.Abc_NodeGetFakeNames(4)
	for i := int32(0); i < 4; i++ {
		fanins[i] = baseabc.Abc_NtkCreatePi(ntk)
		name, _ := vec.Vec_PtrEntry(names, int(i)).(string)
		baseabc.Abc_ObjAssignName(fanins[i], name, "")
	}
	baseabc.Abc_NodeFreeNames(names)
	Abc_ExactStart(10000, 1, fVerbose, 0, "")
	defer func() { Abc_ExactStop(""); baseabc.Abc_NtkDelete(ntk) }()
	assert(Abc_ExactBuildNode(&truth[0], 4, &arrivals[0], fanins[:], ntk) == nil, "no pre-cache implementation")
	perm := make([]int8, 4)
	var cost int32
	assert(Abc_ExactDelayCost(&truth[0], 4, &arrivals[0], perm, &cost, 12) == 1, "delay cost is one")
	assert(Abc_ExactBuildNode(&truth[0], 4, &arrivals[0], fanins[:], ntk) != nil, "cached implementation builds")
	arrivals[0]++
	assert(Abc_ExactBuildNode(&truth[0], 4, &arrivals[0], fanins[:], ntk) == nil, "arrival profile cache key differs")
}

// Helpers for pointer-shaped C arrays at this API boundary.
func unsafeWordSlice(p *word, n int) []uint32 {
	out := make([]uint32, n)
	for i := range out {
		out[i] = uint32(unsafeWordAt(p, i/2) >> (32 * uint(i%2)))
	}
	return out
}
func unsafeWordAt(p *word, i int) word {
	return *(*word)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + uintptr(i)*unsafe.Sizeof(*p)))
}
func unsafeInt32At(p *int32, i int32) int32 {
	return *(*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + uintptr(i)*unsafe.Sizeof(*p)))
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
