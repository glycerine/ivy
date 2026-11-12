package csw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"iter"
)

// Cut records have stable addresses. pNext links hash buckets; nextRecord
// replaces the C layout's byte stride between adjacent records of a node.
type Csw_Cut_t struct {
	pNext      *Csw_Cut_t
	Cost       int32
	uSign      uint32
	iNode      int32
	pFanins    []int32
	truth      []uint32
	nextRecord *Csw_Cut_t
}
type Csw_Man_t struct {
	pManAig, pManRes                                               *aig.Graph
	pEquiv                                                         []aig.Edge
	pCuts                                                          []*Csw_Cut_t
	pnRefs                                                         []int32
	pTable                                                         []*Csw_Cut_t
	nCutsMax, nLeafMax                                             int32
	fVerbose                                                       int32
	puTemp                                                         [4][]uint32
	nNodesTriv0, nNodesTriv1, nNodesTriv2, nNodesCuts, nNodesTried int32
	timeCuts, timeHash, timeOther, timeTotal                       int64
}

func Csw_CutLeaveNum(c *Csw_Cut_t) int32  { return int32(len(c.pFanins)) }
func Csw_CutLeaves(c *Csw_Cut_t) []int32  { return c.pFanins }
func Csw_CutTruth(c *Csw_Cut_t) []uint32  { return c.truth }
func Csw_CutNext(c *Csw_Cut_t) *Csw_Cut_t { return c.nextRecord }
func (p *Csw_Man_t) ensureObject(id int) {
	for len(p.pCuts) <= id {
		p.pCuts = append(p.pCuts, nil)
		p.pnRefs = append(p.pnRefs, 0)
	}
}
func Csw_ObjRefs(p *Csw_Man_t, obj aig.Edge) int32 {
	p.ensureObject(obj.NodeID())
	return p.pnRefs[obj.NodeID()]
}
func Csw_ObjAddRefs(p *Csw_Man_t, obj aig.Edge, n int32) {
	p.ensureObject(obj.NodeID())
	p.pnRefs[obj.NodeID()] += n
}
func Csw_ObjCuts(p *Csw_Man_t, obj aig.Edge) *Csw_Cut_t {
	p.ensureObject(obj.NodeID())
	return p.pCuts[obj.NodeID()]
}
func Csw_ObjSetCuts(p *Csw_Man_t, obj aig.Edge, c *Csw_Cut_t) {
	p.ensureObject(obj.NodeID())
	p.pCuts[obj.NodeID()] = c
}
func Csw_ObjEquiv(p *Csw_Man_t, obj aig.Edge) aig.Edge  { return p.pEquiv[obj.NodeID()] }
func Csw_ObjSetEquiv(p *Csw_Man_t, obj, value aig.Edge) { p.pEquiv[obj.NodeID()] = value }
func cswChildEquiv(p *Csw_Man_t, child aig.Edge) aig.Edge {
	r := Csw_ObjEquiv(p, child)
	if r == aig.InvalidEdge {
		panic("csw: fanin has not been swept")
	}
	if child.Complemented() {
		r = r.Not()
	}
	return r
}
func Csw_ObjChild0Equiv(p *Csw_Man_t, obj aig.Edge) aig.Edge {
	return cswChildEquiv(p, p.pManAig.Node(obj).Fanins[0])
}
func Csw_ObjChild1Equiv(p *Csw_Man_t, obj aig.Edge) aig.Edge {
	return cswChildEquiv(p, p.pManAig.Node(obj).Fanins[1])
}
func Csw_ObjForEachCut(p *Csw_Man_t, obj aig.Edge) iter.Seq2[int, *Csw_Cut_t] {
	return func(yield func(int, *Csw_Cut_t) bool) {
		c := Csw_ObjCuts(p, obj)
		for i := 0; i < int(p.nCutsMax) && c != nil; i++ {
			if !yield(i, c) {
				return
			}
			c = c.nextRecord
		}
	}
}
func Csw_CutForEachLeaf(graph *aig.Graph, c *Csw_Cut_t) iter.Seq2[int, aig.Edge] {
	return func(yield func(int, aig.Edge) bool) {
		for i, id := range c.pFanins {
			e := aig.Edge(id) << 1
			if !graph.ValidEdge(e) {
				panic("csw: cut leaf outside graph")
			}
			if !yield(i, e) {
				return
			}
		}
	}
}
