package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManFromIfLogicCofVars maps a cut's cofactors and encodes their mux
// connection in the mapping record using the original negative mux marker.
func Gia_ManFromIfLogicCofVars(dst *Gia_Man_t, p *If_Man_t, cut *If_Cut_t, leaves, temp, cover, mapping, records *Vec_Int_t) int32 {
	n := int32(cut.nLeaves)
	id := int(Abc_Lit2Var(cut.iCutFunc))
	if n < 1 || n > 13 || int(n) != Vec_IntSize(leaves) || p.vTtVars[n] == nil {
		panic("if: invalid cofactor cut")
	}
	variable := int32(vec.Vec_StrEntry(p.vTtVars[n], id))
	if variable >= n {
		panic("if: cofactor variable outside cut")
	}
	source := If_CutTruthW(p, cut)
	if source == nil {
		panic("if: missing cofactor truth table")
	}
	words := abcutil.Abc_TtWordNum(n)
	truth := append([]word(nil), p.puTempW[:words]...)
	var cofactors [3]int32
	for phase := 0; phase < 2; phase++ {
		tt := make([]word, words)
		vars := make([]int32, n)
		for i := range vars {
			vars[i] = int32(i)
		}
		if phase == 0 {
			abcutil.Abc_TtCofactor0p(tt, truth, words, variable)
		} else {
			abcutil.Abc_TtCofactor1p(tt, truth, words, variable)
		}
		count := abcutil.Abc_TtMinBase(tt, vars, n, max(6, n))
		vec.Vec_IntClear(temp)
		for _, pin := range vars[:count] {
			Vec_IntPush(temp, Vec_IntEntry(leaves, int(pin)))
		}
		cofactors[phase] = Gia_ManFromIfLogicCreateLut(dst, tt, temp, cover, mapping, records)
	}
	cofactors[2] = Vec_IntEntry(leaves, int(variable))
	var result, top int32
	vec.Vec_IntClear(temp)
	if cofactors[0] > 1 && cofactors[1] > 1 {
		for _, lit := range cofactors {
			Vec_IntPush(temp, lit)
		}
		// The LUT truth uses the last input as mux control.
		result = Gia_ManFromIfLogicCreateLut(dst, []word{0xcacacacacacacaca}, temp, cover, mapping, records)
		top = result
		// Replace the normal record with a mux record below.
		offset := int(vec.Vec_IntEntry(mapping, int(result>>1)))
		vec.Vec_IntShrink(records, offset)
	} else {
		assert(cofactors[0] > 1 || cofactors[1] > 1, "at least one cofactor is nonconstant")
		for _, lit := range cofactors {
			if lit > 1 {
				Vec_IntPush(temp, lit)
			}
		}
		assert(Vec_IntSize(temp) == 2, "two nonconstant mux inputs")
		switch {
		case cofactors[0] == 0:
			result = gia.Gia_ManAppendAnd2(dst, cofactors[2], cofactors[1])
		case cofactors[0] == 1:
			result = gia.Gia_ManAppendOr2(dst, cofactors[2]^1, cofactors[1])
		case cofactors[1] == 0:
			result = gia.Gia_ManAppendAnd2(dst, cofactors[2]^1, cofactors[0])
		case cofactors[1] == 1:
			result = gia.Gia_ManAppendOr2(dst, cofactors[2], cofactors[0])
		}
		top = cofactors[2]
	}
	vec.Vec_IntSetEntry(mapping, int(result>>1), int32(Vec_IntSize(records)))
	Vec_IntPush(records, int32(Vec_IntSize(temp)))
	for _, lit := range Vec_IntArray(temp) {
		Vec_IntPush(records, lit>>1)
	}
	Vec_IntPush(records, -(top >> 1))
	return result ^ int32(cut.fCompl)
}

// Gia_ManFromIfLogicAndVars maps two disjoint-support functions and joins
// them with an AND or OR, using a negative mapping marker for the top node.
func Gia_ManFromIfLogicAndVars(dst *Gia_Man_t, p *If_Man_t, cut *If_Cut_t, leaves, temp, cover, mapping, records *Vec_Int_t) int32 {
	n := int32(cut.nLeaves)
	id := int(Abc_Lit2Var(cut.iCutFunc))
	if n < 2 || n > 12 || p.pPars.NLutSize > 13 || n > 2*(p.pPars.NLutSize/2) || int(n) != Vec_IntSize(leaves) || p.vTtDecs[n] == nil {
		panic("if: invalid bi-decomposition cut")
	}
	mask := vec.Vec_IntEntry(p.vTtDecs[n], id)
	if mask <= 0 {
		panic("if: missing bi-decomposition mask")
	}
	masks := [2]int32{mask & 0xffff, (mask >> 16) & 0x3fff}
	orPhase := (mask >> 30) & 1
	source := If_CutTruthWR(p, cut)
	if source == nil {
		panic("if: missing bi-decomposition truth")
	}
	stored := vec.Vec_MemReadEntry(p.vTtMem[n], id)
	truth := make([]word, p.nTruth6Words[n])
	abcutil.Abc_TtCopy(truth, stored, p.nTruth6Words[n], orPhase)
	var parts [2]word
	abcutil.Abc_TtDeriveBiDec(truth, n, masks[0], masks[1], p.pPars.NLutSize/2, parts[0:1], parts[1:2])
	var resultParts [2]int32
	for side := 0; side < 2; side++ {
		if orPhase != 0 {
			parts[side] = ^parts[side]
		}
		vec.Vec_IntClear(temp)
		for pin := int32(0); pin < n; pin++ {
			if masks[side]>>uint(pin)&1 != 0 {
				Vec_IntPush(temp, Vec_IntEntry(leaves, int(pin)))
			}
		}
		assert(Vec_IntSize(temp) >= 1, "bi-decomposition part has inputs")
		resultParts[side] = Gia_ManFromIfLogicCreateLut(dst, []word{parts[side]}, temp, cover, mapping, records) ^ orPhase
	}
	result := gia.Gia_ManAppendAnd(dst, resultParts[0], resultParts[1]) ^ orPhase ^ Abc_LitIsCompl(cut.iCutFunc)
	vec.Vec_IntSetEntry(mapping, int(result>>1), int32(Vec_IntSize(records)))
	Vec_IntPush(records, 2)
	Vec_IntPush(records, resultParts[0]>>1)
	Vec_IntPush(records, resultParts[1]>>1)
	Vec_IntPush(records, -(result >> 1))
	return result ^ int32(cut.fCompl)
}
