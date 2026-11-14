package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

func Mf_CutPrint(p *Mf_Man_t, c *Mf_Cut_t) {
	digits := len(fmt.Sprint(p.pGia.nObjs))
	fmt.Printf("%d  {", c.nLeaves)
	for _, id := range c.pLeaves[:c.nLeaves] {
		fmt.Printf(" %*d", digits, id)
	}
	for i := int32(c.nLeaves); i < p.pPars.NLutSize; i++ {
		fmt.Printf(" %*s", digits, "")
	}
	fmt.Printf("  }   D = %4d  A = %9.4f  F = %6d  ", c.Delay, c.Flow, c.iFunc)
	if p.vTtMem == nil {
		fmt.Println()
		return
	}
	if p.pPars.FGenCnf != 0 {
		fmt.Printf("CNF = %2d  ", vec.Vec_IntEntry(&p.vCnfSizes, int(c.iFunc>>1)))
	}
	if p.pPars.FGenLit != 0 {
		fmt.Printf("Lit = %2d  ", vec.Vec_IntEntry(&p.vCnfSizes, int(c.iFunc>>1)))
	}
	dau.Dau_DsdPrintFromTruth(vec.Vec_MemReadEntry(p.vTtMem, int(c.iFunc>>1)), int32(c.nLeaves))
}
func Mf_SetCheckArray(cuts []*Mf_Cut_t, n int32) int32 {
	if n <= 0 {
		panic("gia: empty cut set")
	}
	for _, a := range cuts[:n] {
		if a.nLeaves > MF_LEAF_MAX || a.Sign != Mf_CutGetSign(a.pLeaves[:], int32(a.nLeaves)) {
			panic("gia: invalid cut")
		}
		for i := uint32(1); i < a.nLeaves; i++ {
			if a.pLeaves[i-1] >= a.pLeaves[i] {
				panic("gia: unordered cut")
			}
		}
		for _, b := range cuts[:n] {
			if a != b && Mf_CutCheck(a, b) != 0 {
				panic("gia: dominated cut")
			}
		}
	}
	return 1
}
func Mf_ManPrintFanoutProfile(p *Mf_Man_t, fanCounts *Vec_Int_t) {
	counts := map[int32]int32{}
	var maxFan int32
	for id, n := range vec.Vec_IntArray(fanCounts) {
		maxFan = max(maxFan, n)
		if n != 0 && Gia_ObjIsAnd(Gia_ManObj(p.pGia0, int32(id))) != 0 {
			counts[n]++
		}
	}
	fmt.Println("\nFanout distribution for internal nodes:")
	for n := int32(0); n <= maxFan; n++ {
		if counts[n] != 0 {
			fmt.Printf("Fanout = %5d : Nodes = %5d.\n", n, counts[n])
		}
	}
	fmt.Printf("Total nodes with fanout = %d. Max fanout = %d.\n\n", len(counts), maxFan)
}
func Mf_ManPrintMfccStats(p *Mf_Man_t, id int32) int32 {
	area := Mf_CutAreaMffc(p, id)
	fmt.Printf("%5d : Level = %5d  Refs = %5d  Mffc = %5d\n", id, Gia_ObjLevelId(p.pGia0, id), Mf_ObjMapRefNum(p, id), area)
	return area
}
func Mf_ManOptimizationOne(p *Mf_Man_t, id int32) {
	g := p.pGia0
	count := Gia_ObjFanoutNumId(g, id)
	fanouts := make([]int32, count)
	for i := range fanouts {
		fanouts[i] = Gia_ObjFanoutId(g, id, int32(i))
		if Gia_ObjIsCo(Gia_ManObj(g, fanouts[i])) != 0 {
			return
		}
	}
	if Mf_ObjMapRefNum(p, id) <= 1 {
		panic("gia: pivot is not shared")
	}
	for _, v := range fanouts {
		if Mf_ObjMapRefNum(p, v) == 0 {
			panic("gia: unmapped fanout")
		}
	}
	fmt.Printf("\nPivot node = %d\nPivot ", id)
	Mf_ManPrintMfccStats(p, id)
	var sum int32
	for _, v := range fanouts {
		fmt.Print("Node  ")
		sum += Mf_ManPrintMfccStats(p, v)
	}
	for _, v := range fanouts {
		Mf_ObjMapRefInc(p, v)
	}
	cuts := make([][]int32, len(fanouts))
	for i, v := range fanouts {
		cuts[i] = Mf_ObjCutBest(p, v)
	}
	area := Mf_CutAreaRefed2Multi(p, id, cuts, int32(len(cuts)))
	for _, v := range fanouts {
		Mf_ObjMapRefDec(p, v)
	}
	fmt.Printf("Sum of MFFC sizes = %d\nShared MFFC size  = %d\n", sum, area)
}
func Mf_ManOptimization(p *Mf_Man_t) {
	g := p.pGia0
	nodes := Mf_ManMappingFromMapping(p)
	Gia_ManLevelNum(g)
	Gia_ManStaticMappingFanoutStart(g)
	Mf_ManPrintFanoutProfile(p, g.vFanoutNums)
	fmt.Printf("\nIndividual logic cones for mapping with %d nodes:\n", nodes)
	for id, n := range vec.Vec_IntArray(g.vFanoutNums) {
		if n >= 2 && n <= 3 && Gia_ObjIsAnd(Gia_ManObj(g, int32(id))) != 0 {
			Mf_ManOptimizationOne(p, int32(id))
		}
	}
	fmt.Println("\nFinished printing individual logic cones.")
	Gia_ManStaticFanoutStop(g)
	g.vMapping = nil
}
