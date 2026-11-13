package dch

import (
	"time"

	aig "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

type Statistics struct {
	Candidates, Proofs, SATCalls, Counterexamples, Unknown, Recycles, Blocked int
	Simulation, Solving, Total                                                time.Duration
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
	proven     map[int]aig.Edge
	blocked    map[int]bool
	fanouts    map[int][]int
	ranks      map[int]int
	lastValues []bool
	stats      Statistics
}
type Dch_Man_t = Manager
