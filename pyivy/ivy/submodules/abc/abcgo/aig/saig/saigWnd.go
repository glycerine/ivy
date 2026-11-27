package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Go port of Saig_ManCollectedDiffNodes .. Saig_ManWindowExtractMiter in
// saigWnd.c. The objects' pData (their copies in the new manager) are kept
// in per-manager slices indexed by object ID; aig.InvalidEdge is NULL.

func wndCopies(p *aig.Graph) []aig.Edge {
	c := make([]aig.Edge, aig.Aig_ManObjNumMax(p))
	for i := range c {
		c[i] = aig.InvalidEdge
	}
	return c
}

// wndChildCopy is Aig_ObjChild0Copy/Aig_ObjChild1Copy.
func wndChildCopy(copies []aig.Edge, child aig.Edge) aig.Edge {
	c := copies[child.NodeID()]
	if c == aig.InvalidEdge {
		panic("saig: window object has no copy")
	}
	return aig.Aig_NotCond(c, aig.Aig_IsComplement(child))
}

// Saig_ManCollectedDiffNodes marks (with the current traversal ID of p0)
// the objects of p0 without a match in p1 and returns the unmatched CIs and
// nodes.
func Saig_ManCollectedDiffNodes(p0, p1 *aig.Graph) *vec.Vec_Ptr_t {
	// collect nodes that are not linked
	aig.Aig_ManIncrementTravId(p0)
	vNodes := vec.Vec_PtrAlloc(1000)
	for _, pObj0 := range p0.LiveEdges() {
		pObj1 := aig.Aig_ObjRepr(p0, pObj0)
		if pObj1 != aig.InvalidEdge {
			if pObj0 != aig.Aig_ObjRepr(p1, pObj1) {
				panic("saig: pObj0 == Aig_ObjRepr( p1, pObj1 )")
			}
			continue
		}
		// mark and collect unmatched objects
		aig.Aig_ObjSetTravIdCurrent(p0, pObj0)
		if strSimIsCiOrNode(p0, pObj0) {
			vec.Vec_PtrPush(vNodes, pObj0)
		}
	}
	// make sure LI/LO are labeled/unlabeled mutually
	for i := int32(0); i < Saig_ManRegNum(p0); i++ {
		if aig.Aig_ObjIsTravIdCurrent(p0, Saig_ManLi(p0, i)) != aig.Aig_ObjIsTravIdCurrent(p0, Saig_ManLo(p0, i)) {
			panic("saig: LI/LO are not labeled mutually")
		}
	}
	return vNodes
}

// Saig_ManWindowCreatePis creates the PIs of the miter for the matched
// fanins of the unmatched objects of p0.
func Saig_ManWindowCreatePis(pNew, p0, p1 *aig.Graph, c0, c1 []aig.Edge, vNodes0 *vec.Vec_Ptr_t) {
	Counter := 0
	createPi := func(pFanin aig.Edge) {
		if !aig.Aig_ObjIsTravIdCurrent(p0, pFanin) && c0[pFanin.NodeID()] == aig.InvalidEdge {
			c0[pFanin.NodeID()] = aig.Aig_ObjCreateCi(pNew)
			pMatch := aig.Aig_ObjRepr(p0, pFanin)
			if pMatch == aig.InvalidEdge || pFanin != aig.Aig_ObjRepr(p1, pMatch) {
				panic("saig: pFanin == Aig_ObjRepr( p1, pMatch )")
			}
			c1[pMatch.NodeID()] = c0[pFanin.NodeID()]
			Counter++
		}
	}
	for i := 0; i < vec.Vec_PtrSize(vNodes0); i++ {
		pObj := vec.Vec_PtrEntry(vNodes0, i).(aig.Edge)
		if strSimIsLo(p0, pObj) {
			createPi(aig.Aig_ObjFanin0(p0, strSimLoToLi(p0, pObj)))
		} else {
			if !strSimIsNode(p0, pObj) {
				panic("saig: Aig_ObjIsNode(pObj)")
			}
			createPi(aig.Aig_ObjFanin0(p0, pObj))
			createPi(aig.Aig_ObjFanin1(p0, pObj))
		}
	}
	//    printf( "Added %d primary inputs.\n", Counter );
}

// Saig_ManWindowCreatePos creates the POs of the miter: one XOR for each
// unmatched fanin of a matched CO or node.
func Saig_ManWindowCreatePos(pNew, p0, p1 *aig.Graph, c0, c1 []aig.Edge) {
	createPo := func(pFanin0, pFanin1 aig.Edge) {
		if aig.Aig_ObjIsTravIdCurrent(p0, pFanin0) != aig.Aig_ObjIsTravIdCurrent(p1, pFanin1) {
			panic("saig: fanins are not labeled mutually")
		}
		if aig.Aig_ObjIsTravIdCurrent(p0, pFanin0) {
			pMiter := aig.Aig_Exor(pNew, c0[pFanin0.NodeID()], c1[pFanin1.NodeID()])
			aig.Aig_ObjCreateCo(pNew, pMiter)
		}
	}
	for _, pObj0 := range p0.LiveEdges() {
		if aig.Aig_ObjIsTravIdCurrent(p0, pObj0) {
			continue
		}
		if aig.Aig_ObjIsConst1(p0, pObj0) {
			continue
		}
		if aig.Aig_ObjIsCi(p0, pObj0) {
			continue
		}
		pObj1 := aig.Aig_ObjRepr(p0, pObj0)
		if pObj0 != aig.Aig_ObjRepr(p1, pObj1) {
			panic("saig: pObj0 == Aig_ObjRepr( p1, pObj1 )")
		}
		if k, _, _ := p0.KindFanins(pObj0); k == aig.Output {
			createPo(aig.Aig_ObjFanin0(p0, pObj0), aig.Aig_ObjFanin0(p1, pObj1))
		} else {
			if !strSimIsNode(p0, pObj0) {
				panic("saig: Aig_ObjIsNode(pObj0)")
			}
			createPo(aig.Aig_ObjFanin0(p0, pObj0), aig.Aig_ObjFanin0(p1, pObj1))
			createPo(aig.Aig_ObjFanin1(p0, pObj0), aig.Aig_ObjFanin1(p1, pObj1))
		}
	}
}

// Saig_ManWindowExtractMiter extracts the miter of the unmatched windows of
// p0 and p1, whose representatives link the matched objects; it adds the
// matching of the POs and of the LIs of matched LOs.
func Saig_ManWindowExtractMiter(p0, p1 *aig.Graph) *aig.Graph {
	// add matching of POs and LIs
	for i := int32(0); i < Saig_ManPoNum(p0); i++ {
		pObj0 := aig.Aig_ManCo(p0, i)
		pObj1 := aig.Aig_ManCo(p1, i)
		aig.Aig_ObjSetRepr(p0, pObj0, pObj1)
		aig.Aig_ObjSetRepr(p1, pObj1, pObj0)
	}
	for i := int32(0); i < Saig_ManRegNum(p0); i++ {
		pObj0 := Saig_ManLi(p0, i)
		pMatch0 := strSimLiToLo(p0, pObj0)
		pMatch1 := aig.Aig_ObjRepr(p0, pMatch0)
		if pMatch1 == aig.InvalidEdge {
			continue
		}
		if pMatch0 != aig.Aig_ObjRepr(p1, pMatch1) {
			panic("saig: pMatch0 == Aig_ObjRepr( p1, pMatch1 )")
		}
		pObj1 := strSimLoToLi(p1, pMatch1)
		aig.Aig_ObjSetRepr(p0, pObj0, pObj1)
		aig.Aig_ObjSetRepr(p1, pObj1, pObj0)
	}
	// clean the markings
	c0, c1 := wndCopies(p0), wndCopies(p1)
	// collect nodes that are not linked
	vNodes0 := Saig_ManCollectedDiffNodes(p0, p1)
	vNodes1 := Saig_ManCollectedDiffNodes(p1, p0)
	// create the new manager
	pNew := aig.Aig_ManStart(int32(vec.Vec_PtrSize(vNodes0) + vec.Vec_PtrSize(vNodes1)))
	pNew.Name = "wnd"
	pNew.Spec = ""
	// map constant nodes
	c0[0] = aig.Aig_ManConst1(pNew)
	c1[0] = aig.Aig_ManConst1(pNew)
	// create real PIs
	Saig_ManWindowCreatePis(pNew, p0, p1, c0, c1, vNodes0)
	Saig_ManWindowCreatePis(pNew, p1, p0, c1, c0, vNodes1)
	// create register outputs
	for _, side := range []struct {
		p      *aig.Graph
		c      []aig.Edge
		vNodes *vec.Vec_Ptr_t
	}{{p0, c0, vNodes0}, {p1, c1, vNodes1}} {
		for i := 0; i < vec.Vec_PtrSize(side.vNodes); i++ {
			pObj := vec.Vec_PtrEntry(side.vNodes, i).(aig.Edge)
			if strSimIsLo(side.p, pObj) {
				side.c[pObj.NodeID()] = aig.Aig_ObjCreateCi(pNew)
			}
		}
	}
	// create internal nodes
	for _, side := range []struct {
		p      *aig.Graph
		c      []aig.Edge
		vNodes *vec.Vec_Ptr_t
	}{{p0, c0, vNodes0}, {p1, c1, vNodes1}} {
		for i := 0; i < vec.Vec_PtrSize(side.vNodes); i++ {
			pObj := vec.Vec_PtrEntry(side.vNodes, i).(aig.Edge)
			if strSimIsNode(side.p, pObj) {
				_, f0, f1 := side.p.KindFanins(pObj)
				side.c[pObj.NodeID()] = aig.Aig_And(pNew, wndChildCopy(side.c, f0), wndChildCopy(side.c, f1))
			}
		}
	}
	// create POs
	Saig_ManWindowCreatePos(pNew, p0, p1, c0, c1)
	//    Saig_ManWindowCreatePos( pNew, p1, p0 );
	// create register inputs
	nRegCount := int32(0)
	for _, side := range []struct {
		p      *aig.Graph
		c      []aig.Edge
		vNodes *vec.Vec_Ptr_t
	}{{p0, c0, vNodes0}, {p1, c1, vNodes1}} {
		for i := 0; i < vec.Vec_PtrSize(side.vNodes); i++ {
			pObj := vec.Vec_PtrEntry(side.vNodes, i).(aig.Edge)
			if strSimIsLo(side.p, pObj) {
				pMatch := strSimLoToLi(side.p, pObj)
				_, f0, _ := side.p.KindFanins(pMatch)
				aig.Aig_ObjCreateCo(pNew, wndChildCopy(side.c, f0))
				nRegCount++
			}
		}
	}
	aig.Aig_ManSetRegNum(pNew, nRegCount)
	aig.Aig_ManCleanup(pNew)
	return pNew
}
