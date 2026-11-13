package cec

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/glucose2"
)

type Cec4_Man_t struct {
	pPars      *Cec_ParFra_t // parameters
	pAig       *Gia_Man_t    // user's AIG
	pNew       *Gia_Man_t    // internal AIG
	simState   *cec4SimulationState
	cnf        *cec4CNF
	packer     *cec4PatternPacker
	candidates *cec4Candidates
	copies     []int32
	pattern    int
	// SAT solving
	pSat          *glucose2.SimpSolver // SAT solver
	vFrontier     *Vec_Ptr_t           // CNF construction
	vFanins       *Vec_Ptr_t           // CNF construction
	vCexMin       *Vec_Int_t           // minimized CEX
	vClassUpdates *Vec_Int_t           // updated equiv classes
	vCexStamps    *Vec_Int_t           // time stamps
	vCands        *Vec_Int_t
	vVisit        *Vec_Int_t
	vPat          *Vec_Int_t
	vDisprPairs   *Vec_Int_t
	vFails        *vec.Vec_Bit_t
	iPosRead      int32 // candidate reading position
	iPosWrite     int32 // candidate writing position
	iLastConst    int32 // last const node proved
	// refinement
	vRefClasses *Vec_Int_t
	vRefNodes   *Vec_Int_t
	vRefBins    *Vec_Int_t
	pTable      *int32
	nTableSize  int32
	// statistics
	nItersSim     int32
	nItersSat     int32
	nAndNodes     int32
	nPatterns     int32
	nSatSat       int32
	nSatUnsat     int32
	nSatUndec     int32
	nCallsSince   int32
	nSimulates    int32
	nRecycles     int32
	nConflicts    [2][3]int32
	nGates        [2]int32
	nFaster       [2]int32
	timeCnf       abctime
	timeGenPats   abctime
	timeSatSat0   abctime
	timeSatUnsat0 abctime
	timeSatSat    abctime
	timeSatUnsat  abctime
	timeSatUndec  abctime
	timeSim       abctime
	timeRefine    abctime
	timeResimGlo  abctime
	timeResimLoc  abctime
	timeStart     abctime
}
