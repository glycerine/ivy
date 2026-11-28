// Package sfm implements SAT-based logic resubstitution and cell mapping.
package sfm

import (
	"time"

	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/sat/bsat"
)

const SFM_FANIN_MAX = 6
const SFM_SUPP_MAX = 8
const SFM_WIN_MAX = 1000
const SFM_DEC_MAX = 4
const SFM_SIM_WORDS = 8
const SFM_SAT_UNDEC uint64 = 0x1234567812345678
const SFM_SAT_SAT uint64 = 0x8765432187654321

type Sfm_Par_t struct {
	TfoLevMax, TfiLevMax, FanoutMax, DepthMax, VarMax, MffcMin, MffcMax, DecMax, WinSizeMax, GrowthLevel                                       int
	BTLimit                                                                                                                                    int64
	NodesMax, NodeOne, FirstFixed, TimeWin, DeltaCrit, DelAreaRatio                                                                            int
	RrOnly, Area, AreaRev, MoreEffort, UseAndOr, ZeroCost, UseSim, UseDcs, PrintDecs, AllBoxes, LibVerbose, DelayVerbose, Verbose, VeryVerbose bool
}

func Sfm_ParSetDefault(p *Sfm_Par_t) {
	*p = Sfm_Par_t{TfoLevMax: 2, FanoutMax: 30, DepthMax: 20, WinSizeMax: 300, BTLimit: 5000}
}
func DefaultParams() Sfm_Par_t { var p Sfm_Par_t; Sfm_ParSetDefault(&p); return p }

type Stats struct{ NodesTried, Removes, Resubs, Improves, TryRemoves, TryResubs, TryImproves, SATCalls, Timeouts, MaxDivs, TotalDivs, NodesBefore, NodesAfter, EdgesBefore, EdgesAfter int }

// Sfm_Ntk_t retains source object numbering: PIs first, POs last. Truth rows
// contain expanded 64-bit truth words; fanin zero is the least significant bit.
// Fixed objects cannot be rewritten; Empty objects add no logic level.
// Gate is a cached native library cell. Area and pin timing use source integer
// units; Cell retains the original genlib handle for application adapters.
type Gate struct {
	Cell   *mio.Mio_Gate_t
	Name   string
	Inputs int
	Truth  []uint64
	Area   int
	Pins   []Pin
}
type Pin struct {
	Phase      int
	Rise, Fall int
	Load       float64
}

// Sfm_Ntk_t is the port of C's struct Sfm_Ntk_t_ (sfmInt.h). The C engine
// (Sfm_NtkConstruct/Sfm_NtkPerform, used by "&mfs" and "mfs2") uses the
// lower-case fields, which mirror the C fields one-to-one, plus the exported
// fields Pis, Pos, Fanins (vFanins), Fanouts (vFanouts), Levels (vLevels),
// Reverse (vLevelsR), Fixed (vFixed), Empty (vEmpty) and LevelMax (nLevelMax).
//
// The Go-native mapped-network optimizer (sfm_dec_*.go, sfmArea.go) shares
// this type and additionally uses Truths, Deleted, Mapping and SourceIDs; see
// sfm_mapped_ntk.go. Those fields are not used by the C engine.
type Sfm_Ntk_t struct {
	// parameters
	pPars  *Sfm_Par_t // parameters
	Params Sfm_Par_t  // storage for pPars when the caller passes none
	// objects
	Pis      int // PI count (PIs should be first objects)
	Pos      int // PO count (POs should be last objects)
	nNodes   int // internal nodes
	nObjs    int // total objects
	LevelMax int // maximum level
	// user data
	Fixed    []bool   // persistent objects
	Empty    []bool   // transparent objects (nil: none)
	vTruths  []uint64 // truth tables
	Fanins   [][]int  // fanins
	vStarts  []int    // offsets
	vTruths2 []uint64 // truth tables (large)
	// attributes
	Fanouts [][]int  // fanouts
	Levels  []int    // logic level
	Reverse []int    // logic level (reverse)
	vCounts []int    // fanin counters
	vId2Var []int32  // ObjId -> SatVar
	vVar2Id []int    // SatVar -> ObjId
	vCnfs   [][]int8 // CNFs
	vCover  []uint32 // temporary
	// traversal IDs
	vTravIds  []int // traversal IDs
	vTravIds2 []int // traversal IDs
	nTravIds  int   // traversal IDs
	nTravIds2 int   // traversal IDs
	// window
	iPivotNode int   // window pivot
	vNodes     []int // internal
	vDivs      []int // divisors
	vRoots     []int // roots
	vTfo       []int // TFO (excluding iNode)
	// SAT solving
	pSat     *bsat.Solver // SAT solver
	nSatVars int32        // the number of variables
	// counter-examples
	nCexes    int      // number of CEXes
	vDivCexes []uint64 // counter-examples
	// intermediate data
	vOrder    []int     // object order
	vDivVars  []int32   // divisor SAT variables
	vDivIds   []int32   // divisors indexes
	vLits     []int32   // literals
	vValues   []int32   // SAT variable values
	vClauses  [][]int32 // CNF clauses for the node
	vFaninMap []int32   // mapping fanins into their SAT vars
	// statistics (the C counters nTryRemoves, nRemoves, nTotalNodesBeg, ...)
	Stats Stats
	// runtime
	timeWin, timeDiv, timeCnf, timeSat, timeOther, timeTotal time.Duration

	// Fields used only by the Go-native mapped-network optimizer.
	Truths  [][]uint64
	Deleted []bool
	// Mapping associates each internal object with a library gate (nil for LUTs).
	// Mapped algorithms share the same object IDs and fanin ordering.
	Mapping []*Gate
	// SourceIDs holds caller object IDs for the original mapped objects.
	SourceIDs []int
}

// Sfm_ParSetDefault3 is sfmDec.c's parameter setup for mfs3.
func Sfm_ParSetDefault3(pPars *Sfm_Par_t) {
	*pPars = Sfm_Par_t{}
	pPars.TfoLevMax = 100  // the maximum fanout levels
	pPars.TfiLevMax = 100  // the maximum fanin levels
	pPars.FanoutMax = 10   // the maximum number of fanouts
	pPars.MffcMin = 1      // the maximum MFFC size
	pPars.MffcMax = 3      // the maximum MFFC size
	pPars.VarMax = 6       // the maximum variable count
	pPars.DecMax = 1       // the maximum number of decompositions
	pPars.WinSizeMax = 0   // the maximum window size
	pPars.GrowthLevel = 0  // the maximum allowed growth in level
	pPars.BTLimit = 0      // the maximum number of conflicts in one SAT run
	pPars.TimeWin = 1      // the size of timing window in percents
	pPars.DeltaCrit = 0    // delta delay in picoseconds
	pPars.UseAndOr = false // enable internal detection of AND/OR gates
	pPars.ZeroCost = false // enable zero-cost replacement
	pPars.MoreEffort = false
	pPars.UseSim = false
	pPars.Area = false
	pPars.Verbose = false
	pPars.VeryVerbose = false
}
