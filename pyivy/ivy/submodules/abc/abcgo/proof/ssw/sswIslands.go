// Go port of sswIslands.c (detection of islands of difference) by Alan
// Mishchenko, UC Berkeley (2008).
package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// The objects' pData (C) is the graph's object data: the matched object of
// the other manager, an aig.Edge; absence is NULL.

func sswIslandMatch(g *aig.Graph, e aig.Edge) (aig.Edge, bool) {
	match, ok := g.ObjectData(e).(aig.Edge)
	return match, ok
}

func sswIslandData(g *aig.Graph, e aig.Edge) aig.Edge {
	if match, ok := sswIslandMatch(g, e); ok {
		return match
	}
	return aig.InvalidEdge
}

func sswIslandSetData(g *aig.Graph, e, data aig.Edge) {
	if data == aig.InvalidEdge {
		g.SetObjectData(aig.Aig_Regular(e), nil)
		return
	}
	g.SetObjectData(aig.Aig_Regular(e), data)
}

// sswIslandChildCopy is Aig_ObjChild0Copy/Aig_ObjChild1Copy over pData.
func sswIslandChildCopy(g *aig.Graph, child aig.Edge) aig.Edge {
	c := sswIslandData(g, child)
	if c == aig.InvalidEdge {
		panic("ssw: fanin has no copy")
	}
	return aig.Aig_NotCond(c, aig.Aig_IsComplement(child))
}

func sswIslandKind(g *aig.Graph, e aig.Edge) aig.NodeKind {
	k, _, _ := g.KindFanins(e)
	return k
}

func sswIslandIsCiOrNode(g *aig.Graph, e aig.Edge) bool {
	k := sswIslandKind(g, e)
	return k == aig.Input || k == aig.And || k == aig.ExclusiveOr
}

func sswIslandIsNode(g *aig.Graph, e aig.Edge) bool {
	k := sswIslandKind(g, e)
	return k == aig.And || k == aig.ExclusiveOr
}

func sswIslandIsLo(g *aig.Graph, e aig.Edge) bool {
	return sswIslandKind(g, e) == aig.Input && aig.Aig_ObjCioId(g, e) >= saig.Saig_ManPiNum(g)
}

func sswIslandIsPi(g *aig.Graph, e aig.Edge) bool {
	return sswIslandKind(g, e) == aig.Input && aig.Aig_ObjCioId(g, e) < saig.Saig_ManPiNum(g)
}

func sswIslandIsPo(g *aig.Graph, e aig.Edge) bool {
	return sswIslandKind(g, e) == aig.Output && aig.Aig_ObjCioId(g, e) < saig.Saig_ManPoNum(g)
}

func sswIslandIsLi(g *aig.Graph, e aig.Edge) bool {
	return sswIslandKind(g, e) == aig.Output && aig.Aig_ObjCioId(g, e) >= saig.Saig_ManPoNum(g)
}

// Ssw_CreatePair establishes relationship between nodes using pairing.
func Ssw_CreatePair(p0, p1 *aig.Graph, vPairs *vec.Vec_Int_t, pObj0, pObj1 aig.Edge) {
	sswIslandSetData(p0, pObj0, pObj1)
	sswIslandSetData(p1, pObj1, pObj0)
	vec.Vec_IntPush(vPairs, int32(pObj0.NodeID()))
	vec.Vec_IntPush(vPairs, int32(pObj1.NodeID()))
}

// Ssw_MatchingStart establishes relationship between nodes using pairing.
func Ssw_MatchingStart(p0, p1 *aig.Graph, vPairs *vec.Vec_Int_t) {
	// create matching
	aig.Aig_ManCleanData(p0)
	aig.Aig_ManCleanData(p1)
	for i := 0; i < vec.Vec_IntSize(vPairs); i += 2 {
		pObj0 := aig.Aig_ManObj(p0, vec.Vec_IntEntry(vPairs, i))
		pObj1 := aig.Aig_ManObj(p1, vec.Vec_IntEntry(vPairs, i+1))
		if sswIslandData(p0, pObj0) != aig.InvalidEdge {
			panic("ssw: pObj0->pData == NULL")
		}
		if sswIslandData(p1, pObj1) != aig.InvalidEdge {
			panic("ssw: pObj1->pData == NULL")
		}
		sswIslandSetData(p0, pObj0, pObj1)
		sswIslandSetData(p1, pObj1, pObj0)
	}
	// make sure constants are matched
	pObj0 := aig.Aig_ManConst1(p0)
	pObj1 := aig.Aig_ManConst1(p1)
	if sswIslandData(p0, pObj0) != pObj1 || sswIslandData(p1, pObj1) != pObj0 {
		panic("ssw: constants are not matched")
	}
	// make sure PIs are matched
	for i := int32(0); i < saig.Saig_ManPiNum(p0); i++ {
		pObj0 = aig.Aig_ManCi(p0, i)
		pObj1 = aig.Aig_ManCi(p1, i)
		if sswIslandData(p0, pObj0) != pObj1 || sswIslandData(p1, pObj1) != pObj0 {
			panic("ssw: PIs are not matched")
		}
	}
	// make sure the POs are not matched
	for i := int32(0); i < aig.Aig_ManCoNum(p0); i++ {
		pObj0 = aig.Aig_ManCo(p0, i)
		pObj1 = aig.Aig_ManCo(p1, i)
		if sswIslandData(p0, pObj0) != aig.InvalidEdge || sswIslandData(p1, pObj1) != aig.InvalidEdge {
			panic("ssw: POs are matched")
		}
	}
	// check that LIs/LOs are matched in sync
	for i := int32(0); i < saig.Saig_ManRegNum(p0); i++ {
		pObj0 = saig.Saig_ManLo(p0, i)
		if pObj1 = sswIslandData(p0, pObj0); pObj1 == aig.InvalidEdge {
			continue
		}
		if !sswIslandIsLo(p1, pObj1) {
			util.Abc_Print(1, "Mismatch between LO pairs.\n")
		}
	}
	for i := int32(0); i < saig.Saig_ManRegNum(p1); i++ {
		pObj1 = saig.Saig_ManLo(p1, i)
		if pObj0 = sswIslandData(p1, pObj1); pObj0 == aig.InvalidEdge {
			continue
		}
		if !sswIslandIsLo(p0, pObj0) {
			util.Abc_Print(1, "Mismatch between LO pairs.\n")
		}
	}
}

// Ssw_MatchingExtendOne collects the matched neighbors of the unmatched
// objects (requires fanouts).
func Ssw_MatchingExtendOne(p *aig.Graph, vNodes *vec.Vec_Ptr_t) {
	vec.Vec_PtrClear(vNodes)
	aig.Aig_ManIncrementTravId(p)
	visit := func(pNext aig.Edge) {
		if sswIslandData(p, pNext) != aig.InvalidEdge && !aig.Aig_ObjIsTravIdCurrent(p, pNext) {
			aig.Aig_ObjSetTravIdCurrent(p, pNext)
			vec.Vec_PtrPush(vNodes, pNext)
		}
	}
	for _, pObj := range p.LiveEdges() {
		if !sswIslandIsCiOrNode(p, pObj) {
			continue
		}
		if sswIslandData(p, pObj) != aig.InvalidEdge {
			continue
		}
		if sswIslandIsLo(p, pObj) {
			pNext := aig.Aig_ObjFanin0(p, saig.Saig_ObjLoToLi(p, pObj))
			if sswIslandData(p, pNext) != aig.InvalidEdge && !aig.Aig_ObjIsTravIdCurrent(p, pNext) && !aig.Aig_ObjIsConst1(p, pNext) {
				aig.Aig_ObjSetTravIdCurrent(p, pNext)
				vec.Vec_PtrPush(vNodes, pNext)
			}
		}
		if sswIslandIsNode(p, pObj) {
			visit(aig.Aig_ObjFanin0(p, pObj))
			visit(aig.Aig_ObjFanin1(p, pObj))
		}
		nRefs := aig.Aig_ObjRefs(p, pObj)
		iFan := int32(-1)
		for k := int32(0); k < nRefs; k++ {
			if k == 0 {
				iFan = aig.Aig_ObjFanout0Int(p, int32(pObj.NodeID()))
			} else {
				iFan = aig.Aig_ObjFanoutNext(p, iFan)
			}
			pNext := aig.Aig_ManObj(p, iFan>>1)
			if sswIslandIsPo(p, pNext) {
				continue
			}
			if sswIslandIsLi(p, pNext) {
				pNext = saig.Saig_ObjLiToLo(p, pNext)
			}
			visit(pNext)
		}
	}
}

// Ssw_MatchingCountUnmached counts the unmatched CIs and nodes.
func Ssw_MatchingCountUnmached(p *aig.Graph) int32 {
	var Counter int32
	for _, pObj := range p.LiveEdges() {
		if !sswIslandIsCiOrNode(p, pObj) {
			continue
		}
		if sswIslandData(p, pObj) != aig.InvalidEdge {
			continue
		}
		Counter++
	}
	return Counter
}

func sswIslandPrintUnmatched(p0 *aig.Graph, d int32) {
	nUnmached := Ssw_MatchingCountUnmached(p0)
	nTotal := aig.Aig_ManCiNum(p0) + aig.Aig_ManNodeNum(p0)
	util.Abc_Print(1, "%2d : Total = %6d. Unmatched = %6d.  Ratio = %6.2f %%\n",
		d, nTotal, nUnmached, 100.0*float64(nUnmached)/float64(nTotal))
}

// Ssw_MatchingExtend removes the matches next to the unmatched objects
// nDist times.
func Ssw_MatchingExtend(p0, p1 *aig.Graph, nDist, fVerbose int32) {
	aig.Aig_ManFanoutStart(p0)
	aig.Aig_ManFanoutStart(p1)
	vNodes0 := vec.Vec_PtrAlloc(1000)
	vNodes1 := vec.Vec_PtrAlloc(1000)
	if fVerbose != 0 {
		util.Abc_Print(1, "Extending islands by %d steps:\n", nDist)
		sswIslandPrintUnmatched(p0, 0)
	}
	for d := int32(0); d < nDist; d++ {
		Ssw_MatchingExtendOne(p0, vNodes0)
		Ssw_MatchingExtendOne(p1, vNodes1)
		// (ABC passes the manager of pNext0 to Saig_ObjIsPi; both managers
		// have the same number of PIs)
		for _, side := range []struct {
			pa, pb *aig.Graph
			vNodes *vec.Vec_Ptr_t
		}{{p0, p1, vNodes0}, {p1, p0, vNodes1}} {
			for k := 0; k < vec.Vec_PtrSize(side.vNodes); k++ {
				pNext0 := vec.Vec_PtrEntry(side.vNodes, k).(aig.Edge)
				pNext1 := sswIslandData(side.pa, pNext0)
				if pNext1 == aig.InvalidEdge {
					continue
				}
				if sswIslandData(side.pb, pNext1) != pNext0 {
					panic("ssw: pNext1->pData == pNext0")
				}
				if sswIslandIsPi(side.pb, pNext1) {
					continue
				}
				sswIslandSetData(side.pa, pNext0, aig.InvalidEdge)
				sswIslandSetData(side.pb, pNext1, aig.InvalidEdge)
			}
		}
		if fVerbose != 0 {
			sswIslandPrintUnmatched(p0, d+1)
		}
	}
	aig.Aig_ManFanoutStop(p0)
	aig.Aig_ManFanoutStop(p1)
}

// Ssw_MatchingComplete uses differences in p0 to complete p1.
func Ssw_MatchingComplete(p0, p1 *aig.Graph) {
	// create register outputs in p0 that are absent in p1
	var vNewLis []aig.Edge
	for i := int32(0); i < saig.Saig_ManRegNum(p0); i++ {
		pObj0Li, pObj0 := saig.Saig_ManLi(p0, i), saig.Saig_ManLo(p0, i)
		if sswIslandData(p0, pObj0) != aig.InvalidEdge {
			continue
		}
		pObj1 := aig.Aig_ObjCreateCi(p1)
		sswIslandSetData(p0, pObj0, pObj1)
		sswIslandSetData(p1, pObj1, pObj0)
		vNewLis = append(vNewLis, pObj0Li)
	}
	// add missing nodes in the topological order
	for _, pObj0 := range p0.LiveEdges() {
		if !sswIslandIsNode(p0, pObj0) {
			continue
		}
		if sswIslandData(p0, pObj0) != aig.InvalidEdge {
			continue
		}
		_, f0, f1 := p0.KindFanins(pObj0)
		pObj1 := aig.Aig_And(p1, sswIslandChildCopy(p0, f0), sswIslandChildCopy(p0, f1))
		sswIslandSetData(p0, pObj0, pObj1)
		sswIslandSetData(p1, pObj1, pObj0)
	}
	// create register outputs in p0 that are absent in p1
	for _, pObj0Li := range vNewLis {
		_, f0, _ := p0.KindFanins(pObj0Li)
		aig.Aig_ObjCreateCo(p1, sswIslandChildCopy(p0, f0))
	}
	// increment the number of registers
	aig.Aig_ManSetRegNum(p1, aig.Aig_ManRegNum(p1)+int32(len(vNewLis)))
}

// Ssw_MatchingPairs derives matching for all pairs.
func Ssw_MatchingPairs(p0, p1 *aig.Graph) *vec.Vec_Int_t {
	// check correctness
	if aig.Aig_ManCiNum(p0) != aig.Aig_ManCiNum(p1) || aig.Aig_ManCoNum(p0) != aig.Aig_ManCoNum(p1) ||
		aig.Aig_ManRegNum(p0) != aig.Aig_ManRegNum(p1) || aig.Aig_ManObjNum(p0) != aig.Aig_ManObjNum(p1) {
		panic("ssw: incomplete matching")
	}
	// create complete pairs
	vPairsNew := vec.Vec_IntAlloc(int(2 * aig.Aig_ManObjNum(p0)))
	for _, pObj0 := range p0.LiveEdges() {
		if sswIslandKind(p0, pObj0) == aig.Output {
			continue
		}
		pObj1 := sswIslandData(p0, pObj0)
		if pObj1 == aig.InvalidEdge {
			panic("ssw: missing match")
		}
		vec.Vec_IntPush(vPairsNew, int32(pObj0.NodeID()))
		vec.Vec_IntPush(vPairsNew, int32(pObj1.NodeID()))
	}
	return vPairsNew
}

// sswIslandCreateMiter is Saig_ManCreateMiter( p0, p1, 0 ), which leaves
// each object's miter copy in its pData.
func sswIslandCreateMiter(p0, p1 *aig.Graph) *aig.Graph {
	pMiter, c0, c1 := saig.Saig_ManCreateMiterWithCopies(p0, p1, 0)
	for _, side := range []struct {
		p *aig.Graph
		c []aig.Edge
	}{{p0, c0}, {p1, c1}} {
		aig.Aig_ManCleanData(side.p)
		for _, e := range side.p.LiveEdges() {
			if sswIslandKind(side.p, e) == aig.Output {
				continue
			}
			if c := side.c[e.NodeID()]; c != aig.InvalidEdge {
				side.p.SetObjectData(e, c)
			}
		}
	}
	return pMiter
}

// Ssw_MatchingMiter transfers the result of matching to the miter.
func Ssw_MatchingMiter(pMiter, p0, p1 *aig.Graph, vPairsAll *vec.Vec_Int_t) *vec.Vec_Int_t {
	// create matching of nodes in the miter
	vPairsMiter := vec.Vec_IntAlloc(int(2 * aig.Aig_ManObjNum(p0)))
	for i := 0; i < vec.Vec_IntSize(vPairsAll); i += 2 {
		pObj0 := aig.Aig_ManObj(p0, vec.Vec_IntEntry(vPairsAll, i))
		pObj1 := aig.Aig_ManObj(p1, vec.Vec_IntEntry(vPairsAll, i+1))
		d0, d1 := sswIslandData(p0, pObj0), sswIslandData(p1, pObj1)
		if d0 == aig.InvalidEdge || d1 == aig.InvalidEdge {
			panic("ssw: pObj->pData != NULL")
		}
		if d0 == d1 {
			continue
		}
		if aig.Aig_ObjIsNone(pMiter, d0) || aig.Aig_ObjIsNone(pMiter, d1) {
			continue
		}
		// get the miter nodes
		pObj0, pObj1 = d0, d1
		if aig.Aig_IsComplement(pObj0) != 0 || aig.Aig_IsComplement(pObj1) != 0 {
			panic("ssw: !Aig_IsComplement(pObj)")
		}
		if aig.Aig_ObjType(pMiter, pObj0) != aig.Aig_ObjType(pMiter, pObj1) {
			panic("ssw: Aig_ObjType(pObj0) == Aig_ObjType(pObj1)")
		}
		if sswIslandKind(pMiter, pObj0) == aig.Output {
			continue
		}
		if pObj0.NodeID() >= pObj1.NodeID() {
			panic("ssw: pObj0->Id < pObj1->Id")
		}
		vec.Vec_IntPush(vPairsMiter, int32(pObj0.NodeID()))
		vec.Vec_IntPush(vPairsMiter, int32(pObj1.NodeID()))
	}
	return vPairsMiter
}

// Ssw_SecWithSimilaritySweep solves SEC using structural similarity; it
// modifies both p0 and p1 by adding extra logic.
func Ssw_SecWithSimilaritySweep(p0, p1 *aig.Graph, vPairs *vec.Vec_Int_t, pPars *Ssw_Pars_t) *aig.Graph {
	// derive full matching
	Ssw_MatchingStart(p0, p1, vPairs)
	if pPars.NIsleDist != 0 {
		Ssw_MatchingExtend(p0, p1, pPars.NIsleDist, pPars.FVerbose)
	}
	Ssw_MatchingComplete(p0, p1)
	Ssw_MatchingComplete(p1, p0)
	vPairsAll := Ssw_MatchingPairs(p0, p1)
	// create miter and transfer matching
	pMiter := sswIslandCreateMiter(p0, p1)
	vPairsMiter := Ssw_MatchingMiter(pMiter, p0, p1, vPairsAll)
	// start the induction manager
	p := Ssw_ManCreate(pMiter, pPars)
	// create equivalence classes using these IDs
	if p.pPars.FPartSigCorr != 0 {
		p.ppClasses = Ssw_ClassesPreparePairsSimple(pMiter, vPairsMiter)
	} else {
		p.ppClasses = Ssw_ClassesPrepare(pMiter, pPars.NFramesK, pPars.FLatchCorr, pPars.FConstCorr, pPars.FOutputCorr, pPars.NMaxLevs, pPars.FVerbose)
	}
	if p.pPars.FDumpSRInit != 0 {
		if p.pPars.FPartSigCorr != 0 {
			pSRed := Ssw_SpeculativeReduction(p)
			if err := aig.Aig_ManDumpBlif(pSRed, "srm_part.blif", nil, nil); err != nil {
				panic(err)
			}
			aig.Aig_ManStop(pSRed)
			util.Abc_Print(1, "Speculatively reduced miter is saved in file \"%s\".\n", "srm_part.blif")
		} else {
			util.Abc_Print(1, "Dumping speculative miter is possible only for partial signal correspondence (switch \"-c\").\n")
		}
	}
	p.pSml = Ssw_SmlStart(pMiter, 0, 1+p.pPars.NFramesAddSim, 1)
	Ssw_ClassesSetData(p.ppClasses, p.pSml, func(data any, e aig.Edge) uint32 { return Ssw_SmlObjHashWord(data.(*Ssw_Sml_t), e) }, func(data any, e aig.Edge) int32 { return Ssw_SmlObjIsConstWord(data.(*Ssw_Sml_t), e) }, func(data any, a, b aig.Edge) int32 { return Ssw_SmlObjsAreEqualWord(data.(*Ssw_Sml_t), a, b) })
	// perform refinement of classes
	pAigNew := Ssw_SignalCorrespondenceRefine(p)
	// cleanup
	Ssw_ManStop(p)
	aig.Aig_ManStop(pMiter)
	return pAigNew
}

// Ssw_SecWithSimilarityPairs solves SEC with structural similarity given the
// pairs of IDs of structurally equivalent nodes of p0 and p1; the managers
// are updated by adding "islands of difference".
func Ssw_SecWithSimilarityPairs(p0, p1 *aig.Graph, vPairs *vec.Vec_Int_t, pPars *Ssw_Pars_t) int32 {
	clk := util.Abc_Clock()
	// derive parameters if not given
	if pPars == nil {
		pPars = new(Ssw_Pars_t)
		Ssw_ManSetDefaultParams(pPars)
	}
	// reduce the AIG with pairs
	pAigRes := Ssw_SecWithSimilaritySweep(p0, p1, vPairs, pPars)
	// report the result of verification
	RetValue := Ssw_MiterStatus(pAigRes, 1)
	if RetValue == 1 {
		util.Abc_Print(1, "Verification successful.  ")
	} else if RetValue == 0 {
		util.Abc_Print(1, "Verification failed with a counter-example.  ")
	} else {
		util.Abc_Print(1, "Verification UNDECIDED. The number of remaining regs = %d (total = %d).  ",
			aig.Aig_ManRegNum(pAigRes), aig.Aig_ManRegNum(p0)+aig.Aig_ManRegNum(p1))
	}
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	aig.Aig_ManStop(pAigRes)
	return RetValue
}

// Saig_StrSimPerformMatching_hack is a dummy procedure to detect structural
// similarity: it pairs every constant, CI and node with the same ID.
func Saig_StrSimPerformMatching_hack(p0, p1 *aig.Graph) *vec.Vec_Int_t {
	// create array of pairs
	vPairs := vec.Vec_IntAlloc(100)
	for _, pObj := range p0.LiveEdges() {
		if !aig.Aig_ObjIsConst1(p0, pObj) && !sswIslandIsCiOrNode(p0, pObj) {
			continue
		}
		i := int32(pObj.NodeID())
		vec.Vec_IntPush(vPairs, i)
		vec.Vec_IntPush(vPairs, i)
	}
	return vPairs
}

// Ssw_SecWithSimilarity solves SEC with structural similarity; with
// p1 == nil, p0 is a miter to be demitered.
func Ssw_SecWithSimilarity(p0, p1 *aig.Graph, pPars *Ssw_Pars_t) int32 {
	var pPart0, pPart1 *aig.Graph
	if pPars == nil {
		pPars = new(Ssw_Pars_t)
		Ssw_ManSetDefaultParams(pPars)
	}
	if pPars.FVerbose != 0 {
		util.Abc_Print(1, "Performing sequential verification using structural similarity.\n")
	}
	// consider the case when a miter is given
	if p1 == nil {
		if pPars.FVerbose != 0 {
			aig.Aig_ManPrintStats(p0)
		}
		// demiter the miter
		if saig.Saig_ManDemiterSimpleDiff(p0, &pPart0, &pPart1) == 0 {
			util.Abc_Print(1, "Demitering has failed.\n")
			return -1
		}
	} else {
		pPart0 = aig.Aig_ManDupSimple(p0)
		pPart1 = aig.Aig_ManDupSimple(p1)
	}
	if aig.Aig_ManRegNum(pPart0) <= 0 || aig.Aig_ManRegNum(pPart1) <= 0 ||
		saig.Saig_ManPiNum(pPart0) != saig.Saig_ManPiNum(pPart1) || saig.Saig_ManPoNum(pPart0) != saig.Saig_ManPoNum(pPart1) {
		panic("ssw: incompatible sequential graphs")
	}
	// derive pairs
	//    vPairs = Saig_StrSimPerformMatching_hack( pPart0, pPart1 );
	vPairs := saig.Saig_StrSimPerformMatching(pPart0, pPart1, 0, pPars.FVerbose, nil)
	RetValue := Ssw_SecWithSimilarityPairs(pPart0, pPart1, vPairs, pPars)
	aig.Aig_ManStop(pPart0)
	aig.Aig_ManStop(pPart1)
	return RetValue
}
