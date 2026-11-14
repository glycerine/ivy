package aig

/**CFile****************************************************************

  FileName    [aig.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aig.h,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/util/utilCex.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Aig_Man_t = Graph
type Aig_Obj_t = Edge

// object types
type Aig_Type_t int32

const (
	AIG_OBJ_NONE Aig_Type_t = iota
	// 0: non-existent object
	AIG_OBJ_CONST1
	// 1: constant 1
	AIG_OBJ_CI
	// 2: combinational input
	AIG_OBJ_CO
	// 3: combinational output
	AIG_OBJ_BUF
	// 4: buffer node
	AIG_OBJ_AND
	// 5: AND node
	AIG_OBJ_EXOR
	// 6: EXOR node
	AIG_OBJ_VOID // 7: unused object

)

// The Go representation stores regular nodes in Graph and inversion in Edge.
// Object accessors which inspect node payload receive the owning manager.

// cut computation
type Aig_ManCut_t = Aig_ManCut_t_

// the cut used to represent node in the AIG
type Aig_Cut_t struct {
	valid    bool
	truth    []uint32
	pNext    *Aig_Cut_t // the next cut in the table
	Cost     int32      // the cost of the cut
	uSign    uint32     // cut signature
	iNode    int32      // the node, for which it is the cut
	nCutSize int16      // the number of bytes in the cut
	nLeafMax byte       // the maximum number of fanins
	nFanins  byte       // the current number of fanins
	pFanins  []int32    // the fanins (followed by the truth table)
}

// the CNF computation manager
type Aig_ManCut_t_ struct {
	// AIG manager
	pAig  *Aig_Man_t   // the input AIG manager
	pCuts []*Aig_Cut_t // the cuts for each node in the output manager
	// parameters
	nCutsMax int32 // the max number of cuts at the node
	nLeafMax int32 // the max number of leaves of a cut
	fTruth   int32 // enables truth table computation
	fVerbose int32 // enables verbose output
	// internal variables
	nCutSize    int32 // the number of bytes needed to store one cut
	nTruthWords int32 // the number of truth table words
}

func Aig_ObjCuts(p *Aig_ManCut_t, e Edge) *Aig_Cut_t       { return p.pCuts[e.NodeID()] }
func Aig_ObjSetCuts(p *Aig_ManCut_t, e Edge, c *Aig_Cut_t) { p.pCuts[e.NodeID()] = c }

func Aig_CutLeaveNum(c *Aig_Cut_t) int32  { return int32(len(c.pFanins)) }
func Aig_CutLeaves(c *Aig_Cut_t) []int32  { return c.pFanins }
func Aig_CutTruth(c *Aig_Cut_t) []uint32  { return c.truth }
func Aig_CutNext(c *Aig_Cut_t) *Aig_Cut_t { return c.pNext }

// iterator over cuts of the node
// Original macro:
// #define Aig_ObjForEachCut( p, pObj, pCut, i )                           \
//     for ( i = 0, pCut = Aig_ObjCuts(p, pObj); i < p->nCutsMax; i++, pCut = Aig_CutNext(pCut) )

// iterator over leaves of the cut
// Original macro:
// #define Aig_CutForEachLeaf( p, pCut, pLeaf, i )                         \
//     for ( i = 0; (i < (int)(pCut)->nFanins) && ((pLeaf) = Aig_ManObj(p, (pCut)->pFanins[i])); i++ )

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Aig_ObjCutSign(ObjId uint32) uint32 { return (1 << (ObjId & 31)) }
func Aig_WordCountOnes(uWord uint32) int32 {
	uWord = (uWord & 0x55555555) + ((uWord >> 1) & 0x55555555)
	uWord = (uWord & 0x33333333) + ((uWord >> 2) & 0x33333333)
	uWord = (uWord & 0x0F0F0F0F) + ((uWord >> 4) & 0x0F0F0F0F)
	uWord = (uWord & 0x00FF00FF) + ((uWord >> 8) & 0x00FF00FF)
	return int32((uWord & 0x0000FFFF) + (uWord >> 16))
}
func Aig_WordFindFirstBit(word uint32) int32 {
	for i := int32(0); i < 32; i++ {
		if word&(1<<uint(i)) != 0 {
			return i
		}
	}
	return -1
}

func Aig_Regular(e Edge) Edge {
	if e == InvalidEdge {
		return e
	}
	return e &^ 1
}
func Aig_Not(e Edge) Edge { return e.Not() }
func Aig_NotCond(e Edge, c int32) Edge {
	if c != 0 {
		return e.Not()
	}
	return e
}
func Aig_IsComplement(e Edge) int32 { return int32(e & 1) }

func Aig_ManCiNum(p *Graph) int32 { return int32(len(p.inputs)) }
func Aig_ManCoNum(p *Graph) int32 { return int32(len(p.outputs)) }
func Aig_ManBufNum(p *Graph) int32 {
	var n int32
	for id, node := range p.nodes {
		if p.deleted[id] {
			continue
		}
		if node.Kind == Buffer {
			n++
		}
	}
	return n
}
func Aig_ManAndNum(p *Graph) int32 {
	var n int32
	for id, node := range p.nodes {
		if p.deleted[id] {
			continue
		}
		if node.Kind == And {
			n++
		}
	}
	return n
}
func Aig_ManExorNum(p *Graph) int32 {
	var n int32
	for id, node := range p.nodes {
		if p.deleted[id] {
			continue
		}
		if node.Kind == ExclusiveOr {
			n++
		}
	}
	return n
}
func Aig_ManNodeNum(p *Graph) int32   { return Aig_ManAndNum(p) + Aig_ManExorNum(p) }
func Aig_ManGetCost(p *Graph) int32   { return Aig_ManAndNum(p) + 3*Aig_ManExorNum(p) }
func Aig_ManObjNum(p *Graph) int32    { return int32(p.NodeCount()) }
func Aig_ManObjNumMax(p *Graph) int32 { return int32(len(p.nodes)) }
func Aig_ManRegNum(p *Graph) int32    { return p.nRegs }
func Aig_ManConstrNum(p *Graph) int32 { return p.nConstrs }

func Aig_ManConst0(p *Graph) Edge      { return p.Const0() }
func Aig_ManConst1(p *Graph) Edge      { return p.Const1() }
func Aig_ManGhost(p *Graph) *Node      { return &p.ghost }
func Aig_ManCi(p *Graph, i int32) Edge { return p.inputs[i] }
func Aig_ManCo(p *Graph, i int32) Edge { return p.outputs[i] }
func Aig_ManLo(p *Graph, i int32) Edge { return p.inputs[int32(len(p.inputs))-p.nRegs+i] }
func Aig_ManLi(p *Graph, i int32) Edge { return p.outputs[int32(len(p.outputs))-p.nRegs+i] }
func Aig_ManObj(p *Graph, i int32) Edge {
	if i < 0 || int(i) >= len(p.nodes) {
		panic("aig: invalid object index")
	}
	return Edge(i) << 1
}

func Aig_ObjType(p *Graph, e Edge) Aig_Type_t {
	switch p.node(e).Kind {
	case Constant:
		return AIG_OBJ_CONST1
	case Input:
		return AIG_OBJ_CI
	case And:
		return AIG_OBJ_AND
	case Buffer:
		return AIG_OBJ_BUF
	case ExclusiveOr:
		return AIG_OBJ_EXOR
	case Output:
		return AIG_OBJ_CO
	}
	return AIG_OBJ_NONE
}
func Aig_ObjIsNone(p *Graph, e Edge) bool   { return !p.valid(e) }
func Aig_ObjIsConst1(p *Graph, e Edge) bool { return e == p.Const1() }
func Aig_ObjIsCi(p *Graph, e Edge) bool     { return p.node(e).Kind == Input }
func Aig_ObjIsCo(p *Graph, e Edge) bool {
	for _, o := range p.outputs {
		if o == e {
			return true
		}
	}
	return false
}
func Aig_ObjIsBuf(p *Graph, e Edge) bool  { return p.node(e).Kind == Buffer }
func Aig_ObjIsAnd(p *Graph, e Edge) bool  { return p.node(e).Kind == And }
func Aig_ObjIsExor(p *Graph, e Edge) bool { return p.node(e).Kind == ExclusiveOr }
func Aig_ObjIsNode(p *Graph, e Edge) bool { k := p.node(e).Kind; return k == And || k == ExclusiveOr }
func Aig_ObjIsTerm(p *Graph, e Edge) bool {
	return p.node(e).Kind == Constant || p.node(e).Kind == Input || Aig_ObjIsCo(p, e)
}
func Aig_ObjIsHash(p *Graph, e Edge) bool { return Aig_ObjIsNode(p, e) }
func Aig_ObjIsChoice(p *Graph, e Edge) bool {
	_, ok := p.equivs[e.NodeID()]
	return ok && p.node(e).Refs > 0
}
func Aig_ObjIsCand(p *Graph, e Edge) bool { return p.node(e).Kind == Input || Aig_ObjIsNode(p, e) }
func Aig_ObjCioId(p *Graph, e Edge) int32 {
	if Aig_ObjIsNode(p, e) {
		panic("aig: internal node has no CioId")
	}
	return p.node(e).CioID
}
func Aig_ObjId(e Edge) int32 { return int32(e.NodeID()) }

func Aig_ObjIsMarkA(p *Graph, e Edge) bool { return p.node(e).MarkA }
func Aig_ObjSetMarkA(p *Graph, e Edge)     { p.node(e).MarkA = true }
func Aig_ObjClearMarkA(p *Graph, e Edge)   { p.node(e).MarkA = false }
func Aig_ObjIsMarkB(p *Graph, e Edge) bool { return p.node(e).MarkB }
func Aig_ObjSetMarkB(p *Graph, e Edge)     { p.node(e).MarkB = true }
func Aig_ObjClearMarkB(p *Graph, e Edge)   { p.node(e).MarkB = false }

func Aig_ObjSetTravId(p *Graph, e Edge, id int32)   { p.node(e).TravID = id }
func Aig_ObjSetTravIdCurrent(p *Graph, e Edge)      { p.node(e).TravID = p.travID }
func Aig_ObjSetTravIdPrevious(p *Graph, e Edge)     { p.node(e).TravID = p.travID - 1 }
func Aig_ObjIsTravIdCurrent(p *Graph, e Edge) bool  { return p.node(e).TravID == p.travID }
func Aig_ObjIsTravIdPrevious(p *Graph, e Edge) bool { return p.node(e).TravID == p.travID-1 }

func Aig_ObjPhase(p *Graph, e Edge) bool     { return p.node(e).Phase }
func Aig_ObjPhaseReal(p *Graph, e Edge) bool { return p.node(e).Phase != e.Complemented() }
func Aig_ObjRefs(p *Graph, e Edge) int32     { return int32(p.node(e).Refs) }
func Aig_ObjRef(p *Graph, e Edge)            { p.node(e).Refs++ }
func Aig_ObjDeref(p *Graph, e Edge) {
	if p.node(e).Refs <= 0 {
		panic("aig: reference underflow")
	}
	p.node(e).Refs--
}
func Aig_ObjClearRef(p *Graph, e Edge) { p.node(e).Refs = 0 }
func Aig_ObjFaninId0(p *Graph, e Edge) int32 {
	x := p.node(e).Fanins[0]
	if x == InvalidEdge {
		return -1
	}
	return int32(x.NodeID())
}
func Aig_ObjFaninId1(p *Graph, e Edge) int32 {
	x := p.node(e).Fanins[1]
	if x == InvalidEdge {
		return -1
	}
	return int32(x.NodeID())
}
func Aig_ObjFaninC0(p *Graph, e Edge) int32   { return int32(p.node(e).Fanins[0] & 1) }
func Aig_ObjFaninC1(p *Graph, e Edge) int32   { return int32(p.node(e).Fanins[1] & 1) }
func Aig_ObjFanin0(p *Graph, e Edge) Edge     { return Aig_Regular(p.node(e).Fanins[0]) }
func Aig_ObjFanin1(p *Graph, e Edge) Edge     { return Aig_Regular(p.node(e).Fanins[1]) }
func Aig_ObjChild0(p *Graph, e Edge) Edge     { return p.node(e).Fanins[0] }
func Aig_ObjChild1(p *Graph, e Edge) Edge     { return p.node(e).Fanins[1] }
func Aig_ObjChild0Copy(p *Graph, e Edge) Edge { return Aig_ObjRealCopy(p, p.node(e).Fanins[0]) }
func Aig_ObjChild1Copy(p *Graph, e Edge) Edge { return Aig_ObjRealCopy(p, p.node(e).Fanins[1]) }
func Aig_ObjChild0Next(p *Graph, e Edge) Edge {
	x := p.node(e).Fanins[0]
	n, ok := p.next[x.NodeID()]
	if !ok {
		return InvalidEdge
	}
	if x.Complemented() {
		n = n.Not()
	}
	return n
}
func Aig_ObjChild1Next(p *Graph, e Edge) Edge {
	x := p.node(e).Fanins[1]
	n, ok := p.next[x.NodeID()]
	if !ok {
		return InvalidEdge
	}
	if x.Complemented() {
		n = n.Not()
	}
	return n
}
func Aig_ObjChild0Flip(p *Graph, e Edge) { p.node(e).Fanins[0] = p.node(e).Fanins[0].Not() }
func Aig_ObjChild1Flip(p *Graph, e Edge) { p.node(e).Fanins[1] = p.node(e).Fanins[1].Not() }
func Aig_ObjCopy(p *Graph, e Edge) Edge {
	if x, ok := p.copies[e.NodeID()]; ok {
		return x
	}
	return InvalidEdge
}
func Aig_ObjSetCopy(p *Graph, e, copy Edge) {
	if p.copies == nil {
		p.copies = make(map[int]Edge)
	}
	p.copies[e.NodeID()] = copy
}
func Aig_ObjRealCopy(p *Graph, e Edge) Edge {
	x := Aig_ObjCopy(p, e)
	if x == InvalidEdge {
		return x
	}
	if e.Complemented() {
		x = x.Not()
	}
	return x
}
func Aig_ObjToLit(e Edge) int32 { return int32(e) }
func Aig_ObjFromLit(p *Graph, lit int32) Edge {
	e := Edge(lit)
	if !p.valid(e) {
		panic("aig: invalid literal")
	}
	return e
}
func Aig_ObjLevel(p *Graph, e Edge) int32 { return int32(p.node(e).Level) }
func Aig_ObjLevelNew(p *Graph, e Edge) int32 {
	n := p.node(e)
	if n.Kind == Buffer || n.Kind == Output {
		return int32(p.node(n.Fanins[0]).Level)
	}
	if n.Kind == Constant || n.Kind == Input {
		return 0
	}
	a, b := p.node(n.Fanins[0]).Level, p.node(n.Fanins[1]).Level
	if a > b {
		return int32(a + 1)
	}
	return int32(b + 1)
}
func Aig_ObjSetLevel(p *Graph, e Edge, l int32) int32 { p.node(e).Level = int(l); return l }
func Aig_ObjClean(p *Graph, e Edge)                   { *p.node(e) = Node{} }
func Aig_ObjFanout0(p *Graph, e Edge) Edge {
	for id, n := range p.nodes {
		if p.deleted[id] {
			continue
		}
		for _, x := range nodeFanins(n) {
			if x.NodeID() == e.NodeID() {
				return Edge(id) << 1
			}
		}
	}
	return InvalidEdge
}
func Aig_ObjEquiv(p *Graph, e Edge) Edge {
	if x, ok := p.equivs[e.NodeID()]; ok {
		return x
	}
	return InvalidEdge
}
func Aig_ObjSetEquiv(p *Graph, e, x Edge) {
	if p.equivs == nil {
		p.equivs = make(map[int]Edge)
	}
	p.equivs[e.NodeID()] = x
}
func Aig_ObjRepr(p *Graph, e Edge) Edge {
	if x, ok := p.reprs[e.NodeID()]; ok {
		return x
	}
	return InvalidEdge
}
func Aig_ObjSetRepr(p *Graph, e, x Edge) {
	if p.reprs == nil {
		p.reprs = make(map[int]Edge)
	}
	p.reprs[e.NodeID()] = x
}
func Aig_ObjWhatFanin(p *Graph, e, fanin Edge) int32 {
	for i, x := range p.node(e).Fanins {
		if x.NodeID() == fanin.NodeID() {
			return int32(i)
		}
	}
	return -1
}
func Aig_ObjFanoutC(p *Graph, e, fanout Edge) int32 {
	i := Aig_ObjWhatFanin(p, fanout, e)
	if i < 0 {
		panic("aig: not a fanout")
	}
	return int32(p.node(fanout).Fanins[i] & 1)
}

// create the ghost of the new node
func Aig_ObjCreateGhost(p *Graph, a, b Edge, kind Aig_Type_t) *Node {
	if a > b {
		a, b = b, a
	}
	k := And
	if kind == AIG_OBJ_EXOR {
		k = ExclusiveOr
	}
	p.ghost = Node{Kind: k, Fanins: [2]Edge{a, b}}
	return &p.ghost
}

// internal memory manager
func Aig_ManFetchMemory(p *Graph) Edge {
	e := Edge(len(p.nodes)) << 1
	p.nodes = append(p.nodes, Node{})
	return e
}
func Aig_ManRecycleMemory(p *Graph, e Edge) {
	if p.node(e).Refs != 0 {
		panic("aig: recycling referenced node")
	}
	p.deleted[e.NodeID()] = true
}

////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

// iterator over the combinational inputs
// Original macro:
// #define Aig_ManForEachCi( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( Aig_Obj_t *, p->vCis, pObj, i )

// Original macro:
// #define Aig_ManForEachCiReverse( p, pObj, i )                                   \
//     Vec_PtrForEachEntryReverse( Aig_Obj_t *, p->vCis, pObj, i )

// iterator over the combinational outputs
// Original macro:
// #define Aig_ManForEachCo( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( Aig_Obj_t *, p->vCos, pObj, i )

// Original macro:
// #define Aig_ManForEachCoReverse( p, pObj, i )                                   \
//     Vec_PtrForEachEntryReverse( Aig_Obj_t *, p->vCos, pObj, i )

// iterators over all objects, including those currently not used
// Original macro:
// #define Aig_ManForEachObj( p, pObj, i )                                         \
//     Vec_PtrForEachEntry( Aig_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL ) {} else

// Original macro:
// #define Aig_ManForEachObjReverse( p, pObj, i )                                  \
//     Vec_PtrForEachEntryReverse( Aig_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL ) {} else

// iterators over the objects whose IDs are stored in an array
// Original macro:
// #define Aig_ManForEachObjVec( vIds, p, pObj, i )                                \
//     for ( i = 0; i < Vec_IntSize(vIds) && (((pObj) = Aig_ManObj(p, Vec_IntEntry(vIds,i))), 1); i++ )

// Original macro:
// #define Aig_ManForEachObjVecReverse( vIds, p, pObj, i )                         \
//     for ( i = Vec_IntSize(vIds) - 1; i >= 0 && (((pObj) = Aig_ManObj(p, Vec_IntEntry(vIds,i))), 1); i-- )

// iterators over all nodes
// Original macro:
// #define Aig_ManForEachNode( p, pObj, i )                                        \
//     Vec_PtrForEachEntry( Aig_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL || !Aig_ObjIsNode(pObj) ) {} else

// Original macro:
// #define Aig_ManForEachNodeReverse( p, pObj, i )                                 \
//     Vec_PtrForEachEntryReverse( Aig_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL || !Aig_ObjIsNode(pObj) ) {} else

// iterator over all nodes
// Original macro:
// #define Aig_ManForEachExor( p, pObj, i )                                        \
//     Vec_PtrForEachEntry( Aig_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL || !Aig_ObjIsExor(pObj) ) {} else

// Original macro:
// #define Aig_ManForEachExorReverse( p, pObj, i )                                 \
//     Vec_PtrForEachEntryReverse( Aig_Obj_t *, p->vObjs, pObj, i ) if ( (pObj) == NULL || !Aig_ObjIsExor(pObj) ) {} else

// Circular fanout accessors for the reference-count-bounded iterator. Traverse
// exactly nRefs entries; a nonempty list returns to its first handle.
func Aig_ObjFanout0Int(p *Graph, id int32) int32 {
	if p.fanouts == nil || id < 0 || int(id) >= len(p.fanouts.data)/5 {
		panic("aig: fanout object outside allocated index")
	}
	return p.fanouts.data[5*id]
}
func Aig_ObjFanoutNext(p *Graph, fan int32) int32 {
	if p.fanouts == nil || fan < 0 || int(fan>>1) >= len(p.fanouts.data)/5 {
		panic("aig: fanout handle outside allocated index")
	}
	return *Aig_FanoutNext(p.fanouts.data, fan)
}

// iterator over the fanouts
// Original macro:
// #define Aig_ObjForEachFanout( p, pObj, pFanout, iFan, i )                       \
//     for ( assert(p->pFanData), i = 0; (i < (int)(pObj)->nRefs) &&               \
//           (((iFan) = i? Aig_ObjFanoutNext(p, iFan) : Aig_ObjFanout0Int(p, pObj->Id)), 1) && \
//           (((pFanout) = Aig_ManObj(p, iFan>>1)), 1); i++ )

////////////////////////////////////////////////////////////////////////
///                     SEQUENTIAL ITERATORS                         ///
////////////////////////////////////////////////////////////////////////

// iterator over the primary inputs
// Original macro:
// #define Aig_ManForEachPiSeq( p, pObj, i )                                       \
//     Vec_PtrForEachEntryStop( Aig_Obj_t *, p->vCis, pObj, i, Aig_ManCiNum(p)-Aig_ManRegNum(p) )

// iterator over the latch outputs
// Original macro:
// #define Aig_ManForEachLoSeq( p, pObj, i )                                       \
//     Vec_PtrForEachEntryStart( Aig_Obj_t *, p->vCis, pObj, i, Aig_ManCiNum(p)-Aig_ManRegNum(p) )

// iterator over the primary outputs
// Original macro:
// #define Aig_ManForEachPoSeq( p, pObj, i )                                       \
//     Vec_PtrForEachEntryStop( Aig_Obj_t *, p->vCos, pObj, i, Aig_ManCoNum(p)-Aig_ManRegNum(p) )

// iterator over the latch inputs
// Original macro:
// #define Aig_ManForEachLiSeq( p, pObj, i )                                       \
//     Vec_PtrForEachEntryStart( Aig_Obj_t *, p->vCos, pObj, i, Aig_ManCoNum(p)-Aig_ManRegNum(p) )

// iterator over the latch input and outputs
// Original macro:
// #define Aig_ManForEachLiLoSeq( p, pObjLi, pObjLo, k )                           \
//     for ( k = 0; (k < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, k)), 1)    \
//         && (((pObjLo)=Aig_ManLo(p, k)), 1); k++ )

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== aigCheck.c ========================================================*/
// Aig_ManCheck is implemented in aigCheck.go.
// extern void            Aig_ManCheckMarkA( Aig_Man_t * p );

// extern void            Aig_ManCheckPhase( Aig_Man_t * p );

/*=== aigCuts.c ========================================================*/
// extern Aig_ManCut_t *  Aig_ComputeCuts( Aig_Man_t * pAig, int nCutsMax, int nLeafMax, int fTruth, int fVerbose );

// extern void            Aig_ManCutStop( Aig_ManCut_t * p );

/*=== aigDfs.c ==========================================================*/
// extern int             Aig_ManVerifyTopoOrder( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManDfs( Aig_Man_t * p, int fNodesOnly );

// extern Vec_Ptr_t *     Aig_ManDfsAll( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManDfsPreorder( Aig_Man_t * p, int fNodesOnly );

// extern Vec_Vec_t *     Aig_ManLevelize( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManDfsNodes( Aig_Man_t * p, Aig_Obj_t ** ppNodes, int nNodes );

// extern Vec_Ptr_t *     Aig_ManDfsChoices( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManDfsReverse( Aig_Man_t * p );

// extern int             Aig_ManLevelNum( Aig_Man_t * p );

// extern int             Aig_ManChoiceLevel( Aig_Man_t * p );

// extern int             Aig_DagSize( Aig_Obj_t * pObj );

// extern int             Aig_SupportSize( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern Vec_Ptr_t *     Aig_Support( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_SupportNodes( Aig_Man_t * p, Aig_Obj_t ** ppObjs, int nObjs, Vec_Ptr_t * vSupp );

// extern void            Aig_ConeUnmark_rec( Aig_Obj_t * pObj );

// extern Aig_Obj_t *     Aig_Transfer( Aig_Man_t * pSour, Aig_Man_t * pDest, Aig_Obj_t * pObj, int nVars );

// extern Aig_Obj_t *     Aig_Compose( Aig_Man_t * p, Aig_Obj_t * pRoot, Aig_Obj_t * pFunc, int iVar );

// extern void            Aig_ObjCollectCut( Aig_Obj_t * pRoot, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vNodes );

// extern int             Aig_ObjCollectSuper( Aig_Obj_t * pObj, Vec_Ptr_t * vSuper );

/*=== aigDup.c ==========================================================*/
// extern Aig_Obj_t *     Aig_ManDupSimpleDfs_rec( Aig_Man_t * pNew, Aig_Man_t * p, Aig_Obj_t * pObj );

// extern Aig_Man_t *     Aig_ManDupSimple( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupSimpleWithHints( Aig_Man_t * p, Vec_Int_t * vHints );

// extern Aig_Man_t *     Aig_ManDupSimpleDfs( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupSimpleDfsPart( Aig_Man_t * p, Vec_Ptr_t * vPis, Vec_Ptr_t * vCos );

// extern Aig_Man_t *     Aig_ManDupOrdered( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupCof( Aig_Man_t * p, int iInput, int Value );

// extern Aig_Man_t *     Aig_ManDupTrim( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupExor( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupDfs( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManOrderPios( Aig_Man_t * p, Aig_Man_t * pOrder );

// extern Aig_Man_t *     Aig_ManDupDfsGuided( Aig_Man_t * p, Vec_Ptr_t * vPios );

// extern Aig_Man_t *     Aig_ManDupLevelized( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupWithoutPos( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupFlopsOnly( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupRepres( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupRepresDfs( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManCreateMiter( Aig_Man_t * p1, Aig_Man_t * p2, int fImpl );

// extern Aig_Man_t *     Aig_ManDupOrpos( Aig_Man_t * p, int fAddRegs );

// extern Aig_Man_t *     Aig_ManDupOneOutput( Aig_Man_t * p, int iPoNum, int fAddRegs );

// extern Aig_Man_t *     Aig_ManDupUnsolvedOutputs( Aig_Man_t * p, int fAddRegs );

// extern Aig_Man_t *     Aig_ManDupArray( Vec_Ptr_t * vArray );

// extern Aig_Man_t *     Aig_ManDupNodes( Aig_Man_t * pMan, Vec_Ptr_t * vArray );

/*=== aigFanout.c ==========================================================*/
// extern void            Aig_ObjAddFanout( Aig_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t * pFanout );

// extern void            Aig_ObjRemoveFanout( Aig_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t * pFanout );

// extern void            Aig_ManFanoutStart( Aig_Man_t * p );

// extern void            Aig_ManFanoutStop( Aig_Man_t * p );

/*=== aigFrames.c ==========================================================*/
// extern Aig_Man_t *     Aig_ManFrames( Aig_Man_t * pAig, int nFs, int fInit, int fOuts, int fRegs, int fEnlarge, Aig_Obj_t *** ppObjMap );

/*=== aigMan.c ==========================================================*/
// extern Aig_Man_t *     Aig_ManStart( int nNodesMax );

// extern Aig_Man_t *     Aig_ManStartFrom( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManExtractMiter( Aig_Man_t * p, Aig_Obj_t * pNode1, Aig_Obj_t * pNode2 );

// extern void            Aig_ManStop( Aig_Man_t * p );

// extern void            Aig_ManStopP( Aig_Man_t ** p );

// extern int             Aig_ManCleanup( Aig_Man_t * p );

// extern int             Aig_ManAntiCleanup( Aig_Man_t * p );

// extern int             Aig_ManCiCleanup( Aig_Man_t * p );

// extern int             Aig_ManCoCleanup( Aig_Man_t * p );

// extern void            Aig_ManPrintStats( Aig_Man_t * p );

// extern void            Aig_ManReportImprovement( Aig_Man_t * p, Aig_Man_t * pNew );

// extern void            Aig_ManSetRegNum( Aig_Man_t * p, int nRegs );

// extern void            Aig_ManFlipFirstPo( Aig_Man_t * p );

// extern void *          Aig_ManReleaseData( Aig_Man_t * p );

/*=== aigMem.c ==========================================================*/
// extern void            Aig_ManStartMemory( Aig_Man_t * p );

// extern void            Aig_ManStopMemory( Aig_Man_t * p );

/*=== aigMffc.c ==========================================================*/
// extern int             Aig_NodeRef_rec( Aig_Obj_t * pNode, unsigned LevelMin );

// extern int             Aig_NodeDeref_rec( Aig_Obj_t * pNode, unsigned LevelMin, float * pPower, float * pProbs );

// extern int             Aig_NodeMffcSupp( Aig_Man_t * p, Aig_Obj_t * pNode, int LevelMin, Vec_Ptr_t * vSupp );

// extern int             Aig_NodeMffcLabel( Aig_Man_t * p, Aig_Obj_t * pNode, float * pPower );

// extern int             Aig_NodeMffcLabelCut( Aig_Man_t * p, Aig_Obj_t * pNode, Vec_Ptr_t * vLeaves );

// extern int             Aig_NodeMffcExtendCut( Aig_Man_t * p, Aig_Obj_t * pNode, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vResult );

/*=== aigObj.c ==========================================================*/
// extern Aig_Obj_t *     Aig_ObjCreateCi( Aig_Man_t * p );

// extern Aig_Obj_t *     Aig_ObjCreateCo( Aig_Man_t * p, Aig_Obj_t * pDriver );

// extern Aig_Obj_t *     Aig_ObjCreate( Aig_Man_t * p, Aig_Obj_t * pGhost );

// extern void            Aig_ObjConnect( Aig_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t * pFan0, Aig_Obj_t * pFan1 );

// extern void            Aig_ObjDisconnect( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_ObjDelete( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_ObjDelete_rec( Aig_Man_t * p, Aig_Obj_t * pObj, int fFreeTop );

// extern void            Aig_ObjDeletePo( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_ObjPrint( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_ObjPatchFanin0( Aig_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t * pFaninNew );

// extern void            Aig_ObjReplace( Aig_Man_t * p, Aig_Obj_t * pObjOld, Aig_Obj_t * pObjNew, int fUpdateLevel );

/*=== aigOper.c =========================================================*/
// extern Aig_Obj_t *     Aig_IthVar( Aig_Man_t * p, int i );

// extern Aig_Obj_t *     Aig_Oper( Aig_Man_t * p, Aig_Obj_t * p0, Aig_Obj_t * p1, Aig_Type_t Type );

// extern Aig_Obj_t *     Aig_And( Aig_Man_t * p, Aig_Obj_t * p0, Aig_Obj_t * p1 );

// extern Aig_Obj_t *     Aig_Or( Aig_Man_t * p, Aig_Obj_t * p0, Aig_Obj_t * p1 );

// extern Aig_Obj_t *     Aig_Exor( Aig_Man_t * p, Aig_Obj_t * p0, Aig_Obj_t * p1 );

// extern Aig_Obj_t *     Aig_Mux( Aig_Man_t * p, Aig_Obj_t * pC, Aig_Obj_t * p1, Aig_Obj_t * p0 );

// extern Aig_Obj_t *     Aig_Maj( Aig_Man_t * p, Aig_Obj_t * pA, Aig_Obj_t * pB, Aig_Obj_t * pC );

// extern Aig_Obj_t *     Aig_Multi( Aig_Man_t * p, Aig_Obj_t ** pArgs, int nArgs, Aig_Type_t Type );

// extern Aig_Obj_t *     Aig_Miter( Aig_Man_t * p, Vec_Ptr_t * vPairs );

// extern Aig_Obj_t *     Aig_MiterTwo( Aig_Man_t * p, Vec_Ptr_t * vNodes1, Vec_Ptr_t * vNodes2 );

// extern Aig_Obj_t *     Aig_CreateAnd( Aig_Man_t * p, int nVars );

// extern Aig_Obj_t *     Aig_CreateOr( Aig_Man_t * p, int nVars );

// extern Aig_Obj_t *     Aig_CreateExor( Aig_Man_t * p, int nVars );

/*=== aigOrder.c =========================================================*/
// extern void            Aig_ManOrderStart( Aig_Man_t * p );

// extern void            Aig_ManOrderStop( Aig_Man_t * p );

// extern void            Aig_ObjOrderInsert( Aig_Man_t * p, int ObjId );

// extern void            Aig_ObjOrderRemove( Aig_Man_t * p, int ObjId );

// extern void            Aig_ObjOrderAdvance( Aig_Man_t * p );

/*=== aigPart.c =========================================================*/
// extern Vec_Ptr_t *     Aig_ManSupports( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManSupportsInverse( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManSupportsRegisters( Aig_Man_t * p );

// extern Vec_Ptr_t *     Aig_ManPartitionSmart( Aig_Man_t * p, int nPartSizeLimit, int fVerbose, Vec_Ptr_t ** pvPartSupps );

// extern Vec_Ptr_t *     Aig_ManPartitionSmartRegisters( Aig_Man_t * pAig, int nSuppSizeLimit, int fVerbose );

// extern Vec_Ptr_t *     Aig_ManPartitionNaive( Aig_Man_t * p, int nPartSize );

// extern Vec_Ptr_t *     Aig_ManMiterPartitioned( Aig_Man_t * p1, Aig_Man_t * p2, int nPartSize, int fSmart );

// extern Aig_Man_t *     Aig_ManChoicePartitioned( Vec_Ptr_t * vAigs, int nPartSize, int nConfMax, int nLevelMax, int fVerbose );

// extern Aig_Man_t *     Aig_ManFraigPartitioned( Aig_Man_t * pAig, int nPartSize, int nConfMax, int nLevelMax, int fVerbose );

// extern Aig_Man_t *     Aig_ManChoiceConstructive( Vec_Ptr_t * vAigs, int fVerbose );

/*=== aigPartReg.c =========================================================*/
// extern Vec_Ptr_t *     Aig_ManRegPartitionSimple( Aig_Man_t * pAig, int nPartSize, int nOverSize );

// extern void            Aig_ManPartDivide( Vec_Ptr_t * vResult, Vec_Int_t * vDomain, int nPartSize, int nOverSize );

// extern Vec_Ptr_t *     Aig_ManRegPartitionSmart( Aig_Man_t * pAig, int nPartSize );

// extern Aig_Man_t *     Aig_ManRegCreatePart( Aig_Man_t * pAig, Vec_Int_t * vPart, int * pnCountPis, int * pnCountRegs, int ** ppMapBack );

// extern Vec_Ptr_t *     Aig_ManRegProjectOnehots( Aig_Man_t * pAig, Aig_Man_t * pPart, Vec_Ptr_t * vOnehots, int fVerbose );

/*=== aigRepr.c =========================================================*/
// extern void            Aig_ManReprStart( Aig_Man_t * p, int nIdMax );

// extern void            Aig_ManReprStop( Aig_Man_t * p );

// extern void            Aig_ObjCreateRepr( Aig_Man_t * p, Aig_Obj_t * pNode1, Aig_Obj_t * pNode2 );

// extern void            Aig_ManTransferRepr( Aig_Man_t * pNew, Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManDupRepr( Aig_Man_t * p, int fOrdered );

// extern Aig_Man_t *     Aig_ManDupReprBasic( Aig_Man_t * p );

// extern int             Aig_ManCountReprs( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManRehash( Aig_Man_t * p );

// extern int             Aig_ObjCheckTfi( Aig_Man_t * p, Aig_Obj_t * pNew, Aig_Obj_t * pOld );

// extern void            Aig_ManMarkValidChoices( Aig_Man_t * p );

// extern int             Aig_TransferMappedClasses( Aig_Man_t * pAig, Aig_Man_t * pPart, int * pMapBack );

/*=== aigRet.c ========================================================*/
// extern Aig_Man_t *     Rtm_ManRetime( Aig_Man_t * p, int fForward, int nStepsMax, int fVerbose );

/*=== aigRetF.c ========================================================*/
// extern Aig_Man_t *     Aig_ManRetimeFrontier( Aig_Man_t * p, int nStepsMax );

/*=== aigScl.c ==========================================================*/
// extern Aig_Man_t *     Aig_ManRemap( Aig_Man_t * p, Vec_Ptr_t * vMap );

// extern int             Aig_ManSeqCleanup( Aig_Man_t * p );

// extern int             Aig_ManSeqCleanupBasic( Aig_Man_t * p );

// extern int             Aig_ManCountMergeRegs( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManReduceLaches( Aig_Man_t * p, int fVerbose );

// extern void            Aig_ManComputeSccs( Aig_Man_t * p );

// extern Aig_Man_t *     Aig_ManScl( Aig_Man_t * pAig, int fLatchConst, int fLatchEqual, int fUseMvSweep, int nFramesSymb, int nFramesSatur, int fVerbose, int fVeryVerbose );

/*=== aigShow.c ========================================================*/
// extern void            Aig_ManShow( Aig_Man_t * pMan, int fHaig, Vec_Ptr_t * vBold );

/*=== aigTable.c ========================================================*/
// extern Aig_Obj_t *     Aig_TableLookup( Aig_Man_t * p, Aig_Obj_t * pGhost );

// extern Aig_Obj_t *     Aig_TableLookupTwo( Aig_Man_t * p, Aig_Obj_t * pFanin0, Aig_Obj_t * pFanin1 );

// extern void            Aig_TableInsert( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_TableDelete( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern int             Aig_TableCountEntries( Aig_Man_t * p );

// extern void            Aig_TableProfile( Aig_Man_t * p );

// extern void            Aig_TableClear( Aig_Man_t * p );

/*=== aigTiming.c ========================================================*/
// extern void            Aig_ObjClearReverseLevel( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern int             Aig_ObjRequiredLevel( Aig_Man_t * p, Aig_Obj_t * pObj );

// extern void            Aig_ManStartReverseLevels( Aig_Man_t * p, int nMaxLevelIncrease );

// extern void            Aig_ManStopReverseLevels( Aig_Man_t * p );

// extern void            Aig_ManUpdateLevel( Aig_Man_t * p, Aig_Obj_t * pObjNew );

// extern void            Aig_ManUpdateReverseLevel( Aig_Man_t * p, Aig_Obj_t * pObjNew );

// extern void            Aig_ManVerifyLevel( Aig_Man_t * p );

// extern void            Aig_ManVerifyReverseLevel( Aig_Man_t * p );

/*=== aigTruth.c ========================================================*/
// extern unsigned *      Aig_ManCutTruth( Aig_Obj_t * pRoot, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vNodes, Vec_Ptr_t * vTruthElem, Vec_Ptr_t * vTruthStore );

/*=== aigTsim.c ========================================================*/
// extern Aig_Man_t *     Aig_ManConstReduce( Aig_Man_t * p, int fUseMvSweep, int nFramesSymb, int nFramesSatur, int fVerbose, int fVeryVerbose );

/*=== aigUtil.c =========================================================*/
// extern void            Aig_ManIncrementTravId( Aig_Man_t * p );

// extern char *          Aig_TimeStamp();

// extern int             Aig_ManHasNoGaps( Aig_Man_t * p );

// extern int             Aig_ManLevels( Aig_Man_t * p );

// extern void            Aig_ManResetRefs( Aig_Man_t * p );

// extern void            Aig_ManCleanMarkA( Aig_Man_t * p );

// extern void            Aig_ManCleanMarkB( Aig_Man_t * p );

// extern void            Aig_ManCleanMarkAB( Aig_Man_t * p );

// extern void            Aig_ManCleanData( Aig_Man_t * p );

// extern void            Aig_ObjCleanData_rec( Aig_Obj_t * pObj );

// extern void            Aig_ManCleanNext( Aig_Man_t * p );

// extern void            Aig_ObjCollectMulti( Aig_Obj_t * pFunc, Vec_Ptr_t * vSuper );

// extern int             Aig_ObjIsMuxType( Aig_Obj_t * pObj );

// extern int             Aig_ObjRecognizeExor( Aig_Obj_t * pObj, Aig_Obj_t ** ppFan0, Aig_Obj_t ** ppFan1 );

// extern Aig_Obj_t *     Aig_ObjRecognizeMux( Aig_Obj_t * pObj, Aig_Obj_t ** ppObjT, Aig_Obj_t ** ppObjE );

// extern Aig_Obj_t *     Aig_ObjReal_rec( Aig_Obj_t * pObj );

// extern int             Aig_ObjCompareIdIncrease( Aig_Obj_t ** pp1, Aig_Obj_t ** pp2 );

// extern void            Aig_ObjPrintEqn( FILE * pFile, Aig_Obj_t * pObj, Vec_Vec_t * vLevels, int Level );

// extern void            Aig_ObjPrintVerilog( FILE * pFile, Aig_Obj_t * pObj, Vec_Vec_t * vLevels, int Level );

// extern void            Aig_ObjPrintVerbose( Aig_Obj_t * pObj, int fHaig );

// extern void            Aig_ManPrintVerbose( Aig_Man_t * p, int fHaig );

// extern void            Aig_ManDump( Aig_Man_t * p );

// extern void            Aig_ManDumpBlif( Aig_Man_t * p, char * pFileName, Vec_Ptr_t * vPiNames, Vec_Ptr_t * vPoNames );

// extern void            Aig_ManDumpVerilog( Aig_Man_t * p, char * pFileName );

// extern void            Aig_ManSetCioIds( Aig_Man_t * p );

// extern void            Aig_ManCleanCioIds( Aig_Man_t * p );

// extern int             Aig_ManChoiceNum( Aig_Man_t * p );

// extern char *          Aig_FileNameGenericAppend( char * pBase, char * pSuffix );

// extern unsigned        Aig_ManRandom( int fReset );

// extern word            Aig_ManRandom64( int fReset );

// extern void            Aig_ManRandomInfo( Vec_Ptr_t * vInfo, int iInputStart, int iWordStart, int iWordStop );

// extern void            Aig_NodeUnionLists( Vec_Ptr_t * vArr1, Vec_Ptr_t * vArr2, Vec_Ptr_t * vArr );

// extern void            Aig_NodeIntersectLists( Vec_Ptr_t * vArr1, Vec_Ptr_t * vArr2, Vec_Ptr_t * vArr );

// extern void            Aig_ManSetPhase( Aig_Man_t * pAig );

// extern Vec_Ptr_t *     Aig_ManMuxesCollect( Aig_Man_t * pAig );

// extern void            Aig_ManMuxesDeref( Aig_Man_t * pAig, Vec_Ptr_t * vMuxes );

// extern void            Aig_ManMuxesRef( Aig_Man_t * pAig, Vec_Ptr_t * vMuxes );

// extern void            Aig_ManInvertConstraints( Aig_Man_t * pAig );

/*=== aigWin.c =========================================================*/
// extern void            Aig_ManFindCut( Aig_Obj_t * pRoot, Vec_Ptr_t * vFront, Vec_Ptr_t * vVisited, int nSizeLimit, int nFanoutLimit );

/*=== aigMem.c ===========================================================*/
// fixed-size-block memory manager
// extern Aig_MmFixed_t * Aig_MmFixedStart( int nEntrySize, int nEntriesMax );

// extern void            Aig_MmFixedStop( Aig_MmFixed_t * p, int fVerbose );

// extern char *          Aig_MmFixedEntryFetch( Aig_MmFixed_t * p );

// extern void            Aig_MmFixedEntryRecycle( Aig_MmFixed_t * p, char * pEntry );

// extern void            Aig_MmFixedRestart( Aig_MmFixed_t * p );

// extern int             Aig_MmFixedReadMemUsage( Aig_MmFixed_t * p );

// extern int             Aig_MmFixedReadMaxEntriesUsed( Aig_MmFixed_t * p );

// flexible-size-block memory manager
// extern Aig_MmFlex_t *  Aig_MmFlexStart();

// extern void            Aig_MmFlexStop( Aig_MmFlex_t * p, int fVerbose );

// extern char *          Aig_MmFlexEntryFetch( Aig_MmFlex_t * p, int nBytes );

// extern void            Aig_MmFlexRestart( Aig_MmFlex_t * p );

// extern int             Aig_MmFlexReadMemUsage( Aig_MmFlex_t * p );

// hierarchical memory manager
// extern Aig_MmStep_t *  Aig_MmStepStart( int nSteps );

// extern void            Aig_MmStepStop( Aig_MmStep_t * p, int fVerbose );

// extern char *          Aig_MmStepEntryFetch( Aig_MmStep_t * p, int nBytes );

// extern void            Aig_MmStepEntryRecycle( Aig_MmStep_t * p, char * pEntry, int nBytes );

// extern int             Aig_MmStepReadMemUsage( Aig_MmStep_t * p );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

// Aig_ManSetConstrNum marks the trailing ordinary outputs as constraints.
func Aig_ManSetConstrNum(p *Graph, count int32) {
	if count < 0 || int(count) > len(p.outputs)-int(p.nRegs) {
		panic("aig: invalid constraint count")
	}
	p.nConstrs = count
}
