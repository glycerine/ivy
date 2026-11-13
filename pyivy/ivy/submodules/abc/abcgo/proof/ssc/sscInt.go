package ssc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
	"math/rand"
)

// Ssc_Man_t owns all mutable simulation, class, and SAT state. It never borrows
// GIA scratch fields and separate managers can operate independently.
type Ssc_Man_t struct {
	Params         Ssc_Pars_t
	Aig, Care      *gia.Gia_Man_t
	Original       *gia.Gia_Man_t
	fraig          *builder
	mapping        []int32
	Sim            *Simulation
	Classes        *Classes
	Pivot, Pattern []bool
	sat            *bsat.Solver
	satVars        [2][]int32
	satInputs      []int32
	satNext        int32
	inconsistent   bool
	Stats          Statistics
}
type Statistics struct{ Calls, Proofs, Counterexamples, Undecided, Recycles, SimulationRounds int }
type Simulation struct {
	Graph           *gia.Gia_Man_t
	Inputs, Values  [][]uint64
	Words, Patterns int
	random          *rand.Rand
}
type Classes struct {
	Groups [][]int32
	Repr   []int32
	Phase  []bool
}
