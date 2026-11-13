// Package rwr implements ABC's DAG-aware rewriting library.
package rwr

import "github.com/glycerine/abcgo/bool/dec"

const RWR_LIMIT = 1048576 / 4

// Rwr_Man_t stores source lookup tables, library nodes, selected rewriting
// graph, counters and timings. Go slices own the memory held by C vectors.
type Rwr_Man_t struct {
	nFuncs                                                                 int32
	puCanons                                                               []uint16
	pPhases, pPerms, pMap                                                  []byte
	pMapInv                                                                []uint16
	pPractical                                                             []bool
	pPerms4                                                                [][]byte
	vForest                                                                []*Rwr_Node_t
	pTable                                                                 []*Rwr_Node_t
	vClasses                                                               [][]*Rwr_Node_t
	nTravIds, nConsidered, nAdded, nClasses                                int32
	fCompl                                                                 int32
	pGraph                                                                 *dec.Dec_Graph_t
	vFanins, vFaninsCur                                                    []any
	vLevNums                                                               []int32
	vNodesTemp                                                             []any
	nNodesConsidered, nNodesRewritten, nNodesGained, nNodesBeg, nNodesEnd  int32
	nScores                                                                [222]int32
	nCutsGood, nCutsBad, nSubgraphs                                        int32
	timeStart, timeCut, timeRes, timeEval, timeMffc, timeUpdate, timeTotal int64
}

// Rwr_NodeRef carries the bit C stores in the low bit of a node pointer.
type Rwr_NodeRef struct {
	node         *Rwr_Node_t
	complemented bool
}
type Rwr_Node_t struct {
	Id, TravId            int32
	nScore, nGain, nAdded int16
	uTruth                uint16
	Volume, Level         uint32
	fUsed, fExor          bool
	p0, p1                Rwr_NodeRef
	pNext                 *Rwr_Node_t
	graph                 *dec.Dec_Graph_t
}
type Rwr_Man_t_ = Rwr_Man_t
type Rwr_Node_t_ = Rwr_Node_t

func Rwr_IsComplement(p Rwr_NodeRef) int32 {
	if p.complemented {
		return 1
	}
	return 0
}
func Rwr_Regular(p Rwr_NodeRef) *Rwr_Node_t { return p.node }
func Rwr_Not(p Rwr_NodeRef) Rwr_NodeRef     { p.complemented = !p.complemented; return p }
func Rwr_NotCond(p Rwr_NodeRef, c int32) Rwr_NodeRef {
	if c != 0 {
		return Rwr_Not(p)
	}
	return p
}
