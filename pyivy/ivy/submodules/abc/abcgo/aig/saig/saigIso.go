package saig

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/misc/util"
)

// Port of src/aig/saig/saigIso.c: canonical forms of sequential AIGs and
// the reduction of isomorphic primary outputs.

// isoPermCos is Saig_ManFindIsoPermCos given the objects' canonical numbers
// (C reads them from pObj->iData): POs by the literal of their canonical
// driver, then the flop inputs in the flop outputs' canonical order.
func isoPermCos(pAig *aig.Graph, vPermCis, iData []int32) []int32 {
	if len(vPermCis) != int(aig.Aig_ManCiNum(pAig)) {
		panic("saig: CI permutation has wrong length")
	}
	nPos, nPis := Saig_ManPoNum(pAig), Saig_ManPiNum(pAig)
	vPermCos := make([]int32, 0, aig.Aig_ManCoNum(pAig))
	if nPos == 1 {
		vPermCos = append(vPermCos, 0)
	} else {
		roots := make([]int32, nPos)
		key := make([]int32, nPos)
		for i := int32(0); i < nPos; i++ {
			po := aig.Aig_ManCo(pAig, i)
			fanin := aig.Aig_ObjFaninId0(pAig, po)
			if fanin != 0 && iData[fanin] <= 0 {
				panic("saig: PO driver without a canonical number")
			}
			roots[i], key[i] = i, 2*iData[fanin]+aig.Aig_ObjFaninC0(pAig, po)
		}
		// POs with equal keys have the same driver literal and write
		// identical lines, so the tie order cannot be observed
		sort.SliceStable(roots, func(a, b int) bool { return key[roots[a]] < key[roots[b]] })
		vPermCos = append(vPermCos, roots...)
	}
	diff := nPos - nPis
	for _, entry := range vPermCis[nPis:] {
		vPermCos = append(vPermCos, entry+diff)
	}
	return vPermCos
}

// Saig_ManFindIsoPermCos orders the COs canonically given vPermCis from
// Saig_ManFindIsoPerm (whose canonical numbers it recomputes).
func Saig_ManFindIsoPermCos(pAig *aig.Graph, vPermCis []int32) []int32 {
	_, iData := isoFindPerm(pAig, 0)
	return isoPermCos(pAig, vPermCis, iData)
}

// Saig_ManDupIsoCanonical is saigIso.c's: CIs in canonical order, nodes in
// a DFS from the canonically ordered COs that visits a node's fanins by
// their canonical numbers, then the COs.
func Saig_ManDupIsoCanonical(pAig *aig.Graph, verbose int32) *aig.Graph {
	vPerm, iData := isoFindPerm(pAig, verbose)
	vPermCo := isoPermCos(pAig, vPerm, iData)
	pNew := aig.Aig_ManStart(aig.Aig_ManNodeNum(pAig))
	pNew.Name = pAig.Name
	aig.Aig_ManIncrementTravId(pAig)
	const1 := aig.Aig_ManConst1(pAig)
	aig.Aig_ObjSetCopy(pAig, const1, aig.Aig_ManConst1(pNew))
	aig.Aig_ObjSetTravIdCurrent(pAig, const1)
	for _, entry := range vPerm {
		ci := aig.Aig_ManCi(pAig, entry)
		aig.Aig_ObjSetCopy(pAig, ci, aig.Aig_ObjCreateCi(pNew))
		aig.Aig_ObjSetTravIdCurrent(pAig, ci)
	}
	var rec func(e aig.Edge)
	rec = func(e aig.Edge) {
		if aig.Aig_ObjIsTravIdCurrent(pAig, e) {
			return
		}
		aig.Aig_ObjSetTravIdCurrent(pAig, e)
		if !aig.Aig_ObjIsNode(pAig, e) {
			panic("saig: canonical DFS reached a non-node")
		}
		f0, f1 := aig.Aig_ObjFanin0(pAig, e), aig.Aig_ObjFanin1(pAig, e)
		if aig.Aig_ObjIsNode(pAig, f0) && aig.Aig_ObjIsNode(pAig, f1) && iData[f1.NodeID()] < iData[f0.NodeID()] {
			rec(f1)
			rec(f0)
		} else {
			rec(f0)
			rec(f1)
		}
		aig.Aig_ObjSetCopy(pAig, e, aig.Aig_And(pNew, aig.Aig_ObjChild0Copy(pAig, e), aig.Aig_ObjChild1Copy(pAig, e)))
	}
	for _, entry := range vPermCo {
		rec(aig.Aig_ObjFanin0(pAig, aig.Aig_ManCo(pAig, entry)))
	}
	for _, entry := range vPermCo {
		aig.Aig_ObjCreateCo(pNew, aig.Aig_ObjChild0Copy(pAig, aig.Aig_ManCo(pAig, entry)))
	}
	aig.Aig_ManSetRegNum(pNew, aig.Aig_ManRegNum(pAig))
	return pNew
}

// Iso_ManCheckMapping reports whether vMap2to1 (CI i of pAig2 to CI
// vMap2to1[i] of pAig1) maps every node of pAig2, in ID order, onto an
// existing node of pAig1, and, for single-output AIGs, the PO driver too.
func Iso_ManCheckMapping(pAig1, pAig2 *aig.Graph, vMap2to1 []int32, verbose int32) int32 {
	if aig.Aig_ManCiNum(pAig1) != aig.Aig_ManCiNum(pAig2) || aig.Aig_ManCoNum(pAig1) != aig.Aig_ManCoNum(pAig2) ||
		aig.Aig_ManRegNum(pAig1) != aig.Aig_ManRegNum(pAig2) || aig.Aig_ManNodeNum(pAig1) != aig.Aig_ManNodeNum(pAig2) {
		panic("saig: Iso_ManCheckMapping on AIGs of different sizes")
	}
	aig.Aig_ManCleanData(pAig2)
	aig.Aig_ObjSetCopy(pAig2, aig.Aig_ManConst1(pAig2), aig.Aig_ManConst1(pAig1))
	for i := int32(0); i < aig.Aig_ManCiNum(pAig2); i++ {
		aig.Aig_ObjSetCopy(pAig2, aig.Aig_ManCi(pAig2, i), aig.Aig_ManCi(pAig1, vMap2to1[i]))
	}
	for _, i := range isoObjIDs(pAig2) {
		e := aig.Aig_ManObj(pAig2, i)
		if !aig.Aig_ObjIsNode(pAig2, e) {
			continue
		}
		found := aig.Aig_TableLookupTwo(pAig1, aig.Aig_ObjChild0Copy(pAig2, e), aig.Aig_ObjChild1Copy(pAig2, e))
		if found == aig.InvalidEdge {
			if verbose != 0 {
				fmt.Printf("Structural equivalence failed at node %d.\n", i)
			}
			return 0
		}
		aig.Aig_ObjSetCopy(pAig2, e, found)
	}
	if aig.Aig_ManCoNum(pAig1)-aig.Aig_ManRegNum(pAig1) == 1 &&
		aig.Aig_ObjChild0Copy(pAig2, aig.Aig_ManCo(pAig2, 0)) != aig.Aig_ObjChild0(pAig1, aig.Aig_ManCo(pAig1, 0)) {
		if verbose != 0 {
			fmt.Printf("Structural equivalence failed at primary output 0.\n")
		}
		return 0
	}
	return 1
}

// Iso_ManNegEdgeNum counts the complemented fanin edges of nodes and COs.
func Iso_ManNegEdgeNum(p *aig.Graph) int32 {
	var count int32
	for _, i := range isoObjIDs(p) {
		e := aig.Aig_ManObj(p, i)
		if aig.Aig_ObjIsNode(p, e) {
			count += aig.Aig_ObjFaninC0(p, e) + aig.Aig_ObjFaninC1(p, e)
		} else if aig.Aig_ObjIsCo(p, e) {
			count += aig.Aig_ObjFaninC0(p, e)
		}
	}
	return count
}

// Iso_ManFindMapping returns the CI map from pAig2 to pAig1 under which the
// AIGs are structurally identical, or nil.
func Iso_ManFindMapping(pAig1, pAig2 *aig.Graph, vPerm1_, vPerm2_ []int32, verbose int32) []int32 {
	if aig.Aig_ManCiNum(pAig1) != aig.Aig_ManCiNum(pAig2) || aig.Aig_ManCoNum(pAig1) != aig.Aig_ManCoNum(pAig2) ||
		aig.Aig_ManRegNum(pAig1) != aig.Aig_ManRegNum(pAig2) || aig.Aig_ManNodeNum(pAig1) != aig.Aig_ManNodeNum(pAig2) ||
		aig.Aig_ManLevelNum(pAig1) != aig.Aig_ManLevelNum(pAig2) {
		return nil
	}
	if verbose != 0 {
		fmt.Printf("AIG1:\n")
	}
	vPerm1 := vPerm1_
	if vPerm1 == nil {
		vPerm1 = Saig_ManFindIsoPerm(pAig1, verbose)
	}
	if verbose != 0 {
		fmt.Printf("AIG1:\n")
	}
	vPerm2 := vPerm2_
	if vPerm2 == nil {
		vPerm2 = Saig_ManFindIsoPerm(pAig2, verbose)
	}
	if len(vPerm1) != int(aig.Aig_ManCiNum(pAig1)) || len(vPerm2) != int(aig.Aig_ManCiNum(pAig2)) {
		panic("saig: CI permutation has wrong length")
	}
	// vInvPerm2[ci of AIG2] = its canonical position, then the AIG1 CI there
	vInvPerm2 := make([]int32, len(vPerm2))
	for i := range vInvPerm2 {
		vInvPerm2[i] = -1
	}
	for canonical, ci := range vPerm2 {
		vInvPerm2[ci] = int32(canonical)
	}
	for i, entry := range vInvPerm2 {
		if entry < 0 || entry >= aig.Aig_ManCiNum(pAig1) {
			panic("saig: CI permutation is not a permutation")
		}
		vInvPerm2[i] = vPerm1[entry]
	}
	if Iso_ManCheckMapping(pAig1, pAig2, vInvPerm2, verbose) == 0 {
		return nil
	}
	return vInvPerm2
}

// Iso_ManFilterPos_old keeps one PO of each class of POs with isomorphic
// cones, checking candidates pairwise with Iso_ManFindMapping.
func Iso_ManFilterPos_old(pAig *aig.Graph, verbose int32) *aig.Graph {
	nPos := aig.Aig_ManCoNum(pAig) - aig.Aig_ManRegNum(pAig)
	parts := make([]*aig.Graph, nPos)
	perms := make([][]int32, nPos)
	for i := int32(0); i < nPos; i++ {
		parts[i] = Saig_ManDupCones(pAig, []int32{i}, 1)
		perms[i] = Saig_ManFindIsoPerm(parts[i], 0)
	}
	var kept []int32
	for i, part := range parts {
		found := false
		for _, k := range kept {
			if Iso_ManFindMapping(parts[k], part, perms[k], perms[i], 0) != nil {
				found = true
				break
			}
		}
		if !found {
			kept = append(kept, int32(i))
		}
	}
	return Saig_ManDupCones(pAig, kept, int32(len(kept)))
}

// Iso_StoCompareVecStr is Vec_StrCompareVec: by size, then by bytes.
func Iso_StoCompareVecStr(a, b []byte) int {
	if len(a) != len(b) {
		return len(a) - len(b)
	}
	return bytes.Compare(a, b)
}

// Iso_ManFilterPos is saigIso.c's: each PO's cone is put in canonical form
// and written as AIGER; POs with identical bytes form a class. It returns
// the AIG of the first PO of each class (classes ordered by it) and the
// classes.
func Iso_ManFilterPos(pAig *aig.Graph, verbose int32) (*aig.Graph, [][]int32) {
	nPos := aig.Aig_ManCoNum(pAig) - aig.Aig_ManRegNum(pAig)
	type buffer struct {
		data []byte
		po   int32
	}
	buffers := make([]buffer, 0, nPos)
	var clkDup, clkIso, clkAig time.Duration
	for i := int32(0); i < nPos; i++ {
		if i%100 == 0 {
			fmt.Printf("%6d finished...\r", i)
		}
		clk := time.Now()
		part := Saig_ManDupCones(pAig, []int32{i}, 1)
		clkDup += time.Since(clk)
		clk = time.Now()
		canon := Saig_ManDupIsoCanonical(part, 0)
		clkIso += time.Since(clk)
		clk = time.Now()
		data, err := ioa.Ioa_WriteAigerIntoMemoryStr(canon)
		if err != nil {
			panic(err)
		}
		clkAig += time.Since(clk)
		buffers = append(buffers, buffer{data, i})
	}
	if verbose != 0 {
		util.Abc_PrintTime(1, "Duplicate time", clkDup.Nanoseconds())
		util.Abc_PrintTime(1, "Isomorph  time", clkIso.Nanoseconds())
		util.Abc_PrintTime(1, "AIGER     time", clkAig.Nanoseconds())
	}
	clk := time.Now()
	sort.SliceStable(buffers, func(a, b int) bool { return Iso_StoCompareVecStr(buffers[a].data, buffers[b].data) < 0 })
	var classes [][]int32
	if len(buffers) > 0 {
		classes = append(classes, []int32{buffers[0].po})
		for i := 1; i < len(buffers); i++ {
			if Iso_StoCompareVecStr(buffers[i-1].data, buffers[i].data) != 0 {
				classes = append(classes, nil)
			}
			classes[len(classes)-1] = append(classes[len(classes)-1], buffers[i].po)
		}
	}
	if verbose != 0 {
		util.Abc_PrintTime(1, "Sorting   time", time.Since(clk).Nanoseconds())
	}
	// canonicize the order: each class sorted, classes by their first PO
	for _, class := range classes {
		sort.Slice(class, func(a, b int) bool { return class[a] < class[b] })
	}
	sort.Slice(classes, func(a, b int) bool { return classes[a][0] < classes[b][0] })
	remain := make([]int32, len(classes))
	for i, class := range classes {
		remain[i] = class[0]
	}
	return Saig_ManDupCones(pAig, remain, int32(len(remain))), classes
}

// Saig_ManIsoReduce keeps one PO per class of isomorphic POs and reports
// the reduction.
func Saig_ManIsoReduce(pAig *aig.Graph, verbose int32) (*aig.Graph, [][]int32) {
	start := time.Now()
	part, classes := Iso_ManFilterPos(pAig, verbose)
	fmt.Printf("Reduced %d outputs to %d outputs.  ", Saig_ManPoNum(pAig), Saig_ManPoNum(part))
	util.Abc_PrintTime(1, "Time", time.Since(start).Nanoseconds())
	if verbose != 0 && classes != nil && int(Saig_ManPoNum(pAig)) != len(classes) {
		fmt.Printf("Nontrivial classes:\n")
		isoPrintClasses(classes)
	}
	return part, classes
}

// isoPrintClasses is Vec_VecPrintInt( vClasses, 1 ): the classes with more
// than one entry.
func isoPrintClasses(classes [][]int32) {
	for i, class := range classes {
		if len(class) == 1 {
			continue
		}
		fmt.Printf(" %4d : {", i)
		for _, po := range class {
			fmt.Printf(" %d", po)
		}
		fmt.Printf(" }\n")
	}
}

// Iso_ManTest times Saig_ManFindIsoPerm.
func Iso_ManTest(p *aig.Graph, verbose int32) *aig.Graph {
	start := time.Now()
	_ = Saig_ManFindIsoPerm(p, verbose)
	util.Abc_PrintTime(1, "Time", time.Since(start).Nanoseconds())
	return nil
}

// Iso_ManTest888 converted the AIG to a network, permuted it and searched
// for the mapping back; the graph package has no network conversion, so it
// keeps only the original nil result.
func Iso_ManTest888(p *aig.Graph, verbose int32) *aig.Graph {
	_ = p
	_ = verbose
	return nil
}
