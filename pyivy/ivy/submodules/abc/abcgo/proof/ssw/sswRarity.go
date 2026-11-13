package ssw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
	"math/bits"
	"math/rand"
	"time"
)

/**CFile****************************************************************

  FileName    [sswRarity.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Rarity-driven refinement of equivalence classes.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswRarity.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"
//#include "aig/gia/giaAig.h"
//#include "base/main/main.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Ssw_RarMan_t struct {
	pPars                *Ssw_RarPars_t
	nGroups, nWordsReg   int32
	pAig                 *aig.Graph
	ppClasses            *Ssw_Cla_t
	vInits               *vec.Vec_Int_t
	pObjData, pPatData   []uint64
	vUpdConst, vUpdClass *vec.Vec_Ptr_t
	pRarity              []int32
	pPatCosts            []float64
	vPatBests            *vec.Vec_Int_t
	iFailPo, iFailPat    int32
	vCexes               *vec.Vec_Ptr_t
}

func Ssw_RarGetBinPat(p *Ssw_RarMan_t, bin, pat int32) int32 {
	if bin < 0 || bin >= p.nGroups || pat < 0 || pat >= 1<<p.pPars.nBinSize {
		panic("ssw: rarity index out of bounds")
	}
	return p.pRarity[bin*(1<<p.pPars.nBinSize)+pat]
}
func Ssw_RarSetBinPat(p *Ssw_RarMan_t, bin, pat, value int32) {
	if bin < 0 || bin >= p.nGroups || pat < 0 || pat >= 1<<p.pPars.nBinSize {
		panic("ssw: rarity index out of bounds")
	}
	p.pRarity[bin*(1<<p.pPars.nBinSize)+pat] = value
}
func Ssw_RarAddToBinPat(p *Ssw_RarMan_t, bin, pat int32) {
	Ssw_RarSetBinPat(p, bin, pat, Ssw_RarGetBinPat(p, bin, pat)+1)
}
func Ssw_RarObjSim(p *Ssw_RarMan_t, id int32) []uint64 {
	if id < 0 || id >= aig.Aig_ManObjNumMax(p.pAig) {
		panic("ssw: invalid rarity object")
	}
	start := id * p.pPars.nWords
	return p.pObjData[start : start+p.pPars.nWords]
}
func Ssw_RarPatSim(p *Ssw_RarMan_t, id int32) []uint64 {
	if id < 0 || id >= 64*p.pPars.nWords {
		panic("ssw: invalid rarity pattern")
	}
	start := id * p.nWordsReg
	return p.pPatData[start : start+p.nWordsReg]
}
func Ssw_RarManAssingRandomPis(p *Ssw_RarMan_t) {
	for i, obj := range p.pAig.Inputs()[:saig.Saig_ManPiNum(p.pAig)] {
		sim := Ssw_RarObjSim(p, int32(obj.NodeID()))
		for w := range sim {
			sim[w] = aig.Aig_ManRandom64(0)
		}
		tail := uint64(0xc)
		if i&1 != 0 {
			tail = 0xa
		}
		sim[0] = sim[0]<<4 | tail
	}
}

/*
*Function*************************************************************

	Synopsis    [Derives the counter-example.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarDeriveCex(p *Ssw_RarMan_t, frame, po, pattern, verbose int32) *util.Abc_Cex_t {
	trace := make([]int32, frame/p.pPars.nFrames+1)
	trace[len(trace)-1] = pattern
	for r := len(trace) - 2; r >= 0; r-- {
		pattern = vec.Vec_IntEntry(p.vPatBests, r*int(p.pPars.nWords)+int(pattern/64))
		trace[r] = pattern
	}
	cex := util.Abc_CexAlloc(aig.Aig_ManRegNum(p.pAig), saig.Saig_ManPiNum(p.pAig), frame+1)
	cex.IFrame = frame
	cex.IPo = po
	bit := cex.NRegs
	for f := int32(0); f <= frame; f++ {
		Ssw_RarManAssingRandomPis(p)
		pattern = trace[f/p.pPars.nFrames]
		for _, obj := range p.pAig.Inputs()[:saig.Saig_ManPiNum(p.pAig)] {
			v := Ssw_RarObjSim(p, int32(obj.NodeID()))
			if v[pattern/64]&(uint64(1)<<uint(pattern%64)) != 0 {
				util.Abc_InfoSetBit(cex.PData, bit)
			}
			bit++
		}
	}
	if bit != cex.NBits {
		panic("ssw: incorrect rarity counterexample size")
	}
	if saig.Saig_ManVerifyCex(p.pAig, cex) == 0 {
		fmt.Println("Ssw_RarDeriveCex(): Counter-example is invalid.")
	}
	return cex
}

/*
*Function*************************************************************

	Synopsis    [Testing the transposing code.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func TransposeTest() {
	var a, b [64]word
	a[0] = ^word(0)
	start := time.Now()
	for i := 0; i < 100001; i++ {
		transpose64Simple(a[:], b[:])
	}
	fmt.Printf("Time %s\n", time.Since(start))
	start = time.Now()
	for i := 0; i < 100001; i++ {
		transpose64(a[:])
	}
	fmt.Printf("Time %s\n", time.Since(start))
	for i := range a {
		if a[i] != b[i] {
			fmt.Println("Mismatch")
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Transposing pObjData[ nRegs x nWords ] -> pPatData[ nWords x nRegs ].]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarTranspose(p *Ssw_RarMan_t) {
	for w := int32(0); w < p.pPars.nWords; w++ {
		for r := int32(0); r < p.nWordsReg; r++ {
			var m [64]uint64
			for i := int32(0); i < 64; i++ {
				idx := r*64 + 63 - i
				if idx < aig.Aig_ManRegNum(p.pAig) {
					li := saig.Saig_ManLi(p.pAig, idx)
					m[i] = Ssw_RarObjSim(p, int32(li.NodeID()))[w]
				}
			}
			transpose64(m[:])
			for i := int32(0); i < 64; i++ {
				Ssw_RarPatSim(p, w*64+63-i)[r] = m[i]
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Sets random inputs and specialied flop outputs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarManInitialize(p *Ssw_RarMan_t, init *vec.Vec_Int_t) {
	constant := Ssw_RarObjSim(p, 0)
	for i := range constant {
		constant[i] = ^uint64(0)
	}
	Ssw_RarManAssingRandomPis(p)
	regs := aig.Aig_ManRegNum(p.pAig)
	if init != nil && vec.Vec_IntSize(init) != int(regs*p.pPars.nWords) {
		panic("ssw: wrong rarity initialization size")
	}
	for i := int32(0); i < regs; i++ {
		lo, li := saig.Saig_ManLo(p.pAig, i), saig.Saig_ManLi(p.pAig, i)
		dst := Ssw_RarObjSim(p, int32(lo.NodeID()))
		if init == nil {
			copy(dst, Ssw_RarObjSim(p, int32(li.NodeID())))
		} else {
			for w := range dst {
				dst[w] = 0
				if vec.Vec_IntEntry(init, w*int(regs)+int(i)) != 0 {
					dst[w] = ^uint64(0)
				}
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if simulation info is composed of all zeros.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarManPoIsConst0(data any, obj aig.Edge) int32 {
	p := data.(*Ssw_RarMan_t)
	for _, v := range Ssw_RarObjSim(p, int32(obj.NodeID())) {
		if v != 0 {
			return 0
		}
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if simulation info is composed of all zeros.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarManObjIsConst(data any, obj aig.Edge) int32 {
	p := data.(*Ssw_RarMan_t)
	var flip uint64
	if aig.Aig_ObjPhase(p.pAig, obj) {
		flip = ^uint64(0)
	}
	for _, v := range Ssw_RarObjSim(p, int32(obj.NodeID())) {
		if v != flip {
			return 0
		}
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if simulation infos are equal.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarManObjsAreEqual(data any, a, b aig.Edge) int32 {
	p := data.(*Ssw_RarMan_t)
	var flip uint64
	if aig.Aig_ObjPhase(p.pAig, a) != aig.Aig_ObjPhase(p.pAig, b) {
		flip = ^uint64(0)
	}
	x, y := Ssw_RarObjSim(p, int32(a.NodeID())), Ssw_RarObjSim(p, int32(b.NodeID()))
	for i := range x {
		if x[i]^y[i]^flip != 0 {
			return 0
		}
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Computes hash value of the node using its simulation info.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Ssw_RarManObjHashWord_s_SPrimes [128]int32 = [128]int32{
	1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459,
	1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997,
	2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543,
	2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089,
	3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671,
	3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243,
	4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871,
	4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471,
	5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073,
	6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689,
	6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309,
	7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933,
	8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
}

func Ssw_RarManObjHashWord(data any, obj aig.Edge) uint32 {
	p := data.(*Ssw_RarMan_t)
	var hash uint32
	primes := []uint32{1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459, 1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997, 2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543, 2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089, 3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671, 3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243, 4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871, 4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471, 5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073, 6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689, 6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309, 7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933, 8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147}
	for i, w := range Ssw_RarObjSim(p, int32(obj.NodeID())) {
		hash ^= uint32(w) * primes[(2*i)&127]
		hash ^= uint32(w>>32) * primes[(2*i+1)&127]
	}
	return hash
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if simulation info is composed of all zeros.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarManObjWhichOne(p *Ssw_RarMan_t, obj aig.Edge) int32 {
	for w, v := range Ssw_RarObjSim(p, int32(obj.NodeID())) {
		if v != 0 {
			return int32(w*64 + bits.TrailingZeros64(v))
		}
	}
	return -1
}

/*
*Function*************************************************************

	Synopsis    [Check if any of the POs becomes non-constant.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarManCheckNonConstOutputs(p *Ssw_RarMan_t, frame int32, elapsed time.Duration) int32 {
	p.iFailPo = -1
	p.iFailPat = -1
	g := p.pAig
	limit := saig.Saig_ManPoNum(g) - aig.Aig_ManConstrNum(g)
	for i, obj := range g.Outputs()[:limit] {
		if p.vCexes != nil && vec.Vec_PtrEntry(p.vCexes, i) != nil {
			continue
		}
		if Ssw_RarManPoIsConst0(p, obj) != 0 {
			continue
		}
		p.iFailPo = int32(i)
		p.iFailPat = Ssw_RarManObjWhichOne(p, obj)
		if p.pPars.fSolveAll == 0 {
			break
		}
		p.pPars.nSolved++
		if p.vCexes == nil {
			p.vCexes = vec.Vec_PtrStart(int(saig.Saig_ManPoNum(g)))
		}
		vec.Vec_PtrWriteEntry(p.vCexes, i, uintptr(1))
		if p.pPars.pFuncOnFail != nil && p.pPars.pFuncOnFail(int32(i), nil) != 0 {
			return 2
		}
		if p.pPars.fNotVerbose == 0 {
			fmt.Printf("Output %d was asserted in frame %d (solved %d out of %d outputs). Time %s\n", i, frame, p.pPars.nSolved, saig.Saig_ManPoNum(g), elapsed)
		}
	}
	if p.iFailPo >= 0 {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Performs one round of simulation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarManSimulate(p *Ssw_RarMan_t, init *vec.Vec_Int_t, update, first int32) {
	g := p.pAig
	Ssw_RarManInitialize(p, init)
	vec.Vec_PtrClear(p.vUpdConst)
	vec.Vec_PtrClear(p.vUpdClass)
	aig.Aig_ManIncrementTravId(g)
	check := func(obj aig.Edge) {
		if update == 0 {
			return
		}
		repr := aig.Aig_ObjRepr(g, obj)
		if repr == aig.InvalidEdge || aig.Aig_ObjIsTravIdCurrent(g, repr) || Ssw_RarManObjsAreEqual(p, obj, repr) != 0 {
			return
		}
		if repr == g.Const1() {
			vec.Vec_PtrPush(p.vUpdConst, obj)
		} else {
			vec.Vec_PtrPush(p.vUpdClass, repr)
			aig.Aig_ObjSetTravIdCurrent(g, repr)
		}
	}
	for _, obj := range g.Inputs() {
		check(obj)
	}
	for _, obj := range aig.Aig_ManDfsAll(g) {
		n := g.Node(obj)
		if n.Kind != aig.And && n.Kind != aig.ExclusiveOr && n.Kind != aig.Buffer {
			continue
		}
		dst := Ssw_RarObjSim(p, int32(obj.NodeID()))
		a := Ssw_RarObjSim(p, int32(n.Fanins[0].NodeID()))
		var b []uint64
		if n.Kind != aig.Buffer {
			b = Ssw_RarObjSim(p, int32(n.Fanins[1].NodeID()))
		}
		for w := range dst {
			x := a[w]
			if n.Fanins[0].Complemented() {
				x = ^x
			}
			if n.Kind == aig.Buffer {
				dst[w] = x
				continue
			}
			y := b[w]
			if n.Fanins[1].Complemented() {
				y = ^y
			}
			if n.Kind == aig.And {
				dst[w] = x & y
			} else {
				dst[w] = x ^ y
			}
		}
		if n.Kind != aig.Buffer {
			check(obj)
		}
	}
	for _, obj := range g.Outputs() {
		child := g.Node(obj).Fanins[0]
		dst, src := Ssw_RarObjSim(p, int32(obj.NodeID())), Ssw_RarObjSim(p, int32(child.NodeID()))
		for w := range dst {
			dst[w] = src[w]
			if child.Complemented() {
				dst[w] = ^dst[w]
			}
		}
	}
	if update != 0 {
		if first != 0 {
			v := vec.Vec_PtrAlloc(1000)
			for _, obj := range g.LiveEdges() {
				if Ssw_ObjIsConst1Cand(g, obj) != 0 {
					vec.Vec_PtrPush(v, obj)
				}
			}
			if int32(vec.Vec_PtrSize(v)) != Ssw_ClassesCand1Num(p.ppClasses) {
				panic("ssw: inconsistent rarity candidate count")
			}
			Ssw_ClassesPrepareRehash(p.ppClasses, v, 0)
		} else {
			Ssw_ClassesRefineConst1Group(p.ppClasses, p.vUpdConst, 1)
			Ssw_ClassesRefineGroup(p.ppClasses, p.vUpdClass, 1)
		}
	}
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarManStart(g *aig.Graph, pars *Ssw_RarPars_t) *Ssw_RarMan_t {
	if pars.nWords <= 0 || pars.nFrames <= 0 || pars.nBinSize <= 0 || pars.nBinSize > 16 {
		panic("ssw: invalid rarity parameters")
	}
	groups := aig.Aig_ManRegNum(g) / pars.nBinSize
	words := Ssw_RarBitWordNum(aig.Aig_ManRegNum(g))
	return &Ssw_RarMan_t{pAig: g, pPars: pars, nGroups: groups, nWordsReg: words, pRarity: make([]int32, int(groups)*(1<<pars.nBinSize)), pPatCosts: make([]float64, int(pars.nWords)*64), pObjData: make([]uint64, int(aig.Aig_ManObjNumMax(g))*int(pars.nWords)), pPatData: make([]uint64, 64*int(pars.nWords)*int(words)), vUpdConst: vec.Vec_PtrAlloc(100), vUpdClass: vec.Vec_PtrAlloc(100), vPatBests: vec.Vec_IntAlloc(100)}
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarManStop(p *Ssw_RarMan_t) {
	if p == nil {
		return
	}
	if p.vCexes != nil {
		if p.pAig.SequentialCounterexamples() != nil {
			panic("ssw: counterexample vector already exists")
		}
		p.pAig.SetSequentialCounterexamples(p.vCexes)
	}
	if p.ppClasses != nil {
		Ssw_ClassesStop(p.ppClasses)
	}
	*p = Ssw_RarMan_t{}
}

/*
*Function*************************************************************

	Synopsis    [Select best patterns.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarTransferPatterns(p *Ssw_RarMan_t, init *vec.Vec_Int_t) {
	Ssw_RarTranspose(p)
	count := int32(64) * p.pPars.nWords
	getBin := func(pattern []uint64, bin int32) int32 {
		start := bin * p.pPars.nBinSize
		word, shift := start/64, uint(start%64)
		value := pattern[word] >> shift
		if shift+uint(p.pPars.nBinSize) > 64 {
			value |= pattern[word+1] << (64 - shift)
		}
		return int32(value & ((uint64(1) << uint(p.pPars.nBinSize)) - 1))
	}
	for k := int32(0); k < count; k++ {
		pat := Ssw_RarPatSim(p, k)
		for i := int32(0); i < p.nGroups; i++ {
			Ssw_RarAddToBinPat(p, i, getBin(pat, i))
		}
	}
	for k := int32(0); k < count; k++ {
		pat := Ssw_RarPatSim(p, k)
		p.pPatCosts[k] = 0
		for i := int32(0); i < p.nGroups; i++ {
			v := Ssw_RarGetBinPat(p, i, getBin(pat, i))
			if v <= 0 {
				panic("ssw: invalid rarity count")
			}
			p.pPatCosts[k] += 1 / (float64(v) * float64(v))
		}
	}
	vec.Vec_IntClear(init)
	for i := int32(0); i < p.pPars.nWords; i++ {
		best := int32(-1)
		cost := math.Inf(-1)
		for k := int32(0); k < count; k++ {
			if cost < p.pPatCosts[k] {
				best = k
				cost = p.pPatCosts[k]
			}
		}
		if best < 0 {
			panic("ssw: rarity selection failed")
		}
		p.pPatCosts[best] = math.Inf(-1)
		pat := Ssw_RarPatSim(p, best)
		for k := int32(0); k < aig.Aig_ManRegNum(p.pAig); k++ {
			v := int32((pat[k/64] >> uint(k%64)) & 1)
			vec.Vec_IntPush(init, v)
		}
		vec.Vec_IntPush(p.vPatBests, best)
	}
}

/*
*Function*************************************************************

	Synopsis    [Performs fraiging for one node.]

	Description [Returns the fraiged node.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarFindStartingState(g *aig.Graph, cex *util.Abc_Cex_t) *vec.Vec_Int_t {
	pos, pis := saig.Saig_ManPoNum(g), saig.Saig_ManPiNum(g)
	for i, li := range g.Outputs()[pos:] {
		sswSetMark(g, li, util.Abc_InfoHasBit(cex.PData, int32(i)))
	}
	p := &Ssw_Man_t{pAig: g}
	bit := cex.NRegs
	for f := int32(0); f <= cex.IFrame; f++ {
		sswFilterSimulateFrame(p, func(int) bool { v := util.Abc_InfoHasBit(cex.PData, bit); bit++; return v })
	}
	if bit != cex.NBits {
		panic("ssw: invalid rarity counterexample")
	}
	init := vec.Vec_IntAlloc(int(g.RegisterCount()))
	for _, lo := range g.Inputs()[pis:] {
		v := int32(0)
		if g.Node(lo).MarkB {
			v = 1
		}
		vec.Vec_IntPush(init, v)
	}
	aig.Aig_ManCleanMarkB(g)
	return init
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarCheckTrivial(g *aig.Graph, verbose int32) int32 {
	limit := saig.Saig_ManPoNum(g) - aig.Aig_ManConstrNum(g)
	for i, obj := range g.Outputs()[:limit] {
		if aig.Aig_ObjPhase(g, obj) {
			cex := util.Abc_CexAlloc(g.RegisterCount(), saig.Saig_ManPiNum(g), 1)
			cex.IPo = int32(i)
			if err := g.SetSequentialCounterexample(cex); err != nil {
				panic(err)
			}
			if verbose != 0 {
				fmt.Printf("Output %d is trivially SAT in frame 0.\n", i)
			}
			return 1
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Perform sequential simulation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarSimulate(g *aig.Graph, pars *Ssw_RarPars_t) int32 {
	if g.RegisterCount() <= 0 || aig.Aig_ManConstrNum(g) != 0 {
		panic("ssw: invalid rarity graph")
	}
	g.ClearSequentialCounterexample()
	seed := pars.nRandSeed
	Ssw_RarManPrepareRandom(seed)
	p := Ssw_RarManStart(g, pars)
	defer Ssw_RarManStop(p)
	p.vInits = vec.Vec_IntStart(int(g.RegisterCount() * pars.nWords))
	pars.nSolved = 0
	start := time.Now()
	lastSolved := start
	restarts := int32(0)
	result := int32(-1)
outer:
	for r := int32(0); pars.nRounds == 0 || restarts*pars.nRestart+r < pars.nRounds; r++ {
		for f := int32(0); f < pars.nFrames; f++ {
			var init *vec.Vec_Int_t
			if f == 0 {
				init = p.vInits
			}
			Ssw_RarManSimulate(p, init, 0, 0)
			status := Ssw_RarManCheckNonConstOutputs(p, r*pars.nFrames+f, time.Since(start))
			if status == 2 {
				break outer
			}
			if status == 1 {
				result = 0
				if pars.fSolveAll == 0 {
					Ssw_RarManPrepareRandom(seed)
					cex := Ssw_RarDeriveCex(p, r*pars.nFrames+f, p.iFailPo, p.iFailPat, pars.fVerbose)
					if err := g.SetSequentialCounterexample(cex); err != nil {
						panic(err)
					}
					break outer
				}
				lastSolved = time.Now()
			}
			if (pars.TimeOut != 0 && time.Since(start) > time.Duration(pars.TimeOut)*time.Second) || (pars.TimeOutGap != 0 && time.Since(lastSolved) > time.Duration(pars.TimeOutGap)*time.Second) {
				break outer
			}
			if pars.fSolveAll != 0 && p.vCexes != nil && vec.Vec_PtrCountZero(p.vCexes) == 0 {
				break outer
			}
		}
		if pars.nRestart != 0 && r == pars.nRestart {
			r = -1
			seed = (seed + 1) % 1000
			Ssw_RarManPrepareRandom(seed)
			vec.Vec_IntFill(p.vInits, int(g.RegisterCount()*pars.nWords), 0)
			restarts++
			vec.Vec_IntClear(p.vPatBests)
		} else {
			Ssw_RarTransferPatterns(p, p.vInits)
		}
	}
	if pars.fSetLastState != 0 && p.vInits != nil {
		vec.Vec_IntShrink(p.vInits, int(g.RegisterCount()))
		g.Data = p.vInits
		p.vInits = nil
	}
	return result
}

/*
*Function*************************************************************

	Synopsis    [Derive random flop permutation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarRandomPermFlop(flops, unused int32) *vec.Vec_Int_t {
	rng := rand.New(rand.NewSource(1))
	v := vec.Vec_IntStartNatural(int(flops))
	if flops == 0 && unused > 0 {
		panic("ssw: cannot add gaps without flops")
	}
	for i := int32(0); i < flops; i++ {
		k := rng.Intn(int(flops))
		a := vec.Vec_IntArray(v)
		a[i], a[k] = a[k], a[i]
	}
	for i := int32(0); i < unused; i++ {
		k := rng.Intn(vec.Vec_IntSize(v))
		vec.Vec_IntPush(v, -1)
		a := vec.Vec_IntArray(v)
		a[len(a)-1], a[k] = a[k], a[len(a)-1]
	}
	return v
}

/*
*Function*************************************************************

	Synopsis    [Perform sequential simulation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarSimulateGia(g *gia.Gia_Man_t, pars *Ssw_RarPars_t) int32 {
	src := g
	var perm *vec.Vec_Int_t
	if pars.fUseFfGrouping != 0 {
		perm = Ssw_RarRandomPermFlop(gia.Gia_ManRegNum(g), 10)
		src = gia.Gia_ManDupPermFlopGap(g, perm)
		defer gia.Gia_ManStop(src)
	}
	graph := gia.Gia_ManToAigSimple(src)
	defer aig.Aig_ManStop(graph)
	result := Ssw_RarSimulate(graph, pars)
	g.ClearSequentialCounterexample()
	cex := graph.SequentialCounterexample()
	if cex != nil {
		if perm != nil {
			out := util.Abc_CexAlloc(gia.Gia_ManRegNum(g), gia.Gia_ManPiNum(g), cex.IFrame+1)
			out.IPo = cex.IPo
			out.IFrame = cex.IFrame
			for i, old := range vec.Vec_IntArray(perm) {
				if old >= 0 && util.Abc_InfoHasBit(cex.PData, int32(i)) {
					util.Abc_InfoSetBit(out.PData, old)
				}
			}
			for i := int32(0); i < (cex.IFrame+1)*cex.NPis; i++ {
				if util.Abc_InfoHasBit(cex.PData, cex.NRegs+i) {
					util.Abc_InfoSetBit(out.PData, out.NRegs+i)
				}
			}
			cex = out
		}
		if err := g.SetSequentialCounterexample(cex); err != nil {
			panic(err)
		}
	}
	return result
}

/*
*Function*************************************************************

	Synopsis    [Filter equivalence classes of nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarSignalFilter(g *aig.Graph, pars *Ssw_RarPars_t) int32 {
	if g.RegisterCount() <= 0 || aig.Aig_ManConstrNum(g) != 0 {
		panic("ssw: invalid rarity filter graph")
	}
	if aig.Aig_ManNodeNum(g) == 0 {
		return -1
	}
	if pars.fMiter != 0 && Ssw_RarCheckTrivial(g, 1) != 0 {
		return 0
	}
	seed := pars.nRandSeed
	Ssw_RarManPrepareRandom(seed)
	p := Ssw_RarManStart(g, pars)
	defer Ssw_RarManStop(p)
	if pars.pCex != nil {
		p.vInits = Ssw_RarFindStartingState(g, pars.pCex)
	} else {
		p.vInits = vec.Vec_IntStart(int(g.RegisterCount()))
	}
	for i := int32(1); i < pars.nWords; i++ {
		for k := int32(0); k < g.RegisterCount(); k++ {
			vec.Vec_IntPush(p.vInits, vec.Vec_IntEntry(p.vInits, int(k)))
		}
	}
	if aig.Aig_ManHasReprs(g) {
		p.ppClasses = Ssw_ClassesPrepareFromReprs(g)
	} else {
		p.ppClasses = Ssw_ClassesPrepareSimple(g, pars.fLatchOnly, 0)
	}
	Ssw_ClassesSetData(p.ppClasses, p, Ssw_RarManObjHashWord, Ssw_RarManObjIsConst, Ssw_RarManObjsAreEqual)
	start := time.Now()
	restarts := int32(0)
	for r := int32(0); pars.nRounds == 0 || restarts*pars.nRestart+r < pars.nRounds; r++ {
		if Ssw_ClassesCand1Num(p.ppClasses) == 0 && Ssw_ClassesClassNum(p.ppClasses) == 0 {
			break
		}
		for f := int32(0); f < pars.nFrames; f++ {
			var init *vec.Vec_Int_t
			if f == 0 {
				init = p.vInits
			}
			first := int32(0)
			if r == 0 && f == 0 {
				first = 1
			}
			Ssw_RarManSimulate(p, init, 1, first)
			if pars.fMiter != 0 && Ssw_RarManCheckNonConstOutputs(p, -1, 0) != 0 {
				Ssw_RarManPrepareRandom(seed)
				cex := Ssw_RarDeriveCex(p, r*pars.nFrames+f, p.iFailPo, p.iFailPat, 1)
				if err := g.SetSequentialCounterexample(cex); err != nil {
					panic(err)
				}
				return 0
			}
			if pars.TimeOut != 0 && time.Since(start) > time.Duration(pars.TimeOut)*time.Second {
				return -1
			}
		}
		if pars.pCex == nil && pars.nRestart != 0 && r == pars.nRestart {
			r = -1
			seed = (seed + 1) % 1000
			Ssw_RarManPrepareRandom(seed)
			vec.Vec_IntFill(p.vInits, int(g.RegisterCount()*pars.nWords), 0)
			restarts++
			vec.Vec_IntClear(p.vPatBests)
		} else {
			Ssw_RarTransferPatterns(p, p.vInits)
		}
	}
	return -1
}

/*
*Function*************************************************************

	Synopsis    [Filter equivalence classes of nodes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ssw_RarSignalFilterGia(g *gia.Gia_Man_t, pars *Ssw_RarPars_t) int32 {
	graph := gia.Gia_ManToAigSimple(g)
	defer aig.Aig_ManStop(graph)
	if g.HasRepresentatives() {
		gia.Gia_ManReprToAigRepr2(graph, g)
		g.ClearRepresentatives()
	}
	result := Ssw_RarSignalFilter(graph, pars)
	gia.Gia_ManReprFromAigRepr(graph, g)
	g.ClearSequentialCounterexample()
	if cex := graph.SequentialCounterexample(); cex != nil {
		if err := g.SetSequentialCounterexample(cex); err != nil {
			panic(err)
		}
	}
	return result
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
