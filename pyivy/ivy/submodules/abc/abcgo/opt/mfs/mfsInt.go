package mfs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/satoko"
)

type Stats struct{ NodesTried, Removes, Resubs, TryRemoves, TryResubs, SATCalls, Counterexamples, Timeouts, WindowsSkipped, Decomposed, NodesGained, EdgesBefore, EdgesAfter, NodesBefore, NodesAfter int }
type Mfs_Man_t struct {
	Params                          Mfs_Par_t
	Network                         *Network
	Target                          int
	Roots, Support, Nodes, Divisors []int
	Window                          *aig.Graph
	Values                          map[int]aig.Edge
	CareOutputs                     int
	Solver                          *satoko.Solver
	projection                      []int
	clauses                         [][]int
	vars                            int
	candidates                      []int
	witnesses                       [][]bool
	Care                            []uint32
	Stats                           Stats
	probabilities                   []float32
	required, levels                []int
	fanouts                         [][]int
	outputDriver                    []bool
}
