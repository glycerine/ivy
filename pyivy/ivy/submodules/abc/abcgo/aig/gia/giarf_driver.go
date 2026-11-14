package gia

import (
	"fmt"
	"time"
)

// Gia_ComputeEquivalencesLevel resolves constant and equivalence candidates
// among the roots of one level using the selected SAT backend.
func Gia_ComputeEquivalencesLevel(p *Hcd_Man_t, reduced *Gia_Man_t, roots *Vec_Ptr_t, level, useMiniSat int32) int32 {
	if Vec_PtrSize(roots) == 0 {
		return 0
	}
	start := time.Now()
	solver := newGiarfSolver(reduced, p.nBTLimit, useMiniSat != 0)
	defer solver.close()
	Vec_PtrCleanSimInfo(p.vSimPres, 0, 1)
	var saved, records, undecided, refinements, queries int32
	check := func(a int32, b *int32) (int32, *Vec_Int_t) {
		queries++
		status, cex, err := solver.check(a, b)
		if err != nil {
			panic(fmt.Sprintf("gia: equivalence SAT: %v", err))
		}
		return status, cex
	}
	g := p.pGia
	for i := 0; i < Vec_PtrSize(roots); i++ {
		root := Vec_PtrEntry(roots, i).(*Gia_Obj_t)
		id := Gia_ObjId(g, root)
		if Gia_ObjIsConst(g, id) == 0 {
			continue
		}
		lit := int32(root.Value ^ root.fPhase)
		status, cex := check(lit, nil)
		switch status {
		case -1:
			undecided++
			Gia_ObjSetFailed(g, id)
		case 1:
			Gia_ObjSetProved(g, id)
		case 0:
			Gia_ObjUnsetRepr(g, id)
			records++
			saved += Gia_GiarfStorePattern(p.vSimInfo, p.vSimPres, cex)
		}
	}
	members := Vec_PtrAlloc(100)
	var iterations int32
	for Vec_PtrSize(roots) > 0 {
		seen := make(map[int32]bool)
		next := Vec_PtrAlloc(Vec_PtrSize(roots))
		for i := 0; i < Vec_PtrSize(roots); i++ {
			root := Vec_PtrEntry(roots, i).(*Gia_Obj_t)
			id := Gia_ObjId(g, root)
			var repr *Gia_Obj_t
			if Gia_ObjIsHead(g, id) != 0 {
				repr = root
			} else if Gia_ObjIsClass(g, id) != 0 {
				repr = Gia_ObjReprObj(g, id)
			} else {
				continue
			}
			reprID := Gia_ObjId(g, repr)
			if seen[reprID] {
				continue
			}
			seen[reprID] = true
			temp := Gia_CollectClassMembers(g, repr, members, level)
			if temp != nil {
				Vec_PtrPush(members, temp)
			}
			if Vec_PtrSize(members) < 2 {
				continue
			}
			failed, pending := false, false
			prev := Vec_PtrEntry(members, Vec_PtrSize(members)-1).(*Gia_Obj_t)
			for k := 0; k < Vec_PtrSize(members); k++ {
				member := Vec_PtrEntry(members, k).(*Gia_Obj_t)
				a := int32(prev.Value ^ prev.fPhase)
				b := int32(member.Value ^ (member.fPhase ^ 1))
				status, cex := check(a, &b)
				switch status {
				case -1:
					pending = true
					undecided++
					Gia_ObjSetFailed(g, prev.id)
					Gia_ObjSetFailed(g, member.id)
				case 0:
					failed = true
					records++
					saved += Gia_GiarfStorePattern(p.vSimInfo, p.vSimPres, cex)
					Gia_GiarfInsertPattern(p, cex, int32(k%31+1))
				}
				prev = member
			}
			if failed {
				refinements++
				for k := 0; k < Vec_PtrSize(members); k++ {
					member := Vec_PtrEntry(members, k).(*Gia_Obj_t)
					if member != temp && Gia_ObjFailed(g, member.id) == 0 {
						Vec_PtrPush(next, member)
					}
				}
				Gia_ResimulateAndRefine(p, reprID)
			} else if !pending {
				for k := 0; k < Vec_PtrSize(members); k++ {
					Gia_ObjSetProved(g, Vec_PtrEntry(members, k).(*Gia_Obj_t).id)
				}
			}
		}
		Vec_PtrClear(roots)
		for i := 0; i < Vec_PtrSize(next); i++ {
			Vec_PtrPush(roots, Vec_PtrEntry(next, i))
		}
		iterations++
	}
	fmt.Printf("nSaved = %d   nRecords = %d   nUndec = %d   nClassRefs = %d   queries = %d   time = %s\n", saved, records, undecided, refinements, queries, time.Since(start))
	Hcd_ManSimulateSimple(p)
	Hcd_ManClassesRefine(p)
	return iterations
}

func Gia_ComputeEquivalences(g *Gia_Man_t, limit, useMiniSat, verbose int32) {
	Gia_ManRandom(1)
	Gia_ManSetPhase(g)
	levels := Gia_ManLevelNum(g)
	Gia_ManIncrementTravId(g)
	p := Gia_ManEquivStart(g, limit, verbose)
	defer Gia_ManEquivStop(p)
	Hcd_ManClassesCreate(p)
	for i := 0; i < 3; i++ {
		Hcd_ManSimulationInit(p)
		Hcd_ManSimulateSimple(p)
		Hcd_ManClassesRefine(p)
	}
	// The source's strict bound skips the deepest gates, leaving simulation
	// candidates there unverified before they are used to rewrite the graph.
	for level := int32(1); level <= levels; level++ {
		start := time.Now()
		var roots *Vec_Ptr_t
		reduced := Gia_GenerateReducedLevel(g, level, &roots)
		iterations := Gia_ComputeEquivalencesLevel(p, reduced, roots, level, useMiniSat)
		Gia_ManStop(reduced)
		fmt.Printf("LEVEL %3d (out of %3d) Iters = %3d Time = %s\n", level, levels, iterations, time.Since(start))
	}
	giaDropUnprovedEquivalences(g)
	Gia_GiarfPrintClasses(g)
}

// Simulation proposes candidates; only successful SAT checks justify
// substituting them. Preserve original logic after an unknown result or
// a final simulation refinement that creates an unchecked candidate.
func giaDropUnprovedEquivalences(g *Gia_Man_t) {
	proved := func(id int32) bool {
		return id == 0 || Gia_ObjProved(g, id) != 0 && Gia_ObjFailed(g, id) == 0
	}
	changed := false
	for id := int32(1); id < g.nObjs; id++ {
		repr := Gia_ObjRepr(g, id)
		// A member may have been proved only against a temporary lower-level
		// member. Its actual class head must have passed checking as well.
		if repr != GIA_VOID && (!proved(id) || !proved(repr)) {
			Gia_ObjUnsetRepr(g, id)
			changed = true
		}
	}
	if changed {
		g.PNexts = Gia_ManDeriveNexts(g)
	}
}
