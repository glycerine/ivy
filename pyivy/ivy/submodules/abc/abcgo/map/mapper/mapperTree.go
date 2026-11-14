package mapper

import (
	"fmt"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/st"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"strconv"
	"strings"
)

/**CFile****************************************************************

  FileName    [mapperTree.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperTree.c,v 1.9 2005/01/23 06:59:45 alanmi Exp $]

***********************************************************************/

//#ifdef __linux__
//#include <libgen.h>
//#endif

//#include "mapperInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void      Map_LibraryAddFaninDelays( Map_SuperLib_t * pLib, Map_Super_t * pGate, Map_Super_t * pFanin, Mio_Pin_t * pPin );

// static int       Map_LibraryGetMaxSuperPi_rec( Map_Super_t * pGate );

// static unsigned  Map_LibraryGetGateSupp_rec( Map_Super_t * pGate );

// fanout limits
var s_MapFanoutLimits = [10]int32{1, 10, 5, 2, 1, 1, 1}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Reads one gate.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryReadGateTree(lib *Map_SuperLib_t, line string, number, maxVars int32) *Map_Super_t {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil
	}
	gate := &Map_SuperStruct_t_{Num: number}
	if fields[0] == "*" {
		gate.fSuper = 1
		fields = fields[1:]
	} else if strings.HasPrefix(fields[0], "*") {
		gate.fSuper = 1
		fields[0] = strings.TrimPrefix(fields[0], "*")
	}
	if len(fields) == 0 {
		return nil
	}
	gate.pRoot = mio.Mio_LibraryReadGateByName(lib.pGenlib, fields[0], "")
	if gate.pRoot == nil {
		return nil
	}
	pinCount := mio.Mio_GateReadPinNum(gate.pRoot)
	if pinCount < 0 || pinCount > int32(len(gate.pFanins)) {
		return nil
	}
	gate.nFanLimit = uint32(s_MapFanoutLimits[pinCount])
	children := make([]int, 0, maxVars)
	formula := ""
	for i := 1; i < len(fields); i++ {
		if strings.HasPrefix(fields[i], "#") {
			formula = strings.TrimPrefix(fields[i], "#")
			if formula == "" && i+1 < len(fields) {
				formula = fields[i+1]
			}
			break
		}
		id, err := strconv.Atoi(fields[i])
		if err != nil || id < 0 || id >= len(lib.ppSupers) {
			return nil
		}
		children = append(children, id)
		if int32(len(children)) > maxVars {
			return nil
		}
	}
	if int32(len(children)) != pinCount {
		return nil
	}
	gate.nFanins = uint32(len(children))
	for i, id := range children {
		gate.pFanins[i] = lib.ppSupers[id]
		if gate.pFanins[i] == nil {
			return nil
		}
	}
	gate.pFormula = formula
	return gate
}

func Vec_StrGets(buffer []byte, bufferSize int32, source *Vec_Str_t, position *int32) int32 {
	if bufferSize <= 1 || source == nil || position == nil || len(buffer) < int(bufferSize) {
		panic("mapper: invalid string reader arguments")
	}
	data := vec.Vec_StrArray(source)
	start := int(*position)
	if start < 0 || start > len(data) {
		panic("mapper: invalid string reader position")
	}
	if start == len(data) {
		buffer[0] = 0
		return 0
	}
	for i := start; i < len(data); i++ {
		offset := i - start
		buffer[offset] = data[i]
		if data[i] == 0 {
			*position += int32(offset)
			return 0
		}
		if data[i] == '\n' || offset == int(bufferSize)-1 {
			*position += int32(offset + 1)
			buffer[offset+1] = 0
			return 1
		}
	}
	return 0
}

func Map_LibraryCompareLibNames(a, b string) int32 {
	normalize := func(s string) string { return strings.NewReplacer(">", "/", "\\", "/").Replace(s) }
	if normalize(a) == normalize(b) {
		return 0
	}
	return 1
}

func Map_LibraryReadFileTreeStr(lib *Map_SuperLib_t, genlib *mio.Mio_Library_t, source *Vec_Str_t, fileName string) int32 {
	if lib == nil || genlib == nil || source == nil {
		return 0
	}
	var lines []string
	var pos int32
	buffer := make([]byte, 5001)
	for Vec_StrGets(buffer, 5000, source, &pos) != 0 {
		end := 0
		for end < len(buffer) && buffer[end] != 0 {
			end++
		}
		lines = append(lines, string(buffer[:end]))
	}
	first := -1
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			first = i
			break
		}
	}
	if first < 0 {
		return 0
	}
	libName := strings.Fields(strings.TrimSpace(lines[first]))
	if len(libName) == 0 || Map_LibraryCompareLibNames(mio.Mio_LibraryReadName(genlib), libName[0]) != 0 {
		fmt.Printf("Supergate library %q does not match the loaded genlib library.\n", fileName)
		return 0
	}
	readCount := func(index int) (int32, bool) {
		if index >= len(lines) {
			return 0, false
		}
		value, err := strconv.Atoi(strings.TrimSpace(lines[index]))
		if err != nil {
			return 0, false
		}
		return int32(value), true
	}
	var ok bool
	if lib.nVarsMax, ok = readCount(first + 1); !ok || lib.nVarsMax < 2 || lib.nVarsMax > 10 {
		return 0
	}
	if lib.nSupersReal, ok = readCount(first + 2); !ok || lib.nSupersReal < 1 || lib.nSupersReal > 10_000_000 {
		return 0
	}
	if lib.nLines, ok = readCount(first + 3); !ok || lib.nLines < 1 || lib.nLines > 10_000_000 {
		return 0
	}
	lib.pGenlib = genlib
	lib.ppSupers = make([]*Map_Super_t, int(lib.nLines)+10_000)
	for i := int32(0); i < lib.nVarsMax; i++ {
		gate := &Map_SuperStruct_t_{Num: i, uTruth: lib.uTruths[i]}
		for k := range gate.tDelaysR {
			gate.tDelaysR[k] = Map_Time_t{Rise: MAP_NO_VAR, Fall: MAP_NO_VAR}
			gate.tDelaysF[k] = Map_Time_t{Rise: MAP_NO_VAR, Fall: MAP_NO_VAR}
		}
		gate.tDelaysR[i].Rise, gate.tDelaysF[i].Fall = 0, 0
		lib.ppSupers[i] = gate
	}
	count := lib.nVarsMax
	for _, line := range lines[first+4:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		gate := Map_LibraryReadGateTree(lib, line, count, lib.nVarsMax)
		if gate == nil {
			return 0
		}
		lib.ppSupers[count] = gate
		count++
	}
	if count != lib.nLines {
		fmt.Printf("The number of lines read (%d) is different from what the file says (%d).\n", count, lib.nLines)
	}
	lib.nSupersAll = count
	real := int32(0)
	for _, gate := range lib.ppSupers[:lib.nLines] {
		if gate == nil {
			panic("mapper: missing declared supergate")
		}
		real += int32(gate.fSuper)
	}
	if real != lib.nSupersReal {
		fmt.Printf("The number of gates read (%d) is different what the file says (%d).\n", real, lib.nSupersReal)
	}
	lib.nSupersReal = real
	return 1
}

func Map_LibraryReadTree(lib *Map_SuperLib_t, genlib *mio.Mio_Library_t, fileName, excludeFile string) int32 {
	assert(lib.pGenlib == nil, "lib.pGenlib == nil")
	var excluded *st.Table
	if excludeFile != "" {
		names := make(map[string]bool)
		count := mio.Mio_LibraryReadExclude(excludeFile, names)
		if count < 0 {
			return 0
		}
		excluded = st.New(st.StringCompare, st.StringHash)
		for name := range names {
			excluded.Insert(name, true)
		}
		fmt.Printf("Read %d gates from exclude file\n", count)
	}
	data, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Cannot open input file %q.\\n", fileName)
		return 0
	}
	text := vec.Vec_StrAllocArrayCopy(data, len(data))
	defer vec.Vec_StrFree(text)
	if Map_LibraryReadFileTreeStr(lib, genlib, text, fileName) == 0 {
		return 0
	}
	return Map_LibraryDeriveGateInfo(lib, excluded)
}

func Map_LibraryDeriveGateInfo(pLib *Map_SuperLib_t, tExcludeGate *st__table) int32 {
	var pGate, pFanin *Map_Super_t
	var pPin *Mio_Pin_t
	var uCanon [2]uint32
	var uTruths [6][2]uint32
	var i, k, nRealVars int32

	// set all the derivable info related to the supergates
	for i = pLib.nVarsMax; i < int32(pLib.nLines); i++ {
		pGate = pLib.ppSupers[i]

		if tExcludeGate != nil {
			// inlining of macro st__is_member
			if _, found := tExcludeGate.Lookup(Mio_GateReadName(pGate.pRoot)); found {
				pGate.fExclude = 1
			}
			for k = 0; k < int32(pGate.nFanins); k++ {
				pFanin = pGate.pFanins[k]
				if pFanin.fExclude != 0 {
					pGate.fExclude = 1
					continue
				}
			}
		}

		// collect the truth tables of the fanins
		for k = 0; k < int32(pGate.nFanins); k++ {
			pFanin = pGate.pFanins[k]
			uTruths[k][0] = pFanin.uTruth[0]
			uTruths[k][1] = pFanin.uTruth[1]
		}
		// derive the new truth table
		Mio_DeriveTruthTable(pGate.pRoot, uTruths[:], int32(pGate.nFanins), 6, pGate.uTruth[:])

		// set the initial delays of the supergate
		for k = 0; k < pLib.nVarsMax; k++ {
			pGate.tDelaysR[k].Rise, pGate.tDelaysR[k].Fall = MAP_NO_VAR, MAP_NO_VAR
			pGate.tDelaysF[k].Rise, pGate.tDelaysF[k].Fall = MAP_NO_VAR, MAP_NO_VAR
		}
		// get the linked list of pins for the given root gate
		pPin = Mio_GateReadPins(pGate.pRoot)
		// update the initial delay of the supergate using info from the corresponding pin
		for k = 0; k < int32(pGate.nFanins); k, pPin = k+1, Mio_PinReadNext(pPin) {
			// if there is no corresponding pin, this is a bug, return fail
			if pPin == nil {
				fmt.Printf("There are less pins than gate inputs.\n")
				return 0
			}
			// update the delay information of k-th fanins info from the corresponding pin
			Map_LibraryAddFaninDelays(pLib, pGate, pGate.pFanins[k], pPin)
		}
		// if there are some pins left, this is a bug, return fail
		if pPin != nil {
			fmt.Printf("There are more pins than gate inputs.\n")
			return 0
		}
		// find the max delay
		pGate.tDelayMax.Rise, pGate.tDelayMax.Fall = MAP_NO_VAR, MAP_NO_VAR
		for k = 0; k < pLib.nVarsMax; k++ {
			// the rise of the output depends on the rise and fall of the output
			if pGate.tDelayMax.Rise < pGate.tDelaysR[k].Rise {
				pGate.tDelayMax.Rise = pGate.tDelaysR[k].Rise
			}
			if pGate.tDelayMax.Rise < pGate.tDelaysR[k].Fall {
				pGate.tDelayMax.Rise = pGate.tDelaysR[k].Fall
			}
			// the fall of the output depends on the rise and fall of the output
			if pGate.tDelayMax.Fall < pGate.tDelaysF[k].Rise {
				pGate.tDelayMax.Fall = pGate.tDelaysF[k].Rise
			}
			if pGate.tDelayMax.Fall < pGate.tDelaysF[k].Fall {
				pGate.tDelayMax.Fall = pGate.tDelaysF[k].Fall
			}

			// inlining of macro MAP_MAX
			pGate.tDelaysF[k].Worst = max(pGate.tDelaysF[k].Fall, pGate.tDelaysF[k].Rise)
			// inlining of macro MAP_MAX
			pGate.tDelaysR[k].Worst = max(pGate.tDelaysR[k].Fall, pGate.tDelaysR[k].Rise)
		}

		// count gates and area of the supergate
		pGate.nGates = 1
		pGate.Area = float32(Mio_GateReadArea(pGate.pRoot))
		for k = 0; k < int32(pGate.nFanins); k++ {
			pGate.nGates += pGate.pFanins[k].nGates
			pGate.Area += pGate.pFanins[k].Area
		}
		// do not add the gate to the table, if this gate is an internal gate
		// of some supegate and does not correspond to a supergate output
		if pGate.fSuper == 0 || pGate.fExclude != 0 {
			continue
		}

		// find the maximum index of a variable in the support of the supergates
		// this is important for two reasons:
		// (1) to limit the number of permutations considered for canonicization
		// (2) to get rid of equivalence phases to speed-up matching
		nRealVars = Map_LibraryGetMaxSuperPi_rec(pGate) + 1
		assert(nRealVars > 0 && nRealVars <= pLib.nVarsMax, "nRealVars > 0 && nRealVars <= pLib.nVarsMax")
		// if there are some problems with this code, try this instead
		//        nRealVars = pLib->nVarsMax;

		// find the N-canonical form of this supergate
		pGate.nPhases = uint32(Map_CanonComputeSlow(pLib.uTruths, pLib.nVarsMax, nRealVars, pGate.uTruth[:], pGate.uPhases[:], uCanon[:]))
		// add the supergate into the table by its N-canonical table
		Map_SuperTableInsertC(pLib.tTableC, uCanon[:], pGate)
		/*
		           {
		               int uCanon1, uCanon2;
		               uCanon1 = uCanon[0];
		               pGate->uTruth[0] = ~pGate->uTruth[0];
		               pGate->uTruth[1] = ~pGate->uTruth[1];
		               Map_CanonComputeSlow( pLib->uTruths, pLib->nVarsMax, nRealVars, pGate->uTruth, pGate->uPhases, uCanon );
		               uCanon2 = uCanon[0];
		   Rwt_Man5ExploreCount( uCanon1 < uCanon2 ? uCanon1 : uCanon2 );
		           }
		*/
	}
	// sort the gates in each line
	Map_SuperTableSortSupergatesByDelay(pLib.tTableC, pLib.nSupersAll)

	// let the glory be manifest
	//    Map_LibraryPrintTree( pLib );
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Finds the largest PI number in the support of the supergate.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryGetMaxSuperPi_rec(pGate *Map_Super_t) int32 {
	var i, VarCur, VarMax int32
	if pGate.pRoot == nil {
		return pGate.Num
	}
	for i = 0; i < int32(pGate.nFanins); i++ {
		VarCur = Map_LibraryGetMaxSuperPi_rec(pGate.pFanins[i])
		if VarMax < VarCur {
			VarMax = VarCur
		}
	}
	return VarMax
}

/*
*Function*************************************************************

	Synopsis    [Finds the largest PI number in the support of the supergate.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryGetGateSupp_rec(pGate *Map_Super_t) uint32 {
	var uSupport uint32
	var i int32
	if pGate.pRoot == nil {
		return uint32((1 << (pGate.Num)))
	}
	uSupport = 0
	for i = 0; i < int32(pGate.nFanins); i++ {
		uSupport |= Map_LibraryGetGateSupp_rec(pGate.pFanins[i])
	}
	return uSupport
}

/*
*Function*************************************************************

	Synopsis    [Derives the pin-to-pin delay constraints for the supergate.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryAddFaninDelays(pLib *Map_SuperLib_t, pGate *Map_Super_t, pFanin *Map_Super_t, pPin *Mio_Pin_t) {
	var PinPhase Mio_PinPhase_t
	var tDelayBlockRise, tDelayBlockFall, tDelayPin float32
	var fMaxDelay int32
	var i int32

	// use this node to enable max-delay model
	if fMaxDelay != 0 {
		var tDelayBlockMax float32
		// get the maximum delay
		tDelayBlockMax = float32(Mio_PinReadDelayBlockMax(pPin))
		// go through the supergate inputs
		for i = 0; i < pLib.nVarsMax; i++ {
			if pFanin.tDelaysR[i].Rise < 0 {
				continue
			}
			tDelayPin = pFanin.tDelaysR[i].Rise + tDelayBlockMax
			if pGate.tDelaysR[i].Rise < tDelayPin {
				pGate.tDelaysR[i].Rise = tDelayPin
			}
		}
		// go through the supergate inputs
		for i = 0; i < pLib.nVarsMax; i++ {
			if pFanin.tDelaysF[i].Fall < 0 {
				continue
			}
			tDelayPin = pFanin.tDelaysF[i].Fall + tDelayBlockMax
			if pGate.tDelaysF[i].Fall < tDelayPin {
				pGate.tDelaysF[i].Fall = tDelayPin
			}
		}
		return
	}

	// get the interesting parameters of this pin
	PinPhase = Mio_PinReadPhase(pPin)
	tDelayBlockRise = float32(Mio_PinReadDelayBlockRise(pPin))
	tDelayBlockFall = float32(Mio_PinReadDelayBlockFall(pPin))

	// update the rise and fall of the output depending on the phase of the pin
	if PinPhase != MIO_PHASE_INV { // NONINV phase is present

		// the rise of the gate is determined by the rise of the fanin
		// the fall of the gate is determined by the fall of the fanin
		for i = 0; i < pLib.nVarsMax; i++ {
			////////////////////////////////////////////////////////
			// consider the rise of the gate
			////////////////////////////////////////////////////////
			// check two types of constraints on the rise of the fanin:
			// (1) the constraints related to the rise of the PIs
			// (2) the constraints related to the fall of the PIs
			if pFanin.tDelaysR[i].Rise >= 0 { // case (1)
				// fanin's rise depends on the rise of i-th PI
				// update the rise of the gate's output
				if pGate.tDelaysR[i].Rise < pFanin.tDelaysR[i].Rise+tDelayBlockRise {
					pGate.tDelaysR[i].Rise = pFanin.tDelaysR[i].Rise + tDelayBlockRise
				}
			}
			if pFanin.tDelaysR[i].Fall >= 0 { // case (2)
				// fanin's rise depends on the fall of i-th PI
				// update the rise of the gate's output
				if pGate.tDelaysR[i].Fall < pFanin.tDelaysR[i].Fall+tDelayBlockRise {
					pGate.tDelaysR[i].Fall = pFanin.tDelaysR[i].Fall + tDelayBlockRise
				}
			}
			////////////////////////////////////////////////////////

			////////////////////////////////////////////////////////
			// consider the fall of the gate (similar)
			////////////////////////////////////////////////////////
			// check two types of constraints on the fall of the fanin:
			// (1) the constraints related to the rise of the PIs
			// (2) the constraints related to the fall of the PIs
			if pFanin.tDelaysF[i].Rise >= 0 { // case (1)

				if pGate.tDelaysF[i].Rise < pFanin.tDelaysF[i].Rise+tDelayBlockFall {
					pGate.tDelaysF[i].Rise = pFanin.tDelaysF[i].Rise + tDelayBlockFall
				}
			}
			if pFanin.tDelaysF[i].Fall >= 0 { // case (2)

				if pGate.tDelaysF[i].Fall < pFanin.tDelaysF[i].Fall+tDelayBlockFall {
					pGate.tDelaysF[i].Fall = pFanin.tDelaysF[i].Fall + tDelayBlockFall
				}
			}
			////////////////////////////////////////////////////////
		}
	}
	if PinPhase != MIO_PHASE_NONINV { // INV phase is present

		// the rise of the gate is determined by the fall of the fanin
		// the fall of the gate is determined by the rise of the fanin
		for i = 0; i < pLib.nVarsMax; i++ {
			////////////////////////////////////////////////////////
			// consider the rise of the gate's output
			////////////////////////////////////////////////////////
			// check two types of constraints on the fall of the fanin:
			// (1) the constraints related to the rise of the PIs
			// (2) the constraints related to the fall of the PIs
			if pFanin.tDelaysF[i].Rise >= 0 { // case (1)
				// fanin's rise depends on the rise of i-th PI
				// update the rise of the gate
				if pGate.tDelaysR[i].Rise < pFanin.tDelaysF[i].Rise+tDelayBlockRise {
					pGate.tDelaysR[i].Rise = pFanin.tDelaysF[i].Rise + tDelayBlockRise
				}
			}
			if pFanin.tDelaysF[i].Fall >= 0 { // case (2)
				// fanin's rise depends on the fall of i-th PI
				// update the rise of the gate
				if pGate.tDelaysR[i].Fall < pFanin.tDelaysF[i].Fall+tDelayBlockRise {
					pGate.tDelaysR[i].Fall = pFanin.tDelaysF[i].Fall + tDelayBlockRise
				}
			}
			////////////////////////////////////////////////////////

			////////////////////////////////////////////////////////
			// consider the fall of the gate (similar)
			////////////////////////////////////////////////////////
			// check two types of constraints on the rise of the fanin:
			// (1) the constraints related to the rise of the PIs
			// (2) the constraints related to the fall of the PIs
			if pFanin.tDelaysR[i].Rise >= 0 { // case (1)

				if pGate.tDelaysF[i].Rise < pFanin.tDelaysR[i].Rise+tDelayBlockFall {
					pGate.tDelaysF[i].Rise = pFanin.tDelaysR[i].Rise + tDelayBlockFall
				}
			}
			if pFanin.tDelaysR[i].Fall >= 0 { // case (2)

				if pGate.tDelaysF[i].Fall < pFanin.tDelaysR[i].Fall+tDelayBlockFall {
					pGate.tDelaysF[i].Fall = pFanin.tDelaysR[i].Fall + tDelayBlockFall
				}
			}
			////////////////////////////////////////////////////////
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Performs phase transformation for one function.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CalculatePhase(uTruths [][2]uint32, nVars int32, uTruth uint32, uPhase uint32) uint32 {
	var v, Shift int32
	for v, Shift = 0, 1; v < nVars; v, Shift = v+1, Shift<<1 {
		if uPhase&uint32(Shift) != 0 {
			uTruth = (((uTruth & ^uTruths[v][0]) << Shift) | ((uTruth & uTruths[v][0]) >> Shift))
		}
	}
	return uTruth
}

/*
*Function*************************************************************

	Synopsis    [Performs phase transformation for one function.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CalculatePhase6(uTruths [][2]uint32, nVars int32, uTruth []uint32, uPhase uint32, uTruthRes []uint32) {
	var uTemp uint32
	var v, Shift int32

	// initialize the result
	uTruthRes[0] = uTruth[0]
	uTruthRes[1] = uTruth[1]
	if uPhase == 0 {
		return
	}
	// compute the phase
	for v, Shift = 0, 1; v < nVars; v, Shift = v+1, Shift<<1 {
		if uPhase&uint32(Shift) != 0 {
			if Shift < 32 {
				uTruthRes[0] = (((uTruthRes[0] & ^uTruths[v][0]) << Shift) | ((uTruthRes[0] & uTruths[v][0]) >> Shift))
				uTruthRes[1] = (((uTruthRes[1] & ^uTruths[v][1]) << Shift) | ((uTruthRes[1] & uTruths[v][1]) >> Shift))

			} else {
				uTemp = uTruthRes[0]
				uTruthRes[0] = uTruthRes[1]
				uTruthRes[1] = uTemp
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Prints the supergate library after deriving parameters.]

	Description [This procedure is very useful to see the library after
	it has been read into the mapper by "read_super" and all the information
	about the supergates derived.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_LibraryPrintTree(pLib *Map_SuperLib_t) {
	var pGate *Map_Super_t
	var i, k int32

	// print all the info related to the supergates
	//    for ( i = pLib->nVarsMax; i < (int)pLib->nLines; i++ )
	for i = pLib.nVarsMax; i < 20; i++ {
		pGate = pLib.ppSupers[i]

		// write the gate's fanin info and formula
		fmt.Printf("%6d  ", pGate.Num)
		marker := ' '
		if pGate.fSuper != 0 {
			marker = '*'
		}
		fmt.Printf("%c ", marker)
		fmt.Printf("%6s", Mio_GateReadName(pGate.pRoot))
		for k = 0; k < int32(pGate.nFanins); k++ {
			fmt.Printf(" %6d", pGate.pFanins[k].Num)
		}
		fmt.Printf("  %s", pGate.pFormula)
		fmt.Printf("\n")

		// write the gate's derived info
		Extra_PrintBinary(os.Stdout, pGate.uTruth[:], 64)
		fmt.Printf("  %3d", pGate.nGates)
		fmt.Printf("  %6.2f", pGate.Area)
		fmt.Printf("  (%4.2f, %4.2f)", pGate.tDelayMax.Rise, pGate.tDelayMax.Fall)
		fmt.Printf("\n")
		for k = 0; k < pLib.nVarsMax; k++ {
			// print the constraint on the rise of the gate in the form (D1, D2),
			// where D1 is the constraint related to the rise of the k-th PI
			// where D2 is the constraint related to the fall of the k-th PI
			if pGate.tDelaysR[k].Rise < 0 && pGate.tDelaysR[k].Fall < 0 {
				fmt.Printf(" (----, ----)")
			} else if pGate.tDelaysR[k].Fall < 0 {
				fmt.Printf(" (%4.2f, ----)", pGate.tDelaysR[k].Rise)
			} else if pGate.tDelaysR[k].Rise < 0 {
				fmt.Printf(" (----, %4.2f)", pGate.tDelaysR[k].Fall)
			} else {
				fmt.Printf(" (%4.2f, %4.2f)", pGate.tDelaysR[k].Rise, pGate.tDelaysR[k].Fall)
			}

			// print the constraint on the fall of the gate in the form (D1, D2),
			// where D1 is the constraint related to the rise of the k-th PI
			// where D2 is the constraint related to the fall of the k-th PI
			if pGate.tDelaysF[k].Rise < 0 && pGate.tDelaysF[k].Fall < 0 {
				fmt.Printf(" (----, ----)")
			} else if pGate.tDelaysF[k].Fall < 0 {
				fmt.Printf(" (%4.2f, ----)", pGate.tDelaysF[k].Rise)
			} else if pGate.tDelaysF[k].Rise < 0 {
				fmt.Printf(" (----, %4.2f)", pGate.tDelaysF[k].Fall)
			} else {
				fmt.Printf(" (%4.2f, %4.2f)", pGate.tDelaysF[k].Rise, pGate.tDelaysF[k].Fall)
			}
			fmt.Printf("\n")
		}
		fmt.Printf("\n")
	}
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
