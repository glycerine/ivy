// Package sfm implements SAT-based logic resubstitution and cell mapping.
package sfm

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/sat/satoko"
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

type Sfm_Ntk_t struct {
	Pis, Pos              int
	Fanins, Fanouts       [][]int
	Truths                [][]uint64
	Fixed, Empty, Deleted []bool
	Levels, Reverse       []int
	LevelMax              int
	Params                Sfm_Par_t
	Stats                 Stats
	// Mapping associates each internal object with a library gate (nil for LUTs).
	// Mapped algorithms share the same object IDs and fanin ordering.
	Mapping                        []*Gate
	pivot                          int
	nodes, divs, roots, tfo, order []int
	solver                         *satoko.Solver
	idVar                          map[int]int
	divIDs                         []int
	witnesses                      [][]bool
	cnfs                           [][][]int
}
