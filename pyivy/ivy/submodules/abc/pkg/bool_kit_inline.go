package abc

import "math/bits"

// Port of the static inline functions of src/bool/kit/kit.h that operate on cubes,
// SOPs, edges, supports and truth tables stored as []uint32 (the "unsigned *" arrays).
// The Kit_Graph* inlines over Kit_Node_t* and Kit_DsdObjTruth (which reinterprets the
// fanin array as truth table words) are not ported: the validated KitGraph
// (bool_kit_graph.go) stores its nodes differently and has its own API.

// kitDsdObjOffset is Kit_DsdObjOffset.
// C: unsigned Kit_DsdObjOffset( int nFans )
func kitDsdObjOffset(nFans int) uint32 {
	r := uint32(nFans >> 1)
	if nFans&1 > 0 {
		r++
	}
	return r
}

// kitDsdNtkObjNum is Kit_DsdNtkObjNum.
// C: int Kit_DsdNtkObjNum( Kit_DsdNtk_t * pNtk )
func kitDsdNtkObjNum(pNtk *KitDsdNtk) int { return int(pNtk.nVars) + int(pNtk.nNodes) }

// kitDsdNtkObj is Kit_DsdNtkObj (nil for a variable).
// C: Kit_DsdObj_t * Kit_DsdNtkObj( Kit_DsdNtk_t * pNtk, int Id )
func kitDsdNtkObj(pNtk *KitDsdNtk, Id int) *KitDsdObj {
	if !(Id >= 0 && Id < int(pNtk.nVars)+int(pNtk.nNodes)) {
		panic("kitDsdNtkObj: Id out of range")
	}
	if Id < int(pNtk.nVars) {
		return nil
	}
	return pNtk.pNodes[Id-int(pNtk.nVars)]
}

// kitDsdNtkRoot is Kit_DsdNtkRoot.
// C: Kit_DsdObj_t * Kit_DsdNtkRoot( Kit_DsdNtk_t * pNtk )
func kitDsdNtkRoot(pNtk *KitDsdNtk) *KitDsdObj { return kitDsdNtkObj(pNtk, int(pNtk.Root)>>1) }

// kitDsdLitIsLeaf is Kit_DsdLitIsLeaf.
// C: int Kit_DsdLitIsLeaf( Kit_DsdNtk_t * pNtk, int Lit )
func kitDsdLitIsLeaf(pNtk *KitDsdNtk, Lit int) int {
	Id := Lit >> 1
	if !(Id >= 0 && Id < int(pNtk.nVars)+int(pNtk.nNodes)) {
		panic("kitDsdLitIsLeaf: Id out of range")
	}
	return kitBool(Id < int(pNtk.nVars))
}

// kitDsdLitSupport is Kit_DsdLitSupport.
// C: unsigned Kit_DsdLitSupport( Kit_DsdNtk_t * pNtk, int Lit )
func kitDsdLitSupport(pNtk *KitDsdNtk, Lit int) uint32 {
	Id := Lit >> 1
	if !(Id >= 0 && Id < int(pNtk.nVars)+int(pNtk.nNodes)) {
		panic("kitDsdLitSupport: Id out of range")
	}
	if pNtk.pSupps == nil {
		return 0
	}
	if Id < int(pNtk.nVars) {
		return 1 << uint(Id)
	}
	return pNtk.pSupps[Id-int(pNtk.nVars)]
}

func kitBool(b bool) int {
	if b {
		return 1
	}
	return 0
}

// kitCubeHasLit is Kit_CubeHasLit.
// C: int Kit_CubeHasLit( unsigned uCube, int i )
func kitCubeHasLit(uCube uint32, i int) int { return kitBool(uCube&(uint32(1)<<uint(i)) > 0) }

// kitCubeSetLit is Kit_CubeSetLit.
// C: unsigned Kit_CubeSetLit( unsigned uCube, int i )
func kitCubeSetLit(uCube uint32, i int) uint32 { return uCube | uint32(1)<<uint(i) }

// kitCubeXorLit is Kit_CubeXorLit.
// C: unsigned Kit_CubeXorLit( unsigned uCube, int i )
func kitCubeXorLit(uCube uint32, i int) uint32 { return uCube ^ uint32(1)<<uint(i) }

// kitCubeRemLit is Kit_CubeRemLit.
// C: unsigned Kit_CubeRemLit( unsigned uCube, int i )
func kitCubeRemLit(uCube uint32, i int) uint32 { return uCube &^ (uint32(1) << uint(i)) }

// kitCubeContains is Kit_CubeContains.
// C: int Kit_CubeContains( unsigned uLarge, unsigned uSmall )
func kitCubeContains(uLarge uint32, uSmall uint32) int { return kitBool(uLarge&uSmall == uSmall) }

// kitCubeSharp is Kit_CubeSharp.
// C: unsigned Kit_CubeSharp( unsigned uCube, unsigned uMask )
func kitCubeSharp(uCube uint32, uMask uint32) uint32 { return uCube &^ uMask }

// kitCubeMask is Kit_CubeMask.
// C: unsigned Kit_CubeMask( int nVar )
func kitCubeMask(nVar int) uint32 { return (^uint32(0)) >> uint(32-nVar) }

// kitCubeIsMarked is Kit_CubeIsMarked.
// C: int Kit_CubeIsMarked( unsigned uCube )
func kitCubeIsMarked(uCube uint32) int { return kitCubeHasLit(uCube, 31) }

// kitCubeMark is Kit_CubeMark.
// C: unsigned Kit_CubeMark( unsigned uCube )
func kitCubeMark(uCube uint32) uint32 { return kitCubeSetLit(uCube, 31) }

// kitCubeUnmark is Kit_CubeUnmark.
// C: unsigned Kit_CubeUnmark( unsigned uCube )
func kitCubeUnmark(uCube uint32) uint32 { return kitCubeRemLit(uCube, 31) }

// kitSopCubeNum is Kit_SopCubeNum.
// C: int Kit_SopCubeNum( Kit_Sop_t * cSop )
func kitSopCubeNum(cSop *KitSop) int { return cSop.nCubes }

// kitSopCube is Kit_SopCube.
// C: unsigned Kit_SopCube( Kit_Sop_t * cSop, int i )
func kitSopCube(cSop *KitSop, i int) uint32 { return cSop.pCubes[i] }

// kitSopShrink is Kit_SopShrink.
// C: void Kit_SopShrink( Kit_Sop_t * cSop, int nCubesNew )
func kitSopShrink(cSop *KitSop, nCubesNew int) { cSop.nCubes = nCubesNew }

// kitSopPushCube is Kit_SopPushCube.
// C: void Kit_SopPushCube( Kit_Sop_t * cSop, unsigned uCube )
func kitSopPushCube(cSop *KitSop, uCube uint32) {
	cSop.pCubes[cSop.nCubes] = uCube
	cSop.nCubes++
}

// kitSopWriteCube is Kit_SopWriteCube.
// C: void Kit_SopWriteCube( Kit_Sop_t * cSop, unsigned uCube, int i )
func kitSopWriteCube(cSop *KitSop, uCube uint32, i int) { cSop.pCubes[i] = uCube }

// kitEdgeCreate is Kit_EdgeCreate.
// C: Kit_Edge_t Kit_EdgeCreate( int Node, int fCompl )
func kitEdgeCreate(Node int, fCompl int) KitEdge { return KitEdge{Node: Node, Complement: fCompl != 0} }

// kitEdgeToInt is Kit_EdgeToInt.
// C: unsigned Kit_EdgeToInt( Kit_Edge_t eEdge )
func kitEdgeToInt(eEdge KitEdge) uint32 {
	return uint32(eEdge.Node)<<1 | uint32(kitBool(eEdge.Complement))
}

// kitIntToEdge is Kit_IntToEdge.
// C: Kit_Edge_t Kit_IntToEdge( unsigned Edge )
func kitIntToEdge(Edge uint32) KitEdge { return kitEdgeCreate(int(Edge>>1), int(Edge&1)) }

// kitEdgeToInt_ is Kit_EdgeToInt_: the C code reinterprets the bit-field struct
// { fCompl:1, Node:30 } as an unsigned word, i.e. fCompl in bit 0 and Node in bits 1..30.
// C: unsigned Kit_EdgeToInt_( Kit_Edge_t m )
func kitEdgeToInt_(m KitEdge) uint32 { return kitEdgeToInt(m) & 0x7FFFFFFF }

// kitIntToEdge_ is Kit_IntToEdge_ (the inverse reinterpretation).
// C: Kit_Edge_t Kit_IntToEdge_( unsigned m )
func kitIntToEdge_(m uint32) KitEdge { return kitEdgeCreate(int((m>>1)&0x3FFFFFFF), int(m&1)) }

// kitSuppIsMinBase is Kit_SuppIsMinBase.
// C: int Kit_SuppIsMinBase( int Supp )
func kitSuppIsMinBase(Supp int) int { return kitBool((Supp & (Supp + 1)) == 0) }

// kitBitWordNum is Kit_BitWordNum.
// C: int Kit_BitWordNum( int nBits )
func kitBitWordNum(nBits int) int { return nBits/32 + kitBool(nBits%32 > 0) }

// kitTruthWordNum is Kit_TruthWordNum.
// C: int Kit_TruthWordNum( int nVars )
func kitTruthWordNum(nVars int) int {
	if nVars <= 5 {
		return 1
	}
	return 1 << uint(nVars-5)
}

// kitBitMask is Kit_BitMask.
// C: unsigned Kit_BitMask( int nBits )
func kitBitMask(nBits int) uint32 {
	if nBits > 32 {
		panic("kitBitMask: nBits > 32")
	}
	return ^((^uint32(0)) << uint(nBits))
}

// kitTruthSetBit is Kit_TruthSetBit.
// C: void Kit_TruthSetBit( unsigned * p, int Bit )
func kitTruthSetBit(p []uint32, Bit int) { p[Bit>>5] |= 1 << uint(Bit&31) }

// kitTruthXorBit is Kit_TruthXorBit.
// C: void Kit_TruthXorBit( unsigned * p, int Bit )
func kitTruthXorBit(p []uint32, Bit int) { p[Bit>>5] ^= 1 << uint(Bit&31) }

// kitTruthHasBit is Kit_TruthHasBit.
// C: int Kit_TruthHasBit( unsigned * p, int Bit )
func kitTruthHasBit(p []uint32, Bit int) int { return kitBool(p[Bit>>5]&(1<<uint(Bit&31)) > 0) }

// kitWordFindFirstBit is Kit_WordFindFirstBit.
// C: int Kit_WordFindFirstBit( unsigned uWord )
func kitWordFindFirstBit(uWord uint32) int {
	if uWord == 0 {
		return -1
	}
	return bits.TrailingZeros32(uWord)
}

// kitWordHasOneBit is Kit_WordHasOneBit.
// C: int Kit_WordHasOneBit( unsigned uWord )
func kitWordHasOneBit(uWord uint32) int { return kitBool((uWord & (uWord - 1)) == 0) }

// kitWordCountOnes is Kit_WordCountOnes.
// C: int Kit_WordCountOnes( unsigned uWord )
func kitWordCountOnes(uWord uint32) int { return bits.OnesCount32(uWord) }

// kitTruthCountOnes is Kit_TruthCountOnes.
// C: int Kit_TruthCountOnes( unsigned * pIn, int nVars )
func kitTruthCountOnes(pIn []uint32, nVars int) int {
	Counter := 0
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		Counter += kitWordCountOnes(pIn[w])
	}
	return Counter
}

// kitTruthFindFirstBit is Kit_TruthFindFirstBit.
// C: int Kit_TruthFindFirstBit( unsigned * pIn, int nVars )
func kitTruthFindFirstBit(pIn []uint32, nVars int) int {
	for w := 0; w < kitTruthWordNum(nVars); w++ {
		if pIn[w] != 0 {
			return 32*w + kitWordFindFirstBit(pIn[w])
		}
	}
	return -1
}

// kitTruthFindFirstZero is Kit_TruthFindFirstZero.
// C: int Kit_TruthFindFirstZero( unsigned * pIn, int nVars )
func kitTruthFindFirstZero(pIn []uint32, nVars int) int {
	for w := 0; w < kitTruthWordNum(nVars); w++ {
		if ^pIn[w] != 0 {
			return 32*w + kitWordFindFirstBit(^pIn[w])
		}
	}
	return -1
}

// kitTruthIsEqual is Kit_TruthIsEqual.
// C: int Kit_TruthIsEqual( unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthIsEqual(pIn0 []uint32, pIn1 []uint32, nVars int) int {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		if pIn0[w] != pIn1[w] {
			return 0
		}
	}
	return 1
}

// kitTruthIsEqualWithCare is Kit_TruthIsEqualWithCare.
// C: int Kit_TruthIsEqualWithCare( unsigned * pIn0, unsigned * pIn1, unsigned * pCare, int nVars )
func kitTruthIsEqualWithCare(pIn0 []uint32, pIn1 []uint32, pCare []uint32, nVars int) int {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		if (pIn0[w] & pCare[w]) != (pIn1[w] & pCare[w]) {
			return 0
		}
	}
	return 1
}

// kitTruthIsOpposite is Kit_TruthIsOpposite.
// C: int Kit_TruthIsOpposite( unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthIsOpposite(pIn0 []uint32, pIn1 []uint32, nVars int) int {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		if pIn0[w] != ^pIn1[w] {
			return 0
		}
	}
	return 1
}

// kitTruthIsEqualWithPhase is Kit_TruthIsEqualWithPhase.
// C: int Kit_TruthIsEqualWithPhase( unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthIsEqualWithPhase(pIn0 []uint32, pIn1 []uint32, nVars int) int {
	if (pIn0[0] & 1) == (pIn1[0] & 1) {
		for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
			if pIn0[w] != pIn1[w] {
				return 0
			}
		}
	} else {
		for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
			if pIn0[w] != ^pIn1[w] {
				return 0
			}
		}
	}
	return 1
}

// kitTruthIsConst0 is Kit_TruthIsConst0.
// C: int Kit_TruthIsConst0( unsigned * pIn, int nVars )
func kitTruthIsConst0(pIn []uint32, nVars int) int {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		if pIn[w] != 0 {
			return 0
		}
	}
	return 1
}

// kitTruthIsConst1 is Kit_TruthIsConst1.
// C: int Kit_TruthIsConst1( unsigned * pIn, int nVars )
func kitTruthIsConst1(pIn []uint32, nVars int) int {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		if pIn[w] != ^uint32(0) {
			return 0
		}
	}
	return 1
}

// kitTruthIsImply is Kit_TruthIsImply.
// C: int Kit_TruthIsImply( unsigned * pIn1, unsigned * pIn2, int nVars )
func kitTruthIsImply(pIn1 []uint32, pIn2 []uint32, nVars int) int {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		if pIn1[w]&^pIn2[w] != 0 {
			return 0
		}
	}
	return 1
}

// kitTruthIsDisjoint is Kit_TruthIsDisjoint.
// C: int Kit_TruthIsDisjoint( unsigned * pIn1, unsigned * pIn2, int nVars )
func kitTruthIsDisjoint(pIn1 []uint32, pIn2 []uint32, nVars int) int {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		if pIn1[w]&pIn2[w] != 0 {
			return 0
		}
	}
	return 1
}

// kitTruthIsDisjoint3 is Kit_TruthIsDisjoint3.
// C: int Kit_TruthIsDisjoint3( unsigned * pIn1, unsigned * pIn2, unsigned * pIn3, int nVars )
func kitTruthIsDisjoint3(pIn1 []uint32, pIn2 []uint32, pIn3 []uint32, nVars int) int {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		if pIn1[w]&pIn2[w]&pIn3[w] != 0 {
			return 0
		}
	}
	return 1
}

// kitTruthCopy is Kit_TruthCopy.
// C: void Kit_TruthCopy( unsigned * pOut, unsigned * pIn, int nVars )
func kitTruthCopy(pOut []uint32, pIn []uint32, nVars int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		pOut[w] = pIn[w]
	}
}

// kitTruthClear is Kit_TruthClear.
// C: void Kit_TruthClear( unsigned * pOut, int nVars )
func kitTruthClear(pOut []uint32, nVars int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		pOut[w] = 0
	}
}

// kitTruthFill is Kit_TruthFill.
// C: void Kit_TruthFill( unsigned * pOut, int nVars )
func kitTruthFill(pOut []uint32, nVars int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		pOut[w] = ^uint32(0)
	}
}

// kitTruthNot is Kit_TruthNot.
// C: void Kit_TruthNot( unsigned * pOut, unsigned * pIn, int nVars )
func kitTruthNot(pOut []uint32, pIn []uint32, nVars int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		pOut[w] = ^pIn[w]
	}
}

// kitTruthAnd is Kit_TruthAnd.
// C: void Kit_TruthAnd( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthAnd(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		pOut[w] = pIn0[w] & pIn1[w]
	}
}

// kitTruthOr is Kit_TruthOr.
// C: void Kit_TruthOr( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthOr(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		pOut[w] = pIn0[w] | pIn1[w]
	}
}

// kitTruthXor is Kit_TruthXor.
// C: void Kit_TruthXor( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthXor(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		pOut[w] = pIn0[w] ^ pIn1[w]
	}
}

// kitTruthSharp is Kit_TruthSharp.
// C: void Kit_TruthSharp( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthSharp(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		pOut[w] = pIn0[w] &^ pIn1[w]
	}
}

// kitTruthNand is Kit_TruthNand.
// C: void Kit_TruthNand( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars )
func kitTruthNand(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		pOut[w] = ^(pIn0[w] & pIn1[w])
	}
}

// kitTruthAndPhase is Kit_TruthAndPhase.
// C: void Kit_TruthAndPhase( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars, int fCompl0, int fCompl1 )
func kitTruthAndPhase(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int, fCompl0 int, fCompl1 int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		switch {
		case fCompl0 != 0 && fCompl1 != 0:
			pOut[w] = ^(pIn0[w] | pIn1[w])
		case fCompl0 != 0 && fCompl1 == 0:
			pOut[w] = ^pIn0[w] & pIn1[w]
		case fCompl0 == 0 && fCompl1 != 0:
			pOut[w] = pIn0[w] &^ pIn1[w]
		default:
			pOut[w] = pIn0[w] & pIn1[w]
		}
	}
}

// kitTruthOrPhase is Kit_TruthOrPhase.
// C: void Kit_TruthOrPhase( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, int nVars, int fCompl0, int fCompl1 )
func kitTruthOrPhase(pOut []uint32, pIn0 []uint32, pIn1 []uint32, nVars int, fCompl0 int, fCompl1 int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		switch {
		case fCompl0 != 0 && fCompl1 != 0:
			pOut[w] = ^(pIn0[w] & pIn1[w])
		case fCompl0 != 0 && fCompl1 == 0:
			pOut[w] = ^pIn0[w] | pIn1[w]
		case fCompl0 == 0 && fCompl1 != 0:
			pOut[w] = pIn0[w] | ^pIn1[w]
		default:
			pOut[w] = pIn0[w] | pIn1[w]
		}
	}
}

// kitTruthMux is Kit_TruthMux.
// C: void Kit_TruthMux( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, unsigned * pCtrl, int nVars )
func kitTruthMux(pOut []uint32, pIn0 []uint32, pIn1 []uint32, pCtrl []uint32, nVars int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		pOut[w] = (pIn0[w] &^ pCtrl[w]) | (pIn1[w] & pCtrl[w])
	}
}

// kitTruthMuxPhase is Kit_TruthMuxPhase.
// C: void Kit_TruthMuxPhase( unsigned * pOut, unsigned * pIn0, unsigned * pIn1, unsigned * pCtrl, int nVars, int fComp0 )
func kitTruthMuxPhase(pOut []uint32, pIn0 []uint32, pIn1 []uint32, pCtrl []uint32, nVars int, fComp0 int) {
	for w := kitTruthWordNum(nVars) - 1; w >= 0; w-- {
		if fComp0 != 0 {
			pOut[w] = (^pIn0[w] &^ pCtrl[w]) | (pIn1[w] & pCtrl[w])
		} else {
			pOut[w] = (pIn0[w] &^ pCtrl[w]) | (pIn1[w] & pCtrl[w])
		}
	}
}

// kitTruthIthVar is Kit_TruthIthVar.
// C: void Kit_TruthIthVar( unsigned * pTruth, int nVars, int iVar )
func kitTruthIthVar(pTruth []uint32, nVars int, iVar int) {
	Masks := [5]uint32{0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000}
	nWords := kitTruthWordNum(nVars)
	if iVar < 5 {
		for k := 0; k < nWords; k++ {
			pTruth[k] = Masks[iVar]
		}
	} else {
		for k := 0; k < nWords; k++ {
			if k&(1<<uint(iVar-5)) != 0 {
				pTruth[k] = ^uint32(0)
			} else {
				pTruth[k] = 0
			}
		}
	}
}
