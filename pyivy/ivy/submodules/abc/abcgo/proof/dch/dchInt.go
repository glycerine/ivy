package dch

import (
	"time"

	aig "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

type Statistics struct {
	Candidates, Proofs, SATCalls, Counterexamples, Unknown, Recycles, Blocked int
	Simulation, Solving, Total                                                time.Duration
	// Solving split by outcome (timeSatSat, timeSatUnsat, timeSatUndec)
	SolvingSat, SolvingUnsat, SolvingUndecided time.Duration
}
type Manager struct {
	graph      *aig.Graph
	satGraph   *aig.Graph
	mapped     []aig.Edge
	params     Params
	order      []aig.Edge
	classes    *Classes
	solver     *bsat.Solver
	variables  map[int]int32
	nextVar    int32
	callsSince int
	coneMax    int // nConeMax: the largest solved cone resimulated
	proven     map[int]aig.Edge
	blocked    map[int]bool
	fanouts    map[int][]int
	ranks      map[int]int
	lastValues []bool
	// scratch for refineCounterexample, reused across counterexamples
	cexVisited []bool
	inputIndex []int32
	// generation-stamped node sets for the counterexample's TFO
	tfoGen                             uint32
	tfoSeen, tfoCandidates, tfoClasses []uint32
	random                             randomSource
	stats                              Statistics
}
type Dch_Man_t = Manager
