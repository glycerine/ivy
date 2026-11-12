// Package rwt implements ABC's four-input AIG rewriting library.
package rwt

import (
	"sync"
	"time"

	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/vec"
)

const RWT_LIMIT = 1048576 / 4

type Rwt_Man_t = Rwt_Man_t_
type Rwt_Node_t = Rwt_Node_t_

// Rwt_Node_t_ is a rewrite-forest node. Complemented edges are represented
// by a lightweight companion handle rather than tagged machine pointers.
type Rwt_Node_t_ struct {
	Id, TravId int32
	uTruth     uint32
	Volume     uint32
	Level      uint32
	fUsed      bool
	fExor      bool
	p0, p1     *Rwt_Node_t
	pNext      *Rwt_Node_t

	regular     *Rwt_Node_t
	neg         *Rwt_Node_t
	fComplement bool
}

type Rwt_Man_t_ struct {
	nFuncs              int32
	puCanons            []uint16
	pPhases, pPerms     []byte
	pMap                []byte
	pMapInv             []uint16
	pPractical          []byte
	pPerms4             [][]byte
	vForest             []*Rwt_Node_t
	pTable              []*Rwt_Node_t
	vClasses            [][]*Rwt_Node_t
	nodeGraphs          map[*Rwt_Node_t]*dec.Dec_Graph_t
	nTravIds            int32
	nConsidered         int32
	nAdded              int32
	nClasses            int32
	fCompl              int32
	pCut                any
	pGraph              *dec.Dec_Graph_t
	pPerm               []byte
	vFanins             *vec.Vec_Ptr_t
	vFaninsCur          *vec.Vec_Ptr_t
	vLevNums            *vec.Vec_Int_t
	vNodesTemp          *vec.Vec_Ptr_t
	nNodesConsidered    int32
	nNodesRewritten     int32
	nNodesGained        int32
	nScores             [222]int32
	nCutsGood, nCutsBad int32
	nSubgraphs          int32
	timeStart           time.Duration
	timeTruth           time.Duration
	timeCut             time.Duration
	timeRes             time.Duration
	timeEval            time.Duration
	timeMffc            time.Duration
	timeUpdate          time.Duration
	timeTotal           time.Duration
}

var rwtGlobal struct {
	sync.Mutex
	canons  []uint16
	phases  []byte
	perms   []byte
	mapping []byte
}

func Rwt_IsComplement(p *Rwt_Node_t) int32 {
	if p != nil && p.fComplement {
		return 1
	}
	return 0
}

func Rwt_Regular(p *Rwt_Node_t) *Rwt_Node_t {
	if p != nil && p.fComplement {
		return p.regular
	}
	return p
}

func Rwt_Not(p *Rwt_Node_t) *Rwt_Node_t {
	if p == nil {
		return nil
	}
	return p.neg
}

func Rwt_NotCond(p *Rwt_Node_t, complement int32) *Rwt_Node_t {
	if complement&1 != 0 {
		return Rwt_Not(p)
	}
	return p
}

// Rwt_ManCanonInfo returns the class, phase, and variable permutation for a
// four-variable truth table. The permutation is a manager-owned immutable
// slice; callers should treat it as read-only.
func Rwt_ManCanonInfo(p *Rwt_Man_t, truth uint16) (class, phase byte, permutation []byte) {
	if p == nil || int(truth) >= len(p.pMap) {
		return 0, 0, nil
	}
	index := p.pPerms[truth]
	return p.pMap[truth], p.pPhases[truth], p.pPerms4[index]
}

func Rwt_ManReadDecs(p *Rwt_Man_t) *dec.Dec_Graph_t {
	if p == nil {
		return nil
	}
	return p.pGraph
}

func Rwt_ManReadLeaves(p *Rwt_Man_t) *vec.Vec_Ptr_t {
	if p == nil {
		return nil
	}
	return p.vFanins
}

func Rwt_ManReadCurrentLeaves(p *Rwt_Man_t) *vec.Vec_Ptr_t {
	if p == nil {
		return nil
	}
	return p.vFaninsCur
}

func Rwt_ManClassSubgraphs(p *Rwt_Man_t, class int32) []*Rwt_Node_t {
	if p == nil || class < 0 || int(class) >= len(p.vClasses) {
		return nil
	}
	return p.vClasses[class]
}

func Rwt_ManNodeDecomposition(p *Rwt_Man_t, node *Rwt_Node_t) *dec.Dec_Graph_t {
	if p == nil || node == nil {
		return nil
	}
	return p.nodeGraphs[Rwt_Regular(node)]
}

func Rwt_ManRecordNodeConsidered(p *Rwt_Man_t) {
	if p != nil {
		p.nNodesConsidered++
	}
}

func Rwt_ManRecordCut(p *Rwt_Man_t, good bool) {
	if p == nil {
		return
	}
	if good {
		p.nCutsGood++
	} else {
		p.nCutsBad++
	}
}

func Rwt_ManAddTimeTruth(p *Rwt_Man_t, elapsed time.Duration) {
	if p != nil {
		p.timeTruth += elapsed
	}
}

func Rwt_ManAddTimeMffc(p *Rwt_Man_t, elapsed time.Duration) {
	if p != nil {
		p.timeMffc += elapsed
	}
}

func Rwt_ManAddTimeEval(p *Rwt_Man_t, elapsed time.Duration) {
	if p != nil {
		p.timeEval += elapsed
	}
}

func Rwt_ManAddTimeResynthesis(p *Rwt_Man_t, elapsed time.Duration) {
	if p != nil {
		p.timeRes += elapsed
	}
}

func Rwt_ManAddSubgraphs(p *Rwt_Man_t, count int32) {
	if p != nil {
		p.nSubgraphs += count
	}
}

// Rwt_ManStoreBest records the winning cut/decomposition selected by Ivy.
func Rwt_ManStoreBest(p *Rwt_Man_t, graph *dec.Dec_Graph_t, complemented bool, fanins []any, truth uint16, gain int32, useZeroCost bool) {
	if p == nil || graph == nil {
		return
	}
	p.pGraph = graph
	p.fCompl = 0
	if complemented {
		p.fCompl = 1
	}
	vec.Vec_PtrClear(p.vFanins)
	for _, fanin := range fanins {
		vec.Vec_PtrPush(p.vFanins, fanin)
	}
	class := p.pMap[truth]
	p.nScores[class]++
	p.nNodesGained += gain
	if useZeroCost || gain > 0 {
		p.nNodesRewritten++
	}
}

func Rwt_ManReadCompl(p *Rwt_Man_t) int32 {
	if p == nil {
		return 0
	}
	return p.fCompl
}
