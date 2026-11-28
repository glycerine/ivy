// Package super precomputes supergate libraries.
//
// superGate.go is a line-by-line port of src/map/super/superGate.c
// (MVSIS Group, UC Berkeley): the same control flow, stmm hash table
// (st__ptrhash on the XOR of the truth-table words, iterated bin by bin),
// glibc qsort (a stable merge sort, hence sort.SliceStable), and the
// Extra_MmFixed recycling of dominated supergates (LIFO free list; a
// recycled entry is reused in place, so arrays and fanin pointers that still
// reference it see the new supergate, as in C).
package super

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"

	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Mio_Library_t = mio.Mio_Library_t
type Mio_Gate_t = mio.Mio_Gate_t
type Vec_Str_t = vec.Vec_Str_t

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// the bit masks
func SUPER_MASK(n int32) uint32 { return (^uint32(0)) >> uint(32-n) }

const SUPER_FULL = ^uint32(0)

// SUPER_NO_VAR and SUPER_EPSILON are double constants in C; comparisons
// with float fields promote the float to double.
const SUPER_NO_VAR float64 = -9999.0
const SUPER_EPSILON float64 = 0.001

// data structure for supergate precomputation
type Super_Man_t struct {
	// parameters
	pName           string  // the original genlib file name
	nVarsMax        int32   // the number of inputs
	nMints          int32   // the number of minterms
	nLevels         int32   // the number of logic levels
	nGatesMax       int32   // the number of gates computed
	tDelayMax       float32 // the max delay of the supergates in the library
	tAreaMax        float32 // the max area of the supergates in the library
	fSkipInv        int32   // the flag says about skipping inverters
	fWriteOldFormat int32   // in addition, writes the file in the old format
	fVerbose        int32

	// supergates
	pInputs [10]*Super_Gate_t // the input supergates
	nGates  int32             // the number of gates in the library
	pGates  []*Super_Gate_t   // the gates themselves
	tTable  *stmm_table       // mapping of truth tables into gates

	// memory manager for the supergates (Extra_MmFixed_t free list)
	pMemFree []*Super_Gate_t

	// statistics
	nTried   int32 // the total number of tried
	nAdded   int32 // the number of entries added
	nRemoved int32 // the number of entries removed
	nClasses int32 // the number of gate classes
	nUnique  int32 // the number of unique gates
	nLookups int32 // the number of hash table lookups
	nAliases int32 // the number of hash table lookups thrown away due to aliasing

	// runtime
	Time      int64 // the runtime of the generation procedure
	TimeLimit int32 // the runtime limit (in seconds)
	TimeSec   int32 // the time passed (in seconds)
	TimeStop  int64 // the time to stop computation (in miliseconds)
	TimePrint int64 // the time to print message
}

type Super_Gate_t struct {
	pRoot     *Mio_Gate_t      // the root gate for this supergate
	fVar      uint32           // (1 bit) the flag signaling the elementary variable
	fSuper    uint32           // (1 bit) the flag signaling the elementary variable
	nFanins   uint32           // (6 bits) the number of fanin gates
	Number    uint32           // (24 bits) the number assigned in the process
	uTruth    [2]uint32        // the truth table of this supergate
	pFanins   [6]*Super_Gate_t // the fanins of the gate
	Area      float32          // the area of this gate
	ptDelays  [6]float32       // the pin-to-pin delays for all inputs
	tDelayMax float32          // the maximum delay
	pNext     *Super_Gate_t    // the next gate in the table

	// fRecycled marks an entry on the memory manager's free list: C stores
	// the free-list link over pRoot, so reading it would be garbage.
	fRecycled bool
}

// Extra_MmFixedEntryFetch: recycled entries first (LIFO), then fresh ones.
func (pMan *Super_Man_t) memFetch() *Super_Gate_t {
	if n := len(pMan.pMemFree); n > 0 {
		pGate := pMan.pMemFree[n-1]
		pMan.pMemFree[n-1] = nil
		pMan.pMemFree = pMan.pMemFree[:n-1]
		return pGate
	}
	return &Super_Gate_t{}
}

// Extra_MmFixedEntryRecycle
func (pMan *Super_Man_t) memRecycle(pGate *Super_Gate_t) {
	pGate.pRoot = nil
	pGate.fRecycled = true
	pMan.pMemFree = append(pMan.pMemFree, pGate)
}

// stmm_table as initialized by stmm_init_table( st__ptrcmp, st__ptrhash ):
// 11 bins, max density 5, grow factor 2.0, no reordering; keys are the
// (zero-extended) unsigned XOR of the truth-table words.
type stmm_table struct {
	num_entries int32
	num_bins    int32
	bins        []*stmm_table_entry
}
type stmm_table_entry struct {
	key    uint64
	record *Super_Gate_t
	next   *stmm_table_entry
}

const stmm_max_density = 5
const stmm_grow_factor = 2.0

func stmm_init_table() *stmm_table {
	return &stmm_table{num_bins: 11, bins: make([]*stmm_table_entry, 11)}
}
func (table *stmm_table) do_hash(key uint64) int32 {
	return int32((key >> 2) % uint64(table.num_bins))
}
func (table *stmm_table) rehash() {
	old_bins := table.bins
	old_num_bins := table.num_bins
	table.num_bins = int32(stmm_grow_factor * float64(old_num_bins))
	if table.num_bins%2 == 0 {
		table.num_bins += 1
	}
	table.num_entries = 0
	table.bins = make([]*stmm_table_entry, table.num_bins)
	for i := int32(0); i < old_num_bins; i++ {
		ptr := old_bins[i]
		for ptr != nil {
			next := ptr.next
			hash_val := table.do_hash(ptr.key)
			ptr.next = table.bins[hash_val]
			table.bins[hash_val] = ptr
			table.num_entries++
			ptr = next
		}
	}
}
func (table *stmm_table) find_entry(key uint64) *stmm_table_entry {
	ptr := table.bins[table.do_hash(key)]
	for ptr != nil && ptr.key != key {
		ptr = ptr.next
	}
	return ptr
}

// stmm_find_or_add returns the record slot and whether the key existed.
func stmm_find_or_add(table *stmm_table, key uint64, slot ***Super_Gate_t) int32 {
	if ptr := table.find_entry(key); ptr != nil {
		*slot = &ptr.record
		return 1
	}
	if table.num_entries/table.num_bins >= stmm_max_density {
		table.rehash()
	}
	hash_val := table.do_hash(key)
	newEntry := &stmm_table_entry{key: key, next: table.bins[hash_val]}
	table.bins[hash_val] = newEntry
	table.num_entries++
	*slot = &newEntry.record
	return 0
}
func stmm_find(table *stmm_table, key uint64, slot ***Super_Gate_t) int32 {
	if ptr := table.find_entry(key); ptr != nil {
		*slot = &ptr.record
		return 1
	}
	return 0
}

// stmm_foreach_item: bins in order, each chain from its head.
func stmm_foreach_item(table *stmm_table, f func(key uint64, record *Super_Gate_t)) {
	for i := int32(0); i < table.num_bins; i++ {
		for ptr := table.bins[i]; ptr != nil; ptr = ptr.next {
			f(ptr.key, ptr.record)
		}
	}
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Precomputes the library of supergates.]

**********************************************************************
*/
func Super_Precompute(pLibGen *Mio_Library_t, nVarsMax, nLevels, nGatesMax int32, tDelayMax, tAreaMax float32, TimeLimit, fSkipInv, fVerbose int32, pFileName string) {
	pFile, err := os.Create(pFileName)
	if err != nil {
		fmt.Printf("Cannot open output file \"%s\".\n", pFileName)
		return
	}
	vStr := Super_PrecomputeStr(pLibGen, nVarsMax, nLevels, nGatesMax, tDelayMax, tAreaMax, TimeLimit, fSkipInv, fVerbose)
	if vStr != nil {
		pFile.Write(vec.Vec_StrArray(vStr)[:vec.Vec_StrSize(vStr)])
	}
	pFile.Close()
	vec.Vec_StrFree(vStr)
	// report the result of writing
	if fVerbose != 0 {
		fmt.Printf("The supergates are written using new format \"%s\" ", pFileName)
		fmt.Printf("(%0.3f MB).\n", float64(extra.Extra_FileSize(pFileName))/float64(1<<20))
	}
}

/*
*Function*************************************************************

	Synopsis    [Precomputes the library of supergates.]

**********************************************************************
*/
func Super_PrecomputeStr(pLibGen *Mio_Library_t, nVarsMax, nLevels, nGatesMax int32, tDelayMax, tAreaMax float32, TimeLimit, fSkipInv, fVerbose int32) *Vec_Str_t {
	var vStr *Vec_Str_t
	var pMan *Super_Man_t
	var ppGates []*Mio_Gate_t
	var nGates, Level int32
	var clk, clockStart int64

	assert(nVarsMax < 7, "nVarsMax < 7")
	if nGatesMax != 0 && nGatesMax < nVarsMax {
		fmt.Fprintf(os.Stderr, "Erro! The number of supergates requested (%d) in less than the number of variables (%d).\n", nGatesMax, nVarsMax)
		fmt.Fprintf(os.Stderr, "The library cannot be computed.\n")
		return nil
	}

	// get the root gates
	ppGates = mio.Mio_CollectRoots(pLibGen, nVarsMax, tDelayMax, 0, &nGates, fVerbose)
	if nGatesMax != 0 && nGates >= nGatesMax {
		fmt.Fprintf(os.Stdout, "Warning! Genlib library contains more gates than can be computed.\n")
		fmt.Fprintf(os.Stdout, "Only one-gate supergates are included in the supergate library.\n")
	}

	// start the manager
	pMan = Super_ManStart()
	pMan.pName = mio.Mio_LibraryReadName(pLibGen)
	pMan.nGatesMax = nGatesMax
	pMan.fSkipInv = fSkipInv
	pMan.tDelayMax = tDelayMax
	pMan.tAreaMax = tAreaMax
	pMan.TimeLimit = TimeLimit // in seconds
	if TimeLimit != 0 {
		pMan.TimeStop = int64(TimeLimit)*util.CLOCKS_PER_SEC + util.Abc_Clock() // in CPU ticks
	}
	pMan.fVerbose = fVerbose

	if nGates == 0 {
		fmt.Fprintf(os.Stderr, "Error: No genlib gates satisfy the limits criteria. Stop.\n")
		fmt.Fprintf(os.Stderr, "Limits: max delay =  %.2f, max area =  %.2f, time limit = %d sec.\n",
			pMan.tDelayMax, pMan.tAreaMax, pMan.TimeLimit)
		// stop the manager
		Super_ManStop(pMan)
		return nil
	}

	// get the starting supergates
	Super_First(pMan, nVarsMax)

	// perform the computation of supergates
	clockStart = util.Abc_Clock()
	if fVerbose != 0 {
		fmt.Printf("Computing supergates with %d inputs, %d levels, and %d max gates.\n",
			pMan.nVarsMax, nLevels, nGatesMax)
		fmt.Printf("Limits: max delay =  %.2f, max area =  %.2f, time limit = %d sec.\n",
			pMan.tDelayMax, pMan.tAreaMax, pMan.TimeLimit)
	}

	for Level = 1; Level <= nLevels; Level++ {
		if pMan.TimeStop != 0 && util.Abc_Clock() > pMan.TimeStop {
			break
		}
		clk = util.Abc_Clock()
		Super_Compute(pMan, ppGates, nGates, nGatesMax, fSkipInv)
		pMan.nLevels = Level
		if fVerbose != 0 {
			fmt.Printf("Lev %d: Try =%12d. Add =%6d. Rem =%5d. Save =%6d. Lookups =%12d. Aliases =%12d. ",
				Level, pMan.nTried, pMan.nAdded, pMan.nRemoved, pMan.nAdded-pMan.nRemoved, pMan.nLookups, pMan.nAliases)
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
		}
	}
	pMan.Time = util.Abc_Clock() - clockStart

	if fVerbose != 0 {
		fmt.Printf("Writing the output file...\n")
	}
	// write them into a file
	vStr = Super_Write(pMan)

	// stop the manager
	Super_ManStop(pMan)
	return vStr
}

/*
*Function*************************************************************

	Synopsis    [Derives the starting supergates.]

**********************************************************************
*/
func Super_First(pMan *Super_Man_t, nVarsMax int32) {
	var pSuper *Super_Gate_t
	var nMintLimit, nVarLimit int32
	var v, m int32
	// set the parameters
	pMan.nVarsMax = nVarsMax
	pMan.nMints = (1 << nVarsMax)
	pMan.nLevels = 0
	// allocate room for the gates
	pMan.nGates = nVarsMax
	pMan.pGates = make([]*Super_Gate_t, nVarsMax+2)
	// create the gates corresponding to the elementary variables
	for v = 0; v < nVarsMax; v++ {
		// get a new gate
		pSuper = pMan.memFetch()
		*pSuper = Super_Gate_t{}
		// assign the elementary variable, the truth table, and the delays
		pSuper.fVar = 1
		pSuper.Number = uint32(v)
		for m = 0; m < nVarsMax; m++ {
			pSuper.ptDelays[m] = float32(SUPER_NO_VAR)
		}
		pSuper.ptDelays[v] = 0.0
		// set the gate
		pMan.pGates[v] = pSuper
		Super_AddGateToTable(pMan, pSuper)
		pMan.pInputs[v] = pSuper
	}
	// set up their truth tables
	if nVarsMax >= 5 {
		nVarLimit = 5
	} else {
		nVarLimit = nVarsMax
	}
	nMintLimit = (1 << nVarLimit)
	for m = 0; m < nMintLimit; m++ {
		for v = 0; v < nVarLimit; v++ {
			if m&(1<<v) != 0 {
				pMan.pGates[v].uTruth[0] |= (1 << m)
			}
		}
	}
	// make adjustments for the case of 6 variables
	if nVarsMax == 6 {
		for v = 0; v < 5; v++ {
			pMan.pGates[v].uTruth[1] = pMan.pGates[v].uTruth[0]
		}
		pMan.pGates[5].uTruth[0] = 0
		pMan.pGates[5].uTruth[1] = ^uint32(0)
	} else {
		for v = 0; v < nVarsMax; v++ {
			pMan.pGates[v].uTruth[1] = 0
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Precomputes one level of supergates.]

	Description [This procedure computes the set of supergates that can be
	derived from the given set of root gates (from genlib library) by composing
	the root gates with the currently available supergates. This procedure is
	smart in the sense that it tries to avoid useless emuration by imposing
	tight bounds by area and delay. Only the supergates and are guaranteed to
	have smaller area and delay are enumereated. See comments below for details.]

**********************************************************************
*/
func Super_Compute(pMan *Super_Man_t, ppGates []*Mio_Gate_t, nGates, nGatesMax, fSkipInv int32) *Super_Man_t {
	var pSupers [6]*Super_Gate_t
	var pGate0, pGate1, pGate2, pGate3, pGate4, pGate5, pGateNew *Super_Gate_t
	var tPinDelaysRes [6]float32
	var ptPinDelays [6][]float32
	var tPinDelayMax, tDelayMio float32
	var Area float32 = 0.0 // Suppress "might be used uninitialized"
	var Area0, Area1, Area2, Area3, Area4, AreaMio float32
	var uTruth [2]uint32
	var uTruths [6][2]uint32
	var i0, i1, i2, i3, i4, i5 int32
	var ppGatesLimit []*Super_Gate_t
	var nFanins, nGatesLimit, k, s, t int32
	var pProgress *extra.ProgressBar
	var fTimeOut int32
	fPrune := true   // Shall we prune?
	iPruneLimit := 3 // Each of the gates plugged into the root gate will have
	// less than these many fanins
	iPruneLimitRoot := 4 // The root gate may have only less than these many fanins

	// derive the delays and the truth table of the root gate over pSupers,
	// check the table, and add the new gate; returns 1 to "goto done"
	tryGate := func() int32 {
		mio.Mio_DeriveGateDelays(ppGates[k], ptPinDelays[:], nFanins, pMan.nVarsMax, float32(SUPER_NO_VAR), tPinDelaysRes[:], &tPinDelayMax)
		mio.Mio_DeriveTruthTable(ppGates[k], uTruths[:], nFanins, pMan.nVarsMax, uTruth[:])
		if Super_CompareGates(pMan, uTruth[:], Area, tPinDelaysRes[:], pMan.nVarsMax) == 0 {
			return 0
		}
		// create a new gate
		pGateNew = Super_CreateGateNew(pMan, ppGates[k], pSupers[:], nFanins, uTruth[:], Area, tPinDelaysRes[:], tPinDelayMax, pMan.nVarsMax)
		Super_AddGateToTable(pMan, pGateNew)
		if nGatesMax != 0 && pMan.nClasses > nGatesMax {
			return 1
		}
		return 0
	}
	areaOver := func(a float32) bool { return pMan.tAreaMax > 0.0 && a > pMan.tAreaMax }

	// put the gates from the unique table into the array
	// the gates from the array will be used to compose other gates
	// the gates in tbe table are used to check uniqueness of collected gates
	Super_TranferGatesToArray(pMan)

	// sort the gates in the increasing order of maximum delay
	if pMan.nGates > 10000 {
		fmt.Printf("Sorting array of %d supergates...\r", pMan.nGates)
	}
	sort.SliceStable(pMan.pGates[:pMan.nGates], func(a, b int) bool { return Super_DelayCompare(pMan.pGates[a], pMan.pGates[b]) < 0 })
	assert(pMan.nGates == 0 || Super_DelayCompare(pMan.pGates[0], pMan.pGates[pMan.nGates-1]) <= 0, "sorted by delay")
	if pMan.nGates > 10000 {
		fmt.Printf("                                       \r")
	}

	pProgress = extra.Extra_ProgressBarStart(os.Stdout, pMan.TimeLimit)
	pMan.TimePrint = util.Abc_Clock() + util.CLOCKS_PER_SEC
	ppGatesLimit = make([]*Super_Gate_t, pMan.nGates)
	// go through the root gates
	// the root gates are sorted in the increasing gelay
	fTimeOut = 0
	for k = 0; k < nGates; k++ {
		if fTimeOut != 0 {
			break
		}

		if fPrune {
			if pMan.nLevels >= 1 { // First level gates have been computed
				if int(mio.Mio_GateReadPinNum(ppGates[k])) >= iPruneLimitRoot {
					continue
				}
			}
		}

		// select the subset of gates to be considered with this root gate
		// all the gates past this point will lead to delay larger than the limit
		tDelayMio = float32(mio.Mio_GateReadDelayMax(ppGates[k]))
		for s, t = 0, 0; s < pMan.nGates; s++ {
			if fPrune && (pMan.nLevels >= 1) && (int(pMan.pGates[s].nFanins) >= iPruneLimit) {
				continue
			}

			ppGatesLimit[t] = pMan.pGates[s]
			t++
			if ppGatesLimit[t-1].tDelayMax+tDelayMio > pMan.tDelayMax && pMan.tDelayMax > 0.0 {
				break
			}
		}
		nGatesLimit = t

		if pMan.fVerbose != 0 {
			fmt.Printf("Trying %d choices for %d inputs\r", t, mio.Mio_GateReadPinNum(ppGates[k]))
		}

		// resort part of this range by area
		// now we can prune the search by going up in the list until we reach the limit on area
		// all the gates beyond this point can be skipped because their area can be only larger
		if nGatesLimit > 10000 {
			fmt.Printf("Sorting array of %d supergates...\r", nGatesLimit)
		}
		lim := ppGatesLimit[:nGatesLimit]
		sort.SliceStable(lim, func(a, b int) bool { return Super_AreaCompare(lim[a], lim[b]) < 0 })
		assert(nGatesLimit == 0 || Super_AreaCompare(lim[0], lim[nGatesLimit-1]) <= 0, "sorted by area")
		if nGatesLimit > 10000 {
			fmt.Printf("                                       \r")
		}

		// consider the combinations of gates with the root gate on top
		AreaMio = float32(mio.Mio_GateReadArea(ppGates[k]))
		nFanins = mio.Mio_GateReadPinNum(ppGates[k])
		switch nFanins {
		case 0: // should not happen
			assert(false, "root gate with no fanins")
		case 1: // interter root
			for i0 = 0; i0 < nGatesLimit; i0++ {
				if pGate0 = ppGatesLimit[i0]; pGate0 == nil {
					break
				}
				if fTimeOut != 0 {
					break
				}
				fTimeOut = Super_CheckTimeout(pProgress, pMan)
				// skip the inverter as the root gate before the elementary variable
				// as a result, the supergates will not have inverters on the input side
				// but inverters still may occur at the output of or inside complex supergates
				if fSkipInv != 0 && pGate0.tDelayMax == 0 {
					continue
				}
				// compute area
				Area = AreaMio + pGate0.Area
				if areaOver(Area) {
					break
				}

				pSupers[0] = pGate0
				uTruths[0][0] = pGate0.uTruth[0]
				uTruths[0][1] = pGate0.uTruth[1]
				ptPinDelays[0] = pGate0.ptDelays[:]
				if tryGate() != 0 {
					goto done
				}
			}
		case 2: // two-input root gate
			for i0 = 0; i0 < nGatesLimit; i0++ {
				if pGate0 = ppGatesLimit[i0]; pGate0 == nil {
					break
				}
				Area0 = AreaMio + pGate0.Area
				if areaOver(Area0) {
					break
				}
				pSupers[0] = pGate0
				uTruths[0][0] = pGate0.uTruth[0]
				uTruths[0][1] = pGate0.uTruth[1]
				ptPinDelays[0] = pGate0.ptDelays[:]
				for i1 = 0; i1 < nGatesLimit; i1++ {
					if pGate1 = ppGatesLimit[i1]; pGate1 == nil {
						break
					}
					if i1 != i0 {
						if fTimeOut != 0 {
							goto done
						}
						fTimeOut = Super_CheckTimeout(pProgress, pMan)
						// compute area
						Area = Area0 + pGate1.Area
						if areaOver(Area) {
							break
						}

						pSupers[1] = pGate1
						uTruths[1][0] = pGate1.uTruth[0]
						uTruths[1][1] = pGate1.uTruth[1]
						ptPinDelays[1] = pGate1.ptDelays[:]
						if tryGate() != 0 {
							goto done
						}
					}
				}
			}
		case 3: // three-input root gate
			for i0 = 0; i0 < nGatesLimit; i0++ {
				if pGate0 = ppGatesLimit[i0]; pGate0 == nil {
					break
				}
				Area0 = AreaMio + pGate0.Area
				if areaOver(Area0) {
					break
				}
				pSupers[0] = pGate0
				uTruths[0][0] = pGate0.uTruth[0]
				uTruths[0][1] = pGate0.uTruth[1]
				ptPinDelays[0] = pGate0.ptDelays[:]

				for i1 = 0; i1 < nGatesLimit; i1++ {
					if pGate1 = ppGatesLimit[i1]; pGate1 == nil {
						break
					}
					if i1 != i0 {
						Area1 = Area0 + pGate1.Area
						if areaOver(Area1) {
							break
						}
						pSupers[1] = pGate1
						uTruths[1][0] = pGate1.uTruth[0]
						uTruths[1][1] = pGate1.uTruth[1]
						ptPinDelays[1] = pGate1.ptDelays[:]

						for i2 = 0; i2 < nGatesLimit; i2++ {
							if pGate2 = ppGatesLimit[i2]; pGate2 == nil {
								break
							}
							if i2 != i0 && i2 != i1 {
								if fTimeOut != 0 {
									goto done
								}
								fTimeOut = Super_CheckTimeout(pProgress, pMan)
								// compute area
								Area = Area1 + pGate2.Area
								if areaOver(Area) {
									break
								}
								pSupers[2] = pGate2
								uTruths[2][0] = pGate2.uTruth[0]
								uTruths[2][1] = pGate2.uTruth[1]
								ptPinDelays[2] = pGate2.ptDelays[:]

								if tryGate() != 0 {
									goto done
								}
							}
						}
					}
				}
			}
		case 4: // four-input root gate
			for i0 = 0; i0 < nGatesLimit; i0++ {
				if pGate0 = ppGatesLimit[i0]; pGate0 == nil {
					break
				}
				Area0 = AreaMio + pGate0.Area
				if areaOver(Area0) {
					break
				}
				pSupers[0] = pGate0
				uTruths[0][0] = pGate0.uTruth[0]
				uTruths[0][1] = pGate0.uTruth[1]
				ptPinDelays[0] = pGate0.ptDelays[:]

				for i1 = 0; i1 < nGatesLimit; i1++ {
					if pGate1 = ppGatesLimit[i1]; pGate1 == nil {
						break
					}
					if i1 != i0 {
						Area1 = Area0 + pGate1.Area
						if areaOver(Area1) {
							break
						}
						pSupers[1] = pGate1
						uTruths[1][0] = pGate1.uTruth[0]
						uTruths[1][1] = pGate1.uTruth[1]
						ptPinDelays[1] = pGate1.ptDelays[:]

						for i2 = 0; i2 < nGatesLimit; i2++ {
							if pGate2 = ppGatesLimit[i2]; pGate2 == nil {
								break
							}
							if i2 != i0 && i2 != i1 {
								Area2 = Area1 + pGate2.Area
								if areaOver(Area2) {
									break
								}
								pSupers[2] = pGate2
								uTruths[2][0] = pGate2.uTruth[0]
								uTruths[2][1] = pGate2.uTruth[1]
								ptPinDelays[2] = pGate2.ptDelays[:]

								for i3 = 0; i3 < nGatesLimit; i3++ {
									if pGate3 = ppGatesLimit[i3]; pGate3 == nil {
										break
									}
									if i3 != i0 && i3 != i1 && i3 != i2 {
										if fTimeOut != 0 {
											goto done
										}
										fTimeOut = Super_CheckTimeout(pProgress, pMan)
										// compute area
										Area = Area2 + pGate3.Area
										if areaOver(Area) {
											break
										}
										pSupers[3] = pGate3
										uTruths[3][0] = pGate3.uTruth[0]
										uTruths[3][1] = pGate3.uTruth[1]
										ptPinDelays[3] = pGate3.ptDelays[:]

										if tryGate() != 0 {
											goto done
										}
									}
								}
							}
						}
					}
				}
			}
		case 5: // five-input root gate
			for i0 = 0; i0 < nGatesLimit; i0++ {
				if pGate0 = ppGatesLimit[i0]; pGate0 == nil {
					break
				}
				Area0 = AreaMio + pGate0.Area
				if areaOver(Area0) {
					break
				}
				pSupers[0] = pGate0
				uTruths[0][0] = pGate0.uTruth[0]
				uTruths[0][1] = pGate0.uTruth[1]
				ptPinDelays[0] = pGate0.ptDelays[:]

				for i1 = 0; i1 < nGatesLimit; i1++ {
					if pGate1 = ppGatesLimit[i1]; pGate1 == nil {
						break
					}
					if i1 != i0 {
						Area1 = Area0 + pGate1.Area
						if areaOver(Area1) {
							break
						}
						pSupers[1] = pGate1
						uTruths[1][0] = pGate1.uTruth[0]
						uTruths[1][1] = pGate1.uTruth[1]
						ptPinDelays[1] = pGate1.ptDelays[:]

						for i2 = 0; i2 < nGatesLimit; i2++ {
							if pGate2 = ppGatesLimit[i2]; pGate2 == nil {
								break
							}
							if i2 != i0 && i2 != i1 {
								Area2 = Area1 + pGate2.Area
								if areaOver(Area2) {
									break
								}
								pSupers[2] = pGate2
								uTruths[2][0] = pGate2.uTruth[0]
								uTruths[2][1] = pGate2.uTruth[1]
								ptPinDelays[2] = pGate2.ptDelays[:]

								for i3 = 0; i3 < nGatesLimit; i3++ {
									if pGate3 = ppGatesLimit[i3]; pGate3 == nil {
										break
									}
									if i3 != i0 && i3 != i1 && i3 != i2 {
										Area3 = Area2 + pGate3.Area
										if areaOver(Area3) {
											break
										}
										pSupers[3] = pGate3
										uTruths[3][0] = pGate3.uTruth[0]
										uTruths[3][1] = pGate3.uTruth[1]
										ptPinDelays[3] = pGate3.ptDelays[:]

										for i4 = 0; i4 < nGatesLimit; i4++ {
											if pGate4 = ppGatesLimit[i4]; pGate4 == nil {
												break
											}
											if i4 != i0 && i4 != i1 && i4 != i2 && i4 != i3 {
												if fTimeOut != 0 {
													goto done
												}
												fTimeOut = Super_CheckTimeout(pProgress, pMan)
												// compute area
												Area = Area3 + pGate4.Area
												if areaOver(Area) {
													break
												}
												pSupers[4] = pGate4
												uTruths[4][0] = pGate4.uTruth[0]
												uTruths[4][1] = pGate4.uTruth[1]
												ptPinDelays[4] = pGate4.ptDelays[:]

												if tryGate() != 0 {
													goto done
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		case 6: // six-input root gate
			for i0 = 0; i0 < nGatesLimit; i0++ {
				if pGate0 = ppGatesLimit[i0]; pGate0 == nil {
					break
				}
				Area0 = AreaMio + pGate0.Area
				if areaOver(Area0) {
					break
				}
				pSupers[0] = pGate0
				uTruths[0][0] = pGate0.uTruth[0]
				uTruths[0][1] = pGate0.uTruth[1]
				ptPinDelays[0] = pGate0.ptDelays[:]

				for i1 = 0; i1 < nGatesLimit; i1++ {
					if pGate1 = ppGatesLimit[i1]; pGate1 == nil {
						break
					}
					if i1 != i0 {
						Area1 = Area0 + pGate1.Area
						if areaOver(Area1) {
							break
						}
						pSupers[1] = pGate1
						uTruths[1][0] = pGate1.uTruth[0]
						uTruths[1][1] = pGate1.uTruth[1]
						ptPinDelays[1] = pGate1.ptDelays[:]

						for i2 = 0; i2 < nGatesLimit; i2++ {
							if pGate2 = ppGatesLimit[i2]; pGate2 == nil {
								break
							}
							if i2 != i0 && i2 != i1 {
								Area2 = Area1 + pGate2.Area
								if areaOver(Area2) {
									break
								}
								pSupers[2] = pGate2
								uTruths[2][0] = pGate2.uTruth[0]
								uTruths[2][1] = pGate2.uTruth[1]
								ptPinDelays[2] = pGate2.ptDelays[:]

								for i3 = 0; i3 < nGatesLimit; i3++ {
									if pGate3 = ppGatesLimit[i3]; pGate3 == nil {
										break
									}
									if i3 != i0 && i3 != i1 && i3 != i2 {
										Area3 = Area2 + pGate3.Area
										if areaOver(Area3) {
											break
										}
										pSupers[3] = pGate3
										uTruths[3][0] = pGate3.uTruth[0]
										uTruths[3][1] = pGate3.uTruth[1]
										ptPinDelays[3] = pGate3.ptDelays[:]

										for i4 = 0; i4 < nGatesLimit; i4++ {
											if pGate4 = ppGatesLimit[i4]; pGate4 == nil {
												break
											}
											if i4 != i0 && i4 != i1 && i4 != i2 && i4 != i3 {
												if fTimeOut != 0 {
													break
												}
												fTimeOut = Super_CheckTimeout(pProgress, pMan)
												// compute area
												Area4 = Area3 + pGate4.Area
												// (C tests the stale Area here, not Area4)
												if areaOver(Area) {
													break
												}
												pSupers[4] = pGate4
												uTruths[4][0] = pGate4.uTruth[0]
												uTruths[4][1] = pGate4.uTruth[1]
												ptPinDelays[4] = pGate4.ptDelays[:]

												for i5 = 0; i5 < nGatesLimit; i5++ {
													if pGate5 = ppGatesLimit[i5]; pGate5 == nil {
														break
													}
													if i5 != i0 && i5 != i1 && i5 != i2 && i5 != i3 && i5 != i4 {
														if fTimeOut != 0 {
															goto done
														}
														fTimeOut = Super_CheckTimeout(pProgress, pMan)
														// compute area
														Area = Area4 + pGate5.Area
														if areaOver(Area) {
															break
														}
														pSupers[5] = pGate5
														uTruths[5][0] = pGate5.uTruth[0]
														uTruths[5][1] = pGate5.uTruth[1]
														ptPinDelays[5] = pGate5.ptDelays[:]

														if tryGate() != 0 {
															goto done
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		default:
			assert(false, "root gate with more than six fanins")
		}
	}
done:
	extra.Extra_ProgressBarStop(pProgress)
	return pMan
}

/*
*Function*************************************************************

	Synopsis    [Transfers gates from table into the array.]

**********************************************************************
*/
func Super_CheckTimeout(pPro *extra.ProgressBar, pMan *Super_Man_t) int32 {
	TimeNow := util.Abc_Clock()
	if TimeNow > pMan.TimePrint {
		pMan.TimeSec++
		extra.Extra_ProgressBarUpdate(pPro, pMan.TimeSec, "")
		pMan.TimePrint = util.Abc_Clock() + util.CLOCKS_PER_SEC
	}
	if pMan.TimeStop != 0 && TimeNow > pMan.TimeStop {
		fmt.Printf("Timeout!\n")
		return 1
	}
	pMan.nTried++
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Transfers gates from table into the array.]

**********************************************************************
*/
func Super_TranferGatesToArray(pMan *Super_Man_t) {
	// put the gates fron the table into the array
	pMan.pGates = make([]*Super_Gate_t, 0, pMan.nAdded)
	pMan.nGates = 0
	stmm_foreach_item(pMan.tTable, func(Key uint64, pList *Super_Gate_t) {
		for pGate := pList; pGate != nil; pGate = pGate.pNext {
			pMan.pGates = append(pMan.pGates, pGate)
			pMan.nGates++
		}
	})
	//    assert( pMan->nGates == pMan->nAdded - pMan->nRemoved );
}

/*
*Function*************************************************************

	Synopsis    [Adds one supergate into the unique table.]

**********************************************************************
*/
func Super_AddGateToTable(pMan *Super_Man_t, pGate *Super_Gate_t) {
	var ppList **Super_Gate_t
	//    Key = pGate->uTruth[0] + 2003 * pGate->uTruth[1];
	Key := uint64(pGate.uTruth[0] ^ pGate.uTruth[1])
	if stmm_find_or_add(pMan.tTable, Key, &ppList) == 0 {
		*ppList = nil
		pMan.nClasses++
	}
	pGate.pNext = *ppList
	*ppList = pGate
	pMan.nAdded++
}

/*
*Function*************************************************************

	Synopsis    [Check the manager's unique table for comparable gates.]

	Description [Returns 0 if the gate is dominated by others. Returns 1
	if the gate is new or is better than the available ones. In this case,
	cleans the table by removing the gates that are worse than the given one.]

**********************************************************************
*/
func Super_CompareGates(pMan *Super_Man_t, uTruth []uint32, Area float32, tPinDelaysRes []float32, nPins int32) int32 {
	var ppList **Super_Gate_t
	var pPrev, pGate, pGate2 *Super_Gate_t
	var i, fNewIsBetter, fGateIsBetter int32

	// skip constant functions
	if pMan.nVarsMax < 6 {
		if uTruth[0] == 0 || ^uTruth[0] == 0 {
			return 0
		}
	} else {
		if (uTruth[0] == 0 && uTruth[1] == 0) || (^uTruth[0] == 0 && ^uTruth[1] == 0) {
			return 0
		}
	}

	// get hold of the place where the entry is stored
	//    Key = uTruth[0] + 2003 * uTruth[1];
	Key := uint64(uTruth[0] ^ uTruth[1])
	if stmm_find(pMan.tTable, Key, &ppList) == 0 {
		return 1
	}
	// the entry with this truth table is found
	pPrev = nil
	for pGate, pGate2 = *ppList, nextOf(*ppList); pGate != nil; pGate, pGate2 = pGate2, nextOf(pGate2) {
		pMan.nLookups++
		if pGate.uTruth[0] != uTruth[0] || pGate.uTruth[1] != uTruth[1] {
			pMan.nAliases++
			continue
		}
		fGateIsBetter = 0
		fNewIsBetter = 0
		if float64(pGate.Area)+SUPER_EPSILON < float64(Area) {
			fGateIsBetter = 1
		} else if float64(pGate.Area) > float64(Area)+SUPER_EPSILON {
			fNewIsBetter = 1
		}
		for i = 0; i < nPins; i++ {
			if float64(pGate.ptDelays[i]) == SUPER_NO_VAR || float64(tPinDelaysRes[i]) == SUPER_NO_VAR {
				continue
			}
			if float64(pGate.ptDelays[i])+SUPER_EPSILON < float64(tPinDelaysRes[i]) {
				fGateIsBetter = 1
			} else if float64(pGate.ptDelays[i]) > float64(tPinDelaysRes[i])+SUPER_EPSILON {
				fNewIsBetter = 1
			}
			if fGateIsBetter != 0 && fNewIsBetter != 0 {
				break
			}
		}
		// consider 4 cases
		if fGateIsBetter != 0 && fNewIsBetter != 0 { // Pareto points; save both
			pPrev = pGate
		} else if fNewIsBetter != 0 { // gate is worse; remove the gate
			if pPrev == nil {
				*ppList = pGate.pNext
			} else {
				pPrev.pNext = pGate.pNext
			}
			pMan.memRecycle(pGate)
			pMan.nRemoved++
		} else if fGateIsBetter != 0 { // new is worse, already dominated no need to see others
			return 0
		} else { // if ( !fGateIsBetter && !fNewIsBetter ) // they are identical, no need to see others
			return 0
		}
	}
	return 1
}

// nextOf is the C loop's "pGate2 = pGate? pGate->pNext: NULL".
func nextOf(pGate *Super_Gate_t) *Super_Gate_t {
	if pGate != nil {
		return pGate.pNext
	}
	return nil
}

/*
*Function*************************************************************

	Synopsis    [Create a new supergate.]

**********************************************************************
*/
func Super_CreateGateNew(pMan *Super_Man_t, pRoot *Mio_Gate_t, pSupers []*Super_Gate_t, nSupers int32,
	uTruth []uint32, Area float32, tPinDelaysRes []float32, tDelayMax float32, nPins int32) *Super_Gate_t {
	pSuper := pMan.memFetch()
	*pSuper = Super_Gate_t{}
	pSuper.pRoot = pRoot
	pSuper.uTruth[0] = uTruth[0]
	pSuper.uTruth[1] = uTruth[1]
	copy(pSuper.ptDelays[:nPins], tPinDelaysRes[:nPins])
	pSuper.Area = Area
	pSuper.nFanins = uint32(nSupers) & 0x3F
	copy(pSuper.pFanins[:nSupers], pSupers[:nSupers])
	pSuper.pNext = nil
	pSuper.tDelayMax = tDelayMax
	return pSuper
}

/*
*Function*************************************************************

	Synopsis    [Starts the manager.]

**********************************************************************
*/
func Super_ManStart() *Super_Man_t {
	pMan := &Super_Man_t{}
	pMan.tTable = stmm_init_table()
	return pMan
}

/*
*Function*************************************************************

	Synopsis    [Stops the manager.]

**********************************************************************
*/
func Super_ManStop(pMan *Super_Man_t) {
	pMan.pMemFree = nil
	pMan.tTable = nil
	pMan.pGates = nil
}

/*
*Function*************************************************************

	Synopsis    [Writes the supergate library into the file.]

**********************************************************************
*/
func Super_Write(pMan *Super_Man_t) *Vec_Str_t {
	var vStr *Vec_Str_t
	var clk int64

	if pMan.nGates < 1 {
		fmt.Printf("The generated library is empty. No output file written.\n")
		return nil
	}

	// Filters the supergates by removing those that have fewer inputs than
	// the given limit, provided that the inputs are not consequtive.
	// For example, NAND2(a,c) is removed, but NAND2(a,b) is left,
	// because a and b are consequtive.
	pMan.pGates = make([]*Super_Gate_t, 0, pMan.nAdded)
	pMan.nGates = 0
	stmm_foreach_item(pMan.tTable, func(Key uint64, pGateRoot *Super_Gate_t) {
		for pGate := pGateRoot; pGate != nil; pGate = pGate.pNext {
			// skip the elementary variables
			if pGate.pRoot == nil {
				continue
			}
			// skip the non-consequtive gates
			fZeroFound := 0
			var v int32
			for v = 0; v < pMan.nVarsMax; v++ {
				if float64(pGate.ptDelays[v]) < SUPER_NO_VAR+SUPER_EPSILON {
					fZeroFound = 1
				} else if fZeroFound != 0 {
					break
				}
			}
			if v < pMan.nVarsMax {
				continue
			}
			// save the unique gate
			pMan.pGates = append(pMan.pGates, pGate)
			pMan.nGates++
		}
	})

	clk = util.Abc_Clock()
	// sort the supergates by truth table
	sort.SliceStable(pMan.pGates, func(a, b int) bool { return Super_WriteCompare(pMan.pGates[a], pMan.pGates[b]) < 0 })
	assert(pMan.nGates == 0 || Super_WriteCompare(pMan.pGates[0], pMan.pGates[pMan.nGates-1]) <= 0, "sorted by truth")
	if pMan.fVerbose != 0 {
		util.Abc_PrintTime(1, "Sorting", util.Abc_Clock()-clk)
	}

	// write library in the old format
	clk = util.Abc_Clock()
	if pMan.fWriteOldFormat != 0 {
		Super_WriteLibrary(pMan)
	}
	if pMan.fVerbose != 0 {
		util.Abc_PrintTime(1, "Writing old format", util.Abc_Clock()-clk)
	}

	// write the tree-like structure of supergates
	clk = util.Abc_Clock()
	vStr = Super_WriteLibraryTreeStr(pMan)
	if pMan.fVerbose != 0 {
		util.Abc_PrintTime(1, "Writing new format", util.Abc_Clock()-clk)
	}
	return vStr
}

/*
*Function*************************************************************

	Synopsis    [Writes the file header.]

**********************************************************************
*/
func superFileHeaderLines(pMan *Super_Man_t) []string {
	skip := "-s"
	if pMan.fSkipInv != 0 {
		skip = ""
	}
	return []string{
		"#\n",
		fmt.Sprintf("# Supergate library derived for \"%s\" on %s.\n", pMan.pName, extra.Extra_TimeStamp()),
		"#\n",
		fmt.Sprintf("# Command line: \"super -I %d -L %d -N %d -T %d -D %.2f -A %.2f %s %s\".\n",
			pMan.nVarsMax, pMan.nLevels, pMan.nGatesMax, pMan.TimeLimit, pMan.tDelayMax, pMan.tAreaMax, skip, pMan.pName),
		"#\n",
		fmt.Sprintf("# The number of inputs      = %10d.\n", pMan.nVarsMax),
		fmt.Sprintf("# The number of levels      = %10d.\n", pMan.nLevels),
		fmt.Sprintf("# The maximum delay         = %10.2f.\n", pMan.tDelayMax),
		fmt.Sprintf("# The maximum area          = %10.2f.\n", pMan.tAreaMax),
		fmt.Sprintf("# The maximum runtime (sec) = %10d.\n", pMan.TimeLimit),
		"#\n",
		fmt.Sprintf("# The number of attempts    = %10d.\n", pMan.nTried),
		fmt.Sprintf("# The number of supergates  = %10d.\n", pMan.nGates),
		fmt.Sprintf("# The number of functions   = %10d.\n", pMan.nUnique),
		fmt.Sprintf("# The total functions       = %.0f (2^%d).\n", math.Pow(2, float64(pMan.nMints)), pMan.nMints),
		"#\n",
		fmt.Sprintf("# Generation time           = %10.2f sec.\n", float32(pMan.Time)/float32(util.CLOCKS_PER_SEC)),
		"#\n",
		fmt.Sprintf("%s\n", pMan.pName),
		fmt.Sprintf("%d\n", pMan.nVarsMax),
		fmt.Sprintf("%d\n", pMan.nGates),
	}
}
func Super_WriteFileHeader(pMan *Super_Man_t, pFile io.Writer) {
	for _, line := range superFileHeaderLines(pMan) {
		fmt.Fprint(pFile, line)
	}
}
func Super_WriteFileHeaderStr(pMan *Super_Man_t, vStr *Vec_Str_t) {
	for _, line := range superFileHeaderLines(pMan) {
		vec.Vec_StrPrintStr(vStr, line)
	}
}

/*
*Function*************************************************************

	Synopsis    [Compares two gates.]

**********************************************************************
*/
func Super_WriteCompare(pG1, pG2 *Super_Gate_t) int {
	pTruth1 := pG1.uTruth
	pTruth2 := pG2.uTruth
	if pTruth1[1] < pTruth2[1] {
		return -1
	}
	if pTruth1[1] > pTruth2[1] {
		return 1
	}
	if pTruth1[0] < pTruth2[0] {
		return -1
	}
	if pTruth1[0] > pTruth2[0] {
		return 1
	}
	return 0
}
func Super_DelayCompare(pG1, pG2 *Super_Gate_t) int {
	if pG1.tDelayMax < pG2.tDelayMax {
		return -1
	}
	if pG1.tDelayMax > pG2.tDelayMax {
		return 1
	}
	return 0
}
func Super_AreaCompare(pG1, pG2 *Super_Gate_t) int {
	if pG1.Area < pG2.Area {
		return -1
	}
	if pG1.Area > pG2.Area {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Writes the gates into the file.]

**********************************************************************
*/
func Super_WriteLibraryGateName_rec(pGate *Super_Gate_t, pBuffer *[]byte) {
	if pGate.pRoot == nil {
		*pBuffer = append(*pBuffer, byte('a'+pGate.Number))
		return
	}
	*pBuffer = append(*pBuffer, mio.Mio_GateReadName(pGate.pRoot)...)
	*pBuffer = append(*pBuffer, '(')
	for i := 0; i < int(pGate.nFanins); i++ {
		if i != 0 {
			*pBuffer = append(*pBuffer, ',')
		}
		Super_WriteLibraryGateName_rec(pGate.pFanins[i], pBuffer)
	}
	*pBuffer = append(*pBuffer, ')')
}
func Super_WriteLibraryGateName(pGate *Super_Gate_t) string {
	var Buffer []byte
	Super_WriteLibraryGateName_rec(pGate, &Buffer)
	return string(Buffer)
}
func Super_WriteLibraryGate(pFile io.Writer, pMan *Super_Man_t, pGate *Super_Gate_t, Num int32) {
	fmt.Fprintf(pFile, "%04d  ", Num)                            // the number
	extra.Extra_PrintBinary(pFile, pGate.uTruth[:], pMan.nMints) // the truth table
	fmt.Fprintf(pFile, "   %5.2f", pGate.tDelayMax)              // the max delay
	fmt.Fprintf(pFile, "  ")
	for i := int32(0); i < pMan.nVarsMax; i++ { // the pin-to-pin delays
		d := float64(pGate.ptDelays[i])
		if d == SUPER_NO_VAR {
			d = 0.0
		}
		fmt.Fprintf(pFile, " %5.2f", d)
	}
	fmt.Fprintf(pFile, "   %5.2f", pGate.Area) // the area
	fmt.Fprintf(pFile, "   ")
	fmt.Fprintf(pFile, "%s", Super_WriteLibraryGateName(pGate)) // the symbolic expression
	fmt.Fprintf(pFile, "\n")
}
func Super_WriteLibrary(pMan *Super_Man_t) {
	var pGate, pGateNext *Super_Gate_t
	var i, Counter int32

	// get the file name
	FileName := fmt.Sprintf("%s.super_old", extra.Extra_FileNameGeneric(pMan.pName))

	// count the number of unique functions
	pMan.nUnique = 1
	for i = 0; i < pMan.nGates; i++ {
		if pGate = pMan.pGates[i]; pGate == nil {
			break
		}
		if i == pMan.nGates-1 {
			break
		}
		// print the newline if this gate is different from the following one
		pGateNext = pMan.pGates[i+1]
		if pGateNext.uTruth[0] != pGate.uTruth[0] || pGateNext.uTruth[1] != pGate.uTruth[1] {
			pMan.nUnique++
		}
	}

	// start the file
	pFile, err := os.Create(FileName)
	if err != nil {
		panic(err)
	}
	Super_WriteFileHeader(pMan, pFile)

	// print the gates
	Counter = 0
	for i = 0; i < pMan.nGates; i++ {
		if pGate = pMan.pGates[i]; pGate == nil {
			break
		}
		Counter++
		Super_WriteLibraryGate(pFile, pMan, pGate, Counter)
		if i == pMan.nGates-1 {
			break
		}
		// print the newline if this gate is different from the following one
		pGateNext = pMan.pGates[i+1]
		if pGateNext.uTruth[0] != pGate.uTruth[0] || pGateNext.uTruth[1] != pGate.uTruth[1] {
			fmt.Fprintf(pFile, "\n")
		}
	}
	assert(Counter == pMan.nGates, "Counter == pMan->nGates")
	pFile.Close()

	if pMan.fVerbose != 0 {
		fmt.Printf("The supergates are written using old format \"%s\" ", FileName)
		fmt.Printf("(%0.3f MB).\n", float64(extra.Extra_FileSize(FileName))/float64(1<<20))
	}
}

// superWriteRecRoot reads pSuper->pRoot as the writers do; C would read the
// memory manager's free-list link of a recycled (dominated) supergate.
func superWriteRecRoot(pSuper *Super_Gate_t) *Mio_Gate_t {
	if pSuper.fRecycled {
		panic("super: writing a recycled supergate (C reads a freed entry)")
	}
	return pSuper.pRoot
}

/*
*Function*************************************************************

	Synopsis    [Recursively writes the gates.]

**********************************************************************
*/
func Super_WriteLibraryTreeFile_rec(pFile io.Writer, pMan *Super_Man_t, pSuper *Super_Gate_t, pCounter *int32) {
	// skip an elementary variable and a gate that was already written
	if pSuper.fVar != 0 || pSuper.Number > 0 {
		return
	}
	// write the fanins
	nFanins := mio.Mio_GateReadPinNum(superWriteRecRoot(pSuper))
	for i := int32(0); i < nFanins; i++ {
		Super_WriteLibraryTreeFile_rec(pFile, pMan, pSuper.pFanins[i], pCounter)
	}
	// finally write the gate
	pSuper.Number = uint32(*pCounter) & 0xFFFFFF
	*pCounter++
	if pSuper.fSuper != 0 {
		fmt.Fprintf(pFile, "* ")
	}
	fmt.Fprintf(pFile, "%s", mio.Mio_GateReadName(pSuper.pRoot))
	for i := int32(0); i < nFanins; i++ {
		fmt.Fprintf(pFile, " %d", pSuper.pFanins[i].Number)
	}
	fmt.Fprintf(pFile, "\n")
}
func Super_WriteLibraryTreeFile(pMan *Super_Man_t) {
	var i, Counter int32

	// get the file name
	FileName := fmt.Sprintf("%s.super", extra.Extra_FileNameGeneric(pMan.pName))

	// write the elementary variables
	pFile, err := os.Create(FileName)
	if err != nil {
		panic(err)
	}
	Super_WriteFileHeader(pMan, pFile)
	// write the place holder for the number of lines
	posStart, _ := pFile.Seek(0, io.SeekCurrent)
	fmt.Fprintf(pFile, "         \n")
	// mark the real supergates
	for i = 0; i < pMan.nGates; i++ {
		pMan.pGates[i].fSuper = 1
	}
	// write the supergates
	Counter = pMan.nVarsMax
	for i = 0; i < pMan.nGates; i++ {
		Super_WriteLibraryTreeFile_rec(pFile, pMan, pMan.pGates[i], &Counter)
	}
	// write the number of lines
	pFile.WriteAt([]byte(fmt.Sprintf("%d", Counter)), posStart)
	pFile.Close()

	if pMan.fVerbose != 0 {
		fmt.Printf("The supergates are written using new format \"%s\" ", FileName)
		fmt.Printf("(%0.3f MB).\n", float64(extra.Extra_FileSize(FileName))/float64(1<<20))
	}
}

/*
*Function*************************************************************

	Synopsis    [Recursively writes the gates.]

**********************************************************************
*/
func Super_WriteLibraryTreeStr_rec(vStr *Vec_Str_t, pMan *Super_Man_t, pSuper *Super_Gate_t, pCounter *int32) {
	// skip an elementary variable and a gate that was already written
	if pSuper.fVar != 0 || pSuper.Number > 0 {
		return
	}
	// write the fanins
	nFanins := mio.Mio_GateReadPinNum(superWriteRecRoot(pSuper))
	for i := int32(0); i < nFanins; i++ {
		Super_WriteLibraryTreeStr_rec(vStr, pMan, pSuper.pFanins[i], pCounter)
	}
	// finally write the gate
	pSuper.Number = uint32(*pCounter) & 0xFFFFFF
	*pCounter++
	if pSuper.fSuper != 0 {
		vec.Vec_StrPrintStr(vStr, "* ")
	} else {
		vec.Vec_StrPrintStr(vStr, "")
	}
	vec.Vec_StrPrintStr(vStr, mio.Mio_GateReadName(pSuper.pRoot))
	for i := int32(0); i < nFanins; i++ {
		vec.Vec_StrPrintStr(vStr, " ")
		vec.Vec_StrPrintNum(vStr, int32(pSuper.pFanins[i].Number))
	}
	vec.Vec_StrPrintStr(vStr, "\n")
}
func Super_WriteLibraryTreeStr(pMan *Super_Man_t) *Vec_Str_t {
	var i, Counter int32
	// write the elementary variables
	vStr := vec.Vec_StrAlloc(1000)
	Super_WriteFileHeaderStr(pMan, vStr)
	// write the place holder for the number of lines
	posStart := vec.Vec_StrSize(vStr)
	for i = 0; i < 9; i++ {
		vec.Vec_StrPush(vStr, ' ')
	}
	vec.Vec_StrPush(vStr, '\n')
	// mark the real supergates
	for i = 0; i < pMan.nGates; i++ {
		pMan.pGates[i].fSuper = 1
	}
	// write the supergates
	Counter = pMan.nVarsMax
	for i = 0; i < pMan.nGates; i++ {
		Super_WriteLibraryTreeStr_rec(vStr, pMan, pMan.pGates[i], &Counter)
	}
	vec.Vec_StrPush(vStr, 0)
	// write the number of lines
	pInsert := fmt.Sprintf("%d", Counter)
	for i = 0; i < int32(len(pInsert)); i++ {
		vec.Vec_StrWriteEntry(vStr, posStart+int(i), pInsert[i])
	}
	return vStr
}
func Super_WriteLibraryTree(pMan *Super_Man_t) {
	pFileName := extra.Extra_FileNameGenericAppend(pMan.pName, ".super")
	pFile, err := os.Create(pFileName)
	if err != nil {
		fmt.Printf("Cannot open output file \"%s\".\n", pFileName)
		return
	}
	vStr := Super_WriteLibraryTreeStr(pMan)
	pFile.Write(vec.Vec_StrArray(vStr)[:vec.Vec_StrSize(vStr)])
	pFile.Close()
	vec.Vec_StrFree(vStr)
	// report the result of writing
	if pMan.fVerbose != 0 {
		fmt.Printf("The supergates are written using new format \"%s\" ", pFileName)
		fmt.Printf("(%0.3f MB).\n", float64(extra.Extra_FileSize(pFileName))/float64(1<<20))
	}
}
